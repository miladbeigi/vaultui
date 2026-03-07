package views

import (
	"strings"
	"testing"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestDiffView_Title(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "apps/myapp/config", 1, 2)
	if !strings.Contains(v.Title(), "v1") || !strings.Contains(v.Title(), "v2") {
		t.Errorf("expected title with version numbers, got %q", v.Title())
	}
}

func TestDiffView_Init_ReturnsCmd(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "apps/myapp/config", 1, 2)
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestDiffView_View_Loading(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "apps/myapp/config", 1, 2)
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestDiffView_Update_Loaded(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "apps/myapp/config", 1, 2)

	oldData := &vault.SecretData{
		Data: map[string]string{"key1": "val1", "key2": "old"},
		Keys: []string{"key1", "key2"},
	}
	newData := &vault.SecretData{
		Data: map[string]string{"key1": "val1", "key3": "new"},
		Keys: []string{"key1", "key3"},
	}

	updated, _ := v.Update(diffLoadedMsg{oldData: oldData, newData: newData})
	dv := updated.(*DiffView)

	if dv.loading {
		t.Error("expected loading to be false")
	}
	if len(dv.lines) != 3 {
		t.Errorf("expected 3 diff lines, got %d", len(dv.lines))
	}
}

func TestDiffView_Update_Error(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "test", 1, 2)

	updated, _ := v.Update(diffLoadedMsg{err: errTest})
	dv := updated.(*DiffView)

	if dv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestComputeDiff(t *testing.T) {
	oldData := &vault.SecretData{
		Data: map[string]string{"a": "1", "b": "2", "c": "3"},
		Keys: []string{"a", "b", "c"},
	}
	newData := &vault.SecretData{
		Data: map[string]string{"a": "1", "b": "changed", "d": "4"},
		Keys: []string{"a", "b", "d"},
	}

	lines := computeDiff(oldData, newData)

	expected := map[string]string{
		"a": "unchanged",
		"b": "changed",
		"c": "removed",
		"d": "added",
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(lines))
	}

	for _, line := range lines {
		want, ok := expected[line.key]
		if !ok {
			t.Errorf("unexpected key %q in diff", line.key)
			continue
		}
		if line.kind != want {
			t.Errorf("key %q: expected kind %q, got %q", line.key, want, line.kind)
		}
	}
}

func TestDiffView_View_NoDifferences(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "test", 1, 2)
	v.loading = false
	v.lines = []diffLine{}

	view := v.View(80, 20)
	if !strings.Contains(view, "No differences") {
		t.Error("expected no differences message")
	}
}

func TestDiffView_KeyHints(t *testing.T) {
	v := NewDiffView(newTestClient(t), "secret/", "test", 1, 2)
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
