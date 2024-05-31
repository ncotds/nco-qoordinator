package restapi_test

import (
	"context"
	"time"

	"github.com/go-faker/faker/v4"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

// MockClient implements Client interface for tests only, use your oun implementation
type MockClient struct {
	DSName  string
	DBTable []qc.QueryResultRow
}

func (c *MockClient) Name() string {
	return c.DSName
}

func (c *MockClient) Exec(ctx context.Context, query qc.Query, user qc.Credentials) qc.QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return qc.QueryResult{Error: ctx.Err()}
	}
	return qc.QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
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

func TableRowsFactory(count int) []qc.QueryResultRow {
	rows := make([]qc.QueryResultRow, 0, count)
	for i := 0; i < count; i++ {
		rows = append(rows, qc.QueryResultRow{
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
