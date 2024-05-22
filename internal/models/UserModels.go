package models

import (
	"time"
)

type User struct {
	ID             uint      `gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"unique" validate:"required"`
	Email          string    `json:"email" gorm:"unique" validate:"required,email"`
	HashedPassword string    `json:"-" validate:"required"`
	BirthDate      string    `json:"birth_date" validate:"required"`
	IsActive       bool      `json:"is_active" validate:"required"`
	IsAdmin        bool      `json:"is_admin" validate:"required"`
	RegisterDate   time.Time `json:"register_date" validate:"required"`
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
