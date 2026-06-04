package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit           key.Binding
	Help           key.Binding
	Tab            key.Binding
	ShiftTab       key.Binding
	Enter          key.Binding
	Back           key.Binding
	Up             key.Binding
	Down           key.Binding
	Top            key.Binding
	Bottom         key.Binding
	HalfPageUp     key.Binding
	HalfPageDown   key.Binding
	NextStack      key.Binding
	PrevStack      key.Binding
	Space          key.Binding
	Plan           key.Binding
	Apply          key.Binding
	ApplyUpdate    key.Binding
	Destroy        key.Binding
	Bootstrap      key.Binding
	SelectAll      key.Binding
	SelectNone     key.Binding
	EllipsisToggle key.Binding
	CancelCmd      key.Binding
	Search         key.Binding
	ViewFiles      key.Binding
	StackGenerate  key.Binding
	StackClean     key.Binding
	StackInit      key.Binding
	Init           key.Binding
	CleanCache     key.Binding
	StateList      key.Binding
	StateRm        key.Binding
	StateShow      key.Binding
	Output         key.Binding
	Validate       key.Binding
	Refresh        key.Binding
	Replace        key.Binding
	Render         key.Binding
	Import         key.Binding
	ForceUnlock    key.Binding
	StateMv        key.Binding
	TabNext          key.Binding
	TabPrev          key.Binding
	OutputFullscreen key.Binding
	ToggleCLI        key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next stack"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev stack"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "drill in"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "h"),
		key.WithHelp("esc/h", "back"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j/↓", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("gg", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	HalfPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	NextStack: key.NewBinding(
		key.WithKeys("L", "tab"),
		key.WithHelp("L/tab", "next stack"),
	),
	PrevStack: key.NewBinding(
		key.WithKeys("H", "shift+tab"),
		key.WithHelp("H/S-tab", "prev stack"),
	),
	Space: key.NewBinding(
		key.WithKeys(" ", "v"),
		key.WithHelp("space/v", "toggle select"),
	),
	Plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	ApplyUpdate: key.NewBinding(
		key.WithKeys("W"),
		key.WithHelp("W", "apply --source-update"),
	),
	Destroy: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "destroy"),
	),
	Bootstrap: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "toggle bootstrap"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "select all"),
	),
	SelectNone: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "select none"),
	),
	EllipsisToggle: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "cycle ellipsis"),
	),
	CancelCmd: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "cancel command"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	ViewFiles: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "view files"),
	),
	// Stack lifecycle
	StackGenerate: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "stack generate"),
	),
	StackClean: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "stack clean"),
	),
	StackInit: key.NewBinding(
		key.WithKeys("I"),
		key.WithHelp("I", "stack init"),
	),
	// Unit detail keys
	Init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	CleanCache: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "clean cache"),
	),
	StateList: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "state list"),
	),
	StateRm: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "state rm (selected)"),
	),
	StateShow: key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "state show"),
	),
	Output: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "outputs"),
	),
	Validate: key.NewBinding(
		key.WithKeys("V"),
		key.WithHelp("V", "validate"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh state"),
	),
	Replace: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "replace resource"),
	),
	Render: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "render config"),
	),
	Import: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "import resource"),
	),
	ForceUnlock: key.NewBinding(
		key.WithKeys("U"),
		key.WithHelp("U", "force-unlock"),
	),
	StateMv: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "state mv"),
	),
	TabNext: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	TabPrev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("S-tab", "prev tab"),
	),
	OutputFullscreen: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "fullscreen output"),
	),
	ToggleCLI: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "show CLI command"),
	),
}
