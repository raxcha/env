package utils

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	date := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	withTime := time.Date(2026, 9, 29, 14, 35, 0, 0, time.UTC)
	fallback := time.Date(2002, 10, 25, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name, input string
		want        time.Time
	}{
		{"release-time date", "2026.09.29", date},
		{"release-time weekday", "2026.09.29 (Tuesday)", date},
		{"date weekday", "2026.09.29 (Tue)", date},
		{"date", "2026.09.29", date},
		{"plain weekday", "2026.09.29 tuesday", date},
		{"plain weekday with time", "2026.09.29 14:35 tuesday", withTime},
		{"time double space weekday", "2026.09.29  14:35 (Tuesday)", withTime},
		{"time trailing space", "2026.09.29 14:35 ", withTime},
		{"last-edited-time trailing space", "2026.09.29 14:35 ", withTime},
		{"last-edited-time weekday", "2026.09.29 14:35 (Tuesday)", withTime},
		{"localized weekday", " 2026.09.29 14:35 (terça-feira)  ", withTime},
		{"legacy time first", "14:35 2026.09.29", withTime},
		{"invalid date", "2026.02.30 (Monday)", fallback},
		{"invalid time", "2026.09.29 25:35 (Tuesday)", fallback},
		{"unclosed weekday", "2026.09.29 (Tuesday", fallback},
		{"trailing garbage", "2026.09.29 (Tuesday) garbage", fallback},
		{"empty", "", fallback},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseTime(tc.input); !got.Equal(tc.want) {
				t.Errorf("ParseTime(%q) = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}
