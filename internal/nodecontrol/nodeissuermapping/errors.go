package nodeissuermapping

import "errors"

var ErrInvalidInput = errors.New("nodeissuermapping: invalid input")
var ErrLimitExceeded = errors.New("nodeissuermapping: limit exceeded")
var ErrInvalidEvidence = errors.New("nodeissuermapping: invalid evidence")
var ErrEvidenceInsufficient = errors.New("nodeissuermapping: evidence insufficient")
var ErrClaimConflict = errors.New("nodeissuermapping: claim conflict")
