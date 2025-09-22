package messages

import "go-pass-keeper/internal/grpcclient/settings"

// ConfigUpdatedMsg - сообщение об установке настроек
type ConfigUpdatedMsg struct {
	Connection settings.Connection
}
