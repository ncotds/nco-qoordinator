package ncoclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	mocks "github.com/ncotds/nco-qoordinator/internal/dbconnector/mocks"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

func TestNcoClient_exec(t *testing.T) {
	ctx := context.Background()
	credentials := models.Credentials{UserName: NameFactory(), Password: SentenceFactory()}
	query := models.Query{SQL: SentenceFactory()}
	seedList := SeedListFactory(5)
	anyValueCtx := mock.AnythingOfType("*context.valueCtx")

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
					conn.EXPECT().Exec(anyValueCtx, query).Return(
						models.RowSet{Columns: []string{WordFactory()}, Rows: [][]any{{WordFactory()}}},
						0,
						nil,
					).Once()

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(anyValueCtx, mock.IsType(db.Addr("")), credentials).
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
					m.EXPECT().Connect(anyValueCtx, mock.IsType(db.Addr("")), credentials).
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
					connBad.EXPECT().Exec(anyValueCtx, query).
						Return(models.RowSet{}, 0, app.Err(app.ErrCodeIncorrectOperation, SentenceFactory()))

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(anyValueCtx, mock.IsType(db.Addr("")), credentials).
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
					connBad.EXPECT().Exec(anyValueCtx, query).
						Return(models.RowSet{}, 0, app.Err(app.ErrCodeUnavailable, SentenceFactory()))
					connBad.EXPECT().Close().Return(nil)

					connOk := mocks.NewMockExecutorCloser(t)
					connOk.EXPECT().Exec(anyValueCtx, query).Return(
						models.RowSet{Columns: []string{WordFactory()}, Rows: [][]any{{WordFactory()}}},
						0,
						nil,
					)

					var failFlag bool

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(anyValueCtx, mock.IsType(db.Addr("")), credentials).
						RunAndReturn(
							func(ctx context.Context, addr db.Addr, credentials models.Credentials) (db.ExecutorCloser, error) {
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
					connBad.EXPECT().Exec(anyValueCtx, query).
						Return(models.RowSet{}, 0, app.Err(app.ErrCodeUnavailable, SentenceFactory()))
					connBad.EXPECT().Close().Return(nil)

					var failFlag bool

					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(anyValueCtx, mock.IsType(db.Addr("")), credentials).
						RunAndReturn(
							func(ctx context.Context, addr db.Addr, credentials models.Credentials) (db.ExecutorCloser, error) {
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

			assert.True(t, !tt.wantRows || len(got.RowSet.Rows) > 0)
			assert.ErrorIs(t, got.Error, tt.wantErrIs)
		})
	}
}
