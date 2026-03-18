package tuiapp

import (
	"fmt"
	"math"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	column := m.baseStyle.Width(m.width).Padding(0, 0, 0, 0).Render
	var tableContent string
	switch m.uiState {
	case mainPage:
		tableContent = m.viewAircraft()
	case globalStats:
		tableContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.viewTypeRarity(),
			m.viewOperatorRarity(),
			m.viewCountryRarity(),
		)
	case aircraftDetails:
	}
	return m.baseStyle.
		Width(m.width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				column(m.viewHeader()),
				column(tableContent),
			),
		)
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

	highest := m.dashboard.Highest
	fastest := m.dashboard.Fastest
	if highest == nil || fastest == nil {
		return ""
	}

	const minutesInHour = 60.0
	const secsInMinute = 60.0
	tSince := time.Since(m.startTime)
	hours := tSince.Hours()
	mins := math.Mod(math.Floor(tSince.Minutes()), minutesInHour)
	secs := math.Mod(math.Floor(tSince.Seconds()), secsInMinute)

	return m.viewStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			list.Border(lipgloss.RoundedBorder()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("   Location %.3f, %.3f", m.dashboard.Lat, m.dashboard.Lon),
					fmt.Sprintf("     UpTime %.0f Hr %02.0f Min %02.0f Sec", hours, mins, secs),
					fmt.Sprintf("Last Update %02.0f seconds ago", time.Since(m.lastUpdate).Seconds())),
			),
			list.Border(lipgloss.RoundedBorder()).Render(
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
			),
		),
	)
}

func (m *model) viewAircraft() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.currentAircraftTbl.table.View())
}

func (m *model) viewTypeRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.typeRarityTbl.table.View())
}

func (m *model) viewOperatorRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.operatorRarityTbl.table.View())
}

func (m *model) viewCountryRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.countryRarityTbl.table.View())
}
