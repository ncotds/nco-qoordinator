package querycoordinator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	. "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	. "github.com/ncotds/nco-qoordinator/pkg/models"
)

var _ Client = (*DemoClient)(nil)

// DemoClient implements Client interface for examples only, use your oun implementation
type DemoClient struct {
	DSName  string
	DBTable RowSet
}

func (c *DemoClient) Name() string {
	return c.DSName
}

func (c *DemoClient) Exec(ctx context.Context, _ Query, _ Credentials) QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return QueryResult{Error: ctx.Err()}
	}
	return QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
}

func ExampleQueryCoordinator_Exec() {
	client1 := &DemoClient{
		DSName: "ds1",
		DBTable: RowSet{
			Columns: []string{"id", "name"},
			Rows:    [][]any{{1, "John"}, {2, "Jane"}},
		},
	}

	client2 := &DemoClient{
		DSName: "ds2",
		DBTable: RowSet{
			Columns: []string{"id", "name"},
			Rows:    [][]any{{1, "Bob"}},
		},
	}

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
	// {"ds1":{"row_set":{"Columns":["id","name"],"Rows":[[1,"John"],[2,"Jane"]]},"affected_rows":0,"error":null},"ds2":{"row_set":{"Columns":["id","name"],"Rows":[[1,"Bob"]]},"affected_rows":0,"error":null}}
}

func ExampleQueryCoordinator_ClientNames() {
	client1 := &DemoClient{"ds1", RowSet{}}
	client2 := &DemoClient{"ds2", RowSet{}}

	coordinator := NewQueryCoordinator(client1, client2)

	names := coordinator.ClientNames()

	sort.Strings(names)
	fmt.Println(names)
	// Output:
	// [ds1 ds2]
}
