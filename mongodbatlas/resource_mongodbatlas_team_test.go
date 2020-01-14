package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	matlas "github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasTeam_basic(t *testing.T) {
	var team matlas.Team

	resourceName := "mongodbatlas_teams.test"
	orgID := os.Getenv("MONGODB_ATLAS_ORG_ID")
	projectID := os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	name := fmt.Sprintf("test-acc-%s", acctest.RandString(10))

	updatedName := fmt.Sprintf("test-acc-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasTeamConfig(orgID, projectID, name,
					[]string{
						"mongodbatlas.testing@gmail.com",
						"francisco.preciado@digitalonus.com",
						"antonio.cabrera@digitalonus.com",
					},
					[]string{"GROUP_READ_ONLY"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasTeamExists(resourceName, &team),
					testAccCheckMongoDBAtlasTeamAttributes(&team, name),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "usernames.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "team_roles.#", "1"),
				),
			},
			{
				Config: testAccMongoDBAtlasTeamConfig(orgID, projectID, updatedName,
					[]string{
						"marin.salinas@digitalonus.com",
						"antonio.cabrera@digitalonus.com",
					},
					[]string{
						"GROUP_DATA_ACCESS_ADMIN",
						"GROUP_READ_ONLY",
					}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasTeamExists(resourceName, &team),
					testAccCheckMongoDBAtlasTeamAttributes(&team, updatedName),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "usernames.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "team_roles.#", "2"),
				),
			},
			{
				Config: testAccMongoDBAtlasTeamConfig(orgID, projectID, updatedName,
					[]string{
						"marin.salinas@digitalonus.com",
						"mongodbatlas.testing@gmail.com",
						"francisco.preciado@digitalonus.com",
					},
					[]string{
						"GROUP_OWNER",
					}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasTeamExists(resourceName, &team),
					testAccCheckMongoDBAtlasTeamAttributes(&team, updatedName),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "usernames.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "team_roles.#", "1"),
				),
			},
		},
	})

}

func TestAccResourceMongoDBAtlasTeam_importBasic(t *testing.T) {
	orgID := os.Getenv("MONGODB_ATLAS_ORG_ID")
	projectID := os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	name := fmt.Sprintf("test-acc-%s", acctest.RandString(10))
	resourceName := "mongodbatlas_teams.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMongoDBAtlasTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasTeamConfig(orgID, projectID, name,
					[]string{"mongodbatlas.testing@gmail.com"},
					[]string{"GROUP_READ_ONLY"}),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckMongoDBAtlasTeamStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMongoDBAtlasTeamExists(resourceName string, team *matlas.Team) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*matlas.Client)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		orgID := rs.Primary.Attributes["org_id"]
		id := rs.Primary.Attributes["team_id"]

		if orgID == "" && id == "" {
			return fmt.Errorf("no ID is set")
		}

		log.Printf("[DEBUG] orgID: %s", orgID)
		log.Printf("[DEBUG] teamID: %s", id)

		teamResp, _, err := conn.Teams.Get(context.Background(), orgID, id)
		if err == nil {
			*team = *teamResp
			return nil
		}
		return fmt.Errorf("team(%s) does not exist", id)
	}
}

func testAccCheckMongoDBAtlasTeamAttributes(team *matlas.Team, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if team.Name != name {
			return fmt.Errorf("bad name: %s", team.Name)
		}
		return nil
	}
}

func testAccCheckMongoDBAtlasTeamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*matlas.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_teams" {
			continue
		}

		orgID := rs.Primary.Attributes["org_id"]
		id := rs.Primary.Attributes["team_id"]

		// Try to find the team
		_, _, err := conn.Teams.Get(context.Background(), orgID, id)
		if err == nil {
			return fmt.Errorf("team (%s) still exists", id)
		}
	}
	return nil
}

func testAccCheckMongoDBAtlasTeamStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		orgID := rs.Primary.Attributes["org_id"]
		id := rs.Primary.Attributes["team_id"]
		projectID := rs.Primary.Attributes["project_id"]

		return fmt.Sprintf("%s-%s-%s", orgID, id, projectID), nil
	}
}

func testAccMongoDBAtlasTeamConfig(orgID, projectID, name string, usernames, roles []string) string {
	var teamRoles string
	if len(roles) > 0 {
		teamRoles = fmt.Sprintf(`
			team_roles = %s
		`, strings.ReplaceAll(fmt.Sprintf("%+q", roles), " ", ","))
	}
	return fmt.Sprintf(`
		resource "mongodbatlas_teams" "test" {
			org_id     = "%s"
			project_id = "%s"
			name       = "%s"
			usernames  = %s
			%s
		}`, orgID, projectID, name,
		strings.ReplaceAll(fmt.Sprintf("%+q", usernames), " ", ","),
		teamRoles,
	)
}
