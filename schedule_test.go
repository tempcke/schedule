package schedule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tempcke/schedule"
)

func TestSchedule(t *testing.T) {
	var (
		today     = schedule.Today()
		yesterday = today.AddDate(0, 0, -1)
		thursday  = schedule.Thursday
		eight     = schedule.ParseClock("08:00")
		ten       = schedule.ParseClock("10:00")

		slot = schedule.WeekdayTimeSlotFromString("Monday 07:00-08:00")
	)
	t.Run("anti panic", func(t *testing.T) {
		new(schedule.WeekdayTimeSlotMap).AddTimeSlot(thursday, eight, ten)
		new(schedule.WeekdayTimeSlotMap).TimeSlots(thursday)
		schedule.ParseTimeSlot("invalid string")
		schedule.ParseClock("invalid string")
	})

	t.Run("IsEmpty", func(t *testing.T) {
		var s schedule.Schedule
		isEmpty := []schedule.Schedule{
			s,                            // no  dr no  slots
			s.WithTimeSlots(slot),        // no  dr has slots
			s.WithFrom(schedule.Today()), // has dr no  slots
			s.WithFrom(schedule.Today()), // has dr has slots no dates in range
			s.WithDateRange(schedule.ZeroDateRange()).
				WithTimeSlots(slot), // zero dr has slots
			s.WithFrom(today).WithUntil(yesterday).
				WithTimeSlots(slot), // has dr has slots but no days in range
		}
		for i, sch := range isEmpty {
			assert.True(t, sch.IsEmpty(), i)
		}
	})
}

func TestScheduleConstructors(t *testing.T) {
	t.Run("new Schedule", func(t *testing.T) {
		dateRange := schedule.NewDateRangeUntil(schedule.Today(), nil)
		sch := schedule.NewSchedule(dateRange)
		assert.NotNil(t, sch)
		assert.Equal(t, dateRange, sch.DateRange)
	})

	t.Run("new Calendar", func(t *testing.T) {
		dateRange := schedule.NewDateRangeUntil(schedule.Today(), nil)
		sch := schedule.NewSchedule(dateRange)
		assert.NotNil(t, sch)
		assert.Equal(t, dateRange, sch.DateRange)
		calendar := schedule.NewCalendar(sch)
		assert.NotNil(t, calendar)
	})
}

func TestSchedulesMerge(t *testing.T) {
	wtsA := []schedule.WeekdayTimeSlot{
		schedule.WeekdayTimeSlotFromString("Sunday"),                // all day with nothing in B
		schedule.WeekdayTimeSlotFromString("Monday"),                // all day with slots in B
		schedule.WeekdayTimeSlotFromString("Tuesday 07:00-08:00"),   // not in B
		schedule.WeekdayTimeSlotFromString("Tuesday 08:00-09:00"),   // in both
		schedule.WeekdayTimeSlotFromString("Tuesday 09:00-09:45"),   // does not match
		schedule.WeekdayTimeSlotFromString("Tuesday 10:15-11:00"),   // does not match
		schedule.WeekdayTimeSlotFromString("Wednesday"),             // invalid, should not be here
		schedule.WeekdayTimeSlotFromString("Wednesday 07:00-08:00"), // in both
		schedule.WeekdayTimeSlotFromString("Thursday 07:00-08:00"),  // B is all day
		schedule.WeekdayTimeSlotFromString("Saturday"),              // both A and B are all day
	}

	wtsB := []schedule.WeekdayTimeSlot{
		schedule.WeekdayTimeSlotFromString("Monday 07:00-08:00"),    // A is all day
		schedule.WeekdayTimeSlotFromString("Tuesday 08:00-09:00"),   // in both
		schedule.WeekdayTimeSlotFromString("Tuesday 09:00-09:30"),   // does not match
		schedule.WeekdayTimeSlotFromString("Tuesday 10:30-11:00"),   // does not match
		schedule.WeekdayTimeSlotFromString("Tuesday 11:00-12:00"),   // not in A
		schedule.WeekdayTimeSlotFromString("Wednesday 07:00-08:00"), // in both
		schedule.WeekdayTimeSlotFromString("Wednesday 08:00-09:00"), // not in A
		schedule.WeekdayTimeSlotFromString("Thursday"),              // all day with slots in B
		schedule.WeekdayTimeSlotFromString("Friday"),                // all day with nothing in B
		schedule.WeekdayTimeSlotFromString("Saturday"),              // both A and B are all day
	}

	wtsC := []schedule.WeekdayTimeSlot{
		schedule.WeekdayTimeSlotFromString("Monday 07:00-08:00"),    // A is all day
		schedule.WeekdayTimeSlotFromString("Tuesday 08:00-09:00"),   // in both
		schedule.WeekdayTimeSlotFromString("Wednesday 07:00-08:00"), // in both
		schedule.WeekdayTimeSlotFromString("Thursday 07:00-08:00"),  // B is all day
		schedule.WeekdayTimeSlotFromString("Saturday"),              // both A and B are all day
	}

	t.Run("merge WeekdayTimeSlots", func(t *testing.T) {
		merged := schedule.MergeWeekdayTimeSlots(wtsA, wtsB)
		assert.Len(t, merged, len(wtsC))

		for _, wts := range wtsC {
			assert.Contains(t, merged, wts, wts.String()+" not found")
		}
	})

	t.Run("merge Schedules", func(t *testing.T) {
		jan1 := schedule.NewDate(2020, 1, 1)
		jan15 := schedule.NewDate(2020, 1, 15)
		jan30 := schedule.NewDate(2020, 1, 30)

		scheduleA := schedule.NewSchedule(schedule.NewDateRangeUntil(jan1, &jan30), wtsA...)
		scheduleB := schedule.NewSchedule(schedule.NewDateRangeUntil(jan15, nil), wtsB...)

		merged := scheduleA.Merge(scheduleB)
		assert.True(t, schedule.NewDateRangeUntil(jan15, &jan30).Equal(merged.DateRange))
		assert.Len(t, merged.TimeSlots, len(wtsC))
		// can be a loop but here's easier to verify with one is missing
		assert.Contains(t, merged.TimeSlots, wtsC[0])
		assert.Contains(t, merged.TimeSlots, wtsC[1])
		assert.Contains(t, merged.TimeSlots, wtsC[2])
		assert.Contains(t, merged.TimeSlots, wtsC[3])
		assert.Contains(t, merged.TimeSlots, wtsC[4])
	})
}

func TestCalendar(t *testing.T) {
	t.Run("byDate", func(t *testing.T) {
		var (
			day0 = schedule.Today()
			day1 = day0.Next()
			day2 = day1.Next()

			slot6 = schedule.ParseTimeSlot("06:00-07:00")
			slot7 = schedule.ParseTimeSlot("07:00-08:00")
			slot8 = schedule.ParseTimeSlot("08:00-09:00")

			d0s6 = schedule.NewWeekdayTimeSlot(day0.Weekday(), slot6)
			d0s7 = schedule.NewWeekdayTimeSlot(day0.Weekday(), slot7)

			d1s6 = schedule.NewWeekdayTimeSlot(day1.Weekday(), slot6)
			d1s7 = schedule.NewWeekdayTimeSlot(day1.Weekday(), slot7)
			d1s8 = schedule.NewWeekdayTimeSlot(day1.Weekday(), slot8)

			d2s7 = schedule.NewWeekdayTimeSlot(day2.Weekday(), slot7)
			d2s8 = schedule.NewWeekdayTimeSlot(day2.Weekday(), slot8)
		)

		// day0 6-7,7-8,___; day1 6-7,7-8,___
		d0d1s6s7 := schedule.NewSchedule(schedule.NewDateRangeUntil(day0, &day1), d0s6, d0s7, d1s6, d1s7)
		// day1 ___,7-8,8-9; day2 ___,7-8,8-9
		d0d1s7s8 := schedule.NewSchedule(schedule.NewDateRangeUntil(day1, &day2), d1s7, d1s8, d2s7, d2s8)

		byDate := schedule.NewCalendar(d0d1s6s7, d0d1s7s8).
			ByDate(day2)

		// day0 6-7,7-8,___; day1 6-7,7-8,8-9; day2 ___,7-8,8-9
		assert.Contains(t, byDate, day0)
		assert.Contains(t, byDate, day1)
		assert.Contains(t, byDate, day2)
		assert.Equal(t, 2, len(byDate[day0]))
		assert.Contains(t, byDate[day0], d0s6)
		assert.Contains(t, byDate[day0], d0s7)
		assert.Equal(t, 3, len(byDate[day1]))
		assert.Contains(t, byDate[day1], d1s6)
		assert.Contains(t, byDate[day1], d1s7)
		assert.Contains(t, byDate[day1], d1s8)
		assert.Equal(t, 2, len(byDate[day2]))
		assert.Contains(t, byDate[day2], d2s7)
		assert.Contains(t, byDate[day2], d2s8)
	})
}
