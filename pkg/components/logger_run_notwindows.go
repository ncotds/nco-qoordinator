//go:build !windows
// +build !windows

package components

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func (l *LoggerComponent) run() error {
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
