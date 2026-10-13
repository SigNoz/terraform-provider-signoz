package model

import (
	"context"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIngestionKey_TagsRoundTrip(t *testing.T) {
	ctx := context.Background()

	original := []string{"env:prod", "team:platform", "service:api"}

	key := &IngestionKey{Tags: original}

	tfList, diags := key.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tfList.Elements()))
	}

	key2 := &IngestionKey{}
	key2.SetTags(ctx, tfList)

	if len(key2.Tags) != len(original) {
		t.Fatalf("expected %d tags, got %d", len(original), len(key2.Tags))
	}
	for i, tag := range original {
		if key2.Tags[i] != tag {
			t.Errorf("tag[%d]: expected %s, got %s", i, tag, key2.Tags[i])
		}
	}
}

func TestIngestionKey_TagsEmpty(t *testing.T) {
	key := &IngestionKey{Tags: []string{}}

	tfList, diags := key.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tfList.Elements()))
	}
}

func TestIngestionKey_TagsNil(t *testing.T) {
	key := &IngestionKey{Tags: nil}

	tfList, diags := key.TagsToTerraform()
	if diags.HasError() {
		t.Fatalf("TagsToTerraform had errors: %v", diags)
	}

	_ = tfList
}

func TestIngestionKey_SetTagsNull(t *testing.T) {
	ctx := context.Background()

	nullList := types.ListNull(types.StringType)

	key := &IngestionKey{Tags: []string{"existing"}}
	key.SetTags(ctx, nullList)

	// Should not modify when null
	if len(key.Tags) != 1 || key.Tags[0] != "existing" {
		t.Errorf("expected tags to remain unchanged, got %v", key.Tags)
	}
}

func TestIngestionKey_SetTagsFromConstructed(t *testing.T) {
	ctx := context.Background()

	tfList, _ := types.ListValue(types.StringType, []tfattr.Value{
		types.StringValue("k1"),
		types.StringValue("k2"),
	})

	key := &IngestionKey{}
	key.SetTags(ctx, tfList)

	if len(key.Tags) != 2 || key.Tags[0] != "k1" || key.Tags[1] != "k2" {
		t.Errorf("unexpected tags: %v", key.Tags)
	}
}

func TestIngestionKey_RateLimitsRoundTrip(t *testing.T) {
	ctx := context.Background()

	key := &IngestionKey{
		RateLimits: &IngestionRateLimit{
			DayLimit:    1000000,
			SecondLimit: 1000,
		},
	}

	tfObj, diags := key.RateLimitsToTerraform()
	if diags.HasError() {
		t.Fatalf("RateLimitsToTerraform had errors: %v", diags)
	}

	if tfObj.IsNull() {
		t.Fatal("expected non-null rate limits object")
	}

	// Verify attributes
	attrs := tfObj.Attributes()
	dayVal, ok := attrs[attr.DayLimit].(types.Int64)
	if !ok {
		t.Fatal("expected day_limit to be Int64")
	}
	if dayVal.ValueInt64() != 1000000 {
		t.Errorf("DayLimit: expected 1000000, got %d", dayVal.ValueInt64())
	}

	secVal, ok := attrs[attr.SecondLimit].(types.Int64)
	if !ok {
		t.Fatal("expected second_limit to be Int64")
	}
	if secVal.ValueInt64() != 1000 {
		t.Errorf("SecondLimit: expected 1000, got %d", secVal.ValueInt64())
	}

	key2 := &IngestionKey{}
	key2.SetRateLimits(ctx, tfObj)

	if key2.RateLimits == nil {
		t.Fatal("expected non-nil rate limits")
	}
	if key2.RateLimits.DayLimit != 1000000 {
		t.Errorf("DayLimit: expected 1000000, got %d", key2.RateLimits.DayLimit)
	}
	if key2.RateLimits.SecondLimit != 1000 {
		t.Errorf("SecondLimit: expected 1000, got %d", key2.RateLimits.SecondLimit)
	}
}

func TestIngestionKey_RateLimitsNil(t *testing.T) {
	key := &IngestionKey{RateLimits: nil}

	tfObj, diags := key.RateLimitsToTerraform()
	if diags.HasError() {
		t.Fatalf("RateLimitsToTerraform had errors: %v", diags)
	}

	if !tfObj.IsNull() {
		t.Error("expected null for nil RateLimits")
	}
}

func TestIngestionKey_SetRateLimitsNull(t *testing.T) {
	ctx := context.Background()

	rateLimitsAttrTypes := attr.RateLimitsAttrTypes()
	nullObj := types.ObjectNull(rateLimitsAttrTypes)

	key := &IngestionKey{}
	key.SetRateLimits(ctx, nullObj)

	if key.RateLimits != nil {
		t.Errorf("expected nil rate limits for null input, got %v", key.RateLimits)
	}
}

func TestIngestionKey_RateLimitsZeroValues(t *testing.T) {
	ctx := context.Background()

	key := &IngestionKey{
		RateLimits: &IngestionRateLimit{
			DayLimit:    0,
			SecondLimit: 0,
		},
	}

	tfObj, diags := key.RateLimitsToTerraform()
	if diags.HasError() {
		t.Fatalf("RateLimitsToTerraform had errors: %v", diags)
	}

	key2 := &IngestionKey{}
	key2.SetRateLimits(ctx, tfObj)

	if key2.RateLimits == nil {
		t.Fatal("expected non-nil rate limits even with zero values")
	}
	if key2.RateLimits.DayLimit != 0 {
		t.Errorf("DayLimit: expected 0, got %d", key2.RateLimits.DayLimit)
	}
	if key2.RateLimits.SecondLimit != 0 {
		t.Errorf("SecondLimit: expected 0, got %d", key2.RateLimits.SecondLimit)
	}
}

func TestIngestionKey_SetRateLimitsFromConstructed(t *testing.T) {
	ctx := context.Background()

	rateLimitsAttrTypes := attr.RateLimitsAttrTypes()
	tfObj, _ := types.ObjectValue(rateLimitsAttrTypes, map[string]tfattr.Value{
		attr.DayLimit:    types.Int64Value(500),
		attr.SecondLimit: types.Int64Value(10),
	})

	key := &IngestionKey{}
	key.SetRateLimits(ctx, tfObj)

	if key.RateLimits == nil {
		t.Fatal("expected non-nil rate limits")
	}
	if key.RateLimits.DayLimit != 500 {
		t.Errorf("DayLimit: expected 500, got %d", key.RateLimits.DayLimit)
	}
	if key.RateLimits.SecondLimit != 10 {
		t.Errorf("SecondLimit: expected 10, got %d", key.RateLimits.SecondLimit)
	}
}
