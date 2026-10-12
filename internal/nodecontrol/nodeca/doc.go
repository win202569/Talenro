// Package nodeca checks only an owned, closed CA material subset and its
// declared subject key and fixed SPKI-SHA256 identifier. Success is not trust,
// signature verification, current validity, parent binding, or runtime authority.
// Summary is a forgeable detached value, not a capability. Input mutation during
// Parse and concurrent Clear/read are unsupported; cancellation is synchronous.
// Clearing covers reachable owned mutable buffers, not allocator/parser copies,
// immutable strings, fatal OOM, or system crashes.
package nodeca
