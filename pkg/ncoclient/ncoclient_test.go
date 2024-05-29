package ncoclient

import (
	"context"
	"testing"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	mocks "github.com/ncotds/nco-qoordinator/pkg/dbconnector/mocks"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

func TestNcoClient_exec(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}
	query := qc.Query{SQL: SentenceFactory()}
	seedList := SeedListFactory(5)

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
						Return(nil, app.Err(app.ErrCodeUnavailable, SentenceFactory()))
					return m
				},
			},
			false,
			app.ErrUnavailable,
		},
		{
			"exec query fails",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					connBad := mocks.NewMockExecutorCloser(t)
					connBad.EXPECT().Exec(ctx, query).
						Return(nil, 0, app.Err(app.ErrCodeIncorrectOperation, SentenceFactory()))

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(connBad, nil)
					return m
				},
			},
			false,
			app.ErrIncorrectOperation,
		},
		{
			"exec loose connection reconnect ok",
			fields{
				WordFactory(),
				func(t *testing.T) *mocks.MockDBConnector {
					connBad := mocks.NewMockExecutorCloser(t)
					connBad.EXPECT().Exec(ctx, query).
						Return(nil, 0, app.Err(app.ErrCodeUnavailable, SentenceFactory()))
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
					connBad.EXPECT().Exec(ctx, query).
						Return(nil, 0, app.Err(app.ErrCodeUnavailable, SentenceFactory()))
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
								return nil, app.Err(app.ErrCodeUnavailable, "test")
							},
						)
					return m
				},
			},
			false,
			app.ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewNcoClient(
				tt.fields.name,
				ClientConfig{Connector: tt.fields.connector(t), SeedList: seedList},
			)
			require.NoError(t, err, "cannot create Client")

			got := c.exec(ctx, query, credentials)

			assert.True(t, !tt.wantRows || len(got.RowSet) > 0)
			assert.ErrorIs(t, got.Error, tt.wantErrIs)
		})
	}
}
