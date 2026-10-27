# `trusty_no_hardcoded_resource_id`

**Severity:** WARNING
**Scope:** static string literals in any expression position

## Summary

Flags string literals that look like an AWS resource ID for a tight set of prefixes. Resource IDs should come from a data source or resource reference, not be hardcoded.

This is a **format smell** (Tier 2): the ID's real source is usually remote state or a data source, which tflint never resolves — so it only nudges.

## How it works

Walks every expression via `hclsyntax.VisitAll` and flags static string literals matching:

```
^(vpc|subnet|sg|ami|igw|rtb|acl|eni|nat|eipalloc|pcx)-[0-9a-f]{8,}$
```

Interpolated strings are skipped. The prefix set is deliberately tight to start; expand it later as false-positive experience allows.

## Examples

### Invalid

```hcl
locals {
  vpc = "vpc-0a1b2c3d4e5f6a7b8" # WARNING
}
```

### Valid

```hcl
locals {
  vpc = data.aws_vpc.this.id
}
```

## Suppression

```hcl
# tflint-ignore: trusty_no_hardcoded_resource_id
vpc_id = "vpc-0a1b2c3d4e5f6a7b8"
```
