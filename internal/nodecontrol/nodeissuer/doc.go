// Package nodeissuer checks an offline, direct CA–leaf issuer signature relation.
//
// A successful result is a partial fact about caller-provided exact materials.
// It is not issuer provenance, chain/root trust, trusted current validity,
// deployment authorization, private-key possession, main-state or runtime proof.
// Both lower obligation tables remain intact. Production/provider compatibility
// remains unknown. No certificate bytes or parser objects are exposed.
//
// Inputs must not be modified concurrently with Evaluate. Summary and Clear are
// not concurrency-safe. Clear erases reachable owned buffers, not inaccessible
// parser/runtime copies; the input budget is not a peak heap or real-time bound.
package nodeissuer
