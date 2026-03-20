package tuiapp

import "github.com/charmbracelet/lipgloss"

// Theme holds colors resolved through the lipgloss renderer’s termenv ColorProfile, so ANSI palette
// indices follow the user’s terminal theme (Windows Terminal, iTerm, etc.).
//
// Indices refer to the xterm 256-color palette where applicable; lipgloss degrades for ANSI 16-color
// and TrueColor terminals automatically.
//
// See: lipgloss.Color, lipgloss.Renderer, termenv.Output.
type Theme struct {
	// Primary is default high-contrast text (still uses Adaptive for light/dark terminal bg).
	Primary lipgloss.AdaptiveColor
	// Secondary / Muted / Border use palette slots that track the terminal theme.
	Secondary lipgloss.TerminalColor
	Muted     lipgloss.TerminalColor
	Highlight lipgloss.TerminalColor // selection background (tables, focused notify row)
	Border    lipgloss.TerminalColor
}

// getDefaultTheme returns palette-based colors. Call lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))
// before rendering so the active terminal profile is detected.
func getDefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{
			Light: "#000000",
			Dark:  "#FFFFFF",
		},
		Secondary: lipgloss.Color("8"),  // bright black / gray
		Muted:     lipgloss.Color("13"), // bright magenta — key labels
		Highlight: lipgloss.Color("5"), // selection bg (palette slot; theme-mapped)
		Border:    lipgloss.Color("8"),
	}
}
