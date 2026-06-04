package model

type Stack struct {
	Name  string
	Dir   string
	Units []Unit
}

type Unit struct {
	Name         string
	Path         string
	RelPath      string
	Dependencies []string
	Dependents   []string
	Status       UnitStatus
	IsStack      bool
	Children     []string
	IncludeOnly  bool
}

type UnitStatus int

const (
	StatusUnknown UnitStatus = iota
	StatusClean
	StatusChanged
	StatusRunning
	StatusError
)

func (s UnitStatus) String() string {
	switch s {
	case StatusClean:
		return "clean"
	case StatusChanged:
		return "changed"
	case StatusRunning:
		return "running"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}
