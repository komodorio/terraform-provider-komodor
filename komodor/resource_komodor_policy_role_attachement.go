package komodor

import (
	"context"
	"fmt"

	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func statePoliciesForRead(currentPolicies []string, apiPolicies []string) []string {
	if len(currentPolicies) > 0 {
		return currentPolicies
	}
	return apiPolicies
}

func resourcePolicyRoleAttachmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     false,
			Deprecated:   "The name field is deprecated and ignored; the role ID is the canonical resource identity.",
			ValidateFunc: validation.NoZeroValues,
			Description:  "Deprecated: ignored for state management. The role ID is the canonical resource identity.",
		},
		"policies": {
			Type:        schema.TypeSet,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Set:         schema.HashString,
			Description: "List of policy IDs to attach to the role. This resource is non-authoritative and preserves the configured subset when importing or diffing.",
		},
		"role": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
			Description:  "The ID of the role to attach policies to. This is the canonical resource identity for imports and state.",
		},
	}
}

func resourcePolicyRoleAttachment() *schema.Resource {
	legacySchema := resourcePolicyRoleAttachmentSchema()
	return &schema.Resource{
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: (&schema.Resource{Schema: legacySchema}).CoreConfigSchema().ImpliedType(),
				Upgrade: func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
					if rawState == nil {
						return nil, nil
					}

					if role, ok := rawState["role"]; ok && role != nil && role.(string) != "" {
						rawState["id"] = role
						return rawState, nil
					}

					if stateID, ok := rawState["id"]; ok && stateID != nil && stateID.(string) != "" {
						rawState["role"] = stateID
						rawState["id"] = stateID
						return rawState, nil
					}

					return rawState, nil
				},
				Version: 0,
			},
		},
		Schema:        legacySchema,
		CreateContext: resourcePolicyRoleAttachmentCreate,
		ReadContext:   resourcePolicyRoleAttachmentRead,
		UpdateContext: resourcePolicyRoleAttachmentUpdate,
		DeleteContext: resourcePolicyRoleAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Creates a non-authoritative logical binding between a Komodor Role and a selected list of Komodor Policies. The canonical state identity is the role ID.",
	}
}

func resourcePolicyRoleAttachmentRoleID(d *schema.ResourceData) string {
	if role, ok := d.GetOk("role"); ok && role.(string) != "" {
		return role.(string)
	}
	if d.Id() != "" {
		return d.Id()
	}
	return ""
}

func resourcePolicyRoleAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	role := resourcePolicyRoleAttachmentRoleID(d)
	policies := ExpandStringSet(d.Get("policies").(*schema.Set))

	err := client.attachPoliciesToRole(role, policies)

	if err != nil {
		return diag.Errorf("Error attaching policy to role: %s", err)
	}

	d.SetId(role)
	return resourcePolicyRoleAttachmentRead(ctx, d, meta)
}

func resourcePolicyRoleAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	roleId := resourcePolicyRoleAttachmentRoleID(d)
	if roleId == "" {
		return diag.Errorf("role is required for policy-role attachment state")
	}
	if err := d.Set("role", roleId); err != nil {
		return diag.Errorf("error setting role: %s", err)
	}
	if d.Id() == "" {
		d.SetId(roleId)
	}

	rolePolicyObject, statusCode, err := client.GetRolePoliciesObject(roleId)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Role-Policy object (%s) was not found - removing from state", roleId)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Role-Policy object: %s", err)
	}

	apiPolicies := make([]string, 0)
	for _, r := range rolePolicyObject {
		apiPolicies = append(apiPolicies, r.Id)
	}

	configuredPolicies := ExpandStringSet(d.Get("policies").(*schema.Set))
	currentPolicies := make([]string, 0, len(configuredPolicies))
	for _, v := range configuredPolicies {
		if v != nil {
			currentPolicies = append(currentPolicies, *v)
		}
	}

	statePolicies := statePoliciesForRead(currentPolicies, apiPolicies)
	log.Printf("Policies attached to role %s are: %s; state will preserve configured subset when present: %v", roleId, apiPolicies, statePolicies)
	if err := d.Set("policies", statePolicies); err != nil {
		return diag.Errorf("error setting policies: %s", err)
	}

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
	role := resourcePolicyRoleAttachmentRoleID(d)
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
	role := resourcePolicyRoleAttachmentRoleID(d)
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
