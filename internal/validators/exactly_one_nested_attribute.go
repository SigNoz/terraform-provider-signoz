package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type exactlyOneNestedAttributeValidator struct {
	names []string
}

func ExactlyOneNestedAttribute(names ...string) validator.Object {
	return exactlyOneNestedAttributeValidator{names: names}
}

func (v exactlyOneNestedAttributeValidator) Description(context.Context) string {
	return fmt.Sprintf("Exactly one of %s must be configured.", strings.Join(v.names, ", "))
}

func (v exactlyOneNestedAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v exactlyOneNestedAttributeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attributes := req.ConfigValue.Attributes()
	var configured []string
	for _, name := range v.names {
		value, ok := attributes[name]
		if !ok || value.IsNull() || value.IsUnknown() {
			continue
		}

		configured = append(configured, name)
	}

	if len(configured) == 1 {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid oneOf configuration",
		fmt.Sprintf("Exactly one of %s must be configured. Currently configured: %s.", strings.Join(v.names, ", "), configuredNames(configured)),
	)
}

func configuredNames(names []string) string {
	if len(names) == 0 {
		return "none"
	}

	return strings.Join(names, ", ")
}
