// Package statetime evaluates conditional offline time-attestation models only.
// Caller observations, IDs and hashes do not authenticate their provenance.
// In particular, this package does not verify metadata root signatures, establish
// committed state, validate TLS, consume nonces, read a clock or prevent replay.
// No production consumer or platform provider is installed. Inputs and Clear
// must not be used concurrently. Success is never a production capability.
package statetime
