package komodor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorPolicyV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKomodorPolicyV2Read,
		Description: "Retrieves an existing Komodor RBAC Policy by name",
		Schema: map[string]*schema.Schema{
			"name": {
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
	}
}

func dataSourceKomodorPolicyV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)

	policy, _, err := client.GetPolicy(name)
	if err != nil {
		return diag.Errorf("Error reading Policy %s: %s", name, err)
	}

	d.SetId(policy.Id)
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", policy.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", policy.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
