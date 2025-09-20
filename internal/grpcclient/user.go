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

func RegisterUser(ctx context.Context, serverAddr string, login string, password string) (string, error) {

	client := newUserClient(serverAddr)
	resp, err := client.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Password: password,
	})

	switch status.Code(err) {
	case codes.OK:
		logger.Info("User registered", login)
	case codes.InvalidArgument:
		logger.Warn("invalid user", err.Error())
		return "", fmt.Errorf("invalid user")
	default:
		logger.Warn("User register error", err.Error())
		return "", fmt.Errorf("internal error")
	}
	return resp.GetToken(), nil
}

// LoginUser - метод клиента для авторизации пользователя
func LoginUser(ctx context.Context, serverAddr string, login, password string) (string, error) {
	client := newUserClient(serverAddr)
	resp, err := client.Login(ctx, &pb.LoginRequest{
		Login:    login,
		Password: password,
	})

	switch status.Code(err) {
	case codes.OK:
		logger.Info("User is authorized", login)
	case codes.Unauthenticated:
		logger.Warn("User unauthenticated", err.Error())
		return "", fmt.Errorf("user unauthenticated")
	default:
		logger.Warn("User login error", err.Error())
		return "", fmt.Errorf("internal error")
	}
	return resp.GetToken(), nil
}

func newUserClient(serverAddr string) pb.UserClient {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("User client error", err.Error())
	}
	return pb.NewUserClient(conn)
}
