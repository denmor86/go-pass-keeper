// Package app предоставляет реализацию инициализацию приложение
// Включает инициализацию конфига и логгера, создание сервера, запуск воркера.
package app

import (
	"fmt"
	"go-pass-keeper/internal/config"
	"go-pass-keeper/internal/grpcserver"
	interceptors "go-pass-keeper/internal/grpcserver/interceptors"
	"go-pass-keeper/internal/logger"
	"go-pass-keeper/internal/token"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
)

// App - модель данных приложения
type App struct {
	config *config.Config
	server *grpcserver.Server
}

// NewApp - создаёт новый сервер, где params - набор параметров
func NewApp(cfg *config.Config) *App {
	return &App{config: cfg}
}

// Run - иницилизация приложения и запуска сервера обработки сообщений
func (a *App) Run() {
	if err := logger.Initialize(a.config.LogLevel); err != nil {
		panic(fmt.Sprintf("can't initialize logger: %s ", errors.Cause(err).Error()))
	}

	logger.Info(
		"Starting server config:", a.config,
	)

	th, err := token.NewJWT(a.config.JWTSecret)
	if err != nil {
		logger.Error("Error token handler", err.Error())
	}

	s := grpcserver.NewServer(
		// адрес
		grpcserver.UseListenAddr(a.config.ListenAddr),
		// перехватчики обычные запросов
		grpcserver.UseUnaryInterceptors(interceptors.CreateUnaryInterceptors(th)...),
		// перехватчики потоковых запросов
		grpcserver.UseStreamInterceptors(interceptors.CreateStreamInterceptors(th)...),
	)

	if err := s.Start(); err != nil {
		logger.Error("Error start server", err.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ждем сигнал остановки
	<-stop
	logger.Info("Shutdown signal received")
	a.shutdown()
}

// shutdown - метод остановки запущенных серверов
func (a *App) shutdown() {
	// Stop для GRPC сервера
	a.server.Stop()
	logger.Info("Shutdown completed")
}
