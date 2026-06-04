package model

import "fmt"

type FilterMode int

const (
	FilterExact            FilterMode = iota
	FilterWithDeps                    // unit... = target + its dependencies (upstream)
	FilterWithDependents              // ...unit = target + its dependents (downstream)
)

type FilterItem struct {
	UnitName string
	Mode     FilterMode
}

func (f FilterItem) String() string {
	switch f.Mode {
	case FilterWithDeps:
		return fmt.Sprintf("%s...", f.UnitName)
	case FilterWithDependents:
		return fmt.Sprintf("...%s", f.UnitName)
	default:
		return f.UnitName
	}
}

type FilterExpr struct {
	Include []FilterItem
	Exclude []FilterItem
}

func (fe FilterExpr) ToFlags() []string {
	var flags []string
	for _, item := range fe.Include {
		flags = append(flags, "--filter", item.String())
	}
	for _, item := range fe.Exclude {
		flags = append(flags, "--filter", "!"+item.String())
	}
	return flags
}
