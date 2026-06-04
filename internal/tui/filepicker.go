package tui

import (
	"path/filepath"
	"strings"
)

type filePicker struct {
	visible bool
	files   []string
	labels  []string
	cursor  int
}

func newFilePicker() filePicker {
	return filePicker{}
}

func (fp *filePicker) Open(files []string, baseDir string) {
	fp.files = files
	fp.labels = make([]string, len(files))
	for i, f := range files {
		rel, err := filepath.Rel(baseDir, f)
		if err != nil {
			fp.labels[i] = filepath.Base(f)
		} else {
			fp.labels[i] = rel
		}
	}
	fp.cursor = 0
	fp.visible = true
}

func (fp *filePicker) Close() {
	fp.visible = false
}

func (fp *filePicker) Up() {
	if fp.cursor > 0 {
		fp.cursor--
	}
}

func (fp *filePicker) Down() {
	if fp.cursor < len(fp.files)-1 {
		fp.cursor++
	}
}

func (fp *filePicker) Selected() string {
	if fp.cursor < len(fp.files) {
		return fp.files[fp.cursor]
	}
	return ""
}

func (fp *filePicker) View(width, height int) string {
	header := paneTitleActiveStyle.Render(" Select File ") +
		keybindDescStyle.Render("  (j/k: navigate  enter: open  esc: cancel)")

	var lines []string
	for i, label := range fp.labels {
		if i == fp.cursor {
			lines = append(lines, unitCursorStyle.Render("  ▸ "+label))
		} else {
			lines = append(lines, unitNormalStyle.Render("    "+label))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, keybindDescStyle.Render("  no files found"))
	}

	return header + "\n\n" + strings.Join(lines, "\n")
}
