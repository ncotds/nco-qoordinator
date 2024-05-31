package httpserver_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver"
)

func ExampleErrJSON_appError() {
	rw := httptest.NewRecorder()
	err := app.Err(app.ErrCodeValidation, "foo")

	httpserver.ErrJSON(rw, &http.Request{}, http.StatusBadRequest, err)

	fmt.Println("status:", rw.Code)
	fmt.Println("content-type:", rw.Header().Get("Content-Type"))
	fmt.Println("body:", rw.Body.String())
	// Output:
	// status: 400
	// content-type: application/json
	// body: {"error":"ERR_VALIDATION","message":"foo"}
}

func ExampleErrJSON_anyError() {
	rw := httptest.NewRecorder()
	err := errors.New("foo")

	httpserver.ErrJSON(rw, &http.Request{}, http.StatusInternalServerError, err)

	fmt.Println("status:", rw.Code)
	fmt.Println("content-type:", rw.Header().Get("Content-Type"))
	fmt.Println("body:", rw.Body.String())
	// Output:
	// status: 500
	// content-type: application/json
	// body: {"error":"ERR_UNKNOWN","message":"foo"}
}
