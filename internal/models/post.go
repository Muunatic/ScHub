package models

import "time"

type Post struct {
	ID         string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ExternalID string    `gorm:"column:externalId;unique;not null" json:"externalId"`
	Content    string    `gorm:"column:content;not null" json:"content"`
	Attachment string    `gorm:"column:attachment;not null" json:"attachment"`
	Likes      int       `gorm:"column:likes;not null;default:0" json:"likes"`
	Comments   int       `gorm:"column:comments;not null;default:0" json:"comments"`
	Thumbnail  *string   `gorm:"column:thumbnail" json:"thumbnail,omitempty"`
	CreatedAt  time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	AccountID  string    `gorm:"column:accountId;not null" json:"accountId"`
	Account    Account   `gorm:"foreignKey:AccountID;references:ID" json:"-"`
	Comment    []Comment `gorm:"foreignKey:PostID" json:"-"`
}

func (Post) TableName() string {
	return "posts"
}
