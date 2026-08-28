package komodor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-cty/cty"
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
		Importer: &schema.ResourceImporter{
			StateContext: resourceKlaudiaFileImport,
		},
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
	raw := d.GetRawConfig()
	if !raw.IsNull() && raw.IsKnown() {
		if pathRaw, ok := extractKlaudiaFileConfigString(raw, "source_path"); ok {
			content, err := os.ReadFile(pathRaw)
			if err != nil {
				return fmt.Errorf("error reading source_path %q: %w", pathRaw, err)
			}
			return d.SetNew("checksum", sha256Hex(content))
		}
		if contentRaw, ok := extractKlaudiaFileConfigString(raw, "content"); ok {
			return d.SetNew("checksum", sha256Hex([]byte(contentRaw)))
		}
	}

	return fmt.Errorf("one of `source_path` or `content` must be configured")
}

func resourceKlaudiaFileImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	fileType, fileID, err := parseKlaudiaFileImportID(d.Id())
	if err != nil {
		return nil, err
	}
	if err := d.Set("type", fileType); err != nil {
		return nil, err
	}
	d.SetId(fileID)
	return []*schema.ResourceData{d}, nil
}

func parseKlaudiaFileImportID(value string) (string, string, error) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected import ID in the form <type>:<file-id>, got %q", value)
	}
	return parts[0], parts[1], nil
}

func extractKlaudiaFileConfigString(raw cty.Value, attr string) (string, bool) {
	if raw.IsNull() || !raw.IsKnown() {
		return "", false
	}
	if !raw.Type().IsObjectType() {
		return "", false
	}
	value := raw.GetAttr(attr)
	if !value.IsKnown() || value.IsNull() || value.Type() != cty.String {
		return "", false
	}
	return value.AsString(), true
}

func resourceKlaudiaFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	fileType := d.Get("type").(string)
	filename := d.Get("filename").(string)
	payload, checksum, err := buildKlaudiaFilePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	existing, statusCode, err := c.ListKlaudiaFiles(fileType)
	if err != nil && statusCode != http.StatusNotFound {
		return diag.Errorf("error listing Klaudia %s files: %s", fileType, err)
	}
	if existing == nil {
		existing = &KlaudiaFileListResponse{}
	}
	for _, file := range existing.Files {
		if file.Name == filename {
			return diag.Errorf("Klaudia %s file %q already exists (id %s); import it instead: terraform import <address> %s:%s", fileType, filename, file.ID, fileType, file.ID)
		}
	}

	uploaded, statusCode, err := c.UploadKlaudiaFile(fileType, payload, expandKlaudiaFileClusters(d))
	if err != nil {
		if statusCode == http.StatusNotFound {
			return diag.Errorf("error uploading Klaudia %s file %q: %s (the %s endpoint was not found)", fileType, filename, err, fileType)
		}
		return diag.Errorf("error uploading Klaudia %s file %q: %s", fileType, filename, err)
	}
	for _, file := range uploaded.Files {
		if file.Name == filename {
			// The file now exists remotely, so record the ID before any follow-up
			// call that can fail — otherwise a failure here orphans the file.
			d.SetId(file.ID)
			_ = d.Set("checksum", checksum)
			if _, _, err := c.UpdateKlaudiaFile(fileType, file.ID, &payload, expandKlaudiaFileClusters(d)); err != nil {
				return diag.Errorf("error updating uploaded Klaudia file %s: %s", file.ID, err)
			}
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

	deleted, statusCode, err := c.DeleteKlaudiaFile(d.Get("type").(string), d.Id())
	if err != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error deleting Klaudia file %s: %s", d.Id(), err)
	}
	if deleted != nil {
		for _, failedID := range deleted.FailedFiles {
			if failedID != d.Id() {
				continue
			}
			// The API reports failures without a reason, so confirm whether the file
			// is actually gone before deciding this is a real error.
			gone, err := klaudiaFileIsAbsent(c, d.Get("type").(string), d.Id())
			if err != nil {
				return diag.Errorf("Klaudia file %s was reported as failed to delete, and confirming its state failed: %s", d.Id(), err)
			}
			if !gone {
				return diag.Errorf("Klaudia file %s failed to delete and is still present", d.Id())
			}
			d.SetId("")
			return nil
		}
	}
	return nil
}

func klaudiaFileIsAbsent(c *Client, fileType string, fileID string) (bool, error) {
	files, statusCode, err := c.ListKlaudiaFiles(fileType)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return true, nil
		}
		return false, err
	}
	if files == nil {
		return true, nil
	}
	for _, file := range files.Files {
		if file.ID == fileID {
			return false, nil
		}
	}
	return true, nil
}

func buildKlaudiaFilePayload(d *schema.ResourceData) (klaudiaFilePayload, string, error) {
	filename := d.Get("filename").(string)

	raw := d.GetRawConfig()
	if !raw.IsNull() && raw.IsKnown() {
		if pathRaw, ok := extractKlaudiaFileConfigString(raw, "source_path"); ok {
			content, err := os.ReadFile(pathRaw)
			if err != nil {
				return klaudiaFilePayload{}, "", fmt.Errorf("error reading source_path %q: %w", pathRaw, err)
			}
			return klaudiaFilePayload{Filename: filename, Content: content}, sha256Hex(content), nil
		}
		if contentRaw, ok := extractKlaudiaFileConfigString(raw, "content"); ok {
			content := []byte(contentRaw)
			return klaudiaFilePayload{Filename: filename, Content: content}, sha256Hex(content), nil
		}
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
