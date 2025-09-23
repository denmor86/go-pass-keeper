package models

import (
	"go-pass-keeper/internal/models"
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
	isViewMode bool
	secretData messages.SecretCard
}

func NewBankCardSecretModel() BankCardSecretModel {
	model := BankCardSecretModel{
		focused:    0,
		isViewMode: false,
	}

	model.cardInputs = make([]textinput.Model, 5)
	for i := range model.cardInputs {
		t := textinput.New()
		t.TextStyle = styles.BlurredStyle
		t.CharLimit = 50
		t.PromptStyle = styles.BlurredStyle

		switch i {
		case 0:
			t.Placeholder = "Имя карты"
			t.CharLimit = 100
		case 1:
			t.Placeholder = "Номер карты"
			t.CharLimit = 19
		case 2:
			t.Placeholder = "Срок действия (MM/YY)"
			t.CharLimit = 5
		case 3:
			t.Placeholder = "CVV код"
			t.CharLimit = 3
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 4:
			t.Placeholder = "Имя владельца"
		}

		model.cardInputs[i] = t
	}

	model.cardInputs[0].Focus()
	model.cardInputs[0].PromptStyle = styles.FocusedStyle
	model.cardInputs[0].TextStyle = styles.FocusedStyle

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

	case messages.GetSecretCardMsg:
		// Обработка сообщения для отображения данных в режиме просмотра
		m.isViewMode = true
		m.secretData = msg.Data

		// Заполняем поля данными (только для отображения)
		m.cardInputs[0].SetValue(msg.Data.Name)
		m.cardInputs[1].SetValue(msg.Data.Number)
		m.cardInputs[2].SetValue(msg.Data.Date)
		m.cardInputs[3].SetValue(msg.Data.CVV)
		m.cardInputs[4].SetValue(msg.Data.Owner)

		// Делаем все поля неактивными в режиме просмотра
		for i := range m.cardInputs {
			m.cardInputs[i].Blur()
		}
		return m, nil

	case tea.KeyMsg:
		// Режим редактирования
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Сбрасываем фокус со всех полей
			for i := range m.cardInputs {
				m.cardInputs[i].Blur()
				m.cardInputs[i].PromptStyle = styles.BlurredStyle
				m.cardInputs[i].TextStyle = styles.BlurredStyle
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > 4 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = 4
			}

			// Устанавливаем фокус только на активное поле
			cmds = append(cmds, m.cardInputs[m.focused].Focus())
			m.cardInputs[m.focused].PromptStyle = styles.FocusedStyle
			m.cardInputs[m.focused].TextStyle = styles.FocusedStyle

			return m, tea.Batch(cmds...)

		case "enter":
			return m, m.attemptAddSecret(
				m.cardInputs[0].Value(),
				m.cardInputs[1].Value(),
				m.cardInputs[2].Value(),
				m.cardInputs[3].Value(),
				m.cardInputs[4].Value())

		case "esc":
			m.isViewMode = false
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	// Обрабатываем ввод ТОЛЬКО для активного поля и только в режиме редактирования
	if !m.isViewMode && m.focused >= 0 && m.focused < len(m.cardInputs) {
		var cmd tea.Cmd
		m.cardInputs[m.focused], cmd = m.cardInputs[m.focused].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m BankCardSecretModel) View() string {
	// Адаптивная ширина контента
	contentWidth := m.windowSize.Width - 10
	if contentWidth > 60 {
		contentWidth = 60
	}
	if contentWidth < 40 {
		contentWidth = 40
	}

	fields := []string{
		m.renderInputField("📝 Имя карты:", m.cardInputs[0], 0, contentWidth),
		m.renderInputField("💳 Номер карты:", m.cardInputs[1], 1, contentWidth),
		m.renderInputField("📅 Срок действия:", m.cardInputs[2], 2, contentWidth),
		m.renderInputField("🔒 CVV код:", m.cardInputs[3], 3, contentWidth),
		m.renderInputField("👤 Владелец:", m.cardInputs[4], 4, contentWidth),
	}

	// Заголовок в зависимости от режима
	title := "💳 Банковская карта"
	if m.isViewMode {
		title = "👁️ Просмотр карты"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(contentWidth).
			Render(title),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(1).Render(""),

		m.renderButtons(contentWidth),
	)

	// Адаптивное размещение с учетом размера окна
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

func (m BankCardSecretModel) renderInputField(label string, input textinput.Model, index int, width int) string {
	var inputStyle lipgloss.Style
	if index == m.focused && !m.isViewMode {
		inputStyle = styles.FocusedInputFieldStyle.Width(width - 10)
	} else {
		inputStyle = styles.InputFieldStyle.Width(width - 10)
	}

	// Для режима просмотра или неактивных полей показываем статическое представление
	var fieldView string
	if index == m.focused && !m.isViewMode {
		fieldView = input.View()
	} else {
		value := input.Value()
		if index == 3 && value != "" { // Для CVV поля
			stars := make([]rune, len(value))
			for i := range stars {
				stars[i] = '•'
			}
			fieldView = string(stars)
		} else {
			fieldView = value
			if fieldView == "" {
				fieldView = "не задано"
			}
		}
	}

	// Ограничиваем длину текста для лучшего отображения
	if len(fieldView) > width-15 {
		fieldView = fieldView[:width-15] + "..."
	}

	fieldContent := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Width(width-10).Render(label),
		inputStyle.Render(fieldView),
	)

	return lipgloss.NewStyle().
		Width(width).
		Render(fieldContent) + "\n"
}

func (m BankCardSecretModel) renderButtons(width int) string {
	if m.isViewMode {
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("ESC - Назад"),
		)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.ButtonStyle.Render("Enter - Сохранить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(buttons)
}

// attemptAddSecret - метод обработки добавления секрета
func (m BankCardSecretModel) attemptAddSecret(name string, number string, date string, cvv string, owner string) tea.Cmd {
	return func() tea.Msg {
		if len(name) == 0 {
			return messages.ErrorMsg("Необходимо задать имя секрета")
		}
		if len(number) == 0 {
			return messages.ErrorMsg("Необходимо задать номер карты")
		}
		if len(date) == 0 {
			return messages.ErrorMsg("Необходимо задать дату выдачи карты")
		}
		if len(cvv) == 0 {
			return messages.ErrorMsg("Необходимо задать CVV карты")
		}
		if len(owner) == 0 {
			return messages.ErrorMsg("Необходимо задать владельца карты")
		}
		return messages.AddSecretCardMsg{Data: messages.SecretCard{Name: name, Type: models.SecretCardType, Number: number, CVV: cvv, Date: date, Owner: owner}}
	}
}
