package store

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/lib/pq"
)

const (
	retryCount               = 10
	retryInitialInterval     = 50 * time.Millisecond
	retryRandomizationFactor = 0.2
	retryMultiplier          = 1.3
	retryMaxInterval         = 5 * time.Second
	retryMaxElapsedTime      = 10 * time.Second

	pqSerializationErrorCode = pq.ErrorCode("40001")
)

func RetryOnPostgresError(f func() error) error {
	e := func(err error) bool {
		pqErr, ok := err.(*pq.Error)
		if ok && containsPQError(pqErr, pqSerializationErrorCode) {
			return true
		}
		return false
	}
	return retry(f, e)
}

func retry(f func() error, e func(err error) bool) error {
	exp := backoff.NewExponentialBackOff()
	exp.InitialInterval = retryInitialInterval
	exp.RandomizationFactor = retryRandomizationFactor
	exp.Multiplier = retryMultiplier
	exp.MaxInterval = retryMaxInterval
	exp.MaxElapsedTime = retryMaxElapsedTime

	var err error
	for n := 0; n < retryCount; n++ {
		err = f()
		if err == nil {
			return nil
		}
		if !e(err) {
			return err
		}
		time.Sleep(exp.NextBackOff())
	}
	return err
}

func containsPQError(pqError *pq.Error, codes ...pq.ErrorCode) bool {
	for _, ec := range codes {
		if pqError.Code == ec {
			return true
		}
	}
	return false
}
