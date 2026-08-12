package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errTask13QueryProbe = errors.New("task13 query probe")

func TestDeviceTokenAuthorityQueriesExposeStableLockPrimitives(t *testing.T) {
	t.Parallel()

	database := &task13QueryProbeDB{}
	queries := New(database)
	ctx := context.Background()
	familyID := uuid.MustParse("0ff820a5-5022-48e6-8867-77761f8e2f07")
	authorizationID := uuid.MustParse("fa01e838-cab9-414e-8f03-ed0418cd4f20")
	deviceID := uuid.MustParse("f353613c-d08f-4141-b287-b37db9fb6f8e")
	digest := make([]byte, 32)

	probes := []struct {
		name string
		run  func() error
		want []string
		deny []string
	}{
		{name: "refresh discovery", run: func() error { _, err := queries.DiscoverDeviceRefreshToken(ctx, digest); return err }, want: []string{"device_refresh_tokens", "device_token_families", "device_authorizations"}, deny: []string{"FOR UPDATE", "identity.accounts"}},
		{name: "access discovery", run: func() error { _, err := queries.DiscoverDeviceAccessToken(ctx, digest); return err }, want: []string{"device_token_families", "device_authorizations"}, deny: []string{"FOR UPDATE", "identity.accounts"}},
		{name: "authorization discovery", run: func() error { _, err := queries.DiscoverDeviceAuthorization(ctx, deviceID); return err }, want: []string{"device_authorizations", "device_id"}, deny: []string{"FOR UPDATE", "identity.accounts"}},
		{name: "stable refresh lock", run: func() error { _, err := queries.LockDeviceFamilyRefreshTokens(ctx, familyID); return err }, want: []string{"ORDER BY token_hash", "FOR UPDATE"}},
		{name: "fresh refresh set", run: func() error { _, err := queries.ListDeviceFamilyRefreshTokens(ctx, familyID); return err }, want: []string{"ORDER BY token_hash"}, deny: []string{"FOR UPDATE"}},
		{name: "stable family discovery", run: func() error { _, err := queries.ListDeviceAuthorizationFamilies(ctx, authorizationID); return err }, want: []string{"ORDER BY id"}, deny: []string{"FOR UPDATE"}},
		{name: "family authority lock", run: func() error { _, err := queries.GetDeviceTokenFamilyForUpdate(ctx, familyID); return err }, want: []string{"FOR UPDATE"}},
		{name: "family revocation", run: func() error {
			_, err := queries.RevokeDeviceTokenFamily(ctx, RevokeDeviceTokenFamilyParams{ID: familyID, UpdatedAt: time.Now()})
			return err
		}, want: []string{"state='revoked'"}},
		{name: "device authority lock", run: func() error { _, err := queries.GetDeviceForUpdate(ctx, deviceID); return err }, want: []string{"FOR UPDATE"}},
		{name: "immutable policy", run: func() error { _, err := queries.GetDevicePolicySnapshot(ctx, authorizationID); return err }, want: []string{"device_policy_snapshots"}, deny: []string{"identity.accounts"}},
	}
	for _, probe := range probes {
		t.Run(probe.name, func(t *testing.T) {
			database.query = ""
			if err := probe.run(); !errors.Is(err, errTask13QueryProbe) {
				t.Fatalf("probe error = %v, want sentinel", err)
			}
			for _, fragment := range probe.want {
				if !strings.Contains(database.query, fragment) {
					t.Fatalf("query %q omitted %q", database.query, fragment)
				}
			}
			for _, fragment := range probe.deny {
				if strings.Contains(database.query, fragment) {
					t.Fatalf("query %q unexpectedly contains %q", database.query, fragment)
				}
			}
		})
	}
}

func TestProvisionalActivationFailsFastBehindTask13AuthorityLock(t *testing.T) {
	t.Parallel()

	database := &task13QueryProbeDB{}
	_, err := New(database).ActivateProvisionalAuthorization(context.Background(), ActivateProvisionalAuthorizationParams{
		PrincipalID: uuid.MustParse("b62139f5-b68c-4cc9-ab36-9b8d5224a875"), UpdatedAt: time.Date(2026, 8, 11, 1, 2, 3, 0, time.UTC),
	})
	if !errors.Is(err, errTask13QueryProbe) {
		t.Fatalf("activation probe error = %v, want sentinel", err)
	}
	if !strings.Contains(database.query, "FOR UPDATE NOWAIT") {
		t.Fatalf("activation query can wait behind reverse account/device lock order: %q", database.query)
	}
}

type task13QueryProbeDB struct{ query string }

func (database *task13QueryProbeDB) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	database.query = query
	return pgconn.CommandTag{}, errTask13QueryProbe
}

func (database *task13QueryProbeDB) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	database.query = query
	return nil, errTask13QueryProbe
}

func (database *task13QueryProbeDB) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	database.query = query
	return task13QueryProbeRow{}
}

type task13QueryProbeRow struct{}

func (task13QueryProbeRow) Scan(...any) error { return errTask13QueryProbe }

var _ DBTX = (*task13QueryProbeDB)(nil)
