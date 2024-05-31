package components

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

type LoggerComponent struct {
	log          *app.Logger
	lvl          slog.Level
	targetCloser io.Closer
	interrupt    chan struct{}
}

// NewLoggerComponent creates new logger component based on config.
//
// If Config.LogFile is not empty - logs will be written into the file,
// otherwise - to stdout.
func NewLoggerComponent(conf *config.Config) (*LoggerComponent, error) {
	target := os.Stdout

	var lvl slog.Level
	err := lvl.UnmarshalText([]byte(conf.LogLevel))
	if err != nil {
		return nil, err
	}

	if conf.LogFile != "" {
		target, err = os.OpenFile(conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		fmt.Println("log will written into file:", conf.LogFile)
	}

	log := app.NewLogger(target, app.WithLogLevel(lvl))

	return &LoggerComponent{
		log:          log,
		lvl:          lvl,
		targetCloser: target,
		interrupt:    make(chan struct{}),
	}, nil
}

// Run enables DEBUG switching on syscall.SIGUSR1.
// The second sending of SIGUSR1 will switch back to initial log level
func (l *LoggerComponent) Run() error {
	toggle := make(chan os.Signal, 1)
	signal.Notify(toggle, syscall.SIGUSR1)

	currentLevel := l.lvl
	for {
		select {
		case <-l.interrupt:
			return nil
		case <-toggle:
			if currentLevel == l.lvl {
				l.log.SetLevel(slog.LevelDebug)
				currentLevel = slog.LevelDebug
			} else {
				l.log.SetLevel(l.lvl)
				currentLevel = l.lvl
			}
		}
	}
}

// Shutdown stops listening of SIGUSR1 and closes the target file descriptor
func (l *LoggerComponent) Shutdown(timeout time.Duration) error {
	// need to close log file
	defer func() {
		_ = l.targetCloser.Close()
	}()

	select {
	// try to stop DEBUG toggle
	case l.interrupt <- struct{}{}:
	case <-time.After(timeout):
	}

	return nil
}

// Logger returns wrapped app.Logger
func (l *LoggerComponent) Logger() *app.Logger {
	return l.log
}
