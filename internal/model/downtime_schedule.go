package model

import (
	"context"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DowntimeSchedule model.
type DowntimeSchedule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	AlertIDs    []string               `json:"alert_ids,omitempty"`
	AllAlerts   bool                   `json:"all_alerts"`
	Schedule    DowntimeScheduleConfig `json:"schedule"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
}

// DowntimeScheduleConfig is the schedule configuration.
type DowntimeScheduleConfig struct {
	StartTime  string              `json:"start_time"`
	EndTime    string              `json:"end_time"`
	Timezone   string              `json:"timezone"`
	Recurrence *DowntimeRecurrence `json:"recurrence,omitempty"`
}

// DowntimeRecurrence describes recurrence rules.
type DowntimeRecurrence struct {
	Type       string `json:"type"`
	DaysOfWeek []int  `json:"days_of_week,omitempty"`
	Duration   string `json:"duration,omitempty"`
}

// RecurrenceTypes lists the allowed recurrence type values.
//
//nolint:gochecknoglobals
var RecurrenceTypes = []string{"once", "daily", "weekly"}

// RecurrenceAttrTypes returns the Terraform attribute types for the recurrence object.
func RecurrenceAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.RecurrenceType: types.StringType,
		attr.DaysOfWeek:     types.ListType{ElemType: types.Int64Type},
		attr.Duration:       types.StringType,
	}
}

// ScheduleAttrTypes returns the Terraform attribute types for the schedule object.
func ScheduleAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.StartTime:  types.StringType,
		attr.EndTime:    types.StringType,
		attr.Timezone:   types.StringType,
		attr.Recurrence: types.ObjectType{AttrTypes: RecurrenceAttrTypes()},
	}
}

// AlertIDsToTerraform converts alert IDs to a Terraform list.
func (d DowntimeSchedule) AlertIDsToTerraform() (types.List, diag.Diagnostics) {
	if d.AlertIDs == nil {
		return types.ListValueMust(types.StringType, []tfattr.Value{}), nil
	}

	alertIDs := utils.Map(d.AlertIDs, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, alertIDs)
}

// ScheduleToTerraform converts the schedule config to a Terraform object.
func (d DowntimeSchedule) ScheduleToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	scheduleAttrTypes := ScheduleAttrTypes()
	recurrenceAttrTypes := RecurrenceAttrTypes()

	var recurrenceObj types.Object
	if d.Schedule.Recurrence != nil {
		rec := d.Schedule.Recurrence

		var daysOfWeekList types.List
		if rec.DaysOfWeek == nil {
			daysOfWeekList = types.ListValueMust(types.Int64Type, []tfattr.Value{})
		} else {
			var diagsDow diag.Diagnostics
			daysOfWeekList, diagsDow = types.ListValueFrom(ctx, types.Int64Type, rec.DaysOfWeek)
			diags.Append(diagsDow...)
			if diags.HasError() {
				return types.ObjectNull(scheduleAttrTypes), diags
			}
		}

		var diagsRec diag.Diagnostics
		recurrenceObj, diagsRec = types.ObjectValue(
			recurrenceAttrTypes,
			map[string]tfattr.Value{
				attr.RecurrenceType: types.StringValue(rec.Type),
				attr.DaysOfWeek:     daysOfWeekList,
				attr.Duration:       types.StringValue(rec.Duration),
			},
		)
		diags.Append(diagsRec...)
		if diags.HasError() {
			return types.ObjectNull(scheduleAttrTypes), diags
		}
	} else {
		recurrenceObj = types.ObjectNull(recurrenceAttrTypes)
	}

	objVal, diagsObj := types.ObjectValue(
		scheduleAttrTypes,
		map[string]tfattr.Value{
			attr.StartTime:  types.StringValue(d.Schedule.StartTime),
			attr.EndTime:    types.StringValue(d.Schedule.EndTime),
			attr.Timezone:   types.StringValue(d.Schedule.Timezone),
			attr.Recurrence: recurrenceObj,
		},
	)
	diags.Append(diagsObj...)

	return objVal, diags
}

// SetAlertIDs populates the model's AlertIDs from a Terraform list.
func (d *DowntimeSchedule) SetAlertIDs(tfAlertIDs types.List) {
	if utils.IsNullOrUnknown(tfAlertIDs) {
		d.AlertIDs = nil
		return
	}

	alertIDs := utils.Map(tfAlertIDs.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	d.AlertIDs = alertIDs
}

// SetSchedule populates the model's Schedule from a Terraform object.
func (d *DowntimeSchedule) SetSchedule(ctx context.Context, tfSchedule types.Object) error {
	if utils.IsNullOrUnknown(tfSchedule) {
		return nil
	}

	attrs := tfSchedule.Attributes()

	config := DowntimeScheduleConfig{}

	if v, ok := attrs[attr.StartTime]; ok {
		if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
			config.StartTime = s.ValueString()
		}
	}

	if v, ok := attrs[attr.EndTime]; ok {
		if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
			config.EndTime = s.ValueString()
		}
	}

	if v, ok := attrs[attr.Timezone]; ok {
		if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
			config.Timezone = s.ValueString()
		}
	}

	if v, ok := attrs[attr.Recurrence]; ok && !utils.IsNullOrUnknown(v) {
		recObj, ok2 := v.(types.Object)
		if !ok2 {
			return nil
		}

		recAttrs := recObj.Attributes()
		rec := &DowntimeRecurrence{}

		if rv, ok3 := recAttrs[attr.RecurrenceType]; ok3 {
			if s, ok4 := rv.(types.String); ok4 && !utils.IsNullOrUnknown(s) {
				rec.Type = s.ValueString()
			}
		}

		if rv, ok3 := recAttrs[attr.DaysOfWeek]; ok3 {
			if list, ok4 := rv.(types.List); ok4 && !utils.IsNullOrUnknown(list) {
				var days []int
				for _, elem := range list.Elements() {
					if intVal, ok5 := elem.(types.Int64); ok5 {
						days = append(days, int(intVal.ValueInt64()))
					}
				}
				rec.DaysOfWeek = days
			}
		}

		if rv, ok3 := recAttrs[attr.Duration]; ok3 {
			if s, ok4 := rv.(types.String); ok4 && !utils.IsNullOrUnknown(s) {
				rec.Duration = s.ValueString()
			}
		}

		config.Recurrence = rec
	}

	d.Schedule = config
	return nil
}
