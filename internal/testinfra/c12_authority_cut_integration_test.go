//go:build integration

package testinfra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestC12AuthorityPITRRecoveryCutStateMachine(t *testing.T) {
	t.Run("new_crash_error_is_finite_and_terminal", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		a := f.observed("0/20", 7)
		sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil {
			t.Fatal(err)
		}
		f.failQuery = `SELECT pg_catalog.pg_walfile_name(pg_catalog.pg_switch_wal())`
		if _, err = c.CrashPrimaryAtCut(context.Background(), sel, a); err != C12PITRDependencyFailure {
			t.Fatal("new API leaked unbounded dependency error", err)
		}
		before := f.effects()
		if _, err = c.CrashPrimaryAtCut(context.Background(), sel, a); err != C12PITRWrongPhase || f.effects() != before || c.state.phase.Load() != 11 {
			t.Fatal("uncertain crash retried", err)
		}
	})
	t.Run("restore_requires_paused_recovery", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		f.paused = false
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		f.afterReplay = cancel
		if err := c.state.waitForCandidateCut(ctx, 0, "0/20"); err == nil {
			t.Fatal("replay at A was accepted before recovery actually paused")
		}
	})
	t.Run("identity_unsigned_boundaries", func(t *testing.T) {
		for _, tc := range []struct {
			system   string
			timeline int64
			ok       bool
		}{{"18446744073709551615", 4294967295, true}, {"18446744073709551616", 1, false}, {"-1", 1, false}, {"0", 1, false}, {"12345", 0, false}, {"12345", 4294967296, false}, {"12345", -1, false}} {
			c, f := newC12CutFixture(t)
			f.system = tc.system
			f.promoted = true
			f.promotedTimeline = tc.timeline
			f.recoveryTimeline = tc.timeline
			_, _, _, err := c12PhysicalIdentity(context.Background(), c.state.database)
			if (err == nil) != tc.ok {
				t.Fatalf("identity %v accepted incorrectly: %v", tc, err)
			}
			_, _, err = c12RecoveryPhysicalIdentity(context.Background(), c.state.database)
			if (err == nil) != tc.ok {
				t.Fatalf("recovery identity %v accepted incorrectly: %v", tc, err)
			}
		}
	})
	t.Run("restore_checks_physical_identity_before_issuance", func(t *testing.T) {
		for _, name := range []string{"system", "timeline"} {
			t.Run(name, func(t *testing.T) {
				c, f := newC12CutFixture(t)
				a := f.observed("0/20", 7)
				sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
				if err != nil {
					t.Fatal(err)
				}
				cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
				if err != nil {
					t.Fatal(err)
				}
				if name == "system" {
					f.system = "54321"
				} else {
					f.recoveryTimeline = 2
				}
				if candidate, err := c.RestoreAtCut(context.Background(), cut); err == nil || candidate.binding.owner != nil || c.state.candidates[0].dto.binding.owner != nil || c.state.candidatePhase[0].Load() != 2 {
					t.Fatal("invalid restored identity issued a candidate", err)
				}
			})
		}
	})
	t.Run("second_backend_is_verified_before_begin", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		a := f.observed("0/20", 7)
		sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil {
			t.Fatal(err)
		}
		cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
		if err != nil {
			t.Fatal(err)
		}
		candidate, err := c.RestoreAtCut(context.Background(), cut)
		if err != nil {
			t.Fatal(err)
		}
		candidate, err = c.PromoteCandidate(context.Background(), candidate)
		if err != nil {
			t.Fatal(err)
		}
		c.state.candidateRole = "candidate_role"
		events := &c12AccessTestFixture{}
		c.state.accessPolicy = func(*c12CandidateBinding, bool, c12AccessOperation, string) bool { return true }
		opens := 0
		c.state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) {
			opens++
			values := []any{"candidate_role", "candidate_role", false, "off", true, true, false, true}
			if opens == 2 {
				values[6] = true
			}
			return &c12PrivilegeBackend{c12AccessTestDriver: &c12AccessTestDriver{fixture: events}, values: values}, nil
		}
		err = c.WithCandidateAuthorityAccess(context.Background(), candidate, func(access C12AuthorityAccess) error {
			tx, e := access.Begin(context.Background())
			if e == nil || tx != nil {
				t.Fatal("bad second connection exposed transaction")
			}
			return nil
		})
		if err != nil || opens != 2 {
			t.Fatal("two-connection verifier", err, opens)
		}
		for _, event := range events.snapshot() {
			if event == "begin" {
				t.Fatal("bad second connection reached BeginTx")
			}
		}
	})
	t.Run("canceled_cut_has_no_effects", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		a := f.observed("0/20", 7)
		sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		before := f.effects()
		if _, err = c.CrashPrimaryAtCut(ctx, sel, a); err != C12PITRCanceled || f.effects() != before || c.state.phase.Load() != 2 {
			t.Fatalf("canceled cut consumed phase: %v phase=%d", err, c.state.phase.Load())
		}
	})
	t.Run("failed_candidate_operations_are_terminal", func(t *testing.T) {
		for _, phase := range []uint32{6, 7} {
			t.Run(fmt.Sprint(phase), func(t *testing.T) {
				c, f := newC12CutFixture(t)
				a := f.observed("0/20", 7)
				sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
				if err != nil {
					t.Fatal(err)
				}
				cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
				if err != nil {
					t.Fatal(err)
				}
				candidate, err := c.RestoreAtCut(context.Background(), cut)
				if err != nil {
					t.Fatal(err)
				}
				if phase == 7 {
					candidate, err = c.PromoteCandidate(context.Background(), candidate)
					if err != nil {
						t.Fatal(err)
					}
				}
				f.failQuery = c12GetNodeControlDatabaseIdentitySQL
				if phase == 6 {
					f.failQuery = `SELECT pg_catalog.pg_promote(true,60)`
				}
				if phase == 6 {
					_, err = c.PromoteCandidate(context.Background(), candidate)
				} else {
					_, err = c.InspectTimeline(context.Background(), candidate)
				}
				if err == nil || c.state.candidatePhase[0].Load() != phase {
					t.Fatal("failed operation was not terminal", err)
				}
				before := f.effects()
				if _, err = c.PromoteCandidate(context.Background(), candidate); err == nil {
					t.Fatal("failed promote retried")
				}
				if _, err = c.InspectTimeline(context.Background(), candidate); err == nil {
					t.Fatal("failed inspect retried")
				}
				called := false
				if err = c.WithCandidateAuthorityAccess(context.Background(), candidate, func(C12AuthorityAccess) error { called = true; return nil }); err == nil || called || f.effects() != before {
					t.Fatal("failed candidate remained usable", err)
				}
				if err = cleanupC12AuthorityPITR(context.Background(), c); err != nil {
					t.Fatal("failed candidate cleanup", err)
				}
			})
		}
	})
	t.Run("backup_retains_actual_physical_identity", func(t *testing.T) {
		for _, mismatch := range []bool{false, true} {
			t.Run(fmt.Sprint(mismatch), func(t *testing.T) {
				c, f := newC12CutFixture(t)
				c.state.phase.Store(1)
				c.state.baseBackup = C12AuthorityPITRBaseBackup{}
				c.state.fixture = &c12FixtureLedger{systemID: 12345, timeline: 3}
				c.state.databaseURL = &url.URL{Path: "/fixture"}
				c.state.descriptor.DatabaseIdentity = strings.Repeat("1", 64)
				if mismatch {
					f.system = "54321"
				}
				backup, err := c.CreateBaseBackup(context.Background())
				if mismatch {
					if err == nil {
						t.Fatal("backup accepted foreign physical system")
					}
					return
				}
				if err != nil || backup.binding == nil || backup.binding.systemID != 12345 || backup.binding.timeline != 3 || backup.binding.containerID != c.state.descriptor.PrimaryID || backup.binding.owner != c.state || backup.binding.generation != 1 {
					t.Fatalf("backup omitted actual tuple: backup=%+v err=%v", backup, err)
				}
			})
		}
	})
	t.Run("candidate_privileges", func(t *testing.T) {
		for _, name := range []string{"valid", "wrong_role", "wrong_session", "superuser", "read_only", "missing_column", "extra_column", "broad_dml", "no_tls", "query_error"} {
			t.Run(name, func(t *testing.T) {
				c, f := newC12CutFixture(t)
				a := f.observed("0/20", 7)
				sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
				if err != nil {
					t.Fatal(err)
				}
				cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
				if err != nil {
					t.Fatal(err)
				}
				candidate, err := c.RestoreAtCut(context.Background(), cut)
				if err != nil {
					t.Fatal(err)
				}
				candidate, err = c.PromoteCandidate(context.Background(), candidate)
				if err != nil {
					t.Fatal(err)
				}
				values := []any{"candidate_role", "candidate_role", false, "off", true, true, false, true}
				switch name {
				case "wrong_role":
					values[0] = "primary"
				case "wrong_session":
					values[1] = "primary"
				case "superuser":
					values[2] = true
				case "read_only":
					values[3] = "on"
				case "missing_column":
					values[4] = false
				case "extra_column":
					values[5] = false
				case "broad_dml":
					values[6] = true
				case "no_tls":
					values[7] = false
				}
				d := &c12PrivilegeBackend{c12AccessTestDriver: &c12AccessTestDriver{fixture: &c12AccessTestFixture{}}, values: values}
				if name == "query_error" {
					d.err = errors.New("query failed")
				}
				c.state.candidateRole = "candidate_role"
				c.state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) { return d, nil }
				err = c12VerifyCandidatePrivileges(context.Background(), d, "candidate_role")
				if (err == nil) != (name == "valid") || d.queries != 1 {
					t.Fatalf("private verifier admitted invalid catalog facts: err=%v queries=%d", err, d.queries)
				}
				called := false
				err = c.WithCandidateAuthorityAccess(context.Background(), candidate, func(C12AuthorityAccess) error { called = true; return nil })
				if called != (name == "valid") || (err == nil) != (name == "valid") {
					t.Fatalf("candidate exposed before verification: callback=%v err=%v", called, err)
				}
			})
		}
	})
	t.Run("candidate_access_exact_dto", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		a := f.observed("0/20", 7)
		sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil {
			t.Fatal(err)
		}
		cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
		if err != nil {
			t.Fatal(err)
		}
		pre, err := c.RestoreAtCut(context.Background(), cut)
		if err != nil {
			t.Fatal(err)
		}
		d := &c12PrivilegeBackend{c12AccessTestDriver: &c12AccessTestDriver{fixture: &c12AccessTestFixture{}}, values: []any{"candidate_role", "candidate_role", false, "off", true, true, false, true}}
		opens := 0
		c.state.candidateRole = "candidate_role"
		c.state.accessOpen = func(context.Context, *c12CandidateBinding) (c12AccessBackend, error) { opens++; return d, nil }
		called := false
		if err = c.WithCandidateAuthorityAccess(context.Background(), pre, func(C12AuthorityAccess) error { called = true; return nil }); err == nil || called || opens != 0 {
			t.Fatal("pre-promote access", err)
		}
		candidate, err := c.PromoteCandidate(context.Background(), pre)
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range []c12CandidateBinding{{owner: c.state, runGeneration: 2, index: 0, cut: candidate.binding.cut}, {owner: c.state, runGeneration: 1, index: 0, cut: &c12CutSelectionState{}}} {
			if err = c12WithAccess(context.Background(), c.state, &bad, func(C12AuthorityAccess) error { called = true; return nil }); err != C12PITRInvalidHandle || called || opens != 0 {
				t.Fatal("private binding resolver ignored supplied identity", err)
			}
		}
		for _, mutate := range []func(*C12AuthorityPITRCandidate){func(c *C12AuthorityPITRCandidate) { c.Name = "altered" }, func(c *C12AuthorityPITRCandidate) { c.TargetLSN = "0/40" }, func(c *C12AuthorityPITRCandidate) { c.Promoted = false }, func(c *C12AuthorityPITRCandidate) { c.DataName = "altered" }, func(c *C12AuthorityPITRCandidate) { *c = pre }} {
			bad := candidate
			mutate(&bad)
			if err = c.WithCandidateAuthorityAccess(context.Background(), bad, func(C12AuthorityAccess) error { called = true; return nil }); err == nil || called || opens != 0 {
				t.Fatal("forged access DTO", err)
			}
		}
		if err = c.WithCandidateAuthorityAccess(context.Background(), candidate, func(C12AuthorityAccess) error { called = true; return nil }); err != nil || !called {
			t.Fatal("legitimate access after rejection", err)
		}
	})
	t.Run("installed_diagnostics_after_partial_drain", func(t *testing.T) {
		c, d, drains := newC12ObservationTest(t)
		a := c12BoundObservation(t, c, true)
		c.state.phase.Store(2)
		b := c12BoundObservation(t, c, true)
		d.committed = append(c12DecodedCommit(7, a.state.marker, "0/30"), c12DecodedCommit(8, b.state.marker, "0/40")...)
		ah, af, err := c.ObserveCommit(context.Background(), a)
		if err != nil {
			t.Fatal(err)
		}
		bh, bf, err := c.ObserveCommit(context.Background(), b)
		if err != nil {
			t.Fatal(err)
		}
		fixture, _ := newC12CutFixture(t)
		backup := fixture.state.baseBackup
		backup.binding.owner = c.state
		backup.binding.containerID = c.state.descriptor.PrimaryID
		c.state.baseBackup = backup
		c.state.wal = fixture.state.wal
		sel, err := c.SelectRecoveryCut(backup, ah)
		if err != nil {
			t.Fatal(err)
		}
		pending := c12BoundObservation(t, c, true)
		c.state.observationQuery = func(context.Context) (pgx.Rows, error) {
			*drains++
			return &c12ObservationTestRows{rows: d.committed, failAt: 1}, nil
		}
		if _, _, err = c.ObserveCommit(context.Background(), pending); err != C12PITRIndeterminate {
			t.Fatal("partial drain", err)
		}
		before := d.fixture.snapshot()
		calls := c.state.dockerCalls.Load()
		phase := c.state.phase.Load()
		drainCount := *drains
		walSequence := c.state.wal.sequence
		if _, err = c.SelectRecoveryCut(backup, ah); err != C12PITRIndeterminate {
			t.Fatal(err)
		}
		if _, err = c.CrashPrimaryAtCut(context.Background(), sel, bh); err != C12PITRIndeterminate {
			t.Fatal(err)
		}
		if _, facts, e := c.ObserveCommit(context.Background(), a); e != nil || facts != af {
			t.Fatal("A diagnostics", e)
		}
		if _, facts, e := c.ObserveCommit(context.Background(), b); e != nil || facts != bf {
			t.Fatal("B diagnostics", e)
		}
		if !reflect.DeepEqual(before, d.fixture.snapshot()) || c.state.dockerCalls.Load() != calls || c.state.phase.Load() != phase || *drains != drainCount || c.state.wal.sequence != walSequence {
			t.Fatal("poison follow-on had effects")
		}
	})
	t.Run("a_survives_b", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		a, b := f.observed("0/20", 7), f.observed("0/40", 8)
		sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil {
			t.Fatal(err)
		}
		again, err := c.SelectRecoveryCut(c.state.baseBackup, a)
		if err != nil || again != sel {
			t.Fatal("same selection was not idempotent", err)
		}
		before := f.effects()
		if _, err := c.SelectRecoveryCut(c.state.baseBackup, b); err != C12PITRCapacityExceeded || f.effects() != before {
			t.Fatal("second target admitted", err)
		}
		cut, err := c.CrashPrimaryAtCut(context.Background(), sel, b)
		if err != nil || cut.TerminalCommit.EndLSN != "0/40" || cut.RecoveryTargetLSN != "0/20" {
			t.Fatalf("A/B cut=%+v err=%v", cut, err)
		}
		candidate, err := c.RestoreAtCut(context.Background(), cut)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(f.script, "recovery_target_lsn = '0/20'") || !strings.Contains(f.script, "recovery_target_inclusive = 'false'") {
			t.Fatal("unsafe produced recovery configuration", f.script)
		}
		if candidate.binding.owner != c.state || candidate.binding.cut != sel.state || candidate.TargetLSN != "0/20" {
			t.Fatal("candidate lost cut identity")
		}
		promoted, err := c.PromoteCandidate(context.Background(), candidate)
		if err != nil || !promoted.Promoted {
			t.Fatal("promotion", err)
		}
		before = f.effects()
		if _, err := c.PromoteCandidate(context.Background(), candidate); err == nil || f.effects() != before {
			t.Fatal("stale_pre_promote_dto accepted")
		}
		if _, err := c.InspectTimeline(context.Background(), candidate); err == nil || f.effects() != before {
			t.Fatal("stale inspect accepted")
		}
		timeline, err := c.InspectTimeline(context.Background(), promoted)
		if err != nil || timeline.TimelineID != 4 || c.state.candidates[0].systemID != 12345 || c.state.candidates[0].timeline != 4 {
			t.Fatal("physical identity not retained", timeline, err)
		}
		before = f.effects()
		if _, err := c.InspectTimeline(context.Background(), promoted); err == nil || f.effects() != before {
			t.Fatal("inspection was not one-shot")
		}
		if err := cleanupC12AuthorityPITR(context.Background(), c); err != nil {
			t.Fatal("retained candidate cleanup", err)
		}
	})
	t.Run("legacy_config", func(t *testing.T) {
		c, f := newC12CutFixture(t)
		c.state.terminalCommit = C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitPrepared, SQLXID: 9, EndLSN: "0/20", GID: "retained-gid"}
		cut, err := c.CrashPrimary(context.Background(), c.state.baseBackup, c.state.terminalCommit)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = c.RestoreAtCut(context.Background(), cut); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(f.script, "recovery_target_inclusive = 'false'") || !strings.Contains(f.script, "recovery_target_lsn = '0/20'") {
			t.Fatal("legacy config does not preserve terminal-end semantics", f.script)
		}
	})
	for _, name := range []string{"foreign_backup", "foreign_observed", "active_callback_blocks_crash", "poison_blocks_crash", "boundary_before_target", "forged_selection"} {
		t.Run(name, func(t *testing.T) {
			c, f := newC12CutFixture(t)
			a, b := f.observed("0/20", 7), f.observed("0/40", 8)
			sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
			if err != nil {
				t.Fatal(err)
			}
			before := f.effects()
			switch name {
			case "foreign_backup":
				backup := c.state.baseBackup
				backup.binding = &c12BackupBinding{}
				_, err = c.SelectRecoveryCut(backup, a)
			case "foreign_observed":
				forged := *a.state
				a.state = &forged
				_, err = c.SelectRecoveryCut(c.state.baseBackup, a)
			case "active_callback_blocks_crash":
				accessErr := c.WithPrimaryAuthorityAccess(context.Background(), func(C12AuthorityAccess) error { _, err = c.CrashPrimaryAtCut(context.Background(), sel, b); return nil })
				if accessErr != nil {
					t.Fatal("primary callback fixture", accessErr)
				}
			case "poison_blocks_crash":
				c.state.observationPoison = true
				_, err = c.CrashPrimaryAtCut(context.Background(), sel, b)
				if err != C12PITRIndeterminate {
					t.Fatal(err)
				}
				if _, e := c.SelectRecoveryCut(c.state.baseBackup, a); e != C12PITRIndeterminate {
					t.Fatal(e)
				}
				c.state.observationPoison = false
			case "boundary_before_target":
				_, err = c.CrashPrimaryAtCut(context.Background(), sel, f.observed("0/10", 9))
			case "forged_selection":
				copy := *sel.state
				_, err = c.CrashPrimaryAtCut(context.Background(), C12AuthorityPITRCutSelection{state: &copy}, b)
			}
			if err == nil || f.effects() != before || c.state.phase.Load() != 2 {
				t.Fatalf("invalid request changed state: %v %s/%s phase=%d", err, before, f.effects(), c.state.phase.Load())
			}
			if _, err = c.CrashPrimaryAtCut(context.Background(), sel, b); err != nil {
				t.Fatal("valid retry", err)
			}
		})
	}
	for _, name := range []string{"forged_cut", "altered_candidate_target", "altered_candidate_name", "altered_candidate_data", "candidate_capacity_8", "active_callback_lifecycle", "physical_system_mismatch", "physical_timeline_not_advanced"} {
		t.Run(name, func(t *testing.T) {
			c, f := newC12CutFixture(t)
			a := f.observed("0/20", 7)
			sel, err := c.SelectRecoveryCut(c.state.baseBackup, a)
			if err != nil {
				t.Fatal(err)
			}
			cut, err := c.CrashPrimaryAtCut(context.Background(), sel, a)
			if err != nil {
				t.Fatal(err)
			}
			if name == "forged_cut" {
				before := f.effects()
				forged := cut
				forged.binding = &c12CutSelectionState{}
				if _, err = c.RestoreAtCut(context.Background(), forged); err == nil || f.effects() != before || c.state.nextCandidate.Load() != 0 {
					t.Fatal("forged cut changed state", err)
				}
			}
			candidate, err := c.RestoreAtCut(context.Background(), cut)
			if err != nil {
				t.Fatal(err)
			}
			before := f.effects()
			switch name {
			case "altered_candidate_target", "altered_candidate_name", "altered_candidate_data":
				bad := candidate
				if name == "altered_candidate_target" {
					bad.TargetLSN = "0/40"
				} else if name == "altered_candidate_name" {
					bad.Name = "wrong"
				} else {
					bad.DataName = "wrong"
				}
				if _, err = c.PromoteCandidate(context.Background(), bad); err == nil || f.effects() != before || c.state.candidatePhase[0].Load() != 3 {
					t.Fatal("forged candidate changed state", err)
				}
			case "candidate_capacity_8":
				for i := 1; i < 8; i++ {
					if _, err = c.RestoreAtCut(context.Background(), cut); err != nil {
						t.Fatal(err)
					}
				}
				before = f.effects()
				if _, err = c.RestoreAtCut(context.Background(), cut); err != C12PITRCapacityExceeded || f.effects() != before || c.state.nextCandidate.Load() != 8 {
					t.Fatal("capacity consumed ninth slot", err)
				}
			case "active_callback_lifecycle":
				c.state.accessActive = true
				if _, err = c.RestoreAtCut(context.Background(), cut); err == nil {
					t.Fatal("restore during callback")
				}
				if _, err = c.PromoteCandidate(context.Background(), candidate); err == nil {
					t.Fatal("promote during callback")
				}
				if _, err = c.CreateBaseBackup(context.Background()); err == nil {
					t.Fatal("backup during callback")
				}
				c.state.accessActive = false
				if f.effects() != before || c.state.candidatePhase[0].Load() != 3 {
					t.Fatal("callback rejection changed state")
				}
			case "physical_system_mismatch":
				f.system = "54321"
			case "physical_timeline_not_advanced":
				f.promotedTimeline = 3
			}
			promoted, err := c.PromoteCandidate(context.Background(), candidate)
			if strings.HasPrefix(name, "physical_") {
				if err == nil {
					t.Fatal("invalid physical identity admitted", promoted)
				}
				return
			}
			if err != nil {
				t.Fatal("valid candidate retry", err)
			}
			for _, mutate := range []func(*C12AuthorityPITRCandidate){func(x *C12AuthorityPITRCandidate) { x.Name = "wrong" }, func(x *C12AuthorityPITRCandidate) { x.TargetLSN = "0/40" }, func(x *C12AuthorityPITRCandidate) { x.DataName = "wrong" }} {
				bad := promoted
				mutate(&bad)
				before = f.effects()
				if _, err = c.InspectTimeline(context.Background(), bad); err == nil || f.effects() != before || c.state.candidatePhase[0].Load() != 4 {
					t.Fatal("forged promoted DTO admitted", err)
				}
			}
		})
	}
	t.Run("order", func(t *testing.T) {
		for _, tc := range []struct {
			backup, target, boundary string
			ok                       bool
		}{
			{"0/10", "0/20", "0/40", true}, {"0/10", "0/20", "0/20", true},
			{"0/20", "0/10", "0/40", false}, {"0/10", "0/40", "0/20", false},
			{"0/F", "0/10", "1/0", true}, {"0/10", "garbage", "0/40", false},
			{"0/0", "100000000/0", "FFFFFFFF/FFFFFFFF", false},
			{"0/0", "0/100000000", "FFFFFFFF/FFFFFFFF", false},
			{"0/0", "00/1", "0/40", false}, {"0/0", "0/01", "0/40", false},
			{"0/0", "0/f", "0/40", false},
		} {
			if err := c12ValidateRecoveryOrder(tc.backup, tc.target, tc.boundary); (err == nil) != tc.ok {
				t.Errorf("order %v: %v", tc, err)
			}
		}
	})
	t.Run("failed_validation_preserves_phase", func(t *testing.T) {
		s := &c12AuthorityPITRState{runGeneration: 1}
		s.phase.Store(2)
		s.baseBackup = C12AuthorityPITRBaseBackup{BackupID: "retained", EndLSN: "0/10"}
		s.terminalCommit = C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitImmediate, SQLXID: 8, EndLSN: "0/20"}
		_, err := (C12AuthorityPITRController{state: s}).CrashPrimary(context.Background(), C12AuthorityPITRBaseBackup{}, s.terminalCommit)
		if err == nil || s.phase.Load() != 2 || s.dockerCalls.Load() != 0 {
			t.Fatalf("rejected backup burned phase: err=%v phase=%d Docker=%d", err, s.phase.Load(), s.dockerCalls.Load())
		}
	})
}

// Only subprocess/database transport is replaced. All cut admission, resource
// identity verification, recovery configuration, phases and WAL writes are real.
type c12CutFixture struct {
	state            *c12AuthorityPITRState
	queries          int
	script           string
	system           string
	promotedTimeline int64
	recoveryTimeline int64
	promoted         bool
	paused           bool
	afterReplay      func()
	failQuery        string
}

func (f *c12CutFixture) effects() string {
	return fmt.Sprintf("%d/%d/%d", f.queries, f.state.dockerCalls.Load(), f.state.wal.sequence)
}
func (f *c12CutFixture) observed(lsn string, xid uint64) C12AuthorityPITRObservedCommit {
	o := &c12ObservationState{owner: f.state, generation: 1, bound: true}
	r := &c12ObservedCommitState{owner: f.state, generation: 1, observation: o, facts: C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitImmediate, SQLXID: xid, EndLSN: lsn}}
	o.result = r
	for i := range f.state.observations {
		if f.state.observations[i] == nil {
			f.state.observations[i] = o
			break
		}
	}
	return C12AuthorityPITRObservedCommit{state: r}
}
func newC12CutFixture(t *testing.T) (C12AuthorityPITRController, *c12CutFixture) {
	t.Helper()
	s := &c12AuthorityPITRState{runGeneration: 1}
	f := &c12CutFixture{state: s, system: "12345", promotedTimeline: 4, recoveryTimeline: 3, paused: true}
	s.phase.Store(2)
	s.descriptor = c12AuthorityPITRDescriptor{RunSuffix: strings.Repeat("a", 32), NonceDigest: strings.Repeat("b", 64), PrimaryID: strings.Repeat("c", 64), PrimaryName: "primary", BaseBackupName: "backup", ArchiveName: "archive", ImageRef: "postgres:18", ImageID: "sha256:" + strings.Repeat("d", 64), MaxCandidates: 8}
	for i := 0; i < 8; i++ {
		s.descriptor.CandidateNames = append(s.descriptor.CandidateNames, fmt.Sprintf("candidate-%02d", i))
		s.descriptor.CandidateDataNames = append(s.descriptor.CandidateDataNames, fmt.Sprintf("data-%02d", i))
	}
	path := filepath.Join(t.TempDir(), "ownership.jsonl")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	var err error
	s.wal, err = openC12AuthorityPITROWAL(path, [32]byte{}, 8192, 4096)
	if err != nil {
		t.Fatal(err)
	}
	s.baseBackup = C12AuthorityPITRBaseBackup{BackupID: strings.Repeat("e", 64), StartLSN: "0/1", EndLSN: "0/10", binding: &c12BackupBinding{owner: s, generation: 1, systemID: 12345, timeline: 3, containerID: s.descriptor.PrimaryID}}
	s.database = sql.OpenDB(c12CutConnector{fixture: f})
	t.Cleanup(func() { s.database.Close() })
	s.candidateDatabaseOpen = func(int) (*sql.DB, error) { return sql.OpenDB(c12CutConnector{fixture: f, candidate: true}), nil }
	s.dockerExecute = func(_ context.Context, a ...string) ([]byte, error) {
		result := ""
		switch {
		case a[0] == "run":
			f.script = a[len(a)-1]
			result = strings.Repeat("f", 64)
		case a[0] == "volume" && a[1] == "create":
			result = a[len(a)-1]
		case a[0] == "volume" && a[1] == "inspect":
			name := a[len(a)-1]
			result = strings.Join([]string{name, "local", "true", s.descriptor.RunSuffix, "authority-v7-pitr", "candidate-" + strings.TrimPrefix(name, "data-") + "-data", s.descriptor.NonceDigest}, "|")
		case a[0] == "container" && a[1] == "inspect":
			if a[3] == "{{.State.Status}}" {
				result = "exited"
			} else {
				index := s.nextCandidate.Load() - 1
				result = strings.Join([]string{strings.Repeat("f", 64), "/" + s.descriptor.CandidateNames[index], "true", s.descriptor.RunSuffix, "authority-v7-pitr", fmt.Sprintf("candidate-%02d", index), s.descriptor.NonceDigest, s.descriptor.ImageRef, s.descriptor.ImageID}, "|")
			}
		case a[0] == "container" && a[1] == "port":
			result = "127.0.0.1:54321"
		case a[0] == "container" && (a[1] == "exec" || a[1] == "kill" || a[1] == "stop" || a[1] == "rm"):
		case a[0] == "volume" && a[1] == "rm":
		default:
			return nil, fmt.Errorf("unexpected Docker args %v", a)
		}
		return []byte(result), nil
	}
	return C12AuthorityPITRController{state: s}, f
}

type c12CutConnector struct {
	fixture   *c12CutFixture
	candidate bool
}

func (c c12CutConnector) Connect(context.Context) (driver.Conn, error) {
	return c12CutConn{c.fixture, c.candidate}, nil
}
func (c12CutConnector) Driver() driver.Driver { return c12InitDriver{} }

type c12CutConn struct {
	fixture   *c12CutFixture
	candidate bool
}

func (c12CutConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unexpected prepare") }
func (c12CutConn) Close() error                        { return nil }
func (c c12CutConn) Begin() (driver.Tx, error)         { return c, nil }
func (c12CutConn) Commit() error                       { return nil }
func (c12CutConn) Rollback() error                     { return nil }
func (c c12CutConn) ExecContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
	c.fixture.queries++
	if q != c12RequiredObjectsSQL {
		return nil, errors.New("unexpected exec")
	}
	return driver.RowsAffected(0), nil
}
func (c c12CutConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	f := c.fixture
	f.queries++
	if f.failQuery == q {
		if q == `SELECT pg_catalog.pg_promote(true,60)` {
			f.promoted = true
		}
		return nil, errors.New("driver result lost")
	}
	var v []driver.Value
	switch q {
	case `SELECT pg_catalog.pg_current_wal_flush_lsn()::text`:
		v = []driver.Value{"0/10"}
	case `SELECT pg_catalog.pg_walfile_name(pg_catalog.pg_switch_wal())`:
		v = []driver.Value{"000000030000000000000001"}
	case `SELECT pg_catalog.pg_is_in_recovery(), pg_catalog.pg_last_wal_replay_lsn()::text`:
		v = []driver.Value{true, "0/20"}
		if f.afterReplay != nil {
			f.afterReplay()
		}
	case `SELECT pg_catalog.pg_is_in_recovery(), pg_catalog.pg_last_wal_replay_lsn()::text, pg_catalog.pg_is_wal_replay_paused()`:
		v = []driver.Value{true, "0/20", f.paused}
		if f.afterReplay != nil {
			f.afterReplay()
		}
	case `SELECT pg_catalog.pg_promote(true,60)`:
		f.promoted = true
		v = []driver.Value{true}
	case `SELECT pg_catalog.pg_is_in_recovery()`:
		v = []driver.Value{false}
	case c12GetNodeControlDatabaseIdentitySQL:
		if c.candidate && !f.promoted {
			return nil, errors.New("WAL insert functions forbidden during recovery")
		}
		timeline := int64(3)
		if f.promoted {
			timeline = f.promotedTimeline
		}
		v = []driver.Value{f.system, timeline, "0/20"}
	case c12RecoveryIdentitySQL:
		v = []driver.Value{f.system, f.recoveryTimeline}
	case `SELECT timeline_id::bigint, pg_catalog.pg_current_wal_flush_lsn()::text FROM pg_catalog.pg_control_checkpoint()`:
		v = []driver.Value{int64(4), "0/20"}
	default:
		return nil, fmt.Errorf("unexpected query %s", q)
	}
	return &c12CutRows{v: v}, nil
}

type c12CutRows struct {
	v    []driver.Value
	done bool
}

func (r *c12CutRows) Columns() []string {
	v := make([]string, len(r.v))
	for i := range v {
		v[i] = fmt.Sprint(i)
	}
	return v
}
func (r *c12CutRows) Close() error { return nil }
func (r *c12CutRows) Next(v []driver.Value) error {
	if r.done {
		return io.EOF
	}
	copy(v, r.v)
	r.done = true
	return nil
}

type c12PrivilegeBackend struct {
	*c12AccessTestDriver
	values  []any
	err     error
	queries int
	query   string
}

func (b *c12PrivilegeBackend) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	if q != c12CandidatePrivilegesSQL {
		return b.c12AccessTestDriver.QueryRow(ctx, q, args...)
	}
	b.fixture.event("candidate_privileges")
	b.queries++
	b.query = q
	return c12PrivilegeRow{b.values, b.err}
}

type c12PrivilegeRow struct {
	values []any
	err    error
}

func (r c12PrivilegeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return errors.New("scan arity")
	}
	for i := range dest {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(r.values[i]))
	}
	return nil
}
