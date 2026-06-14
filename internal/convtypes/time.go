package convtypes

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TimeFromString parses a required RFC3339 timestamp; a null/unknown value is
// an error.
func TimeFromString(v types.String) (time.Time, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || v.IsUnknown() {
		diags.AddError("Time required", "expected a non-null RFC3339 timestamp")
		return time.Time{}, diags
	}
	t, err := time.Parse(time.RFC3339, v.ValueString())
	if err != nil {
		diags.AddError("Invalid RFC3339 timestamp", fmt.Sprintf("could not parse %q: %s", v.ValueString(), err))
	}
	return t, diags
}

// TimePointerFromString is the optional counterpart to TimeFromString:
// null/unknown yields a nil pointer instead of an error.
func TimePointerFromString(v types.String) (*time.Time, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	t, diags := TimeFromString(v)
	if diags.HasError() {
		return nil, diags
	}
	return &t, diags
}

// TimeStringFromValue formats to RFC3339Nano; the field is required so the
// result is always a known string.
func TimeStringFromValue(t time.Time) types.String {
	return types.StringValue(t.Format(time.RFC3339Nano))
}

func TimeStringFromPointer(p *time.Time) types.String {
	if p == nil {
		return types.StringNull()
	}
	return TimeStringFromValue(*p)
}
