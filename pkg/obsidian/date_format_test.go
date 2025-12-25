package obsidian_test

import (
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestExpandDatePattern(t *testing.T) {
	now := time.Date(2024, 1, 15, 14, 30, 52, 0, time.UTC)

	cases := []struct {
		name    string
		pattern string
		want    string
	}{
		{"brace date", "{YYYY-MM-DD}", "2024-01-15"},
		{"plain date", "YYYY-MM-DD", "2024-01-15"},
		{"brace datetime minutes", "{YYYY-MM-DD-HHmm}", "2024-01-15-1430"},
		{"brace datetime seconds", "{YYYY-MM-DD-HHmmss}", "2024-01-15-143052"},
		{"zettel braced", "{YYYYMMDDHHmmss}", "20240115143052"},
		{"zettel plain", "YYYYMMDDHHmmss", "20240115143052"},
		{"weekday", "dddd", "Monday"},
		{"month name", "MMMM", "January"},
		{"month abbrev", "MMM", "Jan"},
		{"bracket literal", "YYYY-[ToDo]-MM", "2024-ToDo-01"},
		{"bracket literal preserved", "[Mon]-YYYY", "Mon-2024"},
		{"braces + literal", "{YYYY}-[Mon]-{MM}", "2024-Mon-01"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, obsidian.ExpandDatePattern(tc.pattern, now))
		})
	}
}

func TestFormatDatePatternErrorsOnEmpty(t *testing.T) {
	_, err := obsidian.FormatDatePattern("   ", time.Now())
	assert.Error(t, err)
}

