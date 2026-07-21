package repository

import (
	"context"

	"github.com/muunatic/schub/internal/models"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Upsert(ctx context.Context, comment *models.Comment) error {
	var existing models.Comment
	result := r.db.WithContext(ctx).
		Where("\"externalId\" = ? AND \"accountId\" = ?", comment.ExternalID, comment.AccountID).
		First(&existing)

	if result.Error != nil {
		return r.db.WithContext(ctx).Create(comment).Error
	}

	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"content":   comment.Content,
		"likes":     comment.Likes,
		"createdAt": comment.CreatedAt,
	}).Error
}

func (r *CommentRepository) FindByAccountID(ctx context.Context, accountID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).
		Where("\"accountId\" = ?", accountID).
		Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepository) FindByAccountIDWithPost(ctx context.Context, accountID string) ([]CommentWithPost, error) {
	var results []CommentWithPost
	err := r.db.WithContext(ctx).
		Table("comments").
		Select(`c.id, c."externalId", c."postId", c.content, c.likes, c."createdAt", c."accountId", p.content as "postContent", p.attachment as "postAttachment"`).
		Joins("INNER JOIN posts p ON p.id = c.\"postId\"").
		Where("c.\"accountId\" = ?", accountID).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

type CommentWithPost struct {
	ID             string `json:"id"`
	ExternalID     string `json:"externalId"`
	PostID         string `json:"postId"`
	Content        string `json:"content"`
	Likes          int    `json:"likes"`
	CreatedAt      string `json:"createdAt"`
	AccountID      string `json:"accountId"`
	PostContent    string `json:"postContent"`
	PostAttachment string `json:"postAttachment"`
}
