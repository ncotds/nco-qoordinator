package tdsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/ncotds/go-dblib/tds"

	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

var (
	_ db.ExecutorCloser = (*Connection)(nil)

	// ErrConnectionInUse returned if caller tries to use Connection concurrently,
	// that is not possible with TDS protocol 'one-query-at-a-time' limitation:
	// OMNIbus server does not support multiplexing
	ErrConnectionInUse = errors.New("connection in use")
	// ErrQueryFailed means that 'transaction error' message has been received from the server
	ErrQueryFailed = errors.New("query failed")

	ErrTranNotCompleted = errors.New("transaction completed message did not found")
	ErrTranFailed       = errors.New("transaction failed")
)

type Connection struct {
	inUse      atomic.Bool
	conn       *tds.Conn
	connCancel context.CancelFunc
	dsn        *tds.Info
	ch         *tds.Channel
	appName    string
}

func (c *Connection) Exec(ctx context.Context, query models.Query) (rows models.RowSet, affectedRows int, err error) {
	if !c.inUse.CompareAndSwap(false, true) {
		return rows, affectedRows, ErrConnectionInUse
	}
	defer c.inUse.Store(false)

	err = c.open(ctx) // ensure that connection exists
	if err != nil {
		return rows, affectedRows, wrapTDSError(err)
	}

	pkgs, err := c.exec(ctx, query)
	err = wrapTDSError(err)
	if errors.Is(err, app.ErrTimeout) {
		// we cannot predict when the query will be completed
		// and cannot use this connection until query in progress...
		// so, force any communication with the server
		c.connCancel()
		_ = c.close()
	}
	if err != nil {
		return rows, affectedRows, err
	}

	rows, affectedRows, err = parseResults(pkgs)
	if err != nil {
		err = app.Err(app.ErrCodeUnknown, err.Error(), errors.Unwrap(err))
	}

	return rows, affectedRows, err
}

func (c *Connection) Close() error {
	if !c.inUse.CompareAndSwap(false, true) {
		return ErrConnectionInUse
	}
	defer c.inUse.Store(false)
	return wrapTDSError(c.close())
}

func (c *Connection) open(ctx context.Context) error {
	if c.conn != nil {
		return nil
	}

	connCtx, connCancel := context.WithCancel(context.Background())

	conn, err := tds.NewConn(connCtx, c.dsn)
	if err != nil {
		connCancel()
		return err
	}

	// NOTE: OMNIbus does not support multiplexing,
	// use the main channel for all communications
	ch, err := conn.NewChannel()
	if err != nil {
		connCancel()
		return err
	}

	login, err := tds.NewLoginConfig(c.dsn)
	if err != nil {
		connCancel()
		return err
	}
	login.AppName = c.appName
	login.Encrypt = 0

	err = ch.Login(ctx, login)
	if err != nil {
		connCancel()
		return err
	}

	c.conn = conn
	c.connCancel = connCancel
	c.ch = ch
	return nil
}

func (c *Connection) close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	c.conn = nil
	c.connCancel = nil
	c.ch = nil
	return err
}

func (c *Connection) exec(ctx context.Context, query models.Query) ([]tds.Package, error) {
	pkg := &tds.LanguagePackage{Cmd: query.SQL}

	err := c.ch.SendPackage(ctx, pkg)
	if err != nil {
		c.ch.Reset() // clear output queue
		return nil, err
	}

	var pkgs []tds.Package

	_, err = c.ch.NextPackageUntil(ctx, true, func(p tds.Package) (bool, error) {
		pkgs = append(pkgs, p)
		switch typedP := p.(type) {
		case *tds.DonePackage:
			if typedP.Status == tds.TDS_DONE_ERROR {
				return false, ErrQueryFailed
			}
			return typedP.Status == tds.TDS_DONE_FINAL, nil
		default:
			return false, nil
		}
	})

	return pkgs, err
}

func wrapTDSError(in error) error {
	var (
		err    error
		eedErr *tds.EEDError
	)
	switch {
	case in == io.EOF:
		// unwrapped io.EOF is OK, it means server has sent all data
		// and next has closed the connection
		err = nil
	case errors.As(in, &eedErr):
		reason := make([]error, len(eedErr.EEDPackages))
		for i, eed := range eedErr.EEDPackages {
			reason[i] = fmt.Errorf("%d: %s", eed.MsgNumber, eed.Msg)
		}
		err = app.Err(app.ErrCodeIncorrectOperation, eedErr.WrappedError.Error(), reason...)
	case errors.Is(in, ErrQueryFailed):
		err = app.Err(app.ErrCodeIncorrectOperation, in.Error(), errors.Unwrap(in))
	case errors.Is(in, context.Canceled), errors.Is(in, context.DeadlineExceeded):
		err = app.Err(app.ErrCodeTimeout, in.Error(), errors.Unwrap(in))
	case in != nil:
		err = app.Err(app.ErrCodeUnavailable, in.Error(), errors.Unwrap(in))
	}
	return err
}

func parseResults(pkgs []tds.Package) (rows models.RowSet, affectedRows int, err error) {
	affectedRows, err = checkTransactionCompleted(pkgs)
	if err != nil {
		return rows, affectedRows, err
	}

	rows = makeRowSet(pkgs, affectedRows)
	return rows, affectedRows, nil
}

func checkTransactionCompleted(pkgs []tds.Package) (affectedRows int, err error) { // check from the last to find 'transaction complete' msg
	var tranCompleted *tds.DonePackage
	for i := len(pkgs) - 1; i >= 0; i-- {
		pkg, ok := pkgs[i].(*tds.DonePackage)
		if ok && pkg.TranState == tds.TDS_TRAN_COMPLETED {
			tranCompleted = pkg
			break
		}
	}

	if tranCompleted == nil {
		return affectedRows, ErrTranNotCompleted
	}

	affectedRows = int(tranCompleted.Count)

	if tranCompleted.Status == tds.TDS_DONE_ERROR {
		err = ErrTranFailed
	}
	return affectedRows, err
}

func makeRowSet(pkgs []tds.Package, affectedRows int) models.RowSet {
	rows := models.RowSet{}
	rows.Columns = makeCols(pkgs)
	if rows.Columns == nil {
		return rows
	}

	rows.Rows = make([][]any, 0, affectedRows)
	for _, pkg := range pkgs {
		rowP, ok := pkg.(*tds.RowPackage)
		if !ok {
			continue
		}
		row := make([]any, len(rowP.DataFields))
		for fIdx, field := range rowP.DataFields {
			val := field.Value()
			if valStr, ok := val.(string); ok {
				val = strings.TrimSuffix(valStr, "\x00")
			}
			row[fIdx] = val
		}
		rows.Rows = append(rows.Rows, row)
	}
	return rows
}

func makeCols(pkgs []tds.Package) []string {
	var rows []string
	for i := 0; i < len(pkgs); i++ {
		rowFmt, ok := pkgs[i].(*tds.RowFmtPackage)
		if ok {
			rows = make([]string, 0, len(rowFmt.Fmts))
			for _, field := range rowFmt.Fmts {
				rows = append(rows, field.Name())
			}
			break
		}
	}
	return rows
}
