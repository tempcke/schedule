package schedule_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

var (
	monday   = schedule.Monday
	tuesday  = schedule.Tuesday
	friday   = schedule.Friday
	saturday = schedule.Saturday

	t08 = schedule.NewClock(8, 0)
	t10 = schedule.NewClock(10, 0)
	t15 = schedule.NewClock(15, 0)
	t17 = schedule.NewClock(17, 0)
)

func TestTimeSlot(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		ts := schedule.NewTimeSlot(t08, t10)
		require.Equal(t, t08, ts.Start)
		require.Equal(t, t10, ts.End)
		require.Equal(t, t08, ts.StartTime())
		require.Equal(t, t10, ts.EndTime())
	})

	t.Run("to and from string", func(t *testing.T) {
		ts := schedule.NewTimeSlot(t08, t10)
		tsString := "08:00-10:00"

		// to string
		require.Equal(t, tsString, ts.String())

		// from string
		require.Equal(t, ts, schedule.ParseTimeSlot(tsString))
	})
}

func TestWeekdayTimeSlotMap(t *testing.T) {
	var (
		ts08 = schedule.NewTimeSlot(t08, t10)
		ts15 = schedule.NewTimeSlot(t15, t17)
	)

	t.Run("Add", func(t *testing.T) {
		wts := schedule.NewWeekdayTimeSlotMap().
			Add(monday, ts08, ts15).
			Add(friday)

		require.Len(t, wts.TimeSlots(monday), 2)

		// friday added without times so should be non nil empty set
		require.NotNil(t, wts.TimeSlots(friday))
		require.Len(t, wts.TimeSlots(friday), 0)

		// saturday was not added, so must be nil
		require.Nil(t, wts.TimeSlots(saturday))
	})

	t.Run("Add no duplicates", func(t *testing.T) {
		wts := schedule.NewWeekdayTimeSlotMap().
			Add(monday, ts08, ts15, ts08).
			Add(friday)

		require.Len(t, wts.TimeSlots(monday), 2)

		// friday added without times so should be non nil empty set
		require.NotNil(t, wts.TimeSlots(friday))
		require.Len(t, wts.TimeSlots(friday), 0)

		// saturday was not added, so must be nil
		require.Nil(t, wts.TimeSlots(saturday))
	})

	t.Run("AddTimeSlot", func(t *testing.T) {
		wts := schedule.NewWeekdayTimeSlotMap().
			AddTimeSlot(monday, t15, t17)

		require.Len(t, wts[monday], 1)
		assert.Equal(t, t15, wts[monday][0].StartTime())
		assert.Equal(t, t17, wts[monday][0].EndTime())
	})

	t.Run("timeslots for day", func(t *testing.T) {
		wts := schedule.NewWeekdayTimeSlotMap().
			AddTimeSlot(monday, t08, t10).
			AddTimeSlot(tuesday, t08, t10).
			AddTimeSlot(tuesday, t15, t17)

		require.Len(t, wts, 2)

		tuesdayTS := wts.TimeSlots(tuesday)
		require.Len(t, tuesdayTS, 2)
	})

	t.Run("without chaining", func(t *testing.T) {
		wts := schedule.NewWeekdayTimeSlotMap()
		wts.AddTimeSlot(tuesday, t08, t10)
		wts.AddTimeSlot(tuesday, t15, t17)

		require.Len(t, wts, 1)
		require.Len(t, wts.TimeSlots(tuesday), 2)
	})

	t.Run("check json encoding", func(t *testing.T) {
		var (
			slot08 = schedule.ParseTimeSlot("08:00-09:45")
			slot10 = schedule.ParseTimeSlot("10:00-11:30")
			slot15 = schedule.ParseTimeSlot("15:00-17:15")
			slot17 = schedule.ParseTimeSlot("17:00-17:45")
		)

		wts := schedule.NewWeekdayTimeSlotMap().
			Add(schedule.Monday, slot08, slot10, slot15).
			Add(schedule.Tuesday, slot10, slot17).
			Add(schedule.Wednesday, slot08, slot10, slot15).
			Add(schedule.Thursday, slot10, slot17).
			Add(schedule.Friday, slot08, slot10).
			Add(schedule.Saturday)

		jsonString := toJsonString(t, wts)
		assert.NotEmpty(t, jsonString)

		// this test really only exists as an example and for manual
		// visual inspection of the resulting data structure in json
		// so to inspect just uncomment this line but re-comment it
		// to keep test running output clean
		// t.Log(jsonString)

		/** This json is from RE-2606 as an example
			"timeSlots": {
			  "Monday":    [
				{"start": "08:00", "end": "09:45"},
				{"start": "10:00", "end": "11:30"},
				{"start": "15:00", "end": "17:15"}],
			  "Tuesday":   [
				{"start": "10:00", "end": "11:30"},
				{"start": "17:00", "end": "17:45"}],
			  "Wednesday": [
				{"start": "08:00", "end": "09:45"},
				{"start": "10:00", "end": "11:30"},
				{"start": "15:00", "end": "17:15"}],
			  "Thursday":  [
				{"start": "10:00", "end": "11:30"},
				{"start": "17:00", "end": "17:45"}],
			  "Friday":    [
				{"start": "08:00", "end": "09:45"},
				{"start": "10:00", "end": "11:30"}],
			  "Saturday"   []
			}
		***************/
	})
}

func TestWeekdayTimeSlot(t *testing.T) {
	var tests = map[string]struct {
		day      schedule.Weekday
		startStr string
		mins     int // duration in minutes
	}{
		"Sunday 00:00-00:00":  {schedule.Sunday, "00:00", 0},
		"Monday 00:00-00:00":  {schedule.Monday, "00:00", 0},
		"Tuesday 01:30-02:15": {schedule.Tuesday, "01:30", 45},
	}
	for wtsString, tc := range tests {
		t.Run(wtsString, func(t *testing.T) {
			var (
				start = schedule.ParseClock(tc.startStr)
				end   = start.Add(tc.mins)
				slot  = schedule.NewTimeSlot(start, end)
				wts   = schedule.NewWeekdayTimeSlot(tc.day, slot)
			)
			assert.Equal(t, tc.day, wts.Weekday())
			assert.Equal(t, start, wts.Start())
			assert.Equal(t, end, wts.End())
			assert.Equal(t, tc.mins, wts.Minutes())
			assert.Equal(t, tc.mins, int(wts.Duration().Minutes()))

			// to and from string
			assert.Equal(t, wtsString, wts.String())
			assert.Equal(t, wtsString, wts.ToString())
			assert.Equal(t, wts, schedule.WeekdayTimeSlotFromString(wtsString))

			// to and from int
			wtsFromInt := schedule.WeekdayTimeSlotFromInt(wts.ToInt())
			// t.Logf("%b", wts.ToInt())
			require.Equal(t, wts, wtsFromInt)
			assert.True(t, wts == wtsFromInt)
		})
	}
}
func TestWeekdayTimeSlot_ZeroValue(t *testing.T) {
	var (
		slotZero = schedule.ParseTimeSlot("00:00-00:00")
	)

	var zeroTests = map[string]schedule.WeekdayTimeSlot{
		"empty":       {},
		"new":         schedule.NewWeekdayTimeSlot(schedule.Sunday, slotZero),
		"from string": schedule.WeekdayTimeSlotFromString(""),
		"from int":    schedule.WeekdayTimeSlotFromInt(0),
	}
	for name, wts := range zeroTests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, 0, wts.ToInt())
			assert.Equal(t, schedule.Sunday, wts.Weekday())
			assert.Equal(t, schedule.NewClock(0, 0), wts.Start())
			assert.Equal(t, schedule.NewClock(0, 0), wts.End())
			assert.Equal(t, 0, wts.Minutes())
			assert.Equal(t, time.Duration(0), wts.Duration())
		})
	}
}
func TestWeekdayTimeSlot_FromString(t *testing.T) {
	var fromStringTests = map[string]struct {
		day      schedule.Weekday
		startStr string
		mins     int // duration in minutes
	}{
		"Sunday 00:00-00:00": {schedule.Sunday, "00:00", 0},
		"Sunday":             {schedule.Sunday, "00:00", 0},
		"00:00-00:00":        {schedule.Sunday, "00:00", 0},
		"Monday 01:30-02:15": {schedule.Monday, "01:30", 45},
		"Monday":             {schedule.Monday, "00:00", 0},
		"01:30-02:15":        {schedule.Sunday, "01:30", 45},
	}
	for input, tc := range fromStringTests {
		t.Run(input, func(t *testing.T) {
			var (
				start = schedule.ParseClock(tc.startStr)
				end   = start.Add(tc.mins)
				slot  = schedule.NewTimeSlot(start, end)
				wts   = schedule.NewWeekdayTimeSlot(tc.day, slot)
			)

			assert.Equal(t, wts, schedule.WeekdayTimeSlotFromString(input))
		})
	}
}
func TestWeekdayTimeSlot_Sort(t *testing.T) {
	// they should order by weekday, start time, duration
	var (
		slot0 = schedule.ParseTimeSlot("00:00-08:30")
		slot1 = schedule.ParseTimeSlot("08:00-08:30")
		slot2 = schedule.ParseTimeSlot("08:00-08:45")
		slot3 = schedule.ParseTimeSlot("08:15-08:30")
		slot4 = schedule.ParseTimeSlot("08:15-08:45")

		wts1 = schedule.NewWeekdayTimeSlot(schedule.Sunday, slot1)
		wts2 = schedule.NewWeekdayTimeSlot(schedule.Sunday, slot2)
		wts3 = schedule.NewWeekdayTimeSlot(schedule.Sunday, slot3)
		wts4 = schedule.NewWeekdayTimeSlot(schedule.Sunday, slot4)
		wts5 = schedule.NewWeekdayTimeSlot(schedule.Monday, slot0)
	)

	assert.True(t, wts1.ToInt() < wts2.ToInt())
	assert.True(t, wts2.ToInt() < wts3.ToInt())
	assert.True(t, wts3.ToInt() < wts4.ToInt())
	assert.True(t, wts4.ToInt() < wts5.ToInt())

	var (
		wtsSliceUnsorted = []schedule.WeekdayTimeSlot{wts5, wts3, wts1, wts2, wts4, wts3}
		sorted           = []schedule.WeekdayTimeSlot{wts1, wts2, wts3, wts3, wts4, wts5}
		uniqueSorted     = []schedule.WeekdayTimeSlot{wts1, wts2, wts3, wts4, wts5}
	)

	assert.Equal(t, sorted, schedule.SortWeekdayTimeSlots(wtsSliceUnsorted...))
	assert.NotEqual(t, sorted, wtsSliceUnsorted) // make sure input was not mutated
	assert.Equal(t, uniqueSorted, schedule.UniqueWeekdayTimeSlots(wtsSliceUnsorted...))
}

func TestWeekdayTimeSlot_fromString(t *testing.T) {
	var fromStringTests = map[string]struct {
		day      schedule.Weekday
		startStr string
		mins     int // duration in minutes
	}{
		"Sunday 00:00-00:00": {schedule.Sunday, "00:00", 0},
		"Sunday":             {schedule.Sunday, "00:00", 0},
		"00:00-00:00":        {schedule.Sunday, "00:00", 0},
		"Monday 01:30-02:15": {schedule.Monday, "01:30", 45},
		"Monday":             {schedule.Monday, "00:00", 0},
		"01:30-02:15":        {schedule.Sunday, "01:30", 45},
	}
	for input, tc := range fromStringTests {
		t.Run(input, func(t *testing.T) {
			var (
				start = schedule.ParseClock(tc.startStr)
				end   = start.Add(tc.mins)
				slot  = schedule.NewTimeSlot(start, end)
				wts   = schedule.NewWeekdayTimeSlot(tc.day, slot)
			)

			assert.Equal(t, wts, schedule.WeekdayTimeSlotFromString(input))
		})
	}
}

func TestWeekdayTimeSlotMap_ToFromWeekdayTimeSlots(t *testing.T) {
	var (
		slot08 = schedule.ParseTimeSlot("08:00-09:45")
		slot10 = schedule.ParseTimeSlot("10:00-11:30")
		slot15 = schedule.ParseTimeSlot("15:00-17:15")
		slot17 = schedule.ParseTimeSlot("17:00-17:45")
	)

	wtsMap := schedule.NewWeekdayTimeSlotMap().
		Add(schedule.Monday, slot08, slot10, slot15).
		Add(schedule.Tuesday, slot10, slot17).
		Add(schedule.Wednesday, slot08, slot10, slot15).
		Add(schedule.Thursday, slot10, slot17).
		Add(schedule.Friday, slot08, slot10).
		Add(schedule.Saturday)

	wtsSlice := []schedule.WeekdayTimeSlot{
		schedule.NewWeekdayTimeSlot(schedule.Monday, slot08),
		schedule.NewWeekdayTimeSlot(schedule.Monday, slot10),
		schedule.NewWeekdayTimeSlot(schedule.Monday, slot15),

		schedule.NewWeekdayTimeSlot(schedule.Tuesday, slot10),
		schedule.NewWeekdayTimeSlot(schedule.Tuesday, slot17),

		schedule.NewWeekdayTimeSlot(schedule.Wednesday, slot08),
		schedule.NewWeekdayTimeSlot(schedule.Wednesday, slot10),
		schedule.NewWeekdayTimeSlot(schedule.Wednesday, slot15),

		schedule.NewWeekdayTimeSlot(schedule.Thursday, slot10),
		schedule.NewWeekdayTimeSlot(schedule.Thursday, slot17),

		schedule.NewWeekdayTimeSlot(schedule.Friday, slot08),
		schedule.NewWeekdayTimeSlot(schedule.Friday, slot10),

		schedule.NewWeekdayTimeSlot(schedule.Saturday, schedule.TimeSlot{}),
	}

	wtsMapToSlice := wtsMap.ToWeekdayTimeSlots()
	wtsMapFromSlice := schedule.WeekdayTimeSlotMapFromSlice(wtsSlice)

	// assert equal on json encoding avoids problems with inconsistent order
	assert.Equal(t, toJsonString(t, wtsSlice), toJsonString(t, wtsMapToSlice))
	assert.Equal(t, toJsonString(t, wtsMap), toJsonString(t, wtsMapFromSlice))

	// these work only because input was in a pre-sorted order
	// the result order is sorted and NOT determined by input order
	assert.Equal(t, wtsSlice, wtsMapToSlice)
	assert.Equal(t, wtsMap, wtsMapFromSlice)
}

func TestWeekdayTimeSlots_Overlap(t *testing.T) {
	t.Run("two slots which overlap", func(t *testing.T) {
		var (
			slot1 = schedule.WeekdayTimeSlotFromString("Monday 06:00-07:00")
			slot2 = schedule.WeekdayTimeSlotFromString("Monday 06:30-07:30")
		)
		assert.True(t, slot1.OverlapsWith(slot2))
	})

	t.Run("two similar slots on different days do not overlap", func(t *testing.T) {
		var (
			slot1 = schedule.WeekdayTimeSlotFromString("Monday 06:00-07:00")
			slot2 = schedule.WeekdayTimeSlotFromString("Tuesday 06:30-07:30")
		)
		assert.False(t, slot1.OverlapsWith(slot2))
	})

	t.Run("overlap crosses midnight", func(t *testing.T) {
		var (
			slot1 = schedule.WeekdayTimeSlotFromString("Monday 23:30-00:30")
			slot2 = schedule.WeekdayTimeSlotFromString("Monday 23:30-23:45")
			slot3 = schedule.WeekdayTimeSlotFromString("Monday 23:40-00:40")
			slot4 = schedule.WeekdayTimeSlotFromString("Tuesday 00:15-01:15")
			slot5 = schedule.WeekdayTimeSlotFromString("Tuesday 00:30-01:30")
			slot6 = schedule.WeekdayTimeSlotFromString("Tuesday 00:40-01:40")
		)
		assert.True(t, slot1.OverlapsWith(slot2))
		assert.True(t, slot1.OverlapsWith(slot3))
		assert.True(t, slot1.OverlapsWith(slot4))
		assert.False(t, slot1.OverlapsWith(slot5))
		assert.False(t, slot1.OverlapsWith(slot6))

		assert.True(t, slot2.OverlapsWith(slot3))
	})

	t.Run("overlap crosses midnight on Saturday", func(t *testing.T) {
		var (
			slot1 = schedule.WeekdayTimeSlotFromString("Saturday 23:30-00:30")
			slot2 = schedule.WeekdayTimeSlotFromString("Saturday 23:30-23:45")
			slot3 = schedule.WeekdayTimeSlotFromString("Saturday 23:40-00:40")
			slot4 = schedule.WeekdayTimeSlotFromString("Sunday 00:15-01:15")
			slot5 = schedule.WeekdayTimeSlotFromString("Sunday 00:30-01:30")
			slot6 = schedule.WeekdayTimeSlotFromString("Sunday 00:40-01:40")
		)
		assert.True(t, slot1.OverlapsWith(slot2))
		assert.True(t, slot1.OverlapsWith(slot3))
		assert.True(t, slot1.OverlapsWith(slot4))
		assert.False(t, slot1.OverlapsWith(slot5))
		assert.False(t, slot1.OverlapsWith(slot6))

		assert.True(t, slot2.OverlapsWith(slot3))
	})
}

func toJsonString(t *testing.T, object interface{}) string {
	t.Helper()
	jsonBytes, err := json.MarshalIndent(object, "", "  ")
	require.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)
	return string(jsonBytes)
}
