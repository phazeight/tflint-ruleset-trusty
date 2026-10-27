# `trusty_positive_variable_names`

**Severity:** ERROR
**Scope:** `variable` blocks with `type = bool`

## Summary

Boolean variables must be named positively. A `bool` whose name starts with `disable_`, `no_`, `not_`, or `prevent_` is flagged.

```hcl
variable "disable_encryption" { type = bool }   # flagged
variable "encryption_enabled" { type = bool }   # preferred
```

## Background

Negative booleans force a double negative at every use site. `disable_encryption = false` means encryption is on, which reads backwards and is easy to misconfigure under review.

They also make the safe default the *falsy* one, so an omitted or zero-valued variable turns protection off. `encryption_enabled` defaults to a state you can reason about, and `count = var.encryption_enabled ? 1 : 0` reads the way the resource behaves.

## Rule checks

| Check | Trigger | Message |
|-------|---------|---------|
| Negative prefix | `type = bool` and the variable name begins with a negative prefix | `variable "<name>": use positive naming (avoid "<prefix>" prefix)` |

Flagged prefixes: `disable_`, `no_`, `not_`, `prevent_`.

At most one issue is emitted per variable — the first matching prefix wins.

## Scope

The `type` must be exactly `bool` (after trimming whitespace). Ignored:

- Variables with no `type` attribute
- Any other type — including `optional(bool)` and `list(bool)`
- Negative wording that is not a leading prefix — `allow_no_public_access` is not flagged, since the prefix match is anchored to the start of the name

## Examples

### Valid

```hcl
variable "encryption_enabled" {
  type        = bool
  default     = true
  description = "Whether to enable server-side encryption."
}

variable "public_access_allowed" {
  type        = bool
  default     = false
  description = "Whether to allow public access."
}
```

### Invalid

```hcl
variable "disable_encryption" {
  type        = bool
  default     = false
  description = "Whether to disable server-side encryption."
  # ERROR: variable "disable_encryption": use positive naming (avoid "disable_" prefix)
}
```

Rename to `encryption_enabled` and invert the default. Remember to invert every use site — the rule cannot do that for you, and a rename without inverting the logic silently flips behaviour.

## Suppressing a finding

```hcl
# tflint-ignore: trusty_positive_variable_names
variable "no_proxy" {
  type        = bool
  description = "Provider-mandated name; cannot be renamed."
}
```
