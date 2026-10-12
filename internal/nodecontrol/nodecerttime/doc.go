// Package nodecerttime binds direct certificate evidence to a conditional,
// caller-supplied time snapshot. It authenticates neither observation provenance
// nor deployment CA trust, selected identity, nonce consumption or runtime rights.
// Inputs, models and returned summaries must not be modified concurrently.
package nodecerttime
