# Talenro C1.2 Batch 03 Node Identity and mTLS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 CSR-bound enrollment/rotation、外部 issuer 与 fence-finalized certificate authorization receipt、host-deployed trust-bundle package、rollback-resistant operator trust guard、彼此隔离且有界的 bootstrap/agent/operator TLS 1.3 listeners，以及 authority v7 所需的 production ClaimV1 provider、runtime/timeline attestors、rollback-resistant commit archive/Fence、staging revocation provider 与 source/Down retirement enforcement。

**Architecture:** PostgreSQL 只保存 grant digest、issuance intent、certificate identity 与 trust high-water，永不保存私钥或 grant plaintext。Identity workflow 先 reserve authority sequence并在短事务消费 grant/建立 immutable intent，锁外调用外部 issuer，独立验证 exact X.509 profile，再经 fence receipt 激活。Listener 的 TLS chain validation 只建立 transport；每次 request 入场和 response commit 都按 exact leaf row、authority/identity epoch、certificate/node state 重授权。Trust package 与 operator guard 从 host deployment 的独立 role keys/rollback-resistant provider取得信任，不能从当前 TLS 连接自举。Authority provider/attestors 只消费 B01 的 canonical contracts/table functions；trusted read observation、primary commit stream、semantic-ID journal、prefix subject-slot reservation、stable once-row 与 Fence/Inspect evidence 存在 PostgreSQL 之外的 rollback-resistant provider failure domain，且在 DB transaction 外完成。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、PostgreSQL trusted read/WAL decoding（含 COMMIT/COMMIT PREPARED end LSN）、rollback-resistant append-only archive、ECDSA P-256、Ed25519、X.509、SHA-256、RFC 8785 JCS、TLS 1.3、HTTP/2、Batch 01 authority contracts/schema/functions、Batch 02 inventory/operator/state services。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`; [approved authority v7 upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`, especially §10 B03 and §11; [approved canonical authority evidence/dispatcher addendum](../specs/2026-08-28-nodecontrol-authority-canonical-dispatcher-addendum-design.md), approved content SHA-256 `7D480607A92214627C1CEF3E81AF76610EC861508CE8D510D5BD046E82600FE2`, especially §§8–10; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json).

## Global Constraints

- 本册只完成 suite index 的 `C1.2-B03`；不得实现 node-agent 本地 keystore/RollbackGuard、supervisor、observation reducer、core adapter 或 production deployment evidence。
- Batch 01 的 three OpenAPI packages、authority types/schema 和 Batch 02 的 inventory/operator/state types逐字消费；不得在 API adapter 中复制领域枚举。
- Production `NodeCertificateIssuer`、`OperatorAuthorizer`、`OperatorClientTrustGuard` provider 和 trusted time均为外部 boundary；local/test implementation 名称必须显式含 `Deterministic` 或 `LocalTest`。
- Node、operator、server leaf 只允许 ECDSA P-256、exact digitalSignature、exact one EKU、exact one SAN、empty subject、fixed extension allowlist；未知 critical/noncritical extension都拒绝。
- Bootstrap 不请求 client certificate；agent/operator 必须 `RequireAndVerifyClientCert`，但 CA-chain success 从不替代 exact PostgreSQL/application authorization。
- Server/client TLS 只用 TLS 1.3 与 ALPN `h2`；禁用 tickets、0-RTT、renegotiation、compression 和 `InsecureSkipVerify`。
- Grant plaintext只生成/返回一次，不落 PostgreSQL、日志、错误、audit、outbox、metrics 或 config；数据库只保存32-byte SHA-256 digest。
- 外部 issuer、authority provider、operator authorizer、trust guard provider均不得在数据库 lock/transaction 内调用。
- 只有 committed authority receipt加最终 activation transaction的 exact certificate可授权；pending/rejected/failed/superseded issuance和迟到 issuer result永不授权。
- Trust package只能由 host-deployed `DeploymentAuthorityKeySetV1` 的 exact role签名；bootstrap/poll/current TLS connection均不能更新 server CA 或 deployment keys。
- 每个 request 入场和 mutation/response commit前重查 exact leaf DER/public key/issuer/serial/status、identity epoch、authority epoch、node/operator state和 trusted time。
- B03 只消费 B01 冻结的 contracts、canonical schema/table/function 与 shared generated model foundations；本册不得创建或修改 migration、schema、wire/JCS/digest domain，也不得拥有 B10 `SpecDigest`。唯一 query/API 例外是本册 Frozen list 中三份 B03 query source、对应三份 sqlc file，以及 pinned generator 对 shared `models.go`/`querier.go` 的机械输出；这些 task 必须串行执行 exact-stage/second-generation drift gate，不能手写或改动其他 batch query API。
- B03 提供 production ClaimV1 provider、runtime/timeline attestors、WAL/archive/Fence/Inspect、staging revocation与 source/Down retirement enforcement，并只按B01固定role/schema签署自身provider/attestor response；不得拥有 B11 的 orchestration、authorization decision、operator-signed capability/intent/evidence、bundle选择或 Release/Abort policy selection。
- rollback-resistant archive/provider 必须独立于 PostgreSQL restore failure domain；所有 provider/attestor/archive/Fence 外部调用都在 DB transaction 外，DB 锁序与 provider append 顺序必须显式且可崩溃恢复。
- Fence once-key 只由稳定的六项 holder tuple 派生，mutable lease digest 不得进入 key；同一 holder lease rollover 必须复用原 key，只有 higher-generation holder 才能取得新 key。每次 signed Inspect freshness 仍逐项核对完整七项 candidate tuple。
- `FreshRestoreImportApplicationV1` 对所有 Fence reason 永久拒绝；lost/unarchived fresh import 不得靠 Fence 或 re-import 恢复，只允许由 recovery revocation application 导向 B11 Abort。
- 本计划中每个 checkbox 是一个可独立执行的 2–5 分钟动作；每个 task 必须 RED、GREEN、REFACTOR 后独立 commit。
- 只 stage task 的精确路径；不得 stage、删除或读取用户未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## Frozen file ownership

```text
db/queries/nodecontrol_identity.sql
db/queries/nodecontrol_recovery.sql
db/queries/nodecontrol_resource_envelope.sql
internal/store/nodecontrol_identity.sql.go
internal/store/nodecontrol_recovery.sql.go
internal/store/nodecontrol_resource_envelope.sql.go
internal/store/models.go                           # shared sqlc output; generator-only, task-scoped collision ownership
internal/store/querier.go                          # shared sqlc output; generator-only, task-scoped collision ownership
internal/nodecontrol/identity/provider.go
internal/nodecontrol/identity/types.go
internal/nodecontrol/identity/csr.go
internal/nodecontrol/identity/x509_profile.go
internal/nodecontrol/identity/receipt.go
internal/nodecontrol/identity/repository.go
internal/nodecontrol/identity/postgres_repository.go
internal/nodecontrol/identity/service.go
internal/nodecontrol/identity/security_fault_binding.go
internal/nodecontrol/identity/trust_conflict_binding.go
internal/nodecontrol/hostevidence/deployment_keys.go
internal/nodecontrol/hostevidence/trust_bundle.go
internal/nodecontrol/hostevidence/operator_guard.go
internal/nodecontrol/hostevidence/remediation.go
internal/nodecontrol/hostevidence/resource_envelope.go
internal/nodecontrol/hostevidence/resource_envelope_service.go
internal/nodecontrol/recovery/types.go
internal/nodecontrol/recovery/repository.go
internal/nodecontrol/recovery/postgres_repository.go
internal/nodecontrol/recovery/security_fault.go
internal/nodecontrol/recovery/trust_conflict.go
internal/nodecontrol/recovery/operator_actions.go
internal/nodecontrol/claimv1/provider.go
internal/nodecontrol/claimv1/head.go
internal/nodecontrol/claimv1/epoch.go
internal/nodecontrol/claimv1/consumption.go
internal/nodecontrol/claimv1/staging.go
internal/nodecontrol/claimv1/retirement.go
internal/nodecontrol/incarnation/runtime_attestor.go
internal/nodecontrol/incarnation/timeline_attestor.go
internal/nodecontrol/incarnation/rebind.go
internal/nodecontrol/commitarchive/archive.go
internal/nodecontrol/commitarchive/wal_decoder.go
internal/nodecontrol/commitarchive/postgres_wal_decoder.go
internal/nodecontrol/commitarchive/inspect.go
internal/nodecontrol/commitarchive/fence.go
internal/nodecontrol/commitarchive/journal.go
internal/nodecontrol/authorityadapter/provider.go
internal/nodebootstrapapi/handler.go
internal/nodebootstrapapi/server.go
internal/nodeagentapi/handler.go
internal/nodeagentapi/authorizer.go
internal/nodeagentapi/waiters.go
internal/nodeagentapi/server.go
internal/nodeoperatorapi/handler.go
internal/nodeoperatorapi/authorizer.go
internal/nodeoperatorapi/server.go
internal/platform/tlsserver.go
internal/config/nodecontrol.go
cmd/control-api/main.go
testdata/c12/integration-node-identity-mtls.v1.json
```

Batch 01 owns every canonical authority/staging table, function, migration and digest/wire contract plus the query APIs listed in B01；B03 consumes them without redefining them and owns only its three listed query sources/sqlc files. Because sqlc regenerates shared `internal/store/models.go`/`querier.go`, those two files are serialized generator outputs under the currently executing query task, not a hand-edited or permanent exclusive API ownership transfer. Batch 10 alone owns `SpecDigest`. Batch 11 owns orchestration, decision authorization, challenge construction, operator-signed capability/intent/evidence bundles and Release/Abort selection；B03 production provider/attestors只以B01封闭role签署其自身response/attestation，并独立重验B11提供的完整preimage，绝不代签operator或接受opaque digest。Batch 04 consumes `CertificateAuthorizationReceiptV1`, trust-package verification, endpoint authorization and signed-state handlers; it owns candidate keystore installation and outbound client transport. B03 does not write node private keys or agent rollback state.

### Task 1: Freeze issuer contracts, CSR parsing, and exact X.509 profiles

**Files:**
- Create: `internal/nodecontrol/identity/provider.go`
- Create: `internal/nodecontrol/identity/types.go`
- Create: `internal/nodecontrol/identity/csr.go`
- Create: `internal/nodecontrol/identity/x509_profile.go`
- Test: `internal/nodecontrol/identity/csr_test.go`
- Test: `internal/nodecontrol/identity/x509_profile_test.go`
- Test: `internal/nodecontrol/identity/x509_profile_fuzz_test.go`

**Interfaces:**
- Consumes: `contracts.Digest`, `contracts.AuthorityVersion`, `uuid.UUID`, standard `crypto/x509`, canonical trust-domain/DNS configuration.
- Produces: suite-index `NodeCertificateIssuer`, exact issue request/result, CSR/template/certificate identities, and validators below.

```go
type NodeCertificateIssuer interface {
	Issue(context.Context, IssueRequest) (IssueResult, error)
}

type IssueRequest struct {
	IssuanceID    uuid.UUID
	IssuerID      string
	CSRDER        []byte
	PublicKeyDER  []byte
	Template      NodeLeafTemplate
	TemplateDigest contracts.Digest
}

type IssueResult struct {
	IssuanceID     uuid.UUID
	IssuerID       string
	TemplateDigest contracts.Digest
	LeafDER        []byte
	IssuerChainDER [][]byte
}

type NodeLeafTemplate struct {
	NodeID       uuid.UUID
	TrustDomain string
	NotBefore   time.Time
	NotAfter    time.Time
}

func ParseNodeCSR(der []byte) (CSRIdentity, error)
func BuildNodeLeafTemplate(NodeTemplateRequest) (NodeLeafTemplate, contracts.Digest, error)
func ValidateIssuedNodeCertificate(IssueRequest, IssueResult, *x509.CertPool) (CertificateIdentity, error)
func ValidateOperatorLeaf(*x509.Certificate, OperatorLeafProfile) (OperatorLeafIdentity, error)
func ValidateServerLeaf(*x509.Certificate, ServerLeafProfile) error
```

- [ ] **Step 1 (5 min): Write RED CSR and extension-table tests**

Generate P-256, P-384, RSA and Ed25519 CSRs; test invalid signature, trailing DER, subject/SAN/extension request injection and duplicate attributes. Build node/operator/server leaves with every missing, extra, duplicate, wrong-criticality and AnyEKU extension case.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodecontrol/identity -run 'TestParseNodeCSR|TestExactLeafProfiles' -count=1`

Expected: FAIL because the identity package does not exist.

- [ ] **Step 3 (4 min): Implement bounded CSR parsing**

Reject DER above 16 KiB before parsing; require one ASN.1 object with no trailing bytes, valid CSR signature, ECDSA P-256 public key, empty subject and no requested extensions/SAN. Return SHA-256 of exact DER and marshaled PKIX public key plus its SHA-256; never retain caller-owned slices.

- [ ] **Step 4 (5 min): Implement exact node leaf validation**

Require `NotBefore=issuer_now-5m`, `NotAfter=NotBefore+24h`, nonzero positive serial at most 20 octets, empty subject, one URI SAN `spiffe://<node-trust-domain>/node/<canonical-node-uuid>`, clientAuth only, KeyUsage exactly digitalSignature, `BasicConstraintsValid=true`, `IsCA=false`, exact SKI/AKI derivation, allowed extension OIDs/criticality, exact CSR SPKI and chain to expected issuer.

- [ ] **Step 5 (4 min): Implement operator and server profile validation**

Operator uses one canonical operator URI SAN, clientAuth only and at most 8 hours. Server uses one configured DNS SAN, serverAuth only and at most 30 days. Both allow at most 5-minute backdate and share the same strict leaf extension rules; CA certificates use a separate CA validator with `IsCA=true` and certSign usage.

- [ ] **Step 6 (3 min): Run GREEN profile tests**

Run: `go test ./internal/nodecontrol/identity -run 'TestParseNodeCSR|TestExactLeafProfiles' -count=1`

Expected: PASS; each single-bit profile mutation is rejected by a finite sentinel error.

- [ ] **Step 7 (4 min): REFACTOR extension inspection and fuzz DER inputs**

Use one OID-indexed validator that counts each extension and checks criticality/value; no `x509.Verify` success may bypass it. Fuzz CSR/leaf parsers and assert no panic, large allocation, or partial identity result on error.

Run: `go test ./internal/nodecontrol/identity -count=1`

Run: `go test ./internal/nodecontrol/identity -run '^$' -fuzz '^FuzzCSRDER$' -fuzztime=10s -timeout 30s`

Run: `go test ./internal/nodecontrol/identity -run '^$' -fuzz '^FuzzLeafDER$' -fuzztime=10s -timeout 30s`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit issuer and X.509 contracts**

```bash
git add internal/nodecontrol/identity/provider.go internal/nodecontrol/identity/types.go internal/nodecontrol/identity/csr.go internal/nodecontrol/identity/x509_profile.go internal/nodecontrol/identity/csr_test.go internal/nodecontrol/identity/x509_profile_test.go internal/nodecontrol/identity/x509_profile_fuzz_test.go
git commit -m "feat(nodecontrol): freeze node certificate profiles"
```

### Task 2: Implement grant, issuance, and certificate persistence

**Files:**
- Create: `db/queries/nodecontrol_identity.sql`
- Create: `internal/nodecontrol/identity/repository.go`
- Create: `internal/nodecontrol/identity/postgres_repository.go`
- Create generated: `internal/store/nodecontrol_identity.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/identity/postgres_repository_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/identity/postgres_repository_integration_test.go`

**Interfaces:**
- Consumes: Batch 01 identity/proof columns、opaque-query `authority.RegisteredEffectResolver`、three-method `authority.RegisteredEffectActivator` and `store.DBTX`, Batch 02 inventory/node-state-transition locks, task 1 identities, immutable trusted-time provider identity/configuration.
- Produces: immutable grant/issuance/certificate records, closed statuses, an activation-only exact certificate lookup bound to the caller DBTX, and one identity handler registered separately for five exact effect kinds. It returns authenticated material, recomputes typed activation inputs, atomically stores the exact prefixed proof group, and validates parsed terminal outcomes with zero writes/external calls. It exposes no independent production serving reader.

```go
type CertificateStatus string

const (
	CertificateActive          CertificateStatus = "active"
	CertificateRecoveryPending CertificateStatus = "recovery_pending"
	CertificateRecoveryLimited CertificateStatus = "recovery_limited"
	CertificateRevoked         CertificateStatus = "revoked"
)

type IssuanceStatus string

const (
	IssuancePending    IssuanceStatus = "pending"
	IssuanceActive     IssuanceStatus = "active"
	IssuanceRejected   IssuanceStatus = "rejected"
	IssuanceSuperseded IssuanceStatus = "superseded"
	IssuanceFailed     IssuanceStatus = "failed"
)

type Repository interface {
	Transact(context.Context, func(context.Context, store.DBTX) error) error
	LockGrant(context.Context, store.DBTX, contracts.Digest) (EnrollmentGrant, error)
	ConsumeGrantAndInsertIssuance(context.Context, store.DBTX, EnrollmentGrant, IssuanceIntent) error
	LoadIssuance(context.Context, uuid.UUID) (IssuanceIntent, error)
	StoreValidatedResult(context.Context, store.DBTX, uuid.UUID, ValidatedIssueResult) error
	LookupExactCertificateForActivation(context.Context, store.DBTX, CertificateLookup) (CertificateAuthorization, error)
	RevokeIdentityEpoch(context.Context, store.DBTX, uuid.UUID, uint64, RevokeReason) error
	ResolveRegisteredAuthorityEffectForUpdate(context.Context, store.DBTX, authority.TransactionalEffectQuery) (authority.TransactionalResolvedEffect, error)
	CaptureActivationDecisionMaterial(context.Context, authority.Receipt) (authority.ActivationDecisionMaterial, error)
	ActivateAuthorityEffect(context.Context, store.DBTX, authority.Receipt, authority.ValidatedActivationDecisionEvidence) error
	ValidatePersistedAuthorityEffect(context.Context, store.DBTX, authority.Receipt, authority.ValidatedActivationDecisionEvidence, authority.AuthorityEffectResolution) error
}

type RecoveryIdentityRepository interface {
	RevokeIdentityEpoch(context.Context, store.DBTX, uuid.UUID, uint64, RevokeReason) error
	SupersedePendingIssuances(context.Context, store.DBTX, uuid.UUID, uint64, time.Time) error
	InvalidateUnconsumedGrants(context.Context, store.DBTX, uuid.UUID, uint64, time.Time) error
	CreateRecoveryGrant(context.Context, store.DBTX, RecoveryGrantIntent) (CreateGrantResult, error)
	SetRecoveryCertificateStatus(context.Context, store.DBTX, uuid.UUID, CertificateStatus, time.Time) error
	RequireNoOldPendingIssuance(context.Context, store.DBTX, uuid.UUID, uint64) error
}
```

- [ ] **Step 1 (5 min): Write RED constraints and exact-lookup integration tests**

Cover one unconsumed grant per node/epoch, 10-minute expiry, atomic claim, changed request digest, one issuance ID binding, terminal immutability, unique issuer/serial and leaf digest, exact public-key digest, max 4 unexpired active leaves per lineage, late old-epoch result, and a serial/SAN copy with changed DER/key. Freeze unique anchors：grant create=`node_enrollment_grants.create_`；grant claim=same row `claim_` consumption group；certificate activation=`node_certificate_issuances.activation_` joined to certificate；certificate revoke=`node_certificates.revoke_`；identity epoch advance=one immutable authority-bound `node_state_transitions` row with `authority_effect_kind=identity_epoch_advance` and explicit disposition. Assert each opaque registered query locks only its real anchor and echoes exact operation/kind/scope/epoch/sequence/effect digest；cross-group/cross-kind or duplicate rows conflict. Every inserted/served certificate joins an exact committed-active `certificate_activate` fence；a valid `grant_claim` fence cannot activate it.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/identity' -Run '^TestPostgresIdentityRepository$' -Timeout 3m`

Expected: FAIL because identity SQL and repository do not exist.

- [ ] **Step 3 (5 min): Add exact grant and issuance queries**

```sql
-- name: LockEnrollmentGrantByDigest :one
SELECT grant_id, token_digest, node_id, identity_epoch, csr_der_sha256,
       expires_at, consumed_at, expired_at
FROM nodecontrol.node_enrollment_grants
WHERE token_digest = sqlc.arg(token_digest)
FOR UPDATE;

-- name: ConsumeEnrollmentGrant :execrows
UPDATE nodecontrol.node_enrollment_grants
SET consumed_at = transaction_timestamp(), consumed_attempt_id = sqlc.arg(attempt_id)
WHERE grant_id = sqlc.arg(grant_id)
  AND consumed_at IS NULL
  AND expired_at IS NULL
  AND expires_at > transaction_timestamp();

-- name: LockCertificateIssuance :one
SELECT issuance_id, attempt_id, operation_id, node_id, identity_epoch,
       lineage_id, issuer_id, csr_der_sha256, public_key_sha256,
       template_digest, authority_epoch, authority_sequence, status
FROM nodecontrol.node_certificate_issuances
WHERE issuance_id = sqlc.arg(issuance_id)
FOR UPDATE;
```

- [ ] **Step 4 (4 min): Add activation-only exact certificate lookup**

The query uses issuer ID, unsigned serial bytes, exact leaf DER SHA-256, public-key SHA-256, node ID, identity epoch and authority epoch in one query; join inventory current identity epoch and the exact fence row. It accepts the Coordinator-owned caller `store.DBTX`, is callable only by the identity transaction-bound activator, and returns one row only when certificate status is one of the finite activation statuses and `not_before <= trusted_now < not_after`. Do not expose this query or repository as the request-serving authorization path；Task 6 must use the bound B01 `serving.Reader`.

- [ ] **Step 5 (3 min): Generate sqlc artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_identity.sql internal/store`

Expected: PASS; generated params retain all digest fields as fixed byte slices validated at the repository boundary, and exact second-generation drift/untracked output is empty.

- [ ] **Step 6 (5 min): Implement row conversion and the transaction-bound identity handler**

Use constant-time digest comparison for grant claims, require exactly one consumed row, expire elapsed grants, mark old-epoch pending issuance superseded inside the identity lock, and count active unexpired lineage leaves before activation. `ResolveRegisteredAuthorityEffectForUpdate` uses only caller DBTX and the opaque query, locks the exact anchor above in B01 order, validates `RegisteredKind`/`Expected` and returns its operation/epoch/sequence echo plus finite state. `CaptureActivationDecisionMaterial` runs outside DBTX, cryptographically validates the approved trusted-time/floor attestation, requires observed identity equal immutable configured identity, derives deadline and all policy inputs from the exact anchor, recomputes the commitment activation-input digest and returns no Provider/Head/checkpoint/evidence.

Fresh `ActivateAuthorityEffect` re-resolves under those locks, recomputes every grant/issuance/certificate/identity-transition typed input, validates the exact branded Receipt/proof, and atomically applies visibility plus the matching `create_|claim_|activation_|revoke_` proof group、resolution/audit/outbox/fence state. The identity transition anchor records both applied and not-applied terminal attempts and is unique by authority operation；it never invents an inventory-only outcome. `ValidatePersistedAuthorityEffect` accepts only parsed origin and one exact terminal anchor, recomputes all typed inputs/proof bindings and makes zero writes/provider/trusted-time/issuer calls. Unknown/multiple/nonterminal/cross-group/corrupt JCS or terminal columns fail closed. The handler never consumes admission；Coordinator consumes rollback-resistant proof after writes and none-time never consumes. Recovery primitives use only caller DBTX and never start a transaction. Never expose token digest、certificate DER or dynamic SQL/domain text in errors.

- [ ] **Step 7 (3 min): Run GREEN repository tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/identity' -Run '^TestPostgresIdentityRepository$' -Timeout 3m`

Expected: PASS with a copied serial/SAN certificate rejected unless every exact lookup component matches.

- [ ] **Step 8 (4 min): REFACTOR corrupt-row, handler-surface and retention predicates**

Reject invalid digest length、zero epoch、unknown status、oversized serial、inconsistent revoked timestamps, opaque-query echo mutation, trusted-time/floor signature/provider identity/expiry/deadline mutation, every typed activation-input mutation, and missing/partial/corrupt commitment/Head/checkpoint/reason/time/identity/evidence/resolution or terminal group. Test fresh activation changes no visibility on failure and parsed validation records zero writes/provider/time/issuer calls. Add compile/static tests proving there is no public `LookupExactCertificate` or independent readiness/serving query in `internal/nodecontrol/identity`, and that only the identity handler calls `LookupExactCertificateForActivation` with the Coordinator-owned DBTX. Add queries that delete only terminal grant/issuance/certificate/transition metadata older than 30 days and not referenced by active/recovery state.

Run: `go test ./internal/nodecontrol/identity ./internal/store -count=1`

Expected: PASS.

- [ ] **Step 9 (2 min): Commit identity persistence**

```bash
git add db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/identity/repository.go internal/nodecontrol/identity/postgres_repository.go internal/nodecontrol/identity/postgres_repository_test.go internal/nodecontrol/identity/postgres_repository_integration_test.go
git commit -m "feat(nodecontrol): persist exact node identities"
```

### Task 3: Implement enrollment, rotation, issuance recovery, and receipts

**Files:**
- Create: `internal/nodecontrol/identity/receipt.go`
- Create: `internal/nodecontrol/identity/service.go`
- Test: `internal/nodecontrol/identity/receipt_test.go`
- Test: `internal/nodecontrol/identity/service_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/identity/service_integration_test.go`
- Test: `internal/nodecontrol/identity/service_crash_test.go`

**Interfaces:**
- Consumes: Batch 01 `authority.Coordinator`, task 1 `NodeCertificateIssuer`, task 2 repository/transaction-bound identity effect handler, Batch 02 inventory/operator audit/outbox, cryptographic random reader and trusted time.
- Produces: one-time grant creation, claim/rotate/recover saga, `CertificateAuthorizationReceiptV1`, and pure response verifier consumed by B04. The service prepares effects and calls Finalize；it never owns a post-Finalize activation transaction.

```go
type CertificateAuthorizationReceiptV1 struct {
	SchemaVersion              string
	IssuanceID                 uuid.UUID
	AttemptID                  uuid.UUID
	NodeID                     uuid.UUID
	IdentityEpoch              uint64
	LineageID                  uuid.UUID
	IssuerID                   string
	CSRDER_SHA256              contracts.Digest
	LeafDER_SHA256             contracts.Digest
	PublicKeySHA256            contracts.Digest
	ControlPlaneAuthorityEpoch uint64
	AuthoritySequence          uint64
	Status                     CertificateStatus
}

type IdentityService interface {
	CreateEnrollmentGrant(context.Context, CreateGrantRequest) (CreateGrantResult, error)
	Claim(context.Context, ClaimRequest) (IssuanceResponse, error)
	Rotate(context.Context, RotateRequest) (IssuanceResponse, error)
	Recover(context.Context, uuid.UUID) (IssuanceResponse, error)
}

func VerifyAuthorizationReceipt(ReceiptVerificationRequest) (CertificateIdentity, error)
```

- [ ] **Step 1 (5 min): Write RED grant secrecy and retry tests**

Use a deterministic random reader to assert exactly 32 bytes are generated, only SHA-256 reaches repository arguments, plaintext appears only in the first successful result, the same idempotency retry returns resource metadata without plaintext, replacement requires a new If-Match and same CSR digest, and claimed/issuance-started grants cannot be replaced. For initial/recovery issuance, assert distinct `grant_claim` and `certificate_activate` operations/anchors；rotation uses only certificate activation, and a claim Receipt/proof cannot cross-route. Test each handler's configured trusted-time provider identity、floor signature/expiry、activation deadline and all typed inputs, plus every stored proof/terminal column mutation. Fresh activation must be atomic；Commit-loss recovery must parse the stored Head/preimages and call the zero-write validator without issuer/time/provider calls. After an intent exists, malformed/dependency/deadline/supersession creates a distinct final-not-applied commitment and follows `EffectCommitted → Finalize/Recover`；Abort is legal only before any domain anchor exists and after exact `EffectAbsent`.

- [ ] **Step 2 (2 min): Run the focused RED service test**

Run: `go test ./internal/nodecontrol/identity -run 'TestCreateEnrollmentGrant|TestIssueCertificateSaga' -count=1`

Expected: FAIL because identity service and receipt verifier do not exist.

- [ ] **Step 3 (5 min): Implement first-enrollment and security-admin grant creation**

Writer creation requires provisioning, identity epoch 0 and never-issued. Reenroll creation requires security-admin and a separately reserved `identity_epoch_advance` operation whose unique immutable terminal anchor is the authority-bound `node_state_transitions` row；only an applied terminal result advances epoch/lineage and revokes old state, while not-applied records an exact tombstone without changing inventory. Grant creation then derives grant ID and `EffectGrantCreate` operation ID from distinct fixed namespace UUIDs over `{node_id,idempotency_key,grant_kind,request_digest}`. Generate secret bytes only when no digest is durably bound, reserve before transaction, persist the exact grant `create_` commitment and expiry, then Finalize outside locks. Crash before the row can prove absence；after digest persistence it never regenerates/rediscloses secret. The registered handler stores the complete `create_` or identity-transition proof group with fence visibility/resolution/audit/outbox. Return plaintext only after exact atomic activation；response-loss retry returns metadata only.

- [ ] **Step 4 (5 min): Implement claim/rotation prepare transactions**

Validate IDs, nonce, CSR and request digest. Before any reservation or durable write, derive `issuance_id`, `claim_operation_id` and `certificate_operation_id` through three distinct fixed namespace UUID derivations over the canonical immutable tuple `{node_id,attempt_id,issuance_kind}` plus the literal role label；changed request fields under the same attempt conflict, and no random process-memory ID participates. Initial/recovery calls `Coordinator.Reserve(EffectGrantClaim)` and `Coordinator.Reserve(EffectCertificateActivate)` outside locks with those exact durable-reconstructible IDs, then locks node/grant. Atomically bind the grant consumption group to the claim operation and insert the immutable issuance bound to the separate certificate-activation operation、exact issuer、CSR/public-key/template digests and pending status. Commit that prepare transaction, then call `Coordinator.Finalize(claim_operation_id,claim_effect_digest)`；the identity handler resolves/audits only the exact one-time claim and never inserts or authorizes a certificate. Only after that exact claim result is committed may issuer work proceed. Rotation skips the claim operation/grant and derives/reserves only `EffectCertificateActivate`, while requiring exact active peer certificate、current epoch/lineage、no more than 4 active leaves、remaining lineage at least 36 hours and a new CSR key. Any partial two-reservation failure retries the same derived IDs, Inspect/Recover classifies both, and aborts only a reservation whose exact immutable effect absence is proven；process death before the prepare transaction therefore cannot strand an undiscoverable reservation.

- [ ] **Step 5 (5 min): Call issuer outside locks and validate independently**

Invoke exact `IssueRequest` by issuance ID; verify result echoes, parse single DER, verify chain/profile/template/SPKI/digests and reject any provider deviation. Persist validated bytes/digests in a short transaction. A retryable issuer timeout before the deadline keeps the exact pending operation recoverable. Once malformed/mismatched output、dependency failure、deadline expiry or supersession is determined after the issuance intent exists, a short fence-first transaction calls B01 `NewAuthorityEffectCommitment` and persists one immutable distinct `final_not_applied` commitment with finite reason and no certificate bytes/pointer. The resolver then returns `EffectCommitted`; the service must Finalize, never Abort. A late issuer result cannot replace the tombstone.

- [ ] **Step 6 (5 min): Finalize fence and activate certificate**

After either issuance commitment commits, call `Coordinator.Finalize(certificate_operation_id,exact_commitment_digest)` outside locks. The identity handler resolves only the `activation_` issuance anchor through the opaque query and returns authenticated material；Coordinator alone binds database point、committed Receipt、same-Provider Head/checkpoint and branded proof. Fresh activation re-locks/recomputes epoch、lineage、issuer/template/public key、node state、active-leaf cap、expiry/deadline and all typed inputs, then atomically inserts the certificate (success only), stores the complete issuance `activation_` proof group、terminal status、resolution/audit/outbox and fence visibility. Tombstone stores the same terminal proof group without certificate/pointer. Parsed recovery accepts only one exact terminal issuance after stored-preimage validation and invokes no issuer/time/provider. A claim operation/proof never routes here；the serving certificate always joins committed-active `certificate_activate`. There is no second service transaction. Recovery-prepared issuance alone may set `recovery_pending`; late results converge to the existing terminal outcome.

- [ ] **Step 7 (4 min): Implement exact receipt and retry response**

Build the receipt only from the active row and committed authority receipt. Same issuance/attempt/request returns identical leaf, chain and receipt until `NotAfter-5m`; after that return `ErrReenrollRequired`. Any changed field returns conflict. `VerifyAuthorizationReceipt` compares every ID/digest/authority field, exact leaf SPKI/profile/chain/time and a fresh private-key proof challenge.

- [ ] **Step 8 (3 min): Run GREEN identity saga tests**

Run: `go test ./internal/nodecontrol/identity -run 'TestCreateEnrollmentGrant|TestIssueCertificateSaga|TestVerifyAuthorizationReceipt' -count=1`

Expected: PASS; no database fake observes plaintext grant and no provider call occurs under lock.

- [ ] **Step 9 (5 min): Add crash recovery and rotation-overlap tests**

Crash after ID derivation, either initial/recovery reservation、grant consume/issuance intent、claim Finalize/activation、issuer response、final-not-applied commitment、validated result、either certificate provider Finalize、each certificate/tombstone activation write before Commit、Commit invocation before response loss and committed fence/domain visibility. Restart with only the caller's exact attempt tuple and prove it reconstructs byte-identical issuance/claim/certificate IDs, Inspect/Recover reaches every reservation, and the provider ends with zero orphan pending fence. Restore PostgreSQL at every prepared/tombstone/provider/activation cut while retaining provider Head and require the same terminal outcome. Prove no crash exposes only one side of each claim or certificate fence/domain/audit/outbox outcome, the two operation IDs never alias, and retry uses one issuance ID and at most one certificate. Rotate at 50% with bounded jitter input, keep old and new exact rows active, reject fifth overlap, revoke all epoch rows together, and reject late issuer results against an existing tombstone.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/identity' -Run '^(TestIdentityCrashRecovery|TestRotationOverlap)$' -Timeout 5m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/identity' -Run '^TestIdentityFinalNotAppliedPITR$' -Timeout 5m`

Expected: PASS.

- [ ] **Step 10 (4 min): REFACTOR failures into finite public results**

Use only invalid request, unauthenticated, forbidden, conflict, rate limited, dependency unavailable and internal mappings; provider, CSR, serial, DER, digest and SQL details remain private. Add a log-capture canary for grant/CSR/certificate bytes.

Run: `go test ./internal/nodecontrol/identity -count=1`

Expected: PASS.

- [ ] **Step 11 (2 min): Commit identity workflows**

```bash
git add internal/nodecontrol/identity/receipt.go internal/nodecontrol/identity/service.go internal/nodecontrol/identity/receipt_test.go internal/nodecontrol/identity/service_test.go internal/nodecontrol/identity/service_integration_test.go internal/nodecontrol/identity/service_crash_test.go
git commit -m "feat(nodecontrol): issue fenced node identities"
```

### Task 4: Verify host-deployed trust-bundle packages against finalized active high-water

**Files:**
- Create: `internal/nodecontrol/hostevidence/deployment_keys.go`
- Create: `internal/nodecontrol/hostevidence/trust_bundle.go`
- Modify: `db/queries/nodecontrol_identity.sql`
- Modify generated: `internal/store/nodecontrol_identity.sql.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/hostevidence/deployment_keys_test.go`
- Test: `internal/nodecontrol/hostevidence/trust_bundle_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/hostevidence/trust_bundle_integration_test.go`
- Create: `testdata/c12/trust-bundle-vectors.json`

**Interfaces:**
- Consumes: host-installed Ed25519 public keys, Batch 01 finalized active-only trust high-water table, task 1 CA profile validator, strict JCS.
- Produces: frozen deployment role registry, `TrustBundleManifestV1`, `TrustBundlePackageV1`, `VerifiedTrustBundle`, pure verifier and a read-only startup/readiness comparator. Per the approved amendments it produces no pending effect、publisher、resolver or activator；`trust_bundle_publish` remains fixed unsupported.

```go
type DeploymentRole string

const (
	RoleTrustBundle          DeploymentRole = "trust_bundle"
	RoleHostRemediation      DeploymentRole = "host_remediation"
	RoleOperatorTrustGuard   DeploymentRole = "operator_trust_guard"
	RoleNodeResourceEnvelope DeploymentRole = "node_resource_envelope"
)

type TrustPurpose string

const (
	PurposeBootstrapServer TrustPurpose = "bootstrap_server"
	PurposeAgentServer     TrustPurpose = "agent_server"
	PurposeOperatorServer  TrustPurpose = "operator_server"
	PurposeNodeClient      TrustPurpose = "node_client"
	PurposeOperatorClient  TrustPurpose = "operator_client"
)

type TrustBundleManifestV1 struct {
	SchemaVersion                    string
	Purpose                          TrustPurpose
	ControlPlaneAuthorityEpoch       uint64
	AuthoritySequence                uint64
	TrustDomain                      string
	BundleVersion                    uint64
	SortedCAEntries                  []CAEntryV1
	CumulativeDeauthorizedCADigests  []contracts.Digest
	IssuedAt                         time.Time
}

type TrustBundlePackageV1 struct {
	Manifest                     TrustBundleManifestV1
	CACertificatesDER            [][]byte
	DeploymentKeyID              string
	Algorithm                    string
	DeploymentAuthoritySignature []byte
}

type TrustBundleVerifier interface {
	Verify(context.Context, *TrustBundleHighWater, TrustBundlePackageV1) (VerifiedTrustBundle, error)
}
```

`CAEntryV1.Status` contains only `active`, `retiring`, `compromised`. `DeploymentAuthorityKeySetV1` requires at least four distinct Ed25519 keys, lowercase SHA-256 key IDs, exactly one role per key, and all four roles present.

- [ ] **Step 1 (5 min): Write RED deployment-key, package and active-high-water tests**

Test missing/duplicate roles, one key in two roles, wrong key ID, wrong signature role, cross-purpose/domain reuse, unsorted/duplicate entries, manifest/DER mismatch, non-CA DER, low version/sequence, same-value fork, illegal status transition, cumulative removal/reorder/re-authorization, 129th cumulative digest and trailing package bytes. Against pre-provisioned finalized fixtures, also test exact active match、active ahead、database rollback/fork、wrong listener mapping、missing committed fence and unavailable high-water. Add a static test proving this package has no Reserve/Finalize/pending/promotion API and cannot register `trust_bundle_publish` as supported.

- [ ] **Step 2 (2 min): Run the focused RED verifier tests**

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestDeploymentAuthorityKeySet|TestTrustBundleVectors' -count=1`

Expected: FAIL because hostevidence trust types and vectors do not exist.

- [ ] **Step 3 (4 min): Implement deployment key validation**

Derive each key ID from raw 32-byte Ed25519 public key, reject private/unknown algorithm bytes, require bytewise unique IDs and unique physical keys, and expose `KeyForRole(role, keyID)` without fallback to another role.

- [ ] **Step 4 (5 min): Implement package canonicalization and signature verification**

Require schema `trust-bundle-manifest.v1`, a closed purpose, canonical trust domain, sorted entries/cumulative set, SHA-256 DER matches one-to-one, and transcript `TALENRO-TRUST-BUNDLE-PACKAGE-V1\x00 || JCS(manifest)`. Verify with exact role `trust_bundle`; package/DER/signature cannot update deployment keys.

- [ ] **Step 5 (4 min): Enforce CA profiles and monotonic transitions**

Validate each CA independently for CA constraints and purpose; allow normal `active→retiring→omitted`, emergency `active|retiring→compromised`, append compromised/removed digest in the same version, forbid current entries in cumulative set and all later reauthorization. Compare version/sequence/digest with `contracts.CompareVersionedDigest` and reject rollback/fork.

- [ ] **Step 6 (4 min): Add the read-only finalized active-high-water query**

```sql
-- name: GetFinalizedTrustBundleHighWater :one
SELECT high_water.purpose, high_water.listener_kind, high_water.trust_domain,
       high_water.operation_id, high_water.authority_epoch,
       high_water.authority_sequence, high_water.bundle_version,
       high_water.bundle_digest, high_water.cumulative_set_digest,
       high_water.cumulative_count, high_water.updated_at
FROM nodecontrol.control_plane_trust_bundle_high_waters AS high_water
JOIN nodecontrol.control_plane_authority_fences AS fence
  ON fence.operation_id = high_water.operation_id
WHERE high_water.purpose = sqlc.arg(purpose)
  AND listener_kind = sqlc.arg(listener_kind)
  AND high_water.trust_domain = sqlc.arg(trust_domain)
  AND fence.provider_status = 'committed'
  AND fence.visibility_state = 'active';
```

Map each closed purpose to exactly one listener kind before the read. The repository exposes no INSERT/UPDATE/lock-for-publish method, and the comparator returns unavailable for absent、partial、uncommitted or fence-mismatched rows. Test fixtures may seed approved finalized rows through test-only migration SQL；that is not a production publisher or evidence for a future OOB protocol.

- [ ] **Step 7 (3 min): Generate store artifacts and run GREEN tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_identity.sql internal/store`

Expected: PASS with exact second-generation drift and untracked output empty.

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestDeploymentAuthorityKeySet|TestTrustBundleVectors' -count=1`

Expected: PASS and checked-in vector digests match lowercase hex exactly.

- [ ] **Step 8 (5 min): Test finalized active-high-water comparison and restart**

Using test-only pre-provisioned finalized snapshots, compare old、old+new and new 48-hour-overlap packages and a separately compromised old CA. Exact current package succeeds；lower/forked package、higher package not represented by the active row、old database snapshot、missing committed fence and wrong purpose/listener/domain all fail closed and keep listener readiness false. Restart re-reads the same finalized active row and never treats a locally supplied package as authority to advance it. Assert `NewEffectDispatcher` still receives the fixed unsupported `trust_bundle_publish` registration with both resolver and activator nil, rejects any non-nil or typed-nil handler there, and no hostevidence production code calls Reserve、Finalize or a trust high-water write.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/hostevidence' -Run '^TestTrustBundleHighWater$' -Timeout 3m`

Expected: PASS.

- [ ] **Step 9 (4 min): REFACTOR DER and manifest limits**

Cap CA entries and DER count at 16, each DER at 16 KiB, total package at 256 KiB and cumulative set at 128 before nested allocation. Errors contain only purpose and finite reason, not DER, signature or host path.

Run: `go test ./internal/nodecontrol/hostevidence -count=1`

Expected: PASS.

- [ ] **Step 10 (2 min): Commit trust-package support**

```bash
git add db/queries/nodecontrol_identity.sql internal/store/nodecontrol_identity.sql.go internal/store/querier.go internal/nodecontrol/hostevidence/deployment_keys.go internal/nodecontrol/hostevidence/trust_bundle.go internal/nodecontrol/hostevidence/deployment_keys_test.go internal/nodecontrol/hostevidence/trust_bundle_test.go internal/nodecontrol/hostevidence/trust_bundle_integration_test.go testdata/c12/trust-bundle-vectors.json
git commit -m "feat(nodecontrol): verify host trust packages"
```

### Task 5: Implement the rollback-resistant operator client trust guard

**Files:**
- Create: `internal/nodecontrol/hostevidence/operator_guard.go`
- Test: `internal/nodecontrol/hostevidence/operator_guard_test.go`
- Test: `internal/nodecontrol/hostevidence/operator_guard_race_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/hostevidence/operator_guard_integration_test.go`

**Interfaces:**
- Consumes: task 4 verified purpose `operator_server` package and deployment role `operator_trust_guard`, external rollback-resistant provider, crypto-random nonce source, trusted platform time, transport close callback.
- Produces: exact attestation DTO, provider boundary, connection lease and fail-closed transport invalidation.

```go
type OperatorTrustGuardProvider interface {
	Attest(context.Context, OperatorTrustGuardRequest) (OperatorTrustGuardAttestationV1, error)
}

type OperatorTrustGuardRequest struct {
	RequestNonce contracts.Nonce32
}

type OperatorTrustGuardAttestationV1 struct {
	SchemaVersion              string
	Purpose                    TrustPurpose
	TrustDomain                string
	ControlPlaneAuthorityEpoch uint64
	AuthoritySequence          uint64
	BundleVersion              uint64
	BundleDigest               contracts.Digest
	CumulativeSetDigest        contracts.Digest
	RequestNonce               contracts.Nonce32
	IssuedAt                   time.Time
	ValidUntil                 time.Time
	DeploymentKeyID            string
	Signature                  []byte
}

type OperatorClientTrustGuard interface {
	AuthorizeConnection(context.Context, VerifiedTrustBundle) (GuardLease, error)
	Invalidate(InvalidationReason)
	Close() error
}
```

- [ ] **Step 1 (5 min): Write RED nonce, rollback, and close tests**

Test a generated `contracts.Nonce32` fresh nonce, echoed nonce, replay, provider unavailable/restart, wrong purpose/domain/key role, lower epoch/sequence/version, same-value fork, 5-minute expiry, future issuance, local package rollback, emergency removal and existing transport closure. Add a cross-package compile/consumer test proving request and attestation use the closed nonce type rather than a semantically interchangeable digest.

- [ ] **Step 2 (2 min): Run the focused RED guard tests**

Run: `go test ./internal/nodecontrol/hostevidence -run TestOperatorClientTrustGuard -count=1`

Expected: FAIL because guard types do not exist.

- [ ] **Step 3 (5 min): Implement exact attestation verification**

Build transcript `TALENRO-OPERATOR-TRUST-GUARD-V1\x00 || JCS(unsigned_payload)`, verify only the exact `operator_trust_guard` role key, require purpose `operator_server`, exact local package trust domain/epoch/sequence/version/bundle/cumulative digests, matching nonce, trusted now within the window and `valid_until-issued_at <= 5m`.

- [ ] **Step 4 (4 min): Implement monotonic provider-backed leases**

Each new connection gets a new provider attestation; a lease expires at the earliest attestation expiry, server-leaf expiry or connected-at plus 30 minutes. Store no file-backed counter claim. Provider error, replay, rollback/fork or package install/seal error rejects new leases and invokes the close callback for all existing transports.

- [ ] **Step 5 (3 min): Run GREEN guard tests**

Run: `go test ./internal/nodecontrol/hostevidence -run TestOperatorClientTrustGuard -count=1`

Expected: PASS; provider failures close existing transports exactly once and no nonce is reused.

- [ ] **Step 6 (4 min): REFACTOR concurrent invalidation and shutdown**

Make `Invalidate` idempotent, prevent new leases after invalidation, avoid callback under the guard mutex, and ensure `Close` waits for callbacks. Run:

`go test -race ./internal/nodecontrol/hostevidence -run TestOperatorClientTrustGuard -count=1`

Expected: PASS with no deadlock, race or leaked lease.

- [ ] **Step 7 (5 min): Add provider conformance-shaped integration cases**

Use a process-isolated deterministic provider to persist its own high-water across restart; replay lower package, same-value fork and nonce after restart, then publish normal overlap/emergency removal. The fake test is an interface gate only and must not be labeled production evidence.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/hostevidence' -Run '^TestOperatorTrustGuardProviderContract$' -Timeout 3m`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit operator guard support**

```bash
git add internal/nodecontrol/hostevidence/operator_guard.go internal/nodecontrol/hostevidence/operator_guard_test.go internal/nodecontrol/hostevidence/operator_guard_race_test.go internal/nodecontrol/hostevidence/operator_guard_integration_test.go
git commit -m "feat(nodecontrol): guard operator client trust"
```

### Task 6: Authorize exact node and operator peers at request and response boundaries

**Files:**
- Create: `internal/nodecontrol/identity/security_fault_binding.go`
- Create: `internal/nodecontrol/identity/trust_conflict_binding.go`
- Create: `internal/nodeagentapi/authorizer.go`
- Create: `internal/nodeagentapi/waiters.go`
- Create: `internal/nodeoperatorapi/authorizer.go`
- Test: `internal/nodeagentapi/authorizer_test.go`
- Test: `internal/nodeagentapi/waiters_test.go`
- Test: `internal/nodeoperatorapi/authorizer_test.go`
- Test (first line `//go:build integration`): `internal/nodeagentapi/authorizer_integration_test.go`

**Interfaces:**
- Consumes: task 1 strict leaf identities, Batch 01 bound `serving.Reader` constructed from the same Coordinator/PostgresRepository as production authority, Batch 02 `operator.OperatorAuthorizer`, trusted time, inventory/recovery/session state and trust-bundle high-water. It does not consume the task 2 identity repository.
- Produces: closed endpoint authorization matrix, immutable authorization snapshots, ordinary response-commit recheck, specialized self-invalidating security-fault receipt commit gate, operator credential extraction and targeted waiter cancellation.

```go
type EndpointClass string

const (
	EndpointRotateCertificate    EndpointClass = "rotate_certificate"
	EndpointDesiredPoll          EndpointClass = "desired_poll"
	EndpointRecoveryPoll         EndpointClass = "recovery_poll"
	EndpointObservation          EndpointClass = "observation"
	EndpointSecurityFault        EndpointClass = "security_fault"
	EndpointTrustConflictEvidence EndpointClass = "trust_conflict_evidence"
	EndpointRecoveryAttestation  EndpointClass = "recovery_attestation"
)

type NodeRequestAuthorizer interface {
	AuthorizeRequest(context.Context, *x509.Certificate, EndpointClass) (NodeAuthorization, error)
	ReauthorizeResponse(context.Context, NodeAuthorization, EndpointClass) error
	AuthorizeSecurityFaultReceiptCommit(context.Context, NodeAuthorization, identity.SecurityFaultCommitBinding) error
	AuthorizeHighWaterConflictCommit(context.Context, NodeAuthorization, identity.HighWaterConflictCommitBinding) error
	AuthorizeTrustConflictAckCommit(context.Context, NodeAuthorization, identity.TrustConflictAckCommitBinding) error
}

type NodeAuthorization struct {
	NodeID                 uuid.UUID
	CertificateID          uuid.UUID
	IssuerID               string
	SerialBytes            []byte
	LeafDERDigest          contracts.Digest
	PublicKeyDigest        contracts.Digest
	IdentityEpoch          uint64
	AuthorityEpoch         uint64
	InventoryVersion       uint64
	AuthorizationVersion   uint64
	CertificateNotAfter    time.Time
	ServerCABundleHighWater contracts.VersionedDigest
}

type NodeCheckpointReader interface {
	CommittedNodeCheckpoint(context.Context, contracts.Digest) (authority.NodeCheckpoint, error)
}

type OperatorPeerAuthenticator interface {
	Authenticate(context.Context, *x509.Certificate) (operator.Credential, error)
}
```

`identity.SecurityFaultCommitBinding` contains operation ID, original exact certificate ID/DER/key/issuer/serial/identity epoch, local fault ID, request/effect digests, incident ID, authority epoch/sequence and committed receipt digest. `identity.HighWaterConflictCommitBinding` binds that same immutable credential snapshot to the exact desired/recovery-poll request/high-water tuple, fence-finalized unverified incident ID, the B01-helper-derived trust-conflict local fault ID, operation/effect/terminal receipt and the one fixed bodyless public `conflict` response plus exact `Talenro-Trust-Conflict-Incident-ID` header value. `identity.TrustConflictAckCommitBinding` binds the exact generated evidence request digest、incident、submitted artifact-set digest、credential snapshot、finite validation/escalation result and the generated `TrustConflictEvidenceAckV1` digest. All three are B03-owned acyclic boundary types；their validators permit serialization of only their respective fixed receipt/error-header/ACK and cannot authorize root、metadata、desired、recovery、time or another endpoint payload. B03 does not add these DTOs to the B01 `contracts` package or schema registry.

- [ ] **Step 1 (5 min): Write the RED endpoint authorization matrix**

Test every certificate status against every endpoint class and node state. Ordinary rotate/desired/observation require active, current epoch, operator enabled/draining and security normal. Recovery poll/attestation require active/recovery-pending/recovery-limited, operator disabled and exact nonterminal recovery session or incident. Security fault accepts those same three current-epoch statuses only as a tightening mutation. For an exact open trust-conflict incident, the exact current-epoch credential that triggered it remains usable while the node is disabled/quarantined for one closed incident-bound endpoint set only: `trust_conflict_evidence`, `recovery_poll`, and `recovery_attestation`. Recovery poll/attestation additionally require that incident's exact nonterminal recovery aggregate、nonce/snapshot predicates；for this trust-conflict row the generated poll/attestation `recovery_id` and the aggregate's recovery locator are byte-for-byte the retained incident UUID, never an independently minted session UUID or caller-provided scalar. The exception grants no rotate、desired、observation、ordinary report、LKG or unrelated recovery access. Unknown classes and revoked/expired/mismatched rows always deny. Separately prove ordinary `ReauthorizeResponse` rejects each self-invalidated desired-poll/security-fault/evidence snapshot while the three specialized commit gates accept only their exact fence-finalized immutable binding and fixed output. Mutate certificate/incident/header UUID/request/artifact-set/operation/effect/receipt/ACK digest, omit/duplicate the specialized header, attach it to an ordinary conflict or swap either commit-binding type and require rejection with zero response bytes. Add a matrix row proving the same certificate can poll the signed clear snapshot and submit its exact stopped attestation after the conflict response, including after transport reconnect, but cannot use any fourth route.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestNodeAuthorizationMatrix|TestOperatorPeerAuthentication' -count=1`

Expected: FAIL because API authorizers do not exist.

- [ ] **Step 3 (5 min): Implement exact node peer extraction through the bound reader**

Require one verified leaf, strict node profile and canonical URI SAN；derive unsigned serial bytes, exact DER/SPKI digests and node ID. Build the exact `serving.CertificateLookup` with the constructor-fixed `agent_server` purpose/listener/configured trust-domain selector and call `Reader.GetAuthorizedNodeCertificate`, which performs readiness→one certificate+server-CA-high-water callback→readiness on one physical connection and returns no facts unless the same authority Head remains ready. Compare every returned certificate/inventory/trust field to the peer, configured listener, trusted time and endpoint's finite allowed statuses, then defensively copy the nonzero server-CA `VersionedDigest` into `NodeAuthorization`；no lookup by serial、SAN、CA or caller-selected trust selector alone. The authorizer cannot import task 2 repository/query code or open its own certificate/trust read transaction.

- [ ] **Step 4 (4 min): Implement operator credential extraction**

Require strict operator profile and canonical operator URI; produce `operator.Credential` containing issuer ID, serial bytes, exact leaf DER/public-key digests, URI SAN, operator ID, authority epoch and credential version. Call external authorizer for every request; security-admin actions bypass cache according to B02.

- [ ] **Step 5 (4 min): Implement final response reauthorization**

`ReauthorizeResponse` calls the same bound `Reader.GetAuthorizedNodeCertificate` again and rechecks the exact certificate row/status、identity/authority epoch、node state、endpoint-specific recovery predicate、trust high-water and trusted time against the entry snapshot. It never opens an independent certificate/readiness transaction and never accepts a repository fallback when the Reader reports authority unavailable. Any change returns only unauthenticated/forbidden and the handler must discard prepared payload before writing headers. The specialized high-water-conflict gate may commit only the fixed no-payload public `conflict` after verifying the exact fence-finalized unverified incident binding；the trust-conflict ACK gate may commit only the exact generated ACK after rechecking the submission/escalation binding. Neither gate reauthorizes the credential for poll、rotate、observation or recovery, and response loss exact-retries the same binding.

- [ ] **Step 6 (3 min): Run GREEN matrix tests**

Run: `go test ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestNodeAuthorizationMatrix|TestOperatorPeerAuthentication|TestResponseReauthorization' -count=1`

Expected: PASS; a copied serial/SAN with different DER or key is rejected.

- [ ] **Step 7 (4 min): Add node/issuer/certificate waiter cancellation**

`WaiterRegistry.Register(nodeID, certificateID, cancel)` returns an idempotent unregister function. Revoke, disable, identity-epoch change, quarantine, recovery-session completion and trust-bundle removal cancel matching waiters without holding registry locks while invoking callbacks.

Run: `go test -race ./internal/nodeagentapi -run TestWaiterRegistry -count=1`

Expected: PASS with no missed cancel, double callback or leak.

- [ ] **Step 8 (5 min): Prove keep-alive reauthorization against PostgreSQL**

Construct the Reader from the exact Coordinator/PostgresRepository pair, authorize once, revoke or increment identity epoch in a second transaction, then reuse the same authorization snapshot for commit. Repeat for certificate expiry、CA compromise、forced source/repository mismatch and an authority Head change between the Reader's two readiness checks. Add a compile/static assertion that `internal/nodeagentapi` has no identity-repository or raw certificate SQL dependency.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodeagentapi' -Run '^TestKeepAliveAuthorizationRecheck$' -Timeout 3m`

Expected: PASS; every stale snapshot is rejected and contains no prepared state payload.

- [ ] **Step 9 (2 min): Commit request authorization**

```bash
git add internal/nodecontrol/identity/security_fault_binding.go internal/nodecontrol/identity/trust_conflict_binding.go internal/nodeagentapi/authorizer.go internal/nodeagentapi/waiters.go internal/nodeoperatorapi/authorizer.go internal/nodeagentapi/authorizer_test.go internal/nodeagentapi/waiters_test.go internal/nodeoperatorapi/authorizer_test.go internal/nodeagentapi/authorizer_integration_test.go
git commit -m "feat(nodecontrol): reauthorize exact mTLS peers"
```

### Task 7: Implement bounded security incidents, recovery sessions, host remediation, and restore reauthorization

**Files:**
- Create: `db/queries/nodecontrol_recovery.sql`
- Create generated: `internal/store/nodecontrol_recovery.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Create: `internal/nodecontrol/hostevidence/remediation.go`
- Create: `internal/nodecontrol/recovery/types.go`
- Create: `internal/nodecontrol/recovery/repository.go`
- Create: `internal/nodecontrol/recovery/postgres_repository.go`
- Create: `internal/nodecontrol/recovery/security_fault.go`
- Create: `internal/nodecontrol/recovery/trust_conflict.go`
- Create: `internal/nodecontrol/recovery/operator_actions.go`
- Modify: `internal/nodecontrol/identity/repository.go`
- Modify: `internal/nodecontrol/identity/postgres_repository.go`
- Modify: `internal/nodecontrol/identity/service.go`
- Test: `internal/nodecontrol/hostevidence/remediation_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/recovery/postgres_repository_integration_test.go`
- Test: `internal/nodecontrol/recovery/security_fault_test.go`
- Test: `internal/nodecontrol/recovery/trust_conflict_test.go`
- Test: `internal/nodecontrol/recovery/operator_actions_test.go`
- Test: `internal/nodecontrol/recovery/service_crash_test.go`
- Create: `testdata/c12/host-remediation-vectors.json`

**Interfaces:**
- Consumes: task 4 `DeploymentAuthorityKeySetV1` and exact `RoleHostRemediation`; task 2 transaction-bound identity handler/`RecoveryIdentityRepository`; task 3 recovery issuance preparation; task 6 node/operator authorization bindings; Batch 01 `authority.Coordinator`、transaction-bound effect contracts、recovery schema and `contracts.DeriveTrustConflictLocalFaultIDV1`; Batch 02 inventory locks、registered state handler/recovery signer saga、`RecoveryTransitionGuard`、atomic audit/outbox and uncached `operator.OperatorAuthorizer`; trusted time.
- Produces: canonical `HostRemediationAction`/`HostRemediationEvidenceV1`/`HostRemediationVerifier`; internal `VerifiedRecoveryAttestation`; field-by-field generated node-agent/operator response construction including exact recovery-poll 200/204；bounded incident/session/evidence/approval repository；`TrustConflictService` with a separate bounded low-priority verifier queue；one transaction-bound `RecoveryEffectHandler` registered only for security-incident open/resolve and standalone operator-transition；`recovery.Service`; and the B02 transaction-bound recovery transition guard implementation.

```go
type DeliverableSecurityFaultReceipt struct {
	SchemaVersion     string
	LocalFaultID      uuid.UUID
	IncidentID        uuid.UUID
	RequestDigest     contracts.Digest
	AuthorityEpoch    uint64
	AuthoritySequence uint64
	Result            string // exactly "accepted"
}

type VerifiedRecoveryAttestation struct {
	RecoveryID               uuid.UUID
	Nonce                    contracts.Nonce32
	RecoverySnapshotDigest   contracts.Digest
	RecoveryGeneration       uint64
	CertificateID            uuid.UUID
	ProofOfPossession        []byte
	AgentBuildDigest         contracts.Digest
	SupervisorBuildDigest    contracts.Digest
	AgentGuardDigest         contracts.Digest
	SupervisorGuardDigest    contracts.Digest
	LatchGuardDigest         contracts.Digest
	TrustedTimeEvidenceDigest contracts.Digest
	AllSlotsStopped          bool
	AttestationDigest        contracts.Digest
}

type GuardCounterEvidenceV1 struct {
	CounterIdentity string
	CounterValue    uint64
	StateDigest     contracts.Digest
}

type HostRemediationAction string

const (
	HostRemediationRegisterIncident         HostRemediationAction = "register_host_security_incident"
	HostRemediationCompleteReenrollment     HostRemediationAction = "complete_reenrollment"
	HostRemediationClearSecurityLatches     HostRemediationAction = "clear_security_latches"
	HostRemediationReauthorizeAfterRestore  HostRemediationAction = "reauthorize_after_restore"
	HostRemediationRegisterResourceEnvelope HostRemediationAction = "register_resource_envelope"
)

type HostRemediationEvidenceV1 struct {
	SchemaVersion           string
	EvidenceID              uuid.UUID
	NodeID                  uuid.UUID
	IncidentID              uuid.UUID
	Action                  HostRemediationAction
	AgentBuildDigest        contracts.Digest
	SupervisorBuildDigest   contracts.Digest
	AgentRollbackGuard      GuardCounterEvidenceV1
	SupervisorRollbackState GuardCounterEvidenceV1
	SecurityLatchGuard      GuardCounterEvidenceV1
	TrustedTimeProviderID   string
	TrustedTimeFloor        time.Time
	InstalledMapVersion     uint64
	InstalledMapDigest      contracts.Digest
	CompletedAt             time.Time
	DeploymentKeyID         string
	Algorithm               string
	DeploymentAuthoritySig  []byte
}

type HostRemediationVerifier interface {
	Verify(context.Context, HostRemediationEvidenceV1) (VerifiedHostRemediation, error)
}

type Service interface {
	ReportSecurityFault(context.Context, NodeCredentialBinding, nodeagentv1.SecurityFaultReportV1) (nodeagentv1.SecurityFaultReceiptV1, error)
	PollRecovery(context.Context, NodeCredentialBinding, nodeagentv1.RecoveryPollRequestV1) (nodeagentv1.NodeRecoveryPollResponseV1, bool, error)
	SubmitRecoveryAttestation(context.Context, NodeCredentialBinding, nodeagentv1.RecoveryAttestationV1) (nodeagentv1.RecoveryAttestationAckV1, error)
	Disable(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.DisableNodeV1) (nodeoperatorv1.RecoveryOperationV1, error)
	Reenroll(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.ReenrollNodeV1) (nodeoperatorv1.RecoveryEnrollmentGrantV1, error)
	CompleteReenrollment(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.CompleteReenrollmentV1) (nodeoperatorv1.RecoveryOperationV1, error)
	RegisterHostSecurityIncident(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.RegisterHostSecurityIncidentV1) (nodeoperatorv1.RecoveryOperationV1, error)
	ClearSecurityQuarantine(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.ClearSecurityQuarantineV1) (nodeoperatorv1.RecoveryOperationV1, error)
	ResumeAfterSecurity(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.ResumeAfterSecurityV1) (nodeoperatorv1.SigningOperationV1, error)
	ReauthorizeAfterRestore(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.ReauthorizeAfterRestoreV1) (nodeoperatorv1.RestoreReauthorizationV1, error)
}

type HighWaterConflictKind string

const (
	HighWaterFork  HighWaterConflictKind = "unverified_client_highwater_conflict"
	HighWaterAhead HighWaterConflictKind = "client_highwater_ahead"
)

type HighWaterStreamKind string

const (
	HighWaterStreamServerCA       HighWaterStreamKind = "server_ca_bundle"
	HighWaterStreamRoot           HighWaterStreamKind = "node_state_root_set"
	HighWaterStreamMetadata       HighWaterStreamKind = "node_state_trust_metadata"
	HighWaterStreamDesired        HighWaterStreamKind = "desired_state"
	HighWaterStreamRecovery       HighWaterStreamKind = "recovery_state"
)

type HighWaterConflictInput struct {
	Kind                  HighWaterConflictKind
	Stream                HighWaterStreamKind
	RequestDigest         contracts.Digest
	ClientVersion         uint64
	ClientAuthoritySequence uint64
	ClientDigest          contracts.Digest
	ServerVersion         uint64
	ServerAuthoritySequence uint64
	ServerDigest          contracts.Digest
}

type TrustConflictService interface {
	MaterializeHighWaterConflict(context.Context, NodeCredentialBinding, HighWaterConflictInput) (identity.HighWaterConflictCommitBinding, error)
	SubmitEvidence(context.Context, NodeCredentialBinding, nodeagentv1.TrustConflictEvidenceRequestV1) (nodeagentv1.TrustConflictEvidenceAckV1, identity.TrustConflictAckCommitBinding, error)
}

type RecoveryEffectHandler interface {
	authority.RegisteredEffectResolver
	authority.RegisteredEffectActivator
}
```

`HighWaterStreamKind` is exactly the five artifact-backed streams above. `node_authority_checkpoint` is deliberately not a member: the node request carries only its scalar checkpoint sequence, not the provider receipt preimage/digest needed by `HighWaterConflictInput`, and no checkpoint artifact may be invented. Each materialized incident freezes one stream and accepts evidence only of that stream's one-to-one generated `ConflictArtifactKind` (`server_ca_bundle`、`node_state_root_set`、`node_state_trust_metadata`、`desired_state` or `recovery_state`)；a different valid artifact kind cannot satisfy or escalate it. A client checkpoint ahead of the DB/provider-consistent checkpoint is a headerless global authority/readiness conflict handled through Coordinator checkpoint/Head/Inspect recovery；it never calls `MaterializeHighWaterConflict`, creates a per-node incident/notice/local latch or asks the node for unavailable checkpoint evidence.

`NodeCredentialBinding` contains the task 6 exact certificate ID/DER/key/issuer/serial, node/identity/authority epochs and authorization version. `operator.AuthorizedMutationBinding` contains the exact uncached B02 credential/authorization、command/idempotency/If-Match digests、target scope and trusted authorization time；recovery defines no wire-compatible shadow. Every request/response DTO in `Service` and `TrustConflictService` is explicitly from the Plan 01 generated package. `PollRecovery` returns `changed=true` only with the complete generated `nodeagentv1.NodeRecoveryPollResponseV1` for HTTP 200；`changed=false` requires its zero value and maps only to bodyless 204. Conflict/error returns no response. The service strict-converts a generated attestation into non-wire-compatible `VerifiedRecoveryAttestation`; after fence finalization it maps `DeliverableSecurityFaultReceipt` field-by-field into `nodeagentv1.SecurityFaultReceiptV1`, never by alias/cast/reflection. An external-package compile test imports both generated packages, assigns every method expression—including the exact three-result poll and two trust-conflict methods—to its signature and statically rejects recovery-local declarations named like generated request/response DTOs. `VerifiedHostRemediation` owns defensive copies of the unsigned evidence, its RFC 8785 JCS digest and verified key/role；only the repository can consume that digest once for its exact node/incident/action. The five action literals are closed and map one-to-one, in declaration order, to `RegisterHostSecurityIncident`、`CompleteReenrollment`、`ClearSecurityQuarantine`、`ReauthorizeAfterRestore` and Task 7B `ResourceEnvelopeService.Register`; empty/unknown/alias/case and every cross-action pairing reject. The recovery handler registers only `security_incident_open`、`security_incident_resolve` and `operator_transition`：open and resolve use independently unique `node_security_incidents.open_`/`resolve_` groups, while standalone operator transition uses one immutable `node_state_transitions` row with `authority_effect_kind=operator_transition` and explicit disposition. Every security-fault、overflow、unverified and typed escalation path maps one operation to exactly one incident anchor；subtype reconciliation never creates a second valid row. Identity-epoch belongs exclusively to task 2 and desired/recovery to B02, so registration never overlaps.

- [ ] **Step 1 (5 min): Write RED host-remediation transcript and freshness vectors**

Generate `testdata/c12/host-remediation-vectors.json` for transcript `TALENRO-HOST-REMEDIATION-EVIDENCE-V1\x00 || JCS(unsigned_evidence)`. Cover every field above, all five literal actions, canonical UUID/time/digest encoding, wrong role/key/algorithm, signature mutation, reordered/duplicate input, node/incident/action mismatch, future completion, exactly 15 minutes, one nanosecond beyond 15 minutes, trusted-time provider/floor rollback and evidence-ID replay. Add the complete 5×5 consumer/action matrix: only the declaration-order diagonal succeeds；empty、case-folded、alias and unknown actions fail before repository/provider use.

Run: `go test ./internal/nodecontrol/hostevidence -run TestHostRemediationEvidenceVectors -count=1`

Expected: FAIL because the verifier and vectors do not exist.

- [ ] **Step 2 (5 min): Implement the pure role-separated verifier**

Strict-decode one schema version and one of the five exact action literals, require nonzero IDs/counters/map version, exact 32-byte digests, nonempty bounded provider/counter identities, `Algorithm=Ed25519`, and the configured exact `host_remediation` deployment key. Reconstruct JCS from typed fields, verify the domain-separated signature, return defensive copies plus the evidence digest/action, and never consult PostgreSQL or accept a `trust_bundle`, `operator_trust_guard`, or `node_resource_envelope` role key.

Run: `go test ./internal/nodecontrol/hostevidence -run TestHostRemediationEvidenceVectors -count=1`

Expected: PASS; every one-field mutation and cross-role signature is rejected.

- [ ] **Step 3 (5 min): Write RED repository bound and immutability tests**

Cover fault subtype slots 1–12, rejection of reserved slots 13–15, saturating overflow slot 16, 16/17 simultaneous incident rows, 64/65 local-fault bindings, 16/17 supervisor-fault bindings, occurrence saturation, bounded aggregation of `unverified_client_highwater_conflict`/`client_highwater_ahead`, immutable finalized security receipt, one nonterminal recovery session per node, one accepted attestation digest per session version/nonce, one-time remediation evidence/action, and proposal/approval uniqueness by exact operator credential and role. Prove `open -> resolution_pending_agent_ack -> resolved` and `overflow -> resolution_pending_agent_ack|resolved` are the only forward incident transitions；typed trust escalation atomically resolves the exact unverified incident and opens/reconciles exactly one mapped typed incident without mutating its immutable subtype. For all three effect kinds, mutate opaque-query echo、operation/kind/scope/epoch/sequence/effect digest、trusted-time/floor signature/provider identity/expiry/deadline、every typed activation input、every stored proof preimage and terminal column；fresh activation changes no visibility on failure, and parsed validation accepts only one terminal anchor with zero writes/provider/time calls.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/recovery' -Run '^TestPostgresRecoveryBounds$' -Timeout 3m`

Expected: FAIL because recovery queries and repository do not exist.

- [ ] **Step 4 (5 min): Implement exact recovery SQL, conversions and transaction-bound handler**

Add lock/upsert queries for the node recovery aggregate, typed/overflow/unverified incident, local/supervisor binding, security receipt intent, recovery session, attestation, consumed remediation evidence, restore proposal/approval and immutable operator-transition anchor. Every write receives caller DBTX；conversions reject invalid versions/enums/digests、partial authority/proof groups、expired evidence and corrupt terminal rows. `ResolveRegisteredAuthorityEffectForUpdate` uses the opaque query and B01 lock order, selects only the exact incident `open_`/`resolve_` group or operator transition row, and returns a validated operation/epoch/sequence echo plus finite state. `CaptureActivationDecisionMaterial` runs lock-free, authenticates configured rollback-resistant trusted-time/floor identity and deadline/input binding, and returns no Provider/Head/checkpoint/evidence.

Fresh activation recomputes every incident subtype/binding、receipt deliverability、operator transition/session/approval and typed policy input, then atomically writes the exact terminal anchor/proof group、incident materialization/escalation/resolution or transition、resolution/audit/outbox/fence visibility；it never consumes admission. Parsed `ValidatePersistedAuthorityEffect` accepts only one exact terminal row/group, recomputes the typed digest and proof binding, and makes zero writes/provider/time/authorizer calls. Unknown/multiple/nonterminal/cross-operation/corrupt rows fail closed. It never registers desired/recovery or identity kinds. Use database uniqueness/check constraints rather than preflight-only counting；trust-conflict processing adds no table/effect kind or unbounded raw artifact.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_recovery.sql internal/store/nodecontrol_recovery.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_recovery.sql internal/store/nodecontrol_recovery.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_recovery.sql internal/store`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/recovery' -Run '^TestPostgresRecoveryBounds$' -Timeout 3m`

Expected: PASS; generated recovery sqlc parameters preserve every UUID, authority version, digest and nullable transition group without lossy strings, with exact second-generation drift/untracked output empty.

- [ ] **Step 5 (5 min): Write the RED self-invalidating security-fault saga**

Start with an exact active current-epoch certificate. Assert `ReportSecurityFault` binds operation/local-fault/body/evidence/credential/effect digests, immediately quarantines and disables the node, saves normalized resume intent, increments identity epoch where required, revokes the epoch, supersedes pending issuance, invalidates grants, creates/aggregates the incident and stores a nondeliverable receipt. Prove ordinary response reauthorization now fails while only task 6's specialized commit gate can emit the exact finalized receipt.

Run: `go test ./internal/nodecontrol/recovery -run TestSecurityFaultSelfInvalidatingSaga -count=1`

Expected: FAIL because the saga is absent.

- [ ] **Step 6 (5 min): Implement reserve, fail-closed commit, finalize, and receipt activation**

Call `authority.Coordinator.Reserve` for the Batch 01 security-fault effect before opening the transaction. In the first short transaction re-lock the exact credential/node, record the reservation, idempotently materialize the binding and all intentionally fail-closed mutations, write their audit/outbox, and persist the immutable nondeliverable receipt intent/effect. After commit call `authority.Coordinator.Finalize(operation_id,effect_digest)` outside all locks；the recovery handler supplies the exact immutable incident/receipt effect for update, and only the Coordinator captures/binds database coordinates、finalizes the provider、captures decision evidence and opens the activation transaction. In that same transaction the handler requires the unchanged operation/credential/effect/incident/request binding and exact receipt/evidence, then atomically marks only that receipt deliverable with fence visibility/resolution/audit/outbox. After the handler returns, only the Coordinator consumes the admission token and immediately attempts Commit. There is no recovery-service final transaction. Timeout or uncertainty leaves the node quarantined and recovers through `Coordinator.Recover(operation_id)`；it never returns a guessed ACK.

Run: `go test ./internal/nodecontrol/recovery -run TestSecurityFaultSelfInvalidatingSaga -count=1`

Expected: PASS; the returned object is only the exact generated `nodeagentv1.SecurityFaultReceiptV1{result:"accepted"}` assembled from a deliverable domain receipt and contains no state/root/time payload.

- [ ] **Step 7 (5 min): Write RED stopped-recovery, generated poll, trust-conflict, and identity/admin flow tests**

Test recovery poll with exact recovery ID, sorted unique known incident IDs, high-water and 32-byte nonce: a strict subset returns the complete newer set, an unknown extra ID or same generation/sequence with another digest conflicts, and administrative-disable/authority-restore may have an empty incident set. For a trust-conflict recovery, assert the request、signed snapshot、attestation and repository aggregate all use the retained incident UUID as that exact recovery ID；a fresh recovery/session UUID, caller-selected scalar or another incident UUID fails before snapshot/attestation use. Assert field-by-field construction of `nodeagentv1.NodeRecoveryPollResponseV1`, `changed=true` only for 200, and zero response plus `changed=false` only for 204. Test proof-of-possession over the fresh server nonce and exact recovery certificate, all-slots-stopped, guard/build/time digests, attestation retry, disable/retire resume normalization, reenroll epoch/lineage/grant creation, `recovery_pending` activation, no-old-pending-issuance, and completion that remains operator-disabled.

Add high-water fork/ahead tests for desired/root/metadata/server-CA and optional recovery streams. Same version/authority-sequence plus another digest or an unexplained client-ahead tuple must derive one immutable operation ID, use the recovery handler's `security_incident_open` path, disable/quarantine only that node while preserving its exact current-epoch credential solely for the incident-bound closed set `trust_conflict_evidence|recovery_poll|recovery_attestation`, and return only the specialized fixed conflict binding；999 other nodes remain serviceable and no client claim mutates a global high-water or readiness. Require the submitted evidence kind to equal the incident's original stream and reject every cross-stream substitution. Independently exercise the provider/DB check and scalar node-checkpoint-ahead case: only a real mismatch signals the existing listener readiness failure path, returns no incident header and creates no notice/local-fault binding.

For generated `TrustConflictEvidenceRequestV1`, cover the declaration-order closed artifact kinds `desired_state`、`recovery_state`、`node_state_trust_metadata`、`node_state_root_set`、`server_ca_trust_bundle`, plus zero/one/two/three artifacts, 1 MiB aggregate boundary, compression rejection at the handler, canonical byte truncation, unknown/empty/alias/case kind, unknown key/schema/role, signature/digest/epoch/sequence/version mutation, cross-kind pairs, one valid artifact without a second independent artifact, one submitted artifact plus the exact server fence-finalized counterpart, and two submitted valid conflicting artifacts. Freeze the only escalation map: `desired_state|recovery_state` → `online_signer_equivocation`; `node_state_trust_metadata|node_state_root_set` → `root_equivocation`; `server_ca_trust_bundle` → `server_trust_bundle_conflict`. Require same kind、authority identity and version/sequence but different digests；a cross-kind pair never escalates. The generated ACK result is closed to `accepted|escalated`: every valid but insufficient set returns `accepted`, and only the atomic typed-incident escalation returns `escalated`; invalid input returns no ACK. Test per node/exact credential four-per-hour limiting and a separate queue of exactly 16 one-MiB jobs with two workers；the 17th fails before copying bytes or calling a verifier/provider. Queue timeout/crash/retry cannot double-escalate, and every insufficient/invalid set leaves the unverified incident open.

Run: `go test ./internal/nodecontrol/recovery -run 'TestStoppedRecovery|TestRecoveryPollGeneratedMapping|TestRecoveryAttestation|TestRecoveryIdentityFlow|TestTrustConflictHighWaterIsolation|TestTrustConflictIncidentRecoveryEndpointsRemainReachable|TestTrustConflictEvidenceEscalation|TestTrustConflictQueueBounds' -count=1`

Expected: FAIL because recovery actions and the B02 guard are absent.

- [ ] **Step 8 (5 min): Implement the stopped recovery and trust-conflict state machines**

Implement identity-compromise, administrative-disable, retire, authority-restore and security-incident sessions. `Disable` performs the fail-closed identity transaction and creates a nonterminal session; retire and authority restore force saved resume state to disabled. `Reenroll` creates a new epoch/lineage CSR-bound one-time grant while disabled. `PollRecovery` signs only bounded `RecoveryStateSnapshotV1`, maps it field-by-field into the generated response and returns the exact 200/204 boolean contract；`SubmitRecoveryAttestation` verifies exact nonce/snapshot/certificate proof and stores one digest. `CompleteReenrollment` requires fresh accepted remediation evidence with action=`complete_reenrollment`, the exact recovery-pending certificate and no old pending issuance, activates it as active or recovery-limited according to the remaining incident set, completes only the exact session, and never resumes ordinary operator state.

`TrustConflictService.MaterializeHighWaterConflict` strict-validates the handler-computed client/server tuple, derives the operation ID from the immutable node/certificate/request/stream/high-water digest, calls `Coordinator.Reserve` before its short transaction, then creates/aggregates the exact unverified subtype and quarantines/disables only that node while leaving the exact current-epoch certificate usable only for the incident-bound closed set `trust_conflict_evidence|recovery_poll|recovery_attestation`. In the same activation transaction, after the incident ID is fixed, it recomputes `TrustConflictLocalFaultBindingV1` from the exact node、incident、poll-request digest、certificate DER digest and identity epoch, calls the B01 helper, and stores that exact `{LocalFaultID,IncidentID}` in the recovery aggregate used by `RecoveryStateSnapshotV1.SortedAcknowledgedLocalFaultBindings`. That aggregate's trust-conflict recovery locator is the same retained incident UUID；the recovery authorizer、poll builder and attestation verifier compare exact equality and have no alternate recovery/session-ID constructor. It commits a B01 canonical effect and calls `Coordinator.Finalize` outside locks；the specialized binding containing that local-fault ID and the exact incident ID/header value is returned only after incident/fault-binding/fence/audit/outbox terminalize together. Response loss uses Inspect/Recover, recomputes the same ID and returns the same incident/header without creating another incident/fence or fault binding. The recovery aggregate remains pollable after reconnect/restart and accepts only the exact stopped attestation needed for the signed clear transition；ordinary service access remains disabled throughout.

`SubmitEvidence` rate-limits before copying/verification and submits one bounded deep copy to the independent cap-16/two-worker low-priority queue, waiting only within the endpoint deadline. Workers strict-parse the complete canonical bytes by generated kind, independently verify root/deployment-authority/online-key signatures、schema/validity、authority identity/epoch/sequence、version/generation and digest, and independently query the B01 provider plus bound server artifact view outside DB locks. Client-reported digests never count as evidence. A single valid unpaired artifact returns the exact generated accepted ACK but makes no escalation. One submitted artifact plus one distinct exact server fence-finalized artifact, or two submitted artifacts, may escalate only when both are independently valid and match the frozen pair predicate/map. That escalation uses a separately derived `security_incident_open` Coordinator operation and one activation transaction to resolve the exact unverified row、open/reconcile the one typed incident、retain quarantine and write resolution/audit/outbox；it never edits an incident subtype or a global high-water. Provider/DB mismatch invokes only the existing readiness recovery path. The result returns a `TrustConflictAckCommitBinding` so the handler can commit only the exact generated ACK even if escalation self-invalidated the entry predicate.

Run: `go test ./internal/nodecontrol/recovery -run 'TestStoppedRecovery|TestRecoveryPollGeneratedMapping|TestRecoveryAttestation|TestRecoveryIdentityFlow|TestTrustConflictHighWaterIsolation|TestTrustConflictIncidentRecoveryEndpointsRemainReachable|TestTrustConflictEvidenceEscalation|TestTrustConflictQueueBounds' -count=1`

Expected: PASS; every success still leaves operator state disabled and every failed prerequisite remains fail closed.

- [ ] **Step 9 (5 min): Write RED host-incident registration and per-incident clear tests**

Verify `RegisterHostSecurityIncident` consumes fresh exact node/incident evidence with action=`register_host_security_incident` once and materializes or reconciles the typed server incident rather than treating absence as cleared. For `ClearSecurityQuarantine`, require action=`clear_security_latches` and cover subtype-specific prerequisites, exact stopped snapshot/attestation/evidence digests, wrong incident, evidence reuse, every other valid action, another concurrent open incident, no local latch binding, local and supervisor bindings, response loss, and clearing one incident without broadening another.

Run: `go test ./internal/nodecontrol/recovery -run 'TestRegisterHostSecurityIncident|TestClearSecurityQuarantine' -count=1`

Expected: FAIL because per-incident recovery actions are absent.

- [ ] **Step 10 (5 min): Implement two-phase latch clear and explicit resume**

Without a local/supervisor binding, finalize only the exact incident after machine-verifying its subtype evidence and exact `clear_security_latches` remediation action. With bindings—including the deterministic trust-conflict local binding materialized before the 409—first move it to `resolution_pending_agent_ack` and publish a strictly higher recovery generation with `required_action=clear_security_latches`, exact fault sets and remediation digest; only a later attestation of both new guard counters/digests finalizes resolution. The signed snapshot and agent independently derive the same trust-conflict local-fault ID; any mismatch rejects rather than creating an alias. The last resolution may set security normal and the certificate active but keeps operator disabled. `ResumeAfterSecurity` requires no open incident, fresh accepted attestation, exact active certificate and a target no broader than saved intent, then uses the B02 guard/signer/fence saga; only its activation transaction restores the explicit target.

Run: `go test ./internal/nodecontrol/recovery -run 'TestRegisterHostSecurityIncident|TestClearSecurityQuarantine|TestResumeAfterSecurity' -count=1`

Expected: PASS; signer failure or a concurrent security mutation leaves the node disabled and supersedes the pending resume intent.

- [ ] **Step 11 (5 min): Write RED two-person authority-restore reauthorization tests**

Start from a completed authority-restore session with the node disabled and old resume/desired discarded. Require proposal then approval from different operator IDs and exact credentials, equal authority epoch/scope/effect/target/evidence/If-Match, host evidence action exactly `reauthorize_after_restore`, fresh uncached authorization at each phase, and deadline `min(proposal.created_at+15m, evidence.completed_at+15m)`. Expire/revoke either credential, use any of the other four valid actions, shrink POP/role scope, change inventory/profile/session/evidence/effect or attempt one credential in both roles; require atomic supersede of proposal, approval and signing intent.

Run: `go test ./internal/nodecontrol/recovery -run TestReauthorizeAfterRestoreTwoPerson -count=1`

Expected: FAIL because restore proposal/approval activation is absent.

- [ ] **Step 12 (5 min): Implement proposal, approval, and finalized restore activation**

Persist one-time proposal/approval rows containing both exact credential snapshots and a shared effect digest；the approver cannot change proposal fields. Only after approval create the B02 `restore_reauthorize` transition/signing intent, whose `RecoveryTransitionCapture` includes immutable approval IDs and credential snapshot/version/scope digests. During the registered state handler's lock-free material phase, call `RecoveryTransitionGuard.CaptureActivationDecisionCapture`；it loads those exact IDs without caller DBTX, calls the external authorizer uncached for each credential, authenticates the approved time/floor source and returns only `state.NewRecoveryTransitionDecisionCapture` sealed over exact authorization results、credential versions/scopes、operation/effect and deadline. The B02 handler recomputes the activation-input digest and returns `ActivationDecisionMaterial`; Coordinator supplies same-Provider Head/checkpoint and branded proof. In the sole activation transaction the state activator passes that capture plus `ValidatedActivationDecisionEvidence` to the guard, which re-locks session/inventory/POP/profile/evidence/approval rows, compares every immutable fact, publishes the new desired generation selecting only enabled or draining, consumes approvals and completes the session atomically with proof group/fence/resolution/audit/outbox. Parsed recovery revalidates the stored terminal group without external authorization. No external authorization occurs under DBTX and there is no re-entered/post-Finalize transaction. Never reuse backup desired or resume intent.

Run: `go test ./internal/nodecontrol/recovery -run TestReauthorizeAfterRestoreTwoPerson -count=1`

Expected: PASS; missing, stale, same-operator or scope-mismatched approval never reaches signer-visible activation.

- [ ] **Step 13 (5 min): Exercise crash and concurrency recovery at every boundary**

Crash before/after reserve, fail-closed transaction, provider finalize, every receipt/fence/resolution/audit/outbox activation write, token consumption, Commit invocation/response loss, recovery-sign prepare/sign/fence/activation, high-water unverified incident, evidence queue acceptance/worker validation/escalation/ACK commit, `resolution_pending_agent_ack`, latch-clear attestation, server resolution, restore proposal, approval and final activation. Race duplicate fault IDs, duplicate evidence requests, queue saturation, evidence escalation versus new high-water conflict/host clear, two reenrolls, incident overflow, clear versus new incident, resume versus disable and restore approval versus credential revocation. Retry exact operation IDs and prove no partial fence/domain visibility, one terminal audit outcome, no reopened authority and no duplicate generation/grant/receipt/typed incident.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/recovery' -Run '^(TestRecoveryCrashMatrix|TestRecoveryConcurrency)$' -Timeout 6m`

Expected: PASS with no provider/authorizer/signer call under a PostgreSQL lock.

- [ ] **Step 14 (4 min): REFACTOR finite errors, retention, and secret-free telemetry**

Map only invalid request, unauthenticated, forbidden, not found, conflict, rate limited, dependency unavailable and internal. Logs/audit/metrics may contain finite subtype/status/artifact kind/queue result and opaque IDs, never canonical conflict bytes、client-reported digest、grant、CSR、certificate DER、proof、remediation signature、guard digest、provider/SQL text or free-form reason. Retain unresolved/session/evidence/approval rows while referenced; delete only terminal unreferenced rows after the Batch 01 windows.

Run: `go test -race ./internal/nodecontrol/recovery ./internal/nodecontrol/hostevidence -count=1`

Expected: PASS with privacy canaries, stable finite reason cardinality and no data race.

- [ ] **Step 15 (2 min): Commit recovery and security workflows**

```bash
git add db/queries/nodecontrol_recovery.sql internal/store/nodecontrol_recovery.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/hostevidence/remediation.go internal/nodecontrol/recovery/types.go internal/nodecontrol/recovery/repository.go internal/nodecontrol/recovery/postgres_repository.go internal/nodecontrol/recovery/security_fault.go internal/nodecontrol/recovery/trust_conflict.go internal/nodecontrol/recovery/operator_actions.go internal/nodecontrol/identity/repository.go internal/nodecontrol/identity/postgres_repository.go internal/nodecontrol/identity/service.go internal/nodecontrol/hostevidence/remediation_test.go internal/nodecontrol/recovery/postgres_repository_integration_test.go internal/nodecontrol/recovery/security_fault_test.go internal/nodecontrol/recovery/trust_conflict_test.go internal/nodecontrol/recovery/operator_actions_test.go internal/nodecontrol/recovery/service_crash_test.go testdata/c12/host-remediation-vectors.json
git commit -m "feat(nodecontrol): recover quarantined node identities"
```

### Task 7B: Register signed node resource envelopes through the closed authority dispatcher

**Files:**
- Create: `db/queries/nodecontrol_resource_envelope.sql`
- Create generated: `internal/store/nodecontrol_resource_envelope.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Create: `internal/nodecontrol/hostevidence/resource_envelope.go`
- Create: `internal/nodecontrol/hostevidence/resource_envelope_service.go`
- Test: `internal/nodecontrol/hostevidence/resource_envelope_test.go`
- Test (first line `//go:build integration`): `internal/nodecontrol/hostevidence/resource_envelope_integration_test.go`
- Test: `internal/nodecontrol/hostevidence/resource_envelope_crash_test.go`

**Interfaces:**
- Consumes: the sole B02 `contracts.NodeResourceEnvelopeV1`/package canonicalizer, task 4 exact deployment key role `node_resource_envelope`, task 7 `HostRemediationVerifier`, B02 `operator.AuthorizedMutationBinding` and inventory DBTX locks, Batch 01 `authority.Coordinator`/registered effect/proof-column contracts and the content-immutable resource-envelope table.
- Produces: `VerifiedResourceEnvelope`, a `ResourceEnvelopeService` for the generated operator route, and the unique registered resolver/three-method activator for `resource_envelope_activate`. It returns authenticated material, permits only the guarded `activation_` proof-group null→terminal transition, validates parsed outcomes with zero writes/external calls, and never lets generic inventory mutation advance the pointer.

```go
type ResourceEnvelopeService interface {
	Register(context.Context, operator.AuthorizedMutationBinding, nodeoperatorv1.RegisterResourceEnvelopeV1) (nodeoperatorv1.ResourceEnvelopeActivationV1, error)
}

type ResourceEnvelopeEffectHandler interface {
	authority.RegisteredEffectResolver
	authority.RegisteredEffectActivator
}
```

- [ ] **Step 1 (5 min): Write RED role, host-evidence and monotonicity tests**

Require exact package JCS/transcript、role=`node_resource_envelope`、deployment key ID/signature、node audience、fresh task-7 host-remediation evidence with action exactly `register_resource_envelope`、matching capacity digest、security-admin binding、If-Match and idempotency. Cover lower authority/version、fork、body/limit/headroom overflow、wrong node/action/key role、other remediation actions、stale/replay、ordinary writer and concurrent N/N+1. Require package authority tuple equal Coordinator reservation. Add opaque-query echo and cross-kind tests；mutate configured trusted-time identity/floor signature/expiry/deadline, every package/key/evidence/capacity/inventory/operator typed input, every `activation_` preimage/digest and terminal pointer column. Fresh failure leaves pointer/audit/outbox unchanged；parsed validation accepts one terminal row and records zero writes/provider/time/verifier calls.

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestResourceEnvelopeVerification|TestResourceEnvelopeService' -count=1`

Expected: FAIL because the verifier/service/handler do not exist.

- [ ] **Step 2 (5 min): Implement pure verification before any database lock**

Use only B02 canonical package bytes/transcript and task 4's exact role key. Verify host-remediation evidence independently, require its closed action exactly equals `HostRemediationRegisterResourceEnvelope`, bind its node/action/capacity/package digests and freshness, perform checked agent+supervisor+core+OS-headroom arithmetic, and return defensive opaque facts. Neither verifier can reserve authority、write PostgreSQL or accept caller-supplied limits outside the signed package.

- [ ] **Step 3 (5 min): Add immutable prepare and caller-DBTX activation queries**

In one prepare transaction lock inventory/current pointer in B01 order, recheck binding/If-Match/version/capacity evidence, and insert one `node_resource_envelopes` row bound to exact operation/authority/commitment and audit inputs. Package/policy columns remain permanently immutable；Batch 01's `activation_` proof columns alone permit one guarded all-null→terminal transition, providing the unique terminal anchor without a generic table or lifecycle rewrite. Before activation the row is unreachable because inventory points to the prior version. The immutable typed row includes the finite activation deadline.

`ResolveRegisteredAuthorityEffectForUpdate` accepts only opaque query and caller DBTX, locks exact row/pointer, validates expected operation/epoch/sequence/kind/scope and returns its echo. `CaptureActivationDecisionMaterial` authenticates configured trusted-time/floor identity outside DBTX, derives deadline/input digest from that row and returns no Provider/final evidence. Fresh activation re-resolves/recomputes package/key/evidence/capacity/inventory/operator inputs, advances the exact pointer and atomically stores terminal `activation_` proof group、resolution/audit/outbox/fence visibility without consuming admission. Parsed `ValidatePersistedAuthorityEffect` accepts one exact terminal envelope/pointer binding and performs zero writes/provider/time/verifier calls. Generated pointer SQL is private；ordinary operator code cannot call it.

- [ ] **Step 4 (4 min): Drive Reserve→prepare→Finalize without a second activation transaction**

Call `Coordinator.Reserve` before prepare, require its tuple equals the signed package, commit the prepared row, then Finalize outside locks. Definite pre-row failure must prove absent before Abort；commit uncertainty Inspect/Recovers first. A terminal mismatch requires a new reservation and newly signed package. Coordinator alone executes `Begin → material/trusted time → same-Provider Head/checkpoint → Complete/New → contextual proof`, opens activation, and for rollback-resistant proof consumes after every write immediately before one Commit；none-time skips consume. Response loss reloads stored Head/preimages from one locked snapshot and invokes parsed zero-write validation；only proved absence recaptures. Retry never creates a second row, edits package bytes or advances pointer directly.

- [ ] **Step 5 (5 min): Run generation, integration and crash matrices**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_resource_envelope.sql internal/store/nodecontrol_resource_envelope.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_resource_envelope.sql internal/store/nodecontrol_resource_envelope.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_resource_envelope.sql internal/store`

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestResourceEnvelopeVerification|TestResourceEnvelopeService' -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/hostevidence|./internal/store' -Run '^(TestResourceEnvelopeActivation|TestResourceEnvelopeCrashMatrix)$' -Timeout 5m`

Crash after Reserve、tuple mismatch/prepare failure、prepared insert、provider Finalize、material/trusted-time/Head/checkpoint/Complete/Validate、each proof/pointer/resolution/audit/outbox write、admission burn、Commit invocation and response loss. Mutate each query/proof/typed/terminal field and PITR cut. Prove definite-failure paths leave no unresolved reservation, uncertain paths never Abort before Inspect, no partial fence/proof/pointer visibility exists, persisted validation performs zero mutations/external calls, stale desired preparation cannot observe the pointer, and one exact retry result survives. Expected: PASS.

- [ ] **Step 6 (3 min): Commit the unique resource-envelope workflow**

```bash
git add db/queries/nodecontrol_resource_envelope.sql internal/store/nodecontrol_resource_envelope.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/hostevidence/resource_envelope.go internal/nodecontrol/hostevidence/resource_envelope_service.go internal/nodecontrol/hostevidence/resource_envelope_test.go internal/nodecontrol/hostevidence/resource_envelope_integration_test.go internal/nodecontrol/hostevidence/resource_envelope_crash_test.go
git commit -m "feat(nodecontrol): activate signed resource envelopes"
```

### Task 8: Adapt generated bootstrap, agent, and operator APIs

**Files:**
- Create: `internal/nodebootstrapapi/handler.go`
- Create: `internal/nodebootstrapapi/server.go`
- Create: `internal/nodeagentapi/handler.go`
- Create: `internal/nodeagentapi/server.go`
- Create: `internal/nodeoperatorapi/handler.go`
- Create: `internal/nodeoperatorapi/server.go`
- Test: `internal/nodebootstrapapi/handler_test.go`
- Test: `internal/nodeagentapi/handler_test.go`
- Test: `internal/nodeoperatorapi/handler_test.go`
- Test: `internal/nodeagentapi/response_commit_test.go`

**Interfaces:**
- Consumes: generated `nodebootstrapv1`, `nodeagentv1`, `nodeoperatorv1` strict server interfaces; the same Batch 01 bound `serving.Reader` used by task 6；task 3 identity service; task 6 authorizers/waiters and specialized conflict commit gates；task 7 recovery and `TrustConflictService`; task 7B resource-envelope service; Batch 02 inventory/operator/state response builders. B07-T03 replaces the exact fail-closed observation implementation defined below.
- Produces: generated-interface-complete handlers, bounded decoder/encoder, finite public error mapper, ordinary response-commit gate, security-fault specialized receipt-commit path, desired high-water fixed-conflict commit path and trust-evidence generated-ACK commit path.

```go
type BootstrapService interface {
	Claim(context.Context, identity.ClaimRequest) (identity.IssuanceResponse, error)
}

var ErrObservationDependencyUnavailable = errors.New("observation dependency unavailable")

type ObservationService interface {
	Accept(context.Context, NodeAuthorization, nodeagentv1.NodeObservationV1) (nodeagentv1.ObservationAckV1, error)
}

type DesiredResponseBuilder interface {
	BuildDesiredPollResponse(context.Context, NodeAuthorization, serving.DesiredStateFacts, nodeagentv1.DesiredPollRequestV1) (nodeagentv1.NodeStatePollResponseV1, bool, error)
}

type UnavailableObservationService struct{}

func (UnavailableObservationService) Accept(_ context.Context, _ NodeAuthorization, _ nodeagentv1.NodeObservationV1) (nodeagentv1.ObservationAckV1, error) {
	return nodeagentv1.ObservationAckV1{}, ErrObservationDependencyUnavailable
}

type AgentServices struct {
	Identity      identity.IdentityService
	AuthorityReader *serving.Reader
	AuthorityCheckpoints NodeCheckpointReader
	Desired       DesiredResponseBuilder
	Recovery      recovery.Service
	TrustConflicts recovery.TrustConflictService
	Observation   ObservationService
	Authorizer    NodeRequestAuthorizer
	Waiters       *WaiterRegistry
}

type OperatorServices struct {
	Inventory     inventory.Repository
	Mutations     operator.Service
	State         state.StateService
	Identity      identity.IdentityService
	Recovery      recovery.Service
	ResourceEnvelopes hostevidence.ResourceEnvelopeService
	Authorizer    operator.OperatorAuthorizer
	Authenticator OperatorPeerAuthenticator
}
```

The desired route first obtains `serving.DesiredStateFacts` from `AuthorityReader.GetActiveDesiredState` and the exact server-CA high-water already captured in `NodeAuthorization`, then calls the injected same-Coordinator `AuthorityCheckpoints.CommittedNodeCheckpoint(SHA256(node_id))` outside every DB transaction. `DesiredResponseBuilder.BuildDesiredPollResponse` is only a bounded response/time-attestation builder and cannot query or select a desired row. It returns `changed=true` only for a `200` payload assembled from the reader-selected finalized desired facts plus root-chain → metadata → optional time-attestation order；`changed=false` means an empty `204`. It never returns pending signer bytes, and it obtains any nonce-bound time attestation through B02 without reserving a new authority sequence. The concrete `nodeagentapi.Handler` owns this call order；construction fails for a nil/mismatched Reader、checkpoint reader or builder.

- [ ] **Step 1 (5 min): Write RED generated-interface and body-limit tests**

Add compile-time assertions against `nodebootstrapv1.ServerInterface`, `nodeagentv1.ServerInterface` and `nodeoperatorv1.ServerInterface`. Test ordinary 64 KiB bodies, conflict-evidence aggregate 1 MiB/two-artifact exception, oversized Content-Length, unknown-length streamed overflow, truncated/chunked body, duplicate/unknown JSON, invalid UTF-8, trailing JSON and non-identity Content-Encoding. Assert construction fails without `TrustConflicts`; the generated trust-conflict request/ACK are mapped field-by-field and no local look-alike DTO exists. Assert `UnavailableObservationService` returns the fixed `dependency_unavailable` public result without reading or persisting observation data; B07-T03 must replace this injected value before observation readiness can become true.

- [ ] **Step 2 (2 min): Run the focused RED handler tests**

Run: `go test ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestGeneratedInterface|TestBodyLimits' -count=1`

Expected: FAIL because handlers do not exist.

- [ ] **Step 3 (5 min): Implement bootstrap claim adapter**

Apply source rate limit before body read, bound wire and decoded size to 64 KiB, strict-decode canonical attempt/node IDs, 32-byte grant/time nonce and CSR DER, call `IdentityService.Claim`, and return only leaf, issuer chain, authorization receipt, node-state metadata and time attestation. Never include other node/operator data or provider detail.

- [ ] **Step 4 (5 min): Implement agent route adapters and endpoint classes**

Map rotate, desired poll, recovery poll, observation, security fault, trust-conflict evidence and recovery attestation to the exact class from task 6. Authorize before full body/heavy dependency, use per-node finite rate limits, register poll waiters, and invoke injected domain service only after strict validation. The desired branch must call the injected `AuthorityReader.GetActiveDesiredState(nodeID)` and `AuthorityCheckpoints.CommittedNodeCheckpoint(SHA256(node_id))`, and use the authorizer-captured server-CA high-water；it cannot call a B02 `LoadActive`/repository query, raw provider or choose the highest generation. Before returning either 200 or 204, require current authority epoch consistency and compare every supplied server-CA/root/metadata/desired/recovery triple against its exact continuous defensive fact. Idempotent equal tuples proceed；a lower explainable stream selects the continuous response, never rewrites authority. A same version/sequence different digest or unexplained ahead tuple for one of those five artifact-backed streams constructs the closed five-value Task 7 `HighWaterConflictInput`, calls `TrustConflicts.MaterializeHighWaterConflict`, discards all prepared root/metadata/desired/time bytes and uses only `AuthorizeHighWaterConflictCommit` to emit the fixed bodyless `409` plus its exact finalized `Talenro-Trust-Conflict-Incident-ID` header. A client scalar node checkpoint ahead of the DB/provider-consistent checkpoint instead returns the same bodyless 409 with no incident header, invokes only the Coordinator checkpoint/Head/Inspect readiness path and creates no per-node incident/notice/latch. Other conflicts likewise emit the bodyless headerless 409. It never closes a global listener or advances a high-water from client input.

Recovery poll first compares its request high-water against the exact Task 7 authoritative recovery/root/metadata facts using the same triple/fork/ahead rules；a high-water conflict takes the same fenced materialization and specialized bodyless-409/header path, while wrong recovery ID or unknown incident remains a headerless conflict. Only after that check does it map Task 7's generated response plus bool exactly to 200/204；recovery attestation calls `recovery.Service`. Trust-conflict evidence requires the exact incident-bound authorizer, enforces four-per-hour before bounded body copy, rejects compression, strict-decodes the generated request, and calls only `TrustConflicts.SubmitEvidence`; after return it discards every non-generated value and uses `AuthorizeTrustConflictAckCommit` to emit exactly the generated ACK. Observation calls the injected service and therefore fails closed until B07-T03；security fault calls `ReportSecurityFault`, discards every non-receipt value, then uses only `AuthorizeSecurityFaultReceiptCommit` with the task 7 immutable binding. None of the three self-invalidating paths calls ordinary `ReauthorizeResponse`; no class may fall through to ordinary authorization.

- [ ] **Step 5 (5 min): Implement operator route adapters**

Authenticate exact operator leaf, call external authorization with exact OpenAPI action/target, require `Idempotency-Key` and finite reason on mutations plus `If-Match` on updates. Bind list cursor reopening to new authorization, 2-second database timeout, item/page byte budget and at most 200 items. Map all nine action operations without a default branch: `drainNode -> State`; `disableNode -> Recovery.Disable`; `reenrollNode -> Recovery.Reenroll`; `completeNodeReenrollment -> Recovery.CompleteReenrollment`; `registerNodeHostSecurityIncident -> Recovery.RegisterHostSecurityIncident`; `registerNodeResourceEnvelope -> ResourceEnvelopes.Register`; `clearNodeSecurityQuarantine -> Recovery.ClearSecurityQuarantine`; `resumeNodeAfterSecurity -> Recovery.ResumeAfterSecurity`; `reauthorizeNodeAfterRestore -> Recovery.ReauthorizeAfterRestore`. The latter eight require a freshly constructed uncached `operator.AuthorizedMutationBinding` with security-admin role；`drainNode` alone permits the writer role. Proposal and approval phases each reauthenticate and reauthorize the exact credential. Generic `Mutations.Execute` has no resource-envelope pointer capability.

- [ ] **Step 6 (4 min): Implement finite public error mapping**

Emit only `invalid_request`, `unauthenticated`, `forbidden`, `not_found`, `conflict`, `rate_limited`, `dependency_unavailable`, `internal`. Use a fixed response schema/request ID; log only low-cardinality internal reason. Provider, TLS, SQL, key, grant, CSR, certificate and filesystem text never reaches JSON.

- [ ] **Step 7 (4 min): Add response-commit authorization for 200 and 204**

Prepare payload in bounded memory, reauthorize immediately before headers through task 6's same bound Reader, then use `http.NewResponseController` to set a response-phase write deadline of trusted/monotonic now plus 10 seconds. For long poll, the listener has no global write timeout；cancellation、authority-unavailable or stale auth discards root/metadata/state/time bytes. The only exceptions are the three exact self-invalidating commit bindings: security-fault may emit only its fence-finalized generated receipt, desired high-water isolation may emit only fixed `conflict` with no ordinary payload, and trust evidence may emit only its exact generated ACK. Add static/recording tests proving desired 200/204 cannot be produced unless `GetActiveDesiredState` completed first, every client high-water triple was compared, and no state repository serving method is reachable；a conflict path must materialize the fenced unverified incident before any response and expose zero prepared state bytes.

- [ ] **Step 8 (3 min): Run GREEN handler tests**

Run: `go test ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi -count=1`

Expected: PASS with all generated methods implemented, exact recovery 200/204 mapping, trust-conflict service injection and every malformed/oversized request bounded；single-node conflict/evidence tests leave unrelated nodes and global readiness unchanged unless the independent provider/DB probe itself proves mismatch.

- [ ] **Step 9 (4 min): REFACTOR common bounded HTTP helpers without merging auth domains**

Share only body counting, identity content-encoding check, error encoding and response-size accounting. Keep three generated packages and listener-auth domains separate; add a dependency-direction test rejecting imports from one API adapter into another.

Run: `go test ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/strictjson -count=1`

Expected: PASS.

- [ ] **Step 10 (2 min): Commit generated API adapters**

```bash
git add internal/nodebootstrapapi/handler.go internal/nodebootstrapapi/server.go internal/nodeagentapi/handler.go internal/nodeagentapi/server.go internal/nodeoperatorapi/handler.go internal/nodeoperatorapi/server.go internal/nodebootstrapapi/handler_test.go internal/nodeagentapi/handler_test.go internal/nodeoperatorapi/handler_test.go internal/nodeagentapi/response_commit_test.go
git commit -m "feat(nodecontrol): adapt isolated node APIs"
```

### Task 9: Build bounded TLS 1.3 and HTTP/2 listeners

**Files:**
- Create: `internal/platform/tlsserver.go`
- Test: `internal/platform/tlsserver_test.go`
- Test: `internal/platform/tlsserver_race_test.go`
- Test (first line `//go:build integration`): `internal/platform/tlsserver_integration_test.go`

**Interfaces:**
- Consumes: exact server certificate/profile, verified purpose-specific client CA pools, task 6 peer identity callbacks, existing platform shutdown conventions.
- Produces: three listener constructors, shared unauthenticated-handshake limiter, per-identity connection limiter and deterministic shutdown.

```go
type TLSListenerKind string

const (
	ListenerBootstrap TLSListenerKind = "bootstrap"
	ListenerAgent     TLSListenerKind = "agent"
	ListenerOperator  TLSListenerKind = "operator"
)

type TLSListenerLimits struct {
	AcceptedConnections uint32
	PerIdentityConnections uint16
	HTTP2Streams        uint32
	MaxHeaderBytes      int
	HandshakeTimeout    time.Duration
	ReadHeaderTimeout   time.Duration
	IdleTimeout         time.Duration
	WriteTimeout        time.Duration
}

type TLSListenerConfig struct {
	Kind              TLSListenerKind
	Address           string
	ServerCertificate tls.Certificate
	ClientCAs         *x509.CertPool
	Limits            TLSListenerLimits
	Handler           http.Handler
	HandshakeLimiter  *HandshakeLimiter
}

func NewTLSListener(TLSListenerConfig) (*TLSListener, error)
```

- [ ] **Step 1 (5 min): Write RED TLS-policy and cap tests**

Inspect configs and live handshakes for TLS 1.2, wrong ALPN, ticket resumption, missing required client cert, wrong node/operator CA, reused CA pool, 0-RTT attempt, slow handshake/header, idle socket, fifth HTTP/2 stream and one connection beyond each cap. Assert the bootstrap config uses `tls.NoClientCert` and never exposes a peer certificate to its handler.

- [ ] **Step 2 (2 min): Run the focused RED listener tests**

Run: `go test ./internal/platform -run TestNodeTLSListener -count=1`

Expected: FAIL because `NewTLSListener` does not exist.

- [ ] **Step 3 (5 min): Implement exact TLS configurations**

Set `MinVersion=MaxVersion=tls.VersionTLS13`, `NextProtos=[]string{"h2"}`, `SessionTicketsDisabled=true`, and never set `InsecureSkipVerify`. Bootstrap uses `tls.NoClientCert` and nil ClientCAs; agent/operator use `tls.RequireAndVerifyClientCert` with their distinct verified purpose pools. Validate the one configured server DNS SAN/profile before listening.

- [ ] **Step 4 (5 min): Implement pre-handler connection and handshake caps**

Wrap `net.Listener.Accept` before goroutine creation. Use listener caps bootstrap 128, agent 1600, operator 64 plus one shared unauthenticated-handshake cap 256. After verified identity, enforce node max 2, operator credential max 4, bootstrap source max 10 and token bucket 60/minute; over-cap pre-TLS connections close without a response.

- [ ] **Step 5 (4 min): Configure HTTP/2 and deadlines**

Set `MaxHeaderBytes=16*1024`, TLS handshake/read-header timeouts 5 seconds, idle timeout 35 seconds and HTTP/2 `MaxConcurrentStreams=4`. Bootstrap/operator write timeout is 10 seconds; agent write timeout is zero so 25-second poll works, while handler response phase remains 10 seconds.

- [ ] **Step 6 (4 min): Enforce connection absolute lifetime and bundle invalidation**

For authenticated peers set deadline to `min(peer.NotAfter, connectedAt+30m)` and recheck trusted time/application auth per request. Maintain indexes by certificate/issuer; revocation, epoch change or CA removal closes matching idle/active transports and cancels waiters.

- [ ] **Step 7 (3 min): Run GREEN live listener tests**

Run: `go test ./internal/platform -run TestNodeTLSListener -count=1`

Expected: PASS; bootstrap does not request or expose a client certificate, agent/operator reject none/wrong CA, and only TLS 1.3 h2 succeeds.

- [ ] **Step 8 (5 min): Test slow clients, poll deadline, and cap release**

Hold sockets at TLS, header and idle phases; prove each deadline releases permits. Run a 25-second poll returning 204 and a changed-state poll returning 200; then use a slow reader and prove response phase stops within 10 seconds.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/platform' -Run '^(TestNodeListenerSlowClients|TestAgentLongPollDeadline)$' -Timeout 3m`

Expected: PASS with permits returning to initial counts.

- [ ] **Step 9 (4 min): REFACTOR close ordering and race behavior**

Stop accepts, cancel handshakes/waiters, close connections, wait for handlers, then clear identity indexes. Ensure release occurs exactly once for handshake failure and HTTP/2 connection close.

Run: `go test -race ./internal/platform -run TestNodeTLSListener -count=1`

Expected: PASS with no race or leaked goroutine.

- [ ] **Step 10 (2 min): Commit bounded listeners**

```bash
git add internal/platform/tlsserver.go internal/platform/tlsserver_test.go internal/platform/tlsserver_race_test.go internal/platform/tlsserver_integration_test.go
git commit -m "feat(platform): add bounded node tls listeners"
```

### Task 10: Wire production configuration, readiness, and the Batch 03 mTLS sub-gate

**Files:**
- Create: `internal/config/nodecontrol.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `cmd/control-api/main.go`
- Modify: `cmd/control-api/main_test.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`
- Create (first line `//go:build integration`): `internal/e2e/nodecontrol_mtls_test.go`

**Interfaces:**
- Consumes: task 3 identity service, task 4 trust-package verifier/deployment role keys, task 7 recovery repository/service/host-remediation verifier, task 7B resource-envelope service/handler, task 8 handlers, task 9 listeners, Batch 02 services/providers, Batch 01 authority readiness/bound Reader, existing public/metrics runtime.
- Produces: strict `NodeControlConfig`, validated provider separation, fail-closed lifecycle for all three node listeners and an end-to-end local mTLS gate.

```go
type NodeControlConfig struct {
	Enabled             bool                       `yaml:"enabled"`
	NodeTrustDomain     string                     `yaml:"node_trust_domain"`
	OperatorTrustDomain string                     `yaml:"operator_trust_domain"`
	Bootstrap           NodeListenerConfig         `yaml:"bootstrap"`
	Agent               NodeListenerConfig         `yaml:"agent"`
	Operator            NodeListenerConfig         `yaml:"operator"`
	Providers           NodeControlProviderConfig  `yaml:"providers"`
}

type NodeListenerConfig struct {
	Address                   string `yaml:"address"`
	ServerDNSName             string `yaml:"server_dns_name"`
	ServerCertificateSecretRef string `yaml:"server_certificate_secret_ref"`
	ServerPrivateKeySecretRef string `yaml:"server_private_key_secret_ref"`
	ClientTrustPackageRef     string `yaml:"client_trust_package_ref"`
}

type NodeControlProviderConfig struct {
	Profile                               string `yaml:"profile"`
	AuthorityFenceProviderID              string `yaml:"authority_fence_provider_id"`
	DeploymentAuthorityKeySetProviderID   string `yaml:"deployment_authority_key_set_provider_id"`
	NodeCertificateIssuerID               string `yaml:"node_certificate_issuer_id"`
	NodeStateSignerID                     string `yaml:"node_state_signer_id"`
	RootShareRegistryID                   string `yaml:"root_share_registry_id"`
	OperatorAuthorizerID                  string `yaml:"operator_authorizer_id"`
	OperatorTrustGuardID                  string `yaml:"operator_trust_guard_id"`
	TrustedTimeSourceID                   string `yaml:"trusted_time_source_id"`
}

func (c NodeControlConfig) Validate(production bool) error
```

- [ ] **Step 1 (5 min): Write RED configuration separation and startup tests**

Reject duplicate listener addresses/DNS names/certificate refs, bootstrap client CA, missing agent/operator client package, reused node/operator CA, reused trust domains, missing deployment-authority key-set provider, provider ID reuse across incompatible roles, missing production provider, `LocalTest`/`Deterministic` production provider, C1.1 config signer/trust path, test private key and enabled listener with authority not ready. Assert the key-set provider returns four distinct role-bound Ed25519 keys and cannot alias the fence, issuer, signer, authorizer, trust-guard or trusted-time provider identity. Freeze the exact 15-kind dispatcher ownership table as one `EffectRegistration` per kind；13 supported entries have registered resolver+activator, while only `trust_bundle_publish` and `operator_authorizer_change` have both nil. Reject missing、duplicate、cross-owner、typed-nil、non-nil unsupported and any dispatcher wrapper/proxy passed to Coordinator.

- [ ] **Step 2 (2 min): Run the focused RED config tests**

Run: `go test ./internal/config ./cmd/control-api -run 'TestNodeControlConfig|TestNodeControlStartup' -count=1`

Expected: FAIL because node-control config and wiring do not exist.

- [ ] **Step 3 (5 min): Implement strict configuration validation**

Require three explicit non-wildcard listener addresses distinct from public/metrics, three controlled DNS names, purpose-matched trust packages, node/operator trust domains that are valid DNS names and unequal, and all external provider IDs. Production profile rejects test implementations by registered capability metadata, not only by substring. Secret refs are opaque resolver IDs; raw PEM/private key/grant fields are absent from config schema.

- [ ] **Step 4 (5 min): Construct domain services before listeners**

Resolve the configured deployment-authority key-set provider and validate its exact four role identities before constructing trust-package、resource-envelope and `HostRemediationVerifier` consumers. Resolve authority provider/repository, issuer, signer/root providers, operator authorizer/guard and trusted time. Before dispatcher/Coordinator/service, construct the identity registered handler, B03 recovery registered handler plus `RecoveryTransitionGuard` (no state-service/Coordinator dependency), inject that guard into the B02 state registered handler, and construct resource-envelope registered handler. Build one literal `[]authority.EffectRegistration` with exactly these 15 entries in bytewise order and with the named owner：`certificate_activate(identity)`、`certificate_revoke(identity)`、`desired_activate(state)`、`grant_claim(identity)`、`grant_create(identity)`、`identity_epoch_advance(identity)`、`metadata_publish(state)`、`operator_authorizer_change(nil,nil)`、`operator_transition(recovery)`、`recovery_activate(state)`、`resource_envelope_activate(resource-envelope)`、`root_publish(state)`、`security_incident_open(recovery)`、`security_incident_resolve(recovery)`、`trust_bundle_publish(nil,nil)`；one object may serve several entries, but every entry remains explicit. Pass the exact value returned by `NewEffectDispatcher` directly to one shared `NewCoordinator` and retain no aggregate wrapper/proxy；the unsupported trust entry is not replaced by task 4's read-only verifier.

Create the `BoundAuthorityReadSource` and one `serving.Reader` only from that Coordinator and its exact `PostgresRepository`. Inject the same Reader into the node request authorizer, `AgentServices.AuthorityReader` desired path and Task 7 trust-conflict server-counterpart verifier；inject the same Coordinator only through the narrow `NodeCheckpointReader` into `AgentServices.AuthorityCheckpoints`. Only after the Coordinator exists construct the B02 state service around the already-built state handler, then construct identity/recovery/trust-conflict/resource-envelope services；the recovery service may depend on that state service, but its previously built guard never does, so there is no service↔Coordinator construction cycle. Require the trust-conflict cap-16/two-worker queue and independent provider/DB probe before building `AgentServices`; shutdown stops admission, drains/cancels bounded jobs and zeroes retained artifact buffers. Then construct all three API handlers and listeners. Do not call `Listen` until every constructor and authority/trust/recovery readiness comparison passes. Missing finalized active trust high-water remains fail closed and cannot be repaired through an invented publisher；production may not substitute `UnavailableObservationService` once B07 observation readiness is declared.

- [ ] **Step 5 (4 min): Add one shared fail-closed lifecycle**

Start bootstrap, agent and operator listeners only after the authority probe is ready. A provider/DB mismatch, unexplained reservation, trust high-water rollback/fork, deployment key-set/remediation verifier failure, corrupt pending recovery operation or provider unavailability marks node-control readiness false, stops new accepts, cancels waiters/signing/recovery operations, and closes node transports. Existing public/metrics servers follow their own lifecycle and cannot make node listeners ready.

- [ ] **Step 6 (3 min): Run GREEN config and startup tests**

Run: `go test ./internal/config ./cmd/control-api ./internal/readiness -run 'TestNodeControlConfig|TestNodeControlStartup|TestAuthorityReadiness' -count=1`

Expected: PASS; no listener starts in every invalid case and shutdown completes once.

- [ ] **Step 7 (5 min): Add a live three-listener mTLS integration test**

Create this composed E2E test with `//go:build integration` as the first line；it must be absent from every untagged package gate.

Use reserved local-test trust domains, three server certificates and distinct node/operator client CAs. Assert bootstrap claim works without a client cert and its handler sees no peer identity；agent rotate reaches handler only with an exact bound-Reader active node row；desired 200/204 first traverses the same Reader and compares every high-water triple；operator list reaches handler only with strict operator leaf plus external authorization. Exercise the recovery-pending poll/attestation path and the specialized self-invalidating security-fault receipt gate. Add `TestNodeControlTrustConflictIsolationAndEscalation`: one legal malicious node submits a same-version false digest through both the would-be 204 and 200 paths, receives only bodyless 409 after its fenced unverified incident, captures the exact generated-client `Talenro-Trust-Conflict-Incident-ID` header, then uses that ID to submit invalid/one-valid/two-valid generated artifact requests through the same exact incident-bound credential. The header is absent on generic conflicts and cannot be replayed with another credential. Only the valid independent pair escalates to the mapped typed incident and exact generated ACK；999 other nodes continue desired poll/observation, and provider-consistent global readiness stays true. Route all nine operator actions to their exact task 8 dependency, including a role-verified/fence-atomic resource-envelope activation with action=`register_resource_envelope`, and prove the eight security-admin actions use uncached authorization. Node cert on operator、operator cert on agent、no cert on either mTLS listener、wrong DNS and TLS 1.2 all fail.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/e2e' -Run '^(TestNodeControlThreeListenerMTLS|TestNodeControlTrustConflictIsolationAndEscalation)$' -Timeout 7m`

Expected: PASS.

- [ ] **Step 8 (5 min): Add revoke, expiry, bundle change, and authority-PITR cases**

Reuse live HTTP/2 connections, then revoke certificate, increment identity epoch, expire trusted time, compromise/remove issuer CA and restore a pre-revocation database while keeping provider head. Assert the next request/response commit fails, existing transports close, and all node listeners stay unavailable on PITR mismatch. Materialize authority-restore sessions, prove stopped reenrollment remains disabled, and require two distinct fresh uncached security-admin credentials plus fresh host-remediation evidence before a new fenced desired generation can reauthorize a node.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/e2e' -Run '^(TestNodeControlKeepAliveRevocation|TestNodeControlPITRFailsClosed)$' -Timeout 5m`

Expected: PASS; no root/metadata/desired/receipt bytes are emitted after invalidation.

- [ ] **Step 9 (4 min): REFACTOR startup diagnostics and cleanup**

Expose only fixed readiness component names `node_authority`, `node_bootstrap`, `node_agent`, `node_operator`, `node_issuer`, `node_signer`, `node_trust`, `node_recovery`, `node_host_remediation`; details remain internal and secret-free. Test partial startup failure closes listeners and releases ports/permits without masking the primary error.

Run: `go test ./cmd/control-api ./internal/readiness ./internal/platform -count=1`

Expected: PASS.

- [ ] **Step 10 (4 min): Run deterministic generation and package gates**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code HEAD -- api gen internal/store`

Run: `powershell -NoProfile -Command "if (git ls-files --others --exclude-standard -- api gen internal/store) { throw 'untracked generated artifact' }"`

Expected: PASS with both drift and untracked-generated outputs empty.

Run: `go test ./internal/nodecontrol/identity ./internal/nodecontrol/hostevidence ./internal/nodecontrol/recovery ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/platform ./internal/config ./cmd/control-api -count=1 -timeout 8m`

Expected: PASS.

- [ ] **Step 11 (5 min): Run the mTLS integration sub-gate**

Run in the current PowerShell process:

```powershell
$expectedBatch03SubgateTags = [ordered]@{
    'internal/nodecontrol/identity/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/identity/service_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/trust_bundle_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/operator_guard_integration_test.go' = '//go:build integration'
    'internal/nodeagentapi/authorizer_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/recovery/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/resource_envelope_integration_test.go' = '//go:build integration'
    'internal/platform/tlsserver_integration_test.go' = '//go:build integration'
    'internal/e2e/nodecontrol_mtls_test.go' = '//go:build integration'
}
$badBatch03SubgateTags = @($expectedBatch03SubgateTags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
} | ForEach-Object Key)
if ($badBatch03SubgateTags.Count -ne 0) {
    $badBatch03SubgateTags
    throw 'B03 Tasks 1-10 integration test missing exact first-line tag'
}
```

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/identity|./internal/nodecontrol/hostevidence|./internal/nodecontrol/recovery|./internal/nodeagentapi|./internal/nodeoperatorapi|./internal/platform' -Timeout 10m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/e2e' -Timeout 10m`

Expected: PASS with issuance/recovery/resource-envelope crash recovery, specialized security receipt validation, host-remediation role/freshness/replay checks, administrative-disable and identity-compromise stopped reenrollment, two-phase latch clear, two-person authority-restore reauthorization, active-high-water trust-package rollback/fork、fixed unsupported publish checks, operator guard failure, bound exact peer/desired authorization, slow-client caps, live three-listener mTLS and authority PITR coverage.

- [ ] **Step 12 (2 min): Commit runtime wiring while keeping Batch 03 open**

```bash
git add internal/config/nodecontrol.go internal/config/config.go internal/config/config_test.go cmd/control-api/main.go cmd/control-api/main_test.go internal/readiness/checker.go internal/readiness/checker_test.go internal/e2e/nodecontrol_mtls_test.go
git commit -m "feat(control-api): start isolated node mtls listeners"
```

This sub-gate closes only Tasks 1–10. It cannot close B03 because the production ClaimV1 provider、runtime/timeline attestors、rollback-resistant commit archive、staging/source retirement and permanent Down enforcement are still absent；only Task 14 Step 8 is the complete B03 exit.

### Task 11: Implement the production ClaimV1 provider and epoch history

**Files:**
- Create: `internal/nodecontrol/claimv1/provider.go`
- Create: `internal/nodecontrol/claimv1/head.go`
- Create: `internal/nodecontrol/claimv1/credential_policy.go`
- Create: `internal/nodecontrol/claimv1/epoch.go`
- Create: `internal/nodecontrol/claimv1/consumption.go`
- Create: `internal/nodecontrol/claimv1/provider_test.go`
- Create (first line `//go:build integration`): `internal/nodecontrol/claimv1/provider_crash_integration_test.go`

**Interfaces:**
- Consumes: B01 canonical genesis/credential/epoch/lease/recovery request, response and evidence types plus pure verifiers; no B03 file defines a schema, digest transcript or DB table.
- Produces: production rollback-resistant `ClaimV1Provider`, exact signed Head/Inspect responses, atomic terminal/recovery/prefix-decision consumers, application-proof first-consumer/history-tail CAS and catchup state. B11 decides call order and supplies complete preimages outside DB transactions.

- [ ] **Step 1 (5 min): Write RED genesis and Head state-machine tests**

Cover absence proof consumption, one-time genesis, credential-policy publication, exact Head projection, request-ID/body exact retry, stale Head, wrong activation/deployment and retired-provider mutation. Assert every response schema uses the unique B01 provider-owned role/key/policy mapping，包括 `claim_v1_security_policy`、`claim_v1_incarnation_registry`、`claim_v1_inventory_anchor`、`claim_v1_provider_history_auditor` 与适用对象的 `claim_v1_provider`；no caller-selected signer metadata is accepted.

- [ ] **Step 2 (2 min): Run the focused provider RED tests**

Run: `go test ./internal/nodecontrol/claimv1 -run 'TestProviderGenesis|TestProviderHead|TestCredentialPolicy' -count=1`

Expected: FAIL because the production provider and rollback-resistant Head store are absent.

- [ ] **Step 3 (5 min): Implement the minimal rollback-resistant Head core**

Persist genesis/effective/terminal/ordinary/policy/runtime/epoch/staging/source/retirement tuples and control sequence outside PostgreSQL restore scope. Validate every B01 envelope before CAS, mechanically select the one provider-owned signer role/key/policy registered for that response schema, return byte-identical exact retries and reject caller-selected/cross-schema roles, same ID/different body, rollback, fork or mutation after permanent retirement.

- [ ] **Step 4 (5 min): Write RED epoch terminal-arbitration and exact-recovery tests**

Cover Prepare, Resolve-vs-Cancel terminal CAS, Inspect after response loss, terminal cancellation rebind, deterministic recovery transcript, same-prefix apply-vs-replacement decision, deferred suffix checkpoints and exact transcript replacement. Add negatives for two terminal winners, same prefix second decision ID, wrong prefix/ordinal, opaque DB proof, history gap and application after replacement.

- [ ] **Step 5 (5 min): Implement epoch, recovery and consumption ledgers**

Store complete immutable request/bundle/response/Head preimages. Require B01-verified commit proof and prefix decision closure, atomically consume the unique winner, maintain deferred suffix/history, and expose first-consumer/history-tail CAS for recovery application, applied rebind, catchup and marker-cleared resume. Provider responses never reference their own digest.

- [ ] **Step 6 (4 min): Add pre/post-catchup and result-reconcile state tests**

Exercise direct catchup, pre-catchup applied rebind, immediate DB result, fresh history-bound proof, unique terminal catchup, post-catchup marker-cleared rebind and pending-result reconcile. Reject a second catchup, wrong first-consumer witness, missing result with later mutation, stale proof, historical Head selection and ordinary mutation during recovery.

- [ ] **Step 7 (5 min): Inject every provider CAS crash seam**

Use deterministic barriers at request journal, pre-Head validation, terminal winner, recovery transcript, prefix-decision consumption, first-consumer append, history tail and response persistence. Restart after each seam and require either zero mutation or one complete immutable transition; exact retry must not increment sequence twice.

- [ ] **Step 8 (4 min): Run provider package, integration and race gates**

Run: `go test ./internal/nodecontrol/claimv1 -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/claimv1' -Run '^(TestProviderCrash|TestEpochRecovery)$' -Timeout 5m`

Run: `go test -race ./internal/nodecontrol/claimv1 -count=1`

Expected: PASS with one terminal winner, direct Head/history continuity and no duplicate consumer.

- [ ] **Step 9 (2 min): Commit only the production provider core**

```bash
git add internal/nodecontrol/claimv1/provider.go internal/nodecontrol/claimv1/head.go internal/nodecontrol/claimv1/credential_policy.go internal/nodecontrol/claimv1/epoch.go internal/nodecontrol/claimv1/consumption.go internal/nodecontrol/claimv1/provider_test.go internal/nodecontrol/claimv1/provider_crash_integration_test.go
git commit -m "feat(nodecontrol): add claim v1 provider history"
```

### Task 12: Implement runtime, timeline, rebind, and recovery attestors

**Files:**
- Create: `internal/nodecontrol/incarnation/runtime_attestor.go`
- Create: `internal/nodecontrol/incarnation/timeline_attestor.go`
- Create: `internal/nodecontrol/incarnation/rebind.go`
- Create: `internal/nodecontrol/incarnation/runtime_attestor_test.go`
- Create: `internal/nodecontrol/incarnation/timeline_attestor_test.go`
- Create (first line `//go:build integration`): `internal/nodecontrol/incarnation/rebind_integration_test.go`

**Interfaces:**
- Consumes: live postmaster/data-directory measurements, B01 incarnation/runtime/lineage/rebind/Head/Gap/PostRecovery contracts and Task 11 current Provider Head/history.
- Produces: fixed-role incarnation/runtime/timeline/Head/Gap/PostRecovery attestations, holder lease/termination, rebind registration/response inputs and deterministic result-reconcile observations. It never decides whether B11 should rebind or recover.

- [ ] **Step 1 (5 min): Write RED single-holder and lease-chain tests**

Require non-exportable incarnation keys outside backup scope, one live binding per incarnation, strictly increasing holder generation, exact process/mount/workload measurements, lease sequence `1..n`, descendant renewal and permanent termination. Reject concurrent postmasters, copied data directory, sibling mount, old lease, sequence gap, lease after termination and caller-provided liveness.

- [ ] **Step 2 (2 min): Run runtime-attestor RED tests**

Run: `go test ./internal/nodecontrol/incarnation -run 'TestRuntimeBinding|TestRuntimeLease|TestRuntimeTermination' -count=1`

Expected: FAIL because attestor state is absent.

- [ ] **Step 3 (5 min): Implement rollback-resistant runtime and timeline history**

Persist binding generation, lease chain, termination and signed timeline history outside PostgreSQL/VM backup scope. Verify parent/child system/timeline/OID/name/incarnation identity, exact fork cut and 1..64-link continuity; reject sibling, gap, rollback, unverifiable history ledger and point comparison ambiguity.

- [ ] **Step 4 (5 min): Write RED rebind, Gap and PostRecovery tests**

Cover ordinary, epoch-recovery-gap, epoch-recovery-applied, staging-held-preserve and staging-exclusion-recovery scopes. Verify authorization ordinal, old-holder termination, new candidate tuple, provider Head, DB Head attestation, deferred suffix, prefix decision, application proof/history and staging evidence set. Add pre/post-catchup and marker-cleared winners plus missing-result first-consumer/reconcile negatives.

- [ ] **Step 5 (5 min): Implement attested rebind and exact result reconciliation**

Trusted-read the DB Head/result on the same bound candidate, construct fixed B01 attestation bodies, and let Task 11 atomically CAS registration/lease/Head. Preserve effective/terminal/ordinary/policy/staging tuples by scope. A successful provider rebind with absent DB result exposes one dedicated reconcile input; it cannot authorize the next mutation until B01 repository records the exact result.

- [ ] **Step 6 (4 min): Add holder death, PITR and response-loss integration tests**

Terminate the holder before/after attestation and provider CAS, restore ancestor timelines, lose rebind response and lose DB result. Require exact Inspect/retry, fresh candidate attestations, strictly later fork-cut witness and one result; reject reuse of a Gap/PostRecovery proof on another candidate or descendant lease as a new holder generation.

- [ ] **Step 7 (4 min): Run attestor package, integration and race gates**

Run: `go test ./internal/nodecontrol/incarnation -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/incarnation' -Run '^(TestRebind|TestTimelinePITR|TestResultReconcile)$' -Timeout 5m`

Run: `go test -race ./internal/nodecontrol/incarnation -count=1`

Expected: PASS with one holder, continuous lineage and no mutation while a result is unreconciled.

- [ ] **Step 8 (2 min): Commit the production attestors**

```bash
git add internal/nodecontrol/incarnation/runtime_attestor.go internal/nodecontrol/incarnation/timeline_attestor.go internal/nodecontrol/incarnation/rebind.go internal/nodecontrol/incarnation/runtime_attestor_test.go internal/nodecontrol/incarnation/timeline_attestor_test.go internal/nodecontrol/incarnation/rebind_integration_test.go
git commit -m "feat(nodecontrol): attest runtime and timeline recovery"
```

### Task 13: Build the rollback-resistant commit archive, Fence, and Inspect runtime

**Files:**
- Create: `internal/nodecontrol/commitarchive/archive.go`
- Create: `internal/nodecontrol/commitarchive/wal_decoder.go`
- Create: `internal/nodecontrol/commitarchive/postgres_wal_decoder.go`
- Create: `internal/nodecontrol/commitarchive/inspect.go`
- Create: `internal/nodecontrol/commitarchive/fence.go`
- Create: `internal/nodecontrol/commitarchive/journal.go`
- Create: `internal/nodecontrol/commitarchive/archive_test.go`
- Create: `internal/nodecontrol/commitarchive/fence_test.go`
- Create: `internal/nodecontrol/commitarchive/inspect_test.go`
- Create (first line `//go:build integration`): `internal/nodecontrol/commitarchive/archive_crash_integration_test.go`

**Interfaces:**
- Consumes: B01 row-kind/related-set/purpose/generation/Fence/Inspect registries, Task 12 live candidate/lease/lineage and a mutually authenticated read-only PostgreSQL connection.
- Produces: database-incarnation commit attestations, rollback-resistant global latches/edges/streams, semantic-ID reservation responses, Fence responses and signed Inspect/history observations. B11 constructs requests; B03 independently trusted-reads, verifies and signs only fixed attestor roles.

- [ ] **Step 1 (5 min): Write RED trusted-read and exact WAL-location tests**

Map the immutable row's inserting top-level xid to the exact same-system/original-timeline `COMMIT` or `COMMIT PREPARED` end LSN. Reject caller body/LSN, later flush/replay/current WAL positions, missing/ambiguous xid, wrong timeline, uncommitted/multiple row and read connection not bound to the attested runtime.

- [ ] **Step 2 (2 min): Run WAL RED tests**

Run: `go test ./internal/nodecontrol/commitarchive -run 'TestTrustedRead|TestWALCommitPosition' -count=1`

Expected: FAIL because the decoder/archive runtime is absent.

- [ ] **Step 3 (5 min): Implement latches, related ownership and primary streams**

Lock sorted global semantic keys, insert/reuse the canonical body latch, CAS the application forward edge and decision reverse owner, then append the primary stream/head and exact-challenge response. Enforce `global latches -> edge/reverse owner -> primary stream -> exact retry`; apply decisions are related-only and replacement decisions primary-only. Commit latch/edge/envelope/entry/head atomically.

- [ ] **Step 4 (5 min): Write RED semantic journal and stable-once tests**

Cover reason-specific reservation, digest-only trusted lookup, rollback floor, context replay, pair/replacement shared prefix subject-slot, `decision_id -> slot` reverse uniqueness and stable Fence once-row. Renew the same binding/ID/generation lease and assert the once-key is unchanged; only a strictly higher holder generation may obtain a new key. Mutate every journal/context/key/six-holder field and reject duplicate records or alias decision IDs.

- [ ] **Step 5 (5 min): Implement atomic multi-key Fence**

Call only B01's exported typed derivation/verifier APIs to obtain every canonical digest, reason-specific semantic key and evidence cardinality；B03 must not copy a schema domain string, JCS formula or parallel canonicalizer. Verify fresh signed zero-row Inspect preconditions and lock `sorted global latches -> stable once-row -> response exact retry`. Advance all generations, successful once-row and response in one archive transaction. Exact retry returns the original response; same holder/different request, partial advance, orphan once-row and FreshRestoreImportApplicationV1 under any reason fail closed. Add a static ownership test plus B01 golden parity vectors proving `internal/nodecontrol/commitarchive` contains no copied domain/JCS formula and every derived value is byte-identical to the B01 helper result.

- [ ] **Step 6 (5 min): Implement signed Inspect and mechanical history selectors**

Echo the complete seven-field candidate tuple, observe live row count 0/1, distinguish absent/primary/related-only, return full archive chain and select current-parent, unconsumed-candidate or recovery-continuity entry mechanically from Provider history. Reject caller-selected history, partial edges, multiple locked bodies, stale generation, wrong Head/candidate and staging related-only evidence.

- [ ] **Step 7 (5 min): Add append-vs-Fence and commit-seam barriers**

Race an already trusted-read/signed append against one- and two-key Fence without sleeps. Inject failure at latch, edge, reverse owner, stream, exact retry, generation, once-row and response seams. Require append-wins exact restore or Fence-wins stale old generation, and storage observes only all-or-none state with no deadlock or reverse lock acquisition.

- [ ] **Step 8 (5 min): Run archive package, integration and race gates**

Run: `go test ./internal/nodecontrol/commitarchive -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/commitarchive' -Run '^(TestArchiveCrash|TestFenceRace|TestInspectPITR)$' -Timeout 8m`

Run: `go test -race ./internal/nodecontrol/commitarchive -count=1`

Expected: PASS with exact COMMIT positions, atomic two-key closure, stable lease-independent once rows and complete signed history.

- [ ] **Step 9 (3 min): Verify B01/B10/B11 ownership remains external**

Run: `git diff --exit-code HEAD -- db/migrations db/schema db/queries internal/store internal/c12evidence docs/superpowers/specs`

Run: `git diff --cached --name-only -- db/migrations db/schema db/queries internal/store internal/c12evidence docs/superpowers/specs`

Run: `git ls-files --others --exclude-standard -- db/migrations db/schema db/queries internal/store internal/c12evidence docs/superpowers/specs`

Expected: all three outputs are empty for this task；B03 created no schema, generated store/query, SpecDigest builder or orchestration artifact, including staged/untracked paths.

- [ ] **Step 10 (2 min): Commit the archive runtime**

```bash
git add internal/nodecontrol/commitarchive/archive.go internal/nodecontrol/commitarchive/wal_decoder.go internal/nodecontrol/commitarchive/postgres_wal_decoder.go internal/nodecontrol/commitarchive/inspect.go internal/nodecontrol/commitarchive/fence.go internal/nodecontrol/commitarchive/journal.go internal/nodecontrol/commitarchive/archive_test.go internal/nodecontrol/commitarchive/fence_test.go internal/nodecontrol/commitarchive/inspect_test.go internal/nodecontrol/commitarchive/archive_crash_integration_test.go
git commit -m "feat(nodecontrol): add rollback resistant commit archive"
```

### Task 14: Wire staging, source retirement, Down enforcement, and the B03 v7 gate

**Files:**
- Create: `internal/nodecontrol/claimv1/staging.go`
- Create: `internal/nodecontrol/claimv1/retirement.go`
- Create: `internal/nodecontrol/claimv1/staging_test.go`
- Create: `internal/nodecontrol/claimv1/retirement_test.go`
- Create: `internal/nodecontrol/authorityadapter/provider.go`
- Create: `internal/nodecontrol/authorityadapter/provider_test.go`
- Modify: `internal/config/nodecontrol.go`
- Modify: `internal/config/config_test.go`
- Modify: `cmd/control-api/main.go`
- Modify: `cmd/control-api/main_test.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`
- Create (first line `//go:build integration`): `internal/e2e/nodecontrol_authority_v7_test.go`
- Create: `testdata/c12/integration-node-identity-mtls.v1.json`

**Interfaces:**
- Consumes: Tasks 11–13, B01's exact six-method `authority.Provider`, and B01 exact staging/source/Down contracts/verifiers/store functions plus B01-typed authorization/request inputs later authored by B11 in production or by fixed tests here；B03 has no package or build dependency on B11.
- Produces: `authorityadapter.OrdinaryProvider`, the narrow production `claim_v1` view over Task 11's rollback-resistant `claimv1.ClaimV1Provider`, plus separate typed v7 adapters for hold/required/abort-only staging, challenge/Release/Abort provider CAS, source membership/retirement/permanent downgrade tombstones and fail-closed readiness. The ordinary view implements only B01's six methods and cannot expose or select a v7 operation. The typed v7 adapters verify policy decisions but do not choose or orchestrate them.

- [ ] **Step 1 (5 min): Write RED ordinary-Provider, staging-state and terminal-barrier tests**

In `authorityadapter/provider_test.go`, first require the compile-time assertion `var _ authority.Provider = (*OrdinaryProvider)(nil)`. Exercise the `claim_v1` profile through all six exact B01 methods: Reserve/Finalize/Abort/Inspect/Head/CommittedNodeCheckpoint, byte-identical same-request retry, changed tuple conflict, Finalize-vs-Abort mutual exclusion, response-loss Inspect, monotonic Head and checkpoint inclusion of only the requested node plus committed global-node-trust effects. Prove no v7 genesis/epoch/staging/source/Down method is present on or dynamically selectable through this view. Also cover acquire only after real capability registration proof, held-preserving rebind, held→required→abort-only recovery, exact request/response retry, import-vs-revocation terminal race, Release with real application proof and normal/recovery Abort with real revocation-application proof. Reject serving/ordinary mutation while held, required direct Abort and any external call under a DB transaction.

- [ ] **Step 2 (5 min): Implement durable staging provider state**

Persist acquisition anchor, capability, recovery request/response, state and terminal consumer in Task 11 Head. Independently verify Task 13 challenge/attestation and Task 12 current holder/lineage before CAS. Release/Abort transitions are exact-ID/body single-use and clear the staging tuple only at the terminal provider transition.

- [ ] **Step 3 (5 min): Test revocation-ID closure and lost fresh import**

Mutate trusted journal key, signed revocation.`revocation_application_id`, recovery intent/request value, application row ID and challenge/attestation outcome ID one at a time. Assert normal and recovery branches cannot splice. Restore after an unarchived import application is lost and prove every Fence reason rejects `FreshRestoreImportApplicationV1`; only intent-backed recovery application proof→signed recovery revocation→revocation application→Abort succeeds, with no reimport.

- [ ] **Step 4 (5 min): Implement and test the staging evidence universe**

Derive the complete candidate-key universe independently from provider request/response and phase. Require a signed Inspect observation for every key, including absent keys; locked items equal the exact primary subset and all seven candidate fields match. Exercise `exact_existing_candidate` and `archived_missing_suffix`, reject null for a nonempty universe, missing/extra/related-only/cross-candidate observations and lease/binding splice.

- [ ] **Step 5 (5 min): Implement source and permanent Down retirement enforcement**

Verify source membership, destructive-action plan, shutdown and permanent tombstones; require all-provider exact coverage before downgrade. After retirement reject genesis/policy/Prepare/rebind/recovery/staging/catchup/history mutation regardless fake-clock expiry. Expose fixed zero-projection counters to the B01 Down guard without generating the B11 evidence.

- [ ] **Step 6 (4 min): Wire production profiles and readiness**

Implement `authorityadapter.OrdinaryProvider` by delegating the exact ordinary reservation/terminal/history tuple to Task 11's rollback-resistant store; it must neither translate an opaque v7 digest into an ordinary request nor accept a caller-selected profile. Keep the compile-time assertion `var _ authority.Provider = (*OrdinaryProvider)(nil)` beside the implementation and run its exact idempotency/conflict/checkpoint tests under the `claim_v1` profile. Resolve distinct rollback-resistant provider, runtime/timeline attestor, WAL archive and signer identities in `NodeControlConfig`; reject deterministic/local implementations in production. Construct them before listeners. `cmd/control-api` constructs exactly one ordinary adapter, passes it to exactly one production `authority.Coordinator`, and injects that Coordinator into every B02/B03 ordinary writer; v7 genesis/epoch/staging/source/Down consumers receive separate narrow typed adapters. Keep listeners closed on Head/DB epoch mismatch, archive corruption, unreconciled result, staging recovery requirement or retirement inconsistency.

- [ ] **Step 7 (5 min): Add the composed authority-v7 integration matrix**

Create the composed test with `//go:build integration` as its first line. First exercise the production ordinary adapter and sole Coordinator with `claim_v1` Reserve→Finalize and Reserve→Abort, response-loss Inspect, Head and scoped checkpoint retries; then run genesis→epoch resolve/cancel races, exact recovery with pair/replacement prefix decisions, holder death before/after proof, pre/post-catchup, repeated PITR, staging normal Release/Abort and recovery Abort, source retirement and permanent Down refusal through the separate v7 interfaces. Assert B11-style calls occur outside SQL transactions and response loss uses exact Inspect/retry. No test may treat `fresh_v7_staging_closed` as completion.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/e2e' -Run '^TestNodeControlAuthorityV7$' -Timeout 12m`

Expected: PASS with all failure points closed and no duplicate terminal/consumer/generation transition.

- [ ] **Step 8 (8 min): Run the sole complete B03 generation, package, integration and race exit**

Create the sorted Batch 03 integration manifest with one explicit package per group, exact top-level test names, required `base`、`authority-v7` or `authority-v7-pitr` profile and bounded timeout. B01's parser gate must prove all three manifests exact-cover every tracked first-line integration test inside their closed package union once. Real WAL/backup/restore/promotion groups select only `authority-v7-pitr`; they consume B01's closed test harness and a test-only external archive/provider adapter, so this B03 gate proves mechanics but does not emit production authority evidence or replace P09's authoritative production PITR scope.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code HEAD -- api gen internal/store`

Run: `powershell -NoProfile -Command "if (git ls-files --others --exclude-standard -- api gen internal/store) { throw 'untracked generated artifact' }"`

Expected: PASS with both drift and untracked-generated outputs empty.

Run: `go test ./internal/nodecontrol/identity ./internal/nodecontrol/hostevidence ./internal/nodecontrol/recovery ./internal/nodecontrol/claimv1 ./internal/nodecontrol/incarnation ./internal/nodecontrol/commitarchive ./internal/nodecontrol/authorityadapter ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/platform ./internal/config ./internal/readiness ./cmd/control-api -count=1 -timeout 15m`

Run in the current PowerShell process:

```powershell
$expectedBatch03Tags = [ordered]@{
    'internal/nodecontrol/identity/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/identity/service_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/trust_bundle_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/operator_guard_integration_test.go' = '//go:build integration'
    'internal/nodeagentapi/authorizer_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/recovery/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/hostevidence/resource_envelope_integration_test.go' = '//go:build integration'
    'internal/platform/tlsserver_integration_test.go' = '//go:build integration'
    'internal/e2e/nodecontrol_mtls_test.go' = '//go:build integration'
    'internal/nodecontrol/claimv1/provider_crash_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/incarnation/rebind_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/commitarchive/archive_crash_integration_test.go' = '//go:build integration'
    'internal/e2e/nodecontrol_authority_v7_test.go' = '//go:build integration'
}
$badBatch03Tags = @($expectedBatch03Tags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
} | ForEach-Object Key)
if ($badBatch03Tags.Count -ne 0) {
    $badBatch03Tags
    throw 'B03 integration test missing exact first-line tag'
}
```

Run: `git add internal/e2e/nodecontrol_authority_v7_test.go testdata/c12/integration-node-identity-mtls.v1.json`

Expected: the new v7 E2E and manifest are tracked in the index before the tracked-file parser runs；no broader path is staged.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Suite batch03 -Timeout 240m`

Run: `go test -race ./internal/nodecontrol/identity ./internal/nodecontrol/hostevidence ./internal/nodecontrol/recovery ./internal/nodecontrol/claimv1 ./internal/nodecontrol/incarnation ./internal/nodecontrol/commitarchive ./internal/nodecontrol/authorityadapter ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/platform ./internal/config ./internal/readiness ./cmd/control-api -count=1 -timeout 18m`

Expected: PASS；the original `nodecontrol_mtls_test` integration set and the composed `TestNodeControlAuthorityV7` path both pass against the same final wiring, production provider/attestors are race-free, and no B01/B10/B11-owned artifact is emitted. This is the only B03 exit；no Task 10 result or focused Task 14 test may be promoted to Batch 03 completion.

- [ ] **Step 9 (2 min): Commit the v7 production composition**

```bash
git add internal/nodecontrol/claimv1/staging.go internal/nodecontrol/claimv1/retirement.go internal/nodecontrol/claimv1/staging_test.go internal/nodecontrol/claimv1/retirement_test.go internal/nodecontrol/authorityadapter/provider.go internal/nodecontrol/authorityadapter/provider_test.go internal/config/nodecontrol.go internal/config/config_test.go cmd/control-api/main.go cmd/control-api/main_test.go internal/readiness/checker.go internal/readiness/checker_test.go internal/e2e/nodecontrol_authority_v7_test.go testdata/c12/integration-node-identity-mtls.v1.json
git commit -m "feat(nodecontrol): wire authority v7 providers"
```

## Batch 03 completion check

- Grant is 32 crypto-random bytes, stored only as digest, CSR-bound, 10-minute, one-time and never re-disclosed on retry.
- Issuer request binds immutable issuance/issuer/public key/template digest; issuer runs outside locks and its output passes independent exact X.509 verification.
- Enrollment/rotation recovery returns one fence-finalized exact certificate and `CertificateAuthorizationReceiptV1`; late old-epoch results cannot activate.
- Node lineages last at most 30 days, rotate before `NotAfter-4h`, require reenroll below 36 hours and cap overlap at 4 active unexpired leaves.
- Deployment key roles are four distinct Ed25519 identities; trust packages enforce purpose/domain/version/sequence/digest/cumulative deauthorization and same-value fork rejection.
- Operator guard requires a fresh nonce-bound, at-most-5-minute rollback-resistant attestation and closes existing transport on failure or package invalidation.
- Host remediation accepts only the configured deployment-authority `host_remediation` role, exact transcript, five-literal consumer/action matrix, 15-minute freshness and one-time node/incident/action binding; the deployment key-set provider identity is distinct from every fence/issuer/signer/authorizer/time provider.
- Security-fault reporting applies quarantine/revocation before provider finalization and can return only the exact fence-finalized generated `nodeagentv1.SecurityFaultReceiptV1` through the specialized commit gate; response loss never reauthorizes the old credential.
- Identity compromise, administrative disable, retire and authority restore follow disabled -> new epoch/lineage -> recovery-pending certificate -> stopped snapshot/attestation -> complete while operator state remains disabled; per-incident clear and explicit resume are separate transitions.
- Authority-restore reauthorization requires fresh host evidence and two distinct fresh uncached security-admin exact credentials bound to one epoch/scope/effect within 15 minutes, then publishes a new fenced desired generation without backup desired/resume reuse.
- Every agent/operator request and response commit matches exact leaf DER/key/issuer/serial/status, authority/identity epoch, trusted time and endpoint-specific node state.
- All nine operator action operations have explicit service/action mappings; no unknown action falls through, and observation returns `dependency_unavailable` until B07-T03 replaces the injected fail-closed service.
- Bootstrap, agent and operator are distinct TLS 1.3 h2 listeners with exact client-auth modes, finite connections/streams/timeouts and working long-poll response deadlines.
- Desired 200/204 compares every client high-water triple；fork/ahead materializes only a bounded fenced unverified incident and fixed conflict for that node. The incident-bound evidence endpoint independently verifies at most two complete generated artifacts on its cap-16/two-worker queue and atomically escalates only a valid independent pair to the exact typed incident；single client claims never affect another node or global readiness.
- Authority/trust mismatch or PITR keeps all node-control listeners fail closed; no C1.1 trust/config signer is accepted as C1.2 authority.
- ClaimV1 provider has one genesis, one terminal winner, exact recovery/prefix arbitration, immutable first-consumer/history and pre/post-catchup result reconciliation；its narrow ordinary adapter implements the exact B01 `authority.Provider`, backs the sole production Coordinator with idempotent Inspect/Head/scoped-checkpoint behavior, and exposes no v7 method.
- Runtime/timeline attestors enforce one live holder, strictly increasing generation, descendant lease continuity, trusted lineage/fork cut and candidate-bound Gap/PostRecovery evidence.
- Commit archive trusted-reads exact rows and COMMIT end LSN, atomically closes global latches/related owner/primary stream, persists prefix subject slots and lease-independent stable once rows, and exposes mechanical signed Inspect history.
- Staging has one import-or-revocation DB winner and one Release-or-Abort provider terminal; `revocation_application_id` is closed end to end, and a lost unarchived fresh import can only recover by revocation→Abort, never Fence/reimport.
- Source membership/retirement and permanent downgrade tombstones block every later provider mutation; B03 emits no SpecDigest or B11 authorization/orchestration evidence.
- No Batch 04 keystore, RollbackGuard persistence, outbound agent loop or local reconciliation production behavior has entered this batch.
