package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/clipboard"
	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type policyLoadedMsg struct {
	body string
	err  error
}

type policyStatusClearMsg struct{}

// PolicyDetailView displays the HCL body of a single policy.
type PolicyDetailView struct {
	client    *vault.Client
	name      string
	body      string
	err       error
	loading   bool
	scroll    int
	statusMsg string
}

var _ ui.View = (*PolicyDetailView)(nil)

func NewPolicyDetailView(client *vault.Client, name string) *PolicyDetailView {
	return &PolicyDetailView{
		client:  client,
		name:    name,
		loading: true,
	}
}

func (v *PolicyDetailView) Init() tea.Cmd {
	return v.fetchPolicy
}

func (v *PolicyDetailView) fetchPolicy() tea.Msg {
	body, err := v.client.GetPolicy(v.name)
	return policyLoadedMsg{body: body, err: err}
}

func (v *PolicyDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case policyLoadedMsg:
		v.loading = false
		v.body = msg.body
		v.err = msg.err
		return v, nil

	case policyStatusClearMsg:
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
			v.scroll = v.maxScroll()
		case "c":
			cmd := v.copyBody()
			return v, cmd
		}
	}

	return v, nil
}

func (v *PolicyDetailView) View(width, height int) string {
	titleLine := v.renderTitle()

	if v.loading {
		body := lipgloss.Place(width, height-policyTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading policy..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-policyTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.body == "" {
		body := lipgloss.Place(width, height-policyTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Empty policy"))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	bodyHeight := height - policyTitleHeight
	lines := strings.Split(v.body, "\n")

	if v.scroll > len(lines)-bodyHeight {
		v.scroll = max(0, len(lines)-bodyHeight)
	}

	end := v.scroll + bodyHeight
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[v.scroll:end]

	hclStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	rendered := make([]string, len(visible))
	for i, line := range visible {
		rendered[i] = hclStyle.Render(line)
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

const policyTitleHeight = 2

func (v *PolicyDetailView) renderTitle() string {
	label := styles.SubtleStyle.Render("Policy: ")
	name := styles.SecondaryStyle.Render(v.name)
	return lipgloss.NewStyle().PaddingBottom(1).Render(label + name)
}

func (v *PolicyDetailView) Title() string {
	return "Policy: " + v.name
}

func (v *PolicyDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "scroll"},
		{Key: "c", Desc: "copy"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *PolicyDetailView) copyBody() tea.Cmd {
	if err := clipboard.WriteWithAutoClear(v.body, 30*time.Second); err != nil {
		v.statusMsg = styles.ErrorStyle.Render(fmt.Sprintf("Copy failed: %v", err))
	} else {
		v.statusMsg = styles.SuccessStyle.Render("Copied policy to clipboard")
	}
	return v.clearStatusAfter()
}

func (v *PolicyDetailView) clearStatusAfter() tea.Cmd {
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return policyStatusClearMsg{}
	})
}

func (v *PolicyDetailView) maxScroll() int {
	lines := strings.Split(v.body, "\n")
	return max(0, len(lines)-5)
}
