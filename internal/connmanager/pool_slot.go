package connmanager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

// PoolSlot is container which stores concrete DB connection implementation
type PoolSlot struct {
	poolUUID, key string
	prev, next    *PoolSlot
	inUse         atomic.Bool
	log           *app.Logger

	conn db.ExecutorCloser
}

// Exec makes DB query using underlying DB connection implementation
func (s *PoolSlot) Exec(ctx context.Context, query models.Query) (rows models.RowSet, affectedRows int, err error) {
	tStart := time.Now()
	rows, affectedRows, err = s.conn.Exec(ctx, query)
	s.log.DebugContext(ctx, "exec completed", "exec_time", time.Since(tStart).String())
	return rows, affectedRows, err
}

func (s *PoolSlot) clear() (err error) {
	if s.conn != nil {
		err = s.conn.Close()
	}
	s.conn = nil
	s.key = ""
	s.prev = nil
	s.next = nil
	return err
}

// poolSlotCache is collection to store DB connections and take it for reuse.
//
// each 'key' contains a set of values, which works in FIFO manner
// with push(v) / pop(key) methods
type poolSlotCache struct {
	mx               sync.Mutex
	slotsCache       map[string][]*PoolSlot
	lruTail, lruHead *PoolSlot
}

func newPoolSlotCache() *poolSlotCache {
	// dummy values to simplify LRU usage
	lruTail, lruHead := &PoolSlot{}, &PoolSlot{}
	lruTail.next = lruHead
	lruHead.prev = lruTail

	return &poolSlotCache{
		slotsCache: make(map[string][]*PoolSlot),
		lruTail:    lruTail,
		lruHead:    lruHead,
	}
}

func (c *poolSlotCache) push(value *PoolSlot) {
	if value == nil {
		// silently ignore
		return
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	c.slotsCache[value.key] = append(c.slotsCache[value.key], value)
	// connect to 'tail'
	value.prev = c.lruHead.prev
	c.lruHead.prev.next = value
	// connect to head
	value.next = c.lruHead
	c.lruHead.prev = value
}

func (c *poolSlotCache) pop(groupKey string) *PoolSlot {
	c.mx.Lock()
	defer c.mx.Unlock()

	group := c.slotsCache[groupKey]
	if len(group) == 0 {
		return nil
	}

	slot := group[len(group)-1]
	slot.next.prev = slot.prev
	slot.prev.next = slot.next
	slot.prev, slot.next = nil, nil

	c.slotsCache[groupKey] = group[:len(group)-1]
	return slot
}

func (c *poolSlotCache) popOldest() *PoolSlot {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.lruTail.next == c.lruHead {
		return nil
	}

	slot := c.lruTail.next
	slot.next.prev = slot.prev
	slot.prev.next = slot.next
	slot.prev, slot.next = nil, nil

	return slot
}
