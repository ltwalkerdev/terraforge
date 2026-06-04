package tui

import (
	"fmt"
	"strings"

	"github.com/ltwalkerdev/terraforge/internal/model"
)

type detailTab int

const (
	tabState detailTab = iota
	tabOutputs
	tabPlan
)

type unitDetail struct {
	unit       model.Unit
	stack      model.Stack
	tab        detailTab
	stateItems []string
	cursor     int
	loaded     bool
}

func newUnitDetail() unitDetail {
	return unitDetail{}
}

func (d *unitDetail) SetUnit(unit model.Unit, stack model.Stack) {
	d.unit = unit
	d.stack = stack
	d.tab = tabState
	d.stateItems = nil
	d.cursor = 0
	d.loaded = false
}

func (d *unitDetail) SetStateItems(items []string) {
	d.stateItems = items
	d.cursor = 0
	d.loaded = true
}

func (d *unitDetail) NextTab() {
	d.tab = (d.tab + 1) % 3
}

func (d *unitDetail) PrevTab() {
	d.tab = (d.tab - 1 + 3) % 3
}

func (d *unitDetail) Up() {
	if d.cursor > 0 {
		d.cursor--
	}
}

func (d *unitDetail) Down() {
	if d.tab == tabState && d.cursor < len(d.stateItems)-1 {
		d.cursor++
	}
}

func (d *unitDetail) GoToTop() {
	d.cursor = 0
}

func (d *unitDetail) GoToBottom() {
	if d.tab == tabState && len(d.stateItems) > 0 {
		d.cursor = len(d.stateItems) - 1
	}
}

func (d *unitDetail) HalfPageUp(viewHeight int) {
	d.cursor -= viewHeight / 2
	if d.cursor < 0 {
		d.cursor = 0
	}
}

func (d *unitDetail) HalfPageDown(viewHeight int) {
	jump := viewHeight / 2
	d.cursor += jump
	max := len(d.stateItems) - 1
	if max < 0 {
		max = 0
	}
	if d.cursor > max {
		d.cursor = max
	}
}

func (d *unitDetail) SelectedResource() string {
	if d.tab == tabState && d.cursor < len(d.stateItems) {
		return d.stateItems[d.cursor]
	}
	return ""
}

func (d *unitDetail) View(width, height int) string {
	var lines []string

	// Breadcrumb
	breadcrumb := unitCursorStyle.Render(fmt.Sprintf("  %s / %s", d.stack.Name, d.unit.Name))
	lines = append(lines, breadcrumb)
	lines = append(lines, "")

	// Tab bar
	tabs := []string{"State", "Outputs", "Plan"}
	var tabLine []string
	for i, t := range tabs {
		if detailTab(i) == d.tab {
			tabLine = append(tabLine, stackActiveStyle.Render(fmt.Sprintf(" %s ", t)))
		} else {
			tabLine = append(tabLine, stackInactiveStyle.Render(fmt.Sprintf(" %s ", t)))
		}
	}
	lines = append(lines, "  "+strings.Join(tabLine, " "))
	lines = append(lines, "")

	// Content
	contentHeight := height - 5
	switch d.tab {
	case tabState:
		lines = append(lines, d.stateView(width, contentHeight)...)
	case tabOutputs:
		lines = append(lines, outputStyle.Render("  press 'o' to load outputs")+"")
	case tabPlan:
		lines = append(lines, outputStyle.Render("  press 'p' to run plan")+"")
	}

	return strings.Join(lines, "\n")
}

func (d *unitDetail) stateView(width, height int) []string {
	if !d.loaded {
		return []string{outputStyle.Render("  press 's' to load state list")}
	}
	if len(d.stateItems) == 0 {
		return []string{outputStyle.Render("  (no resources in state)")}
	}

	var lines []string
	visibleStart := 0
	if d.cursor >= height-1 {
		visibleStart = d.cursor - height + 2
	}

	for i := visibleStart; i < len(d.stateItems) && i < visibleStart+height; i++ {
		item := d.stateItems[i]
		if len(item) > width-6 {
			item = item[:width-6]
		}
		if i == d.cursor {
			lines = append(lines, unitCursorStyle.Render("  ▸ ")+unitSelectedStyle.Render(item))
		} else {
			lines = append(lines, unitNormalStyle.Render("    "+item))
		}
	}

	return lines
}
