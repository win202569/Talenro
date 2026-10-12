// Package statematerials checks a bounded subset of raw format materials and
// caller-declared relationships. It supplies no trust, runtime or commit authority.
// Owners must not be shallow-copied or accessed concurrently with Clear. Callers
// must not mutate input during a call. Other domains and full composition are outstanding.
package statematerials
