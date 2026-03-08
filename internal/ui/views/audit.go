package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type auditDevicesMsg struct {
	devices []vault.AuditDevice
	err     error
}

type auditEntryMsg vault.AuditEntry
type auditStreamErrMsg struct{ err error }
type auditStreamConnectedMsg struct {
	ch <-chan vault.AuditEntry
}

// AuditView displays live Vault audit log entries and audit device configuration.
type AuditView struct {
	client   *vault.Client
	logPath  string
	tab      int // 0 = audit log, 1 = devices
	devices  []vault.AuditDevice
	devErr   error
	devTable *components.Table
	loading  bool

	entries   []vault.AuditEntry
	scroll    int
	paused    bool
	connected bool
	cancel    context.CancelFunc
	logCh     <-chan vault.AuditEntry
	logErr    error
}

const maxAuditEntries = 500

var _ ui.View = (*AuditView)(nil)

var auditDeviceColumns = []components.Column{
	{Title: "PATH", MinWidth: 16},
	{Title: "TYPE", MinWidth: 10},
	{Title: "DESCRIPTION", MinWidth: 30, FlexFill: true},
}

// NewAuditView creates the audit log / devices browser.
// logPath is the path to the audit log file on the local filesystem.
func NewAuditView(client *vault.Client, logPath string) *AuditView {
	return &AuditView{
		client:   client,
		logPath:  logPath,
		devTable: components.NewTable(auditDeviceColumns),
		loading:  true,
	}
}

func (v *AuditView) Init() tea.Cmd {
	return tea.Batch(v.fetchDevices, v.startTail())
}

func (v *AuditView) fetchDevices() tea.Msg {
	devices, err := v.client.ListAuditDevices()
	return auditDevicesMsg{devices: devices, err: err}
}

func (v *AuditView) startTail() tea.Cmd {
	if v.logPath == "" {
		return func() tea.Msg {
			return auditStreamErrMsg{err: fmt.Errorf("no audit log file configured (set --audit-log or configure a file audit device)")}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	logPath := v.logPath
	return func() tea.Msg {
		ch, err := vault.TailAuditLog(ctx, logPath)
		if err != nil {
			return auditStreamErrMsg{err: err}
		}
		return auditStreamConnectedMsg{ch: ch}
	}
}

func waitForAuditEntry(ch <-chan vault.AuditEntry) tea.Cmd {
	return func() tea.Msg {
		entry, ok := <-ch
		if !ok {
			return auditStreamErrMsg{err: fmt.Errorf("audit log stream closed")}
		}
		return auditEntryMsg(entry)
	}
}

func (v *AuditView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case auditDevicesMsg:
		v.loading = false
		v.devErr = msg.err
		v.devices = msg.devices
		v.rebuildDeviceTable()
		return v, nil

	case auditStreamConnectedMsg:
		v.logCh = msg.ch
		v.connected = true
		return v, waitForAuditEntry(msg.ch)

	case auditEntryMsg:
		v.appendEntry(vault.AuditEntry(msg))
		return v, waitForAuditEntry(v.logCh)

	case auditStreamErrMsg:
		v.logErr = msg.err
		return v, nil

	case tea.KeyMsg:
		switch {
		case msg.String() == "tab":
			v.tab = (v.tab + 1) % 2
			return v, nil
		case msg.String() == "p" && v.tab == 0:
			v.paused = !v.paused
			return v, nil
		case msg.String() == "c" && v.tab == 0:
			v.entries = nil
			v.scroll = 0
			return v, nil
		}

		if v.tab == 0 {
			return v, v.handleLogKeys(msg)
		}
		return v, v.handleDeviceKeys(msg)
	}

	return v, nil
}

func (v *AuditView) handleLogKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, navKeys.Up):
		if v.scroll > 0 {
			v.scroll--
		}
	case key.Matches(msg, navKeys.Down):
		v.scroll++
	case key.Matches(msg, navKeys.Top):
		v.scroll = 0
	case key.Matches(msg, navKeys.Bottom):
		v.scroll = len(v.entries)
	case key.Matches(msg, navKeys.PageUp):
		v.scroll -= 10
		if v.scroll < 0 {
			v.scroll = 0
		}
	case key.Matches(msg, navKeys.PageDown):
		v.scroll += 10
	}
	return nil
}

func (v *AuditView) handleDeviceKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, navKeys.Up):
		v.devTable.MoveUp()
	case key.Matches(msg, navKeys.Down):
		v.devTable.MoveDown()
	}
	return nil
}

func (v *AuditView) appendEntry(entry vault.AuditEntry) {
	if v.paused {
		return
	}
	v.entries = append(v.entries, entry)
	if len(v.entries) > maxAuditEntries {
		v.entries = v.entries[len(v.entries)-maxAuditEntries:]
	}
	v.scroll = len(v.entries)
}

func (v *AuditView) rebuildDeviceTable() {
	rows := make([]components.Row, len(v.devices))
	for i, d := range v.devices {
		rows[i] = components.Row{d.Path, d.Type, d.Description}
	}
	v.devTable.SetRows(rows)
}

const auditTitleHeight = 2

func (v *AuditView) View(width, height int) string {
	tabNames := []string{"Audit Log", "Devices"}
	tabs := ""
	for i, name := range tabNames {
		if i == v.tab {
			tabs += styles.SecondaryStyle.Render("["+name+"]") + "  "
		} else {
			tabs += styles.SubtleStyle.Render(" "+name+" ") + "  "
		}
	}
	title := lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(
		styles.ViewTitleStyle.Render("Audit & Logs") + "  " + tabs)

	bodyHeight := height - auditTitleHeight
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	var body string
	if v.tab == 0 {
		body = v.renderAuditLog(width, bodyHeight)
	} else {
		body = v.renderDevices(width, bodyHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}

func (v *AuditView) renderAuditLog(width, height int) string {
	if v.logErr != nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Audit log error: "+v.logErr.Error()))
	}

	if len(v.entries) == 0 {
		var status string
		if v.connected {
			status = styles.SuccessStyle.Render("● Connected") + "  " +
				styles.SubtleStyle.Render("Waiting for audit events...")
		} else {
			status = styles.SubtleStyle.Render("Connecting to audit log...")
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, status)
	}

	statusLine := v.renderLogStatus()
	logHeight := height - 1

	end := v.scroll
	if end > len(v.entries) {
		end = len(v.entries)
	}
	start := end - logHeight
	if start < 0 {
		start = 0
	}

	var lines []string
	for _, entry := range v.entries[start:end] {
		lines = append(lines, renderAuditEntry(entry, width))
	}

	for len(lines) < logHeight {
		lines = append(lines, "")
	}

	logContent := strings.Join(lines, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, logContent, statusLine)
}

func renderAuditEntry(entry vault.AuditEntry, width int) string {
	ts := styles.SubtleStyle.Render(entry.Time.Format("15:04:05"))

	typeStyle := styles.SubtleStyle
	switch entry.Type {
	case "request":
		typeStyle = lipgloss.NewStyle().Foreground(styles.SecondaryColor)
	case "response":
		typeStyle = styles.SuccessStyle
	}
	entryType := typeStyle.Render(fmt.Sprintf("%-4s", entry.Type[:3]))

	opStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	switch entry.Operation {
	case "create", "update", "delete":
		opStyle = lipgloss.NewStyle().Foreground(styles.SecondaryColor).Bold(true)
	}
	op := opStyle.Render(fmt.Sprintf("%-8s", entry.Operation))

	path := lipgloss.NewStyle().Foreground(styles.TextColor).Render(entry.Path)

	line := ts + " " + entryType + " " + op + " " + path

	if entry.Error != "" {
		line += " " + styles.ErrorStyle.Render(entry.Error)
	}

	if lipgloss.Width(line) > width {
		line = line[:width]
	}
	return line
}

func (v *AuditView) renderLogStatus() string {
	parts := []string{
		styles.SubtleStyle.Render(fmt.Sprintf(" %d entries", len(v.entries))),
	}

	if v.paused {
		parts = append(parts, styles.ErrorStyle.Render("PAUSED"))
	} else {
		parts = append(parts, styles.SuccessStyle.Render("LIVE"))
	}

	return strings.Join(parts, "  ")
}

func (v *AuditView) renderDevices(width, height int) string {
	v.devTable.SetSize(width, height)

	if v.loading {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading audit devices..."))
	}

	if v.devErr != nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.devErr.Error()))
	}

	if len(v.devices) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No audit devices configured"))
	}

	return v.devTable.View()
}

func (v *AuditView) Title() string {
	return "Audit & Logs"
}

func (v *AuditView) KeyHints() []ui.KeyHint {
	if v.tab == 0 {
		return []ui.KeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "p", Desc: "pause/resume"},
			{Key: "c", Desc: "clear"},
			{Key: "tab", Desc: "devices"},
			{Key: "esc", Desc: "back"},
		}
	}
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "tab", Desc: "audit log"},
		{Key: "esc", Desc: "back"},
	}
}

// Cleanup stops the audit log tail when leaving the view.
func (v *AuditView) Cleanup() {
	if v.cancel != nil {
		v.cancel()
	}
}
