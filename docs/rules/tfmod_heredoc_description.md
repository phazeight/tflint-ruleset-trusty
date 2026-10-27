# `trusty_heredoc_description`

**Severity:** ERROR
**Scope:** `variable` blocks whose `type` contains `object(` with 2 or more attributes

## Summary

A `variable` with a multi-attribute `object(...)` type must write its `description` as a heredoc (`<<-DOC`), not a single-line string.

## Background

An object-typed variable needs per-attribute documentation — a one-line description cannot explain five keys. Heredoc bodies survive into `terraform-docs` output as formatted markdown, so the rendered README gets a readable table or list instead of one long run-on sentence.

The threshold is 2 attributes because a single-attribute object is effectively a scalar and a one-liner is adequate.

## Rule checks

| Check | Trigger | Message |
|-------|---------|---------|
| Single-line description | `type` source contains `object(`, has ≥2 ` = ` assignments, and `description` does not start with `<` | `variable "<name>": object type with ≥2 attributes must use <<-DOC heredoc for description` |

Detection is source-text based: the rule reads the raw bytes of the `type` expression, requires the literal substring `object(`, and counts occurrences of ` = ` to approximate the attribute count. A `description` is considered a heredoc when its first byte is `<` (covering both `<<DOC` and `<<-DOC`).

## Scope

Ignored:

- Variables with no `type` attribute
- Variables with no `description` attribute (a missing description is out of scope for this rule)
- Non-object types — `string`, `bool`, `list(string)`, `map(string)`
- `object(...)` types with fewer than 2 attributes

## Examples

### Valid

```hcl
variable "logging" {
  type = object({
    bucket        = string
    prefix        = string
    target_prefix = optional(string, "logs/")
  })
  description = <<-DOC
    Access-logging configuration.

    - `bucket`        — destination bucket name
    - `prefix`        — key prefix for delivered logs
    - `target_prefix` — prefix applied to the target bucket
  DOC
}
```

```hcl
# single-attribute object — one-liner is fine
variable "tags_config" {
  type        = object({ enabled = bool })
  description = "Whether to apply default tags."
}
```

```hcl
# not an object type — out of scope
variable "bucket_name" {
  type        = string
  description = "Name of the S3 bucket."
}
```

### Invalid

```hcl
variable "logging" {
  type = object({
    bucket = string
    prefix = string
  })
  description = "Access-logging configuration."
  # ERROR: variable "logging": object type with ≥2 attributes must use <<-DOC heredoc for description
}
```

## Suppressing a finding

```hcl
variable "logging" {
  type = object({
    bucket = string
    prefix = string
  })
  # tflint-ignore: trusty_heredoc_description
  description = "Access-logging configuration."
}
```
