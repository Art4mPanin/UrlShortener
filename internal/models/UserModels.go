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
	Profiles       []UserProfile `gorm:"foreignKey:UserID"`
}

type UserProfile struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"uniqueIndex"`
	AvatarURL       string `json:"avatar_url"`
	DisplayedName   string `json:"displayed_name"`
	ProfileTitle    string `json:"profile_title"`
	Bio             string `json:"bio"`
	Email           string `json:"email"`
	LastVisitDate   time.Time
	LastIP          string
	VkId            string `json:"vk_id"`
	TgId            string
	GoogleID        string `json:"google_id"`
	TotalRedirects  int    `json:"total_redirects"`
	TotalRedirected int    `json:"total_redirected"`
	DailyRedirects  int    `json:"daily_redirects"`
	DailyRedirected int    `json:"daily_redirected"`
	User            User   `gorm:"foreignKey:UserID;references:ID"`
}

type UserInput struct {
	Username       string `json:"username" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	RepeatPassword string `json:"repeat_password" validate:"required,eqfield=Password"`
	BirthDate      string `json:"birth_date" validate:"required"`
}

type PassConfirmation struct {
	OldPassword         string `json:"oldpass" validate:"required"`
	NewPassword         string `json:"newpass" validate:"required"`
	NewPassConfirmation string `json:"newpassconfirm"validate:"required,eqfield=Password"`
}

type UserData struct {
	Email    string
	Password string
}

// pizdec
type Verification struct {
	ID        uint      `gorm:"primaryKey"`
	UserEmail string    `json:"email"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"tstmp"`
}

type VerificationUserInput struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}
type ResetPassVerification struct {
	Email string `json:"email" validate:"required,email"`
}
type ResetPassInput struct {
	Email             string `json:"email" validate:"required,email"`
	Code              string `json:"code" validate:"required"`
	NewPassword       string `json:"newpass" validate:"required"`
	NewPasswordRepeat string `json:"newpassrep" validate:"required, eqfield=NewPassword"`
}
type TgVerification struct {
	AuthDate  int    `json:"auth_date"`
	FirstName string `json:"firstname"`
	TgID      int    `json:"Tg_ID"`
	LastName  string `json:"lastname"`
	Username  string `json:"username"`
	PhotoUrl  string `json:"photo_url"`
	Hash      string `json:"hash"`
}
type LinkInput struct {
	Link               string `json:"link"`
	Duration           string `json:"duration"`
	TimeBeforeRedirect string `json:"timeBeforeRedirect"`
	MaxRedirects       int    `json:"maxRedirects"`
	UniqCounter        bool   `json:"uniqRedirects"`
	Availability       bool   `json:"availability"`
}
type LinkDB struct {
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"foreignKey:UserID"`
	LongLink           string
	ShortLink          string
	ExpiresAt          time.Time
	TimeBeforeRedirect int
	MaxRedirects       int
	UniqCounter        bool
	Availability       bool
	IPLINKs            []IPLINK `gorm:"foreignKey:LinkID"`
}
type IPLINK struct {
	ID     uint `gorm:"primaryKey"`
	IP     string
	LinkID uint
	Link   LinkDB `gorm:"constraint:OnUpdate;OnDelete:CASCADE;"`
}
type LinkDBPublic struct {
	ID        uint   `gorm:"primaryKey"`
	LongLink  string `json:"longLink"`
	ExpiresAt string `json:"expiresAt"`
}
