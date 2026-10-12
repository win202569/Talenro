// Package stateverify verifies bounded signature relationships relative to an
// independently supplied anchor. Success does NOT establish deployment trust,
// current validity, rollback protection, historical signing authority or serving
// authorization. Expired chains may yield evidence here, but must never be used
// to bypass trusted-time network acceptance or deployment recovery rules.
// Desired does not sign a unique metadata identity: compatible metadata can
// yield different, equally conditional evidence. No production consumer exists.
package stateverify
