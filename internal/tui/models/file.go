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
	isViewMode    bool                  // Флаг режима просмотра
	secretData    messages.SecretBinary // Данные для просмотра
}

func NewFileSecretModel() FileSecretModel {
	model := FileSecretModel{
		isViewMode: false,
	}

	model.filePathInput = textinput.New()
	model.filePathInput.Placeholder = "Введите путь к файлу"
	model.filePathInput.CharLimit = 255
	model.filePathInput.TextStyle = styles.FocusedStyle
	model.filePathInput.PromptStyle = styles.FocusedStyle

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

	case messages.GetSecretBinaryMsg:
		// Переключаемся в режим просмотра при получении данных
		m.isViewMode = true
		m.secretData = msg.Data

		// Заполняем поле данными для просмотра
		m.filePathInput.SetValue(msg.Data.Name)
		return m, nil

	case tea.KeyMsg:
		// В режиме просмотра обрабатываем Enter (сохранение) и ESC
		if m.isViewMode {
			switch msg.String() {
			case "enter":
				return m, m.attemptSaveFile(m.secretData)
			case "esc":
				m.isViewMode = false
				return m, func() tea.Msg {
					return messages.SecretAddCancelMsg{}
				}
			}
			return m, nil
		}

		// Режим редактирования
		switch msg.String() {
		case "enter":
			return m, m.attemptAddSecret(m.filePathInput.Value())

		case "esc":
			return m, func() tea.Msg {
				return messages.SecretAddCancelMsg{}
			}
		}
	}

	// В режиме просмотра игнорируем ввод данных
	if m.isViewMode {
		return m, nil
	}

	var cmd tea.Cmd
	m.filePathInput, cmd = m.filePathInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m FileSecretModel) View() string {
	fileInfo := ""

	if m.isViewMode {
		// В режиме просмотра показываем информацию о файле
		fileInfo = "Файл: " + m.secretData.Name
		if m.secretData.Blob != nil {
			fileInfo += " | Готов к сохранению"
		}
	} else if m.filePathInput.Value() != "" {
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
		styles.ButtonStyle.Render("Enter - Сохранить"),
		styles.DividerStyle.Render(),
		styles.ButtonStyle.Render("ESC - Отмена"),
	)
	hint := "Введите полный путь к файлу для сохранения"

	// В режиме просмотра меняем заголовок, кнопки и подсказку
	if m.isViewMode {
		title = "👁️ Просмотр файла"
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ButtonStyle.Render("Enter - Сохранить на диск"),
			styles.DividerStyle.Render(),
			styles.ButtonStyle.Render("ESC - Закрыть"),
		)
		hint = "Нажмите Enter чтобы сохранить файл на диск"
	}

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

func (m FileSecretModel) renderInputField(label string, input textinput.Model) string {
	var inputStyle lipgloss.Style
	if m.isViewMode {
		inputStyle = styles.InputFieldStyle
	} else {
		inputStyle = styles.FocusedInputFieldStyle
	}

	var fieldView string
	if m.isViewMode {
		// В режиме просмотра показываем только значение без курсора
		fieldView = input.Value()
		if fieldView == "" {
			fieldView = " "
		}
	} else {
		fieldView = input.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabelStyle.Render(label),
		inputStyle.Width(60).Render(fieldView),
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

// attemptSaveFile - метод сохранения файла на диск в режиме просмотра
func (m FileSecretModel) attemptSaveFile(secret messages.SecretBinary) tea.Cmd {
	return func() tea.Msg {
		if secret.Blob == nil {
			return messages.ErrorMsg("Нет данных для сохранения")
		}

		// Запрашиваем путь для сохранения
		// В реальной реализации здесь можно добавить диалог выбора пути
		// Сейчас сохраняем в текущую директорию с оригинальным именем
		filename := secret.Name

		// Проверяем, не существует ли файл
		if _, err := os.Stat(filename); err == nil {
			return messages.ErrorMsg("Файл уже существует: " + filename)
		}

		// Сохраняем файл
		err := os.WriteFile(filename, secret.Blob, 0644)
		if err != nil {
			return messages.ErrorMsg("Ошибка сохранения файла: " + err.Error())
		}

		return nil
	}
}
