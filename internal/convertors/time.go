// Helpers used by the codegen-generated Expand/Flatten functions in
// `conv/zz_generated_*.go` and the hand-written Layer 3 entry points.
//
// Only primitive ↔ wire conversions live here. Anything that touches a
// specific customtype belongs in the per-customtype file.
package conv

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Primitive scalars
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Time
// ---------------------------------------------------------------------------
//
// SigNoz's wire format is RFC3339 (`format: date-time` in the OpenAPI
// spec). Framework state holds these as strings — we round-trip through
// time.Time using RFC3339Nano so any sub-second precision the server
// sends survives.

// TimeFromString parses a required RFC3339 string into a time.Time. A
// null/unknown framework value is treated as a hard error — required
// fields must be set.
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

// TimePointerFromString parses an optional RFC3339 string into a
// *time.Time. Null/unknown → nil pointer.
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

// TimeStringFromValue formats a non-pointer time.Time. Always emits a
// known string — the wire shape declared the field as required.
func TimeStringFromValue(t time.Time) types.String {
	return types.StringValue(t.Format(time.RFC3339Nano))
}

// TimeStringFromPointer formats a *time.Time. nil → null framework value.
func TimeStringFromPointer(p *time.Time) types.String {
	if p == nil {
		return types.StringNull()
	}
	return TimeStringFromValue(*p)
}
