package restapi

import (
	"context"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	qc "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
	"github.com/ncotds/nco-qoordinator/pkg/restapi/gen"
)

var _ gen.Handler = (*CoordinatorService)(nil)

type CoordinatorService struct {
	Coordinator *qc.QueryCoordinator
}

func (c CoordinatorService) ClusterNamesGet(_ context.Context, _ gen.ClusterNamesGetParams) ([]string, error) {
	return c.Coordinator.ClientNames(), nil
}

func (c CoordinatorService) RawSQLPost(
	ctx context.Context,
	req *gen.RawSQLRequest,
	_ gen.RawSQLPostParams,
) (gen.RawSQLListResponse, error) {
	qRes := c.Coordinator.Exec(ctx, qc.Query{SQL: req.SQL}, GetCredentials(ctx), req.Clusters...)

	listResp := make(gen.RawSQLListResponse, 0, len(qRes))
	for name, qr := range qRes {
		resp, err := queryResultToResponse(name, qr)
		if err != nil {
			return nil, app.Err(app.ErrCodeUnknown, "cannot parse response", err)
		}
		listResp = append(listResp, resp)
	}
	return listResp, nil
}

// NewError handles not-error returned by 'action' methods
func (c CoordinatorService) NewError(_ context.Context, err error) *gen.ErrorResponseStatusCode {
	return &gen.ErrorResponseStatusCode{
		StatusCode: errCode(err),
		Response:   errToResponseErr(err),
	}
}
