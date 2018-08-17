package client_test

import (
	"errors"

	. "github.com/onsi/gomega"
	"testing"
	"time"
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
	err := client.Do(maxRetries, 0, toRetry)

	// then
	Expect(err).To(HaveLen(maxRetries))
	Expect(executions).To(Equal(maxRetries))
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
	err := client.Do(maxRetries, 0, toRetry)

	// then
	Expect(err).To(HaveLen(1))
	Expect(executions).To(Equal(1))
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
	err := client.Do(10, time.Millisecond * 50, toRetry)

	// then
	Expect(err).To(BeEmpty())
	Expect(executions).To(Equal(3))
}
