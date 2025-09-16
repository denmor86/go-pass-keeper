// Package app предоставляет реализацию инициализацию приложение
// Включает инициализацию конфига и логгера, создание сервера, запуск воркера.
package app

import (
	"fmt"
	"go-pass-keeper/internal/grpcserver"
	"go-pass-keeper/internal/grpcserver/config"
	interceptors "go-pass-keeper/internal/grpcserver/interceptors"
	"go-pass-keeper/internal/logger"
	"go-pass-keeper/internal/services"
	"go-pass-keeper/internal/storage"
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
	db, err := storage.NewDatabase(a.config.DatabaseDSN)
	if err != nil {
		logger.Error("Error create database", err.Error())
	}
	err = db.Initialize()
	if err != nil {
		logger.Error("Error initialize database", err.Error())
	}
	users := storage.NewUserStorage(db)
	// сервис пользователей
	us := services.NewUser(users, th)

	s := grpcserver.NewServer(
		// адрес
		grpcserver.UseListenAddr(a.config.ListenAddr),
		// перехватчики обычные запросов
		grpcserver.UseUnaryInterceptors(interceptors.CreateUnaryInterceptors(th)...),
		// перехватчики потоковых запросов
		grpcserver.UseStreamInterceptors(interceptors.CreateStreamInterceptors(th)...),
		// используемые сервисы
		grpcserver.UseServices(us),
	)

	if err := s.Start(); err != nil {
		logger.Error("Error start server", err.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Ждем сигнал остановки
	<-stop
	logger.Info("Shutdown signal received")

	close(stop)
	a.server.Stop()
	logger.Info("Shutdown completed")
}
