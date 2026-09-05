package authority

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

type VerifiedFreshRestoreImportAdmission struct {
	state *verifiedFreshRestoreImportAdmissionState
}

type freshRestoreImportAdmissionPersistenceView struct {
	Facts contracts.FreshRestoreImportProjectionInputV1

	StagingImportCapabilityBodyJCS                   []byte
	StagingImportCapabilityEnvelopeJCS               []byte
	StagingImportCapabilityEvidenceBundleJCS         []byte
	ManifestBodyJCS                                  []byte
	ManifestEnvelopeJCS                              []byte
	ManifestEvidenceBundleJCS                        []byte
	StagingExclusionLeaseBodyJCS                     []byte
	StagingExclusionLeaseEnvelopeJCS                 []byte
	StagingExclusionLeaseEvidenceBundleJCS           []byte
	AcquisitionLockedProviderHeadBodyJCS             []byte
	AcquisitionLockedProviderHeadEnvelopeJCS         []byte
	CurrentDatabaseIncarnationProofBodyJCS           []byte
	CurrentDatabaseIncarnationProofEnvelopeJCS       []byte
	CurrentDatabaseIncarnationProofEvidenceBundleJCS []byte
}

type verifiedFreshRestoreImportAdmissionState struct {
	view     freshRestoreImportAdmissionPersistenceView
	consumed atomic.Bool
}

type FreshRestoreImportRepository interface {
	ConsumeVerifiedFreshRestoreImport(context.Context, VerifiedFreshRestoreImportAdmission, contracts.FreshImportTopologyProjectionV1) (contracts.FreshRestoreImportApplicationV1, error)
}

func validateFreshRestoreProofWindow(issuedAt, expiresAt, attestorExpiresAt, now time.Time) error {
	if now.Before(issuedAt) || !expiresAt.After(issuedAt) || expiresAt.Sub(issuedAt) > 30*time.Second || expiresAt.After(attestorExpiresAt) || !now.Before(expiresAt) {
		return ErrInvalidArgument
	}
	return nil
}

func validateFreshRestoreAdmissionCurrentAt(facts contracts.FreshRestoreImportProjectionInputV1, now time.Time) error {
	if now.Before(facts.ManifestTopology.IssuedAt) || !now.Before(facts.ManifestTopology.ExpiresAt) ||
		now.Before(facts.StagingImportCapability.IssuedAt) || !now.Before(facts.StagingImportCapability.ExpiresAt) ||
		now.Before(facts.StagingExclusionLease.IssuedAt) || !now.Before(facts.StagingExclusionLease.AdmissionExpiresAt) {
		return ErrInvalidArgument
	}
	return nil
}

func validateFreshRestoreCurrentProofAt(bodyJCS []byte, now time.Time) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &fields) != nil || fields == nil {
		return ErrInvalidArgument
	}
	fieldNames := make([]string, 0, len(fields))
	nullable := make(map[string]bool)
	for field, raw := range fields {
		fieldNames = append(fieldNames, field)
		if bytes.Equal(raw, []byte("null")) {
			nullable[field] = true
		}
	}
	validated, err := validateAuthorityV7StrictJSON(bodyJCS, authorityV7StrictJSONSpec{Fields: fieldNames, NullablePaths: nullable})
	if err != nil {
		return ErrInvalidArgument
	}
	issuedAt, err := authorityV7CanonicalTime(validated["issued_at"])
	if err != nil {
		return ErrInvalidArgument
	}
	expiresAt, err := authorityV7CanonicalTime(validated["expires_at"])
	if err != nil {
		return ErrInvalidArgument
	}
	attestorExpiresAt, err := authorityV7CanonicalTime(validated["attestor_runtime_lease_expires_at"])
	if err != nil {
		return ErrInvalidArgument
	}
	return validateFreshRestoreProofWindow(issuedAt, expiresAt, attestorExpiresAt, now)
}

func newVerifiedFreshRestoreImportAdmission(view freshRestoreImportAdmissionPersistenceView) (VerifiedFreshRestoreImportAdmission, error) {
	cloned := cloneFreshRestoreImportAdmissionPersistenceView(view)
	now := time.Now()
	if err := cloned.Facts.Validate(); err != nil || validateFreshRestoreManifestPayloadDigests(cloned.Facts.ManifestTopology) != nil ||
		validateFreshRestoreAdmissionCurrentAt(cloned.Facts, now) != nil || validateFreshRestoreCurrentProofAt(cloned.CurrentDatabaseIncarnationProofBodyJCS, now) != nil {
		return VerifiedFreshRestoreImportAdmission{}, ErrInvalidArgument
	}
	return VerifiedFreshRestoreImportAdmission{state: &verifiedFreshRestoreImportAdmissionState{view: cloned}}, nil
}

func ViewVerifiedFreshRestoreImportAdmission(admission VerifiedFreshRestoreImportAdmission) (contracts.FreshRestoreImportProjectionInputV1, error) {
	if admission.state == nil {
		return contracts.FreshRestoreImportProjectionInputV1{}, ErrInvalidArgument
	}
	facts := cloneFreshRestoreImportProjectionInput(admission.state.view.Facts)
	now := time.Now()
	if facts.Validate() != nil || validateFreshRestoreManifestPayloadDigests(facts.ManifestTopology) != nil ||
		validateFreshRestoreAdmissionCurrentAt(facts, now) != nil || validateFreshRestoreCurrentProofAt(admission.state.view.CurrentDatabaseIncarnationProofBodyJCS, now) != nil {
		return contracts.FreshRestoreImportProjectionInputV1{}, ErrInvalidArgument
	}
	return facts, nil
}

func consumeVerifiedFreshRestoreImportAdmission(admission VerifiedFreshRestoreImportAdmission) (freshRestoreImportAdmissionPersistenceView, error) {
	if admission.state == nil {
		return freshRestoreImportAdmissionPersistenceView{}, ErrInvalidArgument
	}
	if !admission.state.consumed.CompareAndSwap(false, true) {
		return freshRestoreImportAdmissionPersistenceView{}, ErrConflict
	}
	view := cloneFreshRestoreImportAdmissionPersistenceView(admission.state.view)
	now := time.Now()
	if view.Facts.Validate() != nil || validateFreshRestoreManifestPayloadDigests(view.Facts.ManifestTopology) != nil ||
		validateFreshRestoreAdmissionCurrentAt(view.Facts, now) != nil || validateFreshRestoreCurrentProofAt(view.CurrentDatabaseIncarnationProofBodyJCS, now) != nil {
		return freshRestoreImportAdmissionPersistenceView{}, ErrInvalidArgument
	}
	return view, nil
}

func validateFreshRestoreManifestPayloadDigests(value contracts.FreshRestoreImportManifestTopologyFactsV1) error {
	for _, object := range value.Objects {
		canonical, err := jcs.Transform(object.Payload)
		if err != nil || !equalAuthorityV7Bytes(canonical, object.Payload) {
			return ErrInvalidArgument
		}
		hash := sha256.New()
		_, _ = hash.Write([]byte("talenro.c12.fresh-import-payload-" + string(object.ObjectType) + ".v1"))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(object.Payload)
		var digest contracts.Digest
		copy(digest[:], hash.Sum(nil))
		if digest != object.PayloadDigest {
			return ErrInvalidArgument
		}
	}
	return nil
}

func cloneFreshRestoreImportAdmissionPersistenceView(value freshRestoreImportAdmissionPersistenceView) freshRestoreImportAdmissionPersistenceView {
	result := value
	result.Facts = cloneFreshRestoreImportProjectionInput(value.Facts)
	result.StagingImportCapabilityBodyJCS = cloneAuthorityV7Bytes(value.StagingImportCapabilityBodyJCS)
	result.StagingImportCapabilityEnvelopeJCS = cloneAuthorityV7Bytes(value.StagingImportCapabilityEnvelopeJCS)
	result.StagingImportCapabilityEvidenceBundleJCS = cloneAuthorityV7Bytes(value.StagingImportCapabilityEvidenceBundleJCS)
	result.ManifestBodyJCS = cloneAuthorityV7Bytes(value.ManifestBodyJCS)
	result.ManifestEnvelopeJCS = cloneAuthorityV7Bytes(value.ManifestEnvelopeJCS)
	result.ManifestEvidenceBundleJCS = cloneAuthorityV7Bytes(value.ManifestEvidenceBundleJCS)
	result.StagingExclusionLeaseBodyJCS = cloneAuthorityV7Bytes(value.StagingExclusionLeaseBodyJCS)
	result.StagingExclusionLeaseEnvelopeJCS = cloneAuthorityV7Bytes(value.StagingExclusionLeaseEnvelopeJCS)
	result.StagingExclusionLeaseEvidenceBundleJCS = cloneAuthorityV7Bytes(value.StagingExclusionLeaseEvidenceBundleJCS)
	result.AcquisitionLockedProviderHeadBodyJCS = cloneAuthorityV7Bytes(value.AcquisitionLockedProviderHeadBodyJCS)
	result.AcquisitionLockedProviderHeadEnvelopeJCS = cloneAuthorityV7Bytes(value.AcquisitionLockedProviderHeadEnvelopeJCS)
	result.CurrentDatabaseIncarnationProofBodyJCS = cloneAuthorityV7Bytes(value.CurrentDatabaseIncarnationProofBodyJCS)
	result.CurrentDatabaseIncarnationProofEnvelopeJCS = cloneAuthorityV7Bytes(value.CurrentDatabaseIncarnationProofEnvelopeJCS)
	result.CurrentDatabaseIncarnationProofEvidenceBundleJCS = cloneAuthorityV7Bytes(value.CurrentDatabaseIncarnationProofEvidenceBundleJCS)
	return result
}
