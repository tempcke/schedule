package schedule

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type weekdayTimeSlot struct {
	Weekday  Weekday  `json:"weekday"`
	TimeSlot TimeSlot `json:"timeSlot"`
}

func (s WeekdayTimeSlot) MarshalJSON() ([]byte, error) {
	return json.Marshal(weekdayTimeSlot{
		Weekday:  s.Weekday(),
		TimeSlot: s.Slot(),
	})
}
func (s *WeekdayTimeSlot) UnmarshalJSON(data []byte) error {
	var v weekdayTimeSlot
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*s = NewWeekdayTimeSlot(v.Weekday, v.TimeSlot)
	return nil
}
func (s *WeekdayTimeSlot) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch t := src.(type) {
	case int:
		*s = WeekdayTimeSlotFromInt(t)
	case int64:
		*s = WeekdayTimeSlotFromInt(int(t))
	case string:
		*s = WeekdayTimeSlotFromString(t)
	case []byte:
		*s = WeekdayTimeSlotFromString(string(t))
	default:
		return fmt.Errorf("scan requires an int, string or byte slice but got: %T %v", src, src)
	}
	return nil
}
func (s WeekdayTimeSlot) Value() (driver.Value, error) {
	return s.String(), nil
}
