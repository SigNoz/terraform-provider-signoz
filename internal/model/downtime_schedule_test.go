package model

import (
	"context"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDowntimeSchedule_ScheduleRoundTrip(t *testing.T) {
	ctx := context.Background()

	ds := &DowntimeSchedule{
		Schedule: DowntimeScheduleConfig{
			StartTime: "2024-01-15T00:00:00Z",
			EndTime:   "2024-01-15T06:00:00Z",
			Timezone:  "America/New_York",
			Recurrence: &DowntimeRecurrence{
				Type:       "weekly",
				DaysOfWeek: []int{1, 3, 5},
				Duration:   "6h",
			},
		},
	}

	tfObj, diags := ds.ScheduleToTerraform(ctx)
	if diags.HasError() {
		t.Fatalf("ScheduleToTerraform had errors: %v", diags)
	}

	if tfObj.IsNull() {
		t.Fatal("expected non-null schedule object")
	}

	ds2 := &DowntimeSchedule{}
	err := ds2.SetSchedule(ctx, tfObj)
	if err != nil {
		t.Fatalf("SetSchedule failed: %v", err)
	}

	if ds2.Schedule.StartTime != "2024-01-15T00:00:00Z" {
		t.Errorf("StartTime mismatch: %s", ds2.Schedule.StartTime)
	}
	if ds2.Schedule.EndTime != "2024-01-15T06:00:00Z" {
		t.Errorf("EndTime mismatch: %s", ds2.Schedule.EndTime)
	}
	if ds2.Schedule.Timezone != "America/New_York" {
		t.Errorf("Timezone mismatch: %s", ds2.Schedule.Timezone)
	}
	if ds2.Schedule.Recurrence == nil {
		t.Fatal("expected non-nil recurrence")
	}
	if ds2.Schedule.Recurrence.Type != "weekly" {
		t.Errorf("Recurrence.Type mismatch: %s", ds2.Schedule.Recurrence.Type)
	}
	if len(ds2.Schedule.Recurrence.DaysOfWeek) != 3 {
		t.Fatalf("expected 3 days, got %d", len(ds2.Schedule.Recurrence.DaysOfWeek))
	}
	for i, expected := range []int{1, 3, 5} {
		if ds2.Schedule.Recurrence.DaysOfWeek[i] != expected {
			t.Errorf("DaysOfWeek[%d]: expected %d, got %d", i, expected, ds2.Schedule.Recurrence.DaysOfWeek[i])
		}
	}
	if ds2.Schedule.Recurrence.Duration != "6h" {
		t.Errorf("Duration mismatch: %s", ds2.Schedule.Recurrence.Duration)
	}
}

func TestDowntimeSchedule_ScheduleNoRecurrence(t *testing.T) {
	ctx := context.Background()

	ds := &DowntimeSchedule{
		Schedule: DowntimeScheduleConfig{
			StartTime:  "2024-01-15T00:00:00Z",
			EndTime:    "2024-01-15T06:00:00Z",
			Timezone:   "UTC",
			Recurrence: nil,
		},
	}

	tfObj, diags := ds.ScheduleToTerraform(ctx)
	if diags.HasError() {
		t.Fatalf("ScheduleToTerraform had errors: %v", diags)
	}

	// Verify recurrence attribute is null
	attrs := tfObj.Attributes()
	recVal := attrs[attr.Recurrence]
	if !recVal.IsNull() {
		t.Error("expected null recurrence when nil")
	}

	ds2 := &DowntimeSchedule{}
	err := ds2.SetSchedule(ctx, tfObj)
	if err != nil {
		t.Fatalf("SetSchedule failed: %v", err)
	}

	if ds2.Schedule.Recurrence != nil {
		t.Error("expected nil recurrence")
	}
	if ds2.Schedule.StartTime != "2024-01-15T00:00:00Z" {
		t.Errorf("StartTime mismatch: %s", ds2.Schedule.StartTime)
	}
}

func TestDowntimeSchedule_ScheduleNullInput(t *testing.T) {
	ctx := context.Background()

	scheduleAttrTypes := ScheduleAttrTypes()
	nullObj := types.ObjectNull(scheduleAttrTypes)

	ds := &DowntimeSchedule{}
	err := ds.SetSchedule(ctx, nullObj)
	if err != nil {
		t.Fatalf("SetSchedule failed on null: %v", err)
	}
}

func TestDowntimeSchedule_RecurrenceNoDaysOfWeek(t *testing.T) {
	ctx := context.Background()

	ds := &DowntimeSchedule{
		Schedule: DowntimeScheduleConfig{
			StartTime: "2024-06-01T08:00:00Z",
			EndTime:   "2024-06-01T12:00:00Z",
			Timezone:  "Europe/London",
			Recurrence: &DowntimeRecurrence{
				Type:       "daily",
				DaysOfWeek: nil,
				Duration:   "4h",
			},
		},
	}

	tfObj, diags := ds.ScheduleToTerraform(ctx)
	if diags.HasError() {
		t.Fatalf("ScheduleToTerraform had errors: %v", diags)
	}

	ds2 := &DowntimeSchedule{}
	err := ds2.SetSchedule(ctx, tfObj)
	if err != nil {
		t.Fatalf("SetSchedule failed: %v", err)
	}

	if ds2.Schedule.Recurrence == nil {
		t.Fatal("expected non-nil recurrence")
	}
	if ds2.Schedule.Recurrence.Type != "daily" {
		t.Errorf("expected Type=daily, got %s", ds2.Schedule.Recurrence.Type)
	}
}

func TestDowntimeSchedule_AlertIDsRoundTrip(t *testing.T) {
	ds := &DowntimeSchedule{
		AlertIDs: []string{"alert-1", "alert-2", "alert-3"},
	}

	tfList, diags := ds.AlertIDsToTerraform()
	if diags.HasError() {
		t.Fatalf("AlertIDsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 alert IDs, got %d", len(tfList.Elements()))
	}

	ds2 := &DowntimeSchedule{}
	ds2.SetAlertIDs(tfList)

	if len(ds2.AlertIDs) != 3 {
		t.Fatalf("expected 3 alert IDs, got %d", len(ds2.AlertIDs))
	}

	for i, id := range []string{"alert-1", "alert-2", "alert-3"} {
		if ds2.AlertIDs[i] != id {
			t.Errorf("alertID[%d]: expected %s, got %s", i, id, ds2.AlertIDs[i])
		}
	}
}

func TestDowntimeSchedule_AlertIDsNil(t *testing.T) {
	ds := &DowntimeSchedule{AlertIDs: nil}

	tfList, diags := ds.AlertIDsToTerraform()
	if diags.HasError() {
		t.Fatalf("AlertIDsToTerraform had errors: %v", diags)
	}

	// nil should produce empty list (not null)
	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 elements for nil AlertIDs, got %d", len(tfList.Elements()))
	}
}

func TestDowntimeSchedule_SetAlertIDsNull(t *testing.T) {
	nullList := types.ListNull(types.StringType)

	ds := &DowntimeSchedule{AlertIDs: []string{"existing"}}
	ds.SetAlertIDs(nullList)

	if ds.AlertIDs != nil {
		t.Errorf("expected nil AlertIDs for null input, got %v", ds.AlertIDs)
	}
}

func TestDowntimeSchedule_SetAlertIDsFromConstructed(t *testing.T) {
	tfList, _ := types.ListValue(types.StringType, []tfattr.Value{
		types.StringValue("a-1"),
		types.StringValue("a-2"),
	})

	ds := &DowntimeSchedule{}
	ds.SetAlertIDs(tfList)

	if len(ds.AlertIDs) != 2 || ds.AlertIDs[0] != "a-1" || ds.AlertIDs[1] != "a-2" {
		t.Errorf("unexpected AlertIDs: %v", ds.AlertIDs)
	}
}

func TestScheduleAttrTypes_Structure(t *testing.T) {
	sat := ScheduleAttrTypes()
	if len(sat) != 4 {
		t.Errorf("expected 4 schedule attr types, got %d", len(sat))
	}

	rat := RecurrenceAttrTypes()
	if len(rat) != 3 {
		t.Errorf("expected 3 recurrence attr types, got %d", len(rat))
	}
}
