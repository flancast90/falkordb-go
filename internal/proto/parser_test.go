package proto

import "testing"

func TestToInt(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int
	}{
		{int(42), 42},
		{int64(42), 42},
		{float64(42.9), 42},
		{"42", 42},
		{nil, 0},
	}

	for _, tc := range tests {
		result := ToInt(tc.input)
		if result != tc.expected {
			t.Errorf("ToInt(%v) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int64
	}{
		{int(42), 42},
		{int64(42), 42},
		{float64(42.9), 42},
		{"42", 42},
		{nil, 0},
	}

	for _, tc := range tests {
		result := ToInt64(tc.input)
		if result != tc.expected {
			t.Errorf("ToInt64(%v) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{float64(42.5), 42.5},
		{int(42), 42.0},
		{int64(42), 42.0},
		{"42.5", 42.5},
		{nil, 0},
	}

	for _, tc := range tests {
		result := ToFloat64(tc.input)
		if result != tc.expected {
			t.Errorf("ToFloat64(%v) = %f, expected %f", tc.input, result, tc.expected)
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{42, "42"},
		{nil, ""},
	}

	for _, tc := range tests {
		result := ToString(tc.input)
		if result != tc.expected {
			t.Errorf("ToString(%v) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestParseResult(t *testing.T) {
	// Test metadata-only result
	result := []interface{}{
		[]interface{}{"Query internal execution time: 0.5 ms"},
	}

	raw, err := ParseResult(result)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}
	if raw.Headers != nil {
		t.Error("Expected nil headers for metadata-only result")
	}
	if len(raw.Metadata) != 1 {
		t.Errorf("Expected 1 metadata entry, got %d", len(raw.Metadata))
	}

	// Test full result
	result = []interface{}{
		[]interface{}{[]interface{}{0, "name"}},
		[]interface{}{[]interface{}{[]interface{}{2, "Alice"}}},
		[]interface{}{"Query internal execution time: 0.5 ms"},
	}

	raw, err = ParseResult(result)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}
	if raw.Headers == nil {
		t.Error("Expected non-nil headers")
	}
	if raw.Data == nil {
		t.Error("Expected non-nil data")
	}
}

func TestParseResultInvalidFormat(t *testing.T) {
	// Invalid type
	_, err := ParseResult("not an array")
	if err == nil {
		t.Error("Expected error for invalid input type")
	}

	// Invalid length
	_, err = ParseResult([]interface{}{1, 2})
	if err == nil {
		t.Error("Expected error for invalid array length")
	}
}

func TestValueTypes(t *testing.T) {
	// Verify value type constants
	types := map[ValueType]int{
		ValueTypeUnknown:   0,
		ValueTypeNull:      1,
		ValueTypeString:    2,
		ValueTypeInteger:   3,
		ValueTypeBoolean:   4,
		ValueTypeDouble:    5,
		ValueTypeArray:     6,
		ValueTypeEdge:      7,
		ValueTypeNode:      8,
		ValueTypePath:      9,
		ValueTypeMap:       10,
		ValueTypePoint:     11,
		ValueTypeVectorF32: 12,
		ValueTypeDateTime:  13,
		ValueTypeDate:      14,
		ValueTypeTime:      15,
		ValueTypeDuration:  16,
	}

	for vt, expected := range types {
		if int(vt) != expected {
			t.Errorf("ValueType %d expected to be %d", vt, expected)
		}
	}
}
