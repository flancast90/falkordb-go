# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-01-08

### Added

- Initial release
- Connection support for Standalone, Cluster, and Sentinel modes
- Full Cypher query support with parameterized queries
- Node, Edge, and Path types with property support
- Index management (Range, Fulltext, Vector) for nodes and edges
- Constraint management (Unique, Mandatory)
- Query analysis: Explain, Profile, SlowLog
- Graph operations: Copy, Delete
- Configuration management: ConfigGet, ConfigSet
- Comprehensive test suite with integration tests
- Docker Compose files for local development
- Full documentation and examples

### API

- `Connect()` - Connect to standalone FalkorDB
- `ConnectCluster()` - Connect to FalkorDB cluster
- `ConnectSentinel()` - Connect via Sentinel
- `FalkorDB.SelectGraph()` - Get a graph handle
- `Graph.Query()` - Execute write queries
- `Graph.ROQuery()` - Execute read-only queries
- `Graph.CreateNode*Index()` - Create node indexes
- `Graph.CreateEdge*Index()` - Create edge indexes
- `Graph.DropNode*Index()` - Drop node indexes
- `Graph.DropEdge*Index()` - Drop edge indexes
- `Graph.ConstraintCreate()` - Create constraints
- `Graph.ConstraintDrop()` - Drop constraints
- `Graph.Copy()` - Copy a graph
- `Graph.Delete()` - Delete a graph
- `Graph.Explain()` - Get query execution plan
- `Graph.Profile()` - Profile query execution
- `Graph.SlowLog()` - Get slow query log

