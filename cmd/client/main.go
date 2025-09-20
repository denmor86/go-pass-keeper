package main

import (
	"go-pass-keeper/internal/grpcclient/config"
	"go-pass-keeper/internal/tui/models"
	"go-pass-keeper/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	config := config.NewConfig("go-pass-keeper")
	p := tea.NewProgram(models.NewAppModel(config), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.Error("Error run GophKeeper: %v", err)
	}
}
