package retry_test

import (
	"errors"

	"testing"
	"time"
	"github.com/stretchr/testify/require"
	"github.com/fabric8-services/fabric8-tenant/retry"
)

func TestAccumulateErrorsWhenAllFailed(t *testing.T) {
	// given
	maxRetries := 4
	executions := 0
	toRetry := func() error {
		executions++
		return errors.New("unauthorized")
	}

	// when
	err := retry.Do(maxRetries, 0, toRetry)

	// then
	require.Len(t, err, maxRetries)
	require.Equal(t, executions, maxRetries)
}

func TestRetryExecuteOnce(t *testing.T) {
	// given
	maxRetries := 0
	executions := 0
	toRetry := func() error {
		executions++
		return errors.New("unauthorized")
	}

	// when
	err := retry.Do(maxRetries, 0, toRetry)

	// then
	require.Len(t, err, 1)
	require.Equal(t, executions, 1)
}

func TestStopRetryingWhenSuccessful(t *testing.T) {
	// given
	executions := 0
	toRetry := func() error {
		executions++
		if executions == 3 {
			return nil
		}
		return errors.New("not found")
	}

	// when
	err := retry.Do(10, time.Millisecond*50, toRetry)

	// then
	require.Empty(t, err)
	require.Equal(t, executions, 3)
}
