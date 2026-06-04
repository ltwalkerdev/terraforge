package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ltwalkerdev/terraforge/internal/model"
)

var depHintStyle = lipgloss.NewStyle().Foreground(dimGrey).Italic(true)

type unitList struct {
	units          []model.Unit
	cursor         int
	selected       map[int]model.FilterMode
	ellipsisAnchor int
	ellipsisMode   model.FilterMode
	filter         string
	filtering      bool
	filtered       []int // indices into units that match filter
}

func newUnitList(units []model.Unit) unitList {
	return unitList{
		units:          units,
		cursor:         0,
		selected:       make(map[int]model.FilterMode),
		ellipsisAnchor: -1,
	}
}

func (u *unitList) SetUnits(units []model.Unit) {
	sort.SliceStable(units, func(i, j int) bool {
		if units[i].IncludeOnly != units[j].IncludeOnly {
			return !units[i].IncludeOnly
		}
		return false
	})
	u.units = units
	u.cursor = 0
	u.ellipsisAnchor = -1
	u.selected = make(map[int]model.FilterMode)
	u.filter = ""
	u.filtering = false
	u.filtered = nil
}

func (u *unitList) SelectionCount() (int, int) {
	return len(u.selected), len(u.units)
}

func (u *unitList) HasSelection() bool {
	return len(u.selected) > 0
}

func (u *unitList) ExportSelection() map[string]string {
	result := make(map[string]string)
	for idx, mode := range u.selected {
		if idx < len(u.units) {
			var modeStr string
			switch mode {
			case model.FilterExact:
				modeStr = "exact"
			case model.FilterWithDeps:
				modeStr = "with_deps"
			case model.FilterWithDependents:
				modeStr = "with_dependents"
			}
			result[u.units[idx].Name] = modeStr
		}
	}
	return result
}

func (u *unitList) ImportSelection(sel map[string]string) {
	u.selected = make(map[int]model.FilterMode)
	nameIdx := u.nameIndex()
	for name, modeStr := range sel {
		if idx, ok := nameIdx[name]; ok {
			var mode model.FilterMode
			switch modeStr {
			case "exact":
				mode = model.FilterExact
			case "with_deps":
				mode = model.FilterWithDeps
			case "with_dependents":
				mode = model.FilterWithDependents
			default:
				mode = model.FilterExact
			}
			u.selected[idx] = mode
		}
	}
}

func (u *unitList) AnchorName() string {
	if u.ellipsisAnchor >= 0 && u.ellipsisAnchor < len(u.units) {
		return u.units[u.ellipsisAnchor].Name
	}
	return ""
}

func (u *unitList) SetAnchor(name string, mode model.FilterMode) {
	nameIdx := u.nameIndex()
	if idx, ok := nameIdx[name]; ok {
		u.ellipsisAnchor = idx
		u.ellipsisMode = mode
	}
}

func (u *unitList) StartFilter() {
	u.filtering = true
	u.filter = ""
	u.applyFilter()
}

func (u *unitList) StopFilter(keepFilter bool) {
	u.filtering = false
	if !keepFilter {
		u.filter = ""
		u.filtered = nil
	}
}

func (u *unitList) SetFilter(s string) {
	u.filter = s
	u.applyFilter()
	if len(u.filtered) > 0 {
		u.cursor = u.filtered[0]
	}
}

func (u *unitList) applyFilter() {
	if u.filter == "" {
		u.filtered = nil
		return
	}
	lower := strings.ToLower(u.filter)
	u.filtered = nil
	for i, unit := range u.units {
		if strings.Contains(strings.ToLower(unit.Name), lower) {
			u.filtered = append(u.filtered, i)
		}
	}
}

func (u *unitList) IsFiltering() bool {
	return u.filtering
}

func (u *unitList) HasFilter() bool {
	return u.filter != ""
}

func (u *unitList) ClearFilter() {
	u.filter = ""
	u.filtered = nil
}

func (u *unitList) FilterText() string {
	return u.filter
}

func (u *unitList) visibleIndices() []int {
	if u.filtered != nil {
		return u.filtered
	}
	indices := make([]int, len(u.units))
	for i := range u.units {
		indices[i] = i
	}
	return indices
}

func (u *unitList) cursorInVisible() int {
	visible := u.visibleIndices()
	for vi, idx := range visible {
		if idx == u.cursor {
			return vi
		}
	}
	return 0
}

func (u *unitList) Up() {
	visible := u.visibleIndices()
	if len(visible) == 0 {
		return
	}
	vi := u.cursorInVisible()
	if vi > 0 {
		u.cursor = visible[vi-1]
	}
}

func (u *unitList) Down() {
	visible := u.visibleIndices()
	if len(visible) == 0 {
		return
	}
	vi := u.cursorInVisible()
	if vi < len(visible)-1 {
		u.cursor = visible[vi+1]
	}
}

func (u *unitList) GoToTop() {
	visible := u.visibleIndices()
	if len(visible) > 0 {
		u.cursor = visible[0]
	}
}

func (u *unitList) GoToBottom() {
	visible := u.visibleIndices()
	if len(visible) > 0 {
		u.cursor = visible[len(visible)-1]
	}
}

func (u *unitList) HalfPageUp(viewHeight int) {
	visible := u.visibleIndices()
	if len(visible) == 0 {
		return
	}
	vi := u.cursorInVisible()
	jump := viewHeight / 2
	vi -= jump
	if vi < 0 {
		vi = 0
	}
	u.cursor = visible[vi]
}

func (u *unitList) HalfPageDown(viewHeight int) {
	visible := u.visibleIndices()
	if len(visible) == 0 {
		return
	}
	vi := u.cursorInVisible()
	jump := viewHeight / 2
	vi += jump
	if vi >= len(visible) {
		vi = len(visible) - 1
	}
	u.cursor = visible[vi]
}

func (u *unitList) UpdateStatus(name string, status model.UnitStatus) {
	for i := range u.units {
		if u.units[i].Name == name {
			u.units[i].Status = status
			return
		}
	}
}

func (u *unitList) ToggleSelect() {
	if len(u.units) == 0 {
		return
	}
	if u.units[u.cursor].IncludeOnly {
		return
	}
	if u.ellipsisAnchor >= 0 {
		u.ellipsisAnchor = -1
	}
	if _, ok := u.selected[u.cursor]; ok {
		delete(u.selected, u.cursor)
	} else {
		u.selected[u.cursor] = model.FilterExact
	}
}

func (u *unitList) CycleEllipsis() {
	if len(u.units) == 0 {
		return
	}
	if u.units[u.cursor].IncludeOnly {
		return
	}

	if u.ellipsisAnchor < 0 {
		u.ellipsisAnchor = u.cursor
		u.ellipsisMode = model.FilterWithDeps
		u.applyEllipsisSelection()
		return
	}

	if u.cursor != u.ellipsisAnchor {
		u.ellipsisAnchor = u.cursor
		u.ellipsisMode = model.FilterWithDeps
		u.applyEllipsisSelection()
		return
	}

	switch u.ellipsisMode {
	case model.FilterWithDeps:
		u.ellipsisMode = model.FilterWithDependents
		u.applyEllipsisSelection()
	case model.FilterWithDependents:
		u.ellipsisAnchor = -1
		u.SelectAll()
	}
}

func (u *unitList) applyEllipsisSelection() {
	u.selected = make(map[int]model.FilterMode)
	nameIdx := u.nameIndex()

	var resolved []int
	if u.ellipsisMode == model.FilterWithDeps {
		resolved = u.resolveDeps(u.ellipsisAnchor, nameIdx)
	} else {
		resolved = u.resolveDependents(u.ellipsisAnchor, nameIdx)
	}

	for _, idx := range resolved {
		u.selected[idx] = model.FilterExact
	}
	u.selected[u.ellipsisAnchor] = u.ellipsisMode
}

func (u *unitList) nameIndex() map[string]int {
	m := make(map[string]int, len(u.units))
	for i, unit := range u.units {
		m[unit.Name] = i
	}
	return m
}

func (u *unitList) resolveDeps(anchorIdx int, nameIdx map[string]int) []int {
	visited := make(map[int]bool)
	var walk func(idx int)
	walk = func(idx int) {
		if visited[idx] {
			return
		}
		visited[idx] = true
		for _, dep := range u.units[idx].Dependencies {
			if depIdx, ok := nameIdx[dep]; ok {
				walk(depIdx)
			}
		}
	}
	walk(anchorIdx)
	result := make([]int, 0, len(visited))
	for idx := range visited {
		result = append(result, idx)
	}
	return result
}

func (u *unitList) resolveDependents(anchorIdx int, nameIdx map[string]int) []int {
	visited := make(map[int]bool)
	var walk func(idx int)
	walk = func(idx int) {
		if visited[idx] {
			return
		}
		visited[idx] = true
		for _, dep := range u.units[idx].Dependents {
			if depIdx, ok := nameIdx[dep]; ok {
				walk(depIdx)
			}
		}
	}
	walk(anchorIdx)
	result := make([]int, 0, len(visited))
	for idx := range visited {
		result = append(result, idx)
	}
	return result
}

func (u *unitList) SelectAll() {
	u.ellipsisAnchor = -1
	for i := range u.units {
		if !u.units[i].IncludeOnly {
			u.selected[i] = model.FilterExact
		}
	}
}

func (u *unitList) SelectNone() {
	u.ellipsisAnchor = -1
	u.selected = make(map[int]model.FilterMode)
}

func (u *unitList) CurrentUnit() *model.Unit {
	if len(u.units) == 0 {
		return nil
	}
	return &u.units[u.cursor]
}

func (u *unitList) BuildFilter() model.FilterExpr {
	if u.ellipsisAnchor >= 0 && u.ellipsisAnchor < len(u.units) {
		return model.FilterExpr{
			Include: []model.FilterItem{{
				UnitName: u.units[u.ellipsisAnchor].Name,
				Mode:     u.ellipsisMode,
			}},
		}
	}

	if len(u.selected) == len(u.units) {
		allExact := true
		for _, mode := range u.selected {
			if mode != model.FilterExact {
				allExact = false
				break
			}
		}
		if allExact {
			return model.FilterExpr{}
		}
	}

	var expr model.FilterExpr
	for idx, mode := range u.selected {
		if idx < len(u.units) {
			expr.Include = append(expr.Include, model.FilterItem{
				UnitName: u.units[idx].Name,
				Mode:     mode,
			})
		}
	}
	return expr
}

func (u *unitList) depHint(i int) string {
	if u.ellipsisAnchor < 0 {
		return ""
	}
	unit := u.units[i]
	if u.ellipsisAnchor == i {
		if u.ellipsisMode == model.FilterWithDeps && len(unit.Dependencies) > 0 {
			return depHintStyle.Render(" → " + strings.Join(unit.Dependencies, ", "))
		}
		if u.ellipsisMode == model.FilterWithDependents && len(unit.Dependents) > 0 {
			return depHintStyle.Render(" ← " + strings.Join(unit.Dependents, ", "))
		}
	}
	return ""
}

func (u *unitList) View(width, height int) string {
	visible := u.visibleIndices()
	if len(visible) == 0 {
		if u.filter != "" {
			return unitNormalStyle.Render("  no matches")
		}
		return unitNormalStyle.Render("  no units found — press S to generate stack")
	}

	// Find cursor position in the visible list
	cursorVi := 0
	for vi, idx := range visible {
		if idx == u.cursor {
			cursorVi = vi
			break
		}
	}

	var lines []string
	visibleStart := 0
	if cursorVi >= height-2 {
		visibleStart = cursorVi - height + 3
	}

	separatorShown := false
	for vi := visibleStart; vi < len(visible) && vi < visibleStart+height-1; vi++ {
		i := visible[vi]
		unit := u.units[i]

		if unit.IncludeOnly && !separatorShown {
			separatorShown = true
			lines = append(lines, depHintStyle.Render("  ── includes ──"))
		}

		cursor := "  "
		if i == u.cursor {
			cursor = unitCursorStyle.Render("▸ ")
		}

		if unit.IncludeOnly {
			name := unit.Name
			if i == u.cursor {
				name = unitCursorStyle.Render(name)
			} else {
				name = depHintStyle.Render(name)
			}
			line := fmt.Sprintf("%s    %s", cursor, name)
			lines = append(lines, line)
			continue
		}

		checkbox := "[ ]"
		if mode, ok := u.selected[i]; ok {
			switch mode {
			case model.FilterExact:
				checkbox = unitSelectedStyle.Render("[x]")
			case model.FilterWithDeps:
				checkbox = unitSelectedStyle.Render("[▸]")
			case model.FilterWithDependents:
				checkbox = unitSelectedStyle.Render("[◂]")
			}
		}

		name := unit.Name
		if unit.IsStack {
			suffix := " [stack]"
			if len(unit.Children) > 0 {
				suffix += " (" + strings.Join(unit.Children, ", ") + ")"
			}
			name += suffix
		}
		if i == u.cursor {
			name = unitCursorStyle.Render(name)
		} else if _, ok := u.selected[i]; ok {
			name = unitSelectedStyle.Render(name)
		} else {
			name = unitNormalStyle.Render(name)
		}

		hint := u.depHint(i)
		status := renderStatus(unit.Status)
		line := fmt.Sprintf("%s%s %s%s %s", cursor, checkbox, name, hint, status)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func renderStatus(s model.UnitStatus) string {
	switch s {
	case model.StatusClean:
		return statusCleanStyle.Render("● clean")
	case model.StatusChanged:
		return statusChangedStyle.Render("● changed")
	case model.StatusRunning:
		return purpleRunningStyle.Render("◍ running")
	case model.StatusError:
		return statusErrorStyle.Render("● error")
	default:
		return statusUnknownStyle.Render("○")
	}
}
