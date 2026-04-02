package komodor

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorMonitor() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the monitor.",
				ValidateFunc: validation.NoZeroValues,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The monitor type. Must be one of: `availability`, `node`, `PVC`, `job`, `cronJob`, `deploy`, or `workflow`.",
				ValidateFunc: validation.NoZeroValues,
			},
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Indicates whether the monitor is enabled.",
			},
			"sensors": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "JSON-encoded list defining the scope of monitoring (clusters, namespaces, services, etc.).",
				DiffSuppressFunc: jsonDiffSuppress,
			},
			"variables": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "JSON-encoded additional settings required for specific monitor types (e.g., `duration`, `categories`, `nodeCreationThreshold`).",
				DiffSuppressFunc: jsonDiffSuppress,
			},
			"sinks": {
				Type:             schema.TypeString,
				Elem:             schema.TypeString,
				Optional:         true,
				Description:      "JSON-encoded notification channels for the monitor (e.g., Slack, Teams, PagerDuty, Opsgenie, Webhook).",
				DiffSuppressFunc: jsonDiffSuppress,
			},
			"sinks_options": {
				Type:             schema.TypeString,
				Elem:             schema.TypeString,
				Optional:         true,
				Description:      "JSON-encoded additional notification settings such as `notifyOn`. Valid values depend on the monitor type.",
				DiffSuppressFunc: jsonDiffSuppress,
			},
			"is_deleted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the monitor has been marked for deletion. Defaults to `false`.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		CreateContext: resourceKomodorMonitorCreate,
		ReadContext:   resourceKomodorMonitorRead,
		UpdateContext: resourceKomodorMonitorUpdate,
		DeleteContext: resourceKomodorMonitorDelete,
		Description:   "Creates a new Komodor monitor which allows Komodor to monitor, detect, and analyze failures around infrastructure.",
	}
}

func resourceKomodorMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	sensorsJson := d.Get("sensors")
	var sensors []Sensor
	if err := json.Unmarshal([]byte(sensorsJson.(string)), &sensors); err != nil {
		return diag.Errorf("Error creating sensor statement structure: %s", err)
	}

	newMonitor := &NewMonitor{
		Name:      d.Get("name").(string),
		Type:      d.Get("type").(string),
		Active:    d.Get("active").(bool),
		Sensors:   sensors,
		IsDeleted: d.Get("is_deleted").(bool),
	}

	if variablesJson, variablesJsonExists := d.GetOk("variables"); variablesJsonExists {
		var variables ModelWorkflowConfigurationVariables
		if err := json.Unmarshal([]byte(variablesJson.(string)), &variables); err != nil {
			return diag.Errorf("Error creating variables statement structure: %s", err)
		}
		newMonitor.Variables = &variables
	}

	if sinksJson, sinksJsonExists := d.GetOk("sinks"); sinksJsonExists {
		var sinks Sinks
		if err := json.Unmarshal([]byte(sinksJson.(string)), &sinks); err != nil {
			return diag.Errorf("Error creating sinks statement structure: %s, %s", err, sinksJson)
		}
		newMonitor.Sinks = &sinks
	}

	if sinksOptionsJson, sinksOptionsJsonExists := d.GetOk("sinks_options"); sinksOptionsJsonExists {
		var sinksOptions SinkOptions
		if err := json.Unmarshal([]byte(sinksOptionsJson.(string)), &sinksOptions); err != nil {
			return diag.Errorf("Error creating sinks options statement structure: %s", err)
		}
		newMonitor.SinksOptions = &sinksOptions
	}

	monitor, err := client.CreateMonitor(newMonitor)
	if err != nil {
		return diag.Errorf("Error creating monitor: %s", err)
	}

	d.SetId(monitor.Id)
	log.Printf("[INFO] Monitor created successfully. Monitor Id: %s", monitor.Id)

	return resourceKomodorMonitorRead(ctx, d, meta)
}

func resourceKomodorMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	monitor, statusCode, err := client.GetMonitor(id)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Monitor (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Monitor: %s", err)
	}

	if monitor.Name != nil {
		if err := d.Set("name", *monitor.Name); err != nil {
			return diag.Errorf("error setting name: %s", err)
		}
	}
	if err := d.Set("active", monitor.Active); err != nil {
		return diag.Errorf("error setting active: %s", err)
	}
	if monitor.IsDeleted != nil {
		if err := d.Set("is_deleted", *monitor.IsDeleted); err != nil {
			return diag.Errorf("error setting is_deleted: %s", err)
		}
	}
	if err := d.Set("type", monitor.Type); err != nil {
		return diag.Errorf("error setting type: %s", err)
	}

	if sensorsJSON, err := json.Marshal(monitor.Sensors); err == nil {
		if err := d.Set("sensors", string(sensorsJSON)); err != nil {
			return diag.Errorf("error setting sensors: %s", err)
		}
	}
	if monitor.Sinks != nil {
		if sinksJSON, err := json.Marshal(monitor.Sinks); err == nil {
			if err := d.Set("sinks", string(sinksJSON)); err != nil {
				return diag.Errorf("error setting sinks: %s", err)
			}
		}
	}
	if monitor.SinkOptions != nil {
		if sinksOptionsJSON, err := json.Marshal(monitor.SinkOptions); err == nil {
			if err := d.Set("sinks_options", string(sinksOptionsJSON)); err != nil {
				return diag.Errorf("error setting sinks_options: %s", err)
			}
		}
	}
	if monitor.Variables != nil {
		if variablesJSON, err := json.Marshal(monitor.Variables); err == nil {
			if err := d.Set("variables", string(variablesJSON)); err != nil {
				return diag.Errorf("error setting variables: %s", err)
			}
		}
	}

	return nil
}

func resourceKomodorMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	sensorsJson := d.Get("sensors")
	var sensors []Sensor
	if err := json.Unmarshal([]byte(sensorsJson.(string)), &sensors); err != nil {
		return diag.Errorf("Error creating sensor statement structure: %s", err)
	}

	newMonitor := &NewMonitor{
		Name:      d.Get("name").(string),
		Type:      d.Get("type").(string),
		Active:    d.Get("active").(bool),
		Sensors:   sensors,
		IsDeleted: d.Get("is_deleted").(bool),
	}
	if variablesJson, variablesJsonExists := d.GetOk("variables"); variablesJsonExists {
		var variables ModelWorkflowConfigurationVariables
		if err := json.Unmarshal([]byte(variablesJson.(string)), &variables); err != nil {
			return diag.Errorf("Error creating variables statement structure: %s", err)
		}
		newMonitor.Variables = &variables
	}

	if sinksJson, sinksJsonExists := d.GetOk("sinks"); sinksJsonExists {
		var sinks Sinks
		if err := json.Unmarshal([]byte(sinksJson.(string)), &sinks); err != nil {
			return diag.Errorf("Error creating sinks statement structure: %s", err)
		}
		newMonitor.Sinks = &sinks
	}

	if sinksOptionsJson, sinksOptionsJsonExists := d.GetOk("sinks_options"); sinksOptionsJsonExists {
		var sinksOptions SinkOptions
		if err := json.Unmarshal([]byte(sinksOptionsJson.(string)), &sinksOptions); err != nil {
			return diag.Errorf("Error creating sinks options statement structure: %s", err)
		}
		newMonitor.SinksOptions = &sinksOptions
	}

	_, err := client.UpdateMonitor(id, newMonitor)
	if err != nil {
		return diag.Errorf("Error updating monitor: %s", err)
	}
	return resourceKomodorMonitorRead(ctx, d, meta)
}

func resourceKomodorMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Monitor: %s", id)
	if err := client.DeleteMonitor(id); err != nil {
		return diag.Errorf("Error deleting monitor: %s", err)
	}

	d.SetId("")
	return nil
}
