package repository

import (
	"context"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/models"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Get(ctx context.Context, userId string) (*models.User, error)
	Update(ctx context.Context, user *models.User, fieldMask *fieldmaskpb.FieldMask) error
	Delete(ctx context.Context, userId string) error
}

type postgresRepo struct {
	db *gorm.DB
}

func NewPostgresRepo(db *gorm.DB) UserRepository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresRepo) Get(ctx context.Context, userId string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "user_id = ?", userId).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepo) Update(ctx context.Context, user *models.User, fieldMask *fieldmaskpb.FieldMask) error {
	if fieldMask == nil {
		return r.db.WithContext(ctx).Save(user).Error
	}
	return r.db.WithContext(ctx).
		Model(user).
		Select(fieldMask.GetPaths()).
		Updates(user).Error
}

func (r *postgresRepo) Delete(ctx context.Context, userId string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "user_id = ?", userId).Error
}
