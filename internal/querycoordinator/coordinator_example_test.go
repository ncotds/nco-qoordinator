package querycoordinator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	db "github.com/ncotds/nco-lib/dbconnector"

	. "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
)

var _ Client = (*DemoClient)(nil)

// DemoClient implements Client interface for examples only, use your oun implementation
type DemoClient struct {
	DSName  string
	DBTable db.RowSet
}

func (c *DemoClient) Name() string {
	return c.DSName
}

func (c *DemoClient) Exec(ctx context.Context, _ db.Query, _ db.Credentials) db.QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return db.QueryResult{Error: ctx.Err()}
	}
	return db.QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
}

func ExampleQueryCoordinator_Exec() {
	client1 := &DemoClient{
		DSName: "ds1",
		DBTable: db.RowSet{
			Columns: []string{"id", "name"},
			Rows:    [][]any{{1, "John"}, {2, "Jane"}},
		},
	}

	client2 := &DemoClient{
		DSName: "ds2",
		DBTable: db.RowSet{
			Columns: []string{"id", "name"},
			Rows:    [][]any{{1, "Bob"}},
		},
	}

	coordinator := NewQueryCoordinator(client1, client2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := coordinator.Exec(
		ctx,
		db.Query{SQL: "select * from dbtable"},
		db.Credentials{UserName: "alice", Password: "secret"},
	)

	resultJson, _ := json.Marshal(result)
	fmt.Println(string(resultJson))
	// Output:
	// {"ds1":{"row_set":{"Columns":["id","name"],"Rows":[[1,"John"],[2,"Jane"]]},"affected_rows":0,"error":null},"ds2":{"row_set":{"Columns":["id","name"],"Rows":[[1,"Bob"]]},"affected_rows":0,"error":null}}
}

func ExampleQueryCoordinator_ClientNames() {
	client1 := &DemoClient{"ds1", db.RowSet{}}
	client2 := &DemoClient{"ds2", db.RowSet{}}

	coordinator := NewQueryCoordinator(client1, client2)

	names := coordinator.ClientNames()

	sort.Strings(names)
	fmt.Println(names)
	// Output:
	// [ds1 ds2]
}
