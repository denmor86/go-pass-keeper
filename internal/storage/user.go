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

// UserStorage - хранилище пользователей
type UserStorage struct {
	db *Database // указатель на базу данных
}

// NewUserStorage - метод создаёт подключение к таблице пользователей
func NewUserStorage(db *Database) *UserStorage {
	return &UserStorage{db: db}
}

// Add - метод добавляет пользователя в хранилище
func (s *UserStorage) Add(ctx context.Context, user *models.UserData) (uuid.UUID, error) {
	const query = `
		INSERT INTO users (login, password, salt)
		VALUES ($1, crypt($2, gen_salt('bf')), $3)
		RETURNING id
`
	var uid uuid.UUID
	err := s.db.Pool.QueryRow(ctx, query, user.Login, user.Password, user.Salt).Scan(&uid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(string(pgErr.Code)) {
			return uuid.Nil, ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("failed to add user: %w", err)
	}

	return uid, nil
}

// Get - метод извлекает пользователя из хранилища с использованием логина и пароля
func (s *UserStorage) Get(ctx context.Context, login string, password string) (*models.UserData, error) {
	const query = `
		SELECT id, login, salt FROM users
		WHERE login = $1 AND password = crypt($2, password);
`
	user := &models.UserData{}

	err := s.db.Pool.QueryRow(ctx, query, login, password).Scan(&user.ID, &user.Login, &user.Salt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
