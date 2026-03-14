package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserDomain struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Role      Role      `gorm:"type:varchar(20);not null;default:'user'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
