package komodor

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorKnowledgeBaseFile() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"knowledge-base", "blueprint"}, false),
				Description:  "The type of Klaudia file. Valid values are `knowledge-base` and `blueprint`.",
			},
			"filename": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The name of the file as stored in the Knowledge Base.",
			},
			"content": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "The text content of the file to upload to the Knowledge Base.",
			},
			"clusters": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Optional cluster scoping configuration for this file.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of cluster names that this file applies to.",
						},
						"exclude": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of cluster names to exclude from this file's scope.",
						},
					},
				},
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "File size in bytes.",
			},
			"uploaded_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the file was uploaded.",
			},
			"created_by_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user who uploaded the file.",
			},
		},
		CreateContext: resourceKomodorKnowledgeBaseFileCreate,
		ReadContext:   resourceKomodorKnowledgeBaseFileRead,
		// There is no update endpoint; all mutable attributes use ForceNew.
		DeleteContext: resourceKomodorKnowledgeBaseFileDelete,
		Description: "Manages a file in the Komodor Klaudia Knowledge Base.\n\n" +
			"Knowledge Base files provide contextual runbook-style documentation that Klaudia AI uses " +
			"when performing root cause analysis. Files can be optionally scoped to specific clusters.\n\n" +
			"Note: Because the API does not support in-place updates, any change to `file_type`, `filename`, `content`, " +
			"or `clusters` will cause the resource to be destroyed and re-created.",
	}
}

// expandKnowledgeBaseClusters converts the Terraform schema clusters block into a
// KnowledgeBaseScopedClusters struct. Returns nil if no clusters are configured.
func expandKnowledgeBaseClusters(d *schema.ResourceData) *KnowledgeBaseScopedClusters {
	raw, ok := d.GetOk("clusters")
	if !ok {
		return nil
	}
	list := raw.([]interface{})
	if len(list) == 0 {
		return nil
	}
	m := list[0].(map[string]interface{})
	clusters := &KnowledgeBaseScopedClusters{}
	if includeSet, ok := m["include"].(*schema.Set); ok {
		for _, v := range includeSet.List() {
			clusters.Include = append(clusters.Include, v.(string))
		}
	}
	if excludeSet, ok := m["exclude"].(*schema.Set); ok {
		for _, v := range excludeSet.List() {
			clusters.Exclude = append(clusters.Exclude, v.(string))
		}
	}
	return clusters
}

func resourceKomodorKnowledgeBaseFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	fileType := d.Get("file_type").(string)
	filename := d.Get("filename").(string)
	content := []byte(d.Get("content").(string))
	clusters := expandKnowledgeBaseClusters(d)

	log.Printf("[DEBUG] KnowledgeBaseFile create: filename=%s type=%s", filename, fileType)

	file, err := client.UploadKnowledgeBaseFile(filename, content, clusters, fileType)
	if err != nil {
		return diag.Errorf("Error uploading Knowledge Base file: %s", err)
	}

	d.SetId(file.Id)
	log.Printf("[INFO] Knowledge Base file uploaded successfully. Id: %s", file.Id)

	return resourceKomodorKnowledgeBaseFileRead(ctx, d, meta)
}

func resourceKomodorKnowledgeBaseFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()
	fileType := d.Get("file_type").(string)

	file, statusCode, err := client.GetKnowledgeBaseFile(id, fileType)
	if err != nil {
		if statusCode == 404 {
			log.Printf("[DEBUG] Knowledge Base file (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error reading Knowledge Base file: %s", err)
	}

	var diags diag.Diagnostics

	if err := d.Set("filename", file.Name); err != nil {
		diags = append(diags, diag.Errorf("Error setting filename: %s", err)...)
	}
	if err := d.Set("size", int(file.Size)); err != nil {
		diags = append(diags, diag.Errorf("Error setting size: %s", err)...)
	}
	if err := d.Set("uploaded_at", file.UploadedAt); err != nil {
		diags = append(diags, diag.Errorf("Error setting uploaded_at: %s", err)...)
	}
	if err := d.Set("created_by_email", file.CreatedByEmail); err != nil {
		diags = append(diags, diag.Errorf("Error setting created_by_email: %s", err)...)
	}
	// The API does not return cluster scoping in the read response, so we only
	// update state when the API explicitly provides it. This preserves the
	// configured value and avoids a perpetual diff for a ForceNew attribute.
	if file.Clusters != nil {
		clustersData := []interface{}{
			map[string]interface{}{
				"include": file.Clusters.Include,
				"exclude": file.Clusters.Exclude,
			},
		}
		if err := d.Set("clusters", clustersData); err != nil {
			diags = append(diags, diag.Errorf("Error setting clusters: %s", err)...)
		}
	}

	return diags
}

func resourceKomodorKnowledgeBaseFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id := d.Id()
	fileType := d.Get("file_type").(string)

	log.Printf("[DEBUG] Deleting Knowledge Base file: %s", id)

	resp, err := client.DeleteKnowledgeBaseFiles([]string{id}, fileType)
	if err != nil {
		return diag.Errorf("Error deleting Knowledge Base file: %s", err)
	}

	for _, failedID := range resp.FailedFiles {
		if failedID == id {
			return diag.Errorf("API reported failure deleting Knowledge Base file %s", id)
		}
	}

	log.Printf("[INFO] Knowledge Base file %s successfully deleted", id)
	return nil
}
