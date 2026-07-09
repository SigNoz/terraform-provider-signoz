package customtypes

import (
	"context"
	"testing"
)

func TestNormalizedJSONStringSemanticEquals(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		a    string
		b    string
		want bool
	}{
		"identical":                        {`{"a":1}`, `{"a":1}`, true},
		"key order and whitespace":         {`{"a":1,"b":2}`, "{ \"b\": 2,\n \"a\": 1 }", true},
		"server adds empty string key":     {`{"a":1}`, `{"a":1,"source":""}`, true},
		"server adds null key":             {`{"a":1}`, `{"a":1,"recoveryTarget":null}`, true},
		"server adds empty object/array":   {`{"a":1}`, `{"a":1,"x":{},"y":[]}`, true},
		"nested server-added empties":      {`{"spec":{"name":"A"}}`, `{"spec":{"name":"A","source":"","f":null}}`, true},
		"array element enriched":           {`{"q":[{"name":"A"}]}`, `{"q":[{"name":"A","source":""}]}`, true},
		"real value change not equal":      {`{"target":10}`, `{"target":12}`, false},
		"value cleared to empty not equal": {`{"foo":"bar"}`, `{"foo":""}`, false},
		"array length change not equal":    {`{"q":[1]}`, `{"q":[1,2]}`, false},
		"array element order not equal":    {`{"q":[1,2]}`, `{"q":[2,1]}`, false},
		"different key not equal":          {`{"a":1}`, `{"b":1}`, false},
		"zero-number default dropped":      {`{"target":0}`, `{}`, true},
		"false default dropped":            {`{"flag":false}`, `{}`, true},
		"server adds false/zero defaults":  {`{"spec":{"name":"A"}}`, `{"spec":{"name":"A","disabled":false,"stats":false,"step":0}}`, true},
		"change to a nonzero value diffs":  {`{"target":0}`, `{"target":5}`, false},
		"nonzero to zero diffs":            {`{"target":5}`, `{"target":0}`, false},
		"invalid json falls back equal":    {`{`, `{`, true},
		"invalid json falls back diff":     {`{`, `}`, false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, diags := NewNormalizedJSONValue(tc.a).StringSemanticEquals(context.Background(), NewNormalizedJSONValue(tc.b))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != tc.want {
				t.Fatalf("StringSemanticEquals(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestNormalizedJSONStringSemanticEqualsNullUnknown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	equal, _ := NewNormalizedJSONNull().StringSemanticEquals(ctx, NewNormalizedJSONNull())
	if !equal {
		t.Fatal("null should equal null")
	}

	equal, _ = NewNormalizedJSONNull().StringSemanticEquals(ctx, NewNormalizedJSONValue(`{"a":1}`))
	if equal {
		t.Fatal("null should not equal a value")
	}

	equal, _ = NewNormalizedJSONUnknown().StringSemanticEquals(ctx, NewNormalizedJSONUnknown())
	if !equal {
		t.Fatal("unknown should equal unknown")
	}
}
