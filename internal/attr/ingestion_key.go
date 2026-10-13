package attr

import (
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	KeyName      = "name"
	KeyValue     = "value"
	IngestionURL = "ingestion_url"
	RateLimits   = "rate_limits"
	DayLimit     = "day_limit"
	SecondLimit  = "second_limit"
	ExpiresAt    = "expires_at"
	KeyTags      = "tags"
)

func RateLimitsAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		DayLimit:    types.Int64Type,
		SecondLimit: types.Int64Type,
	}
}
