package storage

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"

	"github.com/google/uuid"
)

type User interface {
	Add(ctx context.Context, user *models.User) (uuid.UUID, error)
	Get(ctx context.Context, login string, password string) (*models.User, error)
}

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrAlreadyExists = errors.New("already exists")
)
