// Package falkordb provides a Go client for FalkorDB, a high-performance
// graph database built on Redis.
//
// # Quick Start
//
// Connect to FalkorDB and execute queries:
//
//	ctx := context.Background()
//
//	// Connect to FalkorDB
//	db, err := falkordb.Connect(ctx, &falkordb.Options{
//	    Addr: "localhost:6379",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Select a graph
//	graph := db.SelectGraph("social")
//
//	// Create data
//	_, err = graph.Query(ctx, `
//	    CREATE (alice:Person {name: 'Alice', age: 30})
//	    CREATE (bob:Person {name: 'Bob', age: 25})
//	    CREATE (alice)-[:KNOWS]->(bob)
//	`)
//
//	// Query with parameters
//	result, err := graph.Query(ctx,
//	    "MATCH (p:Person) WHERE p.age > $minAge RETURN p",
//	    &falkordb.QueryOptions{
//	        Params: map[string]interface{}{"minAge": 20},
//	    },
//	)
//
//	// Process results
//	for _, row := range result.Data {
//	    node := row["p"].(*falkordb.Node)
//	    fmt.Printf("%s is %d years old\n",
//	        node.Properties["name"], node.Properties["age"])
//	}
//
// # Connection Modes
//
// The client supports three connection modes:
//
//   - Standalone: Single FalkorDB instance via [Connect]
//   - Cluster: Redis Cluster mode via [ConnectCluster]
//   - Sentinel: High availability via [ConnectSentinel]
//
// # Graph Operations
//
// The [Graph] type provides methods for:
//
//   - Executing Cypher queries ([Graph.Query], [Graph.ROQuery])
//   - Managing indexes ([Graph.CreateNodeRangeIndex], etc.)
//   - Managing constraints ([Graph.ConstraintCreate], [Graph.ConstraintDrop])
//   - Graph operations ([Graph.Copy], [Graph.Delete])
//   - Query analysis ([Graph.Explain], [Graph.Profile], [Graph.SlowLog])
//
// # Data Types
//
// Query results contain Go representations of FalkorDB types:
//
//   - [Node]: Graph nodes with labels and properties
//   - [Edge]: Relationships with type and properties
//   - [Path]: Paths containing nodes and edges
//   - [Point]: Geographic coordinates
//
// # Thread Safety
//
// All types in this package are safe for concurrent use.
package falkordb

