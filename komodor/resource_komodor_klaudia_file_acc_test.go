package komodor

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
			{
				ResourceName: resourceAddr,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceAddr]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceAddr)
					}
					return fmt.Sprintf("knowledge-base:%s", rs.Primary.ID), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content", "source_path", "checksum"},
			},
		},
	})
}

func TestAcc_komodor_klaudia_file_content(t *testing.T) {
	filename := testResourceName("klaudia-file-content") + ".md"
	resourceAddr := "komodor_klaudia_file.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKlaudiaFileDestroyed("knowledge-base"),
		Steps: []resource.TestStep{
			{
				Config: testAccKlaudiaFileContentConfig("knowledge-base", filename, "# Initial content\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "type", "knowledge-base"),
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "checksum"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					testAccCaptureKlaudiaFileID(resourceAddr),
				),
			},
			{
				Config: testAccKlaudiaFileContentConfig("knowledge-base", filename, "# Updated content\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "filename", filename),
					resource.TestCheckResourceAttrSet(resourceAddr, "checksum"),
					testAccCaptureKlaudiaFileID(resourceAddr),
				),
			},
		},
	})
}

func testAccKlaudiaFileContentConfig(fileType, filename, content string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_file" "test" {
  type     = %q
  filename = %q
  content  = %q
}
`, fileType, filename, content)
}

func TestAcc_komodor_klaudia_file_duplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.md")
	if err := os.WriteFile(path, []byte("# Duplicate\n"), 0600); err != nil {
		t.Fatal(err)
	}

	filename := testResourceName("klaudia-file-dup") + ".md"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKlaudiaFileDestroyed("knowledge-base"),
		Steps: []resource.TestStep{
			{
				Config: testAccKlaudiaFileConfig("knowledge-base", filename, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCaptureKlaudiaFileID("komodor_klaudia_file.test"),
				),
			},
			{
				Config:      testAccKlaudiaFileDuplicateConfig(filename, path),
				ExpectError: regexp.MustCompile(`already exists`),
			},
		},
	})
}

func testAccKlaudiaFileDuplicateConfig(filename, sourcePath string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_file" "test" {
  type        = "knowledge-base"
  filename    = %q
  source_path = %q
}

resource "komodor_klaudia_file" "dup" {
  type        = "knowledge-base"
  filename    = %q
  source_path = %q
}
`, filename, sourcePath, filename, sourcePath)
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
