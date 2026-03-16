# tflint-ruleset-trusty

Custom tflint plugin for `trusty_terraform` linting rules.

## Setup

```bash
go mod tidy    # resolve all dependencies
make test      # run tests
make install   # build and install locally
```

## Testing

```bash
go test ./...        # run all tests
go test ./rules/...  # run rule tests only
go test -v ./...     # verbose output
```

Tests live alongside the code they cover (`<package>/*_test.go`). Every
package has coverage:

| Package | What is tested |
|---------|----------------|
| `main` | Plugin wiring — name, version, and rule registry are non-empty |
| `config` | `New()` returns 150 resources with no duplicates; spot-checks key values |
| `custom` | `parseConfigResource` — flat keys, dot-notation nesting, error on inconsistent prefix |
| `node` | `WrapAttribute` / `WrapBlock` interface methods; `OrderedInspectableNodesFrom` source-position ordering |
| `project` | `RuleName` formatting; `ReferenceLink` URL construction |
| `rules` | `trusty_key_attributes` rule — correct/incorrect ordering, `for_each` skip, data blocks, multi-file, non-custom runner no-op |
| `visit` | `Files` and `Blocks` iterators — multi-file traversal, visitor error propagation, nested block isolation |

The rule tests use `helper.TestRunner` from `tflint-plugin-sdk`, wrapped in
`custom.NewRunner`, so they exercise the full check path against real HCL
without needing a running tflint process.

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
