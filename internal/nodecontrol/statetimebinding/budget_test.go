package statetimebinding

import (
	"context"
	"errors"
	"talenro.local/platform/internal/nodecontrol/statematerials"
	"testing"
	"time"
)

func TestTimeBindingBudgetAllBranches(t *testing.T) {
	for _, path := range []string{"Time.MetadataEnvelope", "Time.AttestationEnvelope"} {
		budgetSlice(t, path, 262144)
	}
	raw := make([]byte, 262144)
	in := budgetInput()
	in.Time.MetadataEnvelope = raw
	in.Time.AttestationEnvelope = raw
	if e := checkBudget(in); e != nil {
		t.Fatal(e)
	}
	in.Time.MetadataEnvelope = make([]byte, 262145)
	if !errors.Is(checkBudget(in), ErrLimitExceeded) {
		t.Fatal("time bound")
	}
	// All 19 L0 collections are exercised by the inherited explicit budget matrix.
	TestTimeBindingBudget(t)
}
func TestTimeBindingTimeNormalization(t *testing.T) {
	in := budgetInput()
	now := time.Date(2020, 1, 1, 1, 2, 3, 123456789, time.FixedZone("offset", 3600))
	in.Time.Time.Now = now
	in.Time.Time.Floor = now
	got, e := ownInput(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	defer clearInput(&got)
	if got.Time.Time.Now.Location() != time.UTC || got.Time.Time.Now.Nanosecond() != 123456789 || !got.Time.Time.Now.Equal(now) {
		t.Fatal("time normalization")
	}
}

func TestTimeBindingCombinedOccurrenceBudget(t *testing.T) {
	in := budgetInput()
	raw := make([]byte, 262144)
	for _, tbl := range []*statematerials.Input{&in.Policy.Binding.Previous, &in.Policy.Binding.Proposed.Materials} {
		tbl.Materials = make([]statematerials.Material, 64)
		for i := range tbl.Materials {
			tbl.Materials[i] = statematerials.Material{Ref: statematerials.MaterialRef{Kind: statematerials.Metadata}, Raw: raw}
		}
	}
	in.Policy.Binding.Proposed.Anchor.Raw = raw
	batch := &in.Policy.Policy.Batch
	batch.AnchorEnvelope = raw
	batch.ActiveRootEnvelope = raw
	batch.MetadataEnvelope = raw
	batch.DesiredEnvelope = raw
	batch.Rotations = make([][]byte, 32)
	for i := range batch.Rotations {
		batch.Rotations[i] = raw
	}
	in.Time.MetadataEnvelope = raw
	in.Time.AttestationEnvelope = raw
	if checkBudget(in) != nil {
		t.Fatal("exact combined 32 MiB + 39 appearances")
	}
	in.Time.AttestationEnvelope = make([]byte, 262145)
	if checkBudget(in) != ErrLimitExceeded {
		t.Fatal("combined +1")
	}
}
