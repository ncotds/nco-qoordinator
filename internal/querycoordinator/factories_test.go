package querycoordinator_test

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	db "github.com/ncotds/nco-lib/dbconnector"

	. "github.com/ncotds/nco-qoordinator/internal/querycoordinator"
	mocks "github.com/ncotds/nco-qoordinator/internal/querycoordinator/mocks"
)

var (
	FakerRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	FakerChoice = []func() any{
		func() any { return FakerRandom.Int() },
		func() any { return FakerRandom.Float64() },
		func() any { return faker.UUIDHyphenated() },
		func() any { return faker.Word() },
		func() any { return faker.Sentence() },
		func() any { return faker.Paragraph() },
		func() any { return faker.URL() },
		func() any { return faker.IPv4() },
		func() any { return faker.RandomUnixTime() },
	}
)

func IntFactory() int {
	return FakerRandom.Int()
}

func UUIDStringFactory() string {
	return faker.UUIDHyphenated()
}

func ErrorFactory() error {
	return errors.New(faker.Sentence())
}

func QueryFactory() db.Query {
	return db.Query{SQL: faker.Sentence()}
}

func CredentialsFactory() db.Credentials {
	return db.Credentials{UserName: faker.Username(), Password: faker.Password()}
}

func QueryResultRowSetFactory(rowCount, colCount uint) db.RowSet {
	schema := make(map[uint]func() any, colCount)
	for i := uint(0); i < colCount; i++ {
		randomIdx := FakerRandom.Intn(len(FakerChoice))
		schema[i] = FakerChoice[randomIdx]
	}

	cols := make([]string, 0, len(schema))
	for range schema {
		cols = append(cols, faker.Word())
	}

	rows := make([][]any, 0, rowCount)
	for i := uint(0); i < rowCount; i++ {
		fakeRow := make([]any, 0, len(cols))
		for j := 0; j < len(cols); j++ {
			fakeRow = append(fakeRow, schema[i]())
		}
		rows = append(rows, fakeRow)
	}

	return db.RowSet{Columns: cols, Rows: rows}
}

func MockClientsFactory(t *testing.T, count int) (names []string, clients []Client) {
	for i := 0; i < count; i++ {
		name := UUIDStringFactory()
		client := mocks.NewMockClient(t)
		client.EXPECT().Name().Return(name)
		clients = append(clients, client)
		names = append(names, name)
	}
	return names, clients
}
