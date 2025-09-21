package models

import (
	"fmt"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BankCardSecretModel struct {
	cardInputs []textinput.Model
	focused    int
	windowSize tea.WindowSizeMsg
}

func NewBankCardSecretModel() BankCardSecretModel {
	model := BankCardSecretModel{
		focused: 0,
	}

	model.cardInputs = make([]textinput.Model, 4)
	for i := range model.cardInputs {
		t := textinput.New()
		t.TextStyle = styles.BlurredStyle
		t.CharLimit = 50

		switch i {
		case 0:
			t.Placeholder = "Номер карты"
			t.CharLimit = 19
		case 1:
			t.Placeholder = "Срок действия (MM/YY)"
			t.CharLimit = 5
		case 2:
			t.Placeholder = "CVV код"
			t.CharLimit = 3
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 3:
			t.Placeholder = "Имя владельца"
		}

		model.cardInputs[i] = t
	}

	model.cardInputs[0].Focus()

	return model
}

func (m BankCardSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m BankCardSecretModel) Update(msg tea.Msg) (BankCardSecretModel, tea.Cmd) {
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

			if m.focused > 3 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = 3
			}

			cmds = append(cmds, m.cardInputs[m.focused].Focus())
			return m, tea.Batch(cmds...)

		case "enter":
			// Проверяем, что все поля заполнены
			allFilled := true
			for _, input := range m.cardInputs {
				if input.Value() == "" {
					allFilled = false
					break
				}
			}

			if allFilled {
				return m, func() tea.Msg {
					return messages.SecretAddCompleteMsg{
						Name: "Банковская карта",
						Type: "Банковская карта",
						Content: fmt.Sprintf("Номер: %s\nСрок: %s\nCVV: %s\nВладелец: %s",
							m.cardInputs[0].Value(),
							m.cardInputs[1].Value(),
							m.cardInputs[2].Value(),
							m.cardInputs[3].Value()),
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

	for i := range m.cardInputs {
		var cmd tea.Cmd
		m.cardInputs[i], cmd = m.cardInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m BankCardSecretModel) View() string {
	fields := []string{
		m.renderInputField("💳 Номер карты:", m.cardInputs[0], 0),
		m.renderInputField("📅 Срок действия:", m.cardInputs[1], 1),
		m.renderInputField("🔒 CVV код:", m.cardInputs[2], 2),
		m.renderInputField("👤 Владелец:", m.cardInputs[3], 3),
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("💳 Банковская карта"),

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

func (m BankCardSecretModel) renderInputField(label string, input textinput.Model, index int) string {
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
