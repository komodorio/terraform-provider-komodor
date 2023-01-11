package komodor

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceKomodorRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceKomodorRoleRead,

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

func dataSourceKomodorRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	role, err := client.GetRoleByName(name)
	if err != nil {
		return err
	}
	d.SetId(role.Id)
	d.Set("name", role.Name)
	d.Set("created_at", role.CreatedAt)
	d.Set("updated_at", role.UpdatedAt)
	d.Set("is_default", role.IsDefault)

	return nil
}
