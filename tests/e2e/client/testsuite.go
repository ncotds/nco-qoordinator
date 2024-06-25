package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/ncotds/nco-qoordinator/pkg/config"
	"github.com/ncotds/nco-qoordinator/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestEnvPrefix = "TEST_E2E"

var (
	TestConfig = config.Config{
		LogLevel: "ERROR",
		HTTPServer: config.HTTPServerConfig{
			Listen: func() string {
				listen := "localhost"
				if addr := os.Getenv(TestEnvPrefix + "_LISTEN_HOST"); addr != "" {
					listen = addr
				}
				if port := os.Getenv(TestEnvPrefix + "_LISTEN_PORT"); port != "" {
					return listen + ":" + port
				}
				return listen
			}(),
		},
		OMNIbus: config.OMNIbus{
			Clusters: func() map[string]config.SeedList {
				res := make(map[string]config.SeedList)
				for i, addr := range strings.Split(os.Getenv(TestEnvPrefix+"_ADDRESS"), ",") {
					res[fmt.Sprintf("OMNI_%d", i)] = config.SeedList{addr}
				}
				return res
			}(),
			ConnectionLabel: "TEST_E2E_QOORDINATOR",
			MaxConnections:  10,
			RandomFailOver:  true,
		},
	}

	TestUser     = os.Getenv(TestEnvPrefix + "_USER")
	TestPassword = os.Getenv(TestEnvPrefix + "_PASSWORD")

	NcoSql = func() goqu.DialectWrapper {
		opts := goqu.DefaultDialectOptions()
		opts.QuoteRune = rune(32)
		opts.SupportsConflictTarget = false
		opts.SupportsWithCTE = false
		goqu.RegisterDialect("netcool", opts)
		return goqu.Dialect("netcool")
	}()

	reqTimeout = 1 * time.Second
)

func DoTestCRUD(t *testing.T, client Client) {
	creds := models.Credentials{UserName: TestUser, Password: TestPassword}
	row := AlertStatusRecordFactory()

	insQ, _, err := NcoSql.Insert(TestAlertsTable).Rows(row).ToSQL()
	require.NoError(t, err, "cannot make insert query")

	selQ, _, err := NcoSql.From(TestAlertsTable).
		Select(&AlertStatusRecord{}).
		Where(goqu.C("AlertGroup").Eq(row.AlertGroup)).
		ToSQL()
	require.NoError(t, err, "cannot make select query")

	updQ, _, err := NcoSql.Update(TestAlertsTable).
		Set(goqu.Record{"Summary": SentenceFactory()}).
		Where(goqu.C("AlertGroup").Eq(row.AlertGroup)).
		ToSQL()
	require.NoError(t, err, "cannot make update query")

	delQ, _, err := NcoSql.Delete(TestAlertsTable).
		Where(goqu.C("AlertGroup").Eq(row.AlertGroup)).
		ToSQL()
	require.NoError(t, err, "cannot make delete query")

	t.Run("insert", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		insResp, err := client.RawSQLPost(ctx, models.Query{SQL: insQ}, creds)
		if assert.NoError(t, err, "query fails") {
			assertResp(t, insResp, []AlertStatusRecord{}, 1)
		}
	})

	t.Run("select", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		selResp, err := client.RawSQLPost(ctx, models.Query{SQL: selQ}, creds)
		if assert.NoError(t, err, "query fails") {
			assertResp(t, selResp, []AlertStatusRecord{row}, 1)
		}
	})

	t.Run("update", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		updResp, err := client.RawSQLPost(ctx, models.Query{SQL: updQ}, creds)
		if assert.NoError(t, err, "query fails") {
			assertResp(t, updResp, []AlertStatusRecord{}, 1)
		}
	})

	t.Run("delete", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		delResp, err := client.RawSQLPost(ctx, models.Query{SQL: delQ}, creds)
		if assert.NoError(t, err, "query fails") {
			assertResp(t, delResp, []AlertStatusRecord{}, 1)
		}
	})

	t.Run("select after delete", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		selResp2, err := client.RawSQLPost(ctx, models.Query{SQL: selQ}, creds)
		if assert.NoError(t, err, "query fails") {
			assertResp(t, selResp2, []AlertStatusRecord{}, 0)
		}
	})
}

func assertResp(
	t *testing.T,
	actual map[string]QueryResult,
	expectedRows []AlertStatusRecord,
	expectedAffected int,
) {
	for k, elem := range actual {
		actualRows := make([]AlertStatusRecord, 0, len(elem.RowSet))
		for _, r := range elem.RowSet {
			actualRows = append(actualRows, NewAlertStatusRecordFromMap(r))
		}
		assert.ElementsMatchf(t, expectedRows, actualRows, "elem %s: %v", k, elem)
		assert.Equalf(t, expectedAffected, elem.AffectedRows, "elem %s: %v", k, elem)
		assert.NoErrorf(t, elem.Error, "elem %s: %v", k, elem)
	}
}
