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

type logEntryMsg vault.LogEntry
type logStreamErrMsg struct{ err error }

// AuditView displays live Vault server logs and audit device configuration.
type AuditView struct {
	client  *vault.Client
	tab     int // 0 = log stream, 1 = devices
	devices []vault.AuditDevice
	devErr  error

	entries  []vault.LogEntry
	scroll   int
	paused   bool
	cancel   context.CancelFunc
	logCh    <-chan vault.LogEntry
	logErr   error
	logLevel string

	devTable *components.Table
	loading  bool
}

const maxLogEntries = 500

var _ ui.View = (*AuditView)(nil)

var auditDeviceColumns = []components.Column{
	{Title: "PATH", MinWidth: 16},
	{Title: "TYPE", MinWidth: 10},
	{Title: "DESCRIPTION", MinWidth: 30, FlexFill: true},
}

// NewAuditView creates the audit log / devices browser.
func NewAuditView(client *vault.Client) *AuditView {
	return &AuditView{
		client:   client,
		devTable: components.NewTable(auditDeviceColumns),
		loading:  true,
		logLevel: "info",
	}
}

func (v *AuditView) Init() tea.Cmd {
	return tea.Batch(v.fetchDevices, v.startLogStream())
}

func (v *AuditView) fetchDevices() tea.Msg {
	devices, err := v.client.ListAuditDevices()
	return auditDevicesMsg{devices: devices, err: err}
}

func (v *AuditView) startLogStream() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	return func() tea.Msg {
		ch, err := v.client.MonitorLogs(ctx, v.logLevel)
		if err != nil {
			return logStreamErrMsg{err: err}
		}
		entry, ok := <-ch
		if !ok {
			return logStreamErrMsg{err: fmt.Errorf("log stream closed")}
		}
		return initLogStreamMsg{ch: ch, first: entry}
	}
}

type initLogStreamMsg struct {
	ch    <-chan vault.LogEntry
	first vault.LogEntry
}

func waitForLog(ch <-chan vault.LogEntry) tea.Cmd {
	return func() tea.Msg {
		entry, ok := <-ch
		if !ok {
			return logStreamErrMsg{err: fmt.Errorf("log stream closed")}
		}
		return logEntryMsg(entry)
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

	case initLogStreamMsg:
		v.logCh = msg.ch
		v.appendEntry(vault.LogEntry(msg.first))
		return v, waitForLog(msg.ch)

	case logEntryMsg:
		v.appendEntry(vault.LogEntry(msg))
		return v, waitForLog(v.logCh)

	case logStreamErrMsg:
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

func (v *AuditView) appendEntry(entry vault.LogEntry) {
	if v.paused {
		return
	}
	v.entries = append(v.entries, entry)
	if len(v.entries) > maxLogEntries {
		v.entries = v.entries[len(v.entries)-maxLogEntries:]
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
	tabNames := []string{"Log Stream", "Devices"}
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
		body = v.renderLogStream(width, bodyHeight)
	} else {
		body = v.renderDevices(width, bodyHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}

func (v *AuditView) renderLogStream(width, height int) string {
	if v.logErr != nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Log stream error: "+v.logErr.Error()))
	}

	if len(v.entries) == 0 {
		status := styles.SubtleStyle.Render("Waiting for log entries...")
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
		lines = append(lines, renderLogEntry(entry, width))
	}

	for len(lines) < logHeight {
		lines = append(lines, "")
	}

	logContent := strings.Join(lines, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, logContent, statusLine)
}

func renderLogEntry(entry vault.LogEntry, width int) string {
	levelStyle := styles.SubtleStyle
	switch strings.ToUpper(entry.Level) {
	case "ERROR":
		levelStyle = styles.ErrorStyle
	case "WARN":
		levelStyle = lipgloss.NewStyle().Foreground(styles.SecondaryColor)
	case "INFO":
		levelStyle = styles.SuccessStyle
	case "DEBUG":
		levelStyle = styles.SubtleStyle
	case "TRACE":
		levelStyle = lipgloss.NewStyle().Foreground(styles.DimTextColor)
	}

	var ts string
	if !entry.Timestamp.IsZero() {
		ts = styles.SubtleStyle.Render(entry.Timestamp.Format("15:04:05")) + " "
	}

	level := levelStyle.Render(fmt.Sprintf("%-5s", entry.Level))
	msg := lipgloss.NewStyle().Foreground(styles.TextColor).Render(entry.Message)

	line := ts + level + " " + msg
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

	parts = append(parts, styles.SubtleStyle.Render("level:"+v.logLevel))

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
		{Key: "tab", Desc: "log stream"},
		{Key: "esc", Desc: "back"},
	}
}

// Cleanup stops the log stream when leaving the view.
func (v *AuditView) Cleanup() {
	if v.cancel != nil {
		v.cancel()
	}
}
