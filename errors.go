package schedule

import (
	"errors"
)

var (
	ErrFromRequired      = errors.New("from is required")
	ErrPastUntil         = errors.New("until can not be before from")
	ErrInvalidDayName    = errors.New("invalid day name")
	ErrInvalidDateString = errors.New("can not parse date, must use yyyy-mm-dd format")
)
