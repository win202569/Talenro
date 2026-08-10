package identity

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestEmailCanonicalizationLowercasesOnlyASCIIDomain(t *testing.T) {
	email, err := CanonicalizeEmail("Mi.Xed+tag@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("CanonicalizeEmail: %v", err)
	}
	if got := string(email.Bytes()); got != "Mi.Xed+tag@example.com" {
		t.Fatalf("canonical email = %q", got)
	}
}

func TestEmailCanonicalizationPreservesUTF8LocalPart(t *testing.T) {
	email, err := CanonicalizeEmail("用户+Tag@Example.COM")
	if err != nil {
		t.Fatalf("CanonicalizeEmail: %v", err)
	}
	if got := string(email.Bytes()); got != "用户+Tag@example.com" {
		t.Fatalf("canonical email = %q", got)
	}
}

func TestEmailCanonicalizationRejectsMalformedInput(t *testing.T) {
	tooLong := strings.Repeat("a", 249) + "@x.com"
	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "missing at", value: "local.example.com"},
		{name: "multiple at", value: "a@b@example.com"},
		{name: "empty local", value: "@example.com"},
		{name: "empty domain", value: "local@"},
		{name: "leading dot", value: ".local@example.com"},
		{name: "trailing dot", value: "local.@example.com"},
		{name: "double dot", value: "lo..cal@example.com"},
		{name: "whitespace", value: " local@example.com"},
		{name: "quoted local", value: "\"local\"@example.com"},
		{name: "domain underscore", value: "local@exam_ple.com"},
		{name: "unicode domain", value: "local@例子.test"},
		{name: "domain leading hyphen", value: "local@-example.com"},
		{name: "domain trailing hyphen", value: "local@example-.com"},
		{name: "domain empty label", value: "local@example..com"},
		{name: "nul", value: "local@example.com\x00"},
		{name: "invalid utf8", value: string([]byte{'a', '@', 0xff, '.', 'c'})},
		{name: "over 254 bytes", value: tooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CanonicalizeEmail(test.value)
			if !errors.Is(err, ErrInvalidEmail) {
				t.Fatalf("error = %v, want ErrInvalidEmail", err)
			}
			if err != nil && test.value != "" && strings.Contains(err.Error(), test.value) {
				t.Fatal("error exposed email input")
			}
		})
	}
}

func TestEmailCanonicalizationOwnsBytesAndRedactsFormatting(t *testing.T) {
	email, err := CanonicalizeEmail("Secret.Local@example.com")
	if err != nil {
		t.Fatalf("CanonicalizeEmail: %v", err)
	}
	first := email.Bytes()
	first[0] = 'X'
	if got := string(email.Bytes()); got != "Secret.Local@example.com" {
		t.Fatalf("mutated canonical email = %q", got)
	}
	for _, formatted := range []string{fmt.Sprintf("%v", email), fmt.Sprintf("%#v", email)} {
		if strings.Contains(formatted, "Secret.Local") {
			t.Fatalf("format exposed email: %q", formatted)
		}
	}
}

func TestEmailCanonicalizationAcceptsExactly254Bytes(t *testing.T) {
	local := strings.Repeat("l", 64)
	domain := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 61)
	value := local + "@" + domain
	if len(value) != 254 {
		t.Fatalf("test fixture length = %d", len(value))
	}
	if _, err := CanonicalizeEmail(value); err != nil {
		t.Fatalf("CanonicalizeEmail at 254 bytes: %v", err)
	}
}
