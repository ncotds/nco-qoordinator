package client

import (
	"context"

	"github.com/ncotds/nco-qoordinator/pkg/models"
)

type QueryResult struct {
	RowSet       []map[string]any
	AffectedRows int
	Error        error
}

type Client interface {
	RawSQLPost(ctx context.Context, query models.Query, credentials models.Credentials) (map[string]QueryResult, error)
}
