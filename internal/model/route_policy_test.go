package model

import (
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRoutePolicy_MatchersRoundTrip(t *testing.T) {
	rp := &RoutePolicy{
		Matchers: []RouteMatcher{
			{Name: "severity", Value: "critical", MatchType: "="},
			{Name: "env", Value: "prod", MatchType: "=~"},
		},
	}

	tfList, diags := rp.MatchersToTerraform()
	if diags.HasError() {
		t.Fatalf("MatchersToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 2 {
		t.Fatalf("expected 2 matchers, got %d", len(tfList.Elements()))
	}

	rp2 := &RoutePolicy{}
	rp2.SetMatchers(tfList)

	if len(rp2.Matchers) != 2 {
		t.Fatalf("expected 2 matchers, got %d", len(rp2.Matchers))
	}

	if rp2.Matchers[0].Name != "severity" {
		t.Errorf("matcher[0].Name: expected severity, got %s", rp2.Matchers[0].Name)
	}
	if rp2.Matchers[0].Value != "critical" {
		t.Errorf("matcher[0].Value: expected critical, got %s", rp2.Matchers[0].Value)
	}
	if rp2.Matchers[0].MatchType != "=" {
		t.Errorf("matcher[0].MatchType: expected =, got %s", rp2.Matchers[0].MatchType)
	}
	if rp2.Matchers[1].Name != "env" {
		t.Errorf("matcher[1].Name: expected env, got %s", rp2.Matchers[1].Name)
	}
	if rp2.Matchers[1].MatchType != "=~" {
		t.Errorf("matcher[1].MatchType: expected =~, got %s", rp2.Matchers[1].MatchType)
	}
}

func TestRoutePolicy_MatchersEmpty(t *testing.T) {
	rp := &RoutePolicy{Matchers: []RouteMatcher{}}

	tfList, diags := rp.MatchersToTerraform()
	if diags.HasError() {
		t.Fatalf("MatchersToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 matchers, got %d", len(tfList.Elements()))
	}
}

func TestRoutePolicy_MatchersNil(t *testing.T) {
	rp := &RoutePolicy{Matchers: nil}

	tfList, diags := rp.MatchersToTerraform()
	if diags.HasError() {
		t.Fatalf("MatchersToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 matchers for nil, got %d", len(tfList.Elements()))
	}
}

func TestRoutePolicy_ChannelIDsRoundTrip(t *testing.T) {
	rp := &RoutePolicy{
		ChannelIDs: []string{"ch-1", "ch-2", "ch-3"},
	}

	tfList, diags := rp.ChannelIDsToTerraform()
	if diags.HasError() {
		t.Fatalf("ChannelIDsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 channel IDs, got %d", len(tfList.Elements()))
	}

	rp2 := &RoutePolicy{}
	rp2.SetChannelIDs(tfList)

	if len(rp2.ChannelIDs) != 3 {
		t.Fatalf("expected 3 channel IDs, got %d", len(rp2.ChannelIDs))
	}

	for i, id := range []string{"ch-1", "ch-2", "ch-3"} {
		if rp2.ChannelIDs[i] != id {
			t.Errorf("channelID[%d]: expected %s, got %s", i, id, rp2.ChannelIDs[i])
		}
	}
}

func TestRoutePolicy_ChannelIDsEmpty(t *testing.T) {
	rp := &RoutePolicy{ChannelIDs: []string{}}

	tfList, diags := rp.ChannelIDsToTerraform()
	if diags.HasError() {
		t.Fatalf("ChannelIDsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 channel IDs, got %d", len(tfList.Elements()))
	}
}

func TestRoutePolicy_SetMatchersFromConstructed(t *testing.T) {
	// Build a Terraform list manually
	matcherAttrTypes := MatcherAttrTypes()
	elemType := types.ObjectType{AttrTypes: matcherAttrTypes}

	obj1, _ := types.ObjectValue(matcherAttrTypes, map[string]tfattr.Value{
		attr.MatchName:  types.StringValue("alertname"),
		attr.MatchValue: types.StringValue("HighCPU"),
		attr.MatchType:  types.StringValue("="),
	})

	tfList, _ := types.ListValue(elemType, []tfattr.Value{obj1})

	rp := &RoutePolicy{}
	rp.SetMatchers(tfList)

	if len(rp.Matchers) != 1 {
		t.Fatalf("expected 1 matcher, got %d", len(rp.Matchers))
	}
	if rp.Matchers[0].Name != "alertname" {
		t.Errorf("expected Name=alertname, got %s", rp.Matchers[0].Name)
	}
	if rp.Matchers[0].Value != "HighCPU" {
		t.Errorf("expected Value=HighCPU, got %s", rp.Matchers[0].Value)
	}
}

func TestRoutePolicy_SetChannelIDsFromConstructed(t *testing.T) {
	tfList, _ := types.ListValue(types.StringType, []tfattr.Value{
		types.StringValue("channel-abc"),
		types.StringValue("channel-def"),
	})

	rp := &RoutePolicy{}
	rp.SetChannelIDs(tfList)

	if len(rp.ChannelIDs) != 2 {
		t.Fatalf("expected 2 channel IDs, got %d", len(rp.ChannelIDs))
	}
	if rp.ChannelIDs[0] != "channel-abc" || rp.ChannelIDs[1] != "channel-def" {
		t.Errorf("unexpected channel IDs: %v", rp.ChannelIDs)
	}
}

func TestMatcherAttrTypes(t *testing.T) {
	attrTypes := MatcherAttrTypes()
	if len(attrTypes) != 3 {
		t.Errorf("expected 3 attr types, got %d", len(attrTypes))
	}
	if _, ok := attrTypes[attr.MatchName]; !ok {
		t.Error("missing match_name attr type")
	}
	if _, ok := attrTypes[attr.MatchValue]; !ok {
		t.Error("missing match_value attr type")
	}
	if _, ok := attrTypes[attr.MatchType]; !ok {
		t.Error("missing match_type attr type")
	}
}
