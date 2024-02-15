package functions

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = EchoFunction{}

func NewProjectFromSelfLinkFunction() function.Function {
	return &ProjectFromSelfLinkFunction{}
}

type ProjectFromSelfLinkFunction struct{}

func (f ProjectFromSelfLinkFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "project_id_from_self_link"
}

func (f ProjectFromSelfLinkFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the project name within the resource self link or id provided as an argument.",
		Description: "Takes a single string argument, which should be a self link or id of a resource. This function will either return the project name from the input string or raise an error due to no project being present in the string. The function uses the presence of \"projects/{{project}}/\" in the input string to identify the project name, e.g. when the function is passed the self link \"https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c/instances/my-instance\" as an argument it will return \"my-project\".",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "self_link",
				Description: "A self link of a resouce, or an id. For example, both \"https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c/instances/my-instance\" and \"projects/my-project/zones/us-central1-c/instances/my-instance\" are valid inputs",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f ProjectFromSelfLinkFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {

	// Load arguments from function call
	var arg0 string
	resp.Diagnostics.Append(req.Arguments.GetArgument(ctx, 0, &arg0)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare how we'll identify project id from input string
	regex := regexp.MustCompile("projects/(?P<ProjectId>[^/]+)/") // Should match the pattern below
	template := "$ProjectId"                                      // Should match the submatch identifier in the regex
	pattern := "projects/{project}/"                              // Human-readable pattern used in errors and warnings

	// Get and return element from input string
	projectId := GetElement(ctx, arg0, regex, template, pattern, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Result.Set(ctx, projectId)...)
}
