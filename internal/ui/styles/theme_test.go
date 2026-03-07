package styles

import (
	"testing"
)

func TestApplyTheme_Dark(t *testing.T) {
	ApplyTheme("dark")
	if CurrentThemeName() != "dark" {
		t.Errorf("expected dark, got %s", CurrentThemeName())
	}
	if PrimaryColor != DarkPalette.Primary {
		t.Error("expected dark primary color")
	}
}

func TestApplyTheme_Light(t *testing.T) {
	ApplyTheme("light")
	if CurrentThemeName() != "light" {
		t.Errorf("expected light, got %s", CurrentThemeName())
	}
	if PrimaryColor != LightPalette.Primary {
		t.Error("expected light primary color")
	}
	ApplyTheme("dark")
}

func TestApplyTheme_Unknown_DefaultsToDark(t *testing.T) {
	ApplyTheme("unknown")
	if CurrentThemeName() != "dark" {
		t.Errorf("expected dark for unknown theme, got %s", CurrentThemeName())
	}
}

func TestApplyTheme_Toggle(t *testing.T) {
	ApplyTheme("dark")
	if CurrentThemeName() != "dark" {
		t.Fatal("expected dark")
	}
	ApplyTheme("light")
	if CurrentThemeName() != "light" {
		t.Fatal("expected light")
	}
	ApplyTheme("dark")
	if CurrentThemeName() != "dark" {
		t.Fatal("expected dark again")
	}
}
