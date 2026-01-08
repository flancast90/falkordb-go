package falkordb

import (
	"fmt"

	"github.com/flancast90/falkordb-go/internal/proto"
)

// QueryResult represents the result of a Cypher query.
type QueryResult struct {
	// Headers contains the column names/definitions from the query.
	Headers []Header

	// Data contains the result rows as maps of column name to value.
	// Values can be: string, int64, float64, bool, nil, *Node, *Edge, *Path, *Point, map, slice
	Data []map[string]interface{}

	// Metadata contains query execution statistics.
	Metadata []string
}

// Header represents a column header in the query result.
type Header struct {
	Type int
	Name string
}

// resultParser handles parsing of FalkorDB results into Go types.
type resultParser struct {
	labels       []string
	relTypes     []string
	propertyKeys []string
}

func newResultParser() *resultParser {
	return &resultParser{}
}

// parseResult converts a raw FalkorDB result into a QueryResult.
func (p *resultParser) parseResult(raw *proto.RawResult) (*QueryResult, error) {
	result := &QueryResult{
		Metadata: raw.Metadata,
	}

	// Parse headers
	if raw.Headers != nil {
		result.Headers = make([]Header, len(raw.Headers))
		for i, h := range raw.Headers {
			if header, ok := h.([]interface{}); ok && len(header) >= 2 {
				result.Headers[i] = Header{
					Type: proto.ToInt(header[0]),
					Name: proto.ToString(header[1]),
				}
			}
		}
	}

	// Parse data rows
	if raw.Data != nil {
		result.Data = make([]map[string]interface{}, len(raw.Data))
		for i, row := range raw.Data {
			if rowData, ok := row.([]interface{}); ok {
				result.Data[i] = p.parseRow(rowData, result.Headers)
			}
		}
	}

	return result, nil
}

func (p *resultParser) parseRow(row []interface{}, headers []Header) map[string]interface{} {
	result := make(map[string]interface{})

	for i, cell := range row {
		var name string
		if i < len(headers) {
			name = headers[i].Name
		} else {
			name = fmt.Sprintf("column_%d", i)
		}

		if cellData, ok := cell.([]interface{}); ok && len(cellData) >= 2 {
			valueType := proto.ValueType(proto.ToInt(cellData[0]))
			result[name] = p.parseValue(valueType, cellData[1])
		} else {
			result[name] = cell
		}
	}

	return result
}

func (p *resultParser) parseValue(valueType proto.ValueType, value interface{}) interface{} {
	switch valueType {
	case proto.ValueTypeNull:
		return nil
	case proto.ValueTypeString:
		return proto.ToString(value)
	case proto.ValueTypeInteger:
		return proto.ToInt64(value)
	case proto.ValueTypeBoolean:
		if b, ok := value.(bool); ok {
			return b
		}
		return proto.ToString(value) == "true"
	case proto.ValueTypeDouble:
		return proto.ToFloat64(value)
	case proto.ValueTypeArray:
		return p.parseArray(value)
	case proto.ValueTypeNode:
		return p.parseNode(value)
	case proto.ValueTypeEdge:
		return p.parseEdge(value)
	case proto.ValueTypePath:
		return p.parsePath(value)
	case proto.ValueTypeMap:
		return p.parseMap(value)
	case proto.ValueTypePoint:
		return p.parsePoint(value)
	default:
		return value
	}
}

func (p *resultParser) parseArray(value interface{}) []interface{} {
	arr, ok := value.([]interface{})
	if !ok {
		return nil
	}

	result := make([]interface{}, len(arr))
	for i, item := range arr {
		if itemData, ok := item.([]interface{}); ok && len(itemData) >= 2 {
			valueType := proto.ValueType(proto.ToInt(itemData[0]))
			result[i] = p.parseValue(valueType, itemData[1])
		} else {
			result[i] = item
		}
	}
	return result
}

func (p *resultParser) parseNode(value interface{}) *Node {
	arr, ok := value.([]interface{})
	if !ok || len(arr) < 3 {
		return nil
	}

	node := &Node{
		ID:         proto.ToInt64(arr[0]),
		Properties: make(map[string]interface{}),
	}

	// Parse labels
	if labels, ok := arr[1].([]interface{}); ok {
		for _, l := range labels {
			labelIdx := proto.ToInt(l)
			if labelIdx < len(p.labels) {
				node.Labels = append(node.Labels, p.labels[labelIdx])
			} else {
				node.Labels = append(node.Labels, fmt.Sprintf("label_%d", labelIdx))
			}
		}
	}

	// Parse properties
	if props, ok := arr[2].([]interface{}); ok {
		node.Properties = p.parseProperties(props)
	}

	return node
}

func (p *resultParser) parseEdge(value interface{}) *Edge {
	arr, ok := value.([]interface{})
	if !ok || len(arr) < 5 {
		return nil
	}

	edge := &Edge{
		ID:            proto.ToInt64(arr[0]),
		SourceID:      proto.ToInt64(arr[2]),
		DestinationID: proto.ToInt64(arr[3]),
		Properties:    make(map[string]interface{}),
	}

	// Parse relationship type
	relTypeIdx := proto.ToInt(arr[1])
	if relTypeIdx < len(p.relTypes) {
		edge.RelationshipType = p.relTypes[relTypeIdx]
	} else {
		edge.RelationshipType = fmt.Sprintf("type_%d", relTypeIdx)
	}

	// Parse properties
	if props, ok := arr[4].([]interface{}); ok {
		edge.Properties = p.parseProperties(props)
	}

	return edge
}

func (p *resultParser) parsePath(value interface{}) *Path {
	// Path structure: [[ArrayType, [nodes...]], [ArrayType, [edges...]]]
	arr, ok := value.([]interface{})
	if !ok || len(arr) < 2 {
		return nil
	}

	path := &Path{
		Nodes: p.parsePathNodes(arr[0]),
		Edges: p.parsePathEdges(arr[1]),
	}
	return path
}

// parsePathNodes extracts nodes from a path wrapper: [ArrayType, [node1, node2, ...]]
func (p *resultParser) parsePathNodes(wrapper interface{}) []*Node {
	items := unwrapArray(wrapper)
	if items == nil {
		return nil
	}

	var nodes []*Node
	for _, item := range items {
		if data := unwrapTypedValue(item); data != nil {
			if node := p.parseNode(data); node != nil {
				nodes = append(nodes, node)
			}
		}
	}
	return nodes
}

// parsePathEdges extracts edges from a path wrapper: [ArrayType, [edge1, edge2, ...]]
func (p *resultParser) parsePathEdges(wrapper interface{}) []*Edge {
	items := unwrapArray(wrapper)
	if items == nil {
		return nil
	}

	var edges []*Edge
	for _, item := range items {
		if data := unwrapTypedValue(item); data != nil {
			if edge := p.parseEdge(data); edge != nil {
				edges = append(edges, edge)
			}
		}
	}
	return edges
}

// unwrapArray extracts the inner array from [ArrayType, [...items]]
func unwrapArray(wrapper interface{}) []interface{} {
	arr, ok := wrapper.([]interface{})
	if !ok || len(arr) < 2 {
		return nil
	}
	items, _ := arr[1].([]interface{})
	return items
}

// unwrapTypedValue extracts data from [Type, data]
func unwrapTypedValue(typed interface{}) interface{} {
	arr, ok := typed.([]interface{})
	if !ok || len(arr) < 2 {
		return nil
	}
	return arr[1]
}

func (p *resultParser) parseMap(value interface{}) map[string]interface{} {
	arr, ok := value.([]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]interface{})
	for i := 0; i < len(arr)-1; i += 2 {
		key := proto.ToString(arr[i])
		if valArr, ok := arr[i+1].([]interface{}); ok && len(valArr) >= 2 {
			valueType := proto.ValueType(proto.ToInt(valArr[0]))
			result[key] = p.parseValue(valueType, valArr[1])
		} else {
			result[key] = arr[i+1]
		}
	}
	return result
}

func (p *resultParser) parsePoint(value interface{}) *Point {
	arr, ok := value.([]interface{})
	if !ok || len(arr) < 2 {
		return nil
	}

	return &Point{
		Latitude:  proto.ToFloat64(arr[0]),
		Longitude: proto.ToFloat64(arr[1]),
	}
}

func (p *resultParser) parseProperties(props []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, prop := range props {
		propArr, ok := prop.([]interface{})
		if !ok || len(propArr) < 3 {
			continue
		}

		keyIdx := proto.ToInt(propArr[0])
		var key string
		if keyIdx < len(p.propertyKeys) {
			key = p.propertyKeys[keyIdx]
		} else {
			key = fmt.Sprintf("prop_%d", keyIdx)
		}

		valueType := proto.ValueType(proto.ToInt(propArr[1]))
		result[key] = p.parseValue(valueType, propArr[2])
	}

	return result
}

// updateMetadata updates the parser's cached metadata from query results.
func (p *resultParser) updateMetadata(labels, relTypes, propertyKeys []string) {
	if labels != nil {
		p.labels = labels
	}
	if relTypes != nil {
		p.relTypes = relTypes
	}
	if propertyKeys != nil {
		p.propertyKeys = propertyKeys
	}
}
