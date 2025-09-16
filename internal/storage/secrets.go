package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-pass-keeper/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserStorage - хранилище секретов пользователей
type SecretStorage struct {
	db *Database // указатель на базу данных
}

// NewUserStorage - метод создаёт подключение к таблице пользователей
func NewSecretStorage(db *Database) *SecretStorage {
	return &SecretStorage{db: db}
}

// Add - метод добавляет секрет пользователя в хранилище
func (s *SecretStorage) Add(ctx context.Context, uid uuid.UUID, secret *models.Secret) (uuid.UUID, error) {
	const query = `
		INSERT INTO secrets (user_id, type, name, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id
`
	var id uuid.UUID
	err := s.db.Pool.QueryRow(ctx, query, secret.UserID, secret.Type, secret.Name, secret.Content).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(string(pgErr.Code)) {
			return uuid.Nil, ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("failed to add secret: %w", err)
	}

	return id, nil
}

func (s *SecretStorage) Get(ctx context.Context, uid uuid.UUID, name string) (*models.Secret, error) {
	const query = `
		SELECT id, type, name, content FROM secrets
		WHERE user_id = $1 AND name = $2;
`
	m := &models.Secret{}
	err := s.db.Pool.QueryRow(ctx, query, uid.String(), name).Scan(&m.ID, &m.Type, &m.Name, &m.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return m, nil
}

// Delete - метод удаляет запись секрета из таблицы
func (s *SecretStorage) Delete(ctx context.Context, uid uuid.UUID, name string) error {
	const query = `
		DELETE FROM secrets
		WHERE user_id = $1 AND name = $2;
`
	res, err := s.db.Pool.Exec(ctx, query, uid.String(), name)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// List - метод возвращает список секретов пользователя
func (s *SecretStorage) List(ctx context.Context, uid uuid.UUID) ([]*models.Secret, error) {
	const SQL = `
		SELECT id, user_id, type_secret, name FROM secrets
		WHERE user_id = $1 ORDER BY name
`
	rows, err := s.db.Pool.Query(ctx, SQL, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}
	res := make([]*models.Secret, 0)

	for rows.Next() {
		var (
			id          uuid.UUID
			user_id     uuid.UUID
			type_secret string
			name        string
		)
		err := rows.Scan(
			&id,
			&user_id,
			&type_secret,
			&name,
		)
		if err != nil {
			return res, fmt.Errorf("failed scan secret data: %w", err)
		}
		res = append(res, &models.Secret{
			ID:     id,
			UserID: user_id,
			Name:   name,
			Type:   type_secret})
	}

	return res, nil
}
