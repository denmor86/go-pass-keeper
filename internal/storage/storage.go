package storage

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"

	"github.com/google/uuid"
)

type User interface {
	// Add - добавление пользователя (возвращает идентификатор добавленного пользователя)
	Add(ctx context.Context, user *models.User) (uuid.UUID, error)
	// Get - получение пользователя (возвращает модель пользователя)
	Get(ctx context.Context, login string, password string) (*models.User, error)
}
type Secret interface {
	// Add - добавление записи с секретом
	Add(ctx context.Context, uid uuid.UUID, m *models.Secret) (*models.Secret, error)
	// Get - получение записи с секретом
	Get(ctx context.Context, uid uuid.UUID, name string) (*models.Secret, error)
	// Delete - удаление записи с секретом
	Delete(ctx context.Context, uid uuid.UUID, name string) error
	// List - список записей с секретами
	List(ctx context.Context, uid uuid.UUID) ([]*models.Secret, error)
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
