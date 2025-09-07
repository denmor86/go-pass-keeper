package usercontext

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ContextKey - тип ключа в передаваемом контексте
type ContextKey string

// UserIDContextKey - имя ключа пользователя в передаваемом контексте
var UserIDContextKey ContextKey = "userID"

// GetUserId - метод получает UUID пользователя из контекста
func GetUserId(ctx context.Context) (uuid.UUID, error) {
	var uid uuid.UUID
	userID := ctx.Value(UserIDContextKey)
	if userID == nil {
		return uid, fmt.Errorf("unknown user")
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uid, fmt.Errorf("invalid user")
	}

	return uid, nil
}

// SetUserId - метод устанавливает UUID пользователя в контекст
func SetUserId(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDContextKey, uid)
}
