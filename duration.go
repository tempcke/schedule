package schedule

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Duration is an alias for time.Duration, but with custom JSON behavior
type Duration time.Duration

func NewDuration(d time.Duration) Duration {
	return Duration(d)
}

func ParseDuration(s string) (Duration, error) {
	d, err := time.ParseDuration(s)
	return Duration(d), err
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Convenience helpers for parity with Clock
func (d Duration) IsZero() bool { return time.Duration(d) == 0 }
func (d Duration) Pointer() *Duration {
	v := d
	return &v
}

// MarshalJSON makes Duration encode as a string (e.g. "1h30m")
// instead of a number of nanoseconds.
func (d Duration) MarshalJSON() ([]byte, error)          { return []byte(`"` + d.String() + `"`), nil }
func (d Duration) MarshalText() (text []byte, err error) { return []byte(d.String()), nil }
func (d Duration) Value() (driver.Value, error)          { return d.String(), nil }

// UnmarshalJSON makes Duration decode from a string (e.g. "1h30m")
// instead of a number of nanoseconds.
func (d *Duration) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		// leave zero
		return nil
	}

	// If first byte is not a quote, try numeric nanoseconds (int64)
	if b[0] != '"' {
		var v time.Duration
		if err := json.Unmarshal(b, &v); err != nil {
			return fmt.Errorf("Duration.UnmarshalJSON expected int but got %v: %w", string(b), err)
		}
		*d = Duration(v)
		return nil
	}

	// Quoted string duration
	return d.UnmarshalText(b)
}

// UnmarshalText should accept plain text (e.g. map key) or quoted JSON string.
// It treats empty string as zero.
func (d *Duration) UnmarshalText(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		// leave zero
		return nil
	}

	var s = strings.Trim(string(b), `" `)

	if s == "" {
		// leave zero
		return nil
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}

	*d = Duration(parsed)
	return nil
}
func (d *Duration) Scan(src any) error {
	switch t := src.(type) {
	case nil:
		// leave zero
		return nil
	case int:
		*d = Duration(time.Duration(t))
	case int64:
		*d = Duration(time.Duration(t))
	case string:
		v, err := ParseDuration(t)
		if err != nil {
			return err
		}
		*d = v
	case []byte:
		v, err := ParseDuration(string(t))
		if err != nil {
			return err
		}
		*d = v
	default:
		return fmt.Errorf("Duration.Scan requires int/int64 nanoseconds, string, or byte array; got %T", src)
	}
	return nil
}
