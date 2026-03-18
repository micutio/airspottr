package tuiapp

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		if m.selectedTable.table.Focused() {
			m.UnfocusSelectedTable()
		} else {
			m.FocusSelectedTable()
		}
	case "up", "k":
		if m.selectedTable.table.Focused() {
			m.selectedTable.table.MoveUp(1)
		}
	case "pgup":
		if m.selectedTable.table.Focused() {
			m.selectedTable.table.MoveUp(m.selectedTable.table.Height() - 1)
		}
	case "down", "j":
		if m.selectedTable.table.Focused() {
			m.selectedTable.table.MoveDown(1)
		}
	case "pgdown":
		if m.selectedTable.table.Focused() {
			m.selectedTable.table.MoveDown(m.selectedTable.table.Height() - 1)
		}
	case "left", "h":
		m.selectTableToTheLeft()
	case "right", "l":
		m.selectTableToTheRight()
	case " ":
		m.toggleGlobalView()
	case "q", "ctrl+c":
		return tea.Quit
	}
	return nil
}
