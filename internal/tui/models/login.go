package models

import (
	"context"
	"fmt"
	"go-pass-keeper/internal/grpcclient"
	"go-pass-keeper/internal/grpcclient/settings"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoginModel - модель окна авторизации пользователя
type LoginModel struct {
	inputs     []textinput.Model
	focused    int
	err        messages.ErrorMsg
	windowSize tea.WindowSizeMsg
	connection *settings.Connection
}

// NewLoginModel - метод для создания окна авторизации пользователя
func NewLoginModel(connection *settings.Connection) LoginModel {
	login := LoginModel{
		inputs:     make([]textinput.Model, 2),
		connection: connection,
	}

	for i := range login.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedStyle
		t.CharLimit = 32
		t.TextStyle = styles.BlurredStyle

		switch i {
		case 0:
			t.Placeholder = "Введите имя пользователя"
			t.PlaceholderStyle = styles.BlurredStyle
			t.Focus()
			t.PromptStyle = styles.FocusedStyle
		case 1:
			t.Placeholder = "Введите пароль"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PlaceholderStyle = styles.BlurredStyle
		}

		login.inputs[i] = t
	}

	return login
}

// Init - метод инициализации окна
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод для обновления окна по внешним сообщениям
func (m LoginModel) Update(msg tea.Msg) (LoginModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" {
				username := m.inputs[0].Value()
				password := m.inputs[1].Value()
				return m, m.attemptLogin(username, password)
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs)-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			for i := range m.inputs {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].Focus())
					m.inputs[i].PromptStyle = styles.FocusedStyle
					m.inputs[i].TextStyle = styles.FocusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = styles.BlurredStyle
				m.inputs[i].TextStyle = styles.BlurredStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View - метод для отрисовки окна, в зависимости от текущего состояния
func (m LoginModel) View() string {
	// Поля ввода
	fields := make([]string, len(m.inputs))
	for i := range m.inputs {
		var inputStyle lipgloss.Style
		if i == m.focused {
			inputStyle = styles.FocusedInputFieldStyle
		} else {
			inputStyle = styles.InputFieldStyle
		}

		fieldName := ""
		switch i {
		case 0:
			fieldName = "👤 Имя пользователя"
		case 1:
			fieldName = "🔒 Пароль"
		}

		fields[i] = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabelStyle.Render(fieldName),
			inputStyle.Render(m.inputs[i].View()),
		)
	}

	// Основной контент
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("🔐 Вход в систему"),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(1).Render(""),

		// Кнопки действий
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Войти"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Назад"),
		),
	)

	// Сообщение об ошибке
	if m.err != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			lipgloss.NewStyle().Height(1).Render(""),
			styles.ErrorStyle.Render("❌ "+string(m.err)),
		)
	}

	// Подсказка
	content = lipgloss.JoinVertical(
		lipgloss.Center,
		content,
		lipgloss.NewStyle().Height(1).Render(""),
		styles.HelpStyle.Render("Tab: переключение полей • Enter: подтвердить • ESC: назад"),
	)

	return styles.ContainerStyle.
		Width(m.windowSize.Width).
		Height(m.windowSize.Height).
		Render(
			lipgloss.Place(
				m.windowSize.Width, m.windowSize.Height,
				lipgloss.Center, lipgloss.Center,
				content,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(styles.BackgroundColor),
			),
		)
}

// attemptLogin - метод обработки прохождения авторизации пользователя
func (m LoginModel) attemptLogin(username string, password string) tea.Cmd {
	return func() tea.Msg {
		if username == "" || password == "" {
			return messages.ErrorMsg("заполните все поля")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.connection.Timeout)*time.Second)
		client := grpcclient.NewUserClient(m.connection.ServerAddress())
		defer func() {
			cancel()
			client.Close()
		}()
		if err := client.Connect(ctx); err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка подключения к %s: %s", m.connection.ServerAddress(), err.Error()))
		}
		token, err := client.Login(username, password)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка авторизации пользователя %s: %s", username, err.Error()))
		}
		return messages.AuthSuccessMsg{Token: token, Email: username}
	}
}
