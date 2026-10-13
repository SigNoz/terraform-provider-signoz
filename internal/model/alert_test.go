package model

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAlert_ConditionRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"compositeQuery": map[string]interface{}{
			"queryType": "builder",
		},
		"op":     "1",
		"target": float64(100),
	}

	alert := &Alert{Condition: original}

	tfValue, err := alert.ConditionToTerraform()
	if err != nil {
		t.Fatalf("ConditionToTerraform failed: %v", err)
	}

	if tfValue.IsNull() || tfValue.IsUnknown() {
		t.Fatal("expected non-null terraform value")
	}

	alert2 := &Alert{}
	err = alert2.SetCondition(tfValue)
	if err != nil {
		t.Fatalf("SetCondition failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(alert2.Condition)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestAlert_ConditionEmpty(t *testing.T) {
	// An empty map flattens to an empty string, which cannot be expanded back.
	// This verifies the known limitation: empty conditions should not round-trip.
	alert := &Alert{Condition: map[string]interface{}{}}

	tfValue, err := alert.ConditionToTerraform()
	if err != nil {
		t.Fatalf("ConditionToTerraform failed: %v", err)
	}

	alert2 := &Alert{}
	err = alert2.SetCondition(tfValue)
	// Empty JSON string cannot be expanded -- this is expected behavior
	if err == nil {
		// If it succeeds, verify the result is sensible
		if len(alert2.Condition) != 0 {
			t.Errorf("expected empty condition, got %v", alert2.Condition)
		}
	}
}

func TestAlert_EvaluationRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"window":   "5m",
		"strategy": "max",
	}

	alert := &Alert{Evaluation: original}

	tfValue, err := alert.EvaluationToTerraform()
	if err != nil {
		t.Fatalf("EvaluationToTerraform failed: %v", err)
	}

	alert2 := &Alert{}
	err = alert2.SetEvaluation(tfValue)
	if err != nil {
		t.Fatalf("SetEvaluation failed: %v", err)
	}

	orig, _ := json.Marshal(original)
	result, _ := json.Marshal(alert2.Evaluation)
	if string(orig) != string(result) {
		t.Errorf("round-trip mismatch:\n  original: %s\n  result:   %s", orig, result)
	}
}

func TestAlert_EvaluationNil(t *testing.T) {
	// A nil Evaluation map flattens to an empty string, which cannot be expanded.
	// This verifies the known behavior for nil/empty evaluation fields.
	alert := &Alert{}

	tfValue, err := alert.EvaluationToTerraform()
	if err != nil {
		t.Fatalf("EvaluationToTerraform failed on nil: %v", err)
	}

	alert2 := &Alert{}
	err = alert2.SetEvaluation(tfValue)
	// Empty JSON string expansion fails -- expected behavior
	if err == nil {
		if alert2.Evaluation != nil && len(alert2.Evaluation) != 0 {
			t.Errorf("expected nil/empty evaluation, got %v", alert2.Evaluation)
		}
	}
}

func TestAlert_LabelsRoundTrip(t *testing.T) {
	alert := &Alert{
		Labels: map[string]string{
			"env":       "prod",
			"team":      "backend",
			"severity":  "critical",
			"managedBy": "terraform",
		},
	}

	tfMap, diags := alert.LabelsToTerraform()
	if diags.HasError() {
		t.Fatalf("LabelsToTerraform had errors: %v", diags)
	}

	// severity and managedBy should be filtered out
	elements := tfMap.Elements()
	if _, ok := elements["severity"]; ok {
		t.Error("severity should be filtered from labels")
	}
	if _, ok := elements["managedBy"]; ok {
		t.Error("managedBy should be filtered from labels")
	}
	if len(elements) != 2 {
		t.Errorf("expected 2 labels, got %d", len(elements))
	}

	alert2 := &Alert{}
	alert2.SetLabels(tfMap, types.StringValue("critical"))

	if alert2.Labels["severity"] != "critical" {
		t.Errorf("expected severity=critical, got %s", alert2.Labels["severity"])
	}
	if alert2.Labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", alert2.Labels["env"])
	}
	if alert2.Labels["team"] != "backend" {
		t.Errorf("expected team=backend, got %s", alert2.Labels["team"])
	}
	// managedBy:terraform label should be re-added
	if alert2.Labels["managedBy"] != "terraform" {
		t.Errorf("expected managedBy=terraform, got %s", alert2.Labels["managedBy"])
	}
}

func TestAlert_LabelsEmpty(t *testing.T) {
	alert := &Alert{Labels: map[string]string{}}

	tfMap, diags := alert.LabelsToTerraform()
	if diags.HasError() {
		t.Fatalf("LabelsToTerraform had errors: %v", diags)
	}

	if len(tfMap.Elements()) != 0 {
		t.Errorf("expected 0 elements, got %d", len(tfMap.Elements()))
	}
}

func TestAlert_PreferredChannelsRoundTrip(t *testing.T) {
	original := []string{"slack-critical", "email-oncall", "pagerduty"}

	alert := &Alert{PreferredChannels: original}

	tfList, diags := alert.PreferredChannelsToTerraform()
	if diags.HasError() {
		t.Fatalf("PreferredChannelsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(tfList.Elements()))
	}

	alert2 := &Alert{}
	alert2.SetPreferredChannels(tfList)

	if len(alert2.PreferredChannels) != len(original) {
		t.Fatalf("expected %d channels, got %d", len(original), len(alert2.PreferredChannels))
	}
	for i, ch := range original {
		if alert2.PreferredChannels[i] != ch {
			t.Errorf("channel[%d]: expected %s, got %s", i, ch, alert2.PreferredChannels[i])
		}
	}
}

func TestAlert_PreferredChannelsEmpty(t *testing.T) {
	alert := &Alert{PreferredChannels: []string{}}

	tfList, diags := alert.PreferredChannelsToTerraform()
	if diags.HasError() {
		t.Fatalf("PreferredChannelsToTerraform had errors: %v", diags)
	}

	if len(tfList.Elements()) != 0 {
		t.Errorf("expected 0 elements, got %d", len(tfList.Elements()))
	}
}

func TestAlert_NotificationSettingsRoundTrip(t *testing.T) {
	ctx := context.Background()

	alert := &Alert{
		NotificationSettings: NotificationSettings{
			GroupBy:   []string{"service", "env"},
			UsePolicy: true,
			Renotify: Renotify{
				Enabled:          true,
				ReNotifyInterval: "5m",
				AlertStates:      []string{"firing", "nodata"},
			},
		},
	}

	tfObj, diags := alert.NotificationSettingsToTerraform(ctx)
	if diags.HasError() {
		t.Fatalf("NotificationSettingsToTerraform had errors: %v", diags)
	}

	if tfObj.IsNull() {
		t.Fatal("expected non-null notification settings object")
	}

	alert2 := &Alert{}
	err := alert2.SetNotificationSettings(ctx, tfObj)
	if err != nil {
		t.Fatalf("SetNotificationSettings failed: %v", err)
	}

	ns := alert2.NotificationSettings
	if !ns.UsePolicy {
		t.Error("expected UsePolicy=true")
	}
	if len(ns.GroupBy) != 2 || ns.GroupBy[0] != "service" || ns.GroupBy[1] != "env" {
		t.Errorf("unexpected GroupBy: %v", ns.GroupBy)
	}
	if !ns.Renotify.Enabled {
		t.Error("expected Renotify.Enabled=true")
	}
	if ns.Renotify.ReNotifyInterval != "5m" {
		t.Errorf("expected Renotify.Interval=5m, got %s", ns.Renotify.ReNotifyInterval)
	}
	if len(ns.Renotify.AlertStates) != 2 {
		t.Errorf("expected 2 alert states, got %d", len(ns.Renotify.AlertStates))
	}
}

func TestAlert_NotificationSettingsNull(t *testing.T) {
	ctx := context.Background()

	nsAttrTypes := attr.NotificationSettingsAttrTypes()
	nullObj := types.ObjectNull(nsAttrTypes)

	alert := &Alert{}
	err := alert.SetNotificationSettings(ctx, nullObj)
	if err != nil {
		t.Fatalf("SetNotificationSettings failed on null: %v", err)
	}

	// Should remain zero-value
	if alert.NotificationSettings.UsePolicy {
		t.Error("expected UsePolicy=false for null input")
	}
}

func TestAlert_NotificationSettingsEmptyGroupBy(t *testing.T) {
	ctx := context.Background()

	alert := &Alert{
		NotificationSettings: NotificationSettings{
			GroupBy: nil,
			Renotify: Renotify{
				Enabled:     false,
				AlertStates: []string{},
			},
		},
	}

	tfObj, diags := alert.NotificationSettingsToTerraform(ctx)
	if diags.HasError() {
		t.Fatalf("NotificationSettingsToTerraform had errors: %v", diags)
	}

	attrs := tfObj.Attributes()
	groupByVal := attrs[attr.GroupBy]
	groupByList, ok := groupByVal.(types.List)
	if !ok {
		t.Fatal("expected group_by to be a list")
	}
	if len(groupByList.Elements()) != 0 {
		t.Errorf("expected empty group_by list, got %d elements", len(groupByList.Elements()))
	}
}

func TestAlert_SetSourceIfEmpty(t *testing.T) {
	alert := &Alert{}
	alert.SetSourceIfEmpty("https://signoz.example.com")
	if alert.Source != "https://signoz.example.com/alerts" {
		t.Errorf("expected source with /alerts suffix, got %s", alert.Source)
	}

	// Should not overwrite existing source
	alert2 := &Alert{Source: "https://custom.example.com/alerts"}
	alert2.SetSourceIfEmpty("https://signoz.example.com")
	if alert2.Source != "https://custom.example.com/alerts" {
		t.Errorf("expected existing source preserved, got %s", alert2.Source)
	}
}

func TestAlert_ToTerraform(t *testing.T) {
	alert := &Alert{
		ID:        "123",
		Alert:     "test-alert",
		AlertType: AlertTypeMetrics,
		RuleType:  AlertRuleTypeThreshold,
		State:     AlertStateInactive,
	}

	result := alert.ToTerraform()
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}

	if m[attr.ID] != "123" {
		t.Errorf("expected id=123, got %v", m[attr.ID])
	}
	if m[attr.Alert] != "test-alert" {
		t.Errorf("expected alert=test-alert, got %v", m[attr.Alert])
	}
}

func TestAlert_GettersAndConstants(t *testing.T) {
	alert := Alert{ID: "42", Alert: "my-alert", AlertType: AlertTypeTraces}

	if alert.GetID() != "42" {
		t.Errorf("expected GetID=42, got %s", alert.GetID())
	}
	if alert.GetName() != "my-alert" {
		t.Errorf("expected GetName=my-alert, got %s", alert.GetName())
	}
	if alert.GetType() != AlertTypeTraces {
		t.Errorf("expected GetType=%s, got %s", AlertTypeTraces, alert.GetType())
	}
}

func TestAlert_SetLabelsNoSeverity(t *testing.T) {
	tfMap, _ := types.MapValue(types.StringType, map[string]tfattr.Value{
		"env": types.StringValue("staging"),
	})

	alert := &Alert{}
	alert.SetLabels(tfMap, types.StringValue(""))

	if _, ok := alert.Labels["severity"]; ok {
		t.Error("severity should not be set when empty string")
	}
	if alert.Labels["env"] != "staging" {
		t.Errorf("expected env=staging, got %s", alert.Labels["env"])
	}
	if alert.Labels["managedBy"] != "terraform" {
		t.Errorf("expected managedBy=terraform, got %s", alert.Labels["managedBy"])
	}
}
