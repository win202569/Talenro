package serving

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

func certificateRowFixture(t *testing.T, agent bool) (CertificateLookup, store.ListExactCertificateProjectionRow, CertificateSnapshot) {
	t.Helper()
	l, s := certificateFixture(t, agent)
	r := store.ListExactCertificateProjectionRow{
		NodeID: s.Identity.NodeID, IssuerID: s.Identity.IssuerID, SerialBytes: append([]byte(nil), s.Identity.SerialBytes...),
		LeafDerSha256: append([]byte(nil), s.Identity.LeafDERDigest[:]...), PublicKeySha256: append([]byte(nil), s.Identity.PublicKeyDigest[:]...),
		CertificateID: s.CertificateID, LineageID: s.LineageID, IdentityEpoch: 2, InventoryIdentityEpoch: 2,
		InventoryLineageID: uuid.NullUUID{UUID: s.LineageID, Valid: true}, InventoryVersion: 3,
		AuthorityOperationID: s.AuthorityOperationID, AuthorityEpoch: 4, AuthoritySequence: 5, Status: "active",
		ValidFrom: pgtype.Timestamptz{Time: s.ValidFrom, Valid: true}, ValidUntil: pgtype.Timestamptz{Time: s.ValidUntil, Valid: true},
		LeafDer: append([]byte(nil), s.LeafDER...),
	}
	if agent {
		r.HasServerTrust = true
		r.TrustPurpose = "agent_server"
		r.TrustListenerKind = "agent"
		r.TrustDomain = "nodes.example"
		r.TrustOperationID = s.ServerTrust.AuthorityOperationID
		r.TrustAuthorityEpoch = 4
		r.TrustAuthoritySequence = 7
		r.TrustBundleVersion = 6
		r.TrustBundleDigest = append([]byte{8}, make([]byte, 31)...)
	}
	return l, r, s
}

// Missing validation before signed/byte conversion must not manufacture facts.
func TestCertificateRowRejectMalformed(t *testing.T) {
	cases := map[string]func(*store.ListExactCertificateProjectionRow){
		"null lineage":  func(r *store.ListExactCertificateProjectionRow) { r.InventoryLineageID.Valid = false },
		"null from":     func(r *store.ListExactCertificateProjectionRow) { r.ValidFrom.Valid = false },
		"null until":    func(r *store.ListExactCertificateProjectionRow) { r.ValidUntil.Valid = false },
		"infinite from": func(r *store.ListExactCertificateProjectionRow) { r.ValidFrom.InfinityModifier = pgtype.Infinity },
		"negative infinite until": func(r *store.ListExactCertificateProjectionRow) {
			r.ValidUntil.InfinityModifier = pgtype.NegativeInfinity
		},
		"unknown time modifier": func(r *store.ListExactCertificateProjectionRow) { r.ValidUntil.InfinityModifier = 2 },
		"epoch mismatch":        func(r *store.ListExactCertificateProjectionRow) { r.AuthorityEpoch = 3; r.TrustAuthorityEpoch = 3 },
		"missing trust":         func(r *store.ListExactCertificateProjectionRow) { r.HasServerTrust = false },
		"corrupt identity":      func(r *store.ListExactCertificateProjectionRow) { r.NodeID[0]++ },
		"corrupt DER":           func(r *store.ListExactCertificateProjectionRow) { r.LeafDer[0]++ },
		"wrong domain":          func(r *store.ListExactCertificateProjectionRow) { r.TrustDomain = "other.example" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			l, r, _ := certificateRowFixture(t, true)
			mutate(&r)
			assertRowConflict(t, l, r)
		})
	}
	coordinates := map[string]func(*store.ListExactCertificateProjectionRow, int64){
		"identity":           func(r *store.ListExactCertificateProjectionRow, n int64) { r.IdentityEpoch = n },
		"inventory identity": func(r *store.ListExactCertificateProjectionRow, n int64) { r.InventoryIdentityEpoch = n },
		"inventory version":  func(r *store.ListExactCertificateProjectionRow, n int64) { r.InventoryVersion = n },
		"authority epoch":    func(r *store.ListExactCertificateProjectionRow, n int64) { r.AuthorityEpoch = n },
		"authority sequence": func(r *store.ListExactCertificateProjectionRow, n int64) { r.AuthoritySequence = n },
		"trust epoch":        func(r *store.ListExactCertificateProjectionRow, n int64) { r.TrustAuthorityEpoch = n },
		"trust sequence":     func(r *store.ListExactCertificateProjectionRow, n int64) { r.TrustAuthoritySequence = n },
		"trust version":      func(r *store.ListExactCertificateProjectionRow, n int64) { r.TrustBundleVersion = n },
	}
	for name, set := range coordinates {
		for _, n := range []int64{0, -1, math.MinInt64} {
			t.Run(fmt.Sprintf("%s/%d", name, n), func(t *testing.T) {
				l, r, _ := certificateRowFixture(t, true)
				set(&r, n)
				assertRowConflict(t, l, r)
			})
		}
	}
	for _, field := range []string{"leaf", "key", "trust"} {
		for _, size := range []int{0, 31, 33} {
			t.Run(fmt.Sprintf("%s digest/%d", field, size), func(t *testing.T) {
				l, r, _ := certificateRowFixture(t, true)
				switch field {
				case "leaf":
					data := make([]byte, size)
					copy(data, r.LeafDerSha256)
					r.LeafDerSha256 = data
				case "key":
					data := make([]byte, size)
					copy(data, r.PublicKeySha256)
					r.PublicKeySha256 = data
				case "trust":
					data := make([]byte, size)
					copy(data, r.TrustBundleDigest)
					r.TrustBundleDigest = data
				}
				assertRowConflict(t, l, r)
			})
		}
	}
}

func assertRowConflict(t *testing.T, l CertificateLookup, r store.ListExactCertificateProjectionRow) {
	t.Helper()
	got, err := certificateFactsFromRows(l, 4, []store.ListExactCertificateProjectionRow{r})
	if err != authority.ErrConflict || !reflect.DeepEqual(got, CertificateFacts{}) {
		t.Fatalf("expected zero facts/conflict, got %v", err)
	}
}

func TestCertificateRowRejectHiddenTrust(t *testing.T) {
	for name, mutate := range map[string]func(*store.ListExactCertificateProjectionRow){
		"flag":      func(r *store.ListExactCertificateProjectionRow) { r.HasServerTrust = true },
		"purpose":   func(r *store.ListExactCertificateProjectionRow) { r.TrustPurpose = "agent_server" },
		"listener":  func(r *store.ListExactCertificateProjectionRow) { r.TrustListenerKind = "agent" },
		"domain":    func(r *store.ListExactCertificateProjectionRow) { r.TrustDomain = "nodes.example" },
		"operation": func(r *store.ListExactCertificateProjectionRow) { r.TrustOperationID = r.CertificateID },
		"epoch":     func(r *store.ListExactCertificateProjectionRow) { r.TrustAuthorityEpoch = 4 },
		"sequence":  func(r *store.ListExactCertificateProjectionRow) { r.TrustAuthoritySequence = 7 },
		"version":   func(r *store.ListExactCertificateProjectionRow) { r.TrustBundleVersion = 6 },
		"digest":    func(r *store.ListExactCertificateProjectionRow) { r.TrustBundleDigest = []byte{8} },
	} {
		t.Run(name, func(t *testing.T) { l, r, _ := certificateRowFixture(t, false); mutate(&r); assertRowConflict(t, l, r) })
	}
}

func TestCertificateRowConversionOwnsFacts(t *testing.T) {
	for _, agent := range []bool{false, true} {
		l, r, want := certificateRowFixture(t, agent)
		got, err := certificateFactsFromRows(l, 4, []store.ListExactCertificateProjectionRow{r})
		if err != nil || !reflect.DeepEqual(got.Snapshot(), want) {
			t.Fatalf("exact conversion failed: %v", err)
		}
		r.LeafDer[0]++
		r.SerialBytes[0]++
		r.LeafDerSha256[0]++
		r.PublicKeySha256[0]++
		if agent {
			r.TrustBundleDigest[0]++
		}
		out := got.Snapshot()
		out.LeafDER[0]++
		out.Identity.SerialBytes[0]++
		if !reflect.DeepEqual(got.Snapshot(), want) {
			t.Fatal("driver/caller mutation changed facts")
		}
	}
}

func TestCertificateRowInputAndCardinality(t *testing.T) {
	l, r, _ := certificateRowFixture(t, true)
	for _, tc := range []struct {
		lookup CertificateLookup
		epoch  uint64
		rows   []store.ListExactCertificateProjectionRow
		want   error
	}{
		{CertificateLookup{}, 4, nil, authority.ErrInvalidArgument}, {l, 0, nil, authority.ErrInvalidArgument},
		{l, math.MaxUint64, nil, authority.ErrInvalidArgument}, {l, 4, nil, authority.ErrNotFound},
		{l, 4, []store.ListExactCertificateProjectionRow{r, r}, authority.ErrConflict},
	} {
		got, err := certificateFactsFromRows(tc.lookup, tc.epoch, tc.rows)
		if err != tc.want || !reflect.DeepEqual(got, CertificateFacts{}) {
			t.Fatalf("got %v want %v", err, tc.want)
		}
	}
}
