package querycoordinator_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	db "github.com/ncotds/nco-lib/dbconnector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	. "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	mocks "github.com/ncotds/nco-qoordinator/internal/querycoordinator/mocks"
)

func TestQueryCoordinator_Exec(t *testing.T) {
	suite.Run(t, &QueryCoordinatorTestSuite{ClientsCount: 5})
}

func TestQueryCoordinator_Exec_SingleDataSource(t *testing.T) {
	suite.Run(t, &QueryCoordinatorTestSuite{ClientsCount: 1})
}

type QueryCoordinatorTestSuite struct {
	suite.Suite
	ClientsCount int

	fixtureQuery            db.Query
	fixtureUser             db.Credentials
	fixtureError            error
	fixtureExecReturnResult map[string]db.QueryResult
}

func (s *QueryCoordinatorTestSuite) SetupSuite() {
	s.Greater(s.ClientsCount, 0, "at least one client is required")

	s.fixtureQuery = QueryFactory()
	s.fixtureUser = CredentialsFactory()
	s.fixtureError = ErrorFactory()

	s.fixtureExecReturnResult = make(map[string]db.QueryResult, s.ClientsCount)
	for i := 0; i < s.ClientsCount; i++ {
		s.fixtureExecReturnResult[UUIDStringFactory()] = db.QueryResult{
			RowSet:       QueryResultRowSetFactory(10, 10),
			AffectedRows: IntFactory(),
		}
	}
}

func (s *QueryCoordinatorTestSuite) TestExecOnAllDataSources() {
	clients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name, resp := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).Return(resp)
		clients = append(clients, client)
	}
	coordinator := NewQueryCoordinator(clients[0], clients[1:]...)

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser)

	s.Equal(s.fixtureExecReturnResult, result)
}

func (s *QueryCoordinatorTestSuite) TestExecOnParticularDataSource() {
	clients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name, resp := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).Maybe().Return(resp)
		clients = append(clients, client)
	}
	coordinator := NewQueryCoordinator(clients[0], clients[1:]...)
	dsName := clients[rand.Intn(len(clients))].Name()

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser, dsName)

	s.Equal(map[string]db.QueryResult{dsName: s.fixtureExecReturnResult[dsName]}, result)
}

func (s *QueryCoordinatorTestSuite) TestExecOnUnknownDataSource() {
	clients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name, resp := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).Maybe().Return(resp)
		clients = append(clients, client)
	}
	coordinator := NewQueryCoordinator(clients[0], clients[1:]...)
	dsName := UUIDStringFactory()

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser, dsName)

	s.Equal(map[string]db.QueryResult{}, result) // unknown DS is skipped
}

func (s *QueryCoordinatorTestSuite) TestExecDataSourceFails() {
	clients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name, resp := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).Maybe().Return(resp)
		clients = append(clients, client)
	}
	failedClient := mocks.NewMockClient(s.T())
	failedClient.EXPECT().Name().Return(UUIDStringFactory())
	failedClient.EXPECT().
		Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).
		Return(db.QueryResult{Error: s.fixtureError})
	coordinator := NewQueryCoordinator(failedClient, clients...)

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser)

	s.Len(result, len(clients)+1) // all clients returned their results
	s.Equal(s.fixtureError, result[failedClient.Name()].Error)
}

func (s *QueryCoordinatorTestSuite) TestExecAllDataSourcesFails() {
	clients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().
			Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).
			Maybe().
			Return(db.QueryResult{Error: s.fixtureError})
		clients = append(clients, client)
	}
	coordinator := NewQueryCoordinator(clients[0], clients[1:]...)

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser)

	s.Equal(len(clients), len(result))
	for _, client := range clients {
		s.Equal(s.fixtureError, result[client.Name()].Error)
	}
}

func (s *QueryCoordinatorTestSuite) TestExecCancel() {
	fastClients := make([]Client, 0, len(s.fixtureExecReturnResult))
	for name, resp := range s.fixtureExecReturnResult {
		client := mocks.NewMockClient(s.T())
		client.EXPECT().Name().Return(name)
		client.EXPECT().Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).Return(resp)
		fastClients = append(fastClients, client)
	}
	slowClient := mocks.NewMockClient(s.T())
	slowClient.EXPECT().Name().Return(UUIDStringFactory())
	slowClient.On("Exec", mock.Anything, s.fixtureQuery, s.fixtureUser).Return(
		func(ctx context.Context, query db.Query, user db.Credentials) db.QueryResult {
			time.Sleep(1 * time.Second)
			return db.QueryResult{Error: context.DeadlineExceeded}
		},
	)

	coordinator := NewQueryCoordinator(slowClient, fastClients...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := coordinator.Exec(ctx, s.fixtureQuery, s.fixtureUser)

	s.Len(result, len(fastClients)+1) // all clients returned their results
	s.Equal(context.DeadlineExceeded, result[slowClient.Name()].Error)
}

func TestQueryCoordinator_ClientNames(t *testing.T) {
	tests := []struct {
		name         string
		clientsCount int
	}{
		{
			name:         "single client",
			clientsCount: 1,
		},
		{
			name:         "a few clients",
			clientsCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, clients := MockClientsFactory(t, tt.clientsCount)

			q := NewQueryCoordinator(clients[0], clients[1:]...)

			gotNames := q.ClientNames()

			assert.ElementsMatch(t, names, gotNames)
		})
	}
}
