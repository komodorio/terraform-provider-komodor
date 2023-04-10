package komodor

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"statements": {
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
		CreateContext: resourceKomodorPolicyCreate,
		ReadContext:   resourceKomodorPolicyRead,
		UpdateContext: resourceKomodorPolicyUpdate,
		DeleteContext: resourceKomodorPolicyDelete,
	}
}

func resourceKomodorPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)

	statementsJson := d.Get("statements")

	var statements []Statement
	if err := json.Unmarshal([]byte(statementsJson.(string)), &statements); err != nil {
		return diag.Errorf("Error creating statement structure: %s", err)
	}

	newPolicy := &NewPolicy{
		Name:       name,
		Statements: statements,
	}

	policy, err := client.CreatePolicy(newPolicy)
	if err != nil {
		return diag.Errorf("Error creating policy: %s", err)
	}

	d.SetId(policy.Id)
	log.Printf("[INFO] Policy created successfully. Policy Id: %s", policy.Id)

	return resourceKomodorPolicyRead(ctx, d, meta)
}

func resourceKomodorPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	policy, err := client.GetPolicy(id)
	if err != nil {
		if err.Error() == "404" {
			log.Printf("[DEBUG] Policy (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Policy: %s", err)
	}

	d.Set("name", policy.Name)
	d.Set("statements", policy.Statements)
	d.Set("created_at", policy.CreatedAt)
	d.Set("updated_at", policy.UpdatedAt)

	return nil
}

func resourceKomodorPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	statementsJson := d.Get("statements")

	var statements []Statement
	if err := json.Unmarshal([]byte(statementsJson.(string)), &statements); err != nil {
		return diag.Errorf("Error creating statement structure: %s", err)
	}

	newPolicy := &NewPolicy{
		Name:       d.Get("name").(string),
		Statements: statements,
	}

	_, err := client.UpdatePolicy(id, newPolicy)

	if err != nil {
		return diag.Errorf("Error updating policy: %s", err)
	}

	log.Printf("[INFO] Policy %s successfully updated", id)
	return resourceKomodorPolicyRead(ctx, d, meta)
}

func resourceKomodorPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Policy: %s", id)
	if err := client.DeletePolicy(id); err != nil {
		return diag.Errorf("Error deleting policy: %s", err)
	}

	d.SetId("")
	return nil
}
