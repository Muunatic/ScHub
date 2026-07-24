package models

import "time"

type PlatformName string

const (
	PlatformTikTok  PlatformName = "TikTok"
	PlatformTwitter PlatformName = "Twitter"
	PlatformYouTube PlatformName = "YouTube"
)

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
