package serving

import (
	"bytes"
	"github.com/google/uuid"
	"strings"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

// CertificateLookupInput is certificate-derived query identity, not proof of
// authorization. Epoch and lineage must come from verified database facts.
type CertificateLookupInput struct {
	NodeID          uuid.UUID
	IssuerID        string
	SerialBytes     []byte
	LeafDERDigest   contracts.Digest
	PublicKeyDigest contracts.Digest
}

// AgentTrustSelector is a detached description of the constructor-fixed scope.
type AgentTrustSelector struct{ Purpose, ListenerKind, TrustDomain string }

// CertificateLookup owns its inputs. Construction validates only their shape;
// it does not verify a certificate or enable a production serving entry point.
type CertificateLookup struct {
	input       CertificateLookupInput
	mode        uint8
	trustDomain string
}

// NewCertificateLookup constructs a non-agent query without a trust selector.
func NewCertificateLookup(input CertificateLookupInput) (CertificateLookup, error) {
	return newCertificateLookup(input, 1, "")
}

// NewAgentCertificateLookup fixes the agent listener and server trust purpose.
// configuredTrustDomain must be the listener's exact configured domain. No
// trimming or case folding is performed, and caller provenance is not proven.
func NewAgentCertificateLookup(input CertificateLookupInput, configuredTrustDomain string) (CertificateLookup, error) {
	return newCertificateLookup(input, 2, configuredTrustDomain)
}

func newCertificateLookup(input CertificateLookupInput, mode uint8, domain string) (CertificateLookup, error) {
	input.SerialBytes = bytes.Clone(input.SerialBytes)
	lookup := CertificateLookup{input: input, mode: mode, trustDomain: domain}
	if err := lookup.Validate(); err != nil {
		return CertificateLookup{}, err
	}
	return lookup, nil
}

// Validate rejects the zero value as well as malformed query identities.
func (l CertificateLookup) Validate() error {
	i := l.input
	if i.NodeID == uuid.Nil || i.NodeID.Variant() != uuid.RFC4122 ||
		len(i.IssuerID) < 1 || len(i.IssuerID) > 128 ||
		len(i.SerialBytes) < 1 || len(i.SerialBytes) > 20 || i.SerialBytes[0] == 0 ||
		i.LeafDERDigest == (contracts.Digest{}) || i.PublicKeyDigest == (contracts.Digest{}) {
		return authority.ErrInvalidArgument
	}
	for _, c := range []byte(i.IssuerID) {
		if !lowerDigit(c) && !(c >= 'A' && c <= 'Z') && c != ':' && c != '_' && c != '-' {
			return authority.ErrInvalidArgument
		}
	}
	switch l.mode {
	case 1:
		if l.trustDomain != "" {
			return authority.ErrInvalidArgument
		}
	case 2:
		if !validLookupTrustDomain(l.trustDomain) {
			return authority.ErrInvalidArgument
		}
	default:
		return authority.ErrInvalidArgument
	}
	return nil
}

// Input returns an owned copy, including serial bytes.
func (l CertificateLookup) Input() CertificateLookupInput {
	i := l.input
	i.SerialBytes = bytes.Clone(i.SerialBytes)
	return i
}

// AgentTrustSelector returns no selector for invalid or non-agent lookups.
func (l CertificateLookup) AgentTrustSelector() (AgentTrustSelector, bool) {
	if l.mode != 2 || l.Validate() != nil {
		return AgentTrustSelector{}, false
	}
	return AgentTrustSelector{Purpose: "agent_server", ListenerKind: "agent", TrustDomain: l.trustDomain}, true
}

func validLookupTrustDomain(domain string) bool {
	if len(domain) < 1 || len(domain) > 253 {
		return false
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) < 1 || len(label) > 63 || !lowerDigit(label[0]) || !lowerDigit(label[len(label)-1]) {
			return false
		}
		for _, c := range []byte(label) {
			if !lowerDigit(c) && c != '-' {
				return false
			}
		}
	}
	return true
}

func lowerDigit(c byte) bool { return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' }
