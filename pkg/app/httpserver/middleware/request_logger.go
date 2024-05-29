package middleware

import (
	"net/http"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

const (
	logLabelRequestLogger = "middleware/RequestLogger"
)

// RequestLogger logs requests using provided logger
//
// Checks response code:
//   - if 0 < code < http.StatusBadRequest logs INFO
//   - otherwise logs ERROR
func RequestLogger(log *app.Logger) func(next http.Handler) http.Handler {
	if log == nil {
		log = app.NewLogger(nil) // noop
	} else {
		log = log.WithComponent(logLabelRequestLogger)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := &respWriterWrapper{ResponseWriter: w}
			start := time.Now()
			defer func() {
				args := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"resp_code", ww.statusCode,
					"bytes", ww.bytesWritten,
					"resp_time", time.Since(start).String(),
				}
				if ww.statusCode > 0 && ww.statusCode < http.StatusBadRequest {
					log.InfoContext(r.Context(), "request completed", args...)
				} else {
					log.ErrorContext(r.Context(), "request failed", args...)
				}

			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

type respWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rr *respWriterWrapper) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *respWriterWrapper) Write(bytes []byte) (int, error) {
	rr.bytesWritten += len(bytes) // TODO handle http/2
	return rr.ResponseWriter.Write(bytes)
}
