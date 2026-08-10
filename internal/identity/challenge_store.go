package identity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	accountRotationProtocolVersion = "account-rotation-v1"
	accountRotationOperation       = "rotate_account_token"
	accountChallengeTTL            = 2 * time.Minute
	challengeRecordPrefix          = "TALENRO-ACCOUNT-CHALLENGE-RECORD-V1\x00"
)

var (
	// ErrInvalidChallenge reports malformed challenge state without retaining it.
	ErrInvalidChallenge = errors.New("identity: invalid account challenge")
	// ErrChallengeNotFound reports consumed, expired, or context-mismatched state.
	ErrChallengeNotFound = errors.New("identity: account challenge unavailable")
	// ErrChallengeUnavailable reports ambiguous Redis or encoded-state failures.
	ErrChallengeUnavailable = errors.New("identity: challenge store unavailable")
)

// ChallengeRecord is the private, bounded Redis value for one account rotation.
// It deliberately contains no token, principal, session, or device identifier.
type ChallengeRecord struct {
	ChallengeID     string
	ProtocolVersion string
	Operation       string
	Challenge       [sha256.Size]byte
	ContextDigest   [sha256.Size]byte
	ExpiresAt       time.Time
}

// ChallengeStore creates and atomically consumes fixed account challenges.
type ChallengeStore interface {
	Create(context.Context, ChallengeRecord, time.Duration) error
	Consume(context.Context, string, [sha256.Size]byte) (ChallengeRecord, error)
}

// Format redacts every challenge field.
func (ChallengeRecord) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("identity.ChallengeRecord([REDACTED])"))
}

// LogValue prevents structured logging from reflecting challenge state.
func (ChallengeRecord) LogValue() slog.Value {
	return slog.StringValue("identity.ChallengeRecord([REDACTED])")
}

// MarshalJSON forbids direct challenge-state serialization.
func (ChallengeRecord) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity: challenge record serialization forbidden")
}

func encodeChallengeRecord(record ChallengeRecord) ([]byte, error) {
	if !validChallengeRecord(record) {
		return nil, ErrInvalidChallenge
	}
	protocol := []byte(record.ProtocolVersion)
	operation := []byte(record.Operation)
	result := make([]byte, 0, len(challengeRecordPrefix)+2+len(protocol)+2+len(operation)+sha256.Size*2+8)
	result = append(result, challengeRecordPrefix...)
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], uint16(len(protocol))) // #nosec G115 -- both fixed strings are far below Uint16.
	result = append(result, length[:]...)
	result = append(result, protocol...)
	binary.BigEndian.PutUint16(length[:], uint16(len(operation))) // #nosec G115 -- both fixed strings are far below Uint16.
	result = append(result, length[:]...)
	result = append(result, operation...)
	result = append(result, record.Challenge[:]...)
	result = append(result, record.ContextDigest[:]...)
	var expiry [8]byte
	binary.BigEndian.PutUint64(expiry[:], uint64(record.ExpiresAt.UTC().UnixNano())) // #nosec G115 -- same-width conversion preserves signed bits.
	result = append(result, expiry[:]...)
	clear(length[:])
	clear(expiry[:])
	return result, nil
}

func decodeChallengeRecord(challengeID string, wire []byte) (ChallengeRecord, error) {
	minimum := len(challengeRecordPrefix) + 2 + 2 + sha256.Size*2 + 8
	if len(wire) < minimum || !bytes.Equal(wire[:len(challengeRecordPrefix)], []byte(challengeRecordPrefix)) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	offset := len(challengeRecordPrefix)
	protocolLength := int(binary.BigEndian.Uint16(wire[offset : offset+2]))
	offset += 2
	if protocolLength != len(accountRotationProtocolVersion) || offset+protocolLength+2 > len(wire) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	protocol := string(wire[offset : offset+protocolLength])
	offset += protocolLength
	operationLength := int(binary.BigEndian.Uint16(wire[offset : offset+2]))
	offset += 2
	wantLength := offset + operationLength + sha256.Size*2 + 8
	if operationLength != len(accountRotationOperation) || wantLength != len(wire) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	operation := string(wire[offset : offset+operationLength])
	offset += operationLength
	record := ChallengeRecord{ChallengeID: challengeID, ProtocolVersion: protocol, Operation: operation}
	copy(record.Challenge[:], wire[offset:offset+sha256.Size])
	offset += sha256.Size
	copy(record.ContextDigest[:], wire[offset:offset+sha256.Size])
	offset += sha256.Size
	// #nosec G115 -- same-width conversion restores the signed two's-complement bits written above.
	expiresUnixNanos := int64(binary.BigEndian.Uint64(wire[offset : offset+8]))
	record.ExpiresAt = time.Unix(0, expiresUnixNanos).UTC()
	if !validChallengeRecord(record) {
		return ChallengeRecord{}, ErrChallengeUnavailable
	}
	return record, nil
}

func validChallengeRecord(record ChallengeRecord) bool {
	parsedID, err := uuid.Parse(record.ChallengeID)
	return err == nil && parsedID != uuid.Nil && parsedID.String() == record.ChallengeID &&
		record.ProtocolVersion == accountRotationProtocolVersion && record.Operation == accountRotationOperation &&
		record.Challenge != [sha256.Size]byte{} && record.ContextDigest != [sha256.Size]byte{} &&
		!record.ExpiresAt.IsZero() && record.ExpiresAt.Year() >= 2020 && record.ExpiresAt.Year() <= 2100
}

func challengeContextMatches(left, right [sha256.Size]byte) bool {
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}
