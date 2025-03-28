package komodor

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKomodorPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the policy",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Policy was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Policy was last updated",
			},
			"statements": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy's statements",
			},
		},
		Description: "Retrieves an existing Komodor Policy by name",
	}
}

func dataSourceKomodorPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	policy, statusCode, err := client.GetPolicy(name)
	if err != nil {
		return diag.FromErr(err)
	}

	if statusCode == http.StatusNotFound {
		return diag.Errorf("Policy not found")
	}

	jsonStatements, err := json.Marshal(policy.Statements)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Id)
	if err := d.Set("created_at", policy.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", policy.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("statements", string(jsonStatements)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
