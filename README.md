# tflint-ruleset-trusty

Custom tflint plugin for `trusty_terraform` linting rules.

## Setup

This directory contains the scaffolded source for the
`phazeight/tflint-ruleset-trusty` Go plugin. To use it:

1. Copy this directory into a new `phazeight/tflint-ruleset-trusty` repository
2. Run `go mod tidy` to resolve all dependencies
3. Run `make test` to verify rules pass
4. Run `make install` to build and install locally

## Usage

Add to your `.tflint.hcl`:

```hcl
plugin "trusty" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/phazeight/tflint-ruleset-trusty"
}
```

Then run:

```bash
tflint --init
tflint
```

## Rules

| Rule | Description | Default |
|------|-------------|---------|
| `trusty_key_attributes` | Key attributes must appear at top of resource blocks in priority order | Enabled (ERROR) |

## Architecture

```text
main.go              Plugin entrypoint (plugin.Serve)
project/main.go      Name, version, helper functions
config/config.go     Default resource key-attribute definitions (140+ types)
custom/
  ruleset.go         Custom RuleSet with config schema support
  runner.go          Custom Runner with resource map
  config.go          Dot-notation key parser (metadata.name -> KeyBlocks + KeyAttributes)
node/
  node.go            Ordered AST node traversal
  attribute.go       HCL attribute wrapper
  block.go           HCL block wrapper
visit/main.go        File and block iteration helpers
rules/
  rules.go           Rule registry
  rule_key_attributes.go  Key-attribute ordering enforcement
```

## Adding new resource types

Edit `config/config.go` and add a new `Resource` entry:

```go
{Kind: "aws_new_resource", Keys: []string{"name"}},
```

For resources with nested key blocks use dot notation:

```go
{Kind: "kubernetes_new_resource", Keys: []string{"metadata.namespace", "metadata.name"}},
```

## Releasing

Requires GPG key configured in GitHub:

```bash
make release
```

Uses GoReleaser to build cross-platform binaries and publish signed releases.
