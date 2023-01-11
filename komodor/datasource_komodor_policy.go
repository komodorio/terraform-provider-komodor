package komodor

import (
	"context"
	"encoding/json"

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
	}
}

func dataSourceKomodorPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	policy, err := client.GetPolicyByName(name)
	if err != nil {
		return diag.FromErr(err)
	}

	jsonStatements, err := json.Marshal(policy.Statements)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.Id)
	d.Set("created_at", policy.CreatedAt) // err not handled intentionally?
	d.Set("updated_at", policy.UpdatedAt)
	d.Set("statements", string(jsonStatements))

	return nil
}
