//go:generate go run github.com/vektra/mockery/v2@v2.42.2
package dbconnector

import (
	"context"
	"io"

	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

type Addr string

type DBConnector interface {
	Connect(ctx context.Context, addr Addr, credentials qc.Credentials) (conn ExecutorCloser, err error)
}

type ExecutorCloser interface {
	Exec(ctx context.Context, query qc.Query) (rows []qc.QueryResultRow, affectedRows int, err error)
	IsConnectionError(err error) bool
	io.Closer
}
