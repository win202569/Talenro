package serving

import (
	"crypto/sha256"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func certificateFixture(t *testing.T, agent bool) (CertificateLookup, CertificateSnapshot) {
	t.Helper()
	i := validLookupInput()
	der := []byte{1, 2, 3}
	i.LeafDERDigest = sha256.Sum256(der)
	l, err := NewCertificateLookup(i)
	if agent {
		l, err = NewAgentCertificateLookup(i, "nodes.example")
	}
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse("23456789-1234-4234-8234-123456789abc")
	s := CertificateSnapshot{Identity: i, CertificateID: id, LineageID: id, IdentityEpoch: 2,
		InventoryIdentityEpoch: 2, InventoryLineageID: id, InventoryVersion: 3,
		AuthorityOperationID: id, Authority: contracts.AuthorityVersion{Epoch: 4, Sequence: 5},
		Status: "active", ValidFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidUntil: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), LeafDER: der}
	if agent {
		s.ServerTrust = &ServerTrustFacts{Selector: AgentTrustSelector{Purpose: "agent_server", ListenerKind: "agent", TrustDomain: "nodes.example"}, AuthorityEpoch: 4, AuthorityOperationID: id, ServerCABundleHighWater: contracts.VersionedDigest{Version: 6, AuthoritySequence: 7, Digest: contracts.Digest{8}}}
	}
	return l, s
}

// These cases catch missing exact-match, completeness and range checks.
func TestCertificateFactsRejectInvalid(t *testing.T) {
	cases := map[string]func(*CertificateSnapshot){
		"node mismatch":       func(s *CertificateSnapshot) { s.Identity.NodeID[0]++ },
		"issuer mismatch":     func(s *CertificateSnapshot) { s.Identity.IssuerID = "other" },
		"serial mismatch":     func(s *CertificateSnapshot) { s.Identity.SerialBytes[0]++ },
		"leaf mismatch":       func(s *CertificateSnapshot) { s.Identity.LeafDERDigest[0]++ },
		"key mismatch":        func(s *CertificateSnapshot) { s.Identity.PublicKeyDigest[0]++ },
		"certificate ID":      func(s *CertificateSnapshot) { s.CertificateID = uuid.Nil },
		"lineage ID":          func(s *CertificateSnapshot) { s.LineageID = uuid.Nil },
		"lineage mismatch":    func(s *CertificateSnapshot) { s.InventoryLineageID[0]++ },
		"identity epoch zero": func(s *CertificateSnapshot) { s.IdentityEpoch = 0 },
		"identity epoch overflow": func(s *CertificateSnapshot) {
			s.IdentityEpoch = math.MaxUint64
			s.InventoryIdentityEpoch = math.MaxUint64
		},
		"identity epoch mismatch": func(s *CertificateSnapshot) { s.InventoryIdentityEpoch++ },
		"inventory zero":          func(s *CertificateSnapshot) { s.InventoryVersion = 0 },
		"inventory overflow":      func(s *CertificateSnapshot) { s.InventoryVersion = math.MaxUint64 },
		"operation ID":            func(s *CertificateSnapshot) { s.AuthorityOperationID = uuid.Nil },
		"authority epoch":         func(s *CertificateSnapshot) { s.Authority.Epoch = 0 },
		"authority sequence":      func(s *CertificateSnapshot) { s.Authority.Sequence = 0 },
		"authority overflow":      func(s *CertificateSnapshot) { s.Authority.Sequence = math.MaxUint64 },
		"revoked":                 func(s *CertificateSnapshot) { s.Status = "revoked" },
		"unknown status":          func(s *CertificateSnapshot) { s.Status = "ACTIVE" },
		"missing start":           func(s *CertificateSnapshot) { s.ValidFrom = time.Time{} },
		"missing end":             func(s *CertificateSnapshot) { s.ValidUntil = time.Time{} },
		"equal dates":             func(s *CertificateSnapshot) { s.ValidUntil = s.ValidFrom },
		"reversed dates":          func(s *CertificateSnapshot) { s.ValidUntil = s.ValidFrom.Add(-time.Second) },
		"empty DER":               func(s *CertificateSnapshot) { s.LeafDER = nil },
		"large DER":               func(s *CertificateSnapshot) { s.LeafDER = make([]byte, 65537) },
		"corrupt DER":             func(s *CertificateSnapshot) { s.LeafDER[0]++ },
		"missing trust":           func(s *CertificateSnapshot) { s.ServerTrust = nil },
		"purpose":                 func(s *CertificateSnapshot) { s.ServerTrust.Selector.Purpose = "operator_server" },
		"listener":                func(s *CertificateSnapshot) { s.ServerTrust.Selector.ListenerKind = "operator" },
		"domain":                  func(s *CertificateSnapshot) { s.ServerTrust.Selector.TrustDomain = "other.example" },
		"trust epoch":             func(s *CertificateSnapshot) { s.ServerTrust.AuthorityEpoch++ },
		"trust operation":         func(s *CertificateSnapshot) { s.ServerTrust.AuthorityOperationID = uuid.Nil },
		"trust version":           func(s *CertificateSnapshot) { s.ServerTrust.ServerCABundleHighWater.Version = 0 },
		"trust version overflow":  func(s *CertificateSnapshot) { s.ServerTrust.ServerCABundleHighWater.Version = math.MaxUint64 },
		"trust sequence":          func(s *CertificateSnapshot) { s.ServerTrust.ServerCABundleHighWater.AuthoritySequence = 0 },
		"trust sequence overflow": func(s *CertificateSnapshot) { s.ServerTrust.ServerCABundleHighWater.AuthoritySequence = math.MaxUint64 },
		"trust digest":            func(s *CertificateSnapshot) { s.ServerTrust.ServerCABundleHighWater.Digest = contracts.Digest{} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			l, s := certificateFixture(t, true)
			mutate(&s)
			got, err := newCertificateFacts(l, s)
			if err != authority.ErrConflict || !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatalf("expected zero facts/conflict, err=%v", err)
			}
		})
	}
}

func TestCertificateFactsAllowedStatusesAndOwnership(t *testing.T) {
	for _, agent := range []bool{false, true} {
		for _, status := range []string{"active", "recovery_pending", "recovery_limited"} {
			l, s := certificateFixture(t, agent)
			s.Status = status
			_, want := certificateFixture(t, agent)
			want.Status = status
			got, err := newCertificateFacts(l, s)
			if err != nil {
				t.Fatal(err)
			}
			s.LeafDER[0]++
			s.Identity.SerialBytes[0]++
			s.Identity.PublicKeyDigest[0]++
			if s.ServerTrust != nil {
				s.ServerTrust.ServerCABundleHighWater.Digest[0]++
			}
			out := got.Snapshot()
			out.LeafDER[0]++
			out.Identity.SerialBytes[0]++
			if out.ServerTrust != nil {
				out.ServerTrust.Selector.TrustDomain = "other.example"
			}
			if !reflect.DeepEqual(got.Snapshot(), want) {
				t.Fatal("caller mutation altered immutable facts")
			}
		}
	}
}

func TestCertificateFactsRejectInvalidLookupAndUnexpectedTrust(t *testing.T) {
	_, s := certificateFixture(t, true)
	got, err := newCertificateFacts(CertificateLookup{}, s)
	if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, CertificateFacts{}) {
		t.Fatal("invalid lookup accepted")
	}
	l, _ := NewCertificateLookup(s.Identity)
	got, err = newCertificateFacts(l, s)
	if err != authority.ErrConflict || !reflect.DeepEqual(got, CertificateFacts{}) {
		t.Fatal("non-agent trust injection accepted")
	}
	if !reflect.DeepEqual((CertificateFacts{}).Snapshot(), CertificateSnapshot{}) {
		t.Fatal("zero facts exposed material")
	}
}
