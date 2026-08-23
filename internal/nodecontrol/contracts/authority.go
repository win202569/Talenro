package contracts

import (
	"errors"
	"math"
)

type Digest [32]byte

type AuthorityVersion struct {
	Epoch    uint64
	Sequence uint64
}

type VersionedDigest struct {
	Version           uint64
	AuthoritySequence uint64
	Digest            Digest
}

type Comparison string

const (
	ComparisonSame     Comparison = "same"
	ComparisonAdvance  Comparison = "advance"
	ComparisonRollback Comparison = "rollback"
	ComparisonFork     Comparison = "fork"
)

var ErrInvalidAuthorityValue = errors.New("nodecontrol contracts: invalid authority value")

func (value AuthorityVersion) Validate() error {
	if value.Epoch == 0 || value.Sequence == 0 ||
		value.Epoch > math.MaxInt64 || value.Sequence > math.MaxInt64 {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func CompareVersionedDigest(current, candidate VersionedDigest) (Comparison, error) {
	if current.Version == 0 || current.Version > math.MaxInt64 ||
		current.AuthoritySequence == 0 || current.AuthoritySequence > math.MaxInt64 ||
		current.Digest == (Digest{}) || candidate.Version == 0 || candidate.Version > math.MaxInt64 ||
		candidate.AuthoritySequence == 0 || candidate.AuthoritySequence > math.MaxInt64 ||
		candidate.Digest == (Digest{}) {
		return "", ErrInvalidAuthorityValue
	}
	if candidate.Version < current.Version || candidate.AuthoritySequence < current.AuthoritySequence {
		return ComparisonRollback, nil
	}
	if candidate.Version == current.Version && candidate.AuthoritySequence == current.AuthoritySequence {
		if candidate.Digest == current.Digest {
			return ComparisonSame, nil
		}
		return ComparisonFork, nil
	}
	if candidate.Version > current.Version && candidate.AuthoritySequence > current.AuthoritySequence {
		return ComparisonAdvance, nil
	}
	return ComparisonFork, nil
}
