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
				ValidateFunc: validation.NoZeroValues,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"active": {
				Type:         schema.TypeBool,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"sensors": {
				Type:     schema.TypeString,
				Required: true,
			},
			"variables": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
			},
			"sinks": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
			},
			"sinks_options": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
			},
			"is_deleted": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceKomodorMonitorCreate,
		ReadContext:   resourceKomodorMonitorRead,
		UpdateContext: resourceKomodorMonitorUpdate,
		DeleteContext: resourceKomodorMonitorDelete,
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
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Active:      d.Get("active").(bool),
		Sensors:     sensors,
		Variables:   d.Get("variables").(map[string]interface{}),
		Sinks:       d.Get("sinks").(map[string]interface{}),
		SinkOptions: d.Get("sinks_options").(map[string]interface{}),
		IsDeleted:   d.Get("is_deleted").(bool),
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

	monitor, err := client.GetMonitor(id)
	if err != nil {
		if err.Error() == "404" {
			log.Printf("[DEBUG] Monitor (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Monitor: %s", err)
	}

	d.Set("name", monitor.Name)
	d.Set("active", monitor.Active)
	d.Set("created_at", monitor.CreatedAt)
	d.Set("updated_at", monitor.UpdatedAt)
	d.Set("is_deleted", monitor.IsDeleted)
	d.Set("sensors", monitor.Sensors)
	d.Set("sink_options", monitor.SinkOptions)
	d.Set("sinks", monitor.Sinks)
	d.Set("type", monitor.Type)
	d.Set("variables", monitor.Variables)

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
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Active:      d.Get("active").(bool),
		Sensors:     sensors,
		Variables:   d.Get("variables").(map[string]interface{}),
		Sinks:       d.Get("sinks").(map[string]interface{}),
		SinkOptions: d.Get("sinks_options").(map[string]interface{}),
		IsDeleted:   d.Get("is_deleted").(bool),
	}

	_, err := client.UpdateMonitor(id, newMonitor)

	if err != nil {
		return diag.Errorf("Error updating monitor: %s", err)
	}
	return resourceKomodorMonitorRead(ctx, d, meta)
}

func resourceKomodorMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Delete monitor is not implemented in api
	return nil
}
