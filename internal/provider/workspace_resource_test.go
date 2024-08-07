package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkspaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "autonomi_workspace" "test_workspace" {
	name = "test_resource_name"
	description = "test_resource_description"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item
					resource.TestCheckResourceAttr("autonomi_workspace.test_workspace", "name", "test_resource_name"),
					resource.TestCheckResourceAttr("autonomi_workspace.test_workspace", "description", "test_resource_description"),
					// check if the ID is set to ensure the resource is created
					resource.TestCheckResourceAttrSet("autonomi_workspace.test_workspace", "id"),
					resource.TestCheckResourceAttrSet("autonomi_workspace.test_workspace", "created_at"),
					resource.TestCheckResourceAttrSet("autonomi_workspace.test_workspace", "updated_at"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
