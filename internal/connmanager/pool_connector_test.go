package connmanager

import (
	"context"
	"testing"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	qc "github.com/ncotds/nco-qoordinator/pkg/models"
	"github.com/stretchr/testify/assert"

	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	mocks "github.com/ncotds/nco-qoordinator/internal/dbconnector/mocks"
)

func Test_iterSlice(t *testing.T) {
	type args struct {
		s       []any
		fromIdx int
	}
	type testCase struct {
		name string
		args args
		want []any
	}
	tests := []testCase{
		{
			"from first",
			args{[]any{1, 2, 3, 4, 5}, 0},
			[]any{1, 2, 3, 4, 5},
		},
		{
			"from middle",
			args{[]any{1, 2, 3, 4, 5}, 2},
			[]any{3, 4, 5, 1, 2},
		},
		{
			"from last",
			args{[]any{1, 2, 3, 4, 5}, 4},
			[]any{5, 1, 2, 3, 4},
		},
		{
			"from negative",
			args{[]any{1, 2, 3, 4, 5}, -3},
			[]any{3, 4, 5, 1, 2},
		},
		{
			"from out of scope positive",
			args{[]any{1, 2, 3, 4, 5}, 7},
			[]any{3, 4, 5, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, iterSlice(tt.args.s, tt.args.fromIdx))
		})
	}
}

func Test_nextSeedRandom(t *testing.T) {
	type args struct {
		startIdx, seedsCount int
	}
	tests := []struct {
		name     string
		args     args
		assertFn func(t *testing.T, got []int, args args)
	}{
		{
			"more that one seed",
			args{FakerRandom.Intn(10), 5},
			func(t *testing.T, got []int, args args) {
				prev := args.startIdx
				for i, val := range got {
					// expected: each call returns value other than previous but in [0, seedCount)
					assert.NotEqualf(t, prev, val, "%d elem in %v", i, got)
					assert.GreaterOrEqualf(t, val, 0, "%d elem in %v", i, got)
					assert.Lessf(t, val, args.seedsCount, "%d elem in %v", i, got)
					prev = val
				}
			},
		},
		{
			"negative start index",
			args{0 - FakerRandom.Intn(10), 5},
			func(t *testing.T, got []int, args args) {
				prev := args.startIdx
				for i, val := range got {
					// expected: each call returns value other than previous but in [0, seedCount)
					assert.NotEqualf(t, prev, val, "%d elem in %v", i, got)
					assert.GreaterOrEqualf(t, val, 0, "%d elem in %v", i, got)
					assert.Lessf(t, val, args.seedsCount, "%d elem in %v", i, got)
					prev = val
				}
			},
		},
		{
			"one seed",
			args{FakerRandom.Intn(10), 1},
			func(t *testing.T, got []int, args args) {
				// only one seed exists, so each call returns idx==0
				for i, val := range got {
					assert.Zerof(t, val, "%d elem in %v", i, got)
				}
			},
		},
		{
			"zero seeds",
			args{FakerRandom.Intn(10), 0},
			func(t *testing.T, got []int, args args) {
				// there is no seeds, so each call zero int value
				for i, val := range got {
					assert.Zerof(t, val, "%d elem in %v", i, got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := nextSeedRandom()

			got := []int{tt.args.startIdx}
			for lastIdx := 0; lastIdx < 10; lastIdx++ {
				got = append(got, next(got[lastIdx], tt.args.seedsCount))
			}
			tt.assertFn(t, got[1:], tt.args)
		})
	}
}

func Test_nextSeedWithFailBack(t *testing.T) {
	type args struct {
		startIdx, seedsCount        int
		failBackDelay, callInterval time.Duration
	}
	tests := []struct {
		name     string
		args     args
		assertFn func(t *testing.T, got []int, args args)
	}{
		{
			"before failback delay expired",
			args{FakerRandom.Intn(10), 5, 1 * time.Minute, 1 * time.Millisecond},
			func(t *testing.T, got []int, args args) {
				expected := args.startIdx % args.seedsCount
				for i, val := range got {
					// failback delay too much, expected stay on current index
					assert.Equalf(t, expected, val, "%d elem in %v", i, got)
				}
			},
		},
		{
			"after failback delay expired",
			args{FakerRandom.Intn(10), 5, 1 * time.Nanosecond, 1 * time.Millisecond},
			func(t *testing.T, got []int, args args) {
				for i, val := range got {
					// failback delay too low, expected back to the first elem
					assert.Zerof(t, val, "%d elem in %v", i, got)
				}
			},
		},
		{
			"one seed",
			args{FakerRandom.Intn(10), 1, 5 * time.Millisecond, 1 * time.Millisecond},
			func(t *testing.T, got []int, args args) {
				for i, val := range got {
					// only one seed, expected always use it
					assert.Zerof(t, val, "%d elem in %v", i, got)
				}
			},
		},
		{
			"zero seeds",
			args{FakerRandom.Intn(10), 0, 5 * time.Millisecond, 1 * time.Millisecond},
			func(t *testing.T, got []int, args args) {
				// there is no seeds, so each call zero int value
				for i, val := range got {
					assert.Zerof(t, val, "%d elem in %v", i, got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := nextSeedWithFailBack(tt.args.failBackDelay)

			got := []int{tt.args.startIdx}
			for lastIdx := 0; lastIdx < 10; lastIdx++ {
				time.Sleep(1 * time.Millisecond)
				got = append(got, next(got[lastIdx], tt.args.seedsCount))
			}
			tt.assertFn(t, got[1:], tt.args)
		})
	}
}

func Test_poolConnector_connect(t *testing.T) {
	mockConn := mocks.NewMockExecutorCloser(t)
	mockErr := app.Err(app.ErrCodeUnavailable, "test")
	ctx := context.Background()
	credentials := qc.Credentials{}
	seedList := SeedListFactory(5)

	type fields struct {
		connector      func(t *testing.T) db.DBConnector
		seedList       []db.Addr
		currentSeedIdx int
	}
	tests := []struct {
		name        string
		fields      fields
		wantConn    db.ExecutorCloser
		wantErrIs   error
		wantAddrIdx int32
	}{
		{
			"first addr ok",
			fields{
				seedList: seedList,
				connector: func(t *testing.T) db.DBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, seedList[0], credentials).Return(mockConn, nil).Once()
					return m
				},
			},
			mockConn,
			nil,
			0,
		},
		{
			"first addr fails, second ok",
			fields{
				seedList: seedList,
				connector: func(t *testing.T) db.DBConnector {
					m := mocks.NewMockDBConnector(t)
					m.EXPECT().Connect(ctx, seedList[0], credentials).Return(nil, mockErr).Once()
					m.EXPECT().Connect(ctx, seedList[1], credentials).Return(mockConn, nil).Once()
					return m
				},
			},
			mockConn,
			nil,
			1,
		},
		{
			"all addr fails",
			fields{
				seedList: seedList,
				connector: func(t *testing.T) db.DBConnector {
					m := mocks.NewMockDBConnector(t)
					for i := 0; i < len(seedList); i++ {
						m.EXPECT().Connect(ctx, seedList[i], credentials).Return(nil, mockErr).Once()
					}
					return m
				},
			},
			nil,
			app.ErrUnavailable,
			0, // couldn't reconnect, stay on current addr
		},
		{
			"all addr fails, unknown err is wrapped",
			fields{
				seedList: seedList,
				connector: func(t *testing.T) db.DBConnector {
					m := mocks.NewMockDBConnector(t)
					for i := 0; i < len(seedList); i++ {
						m.EXPECT().Connect(ctx, seedList[i], credentials).Return(nil, ErrorFactory()).Once()
					}
					return m
				},
			},
			nil,
			app.ErrUnavailable,
			0, // couldn't reconnect, stay on current addr
		},
		{
			"no conn to try",
			fields{
				seedList: SeedListFactory(0),
				connector: func(t *testing.T) db.DBConnector {
					return nil
				},
			},
			nil,
			app.ErrUnavailable,
			0, // couldn't reconnect, stay on current addr
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &poolConnector{
				connector:       tt.fields.connector(t),
				seedList:        tt.fields.seedList,
				failOverSeedIdx: func(currIdx, _ int) (nextIdx int) { return currIdx },
				log:             app.NewLogger(nil),
			}
			c.currentSeedIdx.Store(int32(tt.fields.currentSeedIdx))

			gotConn, err := c.connect(ctx, credentials)

			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantConn, gotConn)
			assert.Equal(t, tt.wantAddrIdx, c.currentSeedIdx.Load())
		})
	}
}
