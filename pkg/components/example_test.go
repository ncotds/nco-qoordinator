package components_test

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ncotds/nco-qoordinator/pkg/components"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

var conf = &config.Config{
	LogLevel: "ERROR",
	HTTPServer: config.HTTPServerConfig{
		Listen: ":8090",
	},
	OMNIbus: config.OMNIbus{
		Clusters: map[string]config.SeedList{"OMNI1": {"localhost:4100"}},
	},
}

func Example() {
	logger, err := components.NewLoggerComponent(conf)
	if err != nil {
		fmt.Println("cannot setup logger", err.Error())
	}

	coord, err := components.InitService(conf.OMNIbus, logger.Logger())
	if err != nil {
		fmt.Println("cannot setup service", err.Error())
	}

	restapi, err := components.NewRESTServerComponent(conf.HTTPServer, coord, logger.Logger())
	if err != nil {
		fmt.Println("cannot setup restapi server", err.Error())
	}

	fmt.Printf("%T\n", restapi)

	gr, fail := errgroup.WithContext(context.Background())
	gr.Go(logger.Run)
	gr.Go(restapi.Run)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // emulate caller interrupt the server

	select {
	case <-fail.Done():
		fmt.Println("error")
	case <-ctx.Done():
		fmt.Println("interrupt")
	}

	fmt.Println("restapi:", restapi.Shutdown(time.Second))
	fmt.Println("logger:", logger.Shutdown(time.Second))
	// Output:
	// *components.RESTServerComponent
	// interrupt
	// restapi: <nil>
	// logger: <nil>
}
