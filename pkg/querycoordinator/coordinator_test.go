package querycoordinator_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	. "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
	mocks "github.com/ncotds/nco-qoordinator/pkg/querycoordinator/mocks"
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

	fixtureQuery            Query
	fixtureUser             Credentials
	fixtureError            error
	fixtureExecReturnResult map[string]QueryResult
}

func (s *QueryCoordinatorTestSuite) SetupSuite() {
	s.Greater(s.ClientsCount, 0, "at least one client is required")

	s.fixtureQuery = QueryFactory()
	s.fixtureUser = CredentialsFactory()
	s.fixtureError = ErrorFactory()

	s.fixtureExecReturnResult = make(map[string]QueryResult, s.ClientsCount)
	for i := 0; i < s.ClientsCount; i++ {
		s.fixtureExecReturnResult[fmt.Sprintf("%s_%d", WordFactory(), i)] = QueryResult{
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

	s.Equal(map[string]QueryResult{dsName: s.fixtureExecReturnResult[dsName]}, result)
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
	dsName := WordFactory()

	result := coordinator.Exec(context.Background(), s.fixtureQuery, s.fixtureUser, dsName)

	s.Equal(map[string]QueryResult{}, result) // unknown DS is skipped
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
	failedClient.EXPECT().Name().Return(WordFactory())
	failedClient.EXPECT().
		Exec(mock.Anything, s.fixtureQuery, s.fixtureUser).
		Return(QueryResult{Error: s.fixtureError})
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
			Return(QueryResult{Error: s.fixtureError})
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
	slowClient.EXPECT().Name().Return(WordFactory())
	slowClient.On("Exec", mock.Anything, s.fixtureQuery, s.fixtureUser).Return(
		func(ctx context.Context, query Query, user Credentials) QueryResult {
			time.Sleep(1 * time.Second)
			return QueryResult{Error: context.DeadlineExceeded}
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
	mocksCount := 5
	names := make([]string, 0, mocksCount)
	clients := make([]Client, 0, mocksCount)
	for i := 0; i < mocksCount; i++ {
		name := WordFactory()
		client := mocks.NewMockClient(t)
		client.EXPECT().Name().Return(name)
		clients = append(clients, client)
		names = append(names, name)
	}

	tests := []struct {
		name      string
		clients   []Client
		wantNames []string
	}{
		{
			name:      "single client",
			clients:   clients[:1],
			wantNames: names[:1],
		},
		{
			name:      "a few clients",
			clients:   clients,
			wantNames: names,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueryCoordinator(tt.clients[0], tt.clients[1:]...)

			gotNames := q.ClientNames()

			assert.ElementsMatch(t, tt.wantNames, gotNames)
		})
	}
}
