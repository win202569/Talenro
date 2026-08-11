// Package secret prevents accidental formatting or JSON encoding of secret bytes.
package secret

import (
	"errors"
	"fmt"
	"log/slog"
)

// Bytes owns a private copy of sensitive bytes.
//
// Clear and Take require a unique owner. Ordinary Go value copies made before
// either operation retain an alias to the same backing storage and cannot have
// their slice headers invalidated by another copy.
type Bytes struct{ value []byte }

// NewBytes copies value into a redacted container.
func NewBytes(value []byte) Bytes {
	return Bytes{value: append([]byte(nil), value...)}
}

// Copy returns an isolated copy for a cryptographic adapter.
func (b Bytes) Copy() []byte { return append([]byte(nil), b.value...) }

// Clear immediately zeroizes the uniquely owned backing storage and releases it.
func (b *Bytes) Clear() {
	if b == nil {
		return
	}
	clear(b.value)
	b.value = nil
}

// Take moves unique ownership into a new Bytes and releases the source.
// Empty, repeated, and nil-source moves fail closed with an empty result.
func (b *Bytes) Take() Bytes {
	if b == nil || len(b.value) == 0 {
		if b != nil {
			b.value = nil
		}
		return Bytes{}
	}
	moved := Bytes{value: b.value}
	b.value = nil
	return moved
}

func (Bytes) String() string { return "[REDACTED]" }

// GoString prevents %#v from exposing the private byte slice.
func (Bytes) GoString() string { return "secret.Bytes([REDACTED])" }

// Format redacts every fmt verb, including byte-oriented verbs.
func (Bytes) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("[REDACTED]"))
}

// LogValue prevents structured handlers from inspecting or serializing secret bytes.
func (Bytes) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }

// MarshalJSON always fails closed.
func (Bytes) MarshalJSON() ([]byte, error) {
	return nil, errors.New("secret: serialization forbidden")
}
