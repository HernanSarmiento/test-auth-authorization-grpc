package handlers

import (
	"context"

	authpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/auth"
	userpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthInterface interface {
	Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error)
	Logout(ctx context.Context) error
}

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	userClient userpb.UserServiceClient
}

var Key = []byte("This-is-my-example-key")

type JWToken struct {
	key []byte
	t   *jwt.Token
	s   string
}

func ComparePasswordHash(hashedPassword []byte, password []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Error: password hashes does not coincide with submitted password %v", err)
	}
	return nil
}
