//go:build integration

package testinfra

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
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
	rows     *c12MaterializedRows
	mu       sync.Mutex
	consumed bool
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
				row[index] = c12CopyBytes(value)
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
	for index, value := range rows.values[rows.index] {
		if dest[index] == nil {
			continue
		}
		decodedBytes, ok := c12ScanDecodedCost(rows.typeMap, rows.fields[index], value, dest[index])
		if !ok || !c12AddCopyCost(&copyBytes, int64(len(value))) || !c12AddCopyCost(&copyBytes, decodedBytes) {
			if !ok {
				rows.err = C12PITRDependencyFailure
				return rows.err
			}
			rows.err = C12PITRCapacityExceeded
			return rows.err
		}
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
		source := c12CopyBytes(rows.values[rows.index][index])
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
	copyBytes := int64(24 + len(rows.fields)*16)
	for index, field := range rows.fields {
		raw := rows.values[rows.index][index]
		if raw == nil {
			continue
		}
		decodedBytes, ok := c12ValuesDecodedCost(rows.typeMap, field, raw)
		if !ok {
			rows.err = C12PITRDependencyFailure
			return nil, rows.err
		}
		if !c12AddCopyCost(&copyBytes, decodedBytes) {
			rows.err = C12PITRCapacityExceeded
			return nil, rows.err
		}
	}
	if err := rows.lease.takeBytes(copyBytes); err != nil {
		rows.err = err
		return nil, err
	}
	decoded := make([]any, len(rows.fields))
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
			value = c12CopyBytes(raw)
		} else {
			err = C12PITRDependencyFailure
		}
		if err != nil {
			rows.err = C12PITRDependencyFailure
			return nil, rows.err
		}
		value, _, ok := c12CopyDecodedValue(value)
		if !ok {
			rows.err = C12PITRDependencyFailure
			return nil, rows.err
		}
		decoded[index] = value
	}
	return decoded, nil
}

const c12PGTypeLSNOID = 3220

func c12AddCopyCost(total *int64, cost int64) bool {
	if cost < 0 || *total > (64<<20)-cost {
		return false
	}
	*total += cost
	return true
}

func c12ValuesDecodedCost(typeMap *pgtype.Map, field pgconn.FieldDescription, raw []byte) (int64, bool) {
	dataType, known := typeMap.TypeForOID(field.DataTypeOID)
	if !known {
		if field.Format == pgx.TextFormatCode {
			return int64(len(raw)), true
		}
		if field.Format == pgx.BinaryFormatCode {
			return int64(len(raw)) * 2, true
		}
		return 0, false
	}
	if !dataType.Codec.FormatSupported(field.Format) {
		return 0, false
	}
	switch dataType.Codec.(type) {
	case pgtype.BoolCodec:
		return 1, true
	case pgtype.ByteaCodec:
		return int64(len(raw)) * 2, true
	case pgtype.Int2Codec:
		return 2, true
	case pgtype.Int4Codec, pgtype.Float4Codec, pgtype.Uint32Codec:
		return 4, true
	case pgtype.Int8Codec, pgtype.Float8Codec:
		return 8, true
	case pgtype.TextCodec:
		return int64(len(raw)), true
	case pgtype.DateCodec, *pgtype.TimestampCodec, *pgtype.TimestamptzCodec:
		return int64(reflect.TypeFor[time.Time]().Size()), true
	default:
		return 0, false
	}
}

func c12ScanDecodedCost(typeMap *pgtype.Map, field pgconn.FieldDescription, raw []byte, target any) (int64, bool) {
	dataType, known := typeMap.TypeForOID(field.DataTypeOID)
	if !c12ScanTargetSupported(target, field, known) {
		return 0, false
	}
	if !known {
		if field.Format != pgx.TextFormatCode && field.Format != pgx.BinaryFormatCode {
			return 0, false
		}
		return int64(len(raw)), true
	}
	if !dataType.Codec.FormatSupported(field.Format) {
		return 0, false
	}
	rawBytes := int64(len(raw))
	switch dataType.Codec.(type) {
	case pgtype.BoolCodec:
		return 8, true
	case pgtype.ByteaCodec, pgtype.TextCodec:
		return rawBytes, true
	case pgtype.Int2Codec, pgtype.Int4Codec, pgtype.Int8Codec,
		pgtype.Float4Codec, pgtype.Float8Codec, pgtype.Uint32Codec:
		return 32, true
	case pgtype.DateCodec:
		return 32, true
	case *pgtype.TimestampCodec, *pgtype.TimestamptzCodec:
		return 128, true
	case pgtype.UUIDCodec:
		return 128, true
	case pgtype.NumericCodec:
		if rawBytes > ((64<<20)-256)/8 {
			return 0, false
		}
		return rawBytes*8 + 256, true
	case pgtype.TimeCodec:
		return 64, true
	default:
		return 0, false
	}
}

func c12ScanTargetSupported(target any, field pgconn.FieldDescription, known bool) bool {
	switch target.(type) {
	case *bool,
		*int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64,
		*float32, *float64, *string, *[]byte, *time.Time, *time.Duration,
		*uuid.UUID, *uuid.NullUUID,
		*pgtype.Bool, *pgtype.Int2, *pgtype.Int4, *pgtype.Int8,
		*pgtype.Float4, *pgtype.Float8, *pgtype.Text,
		*pgtype.Date, *pgtype.Timestamp, *pgtype.Timestamptz,
		*pgtype.UUID, *pgtype.Numeric, *pgtype.Time:
		return true
	case *any:
		return !known && field.DataTypeOID == c12PGTypeLSNOID && field.Format == pgx.TextFormatCode
	default:
		return false
	}
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
			result[index] = c12CopyBytes(value)
		}
	}
	return result
}

func (*c12MaterializedRows) Conn() *pgx.Conn { return nil }

func (row *c12MaterializedRow) Scan(dest ...any) error {
	if row == nil || row.rows == nil {
		return C12PITRInvalidHandle
	}
	row.mu.Lock()
	defer row.mu.Unlock()
	if err := row.rows.lease.status(); err != nil {
		return err
	}
	if row.rows.transaction != nil {
		if err := row.rows.transaction.status(); err != nil {
			return err
		}
	}
	if row.consumed {
		return C12PITRWrongPhase
	}
	row.consumed = true
	if !row.rows.Next() {
		if err := row.rows.Err(); err != nil {
			return err
		}
		return c12NoRows{}
	}
	err := row.rows.Scan(dest...)
	row.rows.Close()
	return err
}

func c12CopyDecodedValue(value any) (any, int64, bool) {
	switch typed := value.(type) {
	case nil:
		return nil, 0, true
	case []byte:
		return c12CopyBytes(typed), int64(len(typed)), true
	case string:
		return typed, int64(len(typed)), true
	case bool, int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64, time.Time:
		return typed, int64(reflect.TypeOf(typed).Size()), true
	default:
		return nil, 0, false
	}
}

func c12CopyBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	result := make([]byte, len(value))
	copy(result, value)
	return result
}
