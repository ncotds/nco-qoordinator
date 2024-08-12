//go:generate ogen --target gen --package gen --clean ../../docs/openapi/openapi.yml
package restapi

import (
	"net/http"
	"time"

	qc "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	"github.com/ncotds/nco-qoordinator/internal/restapi/gen"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver/middleware"
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
