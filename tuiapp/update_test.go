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

func TestNavigateNotifyStripWrapsAround(t *testing.T) {
	t.Parallel()
	m := model{
		inputFocus:     focusNotifyStrip,
		notifyStripIdx: 0,
	}
	// Moving left from index 0 should wrap to index 2, not become -1.
	m.navigateNotifyStrip("left")
	if m.notifyStripIdx != notifyCountry {
		t.Fatalf("notifyStripIdx after left wrap = %d, want %d", m.notifyStripIdx, notifyCountry)
	}

	// Spacebar toggles the selected notify flag.
	m.navigateNotifyStrip(" ")
	if !m.notifyOnCountry {
		t.Fatal("expected notifyOnCountry to be toggled true")
	}

	// Moving right from index 2 should wrap back to index 0.
	m.navigateNotifyStrip("right")
	if m.notifyStripIdx != notifyType {
		t.Fatalf("notifyStripIdx after right wrap = %d, want %d", m.notifyStripIdx, notifyType)
	}
}
