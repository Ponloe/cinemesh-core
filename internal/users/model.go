package users

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:50;unique;not null"`
	Email        string `gorm:"size:100;unique;not null"`
	PasswordHash string `gorm:"not null"`
	AvatarURL    string
	Role         string `gorm:"default:user"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
