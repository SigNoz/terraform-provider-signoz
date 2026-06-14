package convtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringPointer returns nil for null/unknown framework values, else a
// pointer to the underlying string. Use for optional API fields where
// omitting from the JSON body is semantically distinct from sending an
// empty string.
func StringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// StringFromPointer is the inverse of StringPointer: nil pointer becomes
// a null framework value, non-nil becomes a known string.
func StringFromPointer(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// ---------------------------------------------------------------------------
// String slices (pointer-wrapped)
// ---------------------------------------------------------------------------
//
// oapi-codegen emits optional array fields as `*[]E` so the JSON marshal
// can distinguish "omitted" from "empty". The framework List has a
// matching three-state encoding (null / unknown / known) so the
// conversions are direct.

// StringPointerSliceFromList extracts a *[]string from a framework
// List<string>. Null/unknown → nil pointer; known empty → pointer to
// empty slice; known with items → pointer to populated slice.
func StringPointerSliceFromList(ctx context.Context, l types.List) (*[]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	out := make([]string, 0, len(l.Elements()))
	diags := l.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil, diags
	}
	return &out, diags
}

// ListFromStringPointerSlice is the inverse of StringPointerSliceFromList.
func ListFromStringPointerSlice(_ context.Context, p *[]string) (types.List, diag.Diagnostics) {
	if p == nil {
		return types.ListNull(types.StringType), nil
	}
	elems := make([]attr.Value, 0, len(*p))
	for _, s := range *p {
		elems = append(elems, types.StringValue(s))
	}
	return types.ListValue(types.StringType, elems)
}

// TypedStringPointerSliceFromList is StringPointerSliceFromList for any
// `~string` enum type. Used for `*[]apitypes.RuletypesRepeatOn` and
// friends — the wire type is a typed string slice but the framework
// stores plain strings.
func TypedStringPointerSliceFromList[E ~string](ctx context.Context, l types.List) (*[]E, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	var raw []string
	diags := l.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]E, 0, len(raw))
	for _, s := range raw {
		out = append(out, E(s))
	}
	return &out, diags
}

// ListFromTypedStringPointerSlice is the inverse of TypedStringPointerSliceFromList.
func ListFromTypedStringPointerSlice[E ~string](_ context.Context, p *[]E) (types.List, diag.Diagnostics) {
	if p == nil {
		return types.ListNull(types.StringType), nil
	}
	elems := make([]attr.Value, 0, len(*p))
	for _, e := range *p {
		elems = append(elems, types.StringValue(string(e)))
	}
	return types.ListValue(types.StringType, elems)
}

// StringMapPointerFromMap extracts a *map[string]string from a framework
// Map<string>. Same null/empty/known semantics as the slice helpers.
func StringMapPointerFromMap(ctx context.Context, m types.Map) (*map[string]string, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	out := make(map[string]string, len(m.Elements()))
	diags := m.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil, diags
	}
	return &out, diags
}

// MapFromStringPointerMap is the inverse of StringMapPointerFromMap.
func MapFromStringPointerMap(_ context.Context, p *map[string]string) (types.Map, diag.Diagnostics) {
	if p == nil {
		return types.MapNull(types.StringType), nil
	}
	elems := make(map[string]attr.Value, len(*p))
	for k, v := range *p {
		elems[k] = types.StringValue(v)
	}
	return types.MapValue(types.StringType, elems)
}

// InterfaceMapPointerFromMap projects a framework `Map<string>` onto a
// `*map[string]interface{}`. Each map value is stored as the bare
// string in the interface — that's what the server expects for the
// "untyped object property bag" pattern (jira `custom_fields`,
// pagerduty `details`, oauth2 `claims`). Lossy if the user wants
// nested objects, but that surface isn't expressible through tfsdk
// schemas anyway.
func InterfaceMapPointerFromMap(ctx context.Context, m types.Map) (*map[string]interface{}, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	raw := make(map[string]string, len(m.Elements()))
	diags := m.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, diags
	}
	out := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		out[k] = v
	}
	return &out, diags
}

// MapFromInterfaceMapPointer is the inverse. Each interface{} value is
// projected onto a string: `string` passes through; `nil` becomes the
// null framework value; everything else is `fmt.Sprintf("%v", x)` so
// the framework state stays valid even when the server emits numbers
// or booleans inside the bag.
func MapFromInterfaceMapPointer(_ context.Context, p *map[string]interface{}) (types.Map, diag.Diagnostics) {
	if p == nil {
		return types.MapNull(types.StringType), nil
	}
	elems := make(map[string]attr.Value, len(*p))
	for k, v := range *p {
		switch x := v.(type) {
		case string:
			elems[k] = types.StringValue(x)
		case nil:
			elems[k] = types.StringNull()
		default:
			elems[k] = types.StringValue(fmt.Sprintf("%v", x))
		}
	}
	return types.MapValue(types.StringType, elems)
}
