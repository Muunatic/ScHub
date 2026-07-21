package repository

import (
	"context"

	"github.com/muunatic/schub/internal/models"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Upsert(ctx context.Context, post *models.Post) error {
	var existing models.Post
	result := r.db.WithContext(ctx).
		Where("\"externalId\" = ?", post.ExternalID).
		First(&existing)

	if result.Error != nil {
		return r.db.WithContext(ctx).Create(post).Error
	}

	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"content":    post.Content,
		"attachment": post.Attachment,
		"likes":      post.Likes,
		"createdAt":  post.CreatedAt,
	}).Error
}

func (r *PostRepository) FindByAccountID(ctx context.Context, accountID string) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.WithContext(ctx).
		Where("\"accountId\" = ?", accountID).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) FindByExternalID(ctx context.Context, externalID string) (*models.Post, error) {
	var p models.Post
	err := r.db.WithContext(ctx).
		Select("id").
		Where("\"externalId\" = ?", externalID).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostRepository) FindLatestByAccountID(ctx context.Context, accountID string) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.WithContext(ctx).
		Where("\"accountId\" = ?", accountID).
		Order("\"createdAt\" DESC").
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}
