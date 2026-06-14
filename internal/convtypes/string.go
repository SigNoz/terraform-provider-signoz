package convtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringPointer maps null/unknown to a nil pointer so the field is omitted from
// the JSON body — distinct from sending an empty string.
func StringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func StringFromPointer(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// StringPointerSliceFromList maps a null/unknown list to a nil pointer (field
// omitted); a known list — even empty — maps to a non-nil slice.
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

// TypedStringPointerSliceFromList handles wire types that are a typed-string
// slice (e.g. *[]apitypes.RuletypesRepeatOn) while the framework stores plain
// strings.
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

// InterfaceMapPointerFromMap projects a framework Map<string> onto a
// *map[string]interface{}, storing each value as a bare string. This is the
// "untyped object property bag" the server expects (jira custom_fields,
// pagerduty details, oauth2 claims); nested objects aren't expressible through
// the tfsdk schema, so it's intentionally string-only.
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

// MapFromInterfaceMapPointer coerces each value back to a string so framework
// state stays valid when the server emits numbers or booleans in the bag: nil
// becomes null, strings pass through, everything else uses fmt.Sprintf.
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
