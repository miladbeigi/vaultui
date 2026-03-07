package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type dashDataMsg struct {
	health      *vault.HealthStatus
	healthErr   error
	seal        *vault.SealInfo
	ha          *vault.HAInfo
	engineCount int
	authCount   int
	policyCount int
	countsErr   error
}

// DashboardView shows Vault health, resource counts, and quick navigation.
type DashboardView struct {
	client      *vault.Client
	health      *vault.HealthStatus
	healthErr   error
	seal        *vault.SealInfo
	ha          *vault.HAInfo
	engineCount int
	authCount   int
	policyCount int
	countsErr   error
	loading     bool
}

var _ ui.View = (*DashboardView)(nil)

// NewDashboardView creates a new dashboard.
func NewDashboardView(client *vault.Client) *DashboardView {
	return &DashboardView{
		client:  client,
		loading: true,
	}
}

func (v *DashboardView) Init() tea.Cmd {
	return v.fetchData
}

func (v *DashboardView) fetchData() tea.Msg {
	msg := dashDataMsg{}

	msg.health, msg.healthErr = v.client.Health()
	msg.seal, _ = v.client.SealStatus()
	msg.ha, _ = v.client.HAStatus()

	engines, err := v.client.ListSecretEngines()
	if err != nil {
		msg.countsErr = err
	} else {
		msg.engineCount = len(engines)
	}

	auths, err := v.client.ListAuthMethods()
	if err != nil && msg.countsErr == nil {
		msg.countsErr = err
	} else if err == nil {
		msg.authCount = len(auths)
	}

	policies, err := v.client.ListPolicies()
	if err != nil && msg.countsErr == nil {
		msg.countsErr = err
	} else if err == nil {
		msg.policyCount = len(policies)
	}

	return msg
}

func (v *DashboardView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	if msg, ok := msg.(dashDataMsg); ok {
		v.loading = false
		v.health = msg.health
		v.healthErr = msg.healthErr
		v.seal = msg.seal
		v.ha = msg.ha
		v.engineCount = msg.engineCount
		v.authCount = msg.authCount
		v.policyCount = msg.policyCount
		v.countsErr = msg.countsErr
		return v, nil
	}

	return v, nil
}

func (v *DashboardView) View(width, height int) string {
	if v.loading {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading dashboard..."))
	}

	if v.healthErr != nil {
		msg := styles.ErrorStyle.Render("Could not connect to Vault") + "\n\n" +
			styles.SubtleStyle.Render(v.healthErr.Error())
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
	}

	title := styles.ViewTitleStyle.Render("Dashboard")
	healthSection := v.renderHealth()
	countsSection := v.renderCounts()
	quickNav := v.renderQuickNav()

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		healthSection,
		"",
		countsSection,
		"",
		quickNav,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

const (
	dashLabelWidth = 16
	dashColWidth   = 30
	dashGap        = 4
)

func (v *DashboardView) renderHealth() string {
	h := v.health
	if h == nil {
		return ""
	}

	labelStyle := lipgloss.NewStyle().Foreground(styles.DimTextColor).Width(dashLabelWidth)
	valStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	colStyle := lipgloss.NewStyle().Width(dashColWidth)

	var sealVal string
	if h.Sealed {
		sealVal = styles.ErrorStyle.Render("sealed ✗")
	} else {
		sealVal = styles.SuccessStyle.Render("unsealed ✔")
	}

	sealType := styles.SubtleStyle.Render("n/a")
	storageType := styles.SubtleStyle.Render("n/a")
	if v.seal != nil {
		if v.seal.SealType != "" {
			sealType = valStyle.Render(v.seal.SealType)
		}
		if v.seal.StorageType != "" {
			storageType = valStyle.Render(v.seal.StorageType)
		}
	}

	haNodes := styles.SubtleStyle.Render("n/a")
	if v.ha != nil {
		haNodes = valStyle.Render(fmt.Sprintf("%d active, %d standby", v.ha.ActiveNodes, v.ha.StandbyNodes))
	}

	left := lipgloss.JoinVertical(lipgloss.Left,
		colStyle.Render(labelStyle.Render("Status")+sealVal),
		colStyle.Render(labelStyle.Render("Version")+valStyle.Render(h.Version)),
		colStyle.Render(labelStyle.Render("Seal Type")+sealType),
	)

	right := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("HA Nodes")+haNodes,
		labelStyle.Render("Cluster")+valStyle.Render(clusterDisplay(h.ClusterName)),
		labelStyle.Render("Storage")+storageType,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", dashGap), right)
}

func clusterDisplay(name string) string {
	if name == "" {
		return "n/a"
	}
	return name
}

func (v *DashboardView) renderCounts() string {
	labelStyle := lipgloss.NewStyle().Foreground(styles.DimTextColor).Width(dashLabelWidth)
	valStyle := lipgloss.NewStyle().Foreground(styles.TextColor).Bold(true)
	colStyle := lipgloss.NewStyle().Width(dashColWidth)

	if v.countsErr != nil {
		return labelStyle.Render("Resources") + styles.ErrorStyle.Render("error fetching counts")
	}

	left := lipgloss.JoinVertical(lipgloss.Left,
		colStyle.Render(labelStyle.Render("Secret Engines")+valStyle.Render(fmt.Sprintf("%d", v.engineCount))),
		colStyle.Render(labelStyle.Render("Policies")+valStyle.Render(fmt.Sprintf("%d", v.policyCount))),
	)

	right := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("Auth Methods")+valStyle.Render(fmt.Sprintf("%d", v.authCount)),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", dashGap), right)
}

func (v *DashboardView) renderQuickNav() string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SubtleColor).
		Padding(1, 2)

	keyStyle := lipgloss.NewStyle().Foreground(styles.SecondaryColor).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	gap := "   "

	row1 := keyStyle.Render("[1]") + descStyle.Render(" Secret Engines") + gap +
		keyStyle.Render("[2]") + descStyle.Render(" Auth Methods") + gap +
		keyStyle.Render("[3]") + descStyle.Render(" Policies")

	return border.Render(row1)
}

func (v *DashboardView) Title() string {
	return "Dashboard"
}

func (v *DashboardView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "1-3", Desc: "quick nav"},
		{Key: ":", Desc: "command"},
		{Key: "q", Desc: "quit"},
	}
}
