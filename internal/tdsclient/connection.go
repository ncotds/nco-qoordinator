package tdsclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	tds "github.com/minus5/gofreetds"
	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

var (
	_ db.ExecutorCloser = (*Connection)(nil)

	// ErrConnectionInUse returned if caller tries to use Connection concurrently,
	// that is not possible with TDS protocol 'one-query-at-a-time' limitation (by design)
	ErrConnectionInUse = errors.New("connection in use")
)

type Connection struct {
	connStr string
	conn    *tds.Conn
	inUse   atomic.Bool
}

func (c *Connection) Exec(ctx context.Context, query models.Query) (rows models.RowSet, affectedRows int, err error) {
	if !c.inUse.CompareAndSwap(false, true) {
		return rows, affectedRows, ErrConnectionInUse
	}
	defer c.inUse.Store(false)

	err = c.open(ctx) // ensure that connection exists
	if err != nil {
		return rows, affectedRows, err
	}

	var rst []*tds.Result
	var execErr error
	done := make(chan struct{})

	go func() {
		rst, execErr = c.conn.Exec(query.SQL)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return rows, 0, ctx.Err()
	case <-done:
	}

	switch {
	case c.isConnectionError(execErr):
		err = app.Err(app.ErrCodeUnavailable, "connection failed", execErr)
	case execErr != nil:
		err = app.Err(app.ErrCodeIncorrectOperation, "query failed", execErr)
	default:
		rows, affectedRows, err = parseResults(rst)
	}

	return rows, affectedRows, err
}

func (c *Connection) Close() error {
	if !c.inUse.CompareAndSwap(false, true) {
		return ErrConnectionInUse
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.inUse.Store(false)
	return nil
}

func (c *Connection) isConnectionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "dbopen error")
}

func (c *Connection) open(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}

	var conn *tds.Conn
	var err error
	done := make(chan struct{})
	go func() {
		conn, err = tds.NewConn(c.connStr)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	if err != nil {
		err = app.Err(app.ErrCodeUnavailable, "cannot connect db", err)
		return err
	}
	c.conn = conn
	return nil
}

func parseResults(rst []*tds.Result) (rows models.RowSet, affectedRows int, err error) {
	if len(rst) < 2 { // typically, ObjectServer returns 2 results: rowset and metadata
		err = app.Err(app.ErrCodeUnknown, fmt.Sprintf("unexpected db response: %#v", rst))
		return rows, affectedRows, err
	}

	cursor, meta := rst[0], rst[1]

	if cursor == nil || len(cursor.Columns) == 0 {
		// response w/o tabledata
		return rows, meta.RowsAffected, nil
	}

	columns := make([]string, 0, len(cursor.Columns))
	for _, col := range cursor.Columns {
		columns = append(columns, col.Name)
	}

	return models.RowSet{Columns: columns, Rows: cursor.Rows}, meta.RowsAffected, nil
}
