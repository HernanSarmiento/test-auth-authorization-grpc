package handlers

import (
	"context"
	"log"
	"time"

	pb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/blog"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/models"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BlogHandler struct {
	pb.UnimplementedBlogServiceServer
	repo repository.BlogRepository
}

func NewBlogHandler(repo repository.BlogRepository) *BlogHandler {
	return &BlogHandler{repo: repo}
}

func (p *BlogHandler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {

	if req.GetTitle() == "" || req.GetBody() == "" {
		log.Printf("Invalid Argument: All fields must be completed")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Argument: All fields must be completed")
	}
	now := time.Now()
	post := models.Post{
		Title:     req.Title,
		Body:      req.Body,
		CreatedAt: now,
	}
	if err := p.repo.CreatePost(ctx, &post); err != nil {
		log.Printf("Error couldn't create post error %s", err)
		return nil, err
	}

	return &pb.CreatePostResponse{
		Post: &pb.Post{
			Id:        post.PostID.String(),
			Title:     post.Title,
			Body:      post.Body,
			CreatedAt: timestamppb.New(now),
		},
	}, nil
}
