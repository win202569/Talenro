package trust

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"talenro.local/platform/internal/secret"
)

const (
	configKeyIDDomain         = "TALENRO-CONFIG-SIGNING-KEY-ID-V1\x00"
	rootKeyIDDomain           = "TALENRO-ROOT-SIGNING-KEY-ID-V1\x00"
	maximumSignerMessageBytes = maximumCanonicalPayloadBytes + 256
	minimumSignerTimeout      = 500 * time.Millisecond
	maximumSignerTimeout      = 5 * time.Second
	localSignerTimeout        = 2 * time.Second
)

type localSignerState struct {
	mu     sync.RWMutex
	seed   []byte
	public ed25519.PublicKey
	keyID  string
	closed bool
}

// LocalConfigSigner is a local/test-only Ed25519 configuration signer.
type LocalConfigSigner struct{ state *localSignerState }

// LocalRootSigner is a separate local/test-only Ed25519 trust-root signer.
type LocalRootSigner struct{ state *localSignerState }

var _ ConfigSigner = (*LocalConfigSigner)(nil)
var _ ConfigSigner = (*LocalRootSigner)(nil)

// NewLocalConfigSigner copies an exact 32-byte local/test configuration seed.
func NewLocalConfigSigner(seed secret.Bytes) (*LocalConfigSigner, error) {
	state, err := newLocalSignerState(seed, configKeyIDDomain)
	if err != nil {
		return nil, err
	}
	return &LocalConfigSigner{state: state}, nil
}

// NewLocalRootSigner copies an exact 32-byte local/test trust-root seed.
func NewLocalRootSigner(seed secret.Bytes) (*LocalRootSigner, error) {
	state, err := newLocalSignerState(seed, rootKeyIDDomain)
	if err != nil {
		return nil, err
	}
	return &LocalRootSigner{state: state}, nil
}

func newLocalSignerState(seed secret.Bytes, keyIDDomain string) (*localSignerState, error) {
	seedCopy := seed.Copy()
	defer clear(seedCopy)
	if len(seedCopy) != ed25519.SeedSize {
		return nil, ErrInvalidArgument
	}
	privateKey := ed25519.NewKeyFromSeed(seedCopy)
	publicKey := append(ed25519.PublicKey(nil), privateKey.Public().(ed25519.PublicKey)...)
	clear(privateKey)
	return &localSignerState{
		seed: append([]byte(nil), seedCopy...), public: publicKey,
		keyID: deriveKeyID(keyIDDomain, publicKey),
	}, nil
}

// KeyID returns the domain-separated public configuration key identifier.
func (signer *LocalConfigSigner) KeyID() string { return localKeyID(configState(signer)) }

// KeyID returns the domain-separated public root key identifier.
func (signer *LocalRootSigner) KeyID() string { return localKeyID(rootState(signer)) }

// PublicKey returns a defensive copy of the public configuration key.
func (signer *LocalConfigSigner) PublicKey() ed25519.PublicKey {
	return localPublicKey(configState(signer))
}

// PublicKey returns a defensive copy of the public root key.
func (signer *LocalRootSigner) PublicKey() ed25519.PublicKey {
	return localPublicKey(rootState(signer))
}

// Sign implements ConfigSigner without exposing private seed material.
func (signer *LocalConfigSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	return localSign(ctx, configState(signer), message)
}

// Sign signs root-metadata material without exposing private seed material.
func (signer *LocalRootSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	return localSign(ctx, rootState(signer), message)
}

func localKeyID(state *localSignerState) string {
	if state == nil {
		return ""
	}
	state.mu.RLock()
	defer state.mu.RUnlock()
	if state.closed {
		return ""
	}
	return state.keyID
}

func localPublicKey(state *localSignerState) ed25519.PublicKey {
	if state == nil {
		return nil
	}
	state.mu.RLock()
	defer state.mu.RUnlock()
	if state.closed {
		return nil
	}
	return append(ed25519.PublicKey(nil), state.public...)
}

func localSign(ctx context.Context, state *localSignerState, message []byte) ([]byte, error) {
	if state == nil {
		return nil, ErrClosed
	}
	if isNilValue(ctx) || len(message) == 0 || len(message) > maximumSignerMessageBytes || ctx.Err() != nil {
		return nil, ErrInvalidArgument
	}
	messageCopy := bytes.Clone(message)
	defer clear(messageCopy)
	state.mu.RLock()
	defer state.mu.RUnlock()
	if state.closed {
		return nil, ErrClosed
	}
	if ctx.Err() != nil {
		return nil, ErrInvalidArgument
	}
	privateKey := ed25519.NewKeyFromSeed(state.seed)
	defer clear(privateKey)
	signature := ed25519.Sign(privateKey, messageCopy)
	return append([]byte(nil), signature...), nil
}

// Close atomically erases the shared configuration-signing state.
func (signer *LocalConfigSigner) Close() error { return closeLocalSigner(configState(signer)) }

// Close atomically erases the shared root-signing state.
func (signer *LocalRootSigner) Close() error { return closeLocalSigner(rootState(signer)) }

func closeLocalSigner(state *localSignerState) error {
	if state == nil {
		return nil
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed {
		return nil
	}
	clear(state.seed)
	clear(state.public)
	state.seed = nil
	state.public = nil
	state.keyID = ""
	state.closed = true
	return nil
}

// Format redacts every fmt rendering of a local configuration signer.
func (LocalConfigSigner) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trust.LocalConfigSigner([REDACTED])"))
}

// Format redacts every fmt rendering of a local root signer.
func (LocalRootSigner) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trust.LocalRootSigner([REDACTED])"))
}

// LogValue redacts structured logging of a local configuration signer.
func (LocalConfigSigner) LogValue() slog.Value {
	return slog.StringValue("trust.LocalConfigSigner([REDACTED])")
}

// LogValue redacts structured logging of a local root signer.
func (LocalRootSigner) LogValue() slog.Value {
	return slog.StringValue("trust.LocalRootSigner([REDACTED])")
}

func configState(signer *LocalConfigSigner) *localSignerState {
	if signer == nil {
		return nil
	}
	return signer.state
}

func rootState(signer *LocalRootSigner) *localSignerState {
	if signer == nil {
		return nil
	}
	return signer.state
}

// TimeoutConfigSigner owns one bounded external-provider worker. A provider
// that ignores cancellation can strand only this worker, never one per call.
type TimeoutConfigSigner struct {
	signer    ConfigSigner
	keyID     string
	timeout   time.Duration
	requests  chan signerRequest
	done      chan struct{}
	closeOnce sync.Once
	closed    atomic.Bool
}

type signerRequest struct {
	ctx     context.Context
	message []byte
	result  chan signerResult
}

type signerResult struct {
	signature []byte
	err       error
}

// NewTimeoutConfigSigner enforces the configured signer-call timeout boundary.
func NewTimeoutConfigSigner(signer ConfigSigner, timeout time.Duration) (*TimeoutConfigSigner, error) {
	if isNilValue(signer) || timeout < minimumSignerTimeout || timeout > maximumSignerTimeout {
		return nil, ErrInvalidArgument
	}
	keyID := signer.KeyID()
	if !validKeyID(keyID) {
		return nil, ErrInvalidArgument
	}
	bounded := &TimeoutConfigSigner{
		signer: signer, keyID: keyID, timeout: timeout,
		requests: make(chan signerRequest), done: make(chan struct{}),
	}
	go bounded.run()
	return bounded, nil
}

// KeyID returns the validated provider key identifier while the adapter is open.
func (signer *TimeoutConfigSigner) KeyID() string {
	if signer == nil || signer.closed.Load() {
		return ""
	}
	select {
	case <-signer.done:
		return ""
	default:
		return signer.keyID
	}
}

// Sign submits one message to the owned provider worker within its fixed budget.
func (signer *TimeoutConfigSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	if signer != nil && signer.closed.Load() {
		return nil, ErrClosed
	}
	if signer == nil || isNilValue(signer.signer) || isNilValue(ctx) || len(message) == 0 ||
		len(message) > maximumSignerMessageBytes || ctx.Err() != nil {
		return nil, ErrInvalidArgument
	}
	bounded, cancel := context.WithTimeout(ctx, signer.timeout)
	defer cancel()
	request := signerRequest{ctx: bounded, message: bytes.Clone(message), result: make(chan signerResult)}
	select {
	case signer.requests <- request:
	case <-signer.done:
		clear(request.message)
		return nil, ErrClosed
	case <-bounded.Done():
		clear(request.message)
		return nil, ErrSignerTimeout
	}
	select {
	case returned := <-request.result:
		return returned.signature, returned.err
	case <-signer.done:
		return nil, ErrClosed
	case <-bounded.Done():
		return nil, ErrSignerTimeout
	}
}

func (signer *TimeoutConfigSigner) run() {
	for {
		select {
		case <-signer.done:
			return
		case request := <-signer.requests:
			if signer.closed.Load() {
				clear(request.message)
				continue
			}
			signature, err := signer.signer.Sign(request.ctx, request.message)
			clear(request.message)
			result := signerResult{}
			if err != nil || len(signature) != ed25519.SignatureSize {
				clear(signature)
				result.err = ErrSignerFailure
			} else {
				result.signature = append([]byte(nil), signature...)
				clear(signature)
			}
			select {
			case request.result <- result:
			case <-request.ctx.Done():
				clear(result.signature)
			case <-signer.done:
				clear(result.signature)
			}
		}
	}
}

// Close fails the adapter closed and releases an idle owned worker.
func (signer *TimeoutConfigSigner) Close() error {
	if signer == nil {
		return nil
	}
	signer.closeOnce.Do(func() {
		signer.closed.Store(true)
		close(signer.done)
	})
	return nil
}

// Format redacts every fmt rendering of the external signer adapter.
func (*TimeoutConfigSigner) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("trust.TimeoutConfigSigner([REDACTED])"))
}

// LogValue redacts structured logging of the external signer adapter.
func (*TimeoutConfigSigner) LogValue() slog.Value {
	return slog.StringValue("trust.TimeoutConfigSigner([REDACTED])")
}

// SignBundleV1 canonicalizes, hashes, and signs one exact payload.
func SignBundleV1(ctx context.Context, payload PayloadV1, signer ConfigSigner) (SignedBundleV1, error) {
	boundedSigner, bounded := signer.(*TimeoutConfigSigner)
	if isNilValue(ctx) || !bounded || boundedSigner == nil || ctx.Err() != nil || !validKeyID(boundedSigner.KeyID()) {
		return SignedBundleV1{}, ErrInvalidArgument
	}
	canonical, err := CanonicalizePayloadV1(payload)
	if err != nil {
		return SignedBundleV1{}, err
	}
	message := make([]byte, 0, len(bundleSignatureDomain)+len(canonical))
	message = append(message, bundleSignatureDomain...)
	message = append(message, canonical...)
	signature, err := boundedSigner.Sign(ctx, message)
	clear(message)
	if err != nil || len(signature) != ed25519.SignatureSize {
		clear(signature)
		if err != nil {
			return SignedBundleV1{}, err
		}
		return SignedBundleV1{}, ErrInvalidSignature
	}
	digest := sha256.Sum256(canonical)
	value := SignedBundleV1{
		PayloadJCS: string(canonical), PayloadSHA256: encodeBase64URL(digest[:]),
		SignerKeyID: boundedSigner.KeyID(), Algorithm: SignatureAlgorithm,
		Signature: encodeBase64URL(signature), Padding: "",
	}
	clear(signature)
	if err := ValidateSignedBundleV1(value); err != nil {
		return SignedBundleV1{}, err
	}
	return value, nil
}

// VerifySignedBundleV1 validates every binding and returns the trusted payload.
func VerifySignedBundleV1(value SignedBundleV1, publicKey ed25519.PublicKey) (PayloadV1, error) {
	if err := ValidateSignedBundleV1(value); err != nil || len(publicKey) != ed25519.PublicKeySize {
		return PayloadV1{}, ErrInvalidSignature
	}
	payload, err := DecodePayloadV1([]byte(value.PayloadJCS))
	if err != nil {
		return PayloadV1{}, ErrInvalidSignature
	}
	canonical, err := CanonicalizePayloadV1(payload)
	if err != nil || !bytes.Equal(canonical, []byte(value.PayloadJCS)) {
		return PayloadV1{}, ErrInvalidSignature
	}
	digest := sha256.Sum256(canonical)
	if value.PayloadSHA256 != encodeBase64URL(digest[:]) || value.SignerKeyID != deriveKeyID(configKeyIDDomain, publicKey) {
		return PayloadV1{}, ErrInvalidSignature
	}
	signature, err := base64.RawURLEncoding.DecodeString(value.Signature)
	if err != nil {
		return PayloadV1{}, ErrInvalidSignature
	}
	message := make([]byte, 0, len(bundleSignatureDomain)+len(canonical))
	message = append(message, bundleSignatureDomain...)
	message = append(message, canonical...)
	verified := ed25519.Verify(publicKey, message, signature)
	clear(message)
	clear(signature)
	if !verified {
		return PayloadV1{}, ErrInvalidSignature
	}
	return payload, nil
}

func deriveKeyID(domain string, publicKey ed25519.PublicKey) string {
	material := make([]byte, 0, len(domain)+len(publicKey))
	material = append(material, domain...)
	material = append(material, publicKey...)
	digest := sha256.Sum256(material)
	clear(material)
	return encodeBase64URL(digest[:16])
}

func encodeBase64URL(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() { //nolint:exhaustive // Only nil-capable representations matter.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
