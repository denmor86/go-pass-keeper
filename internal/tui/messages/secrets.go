package messages

import (
	"fmt"
	"go-pass-keeper/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

type EncryptConverter interface {
	ToModel([]byte) (*models.SecretInfo, []byte, error)
}
type DecryptConverter interface {
	FromModel([]byte, *models.SecretInfo, []byte) error
}

// SecretPassword - модель с данными логин/пароль
type SecretPassword struct {
	Name     string
	Type     string
	Login    string
	Password string
}

// AddSecretPasswordMsg - сообщение для добавления данными логин/пароль
type AddSecretPasswordMsg struct {
	Data SecretPassword
}

// GetSecretPasswordMsg - сообщение для получения данных логин/пароль
type GetSecretPasswordMsg struct {
	Data SecretPassword
}

// ToModel - метод формирует информацию о секрете и расшифрованный контент
func (msg *AddSecretPasswordMsg) ToModel(key []byte) (*models.SecretInfo, []byte, error) {

	secret := models.NewSecretPassword(msg.Data.Login, msg.Data.Password)
	data, err := secret.Encrypt(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return &models.SecretInfo{Name: msg.Data.Name, Type: msg.Data.Type}, data, nil
}

// FromModel - метод формирует информацию о секрете, шифрованный контент и формирует сообщение
func (msg *GetSecretPasswordMsg) FromModel(key []byte, info *models.SecretInfo, content []byte) error {
	secret := &models.SecretPassword{}
	err := secret.Decrypt(key, content)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	msg.Data = SecretPassword{Name: info.Name, Type: info.Type, Login: secret.Login, Password: secret.Password}
	return nil
}

// SecretCard - модель с данными карты
type SecretCard struct {
	Name   string
	Type   string
	Number string
	Date   string
	CVV    string
	Owner  string
}

// AddSecretCardMsg - сообщение для добавления с данными карты
type AddSecretCardMsg struct {
	Data SecretCard
}

// GetSecretCardMsg - сообщение для получения данных карты
type GetSecretCardMsg struct {
	Data SecretCard
}

// ToModel - метод формирует информацию о секрете и шифрованный контент
func (msg *AddSecretCardMsg) ToModel(key []byte) (*models.SecretInfo, []byte, error) {

	secret := models.NewSecretCard(msg.Data.Number, msg.Data.Date, msg.Data.CVV, msg.Data.Owner)
	data, err := secret.Encrypt(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return &models.SecretInfo{Name: msg.Data.Name, Type: msg.Data.Type}, data, nil
}

// FromModel - метод формирует информацию о секрете, шифрованный контент и формирует сообщение
func (msg *GetSecretCardMsg) FromModel(key []byte, info *models.SecretInfo, content []byte) error {
	secret := &models.SecretCard{}
	err := secret.Decrypt(key, content)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	msg.Data = SecretCard{Name: info.Name, Type: info.Type, Number: secret.Number, CVV: secret.CVV, Date: secret.Date, Owner: secret.Owner}
	return nil
}

// SecretText - модель с текстовыми данными
type SecretText struct {
	Name string
	Type string
	Text string
}

// AddSecretTextMsg - сообщение для добавления секрета с текстовыми данными
type AddSecretTextMsg struct {
	Data SecretText
}

// GetSecretTextMsg - сообщение для получения секрета с  текстовыми данными
type GetSecretTextMsg struct {
	Data SecretText
}

// ToModel - метод формирует информацию о секрете и шифрованный контент
func (msg *AddSecretTextMsg) ToModel(key []byte) (*models.SecretInfo, []byte, error) {

	secret := models.NewSecretText(msg.Data.Text)
	data, err := secret.Encrypt(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return &models.SecretInfo{Name: msg.Data.Name, Type: msg.Data.Type}, data, nil
}

// FromModel - метод формирует информацию о секрете, шифрованный контент и формирует сообщение
func (msg *GetSecretTextMsg) FromModel(key []byte, info *models.SecretInfo, content []byte) error {
	secret := &models.SecretText{}
	err := secret.Decrypt(key, content)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	msg.Data = SecretText{Name: info.Name, Type: info.Type, Text: secret.Text}
	return nil
}

// SecretBinary - модель с бинарными данными
type SecretBinary struct {
	Name string
	Type string
	Blob []byte
}

// AddSecretBinaryMsg - сообщение для добавления секрета с бинарными данными
type AddSecretBinaryMsg struct {
	Data SecretBinary
}

// GetSecretBinarytMsg - сообщение для получения секрета с бинарными данными
type GetSecretBinarytMsg struct {
	Data SecretBinary
}

// ToModel - метод формирует информацию о секрете и шифрованный контент
func (msg *AddSecretBinaryMsg) ToModel(key []byte) (*models.SecretInfo, []byte, error) {

	secret := models.NewSecretBinary(msg.Data.Blob)
	data, err := secret.Encrypt(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return &models.SecretInfo{Name: msg.Data.Name, Type: msg.Data.Type}, data, nil
}

// FromModel - метод формирует информацию о секрете, шифрованный контент и формирует сообщение
func (msg *GetSecretBinarytMsg) FromModel(key []byte, info *models.SecretInfo, content []byte) error {
	secret := &models.SecretBinary{}
	err := secret.Decrypt(key, content)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	msg.Data = SecretBinary{Name: info.Name, Type: info.Type, Blob: secret.Blob}
	return nil
}

// ToMessage - метод формирует сообщение на основе информации о секрете
func ToMessage(key []byte, info *models.SecretInfo, content []byte) tea.Msg {
	switch info.Type {
	case models.SecretPasswordType:
		msg := &GetSecretPasswordMsg{}
		err := msg.FromModel(key, info, content)
		if err != nil {
			return ErrorMsg(fmt.Sprintf("Ошибка разбора сообщения: %s", err.Error()))
		}
		return msg
	case models.SecretCardType:
		msg := &GetSecretCardMsg{}
		err := msg.FromModel(key, info, content)
		if err != nil {
			return ErrorMsg(fmt.Sprintf("Ошибка разбора сообщения: %s", err.Error()))
		}
		return msg
	case models.SecretTextType:
		msg := &GetSecretTextMsg{}
		err := msg.FromModel(key, info, content)
		if err != nil {
			return ErrorMsg(fmt.Sprintf("Ошибка разбора сообщения: %s", err.Error()))
		}
		return msg
	case models.SecretBinaryType:
		msg := &GetSecretBinarytMsg{}
		err := msg.FromModel(key, info, content)
		if err != nil {
			return ErrorMsg(fmt.Sprintf("Ошибка разбора сообщения: %s", err.Error()))
		}
		return msg
	default:
		return ErrorMsg("Неизвестный тип сообщения")
	}
}
