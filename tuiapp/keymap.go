package tuiapp

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if m.inputFocus == focusNotifyStrip {
		m.navigateNotifyStrip(key)
		return nil
	}

	activeTable := m.activeTable()

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
	case "esc":
		m.toggleTableFocus(activeTable.table.Focused())
	case "up", "k":
		m.moveUpInActiveTable(activeTable.table.Focused())
	case "pgup":
		m.pageUpInActiveTable(activeTable.table.Focused())
	case "down", "j":
		m.moveDownInActiveTable(activeTable.table.Focused())
	case "pgdown":
		m.pageDownInActiveTable(activeTable.table.Focused())
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
		m.cycleTableColumn(key, activeTable.table.Focused())
	case "r", "R":
		if activeTable.table.Focused() {
			m.toggleSortDirection()
		}
	case "1", "2", "3", "4", "5", "6", "7", "8":
		m.sortTables(key, activeTable.table.Focused())
	}
	return nil
}

func (m *model) navigateNotifyStrip(key string) {
	notifyStripCount := 3
	switch key {
	case "tab", "shift+tab", "esc":
		m.leaveNotifyStrip()
	case "left", "h", "up", "k":
		m.notifyStripIdx = (m.notifyStripIdx - 1) % notifyStripCount
	case "right", "l", "down", "j":
		m.notifyStripIdx = (m.notifyStripIdx + 1) % notifyStripCount
	case " ":
		m.toggleNotifyAt(m.notifyStripIdx)
	}
}

func (m *model) toggleTableFocus(isActiveTableFocused bool) {
	if isActiveTableFocused {
		m.UnfocusSelectedTable()
	} else {
		m.FocusSelectedTable()
	}
}

func (m *model) moveUpInActiveTable(isActiveTableFocused bool) {
	if isActiveTableFocused {
		m.activeTable().table.MoveUp(1)
	}
}

func (m *model) pageUpInActiveTable(isActiveTableFocused bool) {
	if isActiveTableFocused {
		m.activeTable().table.MoveUp(m.activeTable().table.Height() - 1)
	}
}

func (m *model) moveDownInActiveTable(isActiveTableFocused bool) {
	if isActiveTableFocused {
		m.activeTable().table.MoveDown(1)
	}
}

func (m *model) pageDownInActiveTable(isActiveTableFocused bool) {
	if isActiveTableFocused {
		m.activeTable().table.MoveDown(m.activeTable().table.Height() - 1)
	}
}

func (m *model) cycleTableColumn(key string, isActiveTableFocused bool) {
	if isActiveTableFocused {
		if key == "[" {
			m.cycleSortColumn(-1)
		} else {
			m.cycleSortColumn(1)
		}
	}
}

func (m *model) sortTables(key string, isActiveTableFocused bool) {
	if !isActiveTableFocused {
		return
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
