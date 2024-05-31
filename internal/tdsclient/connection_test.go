//go:build integration
// +build integration

package tdsclient_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/doug-martin/goqu/v9"
	tds "github.com/minus5/gofreetds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/ncotds/nco-qoordinator/internal/dbconnector"
	"github.com/ncotds/nco-qoordinator/internal/tdsclient"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

func TestConnection_Close(t *testing.T) {
	ctx := context.Background()
	credentials := models.Credentials{TestConfig.User, TestConfig.Password}

	client := &tdsclient.TDSConnector{TestConnLabel, TestConnTimeoutSec}
	conn, err := client.Connect(ctx, TestConfig.Address, credentials)
	require.NoError(t, err, "cannot establish connection")

	assert.NoError(t, conn.Close(), "first close")
	assert.NoError(t, conn.Close(), "second close")
}

func TestConnection_Exec_Insert(t *testing.T) {
	conn, _ := setUpTest(t, 0)
	defer tearDownTest(t, conn)

	toInsert := AlertStatusRecordFactory()
	stmt, _, _ := NcoSql.Insert(TestAlertsTable).Rows(toInsert).ToSQL()

	gotRows, gotAffected, gotErr := conn.Exec(context.Background(), models.Query{SQL: stmt})

	assert.Zero(t, gotRows)
	assert.Equal(t, 1, gotAffected)
	assert.NoError(t, gotErr)

	rowsFromDB, _ := selTestRows()
	assert.Equal(t, []AlertStatusRecord{toInsert}, rowsFromDB)
}

func TestConnection_Exec_Select(t *testing.T) {
	conn, existingRows := setUpTest(t, 3)
	defer tearDownTest(t, conn)

	var existingIds []interface{}
	for _, r := range existingRows {
		existingIds = append(existingIds, r.Identifier)
	}
	stmt, _, _ := NcoSql.From(TestAlertsTable).
		Select(&AlertStatusRecord{}).
		Where(goqu.C("Identifier").In(existingIds...)).
		ToSQL()

	gotRows, gotAffected, gotErr := conn.Exec(context.Background(), models.Query{SQL: stmt})

	var gotRecords []AlertStatusRecord
	for _, r := range gotRows {
		gotRecords = append(gotRecords, NewAlertStatusRecordFromMap(r))
	}

	assert.ElementsMatch(t, existingRows, gotRecords)
	assert.Equal(t, len(existingRows), gotAffected)
	assert.NoError(t, gotErr)
}

func TestConnection_Exec_Update(t *testing.T) {
	conn, existingRows := setUpTest(t, 3)
	defer tearDownTest(t, conn)

	updSummary := SentenceFactory()
	var existingIds []interface{}
	for _, r := range existingRows {
		existingIds = append(existingIds, r.Identifier)
	}
	stmt, _, _ := NcoSql.Update(TestAlertsTable).
		Set(goqu.Record{"Summary": updSummary}).
		Where(goqu.C("Identifier").In(existingIds...)).
		ToSQL()

	gotRows, gotAffected, gotErr := conn.Exec(context.Background(), models.Query{SQL: stmt})

	assert.Zero(t, gotRows)
	assert.Equal(t, len(existingRows), gotAffected)
	assert.NoError(t, gotErr)

	var expectedRows []AlertStatusRecord
	for _, r := range existingRows {
		r.Summary = updSummary
		expectedRows = append(expectedRows, r)
	}
	rowsFromDB, _ := selTestRows()
	assert.ElementsMatch(t, expectedRows, rowsFromDB)
}

func TestConnection_Exec_Delete(t *testing.T) {
	conn, existingRows := setUpTest(t, 3)
	defer tearDownTest(t, conn)

	var existingIds []interface{}
	for _, r := range existingRows {
		existingIds = append(existingIds, r.Identifier)
	}
	stmt, _, _ := NcoSql.Delete(TestAlertsTable).Where(goqu.C("Identifier").In(existingIds...)).ToSQL()

	gotRows, gotAffected, gotErr := conn.Exec(context.Background(), models.Query{SQL: stmt})

	assert.Zero(t, gotRows)
	assert.Equal(t, len(existingRows), gotAffected)
	assert.NoError(t, gotErr)

	rowsFromDB, _ := selTestRows()
	assert.Empty(t, rowsFromDB)
}

func TestConnection_Exec_ContextCancel(t *testing.T) {
	conn, _ := setUpTest(t, 0)
	defer tearDownTest(t, conn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := conn.Exec(ctx, models.Query{SQL: "describe status"})

	assert.ErrorIs(t, err, context.Canceled)
}

func TestConnection_Exec_Reconnect(t *testing.T) {
	conn, existingRows := setUpTest(t, 1)
	defer tearDownTest(t, conn)

	err := conn.Close()
	require.NoError(t, err, "cannot close connection")

	stmt, _, err := NcoSql.
		Select(goqu.L("top 1 Identifier")).
		From(TestAlertsTable).
		Where(goqu.C("Identifier").Eq(existingRows[0].Identifier)).
		ToSQL()
	require.NoError(t, err, "cannot make SQL stmt")

	rows, affected, err := conn.Exec(context.Background(), models.Query{SQL: stmt})

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
	assert.Len(t, rows, 1)
}

func TestConnection_Exec_ReconnectCancel(t *testing.T) {
	conn, _ := setUpTest(t, 0)
	defer tearDownTest(t, conn)

	err := conn.Close()
	require.NoError(t, err, "cannot close connection")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = conn.Exec(ctx, models.Query{SQL: "describe status"})

	assert.ErrorIs(t, err, context.Canceled)
}

func setUpTest(t *testing.T, rowsCount int) (conn db.ExecutorCloser, rows []AlertStatusRecord) {
	require.NoError(t, delTestRows(), "cannot delete previous test data")
	for i := 0; i < rowsCount; i++ {
		row, err := insTestRow()
		require.NoError(t, err, "cannot insert test data")
		rows = append(rows, row)
	}

	ctx := context.Background()
	credentials := models.Credentials{TestConfig.User, TestConfig.Password}

	client := &tdsclient.TDSConnector{TestConnLabel, TestConnTimeoutSec}
	conn, err := client.Connect(ctx, TestConfig.Address, credentials)
	require.NoError(t, err, "cannot establish connection")

	return conn, rows
}

func tearDownTest(t *testing.T, conn db.ExecutorCloser) {
	require.NoError(t, delTestRows(), "cannot clear test data")
	require.NoError(t, conn.Close(), "cannot close connection")
}

func makeConn() (*tds.Conn, error) {
	connStr := fmt.Sprintf(
		"user=%s;pwd=%s;host=%s;app=%s;compatibility=%s;tds_version=%s",
		TestConfig.User, TestConfig.Password, TestConfig.Address,
		TestConnLabel, TestCompatibilityMode, TestTDSVersion,
	)
	return tds.NewConn(connStr)
}

func insTestRow() (AlertStatusRecord, error) {
	insRow := AlertStatusRecordFactory()

	conn, err := makeConn()
	if err != nil {
		return insRow, err
	}
	defer conn.Close()

	insQuery, _, err := NcoSql.Insert(TestAlertsTable).Rows(insRow).ToSQL()
	if err != nil {
		return insRow, err
	}
	_, err = conn.Exec(insQuery)
	return insRow, err
}

func delTestRows() error {
	conn, err := makeConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	delQuery, _, err := NcoSql.Delete(TestAlertsTable).
		Where(goqu.Ex{
			"Manager": TestManager,
			"Agent":   TestAgent,
		}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = conn.Exec(delQuery)
	return err
}

func selTestRows() (rows []AlertStatusRecord, err error) {
	conn, err := makeConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	selQuery, _, err := NcoSql.From(TestAlertsTable).
		Select(&AlertStatusRecord{}).
		Where(goqu.Ex{"Manager": TestManager, "Agent": TestAgent}).
		ToSQL()
	if err != nil {
		return nil, err
	}
	rst, err := conn.Exec(selQuery)
	if err != nil {
		return nil, err
	}

	cursor := rst[0]
	for _, row := range cursor.Rows {
		record := AlertStatusRecord{}
		rv := reflect.ValueOf(&record).Elem()
		for i, col := range cursor.Columns {
			field := rv.FieldByName(col.Name)
			field.Set(reflect.ValueOf(row[i]).Convert(field.Type()))
		}
		rows = append(rows, record)
	}

	return rows, err
}
