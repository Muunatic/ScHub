package repository

import (
	"context"

	"github.com/muunatic/schub/internal/models"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) FindByKey(ctx context.Context, key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).Where("\"userId\" = ?", userID).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) Create(ctx context.Context, key, userID string) (*models.APIKey, error) {
	apiKey := models.APIKey{
		Key:    key,
		UserID: userID,
	}
	err := r.db.WithContext(ctx).Create(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}
