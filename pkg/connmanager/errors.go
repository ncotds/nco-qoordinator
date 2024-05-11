package connmanager

import (
	"fmt"
)

var (
	// ErrConnManager is a base error for the package
	ErrConnManager = fmt.Errorf("connection manager fails")

	ErrBadConfiguration = fmt.Errorf("%w: configuration error", ErrConnManager)

	ErrConnectionLimit   = fmt.Errorf("%w: connections limit exceed", ErrConnManager)
	ErrConnectionRelease = fmt.Errorf("%w: cannot release connection", ErrConnManager)

	ErrPoolClosed = fmt.Errorf("%w: pool is closed already", ErrConnManager)
)
