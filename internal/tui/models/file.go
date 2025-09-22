package models

import (
	"go-pass-keeper/internal/models"
	"go-pass-keeper/internal/tui/messages"
	"go-pass-keeper/internal/tui/styles"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileSecretModel struct {
	filePathInput textinput.Model
	windowSize    tea.WindowSizeMsg
}

func NewFileSecretModel() FileSecretModel {
	model := FileSecretModel{}

	model.filePathInput = textinput.New()
	model.filePathInput.Placeholder = "Введите путь к файлу"
	model.filePathInput.CharLimit = 255
	model.filePathInput.TextStyle = styles.BlurredStyle

	model.filePathInput.Focus()

	return model
}

func (m FileSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FileSecretModel) Update(msg tea.Msg) (FileSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.attemptAddSecret(m.filePathInput.Value())

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	var cmd tea.Cmd
	m.filePathInput, cmd = m.filePathInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m FileSecretModel) View() string {
	fileInfo := ""
	if m.filePathInput.Value() != "" {
		fileInfo = "Путь: " + m.filePathInput.Value()

		// Проверяем существование файла
		if _, err := os.Stat(m.filePathInput.Value()); err == nil {
			fileInfo += " ✓ Файл существует"
		} else {
			fileInfo += " ✗ Файл не найден"
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render("📁 Укажите путь к файлу"),

		lipgloss.NewStyle().Height(2).Render(""),

		m.renderInputField("📁 Путь к файлу:", m.filePathInput),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render(fileInfo),

		lipgloss.NewStyle().Height(2).Render(""),

		lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Сохранить"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Отмена"),
		),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render("Введите полный путь к файлу для сохранения"),
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

func (m FileSecretModel) renderInputField(label string, input textinput.Model) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Render(label),
		styles.FocusedInputFieldStyle.Render(input.View()),
	) + "\n"
}

// attemptAddSecret - метод обработки добавления секрета
func (m FileSecretModel) attemptAddSecret(filename string) tea.Cmd {
	return func() tea.Msg {
		if filename == "" {
			return messages.ErrorMsg("Необходимо задать имя файла")
		}
		// Проверяем существование файла
		if _, err := os.Stat(m.filePathInput.Value()); err == nil {
			// Читаем содержимое файла
			content, err := os.ReadFile(m.filePathInput.Value())
			if err != nil {
				return messages.ErrorMsg("Ошибка чтения файла")
			}
			return messages.AddSecretBinaryMsg{Data: messages.SecretBinary{Name: filepath.Base(filename), Type: models.SecretBinaryType, Blob: content}}
		}
		return messages.ErrorMsg("Файл не найден или недоступен")
	}
}
