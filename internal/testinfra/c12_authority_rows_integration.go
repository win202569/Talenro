//go:build integration

package testinfra

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const c12MaterializedColumnLimit = 65536

type c12MaterializedRows struct {
	lease       *c12AccessLease
	transaction *c12AuthorityTx
	typeMap     *pgtype.Map
	fields      []pgconn.FieldDescription
	values      [][][]byte
	commandTag  pgconn.CommandTag

	mu      sync.Mutex
	index   int
	closed  bool
	err     error
	current bool
}

type c12MaterializedRow struct {
	rows *c12MaterializedRows
	once sync.Once
	err  error
}

func c12MaterializeRows(ctx context.Context, lease *c12AccessLease, driverRows pgx.Rows) (pgx.Rows, error) {
	if ctx == nil || lease == nil || driverRows == nil {
		return nil, C12PITRInvalidHandle
	}
	defer driverRows.Close()
	if err := lease.status(); err != nil {
		return nil, err
	}
	fields := driverRows.FieldDescriptions()
	if len(fields) > c12MaterializedColumnLimit {
		return nil, C12PITRCapacityExceeded
	}
	metadataBytes := int64(len(fields)) * 64
	for index := range fields {
		nameBytes := int64(len(fields[index].Name))
		if nameBytes > 1<<20 || metadataBytes > (64<<20)-nameBytes {
			return nil, C12PITRCapacityExceeded
		}
		metadataBytes += nameBytes
	}
	if err := lease.takeBytes(metadataBytes); err != nil {
		return nil, err
	}
	fieldCopies := append([]pgconn.FieldDescription(nil), fields...)
	materialized := make([][][]byte, 0)
	for driverRows.Next() {
		if err := lease.status(); err != nil {
			return nil, err
		}
		raw := driverRows.RawValues()
		if len(raw) != len(fields) {
			return nil, C12PITRDependencyFailure
		}
		rowBytes := int64(24 + len(raw)*24)
		for _, value := range raw {
			valueBytes := int64(len(value))
			if valueBytes > 1<<20 || rowBytes > (1<<20)-valueBytes {
				return nil, C12PITRCapacityExceeded
			}
			rowBytes += valueBytes
		}
		if err := lease.takeRow(rowBytes); err != nil {
			return nil, err
		}
		row := make([][]byte, len(raw))
		for index, value := range raw {
			if value != nil {
				row[index] = append([]byte(nil), value...)
			}
		}
		materialized = append(materialized, row)
	}
	driverRows.Close()
	if err := driverRows.Err(); err != nil {
		return nil, c12AccessError(err, false)
	}
	if err := ctx.Err(); err != nil {
		return nil, C12PITRCanceled
	}
	return &c12MaterializedRows{
		lease:      lease,
		typeMap:    pgtype.NewMap(),
		fields:     fieldCopies,
		values:     materialized,
		commandTag: driverRows.CommandTag(),
		index:      -1,
	}, nil
}

func (rows *c12MaterializedRows) validLocked() error {
	if rows == nil || rows.lease == nil {
		return C12PITRInvalidHandle
	}
	if err := rows.lease.status(); err != nil {
		return err
	}
	if rows.transaction != nil {
		if err := rows.transaction.status(); err != nil {
			return err
		}
	}
	if rows.closed {
		return C12PITRWrongPhase
	}
	return nil
}

func (rows *c12MaterializedRows) Close() {
	if rows == nil {
		return
	}
	rows.mu.Lock()
	rows.closed = true
	rows.current = false
	rows.mu.Unlock()
}

func (rows *c12MaterializedRows) Err() error {
	if rows == nil {
		return C12PITRInvalidHandle
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if rows.err != nil {
		return rows.err
	}
	if err := rows.lease.status(); err != nil {
		return err
	}
	if rows.transaction != nil {
		if err := rows.transaction.status(); err != nil {
			return err
		}
	}
	return nil
}

func (rows *c12MaterializedRows) CommandTag() pgconn.CommandTag {
	if rows == nil {
		return pgconn.CommandTag{}
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if rows.lease.status() != nil {
		return pgconn.CommandTag{}
	}
	return rows.commandTag
}

func (rows *c12MaterializedRows) FieldDescriptions() []pgconn.FieldDescription {
	if rows == nil {
		return nil
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if err := rows.validLocked(); err != nil {
		rows.err = err
		return nil
	}
	copyBytes := int64(len(rows.fields)) * 64
	for index := range rows.fields {
		copyBytes += int64(len(rows.fields[index].Name))
	}
	if err := rows.lease.takeBytes(copyBytes); err != nil {
		rows.err = err
		return nil
	}
	return append([]pgconn.FieldDescription(nil), rows.fields...)
}

func (rows *c12MaterializedRows) Next() bool {
	if rows == nil {
		return false
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if err := rows.validLocked(); err != nil {
		rows.err = err
		rows.current = false
		return false
	}
	rows.index++
	if rows.index >= len(rows.values) {
		rows.closed = true
		rows.current = false
		return false
	}
	rows.current = true
	return true
}

func (rows *c12MaterializedRows) Scan(dest ...any) error {
	if rows == nil {
		return C12PITRInvalidHandle
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if err := rows.validLocked(); err != nil {
		rows.err = err
		return err
	}
	if !rows.current || rows.index < 0 || rows.index >= len(rows.values) || len(dest) != len(rows.fields) {
		rows.err = C12PITRWrongPhase
		return rows.err
	}
	copyBytes := int64(24 + len(rows.fields)*24)
	for _, value := range rows.values[rows.index] {
		copyBytes += int64(len(value))
	}
	if err := rows.lease.takeBytes(copyBytes); err != nil {
		rows.err = err
		return err
	}
	for index, target := range dest {
		if target == nil {
			continue
		}
		field := rows.fields[index]
		source := append([]byte(nil), rows.values[rows.index][index]...)
		if err := rows.typeMap.Scan(field.DataTypeOID, field.Format, source, target); err != nil {
			rows.err = C12PITRDependencyFailure
			return rows.err
		}
	}
	return nil
}

func (rows *c12MaterializedRows) Values() ([]any, error) {
	if rows == nil {
		return nil, C12PITRInvalidHandle
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if err := rows.validLocked(); err != nil {
		rows.err = err
		return nil, err
	}
	if !rows.current || rows.index < 0 || rows.index >= len(rows.values) {
		rows.err = C12PITRWrongPhase
		return nil, rows.err
	}
	decoded := make([]any, len(rows.fields))
	copyBytes := int64(24 + len(rows.fields)*16)
	for index, field := range rows.fields {
		raw := rows.values[rows.index][index]
		if raw == nil {
			continue
		}
		dataType, known := rows.typeMap.TypeForOID(field.DataTypeOID)
		var value any
		var err error
		if known {
			value, err = dataType.Codec.DecodeValue(rows.typeMap, field.DataTypeOID, field.Format, raw)
		} else if field.Format == pgx.TextFormatCode {
			value = string(raw)
		} else if field.Format == pgx.BinaryFormatCode {
			value = append([]byte(nil), raw...)
		} else {
			err = C12PITRDependencyFailure
		}
		if err != nil {
			rows.err = C12PITRDependencyFailure
			return nil, rows.err
		}
		value, valueBytes, ok := c12CopyDecodedValue(value)
		if !ok || copyBytes > (64<<20)-valueBytes {
			rows.err = C12PITRDependencyFailure
			return nil, rows.err
		}
		copyBytes += valueBytes
		decoded[index] = value
	}
	if err := rows.lease.takeBytes(copyBytes); err != nil {
		rows.err = err
		return nil, err
	}
	return decoded, nil
}

func (rows *c12MaterializedRows) RawValues() [][]byte {
	if rows == nil {
		return nil
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if err := rows.validLocked(); err != nil || !rows.current || rows.index < 0 || rows.index >= len(rows.values) {
		if err != nil {
			rows.err = err
		}
		return nil
	}
	current := rows.values[rows.index]
	copyBytes := int64(24 + len(current)*24)
	for _, value := range current {
		copyBytes += int64(len(value))
	}
	if err := rows.lease.takeBytes(copyBytes); err != nil {
		rows.err = err
		return nil
	}
	result := make([][]byte, len(current))
	for index, value := range current {
		if value != nil {
			result[index] = append([]byte(nil), value...)
		}
	}
	return result
}

func (*c12MaterializedRows) Conn() *pgx.Conn { return nil }

func (row *c12MaterializedRow) Scan(dest ...any) error {
	if row == nil || row.rows == nil {
		return C12PITRInvalidHandle
	}
	row.once.Do(func() {
		if !row.rows.Next() {
			if err := row.rows.Err(); err != nil {
				row.err = err
			} else {
				row.err = c12NoRows{}
			}
			return
		}
		row.err = row.rows.Scan(dest...)
		row.rows.Close()
	})
	if row.err == nil && row.rows.closed && len(dest) != len(row.rows.fields) {
		return C12PITRWrongPhase
	}
	return row.err
}

func c12CopyDecodedValue(value any) (any, int64, bool) {
	switch typed := value.(type) {
	case nil:
		return nil, 0, true
	case []byte:
		return append([]byte(nil), typed...), int64(len(typed)), true
	case string:
		return typed, int64(len(typed)), true
	case bool, int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64, time.Time:
		return typed, int64(reflect.TypeOf(typed).Size()), true
	default:
		return nil, 0, false
	}
}
