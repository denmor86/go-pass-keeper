package models

import (
	"crypto/rand"
	"testing"

	"go-pass-keeper/pkg/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretPassword(t *testing.T) {

	testCases := []struct {
		TestName      string
		InputLogin    string
		InputPassword string
	}{
		{
			TestName:      "Success. Encrypt and decrypt password #1",
			InputLogin:    "test",
			InputPassword: "pass",
		},
		{
			TestName:      "Success. Encrypt and decrypt password #2",
			InputLogin:    "123",
			InputPassword: "321",
		},
		{
			TestName:      "Success. Encrypt and decrypt password #3",
			InputLogin:    "123111111111111111111111111111111111",
			InputPassword: "344444444444444444444444444444444421",
		},
		{
			TestName:      "Success. Encrypt and decrypt password #4",
			InputLogin:    "",
			InputPassword: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {

			key, err := crypto.MakeCryptoKey("secret", "salt")
			require.NoError(t, err, "CryptoKey failed")

			sp := NewSecretPassword(tc.InputLogin, tc.InputPassword)

			// Тестируем Encrypt
			encrypted, err := sp.Encrypt(key)
			require.NoError(t, err, "Encrypt failed")

			// Тестируем Decrypt
			if err == nil && encrypted != nil {
				newSP := &SecretPassword{}
				err = newSP.Decrypt(key, encrypted)
				require.NoError(t, err, "Decrypt failed")
				assert.Equal(t, sp, newSP, "SecretPassword not equal")
			}
		})
	}
}

func TestSecretCard(t *testing.T) {

	testCases := []struct {
		TestName    string
		InputNumber string
		InputDate   string
		InputCVV    string
		InputOwner  string
	}{
		{
			TestName:    "Success. Encrypt and decrypt card #1",
			InputNumber: "1234-5678-9012-3456",
			InputDate:   "12/25",
			InputCVV:    "123",
			InputOwner:  "Морозов",
		},
		{
			TestName:    "Success. Encrypt and decrypt card #2",
			InputNumber: "",
			InputDate:   "12/25",
			InputCVV:    "123",
			InputOwner:  "Морозов",
		},
		{
			TestName:    "Success. Encrypt and decrypt card #3",
			InputNumber: "1234-5678-9012-34561111111111111111213124314",
			InputDate:   "12/25",
			InputCVV:    "123",
			InputOwner:  "",
		},
		{
			TestName:    "Success. Encrypt and decrypt card #4",
			InputNumber: "",
			InputDate:   "",
			InputCVV:    "",
			InputOwner:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {

			key, err := crypto.MakeCryptoKey("secret", "salt")
			require.NoError(t, err, "CryptoKey failed")

			sc := NewSecretCard(tc.InputNumber, tc.InputDate, tc.InputCVV, tc.InputOwner)

			// Тестируем Encrypt
			encrypted, err := sc.Encrypt(key)
			require.NoError(t, err, "Encrypt failed")

			// Тестируем Decrypt
			if err == nil && encrypted != nil {
				newSC := &SecretCard{}
				err = newSC.Decrypt(key, encrypted)
				require.NoError(t, err, "Decrypt failed")
				assert.Equal(t, sc, newSC, "SecretCard not equal")
			}
		})
	}
}
func TestSecretText(t *testing.T) {

	testCases := []struct {
		TestName string
		testText string
	}{
		{
			TestName: "Success. Encrypt and decrypt text #1",
			testText: "test",
		},
		{
			TestName: "Success. Encrypt and decrypt text #2",
			testText: "123-11243214",
		},
		{
			TestName: "Success. Encrypt and decrypt text #3",
			testText: "123111111111111111111111111111111111dasdasdasdasdasa131231245",
		},
		{
			TestName: "Success. Encrypt and decrypt text #4",
			testText: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {

			key, err := crypto.MakeCryptoKey("secret", "salt")
			require.NoError(t, err, "CryptoKey failed")

			st := NewSecretText(tc.testText)

			// Тестируем Encrypt
			encrypted, err := st.Encrypt(key)
			require.NoError(t, err, "Encrypt failed")

			// Тестируем Decrypt
			if err == nil && encrypted != nil {
				newST := &SecretText{}
				err = newST.Decrypt(key, encrypted)
				require.NoError(t, err, "Decrypt failed")
				assert.Equal(t, st, newST, "SecretText not equal")
			}
		})
	}
}

func TestSecretBinary(t *testing.T) {

	testCases := []struct {
		TestName string
		Binary   []byte
	}{
		{
			TestName: "Success. Encrypt and decrypt binary #1",
			Binary:   []byte("test"),
		},
		{
			TestName: "Success. Encrypt and decrypt binary #2",
			Binary:   generateRandomBytes(100),
		},
		{
			TestName: "Success. Encrypt and decrypt binary #3",
			Binary:   generateRandomBytes(1000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {

			key, err := crypto.MakeCryptoKey("secret", "salt")
			require.NoError(t, err, "CryptoKey failed")

			sb := NewSecretBinary(tc.Binary)

			// Тестируем Encrypt
			encrypted, err := sb.Encrypt(key)
			require.NoError(t, err, "Encrypt failed")

			// Тестируем Decrypt
			if err == nil && encrypted != nil {
				newSB := &SecretBinary{}
				err = newSB.Decrypt(key, encrypted)
				require.NoError(t, err, "Decrypt failed")
				assert.Equal(t, sb, newSB, "SecretBinary not equal")
			}
		})
	}
}

// Генерация случайных байт
func generateRandomBytes(size int) []byte {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return bytes
}
