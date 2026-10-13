package model

import (
	"fmt"
	"time"

	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Signal values accepted by the gateway limit API
// (pkg/types/gatewaytypes in the SigNoz server).
const (
	SignalLogs    = "logs"
	SignalTraces  = "traces"
	SignalMetrics = "metrics"
)

// IngestionKey mirrors the server's gatewaytypes.IngestionKey response body
// (GET /api/v2/gateway/ingestion_keys/{keyId}).
type IngestionKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Value       string    `json:"value,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	Limits      []Limit   `json:"limits,omitempty"`
}

// CreatedIngestionKey mirrors gatewaytypes.GettableCreatedIngestionKey
// (POST /api/v2/gateway/ingestion_keys). Value is only ever returned here.
type CreatedIngestionKey struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// CreatedIngestionKeyLimit mirrors gatewaytypes.GettableCreatedIngestionKeyLimit
// (POST /api/v2/gateway/ingestion_limits).
type CreatedIngestionKeyLimit struct {
	ID string `json:"id"`
}

// PostableIngestionKey mirrors gatewaytypes.PostableIngestionKey — the shared
// body of key create and key update (PATCH). A zero ExpiresAt means no expiry.
type PostableIngestionKey struct {
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Limit mirrors the server's gatewaytypes.Limit response body.
type Limit struct {
	ID        string      `json:"id"`
	Signal    string      `json:"signal"`
	Config    LimitConfig `json:"config"`
	Tags      []string    `json:"tags"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	KeyID     string      `json:"key_id"`
	Metric    LimitMetric `json:"metric"`
}

// LimitConfig mirrors gatewaytypes.LimitConfig. nil Day/Second mean unbounded.
type LimitConfig struct {
	Day    *LimitValue `json:"day,omitempty"`
	Second *LimitValue `json:"second,omitempty"`
}

// LimitValue mirrors gatewaytypes.LimitValue. nil Size/Count mean unbounded
// for that dimension; Count is requests, Size is bytes.
type LimitValue struct {
	Size  *int64 `json:"size,omitempty"`
	Count *int64 `json:"count,omitempty"`
}

// LimitMetric mirrors gatewaytypes.LimitMetric — the server-computed
// enforcement snapshot of Config (nil arms deref to zero).
type LimitMetric struct {
	Day    *LimitMetricValue `json:"day,omitempty"`
	Second *LimitMetricValue `json:"second,omitempty"`
}

// LimitMetricValue mirrors gatewaytypes.LimitMetricValue.
type LimitMetricValue struct {
	Count int64 `json:"count"`
	Size  int64 `json:"size"`
}

// PostableIngestionKeyLimit mirrors gatewaytypes.PostableIngestionKeyLimit
// (POST /api/v2/gateway/ingestion_limits — note the camelCase keyId).
type PostableIngestionKeyLimit struct {
	KeyID  string      `json:"keyId"`
	Signal string      `json:"signal"`
	Config LimitConfig `json:"config"`
	Tags   []string    `json:"tags"`
}

// UpdatableIngestionKeyLimit mirrors gatewaytypes.UpdatableIngestionKeyLimit
// (PATCH /api/v2/gateway/ingestion_limits/{limitId} — config and tags only;
// keyId and signal are immutable).
type UpdatableIngestionKeyLimit struct {
	Config LimitConfig `json:"config"`
	Tags   []string    `json:"tags"`
}

// NoExpiry reports whether the API's expires_at means "never expires". The
// server stores zero time for keys without expiry and round-trips the same
// instant, but the reconstructed time is not guaranteed to be the literal
// zero Time, so both checks are needed.
func NoExpiry(t time.Time) bool {
	return t.IsZero() || t.Equal(time.Time{}.UTC())
}

// ExpiresAtToAPI converts the TF expires_at attribute to the wire value.
// A null/unknown attribute becomes the zero-time sentinel.
func ExpiresAtToAPI(tfExpiresAt types.String) (time.Time, diag.Diagnostics) {
	if tfExpiresAt.IsNull() || tfExpiresAt.IsUnknown() {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, tfExpiresAt.ValueString())
	if err != nil {
		return time.Time{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid expires_at",
				fmt.Sprintf("expires_at must be an RFC 3339 UTC timestamp: %s", err),
			),
		}
	}

	return parsed, nil
}

// ExpiresAtFromTerraform converts a plain expires_at string to the wire value.
func ExpiresAtFromTerraform(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.RFC3339Nano, value)
}

// ExpiresAtToTerraform converts the wire value back to the TF attribute.
// The zero-time sentinel maps to null so an unexpiring key never diffs.
func ExpiresAtToTerraform(t time.Time) types.String {
	if NoExpiry(t) {
		return types.StringNull()
	}

	return types.StringValue(t.Format(time.RFC3339Nano))
}

// TagsToTerraform converts the wire tags to the TF list. A missing (nil) tag
// set stays null so a key created without tags does not diff.
func (k *IngestionKey) TagsToTerraform() (types.List, diag.Diagnostics) {
	if k.Tags == nil {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValue(types.StringType, utils.Map(k.Tags, func(value string) tfattr.Value {
		return types.StringValue(value)
	}))
}

// TagsFromTerraform converts the TF list to wire tags. Null and unknown both
// become nil so the server stores "no tags".
func TagsFromTerraform(list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	tags := make([]string, 0, len(list.Elements()))

	for i, element := range list.Elements() {
		value, ok := element.(types.String)
		if !ok {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid tags",
					fmt.Sprintf("tag at index %d is not a string", i),
				),
			}
		}

		tags = append(tags, value.ValueString())
	}

	return tags, nil
}
