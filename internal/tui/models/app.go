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

type AppState int

const (
	AuthState AppState = iota
	LoginState
	RegisterState
	MainState
	ViewState
	SettingsState
)

// Кнопки на гланном окне
const (
	LoginButton = iota
	RegisterButton
	ViewButton
	SettingsButton
)

type AppModel struct {
	state      AppState
	login      LoginModel
	register   RegisterModel
	main       MainModel
	view       ViewModel
	settings   SettingsModel
	viewport   viewport.Model
	windowSize tea.WindowSizeMsg
	focused    int
	username   string
	token      string
	config     *config.Config
}

func NewAppModel(config *config.Config) AppModel {

	connection := config.Load()
	return AppModel{
		state:    AuthState,
		login:    NewLoginModel(connection),
		register: NewRegisterModel(connection),
		main:     NewMainModel(),
		view:     NewViewModel(),
		settings: NewSettingsModel(connection),
		viewport: viewport.New(60, 10),
		focused:  0,
		username: "",
		token:    "",
		config:   config,
	}
}

func (m AppModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10

		// Передаем размеры окна всем дочерним моделям
		updatedLogin, loginCmd := m.login.Update(msg)
		m.login = updatedLogin

		updatedRegister, registerCmd := m.register.Update(msg)
		m.register = updatedRegister

		updatedMain, mainCmd := m.main.Update(msg, m.viewport)
		m.main = updatedMain
		m.viewport = m.main.viewport

		updatedView, viewCmd := m.view.Update(msg)
		m.view = updatedView

		updatedSettings, settingsCmd := m.settings.Update(msg)
		m.settings = updatedSettings

		return m, tea.Batch(loginCmd, registerCmd, mainCmd, viewCmd, settingsCmd)

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
			case MainState, ViewState, SettingsState:
				// Выход из главного экрана или просмотра
				m.state = AuthState
				return m, nil
			case AuthState:
				// Выход из приложения
				return m, tea.Quit
			}
		}

	case messages.AuthSuccessMsg:
		m.state = MainState
		m.username = msg.Email
		m.token = msg.Token
		m.main.username = m.username
		m.main.message = fmt.Sprintf("Добро пожаловать, %s!", m.username)
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

	// Делегируем обновление текущему состоянию
	switch m.state {
	case AuthState:
		return m.handleAuthUpdate(msg)
	case LoginState:
		return m.handleLoginUpdate(msg)
	case RegisterState:
		return m.handleRegisterUpdate(msg)
	case MainState:
		return m.handleMainUpdate(msg)
	case ViewState:
		return m.handleViewUpdate(msg)
	case SettingsState:
		return m.handleSettingsUpdate(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	switch m.state {
	case AuthState:
		return m.renderAuthView()
	case LoginState:
		return m.login.View()
	case RegisterState:
		return m.register.View()
	case MainState:
		return m.main.View(m.viewport)
	case ViewState:
		return m.view.View(m.username)
	case SettingsState:
		return m.settings.View()
	default:
		return "Неизвестное состояние"
	}
}

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
			m.renderViewButton()),
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

func (m AppModel) renderViewButton() string {
	text := "👁️ Просмотр"

	if ViewButton == m.focused {
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
		return styles.DisabledButtonStyle.
			Margin(0, 0, 1, 0).
			Render(text + " (требуется вход)")
	}
	return styles.ButtonStyle.
		Margin(0, 0, 1, 0).
		Render(text)
}

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
			case ViewButton:
				if m.isAuthorized() {
					m.state = ViewState
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

func (m AppModel) handleLoginUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.login.Update(msg)
	m.login = updatedModel
	return m, cmd
}

func (m AppModel) handleRegisterUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.register.Update(msg)
	m.register = updatedModel
	return m, cmd
}

func (m AppModel) handleMainUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.main.Update(msg, m.viewport)
	m.main = updatedModel
	m.viewport = m.main.viewport
	return m, cmd
}

func (m AppModel) handleViewUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.view.Update(msg)
	m.view = updatedModel
	return m, cmd
}

func (m AppModel) handleSettingsUpdate(msg tea.Msg) (AppModel, tea.Cmd) {
	updatedModel, cmd := m.settings.Update(msg)
	m.settings = updatedModel
	return m, cmd
}

func (m AppModel) isAuthorized() bool {
	return len(m.token) > 0
}
