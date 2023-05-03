package schedule_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

func TestDate(t *testing.T) {
	var (
		year       = 2022
		month      = time.February
		day        = 1
		dateString = "2022-02-01"
	)
	d := schedule.NewDate(year, month, day)

	assert.Equal(t, dateString, d.String())
	assert.Equal(t, d, *schedule.ParseDate(dateString))

	dt := d.ToTime()
	assert.Equal(t, year, dt.Year())
	assert.Equal(t, month, dt.Month())
	assert.Equal(t, day, dt.Day())

	assert.Equal(t, dateString, schedule.NewDateFromTime(dt).String())

	t.Run("today", func(t *testing.T) {
		now := time.Now()
		d := schedule.Today()
		assert.Equal(t, now.Year(), d.Year())
		assert.Equal(t, now.Month(), d.Month())
		assert.Equal(t, now.Day(), d.Day())
	})

	t.Run("invalid date", func(t *testing.T) {
		d := schedule.NewDate(2022, 13, 99)
		dTime := d.ToTime()
		require.NotEmpty(t, dTime)
		assert.Equal(t, dTime.Year(), d.Year())
		assert.Equal(t, dTime.Month(), d.Month())
		assert.Equal(t, dTime.Day(), d.Day())
	})

	t.Run("parse invalid", func(t *testing.T) {
		var badDate = "foo bar"
		d := schedule.ParseDate(badDate)
		require.Nil(t, d)
	})
}

func TestDate_jsonEncodeDecode(t *testing.T) {
	var (
		year  = 2022
		month = time.February
		day   = 1
	)
	d := schedule.NewDate(year, month, day)

	t.Run("encode date", func(t *testing.T) {
		// encode
		var jsonDate = `"` + d.String() + `"`
		jsonBytes, err := json.Marshal(d)
		require.NoError(t, err)
		assert.Equal(t, jsonDate, string(jsonBytes))

		// decode
		var decodedDate schedule.Date
		err = json.Unmarshal(jsonBytes, &decodedDate)
		require.NoError(t, err)
		assert.Equal(t, d, decodedDate)
	})

	t.Run("UnmarshalJSON empty or null", func(t *testing.T) {
		type Dates struct {
			A, C, E schedule.Date
			B, D, F *schedule.Date
		}
		jsonBytes := []byte(`{"A":"", "B": "", "C": null, "D": null}`)

		// decode
		var dates Dates
		err := json.Unmarshal(jsonBytes, &dates)
		require.NoError(t, err)

		// ""
		assert.True(t, dates.A.IsZero())
		assert.True(t, dates.B.IsZero())

		// null
		assert.True(t, dates.C.IsZero())
		assert.True(t, dates.D.IsZero())

		// ommitted
		assert.True(t, dates.E.IsZero())
		assert.True(t, dates.F.IsZero())
	})
	t.Run("UnmarshalJSON invalid input", func(t *testing.T) {
		type Dates struct {
			A schedule.Date
		}
		jsonBytes := []byte(`{"A":"invalid-input"}`)

		// decode
		var dates Dates
		err := json.Unmarshal(jsonBytes, &dates)
		require.Error(t, err)
	})
	t.Run("decode datetime into date", func(t *testing.T) {
		// so if someone passes a full RFC3339 lets parse the date and ignore the time
		var (
			now   = time.Now()
			today = schedule.Today()
		)
		var tests = []string{
			now.Format(time.RFC3339),
			now.Format(time.RFC3339Nano),
			now.Format("2006-01-02 15:04:05"),
		}
		for _, input := range tests {
			var (
				jsonStr = fmt.Sprintf(`{"Date": "%v"}`, input)
				v       = struct{ Date schedule.Date }{}
			)
			require.NoError(t, json.Unmarshal([]byte(jsonStr), &v), input)
			assert.Equal(t, today.String(), v.Date.String(), input)
		}
		for _, input := range tests {
			var d = schedule.ParseDate(input)
			require.NotNil(t, d)
			assert.Equal(t, today.String(), d.String(), input)
		}
	})
}

func TestDate_Scan(t *testing.T) {
	// this is an example of how you scan into a schedule.Date
	var fromString = "2022-01-01"

	// some entity with a DateRange in it
	var entity = struct {
		DateRange schedule.DateRange
	}{}

	// this represents the values in the database
	// we will pretend there is a from but no until
	var (
		dbFrom  interface{} = []byte(fromString)
		dbUntil interface{} = nil
	)

	// Step1: init the input vars
	// var (
	// 	from  = schedule.ZeroDate()
	// 	until = schedule.Date{}
	// )
	var from schedule.Date
	var until *schedule.Date

	// Step2: Scan
	// this does what happens when you do
	// read.QueryRow(ctx, query, queryArgs...).Scan(from, until)
	require.NoError(t, from.Scan(dbFrom))
	require.NoError(t, until.Scan(dbUntil))

	// Step3: fill out your entity after the scan
	entity.DateRange.From = from
	entity.DateRange.Until = until

	require.Equal(t, fromString, from.String())
	require.Nil(t, entity.DateRange.Until)
}

func TestMinMaxDate(t *testing.T) {
	a := schedule.Today()
	b := schedule.Today().AddDate(0, 0, 10)

	assert.Equal(t, a, *schedule.MinDate(&a, &b))
	assert.Equal(t, a, *schedule.MinDate(&b, &a))
	assert.Equal(t, a, *schedule.MinDate(&a, nil))
	assert.Equal(t, b, *schedule.MinDate(nil, &b))
	assert.Nil(t, schedule.MinDate(nil, nil))

	assert.Equal(t, b, *schedule.MaxDate(&a, &b))
	assert.Equal(t, b, *schedule.MaxDate(&b, &a))
	assert.Equal(t, a, *schedule.MaxDate(&a, nil))
	assert.Equal(t, b, *schedule.MaxDate(nil, &b))
	assert.Nil(t, schedule.MaxDate(nil, nil))
}

func TestDate_Sub(t *testing.T) {
	var (
		// weeks worth of days
		d0 = schedule.NewDate(2022, 5, 1) // Sunday
		d1 = d0.Next()                    // Monday
		d2 = d1.Next()                    // Tuesday
		d3 = d2.Next()                    // Wednesday
		d4 = d3.Next()                    // Thursday
		d5 = d4.Next()                    // Friday
		d6 = d5.Next()                    // Saturday
	)

	assert.Equal(t, 0, d0.Sub(d0))
	assert.Equal(t, -5, d1.Sub(d6))
	assert.Equal(t, 2, d5.Sub(d3))
}
