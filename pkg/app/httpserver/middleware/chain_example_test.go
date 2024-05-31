package middleware_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
)

func ExampleDefaultChain() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	h := middleware.DefaultChain(http.HandlerFunc(okHandler), log)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.XRequestIDHeader, "XXX-YYY")

	h.ServeHTTP(rr, req)

	fmt.Println("response:", rr.Body.String())
	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"INFO","msg":"request completed","component":"middleware/RequestLogger","method":"GET","path":"/","remote_addr":"","user_agent":"","resp_code":200,"bytes":12,"resp_time":"1ms","request_id":"XXX-YYY"}
	// response: XXX-YYY - OK
}
