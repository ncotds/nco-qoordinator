package client

import (
	"context"

	db "github.com/ncotds/nco-lib/dbconnector"
)

type QueryResult struct {
	RowSet       []map[string]any
	AffectedRows int
	Error        error
}

type Client interface {
	RawSQLPost(ctx context.Context, query db.Query, credentials db.Credentials) (map[string]QueryResult, error)
}
