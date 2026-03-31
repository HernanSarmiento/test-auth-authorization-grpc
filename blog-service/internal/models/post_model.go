package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	PostID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title      string    `gorm:"type:varchar(100);not null"`
	Body       string    `gorm:"type:text;not null"`
	AuthorID   string    `gorm:"type:varchar(255);not null"`
	AuthorName string    `gorm:"type:varchar(50);;not null"`
	CreatedAt  time.Time `gorm:""`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	if p.PostID == uuid.Nil {
		p.PostID = uuid.New()
	}
	return nil
}
