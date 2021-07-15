package hjem

import (
	"strings"
	"testing"
	"time"
)

func TestDirtyStringToInt(t *testing.T) {
	tt := []struct {
		name string
		in   string
		out  int
		err  string
	}{
		{name: "with dot", in: "10.000", out: 10000},
		{name: "multiple dots", in: "88.299.199", out: 88299199},
		{name: "additional spaces", in: "  999  ", out: 999},
		{name: "with unit", in: "64kr", out: 64},
		{name: "zero", in: "0", out: 0},
		{name: "letters only", in: "absc", err: "invalid syntax"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			o, err := DirtyStringToInt(tc.in)
			if err != nil {
				if tc.err != "" {
					if strings.Contains(err.Error(), tc.err) {
						return
					}

					t.Fatalf("unexpected error: %s (expected: %s)", err, tc.err)
				}

				t.Fatalf("received unexpected error: %s", err)
			}

			if o != tc.out {
				t.Fatalf("unexpected output: %d (expected: %d)", o, tc.out)
			}
		})
	}
}

func TestDanishDateToTime(t *testing.T) {
	tt := []struct {
		name   string
		in     string
		format string
		out    time.Time
		err    string
	}{
		{name: "basic", in: "10. jan. 2018", format: "2. jan. 2006", out: time.Date(2018, 1, 10, 0, 0, 0, 0, time.UTC)},
		{name: "basic (maj)", in: "12. maj. 2021", format: "2. jan. 2006", out: time.Date(2021, 5, 12, 0, 0, 0, 0, time.UTC)},
		{name: "basic (okt)", in: "28. okt. 2021", format: "2. jan. 2006", out: time.Date(2021, 10, 28, 0, 0, 0, 0, time.UTC)},
		{name: "syntax issue", in: "28-okt-2021", format: "2. jan. 2006", out: time.Date(2021, 10, 28, 0, 0, 0, 0, time.UTC), err: "cannot parse"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			o, err := DanishDateToTime(tc.format, tc.in)
			if err != nil {
				if tc.err != "" {
					if strings.Contains(err.Error(), tc.err) {
						return
					}

					t.Fatalf("unexpected error: %s (expected: %s)", err, tc.err)
				}

				t.Fatalf("received unexpected error: %s", err)
			}

			if o != tc.out {
				t.Fatalf("unexpected output: %v (expected: %v)", o, tc.out)
			}
		})
	}
}
