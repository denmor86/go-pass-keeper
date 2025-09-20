package services

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/storage"
	pb "go-pass-keeper/pkg/proto"
	"go-pass-keeper/pkg/usercontext"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Keeper - модель сервиса секретов
type Keeper struct {
	pb.UnimplementedKeeperServer

	secrets storage.Secret
}

// NewKeeper - метод создания сервиса работы с секретами
func NewKeeper(s storage.Secret) *Keeper {
	return &Keeper{
		secrets: s,
	}
}

// AddSecret - метод для добавления секрета
func (s *Keeper) AddSecret(ctx context.Context, request *pb.AddSecretRequest) (*pb.AddSecretResponse, error) {
	uid, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	m := &models.Secret{
		UserID:  uid,
		Name:    request.GetName(),
		Type:    request.GetType(),
		Content: request.GetContent(),
	}
	_, err = s.secrets.Add(ctx, uid, m)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddSecretResponse{Name: m.Name, Type: m.Type}, nil
}

// GetSecret - метод для получения секрета пользователя
func (s *Keeper) GetSecret(ctx context.Context, request *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	uid, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	secret, err := s.secrets.Get(ctx, uid, request.GetName())
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetSecretResponse{Name: secret.Name, Type: secret.Type, Content: secret.Content}, nil
}

// DeleteSecret - метод удаления секрета пользователя
func (s *Keeper) DeleteSecret(ctx context.Context, request *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	uid, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.secrets.Delete(ctx, uid, request.GetName()); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteSecretResponse{}, nil
}

// GetSecrets - метод получения информации о секретах пользователя
func (s *Keeper) GetSecrets(ctx context.Context, request *pb.GetSecretsRequest) (*pb.GetSecretsResponse, error) {
	uid, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	list, err := s.secrets.List(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.GetSecretsResponse{}
	for _, secret := range list {
		resp.Secrets = append(resp.Secrets, &pb.SecretDescription{
			Name: secret.Name,
			Type: secret.Type,
		})
	}

	return resp, nil
}

func (s *Keeper) RegisterService(r grpc.ServiceRegistrar) {
	pb.RegisterKeeperServer(r, s)
}
