package nodeissuermapping

import (
	"context"
	"github.com/google/uuid"
	"strings"
	"testing"
)

// Catches omitted per-position charging, unsafe addition and malformed declaration acceptance.
func TestIssuerMappingBudget(t *testing.T) {
	c := Claims{Catalog: CatalogClaim{Entries: make([]RegistrationClaim, 128)}}
	for i := range c.Catalog.Entries {
		c.Catalog.Entries[i].IssuerID = strings.Repeat("A", 128)
	}
	c.Association.IssuerID = strings.Repeat("A", 128)
	c.Association.NodeID = strings.Repeat("A", 36)
	c.Association.SerialBytes = make([]byte, 20)
	if err := claimBudget(c); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Claims){
		func(c *Claims) { c.Catalog.Entries = append(c.Catalog.Entries, RegistrationClaim{}) },
		func(c *Claims) { c.Catalog.Entries[0].IssuerID += "A" },
		func(c *Claims) { c.Association.IssuerID += "A" },
		func(c *Claims) { c.Association.NodeID += "A" },
		func(c *Claims) { c.Association.SerialBytes = append(c.Association.SerialBytes, 1) },
	} {
		x := c
		x.Catalog.Entries = append([]RegistrationClaim(nil), c.Catalog.Entries...)
		mutate(&x)
		if claimBudget(x) != ErrLimitExceeded {
			t.Fatal("limit")
		}
	}
	for _, v := range []struct {
		total, size, limit int
		want               bool
	}{{16548, 20, 16568, true}, {16549, 20, 16568, false}, {int(^uint(0) >> 1), 1, 16568, false}, {0, -1, 16568, false}, {-1, 0, 16568, false}} {
		_, ok := addRaw(v.total, v.size, v.limit)
		if ok != v.want {
			t.Fatal(v)
		}
	}
	for _, s := range []string{"ABC:0_-", "a", ""} {
		if validateClaimShape(Claims{Association: AssociationClaim{IssuerID: s}}) != nil {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"a b", "é", "a/b", "a\n"} {
		if validateClaimShape(Claims{Association: AssociationClaim{IssuerID: s}}) != ErrInvalidInput {
			t.Fatal(s)
		}
	}
}
func TestIssuerMappingPreflight(t *testing.T) {
	if preflight(nil, Input{}) != ErrInvalidInput {
		t.Fatal("nil context")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if preflight(ctx, Input{}) != context.Canceled {
		t.Fatal("cancellation")
	}
	if preflight(context.Background(), Input{}) != nil {
		t.Fatal("missing is later")
	}
	bad := Input{Claims: Claims{Association: AssociationClaim{IssuerID: "bad space", CAVersion: 1 << 63, SerialBytes: make([]byte, 21)}}}
	if preflight(context.Background(), bad) != ErrLimitExceeded {
		t.Fatal("budget precedes shape")
	}
	for _, c := range []Claims{
		{ExpectedHead: SnapshotHead{Version: 1 << 63}},
		{Catalog: CatalogClaim{Head: SnapshotHead{Version: 1 << 63}}},
		{Association: AssociationClaim{Head: SnapshotHead{Version: 1 << 63}}},
		{Catalog: CatalogClaim{Entries: []RegistrationClaim{{CAVersion: 1 << 63}}}},
		{Catalog: CatalogClaim{Entries: []RegistrationClaim{{State: 3}}}},
		{Association: AssociationClaim{NodeID: "short"}},
		{Association: AssociationClaim{CertificateID: uuid.UUID{1}}},
	} {
		if preflight(context.Background(), Input{Claims: c}) != ErrInvalidInput {
			t.Fatal(c)
		}
	}
}
