package statecontract

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"talenro.local/platform/internal/nodecontrol/contracts"
)

// Construct expected inputs independently, without any product cloning helper.
func ownershipFixture() ModelInput {
	in := pair()
	for _, x := range []*MainStateContentModel{&in.Previous, &in.Proposed} {
		x.Metadata.RevokedIDs = []contracts.Digest{d(70)}
		x.ServerTrust.Bootstrap.Deauthorized = []contracts.Digest{d(71)}
		x.ServerTrust.Agent.Deauthorized = []contracts.Digest{d(72)}
		x.Recovery = recoveryFixture()
		x.Recovery.Local = []LocalFaultBinding{{faultID(1), faultID(2)}}
		x.Recovery.Supervisor = []SupervisorFaultBinding{{faultID(3), faultID(4), d(73)}}
	}
	in.Proposed.Root.Mark = StreamMark{2, 2, ref(RootPayload, 90)}
	in.RootTransition = []RotationRef{{1, 2, 2, d(20), d(90), ref(RootRotation, 80)}}
	return in
}

func collectionPaths(v reflect.Value, prefix string) []string {
	var paths []string
	for i := 0; i < v.NumField(); i++ {
		name := prefix + v.Type().Field(i).Name
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Struct:
			paths = append(paths, collectionPaths(f, name+".")...)
		case reflect.Slice:
			paths = append(paths, name)
		}
	}
	return paths
}

func zeroFirst(in *ModelInput, path string) {
	v := reflect.ValueOf(in).Elem()
	for _, field := range strings.Split(path, ".") {
		v = v.FieldByName(field)
	}
	v.Index(0).Set(reflect.Zero(v.Type().Elem()))
}

func TestContractOwnershipEveryCollection(t *testing.T) {
	paths := collectionPaths(reflect.ValueOf(ownershipFixture()), "")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			in, expected := ownershipFixture(), ownershipFixture()
			v := reflect.ValueOf(in)
			for _, field := range strings.Split(path, ".") {
				v = v.FieldByName(field)
			}
			if v.Len() == 0 {
				t.Fatal("fixture leaves collection unexercised")
			}
			m, err := EvaluateModel(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			defer m.Clear()
			zeroFirst(&in, path)
			if !reflect.DeepEqual(m.input, expected) {
				t.Fatal("model borrowed caller collection")
			}
			if err := RecheckModel(context.Background(), m, expected); err != nil {
				t.Fatal("owned model changed", err)
			}
			mutated := ownershipFixture()
			zeroFirst(&mutated, path)
			m.Clear()
			if !reflect.DeepEqual(in, mutated) {
				t.Fatal("Clear modified caller collection")
			}
		})
	}
}

func TestContractNestedCoverageErrors(t *testing.T) {
	for _, side := range []string{"previous", "proposed"} {
		for _, field := range []string{"verified", "applied", "transition"} {
			for _, kind := range []string{"unknown", "invalid-tag", "absent-residual", "present-malformed"} {
				t.Run(side+"/"+field+"/"+kind, func(t *testing.T) {
					in := pair()
					x := &in.Proposed
					if side == "previous" {
						x = &in.Previous
					}
					want := ErrIncompleteCoverage
					coverage := Unknown
					switch kind {
					case "invalid-tag":
						coverage = Coverage(255)
					case "absent-residual":
						coverage = NeverEstablished
					case "present-malformed":
						coverage = Present
						want = ErrInvalidInput
					}
					if field == "transition" {
						x.Supervisor.Transition = TransitionRef{Coverage: coverage, Material: ref(TransitionCapsule, 81), PriorApplied: x.Supervisor.Applied.Desired, CandidateVerified: x.Supervisor.Verified.Desired}
						if kind == "present-malformed" {
							x.Supervisor.Transition.Material.Payload = contracts.Digest{}
						}
					} else {
						r := &x.Supervisor.Verified
						if field == "applied" {
							r = &x.Supervisor.Applied
						}
						r.Coverage = coverage
						if kind == "present-malformed" {
							r.Desired.Generation = 0
						}
					}
					m, err := EvaluateModel(context.Background(), in)
					if m != nil {
						m.Clear()
						t.Fatal("invalid coverage produced a model")
					}
					if !errors.Is(err, want) {
						t.Fatalf("got %v, want %v", err, want)
					}
				})
			}
		}
	}
	in := pair()
	dormant(&in.Previous)
	dormant(&in.Proposed)
	m, err := EvaluateModel(context.Background(), in)
	if err != nil {
		t.Fatal("legal absent records", err)
	}
	m.Clear()
}
