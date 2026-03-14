package db

import (
	"fmt"
	"log"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/config"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InnitDB(cfg config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB_HOST,
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_NAME,
		cfg.DB_PORT)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Couldn't open gorm %s", err)
	}
	err = db.AutoMigrate(&models.UserDomain{}, models.BlackListedToken{})
	if err != nil {
		log.Fatalf("Internal: couldn't migrate user model error %v", err)
	}

	DB = db
	return db
}
