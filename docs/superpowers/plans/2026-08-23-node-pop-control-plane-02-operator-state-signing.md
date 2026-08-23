# Talenro C1.2 Batch 02 Operator State and Signing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付受 PostgreSQL 和 Batch 01 authority fence 约束的 POP/node inventory、operator authorization/cursor/audit、desired/recovery signed-state saga，以及与 online signer 完全隔离的 root-threshold metadata publisher。

**Architecture:** Inventory 与 operator audit 是短事务内的数据库权威；opaque list cursor 只封装稳定 keyset continuation，不携带授权。Desired/recovery workflow 先 reserve authority sequence 与永不复用 generation，再在锁外调用有界 `NodeStateSigner`，最后经 fence receipt 和第二次状态校验激活。Root/metadata publication 使用独立的 threshold-share provider、独立 pending cap 和独立 activation path，online signer 没有 root 能力。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、Ed25519、RFC 8785 JCS、SHA-256、AES-256-GCM、Batch 01 `contracts`/`authority`/`store`、现有 outbox 与 strict JSON helpers。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), approved working-tree SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`.

## Global Constraints

- 本册只完成 suite index 的 `C1.2-B02`；不得实现 certificate issuer、bootstrap/agent/operator TLS listener、node-agent、supervisor 或 core adapter。
- Batch 01 的 `contracts.Digest`、`contracts.AuthorityVersion`、`authority.Provider`、`authority.Coordinator`、schema 名和 OpenAPI wire 名称逐字消费，不得复制或重命名；领域 mutation 不直接构造 provider finalize 坐标。
- PostgreSQL transaction 持有 row/advisory lock 时不得调用 authority provider、`NodeStateSigner`、root-share provider、operator authorizer 或 wall-clock/network dependency。
- Production `NodeStateSigner`、`RootShareProvider` 和 `OperatorAuthorizer` 均是外部 provider；local/test implementation 名称必须显式含 `Deterministic` 或 `LocalTest`。
- `NodeStateSigner` 只签 `desired`、`recovery`、`time_attestation`；root set、metadata 和 root rotation 只能由 `RootMetadataPublisher` 收集 threshold shares。
- C1.2 schema/transcript 不接受 C1.1 root、metadata、client bundle signer、config signer 或 runtime delegation。
- Desired canonical payload 最大 64 KiB、最多 8 slots、最长 24 小时；recovery 最长 15 分钟；time attestation 最长 5 分钟且不 reserve authority sequence。
- Pending/failed/superseded intent、partial root shares、未 finalized fence row 和未推进 active pointer 的 bytes 永远不可被 serving query 返回。
- 所有 operator list 使用唯一稳定键的 bytewise ascending keyset；page size 默认 50、最大 200，cursor 最长 15 分钟，单页 decoded/wire 最大 1 MiB，数据库 statement timeout 最多 2 秒。
- 所有 operator mutation 均要求 exact credential、有限 action/target/reason、`Idempotency-Key`；update 还要求 exact `If-Match`，并在同一事务写 audit/outbox。
- 本计划中每个 checkbox 是一个可独立执行的 2–5 分钟动作；每个 task 必须 RED、GREEN、REFACTOR 后独立 commit。
- 只 stage task 的精确路径；不得 stage、删除或读取用户未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## Frozen file ownership

```text
db/queries/nodecontrol_inventory.sql
db/queries/nodecontrol_state.sql
internal/store/nodecontrol_inventory.sql.go
internal/store/nodecontrol_state.sql.go
internal/nodecontrol/inventory/types.go
internal/nodecontrol/inventory/validation.go
internal/nodecontrol/inventory/repository.go
internal/nodecontrol/inventory/postgres_repository.go
internal/nodecontrol/inventory/service.go
internal/nodecontrol/operator/authorizer.go
internal/nodecontrol/operator/cursor.go
internal/nodecontrol/operator/audit.go
internal/nodecontrol/operator/service.go
internal/nodecontrol/state/contracts.go
internal/nodecontrol/state/canonical.go
internal/nodecontrol/state/provider.go
internal/nodecontrol/state/signer_queue.go
internal/nodecontrol/state/repository.go
internal/nodecontrol/state/postgres_repository.go
internal/nodecontrol/state/service.go
internal/nodecontrol/state/root_publisher.go
```

Batch 03 consumes the active inventory/identity predicates and listener-facing authorizers. Batch 04 consumes signed envelopes and verification vectors. Neither later batch may reinterpret B02 generations, cursor binding, signing kinds, transcript domains, root roles, or terminal states.

### Task 1: Freeze inventory and capacity domain contracts

**Files:**
- Create: `internal/nodecontrol/inventory/types.go`
- Create: `internal/nodecontrol/inventory/validation.go`
- Test: `internal/nodecontrol/inventory/types_test.go`
- Test: `internal/nodecontrol/inventory/validation_test.go`

**Interfaces:**
- Consumes: `contracts.Digest`, `uuid.UUID`, canonical ASCII identifiers, checked unsigned arithmetic.
- Produces: `OperatorState`, `SecurityState`, `Node`, `ProcessSlot`, `CapacityProfile`, `CapacityLimits`, `Mutation`, `func (Node) Validate() error`, `func (CapacityLimits) Validate() error`, `func ValidateMutation(Node, Mutation) error`, and `func ValidateAggregate(Node, []ProcessSlot, map[ProfileRef]CapacityProfile) error`.

```go
type OperatorState string

const (
	OperatorProvisioning OperatorState = "provisioning"
	OperatorEnabled      OperatorState = "enabled"
	OperatorDraining     OperatorState = "draining"
	OperatorDisabled     OperatorState = "disabled"
)

type SecurityState string

const (
	SecurityNormal      SecurityState = "normal"
	SecurityQuarantined SecurityState = "quarantined"
)

type Node struct {
	NodeID                 uuid.UUID
	POPCode                string
	OperatorState          OperatorState
	SecurityState          SecurityState
	ResumeOperatorState    *OperatorState
	PendingTransition      string
	IdentityEpoch          uint64
	LineageID              uuid.UUID
	InventoryVersion       uint64
	ResourceEnvelopeVersion uint64
	ResourceEnvelopeDigest contracts.Digest
}

type CapacityLimits struct {
	EgressLimitBPS              uint64
	ConnectionLimit             uint64
	HandshakeLimitPerSecond     uint64
	CPUQuotaMillicores          uint64
	CPULimitBasisPoints         uint64
	MemoryLimitBytes            uint64
	TaskLimit                   uint64
	FileDescriptorLimit         uint64
	QueueLimit                  uint64
	PacketLossLimitBasisPoints  uint64
}
```

- [ ] **Step 1 (3 min): Write the RED enum and numeric-boundary table test**

```go
func TestCapacityLimitsExactBounds(t *testing.T) {
	valid := CapacityLimits{EgressLimitBPS: 1_000_000, ConnectionLimit: 1, HandshakeLimitPerSecond: 1, CPUQuotaMillicores: 100, CPULimitBasisPoints: 1, MemoryLimitBytes: 67_108_864, TaskLimit: 32, FileDescriptorLimit: 64, QueueLimit: 1, PacketLossLimitBasisPoints: 1}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.TaskLimit = 31
	if !errors.Is(invalid.Validate(), ErrCapacityOutOfRange) {
		t.Fatal("task_limit below 32 was accepted")
	}
}
```

- [ ] **Step 2 (2 min): Run the focused RED test**

Run: `go test ./internal/nodecontrol/inventory -run TestCapacityLimitsExactBounds -count=1`

Expected: FAIL because the inventory package and `CapacityLimits` do not exist.

- [ ] **Step 3 (5 min): Add closed enums and exact field bounds**

Implement the types above and reject unknown enum values, noncanonical node/POP/slot/profile IDs, zero versions, more than 8 slots, duplicate slot IDs, duplicate/unsorted required metrics, unsupported adapter values, and every numeric value outside spec §9.

- [ ] **Step 4 (4 min): Add mutation authorization-independent invariants**

`ValidateMutation` must reject `quarantined+draining`, direct writer changes to security/identity fields, enabling a disabled node, resource-envelope downgrade/fork, and a desired slot whose adapter/profile/required/limits differ from active inventory. It must normalize a saved provisioning resume target to disabled.

- [ ] **Step 5 (3 min): Run the GREEN domain tests**

Run: `go test ./internal/nodecontrol/inventory -run 'TestCapacityLimitsExactBounds|TestValidateMutation|TestValidateAggregate' -count=1`

Expected: PASS with checked-add overflow, ninth-slot, unknown adapter, and profile mismatch cases rejected.

- [ ] **Step 6 (4 min): REFACTOR boundary cases into named tables and add fuzz coverage**

Move every numeric range into one unexported table used by validation and tests; fuzz aggregate sums and assert the result is only success or a finite sentinel error, never panic or wraparound.

Run: `go test ./internal/nodecontrol/inventory -run 'Test|Fuzz' -count=1`

Expected: PASS.

- [ ] **Step 7 (2 min): Commit the inventory contracts**

```bash
git add internal/nodecontrol/inventory/types.go internal/nodecontrol/inventory/validation.go internal/nodecontrol/inventory/types_test.go internal/nodecontrol/inventory/validation_test.go
git commit -m "feat(nodecontrol): freeze inventory contracts"
```

### Task 2: Implement inventory SQL and optimistic repository semantics

**Files:**
- Create: `db/queries/nodecontrol_inventory.sql`
- Create: `internal/nodecontrol/inventory/repository.go`
- Create: `internal/nodecontrol/inventory/postgres_repository.go`
- Create: `internal/nodecontrol/inventory/service.go`
- Create generated: `internal/store/nodecontrol_inventory.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/inventory/postgres_repository_test.go`
- Test: `internal/nodecontrol/inventory/postgres_repository_integration_test.go`

**Interfaces:**
- Consumes: Batch 01 `store.DBTX`, `authority.Repository`, migration `nodecontrol.node_*` tables, task 1 types.
- Produces: `Repository`, `TxRepository`, stable `ListRequest`/`ListPage`, and optimistic create/update/get/list methods.

```go
type TxRepository interface {
	LockNode(context.Context, store.DBTX, uuid.UUID) (Node, error)
	UpdateNode(context.Context, store.DBTX, Node, uint64) (Node, error)
	ReplaceFailureDomains(context.Context, store.DBTX, uuid.UUID, []FailureDomainRef) error
	ReplaceEndpoints(context.Context, store.DBTX, uuid.UUID, []Endpoint) error
	ReplaceProcessSlots(context.Context, store.DBTX, uuid.UUID, []ProcessSlot) error
}

type Repository interface {
	Transact(context.Context, func(context.Context, store.DBTX) error) error
	GetNode(context.Context, uuid.UUID) (Node, error)
	ListNodes(context.Context, ListRequest) (ListPage, error)
	GetCapacityProfile(context.Context, ProfileRef) (CapacityProfile, error)
}

type ListRequest struct {
	POPCode     string
	State       OperatorState
	AfterNodeID uuid.UUID
	HasAfter    bool
	Limit       uint16
}
```

- [ ] **Step 1 (4 min): Write the RED repository integration cases**

Cover stale expected version, immutable referenced capacity profile, endpoint replacement atomicity, 8-slot cap, bytewise UUID keyset ordering, no duplicate across pages, and a query canceled at the 2-second statement timeout.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `go test -tags=integration ./internal/nodecontrol/inventory -run TestPostgresInventoryRepository -count=1 -timeout 3m`

Expected: FAIL because `nodecontrol_inventory.sql` and repository types do not exist.

- [ ] **Step 3 (5 min): Add exact locking, update, and keyset queries**

```sql
-- name: LockNodeInventory :one
SELECT node_id, pop_code, operator_state, security_state, resume_operator_state,
       pending_operator_transition, identity_epoch, lineage_id, inventory_version,
       resource_envelope_version, resource_envelope_digest
FROM nodecontrol.node_inventory
WHERE node_id = sqlc.arg(node_id)
FOR UPDATE;

-- name: UpdateNodeInventoryVersioned :one
UPDATE nodecontrol.node_inventory
SET pop_code = sqlc.arg(pop_code),
    operator_state = sqlc.arg(operator_state),
    inventory_version = inventory_version + 1,
    updated_at = transaction_timestamp()
WHERE node_id = sqlc.arg(node_id)
  AND inventory_version = sqlc.arg(expected_inventory_version)
RETURNING node_id, pop_code, operator_state, security_state, resume_operator_state,
          pending_operator_transition, identity_epoch, lineage_id, inventory_version,
          resource_envelope_version, resource_envelope_digest;

-- name: ListNodeInventoryAfter :many
SELECT node_id, pop_code, operator_state, security_state, inventory_version
FROM nodecontrol.node_inventory
WHERE (NOT sqlc.arg(has_after)::boolean OR node_id > sqlc.arg(after_node_id)::uuid)
  AND (sqlc.arg(pop_code)::text = '' OR pop_code = sqlc.arg(pop_code)::text)
  AND (sqlc.arg(operator_state)::text = '' OR operator_state = sqlc.arg(operator_state)::text)
ORDER BY node_id ASC
LIMIT sqlc.arg(page_limit);
```

The row converter maps nullable `resume_operator_state` to `*OperatorState`: SQL NULL stays nil, and a present value must be one of the four closed states. Tests cover NULL, each legal value and corrupt text; no empty-string sentinel is accepted.

- [ ] **Step 4 (3 min): Generate sqlc artifacts and inspect drift**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS; only `internal/store/nodecontrol_inventory.sql.go`, `models.go`, and `querier.go` change for this task.

- [ ] **Step 5 (5 min): Implement transactions, validation, and exact conflict mapping**

In `postgres_repository.go`, execute `SET LOCAL statement_timeout = '2s'` for list transactions, map zero-row optimistic updates to `ErrVersionConflict`, map unique/check violations to finite inventory sentinels, and reject `Limit` outside 1–201 before SQL. Fetch `Limit=page_size+1`, return at most page size items, and derive `NextAfterNodeID` only from the last returned item.

- [ ] **Step 6 (3 min): Run GREEN repository tests**

Run: `go test -tags=integration ./internal/nodecontrol/inventory -run TestPostgresInventoryRepository -count=1 -timeout 3m`

Expected: PASS; query logs contain no OFFSET and the 201st row is used only to decide whether a next cursor is needed.

- [ ] **Step 7 (4 min): REFACTOR row conversion into total functions**

Make database enum/digest/UUID conversion reject unknown or malformed stored values instead of coercing them. Add unit tests for each corrupt-row sentinel.

Run: `go test ./internal/nodecontrol/inventory ./internal/store -count=1`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit inventory persistence**

```bash
git add db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/inventory/repository.go internal/nodecontrol/inventory/postgres_repository.go internal/nodecontrol/inventory/service.go internal/nodecontrol/inventory/postgres_repository_test.go internal/nodecontrol/inventory/postgres_repository_integration_test.go
git commit -m "feat(nodecontrol): persist versioned inventory"
```

### Task 3: Implement exact operator authorization and bound AEAD cursors

**Files:**
- Create: `internal/nodecontrol/operator/authorizer.go`
- Create: `internal/nodecontrol/operator/cursor.go`
- Test: `internal/nodecontrol/operator/authorizer_test.go`
- Test: `internal/nodecontrol/operator/cursor_test.go`
- Test: `internal/nodecontrol/operator/cursor_fuzz_test.go`

**Interfaces:**
- Consumes: exact operator certificate credential from Batch 03, external policy provider, AES-256-GCM key material, `inventory.ListRequest`.
- Produces: suite-index `OperatorAuthorizer`, closed roles/actions, `OperatorListCursorV1`, `CursorCodec`, and normalized filter/scope digests.

```go
type OperatorAuthorizer interface {
	Authorize(context.Context, Credential, Action, Target) (Authorization, error)
}

type Credential struct {
	CredentialID      string
	IssuerID          string
	SerialBytes       []byte
	LeafDERDigest     contracts.Digest
	PublicKeyDigest   contracts.Digest
	URISAN            string
	OperatorID        uuid.UUID
	AuthorityEpoch    uint64
	CredentialVersion uint64
}

type Target struct {
	NodeID  uuid.UUID
	POPCode string
}

type Authorization struct {
	OperatorID        uuid.UUID
	Role              Role
	ExactPOPScope     []string
	ScopeDigest       contracts.Digest
	CredentialVersion uint64
	ExpiresAt         time.Time
}

type OperatorListCursorV1 struct {
	SchemaVersion           string
	Endpoint                string
	OperatorID              uuid.UUID
	ExactRolePOPScopeDigest contracts.Digest
	NormalizedFilterDigest  contracts.Digest
	LastSortKey             string
	PageSize                uint16
	IssuedAt                time.Time
	ExpiresAt               time.Time
}

type CursorCodec interface {
	Seal(context.Context, OperatorListCursorV1) (string, error)
	Open(context.Context, string) (OperatorListCursorV1, error)
}
```

- [ ] **Step 1 (5 min): Write RED authorization and cursor misuse tests**

Test writer, security-admin, and auditor action matrices; exact POP scope; disabled-node escalation rejection; cursor tamper, wrong endpoint/operator/scope/filter, expired/future-issued cursor, page size 0/201, noncanonical sort key, and key ID mismatch.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodecontrol/operator -run 'TestAuthorize|TestOperatorListCursor' -count=1`

Expected: FAIL because the operator package does not exist.

- [ ] **Step 3 (4 min): Add closed authorization contracts and cache rules**

Allow only roles `inventory_reader`, `inventory_writer`, and `node_security_admin`; define a finite `Action` constant for every OpenAPI operation ID. Cache reader/writer allow decisions only up to `min(provider expiry, now+30s)` and never cache security-admin decisions, denies, dependency errors, certificate expiry, or credential-version mismatch.

- [ ] **Step 4 (5 min): Add deterministic cursor plaintext and AEAD framing**

Use schema value `operator-list-cursor.v1`; JCS-encode the exact struct, encrypt with AES-256-GCM using a fresh 12-byte crypto-random nonce, and bind associated data `TALENRO-OPERATOR-LIST-CURSOR-V1\x00 || key_id`. Encode a binary frame of one-byte key-ID length, key ID, nonce, and ciphertext with raw URL-safe base64. Reject decoded frames over 4096 bytes before allocation growth.

- [ ] **Step 5 (3 min): Bind cursor reopening to fresh authorization**

Add `ValidateCursor(cursor, endpoint, authorization, filterDigest, now)`; require exact endpoint, operator, scope digest, filter digest and page size, `issued_at <= now`, `expires_at > now`, and `expires_at-issued_at <= 15m`. A role or POP scope change invalidates the cursor even if the new scope is wider.

- [ ] **Step 6 (3 min): Run GREEN and fuzz tests**

Run: `go test ./internal/nodecontrol/operator -run 'TestAuthorize|TestOperatorListCursor|FuzzCursor' -count=1`

Expected: PASS; arbitrary bytes never panic, authenticate as valid, or expose plaintext fields in errors.

- [ ] **Step 7 (4 min): REFACTOR public errors to finite sentinels**

Return only `ErrUnauthenticated`, `ErrForbidden`, `ErrCursorInvalid`, `ErrCursorExpired`, or `ErrDependencyUnavailable`; preserve provider details only as structured internal causes with no credential, POP list, nonce, cursor plaintext, or key bytes.

Run: `go test ./internal/nodecontrol/operator -count=1`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit operator authorization and cursors**

```bash
git add internal/nodecontrol/operator/authorizer.go internal/nodecontrol/operator/cursor.go internal/nodecontrol/operator/authorizer_test.go internal/nodecontrol/operator/cursor_test.go internal/nodecontrol/operator/cursor_fuzz_test.go
git commit -m "feat(nodecontrol): bind operator authorization cursors"
```

### Task 4: Make operator mutation audit and outbox atomic

**Files:**
- Modify: `db/queries/nodecontrol_inventory.sql`
- Create: `internal/nodecontrol/operator/audit.go`
- Create: `internal/nodecontrol/operator/service.go`
- Modify generated: `internal/store/nodecontrol_inventory.sql.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/operator/audit_test.go`
- Test: `internal/nodecontrol/operator/service_integration_test.go`

**Interfaces:**
- Consumes: task 2 `inventory.TxRepository`, task 3 `OperatorAuthorizer`, Batch 01 `contracts.NodeInventoryChangedV1`, existing `outbox.Repository`, one caller-supplied `store.DBTX`.
- Produces: immutable `AuditEntry`, `AuditWriter`, `MutationCommand`, and `Service.Execute(context.Context, MutationCommand) (MutationResult, error)` with one database transaction for domain row, audit, and outbox.

```go
type AuditResult string

const (
	AuditSucceeded AuditResult = "succeeded"
	AuditRejected  AuditResult = "rejected"
)

type AuditEntry struct {
	AuditID                uuid.UUID
	OperationID            uuid.UUID
	OperatorID             uuid.UUID
	CredentialDigest       contracts.Digest
	Role                    Role
	Action                  Action
	TargetType              string
	TargetID                string
	ReasonCode              string
	Result                  AuditResult
	Authority               contracts.AuthorityVersion
	BeforeInventoryVersion  uint64
	AfterInventoryVersion   uint64
	OccurredAt              time.Time
}

type AuditWriter interface {
	Append(context.Context, store.DBTX, AuditEntry) error
}
```

- [ ] **Step 1 (4 min): Write RED rollback and idempotency integration tests**

Inject failures after inventory update, after audit insert, and after outbox insert; assert all three tables roll back. Retry the same operation ID and idempotency key; assert one domain version, one audit row, and one event ID.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `go test -tags=integration ./internal/nodecontrol/operator -run TestOperatorMutationAtomicAuditOutbox -count=1 -timeout 3m`

Expected: FAIL because audit queries and `operator.Service` do not exist.

- [ ] **Step 3 (4 min): Add immutable audit and idempotency queries**

```sql
-- name: InsertNodeOperatorAudit :one
INSERT INTO nodecontrol.node_operator_audit (
  audit_id, operation_id, operator_id, credential_digest, role, action,
  target_type, target_id, reason_code, result, authority_epoch,
  authority_sequence, before_inventory_version, after_inventory_version,
  occurred_at
) VALUES (
  sqlc.arg(audit_id), sqlc.arg(operation_id), sqlc.arg(operator_id),
  sqlc.arg(credential_digest), sqlc.arg(role), sqlc.arg(action),
  sqlc.arg(target_type), sqlc.arg(target_id), sqlc.arg(reason_code),
  sqlc.arg(result), sqlc.arg(authority_epoch), sqlc.arg(authority_sequence),
  sqlc.arg(before_inventory_version), sqlc.arg(after_inventory_version),
  sqlc.arg(occurred_at)
)
ON CONFLICT (operation_id) DO UPDATE
SET operation_id = EXCLUDED.operation_id
WHERE nodecontrol.node_operator_audit.credential_digest = EXCLUDED.credential_digest
  AND nodecontrol.node_operator_audit.action = EXCLUDED.action
  AND nodecontrol.node_operator_audit.target_type = EXCLUDED.target_type
  AND nodecontrol.node_operator_audit.target_id = EXCLUDED.target_id
RETURNING audit_id;
```

- [ ] **Step 4 (3 min): Regenerate store artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS; the generated audit method accepts fixed-width digest bytes and returns one audit ID.

- [ ] **Step 5 (5 min): Implement the mutation transaction and safe event payload**

Authorize before opening the transaction, then lock/recheck node scope and `If-Match` inside it. Write the domain row, immutable successful audit, and outbox event before commit. The event includes only event ID, node ID, inventory version, POP code, operator lifecycle, and changed-field enum; omit endpoint address, reason text, credential and signed bytes. Provider-authorized actions receive a finalized `AuthorityVersion`; inventory-only writes store a zero authority pair and never invent a sequence.

- [ ] **Step 6 (3 min): Run GREEN atomicity tests**

Run: `go test -tags=integration ./internal/nodecontrol/operator -run TestOperatorMutationAtomicAuditOutbox -count=1 -timeout 3m`

Expected: PASS at every injected failure and retry point.

- [ ] **Step 7 (4 min): REFACTOR rejection audit into a separate bounded transaction**

Only authenticated, known operator attempts may produce `AuditRejected`; use finite rejection reasons and never log the request body or provider error. A rejection-audit dependency failure does not change the public authorization result and emits a fixed low-cardinality metric.

Run: `go test ./internal/nodecontrol/operator ./internal/nodecontrol/inventory -count=1`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit atomic operator audit**

```bash
git add db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/querier.go internal/nodecontrol/operator/audit.go internal/nodecontrol/operator/service.go internal/nodecontrol/operator/audit_test.go internal/nodecontrol/operator/service_integration_test.go
git commit -m "feat(nodecontrol): commit operator audit with mutations"
```

### Task 5: Freeze C1.2 state, trust metadata, and canonical transcripts

**Files:**
- Create: `internal/nodecontrol/contracts/process_spec.go`
- Test: `internal/nodecontrol/contracts/process_spec_test.go`
- Create: `internal/nodecontrol/state/contracts.go`
- Create: `internal/nodecontrol/state/canonical.go`
- Test: `internal/nodecontrol/state/contracts_test.go`
- Test: `internal/nodecontrol/state/canonical_test.go`
- Create: `testdata/c12/node-state-vectors.json`
- Test: `internal/nodecontrol/state/cross_schema_test.go`

**Interfaces:**
- Consumes: `contracts.AuthorityVersion`, `contracts.Digest`, inventory slot/profile types, RFC 8785 JCS, Ed25519 public keys.
- Produces: the one canonical adapter/process-spec IR, closed signed-state/root/metadata DTOs, exact transcript constructors, semantic validators, transition validators, and deterministic vectors consumed by B03/B04/B05/B07.

```go
type SigningKind string

const (
	SigningDesired         SigningKind = "desired"
	SigningRecovery        SigningKind = "recovery"
	SigningTimeAttestation SigningKind = "time_attestation"
)

// These types live in internal/nodecontrol/contracts/process_spec.go. No
// adapter, agent or supervisor package may redefine or alias them.
type Adapter string

const (
	AdapterFixture Adapter = "fixture"
	AdapterXray    Adapter = "xray"
	AdapterSingBox Adapter = "sing_box"
)

type ProcessLifecycle string

const (
	ProcessRunning  ProcessLifecycle = "running"
	ProcessStopped  ProcessLifecycle = "stopped"
	ProcessDraining ProcessLifecycle = "draining"
)

type DesiredReasonV1 string

const (
	DesiredInitial             DesiredReasonV1 = "initial"
	DesiredOperatorUpdate      DesiredReasonV1 = "operator_update"
	DesiredDrain               DesiredReasonV1 = "drain"
	DesiredResume              DesiredReasonV1 = "resume"
	DesiredLeaseRefresh        DesiredReasonV1 = "lease_refresh"
	DesiredClearSlotQuarantine DesiredReasonV1 = "clear_slot_quarantine"
	DesiredRestoreReauthorize  DesiredReasonV1 = "restore_reauthorize"
)

type Nonce32 [32]byte

type ProcessSpecV1 struct {
	SlotID                 string
	Adapter                Adapter
	ReleaseID              string
	Required               bool
	Lifecycle              ProcessLifecycle
	TestProfile            string
	CapacityProfileID      string
	CapacityProfileVersion uint64
	RequiredMetrics        []string
	CapacityLimits         CapacityLimitsV1
	ProbeProfile           string
	RestartProfile         string
}

type CapacityLimitsV1 struct {
	EgressLimitBPS             uint64
	ConnectionLimit            uint64
	HandshakeLimitPerSecond    uint64
	CPUQuotaMillicores         uint64
	CPULimitBasisPoints        uint64
	MemoryLimitBytes           uint64
	TaskLimit                  uint64
	FileDescriptorLimit        uint64
	QueueLimit                 uint64
	PacketLossLimitBasisPoints uint64
}

const (
	DesiredTranscriptDomain     = "TALENRO-NODE-DESIRED-STATE-V1\x00"
	RecoveryTranscriptDomain    = "TALENRO-NODE-RECOVERY-STATE-V1\x00"
	TimeTranscriptDomain        = "TALENRO-NODE-TIME-ATTESTATION-V1\x00"
	MetadataTranscriptDomain    = "TALENRO-NODE-STATE-METADATA-V1\x00"
	RootRotationTranscriptDomain = "TALENRO-NODE-STATE-ROOT-ROTATION-V1\x00"
)

type NodeDesiredStateV1 struct {
	SchemaVersion              string
	ControlPlaneAuthorityEpoch uint64
	AuthoritySequence          uint64
	NodeID                     uuid.UUID
	Generation                 uint64
	InventoryVersion           uint64
	NodeResourceEnvelopeVersion uint64
	NodeResourceEnvelopeDigest contracts.Digest
	IssuedAt                   time.Time
	ValidUntil                 time.Time
	MinimumAgentVersion        string
	MinimumSupervisorVersion   string
	Reason                     contracts.DesiredReasonV1
	Processes                  []contracts.ProcessSpecV1
}

type NodeTimeAttestationV1 struct {
	SchemaVersion                   string
	ControlPlaneAuthorityEpoch     uint64
	NodeAuthorityCheckpointSequence uint64
	NodeID                          uuid.UUID
	RequestNonce                    contracts.Nonce32
	IssuedAt                        time.Time
	ValidUntil                      time.Time
}

type RecoveryStateSnapshotV1 struct {
	SchemaVersion                         string
	ControlPlaneAuthorityEpoch            uint64
	AuthoritySequence                     uint64
	NodeID                                uuid.UUID
	IdentityEpoch                         uint64
	RecoveryID                            uuid.UUID
	RecoveryReason                        RecoveryReason
	SessionVersion                        uint64
	SortedOpenIncidentIDs                 []uuid.UUID
	SortedAcknowledgedLocalFaultBindings  []LocalFaultBindingV1
	SortedSupervisorFaultBindings         []SupervisorFaultBindingV1
	RemediationEvidenceDigest             *contracts.Digest
	RootVersion                           uint64
	MetadataVersion                       uint64
	RecoveryGeneration                    uint64
	IssuedAt                              time.Time
	ValidUntil                            time.Time
	RequiredAction                        RecoveryAction
	AllSlotsStopped                       bool
}

type RecoveryReason string

const (
	RecoveryIdentityCompromise    RecoveryReason = "identity_compromise"
	RecoveryAdministrativeDisable RecoveryReason = "administrative_disable"
	RecoveryRetire                RecoveryReason = "retire"
	RecoveryAuthorityRestore       RecoveryReason = "authority_restore"
	RecoverySecurityIncident       RecoveryReason = "security_incident"
)

type RecoveryAction string

const (
	RecoveryHoldStopped              RecoveryAction = "hold_stopped"
	RecoverySubmitAttestation        RecoveryAction = "submit_recovery_attestation"
	RecoveryClearSecurityLatches     RecoveryAction = "clear_security_latches"
)

type LocalFaultBindingV1 struct {
	LocalFaultID uuid.UUID
	IncidentID   uuid.UUID
}

type SupervisorFaultBindingV1 struct {
	SupervisorFaultID uuid.UUID
	EvidenceDigest    contracts.Digest
	IncidentID        uuid.UUID
}
```

`Nonce32` is raw nonce material, not a digest: its JSON/JCS form is strict unpadded base64url of exactly 43 characters, decode length exactly 32, with canonical re-encode equality; the all-zero value is invalid.

`ProcessSpecV1` validation applies the exact §11.1 ASCII/length rules, at most eight unique slots, the three closed adapters/lifecycles, sorted unique required metrics, immutable inventory/profile equality, and rejects paths, argv, environment, URLs, raw configuration, arbitrary maps and credentials because those fields do not exist. `NodeStateRootSetV1` contains schema version, authority epoch/sequence, root version, threshold and 1–5 bytewise key-ID-sorted unique Ed25519 root keys. `NodeStateTrustMetadataV1` contains a complete key snapshot of at most 64 entries, a sorted cumulative revoked-ID set of at most 4096 entries, metadata version/window, next refresh, authority/root binding and sorted threshold signatures. `NodeStateRootRotationBodyV1` embeds the next root set and requires exact previous version plus current/new signatures.

- [ ] **Step 1 (5 min): Write RED canonical-vector and cross-schema tests**

Load fixed input/output hex from `node-state-vectors.json`; test desired, recovery, time, metadata and root-rotation transcripts. Exercise all four recovery reasons and all three recovery actions, then feed a valid C1.1 `talenro-trust-metadata/v1` payload to every C1.2 decoder and require `ErrSchemaMismatch` before signature acceptance.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodecontrol/state -run 'TestCanonicalVectors|TestRejectC11Schemas' -count=1`

Expected: FAIL because state contracts and vector file do not exist.

- [ ] **Step 3 (5 min): Add strict DTO validation and transcript builders**

Require exact schema strings, UTC whole-second RFC3339 times, canonical UUIDs, sorted unique sets, exact enum values, nonzero authority/version fields, closed desired reason, and JCS roundtrip equality. Compute SHA-256 only after semantic validation. Return transcript bytes as a new allocation containing the fixed ASCII domain followed by canonical payload bytes.

- [ ] **Step 4 (5 min): Add metadata and root transition validation**

Enforce 7-day metadata maximum, refresh no later than 48 hours before expiry, active-key continuous half-open coverage, `active→retiring→revoked`, immutable key material/window, emergency revoke bound to an exact open `online_signer_equivocation` incident, 30-day tombstone retention, cumulative revoked ledger append-only, and exact current/new threshold for rotation.

- [ ] **Step 5 (4 min): Add desired/recovery/time deadline rules**

Desired effective deadline is `min(issued_at+24h, metadata.valid_until, key.not_after)`; recovery uses 15 minutes; time uses 5 minutes and exact 32-byte nonce. Time attestation uses the exact `node_authority_checkpoint_sequence` field, never a global head or another object's creation sequence. Recovery requires `all_slots_stopped=true`, at most 16 sorted unique incidents, at most 64 sorted unique local bindings, at most 16 sorted unique supervisor bindings, and a non-null remediation digest only for `clear_security_latches` (the other two actions require null). Add boundary tests for 16/17 incidents, 64/65 local bindings and 16/17 supervisor bindings.

- [ ] **Step 6 (3 min): Run GREEN vector and transition tests**

Run: `go test ./internal/nodecontrol/state -run 'TestCanonicalVectors|TestRejectC11Schemas|TestMetadataTransition|TestStateDeadlines' -count=1`

Expected: PASS and every vector digest/signature transcript matches the checked-in lowercase hex value.

- [ ] **Step 7 (4 min): REFACTOR canonicalization through one bounded helper**

The helper accepts a 64 KiB maximum for desired/recovery, bounded metadata collections, and rejects unknown/duplicate JSON fields before allocating nested collections. Add fuzz tests asserting decode-canonical-decode equality and no panic.

Run: `go test ./internal/nodecontrol/state -run 'Test|Fuzz' -count=1`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit state contracts and vectors**

```bash
git add internal/nodecontrol/contracts/process_spec.go internal/nodecontrol/contracts/process_spec_test.go internal/nodecontrol/state/contracts.go internal/nodecontrol/state/canonical.go internal/nodecontrol/state/contracts_test.go internal/nodecontrol/state/canonical_test.go internal/nodecontrol/state/cross_schema_test.go testdata/c12/node-state-vectors.json
git commit -m "feat(nodecontrol): freeze signed state contracts"
```

### Task 6: Add the external signer boundary and bounded priority queues

**Files:**
- Create: `internal/nodecontrol/state/provider.go`
- Create: `internal/nodecontrol/state/signer_queue.go`
- Test: `internal/nodecontrol/state/provider_test.go`
- Test: `internal/nodecontrol/state/signer_queue_test.go`
- Test: `internal/nodecontrol/state/signer_queue_race_test.go`

**Interfaces:**
- Consumes: task 5 `SigningKind`, exact canonical transcript and digest, external online signer/root-share providers.
- Produces: suite-index `NodeStateSigner`, `RootShareProvider`, immutable request/result DTOs, and `NewQueuedNodeStateSigner(NodeStateSigner) (*QueuedNodeStateSigner, error)`.

```go
type NodeStateSigner interface {
	Sign(context.Context, SignRequest) (SignResult, error)
}

type RootShareProvider interface {
	SignShare(context.Context, RootShareRequest) (RootShareResult, error)
}

type SignRequest struct {
	SigningID     uuid.UUID
	Kind          SigningKind
	KeyID         string
	PayloadDigest contracts.Digest
	Transcript    []byte
}

type SignResult struct {
	SigningID     uuid.UUID
	KeyID         string
	PayloadDigest contracts.Digest
	Signature     []byte
}

type RootShareRequest struct {
	PublishID     uuid.UUID
	KeyID         string
	PayloadDigest contracts.Digest
	Transcript    []byte
	SignatureRole SignatureRole
}

type RootShareResult struct {
	PublishID       uuid.UUID
	KeyID           string
	PhysicalKeyID   string
	PayloadDigest   contracts.Digest
	SignatureRole   SignatureRole
	Signature       []byte
}
```

`SignatureRole` contains only `current` and `new`. `NodeStateSigner.Sign` rejects any kind outside the three `SigningKind` values; root share types are absent from its request schema and provider credential scope.

- [ ] **Step 1 (5 min): Write RED queue partition and idempotency tests**

Block provider calls and fill recovery, desired, and time queues to 32/64/128. Assert worker limits 2/4/2, total in-flight 8, recovery capacity cannot be borrowed, time flood cannot block recovery, desired receives bounded weighted progress, changed retry fields return conflict, and canceled queued requests never call the provider.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodecontrol/state -run 'TestQueuedSigner|TestProviderRequestBinding' -count=1`

Expected: FAIL because provider and queue types do not exist.

- [ ] **Step 3 (4 min): Add provider request/result validation**

Require nonzero IDs, closed kind/role, lowercase 64-hex key IDs, nonzero digest, transcript domain matching kind, Ed25519 signature length 64, and exact echo of ID/key/digest/role. Return `ErrProviderConflict` for changed retries and `ErrProviderMalformed` for mismatched results.

- [ ] **Step 4 (5 min): Implement fixed queues and workers**

Use constants `recoveryQueueCap=32`, `desiredQueueCap=64`, `timeQueueCap=128`, `recoveryWorkers=2`, `desiredWorkers=4`, `timeWorkers=2`. Copy transcript bytes on enqueue, select no unbounded goroutine per request, reject full queues with `ErrSignerBusy`, and require enough caller deadline for the configured provider timeout.

- [ ] **Step 5 (3 min): Add weighted desired progress without capacity borrowing**

Within each fixed worker group use FIFO. A dispatcher admission counter may observe queues for metrics but cannot move a request between worker pools. Test a continuous recovery/time load while 64 desired requests all reach a terminal result.

- [ ] **Step 6 (3 min): Run GREEN queue tests**

Run: `go test ./internal/nodecontrol/state -run 'TestQueuedSigner|TestProviderRequestBinding' -count=1`

Expected: PASS with exact peak concurrency and queue rejection counts.

- [ ] **Step 7 (4 min): REFACTOR shutdown and race behavior**

Add idempotent `Close`, wait for workers, reject new work after close, and ensure each accepted request receives exactly one result. Run:

`go test -race ./internal/nodecontrol/state -run 'TestQueuedSigner|TestClose' -count=1`

Expected: PASS with no race report or goroutine leak.

- [ ] **Step 8 (2 min): Commit signer boundaries and queues**

```bash
git add internal/nodecontrol/state/provider.go internal/nodecontrol/state/signer_queue.go internal/nodecontrol/state/provider_test.go internal/nodecontrol/state/signer_queue_test.go internal/nodecontrol/state/signer_queue_race_test.go
git commit -m "feat(nodecontrol): bound node state signing"
```

### Task 7: Persist immutable signing intents and active state pointers

**Files:**
- Create: `db/queries/nodecontrol_state.sql`
- Create: `internal/nodecontrol/state/repository.go`
- Create: `internal/nodecontrol/state/postgres_repository.go`
- Create generated: `internal/store/nodecontrol_state.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/state/postgres_repository_test.go`
- Test: `internal/nodecontrol/state/postgres_repository_integration_test.go`

**Interfaces:**
- Consumes: Batch 01 migration and `authority.Repository`, task 5 canonical payloads, `store.DBTX`.
- Produces: immutable `SigningIntent`, closed `IntentStatus`, `RecoveryIntentBinding`, `LockedState`, and transaction-bound persistence methods that can activate only a committed fence record whose recovery capture still matches.

```go
type IntentStatus string

const (
	IntentPending    IntentStatus = "pending"
	IntentActive     IntentStatus = "active"
	IntentFailed     IntentStatus = "failed"
	IntentSuperseded IntentStatus = "superseded"
)

type StateRepository interface {
	Transact(context.Context, func(context.Context, store.DBTX) error) error
	LockState(context.Context, store.DBTX, uuid.UUID, SigningKind) (LockedState, error)
	ReserveGeneration(context.Context, store.DBTX, uuid.UUID, SigningKind) (uint64, error)
	InsertIntent(context.Context, store.DBTX, SigningIntent) error
	LoadIntent(context.Context, uuid.UUID) (SigningIntent, error)
	StoreVerifiedSignature(context.Context, store.DBTX, uuid.UUID, SignResult) error
	Activate(context.Context, store.DBTX, Activation) (SignedEnvelope, error)
	Terminate(context.Context, store.DBTX, uuid.UUID, IntentStatus, FailureCode) error
	LoadActive(context.Context, uuid.UUID, SigningKind) (SignedEnvelope, error)
	ResolveAuthorityEffect(context.Context, uuid.UUID) (authority.ResolvedEffect, error)
}

type RecoveryIntentBinding struct {
	RecoveryID                  uuid.UUID
	Reason                      RecoveryReason
	SessionVersion              uint64
	SessionStatus               string
	IncidentSetDigest           contracts.Digest
	LocalFaultBindingsDigest    contracts.Digest
	SupervisorFaultBindingsDigest contracts.Digest
	RemediationEvidenceDigest   *contracts.Digest
	RequiredAction              RecoveryAction
}
```

- [ ] **Step 1 (5 min): Write RED generation and visibility integration tests**

Assert desired and recovery allocators are independent, aborted generation 2 is never reused, only one pending intent per node/kind exists, terminal state never returns to pending, active pointer cannot target a pending or uncommitted-fence row, and serving query returns no orphan signature.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `go test -tags=integration ./internal/nodecontrol/state -run TestPostgresStateRepository -count=1 -timeout 3m`

Expected: FAIL because state SQL and repository do not exist.

- [ ] **Step 3 (5 min): Add lock, allocator, and immutable intent queries**

```sql
-- name: LockNodeStatePointers :one
SELECT node_id, inventory_version, identity_epoch, security_state,
       active_desired_generation, next_desired_generation,
       active_recovery_generation, next_recovery_generation,
       active_root_version, active_root_publish_id,
       active_metadata_version, active_metadata_publish_id
FROM nodecontrol.node_inventory
WHERE node_id = sqlc.arg(node_id)
FOR UPDATE;

-- name: ReserveDesiredGeneration :one
UPDATE nodecontrol.node_inventory
SET next_desired_generation = next_desired_generation + 1
WHERE node_id = sqlc.arg(node_id)
RETURNING next_desired_generation - 1 AS reserved_generation;

-- name: InsertNodeStateSigningIntent :exec
INSERT INTO nodecontrol.node_state_signing_intents (
  signing_id, operation_id, node_id, signing_kind, idempotency_key_digest,
  authority_epoch, authority_sequence, base_generation, reserved_generation,
  canonical_payload, payload_digest, root_version, metadata_version,
  expected_key_id, expected_public_key_digest, captured_inventory_version,
  captured_identity_epoch, captured_security_version,
  recovery_id, recovery_reason, recovery_session_version, recovery_session_status,
  recovery_incident_set_digest, recovery_local_bindings_digest,
  recovery_supervisor_bindings_digest, recovery_remediation_digest,
  recovery_required_action, activation_deadline, status
) VALUES (
  sqlc.arg(signing_id), sqlc.arg(operation_id), sqlc.arg(node_id),
  sqlc.arg(signing_kind), sqlc.arg(idempotency_key_digest),
  sqlc.arg(authority_epoch), sqlc.arg(authority_sequence),
  sqlc.arg(base_generation), sqlc.arg(reserved_generation),
  sqlc.arg(canonical_payload), sqlc.arg(payload_digest), sqlc.arg(root_version),
  sqlc.arg(metadata_version), sqlc.arg(expected_key_id),
  sqlc.arg(expected_public_key_digest), sqlc.arg(captured_inventory_version),
  sqlc.arg(captured_identity_epoch), sqlc.arg(captured_security_version),
  sqlc.narg(recovery_id), sqlc.narg(recovery_reason),
  sqlc.narg(recovery_session_version), sqlc.narg(recovery_session_status),
  sqlc.narg(recovery_incident_set_digest), sqlc.narg(recovery_local_bindings_digest),
  sqlc.narg(recovery_supervisor_bindings_digest), sqlc.narg(recovery_remediation_digest),
  sqlc.narg(recovery_required_action), sqlc.arg(activation_deadline), 'pending'
);
```

All recovery columns are null for desired intents and all are non-null (except the conditionally nullable remediation digest) for recovery intents. Add migration/schema checks for that XOR. Recovery activation re-locks the session and recomputes every captured digest; canonical payload equality alone is insufficient.

- [ ] **Step 4 (4 min): Add finalized-only activation and serving queries**

Activation must join `control_plane_authority_fences` on operation/epoch/sequence and require `provider_status='committed' AND visibility_state='active'`. `ResolveAuthorityEffect` is read-only, recognizes only desired/recovery activation and root/metadata publication operations, and returns the immutable kind/scope/effect digest and `prepared|committed|terminal` state used by the Batch 01 composition-root dispatcher; unknown kinds are not guessed. Root/metadata lookups initially return absent until Task 9 writes their immutable intents. Serving must join the active pointer, exact kind/generation, `IntentActive`, and the same committed fence; it never falls back to the highest generation.

- [ ] **Step 5 (3 min): Generate and inspect state store artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS; `nodecontrol_state.sql.go` has distinct desired/recovery allocator and active-load methods.

- [ ] **Step 6 (5 min): Implement exact row conversion and transition checks**

Reject zero versions/generations, digest length other than 32, signature length other than 64, unknown kind/status/failure, mismatched reserved generation, mutable canonical bytes, and any transition other than pending to one terminal state. Map the one-pending partial unique violation to `ErrSigningInProgress`.

- [ ] **Step 7 (3 min): Run GREEN repository tests**

Run: `go test -tags=integration ./internal/nodecontrol/state -run TestPostgresStateRepository -count=1 -timeout 3m`

Expected: PASS with desired generations `1,3` after failed generation 2 and no serving visibility for that failed row.

- [ ] **Step 8 (4 min): REFACTOR serving query tests into a corrupt-row matrix**

Directly insert each invalid pointer/fence/status combination in a rolled-back test transaction and require `ErrStateUnavailable`; do not silently skip to an older envelope.

Run: `go test ./internal/nodecontrol/state ./internal/store -count=1`

Expected: PASS.

- [ ] **Step 9 (2 min): Commit state persistence**

```bash
git add db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/state/repository.go internal/nodecontrol/state/postgres_repository.go internal/nodecontrol/state/postgres_repository_test.go internal/nodecontrol/state/postgres_repository_integration_test.go
git commit -m "feat(nodecontrol): persist signed state intents"
```

### Task 8: Implement the desired and recovery signing saga

**Files:**
- Create: `internal/nodecontrol/state/service.go`
- Test: `internal/nodecontrol/state/service_test.go`
- Test: `internal/nodecontrol/state/service_integration_test.go`
- Test: `internal/nodecontrol/state/service_crash_test.go`

**Interfaces:**
- Consumes: Batch 01 `authority.Coordinator`; `authority.Provider` only for read-only `CommittedNodeCheckpoint`; task 6 `NodeStateSigner`; task 7 `StateRepository`/state effect resolver; inventory versions/capacity; root/metadata active key authorization; operator audit/outbox; and a transaction-bound `RecoveryTransitionGuard` implemented by B03.
- Produces: prepare/drive/recover API keyed by immutable signing ID; desired/recovery terminal results; optional transaction-bound drain/resume/restore transition applied only with exact state activation.

```go
type PrepareRequest struct {
	OperationID              uuid.UUID
	SigningID                uuid.UUID
	IdempotencyKeyDigest     contracts.Digest
	NodeID                   uuid.UUID
	Kind                     SigningKind
	ExpectedActiveGeneration uint64
	ExpectedInventoryVersion uint64
	RequestedValidUntil      time.Time
	Desired                  *DesiredInput
	Recovery                 *RecoveryInput
	OperatorTransition       *OperatorTransition
}

type StateService interface {
	Prepare(context.Context, PrepareRequest) (SigningIntent, error)
	Drive(context.Context, uuid.UUID) (TerminalResult, error)
	Recover(context.Context, uuid.UUID) (TerminalResult, error)
	IssueTimeAttestation(context.Context, NodeTimeAttestationRequest) (SignedEnvelope, error)
}

type NodeTimeAttestationRequest struct {
	SigningID   uuid.UUID
	NodeID      uuid.UUID
	RequestNonce contracts.Nonce32
}

type OperatorTransition struct {
	Kind               TransitionKind
	ExpectedState      inventory.OperatorState
	TargetState        inventory.OperatorState
	RecoveryID         uuid.UUID
	RecoverySessionVersion uint64
	EffectDigest       contracts.Digest
	RestoreApprovalIDs [2]uuid.UUID
}

type RecoveryInput struct {
	RecoveryID                    uuid.UUID
	Reason                        RecoveryReason
	SessionVersion                uint64
	SessionStatus                 string
	IdentityEpoch                 uint64
	SortedOpenIncidentIDs         []uuid.UUID
	SortedLocalFaultBindings      []LocalFaultBindingV1
	SortedSupervisorFaultBindings []SupervisorFaultBindingV1
	RemediationEvidenceDigest     *contracts.Digest
	RequiredAction                RecoveryAction
	AllSlotsStopped               bool
}

type RecoveryTransitionCapture struct {
	Binding            RecoveryIntentBinding
	Transition         *OperatorTransition
	AttestationDigest  contracts.Digest
}

type RecoveryTransitionGuard interface {
	ValidatePrepare(context.Context, store.DBTX, RecoveryInput, *OperatorTransition) (RecoveryTransitionCapture, error)
	ValidateActivation(context.Context, store.DBTX, RecoveryTransitionCapture) error
	ApplyActivation(context.Context, store.DBTX, RecoveryTransitionCapture, SignedEnvelope) error
}
```

`TransitionKind` contains `none`, `drain`, `resume`, and `restore_reauthorize`. B02 implements `drain`; B03 implements the guard and may construct `resume` or `restore_reauthorize` only after locking the exact session, attestation, incident set, host evidence and (for restore) two approval rows. Disable/quarantine is fail-closed identity work owned by B03 and cannot be represented as a signer-only transition. `RecoveryTransitionGuard` is invoked only inside the caller-owned PostgreSQL transaction and cannot call an external provider.

- [ ] **Step 1 (5 min): Write the RED happy path and stale activation tests**

Test exact order `Coordinator.Reserve → prepare transaction → Sign → local verify → Coordinator.Finalize → activation transaction`; the coordinator alone captures/binds the post-commit database point and activates the provider receipt. Use lock-probing fakes to fail if coordinator/provider/signer is invoked while a DB lock is held. Change base generation, inventory version, identity epoch, security version, key status, recovery ID/reason/session version/status, incident set, either fault-binding set, remediation digest, required action, attestation, restore approval/credential scope, or activation deadline before activation and require superseded with no active bytes.

- [ ] **Step 2 (2 min): Run the focused RED service test**

Run: `go test ./internal/nodecontrol/state -run 'TestStateServiceHappyPath|TestStateServiceStaleActivation' -count=1`

Expected: FAIL because `StateService` is not implemented.

- [ ] **Step 3 (5 min): Implement `Prepare` with immutable retry binding**

Call `authority.Coordinator.Reserve` for `EffectDesiredActivate` or `EffectRecoveryActivate` before opening a database transaction. Inside the transaction lock node/pointer, record that exact reservation with the immutable intent, verify `If-Match` and prerequisites, call `RecoveryTransitionGuard.ValidatePrepare` for recovery/resume/restore work, reserve one generation, construct exact canonical payload using the reserved authority version, capture all versions/key material/recovery bindings/deadline, write the intent, and return it without signature or outbox. Same signing/operation/idempotency IDs and same fields return the same intent; any changed field returns conflict.

- [ ] **Step 4 (5 min): Implement signer call and independent result verification**

Outside transactions call the queue with exact `SignRequest`; verify echoed fields, Ed25519 signature, canonical bytes/digest, schema/domain, current root/metadata/key authorization, and composite deadline. Persist verified signature in a short transaction. `IssueTimeAttestation` reads `authority.Provider.CommittedNodeCheckpoint(SHA256(node_id))`, constructs the exact five-minute `NodeTimeAttestationV1`, and signs it without calling `Reserve`; equal checkpoint values therefore do not create a fence row or advance an allocator. Signer timeout/malformed/mismatch terminates an intent as failed and asks `authority.Coordinator.Abort` to terminate an unfinalized reservation only when the immutable effect resolver proves the domain effect did not commit.

- [ ] **Step 5 (5 min): Implement fence finalize and activation recheck**

After the verified immutable state effect transaction commits, call `authority.Coordinator.Finalize(operation_id,effect_digest)` outside all transactions. The coordinator reloads the state effect through the registered resolver, captures the same-primary post-commit database point, binds it, finalizes the provider and activates the exact receipt; this service never supplies system ID, timeline or LSN. In the final domain transaction re-lock all captured facts, require that committed active receipt and an unexpired deadline, invoke `RecoveryTransitionGuard.ValidateActivation`, insert the immutable envelope, advance the exact active pointer, invoke `ApplyActivation` for the optional transition/session/approvals, clear pending transition, append audit/outbox, and mark intent active. Any guard mismatch supersedes the intent and leaves the node disabled.

- [ ] **Step 6 (4 min): Implement drain atomicity**

The prepare transaction leaves `operator_state` unchanged, stores `pending_operator_transition='draining'`, and makes derived `accepting_new=false`. Only successful desired activation changes state to draining. Signer failure/abort clears pending without changing the prior state; concurrent disable/quarantine supersedes the intent and retains disabled/quarantined state.

- [ ] **Step 7 (3 min): Run GREEN service tests**

Run: `go test ./internal/nodecontrol/state -run 'TestStateServiceHappyPath|TestStateServiceStaleActivation|TestDrainAtomicity|TestTimeAttestationUsesCommittedNodeCheckpoint' -count=1`

Expected: PASS and the test lock probe records zero external calls under lock.

- [ ] **Step 8 (5 min): Add deterministic crash recovery at every durable boundary**

Crash after reserve, intent commit, signer response, verified-signature commit, visibility commit, provider finalize, and activation commit. Repeat with an incident added, a recovery session completed, an approval credential revoked/expired/scope-shrunk, and an effect digest changed at each boundary. `Recover(signingID)` must inspect durable state/provider receipt and either resume the same exact operation or supersede it; it cannot allocate another generation, request changed bytes or retain stale approvals.

Run: `go test ./internal/nodecontrol/state -run TestStateServiceCrashRecovery -count=1`

Expected: PASS for every crash point and response-loss retry.

- [ ] **Step 9 (4 min): REFACTOR failures into a closed matrix**

Use `validation_failed`, `signer_unavailable`, `signer_malformed`, `authorization_changed`, `deadline_expired`, and `superseded`; map public callers to conflict or dependency unavailable without raw SQL/provider/key details.

Run: `go test -tags=integration ./internal/nodecontrol/state ./internal/nodecontrol/operator -count=1 -timeout 5m`

Expected: PASS.

- [ ] **Step 10 (2 min): Commit the signed-state saga**

```bash
git add internal/nodecontrol/state/service.go internal/nodecontrol/state/service_test.go internal/nodecontrol/state/service_integration_test.go internal/nodecontrol/state/service_crash_test.go
git commit -m "feat(nodecontrol): activate signed node state safely"
```

### Task 9: Implement the isolated threshold root and metadata publisher

**Files:**
- Modify: `db/queries/nodecontrol_state.sql`
- Create: `internal/nodecontrol/state/root_publisher.go`
- Modify generated: `internal/store/nodecontrol_state.sql.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/state/root_publisher_test.go`
- Test: `internal/nodecontrol/state/root_publisher_integration_test.go`
- Test: `internal/nodecontrol/state/root_publisher_crash_test.go`

**Interfaces:**
- Consumes: task 5 root/metadata validators, task 6 `RootShareProvider`, Batch 01 `authority.Coordinator`, task 7 state effect resolver, root key registry, open security incidents, operator audit/outbox.
- Produces: `RootMetadataPublisher` with immutable publish IDs, distinct physical-key share persistence, threshold assembly, stale-intent invalidation, and finalized-only active root/metadata pointers.

```go
type PublishKind string

const (
	PublishMetadata     PublishKind = "metadata"
	PublishRootRotation PublishKind = "root_rotation"
)

type PublishRequest struct {
	PublishID             uuid.UUID
	OperationID           uuid.UUID
	IdempotencyKeyDigest  contracts.Digest
	Kind                  PublishKind
	ExpectedRootVersion   uint64
	ExpectedMetadataVersion uint64
	UnsignedBody          []byte
	ActivationDeadline    time.Time
}

type RootMetadataPublisher interface {
	Prepare(context.Context, PublishRequest) (PublishIntent, error)
	Drive(context.Context, uuid.UUID) (PublishedArtifact, error)
	Recover(context.Context, uuid.UUID) (PublishedArtifact, error)
}
```

- [ ] **Step 1 (5 min): Write RED threshold, role, and pending-cap tests**

Test metadata current threshold, root rotation current plus new thresholds, duplicate key ID, duplicate physical key identity, wrong role, unknown/malformed share, stale base pointer, resolved emergency incident, ninth pending publish, online signer credential used as root provider, and a late share after supersede.

- [ ] **Step 2 (2 min): Run the focused RED publisher tests**

Run: `go test ./internal/nodecontrol/state -run 'TestRootMetadataPublisher|TestRootShareIsolation' -count=1`

Expected: FAIL because publisher logic and share queries do not exist.

- [ ] **Step 3 (4 min): Add immutable intent and share SQL**

```sql
-- name: InsertRootMetadataSignatureShare :exec
INSERT INTO nodecontrol.node_root_metadata_signature_shares (
  publish_id, key_id, physical_key_id, payload_digest, signature_role,
  signature, verified_at
) VALUES (
  sqlc.arg(publish_id), sqlc.arg(key_id), sqlc.arg(physical_key_id),
  sqlc.arg(payload_digest), sqlc.arg(signature_role), sqlc.arg(signature),
  sqlc.arg(verified_at)
)
ON CONFLICT (publish_id, key_id, signature_role) DO NOTHING;

-- name: CountNonterminalRootPublishes :one
SELECT count(*)
FROM nodecontrol.node_root_metadata_publish_intents
WHERE status = 'pending';
```

Add a unique constraint/query guard for `(publish_id,physical_key_id,signature_role)` so one physical identity contributes at most one share to a role, even when it exposes multiple key IDs.

- [ ] **Step 4 (3 min): Regenerate state store artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS and generated queries expose fixed role/digest/signature fields.

- [ ] **Step 5 (5 min): Implement prepare and provider collection outside locks**

Call `authority.Coordinator.Reserve` for `EffectMetadataPublish` or `EffectRootPublish`; in the first short transaction record the exact reservation and require fewer than 8 pending intents, exact base pointers, never-reused next version, sorted expected keys/public-key digests, exact current/new threshold, incident/status snapshot and deadline. After commit, call one provider per expected key outside locks using exact `RootShareRequest`.

- [ ] **Step 6 (5 min): Verify shares and assemble the envelope**

Verify each Ed25519 signature over the exact metadata or rotation transcript; persist valid shares immutably, ignore unknown shares for threshold, and fail the intent on malformed/duplicate known-key material. Sort accepted shares by key ID, assemble the exact envelope, then rerun canonical, coverage, cumulative revoke, incident and current/new threshold validation from zero.

- [ ] **Step 7 (5 min): Finalize and activate with full stale-state recheck**

After the assembled immutable artifact/effect is committed and resolvable by operation ID, call `authority.Coordinator.Finalize(operation_id,effect_digest)` outside locks; only the coordinator captures/binds database coordinates and finalizes visibility. Then re-lock root/metadata pointers and intent. Require the exact committed receipt, base versions/digests, expected keys, key statuses, incident state and deadline before writing the active artifact row, advancing one pointer, marking active, and committing audit/outbox. Otherwise mark intent/shares superseded and keep the artifact unreachable.

- [ ] **Step 8 (3 min): Run GREEN threshold and isolation tests**

Run: `go test ./internal/nodecontrol/state -run 'TestRootMetadataPublisher|TestRootShareIsolation' -count=1`

Expected: PASS; the online signer fake receives zero root calls and cannot be adapted to `RootShareProvider` without an explicit compile-time implementation.

- [ ] **Step 9 (5 min): Add crash recovery and emergency revoke integration tests**

Crash after intent, each share, threshold, fence visibility, provider finalize, and pointer activation. Open an exact `online_signer_equivocation` incident, publish an emergency revoke, then resolve/change it before activation and prove stale shares cannot activate.

Run: `go test -tags=integration ./internal/nodecontrol/state -run 'TestRootPublisherCrashRecovery|TestEmergencyMetadataIncidentBinding' -count=1 -timeout 5m`

Expected: PASS with one authority operation, one active artifact, and the same terminal response on retry.

- [ ] **Step 10 (4 min): REFACTOR provider ACL assertions and terminal cleanup**

Expose required capability names `node_state_sign` and `node_state_root_share` as disjoint constants used by provider wiring tests. Terminal intents retain immutable shares/audit; cleanup removes none before retention and never changes active pointers.

Run: `go test ./internal/nodecontrol/state ./internal/nodecontrol/operator ./internal/nodecontrol/inventory -count=1`

Expected: PASS.

- [ ] **Step 11 (3 min): Run the Batch 02 exit gate**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/inventory ./internal/nodecontrol/operator ./internal/nodecontrol/state ./internal/store -count=1 -timeout 5m`

Expected: PASS.

Run: `go test -tags=integration ./internal/nodecontrol/inventory ./internal/nodecontrol/operator ./internal/nodecontrol/state ./internal/store -count=1 -timeout 8m`

Expected: PASS with inventory cursor/audit, independent generation gaps, signer crash recovery, threshold root rotation, emergency revoke, and finalized-only serving coverage.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS followed by `git diff --exit-code -- db/queries internal/store testdata/c12/node-state-vectors.json` exiting 0.

- [ ] **Step 12 (2 min): Commit the root publisher and Batch 02 gate**

```bash
git add db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/querier.go internal/nodecontrol/state/root_publisher.go internal/nodecontrol/state/root_publisher_test.go internal/nodecontrol/state/root_publisher_integration_test.go internal/nodecontrol/state/root_publisher_crash_test.go
git commit -m "feat(nodecontrol): publish threshold node trust metadata"
```

## Batch 02 completion check

- POP/node/endpoint/slot/profile inventory validates finite enums, exact ranges, optimistic versions, stable keysets and byte budgets.
- Every authenticated operator mutation has one transactional audit row and privacy-safe outbox event.
- Cursor plaintext is AEAD-bound to endpoint, operator, exact role/POP scope, normalized filter, page size and 15-minute lifetime; every page is freshly authorized.
- Desired and recovery generations are independent, monotonic and never reused; only fence-finalized active pointers are served.
- External signer calls occur outside database locks through fixed 32/64/128 queues and 2/4/2 workers.
- Drain becomes visible only with desired activation; signer failure cannot half-transition operator state.
- Root/metadata publication requires distinct verified threshold shares, current/new role separation and no online signer capability.
- Crash recovery and response retry reuse exact operation/signing/publish IDs and never create a second authority effect.
- No Batch 03 certificate, TLS listener, trust-bundle or operator-client guard production behavior has entered this batch.
