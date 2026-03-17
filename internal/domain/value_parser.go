package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ParseValue converts a string value from the database to the appropriate Go type
// based on the valueType parameter
func ParseValue(rawValue string, valueType string) (interface{}, error) {
	switch valueType {
	case "int":
		return strconv.Atoi(rawValue)
	case "float":
		return strconv.ParseFloat(rawValue, 64)
	case "bool":
		return strconv.ParseBool(rawValue)
	case "json":
		var data interface{}
		if err := json.Unmarshal([]byte(rawValue), &data); err != nil {
			return nil, fmt.Errorf("failed to parse json: %w", err)
		}
		return data, nil
	case "string":
		return rawValue, nil
	default:
		// Default to string if unknown type
		return rawValue, nil
	}
}

// SerializeValue converts a typed value to a string for database storage
func SerializeValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		return strconv.FormatBool(v), nil
	case map[string]interface{}, []interface{}, map[interface{}]interface{}:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to serialize json: %w", err)
		}
		return string(bytes), nil
	default:
		// Try JSON serialization as fallback
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v), nil // Last resort: use string representation
		}
		return string(bytes), nil
	}
}

// InferValueType attempts to determine the type of a value
func InferValueType(value interface{}) string {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	case string:
		return "string"
	case map[string]interface{}, []interface{}:
		return "json"
	default:
		return "string"
	}
}
