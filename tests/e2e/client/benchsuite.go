package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

func DoBenchmarkInsert(b *testing.B, client Client, rowCount int, concurrency int) {
	var err error
	inserts, _, del := makeQueries(rowCount)
	creds := models.Credentials{UserName: TestUser, Password: TestPassword}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if concurrency > 1 {
			err = processQueriesConcurrently(client, inserts, creds, concurrency)
		} else {
			err = processQueries(client, inserts, creds)
		}
		if err != nil {
			b.Fatal("inserts fails", err.Error())
		}
	}

	b.StopTimer()
	_, err = client.RawSQLPost(context.Background(), models.Query{SQL: del}, creds)
	if err != nil {
		b.Fatal("delete fails", err.Error())
	}
}

func DoBenchmarkSelect(b *testing.B, client Client, rowCount, selCount int, concurrency int) {
	var err error
	inserts, sel, del := makeQueries(rowCount)
	creds := models.Credentials{UserName: TestUser, Password: TestPassword}

	if err := processQueries(client, inserts, creds); err != nil {
		b.Fatal("cannot insert test rows")
	}

	selects := make([]string, selCount)
	for i := range selects {
		selects[i] = sel
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if concurrency > 1 {
			err = processQueriesConcurrently(client, selects, creds, concurrency)
		} else {
			err = processQueries(client, selects, creds)
		}
		if err != nil {
			b.Fatal("inserts fails", err.Error())
		}
	}

	b.StopTimer()
	_, err = client.RawSQLPost(context.Background(), models.Query{SQL: del}, creds)
	if err != nil {
		b.Fatal("delete fails", err.Error())
	}
}

func processQueries(client Client, queries []string, creds models.Credentials) error {
	var err error
	for i := 0; i < len(queries); i++ {
		_, err = client.RawSQLPost(context.Background(), models.Query{SQL: queries[i]}, creds)
		if err != nil {
			return fmt.Errorf("query %d fails: %w", i, err)
		}
	}
	return nil
}

func processQueriesConcurrently(client Client, queries []string, creds models.Credentials, concurrency int) error {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make(chan error, len(queries))

	for gIdx := 0; gIdx < concurrency; gIdx++ {
		go func(offset int) {
			for i := offset; i < len(queries); i += concurrency {
				_, err := client.RawSQLPost(context.Background(), models.Query{SQL: queries[i]}, creds)
				if err != nil {
					results <- fmt.Errorf("query %d fails: %w", i, err)
					return
				}
			}
			results <- nil
		}(gIdx)
	}

	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			return err
		}
	}
	return nil
}

// makeQueries returns rowCount inserts, one select and one delete stmt
func makeQueries(rowCount int) ([]string, string, string) {
	alertGroup := uuid.NewString()
	queries := make([]string, 0, rowCount) // rowCount inserts + select + delete

	for i := 0; i < rowCount; i++ {
		row := AlertStatusRecordFactory()
		row.AlertGroup = alertGroup
		insQ, _, _ := NcoSql.Insert(TestAlertsTable).Rows(row).ToSQL()
		queries = append(queries, insQ)
	}

	selQ, _, _ := NcoSql.From(TestAlertsTable).
		Select(&AlertStatusRecord{}).
		Where(goqu.C("AlertGroup").Eq(alertGroup)).
		ToSQL()

	delQ, _, _ := NcoSql.Delete(TestAlertsTable).
		Where(goqu.C("AlertGroup").Eq(alertGroup)).
		ToSQL()

	return queries, selQ, delQ
}
