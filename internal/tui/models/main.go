package models

import (
	"go-pass-keeper/internal/tui/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainModel struct {
	username   string
	message    string
	viewport   viewport.Model
	windowSize tea.WindowSizeMsg
}

func NewMainModel() MainModel {
	return MainModel{
		viewport: viewport.New(60, 10),
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg, vp viewport.Model) (MainModel, tea.Cmd) {
	m.viewport = vp
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		// Обновляем размер вьюпорта
		m.viewport.Width = msg.Width - 6
		m.viewport.Height = msg.Height - 16 // Увеличиваем отступ для имени пользователя
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m MainModel) View(vp viewport.Model) string {
	m.viewport = vp
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(50).
			Render("🏠 Главная страница"),

		lipgloss.NewStyle().Height(1).Render(""),

		// Отображаем имя пользователя
		lipgloss.NewStyle().
			Foreground(styles.SuccessColor).
			Bold(true).
			Render("👤 Пользователь: "+m.username),

		lipgloss.NewStyle().Height(1).Render(""),

		styles.SuccessStyle.
			Render("✅ "+m.message),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.ContentStyle.
			Width(m.windowSize.Width-10).
			Height(m.windowSize.Height-18). // Увеличиваем отступ для нового элемента
			Render(m.viewport.View()),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.HelpStyle.
			Render("Нажмите ESC для выхода в меню"),
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
