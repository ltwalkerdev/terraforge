package discovery

import (
	"fmt"
	"os"
	"testing"
)

func TestFindStacks(t *testing.T) {
	ws := "/workspaces/live"
	if _, err := os.Stat(ws); os.IsNotExist(err) {
		t.Skip("workspace not available")
	}

	stacks, err := FindStacks(ws)
	if err != nil {
		t.Fatalf("FindStacks error: %v", err)
	}

	if len(stacks) == 0 {
		t.Fatal("expected at least one stack")
	}

	fmt.Printf("Found %d stacks:\n", len(stacks))
	for _, s := range stacks {
		fmt.Printf("  - %s (%s)\n", s.Name, s.Dir)

		units, err := FindUnits(s)
		if err != nil {
			fmt.Printf("    error finding units: %v\n", err)
			continue
		}
		fmt.Printf("    %d units:\n", len(units))
		for _, u := range units {
			deps := ParseDependencies(u)
			fmt.Printf("      - %s (deps: %v)\n", u.Name, deps)
		}
	}
}
