package mycontext

// DeadlineExceeded is the error returned by Context.Err when the mycontext's
// deadline passes.
var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string {
	return "context deadline exceeded"
}

func (deadlineExceededError) Timeout() bool {
	return true
}
