package schedule_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

func TestWeekday_jsonEncode(t *testing.T) {
	t.Run("encode key and value", func(t *testing.T) {
		type WeekdayMap map[schedule.Weekday]schedule.Weekday

		expectJSON := `{"Monday":"Tuesday"}`

		var in = WeekdayMap{schedule.Monday: schedule.Tuesday}

		bytes, err := json.Marshal(in)
		require.NoError(t, err)
		require.Equal(t, expectJSON, string(bytes))

		var out WeekdayMap
		require.NoError(t, json.Unmarshal(bytes, &out))
		require.Equal(t, in, out)
	})

	t.Run("case insensitive decode", func(t *testing.T) {
		days := []schedule.Weekday{
			schedule.Sunday,
			schedule.Monday,
			schedule.Tuesday,
			schedule.Wednesday,
			schedule.Thursday,
			schedule.Friday,
			schedule.Saturday,
		}
		for _, testDay := range days {
			dayStr := testDay.String()
			dayNames := []string{
				dayStr,
				strings.ToLower(dayStr),
				strings.ToUpper(dayStr),
			}
			for _, dayName := range dayNames {
				var decodedDay schedule.Weekday
				jsonValue := []byte(`"` + dayName + `"`)
				err := json.Unmarshal(jsonValue, &decodedDay)
				require.NoError(t, err, "unexpected error on %s", dayName)
				assert.Equal(t, testDay, decodedDay, "did not match %s", dayName)
			}
		}
	})
}
