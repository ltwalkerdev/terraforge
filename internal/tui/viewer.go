package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	kwStyle          = lipgloss.NewStyle().Foreground(purple).Bold(true)
	strStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#A5D6A7"))
	commentStyle     = lipgloss.NewStyle().Foreground(dimGrey).Italic(true)
	numStyle         = lipgloss.NewStyle().Foreground(purpleLight)
	searchMatchStyle = lipgloss.NewStyle().Background(lipgloss.Color("#4A4A00"))
)

var hclKeywords = []string{
	"terraform", "resource", "data", "variable", "output", "locals",
	"module", "provider", "dependency", "dependencies", "include",
	"remote_state", "inputs", "generate", "source", "config_path",
	"true", "false", "null", "unit", "stack", "values",
}

type fileViewer struct {
	visible       bool
	filePath      string
	fileName      string
	lines         []string
	scrollPos     int
	height        int
	searching     bool
	searchQuery   string
	searchMatches []int
	matchIdx      int
}

func newFileViewer() fileViewer {
	return fileViewer{}
}

func (v *fileViewer) Open(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	v.filePath = path
	v.fileName = filepath.Base(path)
	v.lines = strings.Split(string(data), "\n")
	v.scrollPos = 0
	v.visible = true
	return nil
}

func (v *fileViewer) Close() {
	v.visible = false
	v.lines = nil
}

func (v *fileViewer) ScrollUp() {
	if v.scrollPos > 0 {
		v.scrollPos--
	}
}

func (v *fileViewer) ScrollDown() {
	max := len(v.lines) - v.height
	if max < 0 {
		max = 0
	}
	if v.scrollPos < max {
		v.scrollPos++
	}
}

func (v *fileViewer) HalfPageUp() {
	v.scrollPos -= v.height / 2
	if v.scrollPos < 0 {
		v.scrollPos = 0
	}
}

func (v *fileViewer) HalfPageDown() {
	max := len(v.lines) - v.height
	if max < 0 {
		max = 0
	}
	v.scrollPos += v.height / 2
	if v.scrollPos > max {
		v.scrollPos = max
	}
}

func (v *fileViewer) GoToTop() {
	v.scrollPos = 0
}

func (v *fileViewer) GoToBottom() {
	max := len(v.lines) - v.height
	if max < 0 {
		max = 0
	}
	v.scrollPos = max
}

func (v *fileViewer) StartSearch() {
	v.searching = true
	v.searchQuery = ""
	v.searchMatches = nil
	v.matchIdx = 0
}

func (v *fileViewer) CancelSearch() {
	v.searching = false
	v.searchQuery = ""
	v.searchMatches = nil
}

func (v *fileViewer) ConfirmSearch() {
	v.searching = false
	if len(v.searchMatches) > 0 {
		v.matchIdx = 0
		v.scrollToMatch()
	}
}

func (v *fileViewer) SetSearchQuery(q string) {
	v.searchQuery = q
	v.searchMatches = nil
	v.matchIdx = 0
	if q == "" {
		return
	}
	lower := strings.ToLower(q)
	for i, line := range v.lines {
		if strings.Contains(strings.ToLower(line), lower) {
			v.searchMatches = append(v.searchMatches, i)
		}
	}
	if len(v.searchMatches) > 0 {
		v.scrollToMatch()
	}
}

func (v *fileViewer) NextMatch() {
	if len(v.searchMatches) == 0 {
		return
	}
	v.matchIdx = (v.matchIdx + 1) % len(v.searchMatches)
	v.scrollToMatch()
}

func (v *fileViewer) PrevMatch() {
	if len(v.searchMatches) == 0 {
		return
	}
	v.matchIdx--
	if v.matchIdx < 0 {
		v.matchIdx = len(v.searchMatches) - 1
	}
	v.scrollToMatch()
}

func (v *fileViewer) scrollToMatch() {
	if len(v.searchMatches) == 0 {
		return
	}
	target := v.searchMatches[v.matchIdx]
	v.scrollPos = target - v.height/2
	if v.scrollPos < 0 {
		v.scrollPos = 0
	}
	max := len(v.lines) - v.height
	if max < 0 {
		max = 0
	}
	if v.scrollPos > max {
		v.scrollPos = max
	}
}

func (v *fileViewer) isSearchMatch(lineIdx int) bool {
	for _, m := range v.searchMatches {
		if m == lineIdx {
			return true
		}
	}
	return false
}

func (v *fileViewer) View(width, height int) string {
	v.height = height - 4

	var helpText string
	if v.searching {
		helpText = "  /" + v.searchQuery + "_"
	} else if len(v.searchMatches) > 0 {
		helpText = fmt.Sprintf("  [%d/%d]  n/N: next/prev  /: search", v.matchIdx+1, len(v.searchMatches))
	} else {
		helpText = "  (q/esc: close  j/k: scroll  /: search)"
	}
	header := paneTitleActiveStyle.Render(" " + v.fileName + " ") +
		keybindDescStyle.Render(helpText)

	var content []string
	end := v.scrollPos + v.height
	if end > len(v.lines) {
		end = len(v.lines)
	}

	for i := v.scrollPos; i < end; i++ {
		line := v.lines[i]
		highlighted := highlightLine(line, width-4)
		if v.isSearchMatch(i) {
			highlighted = searchMatchStyle.Render(highlighted)
		}
		content = append(content, "  "+highlighted)
	}

	// Scroll indicator
	total := len(v.lines)
	pos := ""
	if total > v.height {
		pct := (v.scrollPos * 100) / (total - v.height)
		pos = keybindDescStyle.Render(strings.Repeat(" ", width-20)) +
			dimStyle(strings.Repeat(" ", 5)) +
			keybindDescStyle.Render(fmt.Sprintf("%d%%", pct))
	}

	body := strings.Join(content, "\n")
	return header + "\n\n" + body + "\n" + pos
}

func highlightLine(line string, maxWidth int) string {
	if len(line) > maxWidth {
		line = line[:maxWidth]
	}

	trimmed := strings.TrimSpace(line)

	// Comments
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
		return commentStyle.Render(line)
	}

	// Highlight keywords
	result := line
	for _, kw := range hclKeywords {
		if strings.Contains(result, kw) {
			result = highlightKeyword(result, kw)
		}
	}

	return result
}

func highlightKeyword(line, keyword string) string {
	idx := strings.Index(line, keyword)
	if idx < 0 {
		return line
	}

	// Only highlight if it's at a word boundary
	end := idx + len(keyword)
	if idx > 0 {
		prev := line[idx-1]
		if prev != ' ' && prev != '\t' && prev != '"' && prev != '{' && prev != '(' && prev != '=' {
			return line
		}
	}
	if end < len(line) {
		next := line[end]
		if next != ' ' && next != '\t' && next != '"' && next != '{' && next != '(' && next != '=' && next != ')' && next != '}' {
			return line
		}
	}

	return line[:idx] + kwStyle.Render(keyword) + line[end:]
}

func listStackFiles(stackDir string) []string {
	var files []string
	candidates := []string{
		filepath.Join(stackDir, "terragrunt.stack.hcl"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			files = append(files, c)
		}
	}
	// Also check for a values file at the stack's cache level
	stackCache := filepath.Join(stackDir, ".terragrunt-stack")
	if entries, err := os.ReadDir(stackCache); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".hcl") {
				files = append(files, filepath.Join(stackCache, e.Name()))
			}
		}
	}
	return files
}

func listUnitFiles(unitPath string) []string {
	var files []string

	// All HCL files in the unit directory
	if entries, err := os.ReadDir(unitPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".hcl") {
				files = append(files, filepath.Join(unitPath, e.Name()))
			}
		}
	}

	// Scripts at the unit level (.sh files)
	if entries, err := os.ReadDir(unitPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".sh") {
				files = append(files, filepath.Join(unitPath, e.Name()))
			}
		}
	}

	// Module .tf files from .terragrunt-cache (only top-level module, not nested .terraform/)
	cacheDir := filepath.Join(unitPath, ".terragrunt-cache")
	if _, err := os.Stat(cacheDir); err == nil {
		filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				base := filepath.Base(path)
				if base == ".terraform" || base == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".tf") {
				files = append(files, path)
			}
			return nil
		})
	}

	return files
}
