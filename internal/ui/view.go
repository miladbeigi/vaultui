package ui

import tea "github.com/charmbracelet/bubbletea"

// View is the interface that all screen views must implement.
// Each view manages its own state and renders into the body area.
type View interface {
	// Init returns an initial command when the view is first pushed.
	Init() tea.Cmd

	// Update handles messages and returns the updated view and any command.
	Update(msg tea.Msg) (View, tea.Cmd)

	// View renders the view content for the body area.
	View(width, height int) string

	// Title returns the display name for this view (used in breadcrumbs, etc.).
	Title() string

	// KeyHints returns the contextual keybinding hints for the status bar.
	KeyHints() []KeyHint
}

// KeyHint pairs a key label with its description for the status bar.
type KeyHint struct {
	Key  string
	Desc string
}

// PushViewMsg is returned as a tea.Cmd by views that want to navigate
// to a new view. The app model handles this by pushing onto the router.
type PushViewMsg struct {
	View View
}
