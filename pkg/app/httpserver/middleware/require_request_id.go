package middleware

import (
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver"
)

var (
	XRequestIDHeader = textproto.CanonicalMIMEHeaderKey("x-request-id")
)

// RequireRequestId checks if XRequestIDHeader exists.
//
// If ok, adds request id into context.
// If not, sends error response with ErrCodeValidation
func RequireRequestId(next http.Handler) http.Handler {
	msg := fmt.Sprintf("%s header is required", XRequestIDHeader)

	fn := func(w http.ResponseWriter, r *http.Request) {
		xRequestID := r.Header.Get(XRequestIDHeader)
		if xRequestID == "" {
			err := app.Err(app.ErrCodeValidation, msg)
			httpserver.ErrJSON(w, r, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		ctx = app.WithRequestID(ctx, xRequestID)

		w.Header().Set(XRequestIDHeader, xRequestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
