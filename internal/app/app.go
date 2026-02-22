package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/milad/vaultui/internal/ui/styles"
	"github.com/milad/vaultui/internal/vault"
)

type healthMsg struct {
	status *vault.HealthStatus
	err    error
}

// Model is the top-level Bubble Tea model for the application.
type Model struct {
	client    *vault.Client
	health    *vault.HealthStatus
	healthErr error
	width     int
	height    int
	ready     bool
	quitting  bool
}

// New creates the initial application model with the given Vault client.
func New(client *vault.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchHealth
}

func (m Model) fetchHealth() tea.Msg {
	status, err := m.client.Health()
	return healthMsg{status: status, err: err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case healthMsg:
		m.health = msg.status
		m.healthErr = msg.err
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

	addrPart := styles.HeaderLabelStyle.Render(" ◆ ") + styles.HeaderValueStyle.Render(addr)
	nsPart := styles.HeaderLabelStyle.Render("  ns: ") + styles.HeaderValueStyle.Render(ns)

	var statusPart string
	switch {
	case m.healthErr != nil:
		statusPart = styles.HeaderLabelStyle.Render("  ◆  ") +
			styles.ErrorStyle.Render("disconnected")
	case m.health != nil:
		statusPart = styles.HeaderLabelStyle.Render("  ◆  ") + m.renderSealStatus()
	default:
		statusPart = styles.HeaderLabelStyle.Render("  ◆  ") +
			styles.SubtleStyle.Render("connecting...")
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, addrPart, nsPart, statusPart)

	var right string
	if m.health != nil {
		right = m.renderHealthInfo()
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	spacer := lipgloss.NewStyle().Width(gap).Render("")

	headerRow := lipgloss.JoinHorizontal(lipgloss.Center, left, spacer, right)
	return styles.HeaderStyle.Width(m.width).Render(headerRow)
}

func (m Model) renderSealStatus() string {
	if m.health.Sealed {
		return styles.ErrorStyle.Render("sealed")
	}
	return styles.SuccessStyle.Render("unsealed")
}

func (m Model) renderHealthInfo() string {
	h := m.health
	version := styles.HeaderLabelStyle.Render("v") + styles.HeaderValueStyle.Render(h.Version)

	var haMode string
	if h.Standby {
		haMode = styles.HeaderValueStyle.Render("standby")
	} else if h.ClusterID != "" {
		haMode = styles.HeaderValueStyle.Render("active")
	}

	parts := version
	if haMode != "" {
		parts += styles.HeaderLabelStyle.Render("  ha: ") + haMode
	}
	if h.ClusterName != "" {
		parts += styles.HeaderLabelStyle.Render("  cluster: ") + styles.HeaderValueStyle.Render(h.ClusterName)
	}

	return parts
}

func (m Model) renderBody() string {
	bodyHeight := m.height - 4

	var msg string
	if m.healthErr != nil {
		msg = styles.ErrorStyle.Render("Could not connect to Vault") + "\n\n" +
			styles.SubtleStyle.Render(fmt.Sprintf("%v", m.healthErr)) + "\n\n" +
			styles.SubtleStyle.Render("Check VAULT_ADDR and VAULT_TOKEN, then press q to quit")
	} else {
		msg = styles.SubtleStyle.Render("Welcome to VaultUI\n\nPress : for commands, ? for help, q to quit")
	}

	return lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, msg)
}

func (m Model) renderStatusBar() string {
	hints := styles.HintKeyStyle.Render(":") + styles.HintDescStyle.Render(" command  ") +
		styles.HintKeyStyle.Render("/") + styles.HintDescStyle.Render(" filter  ") +
		styles.HintKeyStyle.Render("?") + styles.HintDescStyle.Render(" help  ") +
		styles.HintKeyStyle.Render("q") + styles.HintDescStyle.Render(" quit")

	return styles.StatusBarStyle.Width(m.width).Render(hints)
}
