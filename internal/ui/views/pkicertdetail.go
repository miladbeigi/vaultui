package views

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/clipboard"
	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type pkiCertLoadedMsg struct {
	detail *vault.PKICertDetail
	err    error
}

type pkiCertStatusClearMsg struct{}

// PKICertDetailView shows the PEM content of a certificate.
type PKICertDetailView struct {
	client    *vault.Client
	mount     string
	serial    string
	detail    *vault.PKICertDetail
	err       error
	loading   bool
	scroll    int
	statusMsg string
}

var _ ui.View = (*PKICertDetailView)(nil)

func NewPKICertDetailView(client *vault.Client, mount, serial string) *PKICertDetailView {
	return &PKICertDetailView{
		client:  client,
		mount:   mount,
		serial:  serial,
		loading: true,
	}
}

func (v *PKICertDetailView) Init() tea.Cmd {
	return v.fetchCert
}

func (v *PKICertDetailView) fetchCert() tea.Msg {
	detail, err := v.client.ReadPKICert(v.mount, v.serial)
	return pkiCertLoadedMsg{detail: detail, err: err}
}

func (v *PKICertDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case pkiCertLoadedMsg:
		v.loading = false
		v.detail = msg.detail
		v.err = msg.err
		return v, nil

	case pkiCertStatusClearMsg:
		v.statusMsg = ""
		return v, nil

	case tea.KeyMsg:
		if v.loading {
			return v, nil
		}
		switch msg.String() {
		case "j", "down":
			v.scroll++
		case "k", "up":
			if v.scroll > 0 {
				v.scroll--
			}
		case "g", "home":
			v.scroll = 0
		case "G", "end":
			if v.detail != nil {
				lines := strings.Split(v.detail.Certificate, "\n")
				v.scroll = max(0, len(lines)-5)
			}
		case "c":
			if v.detail != nil {
				if err := clipboard.WriteWithAutoClear(v.detail.Certificate, 30*time.Second); err != nil {
					v.statusMsg = styles.ErrorStyle.Render("Copy failed: " + err.Error())
				} else {
					v.statusMsg = styles.SuccessStyle.Render("Copied certificate to clipboard")
				}
				return v, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
					return pkiCertStatusClearMsg{}
				})
			}
		}
	}

	return v, nil
}

const pkiCertTitleHeight = 2

func (v *PKICertDetailView) View(width, height int) string {
	titleLine := v.renderTitle()

	if v.loading {
		body := lipgloss.Place(width, height-pkiCertTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading certificate..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-pkiCertTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.detail == nil || v.detail.Certificate == "" {
		body := lipgloss.Place(width, height-pkiCertTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No certificate data"))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	bodyHeight := height - pkiCertTitleHeight
	lines := strings.Split(v.detail.Certificate, "\n")

	if v.scroll > len(lines)-bodyHeight {
		v.scroll = max(0, len(lines)-bodyHeight)
	}

	end := v.scroll + bodyHeight
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[v.scroll:end]

	certStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	rendered := make([]string, len(visible))
	for i, line := range visible {
		rendered[i] = certStyle.Render(line)
	}

	content := strings.Join(rendered, "\n")
	padded := lipgloss.NewStyle().Width(width).Height(bodyHeight).Render(content)

	if v.statusMsg != "" {
		padLines := strings.Split(padded, "\n")
		if len(padLines) > 0 {
			padLines[len(padLines)-1] = v.statusMsg
			padded = strings.Join(padLines, "\n")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, padded)
}

func (v *PKICertDetailView) renderTitle() string {
	label := styles.SubtleStyle.Render("Certificate: ")
	serial := styles.SecondaryStyle.Render(v.serial)
	return lipgloss.NewStyle().PaddingBottom(1).Render(label + serial)
}

func (v *PKICertDetailView) Title() string {
	return "Cert: " + v.serial
}

func (v *PKICertDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "scroll"},
		{Key: "c", Desc: "copy"},
		{Key: "esc", Desc: "back"},
	}
}
