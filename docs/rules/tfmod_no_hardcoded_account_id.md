# `trusty_no_hardcoded_account_id`

**Severity:** WARNING
**Scope:** static string literals in any expression position

## Summary

Flags string literals that are exactly 12 digits — the shape of an AWS account ID. Account IDs belong in a variable, `data.aws_caller_identity`, or an account-id map input, not baked into config.

This is a **format smell** (Tier 2): tflint cannot resolve the correct reference, so it only nudges. The account ID's real source is usually remote state or a data source, which tflint never sees.

## How it works

Walks every expression via `hclsyntax.VisitAll` and flags static string literals matching `^[0-9]{12}$`. Interpolated strings (`"...${...}..."`) are skipped — they are already partly a reference.

No baked exceptions. Legitimate account-id map locals are silenced with `# tflint-ignore`.

## Examples

### Invalid

```hcl
locals {
  account = "123456789012" # WARNING
}
```

### Valid

```hcl
locals {
  account = data.aws_caller_identity.current.account_id
}
```

## Suppression

```hcl
# tflint-ignore: trusty_no_hardcoded_account_id
account_map = { prod = "123456789012" }
```
