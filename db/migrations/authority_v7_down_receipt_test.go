package migrations

import (
	"testing"
	"time"
)

func TestAuthorityV7DownConsumeTimeValid(t *testing.T) {
	startedAt := time.Date(2026, time.August, 29, 18, 30, 0, 123456789, time.UTC)
	tests := []struct {
		name       string
		startedAt  time.Time
		consumedAt time.Time
		want       bool
	}{
		{name: "equal", startedAt: startedAt, consumedAt: startedAt, want: true},
		{name: "same instant different location", startedAt: startedAt, consumedAt: startedAt.In(time.FixedZone("fixture", 3*60*60)), want: true},
		{name: "later", startedAt: startedAt, consumedAt: startedAt.Add(time.Nanosecond), want: true},
		{name: "earlier", startedAt: startedAt, consumedAt: startedAt.Add(-time.Nanosecond), want: false},
		{name: "zero consumed", startedAt: startedAt, consumedAt: time.Time{}, want: false},
		{name: "zero started", startedAt: time.Time{}, consumedAt: startedAt, want: false},
		{name: "both zero", startedAt: time.Time{}, consumedAt: time.Time{}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := authorityV7DownConsumeTimeValid(test.startedAt, test.consumedAt); got != test.want {
				t.Fatalf("authorityV7DownConsumeTimeValid(%s, %s) = %t, want %t", test.startedAt, test.consumedAt, got, test.want)
			}
		})
	}
}
