package functions_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tpg_functions "github.com/hashicorp/terraform-provider-google/google/functions"
)

func TestFunctionInternals_GetElementFromSelfLink(t *testing.T) {

	regex := regexp.MustCompile("two/(?P<Element>[^/]+)/")
	template := "$Element"
	pattern := "two/{two}/"

	cases := map[string]struct {
		Input           string
		ExpectedElement string
		ExpectError     bool
		ExpectWarning   bool
	}{
		"it can pull out a value from a string using a regex with a submatch": {
			Input:           "one/element-1/two/element-2/three/element-3",
			ExpectedElement: "element-2",
		},
		"it sets an error in diags if no match is found": {
			Input:       "one/element-1/three/element-3",
			ExpectError: true,
		},
		"it sets a warning in diags if more than one match is found": {
			Input:           "two/element-2/two/element-2/two/element-2",
			ExpectedElement: "element-2",
			ExpectWarning:   true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			// Arrange
			ctx := context.Background()
			req := function.RunRequest{}
			resp := function.RunResponse{
				Diagnostics: diag.Diagnostics{},
				Result:      function.NewResultData(basetypes.StringValue{}), // Required to avoid nil pointer error
			}

			// Act
			result := tpg_functions.GetElementFromSelfLink(ctx, tc.Input, regex, template, pattern, req, &resp)

			// Assert
			if resp.Diagnostics.HasError() && !tc.ExpectError {
				t.Fatalf("Unexpected error(s) were set in response diags: %s", resp.Diagnostics.Errors())
			}
			if !resp.Diagnostics.HasError() && tc.ExpectError {
				t.Fatal("Expected error(s) to be set in response diags, but there were none.")
			}
			if (resp.Diagnostics.WarningsCount() > 0) && !tc.ExpectWarning {
				t.Fatalf("Unexpected warning(s) were set in response diags: %s", resp.Diagnostics.Warnings())
			}
			if (resp.Diagnostics.WarningsCount() == 0) && tc.ExpectWarning {
				t.Fatal("Expected warning(s) to be set in response diags, but there were none.")
			}
			if resp.Diagnostics.HasError() {
				// Stop test before checking returned value if errors occur
				return
			}

			if result != tc.ExpectedElement {
				t.Fatalf("Expected function logic to retrieve %s from input %s, got %s", tc.ExpectedElement, tc.Input, result)
			}
		})
	}
}
