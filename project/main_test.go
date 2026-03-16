package project_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/project"
)

func TestRuleName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"key_attributes", "trusty_key_attributes"},
		{"naming", "trusty_naming"},
	}
	for _, tc := range cases {
		if got := project.RuleName(tc.name); got != tc.want {
			t.Errorf("RuleName(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestReferenceLink(t *testing.T) {
	want := "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/trusty_key_attributes.md"
	if got := project.ReferenceLink("trusty_key_attributes"); got != want {
		t.Errorf("ReferenceLink() = %q, want %q", got, want)
	}
}
