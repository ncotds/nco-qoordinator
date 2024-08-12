package ncoclient_test

import (
	"context"
	"fmt"
	"time"

	db "github.com/ncotds/nco-lib/dbconnector"

	nc "github.com/ncotds/nco-qoordinator/internal/ncoclient"
)

var (
	_ db.DBConnector    = (*DemoConnector)(nil)
	_ db.ExecutorCloser = (*DemoConnection)(nil)
)

// DemoConnector implements dbconnector.DBConnector interface for examples only, use your own implementation
type DemoConnector struct {
	Conn *DemoConnection
}

func (dc *DemoConnector) Connect(_ context.Context, _ db.Addr, _ db.Credentials) (conn db.ExecutorCloser, err error) {
	// working hard
	<-time.After(10 * time.Millisecond)
	return dc.Conn, err
}

// DemoConnection implements dbconnector.ExecutorCloser interface for examples only, use your own implementation
type DemoConnection struct {
	Data     db.RowSet
	Affected int
	Err      error
}

func (c *DemoConnection) Exec(_ context.Context, _ db.Query) (db.RowSet, int, error) {
	<-time.After(10 * time.Millisecond) // working hard
	return c.Data, c.Affected, c.Err
}

func (c *DemoConnection) Close() error {
	return nil
}

func ExampleNewNcoClient() {
	conf := nc.ClientConfig{
		Connector: &DemoConnector{},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	client, err := nc.NewNcoClient("AGG1", conf)

	fmt.Printf("%T, %v", client, err)
	// Output:
	// *ncoclient.NcoClient, <nil>
}

func ExampleNewNcoClient_empty_name_fail() {
	conf := nc.ClientConfig{
		Connector: &DemoConnector{},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	_, err := nc.NewNcoClient("", conf)

	fmt.Println(err)
	// Output:
	// ERR_VALIDATION: invalid client config, empty name
}

func ExampleNcoClient_Name() {
	conf := nc.ClientConfig{
		Connector: &DemoConnector{},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	client, _ := nc.NewNcoClient("AGG1", conf)

	name := client.Name()

	fmt.Println(name)
	// Output:
	// AGG1
}

func ExampleNcoClient_Exec() {
	demoConn := &DemoConnection{
		Data: db.RowSet{
			Columns: []string{"col1", "col2"},
			Rows:    [][]any{{"data1", 3}, {"data2", 5}},
		},
	}

	conf := nc.ClientConfig{
		Connector: &DemoConnector{demoConn},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	client, _ := nc.NewNcoClient("AGG1", conf)

	ctx := context.Background()
	query := db.Query{SQL: "select 1"}
	credentials := db.Credentials{UserName: "someuser", Password: "superpass"}

	result := client.Exec(ctx, query, credentials)

	fmt.Println(result)
	// Output:
	// {{[col1 col2] [[data1 3] [data2 5]]} 0 <nil>}
}

func ExampleNcoClient_Exec_cancel() {
	demoConn := &DemoConnection{
		Data: db.RowSet{
			Columns: []string{"col1", "col2"},
			Rows:    [][]any{{"data1", 3}, {"data2", 5}},
		},
	}
	conf := nc.ClientConfig{
		Connector: &DemoConnector{demoConn},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	client, _ := nc.NewNcoClient("AGG1", conf)

	ctx, cancel := context.WithCancel(context.Background())
	query := db.Query{SQL: "select 1"}
	credentials := db.Credentials{UserName: "someuser", Password: "superpass"}

	cancel()
	result := client.Exec(ctx, query, credentials)

	fmt.Println(result)
	// Output:
	// {{[] []} 0 context canceled}
}

func ExampleNcoClient_Close() {
	demoConn := &DemoConnection{
		Data: db.RowSet{
			Columns: []string{"col1", "col2"},
			Rows:    [][]any{{"data1", 3}, {"data2", 5}},
		},
	}

	conf := nc.ClientConfig{
		Connector: &DemoConnector{demoConn},
		SeedList:  []db.Addr{"host1", "host2"},
	}
	client, _ := nc.NewNcoClient("AGG1", conf)

	ctx := context.Background()
	query := db.Query{SQL: "select 1"}
	credentials := db.Credentials{UserName: "someuser", Password: "superpass"}

	errClose := client.Close()
	result := client.Exec(ctx, query, credentials)
	name := client.Name()

	fmt.Println(errClose)
	fmt.Println(result)
	fmt.Println(name)
	// Output:
	// <nil>
	// {{[] []} 0 ERR_UNKNOWN: pool is closed already}
	// AGG1
}
