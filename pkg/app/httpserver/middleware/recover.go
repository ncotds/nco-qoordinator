package middleware

import (
	"net/http"
	"net/textproto"
	"runtime/debug"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

const (
	logLabelRecoverer = "middleware/Recoverer"
)

var ConnectionHeader = textproto.CanonicalMIMEHeaderKey("connection")

// Recoverer catches handler panic, log it and send http.StatusInternalServerError to client
func Recoverer(log *app.Logger) func(next http.Handler) http.Handler {
	if log == nil {
		log = app.NewLogger(nil) // noop
	} else {
		log = log.WithComponent(logLabelRecoverer)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						panic(rvr)
					}

					if r.Header.Get(ConnectionHeader) != "Upgrade" {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}

					log.ErrorContext(
						r.Context(),
						"fatal error",
						"error", rvr,
						"stacktrace", string(debug.Stack()),
					)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
