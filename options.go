package falkordb

import "time"

// Options configures the FalkorDB client connection.
type Options struct {
	// Addr is the FalkorDB server address in "host:port" format.
	// Default: "localhost:6379"
	Addr string

	// Password for Redis authentication.
	Password string

	// DB is the Redis database number.
	// Default: 0
	DB int

	// DialTimeout is the timeout for establishing new connections.
	// Default: 5s
	DialTimeout time.Duration

	// ReadTimeout is the timeout for socket reads.
	// Default: 3s
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for socket writes.
	// Default: same as ReadTimeout
	WriteTimeout time.Duration

	// PoolSize is the maximum number of connections in the pool.
	// Default: 10 * runtime.GOMAXPROCS
	PoolSize int

	// MinIdleConns is the minimum number of idle connections.
	// Default: 0
	MinIdleConns int
}

func (o *Options) setDefaults() {
	if o.Addr == "" {
		o.Addr = "localhost:6379"
	}
	if o.DialTimeout == 0 {
		o.DialTimeout = 5 * time.Second
	}
	if o.ReadTimeout == 0 {
		o.ReadTimeout = 3 * time.Second
	}
	if o.WriteTimeout == 0 {
		o.WriteTimeout = o.ReadTimeout
	}
}

// QueryOptions configures a Cypher query execution.
type QueryOptions struct {
	// Params are the query parameters to pass to the Cypher query.
	// Parameters are safely escaped and prevent injection attacks.
	Params map[string]interface{}

	// Timeout is the query timeout in milliseconds.
	// A value of 0 means no timeout.
	Timeout int
}
