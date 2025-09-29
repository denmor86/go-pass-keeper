package main

import (
	"fmt"
	"go-pass-keeper/internal/grpcclient/config"
	"go-pass-keeper/internal/tui/models"
	"go-pass-keeper/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	buildVersion string = "N/A" // номер версии
	buildDate    string = "N/A" // дата сборки
	buildCommit  string = "N/A" // хэш комита
)

func makeBuildInfo() string {
	shortCommit := buildCommit
	if len(buildCommit) >= 7 {
		shortCommit = buildCommit[:7]
	}
	return fmt.Sprintf("Build: %s(%s) %s",
		buildVersion, shortCommit, buildDate)
}

func main() {

	config := config.NewConfig("go-pass-keeper")
	p := tea.NewProgram(models.NewAppModel(config, makeBuildInfo()), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.Error("Error run GophKeeper: %v", err)
	}
}
