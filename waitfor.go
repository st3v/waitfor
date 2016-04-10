package waitfor

import (
	"errors"
	"time"

	"golang.org/x/net/context"
)

var ErrTimeoutExceeded = errors.New("timeout exceeded")

type Check func() bool

func ConditionWithTimeout(condition Check, interval, timeout time.Duration) error {
	errChan := make(chan error)
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	go Condition(condition, interval, errChan, ctx)

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ErrTimeoutExceeded
	}
}

func Condition(condition Check, interval time.Duration, errChan chan error, ctx context.Context) {
	if condition() {
		errChan <- nil
		return
	}

	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		case <-time.After(interval):
			if condition() {
				errChan <- nil
				return
			}
		}
	}
}
