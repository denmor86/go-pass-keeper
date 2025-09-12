package models

import (
	"context"
	"go-pass-keeper/internal/grpcclient"
	"go-pass-keeper/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func attemptLogin(serverAddr string, username string, password string) tea.Cmd {
	return func() tea.Msg {
		if username == "" || password == "" {
			return messages.ErrorMsg("заполните все поля")
		}
		if len(password) < 6 {
			return messages.ErrorMsg("пароль должен содержать минимум 6 символов")
		}
		token, err := grpcclient.LoginUser(context.Background(), serverAddr, username, password)
		if err != nil {
			return messages.AuthFailMsg(username)
		}
		return messages.AuthSuccessMsg{Token: token, Email: username}
	}
}

func attemptRegister(serverAddr string, username string, password string, confirm string) tea.Cmd {
	return func() tea.Msg {
		if username == "" || password == "" || confirm == "" {
			return messages.ErrorMsg("заполните все поля")
		}
		if len(username) < 3 {
			return messages.ErrorMsg("имя пользователя должно содержать минимум 3 символа")
		}
		if len(password) < 6 {
			return messages.ErrorMsg("пароль должен содержать минимум 6 символов")
		}
		if password != confirm {
			return messages.ErrorMsg("пароли не совпадают")
		}
		token, err := grpcclient.RegisterUser(context.Background(), serverAddr, username, password)
		if err != nil {
			return messages.AuthFailMsg(username)
		}
		return messages.AuthSuccessMsg{Token: token, Email: username}
	}
}
