package trust

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"talenro.local/platform/internal/securitykit"
)

const (
	// ImmutableContentType is the one C1.1 envelope representation.
	ImmutableContentType = "application/vnd.talenro.bundle+json"
	// ImmutableCacheControl is the exact 24-hour immutable distribution policy.
	ImmutableCacheControl = "public, max-age=86400, immutable"
	locatorDigestDomain   = "TALENRO-BUNDLE-LOCATOR-V1\x00"
)

// ImmutableBundle owns one exact stored envelope copy.
type ImmutableBundle struct {
	Envelope       []byte
	EnvelopeSHA256 [32]byte
}

// ByteStore is the mirror-safe immutable-byte dependency.
type ByteStore interface {
	GetImmutableBundle(context.Context, [32]byte, time.Time) (ImmutableBundle, error)
}

// Distribution serves exact stored bytes without JSON re-encoding.
type Distribution struct {
	store ByteStore
	clock securitykit.Clock
}

var _ http.Handler = (*Distribution)(nil)

// NewDistribution creates the primary immutable byte handler.
func NewDistribution(store ByteStore, clock securitykit.Clock) (*Distribution, error) {
	if nilTrustValue(store) || nilTrustValue(clock) {
		return nil, ErrInvalidArgument
	}
	return &Distribution{store: store, clock: clock}, nil
}

// LocatorDigest computes the public, domain-separated locator lookup digest.
func LocatorDigest(locator []byte) [32]byte {
	if len(locator) != 32 {
		return [32]byte{}
	}
	material := make([]byte, 0, len(locatorDigestDomain)+len(locator))
	material = append(material, locatorDigestDomain...)
	material = append(material, locator...)
	digest := sha256.Sum256(material)
	clear(material)
	return digest
}

// ServeHTTP emits only exact immutable bytes or body-free finite failures.
func (distribution *Distribution) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if writer == nil || request == nil {
		return
	}
	if request.Body != nil {
		_ = request.Body.Close()
	}
	if distribution == nil || nilTrustValue(distribution.store) || nilTrustValue(distribution.clock) {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if request.Header.Get("Range") != "" {
		writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	locatorValue, ok := distributionLocator(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	locator, err := base64.RawURLEncoding.DecodeString(locatorValue)
	if err != nil || len(locator) != 32 || base64.RawURLEncoding.EncodeToString(locator) != locatorValue {
		clear(locator)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	digest := LocatorDigest(locator)
	clear(locator)
	if digest == [32]byte{} {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	now := distribution.clock.Now()
	if now.IsZero() {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	bundle, err := loadImmutableBundle(request.Context(), distribution.store, digest, now.UTC())
	defer bundle.Clear()
	if errors.Is(err, ErrBundleNotFound) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil || len(bundle.Envelope) < 2 || len(bundle.Envelope) > maximumEnvelopeBytes ||
		bundle.EnvelopeSHA256 == [32]byte{} || sha256.Sum256(bundle.Envelope) != bundle.EnvelopeSHA256 {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	writer.Header().Set("Content-Type", ImmutableContentType)
	writer.Header().Set("Cache-Control", ImmutableCacheControl)
	writer.Header().Set("ETag", `"`+hex.EncodeToString(bundle.EnvelopeSHA256[:])+`"`)
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(bundle.Envelope)
}

func loadImmutableBundle(
	ctx context.Context,
	store ByteStore,
	digest [32]byte,
	now time.Time,
) (bundle ImmutableBundle, err error) {
	defer func() {
		if recover() != nil {
			bundle.Clear()
			bundle = ImmutableBundle{}
			err = ErrRepository
		}
	}()
	return store.GetImmutableBundle(ctx, digest, now)
}

// Clear zeroizes the uniquely owned envelope copy.
func (bundle *ImmutableBundle) Clear() {
	if bundle == nil {
		return
	}
	clear(bundle.Envelope)
	bundle.Envelope = nil
	bundle.EnvelopeSHA256 = [32]byte{}
}

func distributionLocator(path string) (string, bool) {
	if !strings.HasPrefix(path, "/b/") {
		return "", false
	}
	locator := strings.TrimPrefix(path, "/b/")
	return locator, locator != "" && !strings.Contains(locator, "/")
}

// Format redacts immutable bytes from diagnostic formatting.
func (ImmutableBundle) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "ImmutableBundle")
}

// LogValue redacts immutable bytes from structured logs.
func (ImmutableBundle) LogValue() slog.Value {
	return slog.StringValue("trust.ImmutableBundle([REDACTED])")
}

// MarshalJSON forbids generic serialization of immutable envelope bytes.
func (ImmutableBundle) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }

// Format redacts distribution dependencies.
func (*Distribution) Format(state fmt.State, _ rune) {
	redactTrustValue(state, "Distribution")
}

// LogValue redacts distribution dependencies.
func (*Distribution) LogValue() slog.Value {
	return slog.StringValue("trust.Distribution([REDACTED])")
}

// MarshalJSON forbids generic serialization of distribution dependencies.
func (*Distribution) MarshalJSON() ([]byte, error) { return nil, errTrustSerialization }
