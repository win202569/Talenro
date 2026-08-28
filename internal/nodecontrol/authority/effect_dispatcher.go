package authority

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/store"
)

var supportedDispatcherKinds = []EffectKind{
	EffectCertificateActivate,
	EffectCertificateRevoke,
	EffectDesiredActivate,
	EffectGrantClaim,
	EffectGrantCreate,
	EffectIdentityEpochAdvance,
	EffectMetadataPublish,
	EffectOperatorTransition,
	EffectRecoveryActivate,
	EffectResourceEnvelopeActivate,
	EffectRootPublish,
	EffectSecurityIncidentOpen,
	EffectSecurityIncidentResolve,
}

type TransactionalEffectQuery struct {
	registeredKind EffectKind
	expected       Reservation
}

func (query TransactionalEffectQuery) RegisteredKind() EffectKind { return query.registeredKind }
func (query TransactionalEffectQuery) Expected() Reservation      { return query.expected }

type TransactionalResolvedEffect struct {
	OperationID uuid.UUID
	Epoch       uint64
	Sequence    uint64
	Effect      ResolvedEffect
}

type RegisteredEffectResolver interface {
	ResolveRegisteredAuthorityEffectForUpdate(context.Context, store.DBTX, TransactionalEffectQuery) (TransactionalResolvedEffect, error)
}

type TransactionalEffectResolver interface {
	ResolveAuthorityEffectForUpdate(context.Context, store.DBTX, Reservation) (ResolvedEffect, error)
}

type RegisteredEffectActivator interface {
	CaptureActivationDecisionMaterial(context.Context, Receipt) (ActivationDecisionMaterial, error)
	ActivateAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence) error
	ValidatePersistedAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence, AuthorityEffectResolution) error
}

type TransactionalEffectActivator interface {
	RegisteredEffectActivator
}

type EffectDispatcher interface {
	TransactionalEffectResolver
	TransactionalEffectActivator
	authorityEffectDispatcher()
}

type EffectRegistration struct {
	Kind      EffectKind
	Resolver  RegisteredEffectResolver
	Activator RegisteredEffectActivator
}

type effectDispatcherRegistration struct {
	resolver  RegisteredEffectResolver
	activator RegisteredEffectActivator
}

type effectDispatcher struct {
	registrations map[EffectKind]effectDispatcherRegistration
}

func NewEffectDispatcher(registrations []EffectRegistration) (EffectDispatcher, error) {
	if len(registrations) != 15 {
		return nil, ErrInvalidArgument
	}
	registry := make(map[EffectKind]effectDispatcherRegistration, len(registrations))
	for _, registration := range registrations {
		if registration.Kind.Validate() != nil {
			return nil, ErrInvalidArgument
		}
		if _, duplicate := registry[registration.Kind]; duplicate {
			return nil, ErrInvalidArgument
		}
		supported := dispatcherKindSupported(registration.Kind)
		if supported {
			if nilDispatcherValue(registration.Resolver) || nilDispatcherValue(registration.Activator) {
				return nil, ErrInvalidArgument
			}
		} else if registration.Kind == EffectTrustBundlePublish || registration.Kind == EffectOperatorAuthorizerChange {
			if registration.Resolver != nil || registration.Activator != nil {
				return nil, ErrInvalidArgument
			}
		} else {
			return nil, ErrInvalidArgument
		}
		registry[registration.Kind] = effectDispatcherRegistration{resolver: registration.Resolver, activator: registration.Activator}
	}
	for _, kind := range supportedDispatcherKinds {
		if _, exists := registry[kind]; !exists {
			return nil, ErrInvalidArgument
		}
	}
	if _, exists := registry[EffectTrustBundlePublish]; !exists {
		return nil, ErrInvalidArgument
	}
	if _, exists := registry[EffectOperatorAuthorizerChange]; !exists {
		return nil, ErrInvalidArgument
	}
	return &effectDispatcher{registrations: registry}, nil
}

func (*effectDispatcher) authorityEffectDispatcher() {}

func (dispatcher *effectDispatcher) ResolveAuthorityEffectForUpdate(ctx context.Context, database store.DBTX, expected Reservation) (ResolvedEffect, error) {
	if err := dispatcherEntryError(ctx, dispatcher); err != nil {
		return ResolvedEffect{}, err
	}
	if !dispatcherKindSupported(expected.Kind) {
		return ResolvedEffect{}, ErrConflict
	}
	if nilDispatcherValue(database) || expected.Validate() != nil {
		return ResolvedEffect{}, ErrInvalidArgument
	}
	resolved := make([]ResolvedEffect, 0, 1)
	for _, kind := range supportedDispatcherKinds {
		registration := dispatcher.registrations[kind]
		result, err := registration.resolver.ResolveRegisteredAuthorityEffectForUpdate(ctx, database, TransactionalEffectQuery{
			registeredKind: kind, expected: expected,
		})
		if mapped := dispatcherHandlerError(ctx, err); mapped != nil {
			return ResolvedEffect{}, mapped
		}
		classification, err := classifyTransactionalResolvedEffect(result, kind, expected)
		if err != nil {
			return ResolvedEffect{}, err
		}
		if classification.State != EffectAbsent {
			resolved = append(resolved, classification)
		}
	}
	switch len(resolved) {
	case 0:
		return ResolvedEffect{State: EffectAbsent}, nil
	case 1:
		return resolved[0], nil
	default:
		return ResolvedEffect{}, ErrConflict
	}
}

func classifyTransactionalResolvedEffect(result TransactionalResolvedEffect, registeredKind EffectKind, expected Reservation) (ResolvedEffect, error) {
	if result.Effect.State == EffectAbsent {
		if result.Effect.Validate() != nil || !validTransactionalEcho(result.OperationID, result.Epoch, result.Sequence) ||
			result.OperationID != expected.OperationID || result.Epoch != expected.Epoch || result.Sequence != expected.Sequence {
			return ResolvedEffect{}, ErrInjectedFailure
		}
		return result.Effect, nil
	}
	if result.Effect.Validate() != nil || !validTransactionalEcho(result.OperationID, result.Epoch, result.Sequence) {
		return ResolvedEffect{}, ErrInjectedFailure
	}
	if result.OperationID != expected.OperationID || result.Epoch != expected.Epoch || result.Sequence != expected.Sequence ||
		result.Effect.Kind != registeredKind || result.Effect.Kind != expected.Kind ||
		result.Effect.ScopeKind != expected.ScopeKind || result.Effect.ScopeDigest != expected.ScopeDigest {
		return ResolvedEffect{}, ErrConflict
	}
	return result.Effect, nil
}

func (dispatcher *effectDispatcher) CaptureActivationDecisionMaterial(ctx context.Context, receipt Receipt) (ActivationDecisionMaterial, error) {
	if err := dispatcherEntryError(ctx, dispatcher); err != nil {
		return ActivationDecisionMaterial{}, err
	}
	if !dispatcherKindSupported(receipt.Kind) {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	if receipt.Validate() != nil || receipt.Status != StatusCommitted || receipt.EffectDigest == nil {
		return ActivationDecisionMaterial{}, ErrInvalidArgument
	}
	material, err := dispatcher.registrations[receipt.Kind].activator.CaptureActivationDecisionMaterial(ctx, cloneReceiptValue(receipt))
	if mapped := dispatcherHandlerError(ctx, err); mapped != nil {
		return ActivationDecisionMaterial{}, mapped
	}
	if !validActivationDecisionMaterial(material) {
		return ActivationDecisionMaterial{}, ErrInjectedFailure
	}
	if !materialCommitmentMatchesReceipt(material, receipt) {
		return ActivationDecisionMaterial{}, ErrConflict
	}
	return cloneActivationDecisionMaterial(material), nil
}

func (dispatcher *effectDispatcher) ActivateAuthorityEffect(ctx context.Context, database store.DBTX, receipt Receipt, proof ValidatedActivationDecisionEvidence) error {
	if err := dispatcherEntryError(ctx, dispatcher); err != nil {
		return err
	}
	if !dispatcherKindSupported(receipt.Kind) {
		return ErrConflict
	}
	if nilDispatcherValue(database) || receipt.Validate() != nil || receipt.Status != StatusCommitted || !validValidatedEvidence(proof) {
		return ErrInvalidArgument
	}
	if proof.origin != activationEvidenceOriginFresh {
		return ErrConflict
	}
	if !receiptEquals(receipt, proof.input.Receipt) || !proofMatchesReceipt(proof, receipt) {
		return ErrConflict
	}
	err := dispatcher.registrations[receipt.Kind].activator.ActivateAuthorityEffect(ctx, database, cloneReceiptValue(receipt), proof)
	return dispatcherHandlerError(ctx, err)
}

func (dispatcher *effectDispatcher) ValidatePersistedAuthorityEffect(ctx context.Context, database store.DBTX, receipt Receipt,
	proof ValidatedActivationDecisionEvidence, resolution AuthorityEffectResolution,
) error {
	if err := dispatcherEntryError(ctx, dispatcher); err != nil {
		return err
	}
	if !dispatcherKindSupported(receipt.Kind) {
		return ErrConflict
	}
	if nilDispatcherValue(database) || receipt.Validate() != nil || receipt.Status != StatusCommitted ||
		!validValidatedEvidence(proof) || !validAuthorityEffectResolution(resolution) {
		return ErrInvalidArgument
	}
	if proof.origin != activationEvidenceOriginParsed {
		return ErrConflict
	}
	if !receiptEquals(receipt, proof.input.Receipt) || !proofMatchesReceipt(proof, receipt) ||
		resolution.Facts().CommitmentDigest != proof.input.Material.Commitment.Digest() {
		return ErrConflict
	}
	if err := ValidateAuthorityEffectResolution(resolution, proof.input.Material.Commitment, proof); err != nil {
		if errors.Is(err, ErrInvalidArgument) {
			return ErrInvalidArgument
		}
		return ErrConflict
	}
	err := dispatcher.registrations[receipt.Kind].activator.ValidatePersistedAuthorityEffect(
		ctx, database, cloneReceiptValue(receipt), proof, resolution,
	)
	return dispatcherHandlerError(ctx, err)
}

func validActivationDecisionMaterial(material ActivationDecisionMaterial) bool {
	if !validAuthorityEffectCommitment(material.Commitment) || !material.Reason.valid() {
		return false
	}
	commitment := material.Commitment.Facts()
	switch material.CheckpointKind {
	case CheckpointNone:
		if material.CheckpointScopeDigest != (contracts.Digest{}) {
			return false
		}
	case CheckpointNode:
		if commitment.ScopeKind != ScopeNode || material.CheckpointScopeDigest != commitment.ScopeDigest {
			return false
		}
	case CheckpointGlobal:
		if (commitment.ScopeKind != ScopeGlobalNodeTrust && commitment.ScopeKind != ScopeGlobalOperatorTrust) ||
			material.CheckpointScopeDigest != commitment.ScopeDigest {
			return false
		}
	default:
		return false
	}
	return validateEvidenceTimeAndMatrix(material, commitment) == nil
}

func materialCommitmentMatchesReceipt(material ActivationDecisionMaterial, receipt Receipt) bool {
	facts := material.Commitment.Facts()
	return receipt.EffectDigest != nil && facts.OperationID == receipt.OperationID && facts.Kind == receipt.Kind &&
		facts.ScopeKind == receipt.ScopeKind && facts.ScopeDigest == receipt.ScopeDigest && facts.Epoch == receipt.Epoch &&
		facts.Sequence == receipt.Sequence && material.Commitment.Digest() == *receipt.EffectDigest
}

func proofMatchesReceipt(proof ValidatedActivationDecisionEvidence, receipt Receipt) bool {
	commitment := proof.input.Material.Commitment.Facts()
	evidence := proof.evidence.Facts()
	return materialCommitmentMatchesReceipt(proof.input.Material, receipt) && evidence.OperationID == receipt.OperationID &&
		evidence.CommitmentDigest == *receipt.EffectDigest && commitment.Kind == receipt.Kind && commitment.ScopeKind == receipt.ScopeKind &&
		commitment.ScopeDigest == receipt.ScopeDigest && commitment.Epoch == receipt.Epoch && commitment.Sequence == receipt.Sequence
}

func receiptEquals(left, right Receipt) bool {
	return left.Reservation == right.Reservation && equalOptionalDigest(left.EffectDigest, right.EffectDigest) &&
		equalOptionalDatabasePoint(left.DatabasePoint, right.DatabasePoint) && left.Status == right.Status &&
		dispatcherEqualOptionalAbortReason(left.AbortReason, right.AbortReason) && left.ReceiptDigest == right.ReceiptDigest
}

func dispatcherEqualOptionalAbortReason(left, right *AbortReason) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneActivationDecisionMaterial(value ActivationDecisionMaterial) ActivationDecisionMaterial {
	value.Commitment.canonical = append([]byte(nil), value.Commitment.canonical...)
	return value
}

func dispatcherEntryError(ctx context.Context, dispatcher *effectDispatcher) error {
	if ctx == nil || dispatcher == nil || dispatcher.registrations == nil {
		return ErrInvalidArgument
	}
	if ctx.Err() != nil {
		return ErrCanceled
	}
	return nil
}

func dispatcherHandlerError(ctx context.Context, err error) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrCanceled
	}
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded), errors.Is(err, ErrCanceled):
		return ErrCanceled
	case errors.Is(err, ErrConflict):
		return ErrConflict
	default:
		return ErrInjectedFailure
	}
}

func dispatcherKindSupported(kind EffectKind) bool {
	for _, supported := range supportedDispatcherKinds {
		if kind == supported {
			return true
		}
	}
	return false
}

func validTransactionalEcho(operationID uuid.UUID, epoch, sequence uint64) bool {
	return validOperationID(operationID) && validPositiveCoordinate(epoch) && validPositiveCoordinate(sequence)
}

func nilDispatcherValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
