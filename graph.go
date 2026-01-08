package falkordb

import (
	"context"
	"fmt"
	"sync"

	"github.com/FalkorDB/falkordb-go/internal/proto"
	"github.com/FalkorDB/falkordb-go/internal/redis"
)

// Graph represents a FalkorDB graph and provides methods to interact with it.
// It is safe for concurrent use by multiple goroutines.
type Graph struct {
	name   string
	client redis.Client
	parser *resultParser
	mu     sync.RWMutex
}

// Name returns the name of the graph.
func (g *Graph) Name() string {
	return g.name
}

// Query executes a Cypher query on the graph.
//
// Example:
//
//	result, err := graph.Query(ctx, "CREATE (n:Person {name: $name}) RETURN n",
//		&falkordb.QueryOptions{
//			Params: map[string]interface{}{"name": "Alice"},
//		},
//	)
func (g *Graph) Query(ctx context.Context, query string, options ...*QueryOptions) (*QueryResult, error) {
	return g.execute(ctx, "GRAPH.QUERY", query, options...)
}

// ROQuery executes a read-only Cypher query on the graph.
// Use this for queries that don't modify data to enable query caching
// and replica reads in cluster mode.
func (g *Graph) ROQuery(ctx context.Context, query string, options ...*QueryOptions) (*QueryResult, error) {
	return g.execute(ctx, "GRAPH.RO_QUERY", query, options...)
}

func (g *Graph) execute(ctx context.Context, cmd, query string, options ...*QueryOptions) (*QueryResult, error) {
	var opts *QueryOptions
	if len(options) > 0 {
		opts = options[0]
	}

	var params map[string]interface{}
	var timeout int
	if opts != nil {
		params = opts.Params
		timeout = opts.Timeout
	}

	args := proto.BuildQueryArgs(cmd, g.name, query, params, timeout, true)
	result, err := g.client.Do(ctx, args...).Result()
	if err != nil {
		return nil, err
	}

	// Update metadata cache if needed
	g.updateMetadataFromResult(ctx)

	raw, err := proto.ParseResult(result)
	if err != nil {
		return nil, err
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.parser.parseResult(raw)
}

// Delete removes the graph from the database.
func (g *Graph) Delete(ctx context.Context) error {
	return g.client.Do(ctx, "GRAPH.DELETE", g.name).Err()
}

// Copy creates a copy of the graph with a new name.
func (g *Graph) Copy(ctx context.Context, destGraph string) error {
	return g.client.Do(ctx, "GRAPH.COPY", g.name, destGraph).Err()
}

// Explain returns the execution plan for a query without executing it.
func (g *Graph) Explain(ctx context.Context, query string) ([]string, error) {
	result, err := g.client.Do(ctx, "GRAPH.EXPLAIN", g.name, query).Result()
	if err != nil {
		return nil, err
	}
	return proto.ParseExplainResult(result)
}

// Profile executes a query and returns execution profiling information.
func (g *Graph) Profile(ctx context.Context, query string) ([]string, error) {
	result, err := g.client.Do(ctx, "GRAPH.PROFILE", g.name, query).Result()
	if err != nil {
		return nil, err
	}
	return proto.ParseExplainResult(result)
}

// SlowLog returns the slow query log for this graph.
func (g *Graph) SlowLog(ctx context.Context) ([]SlowLogEntry, error) {
	result, err := g.client.Do(ctx, "GRAPH.SLOWLOG", g.name).Result()
	if err != nil {
		return nil, err
	}

	raw, err := proto.ParseSlowLogResult(result)
	if err != nil {
		return nil, err
	}

	entries := make([]SlowLogEntry, len(raw))
	for i, r := range raw {
		entries[i] = SlowLogEntry{
			Timestamp: proto.ToInt64(r["timestamp"]),
			Command:   proto.ToString(r["command"]),
			Query:     proto.ToString(r["query"]),
			Took:      proto.ToFloat64(r["took"]),
		}
	}
	return entries, nil
}

// SlowLogEntry represents an entry in the slow query log.
type SlowLogEntry struct {
	Timestamp int64
	Command   string
	Query     string
	Took      float64 // Duration in milliseconds
}

// MemoryUsage returns memory usage statistics for the graph.
func (g *Graph) MemoryUsage(ctx context.Context) ([]interface{}, error) {
	result, err := g.client.Do(ctx, "GRAPH.MEMORY", g.name).Result()
	if err != nil {
		return nil, err
	}

	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, nil
}

// === Index Methods ===
// FalkorDB uses Cypher for index creation/deletion

// CreateNodeRangeIndex creates a range index on a node property.
// Example: CREATE INDEX FOR (e:Person) ON (e.name)
func (g *Graph) CreateNodeRangeIndex(ctx context.Context, label string, properties ...string) (*QueryResult, error) {
	return g.createTypedIndex(ctx, "", "NODE", label, nil, properties...)
}

// CreateNodeFulltextIndex creates a fulltext index on a node property.
// Example: CREATE FULLTEXT INDEX FOR (e:Person) ON (e.bio)
func (g *Graph) CreateNodeFulltextIndex(ctx context.Context, label string, properties ...string) (*QueryResult, error) {
	return g.createTypedIndex(ctx, "FULLTEXT", "NODE", label, nil, properties...)
}

// CreateNodeVectorIndex creates a vector index on a node property.
// Example: CREATE VECTOR INDEX FOR (e:Person) ON (e.embedding) OPTIONS {dimension:128, similarityFunction:'euclidean'}
func (g *Graph) CreateNodeVectorIndex(ctx context.Context, label string, dim int, similarity string, properties ...string) (*QueryResult, error) {
	opts := map[string]interface{}{
		"dimension":          dim,
		"similarityFunction": similarity,
	}
	return g.createTypedIndex(ctx, "VECTOR", "NODE", label, opts, properties...)
}

// CreateEdgeRangeIndex creates a range index on an edge property.
func (g *Graph) CreateEdgeRangeIndex(ctx context.Context, label string, properties ...string) (*QueryResult, error) {
	return g.createTypedIndex(ctx, "", "EDGE", label, nil, properties...)
}

// CreateEdgeFulltextIndex creates a fulltext index on an edge property.
func (g *Graph) CreateEdgeFulltextIndex(ctx context.Context, label string, properties ...string) (*QueryResult, error) {
	return g.createTypedIndex(ctx, "FULLTEXT", "EDGE", label, nil, properties...)
}

// CreateEdgeVectorIndex creates a vector index on an edge property.
func (g *Graph) CreateEdgeVectorIndex(ctx context.Context, label string, dim int, similarity string, properties ...string) (*QueryResult, error) {
	opts := map[string]interface{}{
		"dimension":          dim,
		"similarityFunction": similarity,
	}
	return g.createTypedIndex(ctx, "VECTOR", "EDGE", label, opts, properties...)
}

// DropNodeRangeIndex drops a range index from a node property.
func (g *Graph) DropNodeRangeIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "", "NODE", label, property)
}

// DropNodeFulltextIndex drops a fulltext index from a node property.
func (g *Graph) DropNodeFulltextIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "FULLTEXT", "NODE", label, property)
}

// DropNodeVectorIndex drops a vector index from a node property.
func (g *Graph) DropNodeVectorIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "VECTOR", "NODE", label, property)
}

// DropEdgeRangeIndex drops a range index from an edge property.
func (g *Graph) DropEdgeRangeIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "", "EDGE", label, property)
}

// DropEdgeFulltextIndex drops a fulltext index from an edge property.
func (g *Graph) DropEdgeFulltextIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "FULLTEXT", "EDGE", label, property)
}

// DropEdgeVectorIndex drops a vector index from an edge property.
func (g *Graph) DropEdgeVectorIndex(ctx context.Context, label, property string) (*QueryResult, error) {
	return g.dropTypedIndex(ctx, "VECTOR", "EDGE", label, property)
}

// createTypedIndex creates an index using Cypher syntax
func (g *Graph) createTypedIndex(ctx context.Context, indexType, entityType, label string, options map[string]interface{}, properties ...string) (*QueryResult, error) {
	// Build pattern: (e:Label) for nodes, ()-[e:Label]->() for edges
	var pattern string
	if entityType == "NODE" {
		pattern = fmt.Sprintf("(e:%s)", label)
	} else {
		pattern = fmt.Sprintf("()-[e:%s]->()", label)
	}

	// Build property list: e.prop1, e.prop2
	propList := ""
	for i, p := range properties {
		if i > 0 {
			propList += ", "
		}
		propList += "e." + p
	}

	// Build query: CREATE [FULLTEXT|VECTOR] INDEX FOR pattern ON (props) [OPTIONS {...}]
	var query string
	if indexType != "" {
		query = fmt.Sprintf("CREATE %s INDEX FOR %s ON (%s)", indexType, pattern, propList)
	} else {
		query = fmt.Sprintf("CREATE INDEX FOR %s ON (%s)", pattern, propList)
	}

	// Add options for vector index
	if options != nil && len(options) > 0 {
		optStr := ""
		for k, v := range options {
			if optStr != "" {
				optStr += ", "
			}
			switch val := v.(type) {
			case string:
				optStr += fmt.Sprintf("%s:'%s'", k, val)
			default:
				optStr += fmt.Sprintf("%s:%v", k, val)
			}
		}
		query += fmt.Sprintf(" OPTIONS {%s}", optStr)
	}

	return g.Query(ctx, query)
}

// dropTypedIndex drops an index using Cypher syntax
func (g *Graph) dropTypedIndex(ctx context.Context, indexType, entityType, label, property string) (*QueryResult, error) {
	// Build pattern: (e:Label) for nodes, ()-[e:Label]->() for edges
	var pattern string
	if entityType == "NODE" {
		pattern = fmt.Sprintf("(e:%s)", label)
	} else {
		pattern = fmt.Sprintf("()-[e:%s]->()", label)
	}

	// Build query: DROP [FULLTEXT|VECTOR] INDEX FOR pattern ON (e.prop)
	var query string
	if indexType != "" {
		query = fmt.Sprintf("DROP %s INDEX FOR %s ON (e.%s)", indexType, pattern, property)
	} else {
		query = fmt.Sprintf("DROP INDEX FOR %s ON (e.%s)", pattern, property)
	}

	return g.Query(ctx, query)
}

// === Constraint Methods ===

// ConstraintCreate creates a constraint on the graph.
//
// Example:
//
//	// Create unique constraint
//	graph.ConstraintCreate(ctx, falkordb.ConstraintUnique, falkordb.EntityNode, "Person", "email")
//
//	// Create mandatory constraint
//	graph.ConstraintCreate(ctx, falkordb.ConstraintMandatory, falkordb.EntityNode, "Person", "name")
func (g *Graph) ConstraintCreate(ctx context.Context, constraintType ConstraintType, entityType EntityType, label string, properties ...string) error {
	args := proto.BuildConstraintArgs("CREATE", g.name, string(constraintType), string(entityType), label, properties)
	return g.client.Do(ctx, args...).Err()
}

// ConstraintDrop removes a constraint from the graph.
func (g *Graph) ConstraintDrop(ctx context.Context, constraintType ConstraintType, entityType EntityType, label string, properties ...string) error {
	args := proto.BuildConstraintArgs("DROP", g.name, string(constraintType), string(entityType), label, properties)
	return g.client.Do(ctx, args...).Err()
}

// updateMetadataFromResult fetches and caches graph metadata (labels, types, property keys).
func (g *Graph) updateMetadataFromResult(ctx context.Context) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Fetch labels
	if result, err := g.client.Do(ctx, "GRAPH.RO_QUERY", g.name, "CALL db.labels()", "--compact").Result(); err == nil {
		if labels := extractStringList(result); labels != nil {
			g.parser.labels = labels
		}
	}

	// Fetch relationship types
	if result, err := g.client.Do(ctx, "GRAPH.RO_QUERY", g.name, "CALL db.relationshipTypes()", "--compact").Result(); err == nil {
		if types := extractStringList(result); types != nil {
			g.parser.relTypes = types
		}
	}

	// Fetch property keys
	if result, err := g.client.Do(ctx, "GRAPH.RO_QUERY", g.name, "CALL db.propertyKeys()", "--compact").Result(); err == nil {
		if keys := extractStringList(result); keys != nil {
			g.parser.propertyKeys = keys
		}
	}
}

func extractStringList(result interface{}) []string {
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 2 {
		return nil
	}

	data, ok := arr[1].([]interface{})
	if !ok {
		return nil
	}

	var strings []string
	for _, row := range data {
		if rowArr, ok := row.([]interface{}); ok && len(rowArr) > 0 {
			if cellArr, ok := rowArr[0].([]interface{}); ok && len(cellArr) >= 2 {
				if s, ok := cellArr[1].(string); ok {
					strings = append(strings, s)
				}
			}
		}
	}
	return strings
}

// String returns a string representation of the graph.
func (g *Graph) String() string {
	return fmt.Sprintf("Graph<%s>", g.name)
}
