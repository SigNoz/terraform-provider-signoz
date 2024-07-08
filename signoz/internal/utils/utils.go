package utils

import (
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
