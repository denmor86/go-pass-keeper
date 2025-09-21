package models

import (
	"fmt"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SecretsState int

const (
	SecretsListState SecretsState = iota
	SecretViewState
	SecretAddState
)

type Secret struct {
	ID      int
	Name    string
	Type    string
	Created string
	Updated string
	Value   string
	Login   string
}

type SecretsModel struct {
	state          SecretsState
	table          table.Model
	secrets        []Secret
	windowSize     tea.WindowSizeMsg
	focusedBtn     int
	selectedSecret *Secret
	addModel       SecretAddModel
}

func NewSecretsModel() SecretsModel {
	secrets := getInitialSecrets()

	return SecretsModel{
		state:      SecretsListState,
		table:      createTable(secrets),
		secrets:    secrets,
		focusedBtn: 0,
		addModel:   NewSecretAddModel(),
	}
}

func (m SecretsModel) Init() tea.Cmd {
	return nil
}

func (m SecretsModel) Update(msg tea.Msg) (SecretsModel, tea.Cmd) {
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
			case SecretsListState:
				// Возврат на главный экран
				return m, nil
			case SecretViewState, SecretAddState:
				m.state = SecretsListState
				return m, nil
			}
		}
	}
	switch m.state {
	case SecretAddState:
		return m.handleAddState(msg)
	case SecretViewState:
		return m.handleViewState(msg)
	default:
		return m.handleListState(msg)
	}
}

// updateWindowsSize - метод обновления размеров окон
func (m SecretsModel) updateWindowsSize(msg tea.WindowSizeMsg) (SecretsModel, tea.Cmd) {
	m.windowSize = msg

	// Передаем размеры окна всем дочерним моделям
	updatedAddModel, addModelCmd := m.addModel.Update(msg)
	m.addModel = updatedAddModel

	return m, tea.Batch(addModelCmd)
}

func (m SecretsModel) handleListState(msg tea.Msg) (SecretsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R": // Обновление
			return m.refreshSecrets(), nil

		case "left", "h": // Навигация кнопок
			if m.focusedBtn > 0 {
				m.focusedBtn--
			}
			return m, nil

		case "right", "l": // Навигация кнопок
			if m.focusedBtn < 3 {
				m.focusedBtn++
			}
			return m, nil

		case "enter": // Обработка действий
			return m.handleEnterAction()

		case "esc": // Выход из секретов
			return m, nil
		}
	}

	// Обновление таблицы только если мы в состоянии списка
	if m.state == SecretsListState {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m SecretsModel) handleAddState(msg tea.Msg) (SecretsModel, tea.Cmd) {
	// Передаем сообщение в модель добавления
	updatedModel, cmd := m.addModel.Update(msg)
	m.addModel = updatedModel

	// Проверяем, не завершили ли мы добавление
	switch msg := msg.(type) {
	case messages.SecretAddCompleteMsg:
		m = m.handleAddComplete(msg)
		return m, nil
	}

	return m, cmd
}

func (m SecretsModel) handleViewState(msg tea.Msg) (SecretsModel, tea.Cmd) {
	// ESC в окне просмотра - возврат к списку секретов
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.state = SecretsListState
		m.selectedSecret = nil
	}
	return m, nil
}

func (m SecretsModel) handleAddComplete(msg messages.SecretAddCompleteMsg) SecretsModel {
	newSecret := Secret{
		ID:      getNextID(m.secrets),
		Name:    msg.Name,
		Type:    msg.Type,
		Login:   msg.Login,
		Value:   msg.Password,
		Created: time.Now().Format("2006-01-02"),
		Updated: time.Now().Format("2006-01-02"),
	}

	if msg.Content != "" {
		newSecret.Value = msg.Content
	}
	if msg.FileName != "" {
		newSecret.Value = "Файл: " + msg.FileName
		if msg.Content != "" && msg.Content != "Файл не найден или недоступен" {
			newSecret.Value = msg.Content
		}
	}

	m.secrets = append(m.secrets, newSecret)
	m.table.SetRows(createTableRows(m.secrets))
	m.state = SecretsListState

	return m
}

func (m SecretsModel) handleEnterAction() (SecretsModel, tea.Cmd) {
	// Если выбрана кнопка "Просмотр" И есть выбранная строка
	if m.focusedBtn == 1 && m.table.SelectedRow() != nil {
		selectedID, _ := strconv.Atoi(m.table.SelectedRow()[0])
		for i := range m.secrets {
			if m.secrets[i].ID == selectedID {
				m.selectedSecret = &m.secrets[i]
				m.state = SecretViewState
				return m, nil
			}
		}
	}

	// Если выбрана кнопка "Добавить"
	if m.focusedBtn == 0 {
		m.state = SecretAddState
		return m, m.addModel.Init()
	}

	// Если выбрана кнопка "Удалить" и есть выбранная строка
	if m.focusedBtn == 2 && m.table.SelectedRow() != nil {
		selectedID, _ := strconv.Atoi(m.table.SelectedRow()[0])
		m.secrets = deleteSecret(m.secrets, selectedID)
		m.table.SetRows(createTableRows(m.secrets))
		return m, nil
	}

	// Если выбрана кнопка "Обновить"
	if m.focusedBtn == 3 {
		m = m.refreshSecrets()
		return m, nil
	}

	return m, nil
}

func (m SecretsModel) refreshSecrets() SecretsModel {
	m.secrets = getInitialSecrets()
	m.table.SetRows(createTableRows(m.secrets))
	return m
}

func (m SecretsModel) View() string {
	switch m.state {
	case SecretsListState:
		return m.renderSecretsListView()
	case SecretViewState:
		return m.renderSecretView()
	case SecretAddState:
		return m.addModel.View()
	default:
		return "Неизвестное состояние"
	}
}

// Вспомогательные функции
func getInitialSecrets() []Secret {
	return []Secret{
		{
			ID:      1,
			Name:    "Пароль от почты",
			Type:    "Логин/Пароль",
			Created: time.Now().Format("2006-01-02"),
			Updated: time.Now().Format("2006-01-02"),
			Value:   "mysecretpassword123",
			Login:   "user@example.com",
		},
		{
			ID:      2,
			Name:    "API ключ",
			Type:    "Текст",
			Created: time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			Updated: time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
			Value:   "ak_1234567890abcdef",
			Login:   "",
		},
		{
			ID:      3,
			Name:    "config.txt",
			Type:    "Файл",
			Created: time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
			Updated: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			Value:   "секретные настройки",
			Login:   "",
		},
	}
}

func createTable(secrets []Secret) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Название", Width: 20},
		{Title: "Тип", Width: 12},
		{Title: "Создан", Width: 12},
		{Title: "Обновлен", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(createTableRows(secrets)),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = styles.TableHeaderStyle.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true)

	s.Selected = styles.TableSelectedStyle.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	t.SetStyles(s)

	return t
}

func createTableRows(secrets []Secret) []table.Row {
	rows := make([]table.Row, len(secrets))
	for i, secret := range secrets {
		rows[i] = table.Row{
			strconv.Itoa(secret.ID),
			secret.Name,
			secret.Type,
			secret.Created,
			secret.Updated,
		}
	}
	return rows
}

func deleteSecret(secrets []Secret, id int) []Secret {
	for i, secret := range secrets {
		if secret.ID == id {
			return append(secrets[:i], secrets[i+1:]...)
		}
	}
	return secrets
}

func getNextID(secrets []Secret) int {
	maxID := 0
	for _, secret := range secrets {
		if secret.ID > maxID {
			maxID = secret.ID
		}
	}
	return maxID + 1
}

// Методы отображения
func (m SecretsModel) renderSecretsListView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(m.windowSize.Width-10).
			Render("🔒 Управление секретами"),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.TableStyle.
			Width(m.table.Width()).
			Render(m.table.View()),

		lipgloss.NewStyle().Height(2).Render(""),

		m.renderButtons(),

		lipgloss.NewStyle().Height(1).Render(""),

		m.renderHelpText(),
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

func (m SecretsModel) renderButtons() string {
	buttons := []string{
		m.renderButton("➕ Добавить", 0),
		m.renderButton("👁️ Просмотр", 1),
		m.renderButton("🗑️ Удалить", 2),
		m.renderButton("🔄 Обновить", 3),
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		buttons...,
	)
}

func (m SecretsModel) renderButton(text string, index int) string {
	if index == m.focusedBtn {
		return styles.ActiveButtonStyle.
			Width(15).
			Height(2).
			Render(text)
	}
	return styles.ButtonStyle.
		Width(15).
		Height(2).
		Render(text)
}

func (m SecretsModel) renderHelpText() string {
	helpText := "↑/↓: выбор секрета • ←/→: выбор кнопки • Enter: действие • R: обновить • ESC: выход"

	if m.table.SelectedRow() != nil {
		helpText += " • Выбрано: " + m.table.SelectedRow()[1]
	}

	return lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		Italic(true).
		Render(helpText)
}

func (m SecretsModel) renderSecretView() string {
	if m.selectedSecret == nil {
		return "Ошибка: секрет не выбран"
	}

	secret := m.selectedSecret
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("👁️ Просмотр секрета"),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderSecretField("ID:", fmt.Sprintf("%d", secret.ID)),
			m.renderSecretField("Название:", secret.Name),
			m.renderSecretField("Тип:", secret.Type),
			m.renderSecretField("Создан:", secret.Created),
			m.renderSecretField("Обновлен:", secret.Updated),
			m.renderSecretField("Логин:", secret.Login),
			m.renderSecretField("Значение:", secret.Value),
		),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.ButtonStyle.
			Width(20).
			Render("ESC: Назад к списку"),
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

func (m SecretsModel) renderSecretField(label, value string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Width(12).
			Render(label),
		lipgloss.NewStyle().
			Foreground(styles.TextPrimary).
			Bold(true).
			Render(value),
	) + "\n"
}
