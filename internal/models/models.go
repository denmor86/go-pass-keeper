package models

import (
	"github.com/google/uuid"
)

// User - модель пользователя
type User struct {
	ID       uuid.UUID
	Login    string
	Password string
}

// Secret - модель секрета
type Secret struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Name    string
	Type    string
	Content []byte
}
