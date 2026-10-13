package nodeissuermapping

import (
	"reflect"
	"testing"
)

// Catches ABI changes that could add a trusted capability or omit declared coordinates.
func TestIssuerMappingAPI(t *testing.T) {
	for _, c := range []struct {
		value  any
		fields []string
	}{
		{SnapshotHead{}, []string{"Deployment", "Version", "Digest"}},
		{RegistrationClaim{}, []string{"IssuerID", "CAVersion", "CADERDigest", "State"}},
		{CatalogClaim{}, []string{"Head", "Entries"}},
		{AssociationClaim{}, []string{"Head", "CertificateID", "IssuanceID", "NodeID", "IssuerID", "CAVersion", "LeafDER", "PublicKey", "SerialBytes"}},
		{Claims{}, []string{"ExpectedHead", "Catalog", "Association"}},
		{Input{}, []string{"Material", "Claims"}},
		{Projection{}, []string{"Head", "IssuerID", "CAVersion", "CertificateID", "IssuanceID", "Local"}},
		{Summary{}, []string{"Proof", "Profile", "Evidence", "Proposed", "ObservedAt", "ValidUntil", "Obligations"}},
		{owner{}, []string{"claims", "summary", "valid"}},
	} {
		typ := reflect.TypeOf(c.value)
		if typ.NumField() != len(c.fields) {
			t.Fatal(typ)
		}
		for i, name := range c.fields {
			if typ.Field(i).Name != name {
				t.Fatal(typ, name)
			}
		}
	}
	if None != 0 || DeclaredIssuerSerialCorrespondenceConditionalTimeBound != 1 || UnknownRegistrationState != 0 || DeclaredEligible != 1 || DeclaredWithdrawn != 2 || FullIdentityCompositionCommitAndRuntimeAuthorization != 5 {
		t.Fatal("enum")
	}
	if declaredMappingProfile != "declared-issuer-serial.registry-version-link.v1" || materialReferenceProfile != "main-selected-issuer-serial.ca-der-bound.v1" {
		t.Fatal("profiles")
	}
}
func TestIssuerMappingFixedErrors(t *testing.T) {
	for _, c := range []struct {
		err  error
		want string
	}{
		{ErrInvalidInput, "invalid input"}, {ErrLimitExceeded, "limit exceeded"}, {ErrInvalidEvidence, "invalid evidence"}, {ErrEvidenceInsufficient, "evidence insufficient"}, {ErrClaimConflict, "claim conflict"},
	} {
		if c.err.Error() != "nodeissuermapping: "+c.want {
			t.Fatal(c.err)
		}
	}
}
