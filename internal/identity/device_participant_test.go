package identity

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/securitykit"
	"talenro.local/platform/internal/store"
)

func TestAccountAuthenticatorRejectsDeviceDomainAccessToken(t *testing.T) {
	raw := secret.NewBytes(bytes.Repeat([]byte{0x7c}, 32))
	defer raw.Clear()
	transaction := activeTask10Transaction()
	deviceDigest := securitykit.DigestToken(securitykit.DeviceAccessToken, raw)
	transaction.accessFound = true
	transaction.accessDigest = bytes.Clone(deviceDigest[:])
	transaction.access = store.FindAccountAccessTokenRow{
		ID: uuid.New(), PrincipalID: transaction.account.ID, State: "active", StateVersion: 1,
		AccessExpiresAt: fixedTask10Time.Add(time.Minute), AbsoluteExpiresAt: fixedTask10Time.Add(time.Hour), AccountState: "active",
	}
	application, _, _ := newTask10Application(t, transaction, &task10ChallengeStore{})
	if _, err := application.Authenticate(context.Background(), raw); publicTask10Code(err) != apierrors.AuthenticationFailed {
		t.Fatalf("device-domain digest on account route = %v, want authentication failure", err)
	}
}

func TestPostgresRepositoryValidatesDeviceEnrollmentInsideCallerTransaction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	principalID := uuid.MustParse("a6493384-9407-4ad9-b220-7f3b49ef9054")
	sessionID := uuid.MustParse("46434e83-f7e2-43c1-b7c5-80a3b49eec14")
	digest := [32]byte{1, 2, 3, 4}
	database := &deviceParticipantDBTX{
		grant: store.GetGrantForChallengeRow{
			ID: uuid.MustParse("29250900-c6a2-4262-b393-28eb72197248"), PrincipalID: principalID,
			AccountSessionID: sessionID, TokenHash: bytes.Clone(digest[:]), PolicyMarker: "standard", ExpiresAt: now.Add(time.Minute),
		},
		sessions: []store.IdentityAccountSession{
			{ID: uuid.MustParse("0b0e8065-9ea4-43c1-a88e-7decc5b7bad1"), PrincipalID: principalID, State: "revoked"},
			{ID: sessionID, PrincipalID: principalID, State: "active", AccessExpiresAt: now.Add(time.Minute), AbsoluteExpiresAt: now.Add(time.Hour)},
		},
		account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 3},
	}

	var participant DeviceTransactionParticipant = &PostgresRepository{}
	authority, found, err := participant.ValidateDeviceEnrollment(context.Background(), database, digest, now)
	if err != nil || !found {
		t.Fatalf("validate enrollment = (found %v, error %v)", found, err)
	}
	if authority.PrincipalID() != PrincipalID(principalID.String()) || authority.SessionID() != SessionID(sessionID.String()) ||
		authority.PolicyMarker() != "standard" {
		t.Fatal("authority did not retain the exact redacted enrollment facts")
	}
	if provisional, valid := authority.ProvisionalUntil(); valid || !provisional.IsZero() {
		t.Fatalf("standard provisional boundary = (%v, %v), want absent", provisional, valid)
	}
	if got := strings.Join(database.operations, ","); got != "grant,sessions,account" {
		t.Fatalf("operation order = %q, want grant,sessions,account", got)
	}
	if !strings.Contains(database.sessionQuery, "ORDER BY id") || !strings.Contains(database.sessionQuery, "FOR UPDATE") {
		t.Fatalf("session lock does not preserve stable ordering: %q", database.sessionQuery)
	}
	if database.grantArgs[1] != now || !bytes.Equal(database.grantArgs[0].([]byte), digest[:]) {
		t.Fatal("grant lookup did not receive the exact caller digest and transaction clock")
	}
}

func TestPostgresRepositoryRejectsEnrollmentGrantAtOrBeforeExactExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	principalID := uuid.MustParse("c140a9b1-56f0-4625-9bb0-f7efb5781447")
	sessionID := uuid.MustParse("0e1a37ac-c25e-407f-afbf-7e91180f76e3")
	digest := [32]byte{5, 6, 7, 8}
	for _, expiresAt := range []time.Time{now, now.Add(-time.Nanosecond)} {
		database := &deviceParticipantDBTX{
			grant: store.GetGrantForChallengeRow{
				ID: uuid.MustParse("ee10e4e8-986e-4e25-ab90-8c68e0ca08aa"), PrincipalID: principalID,
				AccountSessionID: sessionID, TokenHash: bytes.Clone(digest[:]), PolicyMarker: "standard", ExpiresAt: expiresAt,
			},
			sessions: []store.IdentityAccountSession{{
				ID: sessionID, PrincipalID: principalID, State: "active", AccessExpiresAt: now.Add(time.Minute), AbsoluteExpiresAt: now.Add(time.Hour),
			}},
			account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 1},
		}

		_, found, err := (&PostgresRepository{}).ValidateDeviceEnrollment(context.Background(), database, digest, now)

		if err != nil || found {
			t.Fatalf("expiry %v validation = (found %v, error %v), want finite not-found", expiresAt, found, err)
		}
		if got := strings.Join(database.operations, ","); got != "grant" {
			t.Fatalf("expiry %v operations = %q, want only grant", expiresAt, got)
		}
	}
}

func TestPostgresRepositoryValidatesTrialEnrollmentAndRejectsAuthorityMismatch(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	principalID := uuid.MustParse("79624f50-ce16-492f-9e7a-cdcf42335798")
	sessionID := uuid.MustParse("293494a7-960b-44c7-b8ed-ed17a7fbce84")
	digest := [32]byte{9, 8, 7, 6}
	provisionalUntil := now.Add(24 * time.Hour)
	validDatabase := func() *deviceParticipantDBTX {
		return &deviceParticipantDBTX{
			grant: store.GetGrantForChallengeRow{
				ID: uuid.MustParse("2167fb84-c843-4208-aab0-3d42a7e6ef09"), PrincipalID: principalID,
				AccountSessionID: sessionID, TokenHash: bytes.Clone(digest[:]), PolicyMarker: "trial_restricted",
				ProvisionalUntil: sql.NullTime{Time: provisionalUntil, Valid: true}, ExpiresAt: now.Add(time.Minute),
			},
			sessions: []store.IdentityAccountSession{{
				ID: sessionID, PrincipalID: principalID, State: "active", AccessExpiresAt: now.Add(time.Minute), AbsoluteExpiresAt: now.Add(time.Hour),
			}},
			account: store.IdentityAccount{ID: principalID, State: "pending_email", StateVersion: 1},
		}
	}

	participant := DeviceTransactionParticipant(&PostgresRepository{})
	authority, found, err := participant.ValidateDeviceEnrollment(context.Background(), validDatabase(), digest, now)
	if err != nil || !found {
		t.Fatalf("validate trial enrollment = (found %v, error %v)", found, err)
	}
	if boundary, valid := authority.ProvisionalUntil(); !valid || !boundary.Equal(provisionalUntil) {
		t.Fatalf("trial boundary = (%v, %v), want %v", boundary, valid, provisionalUntil)
	}

	tests := []struct {
		name   string
		mutate func(*deviceParticipantDBTX)
	}{
		{name: "unknown marker", mutate: func(db *deviceParticipantDBTX) { db.grant.PolicyMarker = "other" }},
		{name: "trial active account", mutate: func(db *deviceParticipantDBTX) { db.account.State = "active" }},
		{name: "standard pending account", mutate: func(db *deviceParticipantDBTX) {
			db.grant.PolicyMarker = "standard"
			db.grant.ProvisionalUntil = sql.NullTime{}
		}},
		{name: "expired session access", mutate: func(db *deviceParticipantDBTX) { db.sessions[0].AccessExpiresAt = now }},
		{name: "bound session", mutate: func(db *deviceParticipantDBTX) {
			db.sessions[0].DeviceAuthorizationID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
		}},
		{name: "digest mismatch", mutate: func(db *deviceParticipantDBTX) { db.grant.TokenHash[0] ^= 0xff }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			database := validDatabase()
			test.mutate(database)
			_, found, err := participant.ValidateDeviceEnrollment(context.Background(), database, digest, now)
			if err != nil || found {
				t.Fatalf("mismatched authority = (found %v, error %v), want finite not-found", found, err)
			}
		})
	}
}

func TestPostgresRepositoryValidatesDeviceEnrollmentSanitizesFailures(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	digest := [32]byte{1, 2, 3, 4}
	participant := DeviceTransactionParticipant(&PostgresRepository{})

	notFoundDatabase := &deviceParticipantDBTX{grantErr: pgx.ErrNoRows}
	if _, found, err := participant.ValidateDeviceEnrollment(context.Background(), notFoundDatabase, digest, now); err != nil || found {
		t.Fatalf("missing grant = (found %v, error %v), want finite not-found", found, err)
	}
	if got := strings.Join(notFoundDatabase.operations, ","); got != "grant" {
		t.Fatalf("missing grant operations = %q, want only grant", got)
	}

	providerFailure := errors.New("SECRET_PROVIDER_CANARY")
	if _, _, err := participant.ValidateDeviceEnrollment(context.Background(), &deviceParticipantDBTX{grantErr: providerFailure}, digest, now); !errors.Is(err, ErrRepository) || strings.Contains(err.Error(), "CANARY") {
		t.Fatalf("provider failure = %v, want fixed repository error", err)
	}

	var typedNil *deviceParticipantDBTX
	if _, _, err := participant.ValidateDeviceEnrollment(context.Background(), typedNil, digest, now); !errors.Is(err, ErrRepository) {
		t.Fatalf("typed-nil DBTX error = %v, want fixed repository error", err)
	}
	if _, _, err := participant.ValidateDeviceEnrollment(context.Background(), panicDeviceParticipantDBTX{}, digest, now); !errors.Is(err, ErrRepository) || strings.Contains(err.Error(), "CANARY") {
		t.Fatalf("panic failure = %v, want fixed repository error", err)
	}
}

func TestDeviceEnrollmentAuthorityRedactsDiagnosticsAndRejectsJSON(t *testing.T) {
	t.Parallel()

	principal := PrincipalID("a6493384-9407-4ad9-b220-7f3b49ef9054")
	session := SessionID("46434e83-f7e2-43c1-b7c5-80a3b49eec14")
	authority, err := NewDeviceEnrollmentAuthority(principal, session, "standard", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	marker := "identity.DeviceEnrollmentAuthority([REDACTED])"
	forbidden := []string{string(principal), string(session), "standard"}
	var nilAuthority *DeviceEnrollmentAuthority
	values := []struct {
		name        string
		value       any
		requireMark bool
	}{
		{name: "value", value: authority, requireMark: true},
		{name: "pointer", value: &authority, requireMark: true},
		{name: "zero", value: DeviceEnrollmentAuthority{}, requireMark: true},
		{name: "nil", value: nilAuthority},
	}
	for _, value := range values {
		for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X"} {
			rendered := fmt.Sprintf(format, value.value)
			assertDeviceEnrollmentAuthorityRedacted(t, value.name+" "+format, rendered, forbidden)
			if value.requireMark && !strings.Contains(rendered, marker) {
				t.Fatalf("%s %s = %q, want fixed marker %q", value.name, format, rendered, marker)
			}
		}
		for _, handler := range []struct {
			name string
			new  func(*bytes.Buffer) slog.Handler
		}{
			{name: "text", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewTextHandler(buffer, nil) }},
			{name: "json", new: func(buffer *bytes.Buffer) slog.Handler { return slog.NewJSONHandler(buffer, nil) }},
		} {
			var output bytes.Buffer
			slog.New(handler.new(&output)).Info("authority state", "authority", value.value)
			rendered := output.String()
			assertDeviceEnrollmentAuthorityRedacted(t, value.name+" slog "+handler.name, rendered, forbidden)
			if strings.Contains(rendered, "!PANIC") || strings.Contains(rendered, "!ERROR") {
				t.Fatalf("%s slog %s used an error fallback: %q", value.name, handler.name, rendered)
			}
			if value.requireMark && !strings.Contains(rendered, marker) {
				t.Fatalf("%s slog %s = %q, want fixed marker %q", value.name, handler.name, rendered, marker)
			}
		}
	}
	for _, value := range []struct {
		name       string
		value      any
		mustReject bool
	}{
		{name: "value", value: authority, mustReject: true},
		{name: "pointer", value: &authority, mustReject: true},
		{name: "zero", value: DeviceEnrollmentAuthority{}, mustReject: true},
		{name: "nil", value: nilAuthority},
	} {
		encoded, err := json.Marshal(value.value)
		assertDeviceEnrollmentAuthorityRedacted(t, value.name+" json", string(encoded), forbidden)
		if value.mustReject && err == nil {
			t.Fatalf("json.Marshal(%s) succeeded: %q", value.name, encoded)
		}
		if !value.mustReject && (err != nil || string(encoded) != "null") {
			t.Fatalf("json.Marshal(nil) = %q, %v, want safe null", encoded, err)
		}
	}
}

func assertDeviceEnrollmentAuthorityRedacted(t *testing.T, context, rendered string, forbidden []string) {
	t.Helper()
	for _, canary := range forbidden {
		if strings.Contains(rendered, canary) {
			t.Fatalf("%s exposed %q in %q", context, canary, rendered)
		}
	}
}

func TestPostgresRepositoryBindsAndRevokesDeviceSessionsThroughCallerDBTX(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 10, 1, 2, 3, 0, time.UTC)
	sessionID := SessionID("46434e83-f7e2-43c1-b7c5-80a3b49eec14")
	authorizationID := uuid.MustParse("b350891f-68af-4e0f-a8b7-b07938354c18")
	database := &deviceParticipantDBTX{bindRows: 1, revokeRows: 2}
	participant := DeviceTransactionParticipant(&PostgresRepository{})

	if err := participant.BindSessionToAuthorization(context.Background(), database, sessionID, authorizationID, now); err != nil {
		t.Fatalf("bind session: %v", err)
	}
	if err := participant.RevokeAuthorizationSessions(context.Background(), database, authorizationID, now.Add(time.Minute)); err != nil {
		t.Fatalf("revoke sessions: %v", err)
	}
	if got := strings.Join(database.operations, ","); got != "bind,event,revoke" {
		t.Fatalf("operation order = %q, want bind,event,revoke", got)
	}
	if database.eventID != authorizationID || database.eventCategory != "device_registered" || database.eventFingerprint != "deviceauth.registration" {
		t.Fatal("binding did not write the fixed collision-resistant registration security event")
	}

	for _, rows := range []int64{0, 2} {
		database := &deviceParticipantDBTX{bindRows: rows}
		if err := participant.BindSessionToAuthorization(context.Background(), database, sessionID, authorizationID, now); !errors.Is(err, ErrRepository) {
			t.Fatalf("bind rows %d error = %v, want fixed repository error", rows, err)
		}
		if database.eventID != uuid.Nil {
			t.Fatalf("bind rows %d wrote a security event", rows)
		}
	}
}

func TestPostgresRepositoryValidatesDeviceAccountAndRevocationAuthorityLast(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 11, 1, 2, 3, 0, time.UTC)
	principalID := uuid.MustParse("bc2d2fd3-a51a-4d3f-9f36-d5df75bb79dc")
	sessionID := uuid.MustParse("61ca9232-71e0-466e-9833-506f52b457e4")
	password := []byte("TASK13-current-password")
	passwordHash := bytes.Repeat([]byte{0x51}, 32)
	database := &deviceParticipantDBTX{
		credential: store.IdentityPasswordCredential{
			PrincipalID: principalID, PolicyVersion: 1, MemoryKib: 65536, TimeCost: 3, Parallelism: 4,
			Salt: bytes.Repeat([]byte{0x31}, 16), PasswordHash: bytes.Clone(passwordHash), UpdatedAt: now.Add(-time.Hour),
		},
		sessions: []store.IdentityAccountSession{{
			ID: sessionID, PrincipalID: principalID, State: "active", AccessExpiresAt: now.Add(time.Minute), AbsoluteExpiresAt: now.Add(time.Hour),
		}},
		account: store.IdentityAccount{ID: principalID, State: "active", StateVersion: 9},
	}
	repository := &PostgresRepository{derive: func(candidate, _ []byte, _ PasswordPolicy) []byte {
		if bytes.Equal(candidate, password) {
			return bytes.Clone(passwordHash)
		}
		return bytes.Repeat([]byte{0xa7}, len(passwordHash))
	}}
	participant := DeviceTransactionParticipant(repository)

	active, err := participant.ValidateDeviceAccountAuthority(context.Background(), database, PrincipalID(principalID.String()), now)
	if err != nil || !active {
		t.Fatalf("validate device account authority = (%v, %v), want active", active, err)
	}
	if got := strings.Join(database.operations, ","); got != "account" {
		t.Fatalf("account authority order = %q, want account", got)
	}
	database.account.State = "pending_email"
	if allowed, err := participant.ValidateDeviceAccountAuthority(context.Background(), database, PrincipalID(principalID.String()), now); err != nil || !allowed {
		t.Fatalf("provisional device account authority = (%v, %v), want grace authority", allowed, err)
	}
	database.account.State = "suspended"
	if allowed, err := participant.ValidateDeviceAccountAuthority(context.Background(), database, PrincipalID(principalID.String()), now); err != nil || allowed {
		t.Fatalf("suspended device account authority = (%v, %v), want finite rejection", allowed, err)
	}
	database.account.State = "active"
	database.operations = nil

	request := DeviceRevocationRequest{
		PrincipalID: PrincipalID(principalID.String()),
		Reauthentication: Reauthentication{
			SessionID: SessionID(sessionID.String()), Method: ReauthPassword, Proof: secret.NewBytes(password),
		},
	}
	defer request.Reauthentication.Proof.Clear()
	authority, found, err := participant.ValidateDeviceRevocation(context.Background(), database, request, now)
	if err != nil || !found {
		t.Fatalf("validate revocation = (found %v, error %v)", found, err)
	}
	if authority.PrincipalID() != request.PrincipalID || authority.SessionID() != request.Reauthentication.SessionID {
		t.Fatal("revocation authority did not preserve the exact verified identity authority")
	}
	if got := strings.Join(database.operations, ","); got != "credential,sessions,account" {
		t.Fatalf("revocation identity lock order = %q, want credential,sessions,account", got)
	}

	badProof := request
	badProof.Reauthentication.Proof = secret.NewBytes([]byte("TASK13-wrong-password"))
	defer badProof.Reauthentication.Proof.Clear()
	database.operations = nil
	if _, found, err := participant.ValidateDeviceRevocation(context.Background(), database, badProof, now); err != nil || found {
		t.Fatalf("wrong reauthentication = (found %v, error %v), want finite rejection", found, err)
	}
}

func TestPostgresRepositoryRecordsOnlyFixedDeviceReplaySecurityCategory(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 11, 2, 3, 4, 0, time.UTC)
	eventID := uuid.MustParse("73fb40c1-7c56-4f4a-8518-2c1529d134ac")
	authorizationID := uuid.MustParse("7e2298b0-cb0c-4af1-a70a-704e3756f95f")
	record, err := NewDeviceTokenReplaySecurityRecord(eventID, authorizationID)
	if err != nil {
		t.Fatal(err)
	}
	database := &deviceParticipantDBTX{}
	participant := DeviceTransactionParticipant(&PostgresRepository{})
	if err := participant.RecordDeviceTokenReplay(context.Background(), database, record, now); err != nil {
		t.Fatal(err)
	}
	if database.eventID != eventID || database.eventCategory != "device_token_replay" || database.eventFingerprint != "deviceauth.refresh_replay" {
		t.Fatal("device replay participant did not write the fixed security classification")
	}
}

func TestClearPasswordCredentialErasesOwnedSaltAndHash(t *testing.T) {
	t.Parallel()
	credential, err := NewPasswordCredential(CurrentPasswordPolicy(), bytes.Repeat([]byte{0x91}, 16), bytes.Repeat([]byte{0x92}, 32))
	if err != nil {
		t.Fatal(err)
	}
	saltOwner, hashOwner := credential.salt, credential.hash
	clearPasswordCredential(&credential)
	if credential.policy != (PasswordPolicy{}) || credential.salt != nil || credential.hash != nil || credential.dummy {
		t.Fatal("cleared password credential retained policy or ownership")
	}
	for _, owner := range [][]byte{saltOwner, hashOwner} {
		for _, value := range owner {
			if value != 0 {
				t.Fatal("password credential clear left owned bytes")
			}
		}
	}
}

func TestDeviceParticipantTask13AuthoritiesRedactDiagnosticsAndRejectJSON(t *testing.T) {
	t.Parallel()

	principal := PrincipalID("bc2d2fd3-a51a-4d3f-9f36-d5df75bb79dc")
	session := SessionID("61ca9232-71e0-466e-9833-506f52b457e4")
	request := DeviceRevocationRequest{
		PrincipalID:      principal,
		Reauthentication: Reauthentication{SessionID: session, Method: ReauthPassword, Proof: secret.NewBytes([]byte("TASK13-secret-proof"))},
	}
	defer request.Reauthentication.Proof.Clear()
	authority, err := NewDeviceRevocationAuthority(principal, session)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewDeviceTokenReplaySecurityRecord(
		uuid.MustParse("73fb40c1-7c56-4f4a-8518-2c1529d134ac"),
		uuid.MustParse("7e2298b0-cb0c-4af1-a70a-704e3756f95f"),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{request, authority, record} {
		rendered := fmt.Sprintf("%+v", value)
		for _, forbidden := range []string{string(principal), string(session), "TASK13-secret-proof", "73fb40c1", "7e2298b0"} {
			if strings.Contains(rendered, forbidden) {
				t.Fatalf("diagnostic formatting exposed %q in %q", forbidden, rendered)
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("generic JSON serialization succeeded for %T", value)
		}
	}
}

type deviceParticipantDBTX struct {
	grant            store.GetGrantForChallengeRow
	grantErr         error
	sessions         []store.IdentityAccountSession
	sessionsErr      error
	account          store.IdentityAccount
	accountErr       error
	credential       store.IdentityPasswordCredential
	credentialErr    error
	bindRows         int64
	revokeRows       int64
	operations       []string
	grantArgs        []any
	sessionQuery     string
	eventID          uuid.UUID
	eventCategory    string
	eventFingerprint string
}

func (database *deviceParticipantDBTX) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	switch {
	case strings.Contains(query, "SET device_authorization_id"):
		database.operations = append(database.operations, "bind")
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", database.bindRows)), nil
	case strings.Contains(query, "INSERT INTO identity.security_events"):
		database.operations = append(database.operations, "event")
		database.eventID, _ = arguments[0].(uuid.UUID)
		database.eventCategory, _ = arguments[2].(string)
		database.eventFingerprint, _ = arguments[3].(string)
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	case strings.Contains(query, "SET state='revoked'") && strings.Contains(query, "identity.account_sessions"):
		database.operations = append(database.operations, "revoke")
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE %d", database.revokeRows)), nil
	default:
		return pgconn.CommandTag{}, errors.New("unexpected Exec call")
	}
}

func (database *deviceParticipantDBTX) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	database.operations = append(database.operations, "sessions")
	database.sessionQuery = query
	if database.sessionsErr != nil {
		return nil, database.sessionsErr
	}
	return &deviceParticipantRows{sessions: database.sessions}, nil
}

func (database *deviceParticipantDBTX) QueryRow(_ context.Context, query string, arguments ...any) pgx.Row {
	switch {
	case strings.Contains(query, "deviceauth.enrollment_grants"):
		database.operations = append(database.operations, "grant")
		database.grantArgs = append([]any(nil), arguments...)
		database.grantArgs[0] = bytes.Clone(arguments[0].([]byte))
		return deviceParticipantRow{values: []any{
			database.grant.ID, database.grant.PrincipalID, database.grant.AccountSessionID, database.grant.TokenHash,
			database.grant.PolicyMarker, database.grant.ProvisionalUntil, database.grant.ExpiresAt,
		}, err: database.grantErr}
	case strings.Contains(query, "identity.accounts"):
		database.operations = append(database.operations, "account")
		return deviceParticipantRow{values: []any{
			database.account.ID, database.account.State, database.account.StateVersion, database.account.Locale,
			database.account.CreatedAt, database.account.UpdatedAt,
		}, err: database.accountErr}
	case strings.Contains(query, "identity.password_credentials"):
		database.operations = append(database.operations, "credential")
		return deviceParticipantRow{values: []any{
			database.credential.PrincipalID, database.credential.PolicyVersion, database.credential.MemoryKib,
			database.credential.TimeCost, database.credential.Parallelism, database.credential.Salt,
			database.credential.PasswordHash, database.credential.ResetTokenHash, database.credential.ResetExpiresAt,
			database.credential.ResetConsumedAt, database.credential.ResetDeliveryID,
			database.credential.ResetDeliveryCiphertext, database.credential.ResetDeliveryKeyVersion,
			database.credential.UpdatedAt,
		}, err: database.credentialErr}
	default:
		return deviceParticipantRow{err: errors.New("unexpected QueryRow call")}
	}
}

type deviceParticipantRows struct {
	sessions []store.IdentityAccountSession
	index    int
	closed   bool
}

func (rows *deviceParticipantRows) Close()                                  { rows.closed = true }
func (*deviceParticipantRows) Err() error                                   { return nil }
func (*deviceParticipantRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (*deviceParticipantRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (rows *deviceParticipantRows) Next() bool {
	if rows.index >= len(rows.sessions) {
		rows.closed = true
		return false
	}
	rows.index++
	return true
}
func (rows *deviceParticipantRows) Scan(destinations ...any) error {
	if rows.index == 0 || rows.index > len(rows.sessions) {
		return errors.New("Scan without current row")
	}
	session := rows.sessions[rows.index-1]
	return assignDeviceParticipantValues(destinations, []any{
		session.ID, session.PrincipalID, session.State, session.StateVersion, session.ClientSigningPublicKey,
		session.AccessTokenHash, session.AccessExpiresAt, session.AbsoluteExpiresAt, session.CreatedAt,
		session.UpdatedAt, session.DeviceAuthorizationID,
	})
}
func (*deviceParticipantRows) Values() ([]any, error) { return nil, errors.New("unused") }
func (*deviceParticipantRows) RawValues() [][]byte    { return nil }
func (*deviceParticipantRows) Conn() *pgx.Conn        { return nil }

type deviceParticipantRow struct {
	values []any
	err    error
}

func (row deviceParticipantRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	return assignDeviceParticipantValues(destinations, row.values)
}

func assignDeviceParticipantValues(destinations, values []any) error {
	if len(destinations) != len(values) {
		return errors.New("unexpected destination count")
	}
	for index := range destinations {
		switch destination := destinations[index].(type) {
		case *uuid.UUID:
			*destination = values[index].(uuid.UUID)
		case *[]byte:
			*destination = bytes.Clone(values[index].([]byte))
		case *string:
			*destination = values[index].(string)
		case *int64:
			*destination = values[index].(int64)
		case *int32:
			*destination = values[index].(int32)
		case *time.Time:
			*destination = values[index].(time.Time)
		case *sql.NullTime:
			*destination = values[index].(sql.NullTime)
		case *uuid.NullUUID:
			*destination = values[index].(uuid.NullUUID)
		case *pgtype.Int4:
			*destination = values[index].(pgtype.Int4)
		default:
			return fmt.Errorf("unexpected destination %T", destinations[index])
		}
	}
	return nil
}

var _ store.DBTX = (*deviceParticipantDBTX)(nil)
var _ pgx.Rows = (*deviceParticipantRows)(nil)

type panicDeviceParticipantDBTX struct{}

func (panicDeviceParticipantDBTX) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("SECRET_EXEC_CANARY")
}
func (panicDeviceParticipantDBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("SECRET_QUERY_CANARY")
}
func (panicDeviceParticipantDBTX) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("SECRET_QUERY_ROW_CANARY")
}

var _ store.DBTX = panicDeviceParticipantDBTX{}
