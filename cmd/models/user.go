package models

import (
	"time"
)

// type User123 struct {
// 	ID     uint   `json:"id" pg:"id"`
// 	FirstName     string `json:"firstName" pg:"first_name" `
// 	LastName     string `json:"lastName" pg:"last_name"`
// 	Email    string `json:"email" pg:"email"`
// 	PasswordHash string `json:"passwordHash,omitempty" pg:"password_hash"`
// 	RegisteredAt time.Time `json:"-" gorm:"column:registered_at"`
// 	LastLogin time.Time `json:"-" gorm:"column:last_login_at"`
// 	IntroDesc string `json:"introDesc" pg:"intro_desc"`
// 	Role     int    `json:"role" pg:"role"`
// 	ProfileDesc string `json:"profileDesc" pg:"profile_desc"`
// 	Token string `json:"token" pg:"token"`
// }

type UserAccount struct {
	ID int64 `json:"id"`
	FirstName string
	LastName string
	Email string
	PasswordHash string `json:"omitempty"`
	RegisteredAt time.Time `json:"-"`
	LastLoginAt time.Time `json:"-"`
	IntroDesc string
	Role int
	ProfileDesc string	
}

// validations

type SignUpUser struct {
	FirstName     string `json:"firstName" pg:"first_name" validate:"required,min=1,max=50"`
	LastName     string `json:"lastName" pg:"last_name" validate:"required,min=1,max=50"`
	Email    string `json:"email" pg:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=256"`
	IntroDesc string `json:"introDesc" pg:"intro_desc" validate:"required,max=500"`
	ProfileDesc string `json:"profileDesc" pg:"profile_desc" validate:"required,max=500"`
}

type LoginUser struct {
	Email    string `json:"email" pg:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=256"`	
}
