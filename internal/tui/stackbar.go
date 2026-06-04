package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ltwalkerdev/terraforge/internal/model"
)

type stackBar struct {
	stacks []model.Stack
	active int
}

func newStackBar(stacks []model.Stack) stackBar {
	return stackBar{stacks: stacks, active: 0}
}

func (s *stackBar) Next() {
	if len(s.stacks) == 0 {
		return
	}
	s.active = (s.active + 1) % len(s.stacks)
}

func (s *stackBar) Prev() {
	if len(s.stacks) == 0 {
		return
	}
	s.active = (s.active - 1 + len(s.stacks)) % len(s.stacks)
}

func (s *stackBar) Active() *model.Stack {
	if len(s.stacks) == 0 {
		return nil
	}
	return &s.stacks[s.active]
}

func (s *stackBar) View(width int) string {
	return s.ViewContent()
}

func (s *stackBar) ViewContent() string {
	if len(s.stacks) == 0 {
		return "  no stacks found"
	}

	var parts []string
	for i, stack := range s.stacks {
		if i == s.active {
			parts = append(parts, stackActiveStyle.Render(fmt.Sprintf(" %s ", stack.Name)))
		} else {
			parts = append(parts, stackInactiveStyle.Render(stack.Name))
		}
	}

	parts = append(parts, dimStyle(fmt.Sprintf(" [%d/%d]", s.active+1, len(s.stacks))))
	return " " + strings.Join(parts, " ")
}

func dimStyle(s string) string {
	return lipgloss.NewStyle().Foreground(dimGrey).Render(s)
}
