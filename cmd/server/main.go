package main

import (
	"go-pass-keeper/internal/app"
	"go-pass-keeper/internal/grpcserver/config"
	"go-pass-keeper/pkg/logger"
)

// функция main вызывается автоматически при запуске приложения
func main() {
	config := config.NewConfig()
	defer logger.Sync()

	a := app.NewApp(config)

	a.Run()
}
