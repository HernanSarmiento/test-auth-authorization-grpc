package handlers

import (
	"context"
	"log"

	pb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/internal/models"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (res *pb.CreateUserResponse, err error) {
	if req.GetUsername() == "" || req.Email == "" || req.Password == "" {
		log.Printf("Invalid Argument: All fields must be completed %s", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Argument: All fields must be completed")
	}

	user := models.User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.Password,
	}
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		log.Printf("Internal: Coudln't hash user password %s", err)
		return nil, status.Errorf(codes.Internal, "Internal: Coudln't hash user password")
	}
	user.Password = hashedPassword

	if err := u.repo.Create(ctx, &user); err != nil {
		log.Printf("Internal: Couldn't create user, error: %s", err)
		return nil, status.Errorf(codes.Internal, "Internal: Couldn't create user, error: %s", err)
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:       user.UserID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil

}
