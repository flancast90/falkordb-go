package main

import (
	"context"
	"fmt"
	"log"

	falkordb "github.com/FalkorDB/falkordb-go"
)

func main() {
	ctx := context.Background()

	// Connect to FalkorDB
	db, err := falkordb.Connect(ctx, &falkordb.Options{
		Addr: "localhost:6379",
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Select a graph
	graph := db.SelectGraph("social")

	// Create some nodes and relationships
	_, err = graph.Query(ctx, `
		CREATE (alice:Person {name: 'Alice', age: 30})
		CREATE (bob:Person {name: 'Bob', age: 25})
		CREATE (charlie:Person {name: 'Charlie', age: 35})
		CREATE (alice)-[:KNOWS {since: 2020}]->(bob)
		CREATE (bob)-[:KNOWS {since: 2021}]->(charlie)
		CREATE (alice)-[:KNOWS {since: 2019}]->(charlie)
	`)
	if err != nil {
		log.Fatalf("Failed to create data: %v", err)
	}
	fmt.Println("Created nodes and relationships")

	// Query with parameters
	result, err := graph.Query(ctx,
		"MATCH (p:Person) WHERE p.age > $minAge RETURN p.name, p.age ORDER BY p.age",
		&falkordb.QueryOptions{
			Params: map[string]interface{}{
				"minAge": 26,
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\nPeople older than 26:")
	for _, row := range result.Data {
		fmt.Printf("  %s (age %v)\n", row["p.name"], row["p.age"])
	}

	// Query relationships
	result, err = graph.Query(ctx, `
		MATCH (p:Person)-[r:KNOWS]->(friend:Person)
		RETURN p.name, friend.name, r.since
	`)
	if err != nil {
		log.Fatalf("Failed to query relationships: %v", err)
	}

	fmt.Println("\nFriendships:")
	for _, row := range result.Data {
		fmt.Printf("  %s knows %s since %v\n", row["p.name"], row["friend.name"], row["r.since"])
	}

	// Get a path
	result, err = graph.Query(ctx, `
		MATCH path = (a:Person {name: 'Alice'})-[:KNOWS*]->(c:Person {name: 'Charlie'})
		RETURN path
	`)
	if err != nil {
		log.Fatalf("Failed to query path: %v", err)
	}

	fmt.Println("\nPaths from Alice to Charlie:")
	for _, row := range result.Data {
		path := row["path"].(*falkordb.Path)
		fmt.Printf("  Path with %d nodes and %d edges\n", len(path.Nodes), len(path.Edges))
	}

	// Create an index
	_, err = graph.CreateNodeRangeIndex(ctx, "Person", "name")
	if err != nil {
		log.Printf("Index may already exist: %v", err)
	}

	// Explain a query
	plan, err := graph.Explain(ctx, "MATCH (p:Person) WHERE p.name = 'Alice' RETURN p")
	if err != nil {
		log.Fatalf("Failed to explain: %v", err)
	}

	fmt.Println("\nQuery execution plan:")
	for _, step := range plan {
		fmt.Printf("  %s\n", step)
	}

	// Profile a query
	profile, err := graph.Profile(ctx, "MATCH (p:Person) RETURN p")
	if err != nil {
		log.Fatalf("Failed to profile: %v", err)
	}

	fmt.Println("\nQuery profile:")
	for _, step := range profile {
		fmt.Printf("  %s\n", step)
	}

	// Get slow log
	slowLog, err := graph.SlowLog(ctx)
	if err != nil {
		log.Fatalf("Failed to get slow log: %v", err)
	}

	fmt.Println("\nSlow log entries:", len(slowLog))

	// List all graphs
	graphs, err := db.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list graphs: %v", err)
	}

	fmt.Println("\nAvailable graphs:", graphs)

	// Get server info
	info, err := db.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get info: %v", err)
	}

	fmt.Printf("\nServer info (first 200 chars): %.200s...\n", info)

	// Clean up
	err = graph.Delete(ctx)
	if err != nil {
		log.Fatalf("Failed to delete graph: %v", err)
	}
	fmt.Println("\nGraph deleted successfully")
}
