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
	DBTable models.RowSet
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

func UUIDFactory() string {
	return faker.UUIDHyphenated()
}

func TableRowsFactory(count int) models.RowSet {
	cols := []string{
		"Identifier",
		"Node",
		"NodeAlias",
		"Agent",
		"Manager",
		"AlertGroup",
		"AlertKey",
		"Type",
		"Severity",
		"Summary",
		"FirstOccurrence",
		"URL",
		"ExtendedAttr",
	}

	rows := make([][]any, 0, count)
	for i := 0; i < count; i++ {
		rows = append(rows, []any{
			faker.UUIDHyphenated(),
			faker.DomainName(),
			faker.IPv4(),
			faker.Word(),
			faker.Word(),
			faker.Sentence(),
			faker.UUIDDigit(),
			1,
			5,
			faker.Sentence(),
			faker.UnixTime(),
			faker.URL(),
			faker.Sentence(),
		})
	}
	return models.RowSet{Columns: cols, Rows: rows}
}

func RowSetToMap(in models.RowSet) []map[string]any {
	result := make([]map[string]any, len(in.Rows))
	for _, row := range in.Rows {
		rowLen := min(len(in.Columns), len(row))
		rowMap := make(map[string]any, rowLen)
		for i := 0; i < rowLen; i++ {
			rowMap[in.Columns[i]] = row[i]
		}
	}
	return result
}
