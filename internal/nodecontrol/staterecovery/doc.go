// Package staterecovery models dual-slot commit ordering and restart selection.
// All observations, including Available and Intact, are caller assumptions:
// this package authenticates no provider, seal, content, authorization, or
// business high-water invariant. It performs no I/O and grants no startup or
// LKG authority. Durable phases do not prove real fsync or hardware behavior.
// Snapshots are value-owned; callers must not concurrently mutate inputs or
// Clear models. Reusing old models is possible: no lock or exactly-once promise
// is provided. Clear does not guarantee erasure of runtime copies.
package staterecovery
