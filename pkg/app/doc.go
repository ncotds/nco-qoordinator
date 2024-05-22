// Package app provides some tools useful for typical Go project.
//
// # Error
//
// Custom Error type and set of standard errors:
//   - ErrUnknown
//   - ErrTimeout
//   - ErrUnavailable
//   - ErrSystem
//   - ErrNotExists
//   - ErrValidation
//   - ErrInvalidArgument
//   - ErrIncorrectOperation
//   - ErrInsufficientPrivileges
//
// Use Err function to create Error instance with custom code, message and optional reason.
//
// Note, that Error instances matched by code:
//
//	errors.Is(app.Err("CODE", "msg1"), app.Err("CODE", "msg2"))  // true
//	errors.Is(app.Err(app.ErrCodeValidation, "msg1"), app.ErrValidation)  // true
//	errors.Is(app.Err(app.ErrCodeValidation, "msg1"), app.ErrUnavailable)  // false
package app
