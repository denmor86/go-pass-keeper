package services

import (
	"context"
	"errors"
	"go-pass-keeper/internal/grpcserver/config"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/storage"
	"go-pass-keeper/internal/storage/mocks"
	"go-pass-keeper/internal/usercontext"
	"go-pass-keeper/pkg/logger"
	pb "go-pass-keeper/pkg/proto"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestNewKeeper(t *testing.T) {
	t.Run("Keeper. NewKeeper", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSecrets := mocks.NewMockSecret(ctrl)

		k := NewKeeper(mockSecrets)
		if k == nil {
			t.Errorf("Expected Keeper to be initialized with Token handler")
		}
		if mockSecrets != nil && k.secrets != mockSecrets {
			t.Errorf("Expected Keeper to be initialized with provided storage")
		}
	})
}

const user_uuid = "e29b9f80-f2b1-4191-a09c-37b05b31baaa"
const secret_uuid = "0789b8d9-cef8-4837-be99-ec36fbf5c536"

func TestAddSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSecrets := mocks.NewMockSecret(ctrl)
	config := config.DefaultConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		ExpectedError error
		Request       *pb.AddSecretRequest
		Responce      *pb.AddSecretResponse
		UserId        uuid.UUID
	}{
		{
			TestName: "Success. Add secret #1",
			SetupMocks: func() {
				mockSecrets.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(uuid.MustParse(secret_uuid), nil)
			},
			ExpectedError: nil,
			Request:       &pb.AddSecretRequest{Name: "Big secret", Type: "file", Content: []byte("0x100")},
			Responce:      &pb.AddSecretResponse{Name: "Big secret", Type: "file"},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Add secret already exists #2",
			SetupMocks: func() {
				mockSecrets.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(uuid.Nil, storage.ErrAlreadyExists)
			},
			ExpectedError: errors.New("rpc error: code = AlreadyExists desc = already exists"),
			Request:       &pb.AddSecretRequest{Name: "Big secret", Type: "file", Content: []byte("0x100")},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Add secret undefined error #3",
			SetupMocks: func() {
				mockSecrets.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("failed to add secret:"))
			},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to add secret:"),
			Request:       &pb.AddSecretRequest{Name: "Big secret", Type: "file", Content: []byte("0x100")},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Add secret unknown user #4",
			SetupMocks: func() {
			},
			ExpectedError: errors.New("rpc error: code = Unauthenticated desc = unknown user"),
			Request:       &pb.AddSecretRequest{Name: "Big secret", Type: "file", Content: []byte("0x100")},
			Responce:      nil,
			UserId:        uuid.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			k := NewKeeper(mockSecrets)

			ctx := context.Background()
			if tc.UserId != uuid.Nil {
				ctx = usercontext.SetUserId(ctx, tc.UserId)
			}

			resp, err := k.AddSecret(ctx, tc.Request)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
			if resp.String() != tc.Responce.String() {
				t.Errorf("Expected responce %v, got %v", tc.Responce.String(), resp.String())
			}
		})
	}
}

func TestGetSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSecrets := mocks.NewMockSecret(ctrl)
	config := config.DefaultConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		ExpectedError error
		Request       *pb.GetSecretRequest
		Responce      *pb.GetSecretResponse
		UserId        uuid.UUID
	}{
		{
			TestName: "Success. Get secret #1",
			SetupMocks: func() {
				mockSecrets.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Secret{UserID: uuid.MustParse(user_uuid), Name: "Big secret", Type: "file", Content: []byte("0x100")}, nil)
			},
			ExpectedError: nil,
			Request:       &pb.GetSecretRequest{Name: "Big secret"},
			Responce:      &pb.GetSecretResponse{Name: "Big secret", Type: "file", Content: []byte("0x100")},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secret already exists #2",
			SetupMocks: func() {
				mockSecrets.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, storage.ErrNotFound)
			},
			ExpectedError: errors.New("rpc error: code = NotFound desc = not found"),
			Request:       &pb.GetSecretRequest{Name: "NotFound"},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secret undefined error #3",
			SetupMocks: func() {
				mockSecrets.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to get secret:"))
			},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to get secret:"),
			Request:       &pb.GetSecretRequest{Name: "Big secret"},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secret unknown user #4",
			SetupMocks: func() {
			},
			ExpectedError: errors.New("rpc error: code = Unauthenticated desc = unknown user"),
			Request:       &pb.GetSecretRequest{Name: "Big secret"},
			Responce:      nil,
			UserId:        uuid.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			k := NewKeeper(mockSecrets)

			ctx := context.Background()
			if tc.UserId != uuid.Nil {
				ctx = usercontext.SetUserId(ctx, tc.UserId)
			}

			resp, err := k.GetSecret(ctx, tc.Request)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
			if resp.String() != tc.Responce.String() {
				t.Errorf("Expected responce %v, got %v", tc.Responce.String(), resp.String())
			}
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSecrets := mocks.NewMockSecret(ctrl)
	config := config.DefaultConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		ExpectedError error
		Request       *pb.DeleteSecretRequest
		UserId        uuid.UUID
	}{
		{
			TestName: "Success. Delete secret #1",
			SetupMocks: func() {
				mockSecrets.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			ExpectedError: nil,
			Request:       &pb.DeleteSecretRequest{Name: "Big secret"},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Delete secret already exists #2",
			SetupMocks: func() {
				mockSecrets.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(storage.ErrNotFound)
			},
			ExpectedError: errors.New("rpc error: code = NotFound desc = not found"),
			Request:       &pb.DeleteSecretRequest{Name: "NotFound"},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Delete secret undefined error #3",
			SetupMocks: func() {
				mockSecrets.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed to delete secret:"))
			},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to delete secret:"),
			Request:       &pb.DeleteSecretRequest{Name: "Big secret"},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Delete secret unknown user #4",
			SetupMocks: func() {
			},
			ExpectedError: errors.New("rpc error: code = Unauthenticated desc = unknown user"),
			Request:       &pb.DeleteSecretRequest{Name: "Big secret"},
			UserId:        uuid.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			k := NewKeeper(mockSecrets)

			ctx := context.Background()
			if tc.UserId != uuid.Nil {
				ctx = usercontext.SetUserId(ctx, tc.UserId)
			}

			_, err := k.DeleteSecret(ctx, tc.Request)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
		})
	}
}

func TestGetSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSecrets := mocks.NewMockSecret(ctrl)
	config := config.DefaultConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}

	testCases := []struct {
		TestName      string
		SetupMocks    func()
		ExpectedError error
		Request       *pb.GetSecretsRequest
		Responce      *pb.GetSecretsResponse
		UserId        uuid.UUID
	}{
		{
			TestName: "Success. Get secrets #1",
			SetupMocks: func() {
				mockSecrets.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*models.Secret{{Type: "password", Name: "Password"}, {Type: "file", Name: "File"}}, nil)
			},
			ExpectedError: nil,
			Request:       &pb.GetSecretsRequest{},
			Responce:      &pb.GetSecretsResponse{Secrets: []*pb.SecretDescription{{Type: "password", Name: "Password"}, {Type: "file", Name: "File"}}},
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secrets already exists #2",
			SetupMocks: func() {
				mockSecrets.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, storage.ErrNotFound)
			},
			ExpectedError: errors.New("rpc error: code = NotFound desc = not found"),
			Request:       &pb.GetSecretsRequest{},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secrets undefined error #3",
			SetupMocks: func() {
				mockSecrets.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to get secrets:"))
			},
			ExpectedError: errors.New("rpc error: code = Internal desc = failed to get secrets:"),
			Request:       &pb.GetSecretsRequest{},
			Responce:      nil,
			UserId:        uuid.MustParse(user_uuid),
		},
		{
			TestName: "Error. Get secrets unknown user #4",
			SetupMocks: func() {
			},
			ExpectedError: errors.New("rpc error: code = Unauthenticated desc = unknown user"),
			Request:       &pb.GetSecretsRequest{},
			Responce:      nil,
			UserId:        uuid.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			k := NewKeeper(mockSecrets)

			ctx := context.Background()
			if tc.UserId != uuid.Nil {
				ctx = usercontext.SetUserId(ctx, tc.UserId)
			}

			resp, err := k.GetSecrets(ctx, tc.Request)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}
			if resp.String() != tc.Responce.String() {
				t.Errorf("Expected responce %v, got %v", tc.Responce.String(), resp.String())
			}
		})
	}
}
