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

	if m.inputFocus == focusNotifyStrip {
		switch key {
		case "tab", "shift+tab", "esc":
			m.leaveNotifyStrip()
		case "left", "h", "up", "k":
			m.notifyStripIdx = (m.notifyStripIdx + 2) % 3
		case "right", "l", "down", "j":
			m.notifyStripIdx = (m.notifyStripIdx + 1) % 3
		case " ":
			m.toggleNotifyAt(m.notifyStripIdx)
		}
		return nil
	}

	at := m.activeTable()
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
	case "tab", "shift+tab":
		m.enterNotifyStrip()
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
