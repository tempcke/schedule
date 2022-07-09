package schedule

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Clock contains minute in a day
type Clock struct {
	min int
}

// NewClock Clock
func NewClock(hour, minute int) Clock {
	m := minute + hour*60
	m %= 24 * 60
	if m < 0 {
		m += 24 * 60
	}
	return Clock{m}
}

// ParseClock takes a string of hours:minutes such as "15:30"
func ParseClock(clockStr string) Clock {
	parts := strings.Split(clockStr, ":")
	if len(parts) < 2 {
		return Clock{}
	}
	h, _ := strconv.Atoi(parts[0]) // h is 0 on error, so not concerned
	m, _ := strconv.Atoi(parts[1])
	return NewClock(h, m)
}

// String of a Clock hh:mm
func (c Clock) String() string {
	return fmt.Sprintf("%02d:%02d", c.Hour(), c.Minute())
}

// Add minutes to a Clock
func (c Clock) Add(minutes int) Clock {
	return NewClock(0, c.min+minutes)
}

// Subtract minutes from a Clock
func (c Clock) Subtract(minutes int) Clock {
	return NewClock(0, c.min-minutes)
}

func (c Clock) Hour() int            { return c.min / 60 }
func (c Clock) Minute() int          { return c.min % 60 }
func (c Clock) Second() int          { return 0 }
func (c Clock) Nanosecond() int      { return 0 }
func (c Clock) Equal(c2 Clock) bool  { return c.min == c2.min }
func (c Clock) Before(c2 Clock) bool { return c.min < c2.min }
func (c Clock) After(c2 Clock) bool  { return c.min > c2.min }
func (c Clock) IsZero() bool         { return c.min == 0 }
func (c Clock) Pointer() *Clock      { return &c }

func (c Clock) ToDuration() time.Duration {
	return time.Duration(c.min) * time.Minute
}

// MarshalJSON marshals the enum as a quoted json string
func (c Clock) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		buffer := bytes.NewBufferString(`"`)
		buffer.WriteString(`"`)
		return buffer.Bytes(), nil
	}
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(c.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string
func (c *Clock) UnmarshalJSON(b []byte) error {
	var hmStr string
	if err := json.Unmarshal(b, &hmStr); err != nil {
		return err
	}

	*c = ParseClock(hmStr)
	return nil
}

// Value implements driver.Valuer which is used parsing sql param values
func (c Clock) Value() (driver.Value, error) {
	return c.String(), nil
}

// Scan implements sql.Scanner so that Scan will be scanned correctly from storage
func (c *Clock) Scan(src interface{}) error {
	switch t := src.(type) {
	case int:
		*c = NewClock(0, t)
	case int64:
		*c = NewClock(0, int(t))
	case string:
		*c = ParseClock(t)
	case []byte:
		*c = ParseClock(string(t))
	default:
		return errors.New("Clock.Scan requires a string or byte array")
	}
	return nil
}

func (c Clock) ToTime(date Date, loc *time.Location) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		c.Hour(), c.Minute(), c.Second(), c.Nanosecond(),
		loc)
}
