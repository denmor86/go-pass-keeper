package messages

import (
	"go-pass-keeper/internal/models"
)

// ErrorMsg - сообщение с ошибкой
type ErrorMsg string

// GotoMainPageMsg - сообщение с переходом на основное окно
type GotoMainPageMsg struct{}

// SecretAddCancelMsg - сообщение с отменой добавления секрета
type SecretAddCancelMsg struct{}

// SecretUpdateMsg - сообщение с необходимость обновления окна
type SecretUpdateMsg struct{}

// SecretDeleteMsg - сообщение с идентификатором удалённого секрета
type SecretDeleteMsg struct {
	Id string
}

// SecretViewMsg - сообщение с идентификатором секрет для просмотра
type SecretViewMsg struct {
	Id string
}

// SecretRefreshMsg - сообщение с обновленным списком секретов
type SecretRefreshMsg struct {
	Secrets []*models.SecretInfo
}
