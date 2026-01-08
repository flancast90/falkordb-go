// Package proto handles FalkorDB protocol encoding and decoding.
package proto

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildQueryArgs constructs the arguments for a GRAPH.QUERY or GRAPH.RO_QUERY command.
func BuildQueryArgs(cmd, graph, query string, params map[string]interface{}, timeout int, compact bool) []interface{} {
	args := []interface{}{cmd, graph}

	// Build query string with params if provided
	if params != nil && len(params) > 0 {
		paramStr := paramsToString(params)
		query = fmt.Sprintf("CYPHER %s %s", paramStr, query)
	}

	args = append(args, query)

	// Add timeout if specified
	if timeout > 0 {
		args = append(args, "TIMEOUT", strconv.Itoa(timeout))
	}

	// Add compact flag
	if compact {
		args = append(args, "--compact")
	}

	return args
}

// BuildConstraintArgs constructs arguments for constraint commands.
func BuildConstraintArgs(action, graph string, constraintType, entityType, label string, properties []string) []interface{} {
	args := []interface{}{
		"GRAPH.CONSTRAINT",
		action,
		graph,
		constraintType,
		entityType,
		label,
		"PROPERTIES",
		len(properties),
	}
	for _, prop := range properties {
		args = append(args, prop)
	}
	return args
}

// BuildIndexArgs constructs arguments for index creation commands.
func BuildIndexArgs(cmd, graph, indexType, entityType, label string, options map[string]interface{}, properties []string) []interface{} {
	args := []interface{}{cmd, graph}

	// Add index type for typed indices
	if indexType != "" {
		args = append(args, indexType)
	}

	// Add entity type for typed indices
	if entityType != "" {
		args = append(args, entityType)
	}

	args = append(args, label)

	// Add options if provided (for vector indices)
	if options != nil {
		for key, value := range options {
			args = append(args, key, value)
		}
	}

	// Add PROPERTIES marker
	args = append(args, "PROPERTIES", len(properties))

	// Add properties
	for _, prop := range properties {
		args = append(args, prop)
	}

	return args
}

// paramsToString converts query parameters to Cypher parameter string format.
func paramsToString(params map[string]interface{}) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, ValueToString(value)))
	}
	return strings.Join(parts, " ")
}

// ValueToString converts a parameter value to its Cypher string representation.
func ValueToString(param interface{}) string {
	if param == nil {
		return "null"
	}

	switch v := param.(type) {
	case string:
		// Escape quotes and backslashes
		escaped := strings.ReplaceAll(v, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v)
	case float32, float64:
		return fmt.Sprint(v)
	case bool:
		return fmt.Sprint(v)
	case []interface{}:
		var items []string
		for _, item := range v {
			items = append(items, ValueToString(item))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ","))
	case map[string]interface{}:
		var items []string
		for key, value := range v {
			items = append(items, fmt.Sprintf("%s:%s", key, ValueToString(value)))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ","))
	default:
		return fmt.Sprint(v)
	}
}

