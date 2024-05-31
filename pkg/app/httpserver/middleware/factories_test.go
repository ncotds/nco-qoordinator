package middleware_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

var (
	hideTimeRe       = regexp.MustCompile(`"time":"[TZ0-9:+.-]+"`)
	hideRespTimeRe   = regexp.MustCompile(`"resp_time":"[^"]+"`)
	hideStackTraceRe = regexp.MustCompile(`"stacktrace":"[^"]+"`)
)

type stdout struct {
}

func (s stdout) Write(b []byte) (int, error) {
	b = hideTimeRe.ReplaceAll(b, []byte(`"time":"2006-01-02T15:05:06.000000000+07:00"`))
	b = hideRespTimeRe.ReplaceAll(b, []byte(`"resp_time":"1ms"`))
	b = hideStackTraceRe.ReplaceAll(b, []byte(`"stacktrace":"..."`))
	return os.Stdout.Write(b)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	reqID := app.RequestID(r.Context())
	msg := fmt.Sprintf("%s - OK", reqID)
	_, _ = w.Write([]byte(msg))
}

func badHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	reqID := app.RequestID(r.Context())
	msg := fmt.Sprintf("%s - FAIL", reqID)
	_, _ = w.Write([]byte(msg))
}

func panicHandler(_ http.ResponseWriter, _ *http.Request) {
	panic("FATAL!!!")
}
