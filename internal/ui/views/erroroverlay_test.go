package views

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorOverlay_Title(t *testing.T) {
	v := NewErrorOverlayView("Test Error", errors.New("something failed"))
	if v.Title() != "Test Error" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestErrorOverlay_View_ShowsError(t *testing.T) {
	v := NewErrorOverlayView("Connection Failed", errors.New("connection refused"))
	view := v.View(80, 24)
	if !strings.Contains(view, "Connection Failed") {
		t.Error("expected title in view")
	}
	if !strings.Contains(view, "connection refused") {
		t.Error("expected error message in view")
	}
}

func TestErrorOverlay_View_ShowsHints(t *testing.T) {
	v := NewErrorOverlayView("Error", errors.New("connection refused"))
	view := v.View(80, 24)
	if !strings.Contains(view, "Troubleshooting") {
		t.Error("expected troubleshooting section")
	}
}

func TestErrorOverlay_KeyHints(t *testing.T) {
	v := NewErrorOverlayView("Error", errors.New("test"))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints")
	}
}

func TestGenerateHints_ConnectionRefused(t *testing.T) {
	hints := generateHints(errors.New("dial tcp: connection refused"))
	found := false
	for _, h := range hints {
		if strings.Contains(h, "Vault is running") {
			found = true
		}
	}
	if !found {
		t.Error("expected connectivity hint for connection refused")
	}
}

func TestGenerateHints_PermissionDenied(t *testing.T) {
	hints := generateHints(errors.New("permission denied"))
	found := false
	for _, h := range hints {
		if strings.Contains(h, "policy") {
			found = true
		}
	}
	if !found {
		t.Error("expected policy hint for permission denied")
	}
}

func TestGenerateHints_TLS(t *testing.T) {
	hints := generateHints(errors.New("x509: certificate signed by unknown authority"))
	found := false
	for _, h := range hints {
		if strings.Contains(h, "VAULT_SKIP_VERIFY") {
			found = true
		}
	}
	if !found {
		t.Error("expected TLS hint for certificate error")
	}
}

func TestGenerateHints_Unknown(t *testing.T) {
	hints := generateHints(errors.New("some random error"))
	if len(hints) == 0 {
		t.Error("expected default hints for unknown error")
	}
}

func TestGenerateHints_Nil(t *testing.T) {
	hints := generateHints(nil)
	if hints != nil {
		t.Error("expected nil hints for nil error")
	}
}
