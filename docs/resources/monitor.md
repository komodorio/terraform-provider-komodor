---
page_title: "komodor_monitor Resource - terraform-provider-komodor"
subcategory: ""
description: |-
  Creates a new Komodor monitor which allows Komodor
  to monitor, detect, and analyze failures around infrastructure.
---

# komodor_monitor (Resource)

Creates a new **Komodor monitor** to observe, detect, and analyze failures across your infrastructure. This resource allows you to define the parameters for monitoring specific components in your Kubernetes clusters and other infrastructure.

---

## Example Usage

### Deployment Monitor

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `webhook`
- **Valid notifyOn Options**:
  - `"Failure"`
  - `"Successful"`
  - `"All"`

```terraform
resource "komodor_monitor" "example-deploy-monitor" {
  name          = "example-deploy-monitor"
  type          = "deploy"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "namespaces": ["komodor"]
  },
  "namespaces": ["default"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["deployment-alerts"],
  "teams": ["Platform-Team"]
}
EOF
  sinks_options = <<EOF
{
  "notifyOn": ["Failure", "Successful"]
}
EOF
}
```

---

### Availability Monitor

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid Duration**: Must be an integer between **5** and **600** (inclusive).
- **Valid Categories for variables.categories**:
  - `"Creating/Initializing"`
  - `"Scheduling"`
  - `"Container Creation"`
  - `"NonZeroExitCode"`
  - `"Unhealthy - failed probes"`
  - `"OOMKilled"`
  - `"BackOff"`
  - `"Infrastructure"`
  - `"Image"`
  - `"Volume/Secret/ConfigMap"`
  - `"Pod Termination"`
  - `"Completed"`
  - `"Other"`
- **Valid notifyOn Options**: `["*"]`
  - `["*"]` (all categories)
  - Or any one of the following:
    - `"Creating/Initializing"`
    - `"Scheduling"`
    - `"Container Creation"`
    - `"NonZeroExitCode"`
    - `"Unhealthy - failed probes"`
    - `"OOMKilled"`
    - `"BackOff"`
    - `"Infrastructure"`
    - `"Image"`
    - `"Volume/Secret/ConfigMap"`
    - `"Pod Termination"`
    - `"Completed"`
    - `"Other"`

```terraform
resource "komodor_monitor" "example-availability-monitor" {
  name          = "example-availability-monitor"
  type          = "availability"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "services": ["default/excluded-service"]
  },
  "services": ["default/important-service"],
  "condition": "and",
  "namespaces": ["default"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["availability-alerts"],
  "teams": ["SRE-Team"]
}
EOF
  variables     = <<EOF
{
  "categories": ["Creating/Initializing", "Unhealthy - failed probes"],
  "duration": 30,
  "minAvailable": "100%"
}
EOF
  sinks_options = <<EOF
{
  "notifyOn": ["*"]
}
EOF
}
```

---

### Node Monitor

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid Duration**: Must be an integer between **5** and **600** (inclusive).
- **Valid NodeCreationThreshold: Should be in the format of `"3m"` (3 minutes) or `"5s"` (5 seconds).
- **Valid Variables**:
  - `duration`: Required
  - `nodeCreationThreshold`: Required

```terraform
resource "komodor_monitor" "example-node-monitor" {
  name          = "example-node-monitor"
  type          = "node"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind"
}]
EOF
  sinks         = <<EOF
{
  "slack": ["node-alerts"]
}
EOF
  variables     = <<EOF
{
  "duration": 60,
  "nodeCreationThreshold": "10m"
}
EOF
}
```

---

### Workflow Monitor

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid notifyOn Options**: None (not applicable for workflows)

```terraform
resource "komodor_monitor" "example-workflow-monitor" {
  name          = "example-workflow-monitor"
  type          = "workflow"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["workflow-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["workflow-alerts"],
  "webhook": ["webhook-url"]
}
EOF
}
```

---

## Schema

### Required

- `active` (Boolean): Indicates whether the monitor is enabled.
- `name` (String): The name of the monitor. Defaults to an empty string if not provided.
- `sensors` (String): Defines the scope of monitoring (e.g., cluster, namespaces, services, etc.).
- `type` (String): The monitor type. Must be one of: `availability`, `node`, `PVC`, `job`, `cronJob`, `deploy`, or `workflow`.

### Optional

- `is_deleted` (Boolean): Default is `false`. Indicates whether the monitor has been marked for deletion.
- `sinks` (String): Defines notification channels for the monitor, such as Slack, Teams, PagerDuty, Opsgenie, or Webhook.
- `sinks_options` (String): Specifies additional notification settings like notifyOn. Valid values depend on the monitor type.
- `variables` (String): Additional settings required for specific monitor types.

### Read-Only

- `id` (String): The ID of this resource.
