package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

// ErrResponse represents standard HTTP error response body
type ErrResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

// ErrJSON encodes error to JSON and sends to client
func ErrJSON(w http.ResponseWriter, _ *http.Request, code int, err error) {
	var resp ErrResponse

	var appErr app.Error
	if errors.As(err, &appErr) {
		resp.Error = appErr.Code()
		resp.Message = appErr.Message()
		if reason := appErr.Unwrap(); reason != nil {
			resp.Reason = reason.Error()
		}
	}

	if resp.Error == "" {
		resp.Error = app.ErrCodeUnknown
	}
	if resp.Message == "" {
		resp.Message = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}
