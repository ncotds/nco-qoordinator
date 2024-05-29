package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
)

func ExampleRecoverer() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	recoverer := middleware.Recoverer(log)

	h := recoverer(http.HandlerFunc(panicHandler))
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rr, req)
	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"ERROR","msg":"fatal error","component":"middleware/Recoverer","error":"FATAL!!!","stacktrace":"..."}
}
