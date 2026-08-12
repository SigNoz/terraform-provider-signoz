package schemas_test

import (
	"context"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type objectValidatable interface {
	ObjectValidators() []validator.Object
}

// Every flattened-oneOf union must reject "no variant set" / "more than one
// variant set" at plan time (see .claude/docs/oneOf.md).
func TestFlattenedOneOfUnionsCarryExactlyOneValidator(t *testing.T) {
	ctx := context.Background()

	dashboard := schemas.DashboardResourceSchema(ctx)
	rule := schemas.RuleResourceSchema(ctx)

	panelSpec := path.Root("spec").AtName("panels").AtMapKey("p").AtName("spec")
	querySpec := panelSpec.AtName("queries").AtListIndex(0).AtName("spec")
	ruleQuery := path.Root("condition").AtName("composite_query").AtName("queries").AtListIndex(0)

	cases := []struct {
		name   string
		schema schema.Schema
		path   path.Path
	}{
		{"dashboard panel plugin", dashboard, panelSpec.AtName("plugin")},
		{"dashboard query plugin", dashboard, querySpec.AtName("plugin")},
		{"dashboard builder_query.spec", dashboard, querySpec.AtName("plugin").AtName("builder_query").AtName("spec")},
		{"rule builder_query.spec", rule, ruleQuery.AtName("builder_query").AtName("spec")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, diags := c.schema.AttributeAtPath(ctx, c.path)
			if diags.HasError() {
				t.Fatalf("resolve %s: %v", c.path, diags)
			}

			ov, ok := a.(objectValidatable)
			if !ok {
				t.Fatalf("%s: %T carries no object validators", c.path, a)
			}

			if len(ov.ObjectValidators()) == 0 {
				t.Errorf("%s: no ExactlyOneNestedAttribute validator", c.path)
			}
		})
	}
}
