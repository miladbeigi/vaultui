package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/milad/vaultui/internal/ui"
	"github.com/milad/vaultui/internal/ui/styles"
	"github.com/milad/vaultui/internal/ui/views"
	"github.com/milad/vaultui/internal/vault"
)

type healthMsg struct {
	status *vault.HealthStatus
	err    error
}

// Model is the top-level Bubble Tea model for the application.
type Model struct {
	client    *vault.Client
	router    *Router
	health    *vault.HealthStatus
	healthErr error
	width     int
	height    int
	ready     bool
	quitting  bool
}

// New creates the initial application model with the given Vault client.
func New(client *vault.Client) Model {
	router := NewRouter()
	router.Push(views.NewHomeView())

	return Model{
		client: client,
		router: router,
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

	case ui.PushViewMsg:
		cmd := m.router.Push(msg.View)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.ForceQuit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			if m.router.Pop() {
				return m, nil
			}
		case key.Matches(msg, keys.Jump1):
			cmd := m.router.Push(views.NewEnginesView(m.client))
			return m, cmd
		}
	}

	if current := m.router.Current(); current != nil {
		updated, cmd := current.Update(msg)
		m.router.Replace(updated)
		return m, cmd
	}

	return m, nil
}

const headerHeight = 4    // 1 top pad + 1 content + 1 bottom pad + 1 border
const statusBarHeight = 4 // 1 border + 1 top pad + 1 content + 1 bottom pad
const bodyPaddingX = 2    // left + right padding (1 each side)

func (m Model) bodyHeight() int {
	h := m.height - headerHeight - statusBarHeight
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) bodyWidth() int {
	w := m.width - bodyPaddingX
	if w < 1 {
		return 1
	}
	return w
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.quitting {
		return ""
	}

	header := m.renderHeader()
	body := lipgloss.NewStyle().
		Width(m.width).
		Height(m.bodyHeight()).
		Padding(0, 1).
		Render(m.renderBody())
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

	innerWidth := m.width - 4 // account for header Padding(1, 2) = 2 left + 2 right
	gap := innerWidth - lipgloss.Width(left) - lipgloss.Width(right)
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
	bw := m.bodyWidth()
	bh := m.bodyHeight()

	if m.healthErr != nil {
		msg := styles.ErrorStyle.Render("Could not connect to Vault") + "\n\n" +
			styles.SubtleStyle.Render(fmt.Sprintf("%v", m.healthErr)) + "\n\n" +
			styles.SubtleStyle.Render("Check VAULT_ADDR and VAULT_TOKEN, then press q to quit")
		return lipgloss.Place(bw, bh, lipgloss.Center, lipgloss.Center, msg)
	}

	if current := m.router.Current(); current != nil {
		return current.View(bw, bh)
	}

	return ""
}

func (m Model) renderStatusBar() string {
	var hints string

	if current := m.router.Current(); current != nil {
		for i, h := range current.KeyHints() {
			if i > 0 {
				hints += "  "
			}
			hints += styles.HintKeyStyle.Render(h.Key) + styles.HintDescStyle.Render(" "+h.Desc)
		}
	}

	return styles.StatusBarStyle.Width(m.width).Render(hints)
}
