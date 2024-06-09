package models

import (
	"time"
)

type User struct {
	ID             uint          `gorm:"primaryKey"`
	Username       string        `json:"username" gorm:"unique" validate:"required"`
	Email          string        `json:"email" gorm:"unique" validate:"required,email"`
	HashedPassword string        `json:"-" validate:"required"`
	BirthDate      string        `json:"birth_date" validate:"required"`
	IsActive       bool          `json:"is_active" validate:"required"`
	IsAdmin        bool          `json:"is_admin" validate:"required"`
	RegisterDate   time.Time     `json:"register_date" validate:"required"`
	Profiles       []UserProfile `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
type UserProfile struct {
	ID              uint `gorm:"primaryKey"`
	UserID          uint `gorm:"uniqueIndex"`
	AvatarURL       string
	DisplayedName   string
	ProfileTitle    string
	Bio             string
	LastVisitDate   time.Time
	LastIP          string
	VkId            string
	TgId            string
	GoogleID        string
	TotalRedirects  int
	TotalRedirected int
	DailyRedirects  int
	DailyRedirected int
	User            User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
type UserInput struct {
	Username       string `json:"username" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	RepeatPassword string `json:"repeat_password" validate:"required,eqfield=Password"`
	BirthDate      string `json:"birth_date" validate:"required"`
}
type UserData struct {
	Email    string
	Password string
}
