package models

import (
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

	model.loginInput = textinput.New()
	model.loginInput.Placeholder = "Логин или email"
	model.loginInput.CharLimit = 50
	model.loginInput.TextStyle = styles.BlurredStyle

	model.passwordInput = textinput.New()
	model.passwordInput.Placeholder = "Пароль"
	model.passwordInput.CharLimit = 50
	model.passwordInput.EchoMode = textinput.EchoPassword
	model.passwordInput.EchoCharacter = '•'
	model.passwordInput.TextStyle = styles.BlurredStyle

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

			switch m.focused {
			case 0:
				cmds = append(cmds, m.nameInput.Focus())
			case 1:
				cmds = append(cmds, m.loginInput.Focus())
			case 2:
				cmds = append(cmds, m.passwordInput.Focus())
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.nameInput.Value() != "" && m.loginInput.Value() != "" && m.passwordInput.Value() != "" {
				return m, func() tea.Msg {
					return messages.SecretAddCompleteMsg{
						Name:     m.nameInput.Value(),
						Type:     "Логин/Пароль",
						Login:    m.loginInput.Value(),
						Password: m.passwordInput.Value(),
					}
				}
			}
			return m, nil

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	cmds = append(cmds, cmd)
	m.loginInput, cmd = m.loginInput.Update(msg)
	cmds = append(cmds, cmd)
	m.passwordInput, cmd = m.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

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

	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Render(label),
		inputStyle.Render(input.View()),
	) + "\n"
}
