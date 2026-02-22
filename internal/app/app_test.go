package app

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/milad/vaultui/internal/vault"
)

func newTestClient(t *testing.T) *vault.Client {
	t.Helper()
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	c, err := vault.NewClient(vault.ClientConfig{
		Address: "http://127.0.0.1:8200",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

func TestNew(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	if m.client != client {
		t.Error("expected model to hold the provided client")
	}
	if m.ready {
		t.Error("expected model to not be ready before WindowSizeMsg")
	}
	if m.quitting {
		t.Error("expected model to not be quitting initially")
	}
}

func TestInit_ReturnsCmd(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command for health check")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	if cmd != nil {
		t.Error("expected no command from WindowSizeMsg")
	}
	if !model.ready {
		t.Error("expected model to be ready after WindowSizeMsg")
	}
	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}
	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
}

func TestUpdate_HealthMsg_Success(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	status := &vault.HealthStatus{
		Initialized: true,
		Sealed:      false,
		Version:     "1.15.4",
		ClusterName: "test-cluster",
		ClusterID:   "abc-123",
	}

	updated, cmd := m.Update(healthMsg{status: status})
	model := updated.(Model)

	if cmd != nil {
		t.Error("expected no command from healthMsg")
	}
	if model.health != status {
		t.Error("expected health status to be stored")
	}
	if model.healthErr != nil {
		t.Error("expected no health error")
	}
}

func TestUpdate_HealthMsg_Error(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	err := errors.New("connection refused")
	updated, _ := m.Update(healthMsg{err: err})
	model := updated.(Model)

	if model.health != nil {
		t.Error("expected no health status on error")
	}
	if model.healthErr == nil {
		t.Error("expected health error to be stored")
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(Model)

	if !model.quitting {
		t.Error("expected model to be quitting after 'q'")
	}
	if cmd == nil {
		t.Error("expected a Quit command")
	}
}

func TestUpdate_ForceQuitKey(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	if !model.quitting {
		t.Error("expected model to be quitting after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected a Quit command")
	}
}

func TestView_NotReady(t *testing.T) {
	client := newTestClient(t)
	m := New(client)

	view := m.View()
	if view != "Initializing..." {
		t.Errorf("expected 'Initializing...', got %q", view)
	}
}

func TestView_Quitting(t *testing.T) {
	client := newTestClient(t)
	m := New(client)
	m.ready = true
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got %q", view)
	}
}

func TestView_Connected(t *testing.T) {
	client := newTestClient(t)
	m := New(client)
	m.ready = true
	m.width = 120
	m.height = 40
	m.health = &vault.HealthStatus{
		Initialized: true,
		Sealed:      false,
		Version:     "1.15.4",
		ClusterName: "test-cluster",
		ClusterID:   "abc-123",
	}

	view := m.View()
	if !strings.Contains(view, "127.0.0.1:8200") {
		t.Error("expected view to contain the vault address")
	}
	if !strings.Contains(view, "unsealed") {
		t.Error("expected view to show unsealed status")
	}
	if !strings.Contains(view, "1.15.4") {
		t.Error("expected view to show vault version")
	}
	if !strings.Contains(view, "Welcome to VaultUI") {
		t.Error("expected view to contain welcome message")
	}
}

func TestView_Sealed(t *testing.T) {
	client := newTestClient(t)
	m := New(client)
	m.ready = true
	m.width = 120
	m.height = 40
	m.health = &vault.HealthStatus{
		Initialized: true,
		Sealed:      true,
		Version:     "1.15.4",
	}

	view := m.View()
	if !strings.Contains(view, "sealed") {
		t.Error("expected view to show sealed status")
	}
}

func TestView_Disconnected(t *testing.T) {
	client := newTestClient(t)
	m := New(client)
	m.ready = true
	m.width = 120
	m.height = 40
	m.healthErr = errors.New("connection refused")

	view := m.View()
	if !strings.Contains(view, "disconnected") {
		t.Error("expected view to show disconnected status")
	}
	if !strings.Contains(view, "Could not connect to Vault") {
		t.Error("expected view to show connection error message")
	}
}

func TestView_Connecting(t *testing.T) {
	client := newTestClient(t)
	m := New(client)
	m.ready = true
	m.width = 120
	m.height = 40

	view := m.View()
	if !strings.Contains(view, "connecting...") {
		t.Error("expected view to show connecting status")
	}
}
