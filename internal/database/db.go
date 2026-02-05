package database

import (
	"log"
	"os"
	"go-auth-api/internal/models"
	"gorm.io/gorm"
)

func InitDB(db *gorm.DB) {
	// 1. Enable UUID support in Postgres
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	
	// 2. Automigrate models
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func SeedSuperuser(db *gorm.DB) {
	username := os.Getenv("SUPER_USER_NAME")
	var count int64
	db.Model(&models.User{}).Where("username = ?", username).Count(&count)

	if count == 0 {
		hashed, _ := models.HashPassword(os.Getenv("SUPER_USER_PASS"))
		admin := models.User{
			Username: username,
			Email:    os.Getenv("SUPER_USER_EMAIL"),
			Password: hashed,
			IsAdmin:  true,
		}
		db.Create(&admin)
		log.Println("⚡ Superuser seeded from .env")
	}
}


