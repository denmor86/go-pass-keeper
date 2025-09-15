package models

import (
	"go-pass-keeper/internal/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewModel - модель окна просмотра секретов
type ViewModel struct {
	windowSize tea.WindowSizeMsg
	content    string
}

// NewViewModel - метод для создания окна просмотра секретов
func NewViewModel() ViewModel {
	return ViewModel{
		content: `Секретные данные`,
	}
}

// Init - метод инициализации окна
func (m ViewModel) Init() tea.Cmd {
	return nil
}

// Update - метод для обновления окна по внешним сообщениям
func (m ViewModel) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil
	}
	return m, nil
}

// View - метод для отрисовки окна, в зависимости от текущего состояния
func (m ViewModel) View(username string) string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(50).
			Render("👁️ Просмотр контента"),

		lipgloss.NewStyle().Height(1).Render(""),

		// Отображаем имя пользователя
		lipgloss.NewStyle().
			Foreground(styles.SuccessColor).
			Bold(true).
			Render("👤 Пользователь: "+username),

		lipgloss.NewStyle().Height(1).Render(""),

		styles.SuccessStyle.
			Render("✅ Вы успешно авторизованы!"),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.ContentStyle.
			Width(m.windowSize.Width-10).
			Height(m.windowSize.Height-16). // Увеличиваем отступ для нового элемента
			Render(m.content),

		lipgloss.NewStyle().Height(2).Render(""),

		styles.HelpStyle.
			Render("Нажмите ESC для возврата в меню"),
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
