package storage

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"
)

type User interface {
	Add(ctx context.Context, login string, password string) error
	Get(ctx context.Context, login string) (*models.User, error)
}

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrAlreadyExists = errors.New("already exists")
)
