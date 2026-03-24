# trusty_variable_location

Ensures that all `variable` blocks are defined in a file named `inputs.tf`
rather than scattered across other files.

## Configuration

This rule requires no configuration. It is enabled by default.

## Examples

### Good

All variables live in `inputs.tf`:

```hcl
# inputs.tf
variable "name" {
  description = "Name of the resource"
  type        = string
}

variable "region" {
  default     = "us-east-1"
  description = "AWS region"
  type        = string
}
```

### Bad

Variables defined in `main.tf` (or any file other than `inputs.tf`):

```hcl
# main.tf
variable "name" {
  description = "Name of the resource"
  type        = string
}
```

Triggers:

```text
variable "name" should be defined in inputs.tf, not main.tf
```

## Severity

ERROR

## How it works

1. Iterates over all HCL files known to the runner
2. Skips files named `inputs.tf`
3. For every other file, checks for `variable` blocks
4. Emits an error for each variable block found outside `inputs.tf`
