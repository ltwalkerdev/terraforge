package tui

import "github.com/charmbracelet/lipgloss"

var (
	purple      = lipgloss.Color("#A855F7")
	purpleLight = lipgloss.Color("#C084FC")
	purpleDim   = lipgloss.Color("#7C3AED")
	white       = lipgloss.Color("#FFFFFF")
	lightGrey   = lipgloss.Color("#D4D4D4")
	midGrey     = lipgloss.Color("#A3A3A3")
	dimGrey     = lipgloss.Color("#737373")
	darkBg      = lipgloss.Color("#0A0A0A")
	panelBg     = lipgloss.Color("#171717")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purpleDim).
			Padding(0, 1)

	stackBarStyle = lipgloss.NewStyle().
			Foreground(lightGrey).
			Background(panelBg).
			Padding(0, 1)

	stackActiveStyle = lipgloss.NewStyle().
				Foreground(darkBg).
				Background(purple).
				Bold(true).
				Padding(0, 1)

	stackInactiveStyle = lipgloss.NewStyle().
				Foreground(dimGrey).
				Padding(0, 1)

	unitSelectedStyle = lipgloss.NewStyle().
				Foreground(purpleLight).
				Bold(true)

	unitNormalStyle = lipgloss.NewStyle().
			Foreground(lightGrey)

	unitCursorStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	statusCleanStyle = lipgloss.NewStyle().
				Foreground(lightGrey)

	statusChangedStyle = lipgloss.NewStyle().
				Foreground(purpleLight)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(purple)

	purpleRunningStyle = lipgloss.NewStyle().
				Foreground(purpleLight).
				Bold(true)

	statusUnknownStyle = lipgloss.NewStyle().
				Foreground(dimGrey)

	outputStyle = lipgloss.NewStyle().
			Foreground(midGrey)

	outputErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lightGrey).
			Background(panelBg).
			Padding(0, 1)

	keybindStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	keybindDescStyle = lipgloss.NewStyle().
				Foreground(dimGrey)

	helpBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2).
			Background(panelBg)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(purpleLight).
			Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true).
			Width(12)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lightGrey)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(purpleDim)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleDim)

	paneActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purple)

	paneTitleStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			Padding(0, 1)

	paneTitleActiveStyle = lipgloss.NewStyle().
				Foreground(white).
				Background(purple).
				Bold(true).
				Padding(0, 1)

	purpleBannerStyle = lipgloss.NewStyle().
				Foreground(purpleLight).
				Bold(true)

	splashStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)
)
