package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scopes": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clusters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"namespaces": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"clusters_patterns":   patternListSchema(),
						"namespaces_patterns": patternListSchema(),
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"author_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_by_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Expand (from TF -> GO)

func expandWorkspace(d *schema.ResourceData) *NewWorkspace {
	return &NewWorkspace{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Scopes:      expandResourcesScope(d.Get("scopes").([]interface{})),
	}
}

// Flatten (from GO -> TF)

func flattenWorkspace(workspace *Workspace, d *schema.ResourceData) error {
	d.Set("name", workspace.Name)
	d.Set("description", workspace.Description)
	d.Set("scopes", flattenResourcesScope(workspace.Scopes))
	d.Set("created_at", workspace.CreatedAt)
	d.Set("updated_at", workspace.LastUpdated)
	d.Set("author_email", workspace.AuthorEmail)
	d.Set("last_updated_by_email", workspace.LastUpdatedByEmail)
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
