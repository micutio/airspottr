package tuiapp

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	at := m.activeTable()
	key := msg.String()

	switch key {
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
	case "[", "]":
		if at.table.Focused() {
			if key == "[" {
				m.cycleSortColumn(-1)
			} else {
				m.cycleSortColumn(1)
			}
		}
	case "r", "R":
		if at.table.Focused() {
			m.toggleSortDirection()
		}
	case "1", "2", "3", "4", "5", "6", "7", "8":
		if !at.table.Focused() {
			break
		}
		if m.uiState == mainPage {
			m.aircraftSortCol = int(key[0] - '1')
			m.updateAllTables()
		} else if m.uiState == globalStats {
			switch key {
			case "1":
				m.raritySortCol[m.selectedRarityIdx] = 0
				m.updateAllTables()
			case "2":
				m.raritySortCol[m.selectedRarityIdx] = 1
				m.updateAllTables()
			}
		}
	}
	return nil
}
