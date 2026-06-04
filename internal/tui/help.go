package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpPopup struct {
	visible bool
}

func newHelpPopup() helpPopup {
	return helpPopup{}
}

func (h *helpPopup) Toggle() {
	h.visible = !h.visible
}

func (h *helpPopup) View(width, height int) string {
	if !h.visible {
		return ""
	}

	title := helpTitleStyle.Render("TForge Help")

	sections := []struct {
		header string
		binds  []struct{ key, desc string }
	}{
		{
			header: "Navigation (vim)",
			binds: []struct{ key, desc string }{
				{"j/k", "up / down"},
				{"h/l", "back / drill in"},
				{"gg/G", "top / bottom"},
				{"ctrl+u/d", "half page up / down"},
				{"H/L", "prev / next stack"},
				{"/", "filter (type, Enter=keep, Esc=clear)"},
				{"q", "quit (or back from detail)"},
			},
		},
		{
			header: "Selection",
			binds: []struct{ key, desc string }{
				{"v/space", "toggle select"},
				{".", "cycle ellipsis (resolves DAG)"},
				{"A", "select all"},
				{"N", "select none"},
			},
		},
		{
			header: "Stack Commands",
			binds: []struct{ key, desc string }{
				{"p", "plan (selected units)"},
				{"a", "apply (selected units)"},
				{"d", "destroy (confirm)"},
				{"S", "stack generate"},
				{"C", "stack clean"},
				{"I", "stack init"},
				{"b", "backend bootstrap"},
				{"f", "view files"},
				{"F", "fullscreen output"},
				{"T", "toggle CLI preview"},
				{"ctrl+x", "cancel running command"},
			},
		},
		{
			header: "Unit Detail (after drill-in)",
			binds: []struct{ key, desc string }{
				{"i", "init"},
				{"p", "plan"},
				{"a", "apply"},
				{"d", "destroy (confirm)"},
				{"r", "refresh state (apply -refresh-only)"},
				{"V", "validate"},
				{"e", "render resolved config"},
				{"s", "state list"},
				{"o", "show outputs"},
				{"l/enter", "state show (on resource)"},
				{"x", "state rm (confirm)"},
				{"R", "replace resource (apply -replace)"},
				{"m", "import resource"},
				{"M", "state mv"},
				{"U", "force-unlock"},
				{"f", "view files"},
				{"C", "clean cache"},
				{"tab/S-tab", "cycle tabs"},
			},
		},
		{
			header: "Ellipsis (DAG-aware selection)",
			binds: []struct{ key, desc string }{
				{"[▸]", "unit... (+ upstream deps)"},
				{"[◂]", "...unit (+ downstream dependents)"},
				{"", "one anchor at a time; resolves"},
				{"", "transitive deps and selects them"},
			},
		},
	}

	var content []string
	content = append(content, title)
	content = append(content, "")

	for _, section := range sections {
		content = append(content, helpTitleStyle.Render("─── "+section.header+" ───"))
		for _, b := range section.binds {
			line := fmt.Sprintf("%s %s", helpKeyStyle.Render(b.key), helpDescStyle.Render(b.desc))
			content = append(content, line)
		}
		content = append(content, "")
	}

	content = append(content, keybindDescStyle.Render("press ? or Esc to close"))

	body := strings.Join(content, "\n")

	popupWidth := 40
	popupHeight := len(content) + 4
	popup := helpBorderStyle.Width(popupWidth).Render(body)

	x := (width - popupWidth - 4) / 2
	y := (height - popupHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, popup)
}
