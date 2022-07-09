package schedule

import (
	"errors"
	"math"
)

var (
	ErrFromRequired = errors.New("from is required")
	ErrPastUntil    = errors.New("until can not be before from")
)

// DateRange represents a set of days
// this set includes the From date and the Until date
// from Jan1 until Jan1 is one day
// from Jan1 until Jan2 is two days
// when Until is nil it means forever
type DateRange struct {
	From  Date  `json:"validFrom"`
	Until *Date `json:"validUntil,omitempty"`
}

func NewDateRange() DateRange {
	return DateRange{
		From: Today(),
	}
}
func NewDateRangeUntil(from Date, until *Date) DateRange {
	return DateRange{
		From:  from,
		Until: until,
	}
}

func ZeroDateRange() DateRange {
	return DateRange{
		From:  *ZeroDate(),
		Until: ZeroDate(),
	}
}
func (dr DateRange) WithFrom(d Date) DateRange {
	dr.From = d
	return dr
}
func (dr DateRange) WithUntil(d Date) DateRange {
	dr.Until = &d
	return dr
}

func (dr DateRange) Validate() error {
	if dr.From == *ZeroDate() {
		return ErrFromRequired
	}

	if dr.Until.IsZero() {
		return nil
	}

	if dr.Until.ToTime().Before(dr.From.ToTime()) {
		return ErrPastUntil
	}

	return nil
}

func (dr DateRange) IsZero() bool {
	return dr.From.IsZero() && dr.Until.IsZero()
}

func (dr DateRange) ContainsDate(date Date) bool {
	if dr.From.Before(date) || dr.From.Equal(date) {
		return dr.Until == nil || dr.Until.After(date) || dr.Until.Equal(date)
	}
	return false
}

func (dr DateRange) Overlaps(dr2 DateRange) bool {
	if dr.Equal(dr2) {
		return true
	}

	a, b := dr, dr2
	if a.Until == nil && b.Until == nil {
		return true
	}

	if a.Until == nil && b.Until != nil {
		return b.Until.After(a.From)
	}

	if b.Until == nil && a.Until != nil {
		return a.Until.After(b.From)
	}

	return b.ContainsDate(a.From) || b.ContainsDate(*a.Until) ||
		a.ContainsDate(b.From) || a.ContainsDate(*b.Until)
}

func (dr DateRange) Exceeds(parentDR DateRange) bool {
	if dr.From.Before(parentDR.From) {
		return true
	}
	if parentDR.Until == nil {
		return false
	}
	return dr.Until == nil || dr.Until.After(*parentDR.Until)
}

func (dr DateRange) Merge(dr2 DateRange) DateRange {
	if !dr.Overlaps(dr2) {
		return ZeroDateRange()
	}

	var (
		from  = *MaxDate(&dr.From, &dr2.From)
		until = MinDate(dr.Until, dr2.Until)
	)

	if until != nil && from.After(*until) {
		return ZeroDateRange()
	}

	return NewDateRangeUntil(from, until)
}

func (dr DateRange) Equal(dr2 DateRange) bool {
	if dr.Until == nil || dr2.Until == nil {
		if dr.Until != dr2.Until {
			return false
		}
		return dr.From.Equal(dr2.From)
	}

	return dr.From == dr2.From && *dr.Until == *dr2.Until
}

func (dr DateRange) String() string {
	var (
		from  = dr.From.String()
		until = "forever"
	)
	if !dr.Until.IsZero() {
		until = dr.Until.String()
	}
	return "from " + from + " until " + until
}

// InfDays represents an infinite number of days
// MaxInt32 days is over 5.8 million years
// we are not using a negative number because someone
// may check DayCount() > 0 to know if it has days or not
const InfDays = math.MaxInt32

// DayCount is the number of days in range
// when until is nil then the range is infinite
// from today until today has one day in range
// from today until tomorrow has two days in range
func (dr DateRange) DayCount() int {
	// DateRange isn't set, so it has no days in range
	if dr.IsZero() {
		return 0
	}

	if dr.Until == nil {
		return InfDays
	}

	if n := dr.Until.Sub(dr.From); n >= 0 {
		return n + 1
	}

	// if we are here then the result of Sub() was negative
	// which happens when from > until, so no days in range.
	return 0
}

func (dr DateRange) HasDays() bool {
	return dr.DayCount() > 0
}
