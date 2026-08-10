// Package identity contains account identity and credential primitives.
package identity

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maximumEmailBytes       = 254
	maximumEmailLocalBytes  = 64
	maximumEmailDomainBytes = 253
)

// ErrInvalidEmail reports a malformed email without retaining its value.
var ErrInvalidEmail = errors.New("identity: invalid email")

// CanonicalEmail owns a canonical email representation without exposing it to
// ordinary formatting or structured logging.
type CanonicalEmail struct {
	value []byte
}

// CanonicalizeEmail validates one SMTPUTF8 dot-atom mailbox. It preserves the
// local part byte-for-byte and lowercases only an ASCII DNS domain. OpenAPI's
// email format is not the idn-email format, so callers must perform an explicit
// future IDNA protocol migration before accepting internationalized domains.
func CanonicalizeEmail(value string) (CanonicalEmail, error) {
	if len(value) < 3 || len(value) > maximumEmailBytes || !utf8.ValidString(value) || strings.IndexByte(value, 0) >= 0 {
		return CanonicalEmail{}, ErrInvalidEmail
	}
	if strings.Count(value, "@") != 1 {
		return CanonicalEmail{}, ErrInvalidEmail
	}
	separator := strings.IndexByte(value, '@')
	local, domain := value[:separator], value[separator+1:]
	if !validEmailLocal(local) || !validEmailDomain(domain) {
		return CanonicalEmail{}, ErrInvalidEmail
	}
	canonical := make([]byte, 0, len(value))
	canonical = append(canonical, local...)
	canonical = append(canonical, '@')
	for index := range len(domain) {
		character := domain[index]
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		canonical = append(canonical, character)
	}
	return CanonicalEmail{value: canonical}, nil
}

// Bytes returns an isolated copy for lookup and field-protection adapters.
func (email CanonicalEmail) Bytes() []byte {
	return append([]byte(nil), email.value...)
}

// Format prevents all fmt verbs from exposing personally identifiable data.
func (CanonicalEmail) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("[REDACTED-EMAIL]"))
}

// LogValue prevents slog from reflecting the private representation.
func (CanonicalEmail) LogValue() slog.Value {
	return slog.StringValue("[REDACTED-EMAIL]")
}

func validEmailLocal(local string) bool {
	if len(local) == 0 || len(local) > maximumEmailLocalBytes || local[0] == '.' || local[len(local)-1] == '.' || strings.Contains(local, "..") {
		return false
	}
	for _, character := range local {
		if character > unicode.MaxASCII {
			if unicode.IsControl(character) || unicode.IsSpace(character) {
				return false
			}
			continue
		}
		if isEmailAtomASCII(character) || character == '.' {
			continue
		}
		return false
	}
	return true
}

func isEmailAtomASCII(character rune) bool {
	if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' {
		return true
	}
	return strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", rune(character))
}

func validEmailDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > maximumEmailDomainBytes || domain[0] == '.' || domain[len(domain)-1] == '.' {
		return false
	}
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for index := range len(label) {
			character := label[index]
			if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' {
				continue
			}
			return false
		}
	}
	return true
}
