package dbconnector

import (
	"fmt"
)

var (
	ErrDBConnector = fmt.Errorf("DB error")
	ErrConnection  = fmt.Errorf("%w: cannot open connection", ErrDBConnector)
	ErrQuery       = fmt.Errorf("%w: query failed", ErrDBConnector)
)
