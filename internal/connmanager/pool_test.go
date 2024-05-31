package connmanager_test

import (
	"context"
	"testing"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	qc "github.com/ncotds/nco-qoordinator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	. "github.com/ncotds/nco-qoordinator/internal/connmanager"
	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	mocks "github.com/ncotds/nco-qoordinator/internal/dbconnector/mocks"
)

func TestNewPool(t *testing.T) {
	mockConn := mocks.NewMockDBConnector(t)

	type args struct {
		connector db.DBConnector
		seedList  []db.Addr
		options   []PoolOption
	}
	tests := []struct {
		name      string
		args      args
		wantErrIs error
	}{
		{
			"default ok",
			args{mockConn, []db.Addr{"1"}, []PoolOption{}},
			nil,
		},
		{
			"with options ok",
			args{mockConn, []db.Addr{"1"}, []PoolOption{
				WithMaxSize(1 + FakerRandom.Intn(MaxConnectionPoolSize)),
				WithRandomFailOver(),
				WithFailBack(DurationFactory()),
			}},
			nil,
		},
		{
			"nil connector fail",
			args{nil, []db.Addr{"1"}, []PoolOption{}},
			app.ErrValidation,
		},
		{
			"empty seed list fail",
			args{mockConn, []db.Addr{}, []PoolOption{}},
			app.ErrValidation,
		},
		{
			"max size too much",
			args{mockConn, []db.Addr{"1"}, []PoolOption{
				WithMaxSize(MaxConnectionPoolSize + FakerRandom.Int()),
			}},
			app.ErrValidation,
		},
		{
			"max size too low",
			args{mockConn, []db.Addr{"1"}, []PoolOption{
				WithMaxSize(0 - FakerRandom.Int()),
			}},
			app.ErrValidation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPool(tt.args.connector, tt.args.seedList, tt.args.options...)
			assert.ErrorIs(t, err, tt.wantErrIs)
			// if expected err is nil, not-nil pool must be returned
			assert.False(t, tt.wantErrIs == nil && got == nil)
		})
	}
}

func TestPool_Acquire(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}
	seedList := SeedListFactory(3)
	poolMaxSize := 3

	type fields struct {
		connector func(t *testing.T) *mocks.MockDBConnector
		setUp     func(p *Pool)
	}
	tests := []struct {
		name      string
		fields    fields
		wantErrIs error
	}{
		{
			"empty pool ok",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(mocks.NewMockExecutorCloser(t), nil).Once()
					return m
				},
			},
			nil,
		},
		{
			"idle conn exists ok",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).
						Return(mocks.NewMockExecutorCloser(t), nil).Times(poolMaxSize)
					return m
				},
				setUp: func(p *Pool) {
					// fill Pool with test credentials, release it to reuse
					conns := make([]*PoolSlot, poolMaxSize)
					for i := range conns {
						c, _ := p.Acquire(ctx, credentials)
						conns[i] = c
					}
					for _, c := range conns {
						_ = p.Release(c)
					}
				},
			},
			nil,
		},
		{
			"idle conn not exists ok",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					c := mocks.NewMockExecutorCloser(t)
					c.EXPECT().Close().Return(nil)
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), mock.IsType(credentials)).
						Return(c, nil).Times(poolMaxSize + 1)
					return m
				},
				setUp: func(p *Pool) {
					// fill Pool with random credentials.
					// it is expected that Pool will drop one of them
					conns := make([]*PoolSlot, poolMaxSize)
					for i := range conns {
						c, _ := p.Acquire(ctx, qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()})
						conns[i] = c
					}
					for _, c := range conns {
						_ = p.Release(c)
					}
				},
			},
			nil,
		},
		{
			"all conn in use fails",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), mock.IsType(credentials)).
						Return(mocks.NewMockExecutorCloser(t), nil)
					return m
				},
				setUp: func(p *Pool) {
					// fill Pool with random credentials, do not release them
					// it is expected that Pool cannot drop anyone to acquire new slot
					for i := 0; i < poolMaxSize; i++ {
						_, err := p.Acquire(ctx, qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()})
						require.NoError(t, err)
					}
				},
			},
			app.ErrUnavailable,
		},
		{
			"cannot open conn fails",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), mock.IsType(credentials)).
						Return(nil, app.Err(app.ErrCodeUnavailable, "test"))
					return m
				},
			},
			app.ErrUnavailable,
		},
		{
			"pool is closed fails",
			fields{
				connector: func(t *testing.T) *mocks.MockDBConnector {
					return mocks.NewMockDBConnector(t)
				},
				setUp: func(p *Pool) {
					p.Close()
				},
			},
			ErrPoolClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.fields.connector(t)
			pool, err := NewPool(connector, seedList, WithMaxSize(poolMaxSize))
			require.NoError(t, err, "cannot create test Pool")
			if tt.fields.setUp != nil {
				tt.fields.setUp(pool)
			}

			gotConn, err := pool.Acquire(ctx, credentials)

			assert.ErrorIsf(t, err, tt.wantErrIs, "got: %v", err)
			assert.False(t, err == nil && gotConn == nil)
			connector.AssertExpectations(t)
		})
	}
}

func TestPool_Close(t *testing.T) {
	type fields struct {
		conn  func(t *testing.T) *mocks.MockExecutorCloser
		setUp func(p *Pool)
	}
	tests := []struct {
		name      string
		fields    fields
		wantErrIs error
	}{
		{
			"close empty ok",
			fields{
				conn: func(t *testing.T) *mocks.MockExecutorCloser {
					return mocks.NewMockExecutorCloser(t)
				},
			},
			nil,
		},
		{
			"close not empty ok",
			fields{
				conn: func(t *testing.T) *mocks.MockExecutorCloser {
					m := mocks.NewMockExecutorCloser(t)
					m.EXPECT().Close().Return(nil).Once()
					return m
				},
				setUp: func(p *Pool) {
					c, _ := p.Acquire(context.Background(), qc.Credentials{UserName: NameFactory()})
					require.NoError(t, p.Release(c))
				},
			},
			nil,
		},
		{
			"already closed fails",
			fields{
				conn: func(t *testing.T) *mocks.MockExecutorCloser {
					m := mocks.NewMockExecutorCloser(t)
					m.EXPECT().Close().Return(nil).Once()
					return m
				},
				setUp: func(p *Pool) {
					c, _ := p.Acquire(context.Background(), qc.Credentials{UserName: NameFactory()})
					require.NoError(t, p.Release(c))
					require.NoError(t, p.Close())
				},
			},
			ErrPoolClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.fields.conn(t)
			connector := mocks.NewMockDBConnector(t)
			connector.EXPECT().
				Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(qc.Credentials{})).
				Return(conn, nil).Maybe()

			pool, err := NewPool(connector, SeedListFactory(1))
			require.NoError(t, err, "cannot create test Pool")
			if tt.fields.setUp != nil {
				tt.fields.setUp(pool)
			}

			err = pool.Close()

			assert.ErrorIsf(t, err, tt.wantErrIs, "got: %v", err)
			conn.AssertExpectations(t)
		})
	}
}

func TestPool_Close_WaitAllReleased(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}

	mockConn := mocks.NewMockExecutorCloser(t)
	mockConn.EXPECT().Close().Return(nil).Once()

	connector := mocks.NewMockDBConnector(t)
	connector.EXPECT().Connect(ctx, mock.IsType(db.Addr("")), credentials).Return(mockConn, nil).Once()

	pool, err := NewPool(connector, SeedListFactory(1))
	require.NoError(t, err, "cannot create test Pool")

	conn, err := pool.Acquire(ctx, credentials)
	if !assert.NoError(t, err, "cannot acquire connection") {
		return
	}

	done := make(chan struct{})
	var closeErr error
	go func() {
		closeErr = pool.Close()
		done <- struct{}{}
	}()

	select {
	case <-done:
		t.Error("pool has closed before all connections are released")
	case <-time.After(5 * time.Millisecond):
	}

	if !assert.NoError(t, pool.Release(conn), "cannot release connection") {
		return
	}

	select {
	case <-done:
	case <-time.After(5 * time.Millisecond):
		t.Error("pool has not closed after all connections are released")
	}

	assert.NoError(t, closeErr)
}

func TestPool_Drop(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}

	type fields struct {
		conn func(t *testing.T) *mocks.MockExecutorCloser
	}
	type args struct {
		conn *PoolSlot
	}
	tests := []struct {
		name      string
		fields    fields
		setUp     func(p *Pool) args
		wantErrIs error
	}{
		{
			"to open pool ok",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Close().Return(nil).Once()
				return m
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				return args{conn: c}
			},
			nil,
		},
		{
			"to closed pool ok",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Close().Return(nil).Once()
				return m
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				go p.Close()
				return args{conn: c}
			},
			nil,
		},
		{
			"not from pool fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				return args{conn: &PoolSlot{}}
			},
			app.ErrUnknown,
		},
		{
			"not used fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Close().Return(nil).Once()
				return m
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				require.NoError(t, p.Drop(c))
				return args{conn: c}
			},
			app.ErrUnknown,
		},
		{
			"nil conn fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				return args{conn: nil}
			},
			app.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.fields.conn(t)
			connector := mocks.NewMockDBConnector(t)
			connector.EXPECT().
				Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(qc.Credentials{})).
				Return(conn, nil).Maybe()

			pool, err := NewPool(connector, SeedListFactory(1))
			require.NoError(t, err, "cannot create test Pool")

			ttArgs := tt.setUp(pool)

			err = pool.Drop(ttArgs.conn)

			assert.ErrorIsf(t, err, tt.wantErrIs, "got: %v", err)
			conn.AssertExpectations(t)
		})
	}
}

func TestPool_Release(t *testing.T) {
	ctx := context.Background()
	credentials := qc.Credentials{UserName: NameFactory(), Password: SentenceFactory()}

	type fields struct {
		conn func(t *testing.T) *mocks.MockExecutorCloser
	}
	type args struct {
		conn *PoolSlot
	}
	tests := []struct {
		name      string
		fields    fields
		setUp     func(p *Pool) args
		wantErrIs error
	}{
		{
			"to open pool ok",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				return args{conn: c}
			},
			nil,
		},
		{
			"to closed pool ok",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Close().Return(nil).Maybe()
				return m
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				go p.Close()
				return args{conn: c}
			},
			nil,
		},
		{
			"not from pool fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				return args{conn: &PoolSlot{}}
			},
			app.ErrUnknown,
		},
		{
			"not used fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				c, _ := p.Acquire(ctx, credentials)
				require.NoError(t, p.Release(c))
				return args{conn: c}
			},
			app.ErrUnknown,
		},
		{
			"nil conn fails",
			fields{conn: func(t *testing.T) *mocks.MockExecutorCloser {
				return mocks.NewMockExecutorCloser(t)
			}},
			func(p *Pool) args {
				return args{conn: nil}
			},
			app.ErrUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.fields.conn(t)
			connector := mocks.NewMockDBConnector(t)
			connector.EXPECT().
				Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(qc.Credentials{})).
				Return(conn, nil).Maybe()

			pool, err := NewPool(connector, SeedListFactory(1))
			require.NoError(t, err, "cannot create test Pool")

			ttArgs := tt.setUp(pool)

			err = pool.Release(ttArgs.conn)

			assert.ErrorIsf(t, err, tt.wantErrIs, "got: %v", err)
			conn.AssertExpectations(t)
		})
	}
}
