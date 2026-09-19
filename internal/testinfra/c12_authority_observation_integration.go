//go:build integration

package testinfra

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type C12AuthorityPITRObservation struct{ state *c12ObservationState }
type C12AuthorityPITRObservedCommit struct{ state *c12ObservedCommitState }

func (controller C12AuthorityPITRController) NewCommitObservation() (C12AuthorityPITRObservation, error) {
	state := controller.state
	if state == nil {
		return C12AuthorityPITRObservation{}, C12PITRInvalidHandle
	}
	state.accessGate.Lock()
	defer state.accessGate.Unlock()
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.observationPoison {
		return C12AuthorityPITRObservation{}, C12PITRIndeterminate
	}
	if state.runGeneration == 0 {
		return C12AuthorityPITRObservation{}, C12PITRInvalidHandle
	}
	if state.transitioning || state.legacyObservation || (state.phase.Load() != 1 && state.phase.Load() != 2) {
		return C12AuthorityPITRObservation{}, C12PITRWrongPhase
	}
	index := -1
	for i, item := range state.observations {
		if item == nil {
			index = i
			break
		}
	}
	if index < 0 {
		return C12AuthorityPITRObservation{}, C12PITRCapacityExceeded
	}
	var random [32]byte
	read := state.observationRandom
	if read == nil {
		read = rand.Read
	}
	if n, err := read(random[:]); err != nil || n != len(random) {
		return C12AuthorityPITRObservation{}, C12PITRDependencyFailure
	}
	marker := hex.EncodeToString(random[:])
	for _, item := range state.observations {
		if item != nil && item.marker == marker {
			return C12AuthorityPITRObservation{}, C12PITRDependencyFailure
		}
	}
	item := &c12ObservationState{owner: state, generation: state.runGeneration, marker: marker}
	state.observations[index] = item
	return C12AuthorityPITRObservation{state: item}, nil
}

// observationRegistered is called under state.mu. Pointer membership, not copied
// facts or marker knowledge, is the capability's authority.
func (state *c12AuthorityPITRState) observationRegistered(item *c12ObservationState) bool {
	if item == nil || item.owner != state || item.generation == 0 || item.generation != state.runGeneration {
		return false
	}
	for _, retained := range state.observations {
		if retained == item {
			return true
		}
	}
	return false
}

func (controller C12AuthorityPITRController) BindCommitObservation(ctx context.Context, observation C12AuthorityPITRObservation, transaction C12AuthorityTransaction) error {
	state := controller.state
	tx, ok := transaction.(*c12AuthorityTx)
	if state == nil || ctx == nil || !ok || tx == nil || tx.lease == nil || tx.lease.owner != state || tx.lease.binding != nil {
		return C12PITRInvalidHandle
	}
	operationContext, cancel, err := tx.lease.operationContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err = tx.acquireOperation(operationContext, c12TxOperation); err != nil {
		return err
	}
	defer tx.releaseOperation(true)
	state.accessGate.Lock()
	state.mu.Lock()
	err = func() error {
		if !state.observationRegistered(observation.state) || tx.runGeneration != state.runGeneration {
			return C12PITRInvalidHandle
		}
		if state.observationPoison {
			return C12PITRIndeterminate
		}
		if state.transitioning || !state.accessActive || (state.phase.Load() != 1 && state.phase.Load() != 2) {
			return C12PITRWrongPhase
		}
		tx.lease.mu.Lock()
		current := tx.lease.activeTx == tx && tx.lease.live.Load() && tx.generation != 0
		tx.lease.mu.Unlock()
		if !current || tx.terminal.Load() != c12TxOperation || operationContext.Err() != nil {
			return C12PITRWrongPhase
		}
		if observation.state.transaction != nil {
			return C12PITRWrongPhase
		}
		for _, item := range state.observations {
			if item != nil && item.transaction == tx {
				return C12PITRWrongPhase
			}
		}
		state.fixtureMu.Lock()
		uncertain := state.writesUncertain
		state.fixtureMu.Unlock()
		if uncertain {
			return C12PITRIndeterminate
		}
		// Reserve once before sending SQL. An error never releases this binding.
		observation.state.transaction = tx
		return nil
	}()
	state.mu.Unlock()
	state.accessGate.Unlock()
	if err != nil {
		return err
	}
	tag, err := tx.driver.Exec(operationContext, c12MarkerInsertSQL, observation.state.marker)
	if err != nil || tag.RowsAffected() != 1 {
		return C12PITRDependencyFailure
	}
	state.mu.Lock()
	observation.state.bound = true
	state.mu.Unlock()
	return nil
}

func (controller C12AuthorityPITRController) ObserveCommit(ctx context.Context, observation C12AuthorityPITRObservation) (C12AuthorityPITRObservedCommit, C12AuthorityPITRCommit, error) {
	empty := C12AuthorityPITRObservedCommit{}
	facts := C12AuthorityPITRCommit{}
	state := controller.state
	if state == nil || ctx == nil {
		return empty, facts, C12PITRInvalidHandle
	}
	// Even cached observations are callback-external; reserve against both new
	// access callbacks and destructive/phase transitions for the whole operation.
	state.accessGate.Lock()
	if state.accessActive || state.transitioning {
		state.accessGate.Unlock()
		return empty, facts, C12PITRWrongPhase
	}
	state.transitioning = true
	state.accessGate.Unlock()
	defer state.endC12Transition()
	state.mu.Lock()
	if !state.observationRegistered(observation.state) {
		state.mu.Unlock()
		return empty, facts, C12PITRInvalidHandle
	}
	if result := observation.state.result; result != nil {
		facts = result.facts
		state.mu.Unlock()
		return C12AuthorityPITRObservedCommit{state: result}, facts, nil
	}
	if state.observationPoison {
		state.mu.Unlock()
		return empty, facts, C12PITRIndeterminate
	}
	if state.phase.Load() != 1 && state.phase.Load() != 2 {
		state.mu.Unlock()
		return empty, facts, C12PITRWrongPhase
	}
	if !observation.state.bound {
		state.mu.Unlock()
		return empty, facts, C12PITRNotObserved
	}
	state.mu.Unlock()
	observationContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := c12DrainObservations(observationContext, state); err != nil {
		return empty, facts, err
	}
	state.mu.Lock()
	result := observation.state.result
	response := state.observationResponse
	if result != nil {
		facts = result.facts
	}
	state.mu.Unlock()
	if result == nil {
		return empty, facts, C12PITRNotObserved
	}
	if response != nil && response() != nil {
		return empty, C12AuthorityPITRCommit{}, C12PITRDependencyFailure
	}
	return C12AuthorityPITRObservedCommit{state: result}, facts, nil
}

const c12ObservationSQL = `SELECT lsn::text, xid::text, data FROM pg_catalog.pg_logical_slot_get_changes($1,NULL,65536,'include-xids','1')`

func c12ReadLegacyObservation(ctx context.Context, state *c12AuthorityPITRState) (decoded []c12AuthorityPITRDecodedRow, result error) {
	if ctx.Err() != nil {
		return nil, C12PITRCanceled
	}
	defer func() {
		if result != nil {
			state.mu.Lock()
			state.observationPoison = true
			state.mu.Unlock()
			result = C12PITRIndeterminate
		}
	}()
	rows, err := state.database.QueryContext(ctx, c12ObservationSQL, state.descriptor.SlotName)
	if err != nil {
		return nil, err
	}
	adapter := &c12SQLObservationRows{rows: rows}
	defer adapter.Close()
	return c12ReadObservedRows(ctx, adapter)
}

// Minimal row surface shared with database/sql legacy probes. No new raw handle
// is exposed to consumers and the prepared parser remains a separate entry.
type c12ObservationRows interface {
	Next() bool
	RawValues() [][]byte
	FieldDescriptions() []pgconn.FieldDescription
	Err() error
	Close()
}
type c12SQLObservationRows struct {
	rows *sql.Rows
	raw  [3]sql.RawBytes
	err  error
}

func (r *c12SQLObservationRows) Next() bool {
	if !r.rows.Next() {
		return false
	}
	r.err = r.rows.Scan(&r.raw[0], &r.raw[1], &r.raw[2])
	return r.err == nil
}
func (r *c12SQLObservationRows) RawValues() [][]byte { return [][]byte{r.raw[0], r.raw[1], r.raw[2]} }
func (r *c12SQLObservationRows) FieldDescriptions() []pgconn.FieldDescription {
	columns, err := r.rows.Columns()
	if err != nil || len(columns) != 3 {
		return nil
	}
	return []pgconn.FieldDescription{{DataTypeOID: pgtype.TextOID}, {DataTypeOID: pgtype.TextOID}, {DataTypeOID: pgtype.TextOID}}
}
func (r *c12SQLObservationRows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Err()
}
func (r *c12SQLObservationRows) Close() {
	if err := r.rows.Close(); r.err == nil {
		r.err = err
	}
}

// The caller holds the controller transition reservation, excluding access and
// legacy slot consumers. A completed response is not proof of an empty slot.
func c12DrainObservations(ctx context.Context, state *c12AuthorityPITRState) (result error) {
	if ctx.Err() != nil {
		return C12PITRCanceled
	}
	state.mu.Lock()
	query := state.observationQuery
	state.mu.Unlock()
	if query == nil {
		if state.databaseURL == nil {
			return C12PITRDependencyFailure
		}
		config, err := pgx.ParseConfig(state.databaseURL.String())
		if err != nil {
			return C12PITRDependencyFailure
		}
		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			return c12AccessError(err, false)
		}
		defer func() {
			closeContext, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = conn.Close(closeContext)
		}()
		query = func(ctx context.Context) (pgx.Rows, error) {
			return conn.Query(ctx, c12ObservationSQL, state.descriptor.SlotName)
		}
	}
	if ctx.Err() != nil {
		return C12PITRCanceled
	}
	// From this point a failure may follow server-side slot advancement, including
	// a Query error. No retry or reopen can recover the missing proof.
	defer func() {
		if result != nil {
			state.mu.Lock()
			state.observationPoison = true
			state.mu.Unlock()
			result = C12PITRIndeterminate
		}
	}()
	rows, err := query(ctx)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return C12PITRIndeterminate
	}
	if rows == nil {
		return C12PITRIndeterminate
	}
	defer rows.Close()
	decoded, err := c12ReadObservedRows(ctx, rows)
	if err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	registered := make(map[string]*c12ObservationState)
	for _, item := range state.observations {
		if item != nil {
			registered[item.marker] = item
		}
	}
	found, err := c12ParseObservedTransactions(decoded, registered)
	if err != nil || ctx.Err() != nil {
		return C12PITRIndeterminate
	}
	// Validation precedes all installations: a later corrupt block cannot leave
	// an earlier proof apparently usable.
	for item, facts := range found {
		item.result = &c12ObservedCommitState{owner: state, generation: state.runGeneration, observation: item, facts: facts}
	}
	return nil
}

func c12ReadObservedRows(ctx context.Context, rows c12ObservationRows) ([]c12AuthorityPITRDecodedRow, error) {
	fields := rows.FieldDescriptions()
	if len(fields) != 3 {
		return nil, C12PITRIndeterminate
	}
	for _, field := range fields {
		if field.DataTypeOID != pgtype.TextOID || field.Format != pgx.TextFormatCode {
			return nil, C12PITRIndeterminate
		}
	}
	budget := c12ReadBudget{}
	decoded := make([]c12AuthorityPITRDecodedRow, 0, 8)
	for rows.Next() {
		if ctx.Err() != nil {
			return nil, C12PITRIndeterminate
		}
		raw := rows.RawValues()
		if len(raw) != 3 {
			return nil, C12PITRIndeterminate
		}
		// Account for the raw payload, its retained copy, decoded-row capacity and
		// parser/map metadata before allocating strings or appending. This allowance
		// also covers amortized slice growth and field/result metadata.
		size := int64(256)
		for _, value := range raw {
			if value == nil || len(value) > 1<<20 {
				return nil, C12PITRCapacityExceeded
			}
			size += 2 * int64(len(value))
		}
		if err := budget.take(size); err != nil {
			return nil, err
		}
		xid, err := strconv.ParseUint(string(raw[1]), 10, 64)
		if err != nil {
			return nil, C12PITRIndeterminate
		}
		decoded = append(decoded, c12AuthorityPITRDecodedRow{LSN: string(raw[0]), XID: xid, Data: string(raw[2])})
	}
	rows.Close()
	if rows.Err() != nil || ctx.Err() != nil {
		return nil, C12PITRIndeterminate
	}
	return decoded, nil
}

// Once a controller chooses a slot consumer family it never switches. This
// preserves the legacy prepared/rollback sequence without consuming ordinary
// observations (including not-yet-bound retained handles).
func c12BeginLegacyObservation(state *c12AuthorityPITRState) error {
	if state == nil {
		return C12PITRInvalidHandle
	}
	state.accessGate.Lock()
	defer state.accessGate.Unlock()
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.observationPoison {
		return C12PITRIndeterminate
	}
	if state.accessActive || state.transitioning || state.phase.Load() != 2 {
		return C12PITRWrongPhase
	}
	for _, item := range state.observations {
		if item != nil {
			return C12PITRWrongPhase
		}
	}
	state.legacyObservation = true
	state.transitioning = true
	return nil
}

type c12ObservationState struct {
	owner       *c12AuthorityPITRState
	generation  uint64
	marker      string
	transaction *c12AuthorityTx
	bound       bool
	result      *c12ObservedCommitState
}

type c12ObservedCommitState struct {
	owner       *c12AuthorityPITRState
	generation  uint64
	observation *c12ObservationState
	facts       C12AuthorityPITRCommit
}

func c12ParseObservedTransactions(rows []c12AuthorityPITRDecodedRow, registered map[string]*c12ObservationState) (map[*c12ObservationState]C12AuthorityPITRCommit, error) {
	results := make(map[*c12ObservationState]C12AuthorityPITRCommit)
	seen := make(map[string]bool)
	var active bool
	var xid, previous uint64
	var observation *c12ObservationState
	for _, row := range rows {
		lsn, valid := c12AuthorityPITRLSNValue(row.LSN)
		if !valid || row.XID == 0 {
			return nil, C12PITRIndeterminate
		}
		if !active {
			if row.Data != "BEGIN "+strconv.FormatUint(row.XID, 10) {
				return nil, C12PITRIndeterminate
			}
			active = true
			xid = row.XID
			previous = lsn
			observation = nil
			continue
		}
		if row.XID != xid || lsn < previous {
			return nil, C12PITRIndeterminate
		}
		previous = lsn
		if row.Data == "COMMIT "+strconv.FormatUint(xid, 10) {
			if observation != nil {
				if _, exists := results[observation]; exists || observation.result != nil {
					return nil, C12PITRIndeterminate
				}
				results[observation] = C12AuthorityPITRCommit{Kind: C12AuthorityPITRCommitImmediate, SQLXID: xid, EndLSN: row.LSN}
			}
			active = false
			continue
		}
		const table = "table public.c12_authority_pitr_markers:"
		const insert = table + " INSERT: marker[text]:'"
		if strings.HasPrefix(row.Data, table) {
			if !strings.HasPrefix(row.Data, insert) || len(row.Data) != len(insert)+65 || row.Data[len(row.Data)-1] != '\'' {
				return nil, C12PITRIndeterminate
			}
			marker := row.Data[len(insert) : len(row.Data)-1]
			if !c12AuthorityPITRDigestPattern.MatchString(marker) || seen[marker] {
				return nil, C12PITRIndeterminate
			}
			seen[marker] = true
			if item := registered[marker]; item != nil {
				if !item.bound || item.marker != marker || observation != nil {
					return nil, C12PITRIndeterminate
				}
				observation = item
			}
		} else if !strings.HasPrefix(row.Data, "table ") || !strings.Contains(row.Data, ": ") {
			// In particular, neither prepared terminals nor stray BEGIN/COMMIT belong here.
			return nil, C12PITRIndeterminate
		}
	}
	if active {
		return nil, C12PITRIndeterminate
	}
	return results, nil
}
