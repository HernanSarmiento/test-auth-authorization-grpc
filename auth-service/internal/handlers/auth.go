package handlers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"time"

	authpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/auth"
	userpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/golang-jwt/jwt/v5"
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
}

func NewAuthService(uc userpb.UserServiceClient, privKey *ecdsa.PrivateKey) *AuthService {
	return &AuthService{
		userClient: uc,
		privateKey: privKey,
	}
}

type JWToken struct {
	Value     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
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
