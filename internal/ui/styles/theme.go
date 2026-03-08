package styles

import "github.com/charmbracelet/lipgloss"

// Color palette (Vault purple/gold + Dracula-inspired).
var (
	PrimaryColor   = lipgloss.Color("#7B61FF")
	SecondaryColor = lipgloss.Color("#FFB86C")
	AccentColor    = lipgloss.Color("#50FA7B")
	ErrorColor     = lipgloss.Color("#FF5555")
	SubtleColor    = lipgloss.Color("#6272A4")
	TextColor      = lipgloss.Color("#F8F8F2")
	DimTextColor   = lipgloss.Color("#6272A4")
	BgColor        = lipgloss.Color("#282A36")
	HeaderBgColor  = lipgloss.Color("#1E1F29")
	StatusBgColor  = lipgloss.Color("#1E1F29")
)

// Styled components.
var (
	HeaderStyle = lipgloss.NewStyle().
			Background(HeaderBgColor).
			Foreground(TextColor).
			Padding(1, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(SubtleColor)

	TitleStyle = lipgloss.NewStyle().
			Background(PrimaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	HeaderLabelStyle = lipgloss.NewStyle().Foreground(DimTextColor)
	HeaderValueStyle = lipgloss.NewStyle().Foreground(TextColor)
	SubtleStyle      = lipgloss.NewStyle().Foreground(DimTextColor)
	SecondaryStyle   = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Background(StatusBgColor).
			Foreground(TextColor).
			Padding(1, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(SubtleColor)

	HintKeyStyle  = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)
	HintDescStyle = lipgloss.NewStyle().Foreground(DimTextColor)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Background(PrimaryColor).
				Bold(true)

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Underline(true)

	ErrorStyle      = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
	SuccessStyle    = lipgloss.NewStyle().Foreground(AccentColor)
	SecretMaskStyle = lipgloss.NewStyle().Foreground(SubtleColor)

	ViewTitleStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Bold(true).
			PaddingBottom(1)

	BreadcrumbStyle       = lipgloss.NewStyle().Foreground(SecondaryColor)
	BreadcrumbActiveStyle = lipgloss.NewStyle().Foreground(TextColor).Bold(true)
)
