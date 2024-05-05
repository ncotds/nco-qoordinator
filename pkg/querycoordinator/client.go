//go:generate go run github.com/vektra/mockery/v2@v2.42.2
package querycoordinator

import "context"

type Client interface {
	// Name returns unique DSName of the client
	Name() string
	// Exec runs a SQL query against ObjectServer
	Exec(ctx context.Context, query Query, user Credentials) QueryResult
}
