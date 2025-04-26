package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKomodorWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Komodor workspace",
		ReadContext: dataSourceKomodorWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the workspace",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the workspace",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the workspace",
			},
			"scopes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The scopes of the workspace",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clusters": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"namespaces": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"clusters_patterns":   patternListSchema(),
						"namespaces_patterns": patternListSchema(),
						"selectors": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     selectorSchema(),
						},
						"selectors_patterns": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     selectorPatternSchema(),
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation timestamp of the workspace",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update timestamp of the workspace",
			},
			"author_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email of the workspace author",
			},
			"last_updated_by_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email of the last user who updated the workspace",
			},
		},
	}
}

func dataSourceKomodorWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Get("id").(string)

	workspace, statusCode, err := client.GetWorkspace(id)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Workspace (%s) was not found", id)
			return diag.Errorf("Workspace not found: %s", id)
		}
		return diag.Errorf("Error reading Workspace: %s", err)
	}

	if err := flattenWorkspace(workspace, d); err != nil {
		return diag.Errorf("Error flattening workspace: %s", err)
	}

	return nil
}
