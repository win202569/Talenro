package authority

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"talenro.local/platform/internal/store"
)

func TestEffectDispatcherClosedRegistry(t *testing.T) {
	registrations, _ := dispatcherRegistrations(t)
	dispatcher, err := NewEffectDispatcher(registrations)
	if err != nil {
		t.Fatalf("valid registry: %v", err)
	}
	if _, ok := dispatcher.(*effectDispatcher); !ok {
		t.Fatalf("dispatcher dynamic type = %T", dispatcher)
	}
	var _ EffectDispatcher = dispatcher

	typedNil := (*dispatcherTestHandler)(nil)
	tests := []struct {
		name   string
		mutate func([]EffectRegistration) []EffectRegistration
	}{
		{name: "nil", mutate: func([]EffectRegistration) []EffectRegistration { return nil }},
		{name: "missing", mutate: func(v []EffectRegistration) []EffectRegistration { return v[:len(v)-1] }},
		{name: "duplicate", mutate: func(v []EffectRegistration) []EffectRegistration { return append(v, v[0]) }},
		{name: "unknown", mutate: func(v []EffectRegistration) []EffectRegistration { v[0].Kind = EffectKind("unknown"); return v }},
		{name: "nil supported resolver", mutate: func(v []EffectRegistration) []EffectRegistration { v[0].Resolver = nil; return v }},
		{name: "nil supported activator", mutate: func(v []EffectRegistration) []EffectRegistration { v[0].Activator = nil; return v }},
		{name: "typed nil resolver", mutate: func(v []EffectRegistration) []EffectRegistration { v[0].Resolver = typedNil; return v }},
		{name: "typed nil activator", mutate: func(v []EffectRegistration) []EffectRegistration { v[0].Activator = typedNil; return v }},
		{name: "non-nil unsupported", mutate: func(v []EffectRegistration) []EffectRegistration {
			for index := range v {
				if v[index].Kind == EffectTrustBundlePublish {
					v[index].Resolver, v[index].Activator = v[0].Resolver, v[0].Activator
				}
			}
			return v
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			copyRegistrations := append([]EffectRegistration(nil), registrations...)
			if _, err := NewEffectDispatcher(test.mutate(copyRegistrations)); err != ErrInvalidArgument {
				t.Fatalf("error = %v, want exact ErrInvalidArgument", err)
			}
		})
	}
}

func TestEffectDispatcherOpaqueQueries(t *testing.T) {
	registrations, handlers := dispatcherRegistrations(t)
	dispatcher, err := NewEffectDispatcher(registrations)
	if err != nil {
		t.Fatal(err)
	}
	expected := committedReceiptFixture(t, "conditional").Reservation
	database := &dispatcherTestDB{}

	t.Run("zero absent and frozen order", func(t *testing.T) {
		resetDispatcherHandlers(handlers)
		resolved, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, expected)
		if err != nil || resolved != (ResolvedEffect{State: EffectAbsent}) {
			t.Fatalf("resolved = %#v, %v", resolved, err)
		}
		var order []EffectKind
		for _, kind := range supportedDispatcherKinds {
			handler := handlers[kind]
			if len(handler.queries) != 1 || handler.queries[0].RegisteredKind() != kind || handler.queries[0].Expected() != expected || handler.databases[0] != database {
				t.Fatalf("query for %s = %#v", kind, handler.queries)
			}
			order = append(order, handler.queries[0].RegisteredKind())
		}
		if fmt.Sprint(order) != fmt.Sprint(supportedDispatcherKinds) {
			t.Fatalf("probe order = %v", order)
		}
	})

	t.Run("one exact effect", func(t *testing.T) {
		resetDispatcherHandlers(handlers)
		want := ResolvedEffect{Kind: expected.Kind, ScopeKind: expected.ScopeKind, ScopeDigest: expected.ScopeDigest,
			EffectDigest: mustAuthorityDigest(t, strings.Repeat("ab", 32)), State: EffectCommitted}
		handlers[expected.Kind].resolved = TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence, Effect: want}
		got, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, expected)
		if err != nil || got != want {
			t.Fatalf("resolved = %#v, %v", got, err)
		}
	})

	t.Run("multiple exact effects", func(t *testing.T) {
		resetDispatcherHandlers(handlers)
		for _, kind := range []EffectKind{EffectCertificateActivate, EffectCertificateRevoke} {
			handlers[kind].resolved = TransactionalResolvedEffect{
				OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence,
				Effect: ResolvedEffect{Kind: kind, ScopeKind: expected.ScopeKind, ScopeDigest: expected.ScopeDigest,
					EffectDigest: mustAuthorityDigest(t, strings.Repeat("ac", 32)), State: EffectCommitted},
			}
		}
		if _, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, expected); err != ErrConflict {
			t.Fatalf("error = %v, want exact ErrConflict", err)
		}
	})

	tests := []struct {
		name   string
		result TransactionalResolvedEffect
		want   error
	}{
		{name: "dirty absent", result: TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence,
			Effect: ResolvedEffect{State: EffectAbsent, Kind: expected.Kind}}, want: ErrInjectedFailure},
		{name: "invalid absent echo", result: TransactionalResolvedEffect{Effect: ResolvedEffect{State: EffectAbsent}}, want: ErrInjectedFailure},
		{name: "different valid absent echo", result: TransactionalResolvedEffect{OperationID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), Epoch: 8, Sequence: 9,
			Effect: ResolvedEffect{State: EffectAbsent}}, want: ErrInjectedFailure},
		{name: "malformed non-absent", result: TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence,
			Effect: ResolvedEffect{Kind: expected.Kind, ScopeKind: expected.ScopeKind, ScopeDigest: expected.ScopeDigest, State: EffectCommitted}}, want: ErrInjectedFailure},
		{name: "different valid non-absent echo", result: TransactionalResolvedEffect{OperationID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), Epoch: expected.Epoch, Sequence: expected.Sequence,
			Effect: ResolvedEffect{Kind: expected.Kind, ScopeKind: expected.ScopeKind, ScopeDigest: expected.ScopeDigest,
				EffectDigest: mustAuthorityDigest(t, strings.Repeat("ad", 32)), State: EffectCommitted}}, want: ErrConflict},
		{name: "cross kind", result: TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence,
			Effect: ResolvedEffect{Kind: EffectCertificateRevoke, ScopeKind: expected.ScopeKind, ScopeDigest: expected.ScopeDigest,
				EffectDigest: mustAuthorityDigest(t, strings.Repeat("ae", 32)), State: EffectCommitted}}, want: ErrConflict},
		{name: "cross scope", result: TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence,
			Effect: ResolvedEffect{Kind: expected.Kind, ScopeKind: ScopeNode, ScopeDigest: mustAuthorityDigest(t, strings.Repeat("af", 32)),
				EffectDigest: mustAuthorityDigest(t, strings.Repeat("ae", 32)), State: EffectCommitted}}, want: ErrConflict},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			resetDispatcherHandlers(handlers)
			handlers[expected.Kind].resolved = test.result
			if _, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, expected); err != test.want {
				t.Fatalf("error = %v, want exact %v", err, test.want)
			}
		})
	}

	t.Run("unknown and unsupported never probe", func(t *testing.T) {
		for _, kind := range []EffectKind{EffectKind("unknown"), EffectTrustBundlePublish, EffectOperatorAuthorizerChange} {
			resetDispatcherHandlers(handlers)
			changed := expected
			changed.Kind = kind
			if _, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, changed); err != ErrConflict {
				t.Fatalf("kind %s error = %v", kind, err)
			}
			if dispatcherHandlerCallCount(handlers) != 0 {
				t.Fatalf("kind %s invoked a handler", kind)
			}
		}
	})
}

func TestEffectDispatcherProofRouting(t *testing.T) {
	registrations, handlers := dispatcherRegistrations(t)
	dispatcher, err := NewEffectDispatcher(registrations)
	if err != nil {
		t.Fatal(err)
	}
	database := &dispatcherTestDB{}
	receipt := committedReceiptFixture(t, "conditional")
	handler := handlers[receipt.Kind]
	handler.material = evidenceInputFixture(t, "may_apply").Material

	material, err := dispatcher.CaptureActivationDecisionMaterial(context.Background(), receipt)
	if err != nil || material.Commitment.Digest() != handler.material.Commitment.Digest() || handler.captureCalls != 1 {
		t.Fatalf("capture = %#v, %v calls=%d", material, err, handler.captureCalls)
	}

	_, freshProof := freshEvidenceProofFixture(t, "may_apply")
	if err := dispatcher.ActivateAuthorityEffect(context.Background(), database, receipt, freshProof); err != nil || handler.activateCalls != 1 {
		t.Fatalf("activate = %v calls=%d", err, handler.activateCalls)
	}

	evidence := freshProof.Evidence()
	parsedEvidence, err := ParseActivationDecisionEvidence(evidence.CanonicalJCS())
	if err != nil {
		t.Fatal(err)
	}
	parsedProof, err := ValidateActivationDecisionEvidence(parsedEvidence, freshProof.Input())
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.ActivateAuthorityEffect(context.Background(), database, receipt, parsedProof); err != ErrConflict || handler.activateCalls != 1 {
		t.Fatalf("parsed activate = %v calls=%d", err, handler.activateCalls)
	}

	resolution := resolutionFixture(t, "applied")
	parsedResolution, err := ParseAuthorityEffectResolution(resolution.CanonicalJCS())
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.ValidatePersistedAuthorityEffect(context.Background(), database, receipt, parsedProof, parsedResolution); err != nil || handler.validateCalls != 1 {
		t.Fatalf("persisted validate = %v calls=%d", err, handler.validateCalls)
	}
	if err := dispatcher.ValidatePersistedAuthorityEffect(context.Background(), database, receipt, freshProof, resolution); err != ErrConflict || handler.validateCalls != 1 {
		t.Fatalf("fresh persisted validate = %v calls=%d", err, handler.validateCalls)
	}

	changed := committedReceiptFixture(t, "final")
	if err := dispatcher.ActivateAuthorityEffect(context.Background(), database, changed, freshProof); err != ErrConflict || handler.activateCalls != 1 {
		t.Fatalf("cross receipt activate = %v calls=%d", err, handler.activateCalls)
	}
	if err := dispatcher.ActivateAuthorityEffect(context.Background(), database, receipt, ValidatedActivationDecisionEvidence{}); err != ErrInvalidArgument || handler.activateCalls != 1 {
		t.Fatalf("zero proof activate = %v", err)
	}
	if err := dispatcher.ValidatePersistedAuthorityEffect(context.Background(), database, receipt, parsedProof, AuthorityEffectResolution{}); err != ErrInvalidArgument || handler.validateCalls != 1 {
		t.Fatalf("zero resolution validate = %v", err)
	}
}

func TestEffectDispatcherFiniteErrors(t *testing.T) {
	registrations, handlers := dispatcherRegistrations(t)
	dispatcher, err := NewEffectDispatcher(registrations)
	if err != nil {
		t.Fatal(err)
	}
	database := &dispatcherTestDB{}
	expected := committedReceiptFixture(t, "conditional").Reservation
	handler := handlers[expected.Kind]

	for _, test := range []struct {
		name string
		err  error
		want error
	}{
		{name: "wrapped conflict", err: fmt.Errorf("secret: %w", ErrConflict), want: ErrConflict},
		{name: "wrapped canceled", err: fmt.Errorf("secret: %w", ErrCanceled), want: ErrCanceled},
		{name: "context canceled", err: fmt.Errorf("secret: %w", context.Canceled), want: ErrCanceled},
		{name: "deadline", err: fmt.Errorf("secret: %w", context.DeadlineExceeded), want: ErrCanceled},
		{name: "dynamic", err: errors.New("SECRET_SQL_TEXT"), want: ErrInjectedFailure},
	} {
		t.Run("resolve_"+test.name, func(t *testing.T) {
			resetDispatcherHandlers(handlers)
			handler.resolveErr = test.err
			if _, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), database, expected); err != test.want || strings.Contains(fmt.Sprint(err), "secret") || strings.Contains(fmt.Sprint(err), "SECRET") {
				t.Fatalf("error = %v, want exact %v", err, test.want)
			}
		})
	}

	resetDispatcherHandlers(handlers)
	handler.material = ActivationDecisionMaterial{}
	if _, err := dispatcher.CaptureActivationDecisionMaterial(context.Background(), committedReceiptFixture(t, "conditional")); err != ErrInjectedFailure {
		t.Fatalf("malformed material = %v", err)
	}
	resetDispatcherHandlers(handlers)
	handler.material = evidenceInputFixture(t, "final").Material
	if _, err := dispatcher.CaptureActivationDecisionMaterial(context.Background(), committedReceiptFixture(t, "conditional")); err != ErrConflict {
		t.Fatalf("cross material = %v", err)
	}

	var nilDatabase *dispatcherTestDB
	var typedNil store.DBTX = nilDatabase
	if _, err := dispatcher.ResolveAuthorityEffectForUpdate(context.Background(), typedNil, expected); err != ErrInvalidArgument {
		t.Fatalf("typed nil resolve = %v", err)
	}
	_, proof := freshEvidenceProofFixture(t, "may_apply")
	if err := dispatcher.ActivateAuthorityEffect(context.Background(), typedNil, committedReceiptFixture(t, "conditional"), proof); err != ErrInvalidArgument {
		t.Fatalf("typed nil activate = %v", err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	resetDispatcherHandlers(handlers)
	if _, err := dispatcher.ResolveAuthorityEffectForUpdate(canceled, database, expected); err != ErrCanceled || dispatcherHandlerCallCount(handlers) != 0 {
		t.Fatalf("entry cancellation = %v calls=%d", err, dispatcherHandlerCallCount(handlers))
	}

	ctx, cancelAfter := context.WithCancel(context.Background())
	resetDispatcherHandlers(handlers)
	handler.resolveFunc = func() { cancelAfter() }
	handler.resolveErr = errors.New("SECRET_AFTER_CANCEL")
	if _, err := dispatcher.ResolveAuthorityEffectForUpdate(ctx, database, expected); err != ErrCanceled {
		t.Fatalf("post-handler cancellation = %v", err)
	}
}

type dispatcherTestDB struct{}

func (*dispatcherTestDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (*dispatcherTestDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (*dispatcherTestDB) QueryRow(context.Context, string, ...interface{}) pgx.Row { return nil }

type dispatcherTestHandler struct {
	mu sync.Mutex

	resolved    TransactionalResolvedEffect
	resolveErr  error
	resolveFunc func()
	material    ActivationDecisionMaterial
	captureErr  error
	activateErr error
	validateErr error

	queries       []TransactionalEffectQuery
	databases     []store.DBTX
	captureCalls  int
	activateCalls int
	validateCalls int
}

func (handler *dispatcherTestHandler) ResolveRegisteredAuthorityEffectForUpdate(_ context.Context, database store.DBTX, query TransactionalEffectQuery) (TransactionalResolvedEffect, error) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.queries = append(handler.queries, query)
	handler.databases = append(handler.databases, database)
	if handler.resolveFunc != nil {
		handler.resolveFunc()
	}
	if handler.resolved == (TransactionalResolvedEffect{}) {
		expected := query.Expected()
		return TransactionalResolvedEffect{OperationID: expected.OperationID, Epoch: expected.Epoch, Sequence: expected.Sequence, Effect: ResolvedEffect{State: EffectAbsent}}, handler.resolveErr
	}
	return handler.resolved, handler.resolveErr
}

func (handler *dispatcherTestHandler) CaptureActivationDecisionMaterial(context.Context, Receipt) (ActivationDecisionMaterial, error) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.captureCalls++
	return handler.material, handler.captureErr
}

func (handler *dispatcherTestHandler) ActivateAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence) error {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.activateCalls++
	return handler.activateErr
}

func (handler *dispatcherTestHandler) ValidatePersistedAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence, AuthorityEffectResolution) error {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.validateCalls++
	return handler.validateErr
}

func dispatcherRegistrations(t *testing.T) ([]EffectRegistration, map[EffectKind]*dispatcherTestHandler) {
	t.Helper()
	handlers := make(map[EffectKind]*dispatcherTestHandler, len(supportedDispatcherKinds))
	registrations := make([]EffectRegistration, 0, 15)
	for _, kind := range supportedDispatcherKinds {
		handler := &dispatcherTestHandler{}
		handlers[kind] = handler
		registrations = append(registrations, EffectRegistration{Kind: kind, Resolver: handler, Activator: handler})
	}
	registrations = append(registrations,
		EffectRegistration{Kind: EffectOperatorAuthorizerChange},
		EffectRegistration{Kind: EffectTrustBundlePublish},
	)
	return registrations, handlers
}

func resetDispatcherHandlers(handlers map[EffectKind]*dispatcherTestHandler) {
	for _, handler := range handlers {
		*handler = dispatcherTestHandler{}
	}
}

func dispatcherHandlerCallCount(handlers map[EffectKind]*dispatcherTestHandler) int {
	total := 0
	for _, handler := range handlers {
		handler.mu.Lock()
		total += len(handler.queries) + handler.captureCalls + handler.activateCalls + handler.validateCalls
		handler.mu.Unlock()
	}
	return total
}

var _ store.DBTX = (*dispatcherTestDB)(nil)
var _ RegisteredEffectResolver = (*dispatcherTestHandler)(nil)
var _ RegisteredEffectActivator = (*dispatcherTestHandler)(nil)
