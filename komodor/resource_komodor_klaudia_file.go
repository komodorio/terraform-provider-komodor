package komodor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKomodorKlaudiaFile() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Klaudia knowledge-base or blueprint file.",
		CreateContext: resourceKlaudiaFileCreate,
		ReadContext:   resourceKlaudiaFileRead,
		UpdateContext: resourceKlaudiaFileUpdate,
		DeleteContext: resourceKlaudiaFileDelete,
		CustomizeDiff: resourceKlaudiaFileCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Klaudia file type. Must be `knowledge-base` or `blueprint`.",
				ValidateFunc: validation.StringInSlice([]string{"knowledge-base", "blueprint"}, false),
			},
			"filename": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Filename to use in Komodor. Together with `type`, this is used to find an existing remote file.",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"source_path": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Local path to read file content from. The path may be stored in state, but the file content is not.",
				ConflictsWith: []string{"content"},
			},
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				WriteOnly:     true,
				Sensitive:     true,
				Description:   "Write-only file content supplied directly in configuration. This value is not persisted in plan or state. Requires Terraform 1.11 or newer.",
				ConflictsWith: []string{"source_path"},
			},
			"checksum": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "SHA-256 checksum used by Terraform to detect file content changes. Computed from `source_path` or `content` when available.",
			},
			"clusters": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Optional cluster include/exclude scope for this file.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Clusters to include.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"exclude": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Clusters to exclude.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Remote file size in bytes.",
			},
			"uploaded_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when Komodor last uploaded or updated the file.",
			},
			"created_by_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email of the user that created the file.",
			},
		},
	}
}

func resourceKlaudiaFileCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if pathRaw, ok := d.GetOk("source_path"); ok {
		content, err := os.ReadFile(pathRaw.(string))
		if err != nil {
			return fmt.Errorf("error reading source_path %q: %w", pathRaw.(string), err)
		}
		return d.SetNew("checksum", sha256Hex(content))
	}

	if contentRaw, ok := d.GetOk("content"); ok {
		return d.SetNew("checksum", sha256Hex([]byte(contentRaw.(string))))
	}

	return fmt.Errorf("one of `source_path` or `content` must be configured")
}

func resourceKlaudiaFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	fileType := d.Get("type").(string)
	filename := d.Get("filename").(string)
	payload, checksum, err := buildKlaudiaFilePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	existing, _, err := c.ListKlaudiaFiles(fileType)
	if err != nil {
		return diag.Errorf("error listing Klaudia %s files: %s", fileType, err)
	}
	for _, file := range existing.Files {
		if file.Name == filename {
			d.SetId(file.ID)
			if _, _, err := c.UpdateKlaudiaFile(fileType, file.ID, &payload, expandKlaudiaFileClusters(d)); err != nil {
				return diag.Errorf("error updating existing Klaudia file %s: %s", file.ID, err)
			}
			_ = d.Set("checksum", checksum)
			return resourceKlaudiaFileRead(ctx, d, meta)
		}
	}

	uploaded, err := c.UploadKlaudiaFile(fileType, payload, expandKlaudiaFileClusters(d))
	if err != nil {
		return diag.Errorf("error uploading Klaudia %s file %q: %s", fileType, filename, err)
	}
	for _, file := range uploaded.Files {
		if file.Name == filename {
			d.SetId(file.ID)
			_ = d.Set("checksum", checksum)
			return resourceKlaudiaFileRead(ctx, d, meta)
		}
	}

	return diag.Errorf("Klaudia file %q was uploaded but was not present in API response", filename)
}

func resourceKlaudiaFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	fileType := d.Get("type").(string)
	files, statusCode, err := c.ListKlaudiaFiles(fileType)
	if err != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error listing Klaudia %s files: %s", fileType, err)
	}

	for _, file := range files.Files {
		if file.ID == d.Id() {
			return flattenKlaudiaFile(d, &file)
		}
	}

	d.SetId("")
	return nil
}

func resourceKlaudiaFileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	var payload *klaudiaFilePayload
	var checksum string
	if d.HasChange("source_path") || d.HasChange("content") || d.HasChange("checksum") {
		nextPayload, nextChecksum, err := buildKlaudiaFilePayload(d)
		if err != nil {
			return diag.FromErr(err)
		}
		payload = &nextPayload
		checksum = nextChecksum
	}

	if _, statusCode, err := c.UpdateKlaudiaFile(d.Get("type").(string), d.Id(), payload, expandKlaudiaFileClusters(d)); err != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return resourceKlaudiaFileCreate(ctx, d, meta)
		}
		return diag.Errorf("error updating Klaudia file %s: %s", d.Id(), err)
	}

	if checksum != "" {
		_ = d.Set("checksum", checksum)
	}
	return resourceKlaudiaFileRead(ctx, d, meta)
}

func resourceKlaudiaFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	deleted, err := c.DeleteKlaudiaFile(d.Get("type").(string), d.Id())
	if err != nil {
		return diag.Errorf("error deleting Klaudia file %s: %s", d.Id(), err)
	}
	for _, failedID := range deleted.FailedFiles {
		if failedID == d.Id() {
			return diag.Errorf("Klaudia file %s failed to delete", d.Id())
		}
	}
	return nil
}

func buildKlaudiaFilePayload(d *schema.ResourceData) (klaudiaFilePayload, string, error) {
	filename := d.Get("filename").(string)

	if pathRaw, ok := d.GetOk("source_path"); ok {
		content, err := os.ReadFile(pathRaw.(string))
		if err != nil {
			return klaudiaFilePayload{}, "", fmt.Errorf("error reading source_path %q: %w", pathRaw.(string), err)
		}
		return klaudiaFilePayload{Filename: filename, Content: content}, sha256Hex(content), nil
	}

	if contentRaw, ok := d.GetOk("content"); ok {
		content := []byte(contentRaw.(string))
		return klaudiaFilePayload{Filename: filename, Content: content}, sha256Hex(content), nil
	}

	return klaudiaFilePayload{}, "", fmt.Errorf("one of `source_path` or `content` must be configured")
}

func flattenKlaudiaFile(d *schema.ResourceData, file *KlaudiaFile) diag.Diagnostics {
	_ = d.Set("filename", file.Name)
	_ = d.Set("size", int(file.Size))
	_ = d.Set("uploaded_at", file.UploadedAt)
	_ = d.Set("created_by_email", file.CreatedByEmail)
	if file.Clusters != nil {
		_ = d.Set("clusters", []interface{}{map[string]interface{}{
			"include": file.Clusters.Include,
			"exclude": file.Clusters.Exclude,
		}})
	} else {
		_ = d.Set("clusters", []interface{}{})
	}
	return nil
}

func expandKlaudiaFileClusters(d *schema.ResourceData) *KlaudiaFileClusters {
	raw := d.Get("clusters").([]interface{})
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}

	data := raw[0].(map[string]interface{})
	clusters := &KlaudiaFileClusters{}
	if includeRaw, ok := data["include"].([]interface{}); ok {
		clusters.Include = expandStringList(includeRaw)
	}
	if excludeRaw, ok := data["exclude"].([]interface{}); ok {
		clusters.Exclude = expandStringList(excludeRaw)
	}
	return clusters
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
