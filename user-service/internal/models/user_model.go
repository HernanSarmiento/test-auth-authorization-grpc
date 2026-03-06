package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username string    `gorm:"type:varchar(50);not null"`
	Email    string    `gorm:"type:varchar(50);not null"`
	Password string    `gorm:"type:varchar(100);not null"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UserID == uuid.Nil {
		u.UserID = uuid.New()
	}
	return nil
}
