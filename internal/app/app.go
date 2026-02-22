package app

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/milad/vaultui/internal/ui/styles"
	"github.com/milad/vaultui/internal/vault"
)

// Model is the top-level Bubble Tea model for the application.
type Model struct {
	client   *vault.Client
	width    int
	height   int
	ready    bool
	quitting bool
}

// New creates the initial application model with the given Vault client.
func New(client *vault.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.ForceQuit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.quitting {
		return ""
	}

	header := m.renderHeader()
	body := m.renderBody()
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)
}

func (m Model) renderHeader() string {
	addr := m.client.Address()
	if addr == "" {
		addr = "not configured"
	}
	ns := m.client.Namespace()
	if ns == "" {
		ns = "root"
	}

	title := styles.TitleStyle.Render(" VaultUI ")
	conn := styles.SubtleStyle.Render(" ◆ " + addr + "  ◆  ns: " + ns)

	headerContent := lipgloss.JoinHorizontal(lipgloss.Center, title, conn)
	return styles.HeaderStyle.Width(m.width).Render(headerContent)
}

func (m Model) renderBody() string {
	bodyHeight := m.height - 4 // account for header + status bar

	content := lipgloss.Place(
		m.width, bodyHeight,
		lipgloss.Center, lipgloss.Center,
		styles.SubtleStyle.Render("Welcome to VaultUI\n\nPress : for commands, ? for help, q to quit"),
	)

	return content
}

func (m Model) renderStatusBar() string {
	hints := styles.HintKeyStyle.Render(":") + styles.HintDescStyle.Render(" command  ") +
		styles.HintKeyStyle.Render("/") + styles.HintDescStyle.Render(" filter  ") +
		styles.HintKeyStyle.Render("?") + styles.HintDescStyle.Render(" help  ") +
		styles.HintKeyStyle.Render("q") + styles.HintDescStyle.Render(" quit")

	return styles.StatusBarStyle.Width(m.width).Render(hints)
}
