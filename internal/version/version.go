package version

import (
	"fmt"
	"runtime"
)

// Set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("vaultui %s (commit: %s, built: %s, %s/%s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH)
}
