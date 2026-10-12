package statebinding

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecomposition"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"talenro.local/platform/internal/nodecontrol/statematerials"
)

// Every expected fact comes from JSON/frozen vectors, not a statebinding helper.
func TestBindingProjection(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			in := bindingFixture(t, n)
			checkBindingLayers(t, in)
			if err := bindProjection(context.Background(), in); err != nil {
				t.Fatal(err)
			}
		})
	}
	edits := []struct {
		name string
		edit func(*statecontract.MainStateContentModel)
		want error
	}{
		{"node", func(x *statecontract.MainStateContentModel) { x.Scope.NodeID = "22345678-1234-4234-8234-123456789abc" }, ErrBindingMismatch},
		{"epoch", func(x *statecontract.MainStateContentModel) { x.Scope.AuthorityEpoch++ }, ErrBindingMismatch},
		{"root_kind", func(x *statecontract.MainStateContentModel) { x.Root.Mark.Material.Kind = statecontract.CertificateDER }, ErrInvalidInput},
		{"root_p", func(x *statecontract.MainStateContentModel) { x.Root.Mark.Material.Payload[0] ^= 1 }, ErrMissingMaterial},
		{"root_e", func(x *statecontract.MainStateContentModel) { x.Root.Mark.Material.Envelope = d(99) }, ErrInvalidInput},
		{"root_version", func(x *statecontract.MainStateContentModel) { x.Root.Mark.Version++ }, ErrBindingMismatch},
		{"root_sequence", func(x *statecontract.MainStateContentModel) { x.Root.Mark.Sequence++ }, ErrBindingMismatch},
		{"root_threshold", func(x *statecontract.MainStateContentModel) { x.Root.Threshold++ }, ErrBindingMismatch},
		{"root_keys", func(x *statecontract.MainStateContentModel) { x.Root.KeyIDs[0][0] ^= 1 }, ErrBindingMismatch},
		{"meta_p", func(x *statecontract.MainStateContentModel) { x.Metadata.Mark.Material.Payload[0] ^= 1 }, ErrMissingMaterial},
		{"meta_e", func(x *statecontract.MainStateContentModel) { x.Metadata.Mark.Material.Envelope[0] ^= 1 }, ErrBindingMismatch},
		{"meta_version", func(x *statecontract.MainStateContentModel) { x.Metadata.Mark.Version++ }, ErrBindingMismatch},
		{"meta_sequence", func(x *statecontract.MainStateContentModel) { x.Metadata.Mark.Sequence++ }, ErrBindingMismatch},
		{"meta_root_version", func(x *statecontract.MainStateContentModel) { x.Metadata.RootVersion++ }, ErrBindingMismatch},
		{"meta_root_p", func(x *statecontract.MainStateContentModel) { x.Metadata.RootPayload[0] ^= 1 }, ErrBindingMismatch},
		{"meta_keys", func(x *statecontract.MainStateContentModel) { x.Metadata.KeyIDs[0][0] ^= 1 }, ErrBindingMismatch},
		{"meta_revoked", func(x *statecontract.MainStateContentModel) { x.Metadata.RevokedIDs[0][0] ^= 1 }, ErrBindingMismatch},
		{"meta_issued", func(x *statecontract.MainStateContentModel) { x.Metadata.Window.Issued++ }, ErrBindingMismatch},
		{"meta_until", func(x *statecontract.MainStateContentModel) { x.Metadata.Window.Until++ }, ErrBindingMismatch},
		{"desired_p", func(x *statecontract.MainStateContentModel) { x.Desired.Mark.Material.Payload[0] ^= 1 }, ErrMissingMaterial},
		{"desired_e", func(x *statecontract.MainStateContentModel) { x.Desired.Mark.Material.Envelope[0] ^= 1 }, ErrBindingMismatch},
		{"desired_generation", func(x *statecontract.MainStateContentModel) { x.Desired.Mark.Version++ }, ErrBindingMismatch},
		{"desired_sequence", func(x *statecontract.MainStateContentModel) { x.Desired.Mark.Sequence++ }, ErrBindingMismatch},
		{"inventory", func(x *statecontract.MainStateContentModel) { x.Desired.InventoryVersion++ }, ErrBindingMismatch},
		{"resource_version", func(x *statecontract.MainStateContentModel) { x.Desired.ResourceVersion++ }, ErrBindingMismatch},
		{"resource_p", func(x *statecontract.MainStateContentModel) { x.Desired.ResourcePayload[0] ^= 1 }, ErrBindingMismatch},
		{"desired_root_p", func(x *statecontract.MainStateContentModel) { x.Desired.RootPayload[0] ^= 1 }, ErrBindingMismatch},
		{"desired_meta_p", func(x *statecontract.MainStateContentModel) { x.Desired.MetadataPayload[0] ^= 1 }, ErrBindingMismatch},
		{"desired_key", func(x *statecontract.MainStateContentModel) { x.Desired.SigningKeyID[0] ^= 1 }, ErrBindingMismatch},
		{"desired_issued", func(x *statecontract.MainStateContentModel) { x.Desired.Window.Issued++ }, ErrBindingMismatch},
		{"desired_until", func(x *statecontract.MainStateContentModel) { x.Desired.Window.Until++ }, ErrBindingMismatch},
		{"desired_slots", func(x *statecontract.MainStateContentModel) { x.Desired.SlotIDs[0] = "different" }, ErrBindingMismatch},
	}
	for _, side := range []string{"previous", "proposed"} {
		for _, tc := range edits {
			t.Run(side+"/"+tc.name, func(t *testing.T) {
				in := currentMetadataFixture(t, true)
				checkBindingLayers(t, in)
				x := &in.Main.Previous
				if side == "proposed" {
					x = &in.Main.Proposed
				}
				tc.edit(x)
				if err := bindProjection(context.Background(), in); !errors.Is(err, tc.want) {
					t.Fatalf("got %v, want %v", err, tc.want)
				}
			})
		}
	}
	for _, side := range []string{"previous", "proposed"} {
		for _, kind := range []statematerials.MaterialKind{statematerials.RootSet, statematerials.Metadata, statematerials.Desired} {
			for _, field := range []string{"Epoch", "Sequence"} {
				t.Run(side+"/expected/"+fmt.Sprint(kind)+"/"+field, func(t *testing.T) {
					in := bindingFixture(t, 1)
					table := &in.Previous
					if side == "proposed" {
						table = &in.Proposed.Materials
					}
					selected := in.Proposed.Target.Desired
					switch kind {
					case statematerials.RootSet:
						selected = tailRef(in.Proposed)
					case statematerials.Metadata:
						selected = in.Proposed.Target.Metadata
					}
					for i := range table.Expected {
						if table.Expected[i].Ref == selected {
							v := reflect.ValueOf(&table.Expected[i]).Elem().FieldByName(field)
							v.SetUint(v.Uint() + 1)
							break
						}
					}
					if err := bindProjection(context.Background(), in); err != ErrBindingMismatch {
						t.Fatal(err)
					}
				})
			}
		}
	}
	for _, kind := range []statecontract.ArtifactKind{statecontract.RootPayload, statecontract.RootRotation, statecontract.Metadata, statecontract.DesiredSnapshot} {
		ref := statecontract.ArtifactRef{Kind: kind, Payload: d(1), Envelope: d(2)}
		if kind == statecontract.RootPayload {
			ref.Envelope = contracts.Digest{}
		}
		projected, err := projectRef(ref)
		if err != nil || artifact(projected) != ref {
			t.Fatal("closed kind mapping", kind, err)
		}
	}
	if _, err := projectRef(statecontract.ArtifactRef{Kind: statecontract.CertificateDER, Payload: d(1)}); err != ErrInvalidInput {
		t.Fatal(err)
	}
	t.Run("previous_ne", func(t *testing.T) {
		in := bindingFixture(t, 1)
		neverPrevious(&in)
		checkBindingLayers(t, in)
		if err := bindProjection(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	})
}

func TestBindingDualEnvelope(t *testing.T) {
	in := bindingFixture(t, 1)
	old := statecomposition.Input{Materials: in.Previous}
	ref := in.Proposed.Target.Metadata
	changed := replaceEnvelope(t, &old, ref, extraShare(t, material(&old, ref).Raw, "signatures"))
	in.Previous = old.Materials
	in.Main.Previous.Metadata.Mark.Material = artifact(changed)
	if changed.Payload != ref.Payload || changed.Envelope == ref.Envelope {
		t.Fatal("fixture not dual envelope")
	}
	checkBindingLayers(t, in)
	if err := bindProjection(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	in.Main.Previous.Metadata.Mark.Material.Envelope = ref.Envelope
	if err := bindProjection(context.Background(), in); err != ErrBindingMismatch {
		t.Fatal("borrowed new E", err)
	}
}

func TestBindingPreviousProjectionMismatch(t *testing.T) {
	in := bindingFixture(t, 1)
	in.Main.Previous.Root.KeyIDs = []contracts.Digest{d(99)}
	checkBindingLayers(t, in)
	if err := bindProjection(context.Background(), in); err != ErrBindingMismatch {
		t.Fatal(err)
	}
}

func TestBindingTransitions(t *testing.T) {
	base := separatedFixture(t)
	checkBindingLayers(t, base)
	if err := bindTransitions(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"Sequence", "PreviousVersion", "NewVersion", "OldRootPayload", "NewRootPayload", "Material.Envelope"} {
		t.Run(field, func(t *testing.T) {
			in := copyBindingInput(t, base)
			v := inputField(reflect.ValueOf(&in.Main.RootTransition[0]).Elem(), field)
			if v.Kind() == reflect.Array {
				v.Index(0).SetUint(v.Index(0).Uint() ^ 1)
			} else {
				v.SetUint(v.Uint() + 1)
			}
			if err := bindTransitions(context.Background(), in); err != ErrBindingMismatch {
				t.Fatal(err)
			}
		})
	}
	for _, side := range []string{"Previous", "Proposed"} {
		for _, field := range []string{"Version", "Sequence", "Material.Payload"} {
			t.Run(side+"/endpoint/"+field, func(t *testing.T) {
				in := copyBindingInput(t, base)
				v := inputField(reflect.ValueOf(&in.Main).Elem(), side+".Root.Mark."+field)
				if v.Kind() == reflect.Array {
					v.Index(0).SetUint(v.Index(0).Uint() ^ 1)
				} else {
					v.SetUint(v.Uint() + 1)
				}
				want := ErrBindingMismatch
				if side == "Previous" && field == "Material.Payload" {
					want = ErrMissingMaterial
				}
				if err := bindTransitions(context.Background(), in); err != want {
					t.Fatal(err)
				}
			})
		}
	}
	in := copyBindingInput(t, base)
	in.Main.RootTransition = nil
	if err := bindTransitions(context.Background(), in); err != ErrBindingMismatch {
		t.Fatal("missing current chain", err)
	}
}

func TestBindingTarget(t *testing.T) {
	base := bindingFixture(t, 1)
	proof := checkBindingLayers(t, base)
	if err := bindTarget(context.Background(), base, proof); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"Status", "Scope.NodeID", "Scope.Epoch", "Subset.Status", "Subset.Scope.NodeID", "Subset.Scope.Epoch", "Anchor.PayloadDigest", "Anchor.EnvelopeDigest", "Target.AnchorRoot.Payload", "Target.Metadata.Payload", "Target.Metadata.Envelope", "Target.Desired.Payload", "Target.Desired.Envelope", "FinalRoot.Payload", "FinalRoot.Envelope", "Coordinates.ExpectedNodeID", "Coordinates.ExpectedEpoch", "Coordinates.FinalRootVersion", "Coordinates.MetadataVersion", "Coordinates.DesiredGeneration", "Coordinates.DesiredSequence", "Coordinates.DesiredKeyID"} {
		t.Run(field, func(t *testing.T) {
			changed := proof
			changed.Target.Links = append([]statecomposition.RotationLink(nil), proof.Target.Links...)
			v := inputField(reflect.ValueOf(&changed).Elem(), field)
			switch v.Kind() {
			case reflect.String:
				v.SetString("different")
			case reflect.Array:
				v.Index(0).SetUint(v.Index(0).Uint() ^ 1)
			default:
				v.SetUint(v.Uint() + 1)
			}
			if err := bindTarget(context.Background(), base, changed); err != ErrBindingMismatch {
				t.Fatal(err)
			}
		})
	}
	changed := proof
	changed.Target.Links = append([]statecomposition.RotationLink(nil), proof.Target.Links...)
	changed.Target.Links[0].Rotation.Envelope[0] ^= 1
	if err := bindTarget(context.Background(), base, changed); err != ErrBindingMismatch {
		t.Fatal("summary links", err)
	}
	in := copyBindingInput(t, base)
	in.Proposed.Target.Metadata.Payload[0] ^= 1
	if err := bindTarget(context.Background(), in, proof); err != ErrBindingMismatch {
		t.Fatal("authorizer", err)
	}
	in = copyBindingInput(t, base)
	in.Proposed.Target.Desired.Kind = statematerials.RootSet
	if err := bindTarget(context.Background(), in, proof); err != ErrInvalidInput {
		t.Fatal("desired kind", err)
	}
	in = copyBindingInput(t, base)
	in.Main.Proposed.Desired.Header.Coverage = statecontract.NeverEstablished
	if err := bindTarget(context.Background(), in, proof); err != ErrInvalidInput {
		t.Fatal(err)
	}
}

func TestBindingSeparatedChains(t *testing.T) {
	in := separatedFixture(t)
	proof := checkBindingLayers(t, in)
	if in.Main.Proposed.Desired.MetadataPayload == in.Main.Proposed.Metadata.Mark.Material.Payload || proof.FinalRoot.Payload == in.Main.Proposed.Root.Mark.Material.Payload {
		t.Fatal("fixture conflated current/original")
	}
	if err := bindProjection(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if err := bindTransitions(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if err := bindTarget(context.Background(), in, proof); err != nil {
		t.Fatal(err)
	}
	in.Proposed.Target.Metadata = statematerials.MaterialRef{Kind: statematerials.Metadata, Payload: in.Main.Proposed.Metadata.Mark.Material.Payload, Envelope: in.Main.Proposed.Metadata.Mark.Material.Envelope}
	if err := bindTarget(context.Background(), in, proof); err != ErrBindingMismatch {
		t.Fatal("latest authorizer substituted", err)
	}
}
