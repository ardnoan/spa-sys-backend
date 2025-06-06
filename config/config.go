// config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB // lowercase agar internal, bisa diakses lewat GetDB()

type Config struct {
	DB     DatabaseConfig
	JWT    JWTConfig
	Server ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret string
	Expire int // hours
}

type ServerConfig struct {
	Port string
	Env  string
}

// Load() memuat environment variable ke struct Config
func Load() *Config {
	expire, _ := strconv.Atoi(getEnv("JWT_EXPIRE", "24"))

	return &Config{
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "spa_ardnoan"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "mysecret"),
			Expire: expire,
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
	}
}

// InitDB initializes the database connection using GORM
func InitDB(cfg *Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Port,
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}

// GetDB returns the *gorm.DB instance
func GetDB() *gorm.DB {
	return db
}

// getEnv returns value from ENV or fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
