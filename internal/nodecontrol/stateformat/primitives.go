package stateformat

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"time"
)

type formatCheck struct{ ok bool }

func (c *formatCheck) require(ok bool)   { c.ok = c.ok && ok }
func (c *formatCheck) u(s string) uint64 { v, e := parseU(s); c.require(e == nil); return v }
func (c *formatCheck) timestamp(s string) time.Time {
	v, e := parseTime(s)
	c.require(e == nil)
	return v
}
func (c *formatCheck) bounded(s string, min, max uint64) {
	v := c.u(s)
	c.require(v >= min && v <= max)
}
func oneOf(s string, options ...string) bool {
	for _, v := range options {
		if v == s {
			return true
		}
	}
	return false
}

// Both operands have already passed the finite wire time bounds. Subtraction
// avoids manufacturing an out-of-range upper date at year 9999.
func validWindow(start, end time.Time, max time.Duration) bool {
	return !start.IsZero() && !end.IsZero() && end.After(start) && end.Sub(start) <= max
}

func parseU(s string) (uint64, error) {
	if len(s) < 1 || len(s) > 19 || s[0] < '1' || s[0] > '9' {
		return 0, ErrInvalidFormat
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, ErrInvalidFormat
		}
	}
	v, err := strconv.ParseUint(s, 10, 63)
	if err != nil {
		return 0, ErrInvalidFormat
	}
	return v, nil
}
func parseDigest(s string) (contracts.Digest, error) {
	var v contracts.Digest
	if len(s) != 64 {
		return v, ErrInvalidFormat
	}
	for i := range s {
		if !(s[i] >= '0' && s[i] <= '9' || s[i] >= 'a' && s[i] <= 'f') {
			return v, ErrInvalidFormat
		}
	}
	_, err := hex.Decode(v[:], []byte(s))
	if err != nil || v == (contracts.Digest{}) {
		return contracts.Digest{}, ErrInvalidFormat
	}
	return v, nil
}
func decodeFixed(s string, out []byte) bool {
	if len(s) != base64.RawURLEncoding.EncodedLen(len(out)) {
		return false
	}
	for i := range s {
		if !(asciiAlnum(s[i]) || s[i] == '_' || s[i] == '-') {
			return false
		}
	}
	n, err := base64.RawURLEncoding.Strict().Decode(out, []byte(s))
	if err != nil || n != len(out) || base64.RawURLEncoding.EncodeToString(out) != s {
		return false
	}
	var nonzero byte
	for _, v := range out {
		nonzero |= v
	}
	return nonzero != 0
}
func parseKey(s string) ([32]byte, error) {
	var v [32]byte
	if !decodeFixed(s, v[:]) {
		return [32]byte{}, ErrInvalidFormat
	}
	return v, nil
}
func parseSignature(s string) ([64]byte, error) {
	var v [64]byte
	if !decodeFixed(s, v[:]) {
		return [64]byte{}, ErrInvalidFormat
	}
	return v, nil
}
func parseTime(s string) (time.Time, error) {
	const layout = "2006-01-02T15:04:05Z"
	if len(s) != 20 {
		return time.Time{}, ErrInvalidFormat
	}
	v, err := time.Parse(layout, s)
	if err != nil || v.IsZero() || v.Year() < 1 || v.Year() > 9999 || v.Format(layout) != s {
		return time.Time{}, ErrInvalidFormat
	}
	return v, nil
}
func parseUUID(s string) (uuid.UUID, error) {
	if len(s) != 36 {
		return uuid.Nil, ErrInvalidFormat
	}
	v, err := uuid.Parse(s)
	if err != nil || v == uuid.Nil || v.Variant() != uuid.RFC4122 || v.String() != s {
		return uuid.Nil, ErrInvalidFormat
	}
	return v, nil
}
func asciiAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
func validID(s string, max int) bool {
	if len(s) == 0 || len(s) > max || !asciiAlnum(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !(asciiAlnum(s[i]) || s[i] == '.' || s[i] == '_' || s[i] == '-') {
			return false
		}
	}
	return true
}
func validBuild(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 || len(p) > 10 || len(p) > 1 && p[0] == '0' {
			return false
		}
		for i := range p {
			if p[i] < '0' || p[i] > '9' {
				return false
			}
		}
		if _, err := strconv.ParseUint(p, 10, 31); err != nil {
			return false
		}
	}
	return true
}
