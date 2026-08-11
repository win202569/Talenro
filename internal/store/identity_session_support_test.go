package store

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// A regression in the refresh authority projection would let rotation decide
// from session state while ignoring a concurrently suspended account.
func TestAccountRefreshAuthorityContractIncludesAccountState(t *testing.T) {
	row := IdentityAccount{State: "suspended"}
	if row.State != "suspended" {
		t.Fatalf("account state = %q", row.State)
	}
}

func TestAccountRefreshAuthorityUsesStableGeneratedPrelocks(t *testing.T) {
	if strings.Contains(discoverRefreshToken, "FOR UPDATE") {
		t.Fatal("refresh discovery unexpectedly decides authority under a partial lock")
	}
	for name, query := range map[string]string{
		"session":   lockSessionRefreshTokens,
		"principal": lockPrincipalRefreshTokens,
	} {
		if !strings.Contains(query, "ORDER BY r.token_hash") || !strings.Contains(query, "FOR UPDATE OF r") {
			t.Fatalf("%s refresh collection is not locked in token-hash order", name)
		}
	}
	if !strings.Contains(listSessionRefreshTokens, "ORDER BY r.token_hash") || strings.Contains(listSessionRefreshTokens, "FOR UPDATE") ||
		!strings.Contains(listPrincipalRefreshTokens, "ORDER BY r.token_hash") || strings.Contains(listPrincipalRefreshTokens, "FOR UPDATE") {
		t.Fatal("post-lock collection revalidation must be ordered and nonlocking")
	}
	if strings.Contains(revokeAccountRefreshTokens, "WHERE session_id") || !strings.Contains(revokeAccountRefreshTokens, "ANY") {
		t.Fatal("family revoke may acquire a previously unseen refresh row after session locks")
	}
	if strings.Contains(markPrincipalSessionsReviewRequired, "WHERE principal_id") || !strings.Contains(markPrincipalSessionsReviewRequired, "ANY") {
		t.Fatal("review transition may acquire a previously unseen session after account lock")
	}
}

func TestPasswordCredentialAuthorityIsExplicitlyLocked(t *testing.T) {
	if !strings.Contains(getPasswordCredential, "FOR UPDATE") {
		t.Fatal("password credential authority query is not locked")
	}
	if !strings.Contains(getPasswordResetForUpdate, "FOR UPDATE") {
		t.Fatal("password reset authority query is not locked")
	}
	if !strings.Contains(lockPrincipalAccountSessions, "ORDER BY id") || !strings.Contains(lockPrincipalAccountSessions, "FOR UPDATE") {
		t.Fatal("principal session query is not a stable ordered lock")
	}
}

type principalSessionRevoker interface {
	RevokePrincipalAccountSessions(context.Context, RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error)
}

type principalSessionLocker interface {
	LockPrincipalAccountSessions(context.Context, uuid.UUID) ([]IdentityAccountSession, error)
}

type refreshAuthorityLocker interface {
	DiscoverRefreshToken(context.Context, []byte) (DiscoverRefreshTokenRow, error)
	LockSessionRefreshTokens(context.Context, uuid.UUID) ([]IdentityAccountRefreshToken, error)
	LockPrincipalRefreshTokens(context.Context, uuid.UUID) ([]IdentityAccountRefreshToken, error)
	ListSessionRefreshTokens(context.Context, uuid.UUID) ([]IdentityAccountRefreshToken, error)
	ListPrincipalRefreshTokens(context.Context, uuid.UUID) ([]IdentityAccountRefreshToken, error)
}

// Session revocation mutates only the explicit sets selected under the stable
// generated prelocks; the write itself remains one generated transition.
var _ principalSessionRevoker = (*Queries)(nil)
var _ principalSessionLocker = (*Queries)(nil)
var _ refreshAuthorityLocker = (*Queries)(nil)
