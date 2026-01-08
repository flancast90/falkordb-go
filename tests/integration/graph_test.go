// Package integration contains end-to-end tests for the FalkorDB Go client.
// These tests require a running FalkorDB instance.
//
// Run with: go test -v ./tests/integration/...
// Or use: make test-integration
package integration

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/flancast90/falkordb-go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomName() string {
	return fmt.Sprintf("test_%d", rand.Intn(999999))
}

func newTestDB(t *testing.T) *falkordb.FalkorDB {
	t.Helper()

	host := os.Getenv("FALKORDB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("FALKORDB_PORT")
	if port == "" {
		port = "6379"
	}

	ctx := context.Background()
	db, err := falkordb.Connect(ctx, &falkordb.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})
	if err != nil {
		t.Skipf("FalkorDB not available at %s:%s: %v", host, port, err)
	}

	return db
}

// =============================================================================
// Connection Tests
// =============================================================================

func TestConnection(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("Ping", func(t *testing.T) {
		if err := db.Ping(ctx); err != nil {
			t.Errorf("Ping failed: %v", err)
		}
	})

	t.Run("Info", func(t *testing.T) {
		info, err := db.Info(ctx)
		if err != nil {
			t.Fatalf("Info failed: %v", err)
		}
		if info == "" {
			t.Error("Expected non-empty info")
		}
	})

	t.Run("InfoWithSection", func(t *testing.T) {
		info, err := db.Info(ctx, "server")
		if err != nil {
			t.Fatalf("Info with section failed: %v", err)
		}
		if !strings.Contains(info, "redis_version") {
			t.Error("Expected redis_version in server info")
		}
	})

	t.Run("List", func(t *testing.T) {
		graphs, err := db.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if graphs == nil {
			t.Error("Expected non-nil graph list")
		}
	})
}

// =============================================================================
// Configuration Tests
// =============================================================================

func TestConfig(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("GetConfig", func(t *testing.T) {
		value, err := db.ConfigGet(ctx, "RESULTSET_SIZE")
		if err != nil {
			t.Fatalf("ConfigGet failed: %v", err)
		}
		t.Logf("RESULTSET_SIZE = %v", value)
	})

	t.Run("SetConfig", func(t *testing.T) {
		// Get original
		original, err := db.ConfigGet(ctx, "RESULTSET_SIZE")
		if err != nil {
			t.Fatalf("ConfigGet failed: %v", err)
		}

		// Set new value
		err = db.ConfigSet(ctx, "RESULTSET_SIZE", 5000)
		if err != nil {
			t.Fatalf("ConfigSet failed: %v", err)
		}

		// Verify
		newVal, err := db.ConfigGet(ctx, "RESULTSET_SIZE")
		if err != nil {
			t.Fatalf("ConfigGet after set failed: %v", err)
		}
		t.Logf("New RESULTSET_SIZE = %v", newVal)

		// Restore
		if original != nil {
			db.ConfigSet(ctx, "RESULTSET_SIZE", original)
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		_, err := db.ConfigGet(ctx, "INVALID_KEY_THAT_DOES_NOT_EXIST")
		if err == nil {
			t.Error("Expected error for invalid config key")
		}
	})
}

// =============================================================================
// Basic Query Tests
// =============================================================================

func TestBasicQueries(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	t.Run("CreateNode", func(t *testing.T) {
		result, err := graph.Query(ctx, "CREATE (n:Person {name: 'Alice', age: 30}) RETURN n")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if len(result.Data) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result.Data))
		}
	})

	t.Run("MatchNode", func(t *testing.T) {
		result, err := graph.Query(ctx, "MATCH (n:Person) RETURN n.name, n.age")
		if err != nil {
			t.Fatalf("Match failed: %v", err)
		}
		if len(result.Data) != 1 {
			t.Fatalf("Expected 1 row, got %d", len(result.Data))
		}
		if result.Data[0]["n.name"] != "Alice" {
			t.Errorf("Expected 'Alice', got %v", result.Data[0]["n.name"])
		}
	})

	t.Run("ROQuery", func(t *testing.T) {
		result, err := graph.ROQuery(ctx, "MATCH (n:Person) RETURN n.name")
		if err != nil {
			t.Fatalf("ROQuery failed: %v", err)
		}
		if len(result.Data) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result.Data))
		}
	})

	t.Run("ROQueryWriteFails", func(t *testing.T) {
		_, err := graph.ROQuery(ctx, "CREATE (n:Test)")
		if err == nil {
			t.Error("Expected error for write in ROQuery")
		}
	})

	t.Run("InvalidSyntax", func(t *testing.T) {
		_, err := graph.Query(ctx, "THIS IS NOT CYPHER")
		if err == nil {
			t.Error("Expected error for invalid syntax")
		}
	})
}

// =============================================================================
// Query Parameters Tests
// =============================================================================

func TestQueryParameters(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	t.Run("StringParam", func(t *testing.T) {
		result, err := graph.Query(ctx,
			"CREATE (n:Person {name: $name}) RETURN n.name",
			&falkordb.QueryOptions{
				Params: map[string]interface{}{"name": "Bob"},
			},
		)
		if err != nil {
			t.Fatalf("Query with string param failed: %v", err)
		}
		if result.Data[0]["n.name"] != "Bob" {
			t.Errorf("Expected 'Bob', got %v", result.Data[0]["n.name"])
		}
	})

	t.Run("IntParam", func(t *testing.T) {
		result, err := graph.Query(ctx,
			"CREATE (n:Person {age: $age}) RETURN n.age",
			&falkordb.QueryOptions{
				Params: map[string]interface{}{"age": 25},
			},
		)
		if err != nil {
			t.Fatalf("Query with int param failed: %v", err)
		}
		// Age might be returned as int64
		age := result.Data[0]["n.age"]
		if age != int64(25) && age != 25 {
			t.Errorf("Expected 25, got %v (%T)", age, age)
		}
	})

	t.Run("MultipleParams", func(t *testing.T) {
		result, err := graph.Query(ctx,
			"CREATE (n:Person {name: $name, age: $age, active: $active}) RETURN n",
			&falkordb.QueryOptions{
				Params: map[string]interface{}{
					"name":   "Charlie",
					"age":    35,
					"active": true,
				},
			},
		)
		if err != nil {
			t.Fatalf("Query with multiple params failed: %v", err)
		}
		if len(result.Data) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result.Data))
		}
	})

	t.Run("ArrayParam", func(t *testing.T) {
		result, err := graph.Query(ctx,
			"CREATE (n:Person {hobbies: $hobbies}) RETURN n.hobbies",
			&falkordb.QueryOptions{
				Params: map[string]interface{}{
					"hobbies": []interface{}{"reading", "coding", "hiking"},
				},
			},
		)
		if err != nil {
			t.Fatalf("Query with array param failed: %v", err)
		}
		hobbies, ok := result.Data[0]["n.hobbies"].([]interface{})
		if !ok {
			t.Fatalf("Expected array, got %T", result.Data[0]["n.hobbies"])
		}
		if len(hobbies) != 3 {
			t.Errorf("Expected 3 hobbies, got %d", len(hobbies))
		}
	})
}

// =============================================================================
// Node and Edge Tests
// =============================================================================

func TestNodesAndEdges(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	t.Run("CreateRelationship", func(t *testing.T) {
		result, err := graph.Query(ctx, `
			CREATE (a:Person {name: 'Alice'})-[r:KNOWS {since: 2020}]->(b:Person {name: 'Bob'})
			RETURN a, r, b
		`)
		if err != nil {
			t.Fatalf("Create relationship failed: %v", err)
		}

		row := result.Data[0]

		// Check source node
		alice, ok := row["a"].(*falkordb.Node)
		if !ok {
			t.Fatalf("Expected Node, got %T", row["a"])
		}
		if alice.Properties["name"] != "Alice" {
			t.Errorf("Expected 'Alice', got %v", alice.Properties["name"])
		}

		// Check relationship
		knows, ok := row["r"].(*falkordb.Edge)
		if !ok {
			t.Fatalf("Expected Edge, got %T", row["r"])
		}
		if knows.RelationshipType != "KNOWS" {
			t.Errorf("Expected 'KNOWS', got %s", knows.RelationshipType)
		}
		if knows.Properties["since"] != int64(2020) {
			t.Errorf("Expected 2020, got %v", knows.Properties["since"])
		}

		// Check destination node
		bob, ok := row["b"].(*falkordb.Node)
		if !ok {
			t.Fatalf("Expected Node, got %T", row["b"])
		}
		if bob.Properties["name"] != "Bob" {
			t.Errorf("Expected 'Bob', got %v", bob.Properties["name"])
		}

		// Check edge connects correct nodes
		if knows.SourceID != alice.ID {
			t.Errorf("Edge source ID mismatch")
		}
		if knows.DestinationID != bob.ID {
			t.Errorf("Edge destination ID mismatch")
		}
	})

	t.Run("MultipleLabels", func(t *testing.T) {
		result, err := graph.Query(ctx, "CREATE (n:Person:Employee:Manager {name: 'Carol'}) RETURN n")
		if err != nil {
			t.Fatalf("Create multi-label node failed: %v", err)
		}

		node, ok := result.Data[0]["n"].(*falkordb.Node)
		if !ok {
			t.Fatalf("Expected Node, got %T", result.Data[0]["n"])
		}
		if len(node.Labels) < 2 {
			t.Errorf("Expected multiple labels, got %d", len(node.Labels))
		}
	})
}

// =============================================================================
// Path Tests
// =============================================================================

func TestPaths(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	// Create a chain
	_, err := graph.Query(ctx, `
		CREATE (a:Person {name: 'Alice'})
		CREATE (b:Person {name: 'Bob'})
		CREATE (c:Person {name: 'Charlie'})
		CREATE (a)-[:KNOWS]->(b)-[:KNOWS]->(c)
	`)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("SingleHopPath", func(t *testing.T) {
		result, err := graph.Query(ctx, `
			MATCH p = (a:Person {name: 'Alice'})-[:KNOWS]->(b)
			RETURN p
		`)
		if err != nil {
			t.Fatalf("Path query failed: %v", err)
		}

		if len(result.Data) == 0 {
			t.Skip("No path results")
		}

		path, ok := result.Data[0]["p"].(*falkordb.Path)
		if !ok {
			t.Skipf("Expected Path, got %T", result.Data[0]["p"])
		}

		if len(path.Nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(path.Nodes))
		}
		if len(path.Edges) != 1 {
			t.Errorf("Expected 1 edge, got %d", len(path.Edges))
		}
	})

	t.Run("VariableLengthPath", func(t *testing.T) {
		result, err := graph.Query(ctx, `
			MATCH p = (a:Person {name: 'Alice'})-[:KNOWS*]->(c:Person {name: 'Charlie'})
			RETURN p
		`)
		if err != nil {
			t.Fatalf("Variable path query failed: %v", err)
		}

		if len(result.Data) == 0 {
			t.Skip("No variable path results")
		}

		path, ok := result.Data[0]["p"].(*falkordb.Path)
		if !ok {
			t.Skipf("Expected Path, got %T", result.Data[0]["p"])
		}

		if path.Length() < 1 {
			t.Errorf("Expected path length >= 1, got %d", path.Length())
		}
	})
}

// =============================================================================
// Graph Operations Tests
// =============================================================================

func TestGraphOperations(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("CopyGraph", func(t *testing.T) {
		srcName := randomName()
		dstName := randomName()

		srcGraph := db.SelectGraph(srcName)
		defer srcGraph.Delete(ctx)

		// Create data
		_, err := srcGraph.Query(ctx, "CREATE (n:Test {value: 42})")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Copy
		err = srcGraph.Copy(ctx, dstName)
		if err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		dstGraph := db.SelectGraph(dstName)
		defer dstGraph.Delete(ctx)

		// Verify copy
		result, err := dstGraph.ROQuery(ctx, "MATCH (n:Test) RETURN n.value")
		if err != nil {
			t.Fatalf("Query copy failed: %v", err)
		}
		if len(result.Data) != 1 {
			t.Errorf("Expected 1 row in copy, got %d", len(result.Data))
		}
	})

	t.Run("DeleteGraph", func(t *testing.T) {
		name := randomName()
		graph := db.SelectGraph(name)

		_, err := graph.Query(ctx, "CREATE (n:Test)")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Verify exists
		graphs, _ := db.List(ctx)
		found := false
		for _, g := range graphs {
			if g == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Graph should exist before delete")
		}

		// Delete
		err = graph.Delete(ctx)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify gone
		graphs, _ = db.List(ctx)
		for _, g := range graphs {
			if g == name {
				t.Error("Graph should not exist after delete")
			}
		}
	})
}

// =============================================================================
// Query Analysis Tests
// =============================================================================

func TestQueryAnalysis(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	// Setup
	_, _ = graph.Query(ctx, "CREATE (:Person {name: 'Alice'})")

	t.Run("Explain", func(t *testing.T) {
		plan, err := graph.Explain(ctx, "MATCH (n:Person) WHERE n.name = 'Alice' RETURN n")
		if err != nil {
			t.Fatalf("Explain failed: %v", err)
		}

		if len(plan) == 0 {
			t.Error("Expected non-empty plan")
		}

		planStr := strings.Join(plan, "\n")
		if !strings.Contains(planStr, "Results") {
			t.Error("Expected 'Results' in plan")
		}
		t.Logf("Execution plan:\n%s", planStr)
	})

	t.Run("Profile", func(t *testing.T) {
		profile, err := graph.Profile(ctx, "MATCH (n:Person) RETURN n")
		if err != nil {
			t.Fatalf("Profile failed: %v", err)
		}

		if len(profile) == 0 {
			t.Error("Expected non-empty profile")
		}

		profileStr := strings.Join(profile, "\n")
		if !strings.Contains(profileStr, "Records produced") {
			t.Error("Expected 'Records produced' in profile")
		}
		t.Logf("Profile:\n%s", profileStr)
	})

	t.Run("SlowLog", func(t *testing.T) {
		// Run a slow query
		_, _ = graph.Query(ctx, "UNWIND range(0, 100000) AS x RETURN max(x)")

		entries, err := graph.SlowLog(ctx)
		if err != nil {
			t.Fatalf("SlowLog failed: %v", err)
		}

		t.Logf("Slow log entries: %d", len(entries))
		for i, entry := range entries {
			if i >= 3 {
				break
			}
			t.Logf("  [%d] %s: %s (%.2fms)", entry.Timestamp, entry.Command, entry.Query, entry.Took)
		}
	})
}

// =============================================================================
// Index Tests
// =============================================================================

func TestIndices(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	// Setup
	_, _ = graph.Query(ctx, "CREATE (:Person {name: 'Alice', bio: 'Software engineer'})")

	t.Run("CreateRangeIndex", func(t *testing.T) {
		_, err := graph.CreateNodeRangeIndex(ctx, "Person", "name")
		if err != nil {
			t.Fatalf("CreateNodeRangeIndex failed: %v", err)
		}

		// Verify
		result, _ := graph.ROQuery(ctx, "CALL db.indexes()")
		if len(result.Data) == 0 {
			t.Error("Expected index to exist")
		}
	})

	t.Run("CreateFulltextIndex", func(t *testing.T) {
		_, err := graph.CreateNodeFulltextIndex(ctx, "Person", "bio")
		if err != nil {
			t.Fatalf("CreateNodeFulltextIndex failed: %v", err)
		}
	})

	t.Run("CreateVectorIndex", func(t *testing.T) {
		_, err := graph.CreateNodeVectorIndex(ctx, "Person", 128, "euclidean", "embedding")
		if err != nil {
			t.Logf("CreateNodeVectorIndex: %v (may not be supported)", err)
		}
	})

	t.Run("DropIndex", func(t *testing.T) {
		_, err := graph.DropNodeRangeIndex(ctx, "Person", "name")
		if err != nil {
			t.Fatalf("DropNodeRangeIndex failed: %v", err)
		}
	})

	t.Run("DuplicateIndexFails", func(t *testing.T) {
		graph.CreateNodeRangeIndex(ctx, "Person", "name")
		_, err := graph.CreateNodeRangeIndex(ctx, "Person", "name")
		if err == nil {
			t.Error("Expected error for duplicate index")
		}
	})
}

// =============================================================================
// Constraint Tests
// =============================================================================

func TestConstraints(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	// Setup
	_, _ = graph.Query(ctx, "CREATE (:Person {name: 'Alice', email: 'alice@example.com'})")
	_, _ = graph.Query(ctx, "CREATE INDEX ON :Person(email)")

	t.Run("CreateUniqueConstraint", func(t *testing.T) {
		err := graph.ConstraintCreate(ctx, falkordb.ConstraintUnique, falkordb.EntityNode, "Person", "email")
		if err != nil {
			t.Fatalf("ConstraintCreate failed: %v", err)
		}

		// Verify
		result, _ := graph.ROQuery(ctx, "CALL db.constraints()")
		if len(result.Data) == 0 {
			t.Error("Expected constraint to exist")
		}
	})

	t.Run("CreateMandatoryConstraint", func(t *testing.T) {
		err := graph.ConstraintCreate(ctx, falkordb.ConstraintMandatory, falkordb.EntityNode, "Person", "name")
		if err != nil {
			t.Fatalf("ConstraintCreate mandatory failed: %v", err)
		}
	})

	t.Run("DropConstraint", func(t *testing.T) {
		err := graph.ConstraintDrop(ctx, falkordb.ConstraintUnique, falkordb.EntityNode, "Person", "email")
		if err != nil {
			t.Fatalf("ConstraintDrop failed: %v", err)
		}
	})

	t.Run("DuplicateConstraintFails", func(t *testing.T) {
		graph.ConstraintCreate(ctx, falkordb.ConstraintMandatory, falkordb.EntityNode, "Person", "name")
		err := graph.ConstraintCreate(ctx, falkordb.ConstraintMandatory, falkordb.EntityNode, "Person", "name")
		if err == nil {
			t.Error("Expected error for duplicate constraint")
		}
	})
}

// =============================================================================
// Data Type Tests
// =============================================================================

func TestDataTypes(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	t.Run("Point", func(t *testing.T) {
		result, err := graph.Query(ctx, "RETURN point({latitude: 40.7128, longitude: -74.0060}) AS p")
		if err != nil {
			t.Fatalf("Point query failed: %v", err)
		}

		point, ok := result.Data[0]["p"].(*falkordb.Point)
		if !ok {
			t.Skipf("Expected Point, got %T", result.Data[0]["p"])
		}

		if point.Latitude < 40.7 || point.Latitude > 40.8 {
			t.Errorf("Unexpected latitude: %f", point.Latitude)
		}
	})

	t.Run("Map", func(t *testing.T) {
		result, err := graph.Query(ctx, "RETURN {name: 'Alice', age: 30, nested: {city: 'NYC'}} AS m")
		if err != nil {
			t.Fatalf("Map query failed: %v", err)
		}

		m, ok := result.Data[0]["m"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", result.Data[0]["m"])
		}

		if m["name"] != "Alice" {
			t.Errorf("Expected 'Alice', got %v", m["name"])
		}
	})

	t.Run("Array", func(t *testing.T) {
		result, err := graph.Query(ctx, "RETURN [1, 2, 3, 'four', true] AS arr")
		if err != nil {
			t.Fatalf("Array query failed: %v", err)
		}

		arr, ok := result.Data[0]["arr"].([]interface{})
		if !ok {
			t.Fatalf("Expected array, got %T", result.Data[0]["arr"])
		}

		if len(arr) != 5 {
			t.Errorf("Expected 5 elements, got %d", len(arr))
		}
	})

	t.Run("NullHandling", func(t *testing.T) {
		result, err := graph.Query(ctx, "RETURN null AS n")
		if err != nil {
			t.Fatalf("Null query failed: %v", err)
		}

		if result.Data[0]["n"] != nil {
			t.Errorf("Expected nil, got %v", result.Data[0]["n"])
		}
	})
}

// =============================================================================
// Concurrent Operations Tests
// =============================================================================

func TestConcurrency(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()
	graph := db.SelectGraph(randomName())
	defer graph.Delete(ctx)

	t.Run("ParallelQueries", func(t *testing.T) {
		// Create some data first
		_, _ = graph.Query(ctx, "UNWIND range(1, 100) AS i CREATE (:Node {id: i})")

		// Run parallel reads
		done := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				_, err := graph.ROQuery(ctx, fmt.Sprintf("MATCH (n:Node {id: %d}) RETURN n", (id%100)+1))
				done <- err
			}(i)
		}

		for i := 0; i < 10; i++ {
			if err := <-done; err != nil {
				t.Errorf("Parallel query failed: %v", err)
			}
		}
	})

	t.Run("ParallelWrites", func(t *testing.T) {
		done := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func(id int) {
				_, err := graph.Query(ctx, fmt.Sprintf("CREATE (:Concurrent {id: %d})", id))
				done <- err
			}(i)
		}

		for i := 0; i < 5; i++ {
			if err := <-done; err != nil {
				t.Errorf("Parallel write failed: %v", err)
			}
		}

		// Verify all created
		result, err := graph.ROQuery(ctx, "MATCH (n:Concurrent) RETURN count(n) AS c")
		if err != nil {
			t.Fatalf("Count query failed: %v", err)
		}
		if len(result.Data) == 0 {
			t.Fatal("No data returned from count query")
		}
		count := result.Data[0]["c"]
		if count != int64(5) {
			t.Errorf("Expected 5 nodes, got %v", count)
		}
	})
}

