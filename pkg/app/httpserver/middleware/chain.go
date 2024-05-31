package middleware

import (
	"net/http"

	"github.com/ncotds/nco-qoordinator/pkg/app"
)

// DefaultChain adds the following middlewares to the handler:
//   - RequireRequestId
//   - RequestLogger(log)
//   - Recoverer(WithLogger(log))
func DefaultChain(h http.Handler, log *app.Logger) http.Handler {
	return Chain(
		h,
		RequireRequestId,
		RequestLogger(log),
		Recoverer(log),
	)
}

// Chain wraps the handler with provided middlewares:
//
// h = mw_0(mw_1( ... mw_N-1(mw_N(h)) ... ))
func Chain(h http.Handler, middlewares ...func(next http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
