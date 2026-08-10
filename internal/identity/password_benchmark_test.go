package identity

import (
	"testing"

	"golang.org/x/crypto/argon2"
)

// BenchmarkArgon2idV1 measures the fixed RFC 9106 second-recommended policy.
// Baseline recorded 2026-08-10 on Windows/amd64, Intel Core 7 240H, Go 1.26.5:
// 3 iterations, 41,906,733 ns/op, 67,121,018 B/op, 85 allocs/op. The policy
// remains 64 MiB, time=3, parallelism=4 regardless of workstation performance.
func BenchmarkArgon2idV1(b *testing.B) {
	policy := CurrentPasswordPolicy()
	password := []byte("benchmark-password")
	salt := []byte("0123456789abcdef")
	b.ReportAllocs()
	b.SetBytes(int64(policy.MemoryKiB) * 1024)
	for b.Loop() {
		result := argon2.IDKey(password, salt, policy.Time, policy.MemoryKiB, policy.Parallelism, policy.TagBytes)
		clear(result)
	}
}
