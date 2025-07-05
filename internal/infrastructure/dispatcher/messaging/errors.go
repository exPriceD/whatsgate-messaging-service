package messaging

import "errors"

var (
	ErrDispatcherClosed = errors.New("dispatcher is closed")
)
