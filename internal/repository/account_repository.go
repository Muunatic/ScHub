package repository

import (
	"context"

	"github.com/muunatic/schub/internal/models"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindByUserIDAndPlatform(ctx context.Context, userID string, platform models.PlatformName) (*models.Account, error) {
	var acc models.Account
	err := r.db.WithContext(ctx).
		Where("\"userId\" = ? AND platform = ?", userID, string(platform)).
		First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.WithContext(ctx).
		Where("\"userId\" = ?", userID).
		Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *AccountRepository) FindPlatformsByUserID(ctx context.Context, userID string) ([]models.PlatformName, error) {
	var platforms []models.PlatformName
	err := r.db.WithContext(ctx).
		Model(&models.Account{}).
		Select("platform").
		Where("\"userId\" = ?", userID).
		Find(&platforms).Error
	if err != nil {
		return nil, err
	}
	return platforms, nil
}

func (r *AccountRepository) Upsert(ctx context.Context, acc *models.Account) error {
	var existing models.Account
	result := r.db.WithContext(ctx).
		Where("\"userId\" = ? AND platform = ?", acc.UserID, string(acc.Platform)).
		First(&existing)

	if result.Error != nil {
		return r.db.WithContext(ctx).Create(acc).Error
	}

	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"email":            acc.Email,
		"username":         acc.Username,
		"accessToken":      acc.AccessToken,
		"refreshToken":     acc.RefreshToken,
		"accessKey":        acc.AccessKey,
		"accesssSecretKey": acc.AccessSecretKey,
	}).Error
}

func (r *AccountRepository) UpdateToken(ctx context.Context, userID string, platform models.PlatformName, accessToken, refreshToken string) error {
	return r.db.WithContext(ctx).
		Model(&models.Account{}).
		Where("\"userId\" = ? AND platform = ?", userID, string(platform)).
		Updates(map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		}).Error
}

func (r *AccountRepository) Delete(ctx context.Context, userID string, platform models.PlatformName) error {
	return r.db.WithContext(ctx).
		Where("\"userId\" = ? AND platform = ?", userID, string(platform)).
		Delete(&models.Account{}).Error
}
