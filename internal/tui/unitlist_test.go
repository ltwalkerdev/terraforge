package tui

import (
	"testing"

	"github.com/ltwalkerdev/terraforge/internal/model"
)

func testUnits() []model.Unit {
	return []model.Unit{
		{Name: "networking", Dependencies: nil, Dependents: []string{"vpc"}},
		{Name: "vpc", Dependencies: []string{"networking"}, Dependents: []string{"rds", "ecs"}},
		{Name: "rds", Dependencies: []string{"vpc"}, Dependents: []string{"app"}},
		{Name: "ecs", Dependencies: []string{"vpc"}, Dependents: []string{"app"}},
		{Name: "app", Dependencies: []string{"rds", "ecs"}, Dependents: nil},
	}
}

func TestSetUnitsStartsEmpty(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	sel, total := ul.SelectionCount()
	if sel != 0 {
		t.Errorf("expected 0 selected, got %d", sel)
	}
	if total != 5 {
		t.Errorf("expected 5 total, got %d", total)
	}
}

func TestCycleEllipsisDeps(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 4 // app
	ul.CycleEllipsis()

	if ul.ellipsisAnchor != 4 {
		t.Fatalf("expected anchor=4, got %d", ul.ellipsisAnchor)
	}
	if ul.ellipsisMode != model.FilterWithDeps {
		t.Fatalf("expected FilterWithDeps, got %d", ul.ellipsisMode)
	}
	// app depends on rds, ecs; rds depends on vpc; ecs depends on vpc; vpc depends on networking
	// so all 5 should be selected
	sel, _ := ul.SelectionCount()
	if sel != 5 {
		t.Errorf("expected 5 selected (full transitive deps of app), got %d", sel)
	}
}

func TestCycleEllipsisDependents(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 0 // networking
	ul.CycleEllipsis()  // -> FilterWithDeps (networking has no deps, just itself)
	ul.CycleEllipsis()  // -> FilterWithDependents

	if ul.ellipsisMode != model.FilterWithDependents {
		t.Fatalf("expected FilterWithDependents, got %d", ul.ellipsisMode)
	}
	// networking's dependents: vpc -> rds, ecs -> app (all 5)
	sel, _ := ul.SelectionCount()
	if sel != 5 {
		t.Errorf("expected 5 selected (full transitive dependents of networking), got %d", sel)
	}
}

func TestCycleEllipsisResetsToAll(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 2 // rds
	ul.CycleEllipsis()  // FilterWithDeps
	ul.CycleEllipsis()  // FilterWithDependents
	ul.CycleEllipsis()  // reset -> all selected

	if ul.ellipsisAnchor != -1 {
		t.Errorf("expected anchor cleared, got %d", ul.ellipsisAnchor)
	}
	sel, total := ul.SelectionCount()
	if sel != total {
		t.Errorf("expected all %d selected, got %d", total, sel)
	}
}

func TestCycleEllipsisDifferentUnitReplacesAnchor(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 2 // rds
	ul.CycleEllipsis()

	ul.cursor = 0 // networking
	ul.CycleEllipsis()

	if ul.ellipsisAnchor != 0 {
		t.Errorf("expected anchor=0, got %d", ul.ellipsisAnchor)
	}
}

func TestToggleSelectClearsAnchor(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 2
	ul.CycleEllipsis()
	ul.ToggleSelect()

	if ul.ellipsisAnchor != -1 {
		t.Errorf("expected anchor cleared after toggle, got %d", ul.ellipsisAnchor)
	}
}

func TestFilterSubstring(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.StartFilter()
	ul.SetFilter("vpc")

	visible := ul.visibleIndices()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible, got %d", len(visible))
	}
	if ul.units[visible[0]].Name != "vpc" {
		t.Errorf("expected vpc, got %s", ul.units[visible[0]].Name)
	}
}

func TestFilterCaseInsensitive(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.SetFilter("VPC")

	visible := ul.visibleIndices()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible, got %d", len(visible))
	}
}

func TestFilterClearRestoresAll(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.StartFilter()
	ul.SetFilter("vpc")
	ul.StopFilter(false)

	visible := ul.visibleIndices()
	if len(visible) != 5 {
		t.Errorf("expected 5 visible after clear, got %d", len(visible))
	}
}

func TestBuildFilterWithAnchor(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 1 // vpc
	ul.CycleEllipsis()

	expr := ul.BuildFilter()
	if len(expr.Include) != 1 {
		t.Fatalf("expected 1 filter item, got %d", len(expr.Include))
	}
	if expr.Include[0].UnitName != "vpc" {
		t.Errorf("expected vpc, got %s", expr.Include[0].UnitName)
	}
	if expr.Include[0].Mode != model.FilterWithDeps {
		t.Errorf("expected FilterWithDeps mode")
	}
}

func TestHasSelectionBlocksEmpty(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	if ul.HasSelection() {
		t.Error("expected no selection on fresh unit list")
	}
}

func TestExportImportSelection(t *testing.T) {
	ul := newUnitList(nil)
	ul.SetUnits(testUnits())
	ul.cursor = 1
	ul.ToggleSelect()
	ul.cursor = 3
	ul.ToggleSelect()

	exported := ul.ExportSelection()
	if len(exported) != 2 {
		t.Fatalf("expected 2, got %d", len(exported))
	}

	ul2 := newUnitList(nil)
	ul2.SetUnits(testUnits())
	ul2.ImportSelection(exported)

	sel, _ := ul2.SelectionCount()
	if sel != 2 {
		t.Errorf("expected 2 after import, got %d", sel)
	}
}
