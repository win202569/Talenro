package authority

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// TestFreshRestoreImportAdmissionHasRepeatableViewAndSingleRepositoryConsume
// catches discarded verifier preimages, caller aliasing, permissive JSON/body
// binding, and a consume right that is not shared by copied opaque values.
func TestFreshRestoreImportAdmissionHasRepeatableViewAndSingleRepositoryConsume(t *testing.T) {
	now := time.Now().UTC()
	facts := task8AdmissionFacts(now)
	if err := facts.Validate(); err != nil {
		t.Fatalf("baseline typed staging facts invalid before private preimage checks: %v", err)
	}
	payload := facts.ManifestTopology.Objects[0].Payload
	canonicalPayload, err := jcs.Transform(payload)
	if err != nil {
		t.Fatalf("baseline manifest payload JCS transform: %v", err)
	}
	if !bytes.Equal(payload, canonicalPayload) {
		t.Fatalf("baseline manifest payload is not JCS: got %s, canonical %s", payload, canonicalPayload)
	}
	if err := validateFreshRestoreManifestPayloadDigests(facts.ManifestTopology); err != nil {
		t.Fatalf("baseline manifest payload is not canonical and domain-bound: %v", err)
	}
	t.Run("pre-acquire provider head is cross-bound", func(t *testing.T) {
		candidate := task8AdmissionFacts(now)
		candidate.StagingImportCapability.ExpectedPreAcquireProviderHeadDigest = task8Digest("orphan-pre-acquire-head")
		if err := candidate.Validate(); err == nil {
			t.Fatal("facts accepted capability and lease with different pre-acquire provider heads")
		}
		candidate = task8AdmissionFacts(now)
		candidate.StagingImportCapability.ExpectedPreAcquireProviderHeadDigest = candidate.CurrentProviderHeadDigest
		candidate.StagingExclusionLease.ExpectedProviderHeadDigest = candidate.CurrentProviderHeadDigest
		if err := candidate.Validate(); err == nil {
			t.Fatal("facts accepted identical pre-acquire and acquisition-locked provider heads")
		}
	})
	t.Run("proof window is strictly positive", func(t *testing.T) {
		instant := now.Add(-time.Second)
		if err := validateFreshRestoreProofWindow(instant, instant, instant.Add(time.Second), now); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("equal proof window = %v, want ErrInvalidArgument", err)
		}
		future := now.Add(time.Second)
		if err := validateFreshRestoreProofWindow(future, future.Add(time.Second), future.Add(2*time.Second), now); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("not-yet-issued proof = %v, want ErrInvalidArgument", err)
		}
	})
	for _, testCase := range []struct {
		name   string
		mutate func(*contracts.FreshRestoreImportProjectionInputV1)
	}{
		{
			name: "expired manifest",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.IssuedAt = now.Add(-2 * time.Minute)
				value.ManifestTopology.ExpiresAt = now.Add(-time.Minute)
			},
		},
		{
			name: "expired capability",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingImportCapability.IssuedAt = now.Add(-2 * time.Minute)
				value.StagingImportCapability.ExpiresAt = now.Add(-time.Minute)
			},
		},
		{
			name: "expired lease admission",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingExclusionLease.IssuedAt = now.Add(-2 * time.Minute)
				value.StagingExclusionLease.RequestedAdmissionExpiresAt = now.Add(-time.Minute)
				value.StagingExclusionLease.AdmissionExpiresAt = now.Add(-time.Minute)
				value.AdmissionCheckedAt = now.Add(-2 * time.Minute)
			},
		},
		{
			name: "future manifest",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.IssuedAt = now.Add(time.Minute)
				value.ManifestTopology.ExpiresAt = now.Add(2 * time.Minute)
			},
		},
		{
			name: "future capability",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingImportCapability.IssuedAt = now.Add(time.Minute)
				value.StagingImportCapability.ExpiresAt = now.Add(2 * time.Minute)
			},
		},
		{
			name: "future lease",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingExclusionLease.IssuedAt = now.Add(time.Minute)
				value.StagingExclusionLease.RequestedAdmissionExpiresAt = now.Add(2 * time.Minute)
				value.StagingExclusionLease.AdmissionExpiresAt = now.Add(2 * time.Minute)
				value.AdmissionCheckedAt = now
			},
		},
	} {
		t.Run("constructor rejects "+testCase.name, func(t *testing.T) {
			candidate := task8AdmissionFacts(now)
			testCase.mutate(&candidate)
			if _, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(candidate)); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("constructor = %v, want ErrInvalidArgument", err)
			}
		})
	}
	view := task8AdmissionPersistenceView(facts)
	admission, err := newVerifiedFreshRestoreImportAdmission(view)
	if err != nil {
		t.Fatal(err)
	}
	copyOfAdmission := admission

	// Construction must own the bytes independently of the caller.
	facts.ManifestTopology.Objects[0].Payload[0] ^= 0xff
	first, err := ViewVerifiedFreshRestoreImportAdmission(admission)
	if err != nil {
		t.Fatal(err)
	}
	if first.ManifestTopology.Objects[0].Payload[0] != '{' {
		t.Fatalf("construction retained caller payload alias: first byte %#x", first.ManifestTopology.Objects[0].Payload[0])
	}

	// Every public view is a fresh clone and does not spend the private right.
	first.ManifestTopology.Objects[0].Payload[0] ^= 0xff
	second, err := ViewVerifiedFreshRestoreImportAdmission(copyOfAdmission)
	if err != nil {
		t.Fatal(err)
	}
	if second.ManifestTopology.Objects[0].Payload[0] != '{' {
		t.Fatal("repeatable admission view returned an alias into sealed state")
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*contracts.FreshRestoreImportProjectionInputV1)
	}{
		{
			name: "manifest",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.IssuedAt = now.Add(-2 * time.Minute)
				value.ManifestTopology.ExpiresAt = now.Add(-time.Minute)
			},
		},
		{
			name: "capability",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingImportCapability.IssuedAt = now.Add(-2 * time.Minute)
				value.StagingImportCapability.ExpiresAt = now.Add(-time.Minute)
			},
		},
		{
			name: "lease",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingExclusionLease.IssuedAt = now.Add(-2 * time.Minute)
				value.StagingExclusionLease.RequestedAdmissionExpiresAt = now.Add(-time.Minute)
				value.StagingExclusionLease.AdmissionExpiresAt = now.Add(-time.Minute)
				value.AdmissionCheckedAt = now.Add(-2 * time.Minute)
			},
		},
		{
			name: "future manifest",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.IssuedAt = now.Add(time.Minute)
				value.ManifestTopology.ExpiresAt = now.Add(2 * time.Minute)
			},
		},
		{
			name: "future capability",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingImportCapability.IssuedAt = now.Add(time.Minute)
				value.StagingImportCapability.ExpiresAt = now.Add(2 * time.Minute)
			},
		},
		{
			name: "future lease",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.StagingExclusionLease.IssuedAt = now.Add(time.Minute)
				value.StagingExclusionLease.RequestedAdmissionExpiresAt = now.Add(2 * time.Minute)
				value.StagingExclusionLease.AdmissionExpiresAt = now.Add(2 * time.Minute)
				value.AdmissionCheckedAt = now
			},
		},
	} {
		t.Run("public view rejects stale "+testCase.name, func(t *testing.T) {
			candidate, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(task8AdmissionFacts(now)))
			if err != nil {
				t.Fatal(err)
			}
			testCase.mutate(&candidate.state.view.Facts)
			if _, err := ViewVerifiedFreshRestoreImportAdmission(candidate); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("view = %v, want ErrInvalidArgument", err)
			}
		})
	}

	consumed, err := consumeVerifiedFreshRestoreImportAdmission(copyOfAdmission)
	if err != nil || consumed.Facts.ManifestTopology.ManifestID != second.ManifestTopology.ManifestID {
		t.Fatalf("first repository consume = %+v, %v", consumed.Facts.ManifestTopology, err)
	}
	consumed.ManifestBodyJCS[0] ^= 0xff
	if admission.state.view.ManifestBodyJCS[0] == consumed.ManifestBodyJCS[0] {
		t.Fatal("private repository consume returned a Manifest preimage alias")
	}
	if _, err := consumeVerifiedFreshRestoreImportAdmission(admission); !errors.Is(err, ErrConflict) {
		t.Fatalf("copied admission retry = %v, want ErrConflict", err)
	}

	// The current facts-only constructor wrongly accepts bytes whose parsed JSON
	// shape is valid but whose exact canonical preimage or digest is not.
	for _, mutation := range []struct {
		name   string
		mutate func(*contracts.FreshRestoreImportProjectionInputV1)
	}{
		{
			name: "noncanonical manifest bytes",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.Objects[0].Payload = []byte(` {"pop_code":"tbs-1","iso_country":"GE","region":"tbilisi"}`)
			},
		},
		{
			name: "manifest body digest mismatch",
			mutate: func(value *contracts.FreshRestoreImportProjectionInputV1) {
				value.ManifestTopology.Objects[0].PayloadDigest = contracts.Digest{0x91, 0x82, 0x73, 0x64}
			},
		},
	} {
		t.Run("rejects "+mutation.name, func(t *testing.T) {
			candidate := task8AdmissionPersistenceView(task8AdmissionFacts(now))
			mutation.mutate(&candidate.Facts)
			if _, err := newVerifiedFreshRestoreImportAdmission(candidate); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("constructor error = %v, want ErrInvalidArgument", err)
			}
		})
	}

	// The typed Manifest/Capability/Lease windows remain current while the
	// independently short-lived sealed current proof has expired. Public view
	// must reject without consuming; repository consume must reject and burn.
	proofExpired, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(task8AdmissionFacts(now)))
	if err != nil {
		t.Fatal(err)
	}
	proofExpiredCopy := proofExpired
	proofExpired.state.view.CurrentDatabaseIncarnationProofBodyJCS = task8CanonicalJSON(map[string]string{
		"issued_at":                         now.Add(-40 * time.Second).Format(time.RFC3339Nano),
		"expires_at":                        now.Add(-10 * time.Second).Format(time.RFC3339Nano),
		"attestor_runtime_lease_expires_at": now.Add(time.Minute).Format(time.RFC3339Nano),
	})
	if _, err := ViewVerifiedFreshRestoreImportAdmission(proofExpired); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expired current proof view = %v, want ErrInvalidArgument", err)
	}
	if _, err := consumeVerifiedFreshRestoreImportAdmission(proofExpired); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expired current proof consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := consumeVerifiedFreshRestoreImportAdmission(proofExpiredCopy); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired current proof copied retry = %v, want ErrConflict", err)
	}

	// Spending precedes every fallible currentness check: a right that goes
	// stale after mint cannot be retried through another copied value.
	expired, err := newVerifiedFreshRestoreImportAdmission(task8AdmissionPersistenceView(task8AdmissionFacts(now)))
	if err != nil {
		t.Fatal(err)
	}
	expired.state.view.Facts.ManifestTopology.IssuedAt = now.Add(-2 * time.Minute)
	expired.state.view.Facts.ManifestTopology.ExpiresAt = now.Add(-time.Minute)
	if _, err := consumeVerifiedFreshRestoreImportAdmission(expired); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expired first consume = %v, want ErrInvalidArgument", err)
	}
	if _, err := consumeVerifiedFreshRestoreImportAdmission(expired); !errors.Is(err, ErrConflict) {
		t.Fatalf("expired copied retry = %v, want ErrConflict", err)
	}
}
