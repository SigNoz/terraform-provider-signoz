package conv_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	conv "github.com/SigNoz/terraform-provider-signoz/internal/convertors"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// probe decodes a terraform-JSON resource body against the dashboard resource
// schema and prints the exact POST body the provider would send on create.
func probe(t *testing.T, body []byte) {
	ctx := context.Background()
	s := schemas.DashboardResourceSchema(ctx)

	raw, err := tftypes.ValueFromJSON(body, s.Type().TerraformType(ctx))
	if err != nil {
		t.Fatalf("ValueFromJSON: %v", err)
	}

	var m schemas.DashboardModel
	st := tfsdk.State{Schema: s, Raw: raw}
	if diags := st.Get(ctx, &m); diags.HasError() {
		t.Fatalf("State.Get: %v", diags)
	}

	in, diags := conv.ExpandDashboardtypesPostableDashboardV2(ctx, m)
	if diags.HasError() {
		t.Fatalf("Expand: %v", diags)
	}

	out, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	t.Logf("POST body:\n%s", out)
}

func TestProbeFile(t *testing.T) {
	path := os.Getenv("PROBE_FILE")
	if path == "" {
		t.Skip("PROBE_FILE unset")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Resource map[string]map[string]json.RawMessage `json:"resource"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	for _, named := range doc.Resource {
		for name, body := range named {
			t.Logf("resource %s", name)
			probe(t, body)
		}
	}
}
