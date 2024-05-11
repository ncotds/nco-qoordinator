package connmanager

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

var randomizer = rand.New(rand.NewSource(time.Now().UnixNano()))

// poolConnector handles failover strategies
type poolConnector struct {
	connector       db.DBConnector
	seedList        []db.Addr
	currentSeedIdx  atomic.Int32
	failOverSeedIdx func(currIdx, seedsCount int) (nextIdx int)
}

// connect provides thread-safe way to open new connection using defined failover strategy
func (c *poolConnector) connect(ctx context.Context, credentials qc.Credentials) (conn db.ExecutorCloser, err error) {
	err = fmt.Errorf("%w: there is no any connection to try", db.ErrConnection)
	nextIdx := c.failOverSeedIdx(int(c.currentSeedIdx.Load()), len(c.seedList))
	for i, addr := range iterSlice(c.seedList, nextIdx) {
		conn, err = c.connector.Connect(ctx, addr, credentials)
		if err == nil {
			c.currentSeedIdx.Store(int32((nextIdx + i) % len(c.seedList)))
			return conn, err
		}
	}
	if !errors.Is(err, ErrConnManager) {
		err = errors.Join(ErrConnManager, err)
	}
	return conn, err
}

func iterSlice[S ~[]E, E any](s S, fromIdx int) S {
	if len(s) == 0 {
		return s
	}
	fromIdx = (len(s) + fromIdx%len(s)) % len(s)
	return append(s[fromIdx:], s[:fromIdx]...)
}

func nextSeedWithFailBack(failBackDelay time.Duration) func(currIdx, seedsCount int) (nextIdx int) {
	var lastFailedAt time.Time

	return func(currIdx, seedsCount int) (nextIdx int) {
		if lastFailedAt.IsZero() {
			lastFailedAt = time.Now()
		}
		if seedsCount > 0 && time.Now().Before(lastFailedAt.Add(failBackDelay)) {
			/*
				fix too low or too high currIdx value:
				- seedsCount + currIdx%seedsCount
					reduces any currIdx, negative or positive, into (0, 2*seedsCount) interval
				- (...) % seedsCount
					moves value to [0, seedsCount)
			*/
			nextIdx = (seedsCount + currIdx%seedsCount) % seedsCount
		}
		lastFailedAt = time.Now()
		return nextIdx
	}
}

func nextSeedRandom() func(currIdx, seedsCount int) (nextIdx int) {
	return func(currIdx, seedsCount int) (nextIdx int) {
		if seedsCount > 1 {
			randomOffset := 1 + randomizer.Intn(seedsCount-1)
			/*
				fix too low or too high currIdx value:
				- seedsCount + currIdx%seedsCount
					reduces any currIdx, negative or positive, into (0, 2*seedsCount) interval
				- (...) + randomOffset
					adds some value to select another index than current
				- (...) % seedsCount
					moves result to [0, seedsCount)
			*/
			nextIdx = (seedsCount + currIdx%seedsCount + randomOffset) % seedsCount
		}
		return nextIdx
	}
}
