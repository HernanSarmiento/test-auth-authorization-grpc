package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	authpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/auth"
	userpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthInterface interface {
	Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error)
	Logout(ctx context.Context) error
}

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	userClient userpb.UserServiceClient
	privateKey *ecdsa.PrivateKey
}

type MyCustomsClaims struct {
	Role string `json:"role"`
	*jwt.RegisteredClaims
}

func NewAuthService(uc userpb.UserServiceClient, privKey *ecdsa.PrivateKey) *AuthService {
	return &AuthService{
		userClient: uc,
		privateKey: privKey,
	}
}

func ComparePasswordHash(hashedPassword []byte, password []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Error: password hashes does not coincide with submitted password ")
	}
	return nil
}

func LoadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private key: %w", err)
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("Error occur while parsing private key %w", err)
	}
	return key, nil
}

func LoadPublicKey(path string) (*ecdsa.PublicKey, error) {
	// 1. Leer el archivo físico (.pem)
	asn1Data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error leyendo llave pública: %w", err)
	}

	// 2. Decodificar el bloque PEM
	block, _ := pem.Decode(asn1Data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("formato de llave pública inválido")
	}
	// 3. Parsear el contenido PKIX (estándar para llaves públicas)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parseando llave pública: %w", err)
	}
	// 4. Asegurarnos de que sea una llave ECDSA (la que elegimos para el proyecto)
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("la llave no es de tipo ECDSA")
	}

	return ecdsaPub, nil
}

func GenerateToken(userID, role string, privKey *ecdsa.PrivateKey, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  exp.Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("error while signing token %w", err)
	}
	return tokenString, nil
}
func mapRoleToProto(role string) authpb.Role {
	switch role {
	case "ADMIN":
		return authpb.Role_ROLE_ADMIN
	case "USER":
		return authpb.Role_ROLE_USER
	default:
		return authpb.Role_ROLE_USER
	}
}

func (a *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error: email and password fields cannot be empty")
	}

	searchByEmail := &userpb.GetUserRequest{
		Email: req.GetEmail(),
	}

	user, err := a.userClient.GetUser(ctx, searchByEmail)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.Unauthenticated, "Error: couldn't find user with email %s", req.GetEmail())
		}
		return nil, status.Errorf(codes.Internal, "Error: Internal database error %v", err)
	}
	err = ComparePasswordHash([]byte(user.HashPassword), []byte(req.GetPassword()))
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(time.Hour * 24)

	token, err := GenerateToken(user.UserId, user.Role, a.privateKey, expirationTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating token %v", err)
	}

	return &authpb.LoginResponse{
		UserId:    string(user.UserId),
		Token:     token,
		Role:      mapRoleToProto(user.Role),
		ExpiredAt: timestamppb.New(expirationTime),
	}, nil
}

func (a *AuthService) VerifyToken(ctx context.Context, pubKey *ecdsa.PublicKey) (*MyCustomsClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Error: no metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Error: missing authorization header")
	}

	tokenString := strings.TrimPrefix(values[0], "bearer ")
	claims := &MyCustomsClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Error: unexpected method %v", t.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "Error: invalid token")
	}

	return claims, nil
}
