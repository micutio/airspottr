package tuiapp

const (
	notifyType     = 0
	notifyOperator = 1
	notifyCountry  = 2
)

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
	case notifyType:
		m.notifyOnType = !m.notifyOnType
	case notifyOperator:
		m.notifyOnOp = !m.notifyOnOp
	case notifyCountry:
		m.notifyOnCountry = !m.notifyOnCountry
	}
}
