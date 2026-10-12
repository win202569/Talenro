// Package nodeleafkeys checks a fixed SKI/AKI derivation relation against exact
// subject SPKI and a caller-declared issuer SPKI. It does not authenticate the
// issuer, signature, trust, current time, private-key possession, or runtime.
// All six nodeleaf obligations remain pending. CA SKI compatibility is unknown
// without separate CA material and authentication. No concurrent input writes
// or concurrent Clear/read are supported.
package nodeleafkeys
