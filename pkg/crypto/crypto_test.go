package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSalt(t *testing.T) {
	testCases := []struct {
		Name string
	}{
		{
			Name: "GenerateSalt Unique #1",
		},
		{
			Name: "GenerateSalt Unique #2",
		},
	}

	var salts []string
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			salt, err := GenerateSalt()
			require.NoError(t, err, "GenerateSalt failed")
			require.NotEmpty(t, salt, "Generated salt should not be empty")

			// Проверяем что salt - валидный base64
			_, err = MakeCryptoKey("password", salt)
			require.NoError(t, err, "Generated salt should be valid base64")

			// Проверяем уникальность
			for _, existingSalt := range salts {
				assert.NotEqual(t, existingSalt, salt, "Generated salts should be unique")
			}
			salts = append(salts, salt)
		})
	}
}

func TestMakeCryptoKey(t *testing.T) {
	testCases := []struct {
		Name     string
		Password string
		Salt     string
	}{
		{
			Name:     "Success #1",
			Password: "password1",
			Salt:     "MDEyMzQ1Njc4OTAxMjM0NQ==",
		},
		{
			Name:     "Success #2",
			Password: "password2",
			Salt:     "OTg3NjU0MzIxMDk4NzY1NA==",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			key1, err := MakeCryptoKey(tc.Password, tc.Salt)
			require.NoError(t, err, "DeriveKey failed")
			require.Len(t, key1, 32, "Derived key should be 32 bytes")

			// Проверяем детерминированность
			key2, err := MakeCryptoKey(tc.Password, tc.Salt)
			require.NoError(t, err, "Second DeriveKey failed")
			assert.Equal(t, key1, key2, "Derived keys should be identical for same inputs")
		})
	}
}
func TestEncrypt(t *testing.T) {
	password := "password"
	salt := "MDEyMzQ1Njc4OTAxMjM0NQ==" // Фиксированная соль

	testCases := []struct {
		Name    string
		Content []byte
	}{
		{
			Name:    "Success #1",
			Content: []byte("Тестовое сообщение"),
		},
		{
			Name: "Success #2",
			Content: []byte(`Затем у вас будет 3 недели на написание кода. \n
			Вы можете и раньше отправить свои наработки на проверку ментору, чтобы иметь возможность доработать проект в срок. \n
			Дедлайн отправки наступает в последний день 12-го спринта)`),
		},
		{
			Name:    "Empty/Nil #3",
			Content: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Формируем ключ
			key, err := MakeCryptoKey(password, salt)
			require.NoError(t, err, "MakeCryptoKey failed")

			// Шифруем с ключом
			data, err := Encrypt(key, tc.Content)
			require.NoError(t, err, "Encrypt failed")
			require.NotEmpty(t, data, "Ciphertext should not be empty")

			// Расшифровываем с ключом
			decrypted, err := Decrypt(key, data)
			require.NoError(t, err, "Decrypt failed")
			assert.Equal(t, tc.Content, decrypted, "Decrypted content should match original")
		})
	}
}

func TestDecrypt(t *testing.T) {
	password := "test-password"
	salt, err := GenerateSalt()
	require.NoError(t, err)

	key, err := MakeCryptoKey(password, salt)
	require.NoError(t, err)

	testData := []byte("secret data for encryption test")

	testCases := []struct {
		TestName      string
		SetupData     func() []byte
		Key           []byte
		ExpectedData  []byte
		ExpectedError string
	}{
		{
			TestName: "Success. Decrypt valid data",
			SetupData: func() []byte {
				encrypted, err := Encrypt(key, testData)
				require.NoError(t, err)
				return encrypted
			},
			Key:           key,
			ExpectedData:  testData,
			ExpectedError: "",
		},
		{
			TestName: "Error. Wrong key",
			SetupData: func() []byte {
				encrypted, err := Encrypt(key, testData)
				require.NoError(t, err)
				return encrypted
			},
			Key:           []byte("wrong-key-32-bytes-long-key!"),
			ExpectedData:  nil,
			ExpectedError: "invalid key size",
		},
		{
			TestName: "Error. Data too short",
			SetupData: func() []byte {
				return []byte("short")
			},
			Key:           key,
			ExpectedData:  nil,
			ExpectedError: "encrypted data too short",
		},
		{
			TestName: "Error. Empty data",
			SetupData: func() []byte {
				return []byte{}
			},
			Key:           key,
			ExpectedData:  nil,
			ExpectedError: "encrypted data too short",
		},
		{
			TestName: "Error. Nil data",
			SetupData: func() []byte {
				return nil
			},
			Key:           key,
			ExpectedData:  nil,
			ExpectedError: "encrypted data too short",
		},
		{
			TestName: "Error. Corrupted data",
			SetupData: func() []byte {
				encrypted, err := Encrypt(key, testData)
				require.NoError(t, err)
				// Повреждаем данные (изменяем последний байт)
				encrypted[len(encrypted)-1] ^= 0xFF
				return encrypted
			},
			Key:           key,
			ExpectedData:  nil,
			ExpectedError: "decryption failed",
		},
		{
			TestName: "Error. Tampered nonce",
			SetupData: func() []byte {
				encrypted, err := Encrypt(key, testData)
				require.NoError(t, err)
				// Изменяем nonce (первые 12 байт для GCM)
				if len(encrypted) >= 12 {
					encrypted[0] ^= 0xFF
				}
				return encrypted
			},
			Key:           key,
			ExpectedData:  nil,
			ExpectedError: "decryption failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			data := tc.SetupData()

			result, err := Decrypt(tc.Key, data)

			if tc.ExpectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.ExpectedData, result)
			}
		})
	}
}
