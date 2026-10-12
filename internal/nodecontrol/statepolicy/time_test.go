package statepolicy

import (
	"context"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func stamp(s string) time.Time {
	v, e := time.Parse(time.RFC3339, s)
	if e != nil {
		panic(e)
	}
	return v
}
func modelTime() TimeObservation {
	return TimeObservation{Status: Available, ProviderID: contracts.Digest{1}, ObservationID: contracts.Digest{2}, Now: stamp("2020-01-01T01:00:00Z"), Floor: stamp("2019-12-31T00:00:00Z")}
}
func factsFor(t *testing.T, n int) *verifiedFacts {
	t.Helper()
	f, e := loadFacts(context.Background(), signedBatch(t, n))
	wantCode(t, e, 0)
	t.Cleanup(f.Clear)
	return f
}

// Unit tests isolate policy branches after real signature verification; the
// integration tests additionally exercise the public entry with signed bytes.
func TestPolicyTimeBoundaries(t *testing.T) {
	for _, c := range []struct {
		name string
		now  time.Time
		want Code
	}{
		{"before metadata", stamp("2020-01-01T00:00:00Z").Add(-time.Nanosecond), Rejected},
		{"issued", stamp("2020-01-01T00:00:00Z"), 0},
		{"before expiry", stamp("2020-01-02T00:00:00Z").Add(-time.Nanosecond), 0},
		{"expiry", stamp("2020-01-02T00:00:00Z"), Rejected},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := factsFor(t, 1)
			o := modelTime()
			o.Now = c.now
			d, e := checkTime(f, o, nil, 0)
			wantCode(t, e, c.want)
			if e == nil && !d.Equal(stamp("2020-01-02T00:00:00Z")) {
				t.Fatal("deadline extended")
			}
		})
	}
	for _, delta := range []time.Duration{5 * time.Minute, 5*time.Minute + time.Nanosecond} {
		f := factsFor(t, 1)
		o := modelTime()
		f.desired.bindings.IssuedAt = o.Now.Add(delta)
		_, e := checkTime(f, o, nil, 0)
		want := Code(0)
		if delta > 5*time.Minute {
			want = Rejected
		}
		wantCode(t, e, want)
	}
	for _, accepted := range []int{1, 2} {
		f := factsFor(t, 1)
		o := modelTime()
		f.rotations[0].bindings.ValidUntil = o.Now
		_, e := checkTime(f, o, nil, accepted)
		want := Code(0)
		if accepted == 1 {
			want = Rejected
		}
		wantCode(t, e, want)
	}
	for _, field := range []string{"metadata", "key", "24h", "zero", "overflow"} {
		t.Run(field, func(t *testing.T) {
			f := factsFor(t, 0)
			o := modelTime()
			switch field {
			case "metadata":
				f.metadata.bindings.ValidUntil = o.Now
			case "key":
				f.metadata.metadata.Keys[0].NotAfter = o.Now
			case "24h":
				f.desired.bindings.ValidUntil = f.desired.bindings.IssuedAt.Add(24*time.Hour + time.Nanosecond)
			case "zero":
				o.Now = time.Time{}
			case "overflow":
				o.Now = stamp("9999-12-31T23:59:59Z")
			}
			_, e := checkTime(f, o, nil, 0)
			if e == nil {
				t.Fatal("invalid time accepted")
			}
		})
	}
}
func TestPolicyTimeUnavailable(t *testing.T) {
	for _, c := range []struct {
		s    ObservationStatus
		want Code
	}{{Unknown, AssuranceMissing}, {Unavailable, AssuranceMissing}, {Rollback, SecurityFault}, {ObservationStatus(99), InvalidInput}} {
		o := modelTime()
		o.Status = c.s
		_, e := checkTime(factsFor(t, 0), o, nil, 0)
		wantCode(t, e, c.want)
	}
	for _, c := range []string{"floor", "provider", "oldfloor", "empty observation"} {
		t.Run(c, func(t *testing.T) {
			o := modelTime()
			old := &StateModel{TimeProviderID: o.ProviderID, TimeFloor: o.Floor}
			switch c {
			case "floor":
				o.Floor = o.Now.Add(time.Second)
			case "provider":
				old.TimeProviderID[0] = 9
			case "oldfloor":
				old.TimeFloor = o.Floor.Add(time.Second)
			case "empty observation":
				o.ObservationID = contracts.Digest{}
			}
			_, e := checkTime(factsFor(t, 0), o, old, 0)
			if e == nil {
				t.Fatal("untrusted observation accepted")
			}
		})
	}
}
func TestPolicyTimeNoRatchet(t *testing.T) {
	f := factsFor(t, 0)
	o := modelTime()
	before := o
	f.metadata.bindings.IssuedAt = o.Now // New metadata may authorize an older desired.
	for i := 0; i < 3; i++ {
		_, e := checkTime(f, o, nil, 0)
		wantCode(t, e, 0)
	}
	if !reflect.DeepEqual(before, o) {
		t.Fatal("time changed")
	}
}
