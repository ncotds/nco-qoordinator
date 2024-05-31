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
//
// # Logger
//
// Logger - use NewLogger to set up.
//
// You can add some key-value pairs to operation context using WithLogAttrs and next,
// when you call any of Logger.DebugContext/Logger.InfoContext/Logger.WarnContext/Logger.ErrorContext methods,
// logger will add those key-value pairs into structured log record.
//
// Logger provides methods to make error logging easier: Logger.Err and Logger.ErrContext.
//
// Logger allows to change it level dynamically using Logger.SetLevel.
package app
