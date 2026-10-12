package serving

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func validLookupInput() CertificateLookupInput {
	return CertificateLookupInput{NodeID: uuid.MustParse("12345678-1234-4234-8234-123456789abc"), IssuerID: "issuer_A:1-b", SerialBytes: []byte{0x80, 1}, LeafDERDigest: contracts.Digest{1}, PublicKeyDigest: contracts.Digest{2}}
}

// Removing any shape check must admit its corresponding malformed query.
func TestCertificateLookupRejectsInvalidInput(t *testing.T) {
	cases := map[string]func(*CertificateLookupInput){
		"nil node":          func(i *CertificateLookupInput) { i.NodeID = uuid.Nil },
		"non RFC node":      func(i *CertificateLookupInput) { i.NodeID[8] = 0 },
		"empty issuer":      func(i *CertificateLookupInput) { i.IssuerID = "" },
		"long issuer":       func(i *CertificateLookupInput) { i.IssuerID = strings.Repeat("a", 129) },
		"issuer whitespace": func(i *CertificateLookupInput) { i.IssuerID = "issuer " },
		"issuer dot":        func(i *CertificateLookupInput) { i.IssuerID = "issuer.a" },
		"issuer unicode":    func(i *CertificateLookupInput) { i.IssuerID = "颁发者" },
		"nil serial":        func(i *CertificateLookupInput) { i.SerialBytes = nil },
		"empty serial":      func(i *CertificateLookupInput) { i.SerialBytes = []byte{} },
		"zero serial":       func(i *CertificateLookupInput) { i.SerialBytes = []byte{0} },
		"padded serial":     func(i *CertificateLookupInput) { i.SerialBytes = []byte{0, 1} },
		"long serial":       func(i *CertificateLookupInput) { i.SerialBytes = bytes.Repeat([]byte{1}, 21) },
		"zero leaf":         func(i *CertificateLookupInput) { i.LeafDERDigest = contracts.Digest{} },
		"zero key":          func(i *CertificateLookupInput) { i.PublicKeyDigest = contracts.Digest{} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			i := validLookupInput()
			mutate(&i)
			for _, agent := range []bool{false, true} {
				var got CertificateLookup
				var err error
				if agent {
					got, err = NewAgentCertificateLookup(i, "nodes.example")
				} else {
					got, err = NewCertificateLookup(i)
				}
				if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, CertificateLookup{}) {
					t.Fatalf("agent=%v got=%+v err=%v", agent, got, err)
				}
			}
		})
	}
}

func TestCertificateLookupValidBoundaries(t *testing.T) {
	for _, serial := range [][]byte{{1}, {0x80}, bytes.Repeat([]byte{0xff}, 20)} {
		i := validLookupInput()
		i.SerialBytes = serial
		i.IssuerID = strings.Repeat("A", 128)
		got, err := NewCertificateLookup(i)
		if err != nil || got.Validate() != nil || !reflect.DeepEqual(got.Input(), i) {
			t.Fatalf("valid input rejected: %v", err)
		}
		if selector, ok := got.AgentTrustSelector(); ok || selector != (AgentTrustSelector{}) {
			t.Fatal("non-agent lookup acquired a trust selector")
		}
	}
}

func TestCertificateLookupAgentDomain(t *testing.T) {
	maxDomain := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	for _, domain := range []string{"a", "node-1.example", maxDomain} {
		got, err := NewAgentCertificateLookup(validLookupInput(), domain)
		if err != nil || got.Validate() != nil {
			t.Fatalf("valid domain rejected: %v", err)
		}
		s, ok := got.AgentTrustSelector()
		if !ok || s.Purpose != "agent_server" || s.ListenerKind != "agent" || s.TrustDomain != domain {
			t.Fatalf("wrong selector: %+v", s)
		}
	}
	for _, domain := range []string{"", "Example.org", " example.org", "example.org ", "a.", ".a", "a..b", "-a", "a-", "a_b", "例子.org", "a/b", strings.Repeat("a", 64), maxDomain + "a"} {
		got, err := NewAgentCertificateLookup(validLookupInput(), domain)
		if err != authority.ErrInvalidArgument || !reflect.DeepEqual(got, CertificateLookup{}) {
			t.Fatalf("malformed domain accepted: %q", domain)
		}
	}
}

// Aliasing either constructor input or getter output would change the query.
func TestCertificateLookupOwnsInput(t *testing.T) {
	i := validLookupInput()
	want := validLookupInput()
	got, err := NewAgentCertificateLookup(i, "nodes.example")
	if err != nil {
		t.Fatal(err)
	}
	i.SerialBytes[0] = 0
	i.LeafDERDigest[0] = 0
	i.PublicKeyDigest[0] = 0
	i.NodeID = uuid.Nil
	out := got.Input()
	out.SerialBytes[0] = 0
	out.LeafDERDigest[0] = 0
	s, _ := got.AgentTrustSelector()
	s.Purpose = "operator_server"
	s.TrustDomain = "other.example"
	if got.Validate() != nil || !reflect.DeepEqual(got.Input(), want) {
		t.Fatal("caller mutation changed validated query")
	}
	s, _ = got.AgentTrustSelector()
	if s.Purpose != "agent_server" || s.TrustDomain != "nodes.example" {
		t.Fatal("caller mutation changed selector")
	}
}

func TestCertificateLookupZeroFailsClosed(t *testing.T) {
	var got CertificateLookup
	if got.Validate() != authority.ErrInvalidArgument {
		t.Fatal("zero lookup accepted")
	}
	if _, ok := got.AgentTrustSelector(); ok {
		t.Fatal("zero lookup has agent selector")
	}
}
