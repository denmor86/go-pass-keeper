package messages

// AuthSuccess - сообщение об успешной аутентификации
type AuthSuccessMsg struct {
	Username string
	Token    string
	Salt     string
}
