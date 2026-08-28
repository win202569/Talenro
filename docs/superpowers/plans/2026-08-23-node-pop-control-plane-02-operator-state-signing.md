# Talenro C1.2 Batch 02 Operator State and Signing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付受 PostgreSQL 和 Batch 01 authority fence 约束的 POP/node inventory、operator authorization/cursor/audit、desired/recovery signed-state saga、与 online signer 完全隔离的 root-threshold metadata publisher，以及严格 consumer-only 的 v7 fresh-restore manifest/state-domain 导入。

**Architecture:** Inventory 与 operator audit 是短事务内的数据库权威；opaque list cursor 只封装稳定 keyset continuation，不携带授权。Desired/recovery workflow 先 reserve authority sequence 与永不复用 generation，再在锁外调用有界 `NodeStateSigner`，最后经 fence receipt 和第二次状态校验激活。Root/metadata publication 使用独立的 threshold-share provider、独立 pending cap 和独立 activation path，online signer 没有 root 能力。Fresh restore 只验证并消费 B01 已冻结的 `FreshRestoreImportManifestV1`/single-use staging capability，通过 B01-owned transaction function 写入 application 与 disabled state-domain projection；B02 不创建 v7 schema，也不拥有 provider、attestor、Fence 或恢复编排。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、Ed25519、RFC 8785 JCS、SHA-256、AES-256-GCM、Batch 01 `contracts`/`authority`/`store`、现有 outbox 与 strict JSON helpers。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`; [approved authority v7 upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`, especially §10 B02 and §11.5; [approved canonical authority evidence/dispatcher addendum](../specs/2026-08-28-nodecontrol-authority-canonical-dispatcher-addendum-design.md), approved content SHA-256 `7D480607A92214627C1CEF3E81AF76610EC861508CE8D510D5BD046E82600FE2`, especially §§8–10; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json).

## Global Constraints

- 本册只完成 suite index 的 `C1.2-B02`；不得实现 certificate issuer、bootstrap/agent/operator TLS listener、node-agent、supervisor 或 core adapter。
- v7 amendment 下本册严格保持 consumer-only：只消费 B01 的 canonical schema/envelope、00007 table/function 与 generated store API；不得新增或修改 migration、table、trigger、security-definer function、wire schema、digest domain 或 SpecDigest。
- B02 不构造/签发 staging capability、recovery intent、revocation、commit challenge/attestation、Fence、Inspect、Release/Abort 或 evidence；B03按B01固定role签署provider/attestor/challenge/Fence/Inspect response，B11独占编排以及v7 §10明列的operator/source/downgrade/staging/recovery/application evidence角色。
- `FreshRestoreImportManifestV1` 只能在 B01-owned `nodecontrol_staging_importer` transaction boundary 中消费一次；transaction 内零 provider/attestor/signer/network 调用，存在 recovery intent、wrong held tuple、过期/replayed apply ID 或 projection mismatch 均零写失败。
- 导入只创建 manifest 明确允许的 disabled node skeleton 与 state-domain current rows；legacy active/pending pointer、legacy generation、resume intent、desired/recovery signing intent/result、certificate/grant/secret/provider result一律不复制。
- `FreshRestoreImportApplicationV1` 永不允许 Fence rerun。若 held 后 PITR 丢失未归档 import application，B02 固定拒绝 reimport；进入 intent-backed recovery revocation→Abort 属于 B03/B11 边界。
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
internal/nodecontrol/contracts/process_spec.go
internal/nodecontrol/contracts/resource_envelope.go
internal/nodecontrol/state/contracts.go
internal/nodecontrol/state/canonical.go
internal/nodecontrol/state/provider.go
internal/nodecontrol/state/signer_queue.go
internal/nodecontrol/state/repository.go
internal/nodecontrol/state/postgres_repository.go
internal/nodecontrol/state/service.go
internal/nodecontrol/state/root_publisher.go
internal/nodecontrol/state/fresh_restore_import.go
internal/nodecontrol/state/fresh_restore_projection.go
testdata/c12/integration-operator-state-signing.v1.json
```

Batch 01 supplies all v7 canonical contracts, exact projection/admission verifiers, opaque `VerifiedFreshRestoreImportAdmission`, tables, triggers, `begin_staging_import` security-definer function and its repository/store adapter. Batch 03 consumes the active inventory/identity predicates and listener-facing authorizers and owns production provider/attestors；Batch 11只签发v7 §10列举的operator/source/staging/recovery evidence、排序调用各服务，并必须通过本册consumer API导入，不能直调store/function。Batch 04 consumes signed envelopes and verification vectors. Neither later batch may reinterpret B02 generations, cursor binding, signing kinds, transcript domains, root roles, terminal states or the disabled-only imported state projection.

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

type IdentityState string

const (
	IdentityNeverEnrolled   IdentityState = "never_enrolled"
	IdentityActive          IdentityState = "active"
	IdentityRecoveryPending IdentityState = "recovery_pending"
	IdentityRecoveryLimited IdentityState = "recovery_limited"
	IdentityRevoked         IdentityState = "revoked"
	IdentityUnauthorized    IdentityState = "unauthorized" // output/import projection only
)

type Node struct {
	NodeID                     uuid.UUID
	POPCode                    string
	OperatorState              OperatorState
	SecurityState              SecurityState
	IdentityState              IdentityState
	ResumeOperatorState        *OperatorState
	PendingTransition          *string
	PendingTransitionSigningID *uuid.UUID
	IdentityEpoch              uint64
	LineageID                  *uuid.UUID
	InventoryVersion           uint64
	SecurityVersion            uint64
	ResourceEnvelopeVersion *uint64
	ResourceEnvelopeDigest *contracts.Digest
	ActiveDesiredGeneration  *uint64
	NextDesiredGeneration    *uint64
	ActiveRecoveryGeneration *uint64
	NextRecoveryGeneration   *uint64
	ActiveRootPublishID       *uuid.UUID
	ActiveRootVersion         *uint64
	ActiveMetadataPublishID   *uuid.UUID
	ActiveMetadataVersion     *uint64
	LastAuthorityOperationID  *uuid.UUID
	LastAuthority             *contracts.AuthorityVersion
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
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

`LineageID` is nil exactly for `never_enrolled` and the B01 verified-import-only `unauthorized` projection；all other identity states require a nonzero lineage and positive identity epoch. Pending transition/signing ID、resource-envelope version/digest、root publish ID/version、metadata publish ID/version and last-authority operation/epoch/sequence are each all-or-none groups；desired/recovery active and next generations obey the B01 nullable ordering checks. `Node` is the complete internal projection used by create/get/update/list response builders, including exact timestamps；there is no partial list-item DTO that can omit required `NodeV1` fields. `Mutation` has no identity-state、identity-epoch、lineage、security-version、resource-pointer、active/next pointer、timestamp or authority fields；ordinary create/update rejects `unauthorized` and cannot select or mutate any output-only group. Only the B01 security-definer fresh-import path can create the exact unauthorized/quarantined/disabled zero-state row.

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

Run: `go test ./internal/nodecontrol/inventory -count=1`

Run: `go test ./internal/nodecontrol/inventory -run '^$' -fuzz '^FuzzValidateAggregate$' -fuzztime=10s -timeout 30s`

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

Cover stale expected version, immutable referenced capacity profile, endpoint replacement atomicity, 8-slot cap, bytewise UUID keyset ordering, no duplicate across pages, and a query canceled at the 2-second statement timeout. Add row/projection cases for every closed identity state、positive `security_version`、nil/non-nil lineage semantics、all-or-none resource and authority groups, plus ordinary create/update attempts to select `unauthorized` or mutate identity/security/authority pointer fields；only the B01 verified-import fixture may materialize the exact unauthorized projection.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/inventory' -Run '^TestPostgresInventoryRepository$' -Timeout 3m`

Expected: FAIL because `nodecontrol_inventory.sql` and repository types do not exist.

- [ ] **Step 3 (5 min): Add exact locking, update, and keyset queries**

```sql
-- name: LockNodeInventory :one
SELECT node_id, pop_code, operator_state, security_state, resume_operator_state,
       identity_state, pending_operator_transition, pending_transition_signing_id,
       identity_epoch, lineage_id,
       inventory_version, security_version, resource_envelope_version,
       resource_envelope_digest, active_desired_generation, next_desired_generation,
       active_recovery_generation, next_recovery_generation,
       active_root_publish_id, active_root_version,
       active_metadata_publish_id, active_metadata_version,
       last_authority_operation_id, last_authority_epoch, last_authority_sequence,
       created_at, updated_at
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
          identity_state, pending_operator_transition, pending_transition_signing_id,
          identity_epoch, lineage_id,
          inventory_version, security_version, resource_envelope_version,
          resource_envelope_digest, active_desired_generation, next_desired_generation,
          active_recovery_generation, next_recovery_generation,
          active_root_publish_id, active_root_version,
          active_metadata_publish_id, active_metadata_version,
          last_authority_operation_id, last_authority_epoch, last_authority_sequence,
          created_at, updated_at;

-- name: ListNodeInventoryAfter :many
SELECT node_id, pop_code, operator_state, security_state, resume_operator_state,
       identity_state, pending_operator_transition, pending_transition_signing_id,
       identity_epoch, lineage_id, inventory_version, security_version,
       resource_envelope_version, resource_envelope_digest,
       active_desired_generation, next_desired_generation,
       active_recovery_generation, next_recovery_generation,
       active_root_publish_id, active_root_version,
       active_metadata_publish_id, active_metadata_version,
       last_authority_operation_id, last_authority_epoch, last_authority_sequence,
       created_at, updated_at
FROM nodecontrol.node_inventory
WHERE (NOT sqlc.arg(has_after)::boolean OR node_id > sqlc.arg(after_node_id)::uuid)
  AND (sqlc.arg(pop_code)::text = '' OR pop_code = sqlc.arg(pop_code)::text)
  AND (sqlc.arg(operator_state)::text = '' OR operator_state = sqlc.arg(operator_state)::text)
ORDER BY node_id ASC
LIMIT sqlc.arg(page_limit);
```

The same total row converter is used for lock/get/update/list and maps nullable resume/transition/lineage/resource/generation/root/metadata/authority groups to the complete `Node`: SQL NULL stays nil, every all-or-none group and generation ordering rule is enforced, timestamps are nonzero UTC instants, and present enum/version/digest values are closed、positive and exact-width. Tests compare every repository path and generated `NodeV1` response field-for-field for never-enrolled、ordinary active/recovery/revoked、verified-import unauthorized projection, each legal resume value, partial nullable groups and corrupt text；no partial list projection, empty-string or zero sentinel is accepted.

- [ ] **Step 4 (3 min): Generate sqlc artifacts and inspect drift**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_inventory.sql internal/store`

Expected: PASS; the second generation changes none of the exact staged inventory outputs and leaves no untracked generated artifact.

- [ ] **Step 5 (5 min): Implement transactions, validation, and exact conflict mapping**

In `postgres_repository.go`, execute `SET LOCAL statement_timeout = '2s'` for list transactions, map zero-row optimistic updates to `ErrVersionConflict`, map unique/check violations to finite inventory sentinels, and reject `Limit` outside 1–201 before SQL. Fetch `Limit=page_size+1`, return at most page size items, and derive `NextAfterNodeID` only from the last returned item.

- [ ] **Step 6 (3 min): Run GREEN repository tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/inventory' -Run '^TestPostgresInventoryRepository$' -Timeout 3m`

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
- Consumes: standard-library `crypto/x509` certificate facts, external policy provider, AES-256-GCM key material, and `inventory.ListRequest`；it has no dependency on future Batch 03 code.
- Produces: suite-index `OperatorAuthorizer`, canonical `operator.Credential`, closed roles/actions, `OperatorListCursorV1`, `CursorCodec`, and normalized filter/scope digests. Batch 03 later validates its mTLS peer and constructs this credential through an adapter.

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

Run: `go test ./internal/nodecontrol/operator -run '^(TestAuthorize|TestOperatorListCursor)' -count=1`

Run: `go test ./internal/nodecontrol/operator -run '^$' -fuzz '^FuzzCursorFrame$' -fuzztime=10s -timeout 30s`

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
- Produces: immutable `AuditEntry`, `AuditWriter`, shared acyclic `AuthorizedMutationBinding`, `MutationCommand`, and `Service.Execute(context.Context, MutationCommand) (MutationResult, error)` with one database transaction for ordinary domain row、audit and outbox. B03 fenced security workflows consume the binding but own their Coordinator orchestration.

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
	Authority               *contracts.AuthorityVersion
	BeforeInventoryVersion  uint64
	AfterInventoryVersion   uint64
	OccurredAt              time.Time
}

type AuthorizedMutationBinding struct {
	OperationID             uuid.UUID
	IdempotencyKeyDigest    contracts.Digest
	IfMatchDigest           contracts.Digest
	TargetScopeDigest       contracts.Digest
	Credential              Credential
	Authorization           Authorization
	AuthorizedAt            time.Time
}

type AuditWriter interface {
	Append(context.Context, store.DBTX, AuditEntry) error
}
```

- [ ] **Step 1 (4 min): Write RED rollback and idempotency integration tests**

Inject failures after inventory update, after audit insert, and after outbox insert; assert all three tables roll back. Retry the same operation ID and idempotency key; assert one domain version, one audit row, and one event ID.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/operator' -Run '^TestOperatorMutationAtomicAuditOutbox$' -Timeout 3m`

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
ON CONFLICT (operation_id) DO NOTHING
RETURNING audit_id;
```

If the insert returns no row, select the immutable row in the same caller transaction and constant-time compare every command、credential、target、reason、result、optional authority and before/after version field；return exact retry only on full equality and conflict otherwise. No idempotency path executes UPDATE, including an apparent no-op update.

- [ ] **Step 4 (3 min): Regenerate store artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_inventory.sql internal/store/nodecontrol_inventory.sql.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_inventory.sql internal/store`

Expected: PASS; the generated audit method accepts fixed-width digest bytes and returns one audit ID, while second-generation drift and untracked output are empty.

- [ ] **Step 5 (5 min): Implement the mutation transaction and safe event payload**

Authorize before opening the transaction and freeze the exact `AuthorizedMutationBinding`, then lock/recheck node scope and `If-Match` inside it. Write the domain row, immutable successful audit, and outbox event before commit. The event includes only event ID, node ID, inventory version, POP code, operator lifecycle, and changed-field enum；omit endpoint address, reason text, credential and signed bytes. Provider-authorized actions receive a non-nil finalized `AuthorityVersion`；inventory-only writes use a nil binding and SQL `NULL/NULL`. Zero authority pairs are invalid and never encode absence.

- [ ] **Step 6 (3 min): Run GREEN atomicity tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/operator' -Run '^TestOperatorMutationAtomicAuditOutbox$' -Timeout 3m`

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
- Create: `internal/nodecontrol/contracts/local_release.go`
- Test: `internal/nodecontrol/contracts/local_release_test.go`
- Create: `internal/nodecontrol/contracts/resource_envelope.go`
- Test: `internal/nodecontrol/contracts/resource_envelope_test.go`
- Create: `internal/localrelease/approved_manifest.go`
- Create: `internal/localrelease/approved-release-manifest.v1.json`
- Test: `internal/localrelease/approved_manifest_test.go`
- Create: `internal/nodecontrol/state/contracts.go`
- Create: `internal/nodecontrol/state/canonical.go`
- Test: `internal/nodecontrol/state/contracts_test.go`
- Test: `internal/nodecontrol/state/canonical_test.go`
- Create: `testdata/c12/node-state-vectors.json`
- Test: `internal/nodecontrol/state/cross_schema_test.go`

**Interfaces:**
- Consumes: `contracts.AuthorityVersion`, `contracts.Digest`, `contracts.LocalVersionedDigestV1`, inventory slot/profile types, RFC 8785 JCS, Ed25519 public keys.
- Produces: the one canonical adapter/process-spec IR；the sole canonical `ApprovedReleaseManifestV1` / `InstalledReleaseMapV1` types, strict verifiers and immutable verified handles；the build-embedded canonical approved-manifest byte source；the sole canonical `NodeResourceEnvelopeV1`/package contract required by B03's server-side publisher；closed signed-state/root/metadata DTOs；exact transcript constructors、semantic validators、transition validators and deterministic vectors consumed by B03/B04/B05/B07. Plan 04 and Plan 05 independently verify the embedded bytes and the deployed map/root facts；neither accepts a caller-built local-version tuple.

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

type ApprovedReleaseManifestV1 struct {
	SchemaVersion   string
	ManifestVersion uint64
	SortedReleases  []ApprovedReleaseV1
}

type ApprovedReleaseV1 struct {
	Adapter                 Adapter
	ReleaseID               string
	ProfileID               string
	ReleaseLockDigest       Digest
	ServerExecutableDigest  Digest
	ClientExecutableDigest  Digest
	DependencyClosureDigest Digest
	CompilerProfileDigest   Digest
	SandboxProfileDigest    Digest
	LicenseRecordDigest     Digest
}

type InstalledReleaseMapV1 struct {
	SchemaVersion          string
	MapVersion             uint64
	ApprovedManifestDigest Digest
	SortedReleases         []InstalledReleaseRootV1
}

type InstalledReleaseRootV1 struct {
	ReleaseID           string
	ImmutableReleaseRoot string
}

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

type ResourceLimitsV1 struct {
	CPUMillicores uint64 `json:"cpu_millicores"`
	MemoryBytes   uint64 `json:"memory_bytes"`
	TaskLimit     uint64 `json:"task_limit"`
	FDLimit       uint64 `json:"fd_limit"`
}

type NodeResourceEnvelopeV1 struct {
	SchemaVersion                   string           `json:"schema_version"`
	NodeID                          uuid.UUID        `json:"node_id"`
	ControlPlaneAuthorityEpoch      uint64           `json:"control_plane_authority_epoch"`
	AuthoritySequence               uint64           `json:"authority_sequence"`
	EnvelopeVersion                 uint64           `json:"envelope_version"`
	MaxSlots                        uint8            `json:"max_slots"`
	AgentLimits                     ResourceLimitsV1 `json:"agent_limits"`
	SupervisorLimits                ResourceLimitsV1 `json:"supervisor_limits"`
	CoreParentLimits                ResourceLimitsV1 `json:"core_parent_limits"`
	AggregateSlotFDReservationLimit uint64           `json:"aggregate_slot_fd_reservation_limit"`
	AggregateTmpfsBytes             uint64           `json:"aggregate_tmpfs_bytes"`
	AggregateTmpfsInodes            uint64           `json:"aggregate_tmpfs_inodes"`
	DetectedHostCapacityDigest      Digest           `json:"detected_host_capacity_digest"`
	IssuedAt                        time.Time         `json:"issued_at"`
}

type NodeResourceEnvelopePackageV1 struct {
	Envelope                     NodeResourceEnvelopeV1 `json:"envelope"`
	DeploymentKeyID              Digest                 `json:"deployment_key_id"`
	Algorithm                    string                 `json:"algorithm"`
	DeploymentAuthoritySignature [64]byte               `json:"deployment_authority_signature"`
}

const NodeResourceEnvelopeTranscriptV1 = "TALENRO-NODE-RESOURCE-ENVELOPE-V1\x00"

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

`ApprovedReleaseManifestV1` has schema `talenro-approved-release-manifest/v1`, size at most 64 KiB and 1–3 entries sorted uniquely by the declaration-order adapter then bytewise release/profile IDs. All versions are in `1..MaxInt64`; every digest is nonzero；adapter/release/profile pairs are unique；the release-lock、server/client executable、dependency closure、compiler profile、sandbox profile and license facts are complete and cannot be supplied by the installed map. Its digest is `SHA-256(ASCII("TALENRO-APPROVED-RELEASE-MANIFEST-V1") || 0x00 || RFC8785_JCS(manifest))`. `VerifyApprovedReleaseManifestV1([]byte)` strict-decodes、requires canonical round-trip and returns a defensive-copy `VerifiedApprovedReleaseManifestV1` with no exported field or bytes/digest/boolean constructor；its sole tuple accessor returns `contracts.LocalVersionedDigestV1{Version:ManifestVersion,Digest:manifest_digest}`. `RequireC12ProductionReleaseSet` additionally requires the exact final `fixture,xray,sing_box` set and rejects the initial fixture-only/partial development instance until Plan 07 has atomically regenerated the embedded instance from both reviewed release locks.

`InstalledReleaseMapV1` has schema `talenro-installed-release-map/v1`, size at most 16 KiB and a sorted unique nonempty subset of the verified manifest's release IDs. It contains only `{release_id,immutable_release_root}` plus its own positive map version and exact approved-manifest digest；adapter、profile、digest、argv、dependency closure、compiler/sandbox policy and license fields do not exist. Roots must be canonical absolute Linux paths below `/opt/talenro/releases/`, with no `.`/`..`/empty segment, and are not deletion authority. `VerifyInstalledReleaseMapV1([]byte, VerifiedApprovedReleaseManifestV1)` performs the bounded canonical/membership checks, derives `SHA-256(ASCII("TALENRO-INSTALLED-RELEASE-MAP-V1") || 0x00 || RFC8785_JCS(map))` and returns defensive-copy `VerifiedInstalledReleaseMapV1`; only after the consumer's independent filesystem checks may its private commit adapter call the handle's tuple accessor `contracts.LocalVersionedDigestV1{Version:MapVersion,Digest:map_digest}`. Plan 04 and Plan 05 must each no-follow open the fixed production file `/etc/talenro/releases/installed-release-map.v1.json`, require root ownership、non-writability、regular-file identity, then no-follow verify every listed immutable root and its manifest-pinned files before they may derive/persist the local tuple. The map path、bytes、tuple or a filesystem-verifier success boolean is never accepted from an agent request, supervisor request, environment variable or caller DTO.

`internal/localrelease/approved_manifest.go` uses `go:embed` for exactly package-local `approved-release-manifest.v1.json` and exposes only a fresh defensive copy of those build-fixed bytes. The initial P02 instance is fixture-only and explicitly production-incomplete so intermediate packages compile but production startup fails closed. Plan 07 `lock-core` is the sole later writer: after each reviewed lock it atomically regenerates this fixed file from all then-present reviewed locks, and after B09 it must satisfy `RequireC12ProductionReleaseSet`. Any direct/manual edit, partial two-file update or embedded/runtime digest disagreement fails the owning gate.

`NodeResourceEnvelopeV1` has schema `node-resource-envelope.v1`, algorithm `ed25519`, canonical UUID/whole-second UTC time, authority/version values in `1..MaxInt64`, `max_slots=8`, and nonzero digests. Each `ResourceLimitsV1` uses CPU `100..64_000`, memory `67_108_864..1_099_511_627_776`, tasks `32..4_096`, and FDs `64..1_000_000`；aggregate FD reservation is `64..9_000_000`, aggregate tmpfs bytes `1..9_895_604_649_984`, and aggregate tmpfs inodes `1..9_000_000`. Checked sums must fit both the declared aggregate and `uint64`. Strict package canonicalization rejects unknown/duplicate/trailing fields, caps JCS at 64 KiB, and returns only `NodeResourceEnvelopeTranscriptV1 || JCS(package.envelope)` plus its digest；signature role/key verification belongs to B03 and cannot change this contract.

`ProcessSpecV1` validation applies the exact §11.1 ASCII/length rules, at most eight unique slots, the three closed adapters/lifecycles, sorted unique required metrics, immutable inventory/profile equality, and rejects paths, argv, environment, URLs, raw configuration, arbitrary maps and credentials because those fields do not exist. `NodeStateRootSetV1` contains schema version, authority epoch/sequence, root version, threshold and 1–5 bytewise key-ID-sorted unique Ed25519 root keys. `NodeStateTrustMetadataV1` contains a complete key snapshot of at most 64 entries, a sorted cumulative revoked-ID set of at most 4096 entries, metadata version/window, next refresh, authority/root binding and sorted threshold signatures. `NodeStateRootRotationBodyV1` embeds the next root set and requires exact previous version plus current/new signatures.

- [ ] **Step 1 (5 min): Write RED canonical-vector and cross-schema tests**

Load fixed input/output hex from `node-state-vectors.json`; test desired, recovery, time, metadata and root-rotation transcripts. Freeze process-spec/Nonce32 semantics, both local-release schemas/domains/immutable verified handles, embedded-byte defensive copies and literal resource-envelope JSON/transcript vectors, numeric edges, checked aggregate overflow, defensive copies and every one-field mutation. Exercise all five recovery reasons and all three recovery actions, then feed a valid C1.1 `talenro-trust-metadata/v1` payload to every C1.2 decoder and require `ErrSchemaMismatch` before signature acceptance. Local-release tests reject unknown/duplicate/trailing fields, unsorted/duplicate entries, map overrides, manifest/map splice, noncanonical/relative/escaping roots, zero/overflow versions, random nonzero tuple injection and production use of the initial partial manifest.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodecontrol/contracts ./internal/localrelease ./internal/nodecontrol/state -run 'TestProcessSpec|TestNonce32|TestLocalRelease|TestEmbeddedApprovedManifest|TestNodeResourceEnvelopeContract|TestCanonicalVectors|TestRejectC11Schemas' -count=1`

Expected: FAIL because state contracts and vector file do not exist.

- [ ] **Step 3 (5 min): Add strict DTO validation and transcript builders**

Require exact schema strings, UTC whole-second RFC3339 times, canonical UUIDs, sorted unique sets, exact enum values, nonzero authority/version fields, closed desired reason, and JCS roundtrip equality. Compute SHA-256 only after semantic validation. Return transcript bytes as a new allocation containing the fixed ASCII domain followed by canonical payload bytes.

- [ ] **Step 4 (5 min): Add metadata and root transition validation**

Enforce 7-day metadata maximum, refresh no later than 48 hours before expiry, active-key continuous half-open coverage, `active→retiring→revoked`, immutable key material/window, emergency revoke bound to an exact open `online_signer_equivocation` incident, 30-day tombstone retention, cumulative revoked ledger append-only, and exact current/new threshold for rotation.

- [ ] **Step 5 (4 min): Add desired/recovery/time deadline rules**

Desired effective deadline is `min(issued_at+24h, metadata.valid_until, key.not_after)`; recovery uses 15 minutes; time uses 5 minutes and exact 32-byte nonce. Time attestation uses the exact `node_authority_checkpoint_sequence` field, never a global head or another object's creation sequence. Recovery requires `all_slots_stopped=true`, at most 16 sorted unique incidents, at most 64 sorted unique local bindings, at most 16 sorted unique supervisor bindings, and a non-null remediation digest only for `clear_security_latches` (the other two actions require null). Add boundary tests for 16/17 incidents, 64/65 local bindings and 16/17 supervisor bindings.

- [ ] **Step 6 (3 min): Run GREEN vector and transition tests**

Run: `go test ./internal/nodecontrol/contracts ./internal/localrelease ./internal/nodecontrol/state -run 'TestProcessSpec|TestNonce32|TestLocalRelease|TestEmbeddedApprovedManifest|TestNodeResourceEnvelopeContract|TestCanonicalVectors|TestRejectC11Schemas|TestMetadataTransition|TestStateDeadlines' -count=1`

Expected: PASS and every vector digest/signature transcript matches the checked-in lowercase hex value.

- [ ] **Step 7 (4 min): REFACTOR canonicalization through one bounded helper**

The helper accepts a 64 KiB maximum for desired/recovery, bounded metadata collections, and rejects unknown/duplicate JSON fields before allocating nested collections. Add fuzz tests asserting decode-canonical-decode equality and no panic.

Run: `go test ./internal/nodecontrol/contracts ./internal/localrelease ./internal/nodecontrol/state -count=1`

Run: `go test ./internal/nodecontrol/contracts -run '^$' -fuzz '^FuzzProcessSpecCanonicalRoundTrip$' -fuzztime=10s -timeout 30s`

Run: `go test ./internal/nodecontrol/state -run '^$' -fuzz '^FuzzStateCanonicalRoundTrip$' -fuzztime=10s -timeout 30s`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit state contracts and vectors**

```bash
git add internal/nodecontrol/contracts/process_spec.go internal/nodecontrol/contracts/process_spec_test.go internal/nodecontrol/contracts/local_release.go internal/nodecontrol/contracts/local_release_test.go internal/nodecontrol/contracts/resource_envelope.go internal/nodecontrol/contracts/resource_envelope_test.go internal/localrelease/approved_manifest.go internal/localrelease/approved-release-manifest.v1.json internal/localrelease/approved_manifest_test.go internal/nodecontrol/state/contracts.go internal/nodecontrol/state/canonical.go internal/nodecontrol/state/contracts_test.go internal/nodecontrol/state/canonical_test.go internal/nodecontrol/state/cross_schema_test.go testdata/c12/node-state-vectors.json
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
- Consumes: Batch 01 migration/proof columns、`authority.Repository`、opaque-query `authority.RegisteredEffectResolver`、three-method `authority.RegisteredEffectActivator` and bound `serving.Reader`, task 5 canonical payloads, `store.DBTX`.
- Produces: immutable `SigningIntent`, closed `IntentStatus`, `RecoveryIntentBinding`, `LockedState`, the B02-owned recovery-guard/capture seam implemented by B03, and one reusable state handler registered separately for desired/recovery and, after Task 9, root/metadata kinds. It returns only authenticated `ActivationDecisionMaterial`, recomputes typed activation inputs during fresh activation, and validates persisted terminal outcomes without writes/external calls. No B02 repository is a production serving reader.

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
	Terminate(context.Context, store.DBTX, uuid.UUID, IntentStatus, FailureCode) error
	ResolveRegisteredAuthorityEffectForUpdate(context.Context, store.DBTX, authority.TransactionalEffectQuery) (authority.TransactionalResolvedEffect, error)
	CaptureActivationDecisionMaterial(context.Context, authority.Receipt) (authority.ActivationDecisionMaterial, error)
	ActivateAuthorityEffect(context.Context, store.DBTX, authority.Receipt, authority.ValidatedActivationDecisionEvidence) error
	ValidatePersistedAuthorityEffect(context.Context, store.DBTX, authority.Receipt, authority.ValidatedActivationDecisionEvidence, authority.AuthorityEffectResolution) error
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

type TransitionKind string

const (
	TransitionNone               TransitionKind = "none"
	TransitionDrain              TransitionKind = "drain"
	TransitionResume             TransitionKind = "resume"
	TransitionRestoreReauthorize TransitionKind = "restore_reauthorize"
)

type OperatorTransition struct {
	Kind                   TransitionKind
	ExpectedState          inventory.OperatorState
	TargetState            inventory.OperatorState
	RecoveryID             uuid.UUID
	RecoverySessionVersion uint64
	EffectDigest           contracts.Digest
	RestoreApprovalIDs     [2]uuid.UUID
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

type RestoreCredentialSnapshot struct {
	ApprovalID               uuid.UUID
	OperatorID               uuid.UUID
	CredentialSnapshotDigest contracts.Digest
	CredentialVersion        uint64
	ScopeDigest              contracts.Digest
}

type RecoveryTransitionCapture struct {
	OperationID               uuid.UUID
	SigningID                 uuid.UUID
	Binding                   RecoveryIntentBinding
	Transition                *OperatorTransition
	AttestationDigest         contracts.Digest
	ActivationDeadline        time.Time
	RestoreCredentialSnapshots *[2]RestoreCredentialSnapshot
}

type RestoreAuthorizationDecision struct {
	ApprovalID               uuid.UUID
	OperatorID               uuid.UUID
	CredentialSnapshotDigest contracts.Digest
	CredentialVersion        uint64
	ScopeDigest              contracts.Digest
	DecisionDigest           contracts.Digest
	Authorized               bool
	ValidUntil               time.Time
}

type RecoveryTransitionDecisionFacts struct {
	OperationID           uuid.UUID
	TransitionKind        TransitionKind
	CaptureBindingDigest  contracts.Digest
	DecisionAnchorDigest  contracts.Digest
	EvidenceValidUntil    time.Time
	RestoreAuthorizations *[2]RestoreAuthorizationDecision
}

type RecoveryTransitionDecisionCapture struct { /* unexported defensive facts + one-use seal */ }

func NewRecoveryTransitionDecisionCapture(RecoveryTransitionDecisionFacts) (RecoveryTransitionDecisionCapture, error)
func (c RecoveryTransitionDecisionCapture) Facts() RecoveryTransitionDecisionFacts

type RecoveryTransitionGuard interface {
	ValidatePrepare(context.Context, store.DBTX, RecoveryInput, *OperatorTransition) (RecoveryTransitionCapture, error)
	CaptureActivationDecisionCapture(context.Context, RecoveryTransitionCapture) (RecoveryTransitionDecisionCapture, error)
	ValidateActivation(context.Context, store.DBTX, RecoveryTransitionCapture, RecoveryTransitionDecisionCapture, authority.ValidatedActivationDecisionEvidence) error
	ApplyActivation(context.Context, store.DBTX, RecoveryTransitionCapture, RecoveryTransitionDecisionCapture, SignedEnvelope) error
}
```

- [ ] **Step 1 (5 min): Write RED generation and visibility integration tests**

Assert desired and recovery allocators are independent, aborted generation 2 is never reused, only one pending intent per node/kind exists, terminal state never returns to pending, and the registered activator cannot point at a pending/unverified signature or a different fence. Increment `security_version` without changing its textual state between prepare/activation and require terminal-not-applied, proving the locked projection selects and rechecks the captured numeric version. Add a fail-closed guard fake and lock probe：nil/unavailable guard rejects transition work；external decision/trusted-time capture occurs only outside DBTX；opaque query exposes only dispatcher-selected kind plus defensive expected Reservation；echo mutations in operation/kind/scope/epoch/sequence/effect digest fail；fresh activation accepts only the exact branded proof；parsed persisted validation accepts only one terminal intent and records zero writes/provider/time/signer calls. Mutate every stored commitment/Head/checkpoint/reason/time/identity/evidence/resolution field and every typed activation input；all fail before pointer/audit/outbox visibility. Prove the repository exposes no independent production active-state load；only the B01 `serving.Reader` integration fixture can return the resulting desired row after its same-connection readiness-query-readiness check.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^TestPostgresStateRepository$' -Timeout 3m`

Expected: FAIL because state SQL and repository do not exist.

- [ ] **Step 3 (5 min): Add lock, allocator, and immutable intent queries**

```sql
-- name: LockNodeStatePointers :one
SELECT node_id, inventory_version, identity_epoch, security_version, security_state,
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

- [ ] **Step 4 (4 min): Add transaction-bound resolution and finalized-only activation queries**

`ResolveRegisteredAuthorityEffectForUpdate` accepts only the caller DBTX and opaque `TransactionalEffectQuery`; it uses `RegisteredKind()` to choose the exact desired/recovery (and later root/metadata) SQL path, locks that real intent/pointer row in B01 order, compares its stored operation/epoch/sequence with `Expected()`, and returns the full echo plus `absent|prepared|committed|terminal`. For absent it returns the exact expected echo and otherwise-zero effect；for non-absent it returns the registered kind、exact scope/digest and finite state. It never guesses unknown/multiple rows or probes another table itself. Root/metadata registrations initially return valid absent echoes until Task 9 adds their intent paths.

The state handler constructor requires immutable expected trusted-time provider identity/configuration and the B02-owned `RecoveryTransitionGuard` seam, with no back-reference to the state service/Coordinator. `CaptureActivationDecisionMaterial` runs outside DBTX after an exact committed Receipt：it loads the immutable prepared capture, calls `RecoveryTransitionGuard.CaptureActivationDecisionCapture`, authenticates the approved rollback-resistant trusted-time/floor attestation and configured provider identity, derives the activation deadline and policy inputs from the exact intent, independently recomputes the commitment activation-input digest, and returns only `ActivationDecisionMaterial`. It never receives Provider、Head/checkpoint or constructs final evidence. Any capture failure discards the domain seal；known-not-committed retry recaptures.

Fresh `ActivateAuthorityEffect` runs inside the Coordinator-owned activation transaction with `ValidatedActivationDecisionEvidence`. It re-resolves the same intent, retrieves the exact domain decision capture, recomputes base generation、inventory/identity/security versions、root/metadata/key facts and every recovery ID/reason/session/status、incident/fault-binding/remediation/action、attestation、restore approval/credential/version/scope/decision/deadline input, and invokes the guard with defensive proof views. It atomically writes envelope/pointer、transition/session/approval disposition、the exact B01 `activation_` proof group、resolution、audit/outbox and terminal intent state；it never consumes admission. `ValidatePersistedAuthorityEffect` accepts only parsed-origin proof, locks/reads exactly one terminal `node_state_signing_intents` row under the supplied DBTX, recomputes the same typed digest, proves every stored preimage/terminal/pointer binding and performs zero writes/provider/trusted-time/signer calls. Missing/duplicate/nonterminal/cross-operation/corrupt rows fail closed. Only Coordinator consumes rollback-resistant admission after all writes and before its one Commit；none-time never consumes. There is no B02 serving query；the B01 reader never falls back to highest generation.

- [ ] **Step 5 (3 min): Generate and inspect state store artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_state.sql internal/store`

Expected: PASS; `nodecontrol_state.sql.go` has distinct desired/recovery allocator methods (no production active-load method), and the exact second-generation drift/untracked outputs are empty.

- [ ] **Step 6 (5 min): Implement exact row conversion and transition checks**

Reject zero versions/generations, digest length other than 32, signature length other than 64, unknown kind/status/failure, mismatched reserved generation, mutable canonical bytes, and any transition other than pending to one terminal state. Map the one-pending partial unique violation to `ErrSigningInProgress`.

- [ ] **Step 7 (3 min): Run GREEN repository tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^TestPostgresStateRepository$' -Timeout 3m`

Expected: PASS with desired generations `1,3` after failed generation 2 and no serving visibility for that failed row.

- [ ] **Step 8 (4 min): REFACTOR activation and bound-reader tests into a corrupt-row matrix**

Directly insert each invalid opaque-query echo、pointer/fence/status、commitment/Head/checkpoint/reason/time/identity/evidence/resolution JCS/digest and typed activation-input combination in a rolled-back test transaction. Require fresh activation to roll back with no pointer/audit/outbox visibility and parsed `ValidatePersistedAuthorityEffect` to reject it with zero observed write/provider/time/signer calls. Through B01 `serving.Reader`, require `ErrStateUnavailable` for every corrupt committed fixture and never silently skip to an older envelope；a compile/static test rejects a public `LoadActive` or independent readiness query in `internal/nodecontrol/state`.

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
- Consumes: Batch 01 `authority.Coordinator`, including its DB/provider-consistent read-only `CommittedNodeCheckpoint`; task 6 `NodeStateSigner`; task 7 `StateRepository` transaction-bound state effect handler and its already-defined recovery-guard/capture seam; inventory versions/capacity; root/metadata active key authorization; operator audit/outbox; and the `RecoveryTransitionGuard` implementation supplied by B03. It never receives raw `authority.Provider`.
- Produces: prepare/drive/recover API keyed by immutable signing ID；desired/recovery terminal results；and the finite inputs consumed by the registered state activator, which alone applies optional drain/resume/restore transition with exact state activation inside the Coordinator transaction.

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

```

`TransitionKind` contains `none`, `drain`, `resume`, and `restore_reauthorize`. B02 implements `drain`; B03 implements the guard and may construct `resume` or `restore_reauthorize` only after locking the exact session, attestation, incident set, host evidence and (for restore) two approval rows. Disable/quarantine is fail-closed identity work owned by B03 and cannot be represented as a signer-only transition. Only `ValidatePrepare`、`ValidateActivation` and `ApplyActivation` receive caller-owned DBTX and they cannot call an external provider；the separate `CaptureActivationDecisionCapture` receives no DBTX, authenticates the uncached authorizer/trusted-time/floor source when required and returns only a defensive sealed domain capture. The state handler converts that capture plus immutable intent facts into `ActivationDecisionMaterial`；it never constructs B01 evidence.

- [ ] **Step 1 (5 min): Write the RED happy path and stale activation tests**

Test exact order `Coordinator.Reserve → prepare transaction → Sign → local verify/store transaction → Coordinator.Finalize{Reservation-in sealed resolve → provider Receipt → Begin → handler material/trusted time → same-Provider Head/checkpoint → Complete/New → contextual proof → one activation transaction}`；there is no service-owned second activation transaction. Use lock-probing fakes to fail if coordinator/provider/signer/guard/trusted-time capture is invoked while a DB lock is held, and prove handler expected provider identity comes only from immutable configuration. For signer dependency failure、malformed result、expired pre-Finalize deadline and pre-Finalize supersession after the intent exists, require a fence-first transaction to call B01 `NewAuthorityEffectCommitment` with the finite `final_not_applied` disposition, persist that distinct immutable commitment/effect digest, make the resolver return `EffectCommitted`, and exact-retry `Coordinator.Finalize/Recover`；a recording provider must fail if this path calls Abort. Abort remains legal only before any intent exists and after `EffectAbsent`. Change any base/inventory/identity/security/key/recovery/session/incident/fault-binding/remediation/action/attestation/approval/credential/scope/decision/deadline input, any trusted-time signature/provider identity/expiry/floor digest, or any persisted proof/terminal column；require fail-closed before active bytes/pointer/audit/outbox. Parsed recovery must accept exactly one atomically terminal row and make zero writes/external calls.

- [ ] **Step 2 (2 min): Run the focused RED service test**

Run: `go test ./internal/nodecontrol/state -run 'TestStateServiceHappyPath|TestStateServiceStaleActivation' -count=1`

Expected: FAIL because `StateService` is not implemented.

- [ ] **Step 3 (5 min): Implement `Prepare` with immutable retry binding**

Call `authority.Coordinator.Reserve` for `EffectDesiredActivate` or `EffectRecoveryActivate` before opening a database transaction. Inside the transaction lock node/pointer, record that exact reservation with the immutable intent, verify `If-Match` and prerequisites, call `RecoveryTransitionGuard.ValidatePrepare` for recovery/resume/restore work, reserve one generation, construct exact canonical payload using the reserved authority version, capture all versions/key material/recovery bindings/deadline, write the intent, and return it without signature or outbox. Same signing/operation/idempotency IDs and same fields return the same intent; any changed field returns conflict.

- [ ] **Step 4 (5 min): Implement signer call and independent result verification**

Outside transactions call the queue with exact `SignRequest`; verify echoed fields, Ed25519 signature, canonical bytes/digest, schema/domain, current root/metadata/key authorization, and composite deadline. Persist verified signature in a short transaction. `IssueTimeAttestation` reads `Coordinator.CommittedNodeCheckpoint(SHA256(node_id))`, whose result already requires the exact provider/repository epoch、sequence and receipt digest to agree, constructs the exact five-minute `NodeTimeAttestationV1`, and signs it without calling `Reserve`; equal checkpoint values therefore do not create a fence row or advance an allocator. A retryable signer timeout before the deadline keeps the exact prepared intent recoverable. Once dependency failure、malformed/mismatched output、deadline expiry or supersession is determined after intent commit, use B01 `NewAuthorityEffectCommitment` in a short fence-first transaction to replace no payload bytes but append one immutable distinct `final_not_applied` commitment with its finite reason and audit inputs；the resolver must then return `EffectCommitted`, and the service calls `Coordinator.Finalize`, never Abort. Only a definite failure before any intent/domain row exists may use Abort after an exact `EffectAbsent` proof.

- [ ] **Step 5 (5 min): Implement fence finalize and activation recheck**

After either the verified immutable state artifact commitment or the Step 4 immutable `final_not_applied` commitment commits, call `authority.Coordinator.Finalize(operation_id,exact_commitment_digest)` outside all transactions. The registered state handler resolves through the opaque query and captures only authenticated material；Coordinator alone binds database point、provider Receipt、same-instance Head/checkpoint and branded proof. In the one activation transaction, `ActivateAuthorityEffect` re-locks/recomputes every typed input, requires exact Receipt/proof/domain capture, invokes `RecoveryTransitionGuard.ValidateActivation`, inserts the immutable envelope, advances the exact pointer, applies the optional transition/session/approvals, clears pending state, and stores the complete `activation_` proof group、resolution、audit/outbox and terminal intent atomically. Final-not-applied stores the same complete terminal proof group without active bytes/pointer. Coordinator alone handles rollback-resistant consume (none-time skips it) and Commit. On response loss, parsed recovery revalidates the stored Head/preimages and calls the handler's zero-write persisted validator；it never recaptures current Head/time unless absence is proved. The service never supplies system ID/timeline/LSN and never opens a post-Finalize transaction.

- [ ] **Step 6 (4 min): Implement drain atomicity**

The prepare transaction leaves `operator_state` unchanged, stores `pending_operator_transition='draining'`, and makes derived `accepting_new=false`. Only successful desired activation changes state to draining. A pre-intent `EffectAbsent` Abort or a post-intent signer-failure/supersession final-not-applied activation clears pending without changing the prior state；concurrent disable/quarantine retains disabled/quarantined state.

- [ ] **Step 7 (3 min): Run GREEN service tests**

Run: `go test ./internal/nodecontrol/state -run 'TestStateServiceHappyPath|TestStateServiceStaleActivation|TestDrainAtomicity|TestTimeAttestationUsesCommittedNodeCheckpoint' -count=1`

Expected: PASS and the test lock probe records zero external calls under lock.

- [ ] **Step 8 (5 min): Add deterministic crash recovery at every durable boundary**

Crash after reserve、intent、signer response、final-not-applied or success commitment、verified signature、provider finalize、Begin/material/trusted-time/Head/checkpoint/Complete/Validate boundaries、activator proof writes、admission burn、Commit invocation and response. Repeat with every typed input or trusted-time/preimage mutation and with corrupt/missing/partial stored terminal groups. `Recover(signingID)` delegates to `Coordinator.Recover`；a committed outcome is accepted only after parsed contextual and zero-write domain validation from the same locked snapshot, while proved absence recaptures fresh material. It cannot Abort `EffectPrepared/EffectCommitted`, allocate another generation, request changed bytes, retain stale approvals, use current Head in place of stored Head or run an independent activation transaction.

Run: `go test ./internal/nodecontrol/state -run TestStateServiceCrashRecovery -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/state' -Run '^TestStateServiceFinalNotAppliedPITR$' -Timeout 5m`

Expected: PASS for every crash point、response-loss retry and database restore between prepared/tombstone/provider/activation boundaries；provider Head prevents resurrection and the operation converges to the same terminal final-not-applied receipt with no pending fence.

- [ ] **Step 9 (4 min): REFACTOR failures into a closed matrix**

Use `validation_failed`, `signer_unavailable`, `signer_malformed`, `authorization_changed`, `deadline_expired`, and `superseded`; freeze which pre-intent cases may prove `EffectAbsent` and which post-intent cases become distinct B01 final-not-applied commitments. Map public callers to conflict or dependency unavailable without raw SQL/provider/key details.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state|./internal/nodecontrol/operator' -Timeout 5m`

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
- Consumes: task 5 root/metadata validators, task 6 `RootShareProvider`, Batch 01 `authority.Coordinator`, task 7 registered state handler/proof-column contract, root key registry, open security incidents, operator audit/outbox.
- Produces: `RootMetadataPublisher` with immutable publish IDs, distinct physical-key share persistence and threshold assembly；it registers the same state handler separately for root/metadata, returns authenticated material, atomically stores the exact `activation_` proof group, and validates parsed terminal outcomes with zero writes/external calls, so finalized-only pointers and audit/outbox commit inside the Coordinator transaction.

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

Verify each Ed25519 signature over the exact metadata or rotation transcript; persist valid shares immutably and ignore unknown shares for threshold. Sort accepted shares by key ID, assemble the exact envelope, then rerun canonical、base pointer/digest、coverage、cumulative revoke、incident、expected-key/status and current/new threshold validation from zero before committing a success artifact. Any determined failure after the publish intent exists but before a success commitment—including provider dependency/deadline、malformed or duplicate known-key material、insufficient threshold at the terminal collection boundary、canonical/coverage/revoke validation failure、stale base、incident or key-status change、threshold change or supersession—must write one distinct B01-canonical immutable `final_not_applied` commitment with a finite reason in a fence-first transaction；do not Abort, leave the intent pending or proceed with partially assembled bytes. A retryable provider timeout before the composite deadline keeps only the exact same intent recoverable；once the deadline or another terminal condition is determined it uses the tombstone path.

- [ ] **Step 7 (5 min): Finalize and activate with full stale-state recheck**

After either the assembled immutable artifact commitment or a Step 6 immutable final-not-applied commitment is committed, call `authority.Coordinator.Finalize(operation_id,exact_commitment_digest)` outside locks. The registered state resolver uses the opaque query and exact echo；material capture authenticates configured trusted time/floor identity and derives deadline/input digest solely from immutable publish facts. The activator re-locks pointers、intent/shares and recomputes base versions/digests、keys/status/threshold、incident、payload/envelope/share and deadline inputs. Success atomically writes artifact/pointer、terminal intent、complete `activation_` proof group、resolution/audit/outbox/fence visibility；not-applied writes the same proof group without artifact/pointer. Parsed `ValidatePersistedAuthorityEffect` accepts one exact terminal `node_root_metadata_publish_intents` row, recomputes every typed input, and performs zero writes/provider/time/share calls. Coordinator owns Head/checkpoint、admission and Commit；`RootMetadataPublisher` opens no post-Finalize transaction and never Aborts an existing intent.

- [ ] **Step 8 (3 min): Run GREEN threshold and isolation tests**

Run: `go test ./internal/nodecontrol/state -run 'TestRootMetadataPublisher|TestRootShareIsolation' -count=1`

Expected: PASS; the online signer fake receives zero root calls and cannot be adapted to `RootShareProvider` without an explicit compile-time implementation.

- [ ] **Step 9 (5 min): Add crash recovery and emergency revoke integration tests**

Crash after intent、each share、dependency failure、malformed/duplicate share、threshold/canonical/coverage/revoke/base/incident/key/status/supersession changes、either commitment、provider finalize、material/trusted-time/Head/checkpoint/proof capture、proof-group write、admission burn、Commit invocation/response and pointer/tombstone activation. Mutate each query echo、configured provider identity、attestation signature/expiry/floor、deadline、typed input、stored JCS/digest and terminal column. Repeat post-intent cases across PITR；committed outcomes must pass parsed zero-write validation from stored Head, while absent outcomes alone may recapture. Open an exact `online_signer_equivocation` incident, publish an emergency revoke, then resolve/change it before commitment and activation and prove stale shares cannot activate.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^(TestRootPublisherCrashRecovery|TestEmergencyMetadataIncidentBinding)$' -Timeout 5m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/state' -Run '^TestRootPublisherFinalNotAppliedPITR$' -Timeout 5m`

Expected: PASS with one authority operation, at most one active artifact, and the same success or final-not-applied terminal response after retry/PITR；no prepared publish fence remains pending.

- [ ] **Step 10 (4 min): REFACTOR provider ACL assertions and terminal cleanup**

Expose required capability names `node_state_sign` and `node_state_root_share` as disjoint constants used by provider wiring tests. Terminal intents retain immutable shares/audit; cleanup removes none before retention and never changes active pointers.

Run: `go test ./internal/nodecontrol/state ./internal/nodecontrol/operator ./internal/nodecontrol/inventory -count=1`

Expected: PASS.

- [ ] **Step 11 (3 min): Run the pre-v7 state/signing regression gate**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/inventory ./internal/nodecontrol/operator ./internal/nodecontrol/state ./internal/store -count=1 -timeout 5m`

Expected: PASS.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/inventory|./internal/nodecontrol/operator|./internal/nodecontrol/state' -Timeout 8m`

Expected: PASS with inventory cursor/audit, independent generation gaps, signer crash recovery, threshold root rotation, emergency revoke, and finalized-only serving coverage.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_state.sql internal/store testdata/c12/node-state-vectors.json`

Expected: PASS；the second generation has empty worktree-vs-index diff and no untracked generated/vector path.

- [ ] **Step 12 (2 min): Commit the root publisher sub-gate**

```bash
git add db/queries/nodecontrol_state.sql internal/store/nodecontrol_state.sql.go internal/store/querier.go internal/nodecontrol/state/root_publisher.go internal/nodecontrol/state/root_publisher_test.go internal/nodecontrol/state/root_publisher_integration_test.go internal/nodecontrol/state/root_publisher_crash_test.go
git commit -m "feat(nodecontrol): publish threshold node trust metadata"
```

### Task 10: Consume the v7 fresh-restore manifest into disabled state only

**Files:**
- Create: `internal/nodecontrol/state/fresh_restore_projection.go`
- Create: `internal/nodecontrol/state/fresh_restore_projection_test.go`
- Create: `internal/nodecontrol/state/fresh_restore_import.go`
- Create: `internal/nodecontrol/state/fresh_restore_import_test.go`
- Create: `internal/nodecontrol/state/fresh_restore_import_integration_test.go`
- Create: `testdata/c12/integration-operator-state-signing.v1.json`
- Modify: `internal/nodecontrol/state/service.go`
- Modify: `internal/nodecontrol/state/service_test.go`

**Interfaces:**
- Consumes: B01 opaque `authority.VerifiedFreshRestoreImportAdmission` plus B01 `ViewVerifiedFreshRestoreImportAdmission`, which returns defensive `contracts.FreshRestoreImportProjectionInputV1` containing only verified manifest topology and the exact capability/held lease/acquisition-locked/current Head/clock/route/identity/lineage/incarnation/rebind/runtime/catalog/pre/post bindings. B02 cannot construct or mutate the admission/view, duplicate its shared consume right, or consume a provider, archive, Fence, Inspect or terminal-transition client.
- Produces: `FreshRestoreProjectionBuilder.Build`, `FreshRestoreImportConsumer.Consume`, exact disabled node skeleton/state-domain arguments through the B01 repository boundary, and typed fail-closed errors `ErrFreshRestoreManifest`, `ErrFreshRestoreReplay` and `ErrStagingRecoveryRequired`.

```go
type FreshRestoreProjectionBuilder struct{}

func (FreshRestoreProjectionBuilder) Build(
	input contracts.FreshRestoreImportProjectionInputV1,
) (contracts.FreshImportTopologyProjectionV1, error)

type FreshRestoreImportConsumer struct {
	repository authority.FreshRestoreImportRepository
}

func NewFreshRestoreImportConsumer(authority.FreshRestoreImportRepository) (*FreshRestoreImportConsumer, error)

func (c *FreshRestoreImportConsumer) Consume(
	ctx context.Context,
	admission authority.VerifiedFreshRestoreImportAdmission,
) (contracts.FreshRestoreImportApplicationV1, error)
```

- [ ] **Step 1 (5 min): Write RED closed-projection unit tests**

Create table-driven tests whose only accepted objects are the B01 view's contract allowlist required to reconstruct the complete node set and state domain. Assert target activation/deployment/database identity and normalized catalog digest enter the projection exactly；mutating returned slices/bytes must not change the opaque admission or a second view. Every imported node must have `operator_state=disabled`、`security_state=quarantined`、`identity_state=unauthorized`、`health_state=unknown`、identity epoch `0`、inventory/security version `1`、`next_desired_generation=1`、`next_recovery_generation=1`、empty authority-anchor/active-pointer sets and null resume/pending transition. Reject duplicate canonical keys, unknown object type, payload/digest mismatch, incomplete node set, nonzero forbidden projection and any payload containing legacy pointer/generation/resume intent, desired/recovery intent/result, certificate, secret or provider state.

- [ ] **Step 2 (2 min): Run the projection RED tests**

Run: `go test ./internal/nodecontrol/state -run 'TestFreshRestoreProjection' -count=1`

Expected: FAIL because `FreshRestoreProjectionBuilder` is absent.

- [ ] **Step 3 (5 min): Implement the minimal disabled-only projection builder**

Read only the normalized target/catalog values and canonical manifest objects in `FreshRestoreImportProjectionInputV1`, defensively copy payloads, sort by `(object_type, canonical_key)`, recompute every payload digest, `complete_node_set_digest`, allowed-object-set digest and `FreshImportTopologyProjectionV1` body digest. Construct the exact fixed disabled/quarantined/unauthorized/unknown/version-1/empty-anchor defaults above；do not copy any legacy generation or pointer, but mechanically initialize both new next-generation counters to `1` as required by §7.2. No B02 reflection/unsafe code may inspect the opaque admission.

- [ ] **Step 4 (3 min): Run the projection GREEN tests**

Run: `go test ./internal/nodecontrol/state -run 'TestFreshRestoreProjection' -count=1`

Expected: PASS, including exact bytewise ordering and forbidden-field negatives.

- [ ] **Step 5 (5 min): Write RED consumer and transaction-boundary tests**

In `fresh_restore_import_integration_test.go` with first line `//go:build integration`, use a recording implementation of B01's `authority.FreshRestoreImportRepository` and B01 integration-tag test factory for an opaque verified admission；B02 must not redeclare a look-alike repository/view interface. Require the consumer to call `ViewVerifiedFreshRestoreImportAdmission`, preserve its exact manifest/capability apply ID, held exclusion lease, acquisition/current Head, target activation/database/incarnation/lineage/rebind/runtime, route/clock admission, normalized catalog and expected-post projection, and then pass the same opaque value once to the repository. Assert malformed/expired/replayed or non-held admissions cannot be constructed by the production verifier, a copied admission cannot obtain a second consume, and view/builder failures make zero repository calls. Add integration fixtures for B01's function proving exact `intent/capability/import/revocation/recovery=0/1/0/0/0` succeeds once, replay fails, recovery-intent present fails, capability/admission/runtime/identity mismatch fails and the transaction exposes zero provider/attestor/signer calls. The untagged `fresh_restore_import_test.go` covers only zero-value rejection、static dependency direction and seams that do not require constructing an opaque positive admission.

- [ ] **Step 6 (2 min): Run the focused consumer RED tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^(TestFreshRestoreImportConsumer|TestFreshRestoreImportTransaction)$' -Timeout 5m`

Expected: FAIL because the consumer and B01 store adapter are not wired.

- [ ] **Step 7 (5 min): Implement the consumer without adding schema or orchestration**

Accept only the already-verified opaque admission, obtain one defensive projection input through B01's view API and build the projection before opening the write transaction. Call the single B01 repository method once with the same opaque value；its adapter alone consumes the shared write right and expands normalized fields into the generated `begin_staging_import` params. Require returned `FreshRestoreImportApplicationV1` to copy capability/manifest/activation/identity/lineage/registration/runtime、staging exclusion lease、acquisition-locked Head、route/pre/post inventory, transaction snapshot and projection digests exactly. Map B01's recovery-intent/held/runtime mismatch to `ErrStagingRecoveryRequired`; never retry with a new apply ID and never call Fence, recovery, revocation or Abort from this package.

- [ ] **Step 8 (4 min): Prove the PITR and fresh-import Fence boundary**

Add a fixture where the original import application was committed but not archived and the restored candidate reports it absent. Assert `Consume` returns `ErrStagingRecoveryRequired`, writes no replacement application and has no archive/Fence dependency. Add compile/static checks over both `fresh_restore_import.go` and `fresh_restore_projection.go` that neither imports B03 provider/attestor nor B11 orchestration, and that no B02 code emits `staging_outcome_row_rerun` for `FreshRestoreImportApplicationV1`. The corresponding recovery revocation→Abort success test is required in Plan03/B11, not duplicated here.

- [ ] **Step 9 (4 min): Run GREEN unit, integration and race tests**

Run: `go test ./internal/nodecontrol/state -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^TestFreshRestoreImport' -Timeout 3m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/state' -Run '^TestFreshRestore' -Timeout 5m -Race`

Expected: PASS; imported projection is disabled-only, replay/recovery/PITR paths are zero-write, and no external dependency is called under the B01 transaction.

- [ ] **Step 10 (5 min): Run the final Batch 02 exit gate after fresh-import wiring**

Create the sorted Batch 02 integration manifest with one explicit package per group, exact top-level test names, required profile and bounded timeout. B01's parser validator must prove the selected Batch 01+02 manifests exactly cover every tracked first-line `//go:build integration` test inside their closed package union once, with no package glob、extra name or shared database group.

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/inventory ./internal/nodecontrol/operator ./internal/nodecontrol/state ./internal/store -count=1 -timeout 5m`

Run in the current PowerShell process:

```powershell
$expectedBatch02IntegrationTags = [ordered]@{
    'internal/nodecontrol/inventory/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/operator/service_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/state/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/state/service_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/state/root_publisher_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/state/fresh_restore_import_integration_test.go' = '//go:build integration'
}
$badBatch02IntegrationTags = @($expectedBatch02IntegrationTags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
} | ForEach-Object Key)
if ($badBatch02IntegrationTags.Count -ne 0) {
    $badBatch02IntegrationTags
    throw 'B02 integration test missing exact first-line tag'
}
```

Run: `git add internal/nodecontrol/state/fresh_restore_import_integration_test.go testdata/c12/integration-operator-state-signing.v1.json`

Expected: the new integration test and manifest are tracked in the index before the tracked-file parser runs；no broader path is staged.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Suite batch02 -Timeout 180m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code HEAD -- api gen internal/store`

Run: `powershell -NoProfile -Command "if (git ls-files --others --exclude-standard -- api gen internal/store) { throw 'untracked generated artifact' }"`

Expected: all PASS, including every Task 9 regression plus disabled-only fresh import, opaque-view/copy/consume negatives and B01 transaction integration；the committed generated tree and untracked set are clean. Only this step may close B02.

- [ ] **Step 11 (3 min): Verify ownership, staged state and generated artifacts stay untouched**

Run: `git diff --exit-code HEAD -- db/migrations db/schema db/queries internal/store api gen internal/nodecontrol/contracts internal/nodecontrol/authority internal/nodecontrol/claimv1 internal/nodecontrol/incarnation internal/nodecontrol/commitarchive internal/nodecontrol/operations internal/c12evidence testdata/c12/authority-v7 docs/superpowers/specs`

Run in the current PowerShell process:

```powershell
$b02ForbiddenOwnershipPaths = @(
    'db', 'internal/store', 'api', 'gen', 'internal/nodecontrol/contracts',
    'internal/nodecontrol/authority', 'internal/nodecontrol/claimv1',
    'internal/nodecontrol/incarnation', 'internal/nodecontrol/commitarchive',
    'internal/nodecontrol/operations', 'internal/c12evidence',
    'testdata/c12/authority-v7', 'docs/superpowers/specs'
)
$b02ForbiddenCached = @(git diff --cached --name-only -- @b02ForbiddenOwnershipPaths)
if ($b02ForbiddenCached.Count -ne 0) {
    $b02ForbiddenCached
    throw 'B02 forbidden path is staged'
}
$b02ForbiddenUntracked = @(git ls-files --others --exclude-standard -- @b02ForbiddenOwnershipPaths)
if ($b02ForbiddenUntracked.Count -ne 0) {
    $b02ForbiddenUntracked
    throw 'B02 forbidden path is untracked'
}
```

Expected: all three outputs are empty for this task；B02 created no schema/contracts, generated wire/store code, production provider/orchestrator, authority-v7 vectors or SpecDigest input, including pre-staged and untracked paths.

- [ ] **Step 12 (2 min): Commit only the B02 consumer boundary**

```bash
git add internal/nodecontrol/state/fresh_restore_projection.go internal/nodecontrol/state/fresh_restore_projection_test.go internal/nodecontrol/state/fresh_restore_import.go internal/nodecontrol/state/fresh_restore_import_test.go internal/nodecontrol/state/fresh_restore_import_integration_test.go internal/nodecontrol/state/service.go internal/nodecontrol/state/service_test.go testdata/c12/integration-operator-state-signing.v1.json
git commit --only -m "feat(nodecontrol): consume v7 fresh restore state" -- internal/nodecontrol/state/fresh_restore_projection.go internal/nodecontrol/state/fresh_restore_projection_test.go internal/nodecontrol/state/fresh_restore_import.go internal/nodecontrol/state/fresh_restore_import_test.go internal/nodecontrol/state/fresh_restore_import_integration_test.go internal/nodecontrol/state/service.go internal/nodecontrol/state/service_test.go testdata/c12/integration-operator-state-signing.v1.json
```

The task commit must not include migrations, generated contracts/store files, B03 provider/attestor code, B10 spec-set output or B11 orchestration/evidence code.

## Batch 02 completion check

- POP/node/endpoint/slot/profile inventory validates finite enums, exact ranges, optimistic versions, stable keysets and byte budgets.
- Every authenticated operator mutation has one transactional audit row and privacy-safe outbox event.
- Cursor plaintext is AEAD-bound to endpoint, operator, exact role/POP scope, normalized filter, page size and 15-minute lifetime; every page is freshly authorized.
- Desired and recovery generations are independent, monotonic and never reused；their envelope/pointer/resolution/audit/outbox visibility is committed only by the registered DBTX activator, and production desired bytes are served only by B01 `serving.Reader`.
- External signer calls occur outside database locks through fixed 32/64/128 queues and 2/4/2 workers.
- Drain becomes visible only with desired activation; signer failure cannot half-transition operator state.
- Root/metadata publication requires distinct verified threshold shares, current/new role separation and no online signer capability.
- Crash recovery and response retry reuse exact operation/signing/publish IDs and never create a second authority effect.
- Fresh restore consumes one exact B01 manifest/capability into a disabled-only node/state projection; recovery intent, replay, wrong tuple or forbidden legacy state produces zero writes.
- No legacy pointer/generation/resume intent/signing result is copied, and a lost unarchived import application is never reimported or Fence-rerun by B02.
- No Batch 03 certificate, TLS listener, trust-bundle or operator-client guard production behavior has entered this batch.
