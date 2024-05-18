package tdsclient

import (
	"context"
	"fmt"

	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

const (
	CompatibilityMode = "sybase"
	TDSVersion        = "1.0.0" // for better TDS client compatibility
)

var _ db.DBConnector = (*TDSConnector)(nil)

type TDSConnector struct {
	AppLabel   string
	TimeoutSec uint
}

func (c *TDSConnector) Connect(
	ctx context.Context,
	addr db.Addr,
	credentials qc.Credentials,
) (conn db.ExecutorCloser, err error) {
	newConn := &Connection{connStr: c.connStr(addr, credentials)}
	err = newConn.open(ctx)
	if err != nil {
		return nil, err
	}
	return newConn, nil
}

func (c *TDSConnector) connStr(addr db.Addr, credentials qc.Credentials) string {
	return fmt.Sprintf(
		"host=%s;user=%s;pwd=%s;app=%s;conn_timeout=%d;compatibility=%s;tds_version=%s",
		addr, credentials.UserName, credentials.Password, c.AppLabel, c.TimeoutSec, CompatibilityMode, TDSVersion,
	)
}
