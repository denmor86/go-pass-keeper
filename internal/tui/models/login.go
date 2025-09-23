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
	connection *settings.Settings
}

// NewLoginModel - метод для создания окна авторизации пользователя
func NewLoginModel(connection *settings.Settings) LoginModel {
	login := LoginModel{
		inputs:     make([]textinput.Model, 2),
		connection: connection,
	}

	for i := range login.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Введите имя пользователя"
			t.PlaceholderStyle = styles.BlurredStyle
			t.Focus()
			t.PromptStyle = styles.FocusedStyle
			t.TextStyle = styles.FocusedStyle
		case 1:
			t.Placeholder = "Введите пароль"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PlaceholderStyle = styles.BlurredStyle
			t.PromptStyle = styles.BlurredStyle
			t.TextStyle = styles.BlurredStyle
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
		case "tab", "shift+tab", "up", "down":
			// Сначала обрабатываем навигацию
			s := msg.String()

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

			// Обновляем фокус только для активного поля
			for i := range m.inputs {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].Focus())
					m.inputs[i].PromptStyle = styles.FocusedStyle
					m.inputs[i].TextStyle = styles.FocusedStyle
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = styles.BlurredStyle
					m.inputs[i].TextStyle = styles.BlurredStyle
				}
			}

			return m, tea.Batch(cmds...)

		case "enter":
			username := m.inputs[0].Value()
			password := m.inputs[1].Value()
			return m, m.attemptLogin(username, password)

		case "esc":
			return m, func() tea.Msg {
				return messages.GotoMainPageMsg{}
			}
		}
	}

	// Ключевое изменение: обрабатываем ввод ТОЛЬКО для активного поля
	// и НЕ обрабатываем для остальных полей
	if m.focused >= 0 && m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
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

		// Для неактивных полей показываем только значение, без курсора и т.д.
		var fieldView string
		if i == m.focused {
			fieldView = m.inputs[i].View()
		} else {
			// Для неактивного поля создаем "статическое" представление
			value := m.inputs[i].Value()
			if i == 1 && value != "" {
				// Для пароля показываем звездочки
				stars := make([]rune, len(value))
				for j := range stars {
					stars[j] = '•'
				}
				fieldView = string(stars)
			} else {
				fieldView = value
				if fieldView == "" {
					fieldView = " "
				}
			}
		}

		fields[i] = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabelStyle.Render(fieldName),
			inputStyle.Render(fieldView),
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
		token, salt, err := client.Login(username, password)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка авторизации пользователя %s: %s", username, err.Error()))
		}
		return messages.AuthSuccessMsg{Token: token, Username: username, Salt: salt}
	}
}
