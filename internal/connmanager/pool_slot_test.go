package connmanager

import (
	"context"
	"testing"

	"github.com/google/uuid"
	db "github.com/ncotds/nco-lib/dbconnector"
	mocks "github.com/ncotds/nco-lib/dbconnector/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

func TestPoolSlot_Exec(t *testing.T) {
	ctx := context.Background()
	query := db.Query{SQL: SentenceFactory()}
	resultRows := db.RowSet{
		Columns: []string{WordFactory(), WordFactory()},
		Rows: [][]any{
			{WordFactory(), SentenceFactory()},
			{WordFactory(), SentenceFactory()},
			{WordFactory(), SentenceFactory()},
		},
	}

	type args struct {
		conn func(t *testing.T) db.ExecutorCloser
	}
	tests := []struct {
		name         string
		args         args
		wantRows     db.RowSet
		wantAffected int
		wantErrIs    error
	}{
		{
			"no errors",
			args{conn: func(t *testing.T) db.ExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Exec(ctx, query).Return(resultRows, len(resultRows.Rows), nil).Once()
				return m
			}},
			resultRows,
			len(resultRows.Rows),
			nil,
		},
		{
			"connection error",
			args{conn: func(t *testing.T) db.ExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Exec(ctx, query).
					Return(db.RowSet{}, 0, app.Err(app.ErrCodeUnavailable, SentenceFactory())).
					Once()
				return m
			}},
			db.RowSet{},
			0,
			app.ErrUnavailable,
		},
		{
			"non-connection error",
			args{conn: func(t *testing.T) db.ExecutorCloser {
				m := mocks.NewMockExecutorCloser(t)
				m.EXPECT().Exec(ctx, query).
					Return(db.RowSet{}, 0, app.Err(app.ErrCodeIncorrectOperation, SentenceFactory())).Once()
				return m
			}},
			db.RowSet{},
			0,
			app.ErrIncorrectOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &PoolSlot{conn: tt.args.conn(t), log: app.NewLogger(nil)}
			rows, affected, err := conn.Exec(ctx, query)

			assert.Equal(t, tt.wantRows, rows)
			assert.Equal(t, tt.wantAffected, affected)
			assert.ErrorIs(t, err, tt.wantErrIs)
		})
	}

}

func TestPoolSlot_clear(t *testing.T) {
	poolUUID := uuid.NewString()
	wantErr := ErrorFactory()

	s := &PoolSlot{
		prev:     &PoolSlot{},
		next:     &PoolSlot{},
		poolUUID: poolUUID,
		key:      WordFactory(),
		conn: func() db.ExecutorCloser {
			m := mocks.NewMockExecutorCloser(t)
			m.EXPECT().Close().Return(wantErr)
			return m
		}(),
	}

	gotErr := s.clear()

	assert.Equal(t, &PoolSlot{poolUUID: poolUUID}, s)
	assert.Equal(t, wantErr, gotErr)

	// clear() should be idempotent
	assert.NoError(t, s.clear(), "second call")
	assert.NoError(t, s.clear(), "third call")
}

func Test_poolSlotCache_pop(t *testing.T) {
	cache := newPoolSlotCache()
	existingSlots := make([]*PoolSlot, 5)
	for i := 0; i < len(existingSlots); i++ {
		existingSlots[i] = &PoolSlot{key: SentenceFactory()}
		cache.push(existingSlots[i])
	}
	anyExistingIdx := FakerRandom.Intn(len(existingSlots))

	type args struct {
		groupKey string
	}
	tests := []struct {
		name string
		args args
		want *PoolSlot
	}{
		{
			"existing key",
			args{existingSlots[anyExistingIdx].key},
			existingSlots[anyExistingIdx],
		},
		{
			"missed key",
			args{SentenceFactory()},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlot := cache.pop(tt.args.groupKey)
			gotSlotSecond := cache.pop(tt.args.groupKey)

			assert.Equal(t, tt.want, gotSlot)
			assert.Nil(t, gotSlotSecond)
		})
	}
}

func Test_poolSlotCache_popOldest(t *testing.T) {
	cache := newPoolSlotCache()
	existingSlots := make([]*PoolSlot, 5)
	for i := 0; i < len(existingSlots); i++ {
		existingSlots[i] = &PoolSlot{key: SentenceFactory()}
		cache.push(existingSlots[i])
	}

	type fields struct {
		cache *poolSlotCache
	}
	tests := []struct {
		name   string
		fields fields
		want   []*PoolSlot
	}{
		{
			"from full cache",
			fields{cache: cache},
			existingSlots,
		},
		{
			"from empty cache",
			fields{cache: newPoolSlotCache()},
			[]*PoolSlot{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]*PoolSlot, 0)
			for slot := tt.fields.cache.popOldest(); slot != nil; slot = tt.fields.cache.popOldest() {
				result = append(result, slot)
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_poolSlotCache_push(t *testing.T) {
	type args struct {
		values []*PoolSlot
	}
	tests := []struct {
		name     string
		args     args
		assertFn func(t *testing.T, cache *poolSlotCache, args args)
	}{
		{
			"random keys",
			args{[]*PoolSlot{
				{key: SentenceFactory()},
				{key: SentenceFactory()},
				{key: SentenceFactory()},
				{key: SentenceFactory()},
				{key: SentenceFactory()},
			}},
			func(t *testing.T, cache *poolSlotCache, args args) {
				assert.True(t, len(cache.slotsCache) > 0 && len(cache.slotsCache) == len(args.values))
				for _, value := range args.values {
					group, ok := cache.slotsCache[value.key]
					assert.True(t, ok)
					assert.Contains(t, group, value)
				}
			},
		},
		{
			"same key",
			args{func() (values []*PoolSlot) {
				key := SentenceFactory()
				for i := 0; i < 5; i++ {
					values = append(values, &PoolSlot{key: key})
				}
				return values
			}()},
			func(t *testing.T, cache *poolSlotCache, args args) {
				assert.Len(t, cache.slotsCache, 1)
				key := args.values[0].key
				group, ok := cache.slotsCache[key]
				assert.True(t, ok)
				for _, value := range args.values {
					assert.Contains(t, group, value)
				}
			},
		},
		{
			"empty key",
			args{[]*PoolSlot{{key: ""}}},
			func(t *testing.T, cache *poolSlotCache, args args) {
				group, ok := cache.slotsCache[""]
				assert.True(t, ok)
				assert.Contains(t, group, args.values[0])
			},
		},
		{
			"nil slot ignored",
			args{[]*PoolSlot{nil}},
			func(t *testing.T, cache *poolSlotCache, args args) {
				assert.Empty(t, cache.slotsCache)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newPoolSlotCache()
			for _, value := range tt.args.values {
				c.push(value)
			}

			tt.assertFn(t, c, tt.args)
		})
	}
}
