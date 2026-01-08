package falkordb

import (
	"context"
	"strings"

	"github.com/flancast90/falkordb-go/internal/redis"
)

// FalkorDB is the main client for interacting with FalkorDB.
// It is safe for concurrent use by multiple goroutines.
type FalkorDB struct {
	client redis.Client
	opts   *Options
}

// Connect establishes a connection to FalkorDB.
//
// The client automatically detects the connection type (standalone, cluster, or sentinel)
// and configures itself accordingly.
//
// Example:
//
//	db, err := falkordb.Connect(ctx, &falkordb.Options{
//		Addr:     "localhost:6379",
//		Password: "secret",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
func Connect(ctx context.Context, opts *Options) (*FalkorDB, error) {
	if opts == nil {
		opts = &Options{}
	}
	opts.setDefaults()

	client, err := redis.NewClient(ctx, &redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})
	if err != nil {
		return nil, err
	}

	return &FalkorDB{
		client: client,
		opts:   opts,
	}, nil
}

// SelectGraph returns a Graph instance for the specified graph name.
// The graph does not need to exist; it will be created on first use.
func (db *FalkorDB) SelectGraph(name string) *Graph {
	return &Graph{
		name:   name,
		client: db.client,
		parser: newResultParser(),
	}
}

// List returns the names of all graphs in the database.
func (db *FalkorDB) List(ctx context.Context) ([]string, error) {
	result, err := db.client.Do(ctx, "GRAPH.LIST").Result()
	if err != nil {
		return nil, err
	}

	arr, ok := result.([]interface{})
	if !ok {
		return []string{}, nil
	}

	graphs := make([]string, len(arr))
	for i, g := range arr {
		graphs[i], _ = g.(string)
	}
	return graphs, nil
}

// ConfigGet retrieves a FalkorDB configuration value.
//
// Example:
//
//	value, _ := db.ConfigGet(ctx, "RESULTSET_SIZE")
func (db *FalkorDB) ConfigGet(ctx context.Context, key string) (interface{}, error) {
	result, err := db.client.Do(ctx, "GRAPH.CONFIG", "GET", key).Result()
	if err != nil {
		return nil, err
	}

	if arr, ok := result.([]interface{}); ok && len(arr) >= 2 {
		return arr[1], nil
	}
	return result, nil
}

// ConfigSet sets a FalkorDB configuration value.
//
// Example:
//
//	err := db.ConfigSet(ctx, "RESULTSET_SIZE", 10000)
func (db *FalkorDB) ConfigSet(ctx context.Context, key string, value interface{}) error {
	return db.client.Do(ctx, "GRAPH.CONFIG", "SET", key, value).Err()
}

// Info returns server information.
// If section is provided, returns information for that specific section.
//
// Example:
//
//	info, _ := db.Info(ctx)              // all info
//	info, _ := db.Info(ctx, "server")    // server section only
func (db *FalkorDB) Info(ctx context.Context, section ...string) (string, error) {
	args := []interface{}{"INFO"}
	if len(section) > 0 {
		args = append(args, section[0])
	}

	result, err := db.client.Do(ctx, args...).Result()
	if err != nil {
		return "", err
	}

	if s, ok := result.(string); ok {
		return s, nil
	}
	return "", nil
}

// Close closes the connection to FalkorDB.
func (db *FalkorDB) Close() error {
	return db.client.Close()
}

// Ping verifies the connection to FalkorDB is alive.
func (db *FalkorDB) Ping(ctx context.Context) error {
	return db.client.Ping(ctx).Err()
}

// parseGraphList parses a comma-separated list of graphs.
func parseGraphList(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
