//go:build integration

package testinfra

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"talenro.local/platform/internal/store"
)

type C12AuthorityPITRCutSelection struct{ state *c12CutSelectionState }
type c12CutSelectionState struct {
	owner      *c12AuthorityPITRState
	generation uint64
	backup     C12AuthorityPITRBaseBackup
	target     *c12ObservedCommitState
}

// Fixed private catalog query, never registered as consumer SQL. Compare the
// complete effective UPDATE-column set in the domain schemas in both directions;
// checking only the three required grants would miss table-wide/inherited grants.
const c12CandidatePrivilegesSQL = `WITH required(table_name,column_name) AS (
 VALUES ('nodecontrol.control_plane_authority_fences','operation_id'),
        ('nodecontrol.node_certificates','certificate_id'),
        ('nodecontrol.authority_task7_crash_effects','operation_id')
), domain_tables AS (
 SELECT c.oid,n.nspname||'.'||c.relname AS table_name
 FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname IN ('nodecontrol','public') AND c.relkind IN ('r','p')
), updates AS (
 SELECT d.table_name,a.attname::text AS column_name FROM domain_tables d
 JOIN pg_catalog.pg_attribute a ON a.attrelid=d.oid
 WHERE a.attnum>0 AND NOT a.attisdropped AND pg_catalog.has_column_privilege(d.oid,a.attnum,'UPDATE')
)
SELECT current_user::text,session_user::text,
       r.rolsuper OR r.rolcreatedb OR r.rolcreaterole OR r.rolreplication OR r.rolbypassrls,
       current_setting('transaction_read_only'),
       NOT EXISTS (SELECT * FROM required EXCEPT SELECT * FROM updates),
       NOT EXISTS (SELECT * FROM updates EXCEPT SELECT * FROM required),
       EXISTS (SELECT 1 FROM domain_tables d WHERE
         pg_catalog.has_table_privilege(d.oid,'INSERT,DELETE,TRUNCATE,REFERENCES,TRIGGER') OR
         pg_catalog.has_any_column_privilege(d.oid,'INSERT,REFERENCES')),
       COALESCE((SELECT ssl FROM pg_catalog.pg_stat_ssl WHERE pid=pg_catalog.pg_backend_pid()),false)
FROM pg_catalog.pg_roles r WHERE r.rolname=current_user`

// pg_current_wal_insert_lsn is forbidden during recovery. This private narrower
// projection retains the frozen query's unsigned identity representation only.
const c12RecoveryIdentitySQL = `SELECT
 (CASE WHEN system_identifier < 0 THEN system_identifier::numeric + 18446744073709551616::numeric
       ELSE system_identifier::numeric END)::numeric(20,0) AS system_id,
 (CASE WHEN timeline_id < 0 THEN timeline_id::bigint + 4294967296::bigint
       ELSE timeline_id::bigint END)::bigint AS timeline
FROM pg_catalog.pg_control_system(), pg_catalog.pg_control_checkpoint()`

func c12VerifyCandidatePrivileges(ctx context.Context, db store.DBTX, role string) error {
	var current, session, readOnly string
	var unsafe, required, exact, broad, tls bool
	if role == "" {
		return C12PITRDependencyFailure
	}
	if err := db.QueryRow(ctx, c12CandidatePrivilegesSQL).Scan(&current, &session, &unsafe, &readOnly, &required, &exact, &broad, &tls); err != nil {
		return c12AccessError(err, false)
	}
	if current != role || session != role || unsafe || readOnly != "off" || !required || !exact || broad || !tls {
		return C12PITRDependencyFailure
	}
	return nil
}

type c12BackupBinding struct {
	owner       *c12AuthorityPITRState
	generation  uint64
	systemID    uint64
	timeline    uint64
	containerID string
}

func (controller C12AuthorityPITRController) SelectRecoveryCut(backup C12AuthorityPITRBaseBackup, observed C12AuthorityPITRObservedCommit) (C12AuthorityPITRCutSelection, error) {
	s := controller.state
	if s == nil {
		return C12AuthorityPITRCutSelection{}, C12PITRInvalidHandle
	}
	s.accessGate.Lock()
	defer s.accessGate.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.observationPoison {
		return C12AuthorityPITRCutSelection{}, C12PITRIndeterminate
	}
	if !s.validC12Backup(backup) || !s.validC12Observed(observed.state) {
		return C12AuthorityPITRCutSelection{}, C12PITRInvalidHandle
	}
	if s.phase.Load() != 2 || s.transitioning || s.accessActive {
		return C12AuthorityPITRCutSelection{}, C12PITRWrongPhase
	}
	if err := c12ValidateRecoveryOrder(backup.EndLSN, observed.state.facts.EndLSN, observed.state.facts.EndLSN); err != nil {
		return C12AuthorityPITRCutSelection{}, err
	}
	if s.selection != nil {
		if s.selection.backup == backup && s.selection.target == observed.state {
			return C12AuthorityPITRCutSelection{state: s.selection}, nil
		}
		return C12AuthorityPITRCutSelection{}, C12PITRCapacityExceeded
	}
	s.selection = &c12CutSelectionState{owner: s, generation: s.runGeneration, backup: backup, target: observed.state}
	return C12AuthorityPITRCutSelection{state: s.selection}, nil
}

func (controller C12AuthorityPITRController) CrashPrimaryAtCut(ctx context.Context, selection C12AuthorityPITRCutSelection, boundary C12AuthorityPITRObservedCommit) (C12AuthorityPITRCut, error) {
	s := controller.state
	if s == nil || ctx == nil {
		return C12AuthorityPITRCut{}, C12PITRInvalidHandle
	}
	if ctx.Err() != nil {
		return C12AuthorityPITRCut{}, C12PITRCanceled
	}
	err := s.claimC12Cut(func() error {
		cut := selection.state
		if cut == nil || cut != s.selection || cut.owner != s || cut.generation != s.runGeneration || !s.validC12Backup(cut.backup) || !s.validC12Observed(cut.target) || !s.validC12Observed(boundary.state) {
			return C12PITRInvalidHandle
		}
		return c12ValidateRecoveryOrder(cut.backup.EndLSN, cut.target.facts.EndLSN, boundary.state.facts.EndLSN)
	})
	if err != nil {
		return C12AuthorityPITRCut{}, err
	}
	defer s.endC12Transition()
	cut, err := c12CrashAtValidatedCut(ctx, s, selection.state.backup, selection.state.target.facts, boundary.state.facts)
	return cut, c12AccessError(err, false)
}

func c12ValidateRecoveryOrder(backupEnd, targetEnd, boundaryEnd string) error {
	b, bok := c12CanonicalLSN(backupEnd)
	a, aok := c12CanonicalLSN(targetEnd)
	z, zok := c12CanonicalLSN(boundaryEnd)
	if !bok || !aok || !zok || a < b || z < a {
		return C12PITRInvalidHandle
	}
	return nil
}

func c12CanonicalLSN(text string) (uint64, bool) {
	v, ok := c12AuthorityPITRLSNValue(text)
	return v, ok && text == fmt.Sprintf("%X/%X", v>>32, uint32(v))
}

// Callers hold accessGate then mu. Validation precedes any phase claim.
func (s *c12AuthorityPITRState) validC12Backup(b C12AuthorityPITRBaseBackup) bool {
	return b == s.baseBackup && b.binding != nil && b.binding.owner == s && b.binding.generation == s.runGeneration && s.runGeneration != 0 && b.binding.systemID != 0 && b.binding.timeline != 0 && b.binding.containerID == s.descriptor.PrimaryID
}
func (s *c12AuthorityPITRState) validC12Observed(r *c12ObservedCommitState) bool {
	return r != nil && r.owner == s && r.generation == s.runGeneration && s.observationRegistered(r.observation) && r.observation.result == r && r.facts.Kind == C12AuthorityPITRCommitImmediate && r.facts.GID == "" && r.facts.SQLXID != 0
}
func (s *c12AuthorityPITRState) claimC12Cut(validate func() error) error {
	s.accessGate.Lock()
	defer s.accessGate.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.observationPoison {
		return C12PITRIndeterminate
	}
	if err := validate(); err != nil {
		return err
	}
	if s.phase.Load() != 2 || s.accessActive || s.transitioning {
		return C12PITRWrongPhase
	}
	s.phase.Store(11)
	s.transitioning = true
	return nil
}

func c12PhysicalIdentity(ctx context.Context, db *sql.DB) (uint64, uint64, string, error) {
	var systemText, lsn string
	var timeline int64
	if err := db.QueryRowContext(ctx, c12GetNodeControlDatabaseIdentitySQL).Scan(&systemText, &timeline, &lsn); err != nil {
		return 0, 0, "", C12PITRDependencyFailure
	}
	system, err := strconv.ParseUint(systemText, 10, 64)
	_, valid := c12CanonicalLSN(lsn)
	if err != nil || system == 0 || timeline <= 0 || timeline > 4294967295 || !valid {
		return 0, 0, "", C12PITRDependencyFailure
	}
	return system, uint64(timeline), lsn, nil
}

func c12RecoveryPhysicalIdentity(ctx context.Context, db *sql.DB) (uint64, uint64, error) {
	var systemText string
	var timeline int64
	if err := db.QueryRowContext(ctx, c12RecoveryIdentitySQL).Scan(&systemText, &timeline); err != nil {
		return 0, 0, C12PITRDependencyFailure
	}
	system, err := strconv.ParseUint(systemText, 10, 64)
	if err != nil || system == 0 || timeline <= 0 || timeline > 4294967295 {
		return 0, 0, C12PITRDependencyFailure
	}
	return system, uint64(timeline), nil
}

func (s *c12AuthorityPITRState) validC12Candidate(c C12AuthorityPITRCandidate) bool {
	if c.Index >= 8 || !s.validC12Backup(s.baseBackup) || c.binding.owner != s || c.binding.runGeneration != s.runGeneration || c.binding.index != c.Index || c.binding.cut == nil || c.binding.cut != s.cut.binding {
		return false
	}
	r := s.candidates[c.Index]
	return r.dto == c && r.targetLSN == c.TargetLSN && r.containerID != "" && r.containerUp && r.volumeExists && r.systemID == s.baseBackup.binding.systemID && r.timeline >= s.baseBackup.binding.timeline
}

func (s *c12AuthorityPITRState) claimC12Candidate(c C12AuthorityPITRCandidate, phase uint32) error {
	s.accessGate.Lock()
	defer s.accessGate.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.validC12Candidate(c) {
		return C12PITRInvalidHandle
	}
	if s.accessActive || s.transitioning || s.candidatePhase[c.Index].Load() != phase {
		return C12PITRWrongPhase
	}
	s.transitioning = true
	// 6/7 are terminal on failure; only successful operations publish 4/5.
	if phase == 3 {
		s.candidatePhase[c.Index].Store(6)
	} else {
		s.candidatePhase[c.Index].Store(7)
	}
	return nil
}
