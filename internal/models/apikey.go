package models

type APIKey struct {
	ID     string `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Key    string `gorm:"column:key;unique;not null" json:"key"`
	UserID string `gorm:"column:userId;unique;not null" json:"userId"`
	User   User   `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (APIKey) TableName() string {
	return "apikey"
}
