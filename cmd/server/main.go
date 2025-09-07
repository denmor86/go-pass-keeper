package main

import (
	"go-pass-keeper/internal/app"
	"go-pass-keeper/internal/config"
	"go-pass-keeper/internal/logger"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	config := config.NewConfig()
	defer logger.Sync()

	a := app.NewApp(config)

	a.Run()
}
