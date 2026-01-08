package proto

import "testing"

func TestValueToString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "null"},
		{"hello", `"hello"`},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{[]interface{}{1, 2, 3}, "[1,2,3]"},
		{map[string]interface{}{"key": "value"}, `{key:"value"}`},
	}

	for _, tc := range tests {
		result := ValueToString(tc.input)
		if result != tc.expected {
			t.Errorf("ValueToString(%v) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestValueToStringEscaping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `"hello"`},
		{`hello "world"`, `"hello \"world\""`},
		{`path\to\file`, `"path\\to\\file"`},
	}

	for _, tc := range tests {
		result := ValueToString(tc.input)
		if result != tc.expected {
			t.Errorf("ValueToString(%q) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestBuildQueryArgs(t *testing.T) {
	// Basic query without params
	args := BuildQueryArgs("GRAPH.QUERY", "myGraph", "MATCH (n) RETURN n", nil, 0, true)

	if len(args) < 3 {
		t.Errorf("Expected at least 3 args, got %d", len(args))
	}
	if args[0] != "GRAPH.QUERY" {
		t.Errorf("Expected first arg GRAPH.QUERY, got %v", args[0])
	}
	if args[1] != "myGraph" {
		t.Errorf("Expected second arg myGraph, got %v", args[1])
	}

	// Query with params
	args = BuildQueryArgs("GRAPH.QUERY", "myGraph", "MATCH (n) RETURN n",
		map[string]interface{}{"name": "test"}, 0, true)

	queryArg := args[2].(string)
	if queryArg[:6] != "CYPHER" {
		t.Errorf("Expected CYPHER prefix, got %s", queryArg[:6])
	}

	// Query with timeout
	args = BuildQueryArgs("GRAPH.QUERY", "myGraph", "MATCH (n) RETURN n", nil, 5000, true)

	foundTimeout := false
	for _, arg := range args {
		if str, ok := arg.(string); ok && str == "TIMEOUT" {
			foundTimeout = true
			break
		}
	}
	if !foundTimeout {
		t.Error("Expected TIMEOUT in args")
	}
}

func TestBuildConstraintArgs(t *testing.T) {
	args := BuildConstraintArgs("CREATE", "myGraph", "UNIQUE", "NODE", "Person", []string{"name", "email"})

	if len(args) < 9 {
		t.Errorf("Expected at least 9 args, got %d", len(args))
	}
	if args[0] != "GRAPH.CONSTRAINT" {
		t.Errorf("Expected GRAPH.CONSTRAINT, got %v", args[0])
	}
	if args[1] != "CREATE" {
		t.Errorf("Expected CREATE, got %v", args[1])
	}
	if args[2] != "myGraph" {
		t.Errorf("Expected myGraph, got %v", args[2])
	}
}

func TestBuildIndexArgs(t *testing.T) {
	args := BuildIndexArgs("GRAPH.CREATE_INDEX", "myGraph", "RANGE", "NODE", "Person", nil, []string{"name"})

	if len(args) < 7 {
		t.Errorf("Expected at least 7 args, got %d", len(args))
	}
	if args[0] != "GRAPH.CREATE_INDEX" {
		t.Errorf("Expected GRAPH.CREATE_INDEX, got %v", args[0])
	}

	// With options (vector index)
	opts := map[string]interface{}{
		"dimension":  128,
		"similarity": "cosine",
	}
	args = BuildIndexArgs("GRAPH.CREATE_INDEX", "myGraph", "VECTOR", "NODE", "Person", opts, []string{"embedding"})

	if len(args) < 10 {
		t.Errorf("Expected at least 10 args for vector index, got %d", len(args))
	}
}

