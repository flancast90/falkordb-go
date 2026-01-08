package falkordb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/flancast90/falkordb-go"
)

func Example() {
	ctx := context.Background()

	// Connect to FalkorDB
	db, err := falkordb.Connect(ctx, &falkordb.Options{
		Addr: "localhost:6379",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Select a graph
	graph := db.SelectGraph("social")

	// Create nodes and relationships
	_, err = graph.Query(ctx, `
		CREATE (alice:Person {name: 'Alice', age: 30})
		CREATE (bob:Person {name: 'Bob', age: 25})
		CREATE (alice)-[:KNOWS]->(bob)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Query the graph
	result, err := graph.Query(ctx, "MATCH (p:Person) RETURN p.name, p.age")
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result.Data {
		fmt.Printf("%s is %v years old\n", row["p.name"], row["p.age"])
	}

	// Clean up
	graph.Delete(ctx)
}

func ExampleGraph_Query_withParams() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	graph := db.SelectGraph("example")

	// Query with parameters
	result, _ := graph.Query(ctx,
		"MATCH (p:Person) WHERE p.age > $minAge RETURN p.name",
		&falkordb.QueryOptions{
			Params: map[string]interface{}{
				"minAge": 21,
			},
		},
	)

	for _, row := range result.Data {
		fmt.Println(row["p.name"])
	}
}

func ExampleGraph_ROQuery() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	graph := db.SelectGraph("example")

	// Read-only queries can be cached and use replicas
	result, _ := graph.ROQuery(ctx, "MATCH (n:Person) RETURN count(n)")

	fmt.Printf("Person count: %v\n", result.Data[0]["count(n)"])
}

func ExampleGraph_Explain() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	graph := db.SelectGraph("example")

	// Get the execution plan without executing
	plan, _ := graph.Explain(ctx, "MATCH (p:Person)-[:KNOWS]->(f) RETURN p, f")

	for _, step := range plan {
		fmt.Println(step)
	}
}

func ExampleGraph_CreateNodeRangeIndex() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	graph := db.SelectGraph("example")

	// Create a range index for faster lookups
	graph.CreateNodeRangeIndex(ctx, "Person", "email")

	// Create a fulltext index for text search
	graph.CreateNodeFulltextIndex(ctx, "Person", "bio")

	// Create a vector index for similarity search
	graph.CreateNodeVectorIndex(ctx, "Person", 128, "cosine", "embedding")
}

func ExampleGraph_ConstraintCreate() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	graph := db.SelectGraph("example")

	// Create a unique constraint
	graph.ConstraintCreate(ctx,
		falkordb.ConstraintUnique,
		falkordb.EntityNode,
		"Person",
		"email",
	)

	// Create a mandatory constraint
	graph.ConstraintCreate(ctx,
		falkordb.ConstraintMandatory,
		falkordb.EntityNode,
		"Person",
		"name",
	)
}

func ExampleFalkorDB_List() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	// List all graphs
	graphs, _ := db.List(ctx)

	fmt.Println("Available graphs:", graphs)
}

func ExampleFalkorDB_ConfigGet() {
	ctx := context.Background()

	db, _ := falkordb.Connect(ctx, &falkordb.Options{Addr: "localhost:6379"})
	defer db.Close()

	// Get configuration
	value, _ := db.ConfigGet(ctx, "RESULTSET_SIZE")
	fmt.Printf("RESULTSET_SIZE: %v\n", value)

	// Set configuration
	db.ConfigSet(ctx, "RESULTSET_SIZE", 10000)
}
