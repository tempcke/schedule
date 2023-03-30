package schedule_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

func TestWeekdayTimeSlot_encoding(t *testing.T) {

	t.Run("interface impl check", func(t *testing.T) {
		// simple compile-time checks that will fail if the
		// type does not correctly implement the interface
		var (
			v = (*schedule.WeekdayTimeSlot)(nil)

			// read/write from/to json
			_ json.Marshaler   = v
			_ json.Unmarshaler = v

			// read/write from/to sql
			_ sql.Scanner   = v
			_ driver.Valuer = v
		)

		wts := schedule.WeekdayTimeSlotFromString("Tuesday 08:00-09:30")
		tLogJson(t, "wts", wts)
		wtsBytes, err := json.Marshal(wts)
		require.NoError(t, err)
		var decodedWts schedule.WeekdayTimeSlot
		require.NoError(t, json.Unmarshal(wtsBytes, &decodedWts))
		assert.Equal(t, wts, decodedWts)
		assert.True(t, wts.Equal(decodedWts))
	})
}

//lint:ignore U1000 keep it for dev testing
func tLogJson(t testing.TB, label string, v interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	t.Log(label + ": " + string(b))
}
