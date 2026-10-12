// Package statepolicy evaluates conditional, isolated trust-state models.
// Observations are assumptions, not authenticated platform evidence. Neither a
// Candidate nor a successful recheck authorizes production or commits state.
// Callers must keep borrowed inputs immutable for the call; owners must not be
// copied or used concurrently with Clear. No provider is installed by this package.
package statepolicy
