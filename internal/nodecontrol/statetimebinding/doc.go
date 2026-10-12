// Package statetimebinding binds current committed metadata, selected policy,
// explicit time, a request nonce and checkpoint declarations in an owned model.
// Success is conditional: no physical provenance, clock/TLS validation, request
// consumption, durable checkpoint commit or runtime authorization is provided.
// Only unchanged current root and metadata are supported; candidates are never
// relabelled committed. Repeated evaluation is not replay protection.
package statetimebinding
