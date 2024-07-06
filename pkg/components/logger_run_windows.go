//go:build windows
// +build windows

package components

// Run do nothing on windows...
// because of missing support for signals (syscall.SIGUSR1)
func (l *LoggerComponent) run() error {
	<-l.interrupt
	return nil
}
