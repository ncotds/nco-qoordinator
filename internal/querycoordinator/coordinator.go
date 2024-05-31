package querycoordinator

import (
	"context"

	"github.com/ncotds/nco-qoordinator/pkg/models"
)

// QueryCoordinator provides methods to run SQL query against a few ObjectServers concurrently and collect all results
type QueryCoordinator struct {
	clients map[string]Client
}

// NewQueryCoordinator creates a ready to use instance of coordinator.
// At least one ObjectServer client must be provided
func NewQueryCoordinator(client Client, clients ...Client) *QueryCoordinator {
	clientsMap := make(map[string]Client, len(clients)+1)
	clientsMap[client.Name()] = client
	for _, cl := range clients {
		clientsMap[cl.Name()] = cl
	}
	return &QueryCoordinator{clients: clientsMap}
}

// Exec runs the given query on the specified NCOS clients and returns the query results for each of them.
//
// If no client names are provided, the query will be executed on all available clients.
// All unknown client names will be ignored
func (q *QueryCoordinator) Exec(
	ctx context.Context,
	query models.Query,
	user models.Credentials,
	clientNames ...string,
) map[string]models.QueryResult {
	if len(clientNames) == 0 {
		clientNames = append(clientNames, q.ClientNames()...)
	}
	actualClients := q.actualClients(clientNames...)

	clientResponses := make(chan clientResponse)
	for _, client := range actualClients {
		go func(client Client) {
			result := client.Exec(ctx, query, user)
			clientResponses <- clientResponse{
				ncosName: client.Name(),
				result:   result,
			}
		}(client)
	}

	result := make(map[string]models.QueryResult, len(actualClients))
	for i := 0; i < len(actualClients); i++ {
		resp := <-clientResponses
		result[resp.ncosName] = resp.result
	}
	close(clientResponses)

	return result
}

// ClientNames returns names of all configured clients
func (q *QueryCoordinator) ClientNames() (names []string) {
	for name := range q.clients {
		names = append(names, name)
	}
	return names
}

// actualClients returns configured clients by it names, unknown names are ignored
func (q *QueryCoordinator) actualClients(names ...string) (clients []Client) {
	for _, name := range names {
		if _, ok := q.clients[name]; ok {
			clients = append(clients, q.clients[name])
		}
	}
	return clients
}

type clientResponse struct {
	ncosName string
	result   models.QueryResult
}
