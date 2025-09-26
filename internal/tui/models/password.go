package models

import (
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoginSecretModel - модель окна секрета (логин/пароль)
type LoginSecretModel struct {
	nameInput     textinput.Model
	loginInput    textinput.Model
	passwordInput textinput.Model
	focused       int
	windowSize    tea.WindowSizeMsg
	isEditMode    bool   // Флаг режима редактирования
	sid           string // id для редактирования
}

// Индексы полей
const (
	fieldNameIndex = iota
	fieldLoginIndex
	fieldPasswordIndex
)

// NewFileSecretModel - метод создания модель окна секрета (логин/пароль)
func NewLoginSecretModel() LoginSecretModel {
	model := LoginSecretModel{
		focused:    fieldNameIndex,
		isEditMode: false,
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
	model.passwordInput.TextStyle = styles.BlurredStyle
	model.passwordInput.PromptStyle = styles.BlurredStyle

	model.nameInput.Focus()

	return model
}

// Init - метод инициализации текущего окна
func (m LoginSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод обновления текущего окна
func (m LoginSecretModel) Update(msg tea.Msg) (LoginSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case messages.GetSecretPasswordMsg:
		// Переключаемся в режим просмотра при получении данных
		m.isEditMode = true
		m.sid = msg.ID
		// Заполняем поля данными для просмотра
		m.nameInput.SetValue(msg.Data.Name)
		m.loginInput.SetValue(msg.Data.Login)
		m.passwordInput.SetValue(msg.Data.Password)
		return m, nil

	case tea.KeyMsg:
		// Режим редактирования
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Сбрасываем стили всех полей
			m.nameInput.Blur()
			m.loginInput.Blur()
			m.passwordInput.Blur()
			m.nameInput.PromptStyle = styles.BlurredStyle
			m.nameInput.TextStyle = styles.BlurredStyle
			m.loginInput.PromptStyle = styles.BlurredStyle
			m.loginInput.TextStyle = styles.BlurredStyle
			m.passwordInput.PromptStyle = styles.BlurredStyle
			m.passwordInput.TextStyle = styles.BlurredStyle

			// Навигация по полям
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > fieldPasswordIndex {
				m.focused = fieldNameIndex
			} else if m.focused < fieldNameIndex {
				m.focused = fieldPasswordIndex
			}

			// Устанавливаем фокус на активное поле
			switch m.focused {
			case fieldNameIndex:
				cmds = append(cmds, m.nameInput.Focus())
				m.nameInput.PromptStyle = styles.FocusedStyle
				m.nameInput.TextStyle = styles.FocusedStyle
			case fieldLoginIndex:
				cmds = append(cmds, m.loginInput.Focus())
				m.loginInput.PromptStyle = styles.FocusedStyle
				m.loginInput.TextStyle = styles.FocusedStyle
			case fieldPasswordIndex:
				cmds = append(cmds, m.passwordInput.Focus())
				m.passwordInput.PromptStyle = styles.FocusedStyle
				m.passwordInput.TextStyle = styles.FocusedStyle
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.isEditMode {
				m.isEditMode = false // сбрасываем режим
				return m, m.attemptEditSecret(m.sid, m.nameInput.Value(), m.loginInput.Value(), m.passwordInput.Value())
			}
			return m, m.attemptAddSecret(m.nameInput.Value(), m.loginInput.Value(), m.passwordInput.Value())

		case "esc":
			m.isEditMode = false
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	// Обновляем активное поле ввода
	var cmd tea.Cmd
	switch m.focused {
	case fieldNameIndex:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case fieldLoginIndex:
		m.loginInput, cmd = m.loginInput.Update(msg)
	case fieldPasswordIndex:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View - метод отрисовки текущего состояния
func (m LoginSecretModel) View() string {
	fields := []string{
		m.renderInputField("📝 Название:", m.nameInput, fieldNameIndex),
		m.renderInputField("👤 Логин:", m.loginInput, fieldLoginIndex),
		m.renderInputField("🔒 Пароль:", m.passwordInput, fieldPasswordIndex),
	}

	title := "🔐 Логин и пароль"
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.ButtonStyle.Render("Enter - Применить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render(title),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(2).Render(""),

		buttons,
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
		fieldView = value
		if fieldView == "" {
			fieldView = " "
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
		return messages.AddSecretPasswordMsg{
			Data: messages.SecretPassword{
				Name:     name,
				Type:     models.SecretPasswordType,
				Login:    username,
				Password: password,
			},
		}
	}
}

// attemptEditSecret - метод обработки изменения секрета
func (m LoginSecretModel) attemptEditSecret(sid string, name string, username string, password string) tea.Cmd {
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
		return messages.EditSecretPasswordMsg{
			ID: sid,
			Data: messages.SecretPassword{
				Name:     name,
				Type:     models.SecretPasswordType,
				Login:    username,
				Password: password,
			},
		}
	}
}
