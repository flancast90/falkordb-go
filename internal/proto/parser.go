package proto

import (
	"fmt"
	"strconv"
)

// ValueType represents the type of a value in a FalkorDB result.
type ValueType int

const (
	ValueTypeUnknown   ValueType = 0
	ValueTypeNull      ValueType = 1
	ValueTypeString    ValueType = 2
	ValueTypeInteger   ValueType = 3
	ValueTypeBoolean   ValueType = 4
	ValueTypeDouble    ValueType = 5
	ValueTypeArray     ValueType = 6
	ValueTypeEdge      ValueType = 7
	ValueTypeNode      ValueType = 8
	ValueTypePath      ValueType = 9
	ValueTypeMap       ValueType = 10
	ValueTypePoint     ValueType = 11
	ValueTypeVectorF32 ValueType = 12
	ValueTypeDateTime  ValueType = 13
	ValueTypeDate      ValueType = 14
	ValueTypeTime      ValueType = 15
	ValueTypeDuration  ValueType = 16
)

// RawResult represents the raw parsed result from FalkorDB.
type RawResult struct {
	Headers  []interface{}
	Data     []interface{}
	Metadata []string
}

// ParseResult parses the raw Redis reply into a RawResult.
func ParseResult(result interface{}) (*RawResult, error) {
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected query result format: %T", result)
	}

	rr := &RawResult{}

	switch len(arr) {
	case 1:
		// Only metadata (no results)
		metadata, err := toStringSlice(arr[0])
		if err != nil {
			return nil, err
		}
		rr.Metadata = metadata
	case 3:
		// Headers, data, metadata
		if headers, ok := arr[0].([]interface{}); ok {
			rr.Headers = headers
		}
		if data, ok := arr[1].([]interface{}); ok {
			rr.Data = data
		}
		metadata, err := toStringSlice(arr[2])
		if err != nil {
			return nil, err
		}
		rr.Metadata = metadata
	default:
		return nil, fmt.Errorf("unexpected query result length: %d", len(arr))
	}

	return rr, nil
}

// ParseExplainResult parses the raw Redis reply for an EXPLAIN command.
func ParseExplainResult(result interface{}) ([]string, error) {
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected explain result format: %T", result)
	}
	return toStringSlice(arr)
}

// ParseSlowLogResult parses the raw Redis reply for a SLOWLOG command.
func ParseSlowLogResult(result interface{}) ([]map[string]interface{}, error) {
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected slowlog result format: %T", result)
	}

	var entries []map[string]interface{}
	for _, item := range arr {
		entry, ok := item.([]interface{})
		if !ok || len(entry) < 4 {
			continue
		}

		entries = append(entries, map[string]interface{}{
			"timestamp": ToInt64(entry[0]),
			"command":   ToString(entry[1]),
			"query":     ToString(entry[2]),
			"took":      ToFloat64(entry[3]),
		})
	}

	return entries, nil
}

// Helper functions for type conversion

func toStringSlice(v interface{}) ([]string, error) {
	arr, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array, got %T", v)
	}

	result := make([]string, len(arr))
	for i, item := range arr {
		result[i] = ToString(item)
	}
	return result, nil
}

// ToInt converts an interface{} to int.
func ToInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// ToInt64 converts an interface{} to int64.
func ToInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}

// ToFloat64 converts an interface{} to float64.
func ToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// ToString converts an interface{} to string.
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
