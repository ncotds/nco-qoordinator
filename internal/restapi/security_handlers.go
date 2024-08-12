package restapi

import (
	"context"

	db "github.com/ncotds/nco-lib/dbconnector"

	gs "github.com/ncotds/nco-qoordinator/internal/restapi/gen"
	"github.com/ncotds/nco-qoordinator/pkg/app"
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
	credentials := db.Credentials{UserName: t.Username, Password: t.Password}
	return context.WithValue(ctx, CtxKeyCredentials, credentials), nil
}

func GetCredentials(ctx context.Context) db.Credentials {
	if cred, ok := ctx.Value(CtxKeyCredentials).(db.Credentials); ok {
		return cred
	}
	return db.Credentials{}
}
