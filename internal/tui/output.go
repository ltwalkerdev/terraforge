package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxOutputLines = 10000

var (
	outputWarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFBD2E"))
	outputInfoStyle    = lipgloss.NewStyle().Foreground(lightGrey)
	outputDebugStyle   = lipgloss.NewStyle().Foreground(dimGrey)
	outputAddStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6ECE6B"))
	outputDestroyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F56"))
	outputChangeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFBD2E"))
	outputBoldStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	outputCyanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#56B6C2"))
)

type outputPane struct {
	lines         []outputLine
	scrollPos     int
	autoScroll    bool
	height        int
	searching     bool
	searchQuery   string
	searchMatches []int
	matchIdx      int
}

type outputLine struct {
	text  string
	isErr bool
}

func newOutputPane() outputPane {
	return outputPane{
		autoScroll: true,
		height:     10,
	}
}

func (o *outputPane) AddLine(text string, isErr bool) {
	o.lines = append(o.lines, outputLine{text: text, isErr: isErr})
	if len(o.lines) > maxOutputLines {
		o.lines = o.lines[len(o.lines)-maxOutputLines:]
	}
	if o.autoScroll {
		o.scrollPos = len(o.lines) - o.height
		if o.scrollPos < 0 {
			o.scrollPos = 0
		}
	}
}

func (o *outputPane) Clear() {
	o.lines = nil
	o.scrollPos = 0
	o.autoScroll = true
}

func (o *outputPane) LoadHistorical(lines []string) {
	o.lines = nil
	for _, l := range lines {
		o.lines = append(o.lines, outputLine{text: l, isErr: false})
	}
	o.scrollPos = len(o.lines) - o.height
	if o.scrollPos < 0 {
		o.scrollPos = 0
	}
	o.autoScroll = true
}

func (o *outputPane) Lines() []string {
	result := make([]string, len(o.lines))
	for i, l := range o.lines {
		result[i] = l.text
	}
	return result
}

func (o *outputPane) ScrollUp() {
	o.autoScroll = false
	if o.scrollPos > 0 {
		o.scrollPos--
	}
}

func (o *outputPane) ScrollDown() {
	if o.scrollPos < len(o.lines)-o.height {
		o.scrollPos++
	}
	if o.scrollPos >= len(o.lines)-o.height {
		o.autoScroll = true
	}
}

func (o *outputPane) HalfPageUp() {
	o.autoScroll = false
	jump := o.height / 2
	o.scrollPos -= jump
	if o.scrollPos < 0 {
		o.scrollPos = 0
	}
}

func (o *outputPane) HalfPageDown() {
	jump := o.height / 2
	o.scrollPos += jump
	max := len(o.lines) - o.height
	if max < 0 {
		max = 0
	}
	if o.scrollPos >= max {
		o.scrollPos = max
		o.autoScroll = true
	}
}

func (o *outputPane) GoToTop() {
	o.autoScroll = false
	o.scrollPos = 0
}

func (o *outputPane) GoToBottom() {
	max := len(o.lines) - o.height
	if max < 0 {
		max = 0
	}
	o.scrollPos = max
	o.autoScroll = true
}

func (o *outputPane) StartSearch() {
	o.searching = true
	o.searchQuery = ""
	o.searchMatches = nil
	o.matchIdx = 0
	o.autoScroll = false
}

func (o *outputPane) CancelSearch() {
	o.searching = false
	o.searchQuery = ""
	o.searchMatches = nil
}

func (o *outputPane) ConfirmSearch() {
	o.searching = false
	if len(o.searchMatches) > 0 {
		o.matchIdx = 0
		o.scrollToMatch()
	}
}

func (o *outputPane) SetSearchQuery(q string) {
	o.searchQuery = q
	o.searchMatches = nil
	o.matchIdx = 0
	if q == "" {
		return
	}
	lower := strings.ToLower(q)
	for i, line := range o.lines {
		if strings.Contains(strings.ToLower(line.text), lower) {
			o.searchMatches = append(o.searchMatches, i)
		}
	}
	if len(o.searchMatches) > 0 {
		o.scrollToMatch()
	}
}

func (o *outputPane) NextMatch() {
	if len(o.searchMatches) == 0 {
		return
	}
	o.matchIdx = (o.matchIdx + 1) % len(o.searchMatches)
	o.scrollToMatch()
}

func (o *outputPane) PrevMatch() {
	if len(o.searchMatches) == 0 {
		return
	}
	o.matchIdx--
	if o.matchIdx < 0 {
		o.matchIdx = len(o.searchMatches) - 1
	}
	o.scrollToMatch()
}

func (o *outputPane) scrollToMatch() {
	if len(o.searchMatches) == 0 {
		return
	}
	target := o.searchMatches[o.matchIdx]
	o.scrollPos = target - o.height/2
	if o.scrollPos < 0 {
		o.scrollPos = 0
	}
	max := len(o.lines) - o.height
	if max < 0 {
		max = 0
	}
	if o.scrollPos > max {
		o.scrollPos = max
	}
}

func (o *outputPane) isSearchMatch(lineIdx int) bool {
	for _, m := range o.searchMatches {
		if m == lineIdx {
			return true
		}
	}
	return false
}

func (o *outputPane) View(width int) string {
	var visible []string

	if len(o.lines) == 0 {
		visible = append(visible, outputStyle.Render("  waiting for command output..."))
	} else {
		end := o.scrollPos + o.height
		if end > len(o.lines) {
			end = len(o.lines)
		}
		start := o.scrollPos
		if start < 0 {
			start = 0
		}

		for i := start; i < end; i++ {
			line := o.lines[i]
			text := line.text
			if len(text) > width-4 {
				text = text[:width-4]
			}
			styled := styleLine(text, line.isErr)
			if o.isSearchMatch(i) {
				styled = searchMatchStyle.Render(styled)
			}
			visible = append(visible, "  "+styled)
		}
	}

	// Pad to fill the full pane height
	for len(visible) < o.height {
		visible = append(visible, "")
	}

	return strings.Join(visible, "\n")
}

func styleLine(text string, isErr bool) string {
	trimmed := strings.TrimSpace(text)

	// Errors
	if isErr || strings.Contains(strings.ToUpper(text), "ERROR") || strings.Contains(strings.ToUpper(text), "FATAL") {
		return outputErrStyle.Render(text)
	}

	// Warnings
	if strings.Contains(strings.ToUpper(text), "WARN") {
		return outputWarnStyle.Render(text)
	}

	// Debug/trace noise
	if strings.Contains(strings.ToUpper(text), "DEBUG") || strings.Contains(strings.ToUpper(text), "TRACE") {
		return outputDebugStyle.Render(text)
	}

	// Terraform plan diff markers
	if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "# ") && strings.Contains(text, "will be created") {
		return outputAddStyle.Render(text)
	}
	if strings.HasPrefix(trimmed, "-") || strings.Contains(text, "will be destroyed") {
		return outputDestroyStyle.Render(text)
	}
	if strings.HasPrefix(trimmed, "~") || strings.Contains(text, "will be updated") || strings.Contains(text, "must be replaced") {
		return outputChangeStyle.Render(text)
	}
	if strings.HasPrefix(trimmed, "<=") || strings.Contains(text, "will be read") {
		return outputCyanStyle.Render(text)
	}

	// Plan/Apply summary lines
	if strings.HasPrefix(trimmed, "Plan:") || strings.HasPrefix(trimmed, "Apply complete!") || strings.HasPrefix(trimmed, "Destroy complete!") {
		return outputBoldStyle.Render(text)
	}
	if strings.Contains(text, "No changes.") || strings.Contains(text, "Your infrastructure matches") {
		return outputAddStyle.Render(text)
	}

	// Resource addresses in plan output
	if strings.HasPrefix(trimmed, "# ") {
		return outputCyanStyle.Render(text)
	}

	return outputInfoStyle.Render(text)
}
