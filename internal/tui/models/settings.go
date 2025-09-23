package models

import (
	"fmt"
	"go-pass-keeper/internal/grpcclient/settings"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsModel - модель окна настроек
type SettingsModel struct {
	inputs     []textinput.Model
	focused    int
	windowSize tea.WindowSizeMsg
	connection *settings.Settings
}

// Константы для именованных индексов полей
const (
	fieldServerURL = iota
	fieldServerPort
	fieldTimeout
	fieldSecretPassword
)

// NewSettingsModel - метод для создания окна настроек
func NewSettingsModel(connection *settings.Settings) SettingsModel {
	model := SettingsModel{
		inputs:     make([]textinput.Model, 4),
		connection: connection,
	}

	// Инициализация полей ввода
	for i := range model.inputs {
		t := textinput.New()
		t.Cursor.Style = styles.FocusedStyle
		t.TextStyle = styles.BlurredStyle
		t.CharLimit = 50

		switch i {
		case fieldServerURL:
			t.Placeholder = "localhost"
			t.Prompt = "URL сервера: "
			t.SetValue(connection.ServerURL)
		case fieldServerPort:
			t.Placeholder = "8080"
			t.Prompt = "Порт: "
			t.SetValue(connection.ServerPort)
		case fieldTimeout:
			t.Placeholder = "30"
			t.Prompt = "Таймаут (секунды): "
			t.SetValue(fmt.Sprintf("%d", connection.Timeout))
		case fieldSecretPassword:
			t.Placeholder = "Секрет"
			t.Prompt = "Секрет: "
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.SetValue(connection.Secret)
		}

		model.inputs[i] = t
	}

	return model
}

// Init - метод инициализации окна
func (m SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод для обновления окна по внешним сообщениям
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" {
				// Сохраняем настройки
				newConnection := settings.Settings{
					ServerURL:  m.inputs[fieldServerURL].Value(),
					ServerPort: m.inputs[fieldServerPort].Value(),
					Secret:     m.inputs[fieldSecretPassword].Value(),
				}

				// Парсим таймаут
				if timeout, err := strconv.Atoi(m.inputs[fieldTimeout].Value()); err == nil {
					newConnection.Timeout = timeout
				} else {
					newConnection.Timeout = 30
				}

				return m, func() tea.Msg {
					return messages.ConfigUpdatedMsg{Connection: newConnection}
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs)-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			for i := range m.inputs {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].Focus())
					m.inputs[i].PromptStyle = styles.FocusedStyle
					m.inputs[i].TextStyle = styles.FocusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = styles.BlurredStyle
				m.inputs[i].TextStyle = styles.BlurredStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View - метод для отрисовки окна, в зависимости от текущего состояния
func (m SettingsModel) View() string {
	// Поля ввода
	fields := make([]string, len(m.inputs))
	for i := range m.inputs {
		var inputStyle lipgloss.Style
		if i == m.focused {
			inputStyle = styles.FocusedInputFieldStyle
		} else {
			inputStyle = styles.InputFieldStyle
		}

		fields[i] = lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(20).Render(m.inputs[i].Prompt),
			inputStyle.Width(30).Render(m.inputs[i].View()),
		)
	}

	// Основной контент
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(50).
			Render("⚙️ Настройки"),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.JoinVertical(lipgloss.Left, fields...),

		lipgloss.NewStyle().Height(2).Render(""),

		// Кнопки действий
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Сохранить"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Отмена"),
		),

		lipgloss.NewStyle().Height(1).Render(""),

		// Текущие настройки
		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render(fmt.Sprintf("Текущее подключение: %s:%s",
				m.connection.ServerURL,
				m.connection.ServerPort)),
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
