package querycoordinator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	. "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

// DemoClient implements Client interface for examples only, use your oun implementation
type DemoClient struct {
	DSName  string
	DBTable []QueryResultRow
}

func (c *DemoClient) Name() string {
	return c.DSName
}

func (c *DemoClient) Exec(ctx context.Context, query Query, user Credentials) QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return QueryResult{Error: ctx.Err()}
	}
	return QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
}

func ExampleQueryCoordinator_Exec() {
	client1 := &DemoClient{"ds1", []QueryResultRow{
		{"id": 1, "name": "John"},
		{"id": 2, "name": "Jane"},
	}}

	client2 := &DemoClient{"ds2", []QueryResultRow{
		{"id": 1, "name": "Bob"},
	}}

	coordinator := NewQueryCoordinator(client1, client2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := coordinator.Exec(
		ctx,
		Query{SQL: "select * from dbtable"},
		Credentials{UserName: "alice", Password: "secret"},
	)

	resultJson, _ := json.Marshal(result)
	fmt.Println(string(resultJson))
	// Output:
	// {"ds1":{"rowset":[{"id":1,"name":"John"},{"id":2,"name":"Jane"}],"affected_rows":0,"error":null},"ds2":{"rowset":[{"id":1,"name":"Bob"}],"affected_rows":0,"error":null}}
}

func ExampleQueryCoordinator_ClientNames() {
	client1 := &DemoClient{"ds1", []QueryResultRow{}}
	client2 := &DemoClient{"ds2", []QueryResultRow{}}

	coordinator := NewQueryCoordinator(client1, client2)

	names := coordinator.ClientNames()

	sort.Strings(names)
	fmt.Println(names)
	// Output:
	// [ds1 ds2]
}
