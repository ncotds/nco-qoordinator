package ncoclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	cm "github.com/ncotds/nco-qoordinator/pkg/connmanager"
	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	mocks "github.com/ncotds/nco-qoordinator/pkg/dbconnector/mocks"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

func TestNcoClient_exec(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}
	query := qc.Query{SQL: SentenceFactory()}
	seedList := SeedListFactory(5)
	anyError := mock.IsType(ErrorFactory())

	type fields struct {
		name      string
		connector func(t *testing.T) *mocks.MockDBConnector
	}
	tests := []struct {
		name      string
		fields    fields
		wantRows  bool
		wantErrIs error
	}{
		{
			"query success",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					conn := mocks.NewMockExecutorCloser(t)
					conn.EXPECT().Exec(ctx, query).Return(
						[]qc.QueryResultRow{
							{WordFactory(): WordFactory()},
							{WordFactory(): WordFactory()},
							{WordFactory(): WordFactory()},
						},
						0,
						nil,
					).Once()
					conn.EXPECT().IsConnectionError(nil).Return(false).Once()

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(conn, nil)
					return m
				},
			},
			true,
			nil,
		},
		{
			"cannot connect fail",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(nil, db.ErrConnection)
					return m
				},
			},
			false,
			db.ErrConnection,
		},
		{
			"exec query fails",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					connBad := mocks.NewMockExecutorCloser(t)
					connBad.EXPECT().Exec(ctx, query).Return(nil, 0, ErrorFactory())
					connBad.EXPECT().IsConnectionError(anyError).Return(false)

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(connBad, nil)
					return m
				},
			},
			false,
			db.ErrQuery,
		},
		{
			"exec loose connection reconnect ok",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					connBad := mocks.NewMockExecutorCloser(t)
					connBad.EXPECT().Exec(ctx, query).Return(nil, 0, ErrorFactory())
					connBad.EXPECT().IsConnectionError(anyError).Return(true)
					connBad.EXPECT().Close().Return(nil)

					connOk := mocks.NewMockExecutorCloser(t)
					connOk.EXPECT().Exec(ctx, query).Return(
						[]qc.QueryResultRow{
							{WordFactory(): WordFactory()},
							{WordFactory(): WordFactory()},
							{WordFactory(): WordFactory()},
						},
						0,
						nil,
					)
					connOk.EXPECT().IsConnectionError(nil).Return(false)

					var failFlag bool

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						RunAndReturn(
							func(ctx context.Context, addr db.Addr, credentials qc.Credentials) (db.ExecutorCloser, error) {
								if !failFlag {
									failFlag = true
									return connBad, nil
								}
								return connOk, nil
							},
						)
					return m
				},
			},
			true,
			nil,
		},
		{
			"exec loose connection reconnect fails",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					connBad := mocks.NewMockExecutorCloser(t)
					connBad.EXPECT().Exec(ctx, query).Return(nil, 0, ErrorFactory())
					connBad.EXPECT().IsConnectionError(anyError).Return(true)
					connBad.EXPECT().Close().Return(nil)

					var failFlag bool

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						RunAndReturn(
							func(ctx context.Context, addr db.Addr, credentials qc.Credentials) (db.ExecutorCloser, error) {
								if !failFlag {
									failFlag = true
									return connBad, nil
								}
								return nil, db.ErrConnection
							},
						)
					return m
				},
			},
			false,
			db.ErrConnection,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := cm.NewPool(tt.fields.connector(t), seedList)
			require.NoError(t, err, "cannot create Pool")
			c, err := NewNcoClient(tt.fields.name, p)
			require.NoError(t, err, "cannot create Client")

			got := c.exec(ctx, query, credentials)

			assert.True(t, !tt.wantRows || len(got.RowSet) > 0)
			assert.ErrorIs(t, got.Error, tt.wantErrIs)
		})
	}
}
