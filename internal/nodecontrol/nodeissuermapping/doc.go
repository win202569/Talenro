// Package nodeissuermapping checks caller-declared registry/version associations
// against real nodeissuerbinding material evidence. Success proves declaration
// consistency only, not authenticated publication, issuance provenance, database
// uniqueness, protected v7 reads, or runtime authorization. Coherent fabricated,
// old, or content-changed declarations can pass. No registry wire format or
// registry content hash is defined here. Lower obligations remain unresolved.
//
// Claims are independently owned; complete Material budget/copy belongs to the
// real lower. Clear covers reachable mutable owned storage, not hidden runtime
// copies or immutable strings. Caller mutation and model concurrency are unsupported.
package nodeissuermapping
