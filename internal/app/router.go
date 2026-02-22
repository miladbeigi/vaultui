package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/milad/vaultui/internal/ui"
)

// Router manages a stack of views for push/pop navigation.
// Pressing Enter typically pushes a new view; Esc/Back pops to the previous one.
// Each view on the stack preserves its full state (scroll position, filters, etc.).
type Router struct {
	stack []ui.View
}

// NewRouter creates an empty router.
func NewRouter() *Router {
	return &Router{}
}

// Push adds a view to the top of the stack and returns its Init command.
func (r *Router) Push(v ui.View) tea.Cmd {
	r.stack = append(r.stack, v)
	return v.Init()
}

// Pop removes the top view from the stack.
// Returns false if the stack has one or fewer views (won't pop the root).
func (r *Router) Pop() bool {
	if len(r.stack) <= 1 {
		return false
	}
	r.stack = r.stack[:len(r.stack)-1]
	return true
}

// Current returns the view at the top of the stack, or nil if empty.
func (r *Router) Current() ui.View {
	if len(r.stack) == 0 {
		return nil
	}
	return r.stack[len(r.stack)-1]
}

// Replace swaps the current top view with a new one and returns its Init command.
// If the stack is empty, this behaves like Push.
func (r *Router) Replace(v ui.View) tea.Cmd {
	if len(r.stack) > 0 {
		r.stack[len(r.stack)-1] = v
	} else {
		r.stack = append(r.stack, v)
	}
	return v.Init()
}

// Depth returns the number of views on the stack.
func (r *Router) Depth() int {
	return len(r.stack)
}
