package tdsclient

import (
	"context"
	"fmt"
	"strings"

	tds "github.com/minus5/gofreetds"
	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

var _ db.ExecutorCloser = (*Connection)(nil)

type Connection struct {
	connStr string
	conn    *tds.Conn
}

func (c *Connection) Exec(ctx context.Context, query models.Query) (rows []models.QueryResultRow, affectedRows int, err error) {
	err = c.open(ctx) // ensure that connection exists
	if err != nil {
		return rows, affectedRows, err
	}

	var rst []*tds.Result
	var execErr error
	done := make(chan struct{})

	go func() {
		rst, execErr = c.conn.Exec(query.SQL)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
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
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
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
		done <- struct{}{}
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

func parseResults(rst []*tds.Result) (rows []models.QueryResultRow, affectedRows int, err error) {
	if len(rst) < 2 { // typically, ObjectServer returns 2 results: rowset and metadata
		err = app.Err(app.ErrCodeUnknown, fmt.Sprintf("unexpected db response: %#v", rst))
		return rows, affectedRows, err
	}

	cursor, meta := rst[0], rst[1]

	if cursor == nil || len(cursor.Columns) == 0 {
		// response w/o tabledata
		return rows, meta.RowsAffected, nil
	}

	rowSet := make([]models.QueryResultRow, 0, len(cursor.Rows))
	for _, cursorRow := range cursor.Rows {
		row := make(models.QueryResultRow, len(cursor.Columns))
		for i, col := range cursor.Columns {
			row[col.Name] = cursorRow[i]
		}
		rowSet = append(rowSet, row)
	}

	return rowSet, meta.RowsAffected, nil
}
