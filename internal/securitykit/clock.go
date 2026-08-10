// Package securitykit provides injectable time, randomness, and secret-token primitives.
package securitykit

import "time"

// Clock supplies the current time to security-sensitive workflows.
type Clock interface {
	Now() time.Time
}
