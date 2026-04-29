package komodor

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorKlaudiaSkill() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Klaudia skill — domain expertise that teaches Klaudia how to use MCP integration tools.",
		CreateContext: resourceKlaudiaSkillCreate,
		ReadContext:   resourceKlaudiaSkillRead,
		UpdateContext: resourceKlaudiaSkillUpdate,
		DeleteContext: resourceKlaudiaSkillDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Unique name for the skill (per account).",
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Short description shown in the UI.",
				ValidateFunc: validation.StringLenBetween(1, 2000),
			},
			"instructions": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Full instructions injected into Klaudia's agent prompt when this skill is active. Supports markdown. Max 50,000 characters.",
				ValidateFunc: validation.StringLenBetween(1, 50000),
			},
			"clusters": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Clusters this skill is active for. Use `[\"*\"]` for all clusters.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the skill is active.",
			},
		},
	}
}

func resourceKlaudiaSkillCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	req := &CreateSkillRequest{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Instructions: d.Get("instructions").(string),
		UseCases:     []string{"rca", "chat"},
		Clusters:     expandStringList(d.Get("clusters").([]interface{})),
	}

	skill, err := c.CreateSkill(req)
	if err != nil {
		return diag.Errorf("error creating Klaudia skill: %s", err)
	}
	d.SetId(skill.ID)
	return resourceKlaudiaSkillRead(ctx, d, meta)
}

func resourceKlaudiaSkillRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	skill, statusCode, err := c.GetSkill(d.Id())
	if err != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading Klaudia skill %s: %s", d.Id(), err)
	}

	_ = d.Set("name", skill.Name)
	_ = d.Set("description", skill.Description)
	_ = d.Set("instructions", skill.Instructions)
	_ = d.Set("clusters", skill.Clusters)
	_ = d.Set("is_enabled", skill.IsEnabled)
	return nil
}

func resourceKlaudiaSkillUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	req := &UpdateSkillRequest{}
	if d.HasChange("name") {
		v := d.Get("name").(string)
		req.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		req.Description = &v
	}
	if d.HasChange("instructions") {
		v := d.Get("instructions").(string)
		req.Instructions = &v
	}
	if d.HasChange("clusters") {
		req.Clusters = expandStringList(d.Get("clusters").([]interface{}))
	}
	if d.HasChange("is_enabled") {
		v := d.Get("is_enabled").(bool)
		req.IsEnabled = &v
	}

	if _, err := c.UpdateSkill(d.Id(), req); err != nil {
		return diag.Errorf("error updating Klaudia skill %s: %s", d.Id(), err)
	}
	return resourceKlaudiaSkillRead(ctx, d, meta)
}

func resourceKlaudiaSkillDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	if err := c.DeleteSkill(d.Id()); err != nil {
		return diag.Errorf("error deleting Klaudia skill %s: %s", d.Id(), err)
	}
	return nil
}
