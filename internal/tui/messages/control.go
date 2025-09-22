package messages

import (
	"go-pass-keeper/internal/models"
)

// ErrorMsg - сообщение с ошибкой
type ErrorMsg string

type GotoMainPageMsg struct{}

type SecretAddCancelMsg struct{}

// SecretUpdateMsg - сообщение с необходимость обновления
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
