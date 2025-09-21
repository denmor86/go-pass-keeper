package models

import (
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TextSecretModel struct {
	nameInput  textinput.Model
	textArea   textarea.Model
	focused    bool // true - textarea, false - name input
	windowSize tea.WindowSizeMsg
}

func NewTextSecretModel() TextSecretModel {
	model := TextSecretModel{
		focused: false,
	}

	model.nameInput = textinput.New()
	model.nameInput.Placeholder = "Название текстового секрета"
	model.nameInput.CharLimit = 50
	model.nameInput.TextStyle = styles.BlurredStyle

	model.textArea = textarea.New()
	model.textArea.Placeholder = "Введите ваш секретный текст здесь..."
	model.textArea.SetWidth(50)
	model.textArea.SetHeight(8)

	model.nameInput.Focus()

	return model
}

func (m TextSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextSecretModel) Update(msg tea.Msg) (TextSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		m.textArea.SetWidth(msg.Width - 20)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.focused {
				m.textArea.Blur()
				m.focused = false
				return m, m.nameInput.Focus()
			} else {
				m.nameInput.Blur()
				m.focused = true
				return m, m.textArea.Focus()
			}

		case "enter":
			if m.nameInput.Value() != "" && m.textArea.Value() != "" {
				return m, func() tea.Msg {
					return messages.SecretAddCompleteMsg{
						Name:    m.nameInput.Value(),
						Type:    "Текст",
						Content: m.textArea.Value(),
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
	if m.focused {
		m.textArea, cmd = m.textArea.Update(msg)
	} else {
		m.nameInput, cmd = m.nameInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m TextSecretModel) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("📝 Текстовый секрет"),

		lipgloss.NewStyle().Height(1).Render(""),

		m.renderInputField("📝 Название:", m.nameInput),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Render("Текст:"),

		styles.InputFieldStyle.
			Height(8).
			Render(m.textArea.View()),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Сохранить"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Отмена"),
		),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render("Tab: переключение между полями"),
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

func (m TextSecretModel) renderInputField(label string, input textinput.Model) string {
	var inputStyle lipgloss.Style
	if !m.focused {
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
