package models

import (
	"github.com/google/uuid"
)

// User - модель пользователя
type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
}
