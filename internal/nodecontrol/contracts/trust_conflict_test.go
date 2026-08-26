package contracts

import (
	"crypto/sha256"
	"errors"
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestDeriveTrustConflictLocalFaultIDV1GoldenAndStable(t *testing.T) {
	binding := validTrustConflictLocalFaultBindingV1()

	first, err := DeriveTrustConflictLocalFaultIDV1(binding)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DeriveTrustConflictLocalFaultIDV1(binding)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("retry/restart derivation changed from %s to %s", first, second)
	}
	if got, want := first.String(), "a872ae37-823b-5935-ba43-439872fd5e4f"; got != want {
		t.Fatalf("local fault ID = %s, want golden %s", got, want)
	}
	if first.Version() != 5 || first.Variant() != uuid.RFC4122 {
		t.Fatalf("local fault ID = %s, want RFC 4122 variant/version 5", first)
	}
}

func TestDeriveTrustConflictLocalFaultIDV1BindsEveryField(t *testing.T) {
	binding := validTrustConflictLocalFaultBindingV1()
	want, err := DeriveTrustConflictLocalFaultIDV1(binding)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		mutate func(*TrustConflictLocalFaultBindingV1)
	}{
		{name: "node", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.NodeID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
		}},
		{name: "incident", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.IncidentID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
		}},
		{name: "poll request digest", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.PollRequestDigest[0] ^= 0xff }},
		{name: "certificate digest", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.CertificateDigest[0] ^= 0xff }},
		{name: "identity epoch", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.IdentityEpoch++ }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			candidate := binding
			test.mutate(&candidate)
			got, err := DeriveTrustConflictLocalFaultIDV1(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if got == want {
				t.Fatalf("mutating %s retained local fault ID %s", test.name, got)
			}
		})
	}
}

func TestDeriveTrustConflictLocalFaultIDV1RejectsInvalidBindings(t *testing.T) {
	valid := validTrustConflictLocalFaultBindingV1()
	cases := []struct {
		name   string
		mutate func(*TrustConflictLocalFaultBindingV1)
	}{
		{name: "nil node", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.NodeID = uuid.Nil }},
		{name: "node non RFC variant", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.NodeID = uuid.MustParse("11111111-1111-4111-0111-111111111111")
		}},
		{name: "node version zero", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.NodeID = uuid.MustParse("11111111-1111-0111-8111-111111111111")
		}},
		{name: "node version six", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.NodeID = uuid.MustParse("11111111-1111-6111-8111-111111111111")
		}},
		{name: "nil incident", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.IncidentID = uuid.Nil }},
		{name: "incident non RFC variant", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.IncidentID = uuid.MustParse("22222222-2222-4222-0222-222222222222")
		}},
		{name: "incident version zero", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.IncidentID = uuid.MustParse("22222222-2222-0222-8222-222222222222")
		}},
		{name: "incident version six", mutate: func(value *TrustConflictLocalFaultBindingV1) {
			value.IncidentID = uuid.MustParse("22222222-2222-6222-8222-222222222222")
		}},
		{name: "zero poll request digest", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.PollRequestDigest = Digest{} }},
		{name: "zero certificate digest", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.CertificateDigest = Digest{} }},
		{name: "zero identity epoch", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.IdentityEpoch = 0 }},
		{name: "identity epoch above max", mutate: func(value *TrustConflictLocalFaultBindingV1) { value.IdentityEpoch = math.MaxInt64 + 1 }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			test.mutate(&candidate)
			got, err := DeriveTrustConflictLocalFaultIDV1(candidate)
			if got != uuid.Nil || !errors.Is(err, ErrInvalidAuthorityValue) {
				t.Fatalf("derivation = %s, %v; want nil UUID and %v", got, err, ErrInvalidAuthorityValue)
			}
		})
	}
}

func TestDeriveTrustConflictLocalFaultIDV1AcceptsMaxIdentityEpoch(t *testing.T) {
	binding := validTrustConflictLocalFaultBindingV1()
	binding.IdentityEpoch = math.MaxInt64
	if got, err := DeriveTrustConflictLocalFaultIDV1(binding); err != nil || got == uuid.Nil {
		t.Fatalf("derivation = %s, %v; want non-nil UUID", got, err)
	}
}

func validTrustConflictLocalFaultBindingV1() TrustConflictLocalFaultBindingV1 {
	return TrustConflictLocalFaultBindingV1{
		NodeID:            uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		IncidentID:        uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		PollRequestDigest: sha256.Sum256([]byte("poll-request-v1")),
		CertificateDigest: sha256.Sum256([]byte("active-leaf-der-v1")),
		IdentityEpoch:     7,
	}
}
