package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

const contentWidth = 40

var (
	top = "╔" + strings.Repeat("═", contentWidth) + "╗"
	mid = "╠" + strings.Repeat("═", contentWidth) + "╣"
	bot = "╚" + strings.Repeat("═", contentWidth) + "╝"
)

func blank() string {
	return "║" + strings.Repeat(" ", contentWidth) + "║"
}

func padded(label, val string) string {
	line := fmt.Sprintf("   %s %-15s", label+":", val)
	pad := contentWidth - len(line)
	if pad < 0 {
		pad = 0
	}
	return "║" + line + strings.Repeat(" ", pad) + "║"
}

func centered(s string) string {
	extra := contentWidth - len(s)
	if extra < 0 {
		return "║" + s + "║"
	}
	left := extra / 2
	return "║" + strings.Repeat(" ", left) + s + strings.Repeat(" ", extra-left) + "║"
}

func Banner() string {
	return strings.Join([]string{
		top,
		blank(),
		centered("VaultUI TUI"),
		blank(),
		mid,
		blank(),
		padded("Version", Version),
		padded("Commit", Commit),
		padded("Built", Date),
		blank(),
		bot,
	}, "\n")
}

func String() string {
	return fmt.Sprintf("vaultui %s (%s, commit: %s, built: %s, %s/%s)\n\n%s",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH, Banner())
}
