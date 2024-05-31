package restapi

import (
	"context"

	gs "github.com/ncotds/nco-qoordinator/internal/restapi/gen"
	"github.com/ncotds/nco-qoordinator/pkg/app"
	qc "github.com/ncotds/nco-qoordinator/pkg/models"
)

type ctxKey int

const CtxKeyCredentials = ctxKey(1)

var (
	_ gs.SecurityHandler = (*SecurityHandler)(nil)
)

type SecurityHandler struct {
}

func (s SecurityHandler) HandleBasicAuth(
	ctx context.Context,
	_ string,
	t gs.BasicAuth,
) (context.Context, error) {
	if t.Username == "" || t.Password == "" {
		return ctx, app.Err(app.ErrCodeInsufficientPrivileges, "username and password are required")
	}
	credentials := qc.Credentials{UserName: t.Username, Password: t.Password}
	return context.WithValue(ctx, CtxKeyCredentials, credentials), nil
}

func GetCredentials(ctx context.Context) qc.Credentials {
	if cred, ok := ctx.Value(CtxKeyCredentials).(qc.Credentials); ok {
		return cred
	}
	return qc.Credentials{}
}
