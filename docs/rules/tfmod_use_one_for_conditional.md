# `trusty_use_one_for_conditional`

**Severity:** ERROR
**Scope:** `[0]` index expressions on resource, data source, and module references

## Summary

Flags `[0]` indexing on a count-conditional object and suggests `one(...)` instead:

```hcl
aws_instance.web[0].id        # flagged
one(aws_instance.web[*].id)   # suggested
```

## Background

A resource with `count = var.enabled ? 1 : 0` is a **list** — one element when enabled, zero when not. Reading it as `aws_instance.web[0].id` works only while the count is 1. When the count is 0 the same expression fails the plan:

```
Error: Invalid index
  The given key does not identify an element in this collection value.
```

The failure is latent: it appears the first time someone flips the flag off, often in an environment nobody was testing. `one()` returns `null` for an empty list and the single element for a one-element list, so the expression degrades instead of erroring.

`one()` also fails loudly if the list ever has more than one element, which is the correct behaviour for a `0..1` conditional — unlike `[0]`, which would silently pick the first of many.

## Rule checks

| Check | Trigger | Message |
|-------|---------|---------|
| `[0]` on a resource | `<type>.<name>[0]` | `use one() instead of [0] indexing on "<ref>" — one(<ref>[*].<attr>) returns null safely when count = 0` |
| `[0]` on a data source | `data.<type>.<name>[0]` | as above |
| `[0]` on a module call | `module.<name>[0]` | as above |

The suggestion drops the `.<attr>` half when the traversal ends at the index (`one(<ref>[*])`).

## Scope

The `[0]` must index the **object itself** — immediately after `<type>.<name>`, `module.<name>`, or `data.<type>.<name>`. An index deeper in the traversal is indexing a list-typed *attribute*, which is a normal list access with no `count` involved and is **ignored**:

```hcl
# ignored — roots is a list attribute, not a count-conditional resource
root_id = aws_organizations_organization.root.roots[0].id
```

Also ignored:

- Non-resource roots: `var`, `local`, `path`, `terraform`, `each`, `count`, `self`
- Any index other than literal `0` — `aws_instance.web[1].id`, `aws_instance.web[each.key].id`
- Splat expressions — `aws_instance.web[*].id`
- The bodies of `moved`, `import`, and `removed` blocks — Terraform requires a literal resource address there, so `one()` is not valid syntax and the `[0]` cannot be rewritten

## Examples

### Valid

```hcl
locals {
  # count-conditional — one() returns null when count = 0
  cert_arn = one(aws_acm_certificate_validation.wildcard[*].certificate_arn)
}

moved {
  # literal address required — not flagged
  from = aws_instance.old
  to   = aws_instance.web[0]
}

locals {
  # list attribute, not a count — not flagged
  root_id = aws_organizations_organization.root.roots[0].id
}
```

### Invalid — resource

```hcl
resource "aws_acm_certificate_validation" "wildcard" {
  count = var.enable_https ? 1 : 0
  # ...
}

resource "aws_lb_listener" "https" {
  certificate_arn = aws_acm_certificate_validation.wildcard[0].certificate_arn
  # ERROR: use one() instead of [0] indexing on "aws_acm_certificate_validation.wildcard"
  #        — one(aws_acm_certificate_validation.wildcard[*].certificate_arn) returns null safely when count = 0
}
```

### Invalid — module

```hcl
module "merge_queue_ruleset" {
  count  = var.require_merge_queue ? 1 : 0
  source = "..."
}

output "ruleset_id" {
  value = module.merge_queue_ruleset[0].id
  # ERROR: use one() instead of [0] indexing on "module.merge_queue_ruleset"
  #        — one(module.merge_queue_ruleset[*].id) returns null safely when count = 0
}
```

Fix with `one(module.merge_queue_ruleset[*].id)`, which yields `null` when `require_merge_queue` is false. Note that a ternary guard — `var.require_merge_queue ? module.merge_queue_ruleset[0].id : null` — is still flagged, and deliberately so: the guard duplicates the count condition, and the two drift apart the moment one of them is edited.

## Suppressing a finding

Use tflint's native annotation on the preceding line:

```hcl
# tflint-ignore: trusty_use_one_for_conditional
value = aws_instance.web[0].id
```
