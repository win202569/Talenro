package secret

import (
	"encoding/json"
	"fmt"
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
