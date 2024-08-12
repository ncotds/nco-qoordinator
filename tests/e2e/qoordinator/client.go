package qoordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	db "github.com/ncotds/nco-lib/dbconnector"

	"github.com/ncotds/nco-qoordinator/pkg/config"
	"github.com/ncotds/nco-qoordinator/tests/e2e/client"
)

var _ client.Client = (*Client)(nil)

type Client struct {
	htClient  *http.Client
	rawSqlUrl string
}

func NewClient(conf *config.Config) (client.Client, error) {
	cl := &Client{
		htClient:  &http.Client{},
		rawSqlUrl: "http://" + conf.HTTPServer.Listen + "/rawSQL",
	}
	return cl, nil
}

func (q *Client) RawSQLPost(
	_ context.Context,
	query db.Query,
	credentials db.Credentials,
) (map[string]client.QueryResult, error) {
	payload, _ := json.Marshal(query)
	req, _ := http.NewRequest(http.MethodPost, q.rawSqlUrl, bytes.NewBuffer(payload))
	req.Header.Add("X-Request-Id", uuid.NewString())
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(credentials.UserName, credentials.Password)

	resp, err := q.htClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		var errBody struct{ Error, Message string }
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("bad response code %d: %#v", resp.StatusCode, errBody)
	}

	var resultPayload []struct {
		ClientName   string
		Rows         []map[string]any
		AffectedRows int
		Error        *struct{ Error, Message, Reason string }
	}
	err = json.NewDecoder(resp.Body).Decode(&resultPayload)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot parse: %w", err)
	}

	result := make(map[string]client.QueryResult, len(resultPayload))
	for _, item := range resultPayload {
		qResult := client.QueryResult{
			RowSet:       item.Rows,
			AffectedRows: item.AffectedRows,
		}
		if item.Error != nil {
			qResult.Error = fmt.Errorf("%+v", item.Error)
		}
		result[item.ClientName] = qResult
	}
	return result, nil
}
