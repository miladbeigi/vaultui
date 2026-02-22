package styles

import "github.com/charmbracelet/lipgloss"

// Color palette — inspired by Vault's purple/gold branding and k9s aesthetics.
var (
	PrimaryColor   = lipgloss.Color("#7B61FF") // Vault purple
	SecondaryColor = lipgloss.Color("#FFB86C") // Warm gold
	AccentColor    = lipgloss.Color("#50FA7B") // Green for success/unsealed
	ErrorColor     = lipgloss.Color("#FF5555") // Red for errors/sealed
	SubtleColor    = lipgloss.Color("#6272A4") // Muted blue-gray
	TextColor      = lipgloss.Color("#F8F8F2") // Light foreground
	DimTextColor   = lipgloss.Color("#6272A4") // Dimmed text
	BgColor        = lipgloss.Color("#282A36") // Dark background
	HeaderBgColor  = lipgloss.Color("#1E1F29") // Slightly darker
	StatusBgColor  = lipgloss.Color("#1E1F29") // Status bar background
)

// HeaderStyle is the top connection/context bar.
var HeaderStyle = lipgloss.NewStyle().
	Background(HeaderBgColor).
	Foreground(TextColor).
	Padding(0, 1).
	BorderStyle(lipgloss.NormalBorder()).
	BorderBottom(true).
	BorderForeground(SubtleColor)

// TitleStyle renders the app name badge.
var TitleStyle = lipgloss.NewStyle().
	Background(PrimaryColor).
	Foreground(lipgloss.Color("#FFFFFF")).
	Bold(true).
	Padding(0, 1)

// HeaderLabelStyle renders labels/separators in the header bar.
var HeaderLabelStyle = lipgloss.NewStyle().
	Foreground(DimTextColor)

// HeaderValueStyle renders values in the header bar.
var HeaderValueStyle = lipgloss.NewStyle().
	Foreground(TextColor)

// SubtleStyle is for secondary/dimmed text.
var SubtleStyle = lipgloss.NewStyle().
	Foreground(DimTextColor)

// StatusBarStyle is the bottom help/keybinding bar.
var StatusBarStyle = lipgloss.NewStyle().
	Background(StatusBgColor).
	Foreground(TextColor).
	Padding(0, 1).
	BorderStyle(lipgloss.NormalBorder()).
	BorderTop(true).
	BorderForeground(SubtleColor)

// HintKeyStyle renders keybinding keys in the status bar.
var HintKeyStyle = lipgloss.NewStyle().
	Foreground(SecondaryColor).
	Bold(true)

// HintDescStyle renders keybinding descriptions in the status bar.
var HintDescStyle = lipgloss.NewStyle().
	Foreground(DimTextColor)

// SelectedRowStyle highlights the currently focused table row.
var SelectedRowStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Background(PrimaryColor).
	Bold(true)

// TableHeaderStyle renders table column headers.
var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(PrimaryColor).
	Bold(true).
	Underline(true)

// ErrorStyle renders error messages.
var ErrorStyle = lipgloss.NewStyle().
	Foreground(ErrorColor).
	Bold(true)

// SuccessStyle renders success indicators.
var SuccessStyle = lipgloss.NewStyle().
	Foreground(AccentColor)

// SecretMaskStyle renders masked secret values.
var SecretMaskStyle = lipgloss.NewStyle().
	Foreground(SubtleColor)

// BreadcrumbStyle renders the path breadcrumb separator.
var BreadcrumbStyle = lipgloss.NewStyle().
	Foreground(SecondaryColor)

// BreadcrumbActiveStyle renders the current breadcrumb segment.
var BreadcrumbActiveStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Bold(true)
