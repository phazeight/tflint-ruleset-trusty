package custom

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/config"
)

func TestParseConfigResource_Empty(t *testing.T) {
	res, err := parseConfigResource(&config.Resource{Kind: "aws_s3_bucket", Keys: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.KeyAttributes) != 0 {
		t.Errorf("want 0 key attributes, got %v", res.KeyAttributes)
	}
	if len(res.KeyBlocks) != 0 {
		t.Errorf("want 0 key blocks, got %v", res.KeyBlocks)
	}
}

func TestParseConfigResource_FlatKeys(t *testing.T) {
	res, err := parseConfigResource(&config.Resource{
		Kind: "aws_instance",
		Keys: []string{"ami", "instance_type"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantAttrs := []string{"ami", "instance_type"}
	if len(res.KeyAttributes) != len(wantAttrs) {
		t.Fatalf("want %d key attributes, got %d", len(wantAttrs), len(res.KeyAttributes))
	}
	for i, a := range wantAttrs {
		if res.KeyAttributes[i] != a {
			t.Errorf("key[%d]: want %q, got %q", i, a, res.KeyAttributes[i])
		}
	}
	if len(res.KeyBlocks) != 0 {
		t.Errorf("want no key blocks for flat keys, got %v", res.KeyBlocks)
	}
}

func TestParseConfigResource_NestedKeys(t *testing.T) {
	res, err := parseConfigResource(&config.Resource{
		Kind: "kubernetes_deployment",
		Keys: []string{"metadata.namespace", "metadata.name"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.KeyBlocks) != 1 || res.KeyBlocks[0] != "metadata" {
		t.Errorf("want KeyBlocks=[metadata], got %v", res.KeyBlocks)
	}
	wantAttrs := []string{"namespace", "name"}
	for i, a := range wantAttrs {
		if res.KeyAttributes[i] != a {
			t.Errorf("key[%d]: want %q, got %q", i, a, res.KeyAttributes[i])
		}
	}
}

func TestParseConfigResource_InconsistentPrefix(t *testing.T) {
	_, err := parseConfigResource(&config.Resource{
		Kind: "bad_resource",
		Keys: []string{"metadata.name", "spec.id"},
	})
	if err == nil {
		t.Error("expected error for inconsistent key prefix, got nil")
	}
}

func TestParseConfigResource_SingleKey(t *testing.T) {
	res, err := parseConfigResource(&config.Resource{
		Kind: "aws_s3_bucket",
		Keys: []string{"bucket"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.KeyAttributes) != 1 || res.KeyAttributes[0] != "bucket" {
		t.Errorf("want [bucket], got %v", res.KeyAttributes)
	}
}
