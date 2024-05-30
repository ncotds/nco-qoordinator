package restapi_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
	"github.com/ncotds/nco-qoordinator/pkg/restapi"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	authBase64 := base64.StdEncoding.EncodeToString([]byte(WordFactory() + ":" + WordFactory()))
	authHeader := "Basic " + authBase64

	type args struct {
		clientName string
		clientRows []qc.QueryResultRow
		request    *http.Request
		headers    http.Header
	}
	tests := []struct {
		name        string
		args        args
		wantStatus  int
		wantErrCode string
		wantBody    func(t *testing.T, args args, rr *httptest.ResponseRecorder)
	}{
		{
			"GET names ok",
			args{
				clientName: WordFactory(),
				clientRows: nil,
				request: httptest.NewRequest(
					http.MethodGet,
					"/clusterNames",
					nil,
				),
				headers: http.Header{"X-Request-Id": {UUIDFactory()}},
			},
			http.StatusOK,
			"",
			func(t *testing.T, args args, rr *httptest.ResponseRecorder) {
				expectedBody := fmt.Sprintf(`["%s"]`, args.clientName)
				assert.Equal(t, expectedBody, rr.Body.String())
			},
		},
		{
			"GET names no request id fail",
			args{
				clientName: WordFactory(),
				clientRows: nil,
				request: httptest.NewRequest(
					http.MethodGet,
					"/clusterNames",
					nil,
				),
				headers: http.Header{},
			},
			http.StatusBadRequest,
			app.ErrCodeValidation,
			nil,
		},
		{
			"POST rawSQL ok",
			args{
				clientName: WordFactory(),
				clientRows: TableRowsFactory(5),
				request: httptest.NewRequest(
					http.MethodPost,
					"/rawSQL",
					bytes.NewBuffer([]byte(`{"sql":"select * from status"}`)),
				),
				headers: http.Header{
					"X-Request-Id":  {UUIDFactory()},
					"Content-Type":  {"application/json"},
					"Authorization": {authHeader},
				},
			},
			http.StatusOK,
			"",
			func(t *testing.T, args args, rr *httptest.ResponseRecorder) {
				var resp []struct {
					ClusterName string
					Rows        []qc.QueryResultRow
				}
				if assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp)) {
					assert.Len(t, resp, 1)
					assert.Equal(t, args.clientName, resp[0].ClusterName)
					assertQueryRows(t, args.clientRows, resp[0].Rows)
				}
			},
		},
		{
			"POST rawSQL no content type fail",
			args{
				clientName: WordFactory(),
				clientRows: TableRowsFactory(5),
				request: httptest.NewRequest(
					http.MethodPost,
					"/rawSQL",
					bytes.NewBuffer([]byte(`{"sql":"select * from status"}`)),
				),
				headers: http.Header{
					"X-Request-Id":  {UUIDFactory()},
					"Authorization": {authHeader},
				},
			},
			http.StatusBadRequest,
			app.ErrCodeValidation,
			nil,
		},
		{
			"POST rawSQL invalid body fail",
			args{
				clientName: WordFactory(),
				clientRows: TableRowsFactory(5),
				request: httptest.NewRequest(
					http.MethodPost,
					"/rawSQL",
					bytes.NewBuffer([]byte(`{"sq":"select * from status"}`)),
				),
				headers: http.Header{
					"X-Request-Id":  {UUIDFactory()},
					"Content-Type":  {"application/json"},
					"Authorization": {authHeader},
				},
			},
			http.StatusBadRequest,
			app.ErrCodeValidation,
			nil,
		},
		{
			"POST rawSQL no auth fail",
			args{
				clientName: WordFactory(),
				clientRows: TableRowsFactory(5),
				request: httptest.NewRequest(
					http.MethodPost,
					"/rawSQL",
					bytes.NewBuffer([]byte(`{"sq":"select * from status"}`)),
				),
				headers: http.Header{
					"X-Request-Id": {UUIDFactory()},
					"Content-Type": {"application/json"},
				},
			},
			http.StatusUnauthorized,
			app.ErrCodeInsufficientPrivileges,
			nil,
		},
		{
			"POST rawSQL empty user-pass fail",
			args{
				clientName: WordFactory(),
				clientRows: TableRowsFactory(5),
				request: httptest.NewRequest(
					http.MethodPost,
					"/rawSQL",
					bytes.NewBuffer([]byte(`{"sq":"select * from status"}`)),
				),
				headers: http.Header{
					"X-Request-Id":  {UUIDFactory()},
					"Content-Type":  {"application/json"},
					"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(":"))},
				},
			},
			http.StatusForbidden,
			app.ErrCodeInsufficientPrivileges,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockClient{DSName: tt.args.clientName, DBTable: tt.args.clientRows}
			svc := qc.NewQueryCoordinator(client)
			srv := restapi.NewServer(svc, restapi.ServerConfig{})

			rr := httptest.NewRecorder()
			tt.args.request.Header = tt.args.headers

			srv.Handler.ServeHTTP(rr, tt.args.request)

			assert.Equalf(t, tt.wantStatus, rr.Code, "status code, resp: %s", rr.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, tt.args, rr)
			}
			if tt.wantErrCode != "" {
				var errResp httpserver.ErrResponse
				if assert.NoErrorf(t, json.Unmarshal(rr.Body.Bytes(), &errResp), "resp: %s", rr.Body.String()) {
					assert.Equal(t, tt.wantErrCode, errResp.Error)
				}
			}
		})
	}
}

func assertQueryRows(t *testing.T, expected, actual []qc.QueryResultRow) {
	assert.Len(t, actual, len(expected), "query rows")
	expectedJSONs := make([][]byte, len(expected))
	for i, elem := range expected {
		expectedJSONs[i], _ = json.Marshal(elem)
	}
	actualJSONs := make([][]byte, len(actual))
	for i, elem := range expected {
		actualJSONs[i], _ = json.Marshal(elem)
	}
	assert.ElementsMatch(t, expectedJSONs, actualJSONs, "query rows")
}
