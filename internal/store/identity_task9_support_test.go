package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentityTask9SupportMigrationAndQueriesAreExact(t *testing.T) {
	t.Parallel()

	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	migration := readTask9SupportFile(t, filepath.Join(repositoryRoot, "db", "migrations", "00005_identity_task9_support.sql"))
	queries := readTask9SupportFile(t, filepath.Join(repositoryRoot, "db", "queries", "identity.sql"))

	for _, statement := range []string{
		"CREATE UNIQUE INDEX identity_email_verification_token_hash_unique",
		"ON identity.email_identities (verification_token_hash)",
		"WHERE verification_token_hash IS NOT NULL;",
		"CREATE UNIQUE INDEX identity_password_reset_token_hash_unique",
		"ON identity.password_credentials (reset_token_hash)",
		"WHERE reset_token_hash IS NOT NULL;",
		"DROP INDEX identity.identity_password_reset_token_hash_unique;",
		"DROP INDEX identity.identity_email_verification_token_hash_unique;",
	} {
		if strings.Count(migration, statement) != 1 {
			t.Fatalf("migration statement %q count = %d, want 1", statement, strings.Count(migration, statement))
		}
	}
	if strings.Index(migration, "DROP INDEX identity.identity_password_reset_token_hash_unique;") >
		strings.Index(migration, "DROP INDEX identity.identity_email_verification_token_hash_unique;") {
		t.Fatal("support indexes are not dropped in reverse creation order")
	}
	for _, forbidden := range []string{"EXECUTE ", "format(", "DO $$"} {
		if strings.Contains(migration, forbidden) {
			t.Fatalf("migration contains dynamic SQL marker %q", forbidden)
		}
	}

	verificationQuery := task9QueryBlock(t, queries, "GetEmailVerificationForUpdate")
	for _, fragment := range []string{
		"WHERE verification_token_hash = $1",
		"FOR UPDATE;",
	} {
		if strings.Count(verificationQuery, fragment) != 1 {
			t.Fatalf("verification query fragment %q count = %d, want 1", fragment, strings.Count(verificationQuery, fragment))
		}
	}
	for _, forbidden := range []string{"verification_consumed_at IS NULL", "verification_expires_at >="} {
		if strings.Contains(verificationQuery, forbidden) {
			t.Fatalf("verification lock query still filters replay state with %q", forbidden)
		}
	}
	lookupQuery := task9QueryBlock(t, queries, "FindIdentityByLookupDigest")
	for _, fragment := range []string{"WHERE lookup_digest = $1", "FOR UPDATE;"} {
		if strings.Count(lookupQuery, fragment) != 1 {
			t.Fatalf("lookup query fragment %q count = %d, want 1", fragment, strings.Count(lookupQuery, fragment))
		}
	}
	lookupLockQuery := task9QueryBlock(t, queries, "LockEmailLookupDigest")
	for _, fragment := range []string{
		"SELECT pg_advisory_xact_lock(",
		"hashtextextended(encode(sqlc.arg(lookup_digest)::bytea, 'hex'), 0)",
	} {
		if strings.Count(lookupLockQuery, fragment) != 1 {
			t.Fatalf("lookup lock query fragment %q count = %d, want 1", fragment, strings.Count(lookupLockQuery, fragment))
		}
	}
	sessionQuery := task9QueryBlock(t, queries, "GetAccountSessionForUpdate")
	for _, fragment := range []string{
		"WHERE id = $1 AND principal_id = $2",
		"FOR UPDATE;",
	} {
		if strings.Count(sessionQuery, fragment) != 1 {
			t.Fatalf("session query fragment %q count = %d, want 1", fragment, strings.Count(sessionQuery, fragment))
		}
	}
}

func readTask9SupportFile(t *testing.T, path string) string {
	t.Helper()
	// #nosec G304 -- callers construct paths exclusively from fixed repository-relative components.
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Task 9 support file: %v", err)
	}
	return string(contents)
}

func task9QueryBlock(t *testing.T, queries, name string) string {
	t.Helper()
	marker := "-- name: " + name + " "
	start := strings.Index(queries, marker)
	if start < 0 {
		t.Fatalf("missing query %s", name)
	}
	remainder := queries[start:]
	if next := strings.Index(remainder[len(marker):], "\n-- name:"); next >= 0 {
		return remainder[:len(marker)+next]
	}
	return remainder
}
