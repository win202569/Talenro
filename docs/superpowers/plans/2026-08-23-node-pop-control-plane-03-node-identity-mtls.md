# Talenro C1.2 Batch 03 Node Identity and mTLS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 CSR-bound enrollment/rotation、外部 issuer 与 fence-finalized certificate authorization receipt、host-deployed trust-bundle package、rollback-resistant operator trust guard，以及彼此隔离且有界的 bootstrap/agent/operator TLS 1.3 listeners。

**Architecture:** PostgreSQL 只保存 grant digest、issuance intent、certificate identity 与 trust high-water，永不保存私钥或 grant plaintext。Identity workflow 先 reserve authority sequence并在短事务消费 grant/建立 immutable intent，锁外调用外部 issuer，独立验证 exact X.509 profile，再经 fence receipt 激活。Listener 的 TLS chain validation 只建立 transport；每次 request 入场和 response commit 都按 exact leaf row、authority/identity epoch、certificate/node state 重授权。Trust package 与 operator guard 从 host deployment 的独立 role keys/rollback-resistant provider取得信任，不能从当前 TLS 连接自举。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、ECDSA P-256、Ed25519、X.509、SHA-256、RFC 8785 JCS、TLS 1.3、HTTP/2、Batch 01 authority fence、Batch 02 inventory/operator/state services。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), approved working-tree SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`.

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
- 本计划中每个 checkbox 是一个可独立执行的 2–5 分钟动作；每个 task 必须 RED、GREEN、REFACTOR 后独立 commit。
- 只 stage task 的精确路径；不得 stage、删除或读取用户未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/`。

---

## Frozen file ownership

```text
db/queries/nodecontrol_identity.sql
db/queries/nodecontrol_recovery.sql
internal/store/nodecontrol_identity.sql.go
internal/store/nodecontrol_recovery.sql.go
internal/nodecontrol/identity/provider.go
internal/nodecontrol/identity/types.go
internal/nodecontrol/identity/csr.go
internal/nodecontrol/identity/x509_profile.go
internal/nodecontrol/identity/receipt.go
internal/nodecontrol/identity/repository.go
internal/nodecontrol/identity/postgres_repository.go
internal/nodecontrol/identity/service.go
internal/nodecontrol/hostevidence/deployment_keys.go
internal/nodecontrol/hostevidence/trust_bundle.go
internal/nodecontrol/hostevidence/operator_guard.go
internal/nodecontrol/hostevidence/remediation.go
internal/nodecontrol/recovery/types.go
internal/nodecontrol/recovery/repository.go
internal/nodecontrol/recovery/postgres_repository.go
internal/nodecontrol/recovery/security_fault.go
internal/nodecontrol/recovery/operator_actions.go
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
```

Batch 04 consumes `CertificateAuthorizationReceiptV1`, trust-package verification, endpoint authorization and signed-state handlers; it owns candidate keystore installation and outbound client transport. B03 does not write node private keys or agent rollback state.

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

Run: `go test ./internal/nodecontrol/identity -run 'Test|Fuzz' -count=1`

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
- Test: `internal/nodecontrol/identity/postgres_repository_integration_test.go`

**Interfaces:**
- Consumes: Batch 01 identity tables/authority repository, Batch 02 inventory node lock, task 1 identities.
- Produces: immutable grant/issuance/certificate records, closed statuses, exact certificate lookup and transaction-bound repository.

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
	ActivateCertificate(context.Context, store.DBTX, CertificateRecord) error
	LookupExactCertificate(context.Context, CertificateLookup) (CertificateAuthorization, error)
	RevokeIdentityEpoch(context.Context, store.DBTX, uuid.UUID, uint64, RevokeReason) error
	ResolveAuthorityEffect(context.Context, uuid.UUID) (authority.ResolvedEffect, error)
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

Cover one unconsumed grant per node/epoch, 10-minute expiry, atomic claim, changed request digest, one issuance ID binding, terminal immutability, unique issuer/serial and leaf digest, exact public-key digest, max 4 unexpired active leaves per lineage, late old-epoch result, and a serial/SAN copy with changed DER/key.

- [ ] **Step 2 (2 min): Run the focused RED integration test**

Run: `go test -tags=integration ./internal/nodecontrol/identity -run TestPostgresIdentityRepository -count=1 -timeout 3m`

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

- [ ] **Step 4 (4 min): Add exact active certificate lookup**

Lookup uses issuer ID, unsigned serial bytes, exact leaf DER SHA-256, public-key SHA-256, node ID, identity epoch and authority epoch in one query; join inventory current identity epoch and a committed active fence. It returns one row only when certificate status is one of the caller-provided finite allowed statuses and `not_before <= trusted_now < not_after`.

- [ ] **Step 5 (3 min): Generate sqlc artifacts**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS; generated params retain all digest fields as fixed byte slices validated at the repository boundary.

- [ ] **Step 6 (5 min): Implement row conversion and transaction guards**

Use constant-time digest comparison for grant claims, require exactly one consumed row, expire elapsed grants, mark all old-epoch pending issuance superseded inside the identity lock, and count active unexpired lineage leaves before activation. Implement `ResolveAuthorityEffect` as a read-only lookup for only grant-create/claim, certificate activate/revoke, and identity-epoch-advance operations; return the immutable kind/node scope/effect digest and finite prepared/committed/terminal state, and reject all other kinds. Implement the recovery primitives above only against the caller-owned transaction; they never authorize an operator or start their own transaction. Never expose token digest or certificate DER in errors.

- [ ] **Step 7 (3 min): Run GREEN repository tests**

Run: `go test -tags=integration ./internal/nodecontrol/identity -run TestPostgresIdentityRepository -count=1 -timeout 3m`

Expected: PASS with a copied serial/SAN certificate rejected unless every exact lookup component matches.

- [ ] **Step 8 (4 min): REFACTOR corrupt-row handling and retention predicates**

Reject invalid digest length, zero epoch, unknown status, oversized serial and inconsistent revoked timestamps. Add queries that delete only terminal grant/issuance/certificate metadata older than 30 days and not referenced by active/recovery state.

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
- Test: `internal/nodecontrol/identity/service_integration_test.go`
- Test: `internal/nodecontrol/identity/service_crash_test.go`

**Interfaces:**
- Consumes: Batch 01 `authority.Coordinator`, task 1 `NodeCertificateIssuer`, task 2 repository/identity effect resolver, Batch 02 inventory/operator audit/outbox, cryptographic random reader and trusted time.
- Produces: one-time grant creation, claim/rotate/recover saga, `CertificateAuthorizationReceiptV1`, and pure response verifier consumed by B04.

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

Use a deterministic random reader to assert exactly 32 bytes are generated, only SHA-256 reaches repository arguments, plaintext appears only in the first successful result, the same idempotency retry returns resource metadata without plaintext, replacement requires a new If-Match and same CSR digest, and claimed/issuance-started grants cannot be replaced.

- [ ] **Step 2 (2 min): Run the focused RED service test**

Run: `go test ./internal/nodecontrol/identity -run 'TestCreateEnrollmentGrant|TestIssueCertificateSaga' -count=1`

Expected: FAIL because identity service and receipt verifier do not exist.

- [ ] **Step 3 (5 min): Implement first-enrollment and security-admin grant creation**

Writer creation requires provisioning, identity epoch 0 and never-issued. Reenroll creation requires security-admin, increments epoch and creates a new lineage under the identity lock after revoking old epoch state. In both cases call `authority.Coordinator.Reserve(EffectGrantCreate)` before the transaction, persist the exact immutable pending grant effect with expiry at trusted now plus 10 minutes, then call `Coordinator.Finalize(operation_id,effect_digest)` outside locks and mark the grant active in a final transaction only with that committed receipt. Return the one-time secret wrapper only after activation; its `String`/JSON/log methods redact, and response-loss retry returns metadata without redisclosing plaintext.

- [ ] **Step 4 (5 min): Implement claim/rotation prepare transactions**

Validate IDs, nonce, CSR and request digest; call `authority.Coordinator.Reserve` for `EffectGrantClaim` on initial/recovery claim or `EffectCertificateActivate` on rotation, then lock node/grant. Atomically record the exact reservation, consume the grant when present and insert immutable issuance with authority binding, exact issuer, CSR/public key/template digests and pending status. Rotation skips grant but requires exact active peer certificate, current epoch/lineage, no more than 4 active leaves, remaining lineage at least 36 hours, and new CSR key.

- [ ] **Step 5 (5 min): Call issuer outside locks and validate independently**

Invoke exact `IssueRequest` by issuance ID; verify result echoes, parse single DER, verify chain/profile/template/SPKI/digests and reject any provider deviation. Persist validated bytes/digests in a short transaction. Issuer timeout keeps the same pending operation recoverable; known malformed output marks rejected.

- [ ] **Step 6 (5 min): Finalize fence and activate certificate**

After the validated immutable issuance effect commits, call `authority.Coordinator.Finalize(operation_id,effect_digest)` outside locks. The registered identity resolver returns the exact grant/issuance effect; only the coordinator captures/binds the same-primary database point and finalizes/activates the provider receipt. Then re-lock node/issuance, require that exact committed receipt, and recheck epoch, lineage, issuer, template, public key, node state, active-leaf cap and expiry before inserting the exact certificate row, marking issuance active and writing audit/outbox. Ordinary initial/rotation activation sets certificate status `active`; an issuance explicitly prepared by the recovery application sets `recovery_pending` and cannot use the ordinary path. Late results become superseded.

- [ ] **Step 7 (4 min): Implement exact receipt and retry response**

Build the receipt only from the active row and committed authority receipt. Same issuance/attempt/request returns identical leaf, chain and receipt until `NotAfter-5m`; after that return `ErrReenrollRequired`. Any changed field returns conflict. `VerifyAuthorizationReceipt` compares every ID/digest/authority field, exact leaf SPKI/profile/chain/time and a fresh private-key proof challenge.

- [ ] **Step 8 (3 min): Run GREEN identity saga tests**

Run: `go test ./internal/nodecontrol/identity -run 'TestCreateEnrollmentGrant|TestIssueCertificateSaga|TestVerifyAuthorizationReceipt' -count=1`

Expected: PASS; no database fake observes plaintext grant and no provider call occurs under lock.

- [ ] **Step 9 (5 min): Add crash recovery and rotation-overlap tests**

Crash after authority reserve, grant consume/intent, issuer response, validated result, fence visibility, provider finalize and activation. Prove retry uses one issuance ID/certificate. Rotate at 50% with bounded jitter input, keep old and new exact rows active, reject fifth overlap, revoke all epoch rows together, and supersede late issuer results.

Run: `go test -tags=integration ./internal/nodecontrol/identity -run 'TestIdentityCrashRecovery|TestRotationOverlap' -count=1 -timeout 5m`

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

### Task 4: Verify and activate host-deployed trust-bundle packages

**Files:**
- Create: `internal/nodecontrol/hostevidence/deployment_keys.go`
- Create: `internal/nodecontrol/hostevidence/trust_bundle.go`
- Modify: `db/queries/nodecontrol_identity.sql`
- Modify generated: `internal/store/nodecontrol_identity.sql.go`
- Modify generated: `internal/store/querier.go`
- Test: `internal/nodecontrol/hostevidence/deployment_keys_test.go`
- Test: `internal/nodecontrol/hostevidence/trust_bundle_test.go`
- Test: `internal/nodecontrol/hostevidence/trust_bundle_integration_test.go`
- Create: `testdata/c12/trust-bundle-vectors.json`

**Interfaces:**
- Consumes: host-installed Ed25519 public keys, Batch 01 authority/fence repository and trust high-water table, task 1 CA profile validator, strict JCS.
- Produces: frozen deployment role registry, `TrustBundleManifestV1`, `TrustBundlePackageV1`, `VerifiedTrustBundle`, pure verifier and finalized activation repository.

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

- [ ] **Step 1 (5 min): Write RED deployment-key and package vector tests**

Test missing/duplicate roles, one key in two roles, wrong key ID, wrong signature role, cross-purpose/domain reuse, unsorted/duplicate entries, manifest/DER mismatch, non-CA DER, low version/sequence, same-value fork, illegal status transition, cumulative removal/reorder/re-authorization, 129th cumulative digest and trailing package bytes.

- [ ] **Step 2 (2 min): Run the focused RED verifier tests**

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestDeploymentAuthorityKeySet|TestTrustBundleVectors' -count=1`

Expected: FAIL because hostevidence trust types and vectors do not exist.

- [ ] **Step 3 (4 min): Implement deployment key validation**

Derive each key ID from raw 32-byte Ed25519 public key, reject private/unknown algorithm bytes, require bytewise unique IDs and unique physical keys, and expose `KeyForRole(role, keyID)` without fallback to another role.

- [ ] **Step 4 (5 min): Implement package canonicalization and signature verification**

Require schema `trust-bundle-manifest.v1`, a closed purpose, canonical trust domain, sorted entries/cumulative set, SHA-256 DER matches one-to-one, and transcript `TALENRO-TRUST-BUNDLE-PACKAGE-V1\x00 || JCS(manifest)`. Verify with exact role `trust_bundle`; package/DER/signature cannot update deployment keys.

- [ ] **Step 5 (4 min): Enforce CA profiles and monotonic transitions**

Validate each CA independently for CA constraints and purpose; allow normal `active→retiring→omitted`, emergency `active|retiring→compromised`, append compromised/removed digest in the same version, forbid current entries in cumulative set and all later reauthorization. Compare version/sequence/digest with `contracts.CompareVersionedDigest` and reject rollback/fork.

- [ ] **Step 6 (4 min): Add high-water locking and finalized activation SQL**

```sql
-- name: LockTrustBundleHighWater :one
SELECT purpose, trust_domain, authority_epoch, authority_sequence,
       bundle_version, bundle_digest, cumulative_set_digest
FROM nodecontrol.control_plane_trust_bundle_high_waters
WHERE purpose = sqlc.arg(purpose) AND trust_domain = sqlc.arg(trust_domain)
FOR UPDATE;

-- name: ActivateTrustBundleHighWater :execrows
UPDATE nodecontrol.control_plane_trust_bundle_high_waters AS high_water
SET authority_epoch = sqlc.arg(authority_epoch),
    authority_sequence = sqlc.arg(authority_sequence),
    bundle_version = sqlc.arg(bundle_version),
    bundle_digest = sqlc.arg(bundle_digest),
    cumulative_set_digest = sqlc.arg(cumulative_set_digest),
    updated_at = transaction_timestamp()
FROM nodecontrol.control_plane_authority_fences AS fence
WHERE high_water.purpose = sqlc.arg(purpose)
  AND high_water.trust_domain = sqlc.arg(trust_domain)
  AND fence.operation_id = sqlc.arg(operation_id)
  AND fence.provider_status = 'committed'
  AND fence.visibility_state = 'active';
```

- [ ] **Step 7 (3 min): Generate store artifacts and run GREEN tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS.

Run: `go test ./internal/nodecontrol/hostevidence -run 'TestDeploymentAuthorityKeySet|TestTrustBundleVectors' -count=1`

Expected: PASS and checked-in vector digests match lowercase hex exactly.

- [ ] **Step 8 (5 min): Test finalized high-water activation and restart**

Activate old, old+new, then new after the 48-hour overlap fixture; separately mark an old CA compromised. Replay the old package and restore an older database snapshot; startup comparison must fail closed rather than reauthorize it.

Run: `go test -tags=integration ./internal/nodecontrol/hostevidence -run TestTrustBundleHighWater -count=1 -timeout 3m`

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
- Test: `internal/nodecontrol/hostevidence/operator_guard_integration_test.go`

**Interfaces:**
- Consumes: task 4 verified purpose `operator_server` package and deployment role `operator_trust_guard`, external rollback-resistant provider, crypto-random nonce source, trusted platform time, transport close callback.
- Produces: exact attestation DTO, provider boundary, connection lease and fail-closed transport invalidation.

```go
type OperatorTrustGuardProvider interface {
	Attest(context.Context, OperatorTrustGuardRequest) (OperatorTrustGuardAttestationV1, error)
}

type OperatorTrustGuardRequest struct {
	RequestNonce contracts.Digest
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
	RequestNonce               contracts.Digest
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

Test 32-byte fresh nonce, echoed nonce, replay, provider unavailable/restart, wrong purpose/domain/key role, lower epoch/sequence/version, same-value fork, 5-minute expiry, future issuance, local package rollback, emergency removal and existing transport closure.

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

Run: `go test -tags=integration ./internal/nodecontrol/hostevidence -run TestOperatorTrustGuardProviderContract -count=1 -timeout 3m`

Expected: PASS.

- [ ] **Step 8 (2 min): Commit operator guard support**

```bash
git add internal/nodecontrol/hostevidence/operator_guard.go internal/nodecontrol/hostevidence/operator_guard_test.go internal/nodecontrol/hostevidence/operator_guard_race_test.go internal/nodecontrol/hostevidence/operator_guard_integration_test.go
git commit -m "feat(nodecontrol): guard operator client trust"
```

### Task 6: Authorize exact node and operator peers at request and response boundaries

**Files:**
- Create: `internal/nodecontrol/contracts/security_fault.go`
- Create: `internal/nodeagentapi/authorizer.go`
- Create: `internal/nodeagentapi/waiters.go`
- Create: `internal/nodeoperatorapi/authorizer.go`
- Test: `internal/nodeagentapi/authorizer_test.go`
- Test: `internal/nodeagentapi/waiters_test.go`
- Test: `internal/nodeoperatorapi/authorizer_test.go`
- Test: `internal/nodeagentapi/authorizer_integration_test.go`

**Interfaces:**
- Consumes: task 1 strict leaf identities, task 2 exact certificate lookup, Batch 02 `operator.OperatorAuthorizer`, trusted time, inventory/recovery/session state and trust-bundle high-water.
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
	AuthorizeSecurityFaultReceiptCommit(context.Context, NodeAuthorization, contracts.SecurityFaultCommitBinding) error
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
}

type OperatorPeerAuthenticator interface {
	Authenticate(context.Context, *x509.Certificate) (operator.Credential, error)
}
```

`contracts.SecurityFaultCommitBinding` contains operation ID, original exact certificate ID/DER/key/issuer/serial/identity epoch, local fault ID, request/effect digests, incident ID, authority epoch/sequence and committed receipt digest. Its validator permits serialization of only the fixed receipt and cannot authorize root, metadata, desired, recovery, time or another endpoint payload.

- [ ] **Step 1 (5 min): Write the RED endpoint authorization matrix**

Test every certificate status against every endpoint class and node state. Ordinary rotate/desired/observation require active, current epoch, operator enabled/draining and security normal. Recovery poll/attestation require active/recovery-pending/recovery-limited, operator disabled and exact nonterminal recovery session or incident. Security fault accepts those same three current-epoch statuses only as a tightening mutation. Conflict evidence additionally requires its exact open unverified incident. Unknown classes and revoked/expired/mismatched rows always deny. Separately prove ordinary `ReauthorizeResponse` rejects the now-revoked fault credential while `AuthorizeSecurityFaultReceiptCommit` accepts only its exact fence-finalized immutable receipt binding.

- [ ] **Step 2 (2 min): Run the focused RED tests**

Run: `go test ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestNodeAuthorizationMatrix|TestOperatorPeerAuthentication' -count=1`

Expected: FAIL because API authorizers do not exist.

- [ ] **Step 3 (5 min): Implement exact node peer extraction and lookup**

Require one verified leaf, strict node profile and canonical URI SAN; derive unsigned serial bytes, exact DER/SPKI digests and node ID. Query one current committed certificate row with trusted time and allowed statuses. Compare every returned field to the peer and inventory; no lookup by serial, SAN or CA alone.

- [ ] **Step 4 (4 min): Implement operator credential extraction**

Require strict operator profile and canonical operator URI; produce `operator.Credential` containing issuer ID, serial bytes, exact leaf DER/public-key digests, URI SAN, operator ID, authority epoch and credential version. Call external authorizer for every request; security-admin actions bypass cache according to B02.

- [ ] **Step 5 (4 min): Implement final response reauthorization**

`ReauthorizeResponse` opens a new short read transaction and rechecks exact certificate row/status, identity/authority epoch, node state, endpoint-specific recovery predicate, trust high-water and trusted time. It compares the returned authorization version to the entry snapshot. Any change returns only unauthenticated/forbidden and the handler must discard prepared payload before writing headers.

- [ ] **Step 6 (3 min): Run GREEN matrix tests**

Run: `go test ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestNodeAuthorizationMatrix|TestOperatorPeerAuthentication|TestResponseReauthorization' -count=1`

Expected: PASS; a copied serial/SAN with different DER or key is rejected.

- [ ] **Step 7 (4 min): Add node/issuer/certificate waiter cancellation**

`WaiterRegistry.Register(nodeID, certificateID, cancel)` returns an idempotent unregister function. Revoke, disable, identity-epoch change, quarantine, recovery-session completion and trust-bundle removal cancel matching waiters without holding registry locks while invoking callbacks.

Run: `go test -race ./internal/nodeagentapi -run TestWaiterRegistry -count=1`

Expected: PASS with no missed cancel, double callback or leak.

- [ ] **Step 8 (5 min): Prove keep-alive reauthorization against PostgreSQL**

Authorize once, revoke or increment identity epoch in a second transaction, then reuse the same authorization snapshot for commit. Repeat for certificate expiry and CA compromise.

Run: `go test -tags=integration ./internal/nodeagentapi -run TestKeepAliveAuthorizationRecheck -count=1 -timeout 3m`

Expected: PASS; every stale snapshot is rejected and contains no prepared state payload.

- [ ] **Step 9 (2 min): Commit request authorization**

```bash
git add internal/nodecontrol/contracts/security_fault.go internal/nodeagentapi/authorizer.go internal/nodeagentapi/waiters.go internal/nodeoperatorapi/authorizer.go internal/nodeagentapi/authorizer_test.go internal/nodeagentapi/waiters_test.go internal/nodeoperatorapi/authorizer_test.go internal/nodeagentapi/authorizer_integration_test.go
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
- Create: `internal/nodecontrol/recovery/operator_actions.go`
- Modify: `internal/nodecontrol/identity/repository.go`
- Modify: `internal/nodecontrol/identity/postgres_repository.go`
- Modify: `internal/nodecontrol/identity/service.go`
- Test: `internal/nodecontrol/hostevidence/remediation_test.go`
- Test: `internal/nodecontrol/recovery/postgres_repository_integration_test.go`
- Test: `internal/nodecontrol/recovery/security_fault_test.go`
- Test: `internal/nodecontrol/recovery/operator_actions_test.go`
- Test: `internal/nodecontrol/recovery/service_crash_test.go`
- Create: `testdata/c12/host-remediation-vectors.json`

**Interfaces:**
- Consumes: task 4 `DeploymentAuthorityKeySetV1` and exact `RoleHostRemediation`; task 2 transaction-bound `RecoveryIdentityRepository`; task 3 recovery issuance/receipt activation and identity effect resolver; task 6 node/operator authorization bindings; Batch 01 `authority.Coordinator` and recovery schema; Batch 02 inventory locks, recovery signer saga, `RecoveryTransitionGuard`, atomic audit/outbox, and uncached `operator.OperatorAuthorizer`; trusted time.
- Produces: canonical `HostRemediationEvidenceV1`/`HostRemediationVerifier`; `SecurityFaultReceiptV1`; `RecoveryAttestationV1`; bounded incident/session/evidence/approval repository; a read-only recovery `ResolveAuthorityEffect` registered in the Batch 01 dispatcher; `recovery.Service`; and the B02 transaction-bound recovery transition guard implementation.

```go
type SecurityFaultReceiptV1 struct {
	SchemaVersion     string
	LocalFaultID      uuid.UUID
	IncidentID        uuid.UUID
	RequestDigest     contracts.Digest
	AuthorityEpoch    uint64
	AuthoritySequence uint64
	Result            string // exactly "accepted"
}

type RecoveryAttestationV1 struct {
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
}

type RecoveryAttestationAckV1 struct {
	RecoveryID       uuid.UUID
	AttestationDigest contracts.Digest
	Result           string // exactly "accepted"
}

type GuardCounterEvidenceV1 struct {
	CounterIdentity string
	CounterValue    uint64
	StateDigest     contracts.Digest
}

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
	ReportSecurityFault(context.Context, NodeCredentialBinding, SecurityFaultReportV1) (SecurityFaultReceiptV1, error)
	PollRecovery(context.Context, NodeCredentialBinding, RecoveryPollRequestV1) (RecoveryPollResult, error)
	SubmitRecoveryAttestation(context.Context, NodeCredentialBinding, RecoveryAttestationV1) (RecoveryAttestationAckV1, error)
	Disable(context.Context, OperatorCommandBinding, DisableNodeV1) (RecoveryOperationV1, error)
	Reenroll(context.Context, OperatorCommandBinding, ReenrollNodeV1) (RecoveryEnrollmentGrantV1, error)
	CompleteReenrollment(context.Context, OperatorCommandBinding, CompleteReenrollmentV1) (RecoveryOperationV1, error)
	RegisterHostSecurityIncident(context.Context, OperatorCommandBinding, RegisterHostSecurityIncidentV1) (RecoveryOperationV1, error)
	ClearSecurityQuarantine(context.Context, OperatorCommandBinding, ClearSecurityQuarantineV1) (RecoveryOperationV1, error)
	ResumeAfterSecurity(context.Context, OperatorCommandBinding, ResumeAfterSecurityV1) (state.SigningOperationV1, error)
	ReauthorizeAfterRestore(context.Context, OperatorCommandBinding, ReauthorizeAfterRestoreV1) (RestoreReauthorizationV1, error)
}

type AuthorityEffectResolver interface {
	ResolveAuthorityEffect(context.Context, uuid.UUID) (authority.ResolvedEffect, error)
}
```

`NodeCredentialBinding` contains the task 6 exact certificate ID/DER/key/issuer/serial, node/identity/authority epochs and authorization version. `OperatorCommandBinding` contains the exact uncached B02 credential/authorization, command/idempotency/If-Match digests, target scope and trusted authorization time. Every request DTO above is the Plan 01 generated schema with no recovery-local duplicate. `VerifiedHostRemediation` owns defensive copies of the unsigned evidence, its RFC 8785 JCS digest and verified key/role; only the repository can consume that digest once for its exact node/incident/action. The recovery resolver recognizes only security-incident open/resolve, identity-epoch-advance, recovery activate and operator-transition effects stored by this task; it returns immutable exact bindings and never calls a provider.

- [ ] **Step 1 (5 min): Write RED host-remediation transcript and freshness vectors**

Generate `testdata/c12/host-remediation-vectors.json` for transcript `TALENRO-HOST-REMEDIATION-EVIDENCE-V1\x00 || JCS(unsigned_evidence)`. Cover every field above, canonical UUID/time/digest encoding, wrong role/key/algorithm, signature mutation, reordered/duplicate input, node/incident/action mismatch, future completion, exactly 15 minutes, one nanosecond beyond 15 minutes, trusted-time provider/floor rollback and evidence-ID replay.

Run: `go test ./internal/nodecontrol/hostevidence -run TestHostRemediationEvidenceVectors -count=1`

Expected: FAIL because the verifier and vectors do not exist.

- [ ] **Step 2 (5 min): Implement the pure role-separated verifier**

Strict-decode one schema version, require nonzero IDs/counters/map version, exact 32-byte digests, nonempty bounded provider/counter identities, `Algorithm=Ed25519`, and the configured exact `host_remediation` deployment key. Reconstruct JCS from typed fields, verify the domain-separated signature, return defensive copies plus the evidence digest, and never consult PostgreSQL or accept a `trust_bundle`, `operator_trust_guard`, or `node_resource_envelope` role key.

Run: `go test ./internal/nodecontrol/hostevidence -run TestHostRemediationEvidenceVectors -count=1`

Expected: PASS; every one-field mutation and cross-role signature is rejected.

- [ ] **Step 3 (5 min): Write RED repository bound and immutability tests**

Cover fault subtype slots 1–12, rejection of reserved slots 13–15, saturating overflow slot 16, 16/17 simultaneous incident rows, 64/65 local-fault bindings, 16/17 supervisor-fault bindings, occurrence saturation, immutable finalized security receipt, one nonterminal recovery session per node, one accepted attestation digest per session version/nonce, one-time remediation evidence, and proposal/approval uniqueness by exact operator credential and role. Prove `open -> resolution_pending_agent_ack -> resolved` and `overflow -> resolution_pending_agent_ack|resolved` are the only forward incident transitions.

Run: `go test -tags=integration ./internal/nodecontrol/recovery -run TestPostgresRecoveryBounds -count=1 -timeout 3m`

Expected: FAIL because recovery queries and repository do not exist.

- [ ] **Step 4 (5 min): Implement exact recovery SQL and repository conversions**

Add lock/upsert queries for the node recovery aggregate, typed/overflow incident, local/supervisor binding, security receipt intent, recovery session, attestation, consumed remediation evidence, restore proposal and restore approval. Every write receives caller-owned `store.DBTX`; conversions reject zero/range-invalid versions, unknown enums, wrong digest lengths, partial authority groups, expired evidence and corrupt terminal rows. Use the database uniqueness/check constraints from Batch 01 rather than preflight-only counting.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `go test -tags=integration ./internal/nodecontrol/recovery -run TestPostgresRecoveryBounds -count=1 -timeout 3m`

Expected: PASS; generated recovery sqlc parameters preserve every UUID, authority version, digest and nullable transition group without lossy strings.

- [ ] **Step 5 (5 min): Write the RED self-invalidating security-fault saga**

Start with an exact active current-epoch certificate. Assert `ReportSecurityFault` binds operation/local-fault/body/evidence/credential/effect digests, immediately quarantines and disables the node, saves normalized resume intent, increments identity epoch where required, revokes the epoch, supersedes pending issuance, invalidates grants, creates/aggregates the incident and stores a nondeliverable receipt. Prove ordinary response reauthorization now fails while only task 6's specialized commit gate can emit the exact finalized receipt.

Run: `go test ./internal/nodecontrol/recovery -run TestSecurityFaultSelfInvalidatingSaga -count=1`

Expected: FAIL because the saga is absent.

- [ ] **Step 6 (5 min): Implement reserve, fail-closed commit, finalize, and receipt activation**

Call `authority.Coordinator.Reserve` for the Batch 01 security-fault effect before opening the transaction. In the first short transaction re-lock the exact credential/node, record the reservation, idempotently materialize the binding and all fail-closed mutations, write audit/outbox, and persist the immutable nondeliverable receipt intent/effect. After commit call `authority.Coordinator.Finalize(operation_id,effect_digest)` outside all locks; the recovery resolver supplies the exact immutable incident/receipt effect and only the coordinator captures/binds database coordinates and activates the provider receipt. In a final transaction require the unchanged operation/credential/effect/incident/request binding and that exact committed receipt, then mark only that receipt deliverable. Timeout or uncertainty leaves the node quarantined and recovers through `Coordinator.Recover(operation_id)`; it never returns a guessed ACK.

Run: `go test ./internal/nodecontrol/recovery -run TestSecurityFaultSelfInvalidatingSaga -count=1`

Expected: PASS; the returned object is only `SecurityFaultReceiptV1{result:"accepted"}` and contains no state/root/time payload.

- [ ] **Step 7 (5 min): Write RED stopped-recovery, attestation, and identity/admin flow tests**

Test recovery poll with exact recovery ID, sorted unique known incident IDs, high-water and 32-byte nonce: a strict subset returns the complete newer set, an unknown extra ID or same generation/sequence with another digest conflicts, and administrative-disable/authority-restore may have an empty incident set. Test proof-of-possession over the fresh server nonce and exact recovery certificate, all-slots-stopped, guard/build/time digests, attestation retry, disable/retire resume normalization, reenroll epoch/lineage/grant creation, `recovery_pending` activation, no-old-pending-issuance, and completion that remains operator-disabled.

Run: `go test ./internal/nodecontrol/recovery -run 'TestStoppedRecovery|TestRecoveryAttestation|TestRecoveryIdentityFlow' -count=1`

Expected: FAIL because recovery actions and the B02 guard are absent.

- [ ] **Step 8 (5 min): Implement the stopped recovery state machine**

Implement identity-compromise, administrative-disable, retire, authority-restore and security-incident sessions. `Disable` performs the fail-closed identity transaction and creates a nonterminal session; retire and authority restore force saved resume state to disabled. `Reenroll` creates a new epoch/lineage CSR-bound one-time grant while disabled. `PollRecovery` signs only bounded `RecoveryStateSnapshotV1`; `SubmitRecoveryAttestation` verifies exact nonce/snapshot/certificate proof and stores one digest. `CompleteReenrollment` requires fresh accepted remediation evidence, the exact recovery-pending certificate and no old pending issuance, activates it as active or recovery-limited according to the remaining incident set, completes only the exact session, and never resumes ordinary operator state.

Run: `go test ./internal/nodecontrol/recovery -run 'TestStoppedRecovery|TestRecoveryAttestation|TestRecoveryIdentityFlow' -count=1`

Expected: PASS; every success still leaves operator state disabled and every failed prerequisite remains fail closed.

- [ ] **Step 9 (5 min): Write RED host-incident registration and per-incident clear tests**

Verify `RegisterHostSecurityIncident` consumes fresh exact node/incident/action evidence once and materializes or reconciles the typed server incident rather than treating absence as cleared. For `ClearSecurityQuarantine`, cover subtype-specific prerequisites, exact stopped snapshot/attestation/evidence digests, wrong incident, evidence reuse, another concurrent open incident, no local latch binding, local and supervisor bindings, response loss, and clearing one incident without broadening another.

Run: `go test ./internal/nodecontrol/recovery -run 'TestRegisterHostSecurityIncident|TestClearSecurityQuarantine' -count=1`

Expected: FAIL because per-incident recovery actions are absent.

- [ ] **Step 10 (5 min): Implement two-phase latch clear and explicit resume**

Without a local/supervisor binding, finalize only the exact incident after machine-verifying its subtype evidence. With bindings, first move it to `resolution_pending_agent_ack` and publish a strictly higher recovery generation with `required_action=clear_security_latches`, exact fault sets and remediation digest; only a later attestation of both new guard counters/digests finalizes resolution. The last resolution may set security normal and the certificate active but keeps operator disabled. `ResumeAfterSecurity` requires no open incident, fresh accepted attestation, exact active certificate and a target no broader than saved intent, then uses the B02 guard/signer/fence saga; only its activation transaction restores the explicit target.

Run: `go test ./internal/nodecontrol/recovery -run 'TestRegisterHostSecurityIncident|TestClearSecurityQuarantine|TestResumeAfterSecurity' -count=1`

Expected: PASS; signer failure or a concurrent security mutation leaves the node disabled and supersedes the pending resume intent.

- [ ] **Step 11 (5 min): Write RED two-person authority-restore reauthorization tests**

Start from a completed authority-restore session with the node disabled and old resume/desired discarded. Require proposal then approval from different operator IDs and exact credentials, equal authority epoch/scope/effect/target/evidence/If-Match, fresh uncached authorization at each phase, and deadline `min(proposal.created_at+15m, evidence.completed_at+15m)`. Expire/revoke either credential, shrink POP/role scope, change inventory/profile/session/evidence/effect or attempt one credential in both roles; require atomic supersede of proposal, approval and signing intent.

Run: `go test ./internal/nodecontrol/recovery -run TestReauthorizeAfterRestoreTwoPerson -count=1`

Expected: FAIL because restore proposal/approval activation is absent.

- [ ] **Step 12 (5 min): Implement proposal, approval, and finalized restore activation**

Persist one-time proposal/approval rows containing both exact credential snapshots and a shared effect digest; the approver cannot change proposal fields. Only after approval create the B02 `restore_reauthorize` transition/signing intent. At activation, re-lock session/inventory/POP/profile/evidence/approval rows, call the external authorizer uncached for each captured credential outside database locks, re-enter the final transaction, require unchanged captures and a committed authority receipt, publish a new higher desired generation selecting only enabled or draining, then complete the session. Never reuse backup desired or resume intent.

Run: `go test ./internal/nodecontrol/recovery -run TestReauthorizeAfterRestoreTwoPerson -count=1`

Expected: PASS; missing, stale, same-operator or scope-mismatched approval never reaches signer-visible activation.

- [ ] **Step 13 (5 min): Exercise crash and concurrency recovery at every boundary**

Crash before/after reserve, fail-closed transaction, provider finalize, receipt activation, recovery-sign prepare/sign/fence/activation, `resolution_pending_agent_ack`, latch-clear attestation, server resolution, restore proposal, approval and final activation. Race duplicate fault IDs, two reenrolls, incident overflow, clear versus new incident, resume versus disable and restore approval versus credential revocation. Retry exact operation IDs and prove one terminal audit outcome, no reopened authority and no duplicate generation/grant/receipt.

Run: `go test -tags=integration ./internal/nodecontrol/recovery -run 'TestRecoveryCrashMatrix|TestRecoveryConcurrency' -count=1 -timeout 6m`

Expected: PASS with no provider/authorizer/signer call under a PostgreSQL lock.

- [ ] **Step 14 (4 min): REFACTOR finite errors, retention, and secret-free telemetry**

Map only invalid request, unauthenticated, forbidden, not found, conflict, rate limited, dependency unavailable and internal. Logs/audit/metrics may contain finite subtype/status and opaque IDs, never grant, CSR, certificate DER, proof, remediation signature, guard digest, provider/SQL text or free-form reason. Retain unresolved/session/evidence/approval rows while referenced; delete only terminal unreferenced rows after the Batch 01 windows.

Run: `go test -race ./internal/nodecontrol/recovery ./internal/nodecontrol/hostevidence -count=1`

Expected: PASS with privacy canaries, stable finite reason cardinality and no data race.

- [ ] **Step 15 (2 min): Commit recovery and security workflows**

```bash
git add db/queries/nodecontrol_recovery.sql internal/store/nodecontrol_recovery.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/hostevidence/remediation.go internal/nodecontrol/recovery/types.go internal/nodecontrol/recovery/repository.go internal/nodecontrol/recovery/postgres_repository.go internal/nodecontrol/recovery/security_fault.go internal/nodecontrol/recovery/operator_actions.go internal/nodecontrol/identity/repository.go internal/nodecontrol/identity/postgres_repository.go internal/nodecontrol/identity/service.go internal/nodecontrol/hostevidence/remediation_test.go internal/nodecontrol/recovery/postgres_repository_integration_test.go internal/nodecontrol/recovery/security_fault_test.go internal/nodecontrol/recovery/operator_actions_test.go internal/nodecontrol/recovery/service_crash_test.go testdata/c12/host-remediation-vectors.json
git commit -m "feat(nodecontrol): recover quarantined node identities"
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
- Consumes: generated `nodebootstrapv1`, `nodeagentv1`, `nodeoperatorv1` strict server interfaces; task 3 identity service; task 6 authorizers/waiters; task 7 recovery service; Batch 02 inventory/operator/state services. B07-T03 replaces the exact fail-closed observation implementation defined below.
- Produces: generated-interface-complete handlers, bounded decoder/encoder, finite public error mapper, ordinary response-commit gate and the security-fault specialized receipt-commit path.

```go
type BootstrapService interface {
	Claim(context.Context, identity.ClaimRequest) (identity.IssuanceResponse, error)
}

var ErrObservationDependencyUnavailable = errors.New("observation dependency unavailable")

type ObservationService interface {
	Accept(context.Context, NodeAuthorization, nodeagentv1.NodeObservationV1) (nodeagentv1.ObservationAckV1, error)
}

type AgentStateService interface {
	PollDesired(context.Context, NodeAuthorization, nodeagentv1.DesiredPollRequestV1) (nodeagentv1.NodeStatePollResponseV1, bool, error)
}

type UnavailableObservationService struct{}

func (UnavailableObservationService) Accept(_ context.Context, _ NodeAuthorization, _ nodeagentv1.NodeObservationV1) (nodeagentv1.ObservationAckV1, error) {
	return nodeagentv1.ObservationAckV1{}, ErrObservationDependencyUnavailable
}

type AgentServices struct {
	Identity      identity.IdentityService
	State         AgentStateService
	Recovery      recovery.Service
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
	Authorizer    operator.OperatorAuthorizer
	Authenticator OperatorPeerAuthenticator
}
```

`AgentStateService.PollDesired` returns `changed=true` only for a `200` payload assembled from finalized active pointers in root-chain → metadata → desired → optional time-attestation order; `changed=false` means an empty `204`. It never returns pending signer bytes, and it obtains any nonce-bound time attestation through B02 without reserving a new authority sequence.

- [ ] **Step 1 (5 min): Write RED generated-interface and body-limit tests**

Add compile-time assertions against `nodebootstrapv1.ServerInterface`, `nodeagentv1.ServerInterface` and `nodeoperatorv1.ServerInterface`. Test ordinary 64 KiB bodies, conflict-evidence 1 MiB/two-artifact exception, oversized Content-Length, unknown-length streamed overflow, truncated/chunked body, duplicate/unknown JSON, invalid UTF-8, trailing JSON and non-identity Content-Encoding. Assert `UnavailableObservationService` returns the fixed `dependency_unavailable` public result without reading or persisting observation data; B07-T03 must replace this injected value before observation readiness can become true.

- [ ] **Step 2 (2 min): Run the focused RED handler tests**

Run: `go test ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi -run 'TestGeneratedInterface|TestBodyLimits' -count=1`

Expected: FAIL because handlers do not exist.

- [ ] **Step 3 (5 min): Implement bootstrap claim adapter**

Apply source rate limit before body read, bound wire and decoded size to 64 KiB, strict-decode canonical attempt/node IDs, 32-byte grant/time nonce and CSR DER, call `IdentityService.Claim`, and return only leaf, issuer chain, authorization receipt, node-state metadata and time attestation. Never include other node/operator data or provider detail.

- [ ] **Step 4 (5 min): Implement agent route adapters and endpoint classes**

Map rotate, desired poll, recovery poll, observation, security fault, trust-conflict evidence and recovery attestation to the exact class from task 6. Authorize before full body/heavy dependency, use per-node finite rate limits, register poll waiters, and invoke injected domain service only after strict validation. Recovery poll/attestation call `recovery.Service`; observation calls the injected service and therefore fails closed until B07-T03; security fault calls `ReportSecurityFault`, discards every non-receipt value, then uses only `AuthorizeSecurityFaultReceiptCommit` with the task 7 immutable binding. It must not call ordinary `ReauthorizeResponse` after the first transaction revoked the presenting credential. No class may fall through to ordinary authorization.

- [ ] **Step 5 (5 min): Implement operator route adapters**

Authenticate exact operator leaf, call external authorization with exact OpenAPI action/target, require `Idempotency-Key` and finite reason on mutations plus `If-Match` on updates. Bind list cursor reopening to new authorization, 2-second database timeout, item/page byte budget and at most 200 items. Map all nine action operations without a default branch: `drainNode -> State`; `disableNode -> Recovery.Disable`; `reenrollNode -> Recovery.Reenroll`; `completeNodeReenrollment -> Recovery.CompleteReenrollment`; `registerNodeHostSecurityIncident -> Recovery.RegisterHostSecurityIncident`; `registerNodeResourceEnvelope -> Mutations.Execute` with exact `ActionRegisterResourceEnvelope`; `clearNodeSecurityQuarantine -> Recovery.ClearSecurityQuarantine`; `resumeNodeAfterSecurity -> Recovery.ResumeAfterSecurity`; `reauthorizeNodeAfterRestore -> Recovery.ReauthorizeAfterRestore`. The latter eight require fresh uncached security-admin authorization; `drainNode` alone permits the writer role. Proposal and approval phases each reauthenticate and reauthorize the exact credential.

- [ ] **Step 6 (4 min): Implement finite public error mapping**

Emit only `invalid_request`, `unauthenticated`, `forbidden`, `not_found`, `conflict`, `rate_limited`, `dependency_unavailable`, `internal`. Use a fixed response schema/request ID; log only low-cardinality internal reason. Provider, TLS, SQL, key, grant, CSR, certificate and filesystem text never reaches JSON.

- [ ] **Step 7 (4 min): Add response-commit authorization for 200 and 204**

Prepare payload in bounded memory, reauthorize immediately before headers, then use `http.NewResponseController` to set a response-phase write deadline of trusted/monotonic now plus 10 seconds. For long poll, the listener has no global write timeout; cancellation or stale auth discards root/metadata/state/time bytes. The only exception is the task 7 self-invalidating security-fault route, which invokes the specialized exact receipt gate and can commit only its fence-finalized `SecurityFaultReceiptV1`.

- [ ] **Step 8 (3 min): Run GREEN handler tests**

Run: `go test ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi -count=1`

Expected: PASS with all generated methods implemented and every malformed/oversized request bounded.

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
- Test: `internal/platform/tlsserver_integration_test.go`

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

Run: `go test -tags=integration ./internal/platform -run 'TestNodeListenerSlowClients|TestAgentLongPollDeadline' -count=1 -timeout 3m`

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

### Task 10: Wire production configuration, readiness, and the Batch 03 mTLS gate

**Files:**
- Create: `internal/config/nodecontrol.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `cmd/control-api/main.go`
- Modify: `cmd/control-api/main_test.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`
- Create: `internal/e2e/nodecontrol_mtls_test.go`

**Interfaces:**
- Consumes: task 3 identity service, task 4 trust packages/deployment role keys, task 7 recovery repository/service/host-remediation verifier, task 8 handlers, task 9 listeners, Batch 02 services/providers, Batch 01 authority readiness, existing public/metrics runtime.
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

Reject duplicate listener addresses/DNS names/certificate refs, bootstrap client CA, missing agent/operator client package, reused node/operator CA, reused trust domains, missing deployment-authority key-set provider, provider ID reuse across incompatible roles, missing production provider, `LocalTest`/`Deterministic` production provider, C1.1 config signer/trust path, test private key and enabled listener with authority not ready. Assert the key-set provider returns four distinct role-bound Ed25519 keys and cannot alias the fence, issuer, signer, authorizer, trust-guard or trusted-time provider identity.

- [ ] **Step 2 (2 min): Run the focused RED config tests**

Run: `go test ./internal/config ./cmd/control-api -run 'TestNodeControlConfig|TestNodeControlStartup' -count=1`

Expected: FAIL because node-control config and wiring do not exist.

- [ ] **Step 3 (5 min): Implement strict configuration validation**

Require three explicit non-wildcard listener addresses distinct from public/metrics, three controlled DNS names, purpose-matched trust packages, node/operator trust domains that are valid DNS names and unequal, and all external provider IDs. Production profile rejects test implementations by registered capability metadata, not only by substring. Secret refs are opaque resolver IDs; raw PEM/private key/grant fields are absent from config schema.

- [ ] **Step 4 (5 min): Construct domain services before listeners**

Resolve the configured deployment-authority key-set provider and validate its exact four role identities before constructing the trust-package and `HostRemediationVerifier` consumers. Resolve authority provider/repository, issuer, signer/root providers, operator authorizer/guard and trusted time; construct the Batch 01 effect dispatcher from identity/state/recovery read-only resolvers and then one shared `authority.Coordinator`. Construct `recovery.NewPostgresRepository`, inject task 2 `RecoveryIdentityRepository`, build `recovery.Service`/B02 recovery guard, then inject that same service and coordinator into both domain composition roots. Build identity/inventory/state services and all three handlers. Do not call `Listen` until every constructor and authority/trust/recovery readiness comparison passes; production may not substitute `UnavailableObservationService` once B07 observation readiness is declared.

- [ ] **Step 5 (4 min): Add one shared fail-closed lifecycle**

Start bootstrap, agent and operator listeners only after the authority probe is ready. A provider/DB mismatch, unexplained reservation, trust high-water rollback/fork, deployment key-set/remediation verifier failure, corrupt pending recovery operation or provider unavailability marks node-control readiness false, stops new accepts, cancels waiters/signing/recovery operations, and closes node transports. Existing public/metrics servers follow their own lifecycle and cannot make node listeners ready.

- [ ] **Step 6 (3 min): Run GREEN config and startup tests**

Run: `go test ./internal/config ./cmd/control-api ./internal/readiness -run 'TestNodeControlConfig|TestNodeControlStartup|TestAuthorityReadiness' -count=1`

Expected: PASS; no listener starts in every invalid case and shutdown completes once.

- [ ] **Step 7 (5 min): Add a live three-listener mTLS integration test**

Use reserved local-test trust domains, three server certificates and distinct node/operator client CAs. Assert bootstrap claim works without a client cert and its handler sees no peer identity; agent rotate reaches handler only with an exact active node row; operator list reaches handler only with strict operator leaf plus external authorization. Exercise the recovery-pending poll/attestation path and the specialized self-invalidating security-fault receipt gate. Route all nine operator actions to their exact task 8 dependency and prove the eight security-admin actions use uncached authorization. Node cert on operator, operator cert on agent, no cert on either mTLS listener, wrong DNS and TLS 1.2 all fail.

Run: `go test -tags=integration ./internal/e2e -run TestNodeControlThreeListenerMTLS -count=1 -timeout 5m`

Expected: PASS.

- [ ] **Step 8 (5 min): Add revoke, expiry, bundle change, and authority-PITR cases**

Reuse live HTTP/2 connections, then revoke certificate, increment identity epoch, expire trusted time, compromise/remove issuer CA and restore a pre-revocation database while keeping provider head. Assert the next request/response commit fails, existing transports close, and all node listeners stay unavailable on PITR mismatch. Materialize authority-restore sessions, prove stopped reenrollment remains disabled, and require two distinct fresh uncached security-admin credentials plus fresh host-remediation evidence before a new fenced desired generation can reauthorize a node.

Run: `go test -tags=integration ./internal/e2e -run 'TestNodeControlKeepAliveRevocation|TestNodeControlPITRFailsClosed' -count=1 -timeout 5m`

Expected: PASS; no root/metadata/desired/receipt bytes are emitted after invalidation.

- [ ] **Step 9 (4 min): REFACTOR startup diagnostics and cleanup**

Expose only fixed readiness component names `node_authority`, `node_bootstrap`, `node_agent`, `node_operator`, `node_issuer`, `node_signer`, `node_trust`, `node_recovery`, `node_host_remediation`; details remain internal and secret-free. Test partial startup failure closes listeners and releases ports/permits without masking the primary error.

Run: `go test ./cmd/control-api ./internal/readiness ./internal/platform -count=1`

Expected: PASS.

- [ ] **Step 10 (4 min): Run deterministic generation and package gates**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS followed by `git diff --exit-code -- api gen internal/store` exiting 0.

Run: `go test ./internal/nodecontrol/identity ./internal/nodecontrol/hostevidence ./internal/nodecontrol/recovery ./internal/nodebootstrapapi ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/platform ./internal/config ./cmd/control-api -count=1 -timeout 8m`

Expected: PASS.

- [ ] **Step 11 (5 min): Run the Batch 03 integration exit gate**

Run: `go test -tags=integration ./internal/nodecontrol/identity ./internal/nodecontrol/hostevidence ./internal/nodecontrol/recovery ./internal/nodeagentapi ./internal/nodeoperatorapi ./internal/platform ./internal/e2e -count=1 -timeout 10m`

Expected: PASS with issuance/recovery crash recovery, specialized security receipt validation, host-remediation role/freshness/replay checks, administrative-disable and identity-compromise stopped reenrollment, two-phase latch clear, two-person authority-restore reauthorization, trust-package rollback/fork, operator guard failure, exact peer authorization, slow-client caps, live three-listener mTLS and authority PITR coverage.

- [ ] **Step 12 (2 min): Commit runtime wiring and Batch 03 gate**

```bash
git add internal/config/nodecontrol.go internal/config/config.go internal/config/config_test.go cmd/control-api/main.go cmd/control-api/main_test.go internal/readiness/checker.go internal/readiness/checker_test.go internal/e2e/nodecontrol_mtls_test.go
git commit -m "feat(control-api): start isolated node mtls listeners"
```

## Batch 03 completion check

- Grant is 32 crypto-random bytes, stored only as digest, CSR-bound, 10-minute, one-time and never re-disclosed on retry.
- Issuer request binds immutable issuance/issuer/public key/template digest; issuer runs outside locks and its output passes independent exact X.509 verification.
- Enrollment/rotation recovery returns one fence-finalized exact certificate and `CertificateAuthorizationReceiptV1`; late old-epoch results cannot activate.
- Node lineages last at most 30 days, rotate before `NotAfter-4h`, require reenroll below 36 hours and cap overlap at 4 active unexpired leaves.
- Deployment key roles are four distinct Ed25519 identities; trust packages enforce purpose/domain/version/sequence/digest/cumulative deauthorization and same-value fork rejection.
- Operator guard requires a fresh nonce-bound, at-most-5-minute rollback-resistant attestation and closes existing transport on failure or package invalidation.
- Host remediation accepts only the configured deployment-authority `host_remediation` role, exact transcript, 15-minute freshness and one-time node/incident/action binding; the deployment key-set provider identity is distinct from every fence/issuer/signer/authorizer/time provider.
- Security-fault reporting applies quarantine/revocation before provider finalization and can return only the exact fence-finalized `SecurityFaultReceiptV1` through the specialized commit gate; response loss never reauthorizes the old credential.
- Identity compromise, administrative disable, retire and authority restore follow disabled -> new epoch/lineage -> recovery-pending certificate -> stopped snapshot/attestation -> complete while operator state remains disabled; per-incident clear and explicit resume are separate transitions.
- Authority-restore reauthorization requires fresh host evidence and two distinct fresh uncached security-admin exact credentials bound to one epoch/scope/effect within 15 minutes, then publishes a new fenced desired generation without backup desired/resume reuse.
- Every agent/operator request and response commit matches exact leaf DER/key/issuer/serial/status, authority/identity epoch, trusted time and endpoint-specific node state.
- All nine operator action operations have explicit service/action mappings; no unknown action falls through, and observation returns `dependency_unavailable` until B07-T03 replaces the injected fail-closed service.
- Bootstrap, agent and operator are distinct TLS 1.3 h2 listeners with exact client-auth modes, finite connections/streams/timeouts and working long-poll response deadlines.
- Authority/trust mismatch or PITR keeps all node-control listeners fail closed; no C1.1 trust/config signer is accepted as C1.2 authority.
- No Batch 04 keystore, RollbackGuard persistence, outbound agent loop or local reconciliation production behavior has entered this batch.
