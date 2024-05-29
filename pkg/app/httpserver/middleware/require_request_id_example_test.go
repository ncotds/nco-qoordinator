package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
)

func ExampleRequireRequestId_ok() {
	h := middleware.RequireRequestId(http.HandlerFunc(okHandler))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.XRequestIDHeader, "XXX-YYY")

	h.ServeHTTP(rr, req)
	fmt.Println("response:", rr.Code, rr.Body.String())
	// Output:
	// response: 200 XXX-YYY - OK
}

func ExampleRequireRequestId_bad() {
	h := middleware.RequireRequestId(http.HandlerFunc(okHandler))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rr, req)
	fmt.Println("response:", rr.Code, rr.Body.String())
	// Output:
	// response: 400 {"error":"ERR_VALIDATION","message":"X-Request-Id header is required"}
}
