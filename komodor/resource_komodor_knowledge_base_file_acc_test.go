package komodor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_knowledge_base_file")
}

func TestAcc_komodor_knowledge_base_file_basic(t *testing.T) {
	filename := testResourceName("knowledge-base-file") + ".md"
	resourceAddr := "komodor_knowledge_base_file.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKnowledgeBaseFileDestroyed(filename, "knowledge-base"),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseFileConfig(filename, "knowledge-base", "# Test Runbook\nThis is a test runbook."),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "file_type", "knowledge-base"),
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttrSet(resourceAddr, "uploaded_at"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_by_email"),
				),
			},
		},
	})
}

func TestAcc_komodor_knowledge_base_file_blueprint(t *testing.T) {
	filename := testResourceName("blueprint-file") + ".md"
	resourceAddr := "komodor_knowledge_base_file.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKnowledgeBaseFileDestroyed(filename, "blueprint"),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseFileConfig(filename, "blueprint", "# Test Blueprint\nThis is a test blueprint."),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "file_type", "blueprint"),
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func TestAcc_komodor_knowledge_base_file_with_clusters(t *testing.T) {
	filename := testResourceName("knowledge-base-file-clusters") + ".md"
	resourceAddr := "komodor_knowledge_base_file.test"

	clusters := &struct{ include, exclude []string }{
		include: []string{"production-*"},
		exclude: []string{"production-dev-*"},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKnowledgeBaseFileDestroyed(filename, "knowledge-base"),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseFileWithClustersConfig(filename, clusters.include, clusters.exclude),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "file_type", "knowledge-base"),
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
		},
	})
}

func testAccCheckKnowledgeBaseFileDestroyed(filename, fileType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)
		files, _, err := client.ListKnowledgeBaseFiles(fileType)
		if err != nil {
			return fmt.Errorf("error listing knowledge base files: %s", err)
		}
		for _, f := range files {
			if f.Name == filename {
				return fmt.Errorf("knowledge base file %q still exists after destroy", filename)
			}
		}
		return nil
	}
}

func testAccKnowledgeBaseFileConfig(filename, fileType, content string) string {
	return fmt.Sprintf(`
resource "komodor_knowledge_base_file" "test" {
  file_type = %q
  filename  = %q
  content   = %q
}
`, fileType, filename, content)
}

func testAccKnowledgeBaseFileWithClustersConfig(filename string, include, exclude []string) string {
	includeStr := `["` + strings.Join(include, `", "`) + `"]`
	excludeStr := `["` + strings.Join(exclude, `", "`) + `"]`
	return fmt.Sprintf(`
resource "komodor_knowledge_base_file" "test" {
  file_type = "knowledge-base"
  filename  = %q
  content   = "# Cluster-scoped Runbook\nThis runbook is scoped to specific clusters."

  clusters {
    include = %s
    exclude = %s
  }
}
`, filename, includeStr, excludeStr)
}
