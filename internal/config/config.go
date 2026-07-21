package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	Port    string `mapstructure:"PORT"`
	Env     string `mapstructure:"ENV"`
	Domain  string `mapstructure:"DOMAIN_URL"`
	ProdURL string `mapstructure:"PRODUCTION_URL"`

	JWTSecret string `mapstructure:"JWT_SECRET"`
	APISecret string `mapstructure:"API_SECRET"`

	DatabaseURL string `mapstructure:"DATABASE_URL"`
	DBHost      string `mapstructure:"DB_HOST"`
	DBPort      string `mapstructure:"DB_PORT"`
	DBUser      string `mapstructure:"DB_USER"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBName      string `mapstructure:"DB_NAME"`
	DBSSLMode   string `mapstructure:"DB_SSLMODE"`

	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI  string `mapstructure:"GOOGLE_REDIRECT_URI"`

	TwitterClientID          string `mapstructure:"TWITTER_CLIENT_ID"`
	TwitterClientSecret      string `mapstructure:"TWITTER_CLIENT_SECRET"`
	TwitterAppKey            string `mapstructure:"TWITTER_APP_KEY"`
	TwitterAppSecretKey      string `mapstructure:"TWITTER_APP_SECRET_KEY"`
	TwitterRedirectURI       string `mapstructure:"TWITTER_REDIRECT_URI"`
	TwitterRedirectURIOAuth1 string `mapstructure:"TWITTER_REDIRECT_URI_OAUTH1"`

	TikTokClientID     string `mapstructure:"TIKTOK_CLIENT_ID"`
	TikTokClientSecret string `mapstructure:"TIKTOK_CLIENT_SECRET"`
	TikTokRedirectURI  string `mapstructure:"TIKTOK_REDIRECT_URI"`

	ReadTimeout  time.Duration `mapstructure:"-"`
	WriteTimeout time.Duration `mapstructure:"-"`
	IdleTimeout  time.Duration `mapstructure:"-"`
}

const (
	TwitterAuthURL  = "https://x.com/i/oauth2/authorize"
	TwitterTokenURL = "https://api.x.com/2/oauth2/token"
	TikTokAuthURL   = "https://www.tiktok.com/v2/auth/authorize/"
	TikTokTokenURL  = "https://open.tiktokapis.com/v2/oauth/token/"
	YouTubeAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	YouTubeTokenURL = "https://oauth2.googleapis.com/token"
)

var AllowedOrigins = []string{
	"https://schub.typeslint.com",
	"https://typeslint.com",
}

var AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
var AllowedHeaders = []string{"Content-Type", "Authorization", "Accept", "Cookie", "Access-Control-Allow-Methods"}

const (
	DefaultTimeout     = 30 * time.Second
	TwitterUploadDelay = 5 * time.Second
)

func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: .env file read error: %v", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.Port == "" {
		cfg.Port = "3000"
	}
	if cfg.Env == "" {
		cfg.Env = "DEVELOPMENT"
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
		)
	}

	cfg.ReadTimeout = DefaultTimeout
	cfg.WriteTimeout = DefaultTimeout
	cfg.IdleTimeout = DefaultTimeout

	return cfg, nil
}

func NewGORM(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
