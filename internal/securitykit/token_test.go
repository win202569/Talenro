package securitykit_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
)

type chunkedRandom struct {
	data      []byte
	chunkSize int
	buffer    []byte
}

func (random *chunkedRandom) Read(target []byte) (int, error) {
	if random.buffer == nil {
		random.buffer = target
	}
	if len(random.data) == 0 {
		return 0, io.EOF
	}
	count := min(len(target), random.chunkSize, len(random.data))
	copy(target, random.data[:count])
	random.data = random.data[count:]
	return count, nil
}

type partialFailureRandom struct {
	buffer []byte
}

type typedNilRandom struct{}

func (*typedNilRandom) Read([]byte) (int, error) { return 0, io.EOF }

type eofRandom struct{}

func (eofRandom) Read([]byte) (int, error) { return 0, io.EOF }

type shortEOFRandom struct {
	buffer []byte
}

func (random *shortEOFRandom) Read(target []byte) (int, error) {
	random.buffer = target
	copy(target, []byte{1, 2, 3})
	return 3, io.EOF
}

type noProgressRandom struct{}

func (noProgressRandom) Read([]byte) (int, error) { return 0, nil }

func (random *partialFailureRandom) Read(target []byte) (int, error) {
	random.buffer = target
	copy(target, []byte("SECRET-CANARY"))
	return len("SECRET-CANARY"), errors.New("provider SECRET-CANARY")
}

func TestNewOpaqueTokenReadsExactlyThirtyTwoBytesAndEncodesCanonicalValue(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 32)
	for index := range raw {
		raw[index] = byte(index)
	}
	random := &chunkedRandom{data: append([]byte(nil), raw...), chunkSize: 7}
	token, err := securitykit.NewOpaqueToken(random)
	if err != nil {
		t.Fatalf("NewOpaqueToken() error = %v", err)
	}
	if got := securitykit.EncodeOpaqueToken(token); got != "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8" {
		t.Fatalf("EncodeOpaqueToken() = %q", got)
	}
	if got := securitykit.EncodeOpaqueToken(token); len(got) != 43 {
		t.Fatalf("encoded token length = %d, want 43", len(got))
	}
	if !bytes.Equal(random.buffer, make([]byte, len(random.buffer))) {
		t.Fatal("random destination retained token bytes")
	}

	decoded, err := securitykit.DecodeOpaqueToken(securitykit.EncodeOpaqueToken(token))
	if err != nil {
		t.Fatalf("DecodeOpaqueToken() error = %v", err)
	}
	decodedCopy := decoded.Copy()
	defer clear(decodedCopy)
	if !bytes.Equal(decodedCopy, raw) {
		t.Fatalf("decoded token = %x, want %x", decodedCopy, raw)
	}
}

func TestNewOpaqueTokenSanitizesRandomFailureAndZeroizesPartialBytes(t *testing.T) {
	t.Parallel()

	random := &partialFailureRandom{}
	token, err := securitykit.NewOpaqueToken(random)
	if !errors.Is(err, securitykit.ErrRandomSource) {
		t.Fatalf("NewOpaqueToken() error = %v, want ErrRandomSource", err)
	}
	if got := token.Copy(); len(got) != 0 {
		clear(got)
		t.Fatalf("NewOpaqueToken() returned %d bytes after failure", len(got))
	}
	if bytes.Contains([]byte(err.Error()), []byte("SECRET-CANARY")) {
		t.Fatalf("NewOpaqueToken() error disclosed provider text: %q", err)
	}
	if !bytes.Equal(random.buffer, make([]byte, len(random.buffer))) {
		t.Fatal("partial random bytes were not zeroized")
	}
}

func TestNewOpaqueTokenRejectsTypedNilAndShortEOF(t *testing.T) {
	t.Parallel()

	var typedNil *typedNilRandom
	tests := []struct {
		name   string
		random securitykit.RandomSource
		want   error
	}{
		{name: "typed nil", random: typedNil, want: securitykit.ErrInvalidArgument},
		{name: "immediate EOF", random: eofRandom{}, want: securitykit.ErrRandomSource},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			token, err := securitykit.NewOpaqueToken(test.random)
			if !errors.Is(err, test.want) {
				t.Fatalf("NewOpaqueToken() error = %v, want %v", err, test.want)
			}
			tokenCopy := token.Copy()
			if len(tokenCopy) != 0 {
				clear(tokenCopy)
				t.Fatalf("NewOpaqueToken() returned %d bytes", len(tokenCopy))
			}
		})
	}
}

func TestNewOpaqueTokenZeroizesShortEOFBytes(t *testing.T) {
	t.Parallel()

	random := &shortEOFRandom{}
	token, err := securitykit.NewOpaqueToken(random)
	if !errors.Is(err, securitykit.ErrRandomSource) {
		t.Fatalf("NewOpaqueToken() error = %v, want ErrRandomSource", err)
	}
	if !bytes.Equal(random.buffer, make([]byte, len(random.buffer))) {
		t.Fatal("NewOpaqueToken() retained bytes from short EOF")
	}
	tokenCopy := token.Copy()
	if len(tokenCopy) != 0 {
		clear(tokenCopy)
		t.Fatalf("NewOpaqueToken() returned %d bytes", len(tokenCopy))
	}
}

func TestNewOpaqueTokenBoundsNoProgressSource(t *testing.T) {
	result := make(chan error, 1)
	go func() {
		_, err := securitykit.NewOpaqueToken(noProgressRandom{})
		result <- err
	}()

	select {
	case err := <-result:
		if !errors.Is(err, securitykit.ErrRandomSource) {
			t.Fatalf("NewOpaqueToken() error = %v, want ErrRandomSource", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("NewOpaqueToken() did not bound a no-progress random source")
	}
}

func TestDecodeOpaqueTokenRejectsNonCanonicalOrWrongLengthValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"",
		"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
		"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh",
		"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh!",
	} {
		decoded, err := securitykit.DecodeOpaqueToken(value)
		if !errors.Is(err, securitykit.ErrInvalidToken) {
			t.Fatalf("DecodeOpaqueToken(%q) error = %v, want ErrInvalidToken", value, err)
		}
		decodedCopy := decoded.Copy()
		if len(decodedCopy) != 0 {
			clear(decodedCopy)
			t.Fatalf("DecodeOpaqueToken(%q) returned bytes", value)
		}
		if bytes.Contains([]byte(err.Error()), []byte(value)) && len(value) >= 8 {
			t.Fatalf("DecodeOpaqueToken() error disclosed input")
		}
	}
}

func TestDigestTokenSeparatesAllSevenDomains(t *testing.T) {
	t.Parallel()

	rawBytes := bytes.Repeat([]byte{0x5a}, 32)
	raw := secret.NewBytes(rawBytes)
	domains := []securitykit.TokenDomain{
		securitykit.AccountAccessToken,
		securitykit.AccountRefreshToken,
		securitykit.DeviceAccessToken,
		securitykit.DeviceRefreshToken,
		securitykit.EnrollmentGrantToken,
		securitykit.EmailVerificationToken,
		securitykit.RecoveryCodeToken,
	}
	seen := make(map[[32]byte]securitykit.TokenDomain, len(domains))
	for _, domain := range domains {
		digest := securitykit.DigestToken(domain, raw)
		if previous, exists := seen[digest]; exists {
			t.Fatalf("domains %q and %q produced the same digest", previous, domain)
		}
		seen[digest] = domain
	}
	if len(seen) != 7 {
		t.Fatalf("unique digest count = %d, want 7", len(seen))
	}
	rawCopy := raw.Copy()
	defer clear(rawCopy)
	if !bytes.Equal(rawCopy, rawBytes) {
		t.Fatal("DigestToken mutated caller-owned secret")
	}
}

func TestDigestTokenFailsClosedForUnknownDomain(t *testing.T) {
	t.Parallel()

	raw := secret.NewBytes(bytes.Repeat([]byte{0x7a}, 32))
	if digest := securitykit.DigestToken(securitykit.TokenDomain("unknown"), raw); digest != ([32]byte{}) {
		t.Fatalf("DigestToken() unknown domain = %x, want all zero", digest)
	}
	rawCopy := raw.Copy()
	defer clear(rawCopy)
	if !bytes.Equal(rawCopy, bytes.Repeat([]byte{0x7a}, 32)) {
		t.Fatal("DigestToken() mutated raw token for unknown domain")
	}
}

var (
	_ securitykit.Clock        = fixedClock{}
	_ securitykit.RandomSource = (*chunkedRandom)(nil)
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Unix(0, 0).UTC() }
