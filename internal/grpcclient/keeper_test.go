package grpcclient

import (
	"context"
	"go-pass-keeper/internal/models"
	pb "go-pass-keeper/pkg/proto"
	"go-pass-keeper/pkg/proto/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestKeeperClient_Connect(t *testing.T) {
	testCases := []struct {
		TestName      string
		ServerAddr    string
		Token         string
		ExpectedError string
	}{
		{
			TestName:      "Success. Connect to server",
			ServerAddr:    "localhost:8080",
			Token:         "secret",
			ExpectedError: "",
		},
		{
			TestName:      "Error. Invalid server address",
			ServerAddr:    "192.0.2.1:-1",
			Token:         "secret",
			ExpectedError: "invalid server address",
		},
		{
			TestName:      "Error. Empty server address",
			ServerAddr:    "",
			Token:         "secret",
			ExpectedError: "invalid server address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			uc := NewKeeperClient(tc.ServerAddr, tc.Token)

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

func TestKeeperClient_AddSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockKeeperClient(ctrl)

	secretInfo := &models.SecretInfo{
		ID:   "secret-123",
		Name: "test-secret",
		Type: "password",
	}
	content := []byte("encrypted-content")
	pbCreatedTime := timestamppb.New(time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC))
	mdCreatedTime := time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC)
	testCases := []struct {
		TestName       string
		SetupMocks     func()
		Client         pb.KeeperClient
		Info           *models.SecretInfo
		Content        []byte
		ExpectedResult *models.SecretInfo
		ExpectedError  string
	}{
		{
			TestName: "Success. Add secret",
			SetupMocks: func() {
				mockClient.EXPECT().AddSecret(gomock.Any(), &pb.AddSecretRequest{
					Meta:    secretInfo.ToProtoMetadata(),
					Content: content,
				}).Return(&pb.AddSecretResponse{
					Meta: &pb.SecretMetadata{
						Id:      "secret-123",
						Name:    "test-secret",
						Type:    "password",
						Created: pbCreatedTime,
						Updated: pbCreatedTime,
					},
				}, nil)
			},
			Client:  mockClient,
			Info:    secretInfo,
			Content: content,
			ExpectedResult: &models.SecretInfo{
				ID:      "secret-123",
				Name:    "test-secret",
				Type:    "password",
				Created: mdCreatedTime,
				Updated: mdCreatedTime,
			},
			ExpectedError: "",
		},
		{
			TestName:       "Error. Client not connected",
			SetupMocks:     func() {},
			Client:         nil,
			Info:           secretInfo,
			Content:        content,
			ExpectedResult: nil,
			ExpectedError:  "client not connected",
		},
		{
			TestName: "Error. User unauthenticated",
			SetupMocks: func() {
				mockClient.EXPECT().AddSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Unauthenticated, "unauthenticated"),
				)
			},
			Client:         mockClient,
			Info:           secretInfo,
			Content:        content,
			ExpectedResult: nil,
			ExpectedError:  "user unauthenticated",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().AddSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:         mockClient,
			Info:           secretInfo,
			Content:        content,
			ExpectedResult: nil,
			ExpectedError:  "internal error",
		},
		{
			TestName:       "Error. Nil secret info",
			SetupMocks:     func() {},
			Client:         mockClient,
			Info:           nil,
			Content:        content,
			ExpectedResult: nil,
			ExpectedError:  "invalid info",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &KeeperClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			result, err := uc.AddSecret(tc.Info, tc.Content)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedResult, result)
			}
		})
	}
}

func TestKeeperClient_GetSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockKeeperClient(ctrl)

	secretID := "secret-123"
	content := []byte("encrypted-secret-content")
	pbCreatedTime := timestamppb.New(time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC))
	mdCreatedTime := time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		TestName        string
		SetupMocks      func()
		Client          pb.KeeperClient
		SecretID        string
		ExpectedInfo    *models.SecretInfo
		ExpectedContent []byte
		ExpectedError   string
	}{
		{
			TestName: "Success. Get secret",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecret(gomock.Any(), &pb.GetSecretRequest{
					Meta: &pb.SecretMetadata{Id: secretID},
				}).Return(&pb.GetSecretResponse{
					Meta: &pb.SecretMetadata{
						Id:      "secret-123",
						Name:    "test-secret",
						Type:    "password",
						Created: pbCreatedTime,
						Updated: pbCreatedTime,
					},
					Content: content,
				}, nil)
			},
			Client:   mockClient,
			SecretID: secretID,
			ExpectedInfo: &models.SecretInfo{
				ID:      "secret-123",
				Name:    "test-secret",
				Type:    "password",
				Created: mdCreatedTime,
				Updated: mdCreatedTime,
			},
			ExpectedContent: content,
			ExpectedError:   "",
		},
		{
			TestName: "Error. Client not connected",
			SetupMocks: func() {
				// No mocks needed for nil client
			},
			Client:          nil,
			SecretID:        secretID,
			ExpectedInfo:    nil,
			ExpectedContent: nil,
			ExpectedError:   "client not connected",
		},
		{
			TestName: "Error. User unauthenticated",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Unauthenticated, "unauthenticated"),
				)
			},
			Client:          mockClient,
			SecretID:        secretID,
			ExpectedInfo:    nil,
			ExpectedContent: nil,
			ExpectedError:   "user unauthenticated",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:          mockClient,
			SecretID:        secretID,
			ExpectedInfo:    nil,
			ExpectedContent: nil,
			ExpectedError:   "internal error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &KeeperClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			info, content, err := uc.GetSecret(tc.SecretID)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
				assert.Nil(t, info)
				assert.Nil(t, content)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedInfo, info)
				assert.Equal(t, tc.ExpectedContent, content)
			}
		})
	}
}

func TestKeeperClient_GetSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockKeeperClient(ctrl)

	pbCreatedTime := timestamppb.New(time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC))
	mdCreatedTime := time.Date(2025, time.September, 21, 10, 30, 0, 0, time.UTC)

	testCases := []struct {
		TestName       string
		SetupMocks     func()
		Client         pb.KeeperClient
		ExpectedResult []*models.SecretInfo
		ExpectedError  string
	}{
		{
			TestName: "Success. Get secrets list",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecrets(gomock.Any(), &pb.GetSecretsRequest{}).Return(
					&pb.GetSecretsResponse{
						Secrets: []*pb.SecretMetadata{
							{Id: "1", Name: "secret1", Type: "password", Created: pbCreatedTime, Updated: pbCreatedTime},
							{Id: "2", Name: "secret2", Type: "card", Created: pbCreatedTime, Updated: pbCreatedTime},
						},
					}, nil,
				)
			},
			Client: mockClient,
			ExpectedResult: []*models.SecretInfo{
				{ID: "1", Name: "secret1", Type: "password", Created: mdCreatedTime, Updated: mdCreatedTime},
				{ID: "2", Name: "secret2", Type: "card", Created: mdCreatedTime, Updated: mdCreatedTime},
			},
			ExpectedError: "",
		},
		{
			TestName:       "Error. Client not connected",
			SetupMocks:     func() {},
			Client:         nil,
			ExpectedResult: nil,
			ExpectedError:  "client not connected",
		},
		{
			TestName: "Error. User unauthenticated",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecrets(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Unauthenticated, "unauthenticated"),
				)
			},
			Client:         mockClient,
			ExpectedResult: nil,
			ExpectedError:  "user unauthenticated",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().GetSecrets(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:         mockClient,
			ExpectedResult: nil,
			ExpectedError:  "internal error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &KeeperClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			result, err := uc.GetSecrets()

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedResult, result)
			}
		})
	}
}

func TestKeeperClient_DeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockKeeperClient(ctrl)

	testCases := []struct {
		TestName       string
		SetupMocks     func()
		Client         pb.KeeperClient
		SecretID       string
		ExpectedResult string
		ExpectedError  string
	}{
		{
			TestName: "Success. Delete secret",
			SetupMocks: func() {
				mockClient.EXPECT().DeleteSecret(gomock.Any(), &pb.DeleteSecretRequest{
					Meta: &pb.SecretMetadata{Id: "secret-123"},
				}).Return(&pb.DeleteSecretResponse{
					Meta: &pb.SecretMetadata{Id: "secret-123"},
				}, nil)
			},
			Client:         mockClient,
			SecretID:       "secret-123",
			ExpectedResult: "secret-123",
			ExpectedError:  "",
		},
		{
			TestName:       "Error. Client not connected",
			SetupMocks:     func() {},
			Client:         nil,
			SecretID:       "secret-123",
			ExpectedResult: "",
			ExpectedError:  "client not connected",
		},
		{
			TestName: "Error. User unauthenticated",
			SetupMocks: func() {
				mockClient.EXPECT().DeleteSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Unauthenticated, "unauthenticated"),
				)
			},
			Client:         mockClient,
			SecretID:       "secret-123",
			ExpectedResult: "",
			ExpectedError:  "user unauthenticated",
		},
		{
			TestName: "Error. Internal error",
			SetupMocks: func() {
				mockClient.EXPECT().DeleteSecret(gomock.Any(), gomock.Any()).Return(
					nil, status.Error(codes.Internal, "internal error"),
				)
			},
			Client:         mockClient,
			SecretID:       "secret-123",
			ExpectedResult: "",
			ExpectedError:  "internal error",
		},
		{
			TestName: "Success. Delete secret with empty ID",
			SetupMocks: func() {
				mockClient.EXPECT().DeleteSecret(gomock.Any(), &pb.DeleteSecretRequest{
					Meta: &pb.SecretMetadata{Id: ""},
				}).Return(&pb.DeleteSecretResponse{
					Meta: &pb.SecretMetadata{Id: ""},
				}, nil)
			},
			Client:         mockClient,
			SecretID:       "",
			ExpectedResult: "",
			ExpectedError:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.SetupMocks()

			uc := &KeeperClient{
				client: tc.Client,
				ctx:    context.Background(),
			}

			result, err := uc.DeleteSecret(tc.SecretID)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedResult, result)
			}
		})
	}
}
