package main

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/project"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
)

// TestPluginWiring verifies that the plugin constants and rule registry are
// consistent — catching cases where a new rule is added to rules.All() but
// the name/version constants are accidentally cleared.
func TestPluginWiring(t *testing.T) {
	if project.Name == "" {
		t.Error("project.Name must not be empty")
	}
	if project.Version == "" {
		t.Error("project.Version must not be empty")
	}

	all := rules.All()
	if len(all) == 0 {
		t.Fatal("rules.All() returned no rules — at least one rule must be registered")
	}

	for _, r := range all {
		if r.Name() == "" {
			t.Error("a registered rule has an empty Name()")
		}
		if r.Link() == "" {
			t.Errorf("rule %q has an empty Link()", r.Name())
		}
	}
}
