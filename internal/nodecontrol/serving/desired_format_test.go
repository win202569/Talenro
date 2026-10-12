package serving

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/store"
)

func formatArtifacts(f desiredFormat) []*stateformat.ParsedArtifact {
	return append([]*stateformat.ParsedArtifact{f.desired, f.root, f.metadata}, f.rotations...)
}
func assertFormatCleared(t *testing.T, artifacts []*stateformat.ParsedArtifact) {
	t.Helper()
	for _, a := range artifacts {
		if a != nil && (a.Payload() != nil || a.Envelope() != nil || a.Bindings() != (stateformat.Bindings{})) {
			t.Fatal("staged artifact retained")
		}
	}
}
func TestDesiredFormatBinding(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		p := desiredFormatFixture(t, n)
		before := p.Snapshot()
		f, e := bindDesiredFormat(t.Context(), p)
		if e != nil || f.desired == nil || f.root == nil || f.metadata == nil || len(f.rotations) != n {
			t.Fatal(n, e)
		}
		if f.desired.Bindings().NodeID != p.snapshot.NodeID || f.root.Bindings().RootVersion != uint64(n+1) {
			t.Fatal("coordinates")
		}
		discardDesiredFormat(&f)
		if !reflect.DeepEqual(before, p.snapshot) {
			t.Fatal("changed input")
		}
	}
	for _, tc := range []struct {
		field string
		value any
	}{
		{"node_id", desiredTestID(999).String()}, {"generation", "8"}, {"inventory_version", "9"}, {"node_resource_envelope_version", "10"}, {"node_resource_envelope_digest", fmt.Sprintf("%064x", 9)}, {"control_plane_authority_epoch", "20"}, {"authority_sequence", "102"}, {"issued_at", "2026-10-09T00:00:01Z"}, {"valid_until", "2026-10-09T23:59:59Z"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			p := desiredFormatFixture(t, 1)
			m := formatJSON(t, p.snapshot.Desired.Payload)
			m[tc.field] = tc.value
			p.snapshot.Desired.Payload = formatBytes(t, m)
			p.snapshot.Desired.PayloadDigest = sha256.Sum256(p.snapshot.Desired.Payload)
			assertFormatConflict(t, p)
		})
	}
	for _, which := range []string{"initial", "rotation", "metadata"} {
		for _, field := range []string{"control_plane_authority_epoch", "authority_sequence", "root_version", "metadata_version"} {
			t.Run(which+"/"+field, func(t *testing.T) {
				n := 1
				if which == "initial" {
					n = 0
				}
				p := desiredFormatFixture(t, n)
				target := &p.snapshot.ActiveRoot
				name := "root-rotation"
				if which == "initial" {
					name = "root-initial"
				}
				if which == "metadata" {
					target = &p.snapshot.ActiveMetadata
					name = "metadata"
				}
				m := formatJSON(t, target.Payload)
				if field == "metadata_version" && which != "metadata" {
					return
				}
				if field == "root_version" && which == "initial" {
					return
				}
				if which == "rotation" && field == "root_version" {
					m["previous_root_version"] = "2"
					m["new_root_set"].(map[string]any)["root_version"] = "3"
				} else {
					m[field] = "99"
					if which == "rotation" {
						m["new_root_set"].(map[string]any)[field] = "99"
					}
				}
				// Keep outer/body internally consistent while leaving projection values old.
				wireProjection := *target
				if field == "control_plane_authority_epoch" {
					wireProjection.Authority.Epoch = 99
				}
				if field == "authority_sequence" {
					wireProjection.Authority.Sequence = 99
				}
				if field == "root_version" {
					if which == "metadata" {
						wireProjection.BaseRootVersion = 99
					} else {
						wireProjection.Version = 3
						wireProjection.BaseRootVersion = 2
					}
				}
				setFormatPublish(t, &wireProjection, name, m)
				target.Payload = wireProjection.Payload
				target.PayloadDigest = wireProjection.PayloadDigest
				target.PublishedEnvelope = wireProjection.PublishedEnvelope
				target.EnvelopeDigest = wireProjection.EnvelopeDigest
				if which == "rotation" {
					p.snapshot.RootRotations[0] = cloneDesiredPublish(*target)
				}
				assertFormatConflict(t, p)
			})
		}
	}
	for _, p := range []desiredProjection{{}, desiredFormatFixture(t, 33)} {
		f, e := bindDesiredFormat(t.Context(), p)
		if e != authority.ErrConflict || !reflect.DeepEqual(f, desiredFormat{}) {
			t.Fatal("structure accepted", e)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, e := bindDesiredFormat(ctx, desiredProjection{}); e != authority.ErrCanceled {
		t.Fatal(e)
	}
	if _, e := bindDesiredFormat(nil, desiredProjection{}); e != authority.ErrAuthorityUnavailable {
		t.Fatal(e)
	}
}
func assertFormatConflict(t *testing.T, p desiredProjection) {
	t.Helper()
	before := p.Snapshot()
	checked, e := newDesiredProjection(p.snapshot.NodeID, 19, p.snapshot)
	if e != nil {
		t.Fatal("mutation must retain valid structural projection", e)
	}
	discardDesiredProjection(&checked)
	f, e := bindDesiredFormat(t.Context(), p)
	if e != authority.ErrConflict || !reflect.DeepEqual(f, desiredFormat{}) {
		t.Fatal("binding accepted", e)
	}
	if !reflect.DeepEqual(before, p.snapshot) {
		t.Fatal("failure mutated input")
	}
}
func TestDesiredFormatSplice(t *testing.T) {
	for _, mode := range []string{"other envelope", "root set only", "initial as rotation", "middle root"} {
		p := desiredFormatFixture(t, 3)
		switch mode {
		case "other envelope":
			m := formatJSON(t, p.snapshot.ActiveMetadata.Payload)
			m["metadata_version"] = "7"
			other := p.snapshot.ActiveMetadata
			setFormatPublish(t, &other, "metadata", m)
			p.snapshot.ActiveMetadata.PublishedEnvelope = other.PublishedEnvelope
			p.snapshot.ActiveMetadata.EnvelopeDigest = other.EnvelopeDigest
		case "root set only":
			m := formatJSON(t, p.snapshot.ActiveRoot.Payload)
			p.snapshot.ActiveRoot.Payload = formatBytes(t, m["new_root_set"])
			p.snapshot.ActiveRoot.PayloadDigest = sha256.Sum256(p.snapshot.ActiveRoot.Payload)
			p.snapshot.RootRotations[2] = cloneDesiredPublish(p.snapshot.ActiveRoot)
		case "initial as rotation":
			initial := desiredFormatFixture(t, 0).snapshot.ActiveRoot
			p.snapshot.RootRotations[0].Payload = initial.Payload
			p.snapshot.RootRotations[0].PayloadDigest = initial.PayloadDigest
			p.snapshot.RootRotations[0].PublishedEnvelope = initial.PublishedEnvelope
			p.snapshot.RootRotations[0].EnvelopeDigest = initial.EnvelopeDigest
		case "middle root":
			p.snapshot.RootRotations[1].Payload = bytes.Clone(p.snapshot.RootRotations[0].Payload)
			p.snapshot.RootRotations[1].PayloadDigest = p.snapshot.RootRotations[0].PayloadDigest
			p.snapshot.RootRotations[1].PublishedEnvelope = bytes.Clone(p.snapshot.RootRotations[0].PublishedEnvelope)
			p.snapshot.RootRotations[1].EnvelopeDigest = p.snapshot.RootRotations[0].EnvelopeDigest
		}
		assertFormatConflict(t, p)
	}
}

// Direct owner tests deliberately inject a panic; no claim that valid codecs panic.
func TestDesiredFormatLifetime(t *testing.T) {
	for _, mode := range []string{"success", "late failure", "cancel", "panic"} {
		t.Run(mode, func(t *testing.T) {
			p := desiredFormatFixture(t, 1)
			f, e := bindDesiredFormat(t.Context(), p)
			if e != nil {
				t.Fatal(e)
			}
			aliases := formatArtifacts(f)
			life := desiredFormatLifetime{staged: f}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var out desiredFormat
			var err error
			var pan any
			func() {
				defer func() { pan = recover() }()
				defer life.cleanup()
				switch mode {
				case "cancel":
					cancel()
				case "panic":
					panic("format owner panic")
				case "late failure":
					err = life.bindPublish(ctx, desiredPublishProjection{}, stateformat.Metadata, &life.staged.metadata)
					return
				}
				out, err = life.finish(ctx)
			}()
			life.cleanup()
			if mode == "success" {
				if err != nil || pan != nil || out.desired == nil || len(out.desired.Payload()) == 0 {
					t.Fatal("success discarded")
				}
				discardDesiredFormat(&out)
				return
			}
			if mode == "cancel" && err != authority.ErrCanceled || mode == "late failure" && err != authority.ErrConflict || mode == "panic" && pan != "format owner panic" {
				t.Fatal(mode, err, pan)
			}
			assertFormatCleared(t, aliases)
		})
	}
}
func TestDesiredFormatGuardComposition(t *testing.T) {
	for _, mode := range []string{"success", "reject", "cancel", "source panic"} {
		t.Run(mode, func(t *testing.T) {
			p := desiredFormatFixture(t, 1)
			before := p.Snapshot()
			rows := &desiredAdapterRows{values: desiredFormatRows(p)}
			db := &desiredAdapterDB{rows: rows}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var captured []*stateformat.ParsedArtifact
			source := boundReadFunc(func(c context.Context, query func(context.Context, store.DBTX) error) error {
				if e := query(c, db); e != nil {
					return e
				}
				switch mode {
				case "reject":
					return errors.New("source private detail")
				case "cancel":
					cancel()
				case "source panic":
					panic("format source panic")
				}
				return nil
			})
			var out desiredFormat
			var err error
			var pan any
			func() {
				defer func() { pan = recover() }()
				out, err = guardedRead(ctx, source, func(c context.Context, d store.DBTX) (desiredFormat, error) {
					projection, e := queryDesiredProjection(c, d, p.snapshot.NodeID, 19)
					if e != nil {
						return desiredFormat{}, e
					}
					defer discardDesiredProjection(&projection)
					f, e := bindDesiredFormat(c, projection)
					captured = formatArtifacts(f)
					return f, e
				}, discardDesiredFormat)
			}()
			assertDesiredAdapterClean(t, rows, db)
			if db.calls != 1 || !reflect.DeepEqual(before, p.snapshot) {
				t.Fatal("composition input/query")
			}
			if mode == "success" {
				if err != nil || pan != nil || out.desired == nil || len(out.desired.Payload()) == 0 {
					t.Fatal(err, pan)
				}
				discardDesiredFormat(&out)
				return
			}
			if mode == "reject" && err != authority.ErrAuthorityUnavailable || mode == "cancel" && err != authority.ErrCanceled || mode == "source panic" && pan != "format source panic" {
				t.Fatal(mode, err, pan)
			}
			if !reflect.DeepEqual(out, desiredFormat{}) {
				t.Fatal("partial result")
			}
			assertFormatCleared(t, captured)
		})
	}
}
