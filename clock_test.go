package schedule_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

func TestClock(t *testing.T) {
	oneThirty := schedule.NewClock(1, 30)

	t.Run("interface impl check", func(t *testing.T) {
		// simple compile-time checks that will fail if the
		// type does not correctly implement the interface
		var (
			// read/write from/to json
			_ json.Marshaler   = (*schedule.Clock)(nil)
			_ json.Unmarshaler = (*schedule.Clock)(nil)

			// read/write from/to sql
			_ sql.Scanner   = (*schedule.Clock)(nil)
			_ driver.Valuer = (*schedule.Clock)(nil)
		)
	})

	tests := map[string]struct {
		hr, min int
	}{
		"01:15": {1, 15},
		"23:59": {23, 59},
	}
	for hmStr, tc := range tests {
		t.Run(hmStr, func(t *testing.T) {
			newHM := schedule.NewClock(tc.hr, tc.min)
			assert.Equal(t, hmStr, newHM.String())
			assert.Equal(t, tc.hr, newHM.Hour())
			assert.Equal(t, tc.min, newHM.Minute())

			// confirm ParseClock and NewClock are ==
			parsedHM := schedule.ParseClock(hmStr)
			assert.Equal(t, parsedHM, newHM)
			assert.True(t, newHM.Equal(parsedHM))

		})
	}

	t.Run("add", func(t *testing.T) {
		c := oneThirty.Add(61)
		require.Equal(t, 2, c.Hour())
		require.Equal(t, 31, c.Minute())
		require.Equal(t, 1, oneThirty.Hour(), "should not mutate")
		require.Equal(t, 30, oneThirty.Minute(), "should not mutate")
	})

	t.Run("subtract", func(t *testing.T) {
		c := oneThirty.Subtract(61)
		require.Equal(t, 0, c.Hour())
		require.Equal(t, 29, c.Minute())
		require.Equal(t, 1, oneThirty.Hour(), "should not mutate")
		require.Equal(t, 30, oneThirty.Minute(), "should not mutate")
	})

	t.Run("negative", func(t *testing.T) {
		t2259 := schedule.NewClock(22, 59)
		assert.Equal(t, t2259, schedule.NewClock(-1, -1))
		assert.Equal(t, t2259, schedule.NewClock(0, -61))
		assert.Equal(t, t2259, schedule.NewClock(0, 0).Subtract(61))
	})

	t.Run("greater than 60 mins", func(t *testing.T) {
		assert.Equal(t, oneThirty, schedule.NewClock(0, 90))
		assert.Equal(t, oneThirty, schedule.ParseClock("00:90"))
		assert.Equal(t, oneThirty, schedule.NewClock(0, 0).Add(90))
	})

	t.Run("greater than 24 hrs", func(t *testing.T) {
		twentyFiveThirty := schedule.NewClock(25, 30)
		assert.Equal(t, oneThirty, twentyFiveThirty)
		assert.Equal(t, "01:30", twentyFiveThirty.String())
		assert.Equal(t, oneThirty, schedule.ParseClock("25:30"))

		oneDayOneMin := schedule.NewClock(23, 59).Add(2)
		assert.Equal(t, 0, oneDayOneMin.Hour())
		assert.Equal(t, 1, oneDayOneMin.Minute())
		assert.Equal(t, "00:01", oneDayOneMin.String())
	})

	t.Run("ToDuration", func(t *testing.T) {
		expectDuration := 90 * time.Minute // 1hr 30m
		assert.Equal(t, expectDuration, oneThirty.ToDuration())
	})

	t.Run("to from json as string", func(t *testing.T) {
		// when you json encode to string, the quotes are part of the value
		jsonHM := `"` + oneThirty.String() + `"`

		// encode newHM
		jsonBytes, err := json.Marshal(oneThirty)
		require.NoError(t, err)
		assert.Equal(t, jsonHM, string(jsonBytes))

		// decode newHM
		var decodedHM schedule.Clock
		err = json.Unmarshal(jsonBytes, &decodedHM)
		require.NoError(t, err)
		assert.Equal(t, oneThirty, decodedHM)
	})

	t.Run("to from SQL as string", func(t *testing.T) {
		// encode to sql
		v, err := oneThirty.Value()
		require.NoError(t, err)
		require.EqualValues(t, oneThirty.String(), v)

		// decode from sql
		var c schedule.Clock
		require.NoError(t, c.Scan(v))
		assert.Equal(t, oneThirty, c)
	})

	scanTests := map[string]struct {
		v interface{}
	}{
		"int":    {90},        // this really should never happen but we can support it if it does
		"int64":  {int64(90)}, // this really should never happen but we can support it if it does
		"string": {"01:30"},
		"bytes":  {[]byte("01:30")},
	}
	for name, tc := range scanTests {
		t.Run(name, func(t *testing.T) {
			var c schedule.Clock
			require.NoError(t, c.Scan(tc.v))
			assert.Equal(t, oneThirty, c)
		})
	}
}

func TestClock_BeforeAfterEqual(t *testing.T) {
	var (
		c1259 = schedule.NewClock(12, 59)
		c1300 = schedule.NewClock(13, 00)
		c1301 = schedule.NewClock(13, 01)
	)
	assert.True(t, c1259.Before(c1300))
	assert.True(t, c1259.Before(c1301))
	assert.True(t, c1300.Before(c1301))

	assert.True(t, c1301.After(c1300))
	assert.True(t, c1301.After(c1259))
	assert.True(t, c1300.After(c1259))

	assert.True(t, c1259.Equal(c1259))
	assert.True(t, c1300.Equal(c1300))
	assert.True(t, c1301.Equal(c1301))
}
