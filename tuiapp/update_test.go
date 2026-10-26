//nolint:exhaustruct_v5 // Don't enforce for bubbletea.
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
	testModel := model{
		inputFocus:     focusNotifyStrip,
		notifyStripIdx: 0,
	}
	// Moving left from index 0 should wrap to index 2, not become -1.
	testModel.navigateNotifyStrip("left")
	if testModel.notifyStripIdx != notifyCountry {
		t.Fatalf("notifyStripIdx after left wrap = %d, want %d", testModel.notifyStripIdx, notifyCountry)
	}

	// Spacebar toggles the selected notify flag.
	testModel.navigateNotifyStrip(" ")
	if !testModel.notifyOnCountry {
		t.Fatal("expected notifyOnCountry to be toggled true")
	}

	// Moving right from index 2 should wrap back to index 0.
	testModel.navigateNotifyStrip("right")
	if testModel.notifyStripIdx != notifyType {
		t.Fatalf("notifyStripIdx after right wrap = %d, want %d", testModel.notifyStripIdx, notifyType)
	}
}
