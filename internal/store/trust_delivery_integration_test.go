//go:build integration

package store_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"talenro.local/platform/internal/store"
	"talenro.local/platform/internal/testinfra"
)

type idempotencyRaceResult struct {
	rowsAffected int64
	locked       store.IdempotencyRecord
	err          error
}

type idempotencyDisposition uint8

const (
	idempotencyInProgress idempotencyDisposition = iota
	idempotencyReplay
	idempotencyConflict
)

type idempotencyOutcome struct {
	disposition idempotencyDisposition
	status      int32
	response    []byte
	keyVersion  int32
}

type publicOutboxState struct {
	availableAt  time.Time
	claimedUntil sql.NullTime
	attempts     int32
	publishedAt  sql.NullTime
}

type testTxBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

func TestTrustBundleVersionsAreMonotonicUnderConcurrency(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	authorizationID := createActiveAuthorization(ctx, t, pool)
	queries := store.New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	versions := make(chan int64, 2)
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			ready <- struct{}{}
			<-start
			version, err := queries.NextBundleVersion(ctx, store.NextBundleVersionParams{
				AuthorizationID: authorizationID,
				UpdatedAt:       now,
			})
			versions <- version
			errs <- err
		}()
	}
	for range 2 {
		select {
		case <-ready:
		case <-ctx.Done():
			t.Fatal("version allocators did not reach the concurrency barrier")
		}
	}
	close(start)

	got := []int64{<-versions, <-versions}
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal("concurrent version allocation failed")
		}
	}
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if got[0] != 1 || got[1] != 2 {
		t.Fatalf("concurrent versions = %v, want [1 2]", got)
	}

	stored, err := queries.GetHighestBundleVersion(ctx, authorizationID)
	if err != nil {
		t.Fatal("read highest bundle version failed")
	}
	if stored != 2 {
		t.Fatalf("stored highest version = %d, want 2", stored)
	}
	third, err := queries.NextBundleVersion(ctx, store.NextBundleVersionParams{
		AuthorizationID: authorizationID,
		UpdatedAt:       now.Add(time.Microsecond),
	})
	if err != nil {
		t.Fatal("third version allocation failed")
	}
	if third != 3 {
		t.Fatalf("third version = %d, want 3", third)
	}
}

func TestTrustBundleStorageIsUniqueAndAcknowledgementIsIdempotent(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queries := store.New(pool)
	authorizationID := createActiveAuthorization(ctx, t, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	rootVersion := now.UnixNano()
	rootPayload := []byte("{}")
	rootSignature := bytes.Repeat([]byte{1}, 64)
	if err := queries.InsertTrustRootMetadata(ctx, store.InsertTrustRootMetadataParams{
		Version:          rootVersion,
		CanonicalPayload: rootPayload,
		Signature:        rootSignature,
		ValidFrom:        now.Add(-time.Hour),
		ValidUntil:       now.Add(48 * time.Hour),
		CreatedAt:        now,
	}); err != nil {
		t.Fatal("insert trust root metadata failed")
	}
	keyID := "task5_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := queries.InsertSigningKeyMetadata(ctx, store.InsertSigningKeyMetadataParams{
		KeyID:               keyID,
		RootMetadataVersion: rootVersion,
		PublicKey:           bytes.Repeat([]byte{2}, 32),
		State:               "active",
		NotBefore:           now.Add(-time.Minute),
		NotAfter:            now.Add(24 * time.Hour),
	}); err != nil {
		t.Fatal("insert signing key metadata failed")
	}

	bundleID := uuid.New()
	locatorSeed := uuid.New()
	locatorHash := bytes.Repeat(locatorSeed[:], 2)
	issuance := store.InsertBundleIssuanceParams{
		ID:                bundleID,
		AuthorizationID:   authorizationID,
		BundleVersion:     1,
		LocatorHash:       locatorHash,
		LocatorCiphertext: bytes.Repeat([]byte{4}, 29),
		LocatorKeyVersion: 1,
		Envelope:          bytes.Repeat([]byte{5}, 64),
		EnvelopeSha256:    bytes.Repeat([]byte{6}, 32),
		SignerKeyID:       keyID,
		IssuedAt:          now,
		NotBefore:         now,
		ExpiresAt:         now.Add(24 * time.Hour),
	}
	if err := queries.InsertBundleIssuance(ctx, issuance); err != nil {
		t.Fatal("insert bundle issuance failed")
	}

	byID, err := queries.GetBundleIssuance(ctx, bundleID)
	if err != nil {
		t.Fatal("read bundle issuance failed")
	}
	byLocator, err := queries.GetBundleByLocatorHash(ctx, store.GetBundleByLocatorHashParams{
		LocatorHash: locatorHash,
		ExpiresAt:   now,
	})
	if err != nil {
		t.Fatal("read bundle by locator hash failed")
	}
	if byID.ID != bundleID || byLocator.ID != bundleID || !bytes.Equal(byLocator.Envelope, issuance.Envelope) {
		t.Fatal("immutable bundle bytes were not returned unchanged")
	}

	latest, err := queries.GetLatestBundleIssuance(ctx, authorizationID)
	if err != nil {
		t.Fatal("read latest bundle issuance failed")
	}
	if latest.ID != bundleID || latest.BundleVersion != 1 || !bytes.Equal(latest.Envelope, issuance.Envelope) {
		t.Fatal("latest issuance did not return the immutable version-one row")
	}

	second := issuance
	second.ID = uuid.New()
	second.BundleVersion = 2
	second.LocatorHash = bytes.Repeat([]byte{7}, 32)
	second.LocatorCiphertext = bytes.Repeat([]byte{8}, 29)
	second.Envelope = bytes.Repeat([]byte{9}, 64)
	second.EnvelopeSha256 = bytes.Repeat([]byte{10}, 32)
	second.IssuedAt = now.Add(time.Second)
	second.NotBefore = now.Add(time.Second)
	second.ExpiresAt = now.Add(24*time.Hour + time.Second)
	if err := queries.InsertBundleIssuance(ctx, second); err != nil {
		t.Fatal("insert second bundle issuance failed")
	}
	latest, err = queries.GetLatestBundleIssuance(ctx, authorizationID)
	if err != nil {
		t.Fatal("read second latest bundle issuance failed")
	}
	if latest.ID != second.ID || latest.BundleVersion != 2 || !bytes.Equal(latest.Envelope, second.Envelope) {
		t.Fatal("latest issuance did not advance to the highest version")
	}

	duplicateLocator := issuance
	duplicateLocator.ID = uuid.New()
	duplicateLocator.BundleVersion = 3
	if err := queries.InsertBundleIssuance(ctx, duplicateLocator); err == nil {
		t.Fatal("duplicate locator hash was accepted")
	}

	ack := store.InsertBundleAcknowledgementParams{
		BundleID:        bundleID,
		AuthorizationID: authorizationID,
		BundleVersion:   1,
		AcknowledgedAt:  now.Add(time.Second),
	}
	if err := queries.InsertBundleAcknowledgement(ctx, ack); err != nil {
		t.Fatal("first bundle acknowledgement failed")
	}
	if err := queries.InsertBundleAcknowledgement(ctx, ack); err != nil {
		t.Fatal("replayed bundle acknowledgement failed")
	}
	var acknowledgementCount int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM trust.bundle_acknowledgements WHERE bundle_id=$1 AND authorization_id=$2`,
		bundleID,
		authorizationID,
	).Scan(&acknowledgementCount); err != nil {
		t.Fatal("count bundle acknowledgements failed")
	}
	if acknowledgementCount != 1 {
		t.Fatalf("acknowledgement count = %d, want 1", acknowledgementCount)
	}
}

func TestIdempotencyRaceLocksAndClassifiesCompletedRecord(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	raceID := uuid.New()
	keyHash := bytes.Repeat(raceID[:], 2)
	requestDigest := bytes.Repeat([]byte{8}, 32)
	params := store.TryBeginIdempotencyParams{
		PrincipalScope:     "principal:" + raceID.String(),
		Operation:          "issue_bundle",
		IdempotencyKeyHash: keyHash,
		RequestDigest:      requestDigest,
		CreatedAt:          now,
		ExpiresAt:          now.Add(24 * time.Hour),
	}

	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	results := make(chan idempotencyRaceResult, 2)
	for range 2 {
		go func() {
			ready <- struct{}{}
			<-start
			results <- beginIdempotency(ctx, pool, params)
		}()
	}
	for range 2 {
		select {
		case <-ready:
		case <-ctx.Done():
			t.Fatal("idempotency contenders did not reach the concurrency barrier")
		}
	}
	close(start)

	winners := 0
	losers := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal("idempotency contender failed")
		}
		switch result.rowsAffected {
		case 1:
			winners++
		case 0:
			losers++
			if !bytes.Equal(result.locked.RequestDigest, requestDigest) || result.locked.State != "in_progress" {
				t.Fatal("lost race did not lock the winning in-progress record")
			}
		default:
			t.Fatalf("idempotency affected rows = %d, want 0 or 1", result.rowsAffected)
		}
	}
	if winners != 1 || losers != 1 {
		t.Fatalf("idempotency outcomes: winners=%d losers=%d, want 1 each", winners, losers)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin idempotency completion transaction failed")
	}
	defer requireRollbackTestTransaction(t, tx) //nolint:contextcheck // rollback must survive operation cancellation
	txQueries := store.New(tx)
	locked, err := txQueries.GetIdempotencyForUpdate(ctx, store.GetIdempotencyForUpdateParams{
		PrincipalScope:     params.PrincipalScope,
		Operation:          params.Operation,
		IdempotencyKeyHash: keyHash,
	})
	if err != nil {
		t.Fatal("lock idempotency record for completion failed")
	}
	if locked.State != "in_progress" || !bytes.Equal(locked.RequestDigest, requestDigest) {
		t.Fatal("winner state changed before completion")
	}
	response := bytes.Repeat([]byte{9}, 48)
	completed, err := txQueries.CompleteIdempotency(ctx, store.CompleteIdempotencyParams{
		PrincipalScope:     params.PrincipalScope,
		Operation:          params.Operation,
		IdempotencyKeyHash: keyHash,
		RequestDigest:      requestDigest,
		ResponseStatus:     pgtype.Int4{Int32: 201, Valid: true},
		ResponseCiphertext: response,
		ResponseKeyVersion: pgtype.Int4{Int32: 4, Valid: true},
	})
	if err != nil {
		t.Fatal("complete idempotency record failed")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal("commit idempotency completion failed")
	}

	replayTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin idempotency replay transaction failed")
	}
	defer requireRollbackTestTransaction(t, replayTx) //nolint:contextcheck // rollback must survive operation cancellation
	stored, err := store.New(replayTx).GetIdempotencyForUpdate(ctx, store.GetIdempotencyForUpdateParams{
		PrincipalScope:     params.PrincipalScope,
		Operation:          params.Operation,
		IdempotencyKeyHash: keyHash,
	})
	if err != nil {
		t.Fatal("lock completed idempotency record failed")
	}
	if err := replayTx.Commit(ctx); err != nil {
		t.Fatal("commit idempotency replay transaction failed")
	}
	if !bytes.Equal(stored.ResponseCiphertext, completed.ResponseCiphertext) {
		t.Fatal("completed idempotency response was not stored")
	}

	replay := classifyIdempotency(stored, requestDigest)
	if replay.disposition != idempotencyReplay || replay.status != 201 || replay.keyVersion != 4 || !bytes.Equal(replay.response, response) {
		t.Fatal("same request digest did not replay the stored response")
	}
	conflict := classifyIdempotency(stored, bytes.Repeat([]byte{10}, 32))
	if conflict.disposition != idempotencyConflict {
		t.Fatal("different request digest was not classified as conflict")
	}
}

func TestIsolatedGlobalTableCleanupSurvivesCanceledOperationContext(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	operationCtx, cancelOperation := context.WithCancel(context.Background())
	defer cancelOperation()

	schemaName := ""
	t.Cleanup(func() {
		if schemaName == "" {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		identifier := pgx.Identifier{schemaName}.Sanitize()
		if _, err := pool.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+identifier+" CASCADE"); err != nil {
			t.Fatal("emergency isolated schema cleanup failed")
		}
	})

	t.Run("canceled operation", func(t *testing.T) {
		schemaName = createIsolatedGlobalTables(operationCtx, t, pool)
		cancelOperation()
	})

	verificationCtx, verificationCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer verificationCancel()
	assertSchemaAbsent(verificationCtx, t, pool, schemaName)
}

func TestPublicOutboxSentinelIgnoresAmbientSearchPath(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shadowSchema := createIsolatedGlobalTables(ctx, t, pool)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal("acquire ambient search-path connection failed")
	}
	defer func() {
		resetCtx, resetCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer resetCancel()
		if _, resetErr := conn.Exec(resetCtx, `RESET search_path`); resetErr != nil {
			t.Error("reset ambient search path failed")
		}
		conn.Release()
	}()
	searchPath := "SET search_path TO " +
		pgx.Identifier{shadowSchema}.Sanitize() + ", " +
		pgx.Identifier{"public"}.Sanitize()
	if _, err := conn.Exec(ctx, searchPath); err != nil {
		t.Fatal("set ambient shadow search path failed")
	}

	sentinelID := uuid.New()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM public.transactional_outbox WHERE event_id=$1`, sentinelID); err != nil {
			t.Fatal("delete ambient-search-path sentinel failed")
		}
	})
	insertPublicOutboxSentinel(ctx, t, conn, sentinelID, time.Now().UTC().Add(-time.Hour))

	shadowTable := pgx.Identifier{shadowSchema, "transactional_outbox"}.Sanitize()
	var shadowCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM "+shadowTable+" WHERE event_id=$1", sentinelID).Scan(&shadowCount); err != nil {
		t.Fatal("count shadow outbox sentinel failed")
	}
	var publicCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM public.transactional_outbox WHERE event_id=$1`, sentinelID).Scan(&publicCount); err != nil {
		t.Fatal("count public outbox sentinel failed")
	}
	if shadowCount != 0 || publicCount != 1 {
		t.Fatalf("sentinel placement: shadow=%d public=%d, want shadow=0 public=1", shadowCount, publicCount)
	}
}

func TestOutboxClaimPublishReleaseAndConsumerDedupe(t *testing.T) {
	pool := testinfra.OpenMigratedPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	publicBaseline := snapshotPublicOutbox(ctx, t, pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	sentinelID := uuid.New()
	insertPublicOutboxSentinel(ctx, t, pool, sentinelID, now.Add(-time.Hour))
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM public.transactional_outbox WHERE event_id=$1`, sentinelID); err != nil {
			t.Fatal("delete public outbox sentinel failed")
		}
	})

	schemaName := ""
	t.Run("isolated global tables", func(t *testing.T) {
		schemaName = createIsolatedGlobalTables(ctx, t, pool)
		exerciseOutboxPersistence(ctx, t, pool, schemaName, now)
	})

	assertSchemaAbsent(ctx, t, pool, schemaName)
	assertPublicOutboxSentinelUnchanged(ctx, t, pool, sentinelID)
	after := snapshotPublicOutbox(ctx, t, pool)
	delete(after, sentinelID)
	assertPublicOutboxBaselineUnchanged(t, publicBaseline, after)
}

func exerciseOutboxPersistence(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
	schemaName string,
	now time.Time,
) {
	t.Helper()
	setupTx := beginIsolatedGlobalTx(ctx, t, pool, schemaName)
	defer requireRollbackTestTransaction(t, setupTx) //nolint:contextcheck // rollback must survive operation cancellation
	setupQueries := store.New(setupTx)
	firstID := uuid.New()
	secondID := uuid.New()
	futureID := uuid.New()
	insertOutboxFixture(ctx, t, setupQueries, firstID, now.Add(-2*time.Second), now.Add(-time.Second))
	insertOutboxFixture(ctx, t, setupQueries, secondID, now.Add(-time.Second), now.Add(-time.Second))
	insertOutboxFixture(ctx, t, setupQueries, futureID, now, now.Add(time.Hour))
	if err := setupTx.Commit(ctx); err != nil {
		t.Fatal("commit isolated outbox fixtures failed")
	}

	firstTx := beginIsolatedGlobalTx(ctx, t, pool, schemaName)
	defer requireRollbackTestTransaction(t, firstTx) //nolint:contextcheck // rollback must survive operation cancellation
	firstClaimedUntil := now.Add(30 * time.Second)
	firstBatch, err := store.New(firstTx).ClaimOutboxBatch(ctx, store.ClaimOutboxBatchParams{
		AvailableAt:  now,
		Limit:        1,
		ClaimedUntil: sql.NullTime{Time: firstClaimedUntil, Valid: true},
	})
	if err != nil || len(firstBatch) != 1 {
		t.Fatal("first outbox batch claim failed")
	}

	secondTx := beginIsolatedGlobalTx(ctx, t, pool, schemaName)
	defer requireRollbackTestTransaction(t, secondTx) //nolint:contextcheck // rollback must survive operation cancellation
	secondClaimedUntil := now.Add(30 * time.Second)
	secondBatch, err := store.New(secondTx).ClaimOutboxBatch(ctx, store.ClaimOutboxBatchParams{
		AvailableAt:  now,
		Limit:        1,
		ClaimedUntil: sql.NullTime{Time: secondClaimedUntil, Valid: true},
	})
	if err != nil || len(secondBatch) != 1 {
		t.Fatal("second outbox batch claim failed")
	}
	if firstBatch[0].EventID == secondBatch[0].EventID {
		t.Fatal("SKIP LOCKED returned the same outbox event to two claimers")
	}
	if firstBatch[0].EventID != firstID || secondBatch[0].EventID != secondID {
		t.Fatal("outbox claims did not preserve occurred_at order")
	}
	if firstBatch[0].Attempts != 1 || secondBatch[0].Attempts != 1 {
		t.Fatal("first outbox claims did not increment attempts once")
	}
	if err := firstTx.Commit(ctx); err != nil {
		t.Fatal("commit first outbox claim failed")
	}
	if err := secondTx.Commit(ctx); err != nil {
		t.Fatal("commit second outbox claim failed")
	}

	workTx := beginIsolatedGlobalTx(ctx, t, pool, schemaName)
	defer requireRollbackTestTransaction(t, workTx) //nolint:contextcheck // rollback must survive operation cancellation
	queries := store.New(workTx)
	published, err := queries.MarkOutboxPublished(ctx, store.MarkOutboxPublishedParams{
		EventID:     firstID,
		PublishedAt: sql.NullTime{Time: now.Add(time.Second), Valid: true},
	})
	if err != nil || published != 1 {
		t.Fatal("mark outbox event published failed")
	}
	publishedAgain, err := queries.MarkOutboxPublished(ctx, store.MarkOutboxPublishedParams{
		EventID:     firstID,
		PublishedAt: sql.NullTime{Time: now.Add(2 * time.Second), Valid: true},
	})
	if err != nil || publishedAgain != 0 {
		t.Fatal("published outbox event was not idempotent")
	}

	retryAt := now.Add(2 * time.Second)
	released, err := queries.ReleaseOutboxClaim(ctx, store.ReleaseOutboxClaimParams{
		EventID:     secondID,
		AvailableAt: retryAt,
	})
	if err != nil || released != 1 {
		t.Fatal("release outbox claim failed")
	}
	retryBatch, err := queries.ClaimOutboxBatch(ctx, store.ClaimOutboxBatchParams{
		AvailableAt:  retryAt,
		Limit:        1,
		ClaimedUntil: sql.NullTime{Time: retryAt.Add(30 * time.Second), Valid: true},
	})
	if err != nil || len(retryBatch) != 1 || retryBatch[0].EventID != secondID || retryBatch[0].Attempts != 2 {
		t.Fatal("released outbox event was not reclaimed with incremented attempts")
	}

	health, err := queries.GetOutboxHealth(ctx, now)
	if err != nil {
		t.Fatal("read outbox health failed")
	}
	if health.Backlog != 2 {
		t.Fatalf("outbox backlog = %d, want 2", health.Backlog)
	}

	consumedExpiry := now.Add(30 * 24 * time.Hour)
	consumed, err := queries.RecordConsumedEvent(ctx, store.RecordConsumedEventParams{
		Consumer:   "task5.trust.publisher",
		EventID:    firstID,
		ConsumedAt: now,
		ExpiresAt:  consumedExpiry,
	})
	if err != nil || consumed != 1 {
		t.Fatal("record consumed event failed")
	}
	duplicate, err := queries.RecordConsumedEvent(ctx, store.RecordConsumedEventParams{
		Consumer:   "task5.trust.publisher",
		EventID:    firstID,
		ConsumedAt: now,
		ExpiresAt:  consumedExpiry,
	})
	if err != nil || duplicate != 0 {
		t.Fatal("duplicate consumed event was not ignored")
	}
	prunedEarly, err := queries.PruneConsumedEvents(ctx, consumedExpiry)
	if err != nil || prunedEarly != 0 {
		t.Fatal("consumer dedupe record was pruned at its inclusive expiry")
	}
	prunedExpired, err := queries.PruneConsumedEvents(ctx, consumedExpiry.Add(time.Microsecond))
	if err != nil || prunedExpired != 1 {
		t.Fatal("expired consumer dedupe record was not pruned")
	}
	if err := workTx.Commit(ctx); err != nil {
		t.Fatal("commit isolated outbox assertions failed")
	}
}

func beginIdempotency(
	ctx context.Context,
	pool *pgxpool.Pool,
	params store.TryBeginIdempotencyParams,
) idempotencyRaceResult {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return idempotencyRaceResult{err: err}
	}
	defer func() { _ = rollbackTestTransaction(tx) }() //nolint:contextcheck // rollback must survive operation cancellation
	queries := store.New(tx)
	rowsAffected, err := queries.TryBeginIdempotency(ctx, params)
	if err != nil {
		return idempotencyRaceResult{err: err}
	}
	result := idempotencyRaceResult{rowsAffected: rowsAffected}
	if rowsAffected == 0 {
		result.locked, result.err = queries.GetIdempotencyForUpdate(ctx, store.GetIdempotencyForUpdateParams{
			PrincipalScope:     params.PrincipalScope,
			Operation:          params.Operation,
			IdempotencyKeyHash: params.IdempotencyKeyHash,
		})
		if result.err != nil {
			return result
		}
	}
	result.err = tx.Commit(ctx)
	return result
}

func classifyIdempotency(record store.IdempotencyRecord, requestDigest []byte) idempotencyOutcome {
	if !bytes.Equal(record.RequestDigest, requestDigest) {
		return idempotencyOutcome{disposition: idempotencyConflict}
	}
	if record.State != "completed" || !record.ResponseStatus.Valid || !record.ResponseKeyVersion.Valid {
		return idempotencyOutcome{disposition: idempotencyInProgress}
	}
	return idempotencyOutcome{
		disposition: idempotencyReplay,
		status:      record.ResponseStatus.Int32,
		response:    record.ResponseCiphertext,
		keyVersion:  record.ResponseKeyVersion.Int32,
	}
}

func createActiveAuthorization(ctx context.Context, t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	queries := store.New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	principalID := uuid.New()
	deviceID := uuid.New()
	authorizationID := uuid.New()
	if err := queries.CreateAccount(ctx, store.CreateAccountParams{
		ID: principalID, State: "active", StateVersion: 1, Locale: "en", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal("create account fixture failed")
	}
	if err := queries.CreateDevice(ctx, store.CreateDeviceParams{
		ID:               deviceID,
		PrincipalID:      principalID,
		SigningPublicKey: bytes.Repeat(deviceID[:], 2),
		HpkePublicKey:    bytes.Repeat(authorizationID[:], 2),
		KeyVersion:       1,
		CreatedAt:        now,
	}); err != nil {
		t.Fatal("create device fixture failed")
	}
	if err := queries.CreateDeviceAuthorization(ctx, store.CreateDeviceAuthorizationParams{
		ID: authorizationID, PrincipalID: principalID, DeviceID: deviceID, State: "active", CreatedAt: now,
	}); err != nil {
		t.Fatal("create device authorization fixture failed")
	}
	return authorizationID
}

func insertOutboxFixture(
	ctx context.Context,
	t *testing.T,
	queries *store.Queries,
	eventID uuid.UUID,
	occurredAt time.Time,
	availableAt time.Time,
) {
	t.Helper()
	if err := queries.InsertOutboxEvent(ctx, store.InsertOutboxEventParams{
		EventID:          eventID,
		EventType:        "talenro.trust.bundle_issued.v1",
		AggregateType:    "bundle",
		AggregateID:      uuid.New(),
		AggregateVersion: 1,
		IdempotencyKey:   "task5:" + eventID.String(),
		Payload:          []byte{1},
		OccurredAt:       occurredAt,
		AvailableAt:      availableAt,
	}); err != nil {
		t.Fatal("insert outbox fixture failed")
	}
}

func insertPublicOutboxSentinel(
	ctx context.Context,
	t *testing.T,
	beginner testTxBeginner,
	eventID uuid.UUID,
	occurredAt time.Time,
) {
	t.Helper()
	tx, err := beginner.Begin(ctx)
	if err != nil {
		t.Fatal("begin public outbox sentinel transaction failed")
	}
	defer requireRollbackTestTransaction(t, tx) //nolint:contextcheck // rollback must survive operation cancellation
	if _, err := tx.Exec(ctx, `SET LOCAL search_path TO public`); err != nil {
		t.Fatal("bind public outbox sentinel transaction failed")
	}
	queries := store.New(tx)
	if err := queries.InsertOutboxEvent(ctx, store.InsertOutboxEventParams{
		EventID:          eventID,
		EventType:        "talenro.review.sentinel.v1",
		AggregateType:    "review_sentinel",
		AggregateID:      uuid.New(),
		AggregateVersion: 1,
		IdempotencyKey:   "review-sentinel:" + eventID.String(),
		Payload:          []byte{1},
		OccurredAt:       occurredAt,
		AvailableAt:      occurredAt,
	}); err != nil {
		t.Fatal("insert public outbox sentinel failed")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal("commit public outbox sentinel failed")
	}
}

func createIsolatedGlobalTables(ctx context.Context, t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	schemaName := "task5_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	schemaIdentifier := pgx.Identifier{schemaName}.Sanitize()
	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schemaIdentifier); err != nil {
		t.Fatal("create isolated global-table schema failed")
	}
	registerIsolatedGlobalTableCleanup(t, pool, schemaName) //nolint:contextcheck // cleanup must outlive operation context
	for _, tableName := range []string{"idempotency_records", "transactional_outbox", "consumed_event_ids"} {
		destination := pgx.Identifier{schemaName, tableName}.Sanitize()
		source := pgx.Identifier{"public", tableName}.Sanitize()
		if _, err := pool.Exec(ctx, "CREATE TABLE "+destination+" (LIKE "+source+" INCLUDING ALL)"); err != nil {
			t.Fatal("clone isolated global table failed")
		}
	}
	return schemaName
}

func registerIsolatedGlobalTableCleanup(t *testing.T, pool *pgxpool.Pool, schemaName string) {
	t.Helper()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		dropIsolatedGlobalTables(cleanupCtx, t, pool, schemaName)
	})
}

func dropIsolatedGlobalTables(ctx context.Context, t *testing.T, pool *pgxpool.Pool, schemaName string) {
	t.Helper()
	identifier := pgx.Identifier{schemaName}.Sanitize()
	if _, err := pool.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
		t.Fatal("drop isolated global-table schema failed")
	}
}

func beginIsolatedGlobalTx(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
	schemaName string,
) pgx.Tx {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("begin isolated global-table transaction failed")
	}
	searchPath := "SET LOCAL search_path TO " +
		pgx.Identifier{schemaName}.Sanitize() + ", " +
		pgx.Identifier{"public"}.Sanitize()
	if _, err := tx.Exec(ctx, searchPath); err != nil {
		_ = rollbackTestTransaction(tx) //nolint:contextcheck // rollback must survive operation cancellation
		t.Fatal("set isolated global-table search path failed")
	}
	return tx
}

func snapshotPublicOutbox(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
) map[uuid.UUID]publicOutboxState {
	t.Helper()
	rows, err := pool.Query(ctx, `
SELECT event_id, available_at, claimed_until, attempts, published_at
FROM public.transactional_outbox
ORDER BY event_id`)
	if err != nil {
		t.Fatal("snapshot public outbox failed")
	}
	defer rows.Close()

	result := make(map[uuid.UUID]publicOutboxState)
	for rows.Next() {
		var eventID uuid.UUID
		var state publicOutboxState
		if err := rows.Scan(
			&eventID,
			&state.availableAt,
			&state.claimedUntil,
			&state.attempts,
			&state.publishedAt,
		); err != nil {
			t.Fatal("scan public outbox snapshot failed")
		}
		result[eventID] = state
	}
	if rows.Err() != nil {
		t.Fatal("read public outbox snapshot failed")
	}
	return result
}

func assertPublicOutboxSentinelUnchanged(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
	eventID uuid.UUID,
) {
	t.Helper()
	var claimedUntil sql.NullTime
	var attempts int32
	var publishedAt sql.NullTime
	if err := pool.QueryRow(ctx, `
SELECT claimed_until, attempts, published_at
FROM public.transactional_outbox
WHERE event_id=$1`, eventID).Scan(&claimedUntil, &attempts, &publishedAt); err != nil {
		t.Fatal("read public outbox sentinel failed")
	}
	if claimedUntil.Valid || attempts != 0 || publishedAt.Valid {
		t.Fatal("isolated outbox test mutated the public sentinel")
	}
}

func assertSchemaAbsent(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
	schemaName string,
) {
	t.Helper()
	var absent bool
	if err := pool.QueryRow(ctx, `SELECT to_regnamespace($1) IS NULL`, schemaName).Scan(&absent); err != nil {
		t.Fatal("verify isolated schema cleanup failed")
	}
	if !absent {
		t.Fatal("isolated global-table schema remained after cleanup")
	}
}

func assertPublicOutboxBaselineUnchanged(
	t *testing.T,
	want map[uuid.UUID]publicOutboxState,
	got map[uuid.UUID]publicOutboxState,
) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("public outbox row count changed: got=%d want=%d", len(got), len(want))
	}
	for eventID, wantState := range want {
		gotState, ok := got[eventID]
		if !ok || !publicOutboxStatesEqual(gotState, wantState) {
			t.Fatal("pre-existing public outbox state changed")
		}
	}
}

func publicOutboxStatesEqual(a publicOutboxState, b publicOutboxState) bool {
	return a.availableAt.Equal(b.availableAt) &&
		nullTimesEqual(a.claimedUntil, b.claimedUntil) &&
		a.attempts == b.attempts &&
		nullTimesEqual(a.publishedAt, b.publishedAt)
}

func nullTimesEqual(a sql.NullTime, b sql.NullTime) bool {
	return a.Valid == b.Valid && (!a.Valid || a.Time.Equal(b.Time))
}

func requireRollbackTestTransaction(t *testing.T, tx pgx.Tx) {
	t.Helper()
	if err := rollbackTestTransaction(tx); err != nil {
		t.Error("rollback test transaction failed")
	}
}

func rollbackTestTransaction(tx pgx.Tx) error {
	rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer rollbackCancel()
	if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return err
	}
	return nil
}

var _ pgx.Tx
