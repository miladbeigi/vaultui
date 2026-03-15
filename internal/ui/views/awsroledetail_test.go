package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestAWSRoleDetailView_Title(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	if v.Title() != "AWS Role: deploy" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestAWSRoleDetailView_Init(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestAWSRoleDetailView_Update_Loaded(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	detail := &vault.AWSRoleDetail{
		Name:            "deploy",
		CredentialTypes: []string{"iam_user"},
		PolicyARNs:      []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
		DefaultSTSTTL:   "15m",
		MaxSTSTTL:       "1h",
	}
	updated, _ := v.Update(awsRoleDetailLoadedMsg{detail: detail})
	rv := updated.(*AWSRoleDetailView)

	if rv.loading {
		t.Error("expected loading to be false")
	}
	if rv.detail == nil {
		t.Fatal("expected detail to be set")
	}
	if rv.detail.Name != "deploy" {
		t.Errorf("unexpected name: %s", rv.detail.Name)
	}
}

func TestAWSRoleDetailView_Update_Error(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	updated, _ := v.Update(awsRoleDetailLoadedMsg{err: errTest})
	rv := updated.(*AWSRoleDetailView)

	if rv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestAWSRoleDetailView_View_Loading(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading role details") {
		t.Error("expected loading message")
	}
}

func TestAWSRoleDetailView_View_Error(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	v.Update(awsRoleDetailLoadedMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestAWSRoleDetailView_Refresh(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	rv := updated.(*AWSRoleDetailView)

	if !rv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func TestAWSRoleDetailView_KeyHints(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	hints := v.KeyHints()
	found := false
	for _, h := range hints {
		if h.Key == "r" {
			found = true
		}
	}
	if !found {
		t.Error("expected refresh hint")
	}
}

func TestAWSRoleDetailView_BuildRows(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	detail := &vault.AWSRoleDetail{
		Name:            "deploy",
		CredentialTypes: []string{"assumed_role"},
		RoleARNs:        []string{"arn:aws:iam::123456789012:role/DeployRole"},
		PolicyDocument:  `{"Version":"2012-10-17"}`,
		DefaultSTSTTL:   "30m",
		MaxSTSTTL:       "1h",
	}
	v.Update(awsRoleDetailLoadedMsg{detail: detail}) //nolint:errcheck // test setup
	view := v.View(120, 30)

	if !strings.Contains(view, "AWS Role: deploy") {
		t.Error("expected title in view")
	}
}
