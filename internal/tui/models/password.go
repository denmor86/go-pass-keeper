package models

import (
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginSecretModel struct {
	nameInput     textinput.Model
	loginInput    textinput.Model
	passwordInput textinput.Model
	focused       int
	windowSize    tea.WindowSizeMsg
}

func NewLoginSecretModel() LoginSecretModel {
	model := LoginSecretModel{
		focused: 0,
	}

	model.nameInput = textinput.New()
	model.nameInput.Placeholder = "Название аккаунта"
	model.nameInput.CharLimit = 50
	model.nameInput.TextStyle = styles.BlurredStyle
	model.nameInput.PromptStyle = styles.FocusedStyle

	model.loginInput = textinput.New()
	model.loginInput.Placeholder = "Логин или email"
	model.loginInput.CharLimit = 50
	model.loginInput.TextStyle = styles.BlurredStyle
	model.loginInput.PromptStyle = styles.BlurredStyle

	model.passwordInput = textinput.New()
	model.passwordInput.Placeholder = "Пароль"
	model.passwordInput.CharLimit = 50
	model.passwordInput.EchoMode = textinput.EchoPassword
	model.passwordInput.EchoCharacter = '•'
	model.passwordInput.TextStyle = styles.BlurredStyle
	model.passwordInput.PromptStyle = styles.BlurredStyle

	model.nameInput.Focus()

	return model
}

func (m LoginSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginSecretModel) Update(msg tea.Msg) (LoginSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			m.nameInput.Blur()
			m.loginInput.Blur()
			m.passwordInput.Blur()
			m.nameInput.PromptStyle = styles.BlurredStyle
			m.nameInput.TextStyle = styles.BlurredStyle
			m.loginInput.PromptStyle = styles.BlurredStyle
			m.loginInput.TextStyle = styles.BlurredStyle
			m.passwordInput.PromptStyle = styles.BlurredStyle
			m.passwordInput.TextStyle = styles.BlurredStyle

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > 2 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = 2
			}

			// Устанавливаем фокус только на активное поле
			switch m.focused {
			case 0:
				cmds = append(cmds, m.nameInput.Focus())
				m.nameInput.PromptStyle = styles.FocusedStyle
				m.nameInput.TextStyle = styles.FocusedStyle
			case 1:
				cmds = append(cmds, m.loginInput.Focus())
				m.loginInput.PromptStyle = styles.FocusedStyle
				m.loginInput.TextStyle = styles.FocusedStyle
			case 2:
				cmds = append(cmds, m.passwordInput.Focus())
				m.passwordInput.PromptStyle = styles.FocusedStyle
				m.passwordInput.TextStyle = styles.FocusedStyle
			}
			return m, tea.Batch(cmds...)

		case "enter":
			return m, m.attemptAddSecret(m.nameInput.Value(), m.loginInput.Value(), m.passwordInput.Value())

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}
	var cmd tea.Cmd
	switch m.focused {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.loginInput, cmd = m.loginInput.Update(msg)
	case 2:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m LoginSecretModel) View() string {
	fields := []string{
		m.renderInputField("📝 Название:", m.nameInput, 0),
		m.renderInputField("👤 Логин:", m.loginInput, 1),
		m.renderInputField("🔒 Пароль:", m.passwordInput, 2),
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("🔐 Логин и пароль"),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Сохранить"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Отмена"),
		),
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

func (m LoginSecretModel) renderInputField(label string, input textinput.Model, index int) string {
	var inputStyle lipgloss.Style
	if index == m.focused {
		inputStyle = styles.FocusedInputFieldStyle
	} else {
		inputStyle = styles.InputFieldStyle
	}

	var fieldView string
	if index == m.focused {
		fieldView = input.View()
	} else {
		value := input.Value()
		if index == 2 && value != "" {
			stars := make([]rune, len(value))
			for i := range stars {
				stars[i] = '•'
			}
			fieldView = string(stars)
		} else {
			fieldView = value
			if fieldView == "" {
				fieldView = " "
			}
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Render(label),
		inputStyle.Render(fieldView),
	) + "\n"
}

// attemptAddSecret - метод обработки добавления секрета
func (m LoginSecretModel) attemptAddSecret(name string, username string, password string) tea.Cmd {
	return func() tea.Msg {
		if len(name) == 0 {
			return messages.ErrorMsg("Необходимо задать имя секрета")
		}
		if len(username) == 0 {
			return messages.ErrorMsg("Необходимо задать пользователя")
		}
		if len(password) == 0 {
			return messages.ErrorMsg("Необходимо задать пароль")
		}
		return messages.AddSecretPasswordMsg{Data: messages.SecretPassword{Name: name, Type: models.SecretPasswordType, Login: username, Password: password}}
	}
}
