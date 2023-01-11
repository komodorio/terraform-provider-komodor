package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorRole() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
		CreateContext: resourceKomodorRoleCreate,
		ReadContext:   resourceKomodorRoleRead,
		UpdateContext: resourceKomodorRoleUpdate,
		DeleteContext: resourceKomodorRoleDelete,
	}
}

func resourceKomodorRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	newRole := &NewRole{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] Role create configuration: %#v", newRole)
	role, err := client.CreateRole(newRole)
	if err != nil {
		return diag.Errorf("Error creating Role: %s", err)
	}

	d.SetId(role.Id)
	log.Printf("[INFO] Role created successfully. Role Id: %s", role.Id)

	return resourceKomodorRoleRead(ctx, d, meta)
}

func resourceKomodorRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	role, err := client.GetRole(id)
	if err != nil {
		statusCode := GetStatusCodeFromErrorMessage(err)
		if statusCode == "404" {
			log.Printf("[DEBUG] Role (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Role: %s", err)
	}

	d.Set("name", role.Name)
	d.Set("created_at", role.CreatedAt)
	d.Set("updated_at", role.UpdatedAt)
	d.Set("is_default", role.IsDefault)

	return nil
}

func resourceKomodorRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()
	if d.HasChange("name") {
		// client.updateRole() // not yet implemented in api, so deleting and recreating
		if err := client.DeleteRole(id); err != nil {
			return diag.Errorf("Error deleting Role: %s", err)
		}

		d.SetId("")
		return resourceKomodorRoleCreate(ctx, d, meta)
	}
	return nil
}

func resourceKomodorRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Role: %s", id)
	if err := client.DeleteRole(id); err != nil {
		return diag.Errorf("Error deleting Role: %s", err)
	}

	d.SetId("")
	return nil
}
