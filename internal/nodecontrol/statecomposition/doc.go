// Package statecomposition checks a complete material subset and verifies one
// explicitly selected desired target against an independently supplied v1 anchor.
// Evidence is conditional on that anchor, not proof of its provenance or of
// current-time validity, unselected signatures, complete state, or authorization.
// Desired does not wire-sign metadata/root payload associations: a fresh coherent
// declaration may succeed without proving a historical association.
// Inputs must not change during a call. Owners must not be shallow-copied or
// accessed concurrently with Clear. Clear zeros reachable mutable owned storage;
// it makes no promise about immutable strings or inaccessible runtime copies.
package statecomposition
