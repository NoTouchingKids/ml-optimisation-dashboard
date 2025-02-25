package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserType string

const (
	UserTypeRegular UserType = "regular"
	UserTypeGuest   UserType = "guest"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     *string   `json:"email" db:"email"`
	Password  *string   `json:"-" db:"password_hash"`
	Name      string    `json:"name" db:"name"`
	Type      UserType  `json:"type" db:"user_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)
	u.Password = &hashStr
	return nil
}

func (u *User) CheckPassword(password string) bool {
	if u.Password == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password))
	return err == nil
}

// For guest users
func NewGuestUser() *User {
	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		Type:      UserTypeGuest,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
