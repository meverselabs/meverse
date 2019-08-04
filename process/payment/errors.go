package payment

import "errors"

// consensus errors
var (
	ErrInvalidRequestPayment  = errors.New("invalid request payment")
	ErrInvalidPaymentAmount   = errors.New("invalid payment amount")
	ErrInvalidTopicName       = errors.New("invalid topic name")
	ErrNotExistRequestPayment = errors.New("not exist request payment")
	ErrExceedContentSize      = errors.New("exceed content size")
	ErrExistTopic             = errors.New("exist topic")
	ErrNotExistTopic          = errors.New("not exist topic")
	ErrExistSubscribe         = errors.New("exist subscribe")
	ErrNotExistSubscribe      = errors.New("not exist subscribe")
	ErrInvalidBillingAmount   = errors.New("invalid billing amount")
)
