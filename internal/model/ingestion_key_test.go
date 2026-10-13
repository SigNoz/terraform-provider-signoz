package model

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNoExpiry(t *testing.T) {
	cases := map[string]struct {
		input time.Time
		want  bool
	}{
		"literal zero":                 {input: time.Time{}, want: true},
		"reconstructed zero":           {input: time.UnixMilli(time.Time{}.UnixMilli()), want: true},
		"unix millisecond zero (1970)": {input: time.UnixMilli(0), want: false},
		"real date":                    {input: time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC), want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := NoExpiry(tc.input); got != tc.want {
				t.Fatalf("NoExpiry(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestExpiresAtSentinelRoundTrip(t *testing.T) {
	// Null attribute -> zero time on the wire.
	wire, diags := ExpiresAtToAPI(types.StringNull())
	if diags.HasError() || !wire.IsZero() {
		t.Fatalf("null -> wire: %v %v", wire, diags)
	}

	// Zero time on the wire -> null attribute (no diff against an omitted expires_at).
	if got := ExpiresAtToTerraform(time.Time{}); !got.IsNull() {
		t.Fatalf("zero wire -> terraform: %v", got)
	}

	// A reconstructed zero (what the server round-trips) also maps to null.
	reconstructed := time.Time{}.UTC()
	if got := ExpiresAtToTerraform(reconstructed); !got.IsNull() {
		t.Fatalf("reconstructed zero -> terraform: %v", got)
	}

	// A real expiry round-trips through the same instant.
	expiry := time.Date(2027, 3, 1, 12, 0, 0, 0, time.UTC)

	wire, diags = ExpiresAtToAPI(types.StringValue("2027-03-01T12:00:00Z"))
	if diags.HasError() || !wire.Equal(expiry) {
		t.Fatalf("value -> wire: %v %v", wire, diags)
	}

	attr := ExpiresAtToTerraform(wire)
	if attr.ValueString() != "2027-03-01T12:00:00Z" {
		t.Fatalf("wire -> value: %v", attr)
	}
}

func TestExpiresAtToAPIRejectsNonRFC3339(t *testing.T) {
	for _, bad := range []string{"2027-03-01", "not-a-time", "2027-03-01T12:00:00 +0000 UTC"} {
		_, diags := ExpiresAtToAPI(types.StringValue(bad))
		if !diags.HasError() {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestIngestionKeyTagsRoundTrip(t *testing.T) {
	// nil wire tags stay null so a key created without tags does not diff.
	key := &IngestionKey{}

	list, diags := key.TagsToTerraform()
	if diags.HasError() || !list.IsNull() {
		t.Fatalf("nil tags -> %v %v", list, diags)
	}

	tags, diags := TagsFromTerraform(list)
	if diags.HasError() || tags != nil {
		t.Fatalf("null list -> %v %v", tags, diags)
	}

	// A populated list round-trips.
	key = &IngestionKey{Tags: []string{"env:prod", "team:platform"}}

	list, diags = key.TagsToTerraform()
	if diags.HasError() || list.IsNull() || len(list.Elements()) != 2 {
		t.Fatalf("tags -> %v %v", list, diags)
	}

	tags, diags = TagsFromTerraform(list)
	if diags.HasError() || len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "team:platform" {
		t.Fatalf("round trip: %v %v", tags, diags)
	}

	// Empty (non-nil) stays an empty list, not null.
	key = &IngestionKey{Tags: []string{}}

	list, diags = key.TagsToTerraform()
	if diags.HasError() || list.IsNull() || len(list.Elements()) != 0 {
		t.Fatalf("empty tags -> %v %v", list, diags)
	}
}
