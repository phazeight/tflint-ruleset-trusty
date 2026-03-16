# CLAUDE.md — tflint-ruleset-trusty

Custom tflint plugin for `trusty_terraform` linting rules.

## Repo Purpose

A Go-based tflint plugin built on `tflint-plugin-sdk`. Currently implements
one rule (`trusty_key_attributes`) that enforces key-attribute ordering on
Terraform resource blocks.

## Package Structure

```
main.go              Plugin entrypoint (plugin.Serve)
project/main.go      Name ("trusty"), version, RuleName/ReferenceLink helpers
config/config.go     Default resource key-attribute definitions (150 types)
custom/
  ruleset.go         Custom RuleSet — config schema + ApplyConfig
  runner.go          Custom Runner — wraps tflint.Runner with resource map
  config.go          Dot-notation key parser (metadata.name → KeyBlocks + KeyAttributes)
node/
  node.go            Ordered AST node traversal
  attribute.go       HCL attribute wrapper
  block.go           HCL block wrapper
visit/main.go        File and block iteration helpers
rules/
  rules.go           Rule registry (All())
  rule_key_attributes.go  Key-attribute ordering enforcement
```

## Adding a New Rule

1. Create `rules/rule_<name>.go` implementing `tflint.Rule`
2. Register it in `rules/rules.go` → `All()`
3. Add tests in `rules/rule_<name>_test.go` using `helper.TestRunner` + `custom.NewRunner`
4. Add a doc page at `docs/trusty_<name>.md`
5. Bump `project.Version` and cut a new release

## Adding Resource Types to key_attributes

Edit `config/config.go` — add a `Resource` entry:

```go
{Kind: "aws_new_resource", Keys: []string{"name"}},
```

Dot notation for nested key blocks:

```go
{Kind: "kubernetes_deployment", Keys: []string{"metadata.namespace", "metadata.name"}},
```

After adding, update `TestNew_ResourceCount` in `config/config_test.go` to match
the new count.

## Testing

```bash
go test ./...       # all packages — no ? marks expected
make test           # same via Makefile
```

## Releasing

Bump `project.Version` in `project/main.go`, commit, then tag:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

The `release.yml` workflow picks up the tag, builds cross-platform binaries,
signs `checksums.txt` with the GPG key, and publishes a GitHub Release.

## Commit Convention

Follow Conventional Commits:

| Prefix | When |
|--------|------|
| `feat(rules):` | New rule or new resource type coverage |
| `fix(rules):` | Bug in rule logic |
| `fix(config):` | Wrong key-attribute definition |
| `chore:` | Deps, tooling, CI |
