package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/milad/vaultui/internal/vault"
)

func testHealth() *vault.HealthStatus {
	return &vault.HealthStatus{
		Initialized: true,
		Sealed:      false,
		Version:     "1.15.4",
		ClusterName: "test-cluster",
		ClusterID:   "abc-123",
	}
}

func TestDashboardView_Title(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	if v.Title() != "Dashboard" {
		t.Errorf("expected title 'Dashboard', got %q", v.Title())
	}
}

func TestDashboardView_Init_ReturnsCmd(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestDashboardView_View_Loading(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	view := v.View(80, 30)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestDashboardView_Update_Loaded(t *testing.T) {
	v := NewDashboardView(newTestClient(t))

	updated, cmd := v.Update(dashDataMsg{
		health:      testHealth(),
		seal:        &vault.SealInfo{SealType: "shamir", StorageType: "raft"},
		ha:          &vault.HAInfo{ActiveNodes: 1, StandbyNodes: 2},
		engineCount: 5,
		authCount:   3,
		policyCount: 10,
	})
	dv := updated.(*DashboardView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if dv.loading {
		t.Error("expected loading to be false")
	}
	if dv.health == nil {
		t.Error("expected health to be populated")
	}
	if dv.seal == nil || dv.seal.StorageType != "raft" {
		t.Error("expected seal info with storage type raft")
	}
	if dv.ha == nil || dv.ha.ActiveNodes != 1 || dv.ha.StandbyNodes != 2 {
		t.Error("expected HA info with 1 active, 2 standby")
	}
	if dv.engineCount != 5 {
		t.Errorf("expected 5 engines, got %d", dv.engineCount)
	}
	if dv.authCount != 3 {
		t.Errorf("expected 3 auth methods, got %d", dv.authCount)
	}
	if dv.policyCount != 10 {
		t.Errorf("expected 10 policies, got %d", dv.policyCount)
	}
}

func TestDashboardView_Update_HealthError(t *testing.T) {
	v := NewDashboardView(newTestClient(t))

	updated, _ := v.Update(dashDataMsg{healthErr: fmt.Errorf("connection refused")})
	dv := updated.(*DashboardView)

	if dv.healthErr == nil {
		t.Error("expected health error to be stored")
	}
}

func TestDashboardView_View_WithData(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	v.loading = false
	v.health = testHealth()
	v.seal = &vault.SealInfo{SealType: "shamir", StorageType: "raft"}
	v.ha = &vault.HAInfo{ActiveNodes: 1, StandbyNodes: 2}
	v.engineCount = 5
	v.authCount = 3
	v.policyCount = 10

	view := v.View(100, 30)

	checks := map[string]string{
		"Dashboard":    "title",
		"unsealed":     "unsealed status",
		"1.15.4":       "version",
		"test-cluster": "cluster name",
		"shamir":       "seal type",
		"raft":         "storage type",
		"1 active":     "active nodes",
		"2 standby":    "standby nodes",
		"5":            "engine count",
		"3":            "auth count",
		"10":           "policy count",
		"[1]":          "quick nav",
	}
	for substr, desc := range checks {
		if !strings.Contains(view, substr) {
			t.Errorf("expected view to show %s (%q)", desc, substr)
		}
	}
}

func TestDashboardView_View_Sealed(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	v.loading = false
	v.health = &vault.HealthStatus{Sealed: true, Version: "1.15.4"}

	view := v.View(100, 30)
	if !strings.Contains(view, "sealed") {
		t.Error("expected view to show sealed status")
	}
}

func TestDashboardView_View_HealthError(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	v.loading = false
	v.healthErr = fmt.Errorf("connection refused")

	view := v.View(100, 30)
	if !strings.Contains(view, "Could not connect") {
		t.Error("expected connection error message")
	}
}

func TestDashboardView_KeyHints(t *testing.T) {
	v := NewDashboardView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
