//go:build integration

package testinfra

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type C12AuthorityPITRCommitKind string

const (
	C12AuthorityPITRCommitImmediate C12AuthorityPITRCommitKind = "commit"
	C12AuthorityPITRCommitPrepared  C12AuthorityPITRCommitKind = "commit_prepared"
)

type C12AuthorityPITRBaseBackup struct {
	BackupID string
	StartLSN string
	EndLSN   string
}

type C12AuthorityPITRCommit struct {
	Kind   C12AuthorityPITRCommitKind
	SQLXID uint64
	EndLSN string
	GID    string
}

type C12AuthorityPITRCut struct {
	BackupID          string
	TerminalCommit    C12AuthorityPITRCommit
	RecoveryTargetLSN string
}

type C12AuthorityPITRCandidate struct {
	Index     uint8
	Name      string
	DataName  string
	TargetLSN string
	Promoted  bool
}

type C12AuthorityPITRTimeline struct {
	CandidateIndex uint8
	TimelineID     uint64
	ReplayLSN      string
	Promoted       bool
}

type C12AuthorityPITRController struct {
	state *c12AuthorityPITRState
}

type c12AuthorityPITRState struct {
	database       *sql.DB
	databaseURL    *url.URL
	descriptor     c12AuthorityPITRDescriptor
	nonce          [32]byte
	wal            *c12AuthorityPITROWAL
	phase          atomic.Uint32
	nextCandidate  atomic.Uint32
	candidatePhase [8]atomic.Uint32
	mu             sync.Mutex
	baseBackup     C12AuthorityPITRBaseBackup
	terminalCommit C12AuthorityPITRCommit
	cut            C12AuthorityPITRCut
	candidates     [8]c12AuthorityPITRCandidateState
	commitSequence uint32
}

type c12AuthorityPITRCandidateState struct {
	containerID  string
	port         int
	volumeExists bool
	containerUp  bool
	targetLSN    string
}

type c12AuthorityPITRDescriptor struct {
	Schema                 string   `json:"schema"`
	Profile                string   `json:"profile"`
	RunSuffix              string   `json:"run_suffix"`
	NonceDigest            string   `json:"nonce_digest"`
	DockerExecutable       string   `json:"docker_executable"`
	DockerExecutableDigest string   `json:"docker_executable_digest"`
	ImageRef               string   `json:"image_ref"`
	ImageID                string   `json:"image_id"`
	DatabaseIdentity       string   `json:"database_identity"`
	ApplicationName        string   `json:"application_name"`
	PrimaryName            string   `json:"primary_name"`
	PrimaryID              string   `json:"primary_id"`
	PrimaryDataName        string   `json:"primary_data_name"`
	ArchiveName            string   `json:"archive_name"`
	BaseBackupName         string   `json:"basebackup_name"`
	TLSPublicDigest        string   `json:"tls_public_digest"`
	ObserverRole           string   `json:"observer_role"`
	SlotName               string   `json:"slot_name"`
	Plugin                 string   `json:"plugin"`
	CandidateNames         []string `json:"candidate_names"`
	CandidateDataNames     []string `json:"candidate_data_names"`
	MaxCandidates          int      `json:"max_candidates"`
	MaxWALLineBytes        int      `json:"max_wal_line_bytes"`
	MaxWALRecords          int      `json:"max_wal_records"`
	FailureSeam            string   `json:"failure_seam"`
	DescriptorDigest       string   `json:"descriptor_digest"`
	DescriptorHMAC         string   `json:"descriptor_hmac"`
}

type c12AuthorityPITROWAL struct {
	path       string
	key        [32]byte
	maxLine    int
	maxRecords int
	mu         sync.Mutex
	sequence   uint64
	previous   string
}

type c12AuthorityPITRWALRecord struct {
	Schema         string `json:"schema"`
	Sequence       uint64 `json:"sequence"`
	PreviousDigest string `json:"previous_digest"`
	Event          string `json:"event"`
	Resource       string `json:"resource"`
	Name           string `json:"name"`
	Identity       string `json:"identity"`
	Labels         string `json:"labels"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	State          string `json:"state"`
	LSN            string `json:"lsn"`
	Timestamp      string `json:"timestamp"`
	PayloadSHA256  string `json:"payload_sha256"`
	RecordSHA256   string `json:"record_sha256"`
	HMACSHA256     string `json:"hmac_sha256"`
}

type c12AuthorityPITRDecodedRow struct {
	LSN  string
	XID  uint64
	Data string
}

var (
	c12AuthorityPITRRunPattern       = regexp.MustCompile(`^[0-9a-f]{32}$`)
	c12AuthorityPITRLSNPattern       = regexp.MustCompile(`^[0-9A-F]+/[0-9A-F]+$`)
	c12AuthorityPITRDigestPattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
	c12AuthorityPITRContainerPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	c12AuthorityPITRRolePattern      = regexp.MustCompile(`^talenro_c12_[0-9a-f]{24}_(observer|slot)$`)
)

var c12AuthorityPITRFailureSeams = map[string]bool{
	"":                                 true,
	"after-intent-before-create":       true,
	"after-create-before-actual":       true,
	"after-actual-before-return":       true,
	"after-clean-intent-before-remove": true,
	"after-remove-before-clean-result": true,
}

const c12AuthorityPITRImageID = "sha256:996d0920e4ff9df1fc19dacb904492f3c1ec0ec1cc338f0ad7123be7731c5f5e"

func OpenC12AuthorityPITR() (C12AuthorityPITRController, error) {
	rawNonce := os.Getenv("TALENRO_C12_AUTHORITY_PITR_NONCE")
	nonceBytes, err := hex.DecodeString(rawNonce)
	if err != nil || len(nonceBytes) != 32 {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR nonce")
	}
	descriptorBytes, err := base64.RawURLEncoding.DecodeString(os.Getenv("TALENRO_C12_AUTHORITY_PITR_DESCRIPTOR"))
	if err != nil || len(descriptorBytes) == 0 || len(descriptorBytes) > 8192 {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR descriptor")
	}
	var descriptor c12AuthorityPITRDescriptor
	if err := json.Unmarshal(descriptorBytes, &descriptor); err != nil {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR descriptor")
	}
	keyBytes, err := hex.DecodeString(os.Getenv("TALENRO_C12_AUTHORITY_PITR_HMAC_KEY"))
	if err != nil || len(keyBytes) != 32 {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR key")
	}
	var key, nonce [32]byte
	copy(key[:], keyBytes)
	copy(nonce[:], nonceBytes)
	clear(keyBytes)
	clear(nonceBytes)
	if err := validateC12AuthorityPITRDescriptor(descriptor, nonce, key); err != nil {
		return C12AuthorityPITRController{}, err
	}
	walPath := os.Getenv("TALENRO_C12_AUTHORITY_PITR_WAL_PATH")
	wantRoot := "talenro-c12-pitr-" + descriptor.RunSuffix
	if !filepath.IsAbs(walPath) || filepath.Base(walPath) != "ownership.wal" || filepath.Base(filepath.Dir(walPath)) != wantRoot {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR ownership WAL path")
	}
	wal, err := openC12AuthorityPITROWAL(walPath, key, descriptor.MaxWALLineBytes, descriptor.MaxWALRecords)
	if err != nil {
		return C12AuthorityPITRController{}, err
	}
	if wal.sequence == 0 {
		if err := wal.append("BOOTSTRAP", "run", descriptor.RunSuffix, descriptor.PrimaryID, c12AuthorityPITRLabels(descriptor, "bootstrap"), 0, walPath, "ready", ""); err != nil {
			return C12AuthorityPITRController{}, err
		}
	}
	rawDatabaseURL := os.Getenv("TALENRO_DATABASE_URL")
	parsedURL, err := url.Parse(rawDatabaseURL)
	if err != nil || parsedURL.Hostname() != "127.0.0.1" || strings.TrimPrefix(parsedURL.Path, "/") != "talenro_c12_"+descriptor.RunSuffix {
		return C12AuthorityPITRController{}, errors.New("invalid authority PITR database endpoint")
	}
	database, err := sql.Open("pgx", rawDatabaseURL)
	if err != nil {
		return C12AuthorityPITRController{}, errors.New("open authority PITR database")
	}
	state := &c12AuthorityPITRState{database: database, databaseURL: parsedURL, descriptor: descriptor, nonce: nonce, wal: wal}
	if err := state.verifyPrimaryAndRuntime(context.Background()); err != nil {
		database.Close()
		return C12AuthorityPITRController{}, err
	}
	state.phase.Store(1)
	if err := wal.append("TRANSITION", "controller", descriptor.PrimaryName, descriptor.PrimaryID, c12AuthorityPITRLabels(descriptor, "controller"), 0, "", "ready", ""); err != nil {
		database.Close()
		return C12AuthorityPITRController{}, err
	}
	return C12AuthorityPITRController{state: state}, nil
}

func validateC12AuthorityPITRDescriptor(descriptor c12AuthorityPITRDescriptor, nonce, key [32]byte) error {
	wantNonceDigest := sha256.Sum256(nonce[:])
	if descriptor.Schema != "talenro-c12-authority-pitr/v1" || descriptor.Profile != "authority-v7-pitr" ||
		!c12AuthorityPITRRunPattern.MatchString(descriptor.RunSuffix) || descriptor.NonceDigest != hex.EncodeToString(wantNonceDigest[:]) ||
		descriptor.ImageRef != "postgres:18.4-alpine3.23" || descriptor.ImageID != c12AuthorityPITRImageID ||
		descriptor.Plugin != "test_decoding" || descriptor.ApplicationName != "talenro-c12-"+descriptor.RunSuffix ||
		descriptor.PrimaryName != "talenro-c12-"+descriptor.RunSuffix+"-pitr-primary" || !c12AuthorityPITRContainerPattern.MatchString(descriptor.PrimaryID) ||
		descriptor.PrimaryDataName != descriptor.PrimaryName+"-data" || descriptor.ArchiveName != "talenro-c12-"+descriptor.RunSuffix+"-pitr-archive" ||
		descriptor.BaseBackupName != "talenro-c12-"+descriptor.RunSuffix+"-pitr-basebackup" ||
		!c12AuthorityPITRDigestPattern.MatchString(descriptor.DatabaseIdentity) || !c12AuthorityPITRDigestPattern.MatchString(descriptor.TLSPublicDigest) ||
		!c12AuthorityPITRRolePattern.MatchString(descriptor.ObserverRole) || !c12AuthorityPITRRolePattern.MatchString(descriptor.SlotName) ||
		descriptor.MaxCandidates != 8 || descriptor.MaxWALLineBytes != 8192 || descriptor.MaxWALRecords != 4096 ||
		len(descriptor.CandidateNames) != 8 || len(descriptor.CandidateDataNames) != 8 || !c12AuthorityPITRFailureSeams[descriptor.FailureSeam] {
		return errors.New("authority PITR descriptor binding mismatch")
	}
	wantDatabaseIdentity := sha256.Sum256([]byte("talenro-c12-authority-v7:" + descriptor.RunSuffix + ":database-identity"))
	if descriptor.DatabaseIdentity != hex.EncodeToString(wantDatabaseIdentity[:]) {
		return errors.New("authority PITR database identity mismatch")
	}
	for index := 0; index < 8; index++ {
		wantName := fmt.Sprintf("talenro-c12-%s-pitr-candidate-%02d", descriptor.RunSuffix, index)
		if descriptor.CandidateNames[index] != wantName || descriptor.CandidateDataNames[index] != wantName+"-data" {
			return errors.New("authority PITR candidate registry mismatch")
		}
	}
	if !filepath.IsAbs(descriptor.DockerExecutable) || !strings.EqualFold(filepath.Ext(descriptor.DockerExecutable), ".exe") || !c12AuthorityPITRDigestPattern.MatchString(descriptor.DockerExecutableDigest) {
		return errors.New("authority PITR Docker executable binding mismatch")
	}
	digest, err := c12AuthorityPITRFileSHA256(descriptor.DockerExecutable)
	if err != nil || digest != descriptor.DockerExecutableDigest {
		return errors.New("authority PITR Docker executable digest mismatch")
	}
	return verifyC12AuthorityPITRDescriptor(descriptor, key[:])
}

func (state *c12AuthorityPITRState) verifyPrimaryAndRuntime(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	identity, err := state.dockerOne(ctx, "container", "inspect", "--format", "{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}|{{.Config.Image}}|{{.Image}}", state.descriptor.PrimaryID)
	if err != nil {
		return errors.New("inspect authority PITR primary")
	}
	wantIdentity := strings.Join([]string{state.descriptor.PrimaryID, "/" + state.descriptor.PrimaryName, "true", state.descriptor.RunSuffix, "authority-v7-pitr", "primary", state.descriptor.NonceDigest, state.descriptor.ImageRef, state.descriptor.ImageID}, "|")
	if identity != wantIdentity {
		return errors.New("authority PITR primary identity mismatch")
	}
	var walLevel, archiveMode, fullPageWrites, maxPrepared, sslEnabled string
	if err := state.database.QueryRowContext(ctx, `SELECT current_setting('wal_level'), current_setting('archive_mode'), current_setting('full_page_writes'), current_setting('max_prepared_transactions'), current_setting('ssl')`).Scan(&walLevel, &archiveMode, &fullPageWrites, &maxPrepared, &sslEnabled); err != nil || walLevel != "logical" || archiveMode != "on" || fullPageWrites != "on" || sslEnabled != "on" {
		return errors.New("authority PITR runtime settings mismatch")
	}
	preparedLimit, err := strconv.Atoi(maxPrepared)
	if err != nil || preparedLimit < 2 {
		return errors.New("authority PITR prepared-transaction limit mismatch")
	}
	var databaseIdentity []byte
	if err := state.database.QueryRowContext(ctx, `SELECT database_identity_digest FROM nodecontrol.control_plane_authority_protocol_migration_latches`).Scan(&databaseIdentity); err != nil || hex.EncodeToString(databaseIdentity) != state.descriptor.DatabaseIdentity {
		return errors.New("authority PITR migration latch identity mismatch")
	}
	var certificatePEM string
	if err := state.database.QueryRowContext(ctx, `SELECT pg_catalog.pg_read_file('server.crt')`).Scan(&certificatePEM); err != nil {
		return errors.New("authority PITR TLS certificate read failed")
	}
	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return errors.New("authority PITR TLS certificate PEM mismatch")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("authority PITR TLS certificate parse failed")
	}
	tlsDigest := sha256.Sum256(certificate.Raw)
	if hex.EncodeToString(tlsDigest[:]) != state.descriptor.TLSPublicDigest {
		return errors.New("authority PITR TLS public digest mismatch")
	}
	if err := state.installObserverAndSlot(ctx); err != nil {
		return err
	}
	return nil
}

func (state *c12AuthorityPITRState) installObserverAndSlot(ctx context.Context) error {
	passwordDigest := sha256.Sum256(append([]byte("TALENRO-C12-PITR-OBSERVER-V1\x00"), state.nonce[:]...))
	observerPassword := hex.EncodeToString(passwordDigest[:])
	quotedRole := `"` + state.descriptor.ObserverRole + `"`
	statements := []string{
		"CREATE ROLE " + quotedRole + " LOGIN REPLICATION NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT PASSWORD '" + observerPassword + "'",
		"ALTER ROLE " + quotedRole + " SET default_transaction_read_only = on",
		"GRANT CONNECT ON DATABASE " + `"` + strings.TrimPrefix(state.databaseURL.Path, "/") + `"` + " TO " + quotedRole,
		"GRANT USAGE ON SCHEMA nodecontrol TO " + quotedRole,
		"GRANT SELECT ON ALL TABLES IN SCHEMA nodecontrol TO " + quotedRole,
		`CREATE TABLE public.c12_authority_pitr_probe(probe_id text PRIMARY KEY, commit_kind text NOT NULL, created_at timestamptz NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := state.database.ExecContext(ctx, statement); err != nil {
			return errors.New("authority PITR observer/probe initialization failed")
		}
	}
	var slotName, startLSN string
	if err := state.database.QueryRowContext(ctx, `SELECT slot_name::text, lsn::text FROM pg_catalog.pg_create_logical_replication_slot($1,$2,false,true,true)`, state.descriptor.SlotName, state.descriptor.Plugin).Scan(&slotName, &startLSN); err != nil || slotName != state.descriptor.SlotName || !validC12AuthorityPITRLSN(startLSN) {
		return errors.New("authority PITR logical slot initialization failed")
	}
	var plugin string
	var twoPhase, failover bool
	if err := state.database.QueryRowContext(ctx, `SELECT plugin, two_phase, failover FROM pg_catalog.pg_replication_slots WHERE slot_name=$1`, state.descriptor.SlotName).Scan(&plugin, &twoPhase, &failover); err != nil || plugin != state.descriptor.Plugin || !twoPhase || !failover {
		return errors.New("authority PITR logical slot facts mismatch")
	}
	observerURL := *state.databaseURL
	observerURL.User = url.UserPassword(state.descriptor.ObserverRole, observerPassword)
	query := observerURL.Query()
	query.Set("sslmode", "require")
	query.Set("application_name", state.descriptor.ApplicationName+"-observer")
	observerURL.RawQuery = query.Encode()
	observer, err := sql.Open("pgx", observerURL.String())
	if err != nil {
		return errors.New("authority PITR TLS observer open failed")
	}
	defer observer.Close()
	var readOnly string
	var tls bool
	if err := observer.QueryRowContext(ctx, `SELECT current_setting('transaction_read_only'), COALESCE((SELECT ssl FROM pg_catalog.pg_stat_ssl WHERE pid=pg_catalog.pg_backend_pid()),false)`).Scan(&readOnly, &tls); err != nil || readOnly != "on" || !tls {
		return errors.New("authority PITR TLS observer facts mismatch")
	}
	if _, err := observer.ExecContext(ctx, `INSERT INTO public.c12_authority_pitr_probe(probe_id,commit_kind,created_at) VALUES('observer-write','forbidden',clock_timestamp())`); err == nil {
		return errors.New("authority PITR observer unexpectedly mutated the database")
	}
	return nil
}

func (controller C12AuthorityPITRController) CreateBaseBackup(ctx context.Context) (C12AuthorityPITRBaseBackup, error) {
	if controller.state == nil || !controller.state.phase.CompareAndSwap(1, 10) {
		return C12AuthorityPITRBaseBackup{}, errors.New("authority PITR base backup already consumed")
	}
	state := controller.state
	if err := state.wal.append("INTENT", "basebackup", state.descriptor.BaseBackupName, "", c12AuthorityPITRLabels(state.descriptor, "basebackup"), 0, state.descriptor.BaseBackupName, "creating", ""); err != nil {
		return C12AuthorityPITRBaseBackup{}, err
	}
	var startLSN string
	if err := state.database.QueryRowContext(ctx, `SELECT pg_catalog.pg_current_wal_flush_lsn()::text`).Scan(&startLSN); err != nil || !validC12AuthorityPITRLSN(startLSN) {
		return C12AuthorityPITRBaseBackup{}, errors.New("authority PITR base-backup start LSN failed")
	}
	databaseName := strings.TrimPrefix(state.databaseURL.Path, "/")
	if _, err := state.docker(ctx, "container", "exec", "--user", "postgres", state.descriptor.PrimaryID,
		"pg_basebackup", "--pgdata=/basebackup", "--format=plain", "--wal-method=stream", "--checkpoint=fast", "--no-password", "--username=talenro", "--dbname=dbname="+databaseName, "--label=talenro-c12-"+state.descriptor.RunSuffix); err != nil {
		return C12AuthorityPITRBaseBackup{}, errors.New("authority PITR pg_basebackup failed")
	}
	manifest, err := state.dockerOne(ctx, "container", "exec", state.descriptor.PrimaryID, "test", "-f", "/basebackup/backup_manifest")
	if err != nil || manifest != "" {
		return C12AuthorityPITRBaseBackup{}, errors.New("authority PITR backup manifest missing")
	}
	var endLSN string
	if err := state.database.QueryRowContext(ctx, `SELECT pg_catalog.pg_current_wal_flush_lsn()::text`).Scan(&endLSN); err != nil || !validC12AuthorityPITRLSN(endLSN) {
		return C12AuthorityPITRBaseBackup{}, errors.New("authority PITR base-backup end LSN failed")
	}
	backupDigest := sha256.Sum256([]byte(state.descriptor.RunSuffix + "\x00" + startLSN + "\x00" + endLSN))
	backup := C12AuthorityPITRBaseBackup{BackupID: hex.EncodeToString(backupDigest[:]), StartLSN: startLSN, EndLSN: endLSN}
	state.mu.Lock()
	state.baseBackup = backup
	state.mu.Unlock()
	state.phase.Store(2)
	if err := state.wal.append("ACTUAL", "basebackup", state.descriptor.BaseBackupName, backup.BackupID, c12AuthorityPITRLabels(state.descriptor, "basebackup"), 0, state.descriptor.BaseBackupName, "base_backup_complete", endLSN); err != nil {
		return C12AuthorityPITRBaseBackup{}, err
	}
	return backup, nil
}

func (controller C12AuthorityPITRController) CrashPrimary(ctx context.Context, backup C12AuthorityPITRBaseBackup, terminal C12AuthorityPITRCommit) (C12AuthorityPITRCut, error) {
	if controller.state == nil || !controller.state.phase.CompareAndSwap(2, 11) || !validC12AuthorityPITRLSN(terminal.EndLSN) {
		return C12AuthorityPITRCut{}, errors.New("invalid authority PITR crash cut")
	}
	state := controller.state
	state.mu.Lock()
	wantBackup, wantTerminal := state.baseBackup, state.terminalCommit
	state.mu.Unlock()
	if backup != wantBackup || terminal != wantTerminal || terminal.SQLXID == 0 ||
		(terminal.Kind != C12AuthorityPITRCommitImmediate && terminal.Kind != C12AuthorityPITRCommitPrepared) ||
		(terminal.Kind == C12AuthorityPITRCommitPrepared) != (terminal.GID != "") {
		return C12AuthorityPITRCut{}, errors.New("authority PITR crash cut is not controller-bound")
	}
	var archivedSegment string
	if err := state.database.QueryRowContext(ctx, `SELECT pg_catalog.pg_walfile_name(pg_catalog.pg_switch_wal())`).Scan(&archivedSegment); err != nil || archivedSegment == "" {
		return C12AuthorityPITRCut{}, errors.New("authority PITR WAL switch failed")
	}
	archiveDeadline := time.Now().Add(30 * time.Second)
	for {
		_, err := state.docker(ctx, "container", "exec", state.descriptor.PrimaryID, "test", "-f", "/archive/"+archivedSegment)
		if err == nil {
			break
		}
		if time.Now().After(archiveDeadline) {
			return C12AuthorityPITRCut{}, errors.New("authority PITR terminal WAL was not archived")
		}
		select {
		case <-ctx.Done():
			return C12AuthorityPITRCut{}, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	if err := state.wal.append("INTENT", "container", state.descriptor.PrimaryName, state.descriptor.PrimaryID, c12AuthorityPITRLabels(state.descriptor, "primary"), 0, "", "crashing", terminal.EndLSN); err != nil {
		return C12AuthorityPITRCut{}, err
	}
	if _, err := state.docker(ctx, "container", "kill", state.descriptor.PrimaryID); err != nil {
		return C12AuthorityPITRCut{}, errors.New("authority PITR primary crash failed")
	}
	status, err := state.dockerOne(ctx, "container", "inspect", "--format", "{{.State.Status}}", state.descriptor.PrimaryID)
	if err != nil || status != "exited" {
		return C12AuthorityPITRCut{}, errors.New("authority PITR primary did not reach crashed state")
	}
	cut := C12AuthorityPITRCut{BackupID: backup.BackupID, TerminalCommit: terminal, RecoveryTargetLSN: terminal.EndLSN}
	state.mu.Lock()
	state.cut = cut
	state.mu.Unlock()
	state.phase.Store(3)
	_ = state.database.Close()
	if err := state.wal.append("ACTUAL", "container", state.descriptor.PrimaryName, state.descriptor.PrimaryID, c12AuthorityPITRLabels(state.descriptor, "primary"), 0, "", "primary_crashed", terminal.EndLSN); err != nil {
		return C12AuthorityPITRCut{}, err
	}
	return cut, nil
}

func (controller C12AuthorityPITRController) RestoreAtCut(ctx context.Context, cut C12AuthorityPITRCut) (C12AuthorityPITRCandidate, error) {
	if controller.state == nil || controller.state.phase.Load() != 3 || !validC12AuthorityPITRLSN(cut.RecoveryTargetLSN) {
		return C12AuthorityPITRCandidate{}, errors.New("invalid authority PITR restore cut")
	}
	state := controller.state
	state.mu.Lock()
	wantCut := state.cut
	state.mu.Unlock()
	if cut != wantCut {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR restore cut is not controller-bound")
	}
	index := state.nextCandidate.Add(1) - 1
	if index >= uint32(state.descriptor.MaxCandidates) || !state.candidatePhase[index].CompareAndSwap(0, 1) {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate limit exceeded")
	}
	name := state.descriptor.CandidateNames[index]
	dataName := state.descriptor.CandidateDataNames[index]
	role := fmt.Sprintf("candidate-%02d-data", index)
	labels := c12AuthorityPITRLabels(state.descriptor, role)
	if err := state.wal.append("INTENT", "volume", dataName, "", labels, 0, dataName, "allocating", cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	if state.descriptor.FailureSeam == "after-intent-before-create" {
		return C12AuthorityPITRCandidate{}, errors.New("injected PITR seam after-intent-before-create")
	}
	output, err := state.dockerOne(ctx, "volume", "create", "--label", "talenro.c12.managed=true", "--label", "talenro.c12.run="+state.descriptor.RunSuffix,
		"--label", "talenro.c12.profile=authority-v7-pitr", "--label", "talenro.c12.role="+role, "--label", "talenro.c12.nonce-digest="+state.descriptor.NonceDigest, dataName)
	if err != nil || output != dataName {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate volume create failed")
	}
	state.mu.Lock()
	state.candidates[index].volumeExists = true
	state.candidates[index].targetLSN = cut.RecoveryTargetLSN
	state.mu.Unlock()
	if state.descriptor.FailureSeam == "after-create-before-actual" {
		return C12AuthorityPITRCandidate{}, errors.New("injected PITR seam after-create-before-actual")
	}
	if err := state.verifyVolume(ctx, dataName, role); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	if err := state.wal.append("ACTUAL", "volume", dataName, dataName, labels, 0, dataName, "allocated", cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	if state.descriptor.FailureSeam == "after-actual-before-return" {
		return C12AuthorityPITRCandidate{}, errors.New("injected PITR seam after-actual-before-return")
	}
	state.candidatePhase[index].Store(2)
	containerRole := fmt.Sprintf("candidate-%02d", index)
	containerLabels := c12AuthorityPITRLabels(state.descriptor, containerRole)
	if err := state.wal.append("INTENT", "container", name, "", containerLabels, 0, "", "restoring", cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	recoveryScript := "set -eu; mkdir -p \"$PGDATA\"; cp -a /source/. \"$PGDATA/\"; touch \"$PGDATA/recovery.signal\"; printf \"%s\\n\" \"restore_command = 'cp /archive/%f %p'\" \"recovery_target_lsn = '" + cut.RecoveryTargetLSN + "'\" \"recovery_target_inclusive = 'true'\" \"recovery_target_action = 'pause'\" >> \"$PGDATA/postgresql.auto.conf\"; chown -R postgres:postgres \"$PGDATA\"; exec su-exec postgres postgres"
	containerID, err := state.dockerOne(ctx, "run", "--detach", "--name", name,
		"--label", "talenro.c12.managed=true", "--label", "talenro.c12.run="+state.descriptor.RunSuffix,
		"--label", "talenro.c12.profile=authority-v7-pitr", "--label", "talenro.c12.role="+containerRole,
		"--label", "talenro.c12.nonce-digest="+state.descriptor.NonceDigest,
		"--publish", "127.0.0.1::5432", "--env", "PGDATA=/var/lib/postgresql/18/docker",
		"--mount", "type=volume,src="+dataName+",dst=/var/lib/postgresql",
		"--mount", "type=volume,src="+state.descriptor.BaseBackupName+",dst=/source,readonly",
		"--mount", "type=volume,src="+state.descriptor.ArchiveName+",dst=/archive,readonly",
		state.descriptor.ImageRef, "sh", "-ceu", recoveryScript)
	if err != nil || !c12AuthorityPITRContainerPattern.MatchString(containerID) {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate container create failed")
	}
	state.mu.Lock()
	state.candidates[index].containerID = containerID
	state.candidates[index].containerUp = true
	state.mu.Unlock()
	port, err := state.verifyCandidateContainer(ctx, int(index), containerID)
	if err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	state.mu.Lock()
	state.candidates[index].port = port
	state.mu.Unlock()
	if err := state.wal.append("ACTUAL", "container", name, containerID, containerLabels, port, "", "restoring", cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	if err := state.waitForCandidateCut(ctx, int(index), cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	state.candidatePhase[index].Store(3)
	if err := state.wal.append("TRANSITION", "candidate", name, containerID, containerLabels, port, "", "paused_at_cut", cut.RecoveryTargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	return C12AuthorityPITRCandidate{Index: uint8(index), Name: name, DataName: dataName, TargetLSN: cut.RecoveryTargetLSN}, nil
}

func (controller C12AuthorityPITRController) PromoteCandidate(ctx context.Context, candidate C12AuthorityPITRCandidate) (C12AuthorityPITRCandidate, error) {
	if controller.state == nil || candidate.Index >= 8 || candidate.Promoted || !validC12AuthorityPITRLSN(candidate.TargetLSN) ||
		!controller.state.candidatePhase[candidate.Index].CompareAndSwap(3, 4) {
		return C12AuthorityPITRCandidate{}, errors.New("invalid authority PITR candidate promotion")
	}
	state := controller.state
	if candidate.Name != state.descriptor.CandidateNames[candidate.Index] || candidate.DataName != state.descriptor.CandidateDataNames[candidate.Index] {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate promotion identity mismatch")
	}
	database, err := state.openCandidateDatabase(int(candidate.Index))
	if err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	defer database.Close()
	var promoted bool
	if err := database.QueryRowContext(ctx, `SELECT pg_catalog.pg_promote(true,60)`).Scan(&promoted); err != nil || !promoted {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate promotion failed")
	}
	var inRecovery bool
	if err := database.QueryRowContext(ctx, `SELECT pg_catalog.pg_is_in_recovery()`).Scan(&inRecovery); err != nil || inRecovery {
		return C12AuthorityPITRCandidate{}, errors.New("authority PITR candidate remained in recovery")
	}
	candidate.Promoted = true
	if err := state.wal.append("TRANSITION", "candidate", candidate.Name, state.candidates[candidate.Index].containerID, c12AuthorityPITRLabels(state.descriptor, fmt.Sprintf("candidate-%02d", candidate.Index)), state.candidates[candidate.Index].port, "", "promoted", candidate.TargetLSN); err != nil {
		return C12AuthorityPITRCandidate{}, err
	}
	return candidate, nil
}

func (controller C12AuthorityPITRController) InspectTimeline(ctx context.Context, candidate C12AuthorityPITRCandidate) (C12AuthorityPITRTimeline, error) {
	if controller.state == nil || candidate.Index >= 8 || !candidate.Promoted ||
		!controller.state.candidatePhase[candidate.Index].CompareAndSwap(4, 5) {
		return C12AuthorityPITRTimeline{}, errors.New("invalid authority PITR timeline inspection")
	}
	state := controller.state
	database, err := state.openCandidateDatabase(int(candidate.Index))
	if err != nil {
		return C12AuthorityPITRTimeline{}, err
	}
	defer database.Close()
	var timeline uint64
	var replayLSN string
	if err := database.QueryRowContext(ctx, `SELECT timeline_id::bigint, pg_catalog.pg_current_wal_flush_lsn()::text FROM pg_catalog.pg_control_checkpoint()`).Scan(&timeline, &replayLSN); err != nil || timeline < 2 || !validC12AuthorityPITRLSN(replayLSN) {
		return C12AuthorityPITRTimeline{}, errors.New("authority PITR timeline inspection failed")
	}
	timelineResult := C12AuthorityPITRTimeline{CandidateIndex: candidate.Index, TimelineID: timeline, ReplayLSN: normalizeC12AuthorityPITRLSN(replayLSN), Promoted: true}
	if err := state.wal.append("TRANSITION", "candidate", candidate.Name, state.candidates[candidate.Index].containerID, c12AuthorityPITRLabels(state.descriptor, fmt.Sprintf("candidate-%02d", candidate.Index)), state.candidates[candidate.Index].port, "", "timeline_inspected", timelineResult.ReplayLSN); err != nil {
		return C12AuthorityPITRTimeline{}, err
	}
	return timelineResult, nil
}

func issueC12AuthorityPITRCommit(ctx context.Context, controller C12AuthorityPITRController, kind C12AuthorityPITRCommitKind) (C12AuthorityPITRCommit, error) {
	if controller.state == nil || controller.state.phase.Load() != 2 || (kind != C12AuthorityPITRCommitImmediate && kind != C12AuthorityPITRCommitPrepared) {
		return C12AuthorityPITRCommit{}, errors.New("invalid authority PITR commit probe")
	}
	state := controller.state
	state.mu.Lock()
	state.commitSequence++
	sequence := state.commitSequence
	state.mu.Unlock()
	probeID := fmt.Sprintf("%s-%s-%d", state.descriptor.RunSuffix, kind, sequence)
	var xid uint64
	gid := ""
	if kind == C12AuthorityPITRCommitImmediate {
		tx, err := state.database.BeginTx(ctx, nil)
		if err != nil {
			return C12AuthorityPITRCommit{}, errors.New("begin immediate authority PITR probe")
		}
		defer tx.Rollback()
		if err := tx.QueryRowContext(ctx, `SELECT pg_catalog.txid_current()`).Scan(&xid); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("read immediate authority PITR xid")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO public.c12_authority_pitr_probe(probe_id,commit_kind,created_at) VALUES($1,$2,clock_timestamp())`, probeID, string(kind)); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("insert immediate authority PITR probe")
		}
		if err := tx.Commit(); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("commit immediate authority PITR probe")
		}
	} else {
		connection, err := state.database.Conn(ctx)
		if err != nil {
			return C12AuthorityPITRCommit{}, errors.New("open prepared authority PITR connection")
		}
		defer connection.Close()
		gid = fmt.Sprintf("talenro_c12_%s_%d", state.descriptor.RunSuffix[:16], sequence)
		if _, err := connection.ExecContext(ctx, `BEGIN`); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("begin prepared authority PITR probe")
		}
		if err := connection.QueryRowContext(ctx, `SELECT pg_catalog.txid_current()`).Scan(&xid); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("read prepared authority PITR xid")
		}
		if _, err := connection.ExecContext(ctx, `INSERT INTO public.c12_authority_pitr_probe(probe_id,commit_kind,created_at) VALUES($1,$2,clock_timestamp())`, probeID, string(kind)); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("insert prepared authority PITR probe")
		}
		if _, err := connection.ExecContext(ctx, `PREPARE TRANSACTION '`+gid+`'`); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("prepare authority PITR probe")
		}
		if _, err := connection.ExecContext(ctx, `COMMIT PREPARED '`+gid+`'`); err != nil {
			return C12AuthorityPITRCommit{}, errors.New("commit prepared authority PITR probe")
		}
	}
	rows, err := state.database.QueryContext(ctx, `SELECT lsn::text, xid::text, data FROM pg_catalog.pg_logical_slot_get_changes($1,NULL,NULL,'include-xids','1')`, state.descriptor.SlotName)
	if err != nil {
		return C12AuthorityPITRCommit{}, errors.New("read authority PITR logical terminal")
	}
	decoded := make([]c12AuthorityPITRDecodedRow, 0, 8)
	for rows.Next() {
		var row c12AuthorityPITRDecodedRow
		var rawXID string
		if err := rows.Scan(&row.LSN, &rawXID, &row.Data); err != nil {
			rows.Close()
			return C12AuthorityPITRCommit{}, errors.New("scan authority PITR logical terminal")
		}
		row.XID, err = strconv.ParseUint(rawXID, 10, 64)
		if err != nil {
			rows.Close()
			return C12AuthorityPITRCommit{}, errors.New("parse authority PITR logical xid")
		}
		decoded = append(decoded, row)
	}
	if err := rows.Err(); err != nil {
		return C12AuthorityPITRCommit{}, errors.New("iterate authority PITR logical terminal")
	}
	rows.Close()
	commit, err := parseC12AuthorityPITRTerminal(decoded, kind, xid, gid)
	if err != nil {
		return C12AuthorityPITRCommit{}, err
	}
	state.mu.Lock()
	state.terminalCommit = commit
	state.mu.Unlock()
	if err := state.wal.append("TRANSITION", "logical_commit", probeID, strconv.FormatUint(xid, 10), c12AuthorityPITRLabels(state.descriptor, "logical_commit"), 0, "", string(kind), commit.EndLSN); err != nil {
		return C12AuthorityPITRCommit{}, err
	}
	return commit, nil
}

func issueC12AuthorityPITRRollbackProbe(ctx context.Context, controller C12AuthorityPITRController) error {
	state := controller.state
	tx, err := state.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO public.c12_authority_pitr_probe(probe_id,commit_kind,created_at) VALUES($1,'rollback',clock_timestamp())`, state.descriptor.RunSuffix+"-rollback"); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Rollback(); err != nil {
		return err
	}
	rows, err := state.database.QueryContext(ctx, `SELECT lsn::text, xid::text, data FROM pg_catalog.pg_logical_slot_get_changes($1,NULL,NULL,'include-xids','1')`, state.descriptor.SlotName)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("authority PITR logical decoding exposed a rolled-back probe")
	}
	return rows.Err()
}

func parseC12AuthorityPITRTerminal(rows []c12AuthorityPITRDecodedRow, kind C12AuthorityPITRCommitKind, xid uint64, gid string) (C12AuthorityPITRCommit, error) {
	if xid == 0 || (kind == C12AuthorityPITRCommitPrepared) != (gid != "") {
		return C12AuthorityPITRCommit{}, errors.New("invalid authority PITR terminal expectation")
	}
	terminals := make([]c12AuthorityPITRDecodedRow, 0, 1)
	for _, row := range rows {
		upper := strings.ToUpper(strings.TrimSpace(row.Data))
		isImmediate := strings.HasPrefix(upper, "COMMIT ") && !strings.HasPrefix(upper, "COMMIT PREPARED")
		isPrepared := strings.HasPrefix(upper, "COMMIT PREPARED")
		if !isImmediate && !isPrepared {
			continue
		}
		if row.XID == xid {
			terminals = append(terminals, row)
		}
	}
	if len(terminals) != 1 {
		return C12AuthorityPITRCommit{}, errors.New("authority PITR logical terminal count mismatch")
	}
	terminal := terminals[0]
	upper := strings.ToUpper(terminal.Data)
	if kind == C12AuthorityPITRCommitImmediate && strings.HasPrefix(upper, "COMMIT PREPARED") {
		return C12AuthorityPITRCommit{}, errors.New("authority PITR immediate terminal kind mismatch")
	}
	if kind == C12AuthorityPITRCommitPrepared && (!strings.HasPrefix(upper, "COMMIT PREPARED") || !strings.Contains(terminal.Data, gid)) {
		return C12AuthorityPITRCommit{}, errors.New("authority PITR prepared terminal GID mismatch")
	}
	endLSN := normalizeC12AuthorityPITRLSN(terminal.LSN)
	if !validC12AuthorityPITRLSN(endLSN) {
		return C12AuthorityPITRCommit{}, errors.New("authority PITR logical terminal LSN mismatch")
	}
	return C12AuthorityPITRCommit{Kind: kind, SQLXID: xid, EndLSN: endLSN, GID: gid}, nil
}

func cleanupC12AuthorityPITR(ctx context.Context, controller C12AuthorityPITRController) error {
	if controller.state == nil {
		return nil
	}
	state := controller.state
	var failures []string
	for index := 7; index >= 0; index-- {
		state.mu.Lock()
		candidate := state.candidates[index]
		state.mu.Unlock()
		name := state.descriptor.CandidateNames[index]
		dataName := state.descriptor.CandidateDataNames[index]
		if candidate.containerID != "" {
			labels := c12AuthorityPITRLabels(state.descriptor, fmt.Sprintf("candidate-%02d", index))
			if err := state.wal.append("CLEAN_INTENT", "container", name, candidate.containerID, labels, candidate.port, "", "removing", candidate.targetLSN); err != nil {
				failures = append(failures, err.Error())
				continue
			}
			if state.descriptor.FailureSeam == "after-clean-intent-before-remove" {
				_ = state.wal.append("RECOVERED_ACTUAL", "container", name, candidate.containerID, labels, candidate.port, "", "recovered_after_clean_intent", candidate.targetLSN)
			}
			_, _ = state.docker(ctx, "container", "stop", "--time", "2", candidate.containerID)
			if _, err := state.docker(ctx, "container", "rm", candidate.containerID); err != nil {
				failures = append(failures, "remove exact authority PITR candidate")
			} else {
				if state.descriptor.FailureSeam == "after-remove-before-clean-result" {
					_ = state.wal.append("NOT_FOUND", "container", name, candidate.containerID, labels, candidate.port, "", "recovered_absence", candidate.targetLSN)
				}
				_ = state.wal.append("CLEAN_RESULT", "container", name, candidate.containerID, labels, candidate.port, "", "absent", candidate.targetLSN)
			}
		}
		if candidate.volumeExists {
			labels := c12AuthorityPITRLabels(state.descriptor, fmt.Sprintf("candidate-%02d-data", index))
			_ = state.wal.append("CLEAN_INTENT", "volume", dataName, dataName, labels, 0, dataName, "removing", candidate.targetLSN)
			if _, err := state.docker(ctx, "volume", "rm", dataName); err != nil {
				failures = append(failures, "remove exact authority PITR candidate volume")
			} else {
				_ = state.wal.append("CLEAN_RESULT", "volume", dataName, dataName, labels, 0, dataName, "absent", candidate.targetLSN)
			}
		}
	}
	if err := state.wal.verify(); err != nil {
		failures = append(failures, err.Error())
	}
	if len(failures) != 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func (state *c12AuthorityPITRState) waitForCandidateCut(ctx context.Context, index int, targetLSN string) error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		database, err := state.openCandidateDatabase(index)
		if err == nil {
			var inRecovery bool
			var replayLSN sql.NullString
			err = database.QueryRowContext(ctx, `SELECT pg_catalog.pg_is_in_recovery(), pg_catalog.pg_last_wal_replay_lsn()::text`).Scan(&inRecovery, &replayLSN)
			_ = database.Close()
			if err == nil && inRecovery && replayLSN.Valid && c12AuthorityPITRLSNGreaterOrEqual(replayLSN.String, targetLSN) {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return errors.New("authority PITR candidate did not pause at the recovery cut")
}

func (state *c12AuthorityPITRState) openCandidateDatabase(index int) (*sql.DB, error) {
	state.mu.Lock()
	port := state.candidates[index].port
	state.mu.Unlock()
	if port < 1 || port > 65535 {
		return nil, errors.New("authority PITR candidate port is invalid")
	}
	candidateURL := *state.databaseURL
	candidateURL.Host = "127.0.0.1:" + strconv.Itoa(port)
	query := candidateURL.Query()
	query.Set("sslmode", "require")
	query.Set("application_name", state.descriptor.ApplicationName+fmt.Sprintf("-candidate-%02d", index))
	candidateURL.RawQuery = query.Encode()
	database, err := sql.Open("pgx", candidateURL.String())
	if err != nil {
		return nil, errors.New("open authority PITR candidate database")
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := database.PingContext(pingCtx); err != nil {
		database.Close()
		return nil, errors.New("ping authority PITR candidate database")
	}
	return database, nil
}

func (state *c12AuthorityPITRState) verifyCandidateContainer(ctx context.Context, index int, containerID string) (int, error) {
	identity, err := state.dockerOne(ctx, "container", "inspect", "--format", "{{.Id}}|{{.Name}}|{{ index .Config.Labels `talenro.c12.managed` }}|{{ index .Config.Labels `talenro.c12.run` }}|{{ index .Config.Labels `talenro.c12.profile` }}|{{ index .Config.Labels `talenro.c12.role` }}|{{ index .Config.Labels `talenro.c12.nonce-digest` }}|{{.Config.Image}}|{{.Image}}", containerID)
	if err != nil {
		return 0, errors.New("inspect authority PITR candidate identity")
	}
	want := strings.Join([]string{containerID, "/" + state.descriptor.CandidateNames[index], "true", state.descriptor.RunSuffix, "authority-v7-pitr", fmt.Sprintf("candidate-%02d", index), state.descriptor.NonceDigest, state.descriptor.ImageRef, state.descriptor.ImageID}, "|")
	if identity != want {
		return 0, errors.New("authority PITR candidate identity mismatch")
	}
	deadline := time.Now().Add(15 * time.Second)
	lastMapping := ""
	for {
		mapping, err := state.dockerOne(ctx, "container", "port", containerID, "5432/tcp")
		lastMapping = mapping
		if err == nil && strings.HasPrefix(mapping, "127.0.0.1:") {
			port, parseErr := strconv.Atoi(strings.TrimPrefix(mapping, "127.0.0.1:"))
			if parseErr == nil && port >= 1 && port <= 65535 {
				return port, nil
			}
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("authority PITR candidate mapped port mismatch: %q", lastMapping)
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (state *c12AuthorityPITRState) verifyVolume(ctx context.Context, name, role string) error {
	identity, err := state.dockerOne(ctx, "volume", "inspect", "--format", "{{.Name}}|{{.Driver}}|{{ index .Labels `talenro.c12.managed` }}|{{ index .Labels `talenro.c12.run` }}|{{ index .Labels `talenro.c12.profile` }}|{{ index .Labels `talenro.c12.role` }}|{{ index .Labels `talenro.c12.nonce-digest` }}", name)
	if err != nil {
		return errors.New("inspect authority PITR candidate volume")
	}
	want := strings.Join([]string{name, "local", "true", state.descriptor.RunSuffix, "authority-v7-pitr", role, state.descriptor.NonceDigest}, "|")
	if identity != want {
		return errors.New("authority PITR candidate volume identity mismatch")
	}
	return nil
}

func (state *c12AuthorityPITRState) docker(ctx context.Context, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, state.descriptor.DockerExecutable, arguments...)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return nil, err
	}
	if output.Len() > 8192 {
		return nil, errors.New("authority PITR Docker output exceeded 8192 bytes")
	}
	return output.Bytes(), nil
}

func (state *c12AuthorityPITRState) dockerOne(ctx context.Context, arguments ...string) (string, error) {
	output, err := state.docker(ctx, arguments...)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(string(output))
	if strings.ContainsAny(trimmed, "\r\n") {
		return "", errors.New("authority PITR Docker command returned multiple lines")
	}
	return trimmed, nil
}

func openC12AuthorityPITROWAL(path string, key [32]byte, maxLine, maxRecords int) (*c12AuthorityPITROWAL, error) {
	if maxLine != 8192 || maxRecords != 4096 {
		return nil, errors.New("authority PITR WAL bounds mismatch")
	}
	wal := &c12AuthorityPITROWAL{path: path, key: key, maxLine: maxLine, maxRecords: maxRecords, previous: strings.Repeat("0", 64)}
	if err := wal.verify(); err != nil {
		return nil, err
	}
	return wal, nil
}

func (wal *c12AuthorityPITROWAL) append(event, resource, name, identity, labels string, port int, path, state, lsn string) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()
	if wal.sequence >= uint64(wal.maxRecords) || !c12AuthorityPITRWALValueValid(event, resource, name, identity, labels, path, state, lsn) || port < 0 || port > 65535 {
		return errors.New("authority PITR WAL record is outside closed bounds")
	}
	record := c12AuthorityPITRWALRecord{
		Schema:         "talenro-c12-authority-pitr-ownership-wal/v1",
		Sequence:       wal.sequence + 1,
		PreviousDigest: wal.previous,
		Event:          event,
		Resource:       resource,
		Name:           name,
		Identity:       identity,
		Labels:         labels,
		Port:           port,
		Path:           path,
		State:          state,
		LSN:            lsn,
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
	}
	payload := strings.Join([]string{record.Event, record.Resource, record.Name, record.Identity, record.Labels, strconv.Itoa(record.Port), record.Path, record.State, record.LSN, record.Timestamp}, "\x00")
	payloadDigest := sha256.Sum256([]byte(payload))
	record.PayloadSHA256 = hex.EncodeToString(payloadDigest[:])
	base, err := c12AuthorityPITRWALBaseBytes(record)
	if err != nil {
		return errors.New("marshal authority PITR WAL record")
	}
	recordDigest := sha256.Sum256(base)
	record.RecordSHA256 = hex.EncodeToString(recordDigest[:])
	mac := hmac.New(sha256.New, wal.key[:])
	_, _ = mac.Write(recordDigest[:])
	record.HMACSHA256 = hex.EncodeToString(mac.Sum(nil))
	line, err := json.Marshal(record)
	if err != nil || len(line)+1 > wal.maxLine || !isC12AuthorityPITRASCII(line) {
		return errors.New("authority PITR WAL line is outside closed bounds")
	}
	file, err := os.OpenFile(wal.path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return errors.New("open authority PITR WAL for append")
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		file.Close()
		return errors.New("append authority PITR WAL")
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return errors.New("flush authority PITR WAL")
	}
	if err := file.Close(); err != nil {
		return errors.New("close authority PITR WAL")
	}
	wal.sequence = record.Sequence
	wal.previous = record.RecordSHA256
	return nil
}

func (wal *c12AuthorityPITROWAL) verify() error {
	wal.mu.Lock()
	defer wal.mu.Unlock()
	file, err := os.Open(wal.path)
	if err != nil {
		return errors.New("open authority PITR WAL")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), wal.maxLine)
	sequence := uint64(0)
	previous := strings.Repeat("0", 64)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		if len(line) == 0 || len(line)+1 > wal.maxLine || !isC12AuthorityPITRASCII(line) || sequence >= uint64(wal.maxRecords) {
			return errors.New("authority PITR WAL line is malformed")
		}
		var record c12AuthorityPITRWALRecord
		if err := json.Unmarshal(line, &record); err != nil || record.Schema != "talenro-c12-authority-pitr-ownership-wal/v1" || record.Sequence != sequence+1 || record.PreviousDigest != previous || !c12AuthorityPITRWALValueValid(record.Event, record.Resource, record.Name, record.Identity, record.Labels, record.Path, record.State, record.LSN) {
			return errors.New("authority PITR WAL chain is malformed")
		}
		payload := strings.Join([]string{record.Event, record.Resource, record.Name, record.Identity, record.Labels, strconv.Itoa(record.Port), record.Path, record.State, record.LSN, record.Timestamp}, "\x00")
		payloadDigest := sha256.Sum256([]byte(payload))
		if record.PayloadSHA256 != hex.EncodeToString(payloadDigest[:]) {
			return errors.New("authority PITR WAL payload digest mismatch")
		}
		base, err := c12AuthorityPITRWALBaseBytes(record)
		if err != nil {
			return errors.New("authority PITR WAL base encoding failed")
		}
		digest := sha256.Sum256(base)
		if record.RecordSHA256 != hex.EncodeToString(digest[:]) {
			return errors.New("authority PITR WAL record digest mismatch")
		}
		wantMAC, err := hex.DecodeString(record.HMACSHA256)
		if err != nil {
			return errors.New("authority PITR WAL HMAC encoding mismatch")
		}
		mac := hmac.New(sha256.New, wal.key[:])
		_, _ = mac.Write(digest[:])
		if !hmac.Equal(wantMAC, mac.Sum(nil)) {
			return errors.New("authority PITR WAL HMAC mismatch")
		}
		sequence = record.Sequence
		previous = record.RecordSHA256
	}
	if err := scanner.Err(); err != nil {
		return errors.New("read authority PITR WAL")
	}
	wal.sequence = sequence
	wal.previous = previous
	return nil
}

func c12AuthorityPITRWALBaseBytes(record c12AuthorityPITRWALRecord) ([]byte, error) {
	record.RecordSHA256 = ""
	record.HMACSHA256 = ""
	return json.Marshal(record)
}

func c12AuthorityPITRWALValueValid(values ...string) bool {
	if len(values) != 8 {
		return false
	}
	closedEvents := map[string]bool{"BOOTSTRAP": true, "INTENT": true, "ACTUAL": true, "RECOVERED_ACTUAL": true, "NOT_FOUND": true, "TRANSITION": true, "CLEAN_INTENT": true, "CLEAN_RESULT": true}
	if !closedEvents[values[0]] {
		return false
	}
	for _, value := range values {
		if len(value) > 1024 || strings.ContainsAny(value, "\x00\r\n") {
			return false
		}
	}
	return true
}

func c12AuthorityPITRLabels(descriptor c12AuthorityPITRDescriptor, role string) string {
	return "managed=true;run=" + descriptor.RunSuffix + ";profile=authority-v7-pitr;role=" + role + ";nonce-digest=" + descriptor.NonceDigest
}

func verifyC12AuthorityPITRDescriptor(value c12AuthorityPITRDescriptor, key []byte) error {
	digestText, hmacText := value.DescriptorDigest, value.DescriptorHMAC
	value.DescriptorDigest = ""
	value.DescriptorHMAC = ""
	body, err := json.Marshal(value)
	if err != nil {
		return errors.New("marshal authority PITR descriptor")
	}
	digest := sha256.Sum256(body)
	if digestText != hex.EncodeToString(digest[:]) {
		return errors.New("authority PITR descriptor digest mismatch")
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(digest[:])
	wantMAC, err := hex.DecodeString(hmacText)
	if err != nil || !hmac.Equal(wantMAC, mac.Sum(nil)) {
		return errors.New("authority PITR descriptor HMAC mismatch")
	}
	return nil
}

func c12AuthorityPITRFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func validC12AuthorityPITRLSN(value string) bool {
	return c12AuthorityPITRLSNPattern.MatchString(normalizeC12AuthorityPITRLSN(value))
}

func normalizeC12AuthorityPITRLSN(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func c12AuthorityPITRLSNGreaterOrEqual(left, right string) bool {
	leftValue, leftOK := c12AuthorityPITRLSNValue(left)
	rightValue, rightOK := c12AuthorityPITRLSNValue(right)
	return leftOK && rightOK && leftValue >= rightValue
}

func c12AuthorityPITRLSNValue(value string) (uint64, bool) {
	parts := strings.Split(normalizeC12AuthorityPITRLSN(value), "/")
	if len(parts) != 2 {
		return 0, false
	}
	high, errHigh := strconv.ParseUint(parts[0], 16, 32)
	low, errLow := strconv.ParseUint(parts[1], 16, 32)
	if errHigh != nil || errLow != nil {
		return 0, false
	}
	return high<<32 | low, true
}

func isC12AuthorityPITRASCII(value []byte) bool {
	for _, character := range value {
		if character < 0x20 || character > 0x7e {
			return false
		}
	}
	return true
}
