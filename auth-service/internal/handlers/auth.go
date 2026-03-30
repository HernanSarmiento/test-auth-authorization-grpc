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

func GenerateToken(userID, role string, AuthorName string, privKey *ecdsa.PrivateKey, exp time.Time) (string, error) {
	id := uuid.NewString()

	claims := &auth.MyCustomsClaims{
		UserID:     userID,
		Role:       role,
		AuthorName: AuthorName,
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
