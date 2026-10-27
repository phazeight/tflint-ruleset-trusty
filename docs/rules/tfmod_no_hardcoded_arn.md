# `trusty_no_hardcoded_arn`

**Severity:** WARNING
**Scope:** static string literals in any expression position

## Summary

Flags string literals that look like an AWS ARN (`^arn:aws[a-z-]*:`, covering `aws`, `aws-cn`, `aws-us-gov` partitions). ARNs of resources managed elsewhere should reference the owning resource's `.arn` attribute, or a variable/data source.

This is a **format smell** (Tier 2): tflint cannot connect the literal to its owning resource, so it only nudges.

## How it works

Walks every expression via `hclsyntax.VisitAll` and flags static string literals matching `^arn:aws[a-z-]*:`. Interpolated strings are skipped.

### Baked exception

`arn:aws:iam::aws:*` (AWS-managed IAM policies and roles, e.g. `arn:aws:iam::aws:policy/AdministratorAccess`) is always allowed — there is no owning resource in the config to reference.

## Examples

### Invalid

```hcl
locals {
  bucket_arn = "arn:aws:s3:::my-bucket" # WARNING
}
```

### Valid

```hcl
locals {
  bucket_arn = aws_s3_bucket.this.arn
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess" # baked exception
}
```

## Suppression

```hcl
# tflint-ignore: trusty_no_hardcoded_arn
some_arn = "arn:aws:service:region:account:resource"
```
