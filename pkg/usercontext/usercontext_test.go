package usercontext

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserId(t *testing.T) {
	testUserID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")

	testCases := []struct {
		TestName       string
		SetupContext   func() context.Context
		ExpectedUserID uuid.UUID
		ExpectedError  error
	}{
		{
			TestName: "Success. Get user ID from context #1",
			SetupContext: func() context.Context {
				return SetUserId(context.Background(), testUserID)
			},
			ExpectedUserID: testUserID,
			ExpectedError:  nil,
		},
		{
			TestName: "Error. Get user ID from empty context #2",
			SetupContext: func() context.Context {
				return context.Background()
			},
			ExpectedUserID: uuid.Nil,
			ExpectedError:  errors.New("unknown user"),
		},
		{
			TestName: "Error. Get user ID from context with nil value #3",
			SetupContext: func() context.Context {
				return context.WithValue(context.Background(), UserIDContextKey, nil)
			},
			ExpectedUserID: uuid.Nil,
			ExpectedError:  errors.New("unknown user"),
		},
		{
			TestName: "Error. Get user ID from context with wrong type #4",
			SetupContext: func() context.Context {
				return context.WithValue(context.Background(), UserIDContextKey, "not-a-uuid")
			},
			ExpectedUserID: uuid.Nil,
			ExpectedError:  errors.New("invalid user"),
		},
		{
			TestName: "Error. Get user ID from context with integer value #5",
			SetupContext: func() context.Context {
				return context.WithValue(context.Background(), UserIDContextKey, 12345)
			},
			ExpectedUserID: uuid.Nil,
			ExpectedError:  errors.New("invalid user"),
		},
		{
			TestName: "Success. Get user ID from context with different key but same value type #6",
			SetupContext: func() context.Context {
				ctx := context.WithValue(context.Background(), ContextKey("otherKey"), testUserID)
				return ctx
			},
			ExpectedUserID: uuid.Nil,
			ExpectedError:  errors.New("unknown user"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			ctx := tc.SetupContext()

			userID, err := GetUserId(ctx)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got: '%v'", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error: '%v', got none", tc.ExpectedError)
			} else if err != nil && tc.ExpectedError != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error: '%v', got: '%v'", tc.ExpectedError, err)
			}

			if err == nil {
				assert.Equal(t, tc.ExpectedUserID, userID, "user ID doesn't match expected value")
			} else {
				assert.Equal(t, tc.ExpectedUserID, userID, "user ID should be nil on error")
			}
		})
	}
}

func TestSetUserId(t *testing.T) {
	testUserID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	anotherUserID := uuid.MustParse("87654321-4321-4321-4321-cba987654321")

	testCases := []struct {
		TestName       string
		InputContext   context.Context
		InputUserID    uuid.UUID
		ExpectedUserID uuid.UUID
	}{
		{
			TestName:       "Success. Set user ID to empty context #1",
			InputContext:   context.Background(),
			InputUserID:    testUserID,
			ExpectedUserID: testUserID,
		},
		{
			TestName:       "Success. Set nil user ID to context #2",
			InputContext:   context.Background(),
			InputUserID:    uuid.Nil,
			ExpectedUserID: uuid.Nil,
		},
		{
			TestName:       "Success. Override existing user ID in context #3",
			InputContext:   SetUserId(context.Background(), testUserID),
			InputUserID:    anotherUserID,
			ExpectedUserID: anotherUserID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			newCtx := SetUserId(tc.InputContext, tc.InputUserID)

			// Проверяем, что значение установлено корректно
			retrievedUserID, err := GetUserId(newCtx)
			require.NoError(t, err, "should be able to get user ID from context")
			assert.Equal(t, tc.ExpectedUserID, retrievedUserID, "set and retrieved user ID should match")
		})
	}
}
