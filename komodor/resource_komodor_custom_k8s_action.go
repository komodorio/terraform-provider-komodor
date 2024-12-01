package komodor

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorCustomK8sAction() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"ruleset": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceKomodorCustomK8sActionCreate,
		ReadContext:   resourceKomodorCustomK8sActionRead,
		UpdateContext: resourceKomodorCustomK8sActionUpdate,
		DeleteContext: resourceKomodorCustomK8sActionDelete,
		Description:   "Creates a new Komodor RBAC action",
	}
}

func resourceKomodorCustomK8sActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	actionName := d.Get("action").(string)
	description := d.Get("description").(string)
	jsonRuleset := d.Get("ruleset")

	var customK8sActionStatements []CustomK8sActionStatement
	if err := json.Unmarshal([]byte(jsonRuleset.(string)), &customK8sActionStatements); err != nil {
		return diag.Errorf("Error creating statement structure: %s", err)
	}

	newCustomK8sAction := &NewCustomK8sAction{
		Action:      actionName,
		Description: description,
		Ruleset:     customK8sActionStatements,
	}

	customK8sAction, statusCode, err := client.CreateCustomK8sAction(newCustomK8sAction)
	if err != nil {
		if statusCode == 409 {
			return diag.Errorf("Action name '%s' is already in use or marked for deletion. Please choose a different name or try again later", actionName)
		}
		return diag.Errorf("Error creating Custom K8S Action: %s, %v", err, customK8sActionStatements)
	}

	d.SetId(customK8sAction.Id)
	log.Printf("[INFO] CustomK8sAction created successfully. CustomK8sAction Id: %s", customK8sAction.Id)

	return resourceKomodorCustomK8sActionRead(ctx, d, meta)
}

func resourceKomodorCustomK8sActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Get("action").(string)

	customK8sAction, statusCode, err := client.GetCustomK8sAction(id)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] CustomK8sAction (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading CustomK8sAction: %s", err)
	}

	d.Set("action", customK8sAction.Action)
	d.Set("description", customK8sAction.Description)
	d.Set("ruleset", customK8sAction.Ruleset)
	d.Set("created_at", customK8sAction.CreatedAt)
	d.Set("updated_at", customK8sAction.UpdatedAt)

	return nil
}

func resourceKomodorCustomK8sActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()
	description := d.Get("description").(string)
	jsonCustomK8sAction := d.Get("ruleset")

	var customK8sActions []CustomK8sActionStatement
	if err := json.Unmarshal([]byte(jsonCustomK8sAction.(string)), &customK8sActions); err != nil {
		return diag.Errorf("Error creating statement structure: %s", err)
	}

	newCustomK8sAction := &NewCustomK8sAction{
		Action:      d.Get("action").(string),
		Description: description,
		Ruleset:     customK8sActions,
	}

	_, err := client.UpdateCustomK8sAction(id, newCustomK8sAction)
	if err != nil {
		return diag.Errorf("Error updating CustomK8sAction: %s", err)
	}

	log.Printf("[INFO] CustomK8sAction %s successfully updated", id)
	return resourceKomodorCustomK8sActionRead(ctx, d, meta)
}

func resourceKomodorCustomK8sActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting CustomK8sAction: %s", id)
	if err := client.DeleteCustomK8sAction(id); err != nil {
		return diag.Errorf("Error deleting CustomK8sAction: %s", err)
	}

	d.SetId("")
	return nil
}
