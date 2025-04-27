package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				ForceNew:     true,
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
		CreateContext: resourceKomodorUserCreate,
		ReadContext:   resourceKomodorUserRead,
		UpdateContext: resourceKomodorUserUpdate,
		DeleteContext: resourceKomodorUserDelete,
		Description:   "Creates a Komodor User",
	}
}

func resourceKomodorUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	newUser := &NewUser{
		DisplayName:      d.Get("display_name").(string),
		Email:            d.Get("email").(string),
		RestoreIfDeleted: true,
	}

	log.Printf("[DEBUG] User create configuration: %#v", newUser)
	user, err := client.CreateUser(newUser)
	if err != nil {
		return diag.Errorf("Error creating User: %s", err)
	}

	d.SetId(user.Id)
	log.Printf("[INFO] User created successfully. User Id: %s", user.Id)

	return resourceKomodorUserRead(ctx, d, meta)
}

func resourceKomodorUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	user, statusCode, err := client.GetUser(id)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] User (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading User: %s", err)
	}

	d.Set("display_name", user.DisplayName)
	d.Set("email", user.Email)
	d.Set("created_at", user.CreatedAt)
	d.Set("updated_at", user.UpdatedAt)

	return nil
}

func resourceKomodorUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	updateUser := &UpdateUser{
		DisplayName: d.Get("display_name").(string),
	}

	_, err := client.UpdateUser(id, updateUser)

	if err != nil {
		return diag.Errorf("Error updating user: %s", err)
	}

	log.Printf("[INFO] User %s successfully updated", id)
	return resourceKomodorUserRead(ctx, d, meta)
}

func resourceKomodorUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting User: %s", id)
	if err := client.DeleteUser(id); err != nil {
		return diag.Errorf("Error deleting User: %s", err)
	}

	d.SetId("")
	return nil
}
