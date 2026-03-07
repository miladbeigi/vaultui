package views

import (
	"strings"
	"testing"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestTransitView_Title(t *testing.T) {
	v := NewTransitView(newTestClient(t), "transit/")
	if v.Title() != "Transit: transit/" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestTransitView_Init(t *testing.T) {
	v := NewTransitView(newTestClient(t), "transit/")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestTransitView_View_Loading(t *testing.T) {
	v := NewTransitView(newTestClient(t), "transit/")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestTransitView_Update_Loaded(t *testing.T) {
	v := NewTransitView(newTestClient(t), "transit/")
	keys := []vault.TransitKey{{Name: "my-key"}}

	updated, _ := v.Update(transitLoadedMsg{keys: keys})
	tv := updated.(*TransitView)

	if tv.loading {
		t.Error("expected loading to be false")
	}
	if len(tv.keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(tv.keys))
	}
}

func TestTransitView_KeyHints(t *testing.T) {
	v := NewTransitView(newTestClient(t), "transit/")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
}

func TestTransitKeyDetailView_Title(t *testing.T) {
	v := NewTransitKeyDetailView(newTestClient(t), "transit/", "my-key")
	if !strings.Contains(v.Title(), "my-key") {
		t.Errorf("expected title to contain key name, got %q", v.Title())
	}
}

func TestTransitKeyDetailView_Update_Loaded(t *testing.T) {
	v := NewTransitKeyDetailView(newTestClient(t), "transit/", "my-key")
	detail := &vault.TransitKeyDetail{
		Name:          "my-key",
		Type:          "aes256-gcm96",
		LatestVersion: 1,
	}

	updated, _ := v.Update(transitKeyLoadedMsg{detail: detail})
	dv := updated.(*TransitKeyDetailView)

	if dv.loading {
		t.Error("expected loading to be false")
	}
	if dv.detail.Type != "aes256-gcm96" {
		t.Errorf("expected type aes256-gcm96, got %s", dv.detail.Type)
	}
}
