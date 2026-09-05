package authority

import (
	"errors"
	"testing"
	"time"

	"talenro.local/platform/internal/nodecontrol/contracts"
)

// TestAuthorityV7DownAuthorizationBurnsBeforeEveryFallibleCheck catches a
// constructor that accepts noncanonical/unbound envelope bytes and proves the
// copied single-use right is burned before expiry and typed validation.
func TestAuthorityV7DownAuthorizationBurnsBeforeEveryFallibleCheck(t *testing.T) {
	now := time.Now().UTC()
	t.Run("effective window boundaries", func(t *testing.T) {
		for _, test := range []struct {
			name      string
			issuedAt  time.Time
			expiresAt time.Time
			wantError bool
		}{
			{name: "issued at now is current", issuedAt: now, expiresAt: now.Add(time.Second)},
			{name: "future issued is not current", issuedAt: now.Add(time.Nanosecond), expiresAt: now.Add(time.Second), wantError: true},
			{name: "expiry at now is not current", issuedAt: now.Add(-time.Second), expiresAt: now, wantError: true},
		} {
			t.Run(test.name, func(t *testing.T) {
				err := validateAuthorityV7EffectiveWindowAt(test.issuedAt, test.expiresAt, now)
				if test.wantError && !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("window error = %v, want ErrInvalidArgument", err)
				}
				if !test.wantError && err != nil {
					t.Fatalf("window error = %v, want nil", err)
				}
			})
		}
	})
	t.Run("retirement provider pairs exactly cover retired members", func(t *testing.T) {
		facts := task8DownFacts(now)
		facts.ProviderRetirementSet.RetiredMembers[0].ProviderIdentityDigest = task8Digest("orphan-retired-member-provider")
		if err := facts.Validate(); err == nil {
			t.Fatal("down facts accepted a retired member whose provider pair has no retirement")
		}

		facts = task8DownFacts(now)
		orphanRetirement := facts.ProviderRetirementSet.Retirements[0]
		for index := range orphanRetirement.ProviderIdentityDigest {
			orphanRetirement.ProviderIdentityDigest[index] = 0xff
			orphanRetirement.ProviderEndpointIdentityDigest[index] = 0xff
		}
		orphanRetirement.ProviderProtocolDowngradeRetirementDigest = contracts.Digest{0xff}
		facts.ProviderRetirementSet.ProviderCount++
		facts.ProviderRetirementSet.Retirements = append(facts.ProviderRetirementSet.Retirements, orphanRetirement)
		if err := facts.Validate(); err == nil {
			t.Fatal("down facts accepted a provider retirement with no retired member")
		}
	})
	for _, mutation := range []struct {
		name   string
		mutate func(*contracts.AuthorityV7DownAuthorizationFactsV1)
	}{
		{
			name: "duplicate-key envelope",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationEnvelopeJCS = []byte(`{"body":"task8","body":"other"}`)
			},
		},
		{
			name: "noncanonical envelope",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationEnvelopeJCS = []byte("{\n\"body\":\"task8\"}")
			},
		},
		{
			name: "body-digest mismatch",
			mutate: func(value *contracts.AuthorityV7DownAuthorizationFactsV1) {
				value.AuthorizationDigest = contracts.Digest{0x10, 0x20, 0x30, 0x40}
			},
		},
	} {
		t.Run("rejects "+mutation.name, func(t *testing.T) {
			view := task8AuthorizationPersistenceView(now)
			mutation.mutate(&view.Facts)
			if _, err := newVerifiedAuthorityV7DownAuthorization(view); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("constructor error = %v, want ErrInvalidArgument", err)
			}
		})
	}

	futureView := task8AuthorizationPersistenceView(now)
	futureView.Facts.Authorization.IssuedAt = now.Add(time.Minute)
	futureView.Facts.Authorization.ExpiresAt = now.Add(2 * time.Minute)
	futureView.AuthorizationBodyJCS = task8CanonicalAuthorizationBody(futureView.Facts.Authorization)
	futureView.Facts.AuthorizationDigest = authorityV7DomainDigest(
		"authority-protocol-downgrade-authorization.v1",
		futureView.AuthorizationBodyJCS,
	)
	future, err := newVerifiedAuthorityV7DownAuthorization(futureView)
	if err != nil {
		t.Fatal(err)
	}
	futureCopy := future
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(future); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("future-issued first consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(futureCopy); !errors.Is(err, ErrConflict) {
		t.Fatalf("future-issued copied retry = %v, want ErrConflict", err)
	}

	expiredView := task8AuthorizationPersistenceView(now.Add(-90 * time.Second))
	expiredView.Facts.Authorization.IssuedAt = now.Add(-90 * time.Second)
	expiredView.Facts.Authorization.ExpiresAt = now.Add(-30 * time.Second)
	expired, err := newVerifiedAuthorityV7DownAuthorization(expiredView)
	if err != nil {
		t.Fatal(err)
	}
	copyOfExpired := expired
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(expired); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expired first consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(copyOfExpired); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired copied retry = %v, want ErrConflict", err)
	}

	invalid, err := newVerifiedAuthorityV7DownAuthorization(task8AuthorizationPersistenceView(now))
	if err != nil {
		t.Fatal(err)
	}
	// Same-package mutation models corruption discovered only at consume time;
	// the right still must burn before Validate returns the error.
	invalid.state.view.Facts.Authorization.AuthorizationScope = "wrong_scope"
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(invalid); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("invalid first consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := ConsumeVerifiedAuthorityV7DownAuthorization(invalid); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid copied retry = %v, want ErrConflict", err)
	}
}
