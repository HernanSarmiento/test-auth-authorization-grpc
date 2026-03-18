package handlers

import (
	"context"
	"errors"
	"log"

	pb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/models"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
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
		log.Printf("Invalid Argument: All fields must be completed")
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
			UserId:   user.UserID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
		}, Message: "Success: User created",
	}, nil

}
func (u *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserCredentialsResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Error: Email field must be provided")
	}
	userFound, err := u.repo.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Error: couldn't find user with email %s %v", req.GetEmail(), err)
		}
		return nil, status.Errorf(codes.Internal, "Error: couldn't parse users registry")
	}

	return &pb.UserCredentialsResponse{
		UserId:       userFound.UserID.String(),
		Email:        userFound.Email,
		HashPassword: userFound.Password,
	}, nil

}
func (u *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error: user_id field cannot be empty")
	}

	if req.GetFieldMask() == nil || len(req.GetFieldMask().GetPaths()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Error:Update mask is required")
	}

	userFound, err := u.repo.GetByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Error: Coudln't find user with id %s, %v", req.GetUserId(), err)
		}
		return nil, status.Errorf(codes.Internal, "Error: Coudln't parse user registry error %s", err)
	}

	if req.GetUser() != nil {
		userFound.Username = req.GetUser().Username
		userFound.Email = req.GetUser().Email
	}

	err = u.repo.Update(ctx, userFound, req.GetFieldMask())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error: couldn't update user info %v", err)
	}

	return &pb.UpdateUserResponse{
		Message: "User update successfully",
	}, nil
}
func (u *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error: user_id is required")
	}

	rows, err := u.repo.Delete(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error: coudln't parse user registry")
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "Error: user with id %s not found", req.GetUserId())
	}

	return &pb.DeleteUserResponse{
		UserId:  req.GetUserId(),
		Message: "user deleted",
	}, nil
}
