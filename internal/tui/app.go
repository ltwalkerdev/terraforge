package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ltwalkerdev/terraforge/internal/config"
	"github.com/ltwalkerdev/terraforge/internal/discovery"
	"github.com/ltwalkerdev/terraforge/internal/model"
	"github.com/ltwalkerdev/terraforge/internal/runner"
	"github.com/ltwalkerdev/terraforge/internal/state"
)

type view int

const (
	viewSplash view = iota
	viewUnitList
	viewUnitDetail
)

var banner = " _____                    __\n" +
	"|_   _|__ _ __ _ __ __ _ / _| ___  _ __ __ _  ___\n" +
	"  | |/ _ \\ '__| '__/ _` | |_ / _ \\| '__/ _` |/ _ \\\n" +
	"  | |  __/ |  | | | (_| |  _| (_) | | | (_| |  __/\n" +
	"  |_|\\___|_|  |_|  \\__,_|_|  \\___/|_|  \\__, |\\___|\n" +
	"                                        |___/"

type Model struct {
	stacks           []model.Stack
	stackBar         stackBar
	unitList         unitList
	unitDetail       unitDetail
	output           outputPane
	help             helpPopup
	runner           *runner.Runner
	width            int
	height           int
	active           view
	running          bool
	workspace        string
	lastKeyTime      time.Time
	lastKey          string
	confirm          *confirmAction
	input            textInput
	viewer           fileViewer
	filePicker       filePicker
	outputFullscreen bool
	showCLI          bool
	lastCLI          string
	cfg              config.Config
	parentUnits      []model.Unit
}

type confirmAction struct {
	action string
	target string
}

type stackUnitsMsg struct {
	units []model.Unit
}

type rescanMsg struct{}
type splashDoneMsg struct{}

func NewModel(stacks []model.Stack, r *runner.Runner, workspace string) Model {
	sb := newStackBar(stacks)
	cfg := config.Load()
	return Model{
		stacks:     stacks,
		stackBar:   sb,
		unitList:   newUnitList(nil),
		unitDetail: newUnitDetail(),
		output:     newOutputPane(),
		help:       newHelpPopup(),
		runner:     r,
		workspace:  workspace,
		active:     viewSplash,
		cfg:        cfg,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadUnits(),
		tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return splashDoneMsg{}
		}),
	)
}

func (m *Model) loadUnits() tea.Cmd {
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	s := *stack
	return func() tea.Msg {
		return stackUnitsMsg{units: s.Units}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		outH := m.outputHeight()
		if outH < 3 {
			outH = 3
		}
		m.output.height = outH
		return m, nil

	case stackUnitsMsg:
		m.unitList.SetUnits(msg.units)
		m.restoreState()
		return m, nil

	case splashDoneMsg:
		if m.active == viewSplash {
			m.active = viewUnitList
		}
		return m, nil

	case rescanMsg:
		stack := m.stackBar.Active()
		if stack != nil {
			units, _ := discovery.FindUnits(*stack)
			for j := range units {
				units[j].Dependencies = discovery.ParseDependencies(units[j])
			}
			nameIdx := make(map[string]int, len(units))
			for j := range units {
				nameIdx[units[j].Name] = j
			}
			for j := range units {
				for _, dep := range units[j].Dependencies {
					if idx, ok := nameIdx[dep]; ok {
						units[idx].Dependents = append(units[idx].Dependents, units[j].Name)
					}
				}
			}
			stack.Units = units
			m.stacks[m.stackBar.active] = *stack
			m.unitList.SetUnits(units)
		}
		return m, nil

	case runner.OutputLineMsg:
		m.output.AddLine(msg.Line, msg.IsErr)
		return m, nil

	case runner.UnitStatusMsg:
		m.unitList.UpdateStatus(msg.UnitName, msg.Status)
		return m, nil

	case runner.ConfirmNeededMsg:
		m.input.Start(inputStdinConfirm, "Enter a value")
		return m, nil

	case runner.CommandDoneMsg:
		m.running = false
		if msg.Err != nil && !m.runner.Cancelled() {
			m.output.AddLine(fmt.Sprintf("command failed: %v", msg.Err), true)
		} else if msg.Err == nil {
			m.output.AddLine("command completed successfully", false)
		}
		if stack := m.stackBar.Active(); stack != nil {
			state.SaveLog(stack.Name, m.output.Lines(), m.cfg.LogLines)
		}
		return m, nil

	case runner.StateListMsg:
		m.running = false
		if msg.Err != nil {
			m.output.AddLine(fmt.Sprintf("state list failed: %v", msg.Err), true)
		} else {
			m.unitDetail.SetStateItems(msg.Items)
			m.output.AddLine(fmt.Sprintf("loaded %d resources", len(msg.Items)), false)
		}
		return m, nil

	case runner.StateShowMsg:
		m.running = false
		if msg.Err != nil {
			m.output.AddLine(fmt.Sprintf("state show failed: %v", msg.Err), true)
		} else {
			m.output.Clear()
			for _, line := range strings.Split(msg.Output, "\n") {
				m.output.AddLine(line, false)
			}
		}
		return m, nil

	case runner.OutputsMsg:
		m.running = false
		if msg.Err != nil {
			m.output.AddLine(fmt.Sprintf("outputs failed: %v", msg.Err), true)
		} else {
			m.output.Clear()
			for _, line := range strings.Split(msg.Output, "\n") {
				m.output.AddLine(line, false)
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.active == viewSplash {
			m.active = viewUnitList
			return m, nil
		}

		if m.outputFullscreen {
			return m.updateOutputFullscreen(msg)
		}

		if m.viewer.visible {
			return m.updateViewer(msg)
		}

		if m.filePicker.visible {
			return m.updateFilePicker(msg)
		}

		if m.help.visible {
			if key.Matches(msg, keys.Help) || key.Matches(msg, keys.Back) || msg.String() == "q" {
				m.help.Toggle()
			}
			return m, nil
		}

		if m.input.Active() {
			return m.handleInput(msg)
		}

		if m.confirm != nil {
			return m.handleConfirm(msg)
		}

		if m.active == viewUnitDetail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}

	return m, nil
}

func (m Model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	wasStdinConfirm := m.input.mode == inputStdinConfirm
	val, submitted := m.input.Update(msg)
	if !submitted {
		if wasStdinConfirm && !m.input.Active() {
			m.runner.Cancel()
			m.output.AddLine("cancelled", false)
		}
		return m, nil
	}

	stack := m.stackBar.Active()
	if stack == nil {
		m.input.Cancel()
		return m, nil
	}

	switch m.input.mode {
	case inputImportAddr:
		m.input.stash = val
		m.input.Start(inputImportID, "resource ID")
		return m, nil
	case inputImportID:
		addr := m.input.stash
		m.input.Cancel()
		m.running = true
		m.output.Clear()
		m.output.AddLine(fmt.Sprintf("running: terraform import %s %s", addr, val), false)
		return m, m.runner.UnitImport(*stack, m.unitDetail.unit, addr, val)
	case inputReplace:
		m.input.Cancel()
		m.running = true
		m.output.Clear()
		m.output.AddLine(fmt.Sprintf("running: terraform apply -replace=%s", val), false)
		return m, m.runner.UnitReplace(*stack, m.unitDetail.unit, val)
	case inputForceUnlock:
		m.input.Cancel()
		m.running = true
		m.output.Clear()
		m.output.AddLine(fmt.Sprintf("running: terraform force-unlock %s", val), false)
		return m, m.runner.UnitForceUnlock(*stack, m.unitDetail.unit, val)
	case inputStateMvSrc:
		m.input.stash = val
		m.input.Start(inputStateMvDst, "destination address")
		return m, nil
	case inputStateMvDst:
		src := m.input.stash
		m.input.Cancel()
		m.running = true
		m.output.Clear()
		m.output.AddLine(fmt.Sprintf("running: terraform state mv %s %s", src, val), false)
		return m, m.runner.UnitStateMv(*stack, m.unitDetail.unit, src, val)
	case inputStdinConfirm:
		m.input.Cancel()
		if val != "yes" && val != "y" {
			m.runner.MarkCancelled()
		}
		m.runner.Confirm(val)
		return m, nil
	}

	m.input.Cancel()
	return m, nil
}

func (m Model) handleConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		action := m.confirm
		m.confirm = nil
		switch action.action {
		case "state_rm":
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				return m, m.runner.UnitStateRm(*stack, m.unitDetail.unit, action.target)
			}
		case "quit":
			m.runner.Cancel()
			m.saveState()
			return m, tea.Quit
		}
	case "n", "N", "esc", "q":
		m.confirm = nil
		m.output.AddLine("cancelled", false)
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.unitList.IsFiltering() {
		switch msg.String() {
		case "esc":
			m.unitList.StopFilter(false)
		case "enter":
			m.unitList.StopFilter(true)
		case "backspace":
			f := m.unitList.FilterText()
			if len(f) > 0 {
				m.unitList.SetFilter(f[:len(f)-1])
			}
		default:
			if len(msg.String()) == 1 {
				m.unitList.SetFilter(m.unitList.FilterText() + msg.String())
			}
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Back):
		if m.unitList.HasFilter() {
			m.unitList.ClearFilter()
			return m, nil
		}
		if m.parentUnits != nil {
			m.unitList.SetUnits(m.parentUnits)
			m.parentUnits = nil
			return m, nil
		}

	case key.Matches(msg, keys.Quit):
		if m.running {
			m.confirm = &confirmAction{action: "quit"}
			m.output.AddLine("command running — quit? (y/n)", true)
			return m, nil
		}
		m.runner.Cancel()
		m.saveState()
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.help.Toggle()
		return m, nil

	case key.Matches(msg, keys.NextStack):
		m.stackBar.Next()
		return m, m.refreshUnits()

	case key.Matches(msg, keys.PrevStack):
		m.stackBar.Prev()
		return m, m.refreshUnits()

	case key.Matches(msg, keys.Top):
		now := time.Now()
		if m.lastKey == "g" && now.Sub(m.lastKeyTime) < 500*time.Millisecond {
			m.unitList.GoToTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
			m.lastKeyTime = now
		}
		return m, nil

	case key.Matches(msg, keys.Bottom):
		m.unitList.GoToBottom()
		return m, nil

	case key.Matches(msg, keys.HalfPageUp):
		m.unitList.HalfPageUp(m.panelHeight())
		return m, nil

	case key.Matches(msg, keys.HalfPageDown):
		m.unitList.HalfPageDown(m.panelHeight())
		return m, nil

	case key.Matches(msg, keys.Up):
		m.unitList.Up()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.unitList.Down()
		return m, nil

	case key.Matches(msg, keys.Space):
		m.unitList.ToggleSelect()
		m.lastCLI = ""
		return m, nil

	case key.Matches(msg, keys.EllipsisToggle):
		m.unitList.CycleEllipsis()
		m.lastCLI = ""
		return m, nil

	case key.Matches(msg, keys.SelectAll):
		m.unitList.SelectAll()
		m.lastCLI = ""
		return m, nil

	case key.Matches(msg, keys.SelectNone):
		m.unitList.SelectNone()
		m.lastCLI = ""
		return m, nil

	case key.Matches(msg, keys.Bootstrap):
		return m, m.runBackendBootstrap()

	case key.Matches(msg, keys.Plan):
		return m, m.runStackCmd("plan")

	case key.Matches(msg, keys.Apply):
		return m, m.runStackCmd("apply")

	case key.Matches(msg, keys.ApplyUpdate):
		return m, m.runStackCmdWithFlags("apply", []string{"--source-update"})

	case key.Matches(msg, keys.Destroy):
		return m, m.runStackCmd("destroy")

	case key.Matches(msg, keys.StackGenerate):
		return m, m.runStackGenerate()

	case key.Matches(msg, keys.StackClean):
		return m, m.runStackLifecycle("clean")

	case key.Matches(msg, keys.StackInit):
		return m, m.runStackCmd("init")

	case key.Matches(msg, keys.CancelCmd):
		m.runner.Cancel()
		m.running = false
		m.output.AddLine("command cancelled", true)
		return m, nil

	case key.Matches(msg, keys.Enter):
		unit := m.unitList.CurrentUnit()
		stack := m.stackBar.Active()
		if unit != nil && stack != nil {
			if unit.IsStack {
				childUnits, _ := discovery.FindUnits(model.Stack{Name: unit.Name, Dir: unit.Path})
				if len(childUnits) > 0 {
					m.parentUnits = m.unitList.units
					m.unitList.SetUnits(childUnits)
					m.output.Clear()
					return m, nil
				}
			}
			m.active = viewUnitDetail
			m.unitDetail.SetUnit(*unit, *stack)
			m.output.Clear()
		}
		return m, nil

	case key.Matches(msg, keys.ViewFiles):
		stack := m.stackBar.Active()
		if stack != nil {
			files := listStackFiles(stack.Dir)
			if len(files) > 0 {
				m.filePicker.Open(files, stack.Dir)
			}
		}
		return m, nil

	case key.Matches(msg, keys.OutputFullscreen):
		m.outputFullscreen = !m.outputFullscreen
		return m, nil

	case key.Matches(msg, keys.ToggleCLI):
		m.showCLI = !m.showCLI
		return m, nil

	case key.Matches(msg, keys.Search):
		m.unitList.StartFilter()
		return m, nil
	}

	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
		m.active = viewUnitList
		m.output.Clear()
		return m, nil

	case key.Matches(msg, keys.Help):
		m.help.Toggle()
		return m, nil

	case key.Matches(msg, keys.TabNext):
		m.unitDetail.NextTab()
		return m, nil

	case key.Matches(msg, keys.TabPrev):
		m.unitDetail.PrevTab()
		return m, nil

	case key.Matches(msg, keys.Up):
		m.unitDetail.Up()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.unitDetail.Down()
		return m, nil

	case key.Matches(msg, keys.Top):
		now := time.Now()
		if m.lastKey == "g" && now.Sub(m.lastKeyTime) < 500*time.Millisecond {
			m.unitDetail.GoToTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
			m.lastKeyTime = now
		}
		return m, nil

	case key.Matches(msg, keys.Bottom):
		m.unitDetail.GoToBottom()
		return m, nil

	case key.Matches(msg, keys.HalfPageUp):
		m.unitDetail.HalfPageUp(m.panelHeight())
		return m, nil

	case key.Matches(msg, keys.HalfPageDown):
		m.unitDetail.HalfPageDown(m.panelHeight())
		return m, nil

	case key.Matches(msg, keys.Init):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt init", false)
				return m, m.runner.UnitRun(*stack, m.unitDetail.unit, "init")
			}
		}
		return m, nil

	case key.Matches(msg, keys.CleanCache):
		cacheDir := m.unitDetail.unit.Path + "/.terragrunt-cache"
		if err := os.RemoveAll(cacheDir); err != nil {
			m.output.AddLine(fmt.Sprintf("error cleaning cache: %v", err), true)
		} else {
			m.output.AddLine("cleaned .terragrunt-cache for "+m.unitDetail.unit.Name, false)
		}
		return m, nil

	case key.Matches(msg, keys.Plan):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt plan", false)
				return m, m.runner.UnitRun(*stack, m.unitDetail.unit, "plan")
			}
		}
		return m, nil

	case key.Matches(msg, keys.Apply):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt apply", false)
				return m, m.runner.UnitRun(*stack, m.unitDetail.unit, "apply")
			}
		}
		return m, nil

	case key.Matches(msg, keys.Destroy):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt destroy", false)
				return m, m.runner.UnitRun(*stack, m.unitDetail.unit, "destroy")
			}
		}
		return m, nil

	case key.Matches(msg, keys.StateList):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt state list", false)
				return m, m.runner.UnitStateList(*stack, m.unitDetail.unit)
			}
		}
		return m, nil

	case key.Matches(msg, keys.StateShow):
		if !m.running && m.unitDetail.tab == tabState {
			resource := m.unitDetail.SelectedResource()
			if resource != "" {
				stack := m.stackBar.Active()
				if stack != nil {
					m.running = true
					m.output.Clear()
					m.output.AddLine("running: terragrunt state show "+resource, false)
					return m, m.runner.UnitStateShow(*stack, m.unitDetail.unit, resource)
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.StateRm):
		if !m.running && m.unitDetail.tab == tabState {
			resource := m.unitDetail.SelectedResource()
			if resource != "" {
				m.confirm = &confirmAction{action: "state_rm", target: resource}
				m.output.AddLine(fmt.Sprintf("confirm state rm '%s'? (y/n)", resource), true)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Output):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt output", false)
				return m, m.runner.UnitOutputs(*stack, m.unitDetail.unit)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Validate):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terraform validate", false)
				return m, m.runner.UnitValidate(*stack, m.unitDetail.unit)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Refresh):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terraform apply -refresh-only", false)
				return m, m.runner.UnitRefresh(*stack, m.unitDetail.unit)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Render):
		if !m.running {
			stack := m.stackBar.Active()
			if stack != nil {
				m.running = true
				m.output.Clear()
				m.output.AddLine("running: terragrunt render", false)
				return m, m.runner.UnitRender(*stack, m.unitDetail.unit)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Replace):
		if !m.running {
			resource := m.unitDetail.SelectedResource()
			if resource != "" {
				m.input.Start(inputReplace, "replace address")
				m.input.value = resource
			} else {
				m.input.Start(inputReplace, "replace address")
			}
		}
		return m, nil

	case key.Matches(msg, keys.Import):
		if !m.running {
			m.input.Start(inputImportAddr, "resource address (e.g. aws_instance.foo)")
		}
		return m, nil

	case key.Matches(msg, keys.ForceUnlock):
		if !m.running {
			m.input.Start(inputForceUnlock, "lock ID")
		}
		return m, nil

	case key.Matches(msg, keys.StateMv):
		if !m.running {
			resource := m.unitDetail.SelectedResource()
			if resource != "" {
				m.input.Start(inputStateMvSrc, "source address")
				m.input.value = resource
			} else {
				m.input.Start(inputStateMvSrc, "source address")
			}
		}
		return m, nil

	case key.Matches(msg, keys.ViewFiles):
		stack := m.stackBar.Active()
		if stack != nil {
			files := listUnitFiles(m.unitDetail.unit.Path)
			if len(files) > 0 {
				m.filePicker.Open(files, stack.Dir)
			}
		}
		return m, nil

	case key.Matches(msg, keys.CancelCmd):
		m.runner.Cancel()
		m.running = false
		m.output.AddLine("command cancelled", true)
		return m, nil
	}

	return m, nil
}

func (m Model) updateOutputFullscreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.output.searching {
		switch msg.String() {
		case "esc":
			m.output.CancelSearch()
		case "enter":
			m.output.ConfirmSearch()
		case "backspace":
			q := m.output.searchQuery
			if len(q) > 0 {
				m.output.SetSearchQuery(q[:len(q)-1])
			}
		default:
			if len(msg.String()) == 1 {
				m.output.SetSearchQuery(m.output.searchQuery + msg.String())
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		if m.running {
			m.confirm = &confirmAction{action: "quit"}
			m.output.AddLine("command running — quit? (y/n)", true)
			m.outputFullscreen = false
			return m, nil
		}
		m.saveState()
		return m, tea.Quit
	case "F", "esc", "q":
		m.outputFullscreen = false
	case "/":
		m.output.StartSearch()
	case "n":
		m.output.NextMatch()
	case "N":
		m.output.PrevMatch()
	case "j", "down":
		m.output.ScrollDown()
	case "k", "up":
		m.output.ScrollUp()
	case "ctrl+d":
		m.output.HalfPageDown()
	case "ctrl+u":
		m.output.HalfPageUp()
	case "G":
		m.output.GoToBottom()
	case "g":
		now := time.Now()
		if m.lastKey == "g" && now.Sub(m.lastKeyTime) < 500*time.Millisecond {
			m.output.GoToTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
			m.lastKeyTime = now
		}
	case "ctrl+x":
		m.runner.Cancel()
		m.running = false
		m.output.AddLine("command cancelled", true)
	}
	return m, nil
}

func (m Model) updateViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.viewer.searching {
		switch msg.String() {
		case "esc":
			m.viewer.CancelSearch()
		case "enter":
			m.viewer.ConfirmSearch()
		case "backspace":
			q := m.viewer.searchQuery
			if len(q) > 0 {
				m.viewer.SetSearchQuery(q[:len(q)-1])
			}
		default:
			if len(msg.String()) == 1 {
				m.viewer.SetSearchQuery(m.viewer.searchQuery + msg.String())
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "esc", "h":
		m.viewer.Close()
	case "/":
		m.viewer.StartSearch()
	case "n":
		m.viewer.NextMatch()
	case "N":
		m.viewer.PrevMatch()
	case "j", "down":
		m.viewer.ScrollDown()
	case "k", "up":
		m.viewer.ScrollUp()
	case "ctrl+d":
		m.viewer.HalfPageDown()
	case "ctrl+u":
		m.viewer.HalfPageUp()
	case "G":
		m.viewer.GoToBottom()
	case "g":
		now := time.Now()
		if m.lastKey == "g" && now.Sub(m.lastKeyTime) < 500*time.Millisecond {
			m.viewer.GoToTop()
			m.lastKey = ""
		} else {
			m.lastKey = "g"
			m.lastKeyTime = now
		}
	}
	return m, nil
}

func (m Model) updateFilePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "h":
		m.filePicker.Close()
	case "j", "down":
		m.filePicker.Down()
	case "k", "up":
		m.filePicker.Up()
	case "enter", "l":
		path := m.filePicker.Selected()
		if path != "" {
			m.filePicker.Close()
			if err := m.viewer.Open(path); err != nil {
				m.output.AddLine(fmt.Sprintf("error opening file: %v", err), true)
			}
		}
	}
	return m, nil
}

func (m *Model) saveState() {
	stack := m.stackBar.Active()
	if stack == nil {
		return
	}
	s := state.SelectionState{
		Stack:    stack.Name,
		Selected: m.unitList.ExportSelection(),
	}
	if name := m.unitList.AnchorName(); name != "" {
		s.Anchor = name
		switch m.unitList.ellipsisMode {
		case model.FilterWithDeps:
			s.AnchorMode = "with_deps"
		case model.FilterWithDependents:
			s.AnchorMode = "with_dependents"
		}
	}
	state.SaveSelection(s)
	state.SaveLog(stack.Name, m.output.Lines(), m.cfg.LogLines)
}

func (m *Model) restoreState() {
	stack := m.stackBar.Active()
	if stack == nil {
		return
	}
	// Restore selection
	s, err := state.LoadSelection()
	if err == nil && s.Stack == stack.Name {
		m.unitList.ImportSelection(s.Selected)
		if s.Anchor != "" {
			var mode model.FilterMode
			switch s.AnchorMode {
			case "with_deps":
				mode = model.FilterWithDeps
			case "with_dependents":
				mode = model.FilterWithDependents
			}
			m.unitList.SetAnchor(s.Anchor, mode)
		}
	}
	// Restore historical output
	lines := state.LoadLog(stack.Name)
	if len(lines) > 0 {
		m.output.LoadHistorical(lines)
	}
}

func (m *Model) cliPreview() string {
	if m.stackBar.Active() == nil || m.active == viewUnitDetail {
		return ""
	}
	if m.lastCLI != "" {
		return dimStyle(m.lastCLI)
	}
	filters := m.unitList.BuildFilter()
	flags := filters.ToFlags()
	if len(flags) == 0 {
		return dimStyle("no filters — runs full stack")
	}
	return dimStyle(strings.Join(flags, " "))
}

func (m *Model) panelHeight() int {
	// Unit list/detail content height (inside border, so subtract 2 for top/bottom border)
	h := m.height*2/5 - 2
	if h < 6 {
		h = 6
	}
	return h
}

func (m *Model) outputHeight() int {
	// Output content height: total - stacks pane - units pane - status bar(1)
	// Each pane has 2 lines of border overhead
	stacksPaneH := 3 // 1 line content + 2 border
	unitsPaneH := m.panelHeight() + 2
	statusH := 1
	return m.height - stacksPaneH - unitsPaneH - statusH - 2 // -2 for output border
}

func (m *Model) refreshUnits() tea.Cmd {
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	s := *stack
	return func() tea.Msg {
		return stackUnitsMsg{units: s.Units}
	}
}

func (m *Model) runBackendBootstrap() tea.Cmd {
	if m.running {
		return nil
	}
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	m.running = true
	m.output.Clear()
	m.output.AddLine("running: terragrunt backend bootstrap --all", false)
	return m.runner.BackendBootstrap(*stack)
}

func (m *Model) runStackGenerate() tea.Cmd {
	if m.running {
		return nil
	}
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	m.running = true
	m.output.Clear()
	m.output.AddLine("running: terragrunt stack generate", false)

	s := *stack
	r := m.runner
	return func() tea.Msg {
		cmd := r.StackLifecycle(s, "generate")
		result := cmd()
		r.Send(rescanMsg{})
		return result
	}
}

func (m *Model) runStackLifecycle(command string) tea.Cmd {
	if m.running {
		return nil
	}
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	m.running = true
	m.output.Clear()
	m.output.AddLine(fmt.Sprintf("running: terragrunt stack %s", command), false)

	s := *stack
	r := m.runner
	return func() tea.Msg {
		cmd := r.StackLifecycle(s, command)
		result := cmd()
		r.Send(rescanMsg{})
		return result
	}
}

func (m *Model) runStackCmd(command string) tea.Cmd {
	return m.runStackCmdWithFlags(command, nil)
}

func (m *Model) runStackCmdWithFlags(command string, extraFlags []string) tea.Cmd {
	if m.running {
		return nil
	}
	stack := m.stackBar.Active()
	if stack == nil {
		return nil
	}
	if !m.unitList.HasSelection() {
		m.output.AddLine("no units selected — use v/space to select or A for all", true)
		return nil
	}
	m.running = true
	m.output.Clear()

	opts := runner.RunOpts{
		Filters:    m.unitList.BuildFilter(),
		ExtraFlags: extraFlags,
	}
	flags := opts.Filters.ToFlags()
	cli := "terragrunt stack run " + command
	if len(extraFlags) > 0 {
		cli += " " + strings.Join(extraFlags, " ")
	}
	if len(flags) > 0 {
		cli += " " + strings.Join(flags, " ")
	}
	m.lastCLI = cli
	m.output.AddLine("running: "+cli, false)

	return m.runner.StackRun(*stack, command, opts)
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	if m.active == viewSplash {
		return m.splashView()
	}

	if m.outputFullscreen {
		innerWidth := m.width - 2
		outH := m.height - 4 // status bar + borders
		var cliLine string
		if m.showCLI {
			cliLine = statusBarStyle.Width(m.width).Render(m.cliPreview())
			outH--
		}
		if outH < 3 {
			outH = 3
		}
		m.output.height = outH
		outputTitle := "Output"
		if m.output.searching {
			outputTitle = "Output  /" + m.output.searchQuery + "_"
		} else if len(m.output.searchMatches) > 0 {
			outputTitle = fmt.Sprintf("Output  [%d/%d]  n/N: next/prev", m.output.matchIdx+1, len(m.output.searchMatches))
		} else if m.running {
			outputTitle = "Output (running...)"
		}
		outputContent := m.output.View(innerWidth)
		outputPane := renderPane(outputTitle, outputContent, innerWidth, outH, true)
		status := m.statusBar()
		fsParts := []string{outputPane}
		if cliLine != "" {
			fsParts = append(fsParts, cliLine)
		}
		fsParts = append(fsParts, status)
		return lipgloss.JoinVertical(lipgloss.Left, fsParts...)
	}

	if m.viewer.visible {
		return m.viewer.View(m.width, m.height)
	}

	if m.filePicker.visible {
		return m.filePicker.View(m.width, m.height)
	}

	if m.help.visible {
		return m.help.View(m.width, m.height)
	}

	innerWidth := m.width - 2 // account for left+right border chars

	// Stacks pane
	stacksContent := m.stackBar.ViewContent()
	stacksPane := renderPane("Stacks", stacksContent, innerWidth, 1, true)

	// Units/Detail pane
	panelH := m.panelHeight()
	var mainContent string
	var mainTitle string
	switch m.active {
	case viewUnitList:
		mainTitle = "Units"
		mainContent = m.unitList.View(innerWidth, panelH)
	case viewUnitDetail:
		mainTitle = m.unitDetail.unit.Name
		mainContent = m.unitDetail.View(innerWidth, panelH)
	}
	mainPane := renderPane(mainTitle, mainContent, innerWidth, panelH, true)

	// CLI preview line (between status bar and output)
	var cliLine string
	if m.showCLI {
		cliLine = statusBarStyle.Width(m.width).Render(m.cliPreview())
	}

	// Output pane
	outH := m.outputHeight()
	if cliLine != "" {
		outH--
	}
	if outH < 3 {
		outH = 3
	}
	m.output.height = outH
	outputTitle := "Output"
	if m.running {
		outputTitle = "Output (running...)"
	}
	outputContent := m.output.View(innerWidth)
	outputPane := renderPane(outputTitle, outputContent, innerWidth, outH, false)

	// Status bar (no border, sits between panes)
	status := m.statusBar()

	parts := []string{stacksPane, mainPane, status}
	if cliLine != "" {
		parts = append(parts, cliLine)
	}
	parts = append(parts, outputPane)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) splashView() string {
	art := splashStyle.Render(banner)
	subtitle := lipgloss.NewStyle().Foreground(dimGrey).Render("press any key to continue")
	content := art + "\n\n" + subtitle
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func renderPane(title, content string, width, height int, active bool) string {
	borderColor := purpleDim
	titleSty := paneTitleStyle
	if active {
		borderColor = purple
		titleSty = paneTitleActiveStyle
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)

	// Build top border: ╭─ Title ─────────╮
	titleStr := titleSty.Render(title)
	// Calculate remaining width for horizontal line
	// width is inner content width; total = width + 2 (borders)
	topLeft := bc.Render("╭─")
	topRight := bc.Render("─╮")
	titleRendered := " " + titleStr + " "
	// Fill remaining with ─
	// We approximate: total outer width = width + 2
	remainLen := width - lipgloss.Width(titleRendered) - 2
	if remainLen < 0 {
		remainLen = 0
	}
	topLine := topLeft + titleRendered + bc.Render(strings.Repeat("─", remainLen)) + topRight

	// Pad/truncate content lines to height
	contentLines := strings.Split(content, "\n")
	for len(contentLines) < height {
		contentLines = append(contentLines, "")
	}
	if len(contentLines) > height {
		contentLines = contentLines[:height]
	}

	// Build middle lines: │ content │
	var middle []string
	for _, line := range contentLines {
		// Pad line to width
		lineWidth := lipgloss.Width(line)
		padding := width - lineWidth
		if padding < 0 {
			padding = 0
		}
		middle = append(middle, bc.Render("│")+line+strings.Repeat(" ", padding)+bc.Render("│"))
	}

	// Build bottom border: ╰──────────────────╯
	bottomLine := bc.Render("╰") + bc.Render(strings.Repeat("─", width)) + bc.Render("╯")

	all := []string{topLine}
	all = append(all, middle...)
	all = append(all, bottomLine)
	return strings.Join(all, "\n")
}

func (m Model) statusBar() string {
	if m.input.Active() {
		return m.input.View(m.width)
	}

	if m.unitList.IsFiltering() {
		prompt := keybindStyle.Render("/") + keybindDescStyle.Render(m.unitList.FilterText()) + unitCursorStyle.Render("▏")
		return statusBarStyle.Width(m.width).Render(prompt)
	}

	var parts []string

	if m.confirm != nil {
		parts = append(parts, statusErrorStyle.Render("CONFIRM (y/n)"))
		return statusBarStyle.Width(m.width).Render(strings.Join(parts, "  "))
	}

	if m.running {
		parts = append(parts, keybindStyle.Render("RUNNING"))
		parts = append(parts, keybindStyle.Render("ctrl+x")+keybindDescStyle.Render(":cancel"))
	} else if m.active == viewUnitDetail {
		parts = append(parts, keybindStyle.Render("i")+keybindDescStyle.Render(":init"))
		parts = append(parts, keybindStyle.Render("p")+keybindDescStyle.Render(":plan"))
		parts = append(parts, keybindStyle.Render("a")+keybindDescStyle.Render(":apply"))
		parts = append(parts, keybindStyle.Render("d")+keybindDescStyle.Render(":destroy"))
		parts = append(parts, keybindStyle.Render("r")+keybindDescStyle.Render(":refresh"))
		parts = append(parts, keybindStyle.Render("s")+keybindDescStyle.Render(":state"))
		parts = append(parts, keybindStyle.Render("o")+keybindDescStyle.Render(":outputs"))
		parts = append(parts, keybindStyle.Render("f")+keybindDescStyle.Render(":files"))
	} else {
		parts = append(parts, keybindStyle.Render("/")+keybindDescStyle.Render(":filter"))
		parts = append(parts, keybindStyle.Render("v")+keybindDescStyle.Render(":select"))
		parts = append(parts, keybindStyle.Render(".")+keybindDescStyle.Render(":ellipsis"))
		parts = append(parts, keybindStyle.Render("p")+keybindDescStyle.Render(":plan"))
		parts = append(parts, keybindStyle.Render("a")+keybindDescStyle.Render(":apply"))
		parts = append(parts, keybindStyle.Render("d")+keybindDescStyle.Render(":destroy"))
		parts = append(parts, keybindStyle.Render("S")+keybindDescStyle.Render(":generate"))
		parts = append(parts, keybindStyle.Render("C")+keybindDescStyle.Render(":clean"))
		parts = append(parts, keybindStyle.Render("F")+keybindDescStyle.Render(":output"))
		parts = append(parts, keybindStyle.Render("T")+keybindDescStyle.Render(":cli"))
	}

	parts = append(parts, keybindStyle.Render("?")+keybindDescStyle.Render(":help"))

	sel, total := m.unitList.SelectionCount()
	parts = append(parts, dimStyle(fmt.Sprintf("[%d/%d]", sel, total)))

	return statusBarStyle.Width(m.width).Render(strings.Join(parts, "  "))
}

