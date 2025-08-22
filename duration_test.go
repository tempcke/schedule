package schedule_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/schedule"
)

type encodable interface {
	// read/write from/to json values
	json.Marshaler
	json.Unmarshaler

	// read/write from/to json keys
	encoding.TextMarshaler
	encoding.TextUnmarshaler

	// read/write from/to sql
	sql.Scanner
	driver.Valuer
}

var _ encodable = (*schedule.Duration)(nil)

func TestDuration_encode(t *testing.T) {
	t.Run("interface impl check", func(t *testing.T) {
		// simple compile-time checks that will fail if the
		// type does not correctly implement the interface
		var (
			// read/write from/to json values
			_ json.Marshaler   = (*schedule.Duration)(nil)
			_ json.Unmarshaler = (*schedule.Duration)(nil)

			// read/write from/to json keys
			_ encoding.TextMarshaler   = (*schedule.Duration)(nil)
			_ encoding.TextUnmarshaler = (*schedule.Duration)(nil)

			// read/write from/to sql
			_ sql.Scanner   = (*schedule.Duration)(nil)
			_ driver.Valuer = (*schedule.Duration)(nil)
		)
	})
	t.Run("encode key and value", func(t *testing.T) {
		type DurationMap map[schedule.Duration]schedule.Duration
		var (
			k          = schedule.NewDuration(time.Second * 2)
			v          = schedule.NewDuration(time.Second * 3)
			expectJSON = `{"2s":"3s"}`
		)
		var in = DurationMap{k: v}
		bytes, err := json.Marshal(in)
		require.NoError(t, err)
		require.Equal(t, expectJSON, string(bytes))
		var out DurationMap
		require.NoError(t, json.Unmarshal(bytes, &out), string(bytes))
		require.Equal(t, in, out)
	})
	t.Run("int decode", func(t *testing.T) {
		// in case it is already encoded as a time.Duration int
		var tests = []time.Duration{
			time.Hour * 2,
			time.Minute * 3,
		}
		for _, tc := range tests {
			t.Run(tc.String(), func(t *testing.T) {
				var (
					jsonStr = fmt.Sprintf(`{"%v":%v}`, tc.String(), int(tc))
					expect  = map[schedule.Duration]schedule.Duration{
						schedule.NewDuration(tc): schedule.NewDuration(tc),
					}
					out map[schedule.Duration]schedule.Duration
				)
				require.NoError(t, json.Unmarshal([]byte(jsonStr), &out), "jsonStr: %s", jsonStr)
				require.Equal(t, expect, out)
			})
		}
	})
	t.Run("empty or null", func(t *testing.T) {
		type Durations struct {
			A, C, E schedule.Duration
			B, D, F *schedule.Duration
		}
		var (
			zero             schedule.Duration
			jsonBytes        = []byte(`{"A":"", "B": "", "C": null, "D": null}`)
			expectDurations  = Durations{A: zero, B: &zero, C: zero, E: zero}
			decodedDurations Durations
		)
		require.NoError(t, json.Unmarshal(jsonBytes, &decodedDurations))
		assert.EqualValues(t, expectDurations, decodedDurations)
	})
	t.Run("string", func(t *testing.T) {
		var d = time.Hour * 3
		assert.Equal(t, d.String(), schedule.NewDuration(d).String())
	})
}
