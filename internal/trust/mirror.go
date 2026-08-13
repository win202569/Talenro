package trust

import (
	"net/http"

	"talenro.local/platform/internal/securitykit"
)

// NewMirror creates a dependency-isolated immutable mirror. Its only runtime
// collaborators are the byte store and clock; it receives no signer, HPKE,
// identity, device authorization, or sensitive-field protector.
func NewMirror(store ByteStore, clock securitykit.Clock) (http.Handler, error) {
	return NewDistribution(store, clock)
}
