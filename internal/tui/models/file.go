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

// FileSecretModel - модель окна создания/просмотра секрета (файл)
type FileSecretModel struct {
	filePathInput textinput.Model
	windowSize    tea.WindowSizeMsg
	isEditMode    bool   // Флаг режима редактирования
	sid           string // id для редактирования
	secretData    []byte // Данные
}

// NewFileSecretModel - метод создания модель окна секрета (файл)
func NewFileSecretModel() FileSecretModel {
	model := FileSecretModel{
		isEditMode: false,
	}

	model.filePathInput = textinput.New()
	model.filePathInput.Placeholder = "Введите путь к файлу"
	model.filePathInput.CharLimit = 255
	model.filePathInput.TextStyle = styles.FocusedStyle
	model.filePathInput.PromptStyle = styles.FocusedStyle

	model.filePathInput.Focus()

	return model
}

// Init - метод инициализации текущего окна
func (m FileSecretModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update - метод обновления текущего окна
func (m FileSecretModel) Update(msg tea.Msg) (FileSecretModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		return m, nil

	case messages.GetSecretBinaryMsg:
		// Переключаемся в режим редактирования при получении данных
		m.isEditMode = true
		m.sid = msg.ID
		m.secretData = msg.Data.Blob
		// Заполняем поле данными для просмотра
		m.filePathInput.SetValue(msg.Data.Name)
		return m, nil

	case tea.KeyMsg:
		// Режим редактирования
		switch msg.String() {
		case "ctrl+s":
			return m, m.attemptSaveFile(m.filePathInput.Value(), m.secretData)

		case "enter":
			if m.isEditMode {
				m.isEditMode = false
				return m, m.attemptEditSecret(m.sid, m.filePathInput.Value())
			}
			return m, m.attemptAddSecret(m.filePathInput.Value())

		case "esc":
			m.isEditMode = false
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

// View - метод отрисовки текущего состояния
func (m FileSecretModel) View() string {
	fileInfo := ""

	if m.filePathInput.Value() != "" {
		// В режиме редактирования проверяем существование файла
		fileInfo = "Путь: " + m.filePathInput.Value()

		// Проверяем существование файла
		if _, err := os.Stat(m.filePathInput.Value()); err == nil {
			fileInfo += " ✓ Файл существует"
		} else {
			fileInfo += " ✗ Файл не найден"
		}
	}

	title := "📁 Укажите путь к файлу"
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.ButtonStyle.Render("Enter - Применить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("Ctrl+S - Сохранить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)
	hint := "Введите полный путь к файлу для сохранения"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.
			Width(40).
			Render(title),

		lipgloss.NewStyle().Height(2).Render(""),

		m.renderInputField("📁 Имя файла:", m.filePathInput),

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render(fileInfo),

		lipgloss.NewStyle().Height(2).Render(""),

		buttons,

		lipgloss.NewStyle().Height(1).Render(""),

		lipgloss.NewStyle().
			Foreground(styles.TextSecondary).
			Italic(true).
			Render(hint),
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

// renderInputField - метод для отрисовки полей ввода
func (m FileSecretModel) renderInputField(label string, input textinput.Model) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Render(label),
		styles.FocusedInputFieldStyle.Width(60).Render(input.View()),
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
			return messages.AddSecretBinaryMsg{
				Data: messages.SecretBinary{
					Name: filepath.Base(filename),
					Type: models.SecretBinaryType,
					Blob: content,
				},
			}
		}
		return messages.ErrorMsg("Файл не найден или недоступен")
	}
}

// attemptEditSecret - метод обработки изменения секрета
func (m FileSecretModel) attemptEditSecret(sid string, filename string) tea.Cmd {
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
			return messages.EditSecretBinaryMsg{
				ID: sid,
				Data: messages.SecretBinary{
					Name: filepath.Base(filename),
					Type: models.SecretBinaryType,
					Blob: content,
				},
			}
		}
		return messages.ErrorMsg("Файл не найден или недоступен")
	}
}

// attemptSaveFile - метод сохранения файла на диск в режиме просмотра
func (m FileSecretModel) attemptSaveFile(filename string, blob []byte) tea.Cmd {
	return func() tea.Msg {
		if blob == nil {
			return messages.ErrorMsg("Нет данных для сохранения")
		}

		// Проверяем, не существует ли файл
		if _, err := os.Stat(filename); err == nil {
			return messages.ErrorMsg("Файл уже существует: " + filename)
		}

		// Сохраняем файл
		err := os.WriteFile(filename, blob, 0644)
		if err != nil {
			return messages.ErrorMsg("Ошибка сохранения файла: " + err.Error())
		}

		return nil
	}
}
