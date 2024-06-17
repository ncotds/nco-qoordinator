package ncorest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ncotds/nco-qoordinator/tests/e2e/client"
)

func TestNcorestCRUD(t *testing.T) {
	cl, err := NewClient(&client.TestConfig)
	require.NoError(t, err, "cannot setup client")

	client.DoTestCRUD(t, cl)
}
