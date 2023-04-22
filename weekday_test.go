package schedule_test

import (
	"encoding/json"
	"fmt"
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

	t.Run("int decode", func(t *testing.T) {
		var tests = map[string]struct {
			num int
			day schedule.Weekday
		}{
			"0 Sunday":    {0, schedule.Sunday},
			"1 Monday":    {1, schedule.Monday},
			"2 Tuesday":   {2, schedule.Tuesday},
			"3 Wednesday": {3, schedule.Wednesday},
			"4 Thursday":  {4, schedule.Thursday},
			"5 Friday":    {5, schedule.Friday},
			"6 Saturday":  {6, schedule.Saturday},
			"7 Sunday":    {7, schedule.Sunday},
			"8 Monday":    {8, schedule.Monday},
		}

		for name, tc := range tests {
			var (
				jsonStr = fmt.Sprintf(`{"Day":%d}`, tc.num)
				v       = struct{ Day schedule.Weekday }{}
			)
			require.NoError(t, json.Unmarshal([]byte(jsonStr), &v), name)
			require.Equal(t, tc.day, v.Day, name)
		}
	})

	t.Run("empty or null", func(t *testing.T) {
		type Days struct {
			A, C schedule.Weekday
			B, D *schedule.Weekday
		}
		jsonBytes := []byte(`{"A": null, "B": null}`)

		var days Days
		require.NoError(t, json.Unmarshal(jsonBytes, &days))

		assert.True(t, days.A == schedule.Sunday)
		assert.True(t, days.B == nil)
		assert.True(t, days.C == schedule.Sunday)
		assert.True(t, days.D == nil)
	})
}
