package attr

import (
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	KeyName   = "name"
	KeyValue  = "value"
	KeyTags   = "tags"
	ExpiresAt = "expires_at"
	KeyID     = "key_id"
	Signal    = "signal"
	Config    = "config"
	Day       = "day"
	Second    = "second"
	Count     = "count"
	Size      = "size"
	Limits    = "limits"
	Metric    = "metric"
	Workspace = "workspace_id"
)

// LimitValueAttrTypes is the {size, count} object type shared by the config
// and metric arms of a limit.
func LimitValueAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		Size:  types.Int64Type,
		Count: types.Int64Type,
	}
}

// LimitConfigAttrTypes is the {day, second} object type of a limit config.
func LimitConfigAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		Day:    types.ObjectType{AttrTypes: LimitValueAttrTypes()},
		Second: types.ObjectType{AttrTypes: LimitValueAttrTypes()},
	}
}

// LimitMetricAttrTypes is the {day, second} object type of a limit metric
// snapshot (non-nullable count/size).
func LimitMetricAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		Day:    types.ObjectType{AttrTypes: map[string]tfattr.Type{Count: types.Int64Type, Size: types.Int64Type}},
		Second: types.ObjectType{AttrTypes: map[string]tfattr.Type{Count: types.Int64Type, Size: types.Int64Type}},
	}
}
