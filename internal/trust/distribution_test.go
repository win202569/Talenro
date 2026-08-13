package trust

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"talenro.local/platform/internal/apierrors"
)

func TestDistributionAndMirrorsServeByteIdenticalImmutableEnvelope(t *testing.T) {
	locator := bytes.Repeat([]byte{0x55}, 32)
	encodedLocator := base64.RawURLEncoding.EncodeToString(locator)
	digest := LocatorDigest(locator)
	envelope := []byte(`{"ciphertext":"opaque","envelope_version":"talenro-config-envelope/v1"}`)
	envelopeDigest := sha256.Sum256(envelope)
	store := &task16ByteStore{records: map[[32]byte]ImmutableBundle{
		digest: {Envelope: bytes.Clone(envelope), EnvelopeSHA256: envelopeDigest},
	}}
	clock := task16Clock{now: time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)}
	handlers := []http.Handler{
		mustTask16Distribution(t, store, clock),
		mustTask16Mirror(t, store, clock),
		mustTask16Mirror(t, store, clock),
	}

	unique := make(map[[32]byte]struct{})
	for index, handler := range handlers {
		body := &task16CloseTrackingBody{Reader: bytes.NewBufferString("ignored")}
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/"+encodedLocator, body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !body.closed {
			t.Fatalf("source %d did not close the ignored request body", index)
		}
		if response.Code != http.StatusOK {
			t.Fatalf("source %d status = %d", index, response.Code)
		}
		if !bytes.Equal(response.Body.Bytes(), envelope) {
			t.Fatalf("source %d changed immutable bytes", index)
		}
		unique[sha256.Sum256(response.Body.Bytes())] = struct{}{}
		if got := response.Header().Get("Content-Type"); got != ImmutableContentType {
			t.Fatalf("source %d content type = %q", index, got)
		}
		if got := response.Header().Get("Cache-Control"); got != ImmutableCacheControl {
			t.Fatalf("source %d cache control = %q", index, got)
		}
		if got := response.Header().Get("ETag"); got != `"`+hex.EncodeToString(envelopeDigest[:])+`"` {
			t.Fatalf("source %d ETag = %q", index, got)
		}
		if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("source %d nosniff = %q", index, got)
		}
		if len(response.Header()) != 4 {
			t.Fatalf("source %d headers = %v, want exactly four immutable headers", index, response.Header())
		}
	}
	if len(unique) != 1 {
		t.Fatalf("three immutable sources produced %d body hashes", len(unique))
	}
}

func TestMirrorPublishedBundleSurvivesSignerStopWhileNewIssueFails(t *testing.T) {
	fixture := newTask16ApplicationFixture(t)
	command := fixture.issueCommand("task16-published-before-stop-1", "1", "published")
	defer command.AccessToken.Clear()
	issued, err := fixture.application.Issue(context.Background(), command)
	if err != nil {
		t.Fatalf("Issue published bundle: %v", err)
	}
	defer issued.Locator.Clear()
	locator := issued.Locator.Copy()
	defer clear(locator)
	digest := LocatorDigest(locator)
	stored := fixture.repository.transaction.bundles[0]
	byteStore := &task16ByteStore{records: map[[32]byte]ImmutableBundle{
		digest: {Envelope: bytes.Clone(stored.Envelope), EnvelopeSHA256: stored.EnvelopeSHA256},
	}}
	if err := fixture.signer.Close(); err != nil {
		t.Fatalf("Close signer: %v", err)
	}
	fixture.sequence.set(TestConfigV1{Message: "new", Sequence: "2"})
	resolve := fixture.resolveQuery("task16-new-after-stop-000001")
	defer resolve.AccessToken.Clear()
	if _, err := fixture.application.Resolve(context.Background(), resolve); task16PublicCode(err) != apierrors.SigningUnavailable {
		t.Fatalf("new Resolve issuance after signer stop = %v, want signing_unavailable", err)
	}
	if fixture.repository.transaction.highest != 1 || len(fixture.repository.transaction.bundles) != 1 {
		t.Fatal("failed Resolve issuance after signer stop was not rolled back")
	}
	encoded := base64.RawURLEncoding.EncodeToString(locator)
	for index, handler := range []http.Handler{
		mustTask16Distribution(t, byteStore, fixture.application.clock.(task16Clock)),
		mustTask16Mirror(t, byteStore, fixture.application.clock.(task16Clock)),
		mustTask16Mirror(t, byteStore, fixture.application.clock.(task16Clock)),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/"+encoded, nil))
		if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), stored.Envelope) {
			t.Fatalf("published source %d after signer stop = status %d body %q", index, response.Code, response.Body.Bytes())
		}
	}
}

func TestMirrorOutageDoesNotAlterOtherImmutableSources(t *testing.T) {
	locator := bytes.Repeat([]byte{0x66}, 32)
	encodedLocator := base64.RawURLEncoding.EncodeToString(locator)
	digest := LocatorDigest(locator)
	body := []byte(`{"ciphertext":"published"}`)
	bodyDigest := sha256.Sum256(body)
	healthy := &task16ByteStore{records: map[[32]byte]ImmutableBundle{digest: {Envelope: body, EnvelopeSHA256: bodyDigest}}}
	failed := &task16ByteStore{err: ErrRepository}
	clock := task16Clock{now: time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)}

	for index, handler := range []http.Handler{
		mustTask16Distribution(t, healthy, clock),
		mustTask16Mirror(t, failed, clock),
		mustTask16Mirror(t, healthy, clock),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/"+encodedLocator, nil))
		if index == 1 {
			if response.Code != http.StatusServiceUnavailable || response.Body.Len() != 0 {
				t.Fatalf("failed mirror = status %d body %q", response.Code, response.Body.String())
			}
			continue
		}
		if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), body) {
			t.Fatalf("healthy source %d changed during mirror outage", index)
		}
	}
}

func TestDistributionCollapsesByteStorePanicToBodyFreeOutage(t *testing.T) {
	handler := mustTask16Distribution(t, &task16ByteStore{panicValue: "BYTE-STORE-PANIC-CANARY"}, task16Clock{
		now: time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC),
	})
	locator := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x75}, 32))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/"+locator, nil))
	if response.Code != http.StatusServiceUnavailable || response.Body.Len() != 0 {
		t.Fatalf("panic outage = status %d body %q", response.Code, response.Body.String())
	}
}

func TestDistributionRejectsRangeAndCollapsesLocatorMisses(t *testing.T) {
	store := &task16ByteStore{records: make(map[[32]byte]ImmutableBundle)}
	handler := mustTask16Distribution(t, store, task16Clock{now: time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)})
	validUnknown := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x77}, 32))

	for _, test := range []struct {
		name        string
		locator     string
		rangeHeader string
		wantStatus  int
	}{
		{name: "malformed", locator: "%%%", wantStatus: http.StatusNotFound},
		{name: "wrong length", locator: base64.RawURLEncoding.EncodeToString([]byte("short")), wantStatus: http.StatusNotFound},
		{name: "unknown", locator: validUnknown, wantStatus: http.StatusNotFound},
		{name: "range", locator: validUnknown, rangeHeader: "bytes=0-3", wantStatus: http.StatusRequestedRangeNotSatisfiable},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/b/placeholder", io.NopCloser(bytes.NewBufferString("ignored")))
			request.URL.Path = "/b/" + test.locator
			request.Header.Set("Range", test.rangeHeader)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus || response.Body.Len() != 0 {
				t.Fatalf("response = status %d body %q", response.Code, response.Body.String())
			}
		})
	}
}

type task16ByteStore struct {
	records    map[[32]byte]ImmutableBundle
	err        error
	panicValue any
}

type task16CloseTrackingBody struct {
	io.Reader
	closed bool
}

func (body *task16CloseTrackingBody) Close() error {
	body.closed = true
	return nil
}

func (store *task16ByteStore) GetImmutableBundle(_ context.Context, digest [32]byte, _ time.Time) (ImmutableBundle, error) {
	if store.panicValue != nil {
		panic(store.panicValue)
	}
	if store.err != nil {
		return ImmutableBundle{}, store.err
	}
	record, ok := store.records[digest]
	if !ok {
		return ImmutableBundle{}, ErrBundleNotFound
	}
	record.Envelope = bytes.Clone(record.Envelope)
	return record, nil
}

func mustTask16Distribution(t *testing.T, store ByteStore, clock task16Clock) http.Handler {
	t.Helper()
	handler, err := NewDistribution(store, clock)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func mustTask16Mirror(t *testing.T, store ByteStore, clock task16Clock) http.Handler {
	t.Helper()
	handler, err := NewMirror(store, clock)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
