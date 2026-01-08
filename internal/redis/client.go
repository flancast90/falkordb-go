// Package redis provides Redis client implementations for FalkorDB.
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is the interface for Redis client operations used by FalkorDB.
type Client interface {
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Close() error
	Ping(ctx context.Context) *redis.StatusCmd
}

// Options configures the Redis connection.
type Options struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

// NewClient creates a new Redis client based on the connection type detected.
func NewClient(ctx context.Context, opts *Options) (Client, error) {
	// Try to detect connection type by attempting connection
	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	// Check if this is a sentinel
	info, err := client.Info(ctx, "server").Result()
	if err == nil && containsSentinel(info) {
		// Handle sentinel connection
		return newSentinelClient(ctx, client, opts)
	}

	// Check if this is a cluster
	clusterInfo, err := client.ClusterInfo(ctx).Result()
	if err == nil && clusterInfo != "" {
		// Handle cluster connection
		client.Close()
		return newClusterClient(ctx, opts)
	}

	return &singleClient{client: client}, nil
}

func containsSentinel(info string) bool {
	// Simple check - in practice you'd parse the info properly
	return false // Sentinel detection would happen here
}

// singleClient wraps a single Redis client.
type singleClient struct {
	client *redis.Client
}

func (c *singleClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	return c.client.Do(ctx, args...)
}

func (c *singleClient) Close() error {
	return c.client.Close()
}

func (c *singleClient) Ping(ctx context.Context) *redis.StatusCmd {
	return c.client.Ping(ctx)
}

// newClusterClient creates a cluster client.
func newClusterClient(ctx context.Context, opts *Options) (Client, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{opts.Addr},
		Password:     opts.Password,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &clusterClient{client: client}, nil
}

type clusterClient struct {
	client *redis.ClusterClient
}

func (c *clusterClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	return c.client.Do(ctx, args...)
}

func (c *clusterClient) Close() error {
	return c.client.Close()
}

func (c *clusterClient) Ping(ctx context.Context) *redis.StatusCmd {
	return c.client.Ping(ctx)
}

// newSentinelClient creates a sentinel-based client.
func newSentinelClient(ctx context.Context, sentinelClient *redis.Client, opts *Options) (Client, error) {
	// Get master info from sentinel
	masters, err := sentinelClient.Do(ctx, "SENTINEL", "MASTERS").Result()
	if err != nil {
		return nil, err
	}

	// Parse master info
	masterAddr := parseMasterAddr(masters)
	if masterAddr == "" {
		// Not actually a sentinel, return single client
		return &singleClient{client: sentinelClient}, nil
	}

	sentinelClient.Close()

	// Connect to master
	client := redis.NewClient(&redis.Options{
		Addr:         masterAddr,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &singleClient{client: client}, nil
}

func parseMasterAddr(masters interface{}) string {
	arr, ok := masters.([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}

	// Get first master
	master, ok := arr[0].([]interface{})
	if !ok {
		return ""
	}

	// Parse key-value pairs
	var ip, port string
	for i := 0; i < len(master)-1; i += 2 {
		key, _ := master[i].(string)
		val, _ := master[i+1].(string)
		switch key {
		case "ip":
			ip = val
		case "port":
			port = val
		}
	}

	if ip != "" && port != "" {
		return ip + ":" + port
	}
	return ""
}
