package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestAWSView_Title(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	if v.Title() != "AWS: aws/" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestAWSView_Init(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestAWSView_View_Loading(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading AWS data") {
		t.Error("expected loading message")
	}
}

func TestAWSView_Update_Loaded(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	msg := awsLoadedMsg{
		roles: []vault.AWSRole{
			{Name: "deploy", CredentialType: "iam_user", PolicyARNs: []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"}},
		},
		config: &vault.AWSConfig{
			AccessKey: "AKIA...",
			Region:    "us-east-1",
		},
		lease: &vault.AWSLeaseConfig{
			Lease:    "30m0s",
			LeaseMax: "1h0m0s",
		},
	}

	updated, _ := v.Update(msg)
	av := updated.(*AWSView)

	if av.loading {
		t.Error("expected loading to be false")
	}
	if len(av.roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(av.roles))
	}
	if av.config == nil {
		t.Error("expected config to be set")
	}
}

func TestAWSView_Update_Error(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	updated, _ := v.Update(awsLoadedMsg{err: errTest})
	av := updated.(*AWSView)

	if av.loading {
		t.Error("expected loading to be false")
	}
	if av.err == nil {
		t.Error("expected error to be set")
	}
}

func TestAWSView_View_Error(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	v.Update(awsLoadedMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestAWSView_View_Loaded(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	v.Update(awsLoadedMsg{ //nolint:errcheck // test setup
		roles:  []vault.AWSRole{{Name: "deploy", CredentialType: "iam_user"}},
		config: &vault.AWSConfig{Region: "us-east-1"},
	})
	view := v.View(100, 24)
	if !strings.Contains(view, "AWS: aws/") {
		t.Error("expected title in view output")
	}
}

func TestAWSView_TabSwitch(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	v.Update(awsLoadedMsg{ //nolint:errcheck // test setup
		roles:  []vault.AWSRole{{Name: "r1"}},
		config: &vault.AWSConfig{Region: "us-east-1"},
		leases: []vault.AWSLease{{LeaseID: "l1", TTL: time.Hour}},
	})

	if v.tab != 0 {
		t.Error("expected initial tab to be 0")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 1 {
		t.Errorf("expected tab 1, got %d", v.tab)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 2 {
		t.Errorf("expected tab 2, got %d", v.tab)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 0 {
		t.Errorf("expected tab 0 (wrap), got %d", v.tab)
	}
}

func TestAWSView_KeyHints(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
	found := false
	for _, h := range hints {
		if h.Key == "tab" {
			found = true
		}
	}
	if !found {
		t.Error("expected tab hint")
	}
}

func TestAWSView_Refresh(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	av := updated.(*AWSView)

	if !av.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func TestAWSView_EmptyTabs(t *testing.T) {
	v := NewAWSView(newTestClient(t), "aws/")
	v.Update(awsLoadedMsg{}) //nolint:errcheck // test setup

	view := v.View(80, 20)
	if !strings.Contains(view, "No roles found") {
		t.Error("expected empty roles message")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	view = v.View(80, 20)
	if !strings.Contains(view, "No config found") {
		t.Error("expected empty config message")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	view = v.View(80, 20)
	if !strings.Contains(view, "No active leases") {
		t.Error("expected empty leases message")
	}
}
