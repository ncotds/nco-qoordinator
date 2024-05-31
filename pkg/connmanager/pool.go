package connmanager

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

const (
	// MaxConnectionPoolSize defines the maximum allowed Pool size.
	// This is a restriction of Netcool Omnibus Object Server software
	MaxConnectionPoolSize = 1000
)

var ErrPoolClosed = app.Err(app.ErrCodeUnknown, "pool is closed already")

// Pool creates and stores DB connections to reuse it
type Pool struct {
	poolUUID string
	maxSize  int

	isClosed atomic.Bool

	idleConnCache *poolSlotCache
	emptySlots    chan *PoolSlot

	poolConnector
}

// NewPool builds new ready for use Pool.
//
// Params:
// - connector: object that can open DB connection
// - seedList: list of DB instances
//
// Options:
//   - WithMaxSize - max connections
//   - WithFailBack - failover strategy used for Aggregation layer of OMNIbus cluster
//   - WithRandomFailOver - failover strategy used for Display level of OMNIBus cluster
func NewPool(connector db.DBConnector, seedList []db.Addr, options ...PoolOption) (*Pool, error) {
	if connector == nil {
		return nil, app.Err(app.ErrCodeValidation, "connector cannot be nil")
	}
	if len(seedList) == 0 {
		return nil, app.Err(app.ErrCodeValidation, "at least one seed is required")
	}

	pool := &Pool{
		maxSize:       MaxConnectionPoolSize,
		poolUUID:      uuid.NewString(),
		idleConnCache: newPoolSlotCache(),
		poolConnector: poolConnector{
			connector: connector,
			seedList:  seedList,
			// by default just try to reconnect to last success addr
			failOverSeedIdx: func(currIdx, _ int) int { return currIdx },
		},
	}

	// apply options
	for _, option := range options {
		if err := option(pool); err != nil {
			return nil, err
		}
	}

	// create available slots
	pool.emptySlots = make(chan *PoolSlot, pool.maxSize)
	for i := 0; i < pool.maxSize; i++ {
		pool.emptySlots <- &PoolSlot{poolUUID: pool.poolUUID}
	}

	return pool, nil
}

// Acquire returns connection for defined credentials from Pool.
//
// If connection not exists yet, tries to acquire free slot and establish the new one.
// If there are no free slots, tries to find the oldest idle connection and close it
func (p *Pool) Acquire(ctx context.Context, credentials qc.Credentials) (conn *PoolSlot, err error) {
	if p.isClosed.Load() {
		return nil, ErrPoolClosed
	}

	key := credentials.UserName
	conn = p.idleConnCache.pop(key)
	if conn == nil {
		// make new one
		conn, err = p.newConn(ctx, credentials)
		if err != nil {
			return nil, err
		}
	}

	conn.key = key
	conn.inUse.Store(true)
	return conn, nil
}

// Release returns connection to Pool to reuse it in future
func (p *Pool) Release(conn *PoolSlot) error {
	if err := p.markUnused(conn); err != nil {
		return err
	}

	if p.isClosed.Load() {
		_ = conn.clear()
		p.emptySlots <- conn
	} else {
		p.idleConnCache.push(conn)
	}
	return nil
}

// Drop returns connection to Pool to close it and mark slot as 'free'
func (p *Pool) Drop(conn *PoolSlot) error {
	if err := p.markUnused(conn); err != nil {
		return err
	}

	err := conn.clear()
	p.emptySlots <- conn
	return err
}

// Close marks Pool as 'closed' to prevent acquiring connections.
//
// Next, Close() waits until all connections are released to Pool and try to close it.
//
// 'Closed' Pool cannot be 'opened' again, you should create the new one if needed
func (p *Pool) Close() error {
	if !p.isClosed.CompareAndSwap(false, true) {
		return ErrPoolClosed
	}

	// drop all idle
	for c := p.idleConnCache.popOldest(); c != nil; c = p.idleConnCache.popOldest() {
		_ = c.clear()
		p.emptySlots <- c
	}
	// wait until all slots are released
	for i := 0; i < p.maxSize; i++ {
		<-p.emptySlots
	}
	close(p.emptySlots)
	return nil
}

func (p *Pool) newConn(ctx context.Context, credentials qc.Credentials) (slot *PoolSlot, err error) {
	slot, err = p.acquireSlot()
	if err != nil {
		return nil, err
	}

	conn, err := p.connect(ctx, credentials)
	if err != nil {
		p.emptySlots <- slot
		return nil, err
	}

	slot.conn = conn
	return slot, nil
}

func (p *Pool) acquireSlot() (slot *PoolSlot, err error) {
	select {
	case slot = <-p.emptySlots:
		return slot, nil
	default:
	}

	// try to get the last used connection to replace it
	slot = p.idleConnCache.popOldest()
	if slot != nil {
		_ = slot.clear()
		return slot, nil
	}

	return slot, app.Err(app.ErrCodeUnavailable, "connections limit exceed")
}

func (p *Pool) markUnused(slot *PoolSlot) (err error) {
	switch {
	case slot == nil:
		err = app.Err(app.ErrCodeUnknown, "cannot release connection, nil slot")
	case slot.poolUUID != p.poolUUID:
		err = app.Err(app.ErrCodeUnknown, "cannot release connection, not from given pool")
	case !slot.inUse.CompareAndSwap(true, false):
		err = app.Err(app.ErrCodeUnknown, "cannot release connection, not in use")
	}
	return err
}

// PoolOption represent optional parameter to configure new Pool
type PoolOption func(pool *Pool) error

// WithMaxSize option sets max opened connections that Pool can store.
//
// When you ask new connection from full Pool,
// it tries to return the idle connection if exists
// or to close the oldest unused one.
//
// maxSize value should be between 0 and MaxConnectionPoolSize
func WithMaxSize(maxSize int) PoolOption {
	return func(pool *Pool) error {
		if maxSize < 1 || maxSize > MaxConnectionPoolSize {
			return app.Err(app.ErrCodeValidation,
				fmt.Sprintf(
					"invalid pool size: %d, allowed values are: 1..%d",
					maxSize,
					MaxConnectionPoolSize,
				),
			)
		}
		pool.maxSize = maxSize
		return nil
	}
}

// WithFailBack option sets Pool's failover policy to 'FailBack'
//
// It means that when current connection is loosed, Pool firstly tries to connect
// to any random address from seed list, except the current one.
// If those address fails, then Pool continue with next seed...
// and makes one attempt to each address from seed list until success or
// attempts to all addresses will fail
func WithFailBack(delay time.Duration) PoolOption {
	return func(pool *Pool) error {
		pool.failOverSeedIdx = nextSeedWithFailBack(delay)
		return nil
	}
}

// WithRandomFailOver option sets Pool's failover strategy to 'RandomFailOver'
//
// It means that when current connection is loosed, Pool firstly tries to connect
// to any random address from seed list, except the current one.
// If those address fails, then Pool continue with next seed...
// and makes one attempt to each address from seed list until success or
// attempts to all addresses will fail
func WithRandomFailOver() PoolOption {
	return func(pool *Pool) error {
		pool.failOverSeedIdx = nextSeedRandom()
		return nil
	}
}
