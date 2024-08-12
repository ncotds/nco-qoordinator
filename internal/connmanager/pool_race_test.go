package connmanager_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	db "github.com/ncotds/nco-lib/dbconnector"
	mocks "github.com/ncotds/nco-lib/dbconnector/mocks"
	. "github.com/ncotds/nco-qoordinator/internal/connmanager"
)

const concurrency = 1000

func TestPool_Acquire_Concurrent(t *testing.T) {
	connector := mocks.NewMockDBConnector(t)
	connector.EXPECT().Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(db.Credentials{})).
		Return(mocks.NewMockExecutorCloser(t), nil)

	pool, err := NewPool(connector, SeedListFactory(1), WithMaxSize(concurrency))
	require.NoError(t, err, "cannot create test Pool")

	ctx, cancel := context.WithCancel(context.Background())
	errs := make([]error, concurrency)

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			<-ctx.Done()
			_, errs[i] = pool.Acquire(
				context.Background(),
				db.Credentials{UserName: NameFactory(), Password: SentenceFactory()},
			)
			wg.Done()
		}(i)
	}

	cancel() // unlock goroutines
	wg.Wait()

	results := map[error]int{}
	for _, val := range errs {
		results[val]++
	}

	assert.Equalf(t, concurrency, results[nil], "results: %v", results)
}

func TestPool_Close_Concurrent(t *testing.T) {
	pool, err := NewPool(mocks.NewMockDBConnector(t), SeedListFactory(1))
	require.NoError(t, err, "cannot create test Pool")

	ctx, cancel := context.WithCancel(context.Background())
	errs := make([]error, concurrency)

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			<-ctx.Done()
			errs[i] = pool.Close()
			wg.Done()
		}(i)
	}

	cancel() // unlock goroutines
	wg.Wait()

	results := map[error]int{}
	for _, val := range errs {
		results[val]++
	}

	assert.Equalf(t, 1, results[nil], "results: %v", results)
	assert.Equalf(t, concurrency-1, results[ErrPoolClosed], "results: %v", results)
}

func TestPool_Drop_Concurrent(t *testing.T) {
	mockConn := mocks.NewMockExecutorCloser(t)
	mockConn.EXPECT().Close().Return(nil)

	connector := mocks.NewMockDBConnector(t)
	connector.EXPECT().Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(db.Credentials{})).
		Return(mockConn, nil)

	pool, err := NewPool(connector, SeedListFactory(1), WithMaxSize(concurrency))
	require.NoError(t, err, "cannot create test Pool")

	conns := make([]*PoolSlot, concurrency)
	for i := 0; i < concurrency; i++ {
		conns[i], _ = pool.Acquire(
			context.Background(),
			db.Credentials{UserName: NameFactory(), Password: SentenceFactory()},
		)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make([]error, concurrency)

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			<-ctx.Done()
			errs[i] = pool.Drop(conns[i])
			wg.Done()
		}(i)
	}

	cancel() // unlock goroutines
	wg.Wait()

	results := map[error]int{}
	for _, val := range errs {
		results[val]++
	}

	assert.Equalf(t, concurrency, results[nil], "results: %v", results)
}

func TestPool_Release_Concurrent(t *testing.T) {
	mockConn := mocks.NewMockExecutorCloser(t)

	connector := mocks.NewMockDBConnector(t)
	connector.EXPECT().Connect(context.Background(), mock.IsType(db.Addr("")), mock.IsType(db.Credentials{})).
		Return(mockConn, nil)

	pool, err := NewPool(connector, SeedListFactory(1), WithMaxSize(concurrency))
	require.NoError(t, err, "cannot create test Pool")

	conns := make([]*PoolSlot, concurrency)
	for i := 0; i < concurrency; i++ {
		conns[i], _ = pool.Acquire(
			context.Background(),
			db.Credentials{UserName: NameFactory(), Password: SentenceFactory()},
		)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errs := make([]error, concurrency)

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			<-ctx.Done()
			errs[i] = pool.Release(conns[i])
			wg.Done()
		}(i)
	}

	cancel() // unlock goroutines
	wg.Wait()

	results := map[error]int{}
	for _, val := range errs {
		results[val]++
	}

	assert.Equalf(t, concurrency, results[nil], "results: %v", results)
}
