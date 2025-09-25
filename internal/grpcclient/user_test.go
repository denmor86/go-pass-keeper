package grpcclient

import (
	"context"
	pb "go-pass-keeper/pkg/proto"
	"testing"

	"go-pass-keeper/pkg/proto/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserClient_Connect(t *testing.T) {
	testCases := []struct {
		TestName      string
		ServerAddr    string
		ExpectedError string
	}{
		{
			TestName:      "Success. Connect to server",
			ServerAddr:    "localhost:8080",
			ExpectedError: "",
		},
		{
			TestName:      "Error. Invalid server address",
			ServerAddr:    "192.0.2.1:-1",
			ExpectedError: "invalid server address",
		},
		{
			TestName:      "Error. Empty server address",
			ServerAddr:    "",
			ExpectedError: "invalid server address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			uc := NewUserClient(tc.ServerAddr)

			ctx := context.Background()
			err := uc.Connect(ctx)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
				assert.Nil(t, uc.conn)
				assert.Nil(t, uc.client)
			} else {
				// В успешном случае проверяем установку полей
				if err == nil {
					assert.NotNil(t, uc.conn)
					assert.NotNil(t, uc.client)
					assert.Equal(t, ctx, uc.ctx)
				}
			}
		})
	}
}

func TestUserClient_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockUserClient(ctrl)

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		Client        pb.UserClient
		Login         string
		Password      string
		ExpectedToken string
		ExpectedSalt  string
		ExpectedError string
	}{
		{
			TestName: "Success. Register user",
			SetupMocks: func() {
				mockClient.EXPECT().Register(gomock.Any(), &pb.RegisterRequest{
					Login:    "testuser",
					Password: "testpass",
				}).Return(&pb.RegisterResponse{
					Token: "jwt-token",
					Salt:  "salt-value",
				}, nil)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "jwt-token",
			ExpectedSalt:  "salt-value",
			ExpectedError: "",
		},
		{
			TestName:      "Error. Client not connected",
			SetupMocks:    func() {},
			Client:        nil,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "client not connected",
		},
		{
			TestName: "Error. Invalid argument",
			SetupMocks: func() {
				mockClient.EXPECT().Register(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.InvalidArgument, "invalid user"),
				)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "invalid user",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().Register(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "internal error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &UserClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			token, salt, err := uc.Register(tc.Login, tc.Password)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedToken, token)
				assert.Equal(t, tc.ExpectedSalt, salt)
			}
		})
	}
}

func TestUserClient_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockUserClient(ctrl)

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		Client        pb.UserClient
		Login         string
		Password      string
		ExpectedToken string
		ExpectedSalt  string
		ExpectedError string
	}{
		{
			TestName: "Success. Login user",
			SetupMocks: func() {
				mockClient.EXPECT().Login(gomock.Any(), &pb.LoginRequest{
					Login:    "testuser",
					Password: "testpass",
				}).Return(&pb.LoginResponse{
					Token: "jwt-token",
					Salt:  "salt-value",
				}, nil)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "jwt-token",
			ExpectedSalt:  "salt-value",
			ExpectedError: "",
		},
		{
			TestName:      "Error. Client not connected",
			SetupMocks:    func() {},
			Client:        nil,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "client not connected",
		},
		{
			TestName: "Error. Unauthenticated",
			SetupMocks: func() {
				mockClient.EXPECT().Login(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Unauthenticated, "invalid credentials"),
				)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "user unauthenticated",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().Login(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:        mockClient,
			Login:         "testuser",
			Password:      "testpass",
			ExpectedToken: "",
			ExpectedSalt:  "",
			ExpectedError: "internal error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &UserClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			token, salt, err := uc.Login(tc.Login, tc.Password)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedToken, token)
				assert.Equal(t, tc.ExpectedSalt, salt)
			}
		})
	}
}
