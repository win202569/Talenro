package serving

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func desiredTestID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("12345678-1234-4234-8234-%012x", n))
}
func desiredTestPublish(n int, version, sequence uint64) desiredPublishProjection {
	p := []byte{byte(n), 1}
	e := []byte{byte(n), 2}
	return desiredPublishProjection{PublishID: desiredTestID(n), AuthorityOperationID: desiredTestID(n + 1000), Version: version, BaseRootVersion: version - 1, Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: sequence}, Payload: p, PayloadDigest: sha256.Sum256(p), PublishedEnvelope: e, EnvelopeDigest: sha256.Sum256(e)}
}
func desiredProjectionFixture(t *testing.T, rotations int) (uuid.UUID, uint64, desiredProjectionSnapshot) {
	t.Helper()
	node := desiredTestID(100)
	root := desiredTestPublish(1, 1, 9)
	var chain []desiredPublishProjection
	for i := 0; i < rotations; i++ {
		chain = append(chain, desiredTestPublish(i+2, uint64(i+2), uint64(i+10)))
		root = desiredTestPublish(i+2, uint64(i+2), uint64(i+10))
	}
	meta := desiredTestPublish(200, 4, 100)
	meta.BaseRootVersion = root.Version
	meta.BaseMetadataVersion = 3
	issued := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	payload := []byte{7, 8, 9}
	d := desiredArtifactProjection{NodeID: node, Generation: 7, SigningID: desiredTestID(300), AuthorityOperationID: desiredTestID(301), Authority: contracts.AuthorityVersion{Epoch: 19, Sequence: 101}, InventoryVersion: 8, ResourceEnvelopeVersion: 9, ResourceEnvelopeDigest: contracts.Digest{3}, RootPublishID: root.PublishID, RootVersion: root.Version, MetadataPublishID: meta.PublishID, MetadataVersion: 4, SigningKeyID: contracts.Digest{4}, Payload: payload, PayloadDigest: sha256.Sum256(payload), Signature: [64]byte{5}, IssuedAt: issued, EffectiveDeadline: issued.Add(time.Hour), ValidUntil: issued.Add(24 * time.Hour)}
	return node, 19, desiredProjectionSnapshot{NodeID: node, ActiveDesiredGeneration: 7, ActiveRootPublishID: root.PublishID, ActiveRootVersion: root.Version, ActiveMetadataPublishID: meta.PublishID, ActiveMetadataVersion: 4, ChainBaseRootVersion: 1, RootRotations: chain, ActiveRoot: root, ActiveMetadata: meta, Desired: d}
}
func checkDesiredResult(t *testing.T, node uuid.UUID, epoch uint64, s desiredProjectionSnapshot, want error) {
	t.Helper()
	got, e := newDesiredProjection(node, epoch, s)
	if e != want {
		t.Fatalf("error=%v want %v", e, want)
	}
	if want != nil && !reflect.DeepEqual(got, desiredProjection{}) {
		t.Fatal("failure returned material")
	}
	if want == nil && !reflect.DeepEqual(got.Snapshot(), s) {
		t.Fatal("projection changed")
	}
}
func TestDesiredProjectionValid(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			node, e, s := desiredProjectionFixture(t, n)
			checkDesiredResult(t, node, e, s, nil)
		})
	}
}
func TestDesiredProjectionOpaqueBoundary(t *testing.T) {
	node, e, s := desiredProjectionFixture(t, 1)
	// Non-JSON envelopes and an arbitrary nonzero signature are intentionally
	// valid STRUCTURE, never evidence of canonical encoding or authentication.
	checkDesiredResult(t, node, e, s, nil)
}

// Walk the fixture's value fields to reject malformed IDs, coordinates and
// hashes at every level, including intermediate roots. This is not a source
// text check: every mutation is passed through the real constructor.
func desiredLeafPaths(v reflect.Value, prefix []int) [][]int {
	if v.Type() == reflect.TypeOf(uuid.UUID{}) || v.Type() == reflect.TypeOf(contracts.Digest{}) || v.Type() == reflect.TypeOf([64]byte{}) || v.Type() == reflect.TypeOf(time.Time{}) || v.Kind() == reflect.Uint64 {
		return [][]int{prefix}
	}
	var out [][]int
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			out = append(out, desiredLeafPaths(v.Field(i), append(append([]int(nil), prefix...), i))...)
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Struct {
			for i := 0; i < v.Len(); i++ {
				out = append(out, desiredLeafPaths(v.Index(i), append(append([]int(nil), prefix...), i))...)
			}
		}
	}
	return out
}
func desiredField(s *desiredProjectionSnapshot, path []int) (reflect.Value, string) {
	v := reflect.ValueOf(s).Elem()
	name := ""
	for _, i := range path {
		if v.Kind() == reflect.Struct {
			name += "/" + v.Type().Field(i).Name
			v = v.Field(i)
		} else {
			name += fmt.Sprintf("[%d]", i)
			v = v.Index(i)
		}
	}
	return v, name
}
func TestDesiredProjectionReject(t *testing.T) {
	node, e, s := desiredProjectionFixture(t, 3)
	for _, bad := range []uuid.UUID{uuid.Nil, {1}} {
		checkDesiredResult(t, bad, e, s, authority.ErrInvalidArgument)
	}
	for _, bad := range []uint64{0, math.MaxInt64 + 1} {
		checkDesiredResult(t, node, bad, s, authority.ErrInvalidArgument)
	}
	for _, path := range desiredLeafPaths(reflect.ValueOf(s), nil) {
		_, name := desiredField(&s, path)
		for _, mode := range []string{"zero", "invalid"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				n, e, f := desiredProjectionFixture(t, 3)
				v, _ := desiredField(&f, path)
				switch v.Type() {
				case reflect.TypeOf(uuid.UUID{}):
					if mode == "zero" {
						v.Set(reflect.ValueOf(uuid.Nil))
					} else {
						v.Set(reflect.ValueOf(uuid.UUID{1}))
					}
				case reflect.TypeOf(contracts.Digest{}):
					if mode == "zero" {
						v.SetZero()
					} else {
						d := v.Interface().(contracts.Digest)
						d[0] ^= 255
						v.Set(reflect.ValueOf(d))
						if name == "/Desired/SigningKeyID" || name == "/Desired/ResourceEnvelopeDigest" {
							return
						}
					}
				case reflect.TypeOf([64]byte{}):
					if mode != "zero" {
						return
					}
					v.SetZero()
				case reflect.TypeOf(time.Time{}):
					if mode != "zero" {
						return
					}
					v.SetZero()
				default:
					if mode == "zero" {
						if len(name) >= 15 && (v.Uint() == 0 || name[len(name)-15:] == "BaseRootVersion") {
							return
						}
						if bytes.HasSuffix([]byte(name), []byte("BaseMetadataVersion")) {
							return
						}
						v.SetUint(0)
					} else {
						v.SetUint(math.MaxInt64 + 1)
					}
				}
				checkDesiredResult(t, n, e, f, authority.ErrConflict)
			})
		}
	}
	cases := map[string]func(*desiredProjectionSnapshot){
		"node":                   func(s *desiredProjectionSnapshot) { s.NodeID = desiredTestID(999) },
		"desired node":           func(s *desiredProjectionSnapshot) { s.Desired.NodeID = desiredTestID(999) },
		"generation":             func(s *desiredProjectionSnapshot) { s.Desired.Generation++ },
		"root pointer":           func(s *desiredProjectionSnapshot) { s.ActiveRootPublishID = desiredTestID(999) },
		"metadata pointer":       func(s *desiredProjectionSnapshot) { s.ActiveMetadataPublishID = desiredTestID(999) },
		"desired root":           func(s *desiredProjectionSnapshot) { s.Desired.RootPublishID = desiredTestID(999) },
		"desired metadata":       func(s *desiredProjectionSnapshot) { s.Desired.MetadataPublishID = desiredTestID(999) },
		"root version":           func(s *desiredProjectionSnapshot) { s.Desired.RootVersion++ },
		"metadata version":       func(s *desiredProjectionSnapshot) { s.Desired.MetadataVersion++ },
		"metadata root":          func(s *desiredProjectionSnapshot) { s.ActiveMetadata.BaseRootVersion++ },
		"metadata base":          func(s *desiredProjectionSnapshot) { s.ActiveMetadata.BaseMetadataVersion++ },
		"root base":              func(s *desiredProjectionSnapshot) { s.ActiveRoot.BaseRootVersion++ },
		"issued equals deadline": func(s *desiredProjectionSnapshot) { s.Desired.EffectiveDeadline = s.Desired.IssuedAt },
		"deadline after validity": func(s *desiredProjectionSnapshot) {
			s.Desired.EffectiveDeadline = s.Desired.ValidUntil.Add(time.Nanosecond)
		},
		"validity before issued": func(s *desiredProjectionSnapshot) { s.Desired.ValidUntil = s.Desired.IssuedAt.Add(-time.Second) },
		"over 24h":               func(s *desiredProjectionSnapshot) { s.Desired.ValidUntil = s.Desired.ValidUntil.Add(time.Nanosecond) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			n, e, f := desiredProjectionFixture(t, 3)
			mutate(&f)
			checkDesiredResult(t, n, e, f, authority.ErrConflict)
		})
	}
	for _, selectP := range []func(*desiredProjectionSnapshot) *contracts.AuthorityVersion{func(s *desiredProjectionSnapshot) *contracts.AuthorityVersion { return &s.ActiveRoot.Authority }, func(s *desiredProjectionSnapshot) *contracts.AuthorityVersion { return &s.ActiveMetadata.Authority }, func(s *desiredProjectionSnapshot) *contracts.AuthorityVersion { return &s.Desired.Authority }, func(s *desiredProjectionSnapshot) *contracts.AuthorityVersion { return &s.RootRotations[1].Authority }} {
		n, e, f := desiredProjectionFixture(t, 3)
		selectP(&f).Epoch = 20
		checkDesiredResult(t, n, e, f, authority.ErrConflict)
	}
}

func TestDesiredProjectionBounds(t *testing.T) {
	for _, where := range []string{"root", "metadata", "middle", "desired"} {
		for _, envelope := range []bool{false, true} {
			if where == "desired" && envelope {
				continue
			}
			limit := 65536
			if envelope {
				limit = 262144
			}
			for _, size := range []int{0, 1, limit, limit + 1} {
				t.Run(fmt.Sprintf("%s/%v/%d", where, envelope, size), func(t *testing.T) {
					n, e, s := desiredProjectionFixture(t, 3)
					b := bytes.Repeat([]byte{42}, size)
					d := sha256.Sum256(b)
					if where == "desired" {
						s.Desired.Payload = b
						s.Desired.PayloadDigest = d
					} else {
						p := &s.ActiveRoot
						switch where {
						case "metadata":
							p = &s.ActiveMetadata
						case "middle":
							p = &s.RootRotations[1]
						}
						if envelope {
							p.PublishedEnvelope = b
							p.EnvelopeDigest = d
						} else {
							p.Payload = b
							p.PayloadDigest = d
						}
						if where == "root" {
							s.RootRotations[2] = s.ActiveRoot
						}
					}
					var want error
					if size == 0 || size > limit {
						want = authority.ErrConflict
					}
					checkDesiredResult(t, n, e, s, want)
				})
			}
		}
	}
	n, e, s := desiredProjectionFixture(t, 0)
	s.Desired.EffectiveDeadline = s.Desired.ValidUntil
	checkDesiredResult(t, n, e, s, nil)
}

func TestDesiredProjectionChain(t *testing.T) {
	cases := map[string]func(*desiredProjectionSnapshot){
		"gap": func(s *desiredProjectionSnapshot) { s.RootRotations = s.RootRotations[1:] },
		"reverse": func(s *desiredProjectionSnapshot) {
			s.RootRotations[0], s.RootRotations[1] = s.RootRotations[1], s.RootRotations[0]
		},
		"duplicate ID":      func(s *desiredProjectionSnapshot) { s.RootRotations[1].PublishID = s.RootRotations[0].PublishID },
		"duplicate version": func(s *desiredProjectionSnapshot) { s.RootRotations[1].Version = s.RootRotations[0].Version },
		"equal sequence": func(s *desiredProjectionSnapshot) {
			s.RootRotations[1].Authority.Sequence = s.RootRotations[0].Authority.Sequence
		},
		"reversed sequence":  func(s *desiredProjectionSnapshot) { s.RootRotations[1].Authority.Sequence = 1 },
		"base mismatch":      func(s *desiredProjectionSnapshot) { s.ChainBaseRootVersion++ },
		"tail ID":            func(s *desiredProjectionSnapshot) { s.RootRotations[2].PublishID = desiredTestID(999) },
		"tail op":            func(s *desiredProjectionSnapshot) { s.RootRotations[2].AuthorityOperationID = desiredTestID(999) },
		"tail sequence":      func(s *desiredProjectionSnapshot) { s.RootRotations[2].Authority.Sequence++ },
		"tail base metadata": func(s *desiredProjectionSnapshot) { s.RootRotations[2].BaseMetadataVersion++ },
		"tail payload": func(s *desiredProjectionSnapshot) {
			p := &s.RootRotations[2]
			p.Payload = []byte{99}
			p.PayloadDigest = sha256.Sum256(p.Payload)
		},
		"tail envelope": func(s *desiredProjectionSnapshot) {
			p := &s.RootRotations[2]
			p.PublishedEnvelope = []byte{99}
			p.EnvelopeDigest = sha256.Sum256(p.PublishedEnvelope)
		},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			n, e, s := desiredProjectionFixture(t, 3)
			m(&s)
			checkDesiredResult(t, n, e, s, authority.ErrConflict)
		})
	}
	n, e, s := desiredProjectionFixture(t, 33)
	checkDesiredResult(t, n, e, s, authority.ErrConflict)
	n, e, s = desiredProjectionFixture(t, 0)
	s.ChainBaseRootVersion = 2
	checkDesiredResult(t, n, e, s, authority.ErrConflict)
	for _, count := range []int{0, 1} {
		n, e, s = desiredProjectionFixture(t, count)
		s.ChainBaseRootVersion = math.MaxInt64 - uint64(count)
		s.ActiveRootVersion = math.MaxInt64
		s.ActiveRoot.Version = math.MaxInt64
		s.ActiveRoot.BaseRootVersion = math.MaxInt64 - 1
		s.ActiveMetadata.BaseRootVersion = math.MaxInt64
		s.Desired.RootVersion = math.MaxInt64
		if count == 1 {
			s.RootRotations[0] = s.ActiveRoot
		}
		checkDesiredResult(t, n, e, s, nil)
		if count == 1 {
			s.ChainBaseRootVersion = math.MaxInt64
			checkDesiredResult(t, n, e, s, authority.ErrConflict)
		}
	}
}

func desiredTestBuffers(s *desiredProjectionSnapshot) [][]byte {
	out := [][]byte{s.ActiveRoot.Payload, s.ActiveRoot.PublishedEnvelope, s.ActiveMetadata.Payload, s.ActiveMetadata.PublishedEnvelope, s.Desired.Payload}
	for i := range s.RootRotations {
		out = append(out, s.RootRotations[i].Payload, s.RootRotations[i].PublishedEnvelope)
	}
	return out
}

func TestDesiredProjectionOwnership(t *testing.T) {
	n, e, input := desiredProjectionFixture(t, 3)
	_, _, want := desiredProjectionFixture(t, 3)
	// Caller deliberately shares the tail with the active-root buffers.
	input.RootRotations[2] = input.ActiveRoot
	p, err := newDesiredProjection(n, e, input)
	if err != nil {
		t.Fatal(err)
	}
	first := p.Snapshot()
	for _, b := range desiredTestBuffers(&input) {
		b[0] ^= 255
	}
	input.RootRotations[0].PublishID = uuid.Nil
	for _, b := range desiredTestBuffers(&first) {
		b[0] ^= 127
	}
	first.RootRotations[1].Authority.Sequence = 0
	first.Desired.Signature[0] = 0
	if !reflect.DeepEqual(p.Snapshot(), want) {
		t.Fatal("input/snapshot mutation reached owned projection")
	}
	// Shared caller buffers must not turn into shared internal mutable buffers.
	p.snapshot.RootRotations[2].Payload[0] ^= 1
	if !bytes.Equal(p.snapshot.ActiveRoot.Payload, want.ActiveRoot.Payload) {
		t.Fatal("owned tail aliases active root")
	}
}

func TestDesiredProjectionDiscard(t *testing.T) {
	n, e, input := desiredProjectionFixture(t, 3)
	_, _, want := desiredProjectionFixture(t, 3)
	p, err := newDesiredProjection(n, e, input)
	if err != nil {
		t.Fatal(err)
	}
	delivered := p.Snapshot()
	buffers := desiredTestBuffers(&p.snapshot)
	roots := p.snapshot.RootRotations
	discardDesiredProjection(&p)
	for _, b := range buffers {
		for _, v := range b {
			if v != 0 {
				t.Fatal("owned bytes retained")
			}
		}
	}
	for _, root := range roots {
		if !reflect.DeepEqual(root, desiredPublishProjection{}) {
			t.Fatal("root container retained material")
		}
	}
	if !reflect.DeepEqual(p, desiredProjection{}) {
		t.Fatal("projection retained material")
	}
	if !reflect.DeepEqual(input, want) || !reflect.DeepEqual(delivered, want) {
		t.Fatal("discard changed caller data")
	}
	discardDesiredProjection(&p)
	if !reflect.DeepEqual(p.Snapshot(), desiredProjectionSnapshot{}) {
		t.Fatal("zero snapshot populated")
	}
	// Rejected construction also leaves all caller bytes intact.
	input.Desired.RootVersion++
	_, _, want = desiredProjectionFixture(t, 3)
	want.Desired.RootVersion++
	checkDesiredResult(t, n, e, input, authority.ErrConflict)
	if !reflect.DeepEqual(input, want) {
		t.Fatal("rejection mutated caller data")
	}
}
