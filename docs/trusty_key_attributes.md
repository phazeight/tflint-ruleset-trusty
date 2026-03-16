# trusty_key_attributes

Ensures that key attributes (those that uniquely identify a resource) appear
at the top of each `resource` and `data` block definition, in priority order.

## Configuration

The rule ships with default key-attribute definitions for 140+ resource types
spanning Google Cloud, AWS, Kubernetes, Helm, Vault, and other providers.

Additional resource types can be configured in `.tflint.hcl`:

```hcl
plugin "trusty" {
  enabled = true

  resource "my_custom_resource" {
    key_attributes = ["project", "name"]
  }
}
```

For resources with nested key blocks (e.g., Kubernetes `metadata`), use
dot notation:

```hcl
resource "kubernetes_custom_resource" {
  key_attributes = ["metadata.namespace", "metadata.name"]
}
```

## Examples

### Good

```hcl
resource "google_compute_instance" "web" {
  project      = var.project_id
  zone         = var.zone
  name         = "web-server"
  machine_type = "e2-medium"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }
}
```

### Bad

```hcl
resource "google_compute_instance" "web" {
  machine_type = "e2-medium"
  name         = "web-server"
  project      = var.project_id
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }
}
```

The bad example triggers:

```text
key-attribute `project` should be defined before non-key `machine_type`
```

## Severity

ERROR

## How it works

1. Iterates over all `resource` and `data` blocks in HCL files
2. Looks up the resource type in the configured key-attributes map
3. For resources with nested key blocks (e.g., Kubernetes `metadata`),
   traverses into the nested block body
4. Verifies that key attributes appear before non-key attributes and
   in the correct priority order
5. Emits an error if ordering is violated
