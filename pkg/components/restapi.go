package components

import (
	"context"
	"errors"
	"net/http"
	"time"

	rest "github.com/ncotds/nco-qoordinator/internal/restapi"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

type RESTServerComponent struct {
	srv *http.Server
	log *app.Logger
}

// NewRESTServerComponent creates http server that handle service methods
func NewRESTServerComponent(
	conf config.HTTPServerConfig,
	service *QueryCoordinatorService,
	log *app.Logger,
) (*RESTServerComponent, error) {
	srv := rest.NewServer(
		service.coordinator,
		rest.ServerConfig{
			Listen:      conf.Listen,
			Timeout:     conf.Timeout,
			IdleTimeout: conf.IdleTimeout,
			Log:         log,
		},
	)

	return &RESTServerComponent{
		log: log.WithComponent("RESTServerComponent"),
		srv: srv,
	}, nil
}

// Run starts listening http host:port
func (s *RESTServerComponent) Run() error {
	err := s.srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.log.Err(err, "failed to start server")
	}
	return err
}

// Shutdown stops http server
func (s *RESTServerComponent) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.log.Err(err, "failed to stop server")
		return err
	}
	s.log.Info("server stopped")
	return nil
}
