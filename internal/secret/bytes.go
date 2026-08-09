// Package secret prevents accidental formatting or JSON encoding of secret bytes.
package secret

import (
	"errors"
	"fmt"
)

// Bytes owns a private copy of sensitive bytes.
type Bytes struct{ value []byte }

// NewBytes copies value into a redacted container.
func NewBytes(value []byte) Bytes {
	return Bytes{value: append([]byte(nil), value...)}
}

// Copy returns an isolated copy for a cryptographic adapter.
func (b Bytes) Copy() []byte { return append([]byte(nil), b.value...) }

func (Bytes) String() string { return "[REDACTED]" }

// GoString prevents %#v from exposing the private byte slice.
func (Bytes) GoString() string { return "secret.Bytes([REDACTED])" }

// Format redacts every fmt verb, including byte-oriented verbs.
func (Bytes) Format(state fmt.State, verb rune) {
	_, _ = state.Write([]byte("[REDACTED]"))
}

// MarshalJSON always fails closed.
func (Bytes) MarshalJSON() ([]byte, error) {
	return nil, errors.New("secret: serialization forbidden")
}
