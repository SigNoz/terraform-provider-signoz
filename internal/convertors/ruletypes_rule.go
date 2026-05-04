// rule is the most quirk-laden resource on the SigNoz API:
//
//   - `step` is a oneOf [string, number] in the OpenAPI; oapi-codegen models
//     it as a `Querybuildertypesv5Step` union with `As*`/`From*` accessors.
//     The hand-written DTO below treats it as `json.RawMessage` so we can
//     normalize either form into a string at flatten time.
//   - `evaluation` decomposes on the wire — the server returns rolling spec
//     fields (`evalWindow`, `frequency`) at the top level rather than inside
//     the envelope. The flat `evaluationSpec` here covers both directions
//     uniformly.
//   - `notificationSettings` is a pass-through `*json.RawMessage` —
//     unsupported in this provider phase, only round-tripped if the
//     server populates it.
//
// To migrate this onto skaff services without rewriting the per-field
// expand/flatten logic, the Layer 3 entry points (`ExpandRuletypesPostableRule`,
// `FlattenRuletypesRule`) JSON-bridge between this hand-written DTO and
// `apitypes.RuletypesPostableRule` / `apitypes.RuletypesRule`. The two
// share the same wire JSON shape, so marshal-then-unmarshal is a stable
// translation. The price is two extra json round-trips per CRUD call;
// the benefit is keeping all the existing Phase 2 minimum-viable scope
// guards (only METRIC_BASED_ALERT, threshold_rule, promql, single-query)
// intact in one place.
package conv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
	"github.com/SigNoz/terraform-provider-signoz/internal/schemas"
	customtypes "github.com/SigNoz/terraform-provider-signoz/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ---------------------------------------------------------------------------
// Hand-written DTOs — deliberately divergent from apitypes; see file header.
// ---------------------------------------------------------------------------

// rule is the JSON shape of both PostableRule (POST/PUT body) and the
// Gettable subset we keep in state. The server adds read-only fields
// like state on Get; we only round-trip the user-managed ones.
//
// **Phase 2 minimum-viable scope** — only one rule shape is supported:
//   - alertType: METRIC_BASED_ALERT
//   - ruleType:  threshold_rule
//   - condition.compositeQuery.queryType: promql
//   - condition.compositeQuery.queries[]: a single PromQL query envelope
//   - condition.thresholds: not modelled (use top-level target/op)
//   - evaluation: rolling or cumulative — both supported
//
// Anything else (anomaly rules, builder/clickhouse queries, multi-query
// composites, multi-threshold conditions, notification_settings, renotify)
// produces a clear "not yet supported" error during expand. Phase 3
// codegen will fill in the rest.
type rule struct {
	Alert                string              `json:"alert"`
	AlertType            string              `json:"alertType"`
	Annotations          map[string]string   `json:"annotations,omitempty"`
	Condition            *ruleCondition      `json:"condition"`
	Description          *string             `json:"description,omitempty"`
	Disabled             *bool               `json:"disabled,omitempty"`
	EvalWindow           *string             `json:"evalWindow,omitempty"`
	Evaluation           *evaluationEnvelope `json:"evaluation,omitempty"`
	Frequency            *string             `json:"frequency,omitempty"`
	ID                   string              `json:"id,omitempty"` // server-set
	Labels               map[string]string   `json:"labels,omitempty"`
	NotificationSettings *json.RawMessage    `json:"notificationSettings,omitempty"` // pass-through
	PreferredChannels    []string            `json:"preferredChannels,omitempty"`
	RuleType             string              `json:"ruleType"`
	SchemaVersion        *string             `json:"schemaVersion,omitempty"`
	Source               *string             `json:"source,omitempty"`
	State                string              `json:"state,omitempty"` // server-set, read-only
	Version              *string             `json:"version,omitempty"`

	// Server-set audit fields. Surfaced on the datasource model only;
	// resource model drops them via config.yml's schema.ignores.
	CreatedAt string `json:"createdAt,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	UpdatedBy string `json:"updatedBy,omitempty"`
}

// ruleCondition mirrors RuletypesRuleCondition. compositeQuery is required.
// thresholds is intentionally not modelled in this phase — users use the
// flat target+op+matchType triple instead.
type ruleCondition struct {
	AbsentFor         *int64               `json:"absentFor,omitempty"`
	AlertOnAbsent     *bool                `json:"alertOnAbsent,omitempty"`
	Algorithm         *string              `json:"algorithm,omitempty"`
	CompositeQuery    *alertCompositeQuery `json:"compositeQuery"`
	MatchType         *string              `json:"matchType,omitempty"`
	Op                *string              `json:"op,omitempty"`
	RequireMinPoints  *bool                `json:"requireMinPoints,omitempty"`
	RequiredNumPoints *int64               `json:"requiredNumPoints,omitempty"`
	Seasonality       *string              `json:"seasonality,omitempty"`
	SelectedQueryName *string              `json:"selectedQueryName,omitempty"`
	Target            *float64             `json:"target,omitempty"`
	TargetUnit        *string              `json:"targetUnit,omitempty"`
}

// alertCompositeQuery wraps a list of query envelopes. Required: queries,
// panelType, queryType.
type alertCompositeQuery struct {
	PanelType string          `json:"panelType"`
	Queries   []queryEnvelope `json:"queries"`
	QueryType string          `json:"queryType"`
	Unit      *string         `json:"unit,omitempty"`
}

// queryEnvelope is the wire shape per query. Each envelope has a `type`
// discriminator and a `spec` whose shape depends on that type. In this
// phase we only emit + read promql variants; everything else errors at
// expand time.
type queryEnvelope struct {
	Spec *promQuery `json:"spec"`
	Type string     `json:"type"`
}

// promQuery mirrors Querybuildertypesv5PromQuery. `step` is a oneOf
// [string, number] in the OpenAPI — the server accepts a Go-duration
// string ("60s") or a bare number of seconds (60); on Read it can
// return either form. Modelling as RawMessage and normalising to a
// string in flatten avoids fighting the framework's typed schema.
type promQuery struct {
	Disabled *bool           `json:"disabled,omitempty"`
	Legend   *string         `json:"legend,omitempty"`
	Name     *string         `json:"name,omitempty"`
	Query    string          `json:"query"`
	Stats    *bool           `json:"stats,omitempty"`
	Step     json.RawMessage `json:"step,omitempty"`
}

// evaluationEnvelope mirrors RuletypesEvaluationEnvelope. Exactly one of
// Cumulative / Rolling is set; `kind` discriminates.
type evaluationEnvelope struct {
	Kind string          `json:"kind"`
	Spec *evaluationSpec `json:"spec"`
}

// evaluationSpec is the union of cumulative + rolling fields. The server
// reads it via the `kind` discriminator; we send only the matching
// subset (omitempty drops the rest). On flatten the deserialiser
// populates whichever kind matches.
type evaluationSpec struct {
	// rolling
	EvalWindow *string `json:"evalWindow,omitempty"`
	Frequency  *string `json:"frequency,omitempty"`
	// cumulative
	Schedule *cumulativeSchedule `json:"schedule,omitempty"`
	Timezone *string             `json:"timezone,omitempty"`
}

// cumulativeSchedule mirrors RuletypesCumulativeSchedule. The fields are
// all optional and depend on `type` (hourly | daily | weekly | monthly).
type cumulativeSchedule struct {
	Day     *int64 `json:"day,omitempty"`
	Hour    *int64 `json:"hour,omitempty"`
	Minute  *int64 `json:"minute,omitempty"`
	Type    string `json:"type"`
	Weekday *int64 `json:"weekday,omitempty"`
}

// ---------------------------------------------------------------------------
// Layer 3 entry points — JSON-bridge between hand-written DTO and apitypes.
// ---------------------------------------------------------------------------

// ExpandRuletypesPostableRule is the Layer 3 expander the codegen-driven
// service template calls for both Create and Update. The hand-written
// `expandRule` produces the per-field DTO; we then JSON-marshal it and
// unmarshal back into the apitype. The two share the wire shape, so the
// round-trip is safe.
func ExpandRuletypesPostableRule(ctx context.Context, m schemas.RuleModel) (*apitypes.RuletypesPostableRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	in, d := expandRule(ctx, m)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	raw, err := json.Marshal(in)
	if err != nil {
		diags.AddError("Marshal rule", err.Error())
		return nil, diags
	}
	out := &apitypes.RuletypesPostableRule{}
	if err := json.Unmarshal(raw, out); err != nil {
		diags.AddError("Bridge rule -> apitypes.RuletypesPostableRule", err.Error())
		return nil, diags
	}
	return out, diags
}

// FlattenRuletypesRule is the Layer 3 flattener for Read — also called
// after Create/Update via the uniform Create+GET / Update+GET refresh
// pattern. The apitype is JSON-bridged into the hand-written DTO, then
// the existing per-field flatten produces the wide DataSource model. The
// service template emits `<r>ResourceFromDS(next)` to narrow on the
// resource side.
func FlattenRuletypesRule(ctx context.Context, g *apitypes.RuletypesRule) (*schemas.RuleDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if g == nil {
		return nil, diags
	}
	raw, err := json.Marshal(g)
	if err != nil {
		diags.AddError("Marshal apitypes.RuletypesRule", err.Error())
		return nil, diags
	}
	var dto rule
	if err := json.Unmarshal(raw, &dto); err != nil {
		diags.AddError("Bridge apitypes.RuletypesRule -> rule", err.Error())
		return nil, diags
	}
	return flattenRule(ctx, &dto)
}

// ---------------------------------------------------------------------------
// expand: framework model -> rule DTO
// ---------------------------------------------------------------------------

const (
	supportedAlertType = "METRIC_BASED_ALERT"
	supportedRuleType  = "threshold_rule"
	supportedQueryType = "promql"
)

func expandRule(ctx context.Context, m schemas.RuleModel) (*rule, diag.Diagnostics) {
	var diags diag.Diagnostics

	if at := m.AlertType.ValueString(); at != "" && at != supportedAlertType {
		diags.AddError("Unsupported alert_type",
			fmt.Sprintf("only %q is supported by terraform-provider-signoz in this phase; got %q", supportedAlertType, at))
		return nil, diags
	}
	if rt := m.RuleType.ValueString(); rt != "" && rt != supportedRuleType && rt != "promql_rule" {
		diags.AddError("Unsupported rule_type",
			fmt.Sprintf("only %q and \"promql_rule\" are supported by terraform-provider-signoz in this phase; got %q", supportedRuleType, rt))
		return nil, diags
	}

	cond, d := expandRuleCondition(ctx, m.Condition)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	eval, d := expandRuleEvaluationEnvelope(ctx, m.Evaluation)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	annotations, d := stringMapFromTFMap(ctx, m.Annotations)
	diags.Append(d...)
	labels, d := stringMapFromTFMap(ctx, m.Labels)
	diags.Append(d...)
	preferred, d := stringSliceFromTFList(ctx, m.PreferredChannels)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.NotificationSettings.IsNull() && !m.NotificationSettings.IsUnknown() {
		diags.AddError("Unsupported field",
			"notification_settings is not yet supported in this provider phase")
		return nil, diags
	}

	rtOut := m.RuleType.ValueString()

	out := &rule{
		Alert:             m.Alert.ValueString(),
		AlertType:         supportedAlertType,
		Annotations:       annotations,
		Condition:         cond,
		Description:       StringPointer(m.Description),
		Disabled:          BoolPointer(m.Disabled),
		EvalWindow:        StringPointer(m.EvalWindow),
		Evaluation:        eval,
		Frequency:         StringPointer(m.Frequency),
		Labels:            labels,
		PreferredChannels: preferred,
		RuleType:          rtOut,
		SchemaVersion:     StringPointer(m.SchemaVersion),
		Source:            StringPointer(m.Source),
		Version:           StringPointer(m.Version),
	}
	return out, diags
}

func expandRuleCondition(ctx context.Context, cv customtypes.RuletypesRuleConditionValue) (*ruleCondition, diag.Diagnostics) {
	var diags diag.Diagnostics
	if cv.IsNull() || cv.IsUnknown() {
		diags.AddError("Missing required field", "condition is required")
		return nil, diags
	}

	if !cv.Thresholds.IsNull() && !cv.Thresholds.IsUnknown() {
		diags.AddError("Unsupported field",
			"condition.thresholds is not yet supported in this provider phase; use the flat target+op+match_type instead")
		return nil, diags
	}
	if alg := cv.Algorithm.ValueString(); alg != "" {
		diags.AddError("Unsupported field",
			fmt.Sprintf("condition.algorithm is not yet supported in this provider phase (got %q); requires anomaly_rule which is also not supported", alg))
		return nil, diags
	}
	if seas := cv.Seasonality.ValueString(); seas != "" {
		diags.AddError("Unsupported field",
			"condition.seasonality is not yet supported in this provider phase")
		return nil, diags
	}

	cq, d := expandAlertCompositeQuery(ctx, cv.CompositeQuery)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	target := NumberFloat64Pointer(cv.Target)

	return &ruleCondition{
		AbsentFor:         Int64Pointer(cv.AbsentFor),
		AlertOnAbsent:     BoolPointer(cv.AlertOnAbsent),
		CompositeQuery:    cq,
		MatchType:         StringPointer(cv.MatchType),
		Op:                StringPointer(cv.Op),
		RequireMinPoints:  BoolPointer(cv.RequireMinPoints),
		RequiredNumPoints: Int64Pointer(cv.RequiredNumPoints),
		SelectedQueryName: StringPointer(cv.SelectedQueryName),
		Target:            target,
		TargetUnit:        StringPointer(cv.TargetUnit),
	}, diags
}

func expandAlertCompositeQuery(ctx context.Context, ov basetypes.ObjectValue) (*alertCompositeQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		diags.AddError("Missing required field", "condition.composite_query is required")
		return nil, diags
	}
	cqv, d := customtypes.NewRuletypesAlertCompositeQueryValue(
		customtypes.RuletypesAlertCompositeQueryValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if qt := cqv.QueryType.ValueString(); qt != "" && qt != supportedQueryType {
		diags.AddError("Unsupported query_type",
			fmt.Sprintf("only %q is supported by terraform-provider-signoz in this phase; got %q", supportedQueryType, qt))
		return nil, diags
	}

	queries, d := expandQueryList(ctx, cqv.Queries)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &alertCompositeQuery{
		PanelType: cqv.PanelType.ValueString(),
		Queries:   queries,
		QueryType: supportedQueryType,
		Unit:      StringPointer(cqv.Unit),
	}, diags
}

func expandQueryList(ctx context.Context, l types.List) ([]queryEnvelope, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() || len(l.Elements()) == 0 {
		diags.AddError("Missing required field", "condition.composite_query.queries must contain at least one entry")
		return nil, diags
	}

	out := make([]queryEnvelope, 0, len(l.Elements()))
	for i, el := range l.Elements() {
		ev, ok := el.(customtypes.Querybuildertypesv5QueryEnvelopeValue)
		if !ok {
			diags.AddError("Unexpected queries element type",
				fmt.Sprintf("element %d: expected Querybuildertypesv5QueryEnvelopeValue, got %T", i, el))
			return nil, diags
		}

		// Reject every variant except promql.
		if !ev.BuilderLog.IsNull() && !ev.BuilderLog.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].builder_log is not yet supported in this provider phase", i))
			return nil, diags
		}
		if !ev.BuilderMetric.IsNull() && !ev.BuilderMetric.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].builder_metric is not yet supported in this provider phase", i))
			return nil, diags
		}
		if !ev.BuilderTrace.IsNull() && !ev.BuilderTrace.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].builder_trace is not yet supported in this provider phase", i))
			return nil, diags
		}
		if !ev.ClickhouseSql.IsNull() && !ev.ClickhouseSql.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].clickhouse_sql is not yet supported in this provider phase", i))
			return nil, diags
		}
		if !ev.Formula.IsNull() && !ev.Formula.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].formula is not yet supported in this provider phase", i))
			return nil, diags
		}
		if !ev.TraceOperator.IsNull() && !ev.TraceOperator.IsUnknown() {
			diags.AddError("Unsupported query envelope variant",
				fmt.Sprintf("queries[%d].trace_operator is not yet supported in this provider phase", i))
			return nil, diags
		}
		if ev.Promql.IsNull() || ev.Promql.IsUnknown() {
			diags.AddError("Missing required field",
				fmt.Sprintf("queries[%d].promql must be set", i))
			return nil, diags
		}

		spec, d := expandPromQuery(ctx, ev.Promql)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		out = append(out, queryEnvelope{
			Spec: spec,
			Type: supportedQueryType,
		})
	}
	return out, diags
}

func expandPromQuery(ctx context.Context, ov basetypes.ObjectValue) (*promQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	pv, d := customtypes.NewQuerybuildertypesv5PromQueryValue(
		customtypes.Querybuildertypesv5PromQueryValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	q := &promQuery{
		Disabled: BoolPointer(pv.Disabled),
		Legend:   StringPointer(pv.Legend),
		Name:     StringPointer(pv.Name),
		Query:    pv.Query.ValueString(),
		Stats:    BoolPointer(pv.Stats),
	}
	if step := StringPointer(pv.Step); step != nil {
		raw, err := json.Marshal(*step)
		if err != nil {
			diags.AddError("Encode promql.step", err.Error())
			return nil, diags
		}
		q.Step = raw
	}
	return q, diags
}

func expandRuleEvaluationEnvelope(ctx context.Context, ev customtypes.RuletypesEvaluationEnvelopeValue) (*evaluationEnvelope, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ev.IsNull() || ev.IsUnknown() {
		return nil, diags
	}

	cumSet := !ev.Cumulative.IsNull() && !ev.Cumulative.IsUnknown()
	rolSet := !ev.Rolling.IsNull() && !ev.Rolling.IsUnknown()
	if cumSet && rolSet {
		diags.AddError("Conflicting evaluation envelope",
			"only one of evaluation.cumulative or evaluation.rolling may be set")
		return nil, diags
	}

	switch {
	case rolSet:
		rv, d := customtypes.NewRuletypesRollingWindowValue(
			customtypes.RuletypesRollingWindowValue{}.AttributeTypes(ctx),
			ev.Rolling.Attributes(),
		)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &evaluationEnvelope{
			Kind: "rolling",
			Spec: &evaluationSpec{
				EvalWindow: StringPointer(rv.EvalWindow),
				Frequency:  StringPointer(rv.Frequency),
			},
		}, diags
	case cumSet:
		cwv, d := customtypes.NewRuletypesCumulativeWindowValue(
			customtypes.RuletypesCumulativeWindowValue{}.AttributeTypes(ctx),
			ev.Cumulative.Attributes(),
		)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		sched, d := expandCumulativeSchedule(ctx, cwv.Schedule)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &evaluationEnvelope{
			Kind: "cumulative",
			Spec: &evaluationSpec{
				Frequency: StringPointer(cwv.Frequency),
				Schedule:  sched,
				Timezone:  StringPointer(cwv.Timezone),
			},
		}, diags
	default:
		return nil, diags
	}
}

func expandCumulativeSchedule(ctx context.Context, ov basetypes.ObjectValue) (*cumulativeSchedule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ov.IsNull() || ov.IsUnknown() {
		diags.AddError("Missing required field",
			"evaluation.cumulative.schedule is required when evaluation kind is cumulative")
		return nil, diags
	}
	csv, d := customtypes.NewRuletypesCumulativeScheduleValue(
		customtypes.RuletypesCumulativeScheduleValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &cumulativeSchedule{
		Day:     Int64Pointer(csv.Day),
		Hour:    Int64Pointer(csv.Hour),
		Minute:  Int64Pointer(csv.Minute),
		Type:    csv.Type_.ValueString(),
		Weekday: Int64Pointer(csv.Weekday),
	}, diags
}

// ---------------------------------------------------------------------------
// flatten: rule DTO -> framework model
// ---------------------------------------------------------------------------

// flattenRule converts a rule DTO into the *wide* datasource model. The
// resource model is a strict subset (no audit fields); the service-layer
// template auto-generates `ruleResourceFromDS` to narrow.
func flattenRule(ctx context.Context, g *rule) (*schemas.RuleDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	annotations, d := tfMapFromStringMap(ctx, g.Annotations)
	diags.Append(d...)
	labels, d := tfMapFromStringMap(ctx, g.Labels)
	diags.Append(d...)
	preferred, d := tfListFromStringSlice(ctx, g.PreferredChannels)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	cond, d := flattenRuleCondition(ctx, g.Condition)
	diags.Append(d...)

	// The server doesn't always echo the `evaluation` envelope — for
	// rolling evaluations it decomposes the spec into top-level
	// `evalWindow` and `frequency` fields. Synthesize the envelope back
	// from those when needed so plan-consistency holds.
	evalIn := g.Evaluation
	if evalIn == nil && (g.EvalWindow != nil && *g.EvalWindow != "") {
		evalIn = &evaluationEnvelope{
			Kind: "rolling",
			Spec: &evaluationSpec{
				EvalWindow: g.EvalWindow,
				Frequency:  g.Frequency,
			},
		}
	}
	eval, d := flattenRuleEvaluationEnvelope(ctx, evalIn)
	diags.Append(d...)

	notif := types.ObjectNull(customtypes.RuletypesNotificationSettingsValue{}.AttributeTypes(ctx))
	notifTyped, d := castNotificationSettingsTyped(ctx, notif)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &schemas.RuleDataSourceModel{
		Alert:                types.StringValue(g.Alert),
		AlertType:            types.StringValue(g.AlertType),
		Annotations:          annotations,
		Condition:            cond,
		CreatedAt:            types.StringValue(g.CreatedAt),
		CreatedBy:            types.StringValue(g.CreatedBy),
		Description:          types.StringValue(strDeref(g.Description)),
		Disabled:             types.BoolValue(boolDeref(g.Disabled)),
		EvalWindow:           types.StringValue(strDeref(g.EvalWindow)),
		Evaluation:           eval,
		Frequency:            types.StringValue(strDeref(g.Frequency)),
		Id:                   types.StringValue(g.ID),
		Labels:               labels,
		NotificationSettings: notifTyped,
		PreferredChannels:    preferred,
		RuleType:             types.StringValue(g.RuleType),
		SchemaVersion:        types.StringValue(strDeref(g.SchemaVersion)),
		Source:               types.StringValue(strDeref(g.Source)),
		State:                types.StringValue(g.State),
		UpdatedAt:            types.StringValue(g.UpdatedAt),
		UpdatedBy:            types.StringValue(g.UpdatedBy),
		Version:              types.StringValue(strDeref(g.Version)),
	}, diags
}

func flattenRuleCondition(ctx context.Context, c *ruleCondition) (customtypes.RuletypesRuleConditionValue, diag.Diagnostics) {
	if c == nil {
		return customtypes.NewRuletypesRuleConditionValueNull(), nil
	}

	var diags diag.Diagnostics

	cq, d := flattenAlertCompositeQuery(ctx, c.CompositeQuery)
	diags.Append(d...)
	if diags.HasError() {
		return customtypes.NewRuletypesRuleConditionValueUnknown(), diags
	}

	thresholds := types.ObjectNull(customtypes.RuletypesRuleThresholdDataValue{}.AttributeTypes(ctx))

	cv, d := customtypes.NewRuletypesRuleConditionValue(
		customtypes.RuletypesRuleConditionValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"absent_for":          Int64FromPointer(c.AbsentFor),
			"alert_on_absent":     BoolFromPointer(c.AlertOnAbsent),
			"algorithm":           StringFromPointer(c.Algorithm),
			"composite_query":     cq,
			"match_type":          StringFromPointer(c.MatchType),
			"op":                  StringFromPointer(c.Op),
			"require_min_points":  BoolFromPointer(c.RequireMinPoints),
			"required_num_points": Int64FromPointer(c.RequiredNumPoints),
			"seasonality":         StringFromPointer(c.Seasonality),
			"selected_query_name": StringFromPointer(c.SelectedQueryName),
			"target":              NumberFloat64FromPointer(c.Target),
			"target_unit":         StringFromPointer(c.TargetUnit),
			"thresholds":          thresholds,
		},
	)
	diags.Append(d...)
	return cv, diags
}

func flattenAlertCompositeQuery(ctx context.Context, c *alertCompositeQuery) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.RuletypesAlertCompositeQueryValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	var diags diag.Diagnostics
	queries, d := flattenQueryList(ctx, c.Queries)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(attrTypes), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		"panel_type": types.StringValue(c.PanelType),
		"queries":    queries,
		"query_type": types.StringValue(c.QueryType),
		"unit":       StringFromPointer(c.Unit),
	})
	diags.Append(d...)
	return obj, diags
}

func flattenQueryList(ctx context.Context, qs []queryEnvelope) (types.List, diag.Diagnostics) {
	elemType := customtypes.Querybuildertypesv5QueryEnvelopeValue{}.Type(ctx)
	if qs == nil {
		return types.ListNull(elemType), nil
	}

	var diags diag.Diagnostics

	// Typed nulls for the 6 envelope variants we don't model. The variant
	// types are the *spec shapes* (skaff inlined them when rewriting the
	// QueryEnvelope oneOf), not wrappers — names are long but mechanical.
	builderLogType := customtypes.Querybuildertypesv5QueryBuilderQueryGithubComSigNozSignozPkgTypesQuerybuildertypesQuerybuildertypesv5LogAggregationValue{}.AttributeTypes(ctx)
	builderMetricType := customtypes.Querybuildertypesv5QueryBuilderQueryGithubComSigNozSignozPkgTypesQuerybuildertypesQuerybuildertypesv5MetricAggregationValue{}.AttributeTypes(ctx)
	builderTraceType := customtypes.Querybuildertypesv5QueryBuilderQueryGithubComSigNozSignozPkgTypesQuerybuildertypesQuerybuildertypesv5TraceAggregationValue{}.AttributeTypes(ctx)
	clickhouseType := customtypes.Querybuildertypesv5ClickHouseQueryValue{}.AttributeTypes(ctx)
	formulaType := customtypes.Querybuildertypesv5QueryBuilderFormulaValue{}.AttributeTypes(ctx)
	traceOperatorType := customtypes.Querybuildertypesv5QueryBuilderTraceOperatorValue{}.AttributeTypes(ctx)

	envelopeAttrTypes := customtypes.Querybuildertypesv5QueryEnvelopeValue{}.AttributeTypes(ctx)
	elems := make([]attr.Value, 0, len(qs))
	for _, q := range qs {
		promql, d := flattenPromQuery(ctx, q.Spec)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListUnknown(elemType), diags
		}
		ev, d := customtypes.NewQuerybuildertypesv5QueryEnvelopeValue(
			envelopeAttrTypes,
			map[string]attr.Value{
				"builder_log":    types.ObjectNull(builderLogType),
				"builder_metric": types.ObjectNull(builderMetricType),
				"builder_trace":  types.ObjectNull(builderTraceType),
				"clickhouse_sql": types.ObjectNull(clickhouseType),
				"formula":        types.ObjectNull(formulaType),
				"promql":         promql,
				"trace_operator": types.ObjectNull(traceOperatorType),
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListUnknown(elemType), diags
		}
		elems = append(elems, ev)
	}
	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)
	return listVal, diags
}

func flattenPromQuery(ctx context.Context, p *promQuery) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.Querybuildertypesv5PromQueryValue{}.AttributeTypes(ctx)
	if p == nil {
		return types.ObjectNull(attrTypes), nil
	}
	obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"disabled": BoolFromPointer(p.Disabled),
		"legend":   StringFromPointer(p.Legend),
		"name":     StringFromPointer(p.Name),
		"query":    types.StringValue(p.Query),
		"stats":    BoolFromPointer(p.Stats),
		"step":     stepFromRaw(p.Step),
	})
	return obj, diags
}

// stepFromRaw normalises Querybuildertypesv5Step from its oneOf wire shape
// into a framework string. JSON null/missing → null; quoted string → that
// string; bare number → "<n>s" (treats numbers as seconds, matching
// signoz's API contract).
func stepFromRaw(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return types.StringValue(s)
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		return types.StringValue(fmt.Sprintf("%ds", int64(n)))
	}
	return types.StringValue(string(raw))
}

func flattenRuleEvaluationEnvelope(ctx context.Context, e *evaluationEnvelope) (customtypes.RuletypesEvaluationEnvelopeValue, diag.Diagnostics) {
	cumAttrTypes := customtypes.RuletypesCumulativeWindowValue{}.AttributeTypes(ctx)
	rolAttrTypes := customtypes.RuletypesRollingWindowValue{}.AttributeTypes(ctx)

	if e == nil || e.Spec == nil {
		return customtypes.NewRuletypesEvaluationEnvelopeValueNull(), nil
	}

	var diags diag.Diagnostics

	cum := types.ObjectNull(cumAttrTypes)
	rol := types.ObjectNull(rolAttrTypes)

	switch e.Kind {
	case "rolling":
		var d diag.Diagnostics
		rol, d = types.ObjectValue(rolAttrTypes, map[string]attr.Value{
			"eval_window": StringFromPointer(e.Spec.EvalWindow),
			"frequency":   StringFromPointer(e.Spec.Frequency),
		})
		diags.Append(d...)
	case "cumulative":
		sched, d := flattenCumulativeSchedule(ctx, e.Spec.Schedule)
		diags.Append(d...)
		if diags.HasError() {
			return customtypes.NewRuletypesEvaluationEnvelopeValueUnknown(), diags
		}
		cum, d = types.ObjectValue(cumAttrTypes, map[string]attr.Value{
			"frequency": StringFromPointer(e.Spec.Frequency),
			"schedule":  sched,
			"timezone":  StringFromPointer(e.Spec.Timezone),
		})
		diags.Append(d...)
	default:
		// Unknown evaluation.kind — leave both null.
	}
	if diags.HasError() {
		return customtypes.NewRuletypesEvaluationEnvelopeValueUnknown(), diags
	}

	ev, d := customtypes.NewRuletypesEvaluationEnvelopeValue(
		customtypes.RuletypesEvaluationEnvelopeValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cumulative": cum,
			"rolling":    rol,
		},
	)
	diags.Append(d...)
	return ev, diags
}

func flattenCumulativeSchedule(ctx context.Context, c *cumulativeSchedule) (basetypes.ObjectValue, diag.Diagnostics) {
	attrTypes := customtypes.RuletypesCumulativeScheduleValue{}.AttributeTypes(ctx)
	if c == nil {
		return types.ObjectNull(attrTypes), nil
	}
	obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
		"day":     Int64FromPointer(c.Day),
		"hour":    Int64FromPointer(c.Hour),
		"minute":  Int64FromPointer(c.Minute),
		"type_":   types.StringValue(c.Type),
		"weekday": Int64FromPointer(c.Weekday),
	})
	return obj, diags
}

func castNotificationSettingsTyped(ctx context.Context, ov basetypes.ObjectValue) (customtypes.RuletypesNotificationSettingsValue, diag.Diagnostics) {
	if ov.IsNull() {
		return customtypes.NewRuletypesNotificationSettingsValueNull(), nil
	}
	if ov.IsUnknown() {
		return customtypes.NewRuletypesNotificationSettingsValueUnknown(), nil
	}
	return customtypes.NewRuletypesNotificationSettingsValue(
		customtypes.RuletypesNotificationSettingsValue{}.AttributeTypes(ctx),
		ov.Attributes(),
	)
}

// ---------------------------------------------------------------------------
// Small private helpers — non-pointer slice/map glue that the existing
// rule expand/flatten was written against. Conv's exported `*[]string` /
// `*map[string]string` variants don't fit here without rewriting every
// call site.
// ---------------------------------------------------------------------------

func tfListFromStringSlice(_ context.Context, ss []string) (types.List, diag.Diagnostics) {
	if ss == nil {
		return types.ListNull(types.StringType), nil
	}
	elems := make([]attr.Value, 0, len(ss))
	for _, s := range ss {
		elems = append(elems, types.StringValue(s))
	}
	return types.ListValue(types.StringType, elems)
}

func stringSliceFromTFList(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	if l.IsNull() || l.IsUnknown() {
		return nil, nil
	}
	out := make([]string, 0, len(l.Elements()))
	var diags diag.Diagnostics
	diags = l.ElementsAs(ctx, &out, false)
	return out, diags
}

func tfMapFromStringMap(_ context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if m == nil {
		return types.MapNull(types.StringType), nil
	}
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	return types.MapValue(types.StringType, elems)
}

func stringMapFromTFMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	out := make(map[string]string, len(m.Elements()))
	diags := m.ElementsAs(ctx, &out, false)
	return out, diags
}

func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func boolDeref(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}
