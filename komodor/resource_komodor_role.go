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
			"is_default": {
				Type:        schema.TypeBool,
				Description: "Set this role as the account wide Default role",
				Optional:    true,
				Default:     false,
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
		},
		CreateContext: resourceKomodorRoleCreate,
		ReadContext:   resourceKomodorRoleRead,
		UpdateContext: resourceKomodorRoleUpdate,
		DeleteContext: resourceKomodorRoleDelete,
		Description: "Creates a Komodor RBAC Role that when combined with a Policy,\n\n" +
			"defines a set of actions one can perform on resources through the Komodor Platform",
	}
}

func resourceKomodorRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	newRole := &NewRole{
		Name:      d.Get("name").(string),
		IsDefault: d.Get("is_default").(bool),
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

	role, statusCode, err := client.GetRole(id)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Role (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Role: %s", err)
	}

	if err := d.Set("name", role.Name); err != nil {
		return diag.Errorf("error setting name: %s", err)
	}
	if err := d.Set("created_at", role.CreatedAt); err != nil {
		return diag.Errorf("error setting created_at: %s", err)
	}
	if err := d.Set("updated_at", role.UpdatedAt); err != nil {
		return diag.Errorf("error setting updated_at: %s", err)
	}
	if err := d.Set("is_default", role.IsDefault); err != nil {
		return diag.Errorf("error setting is_default: %s", err)
	}

	return nil
}

func resourceKomodorRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	newRole := &NewRole{
		Name:      d.Get("name").(string),
		IsDefault: d.Get("is_default").(bool),
	}

	id := d.Id()
	if d.HasChange("name") || d.HasChange("is_default") {
		_, err := client.UpdateRole(id, newRole)
		if err != nil {
			return diag.Errorf("Error updating role: %s", err)
		}
	}
	return resourceKomodorRoleRead(ctx, d, meta)
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
