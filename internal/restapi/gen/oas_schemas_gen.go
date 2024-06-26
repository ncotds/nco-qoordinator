// Code generated by ogen, DO NOT EDIT.

package gen

import (
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

func (s *ErrorResponseStatusCode) Error() string {
	return fmt.Sprintf("code %d: %+v", s.StatusCode, s.Response)
}

type BasicAuth struct {
	Username string
	Password string
}

// GetUsername returns the value of Username.
func (s *BasicAuth) GetUsername() string {
	return s.Username
}

// GetPassword returns the value of Password.
func (s *BasicAuth) GetPassword() string {
	return s.Password
}

// SetUsername sets the value of Username.
func (s *BasicAuth) SetUsername(val string) {
	s.Username = val
}

// SetPassword sets the value of Password.
func (s *BasicAuth) SetPassword(val string) {
	s.Password = val
}

// Ref: #/components/schemas/ErrorResponse
type ErrorResponse struct {
	// Error 'business' code.
	Error ErrorResponseError `json:"error"`
	// Error message.
	Message string `json:"message"`
	// Details about error.
	Reason OptString `json:"reason"`
}

// GetError returns the value of Error.
func (s *ErrorResponse) GetError() ErrorResponseError {
	return s.Error
}

// GetMessage returns the value of Message.
func (s *ErrorResponse) GetMessage() string {
	return s.Message
}

// GetReason returns the value of Reason.
func (s *ErrorResponse) GetReason() OptString {
	return s.Reason
}

// SetError sets the value of Error.
func (s *ErrorResponse) SetError(val ErrorResponseError) {
	s.Error = val
}

// SetMessage sets the value of Message.
func (s *ErrorResponse) SetMessage(val string) {
	s.Message = val
}

// SetReason sets the value of Reason.
func (s *ErrorResponse) SetReason(val OptString) {
	s.Reason = val
}

// Error 'business' code.
type ErrorResponseError string

const (
	ErrorResponseErrorERRUNKNOWN                ErrorResponseError = "ERR_UNKNOWN"
	ErrorResponseErrorERRTIMEOUT                ErrorResponseError = "ERR_TIMEOUT"
	ErrorResponseErrorERRUNAVAILABLE            ErrorResponseError = "ERR_UNAVAILABLE"
	ErrorResponseErrorERRVALIDATION             ErrorResponseError = "ERR_VALIDATION"
	ErrorResponseErrorERRINCORRECTOPERATION     ErrorResponseError = "ERR_INCORRECT_OPERATION"
	ErrorResponseErrorERRINSUFFICIENTPRIVILEGES ErrorResponseError = "ERR_INSUFFICIENT_PRIVILEGES"
)

// AllValues returns all ErrorResponseError values.
func (ErrorResponseError) AllValues() []ErrorResponseError {
	return []ErrorResponseError{
		ErrorResponseErrorERRUNKNOWN,
		ErrorResponseErrorERRTIMEOUT,
		ErrorResponseErrorERRUNAVAILABLE,
		ErrorResponseErrorERRVALIDATION,
		ErrorResponseErrorERRINCORRECTOPERATION,
		ErrorResponseErrorERRINSUFFICIENTPRIVILEGES,
	}
}

// MarshalText implements encoding.TextMarshaler.
func (s ErrorResponseError) MarshalText() ([]byte, error) {
	switch s {
	case ErrorResponseErrorERRUNKNOWN:
		return []byte(s), nil
	case ErrorResponseErrorERRTIMEOUT:
		return []byte(s), nil
	case ErrorResponseErrorERRUNAVAILABLE:
		return []byte(s), nil
	case ErrorResponseErrorERRVALIDATION:
		return []byte(s), nil
	case ErrorResponseErrorERRINCORRECTOPERATION:
		return []byte(s), nil
	case ErrorResponseErrorERRINSUFFICIENTPRIVILEGES:
		return []byte(s), nil
	default:
		return nil, errors.Errorf("invalid value: %q", s)
	}
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *ErrorResponseError) UnmarshalText(data []byte) error {
	switch ErrorResponseError(data) {
	case ErrorResponseErrorERRUNKNOWN:
		*s = ErrorResponseErrorERRUNKNOWN
		return nil
	case ErrorResponseErrorERRTIMEOUT:
		*s = ErrorResponseErrorERRTIMEOUT
		return nil
	case ErrorResponseErrorERRUNAVAILABLE:
		*s = ErrorResponseErrorERRUNAVAILABLE
		return nil
	case ErrorResponseErrorERRVALIDATION:
		*s = ErrorResponseErrorERRVALIDATION
		return nil
	case ErrorResponseErrorERRINCORRECTOPERATION:
		*s = ErrorResponseErrorERRINCORRECTOPERATION
		return nil
	case ErrorResponseErrorERRINSUFFICIENTPRIVILEGES:
		*s = ErrorResponseErrorERRINSUFFICIENTPRIVILEGES
		return nil
	default:
		return errors.Errorf("invalid value: %q", data)
	}
}

// ErrorResponseStatusCode wraps ErrorResponse with StatusCode.
type ErrorResponseStatusCode struct {
	StatusCode int
	Response   ErrorResponse
}

// GetStatusCode returns the value of StatusCode.
func (s *ErrorResponseStatusCode) GetStatusCode() int {
	return s.StatusCode
}

// GetResponse returns the value of Response.
func (s *ErrorResponseStatusCode) GetResponse() ErrorResponse {
	return s.Response
}

// SetStatusCode sets the value of StatusCode.
func (s *ErrorResponseStatusCode) SetStatusCode(val int) {
	s.StatusCode = val
}

// SetResponse sets the value of Response.
func (s *ErrorResponseStatusCode) SetResponse(val ErrorResponse) {
	s.Response = val
}

// NewOptErrorResponse returns new OptErrorResponse with value set to v.
func NewOptErrorResponse(v ErrorResponse) OptErrorResponse {
	return OptErrorResponse{
		Value: v,
		Set:   true,
	}
}

// OptErrorResponse is optional ErrorResponse.
type OptErrorResponse struct {
	Value ErrorResponse
	Set   bool
}

// IsSet returns true if OptErrorResponse was set.
func (o OptErrorResponse) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptErrorResponse) Reset() {
	var v ErrorResponse
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptErrorResponse) SetTo(v ErrorResponse) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptErrorResponse) Get() (v ErrorResponse, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptErrorResponse) Or(d ErrorResponse) ErrorResponse {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptString returns new OptString with value set to v.
func NewOptString(v string) OptString {
	return OptString{
		Value: v,
		Set:   true,
	}
}

// OptString is optional string.
type OptString struct {
	Value string
	Set   bool
}

// IsSet returns true if OptString was set.
func (o OptString) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptString) Reset() {
	var v string
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptString) SetTo(v string) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptString) Get() (v string, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptString) Or(d string) string {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

type RawSQLListResponse []RawSQLResponse

// Ref: #/components/schemas/RawSQLRequest
type RawSQLRequest struct {
	// SQL query to execute.
	SQL string `json:"sql"`
	// List of cluster names to send the query to.
	Clusters []string `json:"clusters"`
}

// GetSQL returns the value of SQL.
func (s *RawSQLRequest) GetSQL() string {
	return s.SQL
}

// GetClusters returns the value of Clusters.
func (s *RawSQLRequest) GetClusters() []string {
	return s.Clusters
}

// SetSQL sets the value of SQL.
func (s *RawSQLRequest) SetSQL(val string) {
	s.SQL = val
}

// SetClusters sets the value of Clusters.
func (s *RawSQLRequest) SetClusters(val []string) {
	s.Clusters = val
}

// Ref: #/components/schemas/RawSQLResponse
type RawSQLResponse struct {
	ClusterName string `json:"clusterName"`
	// Rows returned from the cluster.
	Rows []RawSQLResponseRowsItem `json:"rows"`
	// Number of rows affected by the query.
	AffectedRows int              `json:"affectedRows"`
	Error        OptErrorResponse `json:"error"`
}

// GetClusterName returns the value of ClusterName.
func (s *RawSQLResponse) GetClusterName() string {
	return s.ClusterName
}

// GetRows returns the value of Rows.
func (s *RawSQLResponse) GetRows() []RawSQLResponseRowsItem {
	return s.Rows
}

// GetAffectedRows returns the value of AffectedRows.
func (s *RawSQLResponse) GetAffectedRows() int {
	return s.AffectedRows
}

// GetError returns the value of Error.
func (s *RawSQLResponse) GetError() OptErrorResponse {
	return s.Error
}

// SetClusterName sets the value of ClusterName.
func (s *RawSQLResponse) SetClusterName(val string) {
	s.ClusterName = val
}

// SetRows sets the value of Rows.
func (s *RawSQLResponse) SetRows(val []RawSQLResponseRowsItem) {
	s.Rows = val
}

// SetAffectedRows sets the value of AffectedRows.
func (s *RawSQLResponse) SetAffectedRows(val int) {
	s.AffectedRows = val
}

// SetError sets the value of Error.
func (s *RawSQLResponse) SetError(val OptErrorResponse) {
	s.Error = val
}

type RawSQLResponseRowsItem map[string]jx.Raw

func (s *RawSQLResponseRowsItem) init() RawSQLResponseRowsItem {
	m := *s
	if m == nil {
		m = map[string]jx.Raw{}
		*s = m
	}
	return m
}
