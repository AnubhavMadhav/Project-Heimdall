package ports

import (
	"context"
)

// Gatekeeper defines the primary security boundary.
// It is the only interface the MCP layer is allowed to talk to.
type Gatekeeper interface {
	// ExecuteSafeQuery parses, validates, and runs the query.
	// It returns the result rows or a security violation error.
	ExecuteSafeQuery(ctx context.Context, query string) ([]map[string]interface{}, error)

	// Introspection tools
	ListTables(ctx context.Context) ([]string, error)
	GetSchema(ctx context.Context, tableName string) (string, error)
}
