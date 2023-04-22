package schedule

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	// read/write from/to json values
	_ json.Marshaler   = (*Weekday)(nil)
	_ json.Unmarshaler = (*Weekday)(nil)

	// read/write from/to json keys
	_ encoding.TextMarshaler   = (*Weekday)(nil)
	_ encoding.TextUnmarshaler = (*Weekday)(nil)

	// read/write from/to sql
	_ sql.Scanner   = (*Weekday)(nil)
	_ driver.Valuer = (*Weekday)(nil)
)

// Weekday exists only because time.Weekday does not serialize to a string
// so models.Weekday will be a string when in JSON or database
type Weekday time.Weekday

const (
	Sunday    = Weekday(time.Sunday)
	Monday    = Weekday(time.Monday)
	Tuesday   = Weekday(time.Tuesday)
	Wednesday = Weekday(time.Wednesday)
	Thursday  = Weekday(time.Thursday)
	Friday    = Weekday(time.Friday)
	Saturday  = Weekday(time.Saturday)
)

// NewWeekday converts a string name to a Weekday object
// to get Weekday from time.Weekday simply type assert
// ex: models.Weekday(time.Monday)
func NewWeekday(value string) (Weekday, error) {
	return ParseWeekday(value)
}

var parseWeekdayMap = map[string]Weekday{
	strings.ToLower(Sunday.String()):    Sunday,
	strings.ToLower(Monday.String()):    Monday,
	strings.ToLower(Tuesday.String()):   Tuesday,
	strings.ToLower(Wednesday.String()): Wednesday,
	strings.ToLower(Thursday.String()):  Thursday,
	strings.ToLower(Friday.String()):    Friday,
	strings.ToLower(Saturday.String()):  Saturday,
}

func ParseWeekday(dayName string) (Weekday, error) {
	value := strings.ToLower(dayName)
	if d, ok := parseWeekdayMap[value]; ok {
		return d, nil
	}
	return 0, ErrInvalidDayName
}

func TodayWeekday() Weekday {
	return Today().Weekday()
}

func (w Weekday) String() string { return time.Weekday(w).String() }

func (w Weekday) Next() Weekday {
	if w == Saturday {
		return Sunday
	}
	return w + 1
}

// MarshalJSON marshals the enum as a quoted json string
func (w Weekday) MarshalJSON() ([]byte, error) {
	return []byte(`"` + w.String() + `"`), nil
}

func (w Weekday) MarshalText() (text []byte, err error) {
	return []byte(w.String()), nil
}

func (w *Weekday) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && string(b[0]) != `"` {
		// value is an int, not a string
		var v int
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*w = Weekday(v % 7) // allow large ints to roll over to next week, so 7 is 0 is Sunday
		return nil
	}
	return w.UnmarshalText(b)
}

func (w *Weekday) UnmarshalText(b []byte) error {
	var dayName string
	if err := json.Unmarshal(b, &dayName); err != nil {
		return err
	}

	d, err := NewWeekday(dayName)
	if err != nil {
		return err
	}

	*w = d
	return nil
}

// Value is used for sql exec to persist this type as a string
func (w Weekday) Value() (driver.Value, error) {
	return w.String(), nil
}

// Scan implements sql.Scanner so that Scan will be scanned correctly from storage
func (w *Weekday) Scan(src interface{}) error {
	switch t := src.(type) {
	case int:
		*w = Weekday(t)
	case int64:
		*w = Weekday(int(t))
	case string:
		d, err := NewWeekday(t)
		if err != nil {
			return err
		}
		*w = d
	case []byte:
		d, err := NewWeekday(string(t))
		if err != nil {
			return err
		}
		*w = d
	default:
		return errors.New("Weekday.Scan requires a string or byte array")
	}
	return nil
}
