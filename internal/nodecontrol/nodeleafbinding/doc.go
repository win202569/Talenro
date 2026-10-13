// Package nodeleafbinding binds a proposed main-state leaf projection to
// conditional certificate/time evidence. It authenticates neither platform
// provenance nor complete identity, issuer/serial mapping, key possession or
// runtime rights. The local DER/SPKI/window profile is not a production format.
// Every new observation requires the complete Evaluate input; deadlines do not
// authorize caching or runtime use. Model copies share a Clear lifecycle.
// Concurrent input/model access, including Summary and Clear, is unsupported.
package nodeleafbinding
