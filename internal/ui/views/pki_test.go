package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestPKIView_Title(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	if v.Title() != "PKI: pki/" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestPKIView_Init(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestPKIView_View_Loading(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestPKIView_Update_Loaded(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	certs := []vault.PKICert{{SerialNumber: "aa:bb:cc"}}
	roles := []vault.PKIRole{{Name: "test-role"}}

	updated, _ := v.Update(pkiLoadedMsg{certs: certs, roles: roles})
	pv := updated.(*PKIView)

	if pv.loading {
		t.Error("expected loading to be false")
	}
	if len(pv.certs) != 1 {
		t.Errorf("expected 1 cert, got %d", len(pv.certs))
	}
	if len(pv.roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(pv.roles))
	}
}

func TestPKIView_TabSwitch(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	v.loading = false
	v.certs = []vault.PKICert{{SerialNumber: "aa:bb"}}
	v.roles = []vault.PKIRole{{Name: "role1"}}
	v.rebuildTable()

	if v.tab != 0 {
		t.Error("expected initial tab to be 0 (certs)")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}})
	// Tab key is "tab" string
}

func TestPKIView_KeyHints(t *testing.T) {
	v := NewPKIView(newTestClient(t), "pki/")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
}
