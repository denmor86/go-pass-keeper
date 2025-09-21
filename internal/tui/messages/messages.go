package messages

import "go-pass-keeper/internal/grpcclient/settings"

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

// Сообщения для управления секретами
type SecretAddCompleteMsg struct {
	Name     string
	Type     string
	Login    string
	Password string
	Content  string
	FileName string
}

type SecretAddCancelMsg struct{}

type SecretDeleteMsg struct {
	ID int
}

type SecretViewMsg struct {
	ID int
}

type SecretRefreshMsg struct{}
