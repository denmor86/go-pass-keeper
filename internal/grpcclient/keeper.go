package grpcclient

import (
	"context"
	"fmt"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/pkg/logger"
	pb "go-pass-keeper/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// KeeperClient модель клиента для работы с секретами
type KeeperClient struct {
	serverAddr string
	conn       *grpc.ClientConn
	client     pb.KeeperClient
	opts       []grpc.DialOption
	ctx        context.Context
	token      string
}

// KeeperClientOption определяет тип для опций
type KeeperClientOption func(*KeeperClient)

// NewKeeperClient - метод создает новый экземпляр KeeperClient
func NewKeeperClient(serverAddr string, token string, opts ...KeeperClientOption) *KeeperClient {
	client := &KeeperClient{
		serverAddr: serverAddr,
		opts:       []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		token:      token,
	}

	// Применяем переданные опции
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// UseKeeperOptions - метод добавляет дополнительные grpc опции
func UseKeeperOptions(opts ...grpc.DialOption) KeeperClientOption {
	return func(uc *KeeperClient) {
		uc.opts = append(uc.opts, opts...)
	}
}

// Connect - метод устанавливает соединение с сервером
func (uc *KeeperClient) Connect(ctx context.Context) error {
	conn, err := grpc.NewClient(uc.serverAddr, uc.opts...)
	if err != nil {
		logger.Error("Failed to connect to server", err.Error())
		return fmt.Errorf("failed to connect: %w", err)
	}

	uc.conn = conn
	uc.client = pb.NewKeeperClient(conn)
	uc.ctx = ctx
	return nil
}

// Close - метод закрывает соединение
func (uc *KeeperClient) Close() error {
	if uc.conn != nil {
		return uc.conn.Close()
	}
	return nil
}

// GetSecrets - метод получает список секретов пользователя
func (uc *KeeperClient) GetSecrets() ([]*models.SecretInfo, error) {
	if uc.client == nil {
		return nil, fmt.Errorf("client not connected")
	}
	resp, err := uc.client.GetSecrets(uc.ctx, &pb.GetSecretsRequest{})
	switch status.Code(err) {
	case codes.OK:
		return models.ToModel(resp), nil
	case codes.Unauthenticated:
		logger.Warn("User unauthenticated", err.Error())
		return nil, fmt.Errorf("user unauthenticated")
	default:
		logger.Warn("User login error", err.Error())
		return nil, fmt.Errorf("internal error")
	}
}
