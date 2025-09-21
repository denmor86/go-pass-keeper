package storage

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"

	"github.com/google/uuid"
)

type User interface {
	// Add - добавление пользователя (возвращает идентификатор добавленного пользователя)
	Add(ctx context.Context, user *models.UserData) (uuid.UUID, error)
	// Get - получение пользователя (возвращает модель пользователя)
	Get(ctx context.Context, login string, password string) (*models.UserData, error)
}
type Secret interface {
	// Add - добавление записи с секретом (возвращает идентификатор добавленного секрета)
	Add(ctx context.Context, uid uuid.UUID, m *models.SecretData) (*models.SecretData, error)
	// Get - получение записи с секретом (возвращает модель секрета)
	Get(ctx context.Context, uid uuid.UUID, name string) (*models.SecretData, error)
	// Delete - удаление записи с секретом
	Delete(ctx context.Context, uid uuid.UUID, name string) error
	// List - список записей с секретами (возвращает модель информаций о секретах)
	List(ctx context.Context, uid uuid.UUID) ([]*models.SecretData, error)
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
