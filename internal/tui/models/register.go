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

// RegisterModel - модель окна регистрации пользователя
type RegisterModel struct {
	inputs     []textinput.Model
	focused    int
	err        messages.ErrorMsg
	windowSize tea.WindowSizeMsg
	connection *settings.Connection
}

// NewLoginModel - метод для создания окна регистрации пользователя
func NewRegisterModel(connection *settings.Connection) RegisterModel {
	register := RegisterModel{
		inputs:     make([]textinput.Model, 3),
		connection: connection,
	}

	for i := range register.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedStyle
		t.CharLimit = 32
		t.TextStyle = styles.BlurredStyle

		switch i {
		case 0:
			t.Placeholder = "Введите имя пользователя"
			t.PlaceholderStyle = styles.BlurredStyle
		case 1:
			t.Placeholder = "Введите пароль"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PlaceholderStyle = styles.BlurredStyle
		case 2:
			t.Placeholder = "Подтвердите пароль"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PlaceholderStyle = styles.BlurredStyle
		}

		register.inputs[i] = t
	}

	return register
}

// Init - метод инициализации окна
func (m RegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод для обновления окна по внешним сообщениям
func (m RegisterModel) Update(msg tea.Msg) (RegisterModel, tea.Cmd) {
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
				confirm := m.inputs[2].Value()
				return m, m.attemptRegister(username, password, confirm)
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
func (m RegisterModel) View() string {

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
		case 2:
			fieldName = "✅ Подтверждение пароля"
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
			Render("📝 Регистрация"),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(1).Render(""),

		// Кнопки действий
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Зарегистрироваться"),
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

// attemptRegister - метод обработки прохождения регистрации пользователя
func (m RegisterModel) attemptRegister(username string, password string, confirm string) tea.Cmd {
	return func() tea.Msg {
		if username == "" || password == "" || confirm == "" {
			return messages.ErrorMsg("заполните все поля")
		}
		if password != confirm {
			return messages.ErrorMsg("пароли не совпадают")
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
		token, salt, err := client.Register(username, password)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка регистрации пользователя %s: %s", username, err.Error()))
		}
		return messages.AuthSuccessMsg{Token: token, Username: username, Salt: salt}
	}
}
