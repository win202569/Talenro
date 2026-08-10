package identity

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

type fixedRandom struct {
	data []byte
	err  error
	step int
}

func (r *fixedRandom) Read(target []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.step <= 0 || r.step > len(target) {
		r.step = len(target)
	}
	if len(r.data) == 0 {
		return 0, nil
	}
	n := r.step
	if n > len(r.data) {
		n = len(r.data)
	}
	copy(target, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

func TestPasswordCurrentPolicyIsRFC9106SecondRecommendation(t *testing.T) {
	want := PasswordPolicy{Version: 1, MemoryKiB: 65536, Time: 3, Parallelism: 4, SaltBytes: 16, TagBytes: 32}
	if got := CurrentPasswordPolicy(); got != want {
		t.Fatalf("policy = %+v, want %+v", got, want)
	}
}

func TestPasswordHashAndVerifyRealArgon2id(t *testing.T) {
	password := []byte("correct horse battery staple")
	random := &fixedRandom{data: bytes.Repeat([]byte{0x42}, 16), step: 3}
	credential, err := HashPassword(random, password, CurrentPasswordPolicy())
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	match, needsUpgrade := VerifyPassword(password, credential, CurrentPasswordPolicy())
	if !match || needsUpgrade {
		t.Fatalf("verify = (%v, %v), want (true, false)", match, needsUpgrade)
	}
	match, needsUpgrade = VerifyPassword([]byte("incorrect password"), credential, CurrentPasswordPolicy())
	if match || needsUpgrade {
		t.Fatalf("wrong verify = (%v, %v), want (false, false)", match, needsUpgrade)
	}
}

func TestPasswordRejectsInvalidInputAndPolicy(t *testing.T) {
	validRandom := func() *fixedRandom {
		return &fixedRandom{data: bytes.Repeat([]byte{0x44}, 16), step: 16}
	}
	tests := []struct {
		name     string
		password []byte
		policy   PasswordPolicy
	}{
		{name: "short", password: []byte("eleven-byte"), policy: CurrentPasswordPolicy()},
		{name: "over 1024", password: bytes.Repeat([]byte{'p'}, 1025), policy: CurrentPasswordPolicy()},
		{name: "invalid utf8", password: []byte{'v', 'a', 'l', 'i', 'd', '-', 'l', 'e', 'n', 'g', 't', 'h', 0xff}, policy: CurrentPasswordPolicy()},
		{name: "nul", password: []byte("valid-length\x00password"), policy: CurrentPasswordPolicy()},
		{name: "zero memory", password: []byte("valid-password"), policy: PasswordPolicy{Version: 1, Time: 3, Parallelism: 4, SaltBytes: 16, TagBytes: 32}},
		{name: "huge memory", password: []byte("valid-password"), policy: PasswordPolicy{Version: 1, MemoryKiB: ^uint32(0), Time: 3, Parallelism: 4, SaltBytes: 16, TagBytes: 32}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := HashPassword(validRandom(), test.password, test.policy)
			if !errors.Is(err, ErrInvalidPassword) && !errors.Is(err, ErrInvalidPasswordPolicy) {
				t.Fatalf("error = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), string(test.password)) {
				t.Fatal("error exposed password")
			}
		})
	}
}

func TestPasswordHashFailsClosedForBadRandomSource(t *testing.T) {
	var typedNil *fixedRandom
	tests := []struct {
		name   string
		random interface{ Read([]byte) (int, error) }
	}{
		{name: "nil", random: nil},
		{name: "typed nil", random: typedNil},
		{name: "error", random: &fixedRandom{err: errors.New("SECRET-RANDOM-ERROR")}},
		{name: "no progress", random: &fixedRandom{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := HashPassword(test.random, []byte("valid-password"), CurrentPasswordPolicy())
			if !errors.Is(err, ErrRandomSource) && !errors.Is(err, ErrInvalidRandomSource) {
				t.Fatalf("error = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "SECRET-RANDOM-ERROR") {
				t.Fatal("error exposed random source failure")
			}
		})
	}
}

func TestVerifyPasswordUsesExactlyOneSafeDerivation(t *testing.T) {
	current := CurrentPasswordPolicy()
	valid := PasswordCredential{
		policy: current,
		salt:   bytes.Repeat([]byte{0x11}, 16),
		hash:   bytes.Repeat([]byte{0x22}, 32),
	}
	malformed := PasswordCredential{
		policy: PasswordPolicy{Version: 999, MemoryKiB: ^uint32(0), Time: ^uint32(0), Parallelism: 255, SaltBytes: ^uint32(0), TagBytes: ^uint32(0)},
		salt:   []byte{1},
		hash:   []byte{2},
	}
	tests := []struct {
		name       string
		password   []byte
		credential PasswordCredential
	}{
		{name: "real credential", password: []byte("valid-password"), credential: valid},
		{name: "dummy credential", password: []byte("valid-password"), credential: DummyCredential()},
		{name: "malformed credential", password: []byte("valid-password"), credential: malformed},
		{name: "malformed password", password: []byte{0xff}, credential: valid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			derive := func(password, salt []byte, policy PasswordPolicy) []byte {
				calls++
				if policy != current {
					t.Fatalf("unsafe derivation policy = %+v", policy)
				}
				if len(salt) != 16 || len(password) < 12 || len(password) > 1024 {
					t.Fatalf("unsafe derivation input lengths password=%d salt=%d", len(password), len(salt))
				}
				return bytes.Repeat([]byte{0x22}, 32)
			}
			match, needsUpgrade := VerifyPasswordWithDeriver(test.password, test.credential, current, derive)
			if calls != 1 {
				t.Fatalf("argon2 calls = %d, want 1", calls)
			}
			if test.name != "real credential" && (match || needsUpgrade) {
				t.Fatalf("invalid path verified = (%v, %v)", match, needsUpgrade)
			}
		})
	}
}

func TestPasswordCredentialOwnsReturnedSlices(t *testing.T) {
	random := &fixedRandom{data: bytes.Repeat([]byte{0x33}, 16), step: 16}
	credential, err := HashPassword(random, []byte("valid-password"), CurrentPasswordPolicy())
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	salt := credential.SaltCopy()
	hash := credential.HashCopy()
	salt[0] ^= 0xff
	hash[0] ^= 0xff
	if bytes.Equal(salt, credential.SaltCopy()) || bytes.Equal(hash, credential.HashCopy()) {
		t.Fatal("credential returned mutable backing storage")
	}
}

func TestNewPasswordCredentialValidatesAndOwnsStoredSlices(t *testing.T) {
	salt := bytes.Repeat([]byte{0x51}, 16)
	hash := bytes.Repeat([]byte{0x61}, 32)
	credential, err := NewPasswordCredential(CurrentPasswordPolicy(), salt, hash)
	if err != nil {
		t.Fatalf("NewPasswordCredential: %v", err)
	}
	salt[0] = 0
	hash[0] = 0
	if credential.Policy() != CurrentPasswordPolicy() || credential.SaltCopy()[0] != 0x51 || credential.HashCopy()[0] != 0x61 {
		t.Fatal("credential retained caller-owned storage")
	}
	if _, err := NewPasswordCredential(CurrentPasswordPolicy(), salt[:1], bytes.Repeat([]byte{1}, 32)); !errors.Is(err, ErrInvalidPasswordCredential) {
		t.Fatalf("short salt error = %v", err)
	}
}

type argon2Vector struct {
	Source      string `json:"source"`
	Variant     string `json:"variant"`
	Version     int    `json:"version"`
	PasswordHex string `json:"password_hex"`
	SaltHex     string `json:"salt_hex"`
	Time        uint32 `json:"time"`
	MemoryKiB   uint32 `json:"memory_kib"`
	Parallelism uint8  `json:"parallelism"`
	TagBytes    uint32 `json:"tag_bytes"`
	ExpectedHex string `json:"expected_hex"`
}

func TestPasswordArgon2idCheckedInKnownVector(t *testing.T) {
	encoded, err := os.ReadFile("../../testdata/crypto/rfc9106/argon2id-v1.json")
	if err != nil {
		t.Fatalf("read vector: %v", err)
	}
	var vector argon2Vector
	if err := json.Unmarshal(encoded, &vector); err != nil {
		t.Fatalf("decode vector: %v", err)
	}
	if vector.Source == "" || vector.Variant != "Argon2id" || vector.Version != argon2.Version {
		t.Fatalf("incomplete vector provenance: %+v", vector)
	}
	password, err := hex.DecodeString(vector.PasswordHex)
	if err != nil {
		t.Fatalf("password hex: %v", err)
	}
	salt, err := hex.DecodeString(vector.SaltHex)
	if err != nil {
		t.Fatalf("salt hex: %v", err)
	}
	want, err := hex.DecodeString(vector.ExpectedHex)
	if err != nil {
		t.Fatalf("expected hex: %v", err)
	}
	got := argon2.IDKey(password, salt, vector.Time, vector.MemoryKiB, vector.Parallelism, vector.TagBytes)
	if !bytes.Equal(got, want) {
		t.Fatalf("Argon2id tag = %x, want %x", got, want)
	}
}
