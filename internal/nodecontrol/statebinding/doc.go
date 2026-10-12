// Package statebinding binds complete previous/proposed L0 declarations to two
// separately owned material tables and one conditional signature composition.
// It proves neither trusted persistence, original historical authorization,
// platform provenance, current-time validity, full closure, commit nor runtime
// permission. Unselected and previous signatures retain their lower-layer grade.
// Inputs must not be mutated concurrently. Owners must not be shallow-copied or
// accessed concurrently with Clear. Clearing reaches mutable owned buffers, not
// caller copies, immutable strings or unreachable runtime copies.
package statebinding
