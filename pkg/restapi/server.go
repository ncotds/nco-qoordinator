//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target gen --package gen --clean ../../docs/openapi/openapi.yml
package restapi

import (
	"net/http"
	"time"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
	"github.com/ncotds/nco-qoordinator/pkg/restapi/gen"
)

type ServerConfig struct {
	Listen               string
	Timeout, IdleTimeout time.Duration
	Log                  *app.Logger
}

func NewServer(coordinator *qc.QueryCoordinator, config ServerConfig) *http.Server {
	service := CoordinatorService{Coordinator: coordinator}
	h, _ := gen.NewServer(
		service,
		SecurityHandler{},
		gen.WithErrorHandler(errorHandler),
	)
	s := &http.Server{
		Addr:         config.Listen,
		Handler:      middleware.DefaultChain(h, config.Log),
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		IdleTimeout:  config.IdleTimeout,
	}
	if config.Log != nil {
		s.ErrorLog = config.Log.LogLogger()
	}
	return s
}
