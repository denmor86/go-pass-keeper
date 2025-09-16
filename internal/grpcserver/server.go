package grpcserver

import (
	"fmt"
	"go-pass-keeper/internal/logger"
	"net"

	"google.golang.org/grpc"
)

// Service - интерфейс сервиса
type Service interface {
	// RegisterService - регистрация сервиса
	RegisterService(grpc.ServiceRegistrar)
}

// Server - структура сервера с использованием GRPC
type Server struct {
	listenAddr         string                         // адрес
	server             *grpc.Server                   // указатель на сервер
	services           []Service                      // сервисы
	unaryInterceptors  []grpc.UnaryServerInterceptor  // перехватчики простых запросов
	streamInterceptors []grpc.StreamServerInterceptor // перехватчики потоковых запросов
}

// Params - тип параметров
type Params func(*Server)

// UseListenAddr - метод устанавливает использования адреса сервера
func UseListenAddr(a string) Params {
	return func(server *Server) {
		server.listenAddr = a
	}
}

// UseServices - метод устанавливает используемые сервисы
func UseServices(in ...Service) Params {
	return func(server *Server) {
		server.services = append(server.services, in...)
	}
}

// UseUnaryInterceptors - метод устанавливает используемые интерцепторы для обычных(аутентификация, логирования и т.д)
func UseUnaryInterceptors(in ...grpc.UnaryServerInterceptor) Params {
	return func(server *Server) {
		server.unaryInterceptors = append(server.unaryInterceptors, in...)
	}
}

// UseStreamInterceptors - метод устанавливает используемые интерцепторы для потоковых запросов
func UseStreamInterceptors(in ...grpc.StreamServerInterceptor) Params {
	return func(server *Server) {
		server.streamInterceptors = append(server.streamInterceptors, in...)
	}
}

// NewServer - метод создаёт новый сервер, где params - набор параметров
func NewServer(params ...Params) *Server {
	s := &Server{}

	// применяем параметры сервера
	for _, param := range params {
		param(s)
	}

	return s
}

// RegisterServices - метод регистририет работу сервисов в GRPC
func (s *Server) RegisterServices(services ...Service) {
	for _, service := range services {
		service.RegisterService(s.server)
	}
}

// Start - метод запуска сервера
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("error listen tcp: %w", err)
	}
	// создаем сервер
	s.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.ChainStreamInterceptor(s.streamInterceptors...),
	)
	// регистрируем обработчики
	s.RegisterServices(s.services...)
	//  запускаем сервер
	go func() {
		logger.Info("Starting server on", s.listenAddr)
		if err := s.server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Error("Server error", err.Error())
		}
	}()

	return nil
}

// Stop - метод остановки сервера
func (s *Server) Stop() {
	s.server.GracefulStop()
}
