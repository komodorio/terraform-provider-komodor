package komodor

import (
	"context"
	"fmt"

	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePolicyRoleAttachment() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"policies": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
		CreateContext: resourcePolicyRoleAttachmentCreate,
		ReadContext:   resourcePolicyRoleAttachmentRead,
		UpdateContext: resourcePolicyRoleAttachmentUpdate,
		DeleteContext: resourcePolicyRoleAttachmentDelete,
	}
}

func resourcePolicyRoleAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	role := d.Get("role").(string)
	policies := ExpandStringSet(d.Get("policies").(*schema.Set))

	err := client.attachPoliciesToRole(role, policies)

	if err != nil {
		return diag.Errorf("Error attaching policy to role: %s", err)
	}

	d.SetId(name)
	return resourcePolicyRoleAttachmentRead(ctx, d, meta)
}

func resourcePolicyRoleAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	roleId := d.Get("role").(string)

	rolePolicyObject, statusCode, err := client.GetRolePoliciesObject(roleId)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Role-Policy object (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Role-Policy object: %s", err)
	}

	pl := make([]string, 0)

	for _, r := range rolePolicyObject {
		pl = append(pl, r.PolicyId)
	}

	log.Printf("Policies attached to role %s are: %s", roleId, pl)
	d.Set("policies", pl)

	return nil
}

func resourcePolicyRoleAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	var policiesErr error
	if d.HasChange("policies") {
		policiesErr = client.updatePolicies(d)
	}
	if policiesErr != nil {
		return diag.Errorf("Error updating policies: %s", policiesErr)
	}

	return resourcePolicyRoleAttachmentRead(ctx, d, meta)
}

func resourcePolicyRoleAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	role := d.Get("role").(string)
	policies := ExpandStringSet(d.Get("policies").(*schema.Set))

	err := client.detachPoliciesFromRole(role, policies)

	if err != nil {
		return diag.Errorf("Error detaching policies from role: %s", err)
	}

	d.SetId("")
	return nil
}

func (c *Client) attachPoliciesToRole(role string, policies []*string) error {
	for _, p := range policies {
		err := c.AttachPolicy(*p, role)
		if err != nil {
			return fmt.Errorf("error attaching policy %s to role %s", *p, role)
		}
	}
	return nil
}

func (c *Client) detachPoliciesFromRole(role string, policies []*string) error {
	for _, p := range policies {
		err := c.DetachPolicy(*p, role)
		if err != nil {
			return fmt.Errorf("error detaching policy %s from role %s", *p, role)
		}
	}
	return nil
}

func (c *Client) updatePolicies(d *schema.ResourceData) error {
	role := d.Get("role").(string)
	o, n := d.GetChange("policies")
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

	if rErr := c.detachPoliciesFromRole(role, remove); rErr != nil {
		return rErr
	}
	if aErr := c.attachPoliciesToRole(role, add); aErr != nil {
		return aErr
	}
	return nil
}
