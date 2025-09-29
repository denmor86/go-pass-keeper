package models

import (
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Состояния модели
const (
	SecretTypeSelectState = iota
	LoginPasswordState
	TextState
	FileState
	BankCardState
)

// SecretAddModel - модель выбора создаваемого секрета
type SecretAddModel struct {
	state       int
	windowSize  tea.WindowSizeMsg
	focusedBtn  int
	secretTypes []string

	loginModel LoginSecretModel
	textModel  TextSecretModel
	fileModel  FileSecretModel
	cardModel  BankCardSecretModel
}

// NewViewerModel - метод создания выбора создаваемого секрета
func NewSecretAddModel() SecretAddModel {
	return SecretAddModel{
		state:       SecretTypeSelectState,
		secretTypes: []string{"🔐 Логин/Пароль", "📝 Текст", "📁 Файл", "💳 Банковская карта"},
		focusedBtn:  0,
		loginModel:  NewLoginSecretModel(),
		textModel:   NewTextSecretModel(),
		fileModel:   NewFileSecretModel(),
		cardModel:   NewBankCardSecretModel(),
	}
}

// Init - метод инициализации текущего окна
func (m SecretAddModel) Init() tea.Cmd {
	return tea.Batch(
		m.loginModel.Init(),
		m.textModel.Init(),
		m.fileModel.Init(),
		m.cardModel.Init(),
	)
}

// Update - метод обновления текущего окна
func (m SecretAddModel) Update(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateWindowsSize(msg)

	case messages.SecretAddCancelMsg:
		m.state = SecretTypeSelectState
		return m, nil
	case messages.GetSecretPasswordMsg:
		m.state = LoginPasswordState
	case messages.GetSecretCardMsg:
		m.state = BankCardState
	case messages.GetSecretTextMsg:
		m.state = TextState
	case messages.GetSecretBinaryMsg:
		m.state = FileState
	}

	switch m.state {
	case SecretTypeSelectState:
		return m.updateTypeSelect(msg)
	case LoginPasswordState:
		return m.updateLoginPassword(msg)
	case TextState:
		return m.updateText(msg)
	case FileState:
		return m.updateFile(msg)
	case BankCardState:
		return m.updateBankCard(msg)
	default:
		return m, nil
	}
}

// updateWindowsSize - метод обновления размеров окон
func (m SecretAddModel) updateWindowsSize(msg tea.WindowSizeMsg) (SecretAddModel, tea.Cmd) {
	m.windowSize = msg

	// Передаем размеры окна всем дочерним моделям
	loginModel, loginModelCmd := m.loginModel.Update(msg)
	m.loginModel = loginModel

	fileModel, fileModelCmd := m.fileModel.Update(msg)
	m.fileModel = fileModel

	textModel, textModelCmd := m.textModel.Update(msg)
	m.textModel = textModel

	cardModel, cardModelCmd := m.cardModel.Update(msg)
	m.cardModel = cardModel

	return m, tea.Batch(loginModelCmd, fileModelCmd, textModelCmd, cardModelCmd)
}

// updateTypeSelect - метод обработки выбора типа
func (m SecretAddModel) updateTypeSelect(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.focusedBtn > 0 {
				m.focusedBtn--
			}
			return m, nil

		case "down", "j":
			if m.focusedBtn < len(m.secretTypes)-1 {
				m.focusedBtn++
			}
			return m, nil

		case "enter":
			switch m.focusedBtn {
			case 0: // Логин/Пароль
				m.state = LoginPasswordState
				return m, nil
			case 1: // Текст
				m.state = TextState
				return m, nil
			case 2: // Файл
				m.state = FileState
				return m, nil
			case 3: // Банковская карта
				m.state = BankCardState
				return m, nil
			}
			return m, nil

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}
	return m, nil
}

// updateLoginPassword - метод обновления окна секрета (пароль/логин)
func (m SecretAddModel) updateLoginPassword(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	updatedModel, cmd := m.loginModel.Update(msg)
	m.loginModel = updatedModel
	return m, cmd
}

// updateText - метод обновления окна секрета (текст)
func (m SecretAddModel) updateText(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	updatedModel, cmd := m.textModel.Update(msg)
	m.textModel = updatedModel
	return m, cmd
}

// updateFile - метод обновления окна секрета (файл)
func (m SecretAddModel) updateFile(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	updatedModel, cmd := m.fileModel.Update(msg)
	m.fileModel = updatedModel
	return m, cmd
}

// updateBankCard - метод обновления окна секрета (банковская карта)
func (m SecretAddModel) updateBankCard(msg tea.Msg) (SecretAddModel, tea.Cmd) {
	updatedModel, cmd := m.cardModel.Update(msg)
	m.cardModel = updatedModel
	return m, cmd
}

// View - метод отрисовки текущего состояния
func (m SecretAddModel) View() string {
	switch m.state {
	case SecretTypeSelectState:
		return m.renderTypeSelectView()
	case LoginPasswordState:
		return m.loginModel.View()
	case TextState:
		return m.textModel.View()
	case FileState:
		return m.fileModel.View()
	case BankCardState:
		return m.cardModel.View()
	default:
		return "Неизвестное состояние"
	}
}

// renderTypeSelectView - метод для отрисовки кнопок в окне выбора типа создаваемого секрета
func (m SecretAddModel) renderTypeSelectView() string {
	// Создаем кнопки выбора типа
	buttons := make([]string, len(m.secretTypes))
	for i, secretType := range m.secretTypes {
		if i == m.focusedBtn {
			buttons[i] = styles.ActiveButtonStyle.
				Width(20).
				Height(3).
				Render(secretType)
		} else {
			buttons[i] = styles.ButtonStyle.
				Width(20).
				Height(3).
				Render(secretType)
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("➕ Выберите тип секрета"),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinVertical(lipgloss.Center, buttons...),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render("↑/↓: выбор типа • Enter: подтвердить • ESC: отмена"),
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
