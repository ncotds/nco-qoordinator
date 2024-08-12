package restapi_test

import (
	"context"
	"time"

	"github.com/go-faker/faker/v4"
	db "github.com/ncotds/nco-lib/dbconnector"
)

// MockClient implements Client interface for tests only, use your oun implementation
type MockClient struct {
	DSName  string
	DBTable db.RowSet
}

func (c *MockClient) Name() string {
	return c.DSName
}

func (c *MockClient) Exec(ctx context.Context, _ db.Query, _ db.Credentials) db.QueryResult {
	select {
	case <-time.After(10 * time.Millisecond):
		// working hard
	case <-ctx.Done():
		return db.QueryResult{Error: ctx.Err()}
	}
	return db.QueryResult{RowSet: c.DBTable, AffectedRows: 0, Error: nil}
}

func WordFactory() string {
	return faker.Word()
}

func UUIDFactory() string {
	return faker.UUIDHyphenated()
}

func TableRowsFactory(count int) db.RowSet {
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
	return db.RowSet{Columns: cols, Rows: rows}
}

func RowSetToMap(in db.RowSet) []map[string]any {
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
