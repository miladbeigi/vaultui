package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	clearTimer *time.Timer
	clearMu    sync.Mutex
)

// Write places text on the system clipboard.
func Write(text string) error {
	return writeClipboard(text)
}

// WriteWithAutoClear places text on the system clipboard and schedules
// automatic clearing after the given duration. Pass 0 to skip auto-clear.
func WriteWithAutoClear(text string, clearAfter time.Duration) error {
	if err := writeClipboard(text); err != nil {
		return err
	}

	if clearAfter <= 0 {
		return nil
	}

	clearMu.Lock()
	defer clearMu.Unlock()

	if clearTimer != nil {
		clearTimer.Stop()
	}
	clearTimer = time.AfterFunc(clearAfter, func() {
		_ = writeClipboard("")
	})

	return nil
}

// CancelAutoClear cancels any pending clipboard auto-clear timer.
func CancelAutoClear() {
	clearMu.Lock()
	defer clearMu.Unlock()
	if clearTimer != nil {
		clearTimer.Stop()
		clearTimer = nil
	}
}

func writeClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip, xsel, or wl-copy)")
		}
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
