package settings

import (
	"fmt"
)

// Settings - модель настроек подключения
type Settings struct {
	ServerURL  string `json:"server_url"`
	ServerPort string `json:"server_port"`
	Timeout    int    `json:"timeout"`
	Secret     string // не сохраняем для секурности
}

// ServerAddress - формирование строки адреса сервера
func (s *Settings) ServerAddress() string {
	return fmt.Sprintf("%s:%s", s.ServerURL, s.ServerPort)
}
