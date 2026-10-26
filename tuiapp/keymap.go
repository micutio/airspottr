package tuiapp

import tea "github.com/charmbracelet/bubbletea"

const (
	notifyType = iota
	notifyOperator
	notifyCountry
	notifyStripCount = 3
)

// handleKey processes keyboard input for the TUI.
func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if m.inputFocus == focusNotifyStrip {
		m.navigateNotifyStrip(key)
		return nil
	}

	activeTable := m.activeTable()
	isFocused := activeTable.table.Focused()

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
		if isFocused {
			m.UnfocusSelectedTable()
		} else {
			m.FocusSelectedTable()
		}
		return nil
	case "left", "h":
		if m.uiState == globalStats {
			m.selectRarityNeighbour(-1)
		}
		return nil
	case "right", "l":
		if m.uiState == globalStats {
			m.selectRarityNeighbour(1)
		}
		return nil
	case " ":
		m.toggleGlobalView()
		return nil
	case "tab", "shift+tab":
		m.inputFocus = focusNotifyStrip
		m.UnfocusSelectedTable()
		return nil
	case "[":
		if isFocused {
			activeTable.cycleSortCol(-1)
			m.updateAllTables()
		}
		return nil
	case "]":
		if isFocused {
			activeTable.cycleSortCol(1)
			m.updateAllTables()
		}
		return nil
	case "r", "R":
		if isFocused {
			activeTable.toggleSortDir()
			m.updateAllTables()
		}
		return nil
	case "1", "2", "3", "4", "5", "6", "7", "8":
		if isFocused && activeTable.setSortCol(int(key[0]-'1')) {
			m.updateAllTables()
		}
		return nil
	}

	if isFocused {
		var cmd tea.Cmd
		activeTable.table, cmd = activeTable.table.Update(msg)
		return cmd
	}
	return nil
}

func (m *model) navigateNotifyStrip(key string) {
	switch key {
	case "tab", "shift+tab", "esc":
		m.inputFocus = focusTable
		m.FocusSelectedTable()
	case "left", "h", "up", "k":
		m.notifyStripIdx = (m.notifyStripIdx - 1 + notifyStripCount) % notifyStripCount
	case "right", "l", "down", "j":
		m.notifyStripIdx = (m.notifyStripIdx + 1) % notifyStripCount
	case " ":
		switch m.notifyStripIdx {
		case notifyType:
			m.notifyOnType = !m.notifyOnType
		case notifyOperator:
			m.notifyOnOp = !m.notifyOnOp
		case notifyCountry:
			m.notifyOnCountry = !m.notifyOnCountry
		}
	}
}
