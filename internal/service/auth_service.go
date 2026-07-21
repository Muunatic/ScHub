package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	apiKeyRepo *repository.APIKeyRepository
	cfg        *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, apiKeyRepo *repository.APIKeyRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
		cfg:        cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, email, username, password string) (string, error) {
	existingUser, _ := s.userRepo.FindByEmailOrUsername(ctx, email, username)
	if existingUser != nil {
		return "", errors.New("email or username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.Create(ctx, email, username, string(hashedPassword))
	if err != nil {
		return "", err
	}

	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	apiKey := base64.StdEncoding.EncodeToString(keyBytes)

	if _, err := s.apiKeyRepo.Create(ctx, apiKey, user.ID); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return tokenString, user.ID, nil
}
