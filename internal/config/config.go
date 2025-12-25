package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Storage   StorageConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  string
	Debug bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret             string
	ExpiryHours        time.Duration
	RefreshExpiryHours time.Duration
}

type StorageConfig struct {
	Path          string
	UploadMaxSize int64
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type RateLimitConfig struct {
	Requests int
	Duration int
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Set defaults
	viper.SetDefault("APP_NAME", "investify-api")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_DEBUG", true)
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "investify")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "zx0011")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_TIMEZONE", "Africa/Nairobi")
	viper.SetDefault("JWT_SECRET", "change-this-secret-in-production")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("JWT_REFRESH_EXPIRY_HOURS", 168)
	viper.SetDefault("STORAGE_PATH", "./storage")
	viper.SetDefault("UPLOAD_MAX_SIZE", 10485760)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("CORS_ALLOWED_HEADERS", []string{})
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", 60)

	return &Config{
		App: AppConfig{
			Name:  viper.GetString("APP_NAME"),
			Env:   viper.GetString("APP_ENV"),
			Port:  viper.GetString("APP_PORT"),
			Debug: viper.GetBool("APP_DEBUG"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Name:     viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
			Timezone: viper.GetString("DB_TIMEZONE"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			ExpiryHours:        time.Duration(viper.GetInt("JWT_EXPIRY_HOURS")) * time.Hour,
			RefreshExpiryHours: time.Duration(viper.GetInt("JWT_REFRESH_EXPIRY_HOURS")) * time.Hour,
		},
		Storage: StorageConfig{
			Path:          viper.GetString("STORAGE_PATH"),
			UploadMaxSize: viper.GetInt64("UPLOAD_MAX_SIZE"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetStringSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods: viper.GetStringSlice("CORS_ALLOWED_METHODS"),
			AllowedHeaders: viper.GetStringSlice("CORS_ALLOWED_HEADERS"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: viper.GetInt("RATE_LIMIT_DURATION"),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" port=" + c.Port +
		" sslmode=" + c.SSLMode +
		" TimeZone=" + c.Timezone
}
