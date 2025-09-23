package styles

import "github.com/charmbracelet/lipgloss"

// Цветовая палитра
var (
	PrimaryColor    = lipgloss.Color("#7D56F4") // Фиолетовый
	SecondaryColor  = lipgloss.Color("#F25D94") // Розовый
	AccentColor     = lipgloss.Color("#6EFAFB") // Бирюзовый
	BackgroundColor = lipgloss.Color("#1A1A2E") // Темно-синий
	SurfaceColor    = lipgloss.Color("#252945") // Поверхность
	TextPrimary     = lipgloss.Color("#FFFFFF") // Белый текст
	TextSecondary   = lipgloss.Color("#B0B0B0") // Серый текст
	SuccessColor    = lipgloss.Color("#4ECDC4") // Зеленый успеха
	ErrorColor      = lipgloss.Color("#FF6B6B") // Красный ошибки
	WarningColor    = lipgloss.Color("#FFD166") // Желтый предупреждения
	DisabledColor   = lipgloss.Color("#444444") // Серый для disabled
)

// Базовые стили
var (
	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Background(BackgroundColor)

	TitleStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(PrimaryColor).
			Padding(0, 2).
			Margin(1, 0, 2, 0).
			Bold(true).
			Align(lipgloss.Center)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Padding(0, 1).
			Margin(0, 0, 3, 0).
			Align(lipgloss.Center)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(SurfaceColor).
			Padding(0, 3).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SurfaceColor).
			Width(25).
			Height(3).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

	ActiveButtonStyle = ButtonStyle.
				Foreground(TextPrimary).
				Background(SecondaryColor).
				BorderForeground(AccentColor).
				Bold(true)

	DisabledButtonStyle = ButtonStyle.
				Foreground(TextSecondary).
				Background(DisabledColor).
				BorderForeground(DisabledColor)

	DisabledActiveButtonStyle = ButtonStyle.
					Foreground(TextSecondary).
					Background(DisabledColor).
					BorderForeground(AccentColor).
					Bold(true)

	FocusedStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Italic(true).
			MarginTop(2)

	SmallButtonStyle = lipgloss.NewStyle().
				Foreground(TextSecondary).
				Background(DisabledColor).
				Padding(0, 1).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(DisabledColor).
				Height(1).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)

	ActiveSmallButtonStyle = SmallButtonStyle.Copy().
				Foreground(TextPrimary).
				Background(SecondaryColor).
				BorderForeground(AccentColor).
				Bold(true)

	// Стили для полей ввода
	InputFieldStyle = lipgloss.NewStyle().
			Width(25).
			Height(2).
			Padding(0, 1).
			Background(SurfaceColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SurfaceColor).
			Foreground(TextPrimary)

	FocusedInputFieldStyle = InputFieldStyle.
				BorderForeground(AccentColor).
				Foreground(TextPrimary)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	// Стили для разделителей
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(BackgroundColor)).
			SetString("•").
			Padding(0, 2)

	TableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(AccentColor).
			Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(TextPrimary).
				Background(SurfaceColor).
				Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(SecondaryColor).
				Bold(true)
)
