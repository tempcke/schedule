package schedule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

func TestDateRange_Overlaps(t *testing.T) {
	jan1 := schedule.NewDate(2020, 1, 1)
	jan15 := schedule.NewDate(2020, 1, 15)
	jan30 := schedule.NewDate(2020, 1, 30)
	march1 := schedule.NewDate(2020, 3, 1)
	march30 := schedule.NewDate(2020, 3, 30)

	var tests = map[string]struct {
		dr1, dr2   schedule.DateRange
		doOverlaps bool
	}{
		"not in range": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan30),
			schedule.NewDateRange().WithFrom(march1).WithUntil(march30),
			false,
		},
		"in range": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(march30),
			schedule.NewDateRange().WithFrom(jan30).WithUntil(march30),
			true,
		},
		"in range no until on before": {
			schedule.NewDateRange().WithFrom(jan1),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(jan30),
			true,
		},
		"in range no until on inside": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan30),
			schedule.NewDateRange().WithFrom(jan15),
			true,
		},
		"in range no until after": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan30),
			false,
		},
		"in range no until": {
			schedule.NewDateRange().WithFrom(jan1),
			schedule.NewDateRange().WithFrom(jan15),
			true,
		},
		"in range with from/nil until and empty date range on second": {
			schedule.NewDateRange().WithFrom(jan1),
			schedule.NewDateRange(),
			true,
		},
		"in range no until, from=until": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan15),
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.doOverlaps, test.dr1.Overlaps(test.dr2))
			assert.Equal(t, test.doOverlaps, test.dr2.Overlaps(test.dr1))
		})
	}
}

func TestDateRange_Merge(t *testing.T) {
	jan1 := schedule.NewDate(2020, 1, 1)
	jan15 := schedule.NewDate(2020, 1, 15)
	jan30 := schedule.NewDate(2020, 1, 30)
	march1 := schedule.NewDate(2020, 3, 1)

	var tests = map[string]struct{ dr1, dr2, expect schedule.DateRange }{
		"date range merge with no until": {
			schedule.NewDateRange().WithFrom(jan1),
			schedule.NewDateRange().WithFrom(jan15),
			schedule.NewDateRange().WithFrom(jan15),
		},
		"date range merge zero with no until": {
			schedule.ZeroDateRange(),
			schedule.NewDateRange().WithFrom(jan30),
			schedule.ZeroDateRange(),
		},
		"date range merge with one until (valid merge)": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan30),
			schedule.NewDateRange().WithFrom(jan15),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(jan30),
		},
		"date range merge with one until (zero merge)": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan30),
			schedule.ZeroDateRange(),
		},
		"date range merge with both until (inner merge)": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan30),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(march1),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(jan30),
		},
		"date range merge with both until (zero merge)": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan30).WithUntil(march1),
			schedule.ZeroDateRange(),
		},
		"date range merge with both until (inside merge)": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(march1),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(jan30),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(jan30),
		},
		"date range merge with both until (equal merge same day)": {
			schedule.NewDateRange().WithFrom(march1).WithUntil(march1),
			schedule.NewDateRange().WithFrom(march1).WithUntil(march1),
			schedule.NewDateRange().WithFrom(march1).WithUntil(march1),
		},
		"date range merge with both until (equal merge)": {
			schedule.NewDateRange().WithFrom(jan15).WithUntil(march1),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(march1),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(march1),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			merged := test.dr1.Merge(test.dr2)
			assert.True(t, merged.Equal(test.expect))
		})
	}
}

func TestDateRange(t *testing.T) {
	t.Run("new date range defaults to today", func(t *testing.T) {
		assert.Equal(t, schedule.Today(), schedule.NewDateRange().From)
	})

	t.Run("DayCount", func(t *testing.T) {
		var (
			today    = schedule.Today()
			tomorrow = today.Next()
			lastWeek = today.AddDate(0, 0, -7)

			dr  = schedule.DateRange{}
			dr0 = dr.WithFrom(today)
		)

		tests := map[string]struct {
			dr schedule.DateRange
			c  int
		}{
			"empty":          {dr, 0},
			"one day":        {dr0.WithUntil(today), 1},
			"two days":       {dr0.WithUntil(tomorrow), 2},
			"until nil":      {dr0, schedule.InfDays},
			"until lastWeek": {dr0.WithUntil(lastWeek), 0},
		}
		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, tc.c, tc.dr.DayCount())
			})
		}
	})
}

func TestDateRange_Validate(t *testing.T) {
	var dr schedule.DateRange

	validateTests := map[string]struct {
		dr  schedule.DateRange
		err error
	}{
		// valid cases
		"no until":    {dr.WithFrom(schedule.Today()), nil},
		"zero length": {dr.WithFrom(schedule.Today()).WithUntil(schedule.Today()), nil},
		"one month": {
			dr: dr.WithFrom(*schedule.ParseDate("2022-02-01")).
				WithUntil(*schedule.ParseDate("2022-03-01")),
			err: nil,
		},

		// error cases
		"empty":   {dr, schedule.ErrFromRequired},
		"no from": {dr.WithUntil(schedule.Today()), schedule.ErrFromRequired},
		"past until": {
			dr: dr.WithFrom(*schedule.ParseDate("2022-02-01")).
				WithUntil(*schedule.ParseDate("2022-01-01")),
			err: schedule.ErrPastUntil,
		},
	}
	for name, tc := range validateTests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.err, tc.dr.Validate())
		})
	}
}

func TestDateRange_Equal(t *testing.T) {
	jan1 := schedule.NewDate(2020, 1, 1)
	jan15 := schedule.NewDate(2020, 1, 15)
	march30 := schedule.NewDate(2020, 3, 30)

	var tests = map[string]struct {
		dr1, dr2 schedule.DateRange
		equal    bool
	}{
		"equal with untils": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			true,
		},
		"equal with untils (from==until)": {
			schedule.NewDateRange().WithFrom(march30).WithUntil(march30),
			schedule.NewDateRange().WithFrom(march30).WithUntil(march30),
			true,
		},
		"not equal with untils": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(march30),
			schedule.NewDateRange().WithFrom(jan15).WithUntil(march30),
			false,
		},
		"not equal with one until": {
			schedule.NewDateRange().WithFrom(jan1).WithUntil(jan15),
			schedule.NewDateRange().WithFrom(jan1),
			false,
		},
		"not equal with no until": {
			schedule.NewDateRange().WithFrom(jan15),
			schedule.NewDateRange().WithFrom(jan1),
			false,
		},
		"equal with no until": {
			schedule.NewDateRange().WithFrom(jan15),
			schedule.NewDateRange().WithFrom(jan15),
			true,
		},
		"equal zeros": {
			schedule.ZeroDateRange(),
			schedule.ZeroDateRange(),
			true,
		},
		"not equal with zero": {
			schedule.NewDateRange().WithFrom(jan15),
			schedule.ZeroDateRange(),
			false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.equal, test.dr1.Equal(test.dr2))
		})
	}
}

func TestDateRange_Exceeds(t *testing.T) {
	// a child date range exceeds its parent when
	// the child starts before or ends after
	var (
		jan1 = schedule.NewDate(2020, 1, 1)
		jan2 = schedule.NewDate(2020, 1, 2)
		jan3 = schedule.NewDate(2020, 1, 3)

		drJan1Jan2 = schedule.NewDateRangeUntil(jan1, &jan2)
		drJan2Jan2 = schedule.NewDateRangeUntil(jan2, &jan2)
		drJan2Nil  = schedule.NewDateRangeUntil(jan2, nil)
		drJan2Jan3 = schedule.NewDateRangeUntil(jan2, &jan3)
		drJan1Jan3 = schedule.NewDateRangeUntil(jan2, &jan3)
	)

	assert.NotNil(t, drJan2Jan3, drJan1Jan2, drJan1Jan3) // do not commit this line

	tests := map[string]struct {
		parentDR, childDr schedule.DateRange
		exceeds           bool
	}{
		"exact match":      {drJan2Jan2, drJan2Jan2, false},
		"before from":      {drJan2Jan2, drJan1Jan2, true},
		"after until":      {drJan2Jan2, drJan2Jan3, true},
		"before and after": {drJan2Jan2, drJan1Jan3, true},

		"parent until nil": {drJan2Nil, drJan2Jan3, false},
		"child until nil":  {drJan2Jan2, drJan2Nil, true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.exceeds, tc.childDr.Exceeds(tc.parentDR))
		})
	}
}
