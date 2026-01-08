package main

import (
	"context"
	"fmt"
	"log"

	falkordb "github.com/flancast90/falkordb-go"
)

func main() {
	ctx := context.Background()

	db, graph := setup(ctx)
	defer db.Close()
	defer cleanup(ctx, graph)

	createData(ctx, graph)
	queryPeople(ctx, graph)
	queryRelationships(ctx, graph)
	queryPaths(ctx, graph)
	demonstrateIndexes(ctx, graph)
	showServerInfo(ctx, db)
}

func setup(ctx context.Context) (*falkordb.FalkorDB, *falkordb.Graph) {
	db, err := falkordb.Connect(ctx, &falkordb.Options{
		Addr: "localhost:6379",
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	return db, db.SelectGraph("social")
}

func cleanup(ctx context.Context, graph *falkordb.Graph) {
	if err := graph.Delete(ctx); err != nil {
		log.Printf("Failed to delete graph: %v", err)
	}
	fmt.Println("\nGraph deleted successfully")
}

func createData(ctx context.Context, graph *falkordb.Graph) {
	_, err := graph.Query(ctx, `
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
}

func queryPeople(ctx context.Context, graph *falkordb.Graph) {
	result, err := graph.Query(ctx,
		"MATCH (p:Person) WHERE p.age > $minAge RETURN p.name, p.age ORDER BY p.age",
		&falkordb.QueryOptions{Params: map[string]interface{}{"minAge": 26}},
	)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Println("\nPeople older than 26:")
	for _, row := range result.Data {
		fmt.Printf("  %s (age %v)\n", row["p.name"], row["p.age"])
	}
}

func queryRelationships(ctx context.Context, graph *falkordb.Graph) {
	result, err := graph.Query(ctx, `
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
}

func queryPaths(ctx context.Context, graph *falkordb.Graph) {
	result, err := graph.Query(ctx, `
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
}

func demonstrateIndexes(ctx context.Context, graph *falkordb.Graph) {
	// Create an index
	if _, err := graph.CreateNodeRangeIndex(ctx, "Person", "name"); err != nil {
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
}

func showServerInfo(ctx context.Context, db *falkordb.FalkorDB) {
	graphs, err := db.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list graphs: %v", err)
	}
	fmt.Println("\nAvailable graphs:", graphs)

	info, err := db.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get info: %v", err)
	}
	fmt.Printf("\nServer info (first 200 chars): %.200s...\n", info)
}
