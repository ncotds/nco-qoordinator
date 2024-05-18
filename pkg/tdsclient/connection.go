package tdsclient

import (
	"context"
	"fmt"
	"strings"

	tds "github.com/minus5/gofreetds"
	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

var _ db.ExecutorCloser = (*Connection)(nil)

type Connection struct {
	connStr string
	conn    *tds.Conn
}

func (c *Connection) Exec(ctx context.Context, query qc.Query) (rows []qc.QueryResultRow, affectedRows int, err error) {
	err = c.open(ctx) // ensure that connection exists
	if err != nil {
		return rows, affectedRows, err
	}

	var rst []*tds.Result
	done := make(chan struct{})

	go func() {
		rst, err = c.conn.Exec(query.SQL)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	case <-done:
	}

	if err != nil {
		return rows, affectedRows, err
	}
	return parseResults(rst)
}

func (c *Connection) Close() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	return nil
}

func (c *Connection) IsConnectionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "dbopen error")
}

func (c *Connection) open(ctx context.Context) (err error) {
	if c.conn != nil {
		return nil
	}

	var conn *tds.Conn
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
		return err
	}
	c.conn = conn
	return nil
}

func parseResults(rst []*tds.Result) (rows []qc.QueryResultRow, affectedRows int, err error) {
	if len(rst) < 2 { // typically, ObjectServer returns 2 results: rowset and metadata
		return rows, affectedRows, fmt.Errorf("cannot parse db response")
	}

	cursor, meta := rst[0], rst[1]

	if cursor == nil || len(cursor.Columns) == 0 {
		// response w/o tabledata
		return rows, meta.RowsAffected, nil
	}

	rowSet := make([]qc.QueryResultRow, 0, len(cursor.Rows))
	for _, cursorRow := range cursor.Rows {
		row := make(qc.QueryResultRow, len(cursor.Columns))
		for i, col := range cursor.Columns {
			row[col.Name] = cursorRow[i]
		}
		rowSet = append(rowSet, row)
	}

	return rowSet, meta.RowsAffected, nil
}
