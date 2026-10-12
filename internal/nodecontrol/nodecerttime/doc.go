// Package nodecerttime binds direct certificate evidence to a conditional,
// caller-supplied time snapshot. It authenticates neither observation provenance
// nor deployment CA trust, selected identity, nonce consumption or runtime rights.
// Every new observation requires Evaluate with the complete input. A prior
// Summary and its ValidUntil do not permit caching, skipping fresh observation
// or state checks, extending a connection, or performing runtime actions.
// Shallow Model copies share one lifecycle: Clear on any copy invalidates all
// copies. Concurrent Model access, including Summary and Clear, is unsupported.
// Inputs, models and returned summaries must not be modified concurrently.
package nodecerttime
