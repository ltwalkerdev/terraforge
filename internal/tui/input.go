package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type inputMode int

const (
	inputNone inputMode = iota
	inputImportAddr
	inputImportID
	inputReplace
	inputForceUnlock
	inputStateMvSrc
	inputStateMvDst
	inputStdinConfirm
)

type textInput struct {
	mode    inputMode
	prompt  string
	value   string
	stash   string // holds first value for two-step inputs (import addr, state mv src)
}

func newTextInput() textInput {
	return textInput{}
}

func (t *textInput) Start(mode inputMode, prompt string) {
	t.mode = mode
	t.prompt = prompt
	t.value = ""
}

func (t *textInput) Active() bool {
	return t.mode != inputNone
}

func (t *textInput) Update(msg tea.KeyMsg) (string, bool) {
	switch {
	case key.Matches(msg, keys.Back):
		t.mode = inputNone
		return "", false
	case msg.String() == "enter":
		val := t.value
		t.value = ""
		return val, true
	case msg.String() == "backspace":
		if len(t.value) > 0 {
			t.value = t.value[:len(t.value)-1]
		}
	case msg.String() == "space":
		t.value += " "
	case len(msg.String()) == 1:
		t.value += msg.String()
	}
	return "", false
}

func (t *textInput) Cancel() {
	t.mode = inputNone
	t.value = ""
	t.stash = ""
}

func (t *textInput) View(width int) string {
	promptStr := keybindStyle.Render(t.prompt + ": ")
	cursor := unitCursorStyle.Render("_")
	val := unitNormalStyle.Render(t.value)
	line := promptStr + val + cursor
	padding := width - len(t.prompt) - len(t.value) - 4
	if padding < 0 {
		padding = 0
	}
	return line + strings.Repeat(" ", padding)
}
