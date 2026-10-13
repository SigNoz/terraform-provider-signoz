package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Guards the registry after the legacy-vs-upstream adoption: the four collided
// names must resolve to exactly one implementation each, and the broken
// signoz_downtime_schedule must be gone.
func TestProviderRegistry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	p := New("test", "test", "signoz")()

	var md provider.MetadataRequest

	var mdr provider.MetadataResponse

	p.Metadata(ctx, md, &mdr)

	resourceCtors := p.Resources(ctx)

	var gotResources []string

	for _, ctor := range resourceCtors {
		var resp resource.MetadataResponse

		ctor().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: mdr.TypeName}, &resp)

		gotResources = append(gotResources, resp.TypeName)
	}

	dsCtors := p.DataSources(ctx)

	var gotDataSources []string

	for _, ctor := range dsCtors {
		var resp datasource.MetadataResponse

		ctor().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: mdr.TypeName}, &resp)

		gotDataSources = append(gotDataSources, resp.TypeName)
	}

	wantResources := []string{
		"signoz_alert",
		"signoz_auth_domain",
		"signoz_dashboard",
		"signoz_data_retention",
		"signoz_ingestion_key",
		"signoz_ingestion_key_limit",
		"signoz_log_pipeline",
		"signoz_notification_channel",
		"signoz_planned_maintenance",
		"signoz_plugin_dashboard",
		"signoz_role",
		"signoz_route_policy",
		"signoz_rule",
		"signoz_saved_view",
		"signoz_service_account",
		"signoz_service_account_role",
		"signoz_user",
		"signoz_user_role",
	}

	wantDataSources := []string{
		"signoz_alert",
		"signoz_auth_domain",
		"signoz_dashboard",
		"signoz_data_retention",
		"signoz_ingestion_key",
		"signoz_log_pipeline",
		"signoz_notification_channel",
		"signoz_planned_maintenance",
		"signoz_role",
		"signoz_route_policy",
		"signoz_rule",
		"signoz_saved_view",
		"signoz_service_account",
		"signoz_service_account_role",
		"signoz_user",
		"signoz_user_role",
	}

	assertExactSet(t, "resources", gotResources, wantResources)

	assertExactSet(t, "data sources", gotDataSources, wantDataSources)
}

func assertExactSet(t *testing.T, kind string, got, want []string) {
	t.Helper()

	seen := map[string]int{}

	for _, name := range got {
		seen[name]++

		if seen[name] > 1 {
			t.Errorf("%s: duplicate type name %q", kind, name)
		}
	}

	slices.Sort(got)

	slices.Sort(want)

	if len(got) != len(want) {
		t.Fatalf("%s: got %d type names, want %d (got=%v)", kind, len(got), len(want), got)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s: got %v, want %v", kind, got, want)

			return
		}
	}
}
