package payment

import "errors"

// consensus errors
var (
	ErrInvalidRequestPayment  = errors.New("invalid request payment")
	ErrInvalidTopicName       = errors.New("invalid topic name")
	ErrNotExistRequestPayment = errors.New("not exist request payment")
	ErrExceedContentSize      = errors.New("exceed content size")
	ErrNotExistTopic          = errors.New("not exist topic")
	ErrExistSubscribe         = errors.New("exist subscribe")
)
