package tuiapp

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	at := m.activeTable()
	switch msg.String() {
	case "esc":
		if at.table.Focused() {
			m.UnfocusSelectedTable()
		} else {
			m.FocusSelectedTable()
		}
	case "up", "k":
		if at.table.Focused() {
			at.table.MoveUp(1)
		}
	case "pgup":
		if at.table.Focused() {
			at.table.MoveUp(at.table.Height() - 1)
		}
	case "down", "j":
		if at.table.Focused() {
			at.table.MoveDown(1)
		}
	case "pgdown":
		if at.table.Focused() {
			at.table.MoveDown(at.table.Height() - 1)
		}
	case "left", "h":
		if m.uiState == globalStats {
			m.selectRarityNeighbour(-1)
		}
	case "right", "l":
		if m.uiState == globalStats {
			m.selectRarityNeighbour(1)
		}
	case " ":
		m.toggleGlobalView()
	case "q", "ctrl+c":
		return tea.Quit
	}
	return nil
}
