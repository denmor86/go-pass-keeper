package services

import (
	"context"
	pb "go-pass-keeper/api/proto"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// tokenBuilder интефрейс для работы с токеном
type tokenBuilder interface {
	// BuildJWT - создание токена с ID пользователя
	BuildJWT(userID string) (string, error)
}

// User - модель сервиса пользователей
type User struct {
	pb.UnimplementedUserServer

	users storage.User
	token tokenBuilder
}

// NewUser - метод создания сервиса работы с пользователями
func NewUser(u storage.User, th tokenBuilder) *User {
	return &User{
		users: u,
		token: th,
	}
}

// Register - метод обработки запроса регистрации пользователя
func (s User) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	uid, err := s.users.Add(ctx, &models.User{Login: request.GetLogin(), Password: request.GetPassword()})
	switch err {
	case nil:
	case storage.ErrAlreadyExists:
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		return nil, status.Error(codes.Internal, err.Error())
	}

	t, err := s.token.BuildJWT(uid.String())
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Token: t}, nil
}

// Login - метод обработки запроса автооризации пользователя
func (s User) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	u, err := s.users.Get(ctx, request.GetLogin(), request.GetPassword())
	switch err {
	case nil:
	case storage.ErrNotFound:
		return nil, status.Error(codes.Unauthenticated, err.Error())
	default:
		return nil, status.Error(codes.Internal, err.Error())
	}
	t, err := s.token.BuildJWT(u.ID.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LoginResponse{Token: t}, nil
}

// RegisterService - метод регистрации сервиса
func (s *User) RegisterService(r grpc.ServiceRegistrar) {
	pb.RegisterUserServer(r, s)
}

// AuthFuncOverride - метод для кастомной обработки метода авторизации (использую для исключений проверки авторизации по токену)
func (s *User) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}
