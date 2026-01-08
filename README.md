# FalkorDB Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/FalkorDB/falkordb-go.svg)](https://pkg.go.dev/github.com/FalkorDB/falkordb-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/FalkorDB/falkordb-go)](https://goreportcard.com/report/github.com/FalkorDB/falkordb-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go client library for [FalkorDB](https://falkordb.com), a high-performance graph database built on Redis.

## Features

- Full Cypher query support with parameterized queries
- Node, Edge, and Path types with property support
- Graph operations: create, copy, delete
- Index management (Range, Fulltext, Vector)
- Constraint management (Unique, Mandatory)
- Query analysis: Explain, Profile, SlowLog
- Connection modes: Standalone, Cluster, Sentinel
- Thread-safe and connection pooling
- Context support for cancellation and timeouts

## Installation

```bash
go get github.com/FalkorDB/falkordb-go
```

Requires Go 1.22 or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/FalkorDB/falkordb-go"
)

func main() {
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
        CREATE (alice)-[:KNOWS {since: 2020}]->(bob)
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Query with parameters
    result, err := graph.Query(ctx, 
        "MATCH (p:Person) WHERE p.age > $minAge RETURN p.name, p.age",
        &falkordb.QueryOptions{
            Params: map[string]interface{}{
                "minAge": 20,
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Process results
    for _, row := range result.Data {
        fmt.Printf("Name: %s, Age: %d\n", row["p.name"], row["p.age"])
    }
}
```

## Connection Options

### Standalone

```go
db, err := falkordb.Connect(ctx, &falkordb.Options{
    Addr:     "localhost:6379",
    Password: "secret",      // optional
    DB:       0,             // optional, default database
})
```

### Cluster

```go
db, err := falkordb.ConnectCluster(ctx, &falkordb.ClusterOptions{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    },
    Password: "secret",
})
```

### Sentinel

```go
db, err := falkordb.ConnectSentinel(ctx, &falkordb.SentinelOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"localhost:26379"},
    Password:      "secret",
})
```

## Query Operations

### Basic Queries

```go
// Write query
result, err := graph.Query(ctx, "CREATE (n:Person {name: 'Alice'}) RETURN n")

// Read-only query (enables caching and replica reads)
result, err := graph.ROQuery(ctx, "MATCH (n:Person) RETURN n")
```

### Parameterized Queries

```go
result, err := graph.Query(ctx,
    "CREATE (n:Person {name: $name, age: $age}) RETURN n",
    &falkordb.QueryOptions{
        Params: map[string]interface{}{
            "name": "Alice",
            "age":  30,
        },
        Timeout: 5000, // milliseconds
    },
)
```

### Working with Results

```go
result, _ := graph.Query(ctx, "MATCH (n:Person)-[r:KNOWS]->(m) RETURN n, r, m")

for _, row := range result.Data {
    // Access nodes
    if node, ok := row["n"].(*falkordb.Node); ok {
        fmt.Printf("Node ID: %d, Labels: %v\n", node.ID, node.Labels)
        fmt.Printf("Properties: %v\n", node.Properties)
    }
    
    // Access edges
    if edge, ok := row["r"].(*falkordb.Edge); ok {
        fmt.Printf("Edge: %s, From: %d, To: %d\n", 
            edge.RelationshipType, edge.SourceID, edge.DestinationID)
    }
}
```

### Paths

```go
result, _ := graph.Query(ctx, "MATCH p = (a)-[:KNOWS*]->(b) RETURN p")

for _, row := range result.Data {
    if path, ok := row["p"].(*falkordb.Path); ok {
        fmt.Printf("Path length: %d\n", path.Length())
        fmt.Printf("Nodes: %d, Edges: %d\n", len(path.Nodes), len(path.Edges))
    }
}
```

## Index Management

```go
// Create indexes
graph.CreateNodeRangeIndex(ctx, "Person", "name")
graph.CreateNodeFulltextIndex(ctx, "Person", "bio")
graph.CreateNodeVectorIndex(ctx, "Person", 128, "euclidean", "embedding")

// Edge indexes
graph.CreateEdgeRangeIndex(ctx, "KNOWS", "since")

// Drop indexes
graph.DropNodeRangeIndex(ctx, "Person", "name")
```

## Constraints

```go
// Create constraints
graph.ConstraintCreate(ctx, falkordb.ConstraintUnique, falkordb.EntityNode, "Person", "email")
graph.ConstraintCreate(ctx, falkordb.ConstraintMandatory, falkordb.EntityNode, "Person", "name")

// Drop constraints
graph.ConstraintDrop(ctx, falkordb.ConstraintUnique, falkordb.EntityNode, "Person", "email")
```

## Graph Operations

```go
// Copy a graph
err := graph.Copy(ctx, "social_backup")

// Delete a graph
err := graph.Delete(ctx)

// Get execution plan
plan, err := graph.Explain(ctx, "MATCH (n:Person) RETURN n")

// Profile a query
profile, err := graph.Profile(ctx, "MATCH (n:Person) RETURN n")

// Get slow query log
entries, err := graph.SlowLog(ctx)
for _, entry := range entries {
    fmt.Printf("[%d] %s: %s (%.2fms)\n", 
        entry.Timestamp, entry.Command, entry.Query, entry.Took)
}
```

## Data Types

The client supports all FalkorDB data types:

| FalkorDB Type | Go Type |
|---------------|---------|
| String | `string` |
| Integer | `int64` |
| Float | `float64` |
| Boolean | `bool` |
| Null | `nil` |
| Array | `[]interface{}` |
| Map | `map[string]interface{}` |
| Node | `*falkordb.Node` |
| Edge | `*falkordb.Edge` |
| Path | `*falkordb.Path` |
| Point | `*falkordb.Point` |

## Development

### Prerequisites

- Go 1.22+
- Docker and Docker Compose

### Running Tests

```bash
# Unit tests only (no FalkorDB required)
make test

# Integration tests (starts FalkorDB in Docker)
make test-integration

# With custom port
FALKORDB_PORT=16379 make test-integration

# All tests with coverage
make coverage
```

### Docker Commands

```bash
# Start standalone FalkorDB
make docker-standalone

# Start cluster (6 nodes)
make docker-cluster

# Start with Sentinel
make docker-sentinel

# Stop all
make docker-stop-all
```

### Environment Variables

Create a `.env` file (see `.env.example`):

```bash
FALKORDB_HOST=localhost
FALKORDB_PORT=6379
```

## Project Structure

```
falkordb-go/
├── falkordb.go          # Main entry point, Connect functions
├── graph.go             # Graph type and methods
├── types.go             # Node, Edge, Path, Point types
├── options.go           # QueryOptions, connection options
├── result.go            # Result parsing
├── internal/
│   ├── proto/           # Protocol encoding/parsing
│   └── redis/           # Redis client abstraction
├── tests/
│   └── integration/     # End-to-end tests
├── docker/              # Docker Compose files
└── examples/            # Usage examples
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [FalkorDB](https://falkordb.com) - The graph database
- [FalkorDB Documentation](https://docs.falkordb.com)
- [Cypher Query Language](https://docs.falkordb.com/cypher/)
