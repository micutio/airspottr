package tuiapp

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	column := m.baseStyle.Width(m.width).Padding(0, 0, 0, 0).Render
	var tableContent string
	switch m.uiState {
	case mainPage:
		tableContent = m.viewBorderedTable(&m.tables.aircraft)
	case globalStats:
		parts := make([]string, rarityTableCount)
		for i := range m.tables.rarities {
			parts[i] = m.viewBorderedTable(&m.tables.rarities[i])
		}
		tableContent = lipgloss.JoinHorizontal(lipgloss.Top, parts[0], parts[1], parts[2])
	case aircraftDetails:
	}
	body := []string{column(m.viewHeader()), column(tableContent)}
	if m.uiState == mainPage || m.uiState == globalStats {
		body = append(body, m.viewSortHotkeyHint())
	}
	return m.baseStyle.
		Width(m.width).
		Height(m.height).
		Render(lipgloss.JoinVertical(lipgloss.Left, body...))
}

func (m *model) viewBorderedTable(tbl *autoFormatTable) string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(tbl.table.View())
}

func (m *model) viewHeader() string {
	list := m.baseStyle.
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(m.theme.Border).
		Height(1).
		Padding(0, 0)

	keyStyle := m.baseStyle.Foreground(lipgloss.AdaptiveColor{Light: "#383838", Dark: "#F988F9"})
	listHeader := m.baseStyle.Bold(true).Render

	listItem := func(key, value string) string {
		listItemValue := m.baseStyle.Align(lipgloss.Right).Render(value)
		listItemKey := func(k string) string {
			return keyStyle.Render(k + ":")
		}
		return fmt.Sprintf("%s %s ", listItemKey(key), listItemValue)
	}

	const minutesInHour = 60.0
	const secsInMinute = 60.0
	tSince := time.Since(m.startTime)
	hours := tSince.Hours()
	mins := math.Mod(math.Floor(tSince.Minutes()), minutesInHour)
	secs := math.Mod(math.Floor(tSince.Seconds()), secsInMinute)

	leftPanel := list.Border(lipgloss.RoundedBorder()).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			fmt.Sprintf("   Location %.3f, %.3f", m.dashboard.Lat, m.dashboard.Lon),
			fmt.Sprintf("     UpTime %.0f Hr %02.0f Min %02.0f Sec", hours, mins, secs),
			fmt.Sprintf("Last Update %02.0f seconds ago", time.Since(m.lastUpdate).Seconds())),
	)

	notifyPanel := list.Border(lipgloss.RoundedBorder()).Render(m.viewNotifyListContent())

	var rightPanel string
	highest := m.dashboard.Highest
	fastest := m.dashboard.Fastest
	if highest != nil && fastest != nil {
		rightPanel = list.Border(lipgloss.RoundedBorder()).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				listHeader("Highest"),
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					listItem("ALT", highest.GetAltitudeAsStr()),
					listItem("FNO", highest.GetFlightNoAsStr()),
					listItem("REG", highest.Registration),
					listItem("TID", m.dashboard.IcaoToAircraft[highest.IcaoType].Make),
				),
				listHeader("Fastest"),
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					listItem("SPD", fmt.Sprintf("%5.0f", fastest.GroundSpeed)),
					listItem("FNO", fastest.GetFlightNoAsStr()),
					listItem("REG", fastest.Registration),
					listItem("TID", m.dashboard.IcaoToAircraft[fastest.IcaoType].Make),
				),
			),
		)
	} else {
		rightPanel = list.Border(lipgloss.RoundedBorder()).Render(
			m.baseStyle.Foreground(m.theme.Secondary).Render(
				"  Awaiting\n  aircraft\n  data…",
			),
		)
	}

	return m.viewStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, notifyPanel, rightPanel),
	)
}

// viewNotifyListContent is the third header column: Notify list (Type / Op / Country).
func (m *model) viewNotifyListContent() string {
	check := func(on bool) string {
		if on {
			return "[x]"
		}
		return "[ ]"
	}
	rows := []string{m.baseStyle.Bold(true).Render("Notify")}
	items := []struct {
		label string
		on    bool
		idx   int
	}{
		{"Type", m.notifyOnType, 0},
		{"Op", m.notifyOnOp, 1},
		{"Country", m.notifyOnCountry, 2},
	}
	for _, it := range items {
		text := fmt.Sprintf("%s %s", check(it.on), it.label)
		st := m.baseStyle.Foreground(m.theme.Secondary)
		var line string
		if m.inputFocus == focusNotifyStrip && m.notifyStripIdx == it.idx {
			line = st.Border(lipgloss.NormalBorder()).
				BorderForeground(m.theme.Highlight).
				Padding(0, 1).
				Render(text)
		} else {
			line = " " + st.Render(text)
		}
		rows = append(rows, line)
	}
	return strings.Join(rows, "\n")
}

func (m *model) viewSortHotkeyHint() string {
	msg := "Tab notify · t/o/c toggles · Sort (table): [ ] r 1–8 · rarity 1=count 2=name"
	return m.baseStyle.
		Width(m.width).
		Foreground(m.theme.Secondary).
		Render(msg)
}
