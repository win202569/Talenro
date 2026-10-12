package serving

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

func desiredRowsFixture(t *testing.T, rotations int) (uuid.UUID, []store.ListExactDesiredProjectionRow, desiredProjectionSnapshot) {
	t.Helper()
	node, _, want := desiredProjectionFixture(t, rotations)
	d := want.Desired
	header := store.ListExactDesiredProjectionRow{
		NodeID: node, PointerGroupsValid: true, HasDesired: true, CurrentValid: true,
		ActiveDesiredGeneration: 7, ActiveRootPublishID: want.ActiveRootPublishID, ActiveRootVersion: int64(want.ActiveRootVersion),
		ActiveMetadataPublishID: want.ActiveMetadataPublishID, ActiveMetadataVersion: 4,
		DesiredNodeID:                  d.NodeID,
		DesiredSigningID:               d.SigningID,
		DesiredAuthorityOperationID:    d.AuthorityOperationID,
		DesiredRootPublishID:           d.RootPublishID,
		DesiredMetadataPublishID:       d.MetadataPublishID,
		DesiredGeneration:              int64(d.Generation),
		DesiredInventoryVersion:        int64(d.InventoryVersion),
		DesiredResourceEnvelopeVersion: int64(d.ResourceEnvelopeVersion),
		DesiredRootVersion:             int64(d.RootVersion),
		DesiredMetadataVersion:         int64(d.MetadataVersion),
		DesiredAuthorityEpoch:          19,
		DesiredAuthoritySequence:       101,
		DesiredResourceEnvelopeDigest:  bytes.Clone(d.ResourceEnvelopeDigest[:]),
		DesiredSigningKeyID:            bytes.Clone(d.SigningKeyID[:]),
		DesiredPayloadDigest:           bytes.Clone(d.PayloadDigest[:]),
		DesiredSignature:               bytes.Clone(d.Signature[:]),
		DesiredPayload:                 bytes.Clone(d.Payload),
		DesiredIssuedAt:                pgtype.Timestamptz{Time: d.IssuedAt, Valid: true},
		DesiredEffectiveDeadline:       pgtype.Timestamptz{Time: d.EffectiveDeadline, Valid: true},
		DesiredValidUntil:              pgtype.Timestamptz{Time: d.ValidUntil, Valid: true},
	}
	pRoot := want.ActiveRoot
	header.RootPublishID = pRoot.PublishID
	header.RootAuthorityOperationID = pRoot.AuthorityOperationID
	header.RootVersion = int64(pRoot.Version)
	header.RootBaseRootVersion = int64(pRoot.BaseRootVersion)
	header.RootBaseMetadataVersion = int64(pRoot.BaseMetadataVersion)
	header.RootAuthorityEpoch = int64(pRoot.Authority.Epoch)
	header.RootAuthoritySequence = int64(pRoot.Authority.Sequence)
	header.RootPayload = bytes.Clone(pRoot.Payload)
	header.RootPublishedEnvelope = bytes.Clone(pRoot.PublishedEnvelope)
	header.RootPayloadDigest = bytes.Clone(pRoot.PayloadDigest[:])
	header.RootEnvelopeDigest = bytes.Clone(pRoot.EnvelopeDigest[:])
	pMetadata := want.ActiveMetadata
	header.MetadataPublishID = pMetadata.PublishID
	header.MetadataAuthorityOperationID = pMetadata.AuthorityOperationID
	header.MetadataVersion = int64(pMetadata.Version)
	header.MetadataBaseRootVersion = int64(pMetadata.BaseRootVersion)
	header.MetadataBaseMetadataVersion = int64(pMetadata.BaseMetadataVersion)
	header.MetadataAuthorityEpoch = int64(pMetadata.Authority.Epoch)
	header.MetadataAuthoritySequence = int64(pMetadata.Authority.Sequence)
	header.MetadataPayload = bytes.Clone(pMetadata.Payload)
	header.MetadataPublishedEnvelope = bytes.Clone(pMetadata.PublishedEnvelope)
	header.MetadataPayloadDigest = bytes.Clone(pMetadata.PayloadDigest[:])
	header.MetadataEnvelopeDigest = bytes.Clone(pMetadata.EnvelopeDigest[:])
	if rotations == 0 {
		return node, []store.ListExactDesiredProjectionRow{header}, want
	}
	rows := make([]store.ListExactDesiredProjectionRow, rotations)
	for i, p := range want.RootRotations {
		r := header
		r.HasRotation, r.RotationValid = true, true
		r.RotationExpectedVersion = int64(i + 2)
		r.RotationPublishID = p.PublishID
		r.RotationAuthorityOperationID = p.AuthorityOperationID
		r.RotationVersion = int64(p.Version)
		r.RotationBaseRootVersion = int64(p.BaseRootVersion)
		r.RotationBaseMetadataVersion = int64(p.BaseMetadataVersion)
		r.RotationAuthorityEpoch = int64(p.Authority.Epoch)
		r.RotationAuthoritySequence = int64(p.Authority.Sequence)
		r.RotationPayload = bytes.Clone(p.Payload)
		r.RotationPublishedEnvelope = bytes.Clone(p.PublishedEnvelope)
		r.RotationPayloadDigest = bytes.Clone(p.PayloadDigest[:])
		r.RotationEnvelopeDigest = bytes.Clone(p.EnvelopeDigest[:])
		rows[i] = r
	}
	return node, rows, want
}
func assertDesiredRowError(t *testing.T, node uuid.UUID, epoch uint64, rows []store.ListExactDesiredProjectionRow, want error) {
	t.Helper()
	got, err := desiredProjectionFromRows(node, epoch, rows)
	if err != want || !reflect.DeepEqual(got, desiredProjection{}) {
		t.Fatalf("error=%v want=%v, nonzero=%v", err, want, !reflect.DeepEqual(got, desiredProjection{}))
	}
}
func TestDesiredRowOwnership(t *testing.T) {
	node, rows, want := desiredRowsFixture(t, 3)
	// Deliberately share the caller's tail and active-root buffers.
	rows[2].RotationPayload = rows[2].RootPayload
	rows[2].RotationPublishedEnvelope = rows[2].RootPublishedEnvelope
	got, err := desiredProjectionFromRows(node, 19, rows)
	if err != nil {
		t.Fatal(err)
	}
	for i := range rows {
		v := reflect.ValueOf(&rows[i]).Elem()
		for field := 0; field < v.NumField(); field++ {
			if v.Field(field).Kind() != reflect.Slice {
				continue
			}
			b := v.Field(field).Bytes()
			if len(b) == 0 {
				continue
			}
			old := b[0]
			b[0] ^= 127
			if !reflect.DeepEqual(got.Snapshot(), want) {
				t.Fatalf("input alias %d/%s", i, v.Type().Field(field).Name)
			}
			b[0] = old // Assert BEFORE restoration or the next shared-buffer mutation.
		}
	}
	snap := got.Snapshot()
	for _, b := range desiredTestBuffers(&snap) {
		old := b[0]
		b[0] ^= 127
		if !reflect.DeepEqual(got.Snapshot(), want) {
			t.Fatal("snapshot alias")
		}
		b[0] = old
	}
	// Owned tail and active root must not alias each other either.
	for _, pair := range []struct{ tail, active, want []byte }{
		{got.snapshot.RootRotations[2].Payload, got.snapshot.ActiveRoot.Payload, want.ActiveRoot.Payload},
		{got.snapshot.RootRotations[2].PublishedEnvelope, got.snapshot.ActiveRoot.PublishedEnvelope, want.ActiveRoot.PublishedEnvelope},
	} {
		old := pair.tail[0]
		pair.tail[0] ^= 127
		if !bytes.Equal(pair.active, pair.want) {
			t.Fatal("owned tail aliases active root")
		}
		pair.tail[0] = old
	}
	_, before, _ := desiredRowsFixture(t, 3)
	// shared-vs-independent backing storage is irrelevant to value equality.
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("conversion mutated input")
	}
	rows[0].DesiredAuthorityEpoch = 20
	before[0].DesiredAuthorityEpoch = 20
	assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("failed conversion mutated input")
	}
}

func TestDesiredRowValid(t *testing.T) {
	for _, n := range []int{0, 1, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			node, rows, want := desiredRowsFixture(t, n)
			got, err := desiredProjectionFromRows(node, 19, rows)
			if err != nil || !reflect.DeepEqual(got.Snapshot(), want) {
				t.Fatalf("projection mismatch: %v", err)
			}
		})
	}
}
func TestDesiredRowReject(t *testing.T) {
	// Break caught: omission of any independent scalar/length/NULL check.
	_, sample, _ := desiredRowsFixture(t, 3)
	typ := reflect.TypeOf(sample[0])
	for field := 0; field < typ.NumField(); field++ {
		name := typ.Field(field).Name
		t.Run(name, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 1)
			v := reflect.ValueOf(&rows[0]).Elem().Field(field)
			switch v.Kind() {
			case reflect.Bool:
				v.SetBool(false)
			case reflect.Int64:
				v.SetInt(-1)
			case reflect.Array:
				v.Set(reflect.ValueOf(uuid.Nil))
			case reflect.Slice:
				v.SetBytes(nil)
			case reflect.Struct:
				v.Set(reflect.ValueOf(pgtype.Timestamptz{}))
			default:
				t.Fatalf("unhandled field %s", name)
			}
			assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
		})
	}
	// Single-row corruption cannot rely on header disagreement to reject.
	for _, field := range []string{"DesiredPayloadDigest", "RootPayloadDigest", "RootEnvelopeDigest", "MetadataPayloadDigest", "MetadataEnvelopeDigest", "RotationPayloadDigest", "RotationEnvelopeDigest", "DesiredResourceEnvelopeDigest", "DesiredSigningKeyID", "DesiredSignature"} {
		sizes := []int{31, 33}
		if field == "DesiredSignature" {
			sizes = []int{63, 65}
		}
		for _, n := range sizes {
			t.Run(fmt.Sprintf("%s/%d", field, n), func(t *testing.T) {
				node, rows, _ := desiredRowsFixture(t, 1)
				v := reflect.ValueOf(&rows[0]).Elem().FieldByName(field)
				b := make([]byte, n)
				copy(b, v.Bytes())
				v.SetBytes(b)
				assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
			})
		}
	}
	for _, field := range []string{"DesiredIssuedAt", "DesiredEffectiveDeadline", "DesiredValidUntil"} {
		for _, mod := range []pgtype.InfinityModifier{pgtype.Infinity, pgtype.NegativeInfinity, pgtype.InfinityModifier(7)} {
			t.Run(fmt.Sprintf("%s/%d", field, mod), func(t *testing.T) {
				node, rows, _ := desiredRowsFixture(t, 1)
				v := reflect.ValueOf(&rows[0]).Elem().FieldByName(field)
				stamp := v.Interface().(pgtype.Timestamptz)
				stamp.InfinityModifier = mod
				v.Set(reflect.ValueOf(stamp))
				assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
			})
		}
	}
	for name, mutate := range map[string]func([]store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow{
		"gap": func(r []store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow {
			return append(r[:1], r[2:]...)
		},
		"reorder": func(r []store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow {
			r[0], r[1] = r[1], r[0]
			return r
		},
		"duplicate": func(r []store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow {
			r[1] = r[0]
			return r
		},
		"wrong tail": func(r []store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow {
			r[2].RotationPublishID = desiredTestID(900)
			return r
		},
		"epoch": func(r []store.ListExactDesiredProjectionRow) []store.ListExactDesiredProjectionRow {
			for i := range r {
				r[i].DesiredAuthorityEpoch = 20
			}
			return r
		},
	} {
		t.Run(name, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 3)
			assertDesiredRowError(t, node, 19, mutate(rows), authority.ErrConflict)
		})
	}
}
func TestDesiredRowBounds(t *testing.T) {
	for _, v := range []int64{0, 34, math.MaxInt64} {
		node, rows, _ := desiredRowsFixture(t, 1)
		rows[0].ActiveRootVersion = v
		assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
	}
	node, rows, _ := desiredRowsFixture(t, 32)
	rows = append(rows, rows[0])
	assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
	for _, epoch := range []uint64{0, math.MaxInt64 + 1} {
		assertDesiredRowError(t, node, epoch, nil, authority.ErrInvalidArgument)
	}
	assertDesiredRowError(t, uuid.Nil, 19, nil, authority.ErrInvalidArgument)
}
func TestDesiredRowAbsent(t *testing.T) {
	node, _, _ := desiredRowsFixture(t, 0)
	assertDesiredRowError(t, node, 19, nil, authority.ErrNotFound)
	r := store.ListExactDesiredProjectionRow{NodeID: node, PointerGroupsValid: true}
	assertDesiredRowError(t, node, 19, []store.ListExactDesiredProjectionRow{r}, authority.ErrNotFound)
	r.ActiveRootPublishID, r.ActiveRootVersion = desiredTestID(1), 1
	r.ActiveMetadataPublishID, r.ActiveMetadataVersion = desiredTestID(200), 4
	assertDesiredRowError(t, node, 19, []store.ListExactDesiredProjectionRow{r}, authority.ErrNotFound)
	assertDesiredRowError(t, node, 19, []store.ListExactDesiredProjectionRow{r, r}, authority.ErrConflict)
	r.PointerGroupsValid = false
	assertDesiredRowError(t, node, 19, []store.ListExactDesiredProjectionRow{r}, authority.ErrConflict)
	r.PointerGroupsValid = true
	r.HasDesired = true
	r.ActiveDesiredGeneration = 7
	assertDesiredRowError(t, node, 19, []store.ListExactDesiredProjectionRow{r}, authority.ErrConflict)
	for _, field := range []string{"HasRotation", "RotationValid", "RotationExpectedVersion", "RotationPublishID", "RotationPayload", "RotationEnvelopeDigest"} {
		t.Run("empty chain/"+field, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 0)
			v := reflect.ValueOf(&rows[0]).Elem().FieldByName(field)
			switch v.Kind() {
			case reflect.Bool:
				v.SetBool(true)
			case reflect.Int64:
				v.SetInt(2)
			case reflect.Array:
				v.Set(reflect.ValueOf(desiredTestID(999)))
			case reflect.Slice:
				v.SetBytes([]byte{1})
			}
			assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
		})
	}
}
func TestDesiredRowHeaderEquality(t *testing.T) {
	_, sample, _ := desiredRowsFixture(t, 2)
	typ := reflect.TypeOf(sample[0])
	for field := 0; field < typ.NumField(); field++ {
		name := typ.Field(field).Name
		if strings.HasPrefix(name, "Rotation") || name == "HasRotation" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			node, rows, _ := desiredRowsFixture(t, 2)
			v := reflect.ValueOf(&rows[1]).Elem().Field(field)
			switch v.Kind() {
			case reflect.Bool:
				v.SetBool(!v.Bool())
			case reflect.Int64:
				v.SetInt(v.Int() + 1)
			case reflect.Array:
				v.Set(reflect.ValueOf(desiredTestID(999)))
			case reflect.Slice:
				b := bytes.Clone(v.Bytes())
				b[0] ^= 127
				v.SetBytes(b)
				digestField := ""
				switch name {
				case "RootPayload":
					digestField = "RootPayloadDigest"
				case "RootPublishedEnvelope":
					digestField = "RootEnvelopeDigest"
				case "MetadataPayload":
					digestField = "MetadataPayloadDigest"
				case "MetadataPublishedEnvelope":
					digestField = "MetadataEnvelopeDigest"
				case "DesiredPayload":
					digestField = "DesiredPayloadDigest"
				}
				if digestField != "" {
					hash := sha256.Sum256(b)
					reflect.ValueOf(&rows[1]).Elem().FieldByName(digestField).SetBytes(hash[:])
				}
			case reflect.Struct:
				stamp := v.Interface().(pgtype.Timestamptz)
				stamp.Time = stamp.Time.Add(time.Nanosecond)
				v.Set(reflect.ValueOf(stamp))
			}
			assertDesiredRowError(t, node, 19, rows, authority.ErrConflict)
		})
	}
}
