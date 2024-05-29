package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
)

func ExampleRequestLogger_ok() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	rl := middleware.RequestLogger(log)

	h := rl(http.HandlerFunc(okHandler))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rr, req)
	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"INFO","msg":"request completed","component":"middleware/RequestLogger","method":"GET","path":"/","remote_addr":"","user_agent":"","resp_code":200,"bytes":5,"resp_time":"1ms"}
}

func ExampleRequestLogger_bad() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	rl := middleware.RequestLogger(log)

	h := rl(http.HandlerFunc(badHandler))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rr, req)
	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"ERROR","msg":"request failed","component":"middleware/RequestLogger","method":"GET","path":"/","remote_addr":"","user_agent":"","resp_code":400,"bytes":7,"resp_time":"1ms"}
}
