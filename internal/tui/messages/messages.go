package messages

import (
	"go-pass-keeper/internal/grpcclient/settings"
	"go-pass-keeper/internal/models"
)

// AuthSuccess - сообщение об успешной аутентификации
type AuthSuccessMsg struct {
	Email string
	Token string
}

// ConfigUpdatedMsg - сообщение об установке настроек
type ConfigUpdatedMsg struct {
	Connection settings.Connection
}

// ErrorMsg - сообщение с ошибкой
type ErrorMsg string

type GotoMainPageMsg struct{}

// Сообщения для управления секретами
type SecretAddCompleteMsg struct {
	Id       string
	Name     string
	Type     string
	Login    string
	Password string
	Content  string
	FileName string
}

type SecretAddCancelMsg struct{}

type SecretDeleteMsg struct {
	Id string
}

type SecretViewMsg struct {
	Id string
}

type SecretRefreshMsg struct {
	Secrets []*models.SecretInfo
}
