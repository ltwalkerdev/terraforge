package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ltwalkerdev/terraforge/internal/model"
)

func FindStacks(workspaceRoot string) ([]model.Stack, error) {
	var stacks []model.Stack

	err := filepath.WalkDir(workspaceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".terragrunt-stack" || base == ".terragrunt-cache" || base == ".git" || base == "terraforge" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "terragrunt.stack.hcl" && !strings.Contains(path, ".terragrunt-stack") {
			dir := filepath.Dir(path)
			rel, _ := filepath.Rel(workspaceRoot, dir)
			stacks = append(stacks, model.Stack{
				Name: rel,
				Dir:  dir,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func FindUnits(stack model.Stack) ([]model.Unit, error) {
	stackCacheDir := filepath.Join(stack.Dir, ".terragrunt-stack")
	if _, err := os.Stat(stackCacheDir); os.IsNotExist(err) {
		return findDirectUnits(stack)
	}

	var units []model.Unit

	err := filepath.WalkDir(stackCacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		hclPath := filepath.Join(path, "terragrunt.hcl")
		stackHclPath := filepath.Join(path, "terragrunt.stack.hcl")

		if _, err := os.Stat(stackHclPath); err == nil {
			rel, _ := filepath.Rel(stackCacheDir, path)
			name := filepath.Base(path)
			children := findNestedStackChildren(path)
			units = append(units, model.Unit{
				Name:     name,
				Path:     path,
				RelPath:  rel,
				Status:   model.StatusUnknown,
				IsStack:  true,
				Children: children,
			})
			return filepath.SkipDir
		}

		if _, err := os.Stat(hclPath); err == nil {
			rel, _ := filepath.Rel(stackCacheDir, path)
			name := filepath.Base(path)

			if name == "_envcommon" {
				return filepath.SkipDir
			}

			units = append(units, model.Unit{
				Name:    name,
				Path:    path,
				RelPath: rel,
				Status:  model.StatusUnknown,
			})
			return filepath.SkipDir
		}

		if path != stackCacheDir && hasHCLFiles(path) {
			rel, _ := filepath.Rel(stackCacheDir, path)
			name := filepath.Base(path)
			units = append(units, model.Unit{
				Name:        name,
				Path:        path,
				RelPath:     rel,
				Status:      model.StatusUnknown,
				IncludeOnly: true,
			})
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return units, nil
}

func ParseDependencies(unit model.Unit) []string {
	hclPath := filepath.Join(unit.Path, "terragrunt.hcl")
	data, err := os.ReadFile(hclPath)
	if err != nil {
		return nil
	}

	var deps []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// dependency block: config_path = "../vpc"
		if strings.HasPrefix(trimmed, "config_path") && strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"`)
				dep := cleanDepName(filepath.Base(val))
				if dep != "" && dep != "." && dep != ".." {
					deps = append(deps, dep)
				}
			}
		}

		// dependencies block: paths = ["../vpc", "../database"]
		if strings.HasPrefix(trimmed, "paths") && strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "[] ")
				for _, entry := range strings.Split(val, ",") {
					entry = strings.TrimSpace(entry)
					entry = strings.Trim(entry, `"`)
					dep := cleanDepName(filepath.Base(entry))
					if dep != "" && dep != "." && dep != ".." {
						deps = append(deps, dep)
					}
				}
			}
		}
	}
	return deps
}

func findDirectUnits(stack model.Stack) ([]model.Unit, error) {
	var units []model.Unit
	entries, err := os.ReadDir(stack.Dir)
	if err != nil {
		return nil, nil
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(stack.Dir, e.Name())
		hcl := filepath.Join(path, "terragrunt.hcl")
		if _, err := os.Stat(hcl); err == nil {
			units = append(units, model.Unit{
				Name:    e.Name(),
				Path:    path,
				RelPath: e.Name(),
				Status:  model.StatusUnknown,
			})
		} else if hasHCLFiles(path) {
			units = append(units, model.Unit{
				Name:        e.Name(),
				Path:        path,
				RelPath:     e.Name(),
				Status:      model.StatusUnknown,
				IncludeOnly: true,
			})
		}
	}
	return units, nil
}

func hasHCLFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".hcl") {
			return true
		}
	}
	return false
}

func findNestedStackChildren(stackPath string) []string {
	var children []string
	nestedCache := filepath.Join(stackPath, ".terragrunt-stack")

	if _, err := os.Stat(nestedCache); err == nil {
		filepath.WalkDir(nestedCache, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() || p == nestedCache {
				return nil
			}
			if hasHCLFiles(p) {
				children = append(children, filepath.Base(p))
				return filepath.SkipDir
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(stackPath)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if hasHCLFiles(filepath.Join(stackPath, e.Name())) {
				children = append(children, e.Name())
			}
		}
	}
	return children
}

func cleanDepName(name string) string {
	if !strings.Contains(name, "${") {
		return name
	}
	// Extract quoted string alternatives from ternary expressions
	// e.g. ${values.x ? "msk" : "kafka"} → msk|kafka
	var alternatives []string
	inQuote := false
	var current []byte
	for i := 0; i < len(name); i++ {
		if name[i] == '"' {
			if inQuote {
				if len(current) > 0 {
					alternatives = append(alternatives, string(current))
				}
				current = nil
				inQuote = false
			} else {
				inQuote = true
			}
		} else if inQuote {
			current = append(current, name[i])
		}
	}
	if len(alternatives) > 0 {
		return strings.Join(alternatives, "|")
	}
	// Fallback: strip interpolation markers
	name = strings.ReplaceAll(name, "${", "")
	name = strings.ReplaceAll(name, "}", "")
	return name
}
