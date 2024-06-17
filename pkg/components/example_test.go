package components_test

import (
	"fmt"

	"github.com/ncotds/nco-qoordinator/pkg/components"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

var conf = &config.Config{
	LogLevel: "ERROR",
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

	fmt.Printf("%T", restapi)
	// Output:
	// *components.RESTServerComponent
}
