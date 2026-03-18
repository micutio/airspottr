package tuiapp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKeyQuit(t *testing.T) {
	t.Parallel()
	var m model
	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected non-nil quit command")
	}
}

func TestHandleKeyCtrlC(t *testing.T) {
	t.Parallel()
	var m model
	cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected non-nil quit command")
	}
}
