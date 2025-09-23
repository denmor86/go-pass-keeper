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
	isViewMode    bool                    // Флаг режима просмотра
	secretData    messages.SecretPassword // Данные для просмотра
}

const (
	fieldNameIndex = iota
	fieldLoginIndex
	fieldPasswordIndex
)

func NewLoginSecretModel() LoginSecretModel {
	model := LoginSecretModel{
		focused:    fieldNameIndex,
		isViewMode: false,
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

func (m LoginSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginSecretModel) Update(msg tea.Msg) (LoginSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case messages.GetSecretPasswordMsg:
		// Переключаемся в режим просмотра при получении данных
		m.isViewMode = true
		m.secretData = msg.Data

		// Заполняем поля данными для просмотра
		m.nameInput.SetValue(msg.Data.Name)
		m.loginInput.SetValue(msg.Data.Login)
		m.passwordInput.SetValue(msg.Data.Password)
		return m, nil

	case tea.KeyMsg:
		// В режиме просмотра обрабатываем только ESC
		if m.isViewMode {
			switch msg.String() {
			case "esc":
				m.isViewMode = false
				return m, func() tea.Msg {
					return messages.SecretAddCancelMsg{}
				}
			}
			return m, nil
		}

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
			return m, m.attemptAddSecret(m.nameInput.Value(), m.loginInput.Value(), m.passwordInput.Value())

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	// В режиме просмотра игнорируем ввод данных
	if m.isViewMode {
		return m, nil
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

func (m LoginSecretModel) View() string {
	fields := []string{
		m.renderInputField("📝 Название:", m.nameInput, fieldNameIndex),
		m.renderInputField("👤 Логин:", m.loginInput, fieldLoginIndex),
		m.renderInputField("🔒 Пароль:", m.passwordInput, fieldPasswordIndex),
	}

	title := "🔐 Логин и пароль"
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.ButtonStyle.Render("Enter - Сохранить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)

	// В режиме просмотра меняем заголовок и кнопки
	if m.isViewMode {
		title = "👁️ Просмотр логина и пароля"
		buttons = styles.ButtonStyle.Render("ESC - Назад")
	}

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

func (m LoginSecretModel) renderInputField(label string, input textinput.Model, index int) string {
	var inputStyle lipgloss.Style
	if index == m.focused && !m.isViewMode {
		inputStyle = styles.FocusedInputFieldStyle
	} else {
		inputStyle = styles.InputFieldStyle
	}

	var fieldView string
	if m.isViewMode {
		value := input.Value()
		fieldView = value
		if fieldView == "" {
			fieldView = " "
		}
	} else if index == m.focused {
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
