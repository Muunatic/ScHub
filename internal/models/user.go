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
