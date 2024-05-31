package restapi_test

import (
	"context"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

// MockClient implements Client interface for tests only, use your oun implementation
type MockClient struct {
	DSName  string
	DBTable []models.QueryResultRow
}

func (c *MockClient) Name() string {
	return c.DSName
}

func (c *MockClient) Exec(ctx context.Context, _ models.Query, _ models.Credentials) models.QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return models.QueryResult{Error: ctx.Err()}
	}
	return models.QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
}

func WordFactory() string {
	return faker.Word()
}

func SentenceFactory() string {
	return faker.Sentence()
}

func UUIDFactory() string {
	return faker.UUIDHyphenated()
}

func TableRowsFactory(count int) []models.QueryResultRow {
	rows := make([]models.QueryResultRow, 0, count)
	for i := 0; i < count; i++ {
		rows = append(rows, models.QueryResultRow{
			"Identifier":      faker.UUIDHyphenated(),
			"Node":            faker.DomainName(),
			"NodeAlias":       faker.IPv4(),
			"Agent":           faker.Word(),
			"Manager":         faker.Word(),
			"AlertGroup":      faker.Sentence(),
			"AlertKey":        faker.UUIDDigit(),
			"Type":            1,
			"Severity":        5,
			"Summary":         faker.Sentence(),
			"FirstOccurrence": faker.UnixTime(),
			"URL":             faker.URL(),
			"ExtendedAttr":    faker.Sentence(),
		})
	}
	return rows
}
