resource "komodor_cost_right_sizing_policy" "production_conservative" {
  # step 1 - name
  name        = "production-conservative"
  description = "Conservative right-sizing for prod workloads"
  priority    = 1000

  # step 2 - scope
  scope {
    clusters       = ["prod-us-east-1", "prod-eu-west-1"]
    namespaces     = ["payments", "checkout"]
    resource_types = ["Deployment", "StatefulSet"]
    workload_names_patterns {
      include = "*"
    }
  }

  # step 3 - when to apply
  apply_protocol         = "onCreation"
  allow_restart          = false
  allow_hpa_right_sizing = false

  # step 4 - guardrails
  optimization_preset = "custom"

  guardrails {
    percentile = 95

    managed_resources {
      cpu_requests    = true
      cpu_limits      = false
      memory_requests = true
      memory_limits   = false
    }

    allow_right_sizing_up = true
    allow_qos_upgrade     = false
    allow_qos_downgrade   = false

    constraints {
      increase_cpu_by {
        enabled = true
        value   = 50
      }
      decrease_cpu_by {
        enabled = true
        value   = 20
      }
      increase_memory_by {
        enabled = true
        value   = 50
      }
      decrease_memory_by {
        enabled = true
        value   = 20
      }
    }

    # Absolute constraints — units: CPU in millicores, memory in bytes.
    absolute_constraints {
      cpu_request_millicores_min {
        enabled = true
        value   = 100
      }
      cpu_request_millicores_max {
        enabled = true
        value   = 4000
      }
      cpu_limits_millicores_min {
        enabled = false
        value   = 0
      }
      cpu_limits_millicores_max {
        enabled = false
        value   = 0
      }
      memory_request_bytes_min {
        enabled = true
        value   = 134217728 # 128 MiB
      }
      memory_request_bytes_max {
        enabled = true
        value   = 8589934592 # 8 GiB
      }
      memory_limits_bytes_min {
        enabled = false
        value   = 0
      }
      memory_limits_bytes_max {
        enabled = false
        value   = 0
      }
    }

    buffer {
      cpu {
        enabled = true
        value   = 10
      }
      memory {
        enabled = true
        value   = 15
      }
    }
  }
}
