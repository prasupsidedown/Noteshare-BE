package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Env                 string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	JWTSecret           string
	JWTExpiryHours      int
	MaxFileSizeMB       int64
	AnthropicAPIKey     string
	// Cloudinary — menggantikan local storage
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

var AppConfig Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE_MB", "50"), 10, 64)

	AppConfig = Config{
		Port:                getEnv("PORT", "8080"),
		Env:                 getEnv("ENV", "development"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", ""),
		DBName:              getEnv("DB_NAME", "noteshare"),
		JWTSecret:           getEnv("JWT_SECRET", "fallback-secret"),
		JWTExpiryHours:      jwtExpiry,
		MaxFileSizeMB:       maxFileSize,
		AnthropicAPIKey:     getEnv("ANTHROPIC_API_KEY", ""),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
	}

	// Validasi Cloudinary
	if AppConfig.CloudinaryCloudName == "" || AppConfig.CloudinaryAPIKey == "" || AppConfig.CloudinaryAPISecret == "" {
		log.Println("⚠️  Cloudinary credentials belum diset — upload file tidak akan berfungsi")
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}