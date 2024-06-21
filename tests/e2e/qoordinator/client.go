package qoordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/ncotds/nco-qoordinator/pkg/config"
	"github.com/ncotds/nco-qoordinator/pkg/models"
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
	query models.Query,
	credentials models.Credentials,
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
	}
	err = json.NewDecoder(resp.Body).Decode(&resultPayload)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot parse")
	}

	result := make(map[string]client.QueryResult, len(resultPayload))
	for _, item := range resultPayload {
		result[item.ClientName] = client.QueryResult{
			RowSet:       item.Rows,
			AffectedRows: item.AffectedRows,
		}
	}
	return result, nil
}
