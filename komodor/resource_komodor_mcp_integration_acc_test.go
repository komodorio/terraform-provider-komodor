package komodor

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	registerAccTest("komodor_mcp_integration")
}

var accTestMCPIntegrationID, accTestMCPSkillID string

func TestAcc_komodor_mcp_integration_basic(t *testing.T) {
	skillName := testResourceName("mcp-skill")
	intName := testResourceName("mcp")
	updatedName := intName + "-updated"
	intAddr := "komodor_mcp_integration.test"
	skillAddr := "komodor_klaudia_skill.for_mcp"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMCPIntegrationAndSkillDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccMCPIntegrationConfig(skillName, intName, "https://tf-acc.example.invalid/mcp-initial"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(intAddr, "name", intName),
					resource.TestCheckResourceAttrPair(intAddr, "skill_id", skillAddr, "id"),
					resource.TestCheckResourceAttrSet(intAddr, "id"),
					testAccCaptureMCPIDs(intAddr, skillAddr),
				),
			},
			{
				Config: testAccMCPIntegrationConfig(skillName, updatedName, "https://tf-acc.example.invalid/mcp-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(intAddr, "name", updatedName),
					testAccCaptureMCPIDs(intAddr, skillAddr),
				),
			},
		},
	})
}

func testAccCaptureMCPIDs(integrationAddr, skillAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ir, ok := s.RootModule().Resources[integrationAddr]
		if !ok {
			return fmt.Errorf("not found: %s", integrationAddr)
		}
		sr, ok := s.RootModule().Resources[skillAddr]
		if !ok {
			return fmt.Errorf("not found: %s", skillAddr)
		}
		accTestMCPIntegrationID = ir.Primary.ID
		accTestMCPSkillID = sr.Primary.ID
		return nil
	}
}

func testAccCheckMCPIntegrationAndSkillDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	mcpID := accTestMCPIntegrationID
	skillID := accTestMCPSkillID
	accTestMCPIntegrationID = ""
	accTestMCPSkillID = ""

	if mcpID != "" {
		_, sc, _ := client.GetMCPIntegration(mcpID)
		if sc == http.StatusOK {
			return fmt.Errorf("MCP integration %q still exists after destroy", mcpID)
		}
	}

	if skillID != "" {
		_, sc, _ := client.GetSkill(skillID)
		if sc == http.StatusOK {
			return fmt.Errorf("Klaudia skill %q still exists after destroy", skillID)
		}
	}
	return nil
}

func TestAcc_komodor_mcp_integration_token_exchange(t *testing.T) {
	skillName := testResourceName("te-skill")
	intName := testResourceName("te-mcp")
	updatedName := intName + "-updated"
	intAddr := "komodor_mcp_integration.te_test"
	skillAddr := "komodor_klaudia_skill.te_skill"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMCPIntegrationAndSkillDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccMCPIntegrationTokenExchangeConfig(skillName, intName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(intAddr, "name", intName),
				resource.TestCheckResourceAttr(intAddr, "connectivity.0.mode", "public"),
				resource.TestCheckResourceAttr(intAddr, "mcp_server.0.url", "http://mock-mcp-server.mcp-test.svc:8082/mcp"),
					resource.TestCheckResourceAttr(intAddr, "mcp_server.0.transport", "streamable-http"),
					resource.TestCheckResourceAttr(intAddr, "auth.0.method", "token_exchange"),
					resource.TestCheckResourceAttr(intAddr, "auth.0.token_exchange.0.token_url", "http://mock-auth-server.mcp-test.svc:8081/token"),
					resource.TestCheckResourceAttr(intAddr, "auth.0.token_exchange.0.audience", "mock-mcp-server"),
					resource.TestCheckResourceAttrPair(intAddr, "skill_id", skillAddr, "id"),
					resource.TestCheckResourceAttrSet(intAddr, "id"),
					testAccCaptureMCPIDs(intAddr, skillAddr),
				),
			},
			// Step 2: rename — exercises resourceMCPIntegrationUpdate (line 364)
			{
				Config: testAccMCPIntegrationTokenExchangeConfigUpdated(skillName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(intAddr, "name", updatedName),
					resource.TestCheckResourceAttr(intAddr, "auth.0.token_exchange.0.audience", "mock-mcp-server-v2"),
					testAccCaptureMCPIDs(intAddr, skillAddr),
				),
			},
		},
	})
}

func testAccMCPIntegrationTokenExchangeConfig(skillName, integrationName string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_skill" "te_skill" {
  name         = %q
  description  = "acceptance test skill for token-exchange MCP"
  instructions = "Skill used for token-exchange acceptance test."
  use_cases    = ["chat"]
  clusters     = ["*"]
  is_enabled   = true
}

resource "komodor_mcp_integration" "te_test" {
  name     = %q
  skill_id = komodor_klaudia_skill.te_skill.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "http://mock-mcp-server.mcp-test.svc:8082/mcp"
    transport = "streamable-http"
  }

  auth {
    method = "token_exchange"

    token_exchange {
      token_url  = "http://mock-auth-server.mcp-test.svc:8081/token"
      grant_type = "urn:ietf:params:oauth:grant-type:token-exchange"

      subject_token {
        value = "test-subject-token"
        type  = "urn:ietf:params:oauth:token-type:jwt"
      }

      audience             = "mock-mcp-server"
      requested_token_type = "urn:ietf:params:oauth:token-type:access_token"
    }

    upstream_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }

    response {
      token_field      = "access_token"
      token_type_field = "token_type"
      expires_in_field = "expires_in"
    }
  }
}
`, skillName, integrationName)
}

func testAccMCPIntegrationTokenExchangeConfigUpdated(skillName, integrationName string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_skill" "te_skill" {
  name         = %q
  description  = "acceptance test skill for token-exchange MCP"
  instructions = "Skill used for token-exchange acceptance test."
  use_cases    = ["chat"]
  clusters     = ["*"]
  is_enabled   = true
}

resource "komodor_mcp_integration" "te_test" {
  name     = %q
  skill_id = komodor_klaudia_skill.te_skill.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = "http://mock-mcp-server.mcp-test.svc:8082/mcp"
    transport = "streamable-http"
  }

  auth {
    method = "token_exchange"

    token_exchange {
      token_url  = "http://mock-auth-server.mcp-test.svc:8081/token"
      grant_type = "urn:ietf:params:oauth:grant-type:token-exchange"

      subject_token {
        value = "test-subject-token"
        type  = "urn:ietf:params:oauth:token-type:jwt"
      }

      audience             = "mock-mcp-server-v2"
      requested_token_type = "urn:ietf:params:oauth:token-type:access_token"
    }

    upstream_header {
      name   = "Authorization"
      format = "{token_type} {access_token}"
    }

    response {
      token_field      = "access_token"
      token_type_field = "token_type"
      expires_in_field = "expires_in"
    }
  }
}
`, skillName, integrationName)
}

func testAccMCPIntegrationConfig(skillName, integrationName, mcpURL string) string {
	return fmt.Sprintf(`
resource "komodor_klaudia_skill" "for_mcp" {
  name         = %q
  description  = "acceptance test skill for MCP integration"
  instructions = "Skill used only to satisfy MCP integration acceptance tests."
  use_cases    = ["chat", "rca"]
  clusters     = ["*"]
  is_enabled   = true
}

resource "komodor_mcp_integration" "test" {
  name     = %q
  skill_id = komodor_klaudia_skill.for_mcp.id

  connectivity {
    mode = "public"
  }

  mcp_server {
    url       = %q
    transport = "sse"
  }

  auth {
    method = "none"
  }
}
`, skillName, integrationName, mcpURL)
}
