package ncoclient

import (
	"context"
	"errors"
	"fmt"

	cm "github.com/ncotds/nco-qoordinator/pkg/connmanager"
	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

var (
	_ qc.Client = (*NcoClient)(nil)

	ErrClientConfig = fmt.Errorf("invalid client config")
)

// NcoClient implements querycoordinator.Client interface to work with coordinator.
// On the other hand, NcoClient interacts with the Pool to acquire connections and execute queries
type NcoClient struct {
	name string
	pool *cm.Pool
}

// NewNcoClient returns ready to use instance of client
func NewNcoClient(name string, pool *cm.Pool) (client *NcoClient, err error) {
	if name == "" {
		return nil, fmt.Errorf("%w: empty name", ErrClientConfig)
	}
	if pool == nil {
		return nil, fmt.Errorf("%w: nil pool", ErrClientConfig)
	}

	client = &NcoClient{
		name: name,
		pool: pool,
	}
	return client, err
}

// Name returns name of instance
func (c *NcoClient) Name() string {
	return c.name
}

// Exec runs query and return result or exit on context cancelation
func (c *NcoClient) Exec(ctx context.Context, query qc.Query, credentials qc.Credentials) (result qc.QueryResult) {
	done := make(chan struct{})
	go func() {
		result = c.exec(ctx, query, credentials)
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		result.Error = ctx.Err()
	}
	return result
}

// Close stops underlying Pool to prevent acquire connections,
// so client becomes 'close' too - you cannot use it to run queries anymore
func (c *NcoClient) Close() error {
	return c.pool.Close()
}

func (c *NcoClient) exec(ctx context.Context, query qc.Query, credentials qc.Credentials) qc.QueryResult {
	conn, err := c.pool.Acquire(ctx, credentials)
	if err != nil {
		return qc.QueryResult{Error: err}
	}

	rows, affected, err := conn.Exec(ctx, query)
	// connection is broken, try to establish it again
	if errors.Is(err, db.ErrConnection) {
		_ = c.pool.Drop(conn)
		conn, err = c.pool.Acquire(ctx, credentials)
		if err != nil {
			return qc.QueryResult{Error: err}
		}

		rows, affected, err = conn.Exec(ctx, query)
	}

	// return connection to reuse it later
	_ = c.pool.Release(conn)
	return qc.QueryResult{
		RowSet:       rows,
		AffectedRows: affected,
		Error:        err,
	}
}
