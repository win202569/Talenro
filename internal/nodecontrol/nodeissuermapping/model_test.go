package nodeissuermapping

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Reflection is test-only and fills every exported nested value, including domains
// irrelevant to current lower policy. These are ownership controls, not signed evidence.
func fillOwnedTest(v reflect.Value, stamp time.Time) {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		v.Set(reflect.ValueOf(stamp))
		return
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				fillOwnedTest(v.Field(i), stamp)
			}
		}
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		fillOwnedTest(v.Elem(), stamp)
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 2, 2))
		for i := 0; i < v.Len(); i++ {
			fillOwnedTest(v.Index(i), stamp)
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			fillOwnedTest(v.Index(i), stamp)
		}
	case reflect.String:
		v.SetString("x")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	}
}
func ownedTestClaims() Claims {
	var in Claims
	fillOwnedTest(reflect.ValueOf(&in).Elem(), time.Date(2020, 1, 1, 5, 0, 0, 123, time.FixedZone("explicit", 4*3600)))
	return in
}
func ownedTestJSON(t testing.TB, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func setOwnedTestTimes(v reflect.Value) {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		if !v.Interface().(time.Time).IsZero() {
			v.Set(reflect.ValueOf(time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC)))
		}
		return
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				setOwnedTestTimes(v.Field(i))
			}
		}
	case reflect.Pointer:
		if !v.IsNil() {
			setOwnedTestTimes(v.Elem())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			setOwnedTestTimes(v.Index(i))
		}
	}
}
func ownedTestClone[T any](t testing.TB, v T) T {
	t.Helper()
	var out T
	if e := json.Unmarshal(ownedTestJSON(t, v), &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func ownedTestViews(v reflect.Value) (out []reflect.Value) {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return nil
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			out = append(out, reflect.ValueOf(v.Interface()))
			out = append(out, ownedTestViews(v.Elem())...)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				out = append(out, ownedTestViews(v.Field(i))...)
			}
		}
	case reflect.Slice:
		if v.Len() > 0 {
			out = append(out, reflect.ValueOf(v.Interface()))
		}
		for i := 0; i < v.Len(); i++ {
			out = append(out, ownedTestViews(v.Index(i))...)
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			out = append(out, ownedTestViews(v.Index(i))...)
		}
	}
	return out
}
func assertOwnedTestZero(t testing.TB, views []reflect.Value) {
	t.Helper()
	for _, v := range views {
		if v.Kind() == reflect.Pointer {
			if !v.Elem().IsZero() {
				t.Fatal("owned pointer not cleared", v.Type())
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				if !v.Index(i).IsZero() {
					t.Fatal("owned backing not cleared", v.Type(), i)
				}
			}
		}
	}
}
func assertOwnedUnit[T any](t testing.TB, in T, own func(context.Context, T) (T, error), clearOwned func(*T)) {
	t.Helper()
	// Exercise every reachable raw/collection/pointer independently in both directions.
	for direction := 0; direction < 2; direction++ {
		trial := ownedTestClone(t, in)
		count := len(ownedTestViews(reflect.ValueOf(trial)))
		for index := 0; index < count; index++ {
			trial = ownedTestClone(t, in)
			owned, err := own(context.Background(), trial)
			if err != nil {
				t.Fatal(err)
			}
			var changed, retained any = trial, owned
			if direction == 1 {
				changed, retained = owned, trial
			}
			before := string(ownedTestJSON(t, retained))
			view := ownedTestViews(reflect.ValueOf(changed))[index]
			if view.Kind() == reflect.Pointer {
				view.Elem().SetZero()
			} else {
				view.Index(0).SetZero()
			}
			if string(ownedTestJSON(t, retained)) != before {
				t.Fatal("mutable alias", direction, index)
			}
			clearOwned(&owned)
		}
	}
	before := string(ownedTestJSON(t, in))
	expected := ownedTestClone(t, in)

	out, e := own(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatal("complete normalized value not retained")
	}
	views := ownedTestViews(reflect.ValueOf(out))
	clearOwned(&out)
	assertOwnedTestZero(t, views)
	var zero T
	if !reflect.DeepEqual(out, zero) {
		t.Fatal("not fully zero")
	}
	if string(ownedTestJSON(t, in)) != before {
		t.Fatal("caller mutated")
	}
	var empty T
	emptyOwned, emptyError := own(context.Background(), empty)
	if emptyError != nil || !reflect.DeepEqual(emptyOwned, empty) {
		t.Fatal("empty collections", emptyError)
	}
	clearOwned(&emptyOwned)
	if _, e := own(nil, in); e != ErrInvalidInput {
		t.Fatal("nil context", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := own(ctx, in); e != context.Canceled {
		t.Fatal("cancel", e)
	}
}

type ownershipFailureContext struct {
	context.Context
	calls, stop int
	failure     error
	panicAt     bool
}

func (c *ownershipFailureContext) Err() error {
	c.calls++
	if c.stop > 0 && c.calls >= c.stop {
		if c.panicAt {
			panic("ownership test panic")
		}
		return c.failure
	}
	return nil
}
func TestIssuerMappingModel(t *testing.T) {
	var nilModel *Model
	nilModel.Clear()
	if !reflect.DeepEqual(nilModel.Summary(), Summary{}) {
		t.Fatal("nil")
	}
	var zero Model
	zero.Clear()
	if !reflect.DeepEqual(zero.Summary(), Summary{}) {
		t.Fatal("zero")
	}
	var s Summary
	fillOwnedTest(reflect.ValueOf(&s).Elem(), time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC))
	o := &owner{summary: ownTestSummary(s), valid: true}
	other := &owner{summary: ownTestSummary(s), valid: true}
	defer other.clear()
	m := &Model{owned: o}
	copied := *m
	first, second := m.Summary(), m.Summary()
	expected := ownedTestClone(t, second)
	first.Evidence.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links[0].Rotation.Payload[0] = 9
	reflect.ValueOf(first.Evidence.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage).Index(0).SetZero()
	if !reflect.DeepEqual(second, expected) || !reflect.DeepEqual(m.Summary(), expected) {
		t.Fatal("snapshot alias")
	}
	views := ownedTestViews(reflect.ValueOf(o.summary))
	copied.Clear()
	m.Clear()
	assertOwnedTestZero(t, views)
	if !reflect.DeepEqual(m.Summary(), Summary{}) || !reflect.DeepEqual(copied.Summary(), Summary{}) || !reflect.DeepEqual(second, expected) || !other.valid {
		t.Fatal("shared lifecycle")
	}
}
func ownTestSummary(s Summary) Summary { s.Evidence = ownEvidence(s.Evidence); return s }
func TestIssuerMappingNormalizeTime(t *testing.T) {
	want := time.Date(2020, 1, 1, 1, 0, 0, 123, time.UTC)
	src := time.Date(2020, 1, 1, 5, 0, 0, 123, time.FixedZone("offset", 4*3600))
	if normalizeTime(src) != want || normalizeTime(time.Time{}) != (time.Time{}) || copyRaw([]byte{}) != nil || copySlice([]int{}) != nil {
		t.Fatal("normalize")
	}
}
