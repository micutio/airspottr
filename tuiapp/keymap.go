package tuiapp

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		return tea.Quit
	case "t":
		m.notifyOnType = !m.notifyOnType
		return nil
	case "o":
		m.notifyOnOp = !m.notifyOnOp
		return nil
	case "c":
		m.notifyOnCountry = !m.notifyOnCountry
		return nil
	}

	notifyStripCount := 3
	if m.inputFocus == focusNotifyStrip {
		switch key {
		case "tab", "shift+tab", "esc":
			m.leaveNotifyStrip()
		case "left", "h", "up", "k":
			m.notifyStripIdx = (m.notifyStripIdx + 2) % notifyStripCount
		case "right", "l", "down", "j":
			m.notifyStripIdx = (m.notifyStripIdx + 1) % notifyStripCount
		case " ":
			m.toggleNotifyAt(m.notifyStripIdx)
		}
		return nil
	}

	activeTable := m.activeTable()
	switch key {
	case "esc":
		if activeTable.table.Focused() {
			m.UnfocusSelectedTable()
		} else {
			m.FocusSelectedTable()
		}
	case "up", "k":
		if activeTable.table.Focused() {
			activeTable.table.MoveUp(1)
		}
	case "pgup":
		if activeTable.table.Focused() {
			activeTable.table.MoveUp(activeTable.table.Height() - 1)
		}
	case "down", "j":
		if activeTable.table.Focused() {
			activeTable.table.MoveDown(1)
		}
	case "pgdown":
		if activeTable.table.Focused() {
			activeTable.table.MoveDown(activeTable.table.Height() - 1)
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
	case "tab", "shift+tab":
		m.enterNotifyStrip()
	case "[", "]":
		if activeTable.table.Focused() {
			if key == "[" {
				m.cycleSortColumn(-1)
			} else {
				m.cycleSortColumn(1)
			}
		}
	case "r", "R":
		if activeTable.table.Focused() {
			m.toggleSortDirection()
		}
	case "1", "2", "3", "4", "5", "6", "7", "8":
		if !activeTable.table.Focused() {
			break
		}
		switch m.uiState {
		case mainPage:
			m.aircraftSortCol = int(key[0] - '1')
			m.updateAllTables()
		case globalStats:
			switch key {
			case "1":
				m.raritySortCol[m.selectedRarityIdx] = 0
				m.updateAllTables()
			case "2":
				m.raritySortCol[m.selectedRarityIdx] = 1
				m.updateAllTables()
			}
		default:
			panic("unhandled default case")
		}
	}
	return nil
}
