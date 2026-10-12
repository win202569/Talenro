package serving

import (
	"encoding/hex"
	"math"
	"reflect"
	"testing"

	"talenro.local/platform/internal/nodecontrol/authority"
)

func TestCertificateProjectionArguments(t *testing.T) {
	lookup, _ := certificateFixture(t, true)
	got, err := certificateQueryArguments(lookup, 4)
	if err != nil {
		t.Fatal(err)
	}
	// Independently computed from the canonical UUID bytes, not Windows GUID layout.
	if hex.EncodeToString(got.NodeScopeDigest[:]) != "18823b1e06b1065ff0fb1fc8d6a7b87fa2aa01ea5c7337e75ef3b72ad0d2adee" || hex.EncodeToString(got.TrustScopeDigest[:]) != "f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1" {
		t.Fatal("wrong derived scope")
	}
	if got.AuthorityEpoch != 4 || !got.RequireAgentTrust || got.TrustDomain != "nodes.example" || !reflect.DeepEqual(got.Identity, lookup.Input()) {
		t.Fatal("query identity changed")
	}
	got.Identity.SerialBytes[0]++
	if reflect.DeepEqual(got.Identity.SerialBytes, lookup.Input().SerialBytes) {
		t.Fatal("argument alias changed lookup")
	}
	plain, _ := NewCertificateLookup(lookup.Input())
	nonAgent, err := certificateQueryArguments(plain, 4)
	if err != nil || nonAgent.RequireAgentTrust || nonAgent.TrustDomain != "" {
		t.Fatal("non-agent acquired a selector")
	}
	for _, epoch := range []uint64{0, math.MaxUint64} {
		got, err := certificateQueryArguments(lookup, epoch)
		if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, certificateQueryParams{}) {
			t.Fatal("invalid epoch accepted")
		}
	}
	got, err = certificateQueryArguments(CertificateLookup{}, 4)
	if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, certificateQueryParams{}) {
		t.Fatal("invalid lookup accepted")
	}
}

// An exact query must not publish the first of multiple candidate rows.
func TestCertificateProjectionCardinality(t *testing.T) {
	l, s := certificateFixture(t, true)
	for _, tc := range []struct {
		name string
		rows []CertificateSnapshot
		want error
	}{
		{"missing", nil, authority.ErrNotFound},
		{"duplicate", []CertificateSnapshot{s, s}, authority.ErrConflict},
		{"exact", []CertificateSnapshot{s}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectCertificateProjection(l, tc.rows)
			if err != tc.want {
				t.Fatalf("got %v want %v", err, tc.want)
			}
			if err != nil && !reflect.DeepEqual(got, CertificateFacts{}) {
				t.Fatal("failed query leaked facts")
			}
			if err == nil && !reflect.DeepEqual(got.Snapshot(), s) {
				t.Fatal("lost exact facts")
			}
		})
	}
}

func TestCertificateProjectionValidatesBeforePublishing(t *testing.T) {
	l, s := certificateFixture(t, true)
	s.ServerTrust.Selector.TrustDomain = "wrong.example"
	got, err := selectCertificateProjection(l, []CertificateSnapshot{s})
	if err != authority.ErrConflict || !reflect.DeepEqual(got, CertificateFacts{}) {
		t.Fatal("corrupt projection accepted")
	}
	got, err = selectCertificateProjection(CertificateLookup{}, nil)
	if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, CertificateFacts{}) {
		t.Fatal("invalid lookup misclassified")
	}
}
