package connmanager_test

import (
	"context"
	"fmt"
	"time"

	db "github.com/ncotds/nco-lib/dbconnector"

	cm "github.com/ncotds/nco-qoordinator/internal/connmanager"
)

var _ db.DBConnector = (*DemoConnector)(nil)

// DemoConnector implements dbconnector.DBConnector interface for examples only, use your own implementation
type DemoConnector struct {
}

func (dc *DemoConnector) Connect(_ context.Context, _ db.Addr, _ db.Credentials) (conn db.ExecutorCloser, err error) {
	return nil, err
}

func ExampleNewPool() {
	pool, err := cm.NewPool(
		&DemoConnector{},
		[]db.Addr{"host1:4100", "host2:4100", "host3:4100"},
		// optional params
		cm.WithMaxSize(10),
		cm.WithFailBack(1*time.Minute), // or cm.WithRandomFailOver(),
	)

	fmt.Printf("%T, %v\n", pool, err)
	// Output:
	// *connmanager.Pool, <nil>
}

func ExamplePool_Acquire() {
	pool, _ := cm.NewPool(
		&DemoConnector{},
		[]db.Addr{"host1:4100", "host2:4100", "host3:4100"},
		cm.WithMaxSize(2),
	)
	credentials := db.Credentials{
		UserName: "someuser",
		Password: "superpass",
	}

	conn1, err1 := pool.Acquire(context.Background(), credentials)
	conn2, err2 := pool.Acquire(context.Background(), credentials)
	conn3, err3 := pool.Acquire(context.Background(), credentials)

	fmt.Printf("%T, %v\n", conn1, err1)
	fmt.Printf("%T, %v\n", conn2, err2)
	fmt.Printf("%T, %v\n", conn3, err3)
	// Output:
	// *connmanager.PoolSlot, <nil>
	// *connmanager.PoolSlot, <nil>
	// *connmanager.PoolSlot, ERR_UNAVAILABLE: connections limit exceed
}

func ExamplePool_Release() {
	pool, _ := cm.NewPool(
		&DemoConnector{},
		[]db.Addr{"host1:4100", "host2:4100", "host3:4100"},
		cm.WithMaxSize(2),
	)
	credentials := db.Credentials{
		UserName: "someuser",
		Password: "superpass",
	}
	conn, _ := pool.Acquire(context.Background(), credentials)

	err1 := pool.Release(conn)
	err2 := pool.Release(conn)

	fmt.Println(err1)
	fmt.Println(err2)
	// Output:
	// <nil>
	// ERR_UNKNOWN: cannot release connection, not in use
}

func ExamplePool_Drop() {
	pool, _ := cm.NewPool(
		&DemoConnector{},
		[]db.Addr{"host1:4100", "host2:4100", "host3:4100"},
		cm.WithMaxSize(2),
	)
	credentials := db.Credentials{
		UserName: "someuser",
		Password: "superpass",
	}
	conn, _ := pool.Acquire(context.Background(), credentials)

	err1 := pool.Drop(conn)
	err2 := pool.Drop(conn)

	fmt.Println(err1)
	fmt.Println(err2)
	// Output:
	// <nil>
	// ERR_UNKNOWN: cannot release connection, not in use
}

func ExamplePool_Close() {
	pool, _ := cm.NewPool(
		&DemoConnector{},
		[]db.Addr{"host1:4100", "host2:4100", "host3:4100"},
		cm.WithMaxSize(2),
	)
	credentials := db.Credentials{
		UserName: "someuser",
		Password: "superpass",
	}
	conn, _ := pool.Acquire(context.Background(), credentials)
	_ = pool.Release(conn)

	err1 := pool.Close()
	_, err2 := pool.Acquire(context.Background(), credentials)

	fmt.Println(err1)
	fmt.Println(err2)
	// Output:
	// <nil>
	// ERR_UNKNOWN: pool is closed already
}
