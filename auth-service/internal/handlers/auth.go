package handlers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	authpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/auth"
	userpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/repository"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/common/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
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
	publicKey  *ecdsa.PublicKey
	redisRepo  repository.RedisBlackListRepo
}

func NewAuthService(uc userpb.UserServiceClient, privKey *ecdsa.PrivateKey, rd repository.RedisBlackListRepo, pubKey *ecdsa.PublicKey) *AuthService {
	return &AuthService{
		userClient: uc,
		privateKey: privKey,
		redisRepo:  rd,
		publicKey:  pubKey,
	}
}

func ComparePasswordHash(hashedPassword []byte, password []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Error: password hashes does not coincide with submitted password ")
	}
	return nil
}

<<<<<<< HEAD
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

func GenerateToken(userID, role string, authorName string, privKey *ecdsa.PrivateKey, exp time.Time) (string, error) {
=======
func GenerateToken(userID, role string, privKey *ecdsa.PrivateKey, exp time.Time) (string, error) {
>>>>>>> e20bafb6104eaf37ac88d0de6bee35135c851f97
	id := uuid.NewString()

	claims := &auth.MyCustomsClaims{
		UserID:     userID,
		Role:       role,
		AuthorName: authorName,
		RegisteredClaims: &jwt.RegisteredClaims{
			ID:        id,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
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

	token, err := GenerateToken(user.UserId, user.Role, user.AuthorName, a.privateKey, expirationTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating token %v", err)
	}

	return &authpb.LoginResponse{
		UserId:     string(user.UserId),
		Token:      token,
		AuthorName: user.AuthorName,
		Role:       mapRoleToProto(user.Role),
		ExpiredAt:  timestamppb.New(expirationTime),
	}, nil
}

func (a *AuthService) LogOut(ctx context.Context, req *authpb.LogOutRequest) (*authpb.LogOutResponse, error) {

	claims, err := auth.VerifyToken(ctx, a.publicKey)
	if err != nil {
		return nil, err
	}

	jti := claims.ID

	remainingTime := time.Until(claims.ExpiresAt.Time)

	err = a.redisRepo.Save(ctx, jti, remainingTime)
	if err != nil {
		return nil, err
	}

	return &authpb.LogOutResponse{}, nil
}
