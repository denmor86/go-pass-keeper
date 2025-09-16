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

}
