package ncoclient

import (
	"context"
	"errors"
	"time"

	cm "github.com/ncotds/nco-qoordinator/internal/connmanager"
	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

const logLabelClient = "ncoclient/NcoClient"

var _ qc.Client = (*NcoClient)(nil)

type ClientConfig struct {
	// Connector is object that can open DB connections
	Connector db.DBConnector
	// SeedList is list of OMNIbus cluster nodes
	SeedList []db.Addr
	// MaxPoolSize - max connections that client can open.
	// If MaxPoolSize <= 0, pool with default size will be created
	MaxPoolSize int
	// UseRandomFailOver enables connmanager.WithRandomFailOver pool option.
	// Useful for Display level of OMNIBus cluster
	UseRandomFailOver bool
	// UseFailBack enables connmanager.WithFailBack pool option, overrides UseRandomFailOver.
	// Useful for Aggregation level of OMNIBus cluster
	UseFailBack bool
	// FailBackDelay is time after that client will try to reconnect to the first node in SeedList.
	// Takes effect if UseRandomFailOver is true
	FailBackDelay time.Duration
	// Logger sets logger for client and underlying pool. By default no-op logger is used
	Logger *app.Logger
}

// NcoClient implements querycoordinator.Client interface to work with coordinator.
// On the other hand, NcoClient interacts with the connmanager.Pool to acquire connections and execute queries
type NcoClient struct {
	name string
	pool *cm.Pool
	log  *app.Logger
}

// NewNcoClient returns ready to use instance of client.
// Calls connmanager.NewPool to create underlying connmanager.Pool
func NewNcoClient(name string, config ClientConfig) (client *NcoClient, err error) {
	if name == "" {
		return nil, app.Err(app.ErrCodeValidation, "invalid client config, empty name")
	}

	log := app.NewLogger(nil)
	if config.Logger != nil {
		log = config.Logger.WithComponent(logLabelClient).With("nco_client", name)
	}

	poolOpts := []cm.PoolOption{cm.WithLogger(log)}
	if config.MaxPoolSize > 0 {
		poolOpts = append(poolOpts, cm.WithMaxSize(config.MaxPoolSize))
	}
	if config.UseRandomFailOver {
		poolOpts = append(poolOpts, cm.WithRandomFailOver())
	}
	if config.UseFailBack {
		poolOpts = append(poolOpts, cm.WithFailBack(config.FailBackDelay))
	}

	pool, err := cm.NewPool(config.Connector, config.SeedList, poolOpts...)
	if err != nil {
		return nil, err
	}

	client = &NcoClient{
		name: name,
		pool: pool,
		log:  log,
	}
	return client, err
}

// Name returns name of instance
func (c *NcoClient) Name() string {
	return c.name
}

// Exec runs query and return result or exit on context cancellation
func (c *NcoClient) Exec(ctx context.Context, query models.Query, credentials models.Credentials) models.QueryResult {
	var result models.QueryResult
	done := make(chan struct{})
	go func() {
		result = c.exec(ctx, query, credentials)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return models.QueryResult{Error: ctx.Err()}
	case <-done:
	}
	return result
}

// Close stops underlying Pool to prevent acquire connections,
// so client becomes 'close' too - you cannot use it to run queries anymore
func (c *NcoClient) Close() error {
	return c.pool.Close()
}

func (c *NcoClient) exec(ctx context.Context, query models.Query, credentials models.Credentials) models.QueryResult {
	ctx = app.WithLogAttrs(ctx, app.Attrs{"user": credentials.UserName})

	conn, err := c.pool.Acquire(ctx, credentials)
	if err != nil {
		return models.QueryResult{Error: err}
	}

	rows, affected, err := conn.Exec(ctx, query)
	// connection is broken, try to establish it again
	if errors.Is(err, app.ErrUnavailable) {
		c.log.DebugContext(ctx, "loose connection, try to reconnect")
		if err := c.pool.Drop(conn); err != nil {
			c.log.ErrContext(ctx, err, "cannot return failed connection to pool")
		}
		conn, err = c.pool.Acquire(ctx, credentials)
		if err != nil {
			return models.QueryResult{Error: err}
		}

		rows, affected, err = conn.Exec(ctx, query)
	}

	// return connection to reuse it later
	if releaseErr := c.pool.Release(conn); releaseErr != nil {
		c.log.ErrContext(ctx, releaseErr, "cannot return connection to pool")
	} else {
		c.log.DebugContext(ctx, "connection released to pool")
	}
	return models.QueryResult{
		RowSet:       rows,
		AffectedRows: affected,
		Error:        err,
	}
}
