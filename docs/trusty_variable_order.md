# trusty_variable_order

Ensures that `variable` blocks within `inputs.tf` are sorted alphabetically
(A–Z) by variable name.

## Configuration

This rule requires no configuration. It is enabled by default.

## Examples

### Good

Variables sorted alphabetically:

```hcl
# inputs.tf
variable "force_delete" {
  default     = false
  description = "If true, will delete the repository even if it contains images."
  type        = bool
}

variable "image_tag_mutability" {
  default     = "MUTABLE"
  description = "The tag mutability setting for the repository."
  type        = string
}

variable "name" {
  description = "Name of the Repo"
  type        = string
}
```

### Bad

Variables out of alphabetical order:

```hcl
# inputs.tf
variable "name" {
  description = "Name of the Repo"
  type        = string
}

variable "force_delete" {
  default     = false
  description = "If true, will delete the repository even if it contains images."
  type        = bool
}
```

Triggers:

```text
variable "force_delete" should be defined before "name" (alphabetical order)
```

## Severity

ERROR

## How it works

1. Iterates over all HCL files and selects only `inputs.tf`
2. Walks through `variable` blocks in source order
3. Compares each variable name (case-insensitive) with the previous one
4. Emits an error on the first variable that breaks alphabetical order
