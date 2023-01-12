package komodor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKomodorRoleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the role",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Role was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the Role was last updated",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is default role",
			},
		},
	}
}

func dataSourceKomodorRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	role, err := client.GetRoleByName(name)
	if err != nil {
		return diag.Errorf("Could not get role by name %s", name)
	}
	d.SetId(role.Id)
	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", role.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", role.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default", role.IsDefault); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
