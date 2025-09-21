package models

import (
	"fmt"
	"go-pass-keeper/internal/grpcclient/config"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Тип переменной состояние перехода
type AppState int

// Состояние переходов
const (
	AuthState AppState = iota
	LoginState
	RegisterState
	SecretState
	SettingsState
)

// Кнопки на главном окне
const (
	LoginButton = iota
	RegisterButton
	SecretButton
	SettingsButton
)

// AppModel - модель главного окна
type AppModel struct {
	state      AppState
	login      LoginModel
	register   RegisterModel
	secrets    SecretsModel
	settings   SettingsModel
	viewport   viewport.Model
	windowSize tea.WindowSizeMsg
	focused    int
	username   string
	token      string
	config     *config.Config
}

// NewAppModel - метод для создания главного окна
func NewAppModel(config *config.Config) AppModel {

	connection := config.Load()
	return AppModel{
		state:    AuthState,
		login:    NewLoginModel(connection),
		register: NewRegisterModel(connection),
		secrets:  NewSecretsModel(),
		settings: NewSettingsModel(connection),
		viewport: viewport.New(80, 20),
		focused:  0,
		username: "",
		token:    "",
		config:   config,
	}
}

// Init - метод инициализации окна
func (m AppModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод для обновления окна по внешним сообщениям
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateWindowsSize(msg)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			// Обработка ESC в зависимости от текущего состояния
			switch m.state {
			case LoginState, RegisterState:
				// Возврат на главный экран
				m.state = AuthState
				m.login.err = ""
				m.register.err = ""
				return m, nil
			case SettingsState:
				// Выход из главного экрана или просмотра
				m.state = AuthState
				return m, nil
			case AuthState:
				// Выход из приложения
				return m, tea.Quit
			}
		}

	case messages.AuthSuccessMsg:
		m.state = AuthState
		m.username = msg.Email
		m.token = msg.Token
		m.viewport.SetContent(fmt.Sprintf("Приветствуем, %s!\n\nВы успешно вошли в систему.\n\nЗдесь может быть ваш контент...", m.username))
		return m, nil

	case messages.ErrorMsg:
		switch m.state {
		case LoginState:
			m.login.err = msg
		case RegisterState:
			m.register.err = msg
		}
		return m, nil

	case messages.ConfigUpdatedMsg:
		m.config.Save(&msg.Connection)
		m.state = AuthState
		return m, nil
	}

	// Обновление текущего состояния
	switch m.state {
	case AuthState:
		return m.handleAuthUpdate(msg)
	case LoginState:
		return m.handleLoginUpdate(msg)
	case RegisterState:
		return m.handleRegisterUpdate(msg)
	case SecretState:
		return m.handleSecretUpdate(msg)
	case SettingsState:
		return m.handleSettingsUpdate(msg)
	}

	return m, tea.Batch(cmds...)
}

// View - метод для отрисовки окна, в зависимости от текущего состояния
func (m AppModel) View() string {
	switch m.state {
	case AuthState:
		return m.renderAuthView()
	case LoginState:
		return m.login.View()
	case RegisterState:
		return m.register.View()
	case SecretState:
		return m.secrets.View()
	case SettingsState:
		return m.settings.View()
	default:
		return "Неизвестное состояние"
	}
}

// updateWindowsSize - метод обновления размеров окон
func (m AppModel) updateWindowsSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.windowSize = msg
	m.viewport.Width = msg.Width
	m.viewport.Height = msg.Height

	// Передаем размеры окна всем дочерним моделям
	updatedLogin, loginCmd := m.login.Update(msg)
	m.login = updatedLogin

	updatedRegister, registerCmd := m.register.Update(msg)
	m.register = updatedRegister

	updatedSecrets, secretsCmd := m.secrets.Update(msg)
	m.secrets = updatedSecrets

	updatedSettings, settingsCmd := m.settings.Update(msg)
	m.settings = updatedSettings

	return m, tea.Batch(loginCmd, registerCmd, secretsCmd, settingsCmd)
}

// renderAuthView - метод отрисовки основного окна
func (m AppModel) renderAuthView() string {
	// Статус пользователя
	userStatus := m.getUserStatus()

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(50).
			Render("✨ Добро пожаловать!"),

		userStatus,
		lipgloss.NewStyle().Height(1).Render(""),

		styles.SubtitleStyle.
			Width(50).
			Render("Выберите действие для продолжения работы"),

		lipgloss.NewStyle().Height(2).Render(""),
		lipgloss.JoinVertical(lipgloss.Center,
			m.renderLoginButton(),
			m.renderRegisterButton(),
			m.renderSecretButton()),
		lipgloss.NewStyle().Height(2).Render(""),
		m.renderSettingsButton(),
		lipgloss.NewStyle().Height(1).Render(""),

		styles.HelpStyle.
			Render("↑/↓: выбор • Enter: подтвердить • S: настройки • ESC: выход"),
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

// getUserStatus - метод для формирования текущего статуса авторизации
func (m AppModel) getUserStatus() string {
	if m.isAuthorized() {
		return lipgloss.NewStyle().
			Foreground(styles.SuccessColor).
			Bold(true).
			Padding(0, 1).
			Render("👤 Вы вошли как: " + m.username)
	}
	return lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		Italic(true).
		Render("🔒 Не авторизован")
}

// renderLoginButton - метод отрисовки кнопки авторизации пользователя
func (m AppModel) renderLoginButton() string {
	text := "🔐 Вход в систему"
	if LoginButton == m.focused {
		return styles.ActiveButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text)
	}
	return styles.ButtonStyle.
		Margin(0, 0, 1, 0).
		Render(text)
}

// renderRegisterButton - метод отрисовки кнопки регистрации пользователя
func (m AppModel) renderRegisterButton() string {
	text := "📝 Регистрация"
	if RegisterButton == m.focused {
		return styles.ActiveButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text)
	}
	return styles.ButtonStyle.
		Margin(0, 0, 1, 0).
		Render(text)
}

// renderSecretButton - метод отрисовки кнопки просмотра секретов
func (m AppModel) renderSecretButton() string {
	text := "👁️ Просмотр"

	if SecretButton == m.focused {
		if m.isAuthorized() {
			return styles.ActiveButtonStyle.
				Margin(0, 0, 1, 0).
				Render(text)
		}
		return styles.DisabledActiveButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text + " (требуется вход)")
	}
	if m.isAuthorized() {
		return styles.ButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text)
	}
	return styles.DisabledButtonStyle.
		Margin(0, 0, 1, 0).
		Render(text + " (требуется вход)")

}

// renderSettingsButton - метод отрисовки кнопки настроек клиента
func (m AppModel) renderSettingsButton() string {
	text := "⚙️ Настройки подключения"
	if SettingsButton == m.focused {
		return styles.ActiveSmallButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text)
	}
	return styles.SmallButtonStyle.
		Margin(0, 0, 1, 0).
		Render(text)
}

// handleAuthUpdate - метод обработчик действий на кнопках основного окна
func (m AppModel) handleAuthUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.focused >= RegisterButton {
				m.focused--
			}
			return m, nil
		case "down", "j":
			if m.focused < SettingsButton {
				m.focused++
			}
			return m, nil
		case "enter":
			switch m.focused {
			case LoginButton:
				m.state = LoginState
				return m, m.login.inputs[0].Focus()
			case RegisterButton:
				m.state = RegisterState
				return m, m.register.inputs[0].Focus()
			case SecretButton:
				if m.isAuthorized() {
					m.state = SecretState
					return m, nil
				}
				return m, nil
			case SettingsButton:
				m.state = SettingsState
				return m, m.settings.inputs[0].Focus()
			}
		}
	}
	return m, nil
}

// handleLoginUpdate - метод обработчик действий на кнопке авторизации пользователя
func (m AppModel) handleLoginUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.login.Update(msg)
	m.login = updatedModel
	return m, cmd
}

// handleRegisterUpdate - метод обработчик действий на кнопке регистрации пользователя
func (m AppModel) handleRegisterUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.register.Update(msg)
	m.register = updatedModel
	return m, cmd
}

// handleSecretUpdate - метод обработчик действий на кнопке просмотра секретов
func (m AppModel) handleSecretUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.secrets.Update(msg)
	m.secrets = updatedModel
	return m, cmd
}

// handleSettingsUpdate - метод обработчик действий на кнопке настроек клиента
func (m AppModel) handleSettingsUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.settings.Update(msg)
	m.settings = updatedModel
	return m, cmd
}

// isAuthorized - метод определения наличия авторизации пользователя
func (m AppModel) isAuthorized() bool {
	return true
	//return len(m.token) > 0
}
