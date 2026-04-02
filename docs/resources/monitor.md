---
page_title: "komodor_monitor Resource - terraform-provider-komodor"
subcategory: ""
description: |-
  Creates a new Komodor monitor which allows Komodor to monitor, detect, and analyze failures around infrastructure.
---

# komodor_monitor (Resource)

Creates a new Komodor monitor which allows Komodor to monitor, detect, and analyze failures around infrastructure.

## Example Usage

### Deployment Monitor

#### Valid Configurations:

- **Valid Sinks**: `slack`, `teams`, `webhook`
- **Valid notifyOn Options**: Only one option is valid:
  - `"Failure"`
  - `"Successful"`
  - `"All"`

```terraform
resource "komodor_monitor" "example-availability-monitor" {
  name          = "example-availability-monitor"
  type          = "availability"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "exclude": {
    "services": ["default/service-to-exclude"]
  },
  "services": [
    "default/service-to-include"
  ],
  "condition": "and",
  "namespaces": ["default"]
}]
EOF
  sinks         = <<EOF
{
  "slack": [
    "default"
  ],
  "teams": [
    "default"
  ],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-pagerduty-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables     = <<EOF
{
  "categories": [
    "*"
  ],
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
  - Or any one of the following: Each notifyOn categories must be included in the "Categories" field
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

### Node Monitor

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

#### Valid Configurations:

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

#### Valid Configurations:

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

### PVC Monitor

```terraform
resource "komodor_monitor" "example-pvc-monitor" {
  name          = "example-pvc-monitor"
  type          = "PVC"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["storage-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["storage-alerts"],
  "teams": ["Storage-Team"],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables     = <<EOF
{
  "duration": 300
}
EOF
}
```

#### Valid Configurations:

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid Duration**: Must be an integer between **5** and **600** (inclusive).
- **Valid NodeCreationThreshold**: Not applicable for PVC.
- **Valid Categories for variables.categories**: Not applicable for PVC.
- **Valid notifyOn Options**: Not applicable for PVC.

```terraform
resource "komodor_monitor" "example-pvc-monitor" {
  name          = "example-pvc-monitor"
  type          = "PVC"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["storage-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["storage-alerts"],
  "teams": ["Storage-Team"],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables     = <<EOF
{
  "duration": 300
}
EOF
}
```

---

### CronJob Monitor

#### Valid Configurations
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid Duration**: Must be an integer between **5** and **600** (inclusive).
- **Valid CronJobCondition**: `"first"` or `"any"`.

```terraform
resource "komodor_monitor" "example-cronjob-monitor" {
  name          = "example-cronjob-monitor"
  type          = "cronJob"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["jobs-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["cronjob-alerts"],
  "teams": ["SRE-Team"],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables     = <<EOF
{
  "duration": 120,
  "cronJobCondition": "first"
}
EOF
}
```

### Job Monitor

#### Valid Configurations:
- **Valid Sinks**: `slack`, `teams`, `opsgenie`, `pagerduty`, `webhook`
- **Valid Duration**: Must be an integer between **5** and **600** (inclusive).

```terraform
resource "komodor_monitor" "example-job-monitor" {
  name          = "example-job-monitor"
  type          = "job"
  active        = true
  sensors       = <<EOF
[{
  "cluster": "kind-kind",
  "namespaces": ["job-namespace"]
}]
EOF
  sinks         = <<EOF
{
  "slack": ["job-alerts"],
  "teams": ["SRE-Team"],
  "pagerduty": [{
    "channel": "example-channel",
    "integrationKey": "example-integration-key",
    "pagerDutyAccountName": "example-pagerduty-account-name"
  }]
}
EOF
  variables     = <<EOF
{
  "duration": 300
}
EOF
}
```

## Argument Reference

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `active` (Boolean) Indicates whether the monitor is enabled.
- `name` (String) The name of the monitor.
- `sensors` (String) JSON-encoded list defining the scope of monitoring (clusters, namespaces, services, etc.).
- `type` (String) The monitor type. Must be one of: `availability`, `node`, `PVC`, `job`, `cronJob`, `deploy`, or `workflow`.

### Optional

- `is_deleted` (Boolean) Indicates whether the monitor has been marked for deletion. Defaults to `false`.
- `sinks` (String) JSON-encoded notification channels for the monitor (e.g., Slack, Teams, PagerDuty, Opsgenie, Webhook).
- `sinks_options` (String) JSON-encoded additional notification settings such as `notifyOn`. Valid values depend on the monitor type.
- `variables` (String) JSON-encoded additional settings required for specific monitor types (e.g., `duration`, `categories`, `nodeCreationThreshold`).

### Read-Only

- `id` (String) The ID of this resource.

## Import

This resource can be imported using the monitor ID:

```sh
terraform import komodor_monitor.example <monitor_id>
```
