package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ltwalkerdev/terraforge/internal/discovery"
	"github.com/ltwalkerdev/terraforge/internal/runner"
	"github.com/ltwalkerdev/terraforge/internal/tui"
)

var Version = "0.1.13"

var workspace string

var rootCmd = &cobra.Command{
	Use:     "tforge",
	Short:   "TUI for managing Terragrunt stacks",
	Version: Version,
	RunE:    run,
}

func init() {
	rootCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "workspace root directory (defaults to current directory or nearest parent with .envrc)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ws, err := resolveWorkspace()
	if err != nil {
		return fmt.Errorf("could not resolve workspace: %w", err)
	}

	stacks, err := discovery.FindStacks(ws)
	if err != nil {
		return fmt.Errorf("error discovering stacks: %w", err)
	}

	for i := range stacks {
		units, err := discovery.FindUnits(stacks[i])
		if err != nil {
			continue
		}
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
		stacks[i].Units = units
	}

	r := runner.New()
	m := tui.NewModel(stacks, r, ws)

	p := tea.NewProgram(m, tea.WithAltScreen())
	r.SetProgram(p)

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

const defaultWorkspace = "/workspaces/live"

func resolveWorkspace() (string, error) {
	if workspace != "" {
		abs, err := filepath.Abs(workspace)
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	if _, err := os.Stat(filepath.Join(defaultWorkspace, "root.hcl")); err == nil {
		return defaultWorkspace, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".envrc")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "root.hcl")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return os.Getwd()
}
