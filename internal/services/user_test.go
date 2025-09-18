package services

import (
	"context"
	"errors"
	pb "go-pass-keeper/api/proto"
	"go-pass-keeper/internal/grpcserver/config"
	"go-pass-keeper/internal/logger"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/storage"
	"go-pass-keeper/internal/token"
	"testing"
	"time"

	"go-pass-keeper/internal/storage/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewUser(t *testing.T) {
	t.Run("User. NewUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUsers := mocks.NewMockUser(ctrl)

		config := config.DefaultConfig()

		th, err := token.NewJWT(config.JWTSecret)
		if err != nil {
			logger.Error("Error token handler", err.Error())
		}

		u := NewUser(mockUsers, th)
		if u == nil || th == nil {
			t.Errorf("Expected Users to be initialized with Token handler")
		}
		if mockUsers != nil && u.users != mockUsers {
			t.Errorf("Expected Users to be initialized with provided storage")
		}
	})
}

const uid = "0789b8d9-cef8-4837-be99-ec36fbf5c536"

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsers := mocks.NewMockUser(ctrl)
	config := config.DefaultConfig()

	th, err := token.NewJWT(config.JWTSecret)
	if err != nil {
		logger.Error("Error token handler", err.Error())
	}

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		ExpectedError error
		User          *pb.RegisterRequest
		UserID        string
	}{
		{
			TestName: "Success. Register user #1",
			SetupMocks: func() {
				mockUsers.EXPECT().Add(gomock.Any(), gomock.Any()).Return(uuid.MustParse(uid), nil)
			},
			ExpectedError: nil,
			User:          &pb.RegisterRequest{Login: "mda", Password: "test_pass"},
			UserID:        uid,
		},
		{
			TestName: "Error. Register user already exists #2",
			SetupMocks: func() {
				mockUsers.EXPECT().Add(gomock.Any(), gomock.Any()).Return(uuid.Nil, storage.ErrAlreadyExists)
			},
			ExpectedError: errors.New("rpc error: code = InvalidArgument desc = already exists"),
			User:          &pb.RegisterRequest{Login: "mda", Password: "test_pass"},
		},
		{
			TestName: "Error. Register user undefined error #3",
			SetupMocks: func() {
				mockUsers.EXPECT().Add(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("failed to add user"))
			},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to add user"),
			User:          &pb.RegisterRequest{Login: "mda", Password: "test_pass"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			u := NewUser(mockUsers, th)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := u.Register(ctx, tc.User)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
			if err == nil {
				// Парсим токен для проверки его claims
				claims, err := th.ParseJWT(resp.GetToken())
				require.NoError(t, err, "invalid claims")

				assert.Equal(t, tc.UserID, claims.Id, "user ID in claims doesn't match")
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsers := mocks.NewMockUser(ctrl)

	config := config.DefaultConfig()

	th, err := token.NewJWT(config.JWTSecret)
	if err != nil {
		logger.Error("Error token handler", err.Error())
	}

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		User          *pb.LoginRequest
		ExpectedError error
		UserID        string
	}{
		{
			TestName: "AuthenticateUser Success #1",
			SetupMocks: func() {
				mockUsers.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.User{ID: uuid.MustParse(uid), Login: "mda"}, nil)
			},
			User:          &pb.LoginRequest{Login: "mda", Password: "test_pass"},
			ExpectedError: nil,
			UserID:        uid,
		},
		{
			TestName: "AuthenticateUser NotFound #2",
			SetupMocks: func() {
				mockUsers.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, storage.ErrNotFound)
			},
			User:          &pb.LoginRequest{Login: "mda", Password: "test_pass"},
			ExpectedError: errors.New("rpc error: code = Unauthenticated desc = not found"),
		},
		{
			TestName: "AuthenticateUser InvalidPassword #3",
			SetupMocks: func() {
				mockUsers.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to add user"))
			},
			User:          &pb.LoginRequest{Login: "mda", Password: "test_pass"},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to add user"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			u := NewUser(mockUsers, th)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := u.Login(ctx, tc.User)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
			if err == nil {
				// Парсим токен для проверки его claims
				claims, err := th.ParseJWT(resp.GetToken())
				require.NoError(t, err, "invalid claims")

				assert.Equal(t, tc.UserID, claims.Id, "user ID in claims doesn't match")
			}
		})
	}
}
