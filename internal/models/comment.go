package models

import "time"

type Comment struct {
	ID         string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ExternalID string    `gorm:"column:externalId;unique;not null" json:"externalId"`
	PostID     string    `gorm:"column:postId;not null" json:"postId"`
	Content    string    `gorm:"column:content;not null" json:"content"`
	Likes      int       `gorm:"column:likes;not null;default:0" json:"likes"`
	CreatedAt  time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	AccountID  string    `gorm:"column:accountId;not null" json:"accountId"`
	Account    Account   `gorm:"foreignKey:AccountID;references:ID" json:"-"`
	Post       Post      `gorm:"foreignKey:PostID;references:ID" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}
