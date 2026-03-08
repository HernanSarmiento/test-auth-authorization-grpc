package handlers

import (
	"context"
	"errors"
	"log"
	"time"

	pb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/blog"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/models"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
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
			PostId:    post.PostID.String(),
			Title:     post.Title,
			Body:      post.Body,
			CreatedAt: timestamppb.New(now),
		},
	}, nil
}
func (p *BlogHandler) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	if req.GetPostId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Argument: missing postid")
	}
	post, err := p.repo.GetPost(ctx, req.GetPostId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Error: post not found or doesn't exist %s", err)
		}
		return nil, status.Errorf(codes.Internal, "Error has occur while communicating with db %s", err)
	}

	return &pb.GetPostResponse{
		Post: &pb.Post{
			PostId:    post.PostID.String(),
			Title:     post.Title,
			Body:      post.Title,
			CreatedAt: timestamppb.New(post.CreatedAt),
		},
	}, nil

}
func (p *BlogHandler) GetAllPosts(ctx context.Context, req *pb.GetAllPostRequest) (*pb.GetAllPostResponse, error) {
	limit := int(req.GetResultPerPage())

	if limit <= 0 {
		limit = 12
	}

	page := int(req.GetPageNumber())
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	totalResults, err := p.repo.CountPost(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while counting post %s", err)
	}

	postsDB, err := p.repo.GetAllPosts(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error: Couldn't fetch posts from db %s", err)
	}

	totalPages := int32(totalResults + int64(limit) - 1/int64(limit))

	var protoPosts []*pb.Post

	for _, p := range postsDB {
		protoPosts = append(protoPosts, &pb.Post{
			PostId:    p.PostID.String(),
			Title:     p.Title,
			Body:      p.Body,
			CreatedAt: timestamppb.New(p.CreatedAt),
		})
	}

	return &pb.GetAllPostResponse{
		Post:         protoPosts,
		TotalResults: totalResults,
		TotalPages:   totalPages,
	}, nil
}
func (p *BlogHandler) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {

	if req.GetPostId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error: id cannot be empty")
	}

	postUpdate := models.Post{
		Title: req.Post.GetTitle(),
		Body:  req.Post.GetBody(),
	}

	pID, err := uuid.Parse(req.GetPost().GetPostId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Error:Wrong id format %v", err)
	}
	postUpdate.PostID = pID

	err = p.repo.UpdatePost(ctx, &postUpdate, req.GetUpdateMask())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Error: Couldn't find the post %s ", req.GetPost().GetPostId())
		}
		return nil, status.Errorf(codes.Internal, "Error: Couldn't update post %v", err)
	}

	return &pb.UpdatePostResponse{
		Post: &pb.Post{
			PostId: postUpdate.PostID.String(),
			Title:  postUpdate.Title,
			Body:   postUpdate.Body,
		},
	}, nil
}
func (p *BlogHandler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	if req.GetPostId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Error:postid field cannot be empty")
	}

	err := p.repo.DeletePost(ctx, req.GetPostId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Error: Couldn't find post with id %s", req.GetPostId())
		}
		return nil, status.Errorf(codes.Internal, "Error:Coudln't delete requested post %v", err)
	}
	return &pb.DeletePostResponse{
		PostId:  req.GetPostId(),
		Message: "Post deleted successfully",
	}, nil
}
