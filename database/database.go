package database

import (
	"fmt"
	"log"
	"noteshare-be/config"
	"noteshare-be/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	cfg := config.AppConfig

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	logLevel := logger.Silent
	if cfg.Env == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected successfully")
	DB = db

	migrate()
}

func migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Note{},
		&models.TodoList{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrated successfully")
}