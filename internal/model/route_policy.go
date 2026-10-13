package model

import (
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/attr"
	"github.com/SigNoz/terraform-provider-signoz/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoutePolicy model.
type RoutePolicy struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	Enabled        bool           `json:"enabled"`
	Matchers       []RouteMatcher `json:"matchers,omitempty"`
	Expression     string         `json:"expression,omitempty"`
	ChannelIDs     []string       `json:"channel_ids,omitempty"`
	Channels       []string       `json:"channels,omitempty"`
	Continue       bool           `json:"continue"`
	GroupWait      string         `json:"group_wait,omitempty"`
	GroupInterval  string         `json:"group_interval,omitempty"`
	RepeatInterval string         `json:"repeat_interval,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	CreatedBy      string         `json:"created_by,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
	UpdatedBy      string         `json:"updated_by,omitempty"`
}

// PrepareForAPI populates v0.104.0 fields (expression, channels) from terraform fields (matchers, channel_ids).
func (r *RoutePolicy) PrepareForAPI() {
	r.BuildExpression()
	if len(r.Channels) == 0 && len(r.ChannelIDs) > 0 {
		r.Channels = r.ChannelIDs
	}
}

// BuildExpression converts matchers to an expression string for v0.104.0 API.
func (r *RoutePolicy) BuildExpression() {
	if len(r.Matchers) == 0 {
		return
	}
	parts := make([]string, 0, len(r.Matchers))
	for _, m := range r.Matchers {
		switch m.MatchType {
		case "regex":
			parts = append(parts, m.Name+" matches \""+m.Value+"\"")
		case "not_regex":
			parts = append(parts, m.Name+" not matches \""+m.Value+"\"")
		case "not_equals":
			parts = append(parts, m.Name+" != \""+m.Value+"\"")
		default:
			parts = append(parts, m.Name+" == \""+m.Value+"\"")
		}
	}
	r.Expression = strings.Join(parts, " AND ")
}

// PopulateFromAPI fills terraform fields from v0.104.0 API response.
func (r *RoutePolicy) PopulateFromAPI() {
	r.ParseExpression()
	if len(r.ChannelIDs) == 0 && len(r.Channels) > 0 {
		r.ChannelIDs = r.Channels
	}
}

// ParseExpression converts an expression string back to matchers.
func (r *RoutePolicy) ParseExpression() {
	if r.Expression == "" || len(r.Matchers) > 0 {
		return
	}
	clauses := strings.Split(r.Expression, " AND ")
	for _, clause := range clauses {
		clause = strings.TrimSpace(clause)
		var name, value, matchType string
		if idx := strings.Index(clause, " matches "); idx > 0 {
			name = clause[:idx]
			value = strings.Trim(clause[idx+9:], "\"")
			matchType = "regex"
		} else if idx := strings.Index(clause, " not matches "); idx > 0 {
			name = clause[:idx]
			value = strings.Trim(clause[idx+13:], "\"")
			matchType = "not_regex"
		} else if idx := strings.Index(clause, " != "); idx > 0 {
			name = clause[:idx]
			value = strings.Trim(clause[idx+4:], "\"")
			matchType = "not_equals"
		} else if idx := strings.Index(clause, " == "); idx > 0 {
			name = clause[:idx]
			value = strings.Trim(clause[idx+4:], "\"")
			matchType = "equals"
		} else {
			continue
		}
		r.Matchers = append(r.Matchers, RouteMatcher{
			Name:      strings.TrimSpace(name),
			Value:     value,
			MatchType: matchType,
		})
	}
}

// RouteMatcher model.
type RouteMatcher struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	MatchType string `json:"match_type"`
}

// MatcherAttrTypes returns the attribute types for a single matcher object.
func MatcherAttrTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.MatchName:  types.StringType,
		attr.MatchValue: types.StringType,
		attr.MatchType:  types.StringType,
	}
}

// MatchersToTerraform converts matchers to a Terraform list value.
func (r RoutePolicy) MatchersToTerraform() (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: MatcherAttrTypes()}

	if len(r.Matchers) == 0 {
		return types.ListValueMust(elemType, []tfattr.Value{}), nil
	}

	elements := make([]tfattr.Value, 0, len(r.Matchers))
	for _, m := range r.Matchers {
		obj, diags := types.ObjectValue(MatcherAttrTypes(), map[string]tfattr.Value{
			attr.MatchName:  types.StringValue(m.Name),
			attr.MatchValue: types.StringValue(m.Value),
			attr.MatchType:  types.StringValue(m.MatchType),
		})
		if diags.HasError() {
			return types.ListNull(elemType), diags
		}
		elements = append(elements, obj)
	}

	return types.ListValue(elemType, elements)
}

// ChannelIDsToTerraform converts channel IDs to a Terraform list value.
func (r RoutePolicy) ChannelIDsToTerraform() (types.List, diag.Diagnostics) {
	channelIDs := utils.Map(r.ChannelIDs, func(value string) tfattr.Value {
		return types.StringValue(value)
	})

	return types.ListValue(types.StringType, channelIDs)
}

// SetMatchers sets matchers from a Terraform list value.
func (r *RoutePolicy) SetMatchers(tfMatchers types.List) {
	matchers := make([]RouteMatcher, 0, len(tfMatchers.Elements()))
	for _, elem := range tfMatchers.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		matcher := RouteMatcher{}
		if v, ok := attrs[attr.MatchName]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.Name = s.ValueString()
			}
		}
		if v, ok := attrs[attr.MatchValue]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.Value = s.ValueString()
			}
		}
		if v, ok := attrs[attr.MatchType]; ok {
			if s, ok2 := v.(types.String); ok2 && !utils.IsNullOrUnknown(s) {
				matcher.MatchType = s.ValueString()
			}
		}
		matchers = append(matchers, matcher)
	}
	r.Matchers = matchers
}

// SetChannelIDs sets channel IDs from a Terraform list value.
func (r *RoutePolicy) SetChannelIDs(tfChannelIDs types.List) {
	channelIDs := utils.Map(tfChannelIDs.Elements(), func(value tfattr.Value) string {
		return strings.Trim(value.String(), "\"")
	})
	r.ChannelIDs = channelIDs
}
