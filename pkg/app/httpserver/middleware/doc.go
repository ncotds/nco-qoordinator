// Package middleware contains common used HTTP-server middlewares:
//   - RequireRequestId
//   - RequestLogger
//   - Recoverer
//
// ... and DefaultChain func to add those middlewares to http-handler.
//
// All middlewares are compatible with std http.Handler
package middleware
