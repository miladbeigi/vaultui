package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/config"
	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/ui/views"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type healthMsg struct {
	status *vault.HealthStatus
	err    error
}

// Model is the top-level Bubble Tea model for the application.
type Model struct {
	client    *vault.Client
	cfg       *config.Config
	cfgPath   string
	router    *Router
	health    *vault.HealthStatus
	healthErr error
	renewer   *vault.TokenRenewer
	width     int
	height    int
	ready     bool
	quitting  bool
	initCmd   tea.Cmd
	cmdActive bool
	cmdInput  string
	cmdError  string
}

// New creates the initial application model with the given Vault client.
func New(client *vault.Client, cfg *config.Config, cfgPath string) Model {
	router := NewRouter()
	dashView := views.NewDashboardView(client)
	router.Push(dashView)

	if cfg == nil {
		cfg = &config.Config{}
	}

	return Model{
		client:  client,
		cfg:     cfg,
		cfgPath: cfgPath,
		router:  router,
		initCmd: dashView.Init(),
	}
}

func (m Model) Init() tea.Cmd {
	m.renewer = vault.StartTokenRenewer(m.client)
	return tea.Batch(m.fetchHealth, m.initCmd)
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

	case views.SwitchContextMsg:
		return m.switchContext(msg.Context)

	case tea.KeyMsg:
		if m.cmdActive {
			return m.updateCommandInput(msg)
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.stopRenewer()
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.ForceQuit):
			m.stopRenewer()
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			if m.router.Pop() {
				return m, nil
			}
		case key.Matches(msg, keys.Command):
			m.cmdActive = true
			m.cmdInput = ""
			m.cmdError = ""
			return m, nil
		case key.Matches(msg, keys.Jump1):
			cmd := m.router.ResetToRoot(views.NewEnginesView(m.client))
			return m, cmd
		case key.Matches(msg, keys.Jump2):
			cmd := m.router.ResetToRoot(views.NewAuthMethodsView(m.client))
			return m, cmd
		case key.Matches(msg, keys.Jump3):
			cmd := m.router.ResetToRoot(views.NewPoliciesView(m.client))
			return m, cmd
		case key.Matches(msg, keys.Jump4):
			cmd := m.router.ResetToRoot(views.NewIdentityView(m.client))
			return m, cmd
		case key.Matches(msg, keys.Jump5):
			cmd := m.router.ResetToRoot(views.NewPKIView(m.client, "pki/"))
			return m, cmd
		case key.Matches(msg, keys.Jump6):
			cmd := m.router.ResetToRoot(views.NewTransitView(m.client, "transit/"))
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

func (m Model) updateCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.executeCommand()
	case tea.KeyEsc:
		m.cmdActive = false
		m.cmdInput = ""
		m.cmdError = ""
		return m, nil
	case tea.KeyBackspace:
		if m.cmdInput != "" {
			m.cmdInput = m.cmdInput[:len(m.cmdInput)-1]
		}
		m.cmdError = ""
		return m, nil
	case tea.KeyRunes:
		m.cmdInput += string(msg.Runes)
		m.cmdError = ""
		return m, nil
	}
	return m, nil
}

func (m Model) executeCommand() (tea.Model, tea.Cmd) {
	cmd := m.cmdInput
	m.cmdActive = false
	m.cmdInput = ""
	m.cmdError = ""

	switch cmd {
	case "secrets":
		c := m.router.ResetToRoot(views.NewEnginesView(m.client))
		return m, c
	case "auth":
		c := m.router.ResetToRoot(views.NewAuthMethodsView(m.client))
		return m, c
	case "policies":
		c := m.router.ResetToRoot(views.NewPoliciesView(m.client))
		return m, c
	case "dash", "dashboard":
		c := m.router.ResetToRoot(views.NewDashboardView(m.client))
		return m, c
	case "identity":
		c := m.router.ResetToRoot(views.NewIdentityView(m.client))
		return m, c
	case "pki":
		c := m.router.ResetToRoot(views.NewPKIView(m.client, "pki/"))
		return m, c
	case "transit":
		c := m.router.ResetToRoot(views.NewTransitView(m.client, "transit/"))
		return m, c
	case "audit", "logs":
		c := m.router.ResetToRoot(views.NewAuditView(m.client))
		return m, c
	case "ctx", "contexts":
		c := m.router.ResetToRoot(views.NewContextsView(m.cfg))
		return m, c
	case "q", "quit":
		m.quitting = true
		return m, tea.Quit
	default:
		m.cmdActive = true
		m.cmdInput = cmd
		m.cmdError = fmt.Sprintf("unknown command: %s", cmd)
		return m, nil
	}
}

func (m *Model) stopRenewer() {
	if m.renewer != nil {
		m.renewer.Stop()
		m.renewer = nil
	}
}

func (m Model) switchContext(ctx config.Context) (tea.Model, tea.Cmd) {
	newClient, err := vault.NewClient(vault.ClientConfig{
		Address:   ctx.Address,
		Token:     ctx.Token,
		Namespace: ctx.Namespace,
	})
	if err != nil {
		m.healthErr = err
		return m, nil
	}

	if ctx.Auth.Method != "" && ctx.Auth.Method != "token" {
		err := newClient.Authenticate(vault.AuthConfig{
			Method:    vault.AuthMethod(ctx.Auth.Method),
			MountPath: ctx.Auth.MountPath,
			Username:  ctx.Auth.Username,
			Password:  ctx.Auth.Password,
			RoleID:    ctx.Auth.RoleID,
			SecretID:  ctx.Auth.SecretID,
		})
		if err != nil {
			m.healthErr = err
			return m, nil
		}
	}

	m.stopRenewer()
	m.client = newClient
	m.cfg.CurrentContext = ctx.Name
	m.health = nil
	m.healthErr = nil

	_ = config.Save(m.cfgPath, m.cfg)

	m.renewer = vault.StartTokenRenewer(m.client)

	router := NewRouter()
	dashView := views.NewDashboardView(m.client)
	router.Push(dashView)
	m.router = router

	return m, tea.Batch(m.fetchHealth, dashView.Init())
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

func (m Model) isCompact() bool {
	return m.width < 80
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.quitting {
		return ""
	}

	if m.width < 40 || m.height < 10 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Terminal too small\nResize to at least 40x10"))
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

	if m.isCompact() {
		left := lipgloss.JoinHorizontal(lipgloss.Center, title, statusPart)
		return styles.HeaderStyle.Width(m.width).Render(left)
	}

	addrPart := styles.HeaderLabelStyle.Render(" ◆ ") + styles.HeaderValueStyle.Render(addr)
	nsPart := styles.HeaderLabelStyle.Render("  ns: ") + styles.HeaderValueStyle.Render(ns)
	left := lipgloss.JoinHorizontal(lipgloss.Center, title, addrPart, nsPart, statusPart)

	var right string
	if m.health != nil {
		right = m.renderHealthInfo()
	}

	innerWidth := m.width - 4
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

const cmdInputHeight = 3 // top border + content + bottom border

func (m Model) renderCommandInput() string {
	prompt := styles.SecondaryStyle.Render(": ")
	input := styles.HeaderValueStyle.Render(m.cmdInput)
	cursor := styles.SecondaryStyle.Render("█")

	line := prompt + input + cursor
	if m.cmdError != "" {
		line += "  " + styles.ErrorStyle.Render(m.cmdError)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Width(m.bodyWidth() - 2).
		Render(line)
}

func (m Model) renderBody() string {
	bw := m.bodyWidth()
	bh := m.bodyHeight()

	if m.healthErr != nil {
		overlay := views.NewErrorOverlayView("Could not connect to Vault", m.healthErr)
		return overlay.View(bw, bh)
	}

	viewHeight := bh
	var cmdLine string
	if m.cmdActive {
		cmdLine = m.renderCommandInput()
		viewHeight -= cmdInputHeight
	}

	var viewContent string
	if current := m.router.Current(); current != nil {
		viewContent = current.View(bw, viewHeight)
	}

	if m.cmdActive {
		return lipgloss.JoinVertical(lipgloss.Left, cmdLine, viewContent)
	}
	return viewContent
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
