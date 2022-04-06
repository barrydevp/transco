package transco

import "errors"

var (
	ErrNotLeader       = NewRetryableError(errors.New("not leader"))
	ErrNoNodeAvailable = errors.New("no node available")
	errRetryable       = &RetryableError{}
)

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func NewRetryableError(err error) *RetryableError {
	return &RetryableError{err}
}

func IsRetryableErr(err error) bool {
	return errors.As(err, &errRetryable)
}
