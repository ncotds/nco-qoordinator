//go:build integration
// +build integration

package tdsclient_test

import (
	"context"
	"testing"

	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
	"github.com/ncotds/nco-qoordinator/pkg/tdsclient"
	"github.com/stretchr/testify/assert"
)

func TestTDSConnector_Connect(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{TestConfig.User, TestConfig.Password}

	type args struct {
		ctx         context.Context
		addr        db.Addr
		credentials qc.Credentials
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"connect ok",
			args{ctx: ctx, addr: TestConfig.Address, credentials: credentials},
			false,
		},
		{
			"bad address fails",
			args{ctx: ctx, addr: db.Addr(WordFactory()), credentials: credentials},
			true,
		},
		{
			"bad credentials fails",
			args{ctx: ctx, addr: TestConfig.Address, credentials: qc.Credentials{WordFactory(), WordFactory()}},
			true,
		},
		{
			"context cancel",
			args{
				ctx: func() context.Context {
					c, cancel := context.WithCancel(context.Background())
					cancel()
					return c
				}(),
				addr:        TestConfig.Address,
				credentials: credentials,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &tdsclient.TDSConnector{AppLabel: TestConnLabel, TimeoutSec: TestConnTimeoutSec}
			gotConn, err := c.Connect(tt.args.ctx, tt.args.addr, tt.args.credentials)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, gotConn) {
				assert.NoError(t, gotConn.Close())
			}
		})
	}
}
