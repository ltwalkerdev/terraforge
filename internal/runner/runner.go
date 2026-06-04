package runner

import (
	"context"
	"io"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ltwalkerdev/terraforge/internal/model"
)

type OutputLineMsg struct {
	Line  string
	IsErr bool
}

type CommandDoneMsg struct {
	Err      error
	UnitName string
	Command  string
}

type UnitStatusMsg struct {
	UnitName string
	Status   model.UnitStatus
}

type StateListMsg struct {
	Items []string
	Err   error
}

type StateShowMsg struct {
	Output string
	Err    error
}

type OutputsMsg struct {
	Output string
	Err    error
}

type ConfirmNeededMsg struct{}

type RunOpts struct {
	Filters    model.FilterExpr
	ExtraFlags []string
}

type Runner struct {
	program        *tea.Program
	cancel         context.CancelFunc
	stdin          io.WriteCloser
	prompted       bool
	cancelled      bool
	hasChanges     bool
	noChanges      bool
	collectOutput  bool
	collectedLines []string
}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) SetProgram(p *tea.Program) {
	r.program = p
}

func (r *Runner) Cancel() {
	if r.cancel != nil {
		r.cancelled = true
		r.cancel()
		r.cancel = nil
	}
}

func (r *Runner) Send(msg tea.Msg) {
	if r.program != nil {
		r.program.Send(msg)
	}
}

func (r *Runner) Cancelled() bool {
	return r.cancelled
}

func (r *Runner) MarkCancelled() {
	r.cancelled = true
}

func (r *Runner) Confirm(response string) {
	if r.stdin != nil {
		io.WriteString(r.stdin, response+"\n")
	}
}

func (r *Runner) run(dir string, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	cmd := exec.CommandContext(ctx, "direnv", args...)
	cmd.Dir = dir

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		r.cancel = nil
		return err
	}
	r.stdin = stdinPipe

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		r.cancel = nil
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		r.cancel = nil
		return err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		r.cancel = nil
		return err
	}

	r.prompted = false
	r.cancelled = false
	r.hasChanges = false
	r.noChanges = false
	r.collectedLines = nil

	go r.readInteractive(stderr, true)
	r.readInteractive(stdout, false)

	err = cmd.Wait()
	cancel()
	r.cancel = nil
	r.stdin = nil
	return err
}

func (r *Runner) runCollect(dir string, args []string) ([]string, error) {
	r.collectOutput = true
	err := r.run(dir, args)
	r.collectOutput = false
	return r.collectedLines, err
}

func (r *Runner) StackRun(stack model.Stack, command string, opts RunOpts) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "stack", "run", command, "--no-color"}
		args = append(args, opts.Filters.ToFlags()...)
		args = append(args, opts.ExtraFlags...)
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) StackLifecycle(stack model.Stack, command string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "stack", command, "--no-color"}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) UnitRun(stack model.Stack, unit model.Unit, command string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--no-color", "--working-dir", unit.Path, "--", command}

		if r.program != nil {
			r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusRunning})
		}

		err := r.run(stack.Dir, args)

		if r.program != nil {
			if err != nil && !r.cancelled {
				r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusError})
			} else if command == "plan" {
				if r.noChanges {
					r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusClean})
				} else if r.hasChanges {
					r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusChanged})
				} else {
					r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusClean})
				}
			} else if command == "apply" {
				r.program.Send(UnitStatusMsg{UnitName: unit.Name, Status: model.StatusClean})
			}
		}

		return CommandDoneMsg{Err: err, UnitName: unit.Name, Command: command}
	}
}

func (r *Runner) UnitStateList(stack model.Stack, unit model.Unit) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "state", "list"}
		lines, err := r.runCollect(stack.Dir, args)
		if err != nil {
			return StateListMsg{Err: err}
		}
		var items []string
		for _, l := range lines {
			if t := strings.TrimSpace(l); t != "" {
				items = append(items, t)
			}
		}
		return StateListMsg{Items: items}
	}
}

func (r *Runner) UnitStateShow(stack model.Stack, unit model.Unit, resource string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "state", "show", resource}
		lines, err := r.runCollect(stack.Dir, args)
		if err != nil {
			return StateShowMsg{Err: err}
		}
		return StateShowMsg{Output: strings.Join(lines, "\n")}
	}
}

func (r *Runner) UnitStateRm(stack model.Stack, unit model.Unit, resource string) tea.Cmd {
	return func() tea.Msg {
		if r.program != nil {
			r.program.Send(OutputLineMsg{Line: "removing: " + resource, IsErr: false})
		}
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "state", "rm", resource}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) UnitStateMv(stack model.Stack, unit model.Unit, src, dst string) tea.Cmd {
	return func() tea.Msg {
		if r.program != nil {
			r.program.Send(OutputLineMsg{Line: "moving: " + src + " → " + dst, IsErr: false})
		}
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "state", "mv", src, dst}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) UnitForceUnlock(stack model.Stack, unit model.Unit, lockID string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "force-unlock", lockID}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) UnitImport(stack model.Stack, unit model.Unit, addr, id string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--working-dir", unit.Path, "--", "import", addr, id}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err, UnitName: unit.Name, Command: "import"}
	}
}

func (r *Runner) UnitReplace(stack model.Stack, unit model.Unit, addr string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--no-color", "--working-dir", unit.Path, "--", "apply", "-replace=" + addr}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err, UnitName: unit.Name, Command: "replace"}
	}
}

func (r *Runner) UnitRefresh(stack model.Stack, unit model.Unit) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--no-color", "--working-dir", unit.Path, "--", "apply", "-refresh-only"}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err, UnitName: unit.Name, Command: "refresh"}
	}
}

func (r *Runner) UnitValidate(stack model.Stack, unit model.Unit) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "run", "--no-color", "--working-dir", unit.Path, "--", "validate"}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err, UnitName: unit.Name, Command: "validate"}
	}
}

func (r *Runner) UnitRender(stack model.Stack, unit model.Unit) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "render", "--format=json", "--working-dir", unit.Path}
		lines, err := r.runCollect(stack.Dir, args)
		if err != nil {
			return OutputsMsg{Err: err}
		}
		return OutputsMsg{Output: strings.Join(lines, "\n")}
	}
}

func (r *Runner) UnitOutputs(stack model.Stack, unit model.Unit) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "output", "--working-dir", unit.Path}
		lines, err := r.runCollect(stack.Dir, args)
		if err != nil {
			return OutputsMsg{Err: err}
		}
		return OutputsMsg{Output: strings.Join(lines, "\n")}
	}
}

func (r *Runner) BackendBootstrap(stack model.Stack) tea.Cmd {
	return func() tea.Msg {
		args := []string{"exec", stack.Dir, "terragrunt", "backend", "bootstrap", "--all"}
		err := r.run(stack.Dir, args)
		return CommandDoneMsg{Err: err}
	}
}

func (r *Runner) readInteractive(rd io.Reader, isErr bool) {
	buf := make([]byte, 1)
	var line []byte

	for {
		n, err := rd.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				text := string(line)
				if r.program != nil {
					r.program.Send(OutputLineMsg{Line: text, IsErr: isErr})
				}
				if !isErr {
					if r.collectOutput {
						r.collectedLines = append(r.collectedLines, text)
					}
					if strings.Contains(text, "No changes.") || strings.Contains(text, "Your infrastructure matches the configuration") {
						r.noChanges = true
					}
					if strings.Contains(text, "Plan:") && (strings.Contains(text, "to add") || strings.Contains(text, "to change") || strings.Contains(text, "to destroy")) {
						r.hasChanges = true
					}
				}
				line = line[:0]
			} else {
				line = append(line, buf[0])
				if !r.prompted && isConfirmPrompt(string(line)) {
					r.prompted = true
					if r.program != nil {
						r.program.Send(OutputLineMsg{Line: string(line), IsErr: isErr})
						r.program.Send(ConfirmNeededMsg{})
					}
					line = line[:0]
				}
			}
		}
		if err != nil {
			if len(line) > 0 && r.program != nil {
				r.program.Send(OutputLineMsg{Line: string(line), IsErr: isErr})
			}
			break
		}
	}
}

func isConfirmPrompt(line string) bool {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "enter a value") {
		return true
	}
	if strings.Contains(lower, "(y/n)") {
		return true
	}
	return false
}
