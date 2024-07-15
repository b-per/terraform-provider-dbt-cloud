package service_token_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudServiceTokenResource(t *testing.T) {

	serviceTokenName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	serviceTokenName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudServiceTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudServiceTokenResourceBasicConfig(
					serviceTokenName,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudServiceTokenExists(
						"dbtcloud_service_token.test_service_token",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"name",
						serviceTokenName,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_service_token.test_service_token",
						"token_string",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.#",
						"3",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.permission_set",
						"job_admin",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.all_projects",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.project_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.1.permission_set",
						"git_admin",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.1.all_projects",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.2.permission_set",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.2.all_projects",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.2.writable_environment_categories.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.2.writable_environment_categories.0",
						"development",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudServiceTokenResourceFullConfig(
					serviceTokenName2,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudServiceTokenExists(
						"dbtcloud_service_token.test_service_token",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"name",
						serviceTokenName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.permission_set",
						"git_admin",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.all_projects",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.0.project_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.1.all_projects",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_service_token.test_service_token",
						"service_token_permissions.1.permission_set",
						"job_admin",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_service_token.test_service_token",
						"token_string",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_service_token.test_service_token",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"token_string",
					// being a set, we need to ignore all the project_id as we don't know which one is the one with 0
					"service_token_permissions.0.project_id",
					"service_token_permissions.1.project_id",
					"service_token_permissions.2.project_id",
				},
			},
		},
	})
}

func testAccDbtCloudServiceTokenResourceBasicConfig(serviceTokenName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_service_token" "test_service_token" {
    name = "%s"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = true
    }
    service_token_permissions {
		permission_set = "job_admin"
        all_projects = false
        project_id = dbtcloud_project.test_project.id
    }
    service_token_permissions {
	    permission_set = "developer"
	    all_projects = true
	    writable_environment_categories = ["development"]
    }
}
`, projectName, serviceTokenName)
}

func testAccDbtCloudServiceTokenResourceFullConfig(serviceTokenName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_service_token" "test_service_token" {
    name = "%s"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = false
        project_id = dbtcloud_project.test_project.id
    }
    service_token_permissions {
        permission_set = "job_admin"
        all_projects = true
    }
    // service_token_permissions {
    //     permission_set = "developer"
    //     all_projects = true
    //     // writable_environment_categories = ["development", "staging"]
    // }
}
`, projectName, serviceTokenName)
}

func testAccCheckDbtCloudServiceTokenExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		ServiceTokenID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get ServiceTokenID")
		}
		_, err = apiClient.GetServiceToken(ServiceTokenID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudServiceTokenDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_service_token" {
			continue
		}
		ServiceTokenID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get ServiceTokenID")
		}
		_, err = apiClient.GetServiceToken(ServiceTokenID)
		if err == nil {
			return fmt.Errorf("ServiceToken still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
