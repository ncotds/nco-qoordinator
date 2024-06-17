package client

import (
	"context"

	"github.com/ncotds/nco-qoordinator/pkg/models"
)

type Client interface {
	RawSQLPost(ctx context.Context, query models.Query, credentials models.Credentials) (map[string]models.QueryResult, error)
}
