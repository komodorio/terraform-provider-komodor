package komodor

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUserRoleBinding() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "A unique name for this user-role binding (for Terraform state management)",
			},
			"user_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The ID or email of the user",
			},
			"roles": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of role IDs or names to assign to the user",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
				Set: schema.HashString,
			},
			"expiration": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Optional expiration date for the user-role assignments (ISO 8601 format)",
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
		CreateContext: resourceUserRoleBindingCreate,
		ReadContext:   resourceUserRoleBindingRead,
		UpdateContext: resourceUserRoleBindingUpdate,
		DeleteContext: resourceUserRoleBindingDelete,
		Description:   "Creates a binding between a Komodor User and one or more Komodor Roles",
	}
}

func resourceUserRoleBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	userId := d.Get("user_id").(string)
	roles := ExpandStringSet(d.Get("roles").(*schema.Set))
	expiration := d.Get("expiration").(string)

	err := client.attachRolesToUser(userId, roles, expiration)
	if err != nil {
		return diag.Errorf("Error attaching roles to user: %s", err)
	}

	d.SetId(name)
	return resourceUserRoleBindingRead(ctx, d, meta)
}

func resourceUserRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	userId := d.Get("user_id").(string)

	userRoles, statusCode, err := client.GetUserRoles(userId)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] User-Role binding (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading User-Role binding: %s", err)
	}

	roleIds := make([]string, 0, len(userRoles))
	for _, userRole := range userRoles {
		roleIds = append(roleIds, userRole.RoleId)
	}

	log.Printf("Roles attached to user %s are: %s", userId, roleIds)
	d.Set("roles", roleIds)

	// Set expiration if present in any of the roles
	if len(userRoles) > 0 && userRoles[0].Expiration != "" {
		d.Set("expiration", userRoles[0].Expiration)
	}

	return nil
}

func resourceUserRoleBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	userId := d.Get("user_id").(string)

	if d.HasChange("roles") {
		o, n := d.GetChange("roles")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := ExpandStringSet(os.Difference(ns))
		add := ExpandStringSet(ns.Difference(os))

		if len(remove) > 0 {
			if err := client.detachRolesFromUser(userId, remove); err != nil {
				return diag.Errorf("Error detaching roles from user: %s", err)
			}
		}

		if len(add) > 0 {
			expiration := d.Get("expiration").(string)
			if err := client.attachRolesToUser(userId, add, expiration); err != nil {
				return diag.Errorf("Error attaching roles to user: %s", err)
			}
		}
	}

	if d.HasChange("expiration") && !d.HasChange("roles") {
		// If only expiration changed, update all existing roles
		roles := ExpandStringSet(d.Get("roles").(*schema.Set))
		expiration := d.Get("expiration").(string)
		if err := client.updateUserRoles(userId, roles, expiration); err != nil {
			return diag.Errorf("Error updating user role expirations: %s", err)
		}
	}

	return resourceUserRoleBindingRead(ctx, d, meta)
}

func resourceUserRoleBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	userId := d.Get("user_id").(string)
	roles := ExpandStringSet(d.Get("roles").(*schema.Set))

	err := client.detachRolesFromUser(userId, roles)
	if err != nil {
		return diag.Errorf("Error detaching roles from user: %s", err)
	}

	d.SetId("")
	return nil
}

func (c *Client) attachRolesToUser(userId string, roles []*string, expiration string) error {
	for _, roleId := range roles {
		err := c.AttachUserToRole(userId, *roleId, expiration)
		if err != nil {
			return fmt.Errorf("error attaching role %s to user %s: %w", *roleId, userId, err)
		}
	}
	return nil
}

func (c *Client) detachRolesFromUser(userId string, roles []*string) error {
	for _, roleId := range roles {
		err := c.DetachUserFromRole(userId, *roleId)
		if err != nil {
			return fmt.Errorf("error detaching role %s from user %s: %w", *roleId, userId, err)
		}
	}
	return nil
}

func (c *Client) updateUserRoles(userId string, roles []*string, expiration string) error {
	for _, roleId := range roles {
		err := c.UpdateUserRole(userId, *roleId, expiration)
		if err != nil {
			return fmt.Errorf("error updating role %s for user %s: %w", *roleId, userId, err)
		}
	}
	return nil
}
