package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// ErrorOverlayView displays an error message with troubleshooting hints.
type ErrorOverlayView struct {
	title string
	err   error
	hints []string
}

var _ ui.View = (*ErrorOverlayView)(nil)

// NewErrorOverlayView creates an error overlay with contextual troubleshooting hints.
func NewErrorOverlayView(title string, err error) *ErrorOverlayView {
	return &ErrorOverlayView{
		title: title,
		err:   err,
		hints: generateHints(err),
	}
}

func (v *ErrorOverlayView) Init() tea.Cmd { return nil }

func (v *ErrorOverlayView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	return v, nil
}

func (v *ErrorOverlayView) View(width, height int) string {
	boxWidth := min(width-4, 70)
	if boxWidth < 30 {
		boxWidth = width - 2
	}

	errorTitle := styles.ErrorStyle.Render("⚠  " + v.title)

	var parts []string
	parts = append(parts, errorTitle)
	parts = append(parts, "")

	if v.err != nil {
		errMsg := lipgloss.NewStyle().
			Foreground(styles.TextColor).
			Width(boxWidth - 4).
			Render(v.err.Error())
		parts = append(parts, errMsg)
		parts = append(parts, "")
	}

	if len(v.hints) > 0 {
		hintsTitle := styles.SecondaryStyle.Render("Troubleshooting:")
		parts = append(parts, hintsTitle)
		for _, hint := range v.hints {
			bullet := styles.SubtleStyle.Render("  • ") + styles.SubtleStyle.Render(hint)
			parts = append(parts, bullet)
		}
	}

	content := strings.Join(parts, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ErrorColor).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (v *ErrorOverlayView) Title() string {
	return v.title
}

func (v *ErrorOverlayView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func generateHints(err error) []string {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	var hints []string

	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "dial tcp") {
		hints = append(hints,
			"Check that Vault is running and accessible",
			"Verify VAULT_ADDR is set correctly",
			"Check firewall rules and network connectivity",
		)
	}

	if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "403") {
		hints = append(hints,
			"Your token may lack the required policy permissions",
			"Verify your token with: vault token lookup",
			"Check ACL policies attached to your token",
		)
	}

	if strings.Contains(errStr, "missing client token") || strings.Contains(errStr, "token") {
		hints = append(hints,
			"Set VAULT_TOKEN environment variable",
			"Or use --token flag to provide a token",
			"Or use --auth-method with userpass/approle",
			"Check ~/.vault-token file",
		)
	}

	if strings.Contains(errStr, "tls") || strings.Contains(errStr, "certificate") || strings.Contains(errStr, "x509") {
		hints = append(hints,
			"Vault may be using a self-signed certificate",
			"Set VAULT_SKIP_VERIFY=true to skip TLS verification",
			"Or set VAULT_CACERT to your CA certificate path",
		)
	}

	if strings.Contains(errStr, "sealed") {
		hints = append(hints,
			"Vault is in sealed state and cannot serve requests",
			"Unseal Vault using: vault operator unseal",
			"Contact your Vault administrator",
		)
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
		hints = append(hints,
			"Request timed out — Vault may be overloaded",
			"Check network latency to the Vault server",
			"Try again in a few moments",
		)
	}

	if len(hints) == 0 {
		hints = append(hints,
			"Check Vault server logs for more details",
			"Verify VAULT_ADDR and VAULT_TOKEN are correct",
			"Run: vault status (to check connectivity)",
		)
	}

	return hints
}
