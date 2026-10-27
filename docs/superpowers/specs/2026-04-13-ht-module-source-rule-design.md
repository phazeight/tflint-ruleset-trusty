# Design: `trusty_module_source` tflint rule

**Date:** 2026-04-13  
**Ticket:** DO-151  
**Repo:** `phazeight/tflint-ruleset-trusty`

## Problem

`terraform-modules` uses `terraform-module-releaser` which creates flat-commit tags — only the changed module's files at repo root, no directory tree. Three common mistakes cause broken `terraform init`:

1. **SSH sources** (`git::ssh://`) — breaks CI without a deploy key; HTTPS is the standard
2. **`//subdir` notation** (`git::https://...//modules/atoms/foo?ref=...`) — fails at init because the subdirectory doesn't exist in the flat commit
3. **Missing `?ref=`** — floating sources are non-reproducible

## Scope

Only module `source` values that reference `github.com/isapp/terraform-modules`. Local paths, registry sources, and other git repos are ignored.

## Rule: `trusty_module_source`

### Implementation

- **File:** `rules/trusty_module_source.go`
- **Pattern:** `GetFiles()` + `hclsyntax` — consistent with all existing rules
- **Severity:** `tflint.ERROR`

Walk all files. For each `module` block, extract the `source` attribute value using `sourceAttr.Expr.Value(nil).AsString()`. Skip if source doesn't contain `github.com/isapp/terraform-modules`. Run three independent checks, emit a separate issue per violation:

| Check | Condition | Message |
|---|---|---|
| Wrong scheme | `strings.Contains(source, "git::ssh://")` | `module source must use git::https://, not git::ssh://` |
| Subdir present | `strings.Contains(source, "//")` | `module source must not use //subdir notation — module repo tags are flat commits` |
| Missing ref | `!strings.Contains(source, "?ref=")` | `module source must be pinned with ?ref=<tag>` |

### Tests

- **File:** `rules/trusty_module_source_test.go`
- Uses `helper.TestRunner()` + `helper.AssertIssues()` from tflint-plugin-sdk
- Inline content (no testdata fixture files needed)

| Test case | Expected issues |
|---|---|
| Valid HTTPS + `?ref=` | 0 |
| Non-terraform-modules sources (local, registry) | 0 |
| SSH scheme | 1 |
| `//subdir` present | 1 |
| Missing `?ref=` | 1 |
| All three violations at once | 3 |

### Registration

Add `rules.NewModuleSourceRule()` to the `Rules` slice in `main.go`.
