package recurrence

import "errors"

var ErrDateOutOfRange = errors.New("date out of recurrence range")

var ErrDateNotMatched = errors.New("date does not match recurrence schedule")
