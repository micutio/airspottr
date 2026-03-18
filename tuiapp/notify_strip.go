package tuiapp

func (m *model) enterNotifyStrip() {
	m.inputFocus = focusNotifyStrip
	m.UnfocusSelectedTable()
}

func (m *model) leaveNotifyStrip() {
	m.inputFocus = focusTable
	m.FocusSelectedTable()
}

func (m *model) toggleNotifyAt(idx int) {
	switch idx {
	case 0:
		m.notifyOnType = !m.notifyOnType
	case 1:
		m.notifyOnOp = !m.notifyOnOp
	case 2:
		m.notifyOnCountry = !m.notifyOnCountry
	}
}
