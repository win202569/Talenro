package securitykit

// RandomSource supplies cryptographically secure bytes in production.
type RandomSource interface {
	Read([]byte) (int, error)
}
