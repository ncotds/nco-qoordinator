package restapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	qc "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	"github.com/ncotds/nco-qoordinator/internal/restapi"
)

/*
	goos: linux
	goarch: amd64
	pkg: github.com/ncotds/nco-qoordinator/internal/restapi
	cpu: 12th Gen Intel(R) Core(TM) i7-1260P
	BenchmarkNewServer_rawSQLPost_select_1-8        100    10594548 ns/op       11577 B/op       101 allocs/op
	BenchmarkNewServer_rawSQLPost_select_100-8      100    11347381 ns/op      724276 B/op      3141 allocs/op
	BenchmarkNewServer_rawSQLPost_select_10000-8     15    77095958 ns/op    85613013 B/op    305460 allocs/op
	PASS
*/

func BenchmarkNewServer_rawSQLPost_select_1(b *testing.B) {
	benchmarkNewServerRawSQLPostSelect(b, 1)
}

func BenchmarkNewServer_rawSQLPost_select_100(b *testing.B) {
	benchmarkNewServerRawSQLPostSelect(b, 100)
}

func BenchmarkNewServer_rawSQLPost_select_10000(b *testing.B) {
	benchmarkNewServerRawSQLPostSelect(b, 10_000)
}

func benchmarkNewServerRawSQLPostSelect(b *testing.B, rowsCount int) {
	client1 := &MockClient{DSName: WordFactory(), DBTable: TableRowsFactory(rowsCount)}
	client2 := &MockClient{DSName: WordFactory(), DBTable: TableRowsFactory(rowsCount)}
	svc := qc.NewQueryCoordinator(client1, client2)
	srv := restapi.NewServer(svc, restapi.ServerConfig{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rr := httptest.NewRecorder()
		req := testSelectRequest()
		b.StartTimer()

		srv.Handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatal("bad response code", rr.Code, rr.Body.String())
		}
	}

}

func testSelectRequest() *http.Request {
	req := httptest.NewRequest(
		http.MethodPost,
		"/rawSQL",
		bytes.NewBuffer([]byte(`{"sql":"select * from status"}`)),
	)
	req.Header.Add("X-Request-Id", UUIDFactory())
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(WordFactory(), WordFactory())
	return req
}
