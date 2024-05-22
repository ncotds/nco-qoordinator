package app

const (
	// ErrCodeUnknown is the worst case: app has not any suggestion, how to handle it.
	// Caller should decide itself, repeat request ot not
	ErrCodeUnknown = "ERR_UNKNOWN"
)

// The following errors represents temporal problems on the server side, we expect
// it will be resolved soon... so caller can repeat request a bit later and get success response
//
//   - ErrCodeTimeout is given when the app cannot perform action
//     (or get response from external system) in time
//   - ErrCodeUnavailable means the app is currently unable to perform action (or external request),
//     but we expect that it will be resolved soon
//   - ErrCodeSystem - unclassified error on server side
const (
	ErrCodeTimeout     = "ERR_TIMEOUT"
	ErrCodeUnavailable = "ERR_UNAVAILABLE"
	ErrCodeSystem      = "ERR_SYSTEM"
)

// The following errors represents problems on client side, so caller must change request to
// get success response
//
//   - ErrCodeNotExists
//   - ErrCodeValidation
//   - ErrCodeInvalidArgument
//   - ErrCodeIncorrectOperation
//   - ErrCodeInsufficientPrivileges
const (
	ErrCodeNotExists              = "ERR_NOT_FOUND"
	ErrCodeValidation             = "ERR_VALIDATION"
	ErrCodeInvalidArgument        = "ERR_INVALID_ARGUMENT"
	ErrCodeIncorrectOperation     = "ERR_INCORRECT_OPERATION"
	ErrCodeInsufficientPrivileges = "ERR_INSUFFICIENT_PRIVILEGES"
	// NOTE: add app specific error codes below
	// ErrCodeLowBalance = "ERR_LOW_BALANCE"
)

// Sentinel errors to compare using errors.Is/As
var (
	ErrUnknown                = Error{code: ErrCodeUnknown, message: "Unknown error"}
	ErrTimeout                = Error{code: ErrCodeTimeout, message: "Operation timeout"}
	ErrUnavailable            = Error{code: ErrCodeUnavailable, message: "Service cannot handle requests"}
	ErrSystem                 = Error{code: ErrCodeSystem, message: "Unclassified system error"}
	ErrNotExists              = Error{code: ErrCodeNotExists, message: "Object not found"}
	ErrValidation             = Error{code: ErrCodeValidation, message: "Bad request"}
	ErrInvalidArgument        = Error{code: ErrCodeInvalidArgument, message: "Invalid value"}
	ErrIncorrectOperation     = Error{code: ErrCodeIncorrectOperation, message: "Incorrect operation"}
	ErrInsufficientPrivileges = Error{code: ErrCodeInsufficientPrivileges, message: "Not enough permission to perform this action"}
	// NOTE: add app specific errors below
	// ErrLowBalance = Error{code: ErrCodeLowBalance, message: "Account too low balance"}
)
