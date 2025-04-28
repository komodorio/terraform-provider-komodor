package komodor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKomodorUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the user",
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the User was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the User was last updated",
			},
		},
		Description: "Retrieves an existing Komodor User by email",
	}
}

func dataSourceKomodorUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	email := d.Get("email").(string)
	user, _, err := client.GetUser(email)
	if err != nil {
		return diag.Errorf("Could not get user by email %s", email)
	}
	d.SetId(user.Id)
	if err := d.Set("display_name", user.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", user.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", user.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
