package schedule

type Schedule struct {
	DateRange DateRange
	TimeSlots []WeekdayTimeSlot
}

func NewSchedule(dr DateRange, slots ...WeekdayTimeSlot) Schedule {
	return Schedule{
		DateRange: dr,
		TimeSlots: slots,
	}
}

func (s Schedule) WithDateRange(dateRange DateRange) Schedule {
	s.DateRange = dateRange
	return s
}

func (s Schedule) WithFrom(from Date) Schedule {
	s.DateRange = s.DateRange.WithFrom(from)
	return s
}

func (s Schedule) WithUntil(until Date) Schedule {
	s.DateRange = s.DateRange.WithUntil(until)
	return s
}

func (s Schedule) WithTimeSlots(slots ...WeekdayTimeSlot) Schedule {
	s.TimeSlots = append(s.TimeSlots, slots...)
	return s
}

func (s Schedule) From() Date   { return s.DateRange.From }
func (s Schedule) Until() *Date { return s.DateRange.Until }

// IsEmpty means there are either no days in range
// or there are no timeslots, therefore nothing on schedule
func (s Schedule) IsEmpty() bool {
	return !s.DateRange.HasDays() || !s.HasTimeSlots()
}

func (s Schedule) HasTimeSlots() bool {
	return len(s.TimeSlots) > 0
}

// Merge does a merge on both the schedule dateRanges and the timeslots
//   The intended use for this is to merge a parent schedule with a sub schedule
//   the sub schedule is a subset of the parent, however it may have all-day entries
//   which allow for config on "any" timeslots for that day
//   if timeslots exist on only the parent, they are excluded
//   if timeslots exist on only the child, they are invalid, and therefore excluded
// dateRange
//   a merged dateRange is where they intersect
//   - ex: jan01-jan30 merged with jan15-feb15 results in jan15-jan30
// timeslots
//   are only kept in a merge when they exist in all schedules without conflicting
//   exception is when one schedule has an all day timeslot and another specific timeslots
//   this results in the specific timeslots being kept in favor of the all day
//   - ex: Mon7-8,Tues8-9,Wed merged with Mon7-8,Tues6-7,Wed7-8 results in Mon7-8,Wed7-8
//         Tues809,Tues6-7 were excluded because they only exist in one schedule
//         Wed7-8 was included because the other schedule was for Wed all day
// see TestSchedulesMerge for a good example
func (s Schedule) Merge(schedules ...Schedule) Schedule {
	if len(schedules) == 0 {
		return s
	}
	ret := NewSchedule(s.DateRange, s.TimeSlots...)

	for _, schedule := range schedules {
		ret.DateRange = ret.DateRange.Merge(schedule.DateRange)
		if ret.DateRange.IsZero() {
			ret.TimeSlots = make([]WeekdayTimeSlot, 0)
			return ret
		}

		ret.TimeSlots = MergeWeekdayTimeSlots(ret.TimeSlots, schedule.TimeSlots)
		if len(ret.TimeSlots) == 0 {
			return ret
		}
	}

	return ret
}

type CalendarMap map[Date][]WeekdayTimeSlot

func (cm CalendarMap) HasDate(date Date) bool {
	_, ok := cm[date]
	return ok
}

func (cm CalendarMap) GetTimeslots() []WeekdayTimeSlot {
	var calendarTimeslots []WeekdayTimeSlot
	for _, timeslots := range cm {
		calendarTimeslots = append(calendarTimeslots, timeslots...)
	}
	return calendarTimeslots
}

type Calendar struct {
	schedules []Schedule
}

func NewCalendar(schedules ...Schedule) Calendar {
	var c Calendar
	return c.WithSchedules(schedules...)
}

// WithSchedules appends schedules to existing Calendar
func (c Calendar) WithSchedules(schedules ...Schedule) Calendar {
	c.schedules = append(c.schedules, schedules...)
	return c
}

func (c Calendar) ByDate(limit Date) CalendarMap {
	var cm = make(CalendarMap)

	for _, s := range c.schedules {
		var dr = s.DateRange
		if dr.Until == nil || dr.Until.After(limit) {
			dr.Until = &limit
		}

		for date := dr.From; !date.After(*dr.Until); date = date.Next() {
			if cm[date] == nil {
				cm[date] = make([]WeekdayTimeSlot, 0, len(s.TimeSlots))
			}
			for _, slot := range s.TimeSlots {
				if slot.Weekday() == date.Weekday() {
					cm[date] = append(cm[date], slot)
				}
			}
		}
	}

	// ensure
	for date, slots := range cm {
		cm[date] = UniqueWeekdayTimeSlots(slots...)
	}

	return cm
}
