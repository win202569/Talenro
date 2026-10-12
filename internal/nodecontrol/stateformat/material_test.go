package stateformat

import (
	"bytes"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
	"time"
)

func TestMaterialKindAndOwnership(t *testing.T) {
	for name, kind := range formatKinds {
		t.Run(name, func(t *testing.T) {
			a, err := ParseEnvelope(kind, fixture(t, name+".envelope"))
			if err != nil {
				t.Fatal(err)
			}
			defer a.Clear()
			r, re := a.RootMaterial()
			m, me := a.MetadataMaterial()
			s, se := a.SignatureMaterial()
			defer r.Clear()
			defer m.Clear()
			defer s.Clear()
			if (re == nil) != (kind == InitialRoot || kind == RootRotation) || (me == nil) != (kind == Metadata) || (se == nil) != (kind == Metadata || kind == RootRotation) {
				t.Fatal("kind gate", re, me, se)
			}
			if re == nil {
				if r.Version != a.Bindings().RootVersion || r.Threshold != 1 || len(r.Keys) != 1 || r.Keys[0].PublicKey != ([32]byte{1}) {
					t.Fatal("root material", r)
				}
				want := r.Keys[0]
				r.Keys[0].PublicKey[0] ^= 1
				again, _ := a.RootMaterial()
				if again.Keys[0] != want {
					t.Fatal("root alias")
				}
				again.Clear()
			}
			if me == nil {
				if m.Version != 6 || m.RootVersion != 2 || len(m.Keys) != 1 || m.Keys[0].Status != "active" || m.Keys[0].HasRevocation || m.Keys[0].HasIncident || m.Keys[0].NotBefore != time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC) {
					t.Fatal("metadata material", m)
				}
				want := m.Keys[0]
				m.Keys[0].PublicKey[0] ^= 1
				again, _ := a.MetadataMaterial()
				if again.Keys[0] != want {
					t.Fatal("metadata alias")
				}
				again.Clear()
			}
			if se == nil {
				var group []SignatureShare
				if kind == Metadata {
					group = s.Metadata
				} else {
					group = s.Current
				}
				if len(group) > 0 {
					want := group[0]
					group[0].Signature[0] ^= 1
					again, _ := a.SignatureMaterial()
					other := again.Metadata
					if kind == RootRotation {
						other = again.Current
					}
					if other[0] != want {
						t.Fatal("share alias")
					}
					again.Clear()
				}
			}
			a.Clear()
			for _, empty := range []*ParsedArtifact{nil, {}, a} {
				if _, e := empty.RootMaterial(); e != ErrInvalidFormat {
					t.Fatal("empty root")
				}
				if _, e := empty.MetadataMaterial(); e != ErrInvalidFormat {
					t.Fatal("empty metadata")
				}
				if _, e := empty.SignatureMaterial(); e != ErrInvalidFormat {
					t.Fatal("empty shares")
				}
			}
		})
	}
	for name, kind := range map[string]Kind{"metadata": Metadata, "root-rotation": RootRotation} {
		obj := object(t, fixture(t, name+".envelope"))
		if kind == Metadata {
			obj["signatures"] = []any{}
		} else {
			obj["current_signatures"] = []any{}
			obj["new_signatures"] = []any{}
		}
		a, e := ParseEnvelope(kind, encoded(t, obj))
		if e != nil {
			t.Fatal(e)
		}
		s, e := a.SignatureMaterial()
		if e != nil || len(s.Current)+len(s.New)+len(s.Metadata) != 0 {
			t.Fatal("valid empty shares", e)
		}
		s.Clear()
		a.Clear()
	}
}

func TestMaterialFailureCleanup(t *testing.T) {
	r := RootMaterial{Keys: []PublicKeyMaterial{{PublicKey: [32]byte{1}}}}
	m := MetadataMaterial{Keys: []MetadataKeyMaterial{{Status: "revoked"}}, RevokedIDs: []contracts.Digest{{1}}}
	s := SignatureMaterial{Current: []SignatureShare{{Signature: [64]byte{1}}}, New: []SignatureShare{{Signature: [64]byte{2}}}, Metadata: []SignatureShare{{Signature: [64]byte{3}}}}
	rk, mk, ids, cs, ns, ms := r.Keys, m.Keys, m.RevokedIDs, s.Current, s.New, s.Metadata
	func() { defer r.Clear(); defer m.Clear(); defer s.Clear() }()
	if rk[0] != (PublicKeyMaterial{}) || mk[0] != (MetadataKeyMaterial{}) || ids[0] != (contracts.Digest{}) || cs[0] != (SignatureShare{}) || ns[0] != (SignatureShare{}) || ms[0] != (SignatureShare{}) {
		t.Fatal("owned alias not cleared")
	}
	r.Clear()
	m.Clear()
	s.Clear()
	for _, name := range []string{"root-initial", "root-rotation", "metadata"} {
		a, e := ParseEnvelope(formatKinds[name], fixture(t, name+".envelope"))
		if e != nil {
			t.Fatal(e)
		}
		rootKeys, metaKeys, shares := a.root.Keys, a.metadata.Keys, a.signatures.Metadata
		a.Clear()
		for _, v := range rootKeys {
			if v != (PublicKeyMaterial{}) {
				t.Fatal("artifact root clear")
			}
		}
		for _, v := range metaKeys {
			if v != (MetadataKeyMaterial{}) {
				t.Fatal("artifact metadata clear")
			}
		}
		for _, v := range shares {
			if v != (SignatureShare{}) {
				t.Fatal("artifact share clear")
			}
		}
	}
}

func TestCanonicalOwnedBufferClear(t *testing.T) {
	for _, s := range []string{"same", "other"} {
		body := []byte("same")
		canonical := []byte(s)
		if equalCanonicalAndClear(body, canonical) != (s == "same") {
			t.Fatal("comparison")
		}
		if string(body) != "same" || !bytes.Equal(canonical, make([]byte, len(canonical))) {
			t.Fatal("owned canonical not cleared")
		}
	}
}

func TestMaterialNonemptySharesOwnership(t *testing.T) {
	for name, kind := range map[string]Kind{"metadata": Metadata, "root-rotation": RootRotation} {
		obj := object(t, fixture(t, name+".envelope"))
		share := map[string]any{"key_id": "01d0fabd251fcbbe2b93b4b927b26ad2a1a99077152e45ded1e678afa45dbec5", "signature": "AgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}
		if kind == Metadata {
			obj["signatures"] = []any{share}
		} else {
			obj["current_signatures"] = []any{share}
			obj["new_signatures"] = []any{share}
		}
		a, e := ParseEnvelope(kind, encoded(t, obj))
		if e != nil {
			t.Fatal(e)
		}
		s, e := a.SignatureMaterial()
		if e != nil {
			t.Fatal(e)
		}
		for index, group := range [][]SignatureShare{s.Current, s.New, s.Metadata} {
			if len(group) == 0 {
				continue
			}
			if group[0].Signature != ([64]byte{2}) {
				t.Fatal("signature material")
			}
			group[0].Signature[0] ^= 1
			again, e := a.SignatureMaterial()
			if e != nil {
				t.Fatal(e)
			}
			other := [][]SignatureShare{again.Current, again.New, again.Metadata}[index]
			if other[0].Signature != ([64]byte{2}) {
				t.Fatal("share aliases artifact")
			}
			again.Clear()
		}
		backing := [][]SignatureShare{a.signatures.Current, a.signatures.New, a.signatures.Metadata}
		a.Clear()
		s.Clear()
		for _, group := range backing {
			for _, v := range group {
				if v != (SignatureShare{}) {
					t.Fatal("share owner not cleared")
				}
			}
		}
	}
}
func TestCanonicalExactSizeLimit(t *testing.T) {
	for _, limit := range []int{65536, 262144} {
		for _, extra := range []int{0, 1} {
			body := []byte("\"" + strings.Repeat("a", limit-2+extra) + "\"")
			want := error(nil)
			if extra == 1 {
				want = ErrInvalidFormat
			}
			if got := strictCanonical(body, limit, wireString); got != want {
				t.Fatalf("limit %d extra %d: %v", limit, extra, got)
			}
		}
	}
}
func TestMaterialRevocationFields(t *testing.T) {
	p := object(t, fixture(t, "metadata.payload"))
	keys := p["keys"].([]any)
	k := object(t, encoded(t, keys[0]))
	k["status"] = "revoked"
	k["revoked_at"] = "2026-10-09T00:00:00Z"
	k["revocation_reason"] = "security_incident"
	k["incident_id"] = "11111111-1111-4111-8111-111111111111"
	// Keep an active covering key and add a distinct revoked key, sorted by ID.
	k["public_key"] = "AgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	k["key_id"] = "5778f985db754c6628691f56fadae50c65fddbe8eb2e93039633fefa05d45e31"
	p["keys"] = []any{keys[0], k}
	p["revoked_key_ids"] = []any{k["key_id"]}
	a, e := ParseEnvelope(Metadata, rewrap(t, "metadata", encoded(t, p)))
	if e != nil {
		t.Fatal(e)
	}
	defer a.Clear()
	m, e := a.MetadataMaterial()
	if e != nil {
		t.Fatal(e)
	}
	defer m.Clear()
	if len(m.Keys) != 2 || !m.Keys[1].HasRevocation || !m.Keys[1].HasIncident || m.Keys[1].RevocationReason != "security_incident" || m.Keys[1].IncidentID.String() != "11111111-1111-4111-8111-111111111111" || len(m.RevokedIDs) != 1 {
		t.Fatal("revocation lost")
	}
	want := m.RevokedIDs[0]
	m.RevokedIDs[0][0] ^= 1
	again, _ := a.MetadataMaterial()
	defer again.Clear()
	if !reflect.DeepEqual(again.RevokedIDs, []contracts.Digest{want}) {
		t.Fatal("ledger alias")
	}
}
