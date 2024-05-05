package querycoordinator_test

import (
	"errors"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	. "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
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

func WordFactory() string {
	return faker.Word()
}

func ErrorFactory() error {
	return errors.New(faker.Sentence())
}

func QueryFactory() Query {
	return Query{SQL: faker.Sentence()}
}

func CredentialsFactory() Credentials {
	return Credentials{UserName: faker.Username(), Password: faker.Password()}
}

func QueryResultRowSetFactory(rowCount, colCount uint) []QueryResultRow {
	schema := make(map[string]func() any, colCount)
	for i := uint(0); i < colCount; i++ {
		randomIdx := FakerRandom.Intn(len(FakerChoice))
		schema[faker.Word()] = FakerChoice[randomIdx]
	}

	fakeRow := make([]QueryResultRow, 0, rowCount)
	for i := uint(0); i < rowCount; i++ {
		fakeRow = append(fakeRow, QueryResultRowFactory(schema))
	}
	return fakeRow
}

func QueryResultRowFactory(schema map[string]func() any) QueryResultRow {
	if len(schema) == 0 {
		schema[faker.Word()] = func() any { return faker.Word() }
	}

	fakeRow := make(map[string]any, len(schema))
	for name, val := range schema {
		fakeRow[name] = val()
	}

	return fakeRow
}
