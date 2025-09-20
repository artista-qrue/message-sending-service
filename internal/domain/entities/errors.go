package entities

import "errors"

var (
	ErrInvalidMessageContent   = errors.New("message content cannot be empty")
	ErrMessageTooLong          = errors.New("message content exceeds 160 characters")
	ErrInvalidPhoneNumber      = errors.New("phone number cannot be empty")
	ErrMessageNotFound         = errors.New("message not found")
	ErrSchedulerNotRunning     = errors.New("scheduler is not running")
	ErrSchedulerAlreadyRunning = errors.New("scheduler is already running")
)
