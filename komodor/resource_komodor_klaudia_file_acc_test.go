package komodor

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_klaudia_file")
}

var accTestKlaudiaFileID string

func TestAcc_komodor_klaudia_file_basic(t *testing.T) {
	dir := t.TempDir()
	initialPath := filepath.Join(dir, "initial.md")
	updatedPath := filepath.Join(dir, "updated.md")
	if err := os.WriteFile(initialPath, []byte("# Initial knowledge\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(updatedPath, []byte("# Updated knowledge\n"), 0600); err != nil {
		t.Fatal(err)
	}

	filename := testResourceName("klaudia-file") + ".md"
	resourceAddr := "komodor_klaudia_file.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKlaudiaFileDestroyed("knowledge-base"),
		Steps: []resource.TestStep{
			{
				Config: testAccKlaudiaFileConfig("knowledge-base", filename, initialPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "type", "knowledge-base"),
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "checksum"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					testAccCaptureKlaudiaFileID(resourceAddr),
				),
			},
			{
				Config: testAccKlaudiaFileConfig("knowledge-base", filename, updatedPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "checksum"),
					testAccCaptureKlaudiaFileID(resourceAddr),
				),
			},
		},
	})
}

func testAccCaptureKlaudiaFileID(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("not found: %s", addr)
		}
		accTestKlaudiaFileID = rs.Primary.ID
		return nil
	}
}

func testAccCheckKlaudiaFileDestroyed(fileType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := accTestKlaudiaFileID
		accTestKlaudiaFileID = ""
		if id == "" {
			return nil
		}
		client := testAccProvider.Meta().(*Client)
		files, sc, _ := client.ListKlaudiaFiles(fileType)
		if sc == http.StatusNotFound {
			return nil
		}
		if sc != http.StatusOK {
			return nil
		}
		for _, file := range files.Files {
			if file.ID == id {
				return fmt.Errorf("Klaudia file %q still exists after destroy", id)
			}
		}
		return nil
	}
}

func testAccKlaudiaFileConfig(fileType, filename, sourcePath string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_file" "test" {
  type        = %q
  filename    = %q
  source_path = %q
}
`, fileType, filename, sourcePath)
}
