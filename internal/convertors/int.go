package conv

import "github.com/hashicorp/terraform-plugin-framework/types"

// Int64Pointer mirrors StringPointer for int64.
func Int64Pointer(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	n := v.ValueInt64()
	return &n
}

// Int64FromPointer mirrors StringFromPointer for int64.
func Int64FromPointer(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

// IntPointer is Int64Pointer for the platform-default `int`. Used for
// optional integer fields where oapi-codegen emits `*int` (the Go
// default for `integer` schema with no explicit format).
func IntPointer(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	n := int(v.ValueInt64())
	return &n
}

// IntFromPointer is the inverse of IntPointer.
func IntFromPointer(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
