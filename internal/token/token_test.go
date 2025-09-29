package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userID = "0789b8d9-cef8-4837-be99-ec36fbf5c536"

func TestBuildJWT(t *testing.T) {
	// Определяем тестовые случаи
	testCases := []struct {
		name      string
		userID    string
		secret    string
		wantError bool
	}{
		{
			name:      "Successful test #1 (good)",
			userID:    userID,
			secret:    "valid-secret-key",
			wantError: false,
		},
		{
			name:      "Empty user #2 (bad)",
			userID:    "",
			secret:    "valid-secret-key",
			wantError: false,
		},
		{
			name:      "Empty secret #3 (bad)",
			userID:    userID,
			secret:    "",
			wantError: true, // пустой секретный ключ должен вызывать ошибку
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			th, err := NewJWT(tc.secret)
			if err != nil {
				assert.True(t, tc.wantError, "check empty secret")
				return
			}
			// Вызываем тестируемую функцию
			tokenString, err := th.BuildJWT(tc.userID)

			// Проверяем ожидаемую ошибку
			if tc.wantError {
				assert.Error(t, err, "expected error but got none")
				assert.Empty(t, tokenString, "token should be empty when error occurs")
				return
			}

			// Если ошибка не ожидается
			require.NoError(t, err, "unexpected error")
			assert.NotEmpty(t, tokenString, "token should not be empty")

			// Парсим токен для проверки его claims
			claims, err := th.ParseJWT(tokenString)
			require.NoError(t, err, "invalid claims")

			assert.Equal(t, tc.userID, claims.Id, "user ID in claims doesn't match")
			assert.WithinDuration(t, time.Now().Add(JWTExpire), time.Unix(claims.ExpiresAt, 0), time.Second, "expiration time is not correct")
		})
	}
}

func TestParseJWT(t *testing.T) {
	// Создадим валидный токен для тестов
	validUserID := "mda"
	th, err := NewJWT("valid-secret-key")
	require.NoError(t, err, "failed to create token handler")
	validToken, err := th.BuildJWT(validUserID)
	require.NoError(t, err, "failed to create valid test token")

	testCases := []struct {
		name        string
		tokenString string
		secret      string
		wantError   bool
		errorText   string
	}{
		{
			name:        "Successful test #1",
			tokenString: validToken,
			wantError:   false,
		},
		{
			name:        "Empty token #2",
			tokenString: "",
			wantError:   true,
			errorText:   "token contains an invalid number of segments",
		},
		{
			name:        "Invalid token #3",
			tokenString: "invalid",
			wantError:   true,
			errorText:   "token contains an invalid number of segments",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := th.ParseJWT(tc.tokenString)

			if tc.wantError {
				require.Error(t, err, "expected error but got none")
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText, "unexpected error text")
				}
				assert.Nil(t, claims, "claims should be nil when error occurs")
				return
			}

			require.NoError(t, err, "unexpected error")
			require.NotNil(t, claims, "claims should not be nil")
			assert.Equal(t, validUserID, claims.Id, "user ID in claims doesn't match")
		})
	}
}

func TestJWT_DecodeUserId(t *testing.T) {

	jwtSecret := "valid-secret-key"
	j, err := NewJWT(jwtSecret)
	require.NoError(t, err)

	userID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		TestName       string
		SetupMocks     func() string
		Token          string
		ExpectedUserID string
		ExpectedError  string
	}{
		{
			TestName: "Success. Decode valid token",
			SetupMocks: func() string {
				token, err := j.BuildJWT(userID)
				require.NoError(t, err)
				return token
			},
			ExpectedUserID: userID,
			ExpectedError:  "",
		},
		{
			TestName: "Error. Invalid token signature",
			SetupMocks: func() string {
				// Создаем токен с другим секретом
				otherJWT, _ := NewJWT("different-secret-key")
				token, err := otherJWT.BuildJWT(userID)
				require.NoError(t, err)
				return token
			},
			ExpectedUserID: "",
			ExpectedError:  "signature is invalid",
		},
		{
			TestName: "Error. Empty token",
			SetupMocks: func() string {
				return ""
			},
			Token:          "",
			ExpectedUserID: "",
			ExpectedError:  "token contains an invalid number of segments",
		},
		{
			TestName: "Error. Expired token",
			SetupMocks: func() string {
				// Создаем токен с истекшим сроком
				now := time.Now().Add(-2 * time.Hour)
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
					StandardClaims: jwt.StandardClaims{
						Id:        userID,
						ExpiresAt: now.Unix(),
						IssuedAt:  now.Unix(),
						NotBefore: now.Unix(),
					},
				})
				tokenString, err := token.SignedString([]byte(jwtSecret))
				require.NoError(t, err)
				return tokenString
			},
			ExpectedUserID: "",
			ExpectedError:  "token is expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			token := tc.SetupMocks()

			userID, err := j.DecodeUserId(token)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
				assert.Equal(t, "", userID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedUserID, userID)
			}
		})
	}
}
