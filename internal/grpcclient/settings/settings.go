package settings

import (
	"fmt"
)

// Connection - модель настроек подключения
type Connection struct {
	ServerURL  string `json:"server_url"`
	ServerPort string `json:"server_port"`
	Timeout    int    `json:"timeout"`
}

func (s *Connection) ServerAddress() string {
	return fmt.Sprintf("%s:%s", s.ServerURL, s.ServerPort)
}
