package secret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestBytesNeverFormatsOrMarshalsSecret(t *testing.T) {
	value := NewBytes([]byte("SECRET-CANARY"))
	if got := fmt.Sprintf("%v", value); got != "[REDACTED]" {
		t.Fatalf("formatted secret = %q", got)
	}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("secret unexpectedly marshaled")
	}
	copyValue := value.Copy()
	copyValue[0] = 'X'
	if string(value.Copy()) != "SECRET-CANARY" {
		t.Fatal("Copy exposed mutable backing storage")
	}
}

func TestBytesClearZeroizesOwnedBackingAndReleasesOwner(t *testing.T) {
	owner := NewBytes([]byte("CLEAR-SECRET-CANARY"))
	backing := owner.value

	owner.Clear()

	if owner.value != nil || len(owner.Copy()) != 0 {
		t.Fatal("Clear retained the source owner")
	}
	if !bytes.Equal(backing, make([]byte, len(backing))) {
		t.Fatalf("Clear retained owned backing bytes: %q", backing)
	}
	owner.Clear()
	if owner.value != nil {
		t.Fatal("repeated Clear changed the released state")
	}
}

func TestBytesTakeMovesUniqueOwnership(t *testing.T) {
	source := NewBytes([]byte("TAKE-SECRET-CANARY"))
	backing := source.value

	destination := source.Take()

	if source.value != nil || len(source.Copy()) != 0 {
		t.Fatal("Take retained the source owner")
	}
	destinationCopy := destination.Copy()
	defer clear(destinationCopy)
	if string(destinationCopy) != "TAKE-SECRET-CANARY" {
		t.Fatalf("Take destination = %q", destinationCopy)
	}

	destination.Clear()
	if destination.value != nil || !bytes.Equal(backing, make([]byte, len(backing))) {
		t.Fatalf("destination Clear retained moved bytes: %q", backing)
	}
}

func TestBytesTakeFailsClosedForEmptyNilAndRepeatedSources(t *testing.T) {
	source := NewBytes([]byte("SINGLE-MOVE-CANARY"))
	first := source.Take()
	defer first.Clear()
	if second := source.Take(); second.value != nil || len(second.Copy()) != 0 {
		t.Fatal("repeated Take returned an owner")
	}

	var empty Bytes
	if moved := empty.Take(); moved.value != nil || len(moved.Copy()) != 0 {
		t.Fatal("empty Take returned an owner")
	}
	var nilSource *Bytes
	if moved := nilSource.Take(); moved.value != nil || len(moved.Copy()) != 0 {
		t.Fatal("nil Take returned an owner")
	}
}

func TestBytesOwnershipOperationsRequireUniqueOwner(t *testing.T) {
	source := NewBytes([]byte("UNIQUE-OWNER-CANARY"))
	preexistingCopy := source
	destination := source.Take()

	if source.value != nil {
		t.Fatal("Take retained the designated source owner")
	}
	copyBytes := preexistingCopy.Copy()
	if string(copyBytes) != "UNIQUE-OWNER-CANARY" {
		clear(copyBytes)
		t.Fatalf("Take claimed to invalidate an arbitrary preexisting copy: %q", copyBytes)
	}
	clear(copyBytes)

	destination.Clear()
	copyBytes = preexistingCopy.Copy()
	defer clear(copyBytes)
	if !bytes.Equal(copyBytes, make([]byte, len(copyBytes))) {
		t.Fatalf("moved owner Clear did not zero shared backing: %q", copyBytes)
	}
}

func TestBytesRedactsFmtSlogAndJSON(t *testing.T) {
	const canary = "REDACTION-SECRET-CANARY"
	value := NewBytes([]byte(canary))
	defer value.Clear()
	valueCopy := value
	var nilValue *Bytes
	subjects := []struct {
		name  string
		value any
	}{
		{name: "value", value: valueCopy},
		{name: "pointer", value: &value},
		{name: "zero", value: Bytes{}},
		{name: "nil pointer", value: nilValue},
	}
	for _, subject := range subjects {
		for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X"} {
			rendered := fmt.Sprintf(format, subject.value)
			if strings.Contains(rendered, canary) {
				t.Fatalf("%s %s exposed secret: %q", subject.name, format, rendered)
			}
		}
	}

	handlers := []struct {
		name string
		new  func(*bytes.Buffer) slog.Handler
	}{
		{name: "text", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewTextHandler(buffer, nil) }},
		{name: "json", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewJSONHandler(buffer, nil) }},
	}
	for _, handler := range handlers {
		var output bytes.Buffer
		slog.New(handler.new(&output)).Info("secret state", "value", valueCopy, "pointer", &value, "zero", Bytes{}, "nil", nilValue)
		if rendered := output.String(); strings.Contains(rendered, canary) || !strings.Contains(rendered, "[REDACTED]") {
			t.Fatalf("%s slog output did not redact: %q", handler.name, rendered)
		}
	}

	for _, subject := range []any{valueCopy, &value, Bytes{}} {
		if encoded, err := json.Marshal(subject); err == nil || bytes.Contains(encoded, []byte(canary)) {
			t.Fatalf("json.Marshal(%T) = %q, %v", subject, encoded, err)
		}
	}
}
