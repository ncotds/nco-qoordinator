package restapi

import (
	"encoding/json"
	"fmt"

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
	if qr.Error != nil {
		resp := gs.RawSQLResponse{
			ClusterName: name,
			Rows:        []gs.RawSQLResponseRowsItem{},
			Error:       gs.NewOptErrorResponse(errToResponseErr(qr.Error)),
		}
		return resp, nil
	}

	rows := make([]gs.RawSQLResponseRowsItem, 0, len(qr.RowSet.Rows))
	for i, row := range qr.RowSet.Rows {
		outRow, err := queryRowToResponseRow(qr.RowSet.Columns, row)
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
	return resp, nil
}

func queryRowToResponseRow(keys []string, values []any) (gs.RawSQLResponseRowsItem, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("keys and values count does not match")
	}

	out := make(gs.RawSQLResponseRowsItem, len(keys))
	for i, k := range keys {
		b, err := marshalValue(values[i])
		if err != nil {
			err = fmt.Errorf("%w: cannot marshal key %s value %v", err, k, values[i])
			return nil, err
		}
		out[k] = b
	}
	return out, nil
}

func marshalValue(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	return b, err
}
