package functions_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/envvar"
)

func TestAccProviderFunction_project_id_from_self_link(t *testing.T) {
	t.Parallel()
	acctest.SkipIfVcr(t) // Need to determine if compatible with VCR, as functions are implemented in PF provider

	projectId := envvar.GetTestProjectFromEnv()
	projectIdRegex := regexp.MustCompile(fmt.Sprintf("^%s$", projectId))

	validSelfLink := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/us-central1-c/instances/my-instance", projectId)
	validId := fmt.Sprintf("projects/%s/zones/us-central1-c/instances/my-instance", projectId)
	repetitiveInput := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/projects/not-this-1/projects/not-this-2/instances/my-instance", projectId)
	invalidInput := "zones/us-central1-c/instances/my-instance"

	context := map[string]interface{}{
		"function_name": "project_id_from_self_link",
		"output_name":   "project_id",
		"self_link":     "", // overridden in test cases
	}

	acctest.VcrTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				// Given valid resource self_link input, the output value matches the expected value
				Config: testProviderFunction_generic_element_from_self_link(context, validSelfLink),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
			{
				// Given valid resource id input, the output value matches the expected value
				Config: testProviderFunction_generic_element_from_self_link(context, validId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
			{
				// Given repetitive input, the output value is the left-most match in the input
				Config: testProviderFunction_generic_element_from_self_link(context, repetitiveInput),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
			{
				// Given invalid input, an error occurs
				Config:      testProviderFunction_generic_element_from_self_link(context, invalidInput),
				ExpectError: regexp.MustCompile("Error in function call"), // ExpectError doesn't inspect the specific error messages
			},
			{
				// Can get the project from a resource's id in one step
				// Uses google_service_account resource's id attribute with format projects/{{project}}/serviceAccounts/{{email}}
				Config: testProviderFunction_get_project_from_resource_id(context),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
			{
				// Can get the project from a resource's self_link in one step
				// Uses google_compute_subnetwork resource's self_link attribute
				Config: testProviderFunction_get_project_from_resource_self_link(context),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
		},
	})
}

func testProviderFunction_generic_element_from_self_link(context map[string]interface{}, selfLink string) string {
	context["self_link"] = selfLink

	return acctest.Nprintf(`
	# terraform block required for provider function to be found
	terraform {
		required_providers {
			google = {
				source = "hashicorp/google"
			}
		}
	}

	output "%{output_name}" {
		value = provider::google::%{function_name}("%{self_link}")
	}
`, context)
}

func testProviderFunction_get_project_from_resource_id(context map[string]interface{}) string {
	return acctest.Nprintf(`
# terraform block required for provider function to be found
terraform {
	required_providers {
		google = {
			source = "hashicorp/google"
		}
	}
}

resource "google_service_account" "service_account" {
	account_id   = "tf-test-project-id-func"
	display_name = "Testing use of provider function %{function_name}"
}

output "%{output_name}" {
	value = provider::google::%{function_name}(google_service_account.service_account.id)
}
`, context)
}

func testProviderFunction_get_project_from_resource_self_link(context map[string]interface{}) string {
	return acctest.Nprintf(`
# terraform block required for provider function to be found
terraform {
	required_providers {
		google = {
			source = "hashicorp/google"
		}
	}
}

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_subnetwork" "subnet" {
  name          = "tf-test-project-id-func"
  ip_cidr_range = "10.2.0.0/16"
  network        = data.google_compute_network.default.id
}

output "%{output_name}" {
	value = provider::google::%{function_name}(google_compute_subnetwork.subnet.self_link)
}
`, context)
}
