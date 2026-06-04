package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type SelectionState struct {
	Stack      string            `json:"stack"`
	Selected   map[string]string `json:"selected"`
	Anchor     string            `json:"anchor,omitempty"`
	AnchorMode string            `json:"anchorMode,omitempty"`
}

func dataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "terraforge")
}

func statePath() string {
	return filepath.Join(dataDir(), "state.json")
}

func logDir() string {
	return filepath.Join(dataDir(), "logs")
}

func logPath(stackName string) string {
	safe := strings.ReplaceAll(stackName, "/", "_")
	return filepath.Join(logDir(), safe+".log")
}

func SaveSelection(s SelectionState) error {
	dir := dataDir()
	os.MkdirAll(dir, 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(), data, 0644)
}

func LoadSelection() (SelectionState, error) {
	var s SelectionState
	data, err := os.ReadFile(statePath())
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(data, &s)
	return s, err
}

func SaveLog(stackName string, lines []string, maxLines int) error {
	if maxLines <= 0 {
		return nil
	}
	os.MkdirAll(logDir(), 0755)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	content := strings.Join(lines, "\n")
	return os.WriteFile(logPath(stackName), []byte(content), 0644)
}

func LoadLog(stackName string) []string {
	data, err := os.ReadFile(logPath(stackName))
	if err != nil {
		return nil
	}
	content := string(data)
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}
