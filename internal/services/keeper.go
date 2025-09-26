package services

import (
	"context"
	"errors"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/storage"
	pb "go-pass-keeper/pkg/proto"
	"go-pass-keeper/pkg/usercontext"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	m := &models.SecretData{
		UserID:  uid,
		Name:    request.GetMeta().GetName(),
		Type:    request.GetMeta().GetType(),
		Content: request.GetContent(),
	}
	secret, err := s.secrets.Add(ctx, m)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddSecretResponse{Meta: &pb.SecretMetadata{
		Id:      secret.ID.String(),
		Name:    m.Name,
		Type:    m.Type,
		Created: timestamppb.New(secret.Created),
		Updated: timestamppb.New(secret.Updated)}}, nil
}

// GetSecret - метод для получения секрета пользователя
func (s *Keeper) GetSecret(ctx context.Context, request *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	_, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	sid, err := uuid.Parse(request.GetMeta().GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	secret, err := s.secrets.Get(ctx, sid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetSecretResponse{Meta: &pb.SecretMetadata{
		Id:      secret.ID.String(),
		Name:    secret.Name,
		Type:    secret.Type,
		Created: timestamppb.New(secret.Created),
		Updated: timestamppb.New(secret.Updated)},
		Content: secret.Content}, nil
}

// DeleteSecret - метод удаления секрета пользователя
func (s *Keeper) DeleteSecret(ctx context.Context, request *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	_, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	sid, err := uuid.Parse(request.GetMeta().GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := s.secrets.Delete(ctx, sid); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteSecretResponse{Meta: request.GetMeta()}, nil
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
		resp.Secrets = append(resp.Secrets, &pb.SecretMetadata{
			Id:      secret.ID.String(),
			Name:    secret.Name,
			Type:    secret.Type,
			Created: timestamppb.New(secret.Created),
			Updated: timestamppb.New(secret.Updated),
		})
	}

	return resp, nil
}

// EditSecret - метод для добавления секрета
func (s *Keeper) EditSecret(ctx context.Context, request *pb.EditSecretRequest) (*pb.EditSecretResponse, error) {
	uid, err := usercontext.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	m := &models.SecretData{
		UserID:  uid,
		Name:    request.GetMeta().GetName(),
		Type:    request.GetMeta().GetType(),
		Content: request.GetContent(),
	}
	secret, err := s.secrets.Edit(ctx, m)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.EditSecretResponse{Meta: &pb.SecretMetadata{
		Id:      secret.ID.String(),
		Name:    m.Name,
		Type:    m.Type,
		Created: timestamppb.New(secret.Created),
		Updated: timestamppb.New(secret.Updated)}}, nil
}

func (s *Keeper) RegisterService(r grpc.ServiceRegistrar) {
	pb.RegisterKeeperServer(r, s)
}
