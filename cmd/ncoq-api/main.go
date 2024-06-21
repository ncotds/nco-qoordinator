package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ncotds/nco-qoordinator/pkg/components"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

var (
	version = "development"

	configPath  = flag.String("c", "config.yml", "path to config file")
	showVersion = flag.Bool("version", false, "print version and exit")
)

func main() {
	os.Exit(run())
}

func run() (rc int) {
	rc = 1

	flag.Parse()
	if *showVersion {
		fmt.Println("version:", version)
		return 0
	}

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println("cannot get config", err.Error())
		return rc
	}

	logger, err := components.NewLoggerComponent(conf)
	if err != nil {
		fmt.Println("cannot setup logger", err.Error())
		return rc
	}

	coord, err := components.InitService(conf.OMNIbus, logger.Logger())
	if err != nil {
		fmt.Println("cannot setup service", err.Error())
		return rc
	}

	restapi, err := components.NewRESTServerComponent(conf.HTTPServer, coord, logger.Logger())
	if err != nil {
		fmt.Println("cannot setup restapi server", err.Error())
		return rc
	}

	gr, fail := errgroup.WithContext(context.Background())
	gr.Go(logger.Run)
	gr.Go(restapi.Run)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case <-fail.Done():
		rc = 1
	case <-interrupt:
		rc = 0
	}

	stopTimeout := 10 * time.Second // TODO add to config

	defer func() { _ = logger.Shutdown(stopTimeout) }() // stop logger the last

	if err := restapi.Shutdown(stopTimeout); err != nil {
		rc = 1
	}

	return rc
}
