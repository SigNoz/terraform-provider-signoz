package model

import (
	"context"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IngestionKey model.
type IngestionKey struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Value        string              `json:"value,omitempty"`
	IngestionURL string              `json:"ingestion_url,omitempty"`
	Tags         []string            `json:"tags,omitempty"`
	CreatedAt    string              `json:"created_at,omitempty"`
	UpdatedAt    string              `json:"updated_at,omitempty"`
	ExpiresAt    string              `json:"expires_at,omitempty"`
	RateLimits   *IngestionRateLimit `json:"rate_limits,omitempty"`
}

// IngestionRateLimit model.
type IngestionRateLimit struct {
	DayLimit    int64 `json:"day_limit"`
	SecondLimit int64 `json:"second_limit"`
}

func (k IngestionKey) TagsToTerraform() (types.List, diag.Diagnostics) {
	tags := utils.Map(k.Tags, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, tags)
}

func (k IngestionKey) RateLimitsToTerraform() (types.Object, diag.Diagnostics) {
	rateLimitsAttrTypes := attr.RateLimitsAttrTypes()

	if k.RateLimits == nil {
		return types.ObjectNull(rateLimitsAttrTypes), nil
	}

	return types.ObjectValue(
		rateLimitsAttrTypes,
		map[string]tfattr.Value{
			attr.DayLimit:    types.Int64Value(k.RateLimits.DayLimit),
			attr.SecondLimit: types.Int64Value(k.RateLimits.SecondLimit),
		},
	)
}

func (k *IngestionKey) SetTags(ctx context.Context, tfTags types.List) {
	if utils.IsNullOrUnknown(tfTags) {
		return
	}

	tags := utils.Map(tfTags.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	k.Tags = tags
}

func (k *IngestionKey) SetRateLimits(ctx context.Context, tfRateLimits types.Object) {
	if utils.IsNullOrUnknown(tfRateLimits) {
		return
	}

	attrs := tfRateLimits.Attributes()

	rateLimits := &IngestionRateLimit{}

	if v, ok := attrs[attr.DayLimit]; ok {
		if i, ok2 := v.(types.Int64); ok2 && !utils.IsNullOrUnknown(i) {
			rateLimits.DayLimit = i.ValueInt64()
		}
	}

	if v, ok := attrs[attr.SecondLimit]; ok {
		if i, ok2 := v.(types.Int64); ok2 && !utils.IsNullOrUnknown(i) {
			rateLimits.SecondLimit = i.ValueInt64()
		}
	}

	k.RateLimits = rateLimits
}
