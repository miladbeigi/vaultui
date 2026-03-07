package styles

import "github.com/charmbracelet/lipgloss"

// Palette holds all the colors for a theme.
type Palette struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Error     lipgloss.Color
	Subtle    lipgloss.Color
	Text      lipgloss.Color
	DimText   lipgloss.Color
	Bg        lipgloss.Color
	HeaderBg  lipgloss.Color
	StatusBg  lipgloss.Color
	TitleFg   lipgloss.Color
}

// DarkPalette is the default dark theme (Vault purple/gold + Dracula-inspired).
var DarkPalette = Palette{
	Primary:   lipgloss.Color("#7B61FF"),
	Secondary: lipgloss.Color("#FFB86C"),
	Accent:    lipgloss.Color("#50FA7B"),
	Error:     lipgloss.Color("#FF5555"),
	Subtle:    lipgloss.Color("#6272A4"),
	Text:      lipgloss.Color("#F8F8F2"),
	DimText:   lipgloss.Color("#6272A4"),
	Bg:        lipgloss.Color("#282A36"),
	HeaderBg:  lipgloss.Color("#1E1F29"),
	StatusBg:  lipgloss.Color("#1E1F29"),
	TitleFg:   lipgloss.Color("#FFFFFF"),
}

// LightPalette is a light theme with good contrast.
var LightPalette = Palette{
	Primary:   lipgloss.Color("#5A45D6"),
	Secondary: lipgloss.Color("#B86800"),
	Accent:    lipgloss.Color("#1B8C3A"),
	Error:     lipgloss.Color("#CC2233"),
	Subtle:    lipgloss.Color("#8890A6"),
	Text:      lipgloss.Color("#1A1A2E"),
	DimText:   lipgloss.Color("#6B7394"),
	Bg:        lipgloss.Color("#F5F5FA"),
	HeaderBg:  lipgloss.Color("#E8E8F0"),
	StatusBg:  lipgloss.Color("#E8E8F0"),
	TitleFg:   lipgloss.Color("#FFFFFF"),
}

// Current color variables (used by all styles below).
var (
	PrimaryColor   = DarkPalette.Primary
	SecondaryColor = DarkPalette.Secondary
	AccentColor    = DarkPalette.Accent
	ErrorColor     = DarkPalette.Error
	SubtleColor    = DarkPalette.Subtle
	TextColor      = DarkPalette.Text
	DimTextColor   = DarkPalette.DimText
	BgColor        = DarkPalette.Bg
	HeaderBgColor  = DarkPalette.HeaderBg
	StatusBgColor  = DarkPalette.StatusBg
)

// Styled components — initialized with default dark theme.
var (
	HeaderStyle           lipgloss.Style
	TitleStyle            lipgloss.Style
	HeaderLabelStyle      lipgloss.Style
	HeaderValueStyle      lipgloss.Style
	SubtleStyle           lipgloss.Style
	SecondaryStyle        lipgloss.Style
	StatusBarStyle        lipgloss.Style
	HintKeyStyle          lipgloss.Style
	HintDescStyle         lipgloss.Style
	SelectedRowStyle      lipgloss.Style
	TableHeaderStyle      lipgloss.Style
	ErrorStyle            lipgloss.Style
	SuccessStyle          lipgloss.Style
	SecretMaskStyle       lipgloss.Style
	ViewTitleStyle        lipgloss.Style
	BreadcrumbStyle       lipgloss.Style
	BreadcrumbActiveStyle lipgloss.Style
)

func init() {
	ApplyTheme("dark")
}

// ApplyTheme switches the global styles to the named theme.
func ApplyTheme(name string) {
	p := DarkPalette
	if name == "light" {
		p = LightPalette
	}
	applyPalette(p)
}

func applyPalette(p Palette) {
	PrimaryColor = p.Primary
	SecondaryColor = p.Secondary
	AccentColor = p.Accent
	ErrorColor = p.Error
	SubtleColor = p.Subtle
	TextColor = p.Text
	DimTextColor = p.DimText
	BgColor = p.Bg
	HeaderBgColor = p.HeaderBg
	StatusBgColor = p.StatusBg

	HeaderStyle = lipgloss.NewStyle().
		Background(HeaderBgColor).
		Foreground(TextColor).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(SubtleColor)

	TitleStyle = lipgloss.NewStyle().
		Background(PrimaryColor).
		Foreground(p.TitleFg).
		Bold(true).
		Padding(0, 1)

	HeaderLabelStyle = lipgloss.NewStyle().Foreground(DimTextColor)
	HeaderValueStyle = lipgloss.NewStyle().Foreground(TextColor)
	SubtleStyle = lipgloss.NewStyle().Foreground(DimTextColor)
	SecondaryStyle = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
		Background(StatusBgColor).
		Foreground(TextColor).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(SubtleColor)

	HintKeyStyle = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)
	HintDescStyle = lipgloss.NewStyle().Foreground(DimTextColor)

	SelectedRowStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Background(PrimaryColor).
		Bold(true)

	TableHeaderStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Underline(true)

	ErrorStyle = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(AccentColor)
	SecretMaskStyle = lipgloss.NewStyle().Foreground(SubtleColor)

	ViewTitleStyle = lipgloss.NewStyle().
		Foreground(TextColor).
		Bold(true).
		PaddingBottom(1)

	BreadcrumbStyle = lipgloss.NewStyle().Foreground(SecondaryColor)
	BreadcrumbActiveStyle = lipgloss.NewStyle().Foreground(TextColor).Bold(true)
}

// CurrentThemeName returns the name of the active theme.
func CurrentThemeName() string {
	if PrimaryColor == LightPalette.Primary {
		return "light"
	}
	return "dark"
}
