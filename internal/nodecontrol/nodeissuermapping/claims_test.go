package nodeissuermapping

import (
	"context"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"testing"
)

// Catches missing declarations incorrectly succeeding, automatic selection and domain-order reversal.
func TestIssuerMappingClaims(t *testing.T) {
	base := issuerMappingFixture(t, "serial-one-high").Claims
	cases := []struct {
		name   string
		mutate func(*Claims)
		want   error
	}{
		{"expected scope", func(c *Claims) { c.ExpectedHead.Deployment = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"expected version", func(c *Claims) { c.ExpectedHead.Version = 0 }, ErrEvidenceInsufficient},
		{"expected digest", func(c *Claims) { c.ExpectedHead.Digest = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"catalog scope", func(c *Claims) { c.Catalog.Head.Deployment = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"association scope", func(c *Claims) { c.Association.Head.Deployment = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"empty table", func(c *Claims) { c.Catalog.Entries = nil }, ErrEvidenceInsufficient},
		{"issuer", func(c *Claims) { c.Catalog.Entries[0].IssuerID = "" }, ErrEvidenceInsufficient},
		{"version", func(c *Claims) { c.Catalog.Entries[0].CAVersion = 0 }, ErrEvidenceInsufficient},
		{"ca", func(c *Claims) { c.Catalog.Entries[0].CADERDigest = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"unknown", func(c *Claims) { c.Catalog.Entries[0].State = UnknownRegistrationState }, ErrEvidenceInsufficient},
		{"certificate", func(c *Claims) { c.Association.CertificateID = uuid.Nil }, ErrEvidenceInsufficient},
		{"issuance", func(c *Claims) { c.Association.IssuanceID = uuid.Nil }, ErrEvidenceInsufficient},
		{"node", func(c *Claims) { c.Association.NodeID = "" }, ErrEvidenceInsufficient},
		{"association issuer", func(c *Claims) { c.Association.IssuerID = "" }, ErrEvidenceInsufficient},
		{"association version", func(c *Claims) { c.Association.CAVersion = 0 }, ErrEvidenceInsufficient},
		{"leaf", func(c *Claims) { c.Association.LeafDER = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"spki", func(c *Claims) { c.Association.PublicKey = contracts.Digest{} }, ErrEvidenceInsufficient},
		{"serial", func(c *Claims) { c.Association.SerialBytes = nil }, ErrEvidenceInsufficient},
		{"no version", func(c *Claims) { c.Association.CAVersion++ }, ErrEvidenceInsufficient},
		{"case", func(c *Claims) { c.Association.IssuerID = strings.ToUpper(c.Association.IssuerID) }, ErrEvidenceInsufficient},
		{"withdrawn no fallback", func(c *Claims) {
			r := c.Catalog.Entries[0]
			r.CAVersion++
			c.Catalog.Entries = append(c.Catalog.Entries, r)
			c.Catalog.Entries[0].State = DeclaredWithdrawn
		}, ErrClaimConflict},
		{"incomplete precedes duplicate", func(c *Claims) {
			c.Catalog.Entries = append(c.Catalog.Entries, c.Catalog.Entries[0], RegistrationClaim{})
		}, ErrEvidenceInsufficient},
		{"duplicate precedes association missing", func(c *Claims) {
			c.Catalog.Entries = append(c.Catalog.Entries, c.Catalog.Entries[0])
			c.Association.CertificateID = uuid.Nil
		}, ErrClaimConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := ownedTestClone(t, base)
			tc.mutate(&c)
			if _, e := validateDeclarations(context.Background(), c); e != tc.want {
				t.Fatal(e, tc.want)
			}
		})
	}
	r, e := validateDeclarations(context.Background(), base)
	if e != nil || r != base.Catalog.Entries[0] {
		t.Fatal(r, e)
	}
	if _, e := validateDeclarations(nil, base); e != ErrInvalidInput {
		t.Fatal(e)
	}
}
func TestIssuerMappingDuplicateKeys(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		c := issuerMappingFixture(t, "serial-one-high").Claims
		r := c.Catalog.Entries[0]
		if conflict {
			r.CADERDigest[0] ^= 1
		}
		c.Catalog.Entries = append(c.Catalog.Entries, r)
		if _, e := validateDeclarations(context.Background(), c); e != ErrClaimConflict {
			t.Fatal(e)
		}
	}
	c := issuerMappingFixture(t, "serial-one-high").Claims
	r := c.Catalog.Entries[0]
	r.CAVersion++
	c.Catalog.Entries = append(c.Catalog.Entries, r)
	if _, e := validateDeclarations(context.Background(), c); e != nil {
		t.Fatal("distinct versions", e)
	}
}
func TestIssuerMappingUnselectedDuplicates(t *testing.T) {
	c := issuerMappingFixture(t, "serial-one-high").Claims
	r := c.Catalog.Entries[0]
	r.IssuerID = "other"
	c.Catalog.Entries = append(c.Catalog.Entries, r, r)
	if _, e := validateDeclarations(context.Background(), c); e != ErrClaimConflict {
		t.Fatal(e)
	}
	c.Catalog.Entries[2].CADERDigest[0] ^= 1
	if _, e := validateDeclarations(context.Background(), c); e != ErrClaimConflict {
		t.Fatal(e)
	}
}
func TestIssuerMappingDeclaredHeadMismatch(t *testing.T) {
	base := issuerMappingFixture(t, "serial-one-high").Claims
	for head := 0; head < 3; head++ {
		for field := 0; field < 3; field++ {
			c := ownedTestClone(t, base)
			h := []*SnapshotHead{&c.ExpectedHead, &c.Catalog.Head, &c.Association.Head}[head]
			switch field {
			case 0:
				h.Deployment[0] ^= 1
			case 1:
				h.Version++
			case 2:
				h.Digest[0] ^= 1
			}
			if _, e := validateDeclarations(context.Background(), c); e != ErrClaimConflict {
				t.Fatal(head, field, e)
			}
		}
	}
	probe := &ownershipFailureContext{Context: context.Background()}
	_, e := validateDeclarations(probe, base)
	if e != nil {
		t.Fatal(e)
	}
	for stop := 1; stop <= probe.calls; stop++ {
		out, e := validateDeclarations(&ownershipFailureContext{Context: context.Background(), stop: stop, failure: context.Canceled}, base)
		if e != context.Canceled || !reflect.DeepEqual(out, RegistrationClaim{}) {
			t.Fatal(stop, e)
		}
	}
}
