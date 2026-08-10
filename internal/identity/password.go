package identity

import (
	"crypto/subtle"
	"errors"
	"io"
	"reflect"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"

	"talenro.local/platform/internal/securitykit"
)

const (
	minimumPasswordBytes = 12
	maximumPasswordBytes = 1024
	maximumEmptyReads    = 8
)

var (
	// ErrInvalidPassword reports an invalid password without including it.
	ErrInvalidPassword = errors.New("identity: invalid password")
	// ErrInvalidPasswordPolicy reports a policy outside the finite approved set.
	ErrInvalidPasswordPolicy = errors.New("identity: invalid password policy")
	// ErrInvalidPasswordCredential reports malformed stored credential state.
	ErrInvalidPasswordCredential = errors.New("identity: invalid password credential")
	// ErrInvalidRandomSource reports a nil entropy source.
	ErrInvalidRandomSource = errors.New("identity: invalid random source")
	// ErrRandomSource reports a sanitized entropy-source failure.
	ErrRandomSource = errors.New("identity: random source failure")
)

// PasswordPolicy identifies a finite, versioned Argon2id parameter set.
type PasswordPolicy struct {
	Version     uint32
	MemoryKiB   uint32
	Time        uint32
	Parallelism uint8
	SaltBytes   uint32
	TagBytes    uint32
}

// CurrentPasswordPolicy is RFC 9106 section 4's second recommended option.
func CurrentPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		Version:     1,
		MemoryKiB:   65536,
		Time:        3,
		Parallelism: 4,
		SaltBytes:   16,
		TagBytes:    32,
	}
}

// PasswordCredential owns a versioned salt and Argon2id tag.
type PasswordCredential struct {
	policy PasswordPolicy
	salt   []byte
	hash   []byte
	dummy  bool
}

// NewPasswordCredential validates and copies a credential loaded from storage.
func NewPasswordCredential(policy PasswordPolicy, salt, hash []byte) (PasswordCredential, error) {
	if !approvedPasswordPolicy(policy) || len(salt) != int(policy.SaltBytes) || len(hash) != int(policy.TagBytes) {
		return PasswordCredential{}, ErrInvalidPasswordCredential
	}
	return PasswordCredential{
		policy: policy,
		salt:   append([]byte(nil), salt...),
		hash:   append([]byte(nil), hash...),
	}, nil
}

// Policy returns the immutable credential policy value.
func (credential PasswordCredential) Policy() PasswordPolicy { return credential.policy }

// SaltCopy returns an isolated salt for persistence.
func (credential PasswordCredential) SaltCopy() []byte {
	return append([]byte(nil), credential.salt...)
}

// HashCopy returns an isolated Argon2id tag for persistence.
func (credential PasswordCredential) HashCopy() []byte {
	return append([]byte(nil), credential.hash...)
}

// PasswordDeriver is the exact injectable boundary used to verify constant-work
// control flow without weakening production Argon2id.
type PasswordDeriver func(password, salt []byte, policy PasswordPolicy) []byte

// HashPassword validates and hashes one password with a fresh random salt.
func HashPassword(random securitykit.RandomSource, password []byte, policy PasswordPolicy) (PasswordCredential, error) {
	if !approvedPasswordPolicy(policy) {
		return PasswordCredential{}, ErrInvalidPasswordPolicy
	}
	if !validPassword(password) {
		return PasswordCredential{}, ErrInvalidPassword
	}
	if nilRandomSource(random) {
		return PasswordCredential{}, ErrInvalidRandomSource
	}

	passwordCopy := append([]byte(nil), password...)
	defer clear(passwordCopy)
	salt := make([]byte, policy.SaltBytes)
	bounded := &boundedRandomSource{source: random}
	count, err := io.ReadFull(bounded, salt)
	if err != nil || count != len(salt) {
		clear(salt)
		return PasswordCredential{}, ErrRandomSource
	}
	derived := deriveArgon2id(passwordCopy, salt, policy)
	hash := append([]byte(nil), derived...)
	clear(derived)
	return PasswordCredential{policy: policy, salt: salt, hash: hash}, nil
}

// VerifyPassword performs exactly one approved-cost Argon2id derivation.
func VerifyPassword(password []byte, credential PasswordCredential, current PasswordPolicy) (match, needsUpgrade bool) {
	return VerifyPasswordWithDeriver(password, credential, current, deriveArgon2id)
}

// VerifyPasswordWithDeriver exposes only the derivation boundary for tests.
// Invalid passwords, policies, and stored credentials use safe dummy inputs and
// still perform exactly one current-cost derivation.
func VerifyPasswordWithDeriver(password []byte, credential PasswordCredential, current PasswordPolicy, derive PasswordDeriver) (match, needsUpgrade bool) {
	approvedCurrent := approvedPasswordPolicy(current)
	workPolicy := current
	if !approvedCurrent {
		workPolicy = CurrentPasswordPolicy()
	}
	if derive == nil {
		derive = deriveArgon2id
	}

	passwordOK := validPassword(password)
	var passwordCopy []byte
	if passwordOK {
		passwordCopy = append([]byte(nil), password...)
	} else {
		passwordCopy = append([]byte(nil), dummyPassword...)
	}
	defer clear(passwordCopy)

	credentialOK := approvedPasswordPolicy(credential.policy) &&
		len(credential.salt) == int(credential.policy.SaltBytes) &&
		len(credential.hash) == int(credential.policy.TagBytes)
	workCredential := credential
	if !credentialOK {
		workCredential = DummyCredential()
	}
	salt := append([]byte(nil), workCredential.salt...)
	expected := append([]byte(nil), workCredential.hash...)
	defer clear(salt)
	defer clear(expected)

	derived := derive(passwordCopy, salt, workPolicy)
	defer clear(derived)
	equal := subtle.ConstantTimeCompare(derived, expected) == 1
	match = approvedCurrent && passwordOK && credentialOK && !workCredential.dummy && equal
	// C1.1 approves only policy v1, so every successful match is current and
	// needsUpgrade remains false. A future version must receive a separate
	// security review: verify legacy hashes with their finite approved policy
	// while retaining current-cost dummy work on every enumeration path.
	needsUpgrade = match && credential.policy != current
	return match, needsUpgrade
}

var (
	dummyPassword = []byte("Talenro-Dummy-Password-V1")
	dummySalt     = [16]byte{0x5e, 0x28, 0x1f, 0x93, 0xe2, 0x8b, 0xc6, 0x4a, 0x8d, 0x11, 0x7c, 0xb0, 0x39, 0x6a, 0xf4, 0x72}
	dummyHash     = [32]byte{0x8b, 0x42, 0x97, 0x6d, 0x2a, 0x31, 0xde, 0xc8, 0x61, 0x05, 0xb7, 0xf4, 0x98, 0x23, 0xa0, 0x6c, 0x55, 0xf1, 0x9b, 0x3d, 0x72, 0xce, 0x0e, 0xa4, 0x17, 0xd9, 0x63, 0xb5, 0x40, 0x2c, 0xee, 0x19}
)

// DummyCredential returns deterministic process data at the current cost. It
// is never accepted as a real credential, including if its tag happens to match.
func DummyCredential() PasswordCredential {
	return PasswordCredential{
		policy: CurrentPasswordPolicy(),
		salt:   append([]byte(nil), dummySalt[:]...),
		hash:   append([]byte(nil), dummyHash[:]...),
		dummy:  true,
	}
}

func deriveArgon2id(password, salt []byte, policy PasswordPolicy) []byte {
	passwordCopy := append([]byte(nil), password...)
	saltCopy := append([]byte(nil), salt...)
	defer clear(passwordCopy)
	defer clear(saltCopy)
	return argon2.IDKey(passwordCopy, saltCopy, policy.Time, policy.MemoryKiB, policy.Parallelism, policy.TagBytes)
}

func approvedPasswordPolicy(policy PasswordPolicy) bool {
	return policy == CurrentPasswordPolicy()
}

func validPassword(password []byte) bool {
	return len(password) >= minimumPasswordBytes && len(password) <= maximumPasswordBytes &&
		utf8.Valid(password) && !containsNUL(password)
}

func containsNUL(value []byte) bool {
	for _, character := range value {
		if character == 0 {
			return true
		}
	}
	return false
}

type boundedRandomSource struct {
	source     securitykit.RandomSource
	emptyReads int
}

func (random *boundedRandomSource) Read(target []byte) (int, error) {
	count, err := random.source.Read(target)
	if count < 0 || count > len(target) {
		return 0, io.ErrNoProgress
	}
	if count == 0 && err == nil {
		random.emptyReads++
		if random.emptyReads >= maximumEmptyReads {
			return 0, io.ErrNoProgress
		}
	} else {
		random.emptyReads = 0
	}
	return count, err
}

func nilRandomSource(random securitykit.RandomSource) bool {
	if random == nil {
		return true
	}
	value := reflect.ValueOf(random)
	//nolint:exhaustive // Only nil-capable interface representations are relevant.
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
