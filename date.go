package schedule

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

const ymdFormat = "2006-01-02"

var _ json.Marshaler = (*Date)(nil)
var _ json.Unmarshaler = (*Date)(nil)
var _ encoding.TextMarshaler = (*Date)(nil)
var _ encoding.TextUnmarshaler = (*Date)(nil)
var _ sql.Scanner = (*Date)(nil)
var _ driver.Valuer = (*Date)(nil)

type Date struct {
	year  int
	month time.Month
	day   int
}

// ZeroDate is just a zero value Date
// it is good for json decoding and sql scanning
func ZeroDate() *Date {
	return &Date{}
}

func Today() Date {
	t := time.Now()
	return NewDate(t.Year(), t.Month(), t.Day())
}

func NewDate(year int, month time.Month, day int) Date {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Date{
		year:  t.Year(),
		month: t.Month(),
		day:   t.Day(),
	}
}

func NewDateFromTime(t time.Time) Date {
	return NewDate(t.Year(), t.Month(), t.Day())
}

func ParseDate(s string) *Date {
	if len(s) > len(ymdFormat) {
		s = s[0:len(ymdFormat)] // only keep the first bit in case string includes time info
	}
	t, err := time.Parse(ymdFormat, s)
	if err != nil {
		return nil
	}
	d := NewDate(t.Year(), t.Month(), t.Day())
	return &d
}

func newDateFromTime(t time.Time) Date {
	return Date{
		year:  t.Year(),
		month: t.Month(),
		day:   t.Day(),
	}
}

func (d Date) String() string        { return d.ToTime().Format(ymdFormat) }
func (d Date) Year() int             { return d.year }
func (d Date) Month() time.Month     { return d.month }
func (d Date) Day() int              { return d.day }
func (d Date) Weekday() Weekday      { return Weekday(d.ToTime().Weekday()) }
func (d Date) Before(date Date) bool { return d.ToTime().Before(date.ToTime()) }
func (d Date) After(date Date) bool  { return d.ToTime().After(date.ToTime()) }
func (d Date) Equal(date Date) bool  { return d.ToTime().Equal(date.ToTime()) }
func (d Date) Next() Date            { return d.AddDate(0, 0, 1) }
func (d Date) Pointer() *Date        { return &d }
func (d *Date) IsZero() bool         { return d == nil || *d == Date{} }

// Sub subtracts two dates, returning the number of days between
//
//	today.Sub(today)     = 0
//	today.Sub(yesterday) = 1
//	yesterday.Sub(today) = -1
func (d Date) Sub(date Date) int {
	return int(math.Round(d.ToTime().Sub(date.ToTime()).Hours() / 24))
}

func (d Date) ToTime() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
}

func (d *Date) Date() *Date {
	if d.IsZero() {
		return nil
	}
	return d
}

func (d Date) AddDate(year, month, day int) Date {
	return NewDate(d.Year()+year, d.Month()+time.Month(month), d.Day()+day)
}

func (d Date) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	*d = *ParseDate(string(text))
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(d.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (d *Date) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	if s == "" {
		return nil
	}

	if len(s) < len(ymdFormat) {
		return fmt.Errorf(`%w: %s`, ErrInvalidDateString, s)
	}

	s = s[0:len(ymdFormat)] // only keep the first bit in case string includes time info
	t, err := time.Parse(ymdFormat, s)
	if err != nil {
		return fmt.Errorf(`%w: %s`, ErrInvalidDateString, s)
	}
	*d = NewDate(t.Year(), t.Month(), t.Day())

	return nil
}

func (d *Date) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch t := src.(type) {
	case int:
		dt := time.Unix(int64(t), 0)
		*d = newDateFromTime(dt)
	case int64:
		dt := time.Unix(t, 0)
		*d = newDateFromTime(dt)
	case time.Time:
		*d = newDateFromTime(t)
	case string:
		dt, err := time.Parse(ymdFormat, t)
		if err != nil {
			return err
		}
		*d = newDateFromTime(dt)
	case []byte:
		dt, err := time.Parse(ymdFormat, string(t))
		if err != nil {
			return err
		}
		*d = newDateFromTime(dt)
	default:
		return fmt.Errorf("Date.Scan requires a string or byte array in yyyy-mm-dd format got %T %v", src, src)
	}
	return nil
}

func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.String(), nil
}

func MinDate(a, b *Date) *Date {
	if a == nil && b == nil {
		return nil
	}

	if a != nil && (b == nil || a.Before(*b)) {
		return a
	}

	return b
}

func MaxDate(a, b *Date) *Date {
	if a == nil && b == nil {
		return nil
	}

	if a != nil && (b == nil || a.After(*b)) {
		return a
	}

	return b

}
