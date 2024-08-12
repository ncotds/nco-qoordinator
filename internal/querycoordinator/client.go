//go:generate mockery
package querycoordinator

import (
	"context"

	db "github.com/ncotds/nco-lib/dbconnector"
)

type Client interface {
	// Name returns unique DSName of the client
	Name() string
	// Exec runs a SQL query against ObjectServer
	Exec(ctx context.Context, query db.Query, user db.Credentials) db.QueryResult
}
