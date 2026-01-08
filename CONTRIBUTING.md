# Contributing to FalkorDB Go Client

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## Getting Started

1. **Fork the repository** and clone it locally
2. **Install dependencies**: `go mod download`
3. **Run tests**: `make test`

## Development Setup

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose (for integration tests)
- Make

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
make test

# Integration tests (requires Docker)
make test-integration

# All tests with coverage report
make coverage

# Lint the code
make lint
```

### Code Style

- Follow standard Go conventions and [Effective Go](https://go.dev/doc/effective_go)
- Run `make fmt` before committing
- Run `make vet` to check for issues
- Add tests for new functionality

## Making Changes

### Branch Naming

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring

### Commit Messages

Use clear, descriptive commit messages:

```
Add vector index support for nodes and edges

- Implement CreateNodeVectorIndex and CreateEdgeVectorIndex
- Add OPTIONS clause support for similarity functions
- Add tests for vector index operations
```

### Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** with tests
3. **Run the test suite**: `make test-all`
4. **Update documentation** if needed
5. **Submit a pull request** with a clear description

## Code Structure

```
falkordb-go/
├── falkordb.go          # Public API: Connect, FalkorDB type
├── graph.go             # Graph operations
├── types.go             # Data types (Node, Edge, Path, etc.)
├── options.go           # Configuration options
├── result.go            # Result parsing
├── internal/
│   ├── proto/           # Wire protocol (not exported)
│   │   ├── args.go      # Command argument building
│   │   └── parser.go    # Response parsing
│   └── redis/           # Redis client wrapper
├── tests/
│   └── integration/     # Integration tests
└── examples/            # Usage examples
```

### Guidelines

- **Public API** goes in the root package
- **Internal implementation** goes in `internal/`
- **Tests** go alongside the code they test (`*_test.go`)
- **Integration tests** go in `tests/integration/`

## Testing

### Unit Tests

Unit tests should:
- Be fast (no external dependencies)
- Test edge cases
- Use table-driven tests where appropriate

```go
func TestValueToString(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected string
    }{
        {"string", "hello", `"hello"`},
        {"int", 42, "42"},
        {"bool", true, "true"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := valueToString(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Integration Tests

Integration tests:
- Require a running FalkorDB instance
- Use `t.Skip()` if the database isn't available
- Clean up after themselves

```go
func TestSomething(t *testing.T) {
    db := newTestDB(t) // Skips if FalkorDB unavailable
    defer db.Close()
    
    ctx := context.Background()
    graph := db.SelectGraph(randomName())
    defer graph.Delete(ctx) // Cleanup
    
    // Test code...
}
```

## Reporting Issues

When reporting issues, please include:

1. **Go version**: `go version`
2. **FalkorDB version**: Check Docker image tag
3. **Steps to reproduce**
4. **Expected vs actual behavior**
5. **Relevant code snippets**

## Feature Requests

For feature requests:

1. **Check existing issues** to avoid duplicates
2. **Describe the use case** - why is this needed?
3. **Propose an API** if you have ideas

## Questions?

- Open a [GitHub Discussion](https://github.com/flancast90/falkordb-go/discussions)
- Check [FalkorDB Documentation](https://docs.falkordb.com)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

