package identity

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"

	"talenro.local/platform/internal/strictjson"
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

func TestVerifyPasswordInvalidInputsUseBoundedDummyWork(t *testing.T) {
	credential := DummyCredential()
	tests := []struct {
		name     string
		password []byte
	}{
		{name: "nil", password: nil},
		{name: "short", password: []byte("short")},
		{name: "invalid utf8", password: append(bytes.Repeat([]byte{'a'}, 12), 0xff)},
		{name: "over maximum", password: bytes.Repeat([]byte{'p'}, 1025)},
		{name: "attacker sized", password: bytes.Repeat([]byte{'p'}, 2<<20)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append([]byte(nil), test.password...)
			calls := 0
			derive := func(password, salt []byte, policy PasswordPolicy) []byte {
				calls++
				if string(password) != string(dummyPassword) {
					t.Fatalf("work password length = %d, want fixed dummy", len(password))
				}
				if policy != CurrentPasswordPolicy() || len(salt) != 16 {
					t.Fatalf("unsafe dummy work policy=%+v salt=%d", policy, len(salt))
				}
				return make([]byte, 32)
			}
			match, needsUpgrade := VerifyPasswordWithDeriver(test.password, credential, CurrentPasswordPolicy(), derive)
			if calls != 1 || match || needsUpgrade {
				t.Fatalf("verify = calls:%d match:%v upgrade:%v", calls, match, needsUpgrade)
			}
			if !bytes.Equal(test.password, original) {
				t.Fatal("VerifyPassword modified caller input")
			}
		})
	}
}

func TestVerifyPasswordRejectsAttackerSizedInputWithoutCopyingIt(t *testing.T) {
	password := bytes.Repeat([]byte{'p'}, 2<<20)
	credential := DummyCredential()
	result := testing.Benchmark(func(benchmark *testing.B) {
		derive := func(_, _ []byte, _ PasswordPolicy) []byte { return make([]byte, 32) }
		for benchmark.Loop() {
			VerifyPasswordWithDeriver(password, credential, CurrentPasswordPolicy(), derive)
		}
	})
	const maximumVerificationAllocationBytes = 128 << 10
	if allocated := result.AllocedBytesPerOp(); allocated > maximumVerificationAllocationBytes {
		t.Fatalf("allocated bytes/op = %d, want <= %d", allocated, maximumVerificationAllocationBytes)
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

func TestVerifyPasswordSuccessfulCurrentV1NeverNeedsUpgrade(t *testing.T) {
	credential, err := NewPasswordCredential(
		CurrentPasswordPolicy(),
		bytes.Repeat([]byte{0x11}, 16),
		bytes.Repeat([]byte{0x22}, 32),
	)
	if err != nil {
		t.Fatalf("NewPasswordCredential: %v", err)
	}
	derive := func(_, _ []byte, policy PasswordPolicy) []byte {
		if policy != CurrentPasswordPolicy() {
			t.Fatalf("derive policy = %+v", policy)
		}
		return bytes.Repeat([]byte{0x22}, 32)
	}
	match, needsUpgrade := VerifyPasswordWithDeriver([]byte("valid-password"), credential, CurrentPasswordPolicy(), derive)
	if !match || needsUpgrade {
		t.Fatalf("verify = (%v, %v), want (true, false)", match, needsUpgrade)
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

const (
	maximumArgon2VectorBytes int64 = 4096
	argon2VectorSource             = "P-H-C/phc-winner-argon2 reference C implementation commit f57e61e19229e23c4445b85494dbf7c07de721cb; argon2id_hash_raw(t=3,m=256,p=2,password=password,salt=somesalt,tag=32,v=19)"
	// #nosec G101 -- public known-answer test input, not a credential.
	argon2VectorPasswordHex = "70617373776f7264"
	argon2VectorSaltHex     = "736f6d6573616c74"
	argon2VectorExpectedHex = "a3161de99d0e7c0762364b2c4b3ea2b950005973f8879d54287fd8bd56921f36"
)

var errInvalidArgon2Vector = errors.New("identity test: invalid Argon2 vector")

type checkedArgon2Vector struct {
	password []byte
	salt     []byte
	expected []byte
	params   argon2Vector
}

type argon2VectorDeriver func(password, salt []byte, time, memoryKiB uint32, parallelism uint8, tagBytes uint32) []byte

func decodeArgon2Vector(reader io.Reader) (checkedArgon2Vector, error) {
	var vector argon2Vector
	if err := strictjson.Decode(reader, maximumArgon2VectorBytes, &vector); err != nil {
		return checkedArgon2Vector{}, err
	}
	if vector.Source != argon2VectorSource || vector.Variant != "Argon2id" || vector.Version != argon2.Version ||
		vector.PasswordHex != argon2VectorPasswordHex || vector.SaltHex != argon2VectorSaltHex ||
		vector.Time != 3 || vector.MemoryKiB != 256 || vector.Parallelism != 2 || vector.TagBytes != 32 ||
		vector.ExpectedHex != argon2VectorExpectedHex {
		return checkedArgon2Vector{}, errInvalidArgon2Vector
	}
	password, passwordErr := hex.DecodeString(vector.PasswordHex)
	salt, saltErr := hex.DecodeString(vector.SaltHex)
	expected, expectedErr := hex.DecodeString(vector.ExpectedHex)
	if passwordErr != nil || saltErr != nil || expectedErr != nil || len(password) != 8 || len(salt) != 8 || len(expected) != 32 {
		clear(password)
		clear(salt)
		clear(expected)
		return checkedArgon2Vector{}, errInvalidArgon2Vector
	}
	return checkedArgon2Vector{password: password, salt: salt, expected: expected, params: vector}, nil
}

func verifyArgon2Vector(reader io.Reader, derive argon2VectorDeriver) error {
	vector, err := decodeArgon2Vector(reader)
	if err != nil {
		return err
	}
	defer clear(vector.password)
	defer clear(vector.salt)
	defer clear(vector.expected)
	got := derive(
		vector.password,
		vector.salt,
		vector.params.Time,
		vector.params.MemoryKiB,
		vector.params.Parallelism,
		vector.params.TagBytes,
	)
	defer clear(got)
	if !bytes.Equal(got, vector.expected) {
		return errInvalidArgon2Vector
	}
	return nil
}

func TestPasswordArgon2idCheckedInKnownVector(t *testing.T) {
	fixture, err := os.Open("../../testdata/crypto/rfc9106/argon2id-v1.json")
	if err != nil {
		t.Fatalf("open vector: %v", err)
	}
	t.Cleanup(func() { _ = fixture.Close() })
	derive := func(password, salt []byte, time, memoryKiB uint32, parallelism uint8, tagBytes uint32) []byte {
		return argon2.IDKey(password, salt, time, memoryKiB, parallelism, tagBytes)
	}
	if err := verifyArgon2Vector(fixture, derive); err != nil {
		t.Fatalf("verify vector: %v", err)
	}
}

func TestPasswordArgon2idRejectsMutatedFixtureBeforeDerivation(t *testing.T) {
	encoded, err := os.ReadFile("../../testdata/crypto/rfc9106/argon2id-v1.json")
	if err != nil {
		t.Fatalf("read vector: %v", err)
	}
	replace := func(old, replacement string) []byte {
		t.Helper()
		mutated := bytes.Replace(encoded, []byte(old), []byte(replacement), 1)
		if bytes.Equal(mutated, encoded) {
			t.Fatalf("mutation source not found: %q", old)
		}
		return mutated
	}
	tests := []struct {
		name string
		body []byte
	}{
		{name: "unknown member", body: replace(`"variant": "Argon2id",`, `"variant": "Argon2id", "unknown": true,`)},
		{name: "duplicate member", body: replace(`"time": 3,`, `"time": 3, "time": 3,`)},
		{name: "wrong source", body: replace(argon2VectorSource, "untrusted source")},
		{name: "wrong variant", body: replace(`"variant": "Argon2id"`, `"variant": "Argon2i"`)},
		{name: "wrong version", body: replace(`"version": 19`, `"version": 16`)},
		{name: "unsafe time", body: replace(`"time": 3`, `"time": 4294967295`)},
		{name: "unsafe memory", body: replace(`"memory_kib": 256`, `"memory_kib": 4294967295`)},
		{name: "unsafe parallelism", body: replace(`"parallelism": 2`, `"parallelism": 255`)},
		{name: "wrong tag bytes", body: replace(`"tag_bytes": 32`, `"tag_bytes": 24`)},
		{name: "wrong password", body: replace(argon2VectorPasswordHex, "70617373")},
		{name: "wrong salt", body: replace(argon2VectorSaltHex, "73616c74")},
		{name: "wrong expected", body: replace(argon2VectorExpectedHex, strings.Repeat("00", 32))},
		{name: "over byte limit", body: bytes.Repeat([]byte{' '}, int(maximumArgon2VectorBytes)+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			derive := func(_, _ []byte, _, _ uint32, _ uint8, _ uint32) []byte {
				calls++
				return make([]byte, 32)
			}
			if err := verifyArgon2Vector(bytes.NewReader(test.body), derive); err == nil {
				t.Fatal("mutated vector verified")
			}
			if calls != 0 {
				t.Fatalf("Argon2 calls = %d, want 0", calls)
			}
		})
	}
}
