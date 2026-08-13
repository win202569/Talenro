package trustclient_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"talenro.local/platform/internal/trustclient"
)

// FuzzVerifyEnvelope exercises the independent strict outer boundary with a
// fixed non-secret key and finite in-memory store. It never logs or returns
// attacker-controlled bytes.
func FuzzVerifyEnvelope(f *testing.F) {
	const canary = "TASK15-FUZZ-SECRET-CANARY-5C8B1F"
	fixture := newVerifyFixture(f, "1")
	f.Add(fixture.envelope)
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"envelope_version":"talenro-config-envelope/v1"}`))
	f.Add([]byte(canary))
	f.Add(bytes.Repeat([]byte{'x'}, (1<<20)+1))
	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > (1<<20)+1 {
			return
		}
		_, err := trustclient.VerifyAndStage(
			context.Background(), body, fixture.recipient, fixture.metadata, fixture.expected, trustclient.NewMemoryStore(),
			fixture.now, 120*time.Second,
		)
		if err == nil {
			return
		}
		finite := errors.Is(err, trustclient.ErrInvalidArgument) || errors.Is(err, trustclient.ErrEnvelope) || errors.Is(err, trustclient.ErrRecipient) ||
			errors.Is(err, trustclient.ErrMetadata) || errors.Is(err, trustclient.ErrMetadataRollback) || errors.Is(err, trustclient.ErrSignature) ||
			errors.Is(err, trustclient.ErrPayload) || errors.Is(err, trustclient.ErrRollback) || errors.Is(err, trustclient.ErrStore)
		if !finite {
			t.Fatalf("non-finite error: %v", err)
		}
		if bytes.Contains(body, []byte(canary)) && strings.Contains(err.Error(), canary) {
			t.Fatal("error echoed fuzz input")
		}
	})
}
