package config

import (
	"os"
	"strconv"
)

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

func LoadConfig() *Config {
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
