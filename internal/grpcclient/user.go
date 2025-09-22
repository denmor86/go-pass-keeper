package grpcclient

import (
	"context"
	"fmt"
	"go-pass-keeper/pkg/logger"
	pb "go-pass-keeper/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// UserClient модель клиента для работы с пользователем
type UserClient struct {
	serverAddr string
	conn       *grpc.ClientConn
	client     pb.UserClient
	opts       []grpc.DialOption
	ctx        context.Context
}

// UserClientOption определяет тип для опций
type UserClientOption func(*UserClient)

// NewUserClient - метод создает новый экземпляр UserClient
func NewUserClient(serverAddr string, opts ...UserClientOption) *UserClient {
	client := &UserClient{
		serverAddr: serverAddr,
		opts:       []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	}

	// Применяем переданные опции
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// UseUserOptions - метод добавляет дополнительные grpc опции
func UseUserOptions(opts ...grpc.DialOption) UserClientOption {
	return func(uc *UserClient) {
		uc.opts = append(uc.opts, opts...)
	}
}

// Connect - метод устанавливает соединение с сервером
func (uc *UserClient) Connect(ctx context.Context) error {
	conn, err := grpc.NewClient(uc.serverAddr, uc.opts...)
	if err != nil {
		logger.Error("Failed to connect to server", err.Error())
		return fmt.Errorf("failed to connect: %w", err)
	}

	uc.conn = conn
	uc.client = pb.NewUserClient(conn)
	uc.ctx = ctx
	return nil
}

// Close - метод закрывает соединение
func (uc *UserClient) Close() error {
	if uc.conn != nil {
		return uc.conn.Close()
	}
	return nil
}

// Register - метод регистрирует нового пользователя
func (uc *UserClient) Register(login string, password string) (string, string, error) {
	if uc.client == nil {
		return "", "", fmt.Errorf("client not connected")
	}

	resp, err := uc.client.Register(uc.ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})

	switch status.Code(err) {
	case codes.OK:
		logger.Info("User registered", login)
		return resp.GetToken(), resp.GetSalt(), nil
	case codes.InvalidArgument:
		logger.Warn("invalid user", err.Error())
		return "", "", fmt.Errorf("invalid user")
	default:
		logger.Warn("User register error", err.Error())
		return "", "", fmt.Errorf("internal error")
	}
}

// Login - метод авторизует пользователя
func (uc *UserClient) Login(login, password string) (string, string, error) {
	if uc.client == nil {
		return "", "", fmt.Errorf("client not connected")
	}

	resp, err := uc.client.Login(uc.ctx, &pb.LoginRequest{
		Login:    login,
		Password: password,
	})

	switch status.Code(err) {
	case codes.OK:
		logger.Info("User is authorized", login)
		return resp.GetToken(), resp.GetSalt(), nil
	case codes.Unauthenticated:
		logger.Warn("User unauthenticated", err.Error())
		return "", "", fmt.Errorf("user unauthenticated")
	default:
		logger.Warn("User login error", err.Error())
		return "", "", fmt.Errorf("internal error")
	}
}
