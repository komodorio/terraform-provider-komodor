package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"
)

func resourceKomodorWorkspace() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a Komodor Workspace",
		CreateContext: resourceKomodorWorkspaceCreate,
		ReadContext:   resourceKomodorWorkspaceRead,
		UpdateContext: resourceKomodorWorkspaceUpdate,
		DeleteContext: resourceKomodorWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The name of the workspace.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A human-readable description of the workspace.",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "One or more scopes defining the Kubernetes resources visible in this workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clusters": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of cluster names to include in the scope.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"namespaces": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of namespace names to include in the scope.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"clusters_patterns":   patternListMaxOneSchema(),
						"namespaces_patterns": patternListMaxOneSchema(),
						"selectors": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     selectorSchema(),
						},
						"selectors_patterns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     selectorPatternSchema(),
						},
					},
				},
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the workspace.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the workspace was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the workspace was last updated.",
			},
			"author_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email of the user who created the workspace.",
			},
			"last_updated_by_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email of the user who last updated the workspace.",
			},
		},
	}
}

// Expand (from TF -> GO)

func expandWorkspace(d *schema.ResourceData) *NewWorkspace {
	scopes := d.Get("scopes").([]interface{})
	expandedScopes := lo.Map(scopes, func(item interface{}, _ int) ResourcesScope {
		data := item.(map[string]interface{})
		scope := expandResourcesScope([]interface{}{data})
		if scope == nil {
			return ResourcesScope{}
		}
		return *scope
	})

	return &NewWorkspace{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Scopes:      expandedScopes,
	}
}

// Flatten (from GO -> TF)

func flattenWorkspace(workspace *Workspace, d *schema.ResourceData) error {
	if err := d.Set("name", workspace.Name); err != nil {
		return err
	}
	if err := d.Set("description", workspace.Description); err != nil {
		return err
	}
	scopesList := lo.Map(workspace.Scopes, func(scope ResourcesScope, _ int) interface{} {
		return flattenResourcesScope(&scope)
	})
	if err := d.Set("scopes", scopesList); err != nil {
		return err
	}
	if err := d.Set("created_at", workspace.CreatedAt); err != nil {
		return err
	}
	if err := d.Set("updated_at", workspace.LastUpdated); err != nil {
		return err
	}
	if err := d.Set("author_email", workspace.AuthorEmail); err != nil {
		return err
	}
	if err := d.Set("last_updated_by_email", workspace.LastUpdatedByEmail); err != nil {
		return err
	}
	return nil
}

func resourceKomodorWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	newWorkspace := expandWorkspace(d)

	workspace, err := client.CreateWorkspace(newWorkspace)
	if err != nil {
		return diag.Errorf("Error creating workspace: %s", err)
	}

	d.SetId(workspace.Id)

	log.Printf("[INFO] Workspace created successfully. Workspace Id: %s", workspace.Id)

	return resourceKomodorWorkspaceRead(ctx, d, meta)
}

func resourceKomodorWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspace, statusCode, err := client.GetWorkspace(d.Id())
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Workspace (%s) was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Workspace: %s", err)
	}

	if err := flattenWorkspace(workspace, d); err != nil {
		return diag.Errorf("Error flattening workspace: %s", err)
	}

	return nil
}

func resourceKomodorWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	newWorkspace := expandWorkspace(d)

	_, err := client.UpdateWorkspace(d.Id(), newWorkspace)
	if err != nil {
		return diag.Errorf("Error updating workspace: %s", err)
	}

	log.Printf("[INFO] Workspace %s successfully updated", d.Id())
	return resourceKomodorWorkspaceRead(ctx, d, meta)
}

func resourceKomodorWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()

	log.Printf("[INFO] Deleting Workspace: %s", id)
	if err := client.DeleteWorkspace(id); err != nil {
		return diag.Errorf("Error deleting workspace: %s", err)
	}

	d.SetId("")
	return nil
}
