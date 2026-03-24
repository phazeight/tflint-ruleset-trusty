# trusty_variable_field_order

Ensures that attributes within each `variable` block in `inputs.tf` follow
the canonical order: `type`, `default`, `description`.

## Configuration

This rule requires no configuration. It is enabled by default.

## Examples

### Good

```hcl
# inputs.tf
variable "name" {
  type        = string
  description = "Name of the resource"
}

variable "force_delete" {
  type        = bool
  default     = false
  description = "If true, will delete the repository even if it contains images."
}
```

### Bad

```hcl
# inputs.tf
variable "name" {
  description = "Name of the resource"
  type        = string
}
```

Triggers:

```text
variable "name": "type" should appear before "description" (expected order: type, default, description)
```

## Severity

ERROR

## How it works

1. Iterates over all HCL files and selects only `inputs.tf`
2. For each `variable` block, collects known fields (`type`, `default`, `description`) in source order
3. Checks that the fields appear in the canonical order: `type` first, then `default`, then `description`
4. Unknown fields (e.g., `sensitive`, `nullable`, `validation`) are ignored
5. Emits an error on the first variable that breaks the expected field order
