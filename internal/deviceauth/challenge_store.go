package deviceauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	challengeRecordPrefix = "TALENRO-DEVICE-CHALLENGE-V1\x00"
	deviceChallengeTTL    = 2 * time.Minute
)

var (
	// ErrInvalidChallenge reports malformed device challenge input without retaining it.
	ErrInvalidChallenge = errors.New("deviceauth: invalid challenge")
	// ErrChallengeNotFound reports consumed, expired, or context-mismatched challenge state.
	ErrChallengeNotFound = errors.New("deviceauth: challenge unavailable")
	// ErrChallengeUnavailable reports ambiguous Redis or encoded-state failures.
	ErrChallengeUnavailable = errors.New("deviceauth: challenge store unavailable")
)

// ChallengeRecord is the private bounded Redis representation of one registration challenge.
type ChallengeRecord struct {
	ChallengeID     string
	Kind            ChallengeKind
	ProtocolVersion string
	Operation       string
	Challenge       [sha256.Size]byte
	GrantDigest     [sha256.Size]byte
	ContextDigest   [sha256.Size]byte
	ExpiresAt       time.Time
}

// ChallengeStore creates and atomically consumes one-time device challenges.
type ChallengeStore interface {
	Create(context.Context, ChallengeRecord, time.Duration) error
	Consume(context.Context, string, [sha256.Size]byte, [sha256.Size]byte) (ChallengeRecord, error)
}

// Format prevents diagnostic formatting from exposing challenge state.
func (ChallengeRecord) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("deviceauth.ChallengeRecord([REDACTED])"))
}

// LogValue prevents structured logs from reflecting challenge state.
func (ChallengeRecord) LogValue() slog.Value {
	return slog.StringValue("deviceauth.ChallengeRecord([REDACTED])")
}

// MarshalJSON forbids direct serialization of challenge state.
func (ChallengeRecord) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: challenge record serialization forbidden")
}

func encodeChallengeRecord(record ChallengeRecord) ([]byte, error) {
	if !validChallengeRecord(record) {
		return nil, ErrInvalidChallenge
	}
	kind := []byte(record.Kind)
	protocol := []byte(record.ProtocolVersion)
	operation := []byte(record.Operation)
	result := make([]byte, 0, len(challengeRecordPrefix)+6+len(kind)+len(protocol)+len(operation)+sha256.Size*3+8)
	result = append(result, challengeRecordPrefix...)
	result = appendChallengeRecordString(result, kind)
	result = appendChallengeRecordString(result, protocol)
	result = appendChallengeRecordString(result, operation)
	result = append(result, record.Challenge[:]...)
	result = append(result, record.GrantDigest[:]...)
	result = append(result, record.ContextDigest[:]...)
	var expiry [8]byte
	binary.BigEndian.PutUint64(expiry[:], uint64(record.ExpiresAt.UTC().UnixNano())) // #nosec G115 -- same-width conversion preserves bits.
	result = append(result, expiry[:]...)
	clear(expiry[:])
	return result, nil
}

func decodeChallengeRecord(challengeID string, wire []byte) (ChallengeRecord, error) {
	minimum := len(challengeRecordPrefix) + 6 + sha256.Size*3 + 8
	if len(wire) < minimum || !bytes.Equal(wire[:len(challengeRecordPrefix)], []byte(challengeRecordPrefix)) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	offset := len(challengeRecordPrefix)
	kind, next, ok := decodeChallengeRecordString(wire, offset, len(ChallengeRegistration))
	if !ok {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	offset = next
	protocol, next, ok := decodeChallengeRecordString(wire, offset, len(deviceProofProtocolVersion))
	if !ok {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	offset = next
	operation, next, ok := decodeChallengeRecordString(wire, offset, len(registerDeviceOperation))
	if !ok || next+sha256.Size*3+8 != len(wire) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	offset = next
	record := ChallengeRecord{ChallengeID: challengeID, Kind: ChallengeKind(kind), ProtocolVersion: protocol, Operation: operation}
	copy(record.Challenge[:], wire[offset:offset+sha256.Size])
	offset += sha256.Size
	copy(record.GrantDigest[:], wire[offset:offset+sha256.Size])
	offset += sha256.Size
	copy(record.ContextDigest[:], wire[offset:offset+sha256.Size])
	offset += sha256.Size
	record.ExpiresAt = time.Unix(0, int64(binary.BigEndian.Uint64(wire[offset:offset+8]))).UTC() // #nosec G115 -- same-width restoration.
	if !validChallengeRecord(record) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	return record, nil
}

func appendChallengeRecordString(target, value []byte) []byte {
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], uint16(len(value))) // #nosec G115 -- all values are fixed short contract strings.
	target = append(target, length[:]...)
	clear(length[:])
	return append(target, value...)
}

func decodeChallengeRecordString(wire []byte, offset, exactLength int) (string, int, bool) {
	if offset+2 > len(wire) {
		return "", 0, false
	}
	length := int(binary.BigEndian.Uint16(wire[offset : offset+2]))
	offset += 2
	if length != exactLength || offset+length > len(wire) {
		return "", 0, false
	}
	return string(wire[offset : offset+length]), offset + length, true
}

func validChallengeRecord(record ChallengeRecord) bool {
	parsedID, err := uuid.Parse(record.ChallengeID)
	return err == nil && parsedID != uuid.Nil && parsedID.String() == record.ChallengeID &&
		record.Kind == ChallengeRegistration && record.ProtocolVersion == deviceProofProtocolVersion && record.Operation == registerDeviceOperation &&
		record.Challenge != [sha256.Size]byte{} && record.GrantDigest != [sha256.Size]byte{} && record.ContextDigest != [sha256.Size]byte{} &&
		!record.ExpiresAt.IsZero() && record.ExpiresAt.Year() >= 2020 && record.ExpiresAt.Year() <= 2100
}

func challengeDigestsMatch(record ChallengeRecord, grantDigest, contextDigest [sha256.Size]byte) bool {
	grantMatch := subtle.ConstantTimeCompare(record.GrantDigest[:], grantDigest[:])
	contextMatch := subtle.ConstantTimeCompare(record.ContextDigest[:], contextDigest[:])
	return grantMatch&contextMatch == 1
}

var _ json.Marshaler = ChallengeRecord{}
