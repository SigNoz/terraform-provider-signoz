package utils

import "encoding/json"

// FlattenJsonToString compacts a JSON object into a string; an empty object
// flattens to "". Replaces the SDKv1 helper/structure function of the same
// name, which upstream no longer depends on.
func FlattenJsonToString(input map[string]interface{}) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	result, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// ExpandJsonFromString unmarshals a JSON string into a generic object.
// Replaces the SDKv1 helper/structure function of the same name.
func ExpandJsonFromString(jsonString string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := json.Unmarshal([]byte(jsonString), &result)

	return result, err
}
