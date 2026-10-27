# `trusty_managed_repo_ref`

**Severity:** WARNING
**Scope:** `repository` attribute of GitHub provider resources, when the value matches a repo provisioned by a `managed-repo` module in the same configuration

## Summary

When a `module` block whose `source` contains `managed-repo` declares a `repo_name`, that repository is already modelled in the configuration. Hardcoding the same repo name string elsewhere duplicates the value and drifts if the module's `repo_name` changes. This rule flags string-literal `repository` values that match a managed repo name and nudges toward referencing the module output.

## How it works

Two passes over the module files:

1. **Collect** — every `module` block whose `source` contains `managed-repo` contributes its `repo_name` literal, mapped to the module label.
2. **Check** — for the covered resource types below, if `repository` is a static string literal that matches a collected repo name, emit a warning.

Only fires when a real in-module referent exists, so false positives are near-zero.

## Covered resource types

- `github_actions_secret`
- `github_actions_variable`
- `github_actions_environment`
- `github_team_repository`
- `github_repository_collaborator`

`data "github_repository".full_name` is **not** covered — external repos have no managed referent.

## Correct form

Reference the managed-repo module output instead of a literal:

```hcl
split("/", module.<module_label>.repository_full_name)[1]
```

## Examples

### Invalid

```hcl
module "my_repo" {
  source    = "git::https://github.com/isapp/terraform-modules?ref=modules/molecules/managed-repo/v1.0.0"
  repo_name = "my-repo"
}

resource "github_actions_secret" "token" {
  repository  = "my-repo" # WARNING: managed by module.my_repo
  secret_name = "TOKEN"
}
```

### Valid

```hcl
resource "github_actions_secret" "token" {
  repository  = split("/", module.my_repo.repository_full_name)[1]
  secret_name = "TOKEN"
}
```

## Suppression

Use tflint's native annotation on the offending line:

```hcl
# tflint-ignore: trusty_managed_repo_ref
repository = "my-repo"
```
