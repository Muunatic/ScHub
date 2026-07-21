package models

import "time"

type User struct {
	ID        string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username  string    `gorm:"column:username;unique;not null" json:"username"`
	Email     string    `gorm:"column:email;unique;not null" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"-"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	Accounts  []Account `gorm:"foreignKey:UserID" json:"-"`
	APIKey    *APIKey   `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type Account struct {
	ID              string       `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email           *string      `gorm:"column:email" json:"email,omitempty"`
	Username        *string      `gorm:"column:username" json:"username,omitempty"`
	AccessToken     string       `gorm:"column:accessToken;not null" json:"accessToken"`
	RefreshToken    *string      `gorm:"column:refreshToken" json:"refreshToken,omitempty"`
	AccessKey       *string      `gorm:"column:accessKey" json:"accessKey,omitempty"`
	AccessSecretKey *string      `gorm:"column:accesssSecretKey" json:"accessSecretKey,omitempty"`
	UserID          string       `gorm:"column:userId;not null" json:"userId"`
	Platform        PlatformName `gorm:"column:platform;type:platform_name;not null" json:"platform"`
	CreatedAt       time.Time    `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time    `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	User            User         `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Posts           []Post       `gorm:"foreignKey:AccountID" json:"-"`
	Comments        []Comment    `gorm:"foreignKey:AccountID" json:"-"`
}

func (Account) TableName() string {
	return "accounts"
}

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

type APIKey struct {
	ID     string `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Key    string `gorm:"column:key;unique;not null" json:"key"`
	UserID string `gorm:"column:userId;unique;not null" json:"userId"`
	User   User   `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (APIKey) TableName() string {
	return "apikey"
}

type PlatformName string

const (
	PlatformTikTok  PlatformName = "TikTok"
	PlatformTwitter PlatformName = "Twitter"
	PlatformYouTube PlatformName = "YouTube"
)
