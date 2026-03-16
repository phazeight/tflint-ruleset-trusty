package config_test

import (
	"fmt"
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/config"
)

func TestNew_NotNil(t *testing.T) {
	c := config.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Resources == nil {
		t.Fatal("New().Resources is nil")
	}
}

func TestNew_ResourceCount(t *testing.T) {
	c := config.New()
	if len(c.Resources) != 150 {
		t.Errorf("New() has %d resources, want 150", len(c.Resources))
	}
}

func TestNew_NoDuplicateKinds(t *testing.T) {
	c := config.New()
	seen := make(map[string]int, len(c.Resources))
	for i, r := range c.Resources {
		if prev, dup := seen[r.Kind]; dup {
			t.Errorf("duplicate Kind %q at index %d (first seen at %d)", r.Kind, i, prev)
		}
		seen[r.Kind] = i
	}
}

func TestNew_NoEmptyKinds(t *testing.T) {
	c := config.New()
	for i, r := range c.Resources {
		if r.Kind == "" {
			t.Errorf("resource at index %d has empty Kind", i)
		}
	}
}

func TestNew_SpotChecks(t *testing.T) {
	c := config.New()

	index := make(map[string]*config.Resource, len(c.Resources))
	for _, r := range c.Resources {
		index[r.Kind] = r
	}

	cases := []struct {
		kind string
		keys []string
	}{
		{"archive_file", []string{"source_file", "output_path"}},
		{"external", []string{}},
		{"aws_s3_bucket", []string{"bucket"}},
		{"aws_instance", []string{"ami", "instance_type"}},
		{"google_container_cluster", []string{"project", "location", "name"}},
		{"kubernetes_deployment", []string{"metadata.namespace", "metadata.name"}},
	}

	for _, tc := range cases {
		r, ok := index[tc.kind]
		if !ok {
			t.Errorf("resource %q not found in defaults", tc.kind)
			continue
		}
		if fmt.Sprintf("%v", r.Keys) != fmt.Sprintf("%v", tc.keys) {
			t.Errorf("%s: keys = %v, want %v", tc.kind, r.Keys, tc.keys)
		}
	}
}
