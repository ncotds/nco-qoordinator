package qoordinator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ncotds/nco-qoordinator/tests/e2e/client"
)

func TestQoordinatorCRUD(t *testing.T) {
	cl, err := NewClient(&client.TestConfig)
	require.NoError(t, err, "cannot setup client")

	client.DoTestCRUD(t, cl)
}
