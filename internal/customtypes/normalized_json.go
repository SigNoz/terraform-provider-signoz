package customtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// NormalizedJSON is a string attribute holding a JSON document whose semantic equality ignores
// differences that the SigNoz API introduces on read-back: key ordering, whitespace, and — crucially —
// keys the server adds with empty/null default values (e.g. `spec.source = ""`, `recoveryTarget = null`).
//
// Without this, `condition`/`evaluation` show a perpetual in-place update on every plan: the config
// carries the user's JSON, but a refresh stores the server-enriched JSON, and a plain string compare
// treats them as different. Semantic equality lets Terraform recognize the two as the same document.
//
// Equality is intentionally conservative: it only collapses null / empty-string / empty-object /
// empty-array entries. Any real value change (a different number, a changed expression, a value set to
// or cleared from a non-empty value) still produces a diff, so genuine changes are never masked.

var _ basetypes.StringTypable = NormalizedJSONType{}

type NormalizedJSONType struct {
	basetypes.StringType
}

func (t NormalizedJSONType) Equal(o attr.Type) bool {
	other, ok := o.(NormalizedJSONType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t NormalizedJSONType) String() string {
	return "customtypes.NormalizedJSONType"
}

func (t NormalizedJSONType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return NormalizedJSON{StringValue: in}, nil
}

func (t NormalizedJSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T for NormalizedJSONType", attrValue)
	}

	return NormalizedJSON{StringValue: stringValue}, nil
}

func (t NormalizedJSONType) ValueType(_ context.Context) attr.Value {
	return NormalizedJSON{}
}

var (
	_ basetypes.StringValuable                   = NormalizedJSON{}
	_ basetypes.StringValuableWithSemanticEquals = NormalizedJSON{}
)

type NormalizedJSON struct {
	basetypes.StringValue
}

func (v NormalizedJSON) Type(_ context.Context) attr.Type {
	return NormalizedJSONType{}
}

func (v NormalizedJSON) Equal(o attr.Value) bool {
	other, ok := o.(NormalizedJSON)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals reports whether two JSON documents are equal after normalizing away key
// ordering, whitespace, and empty/null entries. If either side is not valid JSON, it falls back to a
// plain string comparison so nothing is silently treated as equal.
func (v NormalizedJSON) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(NormalizedJSON)
	if !ok {
		return false, diags
	}

	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.StringValue.Equal(newValue.StringValue), diags
	}

	left, errLeft := normalizeJSON(v.ValueString())
	right, errRight := normalizeJSON(newValue.ValueString())
	if errLeft != nil || errRight != nil {
		return v.ValueString() == newValue.ValueString(), diags
	}

	return reflect.DeepEqual(left, right), diags
}

func NewNormalizedJSONValue(value string) NormalizedJSON {
	return NormalizedJSON{StringValue: basetypes.NewStringValue(value)}
}

func NewNormalizedJSONNull() NormalizedJSON {
	return NormalizedJSON{StringValue: basetypes.NewStringNull()}
}

func NewNormalizedJSONUnknown() NormalizedJSON {
	return NormalizedJSON{StringValue: basetypes.NewStringUnknown()}
}

// normalizeJSON parses a JSON string and recursively drops empty entries so server-added default keys
// do not register as drift.
func normalizeJSON(s string) (interface{}, error) {
	var parsed interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return nil, err
	}

	return stripEmpty(parsed), nil
}

// stripEmpty recursively removes null / empty-string / empty-object / empty-array values from maps.
// Slice elements are normalized but never dropped, so array length (which is meaningful) is preserved.
func stripEmpty(v interface{}) interface{} {
	switch value := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(value))
		for k, elem := range value {
			normalized := stripEmpty(elem)
			if isEmpty(normalized) {
				continue
			}
			out[k] = normalized
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(value))
		for i, elem := range value {
			out[i] = stripEmpty(elem)
		}
		return out
	default:
		return value
	}
}

// isEmpty reports whether a value equals its JSON zero value. The SigNoz API fills omitted condition
// fields with their zero value on read-back (e.g. `disabled: false`, `step: 0`, `source: ""`,
// `recoveryTarget: null`), so treating zero values as absent is what lets an omitted field and a
// server-defaulted field compare equal. A change to a non-zero value is still a real diff, because the
// other side carries the previous non-zero value.
func isEmpty(v interface{}) bool {
	switch value := v.(type) {
	case nil:
		return true
	case string:
		return value == ""
	case bool:
		return !value
	case float64:
		return value == 0
	case map[string]interface{}:
		return len(value) == 0
	case []interface{}:
		return len(value) == 0
	default:
		return false
	}
}
