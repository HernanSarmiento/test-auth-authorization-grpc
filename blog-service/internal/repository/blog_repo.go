package repository

import (
	"context"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/models"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

type BlogRepository interface {
	CountPost(ctx context.Context) (int64, error)
	CreatePost(ctx context.Context, post *models.Post) error
	GetPost(ctx context.Context, postId string) (*models.Post, error)
	GetAllPosts(ctx context.Context, limit int, offset int) ([]models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post, fieldMask *fieldmaskpb.FieldMask) error
	DeletePost(ctx context.Context, postId string) error
}

type PostgresRepo struct {
	db *gorm.DB
}

func NewPostgresRepo(db *gorm.DB) BlogRepository {
	return &PostgresRepo{db: db}
}

func (p *PostgresRepo) CountPost(ctx context.Context) (int64, error) {
	var count int64

	if err := p.db.WithContext(ctx).Model(models.Post{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (p *PostgresRepo) CreatePost(ctx context.Context, post *models.Post) error {
	return p.db.WithContext(ctx).Create(post).Error
}

func (p *PostgresRepo) GetPost(ctx context.Context, postId string) (*models.Post, error) {
	var post models.Post

	err := p.db.WithContext(ctx).First(&post, "post_id = ?", postId).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}
func (p *PostgresRepo) GetAllPosts(ctx context.Context, limit int, offset int) ([]models.Post, error) {
	var posts []models.Post

	result := p.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at desc").
		Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}
	return posts, nil
}
func (p *PostgresRepo) UpdatePost(ctx context.Context, post *models.Post, fieldMask *fieldmaskpb.FieldMask) error {

	if fieldMask == nil {
		return p.db.WithContext(ctx).Save(post).Error
	}

	return p.db.WithContext(ctx).
		Model(post).
		Select(fieldMask.GetPaths()).
		Updates(post).Error
}
func (p *PostgresRepo) DeletePost(ctx context.Context, postId string) error {
	return p.db.WithContext(ctx).Delete(models.Post{}, "post_id = ?", postId).Error
}
