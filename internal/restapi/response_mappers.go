package restapi

import (
	"encoding/json"
	"fmt"

	"github.com/go-faster/jx"
	gs "github.com/ncotds/nco-qoordinator/internal/restapi/gen"
	"github.com/ncotds/nco-qoordinator/pkg/models"
)

func errToResponseErr(in error) (out gs.ErrorResponse) {
	if in == nil {
		return out
	}

	appErr := appError(in)

	out.Error = gs.ErrorResponseError(appErr.Code())
	out.Message = appErr.Message()
	if reason := appErr.Unwrap(); reason != nil {
		out.Reason = gs.NewOptString(reason.Error())
	}

	return out
}

func queryResultToResponse(name string, qr models.QueryResult) (gs.RawSQLResponse, error) {
	rows := make([]gs.RawSQLResponseRowsItem, 0, len(qr.RowSet))
	for i, row := range qr.RowSet {
		outRow, err := queryRowToResponseRow(row)
		if err != nil {
			err = fmt.Errorf("%w: fail on row %d of %s response", err, i, name)
			return gs.RawSQLResponse{}, err
		}
		rows = append(rows, outRow)
	}

	resp := gs.RawSQLResponse{
		ClusterName:  name,
		Rows:         rows,
		AffectedRows: qr.AffectedRows,
	}
	if qr.Error != nil {
		resp.Error = gs.NewOptErrorResponse(errToResponseErr(qr.Error))
	}
	return resp, nil
}

func queryRowToResponseRow(in models.QueryResultRow) (map[string]jx.Raw, error) {
	out := make(map[string]jx.Raw, len(in))
	for k, v := range in {
		b, err := json.Marshal(v)
		if err != nil {
			err = fmt.Errorf("%w: cannot marshal key %s value %v", err, k, v)
			return nil, err
		}
		out[k] = b
	}
	return out, nil
}
