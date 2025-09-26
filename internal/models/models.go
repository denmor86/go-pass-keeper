package models

import (
	"time"

	"github.com/google/uuid"
)

// UserData - модель пользователя из БД
type UserData struct {
	ID       uuid.UUID
	Login    string
	Password string
	Salt     string
}

// SecretData - модель секрета  из БД
type SecretData struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Name    string
	Type    string
	Created time.Time
	Updated time.Time
	Content []byte
}
