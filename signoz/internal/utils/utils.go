package utils

import (
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetValueString - get value from string or return default.
func GetValueString(element types.String, defaultValue string) string {
	if element.IsNull() || element.IsUnknown() || element.ValueString() == "" {
		return defaultValue
	}

	return element.ValueString()
}

// GetValueBool - get value from bool or return default.
func GetValueBool(element types.Bool, defaultValue bool) bool {
	if element.IsNull() || element.IsUnknown() {
		return defaultValue
	}

	return element.ValueBool()
}

// WithDefault - return default value if value is zero.
func WithDefault[T comparable](val, defaultVal T) T {
	var zeroValue T
	if val == zeroValue {
		return defaultVal
	}

	return val
}

// Map - transform giving slice of items by applying the func.
func Map[T, R any](items []T, f func(item T) R) []R {
	result := make([]R, 0, len(items))

	for _, item := range items {
		result = append(result, f(item))
	}

	return result
}

// Filter - filter down the elements from the given array that
// pass the test implemented by the provided function.
func Filter[T any](items []T, ok func(item T) bool) []T {
	result := make([]T, 0, len(items))

	for _, item := range items {
		if ok(item) {
			result = append(result, item)
		}
	}

	return result
}

// Contains - checks if element exists in the slice.
func Contains[T comparable](items []T, element T) bool {
	for _, item := range items {
		if item == element {
			return true
		}
	}

	return false
}

// TfValueToGo converts a Terraform attr.Value to a Go value.
func TfValueToGo(val tfattr.Value) interface{} {
	switch v := val.(type) {
	case types.String:
		return v.ValueString()
	case types.Bool:
		return v.ValueBool()
	case types.Int64:
		return v.ValueInt64()
	case types.Float64:
		return v.ValueFloat64()
	case types.List:
		result := make([]interface{}, 0, len(v.Elements()))
		for _, elem := range v.Elements() {
			result = append(result, TfValueToGo(elem))
		}
		return result
	case types.Object:
		result := make(map[string]interface{})
		for k, attrVal := range v.Attributes() {
			result[k] = TfValueToGo(attrVal)
		}
		return result
	}
	return nil
}

// IsNullOrUnknown checks if a Terraform value is null or unknown.
func IsNullOrUnknown(v tfattr.Value) bool {
	return v.IsNull() || v.IsUnknown()
}
