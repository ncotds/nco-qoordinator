package restapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/ncotds/nco-qoordinator/pkg/app"
	"github.com/ncotds/nco-qoordinator/pkg/app/httpserver"
	"github.com/ogen-go/ogen/ogenerrors"
)

var errorCodeToHttpCodeMap = map[string]int{
	app.ErrCodeUnknown:                http.StatusInternalServerError,
	app.ErrCodeTimeout:                http.StatusServiceUnavailable,
	app.ErrCodeUnavailable:            http.StatusServiceUnavailable,
	app.ErrCodeSystem:                 http.StatusServiceUnavailable,
	app.ErrCodeNotExists:              http.StatusBadRequest, // to make difference with 'no such route'
	app.ErrCodeValidation:             http.StatusBadRequest,
	app.ErrCodeInvalidArgument:        http.StatusBadRequest,
	app.ErrCodeIncorrectOperation:     http.StatusBadRequest,
	app.ErrCodeInsufficientPrivileges: http.StatusForbidden,
}

func errorHandler(_ context.Context, w http.ResponseWriter, r *http.Request, err error) {
	httpserver.ErrJSON(w, r, errCode(err), appError(err))
}

func errCode(err error) int {
	code := http.StatusInternalServerError

	var (
		appErr  app.Error
		ogenErr ogenerrors.Error
	)

	switch {
	case errors.As(err, &appErr):
		if c, ok := errorCodeToHttpCodeMap[appErr.Code()]; ok {
			code = c
		}
	case errors.As(err, &ogenErr):
		code = ogenErr.Code()
	}

	return code
}

func appError(in error) app.Error {
	if in == nil {
		return app.Err(app.ErrCodeUnknown, "")
	}

	var (
		appErr     app.Error
		ogenErr    ogenerrors.Error
		ogenSecErr *ogenerrors.SecurityError
	)

	switch {
	case errors.As(in, &appErr):
		return appErr
	case errors.As(in, &ogenSecErr):
		return app.Err(app.ErrCodeInsufficientPrivileges, ogenSecErr.Error())
	case errors.As(in, &ogenErr):
		return app.Err(app.ErrCodeValidation, ogenErr.Error())
	default:
		return app.Err(app.ErrCodeUnknown, in.Error())
	}
}
