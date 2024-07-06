package components_test

import (
	"os"
	"testing"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/components"
	"github.com/ncotds/nco-qoordinator/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoggerComponent(t *testing.T) {
	conf := &config.Config{LogLevel: "ERROR"}

	logger, err := components.NewLoggerComponent(conf)
	require.NoError(t, err)

	errRun := make(chan error, 1)
	go func() {
		errRun <- logger.Run()
	}()

	errShutdown := logger.Shutdown(time.Second)

	assert.NoError(t, <-errRun, "logger Run()")
	assert.NoError(t, errShutdown, "logger Shutdown()")
}

func TestNewLoggerComponent_toFile(t *testing.T) {
	logFile, _ := os.CreateTemp(os.TempDir(), "ncoq-api-*.log")
	defer os.Remove(logFile.Name())

	conf := &config.Config{
		LogLevel: "ERROR",
		LogFile:  logFile.Name(),
	}

	logger, err := components.NewLoggerComponent(conf)
	require.NoError(t, err)

	errRun := make(chan error, 1)
	go func() {
		errRun <- logger.Run()
	}()

	errShutdown := logger.Shutdown(time.Second)

	assert.NoError(t, <-errRun, "logger Run()")
	assert.NoError(t, errShutdown, "logger Shutdown()")
}

func TestNewLoggerComponent_badLevel(t *testing.T) {
	conf := &config.Config{LogLevel: "QQQ"}

	_, err := components.NewLoggerComponent(conf)
	assert.Error(t, err)
}
