package schedule

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type WeekdayTimeSlotMap map[Weekday][]TimeSlot

func NewWeekdayTimeSlotMap() WeekdayTimeSlotMap {
	return make(WeekdayTimeSlotMap)
}

func (w WeekdayTimeSlotMap) Add(day Weekday, slots ...TimeSlot) WeekdayTimeSlotMap {
	if w == nil { // anti-panic
		return NewWeekdayTimeSlotMap().Add(day, slots...)
	}
	if w[day] == nil {
		w[day] = make([]TimeSlot, 0)
	}
	for _, slot := range slots {
		if w.Has(day, slot) {
			continue
		}
		w[day] = append(w[day], slot)
	}
	return w
}
func (w WeekdayTimeSlotMap) Has(day Weekday, slot TimeSlot) bool {
	for _, s := range w[day] {
		if s == slot {
			return true
		}
	}
	return false
}

func (w WeekdayTimeSlotMap) AddTimeSlot(day Weekday, start, end Clock) WeekdayTimeSlotMap {
	return w.Add(day, NewTimeSlot(start, end))
}

func (w WeekdayTimeSlotMap) TimeSlots(day Weekday) []TimeSlot {
	return w[day]
}

func (w WeekdayTimeSlotMap) ToWeekdayTimeSlots() []WeekdayTimeSlot {
	wtsSlice := make([]WeekdayTimeSlot, 0)
	for day, slots := range w {
		if len(slots) == 0 {
			wtsSlice = append(wtsSlice, NewWeekdayTimeSlot(day, TimeSlot{}))
			continue
		}
		for _, slot := range slots {
			wtsSlice = append(wtsSlice, NewWeekdayTimeSlot(day, slot))
		}
	}
	return SortWeekdayTimeSlots(wtsSlice...)
}

func WeekdayTimeSlotMapFromSlice(wtsSlice []WeekdayTimeSlot) WeekdayTimeSlotMap {
	wtsMap := make(WeekdayTimeSlotMap)
	for _, wts := range wtsSlice {
		if wts.slot.IsZero() {
			wtsMap.Add(wts.day)
			continue
		}
		wtsMap.Add(wts.day, wts.slot)
	}
	return wtsMap
}

type TimeSlot struct {
	Start Clock `json:"start"`
	End   Clock `json:"end"`
}

func NewTimeSlot(start, end Clock) TimeSlot {
	return TimeSlot{
		Start: start,
		End:   end,
	}
}

func ParseTimeSlot(tsString string) TimeSlot {
	parts := strings.Split(tsString, "-")
	if len(parts) != 2 {
		return TimeSlot{}
	}
	return NewTimeSlot(ParseClock(parts[0]), ParseClock(parts[1]))
}

func (ts TimeSlot) StartTime() Clock { return ts.Start }
func (ts TimeSlot) EndTime() Clock   { return ts.End }
func (ts TimeSlot) String() string {
	return ts.StartTime().String() + "-" + ts.EndTime().String()
}
func (ts TimeSlot) Minutes() int {
	return ts.EndTime().min - ts.StartTime().min
}
func (ts TimeSlot) Duration() time.Duration {
	return time.Duration(ts.Minutes()) * time.Minute
}

// IsZero returns true only when the start and time are both 00:00
func (ts TimeSlot) IsZero() bool {
	return ts.Start.min+ts.End.min == 0
}

type WeekdayTimeSlot struct {
	day  Weekday
	slot TimeSlot
}

func NewWeekdayTimeSlot(day Weekday, slot TimeSlot) WeekdayTimeSlot {
	return WeekdayTimeSlot{day: day, slot: slot}
}

func NewWeekdayAllDayTimeSlot(day Weekday) WeekdayTimeSlot {
	slot := TimeSlot{
		Start: Clock{},
		End:   Clock{},
	}
	return WeekdayTimeSlot{day: day, slot: slot}
}

func WeekdayTimeSlotFromString(wtsString string) WeekdayTimeSlot {
	var wts WeekdayTimeSlot

	if wtsString == "" {
		return wts
	}

	// Monday 01:30-02:15
	if parts := strings.Split(wtsString, " "); len(parts) == 2 {
		wts.slot = ParseTimeSlot(parts[1])
		day, _ := NewWeekday(parts[0])
		wts.day = day
		return wts
	}

	// Monday
	if day, err := NewWeekday(wtsString); err == nil {
		wts.day = day
		return wts
	}

	// 01:30-02:15
	wts.slot = ParseTimeSlot(wtsString)
	return wts
}

func (s WeekdayTimeSlot) Slot() TimeSlot { return s.slot }
func (s WeekdayTimeSlot) String() string { return s.ToString() }
func (s WeekdayTimeSlot) ToString() string {
	return fmt.Sprintf("%s %s-%s",
		s.day.String(),
		s.Start().String(),
		s.End().String())
}

func WeekdayTimeSlotFromInt(wtsInt int) WeekdayTimeSlot {
	var (
		dayInt    = wtsInt >> 22                   // bits 23-25
		startMins = (wtsInt >> 11) & 0b11111111111 // bits 12-22
		endMins   = wtsInt & 0b11111111111         // bits 1-11
		start     = NewClock(0, startMins)
		end       = NewClock(0, endMins)
	)
	return WeekdayTimeSlot{
		day:  Weekday(dayInt),
		slot: NewTimeSlot(start, end),
	}
}

// ToInt stores the object in binary
// 3 bits for day, 11 bits for start, 11 bits for end
// this could be used to check equality or sorting
func (s WeekdayTimeSlot) ToInt() int {
	dayInt := int(s.day) << 22
	startInt := s.Start().min << 11
	return dayInt + startInt + s.End().min
}

func (s WeekdayTimeSlot) Weekday() Weekday        { return s.day }
func (s WeekdayTimeSlot) Start() Clock            { return s.slot.StartTime() }
func (s WeekdayTimeSlot) End() Clock              { return s.slot.EndTime() }
func (s WeekdayTimeSlot) Minutes() int            { return s.slot.Minutes() }
func (s WeekdayTimeSlot) Duration() time.Duration { return s.slot.Duration() }
func (s WeekdayTimeSlot) IsAllDay() bool {
	return s.slot.IsZero()
}

func (s WeekdayTimeSlot) OverlapsWith(wts2 WeekdayTimeSlot) bool {
	slots := SortWeekdayTimeSlots(s, wts2)
	s0, s1 := slots[0], slots[1]

	if s0.Weekday() == s1.Weekday() {
		return s1.Start().Before(s0.End()) || s0.crossesMidnight()
	}

	if s0.crossesMidnight() && s1.Weekday() == s0.Weekday().Next() {
		// example: Monday 23:30-00:30, Tuesday 00:15-01:15
		return s1.Start().Before(s0.End())
	}

	if s1.Weekday() == Saturday && s0.Weekday() == Sunday {
		// example: Sunday 00:15-01:15, Saturday 23:30-00:30
		return s0.Start().Before(s1.End()) && s1.crossesMidnight()
	}
	return false
}

func (s WeekdayTimeSlot) crossesMidnight() bool { return s.End().Before(s.Start()) }

func (s WeekdayTimeSlot) Equal(s2 WeekdayTimeSlot) bool {
	return s.ToInt() == s2.ToInt()
}

func SortWeekdayTimeSlots(wtsSlice ...WeekdayTimeSlot) []WeekdayTimeSlot {
	result := append([]WeekdayTimeSlot{}, wtsSlice...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ToInt() < result[j].ToInt()
	})
	return result
}

// UniqueWeekdayTimeSlots sorts and removes duplicates
func UniqueWeekdayTimeSlots(wtsSlice ...WeekdayTimeSlot) []WeekdayTimeSlot {
	if len(wtsSlice) == 0 {
		return wtsSlice
	}
	return WeekdayTimeSlotMapFromSlice(wtsSlice).ToWeekdayTimeSlots()
}

func SlotKeys(slots ...WeekdayTimeSlot) []int {
	var keys = make([]int, len(slots))
	for i, slot := range slots {
		keys[i] = slot.ToInt()
	}
	return keys
}

func MergeWeekdayTimeSlots(a, b []WeekdayTimeSlot) []WeekdayTimeSlot {
	var ret = make([]WeekdayTimeSlot, 0)
	a = UniqueWeekdayTimeSlots(a...)
	b = UniqueWeekdayTimeSlots(b...)
	for _, aSlot := range a {
		aSlotInt := aSlot.ToInt()
		for _, bSlot := range b {
			if aSlot.IsAllDay() && !bSlot.IsAllDay() && aSlot.Weekday() == bSlot.Weekday() {
				ret = append(ret, bSlot)
				continue
			}

			if bSlot.IsAllDay() && !aSlot.IsAllDay() && aSlot.Weekday() == bSlot.Weekday() {
				ret = append(ret, aSlot)
				continue
			}

			if aSlotInt == bSlot.ToInt() {
				ret = append(ret, bSlot)
			}
		}
	}

	return ret
}
