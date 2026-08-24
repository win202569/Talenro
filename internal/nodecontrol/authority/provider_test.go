package authority

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func TestDeterministicProviderIsIdempotentAndTerminal(t *testing.T) {
	provider, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	request := ReserveRequest{
		OperationID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("22222222-2222-4222-8222-222222222222")),
	}
	first, err := provider.Reserve(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := provider.Reserve(t.Context(), request)
	if err != nil || first != second {
		t.Fatalf("reserve retry changed result")
	}
	request.ScopeDigest = contracts.Digest{2}
	if _, err := provider.Reserve(t.Context(), request); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed reserve error = %v", err)
	}
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID:  first.OperationID,
		EffectDigest: contracts.Digest{3},
		DBSystemID:   72057594037927937,
		DBTimeline:   1,
		RequiredLSN:  "0/16B6C50",
	})
	if err != nil || receipt.Status != StatusCommitted {
		t.Fatalf("finalize = %v, %v", receipt, err)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: first.OperationID, Reason: AbortSuperseded}); !errors.Is(err, ErrTerminalConflict) {
		t.Fatalf("abort after commit error = %v", err)
	}

	pending, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		Kind:        EffectRecoveryActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("22222222-2222-4222-8222-222222222222")),
	})
	if err != nil {
		t.Fatal(err)
	}
	abortedReservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		Kind:        EffectOperatorTransition,
		ScopeKind:   ScopeNode,
		ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("22222222-2222-4222-8222-222222222222")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: abortedReservation.OperationID, Reason: AbortSuperseded}); err != nil {
		t.Fatal(err)
	}
	head, err := provider.Head(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if head.LatestReservedSequence != abortedReservation.Sequence || head.LatestReservationDigest != abortedReservation.ReservationDigest {
		t.Fatalf("reserved head = %#v, want sequence/digest from %#v", head, abortedReservation)
	}
	if head.LatestCommittedSequence != receipt.Sequence || head.LatestCommittedOperationID != receipt.OperationID ||
		head.LatestCommittedReceiptDigest != receipt.ReceiptDigest || head.LatestCommittedDatabasePoint == nil ||
		receipt.DatabasePoint == nil || *head.LatestCommittedDatabasePoint != *receipt.DatabasePoint {
		t.Fatalf("committed head spliced anchors: %#v, receipt %#v", head, receipt)
	}
	if pending.Sequence >= head.LatestReservedSequence || head.LatestCommittedSequence >= pending.Sequence {
		t.Fatalf("test setup did not leave later pending/aborted rows: head %#v, pending %#v", head, pending)
	}
}

func TestProviderValueEmptyHeadGlobalVectorsAndClosedScopeMatrix(t *testing.T) {
	provider, err := NewDeterministicProvider(11)
	if err != nil {
		t.Fatal(err)
	}
	head, err := provider.Head(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if head != (Head{Epoch: 11}) || head.Validate() != nil {
		t.Fatalf("empty head = %#v, validation %v", head, head.Validate())
	}

	globalNodeTranscript := []byte("TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust")
	globalOperatorTranscript := []byte("TALENRO-GLOBAL-OPERATOR-TRUST-SCOPE-V1\x00global_operator_trust")
	globalNode := sha256.Sum256(globalNodeTranscript)
	globalOperator := sha256.Sum256(globalOperatorTranscript)
	if got := hex.EncodeToString(globalNode[:]); got != "f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1" {
		t.Fatalf("global node digest = %s", got)
	}
	if got := hex.EncodeToString(globalOperator[:]); got != "128f2a7c781f323050ad3ef1541bfad20b7eb32fb2dfae1ca03d59cf671547c8" {
		t.Fatalf("global operator digest = %s", got)
	}

	valid := []struct {
		kind   EffectKind
		scope  ScopeKind
		digest contracts.Digest
	}{
		{EffectTrustBundlePublish, ScopeGlobalNodeTrust, globalNode},
		{EffectTrustBundlePublish, ScopeGlobalOperatorTrust, globalOperator},
		{EffectRootPublish, ScopeGlobalNodeTrust, globalNode},
		{EffectMetadataPublish, ScopeGlobalNodeTrust, globalNode},
		{EffectGrantCreate, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectGrantClaim, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectCertificateActivate, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectCertificateRevoke, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectIdentityEpochAdvance, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectSecurityIncidentOpen, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectSecurityIncidentResolve, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectResourceEnvelopeActivate, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectDesiredActivate, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectRecoveryActivate, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectOperatorTransition, ScopeNode, validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{EffectOperatorAuthorizerChange, ScopeGlobalOperatorTrust, globalOperator},
	}
	for index, test := range valid {
		operationID := uuid.UUID{0x10, byte(index + 1), 0, 0, 0, 0, 0x40, 0, 0x80, 0, 0, 0, 0, 0, 0, byte(index + 1)}
		request := ReserveRequest{OperationID: operationID, Kind: test.kind, ScopeKind: test.scope, ScopeDigest: test.digest}
		if request.Validate() != nil {
			t.Fatalf("valid matrix row %d rejected by value validation: %#v", index, request)
		}
		if _, err := provider.Reserve(t.Context(), request); err != nil {
			t.Fatalf("valid matrix row %d rejected: %v", index, err)
		}
	}

	badGlobal := globalNode
	badGlobal[31] ^= 1
	invalid := []ReserveRequest{
		{OperationID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1"), Kind: EffectRootPublish, ScopeKind: ScopeGlobalNodeTrust, ScopeDigest: badGlobal},
		{OperationID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2"), Kind: EffectRootPublish, ScopeKind: ScopeNode, ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))},
		{OperationID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3"), Kind: EffectDesiredActivate, ScopeKind: ScopeGlobalNodeTrust, ScopeDigest: globalNode},
		{OperationID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4"), Kind: EffectOperatorAuthorizerChange, ScopeKind: ScopeGlobalNodeTrust, ScopeDigest: globalNode},
		{OperationID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5"), Kind: EffectKind("escape_hatch"), ScopeKind: ScopeKind("other"), ScopeDigest: contracts.Digest{1}},
	}
	for index, request := range invalid {
		if _, err := provider.Reserve(t.Context(), request); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("invalid matrix row %d error = %v", index, err)
		}
	}
}

func TestProviderValueTrustBundlePublishAcceptsExactlyBothGlobalScopes(t *testing.T) {
	provider, err := NewDeterministicProvider(12)
	if err != nil {
		t.Fatal(err)
	}
	nodeScope := validNodeScopeDigest(t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))

	valid := []struct {
		name   string
		scope  ScopeKind
		digest contracts.Digest
	}{
		{name: "node-purpose global scope", scope: ScopeGlobalNodeTrust, digest: globalNodeScopeDigest},
		{name: "operator-purpose global scope", scope: ScopeGlobalOperatorTrust, digest: globalOperatorScopeDigest},
	}
	for index, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			request := ReserveRequest{
				OperationID: uuid.UUID{0x12, byte(index + 1), 0, 0, 0, 0, 0x40, 0, 0x80, 0, 0, 0, 0, 0, 0, byte(index + 1)},
				Kind:        EffectTrustBundlePublish, ScopeKind: test.scope, ScopeDigest: test.digest,
			}
			if err := request.Validate(); err != nil {
				t.Fatalf("valid trust-bundle scope rejected by value validation: %v", err)
			}
			if _, err := provider.Reserve(t.Context(), request); err != nil {
				t.Fatalf("valid trust-bundle scope rejected by provider: %v", err)
			}
		})
	}

	invalid := []struct {
		name   string
		scope  ScopeKind
		digest contracts.Digest
	}{
		{name: "node scope", scope: ScopeNode, digest: nodeScope},
		{name: "node global with operator digest", scope: ScopeGlobalNodeTrust, digest: globalOperatorScopeDigest},
		{name: "operator global with node digest", scope: ScopeGlobalOperatorTrust, digest: globalNodeScopeDigest},
		{name: "unknown scope", scope: ScopeKind("global_other_trust"), digest: contracts.Digest{1}},
	}
	for index, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			request := ReserveRequest{
				OperationID: uuid.UUID{0x13, byte(index + 1), 0, 0, 0, 0, 0x40, 0, 0x80, 0, 0, 0, 0, 0, 0, byte(index + 1)},
				Kind:        EffectTrustBundlePublish, ScopeKind: test.scope, ScopeDigest: test.digest,
			}
			if err := request.Validate(); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("invalid trust-bundle scope validation error = %v", err)
			}
			if _, err := provider.Reserve(t.Context(), request); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("invalid trust-bundle scope provider error = %v", err)
			}
		})
	}
}

func TestProviderValueCanonicalReservationAndReceiptDigests(t *testing.T) {
	provider, err := NewDeterministicProvider(math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	nodeID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	operationID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	scopeDigest := validNodeScopeDigest(t, nodeID)
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: operationID,
		Kind:        EffectDesiredActivate,
		ScopeKind:   ScopeNode,
		ScopeDigest: scopeDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	reservationJSON := fmt.Sprintf(
		`{"authority_epoch":"9223372036854775807","authority_sequence":"1","effect_kind":"desired_activate","operation_id":"11111111-1111-4111-8111-111111111111","scope_digest":"%s","scope_kind":"node"}`,
		hex.EncodeToString(scopeDigest[:]),
	)
	wantReservationDigest := sha256.Sum256(append([]byte("TALENRO-CONTROL-PLANE-AUTHORITY-RESERVATION-V1\x00"), []byte(reservationJSON)...))
	if reservation.ReservationDigest != wantReservationDigest || reservation.Validate() != nil {
		t.Fatalf("reservation digest = %x, validation %v; want %x", reservation.ReservationDigest, reservation.Validate(), wantReservationDigest)
	}

	effectDigest := contracts.Digest{3}
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID:  operationID,
		EffectDigest: effectDigest,
		DBSystemID:   72057594037927937,
		DBTimeline:   math.MaxUint32,
		RequiredLSN:  "0000/016B6C50",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.DatabasePoint == nil || receipt.DatabasePoint.RequiredLSN != "0/16B6C50" {
		t.Fatalf("canonical database point = %#v", receipt.DatabasePoint)
	}
	receiptJSON := fmt.Sprintf(
		`{"abort_reason":null,"authority_epoch":"9223372036854775807","authority_sequence":"1","database_point":{"db_system_id":"72057594037927937","db_timeline":"4294967295","required_lsn":"0/16B6C50"},"effect_digest":"%s","effect_kind":"desired_activate","operation_id":"11111111-1111-4111-8111-111111111111","reservation_digest":"%s","scope_digest":"%s","scope_kind":"node","status":"committed"}`,
		hex.EncodeToString(effectDigest[:]), hex.EncodeToString(wantReservationDigest[:]), hex.EncodeToString(scopeDigest[:]),
	)
	wantReceiptDigest := sha256.Sum256(append([]byte("TALENRO-CONTROL-PLANE-AUTHORITY-RECEIPT-V1\x00"), []byte(receiptJSON)...))
	if receipt.ReceiptDigest != wantReceiptDigest || receipt.Validate() != nil {
		t.Fatalf("receipt digest = %x, validation %v; want %x", receipt.ReceiptDigest, receipt.Validate(), wantReceiptDigest)
	}

	firstCopy := cloneReceiptForTest(receipt)
	*receipt.EffectDigest = contracts.Digest{9}
	receipt.DatabasePoint.RequiredLSN = "BAD"
	retry, err := provider.Finalize(t.Context(), FinalizeRequest{
		OperationID: operationID, EffectDigest: effectDigest, DBSystemID: 72057594037927937,
		DBTimeline: math.MaxUint32, RequiredLSN: "0/16B6C50",
	})
	if err != nil || !reflect.DeepEqual(retry, firstCopy) {
		t.Fatalf("provider state escaped through receipt pointers: %#v, %v; want %#v", retry, err, firstCopy)
	}
}

func TestProviderValueAbortedReceiptInteroperabilityGoldens(t *testing.T) {
	const (
		scopeDigestHex       = "78e35ee017c068429af3d13922e99686dc26b3e4282c7c679e4ec045c1cf6d21"
		reservationDigestHex = "3feddebf8af0692e26481c0a7c9de75fc8b33ea8b3caac4020064716e586a9d8"
		unboundJCS           = `{"abort_reason":"superseded","authority_epoch":"9223372036854775807","authority_sequence":"1","database_point":null,"effect_digest":null,"effect_kind":"desired_activate","operation_id":"11111111-1111-4111-8111-111111111111","reservation_digest":"3feddebf8af0692e26481c0a7c9de75fc8b33ea8b3caac4020064716e586a9d8","scope_digest":"78e35ee017c068429af3d13922e99686dc26b3e4282c7c679e4ec045c1cf6d21","scope_kind":"node","status":"aborted"}`
		unboundDigestHex     = "8caee31658c50c1291ae20344b1d3a6c6fbccb0017db2ac77087e0eeccb1f918"
		boundJCS             = `{"abort_reason":"provider_dependency_failed","authority_epoch":"9223372036854775807","authority_sequence":"1","database_point":{"db_system_id":"72057594037927937","db_timeline":"4294967295","required_lsn":"0/16B6C50"},"effect_digest":"0300000000000000000000000000000000000000000000000000000000000000","effect_kind":"desired_activate","operation_id":"11111111-1111-4111-8111-111111111111","reservation_digest":"3feddebf8af0692e26481c0a7c9de75fc8b33ea8b3caac4020064716e586a9d8","scope_digest":"78e35ee017c068429af3d13922e99686dc26b3e4282c7c679e4ec045c1cf6d21","scope_kind":"node","status":"aborted"}`
		boundDigestHex       = "00081b2995c7e375becb40ecb82befce233bac24aa9847f0d1feda6eb2a299cd"
	)
	reservation := Reservation{
		OperationID:       uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Kind:              EffectDesiredActivate,
		ScopeKind:         ScopeNode,
		ScopeDigest:       mustDigestHexForTest(t, scopeDigestHex),
		Epoch:             math.MaxInt64,
		Sequence:          1,
		ReservationDigest: mustDigestHexForTest(t, reservationDigestHex),
	}
	if err := reservation.Validate(); err != nil {
		t.Fatalf("frozen reservation vector is invalid: %v", err)
	}

	tests := []struct {
		name       string
		reason     AbortReason
		effect     *contracts.Digest
		database   *DatabasePoint
		wantJCS    string
		wantDigest string
	}{
		{
			name: "unbound aborted", reason: AbortSuperseded,
			wantJCS: unboundJCS, wantDigest: unboundDigestHex,
		},
		{
			name: "bound aborted", reason: AbortProviderDependencyFailed,
			effect:     &contracts.Digest{3},
			database:   &DatabasePoint{SystemID: 72057594037927937, Timeline: math.MaxUint32, RequiredLSN: "0/16B6C50"},
			wantJCS:    boundJCS,
			wantDigest: boundDigestHex,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason := test.reason
			receipt := Receipt{
				Reservation: reservation, EffectDigest: test.effect, DatabasePoint: test.database,
				Status: StatusAborted, AbortReason: &reason,
			}
			if got := independentAbortedReceiptJCSForTest(receipt); got != test.wantJCS {
				t.Fatalf("aborted receipt JCS = %s\nwant %s", got, test.wantJCS)
			}
			digest, err := receiptDigest(receipt)
			if err != nil {
				t.Fatal(err)
			}
			wantDigest := mustDigestHexForTest(t, test.wantDigest)
			if digest != wantDigest {
				t.Fatalf("aborted receipt digest = %x, want independently frozen %s for JCS %s", digest, test.wantDigest, test.wantJCS)
			}
			receipt.ReceiptDigest = digest
			if err := receipt.Validate(); err != nil {
				t.Fatalf("aborted receipt validation: %v", err)
			}
		})
	}
}

func TestProviderValueValidationRejectsMalformedValues(t *testing.T) {
	if _, err := NewDeterministicProvider(0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("zero epoch error = %v", err)
	}
	if _, err := NewDeterministicProvider(math.MaxInt64 + 1); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("large epoch error = %v", err)
	}
	provider, err := NewDeterministicProvider(1)
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := provider.Reserve(t.Context(), ReserveRequest{
		OperationID: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"), Kind: EffectDesiredActivate,
		ScopeKind: ScopeNode, ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")),
	})
	if err != nil {
		t.Fatal(err)
	}
	badWAL := []WALPosition{"", "0", "0/", "/0", "0/abc", "0/0A-G", "100000000/0", "0/100000000", " 0/1", "0/1\n"}
	for index, wal := range badWAL {
		_, err := provider.Finalize(t.Context(), FinalizeRequest{
			OperationID: reservation.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, DBTimeline: 1, RequiredLSN: wal,
		})
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("bad WAL %d (%q) error = %v", index, wal, err)
		}
	}
	badFinalize := []FinalizeRequest{
		{OperationID: reservation.OperationID, DBSystemID: 1, DBTimeline: 1, RequiredLSN: "0/1"},
		{OperationID: reservation.OperationID, EffectDigest: contracts.Digest{1}, DBTimeline: 1, RequiredLSN: "0/1"},
		{OperationID: reservation.OperationID, EffectDigest: contracts.Digest{1}, DBSystemID: 1, RequiredLSN: "0/1"},
	}
	for index, request := range badFinalize {
		if _, err := provider.Finalize(t.Context(), request); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("bad finalize %d error = %v", index, err)
		}
	}

	badHead := Head{Epoch: 1, LatestReservedSequence: 1}
	if badHead.Validate() == nil {
		t.Fatal("head accepted a reserved sequence without reservation digest")
	}
	badCheckpoint := NodeCheckpoint{AuthorityEpoch: 1, Sequence: 1}
	if badCheckpoint.Validate() == nil {
		t.Fatal("checkpoint accepted a sequence without receipt digest")
	}
	tampered := reservation
	tampered.ReservationDigest[0] ^= 1
	if tampered.Validate() == nil {
		t.Fatal("reservation accepted a provider-selected digest")
	}
	if strings.ToLower(string(WALPosition("A/B"))) == string(WALPosition("A/B")) {
		t.Fatal("test setup must exercise uppercase WAL text")
	}
}

func validNodeScopeDigest(t *testing.T, nodeID uuid.UUID) contracts.Digest {
	t.Helper()
	transcript := append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...)
	return sha256.Sum256(transcript)
}

func cloneReceiptForTest(value Receipt) Receipt {
	clone := value
	if value.EffectDigest != nil {
		digest := *value.EffectDigest
		clone.EffectDigest = &digest
	}
	if value.DatabasePoint != nil {
		point := *value.DatabasePoint
		clone.DatabasePoint = &point
	}
	if value.AbortReason != nil {
		reason := *value.AbortReason
		clone.AbortReason = &reason
	}
	return clone
}

func mustDigestHexForTest(t *testing.T, value string) contracts.Digest {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		t.Fatalf("invalid frozen digest %q: %v", value, err)
	}
	var digest contracts.Digest
	copy(digest[:], decoded)
	return digest
}

func independentAbortedReceiptJCSForTest(value Receipt) string {
	abortReason := "null"
	if value.AbortReason != nil {
		abortReason = fmt.Sprintf("%q", string(*value.AbortReason))
	}
	databasePoint := "null"
	if value.DatabasePoint != nil {
		databasePoint = fmt.Sprintf(
			`{"db_system_id":"%d","db_timeline":"%d","required_lsn":"%s"}`,
			value.DatabasePoint.SystemID, value.DatabasePoint.Timeline, value.DatabasePoint.RequiredLSN,
		)
	}
	effectDigest := "null"
	if value.EffectDigest != nil {
		effectDigest = fmt.Sprintf("%q", hex.EncodeToString(value.EffectDigest[:]))
	}
	return fmt.Sprintf(
		`{"abort_reason":%s,"authority_epoch":"%d","authority_sequence":"%d","database_point":%s,"effect_digest":%s,"effect_kind":"%s","operation_id":"%s","reservation_digest":"%s","scope_digest":"%s","scope_kind":"%s","status":"%s"}`,
		abortReason, value.Epoch, value.Sequence, databasePoint, effectDigest, value.Kind, value.OperationID,
		hex.EncodeToString(value.ReservationDigest[:]), hex.EncodeToString(value.ScopeDigest[:]), value.ScopeKind, value.Status,
	)
}
