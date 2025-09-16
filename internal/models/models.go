package models

import (
	"github.com/google/uuid"
)

// User - модель пользователя
type User struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"-"`
}
