package app_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"regexp"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

func ExampleNewLogger_stdout() {
	ctx := context.Background()
	ctx = app.WithRequestID(ctx, "ZZZ")
	ctx = app.WithLogAttrs(ctx, app.Attrs{"XXX": "YYY"})

	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelDebug))

	log.DebugContext(ctx, "hello", "key", "value")
	log.InfoContext(ctx, "hello", "key", "value")
	log.WarnContext(ctx, "hello", "key", "value")
	log.ErrorContext(ctx, "hello", "key", "value")

	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"DEBUG","msg":"hello","key":"value","request_id":"ZZZ","XXX":"YYY"}
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"INFO","msg":"hello","key":"value","request_id":"ZZZ","XXX":"YYY"}
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"WARN","msg":"hello","key":"value","request_id":"ZZZ","XXX":"YYY"}
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"ERROR","msg":"hello","key":"value","request_id":"ZZZ","XXX":"YYY"}
}

func ExampleNewLogger_noop() {
	ctx := context.Background()
	log := app.NewLogger(nil, app.WithLogLevel(slog.LevelDebug))

	log.DebugContext(ctx, "hello", "key", "value")
	log.InfoContext(ctx, "hello", "key", "value")
	log.WarnContext(ctx, "hello", "key", "value")
	log.ErrorContext(ctx, "hello", "key", "value")

	// Output:
	//
}

func ExampleLogger_Err() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelDebug))
	reason := errors.New("foo")
	err := app.Err(app.ErrCodeValidation, "baz", reason)

	log.Err(err, "hello", "key", "value")

	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"ERROR","msg":"hello","key":"value","error":"ERR_VALIDATION: baz","reason":"foo"}
}

func ExampleLogger_ErrContext() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelDebug))
	ctx := app.WithRequestID(context.Background(), "ZZZ")
	reason := errors.New("foo")
	err := app.Err(app.ErrCodeValidation, "baz", reason)

	log.ErrContext(ctx, err, "hello", "key", "value")

	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"ERROR","msg":"hello","key":"value","error":"ERR_VALIDATION: baz","reason":"foo","request_id":"ZZZ"}
}

func ExampleLogger_SetLevel() {
	ctx := context.Background()

	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	// debug record will be skipped
	log.DebugContext(ctx, "hello first", "key", "value")

	log.SetLevel(slog.LevelDebug)
	// debug record will now be printed
	log.DebugContext(ctx, "hello second", "key", "value")

	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"DEBUG","msg":"hello second","key":"value"}
}

func ExampleLogger_LogLogger() {
	log := app.NewLogger(stdout{}, app.WithLogLevel(slog.LevelInfo))
	logLogger := log.LogLogger()

	logLogger.Printf("some key: %s", "val")
	// Output:
	// {"time":"2006-01-02T15:05:06.000000000+07:00","level":"INFO","msg":"some key: val"}
}

var (
	hideTimeRe = regexp.MustCompile(`"time":"[TZ0-9:+.-]+"`)
)

type stdout struct {
}

func (s stdout) Write(b []byte) (int, error) {
	b = hideTimeRe.ReplaceAll(b, []byte(`"time":"2006-01-02T15:05:06.000000000+07:00"`))
	return os.Stdout.Write(b)
}
