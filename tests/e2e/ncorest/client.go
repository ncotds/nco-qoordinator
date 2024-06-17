package ncorest

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
	htClient   *http.Client
	rawSqlUrls map[string]string
}

func NewClient(conf *config.Config) (client.Client, error) {
	urls := make(map[string]string, len(conf.OMNIbus.Clusters))
	for name, item := range conf.OMNIbus.Clusters {
		if len(item) < 1 {
			return nil, fmt.Errorf("cluster %s has no any addr to connect", name)
		}
		urls[name] = "http://" + item[0] + "/objectserver/restapi/sql/factory"
	}
	cl := &Client{
		htClient:   &http.Client{},
		rawSqlUrls: urls,
	}
	return cl, nil
}

func (n Client) RawSQLPost(
	_ context.Context,
	query models.Query,
	credentials models.Credentials,
) (map[string]models.QueryResult, error) {
	responses := make(chan struct {
		name string
		models.QueryResult
	})

	for name, url := range n.rawSqlUrls {
		go func(name, url string) {
			resp, err := n.doRawSQLReq(url, query, credentials)
			qResult := models.QueryResult{
				AffectedRows: resp.Rowset.AffectedRows,
				RowSet:       resp.Rowset.Rows,
				Error:        err,
			}

			responses <- struct {
				name string
				models.QueryResult
			}{
				name:        name,
				QueryResult: qResult,
			}
		}(name, url)
	}

	result := make(map[string]models.QueryResult, len(n.rawSqlUrls))
	for i := 0; i < len(n.rawSqlUrls); i++ {
		resp := <-responses
		result[resp.name] = resp.QueryResult
	}

	return result, nil
}

func (n Client) doRawSQLReq(
	url string,
	query models.Query,
	credentials models.Credentials,
) (*SQLFactoryResponse, error) {
	payload, _ := json.Marshal(SQLFactoryRequest{Sqlcmd: query.SQL})
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	req.Header.Add("X-Request-Id", uuid.NewString())
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Connection", "Keep-Alive")
	req.SetBasicAuth(credentials.UserName, credentials.Password)

	resp, err := n.htClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		var errBody interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("bad response code %d: %#v", resp.StatusCode, errBody)
	}

	result := &SQLFactoryResponse{}
	err = json.NewDecoder(resp.Body).Decode(result)
	return result, err
}

type SQLFactoryRequest struct {
	Sqlcmd string `json:"sqlcmd"`
}

type SQLFactoryResponse struct {
	Rowset struct {
		AffectedRows int
		Rows         []models.QueryResultRow `json:"rows"`
	} `json:"rowset"`
}
