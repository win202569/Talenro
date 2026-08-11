package deviceauth

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	deviceProofPrefix          = "TALENRO-DEVICE-POP-V1\x00"
	deviceProofProtocolVersion = "device-pop-v1"
	registerDeviceOperation    = "register_device"
	maximumProofNameBytes      = 64
	maximumProofAudienceBytes  = 2048
)

var (
	// ErrInvalidProof reports malformed or unauthenticated proof input without retaining it.
	ErrInvalidProof = errors.New("deviceauth: invalid proof")
)

// ProofBytes returns the fixed, length-framed proof-of-possession transcript.
func ProofBytes(input ProofInput) []byte {
	if !validProofString(input.ProtocolVersion, maximumProofNameBytes) ||
		!validProofString(input.Operation, maximumProofNameBytes) ||
		!validProofString(input.Audience, maximumProofAudienceBytes) ||
		input.Challenge == [32]byte{} || input.GrantDigest == [32]byte{} ||
		input.SigningPublicKey == [32]byte{} || input.HPKEPublicKey == [32]byte{} || input.RequestNonce == [32]byte{} {
		return nil
	}
	result := make([]byte, 0, len(deviceProofPrefix)+12+len(input.ProtocolVersion)+len(input.Operation)+len(input.Audience)+32*5)
	result = append(result, deviceProofPrefix...)
	result = appendProofString(result, input.ProtocolVersion)
	result = append(result, input.Challenge[:]...)
	result = append(result, input.GrantDigest[:]...)
	result = append(result, input.SigningPublicKey[:]...)
	result = append(result, input.HPKEPublicKey[:]...)
	result = appendProofString(result, input.Operation)
	result = appendProofString(result, input.Audience)
	result = append(result, input.RequestNonce[:]...)
	return result
}

func verifyProof(input ProofInput, signature [64]byte, configuredAudience string) error {
	if input.ProtocolVersion != deviceProofProtocolVersion || input.Operation != registerDeviceOperation ||
		input.Audience != configuredAudience || signature == [64]byte{} || !validPublicOrigin(configuredAudience) {
		return ErrInvalidProof
	}
	transcript := ProofBytes(input)
	if len(transcript) == 0 {
		return ErrInvalidProof
	}
	defer clear(transcript)
	if !ed25519.Verify(ed25519.PublicKey(input.SigningPublicKey[:]), transcript, signature[:]) {
		return ErrInvalidProof
	}
	return nil
}

func appendProofString(target []byte, value string) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) // #nosec G115 -- caller validates every string below MaxUint32.
	target = append(target, length[:]...)
	clear(length[:])
	return append(target, value...)
}

func validProofString(value string, maximum int) bool {
	return len(value) > 0 && len(value) <= maximum && utf8.ValidString(value)
}

func validPublicOrigin(value string) bool {
	if !validProofString(value, maximumProofAudienceBytes) {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" || parsed.ForceQuery || parsed.RawPath != "" ||
		(parsed.Path != "" && parsed.Path != "/") || parsed.Host != strings.ToLower(parsed.Host) {
		return false
	}
	return parsed.String() == value
}
