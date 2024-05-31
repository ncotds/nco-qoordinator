package components

import (
	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	nc "github.com/ncotds/nco-qoordinator/internal/ncoclient"
	qc "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	tds "github.com/ncotds/nco-qoordinator/internal/tdsclient"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/config"
)

type QueryCoordinatorService struct {
	coordinator *qc.QueryCoordinator
}

func InitService(conf config.OMNIbus, log *app.Logger) (*QueryCoordinatorService, error) {
	var clients []qc.Client

	for name, addresses := range conf.Clusters {
		seedList := make([]db.Addr, 0, len(addresses))
		for _, addr := range addresses {
			seedList = append(seedList, db.Addr(addr))
		}
		client, err := nc.NewNcoClient(
			name, nc.ClientConfig{
				Connector:         &tds.TDSConnector{AppLabel: conf.ConnectionLabel},
				SeedList:          seedList,
				MaxPoolSize:       conf.MaxConnections,
				UseRandomFailOver: conf.RandomFailOver,
				UseFailBack:       conf.FailBack,
				FailBackDelay:     conf.FailBackDelay,
				Logger:            log,
			},
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	coordinator := qc.NewQueryCoordinator(clients[0], clients[1:]...)
	service := QueryCoordinatorService{coordinator: coordinator}
	return &service, nil
}
