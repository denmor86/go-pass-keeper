package models

import (
	"context"
	"fmt"
	"go-pass-keeper/internal/grpcclient"
	"go-pass-keeper/internal/grpcclient/settings"
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"
	"go-pass-keeper/pkg/crypto"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewerState int

const (
	ViewerListState ViewerState = iota
	SecretViewState
	SecretAddState
)

// Кнопки на главном окне
const (
	AddButton = iota
	ViewButton
	DeleteButton
	UpdateButton
)

type ViewerModel struct {
	state      ViewerState
	table      table.Model
	secrets    []*models.SecretInfo
	windowSize tea.WindowSizeMsg
	focusedBtn int
	addModel   SecretAddModel
	connection *settings.Connection
	token      string
	cryptoKey  []byte
}

func NewViewerModel(connection *settings.Connection) ViewerModel {
	return ViewerModel{
		state:      ViewerListState,
		table:      createTable(),
		focusedBtn: 0,
		addModel:   NewSecretAddModel(),
		connection: connection,
	}
}
func (m ViewerModel) Init() tea.Cmd {
	return m.attemptGetSecrets()
}

func (m ViewerModel) Update(msg tea.Msg) (ViewerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateWindowsSize(msg)

	case messages.SecretAddCancelMsg:
		m.state = ViewerListState
		return m.handleAddState(msg)

	case messages.AuthSuccessMsg:
		return m.handleAuthAction(msg)

	case messages.AddSecretPasswordMsg:
		m.state = ViewerListState
		return m, m.attemptAddSecret(&msg)

	case messages.AddSecretCardMsg:
		m.state = ViewerListState
		return m, m.attemptAddSecret(&msg)

	case messages.AddSecretTextMsg:
		m.state = ViewerListState
		return m, m.attemptAddSecret(&msg)

	case messages.AddSecretBinaryMsg:
		m.state = ViewerListState
		return m, m.attemptAddSecret(&msg)

	case messages.SecretRefreshMsg:
		m.secrets = msg.Secrets
		return m.refreshViewer(), nil
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
func (m ViewerModel) updateWindowsSize(msg tea.WindowSizeMsg) (ViewerModel, tea.Cmd) {
	m.windowSize = msg

	// Передаем размеры окна всем дочерним моделям
	updatedAddModel, addModelCmd := m.addModel.Update(msg)
	m.addModel = updatedAddModel

	return m, tea.Batch(addModelCmd)
}

func (m ViewerModel) handleListState(msg tea.Msg) (ViewerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R": // Обновление
			return m.refreshViewer(), nil

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
			return m, func() tea.Msg {
				return messages.GotoMainPageMsg{}
			}
		}
	}

	// Обновление таблицы только если мы в состоянии списка
	if m.state == ViewerListState {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m ViewerModel) handleAddState(msg tea.Msg) (ViewerModel, tea.Cmd) {
	// Передаем сообщение в модель добавления
	updatedModel, cmd := m.addModel.Update(msg)
	m.addModel = updatedModel
	return m, cmd
}

func (m ViewerModel) handleViewState(msg tea.Msg) (ViewerModel, tea.Cmd) {
	// ESC в окне просмотра - возврат к списку секретов
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.state = ViewerListState
	}
	return m, nil
}

func (m ViewerModel) handleEnterAction() (ViewerModel, tea.Cmd) {
	// Если выбрана кнопка "Добавить"
	if m.focusedBtn == AddButton {
		m.state = SecretAddState
		return m, m.addModel.Init()
	}
	// Если выбрана кнопка "Обновить"
	if m.focusedBtn == UpdateButton {
		return m, m.attemptGetSecrets()
	}

	if len(m.table.Rows()) == 0 {
		return m, nil
	}
	selectedID := m.table.SelectedRow()[0]

	// Если выбрана кнопка "Просмотр" И есть выбранная строка
	if m.focusedBtn == ViewButton && m.table.SelectedRow() != nil {
		m.state = SecretAddState
		return m, m.attemptGetSecret(selectedID)
	}

	// Если выбрана кнопка "Удалить" и есть выбранная строка
	if m.focusedBtn == DeleteButton && m.table.SelectedRow() != nil {
		return m, m.attemptDeleteSecret(selectedID)
	}

	return m, nil
}

// handleAuthAction - обработчик добавления секрета
func (m ViewerModel) handleAuthAction(msg messages.AuthSuccessMsg) (ViewerModel, tea.Cmd) {
	m.token = msg.Token
	key, err := crypto.MakeCryptoKey("secret", msg.Salt)
	if err != nil {
		return m, func() tea.Msg {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка формирования ключа: %s", err.Error()))
		}
	}
	m.cryptoKey = key
	return m, nil
}

func (m ViewerModel) refreshViewer() ViewerModel {
	m.table.SetRows(createTableRows(m.secrets))
	return m
}

func (m ViewerModel) View() string {
	switch m.state {
	case ViewerListState:
		return m.renderViewerListView()
	case SecretAddState:
		return m.addModel.View()
	default:
		return "Неизвестное состояние"
	}
}

func createTable() table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Название", Width: 20},
		{Title: "Тип", Width: 12},
		{Title: "Создан", Width: 12},
		{Title: "Обновлен", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
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

func createTableRows(secrets []*models.SecretInfo) []table.Row {
	rows := make([]table.Row, len(secrets))
	for i, secret := range secrets {
		rows[i] = table.Row{
			secret.ID,
			secret.Name,
			secret.Type,
			secret.Created.Local().Format(time.UnixDate),
			secret.Updated.Local().Format(time.UnixDate),
		}
	}
	return rows
}

// Методы отображения
func (m ViewerModel) renderViewerListView() string {
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

func (m ViewerModel) renderButtons() string {
	buttons := []string{
		m.renderButton("➕ Добавить", AddButton),
		m.renderButton("👁️ Просмотр", ViewButton),
		m.renderButton("🗑️ Удалить", DeleteButton),
		m.renderButton("🔄 Обновить", UpdateButton),
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		buttons...,
	)
}

func (m ViewerModel) renderButton(text string, index int) string {
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

func (m ViewerModel) renderHelpText() string {
	helpText := "↑/↓: выбор секрета • ←/→: выбор кнопки • Enter: действие • R: обновить • ESC: выход"

	if m.table.SelectedRow() != nil {
		helpText += " • Выбрано: " + m.table.SelectedRow()[1]
	}

	return lipgloss.NewStyle().
		Foreground(styles.TextSecondary).
		Italic(true).
		Render(helpText)
}

// attemptGetSecrets - обработчик получения секретов
func (m ViewerModel) attemptGetSecrets() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.connection.Timeout)*time.Second)
		client := grpcclient.NewKeeperClient(m.connection.ServerAddress(), m.token)
		defer func() {
			cancel()
			client.Close()
		}()
		if err := client.Connect(ctx); err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка подключения к %s: %s", m.connection.ServerAddress(), err.Error()))
		}
		secrets, err := client.GetSecrets()
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка получения данных: %s", err.Error()))
		}
		return messages.SecretRefreshMsg{Secrets: secrets}
	}
}

// attemptDeleteSecret - обработчик удаления секрета
func (m ViewerModel) attemptDeleteSecret(sid string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.connection.Timeout)*time.Second)
		client := grpcclient.NewKeeperClient(m.connection.ServerAddress(), m.token)
		defer func() {
			cancel()
			client.Close()
		}()
		if err := client.Connect(ctx); err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка подключения к %s: %s", m.connection.ServerAddress(), err.Error()))
		}
		id, err := client.DeleteSecret(sid)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка удаления секрета: %s", err.Error()))
		}
		return messages.SecretDeleteMsg{Id: id}
	}
}

// attemptAddSecret - обработчик добавления секрета
func (m ViewerModel) attemptAddSecret(converter messages.EncryptConverter) tea.Cmd {
	return func() tea.Msg {
		info, content, err := converter.ToModel(m.cryptoKey)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка добавления секрета: %s", err.Error()))
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.connection.Timeout)*time.Second)
		client := grpcclient.NewKeeperClient(m.connection.ServerAddress(), m.token)
		defer func() {
			cancel()
			client.Close()
		}()
		if err := client.Connect(ctx); err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка подключения к %s: %s", m.connection.ServerAddress(), err.Error()))
		}
		_, err = client.AddSecret(info, content)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка добавления секрета: %s", err.Error()))
		}
		return messages.SecretUpdateMsg{}
	}
}

// attemptGetSecret - обработчик получения секрета
func (m ViewerModel) attemptGetSecret(sid string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.connection.Timeout)*time.Second)
		client := grpcclient.NewKeeperClient(m.connection.ServerAddress(), m.token)
		defer func() {
			cancel()
			client.Close()
		}()
		if err := client.Connect(ctx); err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка подключения к %s: %s", m.connection.ServerAddress(), err.Error()))
		}
		info, content, err := client.GetSecret(sid)
		if err != nil {
			return messages.ErrorMsg(fmt.Sprintf("Ошибка добавления секрета: %s", err.Error()))
		}
		return messages.ToMessage(m.cryptoKey, info, content)
	}
}
