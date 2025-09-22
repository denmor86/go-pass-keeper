package models

import (
	"encoding/json"
	"fmt"
	"go-pass-keeper/pkg/crypto"
)

const (
	SecretPasswordType = "password"
	SecretCardType     = "card"
	SecretTextType     = "text"
	SecretBinaryType   = "binary"
)

// SecretPassword - данные логин/пароль
type SecretPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SecretCard - данные банковская карта
type SecretCard struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	CVV    string `json:"cvv"`
	Owner  string `json:"owner"`
}

// SecretCrypter - интерфейс для обобщения типов секретных данных
type SecretCrypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte, []byte) error
}

// NewSecretPassword - базовый конструктор
func NewSecretPassword(login, password string) *SecretPassword {
	return &SecretPassword{
		Login:    login,
		Password: password,
	}
}

// Encrypt - метод шифрует данные пароля и логина
func (sp *SecretPassword) Encrypt(key []byte) ([]byte, error) {
	data, err := json.Marshal(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	return crypto.Encrypt(key, data)
}

// Decrypt - метод рашифровывает данные пароля и логина
func (sp *SecretPassword) Decrypt(key []byte, content []byte) error {
	data, err := crypto.Decrypt(key, content)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, sp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	return nil
}

// NewSecretCard - базовый конструктор
func NewSecretCard(number, date, cvv, owner string) *SecretCard {
	return &SecretCard{
		Number: number,
		Date:   date,
		CVV:    cvv,
		Owner:  owner,
	}
}

// Encrypt - метод шифрует данные кредитной карты
func (sc *SecretCard) Encrypt(key []byte) ([]byte, error) {
	data, err := json.Marshal(sc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	return crypto.Encrypt(key, data)
}

// Decrypt - метод рашифровывает данные кредитной карты
func (sc *SecretCard) Decrypt(key []byte, content []byte) error {
	data, err := crypto.Decrypt(key, content)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, sc)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	return nil
}

// SecretText - текстовые данные
type SecretText struct {
	Text string
}

// NewSecretText - базовый конструктор
func NewSecretText(text string) *SecretText {
	return &SecretText{
		Text: text,
	}
}

// Encrypt - метод шифрует текстовые данные
func (sc *SecretText) Encrypt(key []byte) ([]byte, error) {
	return crypto.Encrypt(key, []byte(sc.Text))
}

// Decrypt - метод рашифровывает текстовые данные
func (sc *SecretText) Decrypt(key []byte, content []byte) error {
	data, err := crypto.Decrypt(key, content)
	if err != nil {
		return err
	}
	sc.Text = string(data)
	return nil
}

// SecretBinary - бинарные данные
type SecretBinary struct {
	Blob []byte
}

// NewSecretBinary- базовый конструктор
func NewSecretBinary(bin []byte) *SecretBinary {
	return &SecretBinary{
		Blob: bin,
	}
}

// Encrypt - метод шифрует бинарные данные
func (sc *SecretBinary) Encrypt(key []byte) ([]byte, error) {
	return crypto.Encrypt(key, sc.Blob)
}

// Decrypt - метод рашифровывает бинарные данные
func (sc *SecretBinary) Decrypt(key []byte, content []byte) error {
	data, err := crypto.Decrypt(key, content)
	if err != nil {
		return err
	}
	sc.Blob = data
	return nil
}
