package statecontract

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrIncompleteCoverage = errors.New("incomplete coverage")
	ErrBindingMismatch    = errors.New("binding mismatch")
	ErrRollback           = errors.New("rollback")
	ErrConflict           = errors.New("conflict")
	ErrStaleModel         = errors.New("stale model")
)

// ContractError exposes only a bounded domain and a fixed error category.
type ContractError struct {
	Domain Domain
	Cause  error
}

func (e *ContractError) Error() string { return "state contract rejected" }
func (e *ContractError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
