//go:generate go run github.com/vektra/mockery/v2
package querycoordinator

import (
	"context"

	"github.com/ncotds/nco-qoordinator/pkg/models"
)

type Client interface {
	// Name returns unique DSName of the client
	Name() string
	// Exec runs a SQL query against ObjectServer
	Exec(ctx context.Context, query models.Query, user models.Credentials) models.QueryResult
}
