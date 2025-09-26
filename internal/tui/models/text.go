package models

import (
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextSecretModel - модель окна создания/просмотра текстового секрета
type TextSecretModel struct {
	nameInput  textinput.Model
	textArea   textarea.Model
	focused    bool
	windowSize tea.WindowSizeMsg
	isEditMode bool   // Флаг режима редактирования
	sid        string // id для редактирования
}

// NewTextSecretModel - метод создания модель окна создания/просмотра текстового секрета
func NewTextSecretModel() TextSecretModel {
	model := TextSecretModel{
		focused:    false,
		isEditMode: false,
	}

	model.nameInput = textinput.New()
	model.nameInput.Placeholder = "Название"
	model.nameInput.CharLimit = 50
	model.nameInput.TextStyle = styles.BlurredStyle

	model.textArea = textarea.New()
	model.textArea.Placeholder = "Введите текст здесь..."
	model.textArea.SetWidth(50)
	model.textArea.SetHeight(8)

	model.nameInput.Focus()

	return model
}

// Init - метод инициализации текущего окна
func (m TextSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод обновления текущего окна
func (m TextSecretModel) Update(msg tea.Msg) (TextSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		m.textArea.SetWidth(msg.Width - 20)
		return m, nil

	case messages.GetSecretTextMsg:
		// Переключаемся в режим редактирования при получении данных
		m.isEditMode = true
		m.sid = msg.ID

		// Заполняем поля данными для просмотра
		m.nameInput.SetValue(msg.Data.Name)
		m.textArea.SetValue(msg.Data.Text)
		return m, nil

	case tea.KeyMsg:
		// Режим редактирования
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
			if m.isEditMode {
				m.isEditMode = false
				return m, m.attemptEditSecret(m.sid, m.nameInput.Value(), m.textArea.Value())
			}
			return m, m.attemptAddSecret(m.nameInput.Value(), m.textArea.Value())

		case "esc":
			m.isEditMode = false
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

// View - метод отрисовки текущего состояния
func (m TextSecretModel) View() string {
	title := "📝 Текст"
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.ButtonStyle.Render("Enter - Применить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)
	hint := "Tab: переключение между полями"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render(title),

		lipgloss.NewStyle().Height(1).Render(""),

		m.renderInputField("📝 Название:", m.nameInput),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Render("Текст:"),

		m.renderTextArea(m.textArea),

		lipgloss.NewStyle().Height(2).Render(""),

		buttons,

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render(hint),
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

// renderInputField - метод для отрисовки полей ввода
func (m TextSecretModel) renderInputField(label string, input textinput.Model) string {
	var inputStyle lipgloss.Style
	if m.focused {
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

// renderTextArea - метод отрисовки окна для ввода текста
func (m TextSecretModel) renderTextArea(area textarea.Model) string {
	return styles.InputFieldStyle.Width(60).Height(12).Render(area.View())
}

// attemptAddSecret - метод обработки добавления секрета
func (m TextSecretModel) attemptAddSecret(name string, text string) tea.Cmd {
	return func() tea.Msg {
		if len(name) == 0 {
			return messages.ErrorMsg("Необходимо задать имя секрета")
		}
		if len(text) == 0 {
			return messages.ErrorMsg("Пустой текст секрета")
		}
		return messages.AddSecretTextMsg{
			Data: messages.SecretText{
				Name: name,
				Type: models.SecretTextType,
				Text: text,
			},
		}
	}
}

// attemptAddSecret - метод обработки добавления секрета
func (m TextSecretModel) attemptEditSecret(sid string, name string, text string) tea.Cmd {
	return func() tea.Msg {
		if len(name) == 0 {
			return messages.ErrorMsg("Необходимо задать имя секрета")
		}
		if len(text) == 0 {
			return messages.ErrorMsg("Пустой текст секрета")
		}
		return messages.EditSecretTextMsg{
			ID: sid,
			Data: messages.SecretText{
				Name: name,
				Type: models.SecretTextType,
				Text: text,
			},
		}
	}
}
