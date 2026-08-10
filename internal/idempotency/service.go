// Package idempotency provides transaction-bound request replay protection.
package idempotency

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const (
	requestDigestDomain          = "TALENRO-IDEMPOTENCY-REQUEST-V1\x00"
	keyDigestDomain              = "TALENRO-IDEMPOTENCY-KEY-V1\x00"
	responseProtectionDomain     = "idempotency/response/v1"
	maximumCanonicalRequestBytes = 64 * 1024
	maximumResponseBytes         = 1 << 20
	maximumStoredResponseBytes   = 1_048_608
	responseFrameVersion         = byte(1)

	// AnonymousRegistrationPrincipal is the only unauthenticated principal scope.
	AnonymousRegistrationPrincipal = "anonymous_registration"
	// AnonymousRegistrationOperation is the only operation allowed for the anonymous scope.
	AnonymousRegistrationOperation = "register_account"
)

var (
	// ErrInvalidArgument reports malformed input without retaining its value.
	ErrInvalidArgument = errors.New("idempotency: invalid argument")
	// ErrStore reports a transaction-bound persistence failure.
	ErrStore = errors.New("idempotency: store failed")
	// ErrProtection reports response encryption or decryption failure.
	ErrProtection = errors.New("idempotency: protection failed")
	// ErrConflict reports that a completion lost its state or digest guard.
	ErrConflict = errors.New("idempotency: state conflict")
	// ErrRecordUnavailable reports an expired, failed, or malformed stored record.
	ErrRecordUnavailable = errors.New("idempotency: record unavailable")
	// ErrCanceled reports cancellation without wrapping caller or provider details.
	ErrCanceled = errors.New("idempotency: canceled")
)

// Outcome classifies a request without exposing stored material.
type Outcome string

const (
	// Started means the caller owns a newly inserted in-progress record.
	Started Outcome = "started"
	// Replay means a byte-identical completed response is available.
	Replay Outcome = "replay"
	// Conflict means the key was previously used for a different request.
	Conflict Outcome = "conflict"
	// InProgress means an identical request is already being processed.
	InProgress Outcome = "in_progress"
)

// Scope is the finite principal and operation namespace for a key.
type Scope struct {
	Principal string
	Operation string
}

// AnonymousRegistrationScope returns the one permitted unauthenticated scope.
func AnonymousRegistrationScope() Scope {
	return Scope{Principal: AnonymousRegistrationPrincipal, Operation: AnonymousRegistrationOperation}
}

// AuthenticatedScope binds an opaque principal UUID to a credential domain.
func AuthenticatedScope(principalID uuid.UUID, credentialDomain, operation string) (Scope, error) {
	if principalID == uuid.Nil || !validCredentialDomain(credentialDomain) || !validOperation(operation) {
		return Scope{}, ErrInvalidArgument
	}
	return Scope{
		Principal: "principal:" + principalID.String() + ":" + credentialDomain,
		Operation: operation,
	}, nil
}

// Record is an immutable, redacted idempotency handle. Byte accessors return copies.
type Record struct {
	principalScope string
	operation      string
	keyDigest      [sha256.Size]byte
	requestDigest  [sha256.Size]byte
	state          string
	responseStatus int
	responseBody   []byte
	createdAt      time.Time
	expiresAt      time.Time
	canComplete    bool
}

// RequestDigest returns the request digest by value.
func (record Record) RequestDigest() [sha256.Size]byte { return record.requestDigest }

// ResponseStatus returns the completed HTTP status, or zero before completion.
func (record Record) ResponseStatus() int { return record.responseStatus }

// ResponseBody returns a defensive copy of replay or completion bytes.
func (record Record) ResponseBody() []byte { return append([]byte(nil), record.responseBody...) }

// Format prevents diagnostic formatting from exposing record digests or responses.
func (Record) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("idempotency.Record([REDACTED])"))
}

// LogValue prevents structured logging from exposing record digests or responses.
func (Record) LogValue() slog.Value { return slog.StringValue("idempotency.Record([REDACTED])") }

// Repository begins and completes records through one caller-owned transaction.
type Repository interface {
	Begin(context.Context, Scope, string, []byte, time.Time, time.Time) (Record, Outcome, error)
	Complete(context.Context, Record, int, []byte) (Record, error)
}

type idempotencyStore interface {
	TryBeginIdempotency(context.Context, store.TryBeginIdempotencyParams) (int64, error)
	GetIdempotencyForUpdate(context.Context, store.GetIdempotencyForUpdateParams) (store.IdempotencyRecord, error)
	CompleteIdempotency(context.Context, store.CompleteIdempotencyParams) (store.IdempotencyRecord, error)
}

type repository struct {
	store     idempotencyStore
	protector sensitive.Protector
}

var _ Repository = (*repository)(nil)

// New binds all idempotency operations to the supplied caller-owned DBTX.
func New(db store.DBTX, protector sensitive.Protector) (Repository, error) {
	if nilDependency(db) || nilDependency(protector) {
		return nil, ErrInvalidArgument
	}
	return newWithStore(store.New(db), protector), nil
}

func newWithStore(boundStore idempotencyStore, protector sensitive.Protector) *repository {
	return &repository{store: boundStore, protector: protector}
}

// RequestDigest hashes exact bounded canonical bytes using the frozen request domain.
func RequestDigest(operation string, canonical []byte) ([sha256.Size]byte, error) {
	if !validOperation(operation) || len(canonical) == 0 || len(canonical) > maximumCanonicalRequestBytes {
		return [sha256.Size]byte{}, ErrInvalidArgument
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(requestDigestDomain))
	_, _ = hash.Write([]byte(operation))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(canonical)
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

// KeyDigest hashes a validated key under a distinct length-framed domain.
func KeyDigest(key string) ([sha256.Size]byte, error) {
	if !validIdempotencyKey(key) {
		return [sha256.Size]byte{}, ErrInvalidArgument
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(keyDigestDomain))
	var size [2]byte
	binary.BigEndian.PutUint16(size[:], uint16(len(key))) // #nosec G115 -- key length is validated as 22..86 above.
	_, _ = hash.Write(size[:])
	_, _ = hash.Write([]byte(key))
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

func (service *repository) Begin(ctx context.Context, scope Scope, key string, canonical []byte, createdAt, expiresAt time.Time) (Record, Outcome, error) {
	if ctx == nil || service == nil || service.store == nil || service.protector == nil ||
		nilDependency(service.store) || nilDependency(service.protector) ||
		!validScope(scope) || createdAt.IsZero() || !expiresAt.After(createdAt) {
		return Record{}, "", ErrInvalidArgument
	}
	if ctx.Err() != nil {
		return Record{}, "", ErrCanceled
	}
	requestDigest, err := RequestDigest(scope.Operation, canonical)
	if err != nil {
		return Record{}, "", err
	}
	keyDigest, err := KeyDigest(key)
	if err != nil {
		return Record{}, "", err
	}
	params := store.TryBeginIdempotencyParams{
		PrincipalScope:     scope.Principal,
		Operation:          scope.Operation,
		IdempotencyKeyHash: append([]byte(nil), keyDigest[:]...),
		RequestDigest:      append([]byte(nil), requestDigest[:]...),
		CreatedAt:          createdAt,
		ExpiresAt:          expiresAt,
	}
	rows, storeErr := service.store.TryBeginIdempotency(ctx, params)
	if storeErr != nil {
		return Record{}, "", mapContextOrStore(ctx)
	}
	started := Record{
		principalScope: scope.Principal, operation: scope.Operation,
		keyDigest: keyDigest, requestDigest: requestDigest, state: "in_progress",
		createdAt: createdAt, expiresAt: expiresAt,
		canComplete: true,
	}
	if rows == 1 {
		return started, Started, nil
	}
	if rows != 0 {
		return Record{}, "", ErrStore
	}
	stored, storeErr := service.store.GetIdempotencyForUpdate(ctx, store.GetIdempotencyForUpdateParams{
		PrincipalScope:     scope.Principal,
		Operation:          scope.Operation,
		IdempotencyKeyHash: append([]byte(nil), keyDigest[:]...),
	})
	if storeErr != nil {
		return Record{}, "", mapContextOrStore(ctx)
	}
	record, valid := recordFromStore(stored, scope, keyDigest)
	if !valid || !record.expiresAt.After(createdAt) {
		return Record{}, "", ErrRecordUnavailable
	}
	if subtle.ConstantTimeCompare(record.requestDigest[:], requestDigest[:]) != 1 {
		record.responseBody = nil
		record.responseStatus = 0
		return record, Conflict, nil
	}
	switch record.state {
	case "in_progress":
		return record, InProgress, nil
	case "completed":
		if !stored.ResponseStatus.Valid || stored.ResponseStatus.Int32 < 100 || stored.ResponseStatus.Int32 > 599 ||
			!stored.ResponseKeyVersion.Valid || stored.ResponseKeyVersion.Int32 <= 0 ||
			len(stored.ResponseCiphertext) <= 28 || len(stored.ResponseCiphertext) > maximumStoredResponseBytes {
			return Record{}, "", ErrRecordUnavailable
		}
		plaintext, decryptErr := service.protector.Decrypt(responseProtectionDomain, sensitive.EncryptedField{
			KeyVersion: uint32(stored.ResponseKeyVersion.Int32),
			Ciphertext: append([]byte(nil), stored.ResponseCiphertext...),
		})
		if decryptErr != nil || len(plaintext) == 0 || len(plaintext) > maximumResponseBytes+1 || plaintext[0] != responseFrameVersion {
			clear(plaintext)
			return Record{}, "", ErrProtection
		}
		record.responseStatus = int(stored.ResponseStatus.Int32)
		record.responseBody = append([]byte(nil), plaintext[1:]...)
		clear(plaintext)
		return record, Replay, nil
	case "failed":
		return Record{}, "", ErrRecordUnavailable
	default:
		return Record{}, "", ErrRecordUnavailable
	}
}

func (service *repository) Complete(ctx context.Context, record Record, status int, body []byte) (Record, error) {
	if ctx == nil || service == nil || service.store == nil || service.protector == nil ||
		nilDependency(service.store) || nilDependency(service.protector) ||
		!validInternalRecord(record) || status < 100 || status > 599 || len(body) > maximumResponseBytes {
		return Record{}, ErrInvalidArgument
	}
	if !record.canComplete || record.state != "in_progress" {
		return Record{}, ErrConflict
	}
	if ctx.Err() != nil {
		return Record{}, ErrCanceled
	}
	responseCopy := append([]byte(nil), body...)
	defer clear(responseCopy)
	frame := make([]byte, 1, len(responseCopy)+1)
	frame[0] = responseFrameVersion
	frame = append(frame, responseCopy...)
	defer clear(frame)
	protected, err := service.protector.Encrypt(responseProtectionDomain, frame)
	if err != nil || protected.KeyVersion == 0 || protected.KeyVersion > math.MaxInt32 ||
		len(protected.Ciphertext) == 0 || len(protected.Ciphertext) > maximumStoredResponseBytes {
		clear(protected.Ciphertext)
		return Record{}, ErrProtection
	}
	if ctx.Err() != nil {
		clear(protected.Ciphertext)
		return Record{}, ErrCanceled
	}
	ciphertext := append([]byte(nil), protected.Ciphertext...)
	clear(protected.Ciphertext)
	stored, storeErr := service.store.CompleteIdempotency(ctx, store.CompleteIdempotencyParams{
		PrincipalScope:     record.principalScope,
		Operation:          record.operation,
		IdempotencyKeyHash: append([]byte(nil), record.keyDigest[:]...),
		RequestDigest:      append([]byte(nil), record.requestDigest[:]...),
		ResponseStatus:     pgtype.Int4{Int32: int32(status), Valid: true},
		ResponseCiphertext: ciphertext,
		ResponseKeyVersion: pgtype.Int4{Int32: int32(protected.KeyVersion), Valid: true},
	})
	clear(ciphertext)
	if storeErr != nil {
		if errors.Is(storeErr, pgx.ErrNoRows) {
			return Record{}, ErrConflict
		}
		return Record{}, mapContextOrStore(ctx)
	}
	completed, valid := recordFromStore(stored, Scope{Principal: record.principalScope, Operation: record.operation}, record.keyDigest)
	if !valid || completed.state != "completed" || subtle.ConstantTimeCompare(completed.requestDigest[:], record.requestDigest[:]) != 1 {
		return Record{}, ErrConflict
	}
	completed.responseStatus = status
	completed.responseBody = append([]byte(nil), responseCopy...)
	return completed, nil
}

func recordFromStore(stored store.IdempotencyRecord, scope Scope, keyDigest [sha256.Size]byte) (Record, bool) {
	if stored.PrincipalScope != scope.Principal || stored.Operation != scope.Operation ||
		len(stored.IdempotencyKeyHash) != sha256.Size || len(stored.RequestDigest) != sha256.Size ||
		subtle.ConstantTimeCompare(stored.IdempotencyKeyHash, keyDigest[:]) != 1 ||
		stored.CreatedAt.IsZero() || !stored.ExpiresAt.After(stored.CreatedAt) {
		return Record{}, false
	}
	record := Record{
		principalScope: stored.PrincipalScope,
		operation:      stored.Operation,
		state:          stored.State,
		createdAt:      stored.CreatedAt,
		expiresAt:      stored.ExpiresAt,
	}
	copy(record.keyDigest[:], stored.IdempotencyKeyHash)
	copy(record.requestDigest[:], stored.RequestDigest)
	return record, true
}

func validInternalRecord(record Record) bool {
	return validScope(Scope{Principal: record.principalScope, Operation: record.operation}) &&
		(record.state == "in_progress" || record.state == "completed") && !record.createdAt.IsZero() && record.expiresAt.After(record.createdAt) &&
		record.keyDigest != [sha256.Size]byte{} && record.requestDigest != [sha256.Size]byte{}
}

func validScope(scope Scope) bool {
	if !validOperation(scope.Operation) || len(scope.Principal) == 0 || len(scope.Principal) > 128 {
		return false
	}
	if scope.Principal == AnonymousRegistrationPrincipal {
		return scope.Operation == AnonymousRegistrationOperation
	}
	parts := strings.Split(scope.Principal, ":")
	if len(parts) != 3 || parts[0] != "principal" || !validCredentialDomain(parts[2]) {
		return false
	}
	parsed, err := uuid.Parse(parts[1])
	return err == nil && parsed != uuid.Nil && parsed.String() == parts[1]
}

func validCredentialDomain(value string) bool {
	if len(value) == 0 || len(value) > 32 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character < 'a' || character > 'z' {
			if character < '0' || character > '9' {
				if character != '_' {
					return false
				}
			}
		}
	}
	return true
}

func validOperation(value string) bool {
	if len(value) == 0 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character < 'a' || character > 'z' {
			if character < '0' || character > '9' {
				if character != '_' {
					return false
				}
			}
		}
	}
	return true
}

func validIdempotencyKey(value string) bool {
	if len(value) < 22 || len(value) > 86 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func mapContextOrStore(ctx context.Context) error {
	if ctx != nil && ctx.Err() != nil {
		return ErrCanceled
	}
	return ErrStore
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	kind := reflected.Kind()
	nilCapable := kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice
	return nilCapable && reflected.IsNil()
}
