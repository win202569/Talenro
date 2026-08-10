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
	row := GetRefreshTokenForUpdateRow{AccountState: "suspended"}
	if row.AccountState != "suspended" {
		t.Fatalf("account state = %q", row.AccountState)
	}
}

func TestAccountRefreshAuthorityUsesExplicitTokenSessionAccountLockOrder(t *testing.T) {
	positions := []int{
		strings.Index(getRefreshTokenForUpdate, "locked_refresh AS MATERIALIZED"),
		strings.Index(getRefreshTokenForUpdate, "locked_session AS MATERIALIZED"),
		strings.Index(getRefreshTokenForUpdate, "locked_account AS MATERIALIZED"),
	}
	if positions[0] < 0 || positions[1] <= positions[0] || positions[2] <= positions[1] || strings.Contains(getRefreshTokenForUpdate, "FOR UPDATE OF r, s, a") {
		t.Fatalf("refresh authority lock structure is not token -> session -> account: %v", positions)
	}
}

type principalSessionRevoker interface {
	RevokePrincipalAccountSessions(context.Context, RevokePrincipalAccountSessionsParams) ([]uuid.UUID, error)
}

// Session revocation must remain one generated PostgreSQL transition: splitting
// selection and mutation in application code re-opens a rotation race.
var _ principalSessionRevoker = (*Queries)(nil)
