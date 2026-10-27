# tflint-ruleset-trusty

A custom tflint ruleset enforcing HT Terraform module conventions.

## Rules

| Rule | Severity | Description |
|------|----------|-------------|
| `trusty_heredoc_description` | ERROR | `variable` blocks with a multi-attribute `object(...)` type must use `<<-DOC` heredoc for `description` |
| `trusty_key_attributes` | ERROR | Identifying resource attributes must appear first, followed by remaining attributes sorted A-Z |
| `trusty_managed_repo_ref` | WARNING | `repository` values that match a `managed-repo` module's `repo_name` should reference the module output, not hardcode the string |
| `trusty_module_source` | ERROR | Module sources referencing the repo set in `repo` must use `git::https://`, include `?ref=`, and omit `//subdir` notation |
| `trusty_no_hardcoded_account_id` | WARNING | String literals shaped like a 12-digit AWS account ID should be a variable or `data.aws_caller_identity` |
| `trusty_no_hardcoded_arn` | WARNING | Hardcoded ARNs should reference the owning resource's `.arn` (AWS-managed `arn:aws:iam::aws:*` excepted) |
| `trusty_no_hardcoded_resource_id` | WARNING | String literals shaped like an AWS resource ID (`vpc-`, `subnet-`, `sg-`, …) should come from a data source or reference |
| `trusty_positive_variable_names` | ERROR | `bool` variables must use positive naming — no `disable_`, `no_`, `not_`, `prevent_` prefix |
| `trusty_reserved_variable_names` | ERROR | Variable names Terraform reserves as `module` meta-arguments (`count`, `source`, `version`, …) prevent the module from initialising |
| `trusty_use_one_for_conditional` | ERROR | `[0]` indexing on a count-conditional resource, data source, or module must use `one(...)` instead |
| `trusty_variable_field_order` | ERROR | Fields within each variable block must be ordered: `type`, `default`, `description` |
| `trusty_variable_location` | ERROR | All `variable` blocks must be defined in `inputs.tf` |
| `trusty_variable_order` | ERROR | Variables in `inputs.tf` must be sorted A-Z within each section (required first, then optional) |
| `trusty_variable_section_order` | ERROR | In `inputs.tf`, required and optional variables must be separated by a `DEFAULTS` comment, with required variables before it and optional variables after it |

## Rule Details

### `trusty_heredoc_description`

A `variable` whose `type` contains `object(...)` with 2 or more attributes must write its `description` as a heredoc (`<<-DOC`). One line cannot document five keys, and heredoc bodies carry through to `terraform-docs` output as formatted markdown.

Single-attribute objects and non-object types are out of scope.

See [`docs/rules/trusty_heredoc_description.md`](docs/rules/trusty_heredoc_description.md).

### `trusty_key_attributes`

For resource types with a configured identifying attribute (e.g. `name` for `aws_s3_bucket`, `bucket` + `key` for `aws_s3_object`), the key attribute(s) must appear before all other attributes. Non-key attributes after the key section must be sorted A-Z.

Violation: a non-key attribute appears on an earlier line than the last key attribute, or non-key attributes after the key section are out of alphabetical order.

### `trusty_module_source`

Module `source` values referencing the repo set in `repo` must follow three conventions:

1. **HTTPS scheme** — `git::https://` is required; `git::ssh://` breaks CI without a deploy key
2. **No `//subdir` notation** — `terraform-module-releaser` creates flat-commit tags with module files at repo root; subdirectories don't exist at tag time and `terraform init` will fail
3. **Pinned `?ref=`** — floating sources are non-reproducible and must be rejected

Sources that do not reference the `repo` value (local paths, registry sources, other git repos) are ignored by this rule.

See [`docs/rules/trusty_module_source.md`](docs/rules/trusty_module_source.md) for full details and examples.

### Prefer references over hardcoded values (WARNING)

Four rules discourage literals where a Terraform *reference* belongs. All are `WARNING` severity, so with tflint's default `--minimum-failure-severity=error` they are surfaced but do not fail CI.

- **`trusty_managed_repo_ref`** — a `repository` string literal that matches the `repo_name` of a `managed-repo` module in the same configuration should reference `split("/", module.<x>.repository_full_name)[1]` instead. See [`docs/rules/trusty_managed_repo_ref.md`](docs/rules/trusty_managed_repo_ref.md).
- **`trusty_no_hardcoded_account_id`** — flags exact 12-digit strings (AWS account ID shape). See [`docs/rules/trusty_no_hardcoded_account_id.md`](docs/rules/trusty_no_hardcoded_account_id.md).
- **`trusty_no_hardcoded_arn`** — flags `arn:aws...` strings, excepting AWS-managed `arn:aws:iam::aws:*`. See [`docs/rules/trusty_no_hardcoded_arn.md`](docs/rules/trusty_no_hardcoded_arn.md).
- **`trusty_no_hardcoded_resource_id`** — flags AWS resource-ID-shaped strings (`vpc-`, `subnet-`, `sg-`, `ami-`, …). See [`docs/rules/trusty_no_hardcoded_resource_id.md`](docs/rules/trusty_no_hardcoded_resource_id.md).

The account-id / ARN / resource-id rules are format smells: tflint cannot resolve the correct reference (it usually lives in remote state), so they nudge toward a variable or data source rather than resolving it.

#### Suppressing a finding

There is no per-rule config in v1. Suppress a specific finding with tflint's native annotation on the preceding line:

```hcl
# tflint-ignore: trusty_no_hardcoded_account_id
account_map = { prod = "123456789012" }
```

### `trusty_positive_variable_names`

`bool` variables must be named positively — no `disable_`, `no_`, `not_`, or `prevent_` prefix. A negative boolean forces a double negative at every use site (`disable_encryption = false`) and makes the falsy default the unsafe one.

Only exact `type = bool` is in scope; the prefix match is anchored to the start of the name.

See [`docs/rules/trusty_positive_variable_names.md`](docs/rules/trusty_positive_variable_names.md).

### `trusty_reserved_variable_names`

Terraform reserves `count`, `depends_on`, `for_each`, `lifecycle`, `providers`, `source`, and `version` as meta-arguments inside `module` blocks. A module declaring a variable with one of those names **cannot be initialised at all** — `terraform init` fails with "Invalid variable name". Rename the variable and map it back in `main.tf`.

See [`docs/rules/trusty_reserved_variable_names.md`](docs/rules/trusty_reserved_variable_names.md).

### `trusty_use_one_for_conditional`

A resource with `count = var.enabled ? 1 : 0` is a list. Reading it as `aws_thing.x[0].id` fails the plan with "Invalid index" the first time the count is 0 — a latent break that surfaces whenever someone flips the flag. `one(aws_thing.x[*].id)` returns `null` instead.

Applies to resources, data sources, and `module` calls. The `[0]` must index the object itself: a deeper index is a list-typed *attribute* (`aws_organizations_organization.root.roots[0].id`) and is ignored. Bodies of `moved`, `import`, and `removed` blocks are skipped — Terraform requires a literal address there, so `one()` is not valid.

This is a correctness rule, not a style rule. Consumers that lint hand-authored configuration may still opt out via `.tflint.hcl`.

See [`docs/rules/trusty_use_one_for_conditional.md`](docs/rules/trusty_use_one_for_conditional.md).

### `trusty_variable_field_order`

Within any `variable` block, attributes must appear in the order: `type` → `default` → `description`.

Violation: `type` appears after `default` or `description`; `default` appears after `description`.

### `trusty_variable_location`

All `variable` blocks must be declared in a file named `inputs.tf`.

Violation: a `variable` block is found in any file other than `inputs.tf`.

### `trusty_variable_order`

In `inputs.tf`, required variables (no `default`) must all appear before optional variables (have `default`). Within each group, variables must be sorted alphabetically.

Violation: a required variable appears after an optional one, or variables within a group are out of alphabetical order.

### `trusty_variable_section_order`

In `inputs.tf` files containing both required and optional variables, a separator comment containing the word `DEFAULTS` (e.g. `### DEFAULTS ###`) must be present. All required variables must appear before this separator and all optional variables must appear after it.

Violation: the separator comment is missing; a required variable appears below the separator; an optional variable appears above the separator.
