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
type AuthFailMsg string
type ErrorMsg string
type ViewMsg string // Новое сообщение для просмотра
