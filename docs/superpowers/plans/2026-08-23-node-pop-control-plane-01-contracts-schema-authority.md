# Talenro C1.2 Batch 01 Contracts, Schema, and Authority Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付三份隔离的 node OpenAPI、nodecontrol Protobuf event、完整 C1.2 PostgreSQL 权威 schema，以及可在 crash/PITR 场景中确定性验证的 `ControlPlaneAuthorityFence` provider client、repository 和 local/test fake。

**Architecture:** 契约先行：OpenAPI、Protobuf 和 `internal/nodecontrol/contracts` 冻结所有跨分册 wire/domain 名称；`db/migrations/00006_nodecontrol.sql` 一次建立互不合并的 C1.2 权威表。Fence provider 只保存 epoch/sequence/effect receipt，PostgreSQL repository 只保存其数据库可见性镜像；后续领域 workflow 在自己的短事务中调用 transaction-bound fence repository，provider 调用始终发生在锁外。

**Tech Stack:** Go 1.26.5、OpenAPI 3.0.3、oapi-codegen 2.8.0、Protobuf Go 1.36.11、Buf 1.72.0、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、goose 3.27.1、RFC 8785 JCS、SHA-256、现有 strictjson/outbox 约束。

**Spec:** [Approved C1.2 node and POP control-plane design](../specs/2026-08-23-node-pop-control-plane-design.md), approved working-tree SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`.

## Global Constraints

- 本册只完成索引中的 `C1.2-B01`；不得实现 operator mutation、state signing、certificate issuance、agent、supervisor 或 core adapter。
- [Suite index](2026-08-23-node-pop-control-plane.md) 冻结的路径与跨册接口逐字生效；本册不得重命名 `contracts.AuthorityVersion`、`authority.Provider` 或三个生成包。
- C1.1 `internal/trust`、C1.1 root schema、config signer key 与 `securitykit.EnrollmentGrantToken` 不得成为 C1.2 authority。
- PostgreSQL 是领域事实源；production anti-rollback authority 必须在独立 provider，普通表、文件、wall clock 和 Redis 都不能冒充 provider。
- Access-granting effect 在 provider committed receipt 和最终 visibility transaction 前不可被 serving query、listener authorizer、signer 或 active pointer读取。
- Fail-closed effect 可以先写数据库，但 provider finalize 前不得返回成功；任何不确定结果保持不可服务并按 operation ID 恢复。
- OpenAPI 普通 request body 最大 64 KiB；`trust-conflict-evidence` 是唯一 1 MiB、最多两个 artifact 的例外。
- 所有 JSON 拒绝未知字段、重复字段、无效 UTF-8、尾随内容、非法枚举、越界集合和非 canonical UUID。
- Nodecontrol event 不得携带 endpoint address、grant、CSR、certificate serial/DER、signed desired bytes、credential、core output 或 raw capacity sample。
- 所有生成物提交仓库；`scripts/generate.ps1` 与 `scripts/generate.sh` 必须确定性生成三份 OpenAPI、Protobuf 和 sqlc artifact。
- 每个 task 先 RED，再 GREEN，再 REFACTOR；每个 task 使用独立 commit，且只 stage 该 task 列出的路径。
- 用户未跟踪目录 `.cache/`、`.superpowers/`、`.task19-go/` 不得 stage、删除或作为测试根。

---

## Frozen file ownership

```text
api/openapi/node-bootstrap-api.v1.yaml
api/openapi/node-bootstrap-oapi-codegen.yaml
api/openapi/node-agent-api.v1.yaml
api/openapi/node-agent-oapi-codegen.yaml
api/openapi/node-operator-api.v1.yaml
api/openapi/node-operator-oapi-codegen.yaml
api/proto/talenro/nodecontrol/v1/events.proto
gen/go/talenro/nodebootstrap/v1/server.gen.go
gen/go/talenro/nodeagent/v1/server.gen.go
gen/go/talenro/nodeoperator/v1/server.gen.go
gen/go/talenro/nodecontrol/v1/events.pb.go
db/migrations/00006_nodecontrol.sql
db/queries/nodecontrol_authority.sql
internal/store/nodecontrol_authority.sql.go
internal/nodecontrol/contracts/authority.go
internal/nodecontrol/contracts/events.go
internal/nodecontrol/authority/provider.go
internal/nodecontrol/authority/deterministic_fake.go
internal/nodecontrol/authority/repository.go
internal/nodecontrol/authority/postgres_repository.go
```

`db/queries/nodecontrol_inventory.sql`、`nodecontrol_state.sql`、`nodecontrol_identity.sql` 和 `nodecontrol_observation.sql` 由后续对应分册创建；本册 migration 必须已经建立它们将使用的表。

### Task 1: Freeze pure authority comparison contracts

**Files:**
- Create: `internal/nodecontrol/contracts/authority.go`
- Test: `internal/nodecontrol/contracts/authority_test.go`
- Test: `internal/nodecontrol/contracts/authority_fuzz_test.go`

**Interfaces:**
- Consumes: `errors.New`, fixed-width `[32]byte` values, unsigned integer ordering.
- Produces: `type Digest [32]byte`; `type AuthorityVersion struct { Epoch uint64; Sequence uint64 }`; `type VersionedDigest struct { Version uint64; AuthoritySequence uint64; Digest Digest }`; `func (AuthorityVersion) Validate() error`; `func CompareVersionedDigest(VersionedDigest, VersionedDigest) (Comparison, error)`.

- [ ] **Step 1: Write the RED table test for validation, advance, rollback, and fork**

```go
package contracts

import (
	"errors"
	"testing"
)

func TestCompareVersionedDigestClosedOutcomes(t *testing.T) {
	digestA := Digest{1}
	digestB := Digest{2}
	cases := []struct {
		name      string
		current   VersionedDigest
		candidate VersionedDigest
		want      Comparison
		wantErr   error
	}{
		{name: "same", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, want: ComparisonSame},
		{name: "advance", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 9, Digest: digestB}, want: ComparisonAdvance},
		{name: "version rollback", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 3, AuthoritySequence: 9, Digest: digestB}, want: ComparisonRollback},
		{name: "sequence rollback", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 7, Digest: digestB}, want: ComparisonRollback},
		{name: "same coordinate fork", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestB}, want: ComparisonFork},
		{name: "crossed coordinate fork", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 9, Digest: digestB}, want: ComparisonFork},
		{name: "zero digest", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 9}, wantErr: ErrInvalidAuthorityValue},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := CompareVersionedDigest(test.current, test.candidate)
			if !errors.Is(err, test.wantErr) || got != test.want {
				t.Fatalf("comparison = %q, %v; want %q, %v", got, err, test.want, test.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run the focused RED test**

Run: `go test ./internal/nodecontrol/contracts -run TestCompareVersionedDigestClosedOutcomes -count=1`

Expected: FAIL because package symbols `Digest`, `VersionedDigest`, and `CompareVersionedDigest` do not exist.

- [ ] **Step 3: Add the minimal closed types and exhaustive comparison**

```go
package contracts

import (
	"errors"
	"math"
)

type Digest [32]byte

type AuthorityVersion struct {
	Epoch    uint64
	Sequence uint64
}

type VersionedDigest struct {
	Version           uint64
	AuthoritySequence uint64
	Digest            Digest
}

type Comparison string

const (
	ComparisonSame     Comparison = "same"
	ComparisonAdvance  Comparison = "advance"
	ComparisonRollback Comparison = "rollback"
	ComparisonFork     Comparison = "fork"
)

var ErrInvalidAuthorityValue = errors.New("nodecontrol contracts: invalid authority value")

func (value AuthorityVersion) Validate() error {
	if value.Epoch == 0 || value.Sequence == 0 ||
		value.Epoch > math.MaxInt64 || value.Sequence > math.MaxInt64 {
		return ErrInvalidAuthorityValue
	}
	return nil
}

func CompareVersionedDigest(current, candidate VersionedDigest) (Comparison, error) {
	if current.Version == 0 || current.Version > math.MaxInt64 ||
		current.AuthoritySequence == 0 || current.AuthoritySequence > math.MaxInt64 ||
		current.Digest == (Digest{}) || candidate.Version == 0 || candidate.Version > math.MaxInt64 ||
		candidate.AuthoritySequence == 0 || candidate.AuthoritySequence > math.MaxInt64 ||
		candidate.Digest == (Digest{}) {
		return "", ErrInvalidAuthorityValue
	}
	if candidate.Version < current.Version || candidate.AuthoritySequence < current.AuthoritySequence {
		return ComparisonRollback, nil
	}
	if candidate.Version == current.Version && candidate.AuthoritySequence == current.AuthoritySequence {
		if candidate.Digest == current.Digest {
			return ComparisonSame, nil
		}
		return ComparisonFork, nil
	}
	if candidate.Version > current.Version && candidate.AuthoritySequence > current.AuthoritySequence {
		return ComparisonAdvance, nil
	}
	return ComparisonFork, nil
}
```

- [ ] **Step 4: Run the focused GREEN test**

Run: `go test ./internal/nodecontrol/contracts -run TestCompareVersionedDigestClosedOutcomes -count=1`

Expected: PASS, including exact `math.MaxInt64` acceptance and `uint64(math.MaxInt64)+1` rejection for every PostgreSQL-backed version/sequence field.

- [ ] **Step 5: REFACTOR with a fuzz invariant that forbids unknown outcomes**

```go
func FuzzCompareVersionedDigestNeverInventsOutcome(f *testing.F) {
	f.Add(uint64(1), uint64(1), uint64(1), uint64(1), byte(1), byte(1))
	f.Fuzz(func(t *testing.T, cv, cs, nv, ns uint64, cd, nd byte) {
		current := VersionedDigest{Version: cv, AuthoritySequence: cs, Digest: Digest{cd}}
		candidate := VersionedDigest{Version: nv, AuthoritySequence: ns, Digest: Digest{nd}}
		got, err := CompareVersionedDigest(current, candidate)
		if err != nil {
			return
		}
		switch got {
		case ComparisonSame, ComparisonAdvance, ComparisonRollback, ComparisonFork:
			return
		default:
			t.Fatalf("unknown comparison %q", got)
		}
	})
}
```

- [ ] **Step 6: Format and run the complete contracts package**

Run: `gofmt -w internal/nodecontrol/contracts/authority.go internal/nodecontrol/contracts/authority_test.go internal/nodecontrol/contracts/authority_fuzz_test.go`

Expected: command exits 0.

Run: `go test ./internal/nodecontrol/contracts -count=1`

Expected: PASS with unit and fuzz seed coverage.

- [ ] **Step 7: Commit the authority primitives**

```bash
git add internal/nodecontrol/contracts/authority.go internal/nodecontrol/contracts/authority_test.go internal/nodecontrol/contracts/authority_fuzz_test.go
git commit -m "feat(nodecontrol): add authority comparison contracts"
```

### Task 2: Add three isolated OpenAPI contracts and deterministic generation

**Files:**
- Create: `api/openapi/node-bootstrap-api.v1.yaml`
- Create: `api/openapi/node-bootstrap-oapi-codegen.yaml`
- Create: `api/openapi/node-agent-api.v1.yaml`
- Create: `api/openapi/node-agent-oapi-codegen.yaml`
- Create: `api/openapi/node-operator-api.v1.yaml`
- Create: `api/openapi/node-operator-oapi-codegen.yaml`
- Create: `gen/go/talenro/nodebootstrap/v1/server.gen.go`
- Create: `gen/go/talenro/nodeagent/v1/server.gen.go`
- Create: `gen/go/talenro/nodeoperator/v1/server.gen.go`
- Modify: `scripts/generate.ps1`
- Modify: `scripts/generate.sh`
- Test: `internal/nodecontrol/contracts/openapi_contract_test.go`

**Interfaces:**
- Consumes: OpenAPI 3.0.3, existing `oapi-codegen` std-http server conventions, canonical JSON names from the approved spec.
- Produces: packages `nodebootstrapv1`, `nodeagentv1`, and `nodeoperatorv1`; their generated `ServerInterface`, typed clients (`ClientInterface`/`ClientWithResponsesInterface`), request/response models, parameter structs, `HandlerWithOptions`, and embedded specs.

Freeze these operation IDs:

| Package | Method and path | Operation ID |
| --- | --- | --- |
| bootstrap | `POST /v1/node-enrollments/claim` | `claimNodeEnrollment` |
| agent | `POST /v1/node-agent/certificates/rotate` | `rotateNodeCertificate` |
| agent | `POST /v1/node-agent/desired-state:poll` | `pollNodeDesiredState` |
| agent | `POST /v1/node-agent/recovery-state:poll` | `pollNodeRecoveryState` |
| agent | `POST /v1/node-agent/observations` | `createNodeObservation` |
| agent | `POST /v1/node-agent/security-faults` | `createNodeSecurityFault` |
| agent | `POST /v1/node-agent/trust-conflict-evidence` | `createNodeTrustConflictEvidence` |
| agent | `POST /v1/node-agent/recovery-attestations` | `createNodeRecoveryAttestation` |
| operator | `POST/GET /v1/operator/pops` and `GET/PUT /v1/operator/pops/{pop_code}` | `createNodePOP`, `listNodePOPs`, `getNodePOP`, `updateNodePOP` |
| operator | `POST/GET /v1/operator/failure-domains` and `GET/PUT /v1/operator/failure-domains/{failure_domain_id}` | `createNodeFailureDomain`, `listNodeFailureDomains`, `getNodeFailureDomain`, `updateNodeFailureDomain` |
| operator | `POST/GET /v1/operator/nodes` and `GET/PUT /v1/operator/nodes/{node_id}` | `createNode`, `listNodes`, `getNode`, `updateNode` |
| operator | `POST/GET /v1/operator/nodes/{node_id}/endpoints` and `GET/PUT /v1/operator/nodes/{node_id}/endpoints/{endpoint_id}` | `createNodeEndpoint`, `listNodeEndpoints`, `getNodeEndpoint`, `updateNodeEndpoint` |
| operator | `POST/GET /v1/operator/nodes/{node_id}/process-slots` and `GET/PUT /v1/operator/nodes/{node_id}/process-slots/{slot_id}` | `createNodeProcessSlot`, `listNodeProcessSlots`, `getNodeProcessSlot`, `updateNodeProcessSlot` |
| operator | `POST /v1/operator/nodes/{node_id}/enrollment-grants` | `createNodeEnrollmentGrant` |
| operator | `PUT /v1/operator/nodes/{node_id}/desired-state` | `putNodeDesiredState` |
| operator | each `/actions/<action>` path from spec §8.3 | camel-case operation matching `drainNode`, `disableNode`, `reenrollNode`, `completeNodeReenrollment`, `registerNodeHostSecurityIncident`, `registerNodeResourceEnvelope`, `clearNodeSecurityQuarantine`, `resumeNodeAfterSecurity`, `reauthorizeNodeAfterRestore` |

- [ ] **Step 1: Write the RED embedded-spec contract test**

```go
package contracts_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	nodeagentv1 "talenro.local/platform/gen/go/talenro/nodeagent/v1"
	nodebootstrapv1 "talenro.local/platform/gen/go/talenro/nodebootstrap/v1"
	nodeoperatorv1 "talenro.local/platform/gen/go/talenro/nodeoperator/v1"
)

func TestNodeOpenAPISpecsAreIsolatedAndValid(t *testing.T) {
	tests := []struct {
		name       string
		get        func() (*openapi3.T, error)
		wantPath   string
		forbidPath string
	}{
		{name: "bootstrap", get: nodebootstrapv1.GetSwagger, wantPath: "/v1/node-enrollments/claim", forbidPath: "/v1/node-agent/desired-state:poll"},
		{name: "agent", get: nodeagentv1.GetSwagger, wantPath: "/v1/node-agent/desired-state:poll", forbidPath: "/v1/operator/nodes"},
		{name: "operator", get: nodeoperatorv1.GetSwagger, wantPath: "/v1/operator/nodes", forbidPath: "/v1/node-enrollments/claim"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec, err := test.get()
			if err != nil {
				t.Fatal(err)
			}
			if err := spec.Validate(context.Background()); err != nil {
				t.Fatal(err)
			}
			if spec.Paths.Find(test.wantPath) == nil || spec.Paths.Find(test.forbidPath) != nil {
				t.Fatalf("listener path isolation failed")
			}
		})
	}
}
```

- [ ] **Step 2: Run the RED contract test**

Run: `go test ./internal/nodecontrol/contracts -run TestNodeOpenAPISpecsAreIsolatedAndValid -count=1`

Expected: FAIL because the three generated packages do not exist.

- [ ] **Step 3: Create all three specs with closed errors, bounds, and security semantics**

Each file starts with its own title/version and contains only its listener's paths. Use this exact finite error schema in all three files:

```yaml
openapi: 3.0.3
components:
  schemas:
    PublicErrorV1:
      type: object
      additionalProperties: false
      required: [code, request_id]
      properties:
        code:
          type: string
          enum: [invalid_request, unauthenticated, forbidden, not_found, conflict, rate_limited, dependency_unavailable, internal]
        request_id:
          type: string
          minLength: 16
          maxLength: 64
          pattern: '^[A-Za-z0-9_-]+$'
        reason:
          type: string
          enum: [malformed, precondition_failed, stale_version, reenroll_required, incident_capacity_exceeded, authority_unavailable, credential_invalid, scope_changed, operation_in_progress, deadline_expired, unsupported_capability]
        retry_after_ms:
          type: integer
          format: int64
          minimum: 1
          maximum: 10000
  parameters:
    IdempotencyKey:
      name: Idempotency-Key
      in: header
      required: true
      schema:
        type: string
        minLength: 22
        maxLength: 86
        pattern: '^[A-Za-z0-9_-]+$'
    IfMatch:
      name: If-Match
      in: header
      required: true
      schema:
        type: string
        pattern: '^"[1-9][0-9]{0,18}"$'
```

The bootstrap claim schema must bind the exact node, attempt, grant, CSR, and nonce:

```yaml
paths:
  /v1/node-enrollments/claim:
    post:
      operationId: claimNodeEnrollment
      security: []
      x-talenro-listener-auth: bootstrap_server_tls
      x-talenro-content-encoding: identity
      x-talenro-max-wire-bytes: 65536
      x-talenro-max-decoded-bytes: 65536
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              additionalProperties: false
              required: [attempt_id, node_id, enrollment_grant, csr_der, time_nonce]
              properties:
                attempt_id: {$ref: '#/components/schemas/CanonicalUUID'}
                node_id: {$ref: '#/components/schemas/CanonicalUUID'}
                enrollment_grant: {$ref: '#/components/schemas/Base64URL32'}
                csr_der: {$ref: '#/components/schemas/CSRDER'}
                time_nonce: {$ref: '#/components/schemas/Base64URL32'}
      responses:
        "200": {$ref: '#/components/responses/CertificateAuthorization'}
        "400": {$ref: '#/components/responses/InvalidRequest'}
        "409": {$ref: '#/components/responses/Conflict'}
        "429": {$ref: '#/components/responses/RateLimited'}
        "503": {$ref: '#/components/responses/DependencyUnavailable'}
```

OpenAPI remains 3.0.3, which has no standard `mutualTLS` security-scheme type. Do not encode mTLS as an API key or client-controlled header. Every operation carries exactly one runtime-enforced extension: `x-talenro-listener-auth: bootstrap_server_tls`, `node_mtls`, or `operator_mtls`; the generated handlers still receive peer identity only from the TLS middleware. Every request body also carries `x-talenro-content-encoding: identity`, `x-talenro-max-wire-bytes`, and `x-talenro-max-decoded-bytes`; ordinary values are 65,536/65,536 and trust-conflict evidence alone is 1,048,576/1,048,576.

Freeze these reusable schemas with `additionalProperties: false` on every object:

| Schema | Exact shape |
| --- | --- |
| `CanonicalUUID` | lowercase canonical UUID string, pattern `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$` |
| `DigestHex32` | lowercase SHA-256, pattern `^[0-9a-f]{64}$` |
| `Base64URL32` | unpadded base64url 32 bytes, length 43 and pattern `^[A-Za-z0-9_-]{43}$` |
| `CSRDER` | the exact YAML component below; strict RFC 4648 raw URL encoding, decoded length 1–16,384, and one signature-valid PKCS#10 DER object with no trailing bytes |
| `NodeHighWaterV1` | required positive `version_or_generation`, positive `authority_sequence`, `DigestHex32 digest` |
| `NodeActorHighWaterV1` | required positive `control_plane_authority_epoch`, nonnegative `node_authority_checkpoint`, required `server_ca_bundle/root/metadata`, optional `desired_seen/desired_applied/recovery_seen` `NodeHighWaterV1`; absent means that stream has never been observed and zero-valued fake objects are rejected |
| `PublicErrorV1` | the single YAML schema above: required `request_id` and closed `code`; optional closed `reason` and `retry_after_ms` 1–10,000; no second error schema/unit or detail/provider/error string |

`CSRDER` is copied verbatim into each spec that accepts a CSR:

```yaml
CSRDER:
  type: string
  minLength: 2
  maxLength: 21846
  x-talenro-base64url-padding: forbidden
  x-talenro-min-decoded-bytes: 1
  x-talenro-max-decoded-bytes: 16384
  oneOf:
    - pattern: '^(?:[A-Za-z0-9_-]{4})+$'
    - pattern: '^(?:[A-Za-z0-9_-]{4})*[A-Za-z0-9_-]{2}$'
    - pattern: '^(?:[A-Za-z0-9_-]{4})*[A-Za-z0-9_-]{3}$'
```

The handler decodes with `base64.RawURLEncoding.Strict()`, rejects `=`, whitespace and non-alphabet bytes, requires `RawURLEncoding.EncodeToString(decoded) == input`, enforces decoded length again, parses exactly one `x509.CertificateRequest` with no trailing ASN.1 bytes, and requires `CheckSignature()` to succeed before hashing or persistence. The three mutually exclusive regex branches reject impossible unpadded lengths congruent to one modulo four.

There is no generic operator-list wrapper. Freeze these five resource-specific list schemas verbatim; every referenced item schema is itself a fully enumerated object with `additionalProperties: false`, the required fields listed below, and the column-derived string/integer/time bounds. `OperatorListCursorV1` is an opaque AEAD value with `minLength: 16`, `maxLength: 2048`, and pattern `^[A-Za-z0-9_-]+$`.

```yaml
OperatorListCursorV1:
  type: string
  minLength: 16
  maxLength: 2048
  pattern: '^[A-Za-z0-9_-]+$'
NodePOPListV1:
  type: object
  additionalProperties: false
  required: [items]
  properties:
    items:
      type: array
      minItems: 0
      maxItems: 200
      items: {$ref: '#/components/schemas/NodePOPV1'}
    next_cursor: {$ref: '#/components/schemas/OperatorListCursorV1'}
FailureDomainListV1:
  type: object
  additionalProperties: false
  required: [items]
  properties:
    items:
      type: array
      minItems: 0
      maxItems: 200
      items: {$ref: '#/components/schemas/FailureDomainV1'}
    next_cursor: {$ref: '#/components/schemas/OperatorListCursorV1'}
NodeListV1:
  type: object
  additionalProperties: false
  required: [items]
  properties:
    items:
      type: array
      minItems: 0
      maxItems: 200
      items: {$ref: '#/components/schemas/NodeV1'}
    next_cursor: {$ref: '#/components/schemas/OperatorListCursorV1'}
NodeEndpointListV1:
  type: object
  additionalProperties: false
  required: [items]
  properties:
    items:
      type: array
      minItems: 0
      maxItems: 200
      items: {$ref: '#/components/schemas/NodeEndpointV1'}
    next_cursor: {$ref: '#/components/schemas/OperatorListCursorV1'}
NodeProcessSlotListV1:
  type: object
  additionalProperties: false
  required: [items]
  properties:
    items:
      type: array
      minItems: 0
      maxItems: 200
      items: {$ref: '#/components/schemas/NodeProcessSlotV1'}
    next_cursor: {$ref: '#/components/schemas/OperatorListCursorV1'}
```

The item schemas have these exact required sets: `NodePOPV1={pop_code,iso_country,region,operator_state,version,created_at,updated_at}`; `FailureDomainV1={failure_domain_id,domain_type,stable_id,version,created_at,updated_at}`; `NodeV1={node_id,pop_code,operator_state,security_state,identity_state,identity_epoch,inventory_version,security_version,created_at,updated_at}`; `NodeEndpointV1={endpoint_id,node_id,address,port,transport,protocol_capability,operator_state,inventory_version,created_at,updated_at}`; `NodeProcessSlotV1={node_id,slot_id,adapter,capacity_profile_id,capacity_profile_version,required,operator_state,inventory_version,created_at,updated_at}`. `NodeV1` may additionally contain only these explicitly nullable properties: `resume_operator_state`, `pending_operator_transition`, `pending_transition_signing_id`, `lineage_id`, `resource_envelope_version`, `resource_envelope_digest`, `active_desired_generation`, `next_desired_generation`, `active_recovery_generation`, `next_recovery_generation`, `active_root_publish_id`, `active_root_version`, `active_metadata_publish_id`, `active_metadata_version`, `last_authority_operation_id`, `last_authority_epoch`, and `last_authority_sequence`; each all-or-none pair/group is expressed with `oneOf` required/not-required branches rather than undocumented null combinations. Each list operation references exactly its matching list schema, defaults `page_size` to 50, rejects values outside 1–200, and has only its explicitly enumerated closed filters; no schema exposes `total`, `count`, `offset`, caller-selected `sort`, unbounded `include`, or an empty `items: {}` schema.

The operation-to-model matrix is exact:

| Operations | Request model | Success model | Auth/headers |
| --- | --- | --- | --- |
| `claimNodeEnrollment` | `ClaimNodeEnrollmentRequestV1={attempt_id,node_id,enrollment_grant,csr_der,time_nonce}` | `CertificateAuthorizationV1={leaf_der,issuer_chain,receipt,node_state_trust_metadata,time_attestation}` | bootstrap server TLS; source limiter runs before body read |
| `rotateNodeCertificate` | `RotateNodeCertificateRequestV1={attempt_id,csr_der,time_nonce}` | same `CertificateAuthorizationV1` | node mTLS, active exact certificate only |
| `pollNodeDesiredState` | `DesiredPollRequestV1={boot_id,agent_build_digest,capability_schema_digest,high_water,time_nonce?}` | `200 NodeStatePollResponseV1={root_chain?,metadata?,desired_state?,time_attestation?}` or empty `204` | node mTLS; 25-second wait; no delta |
| `pollNodeRecoveryState` | `RecoveryPollRequestV1={boot_id,recovery_id,sorted_known_incident_ids,high_water,time_nonce}` | `200 NodeRecoveryPollResponseV1={root_chain?,metadata?,recovery_state,time_attestation}` or empty `204` | node mTLS and exact recovery session/incident |
| `createNodeObservation` | `NodeObservationV1={node_id,identity_epoch,boot_id,sequence,observed_at,agent_build_digest,capability_schema_digest,seen_generation,seen_authority_sequence,seen_digest,applied_generation,applied_authority_sequence,applied_digest,sorted_slot_facts,reducer_version,reducer_digest,reducer_output}` | `ObservationAckV1={boot_id,sequence,report_digest,result}` | active node mTLS; at most 8 slot facts |
| `createNodeSecurityFault` | `SecurityFaultReportV1={operation_id,local_fault_id,subtype,evidence_digest,identity_epoch,boot_id,request_digest,supervisor_fault?}` | fixed `SecurityFaultReceiptV1` only | active/recovery-pending/recovery-limited node mTLS; specialized self-invalidating receipt gate |
| `createNodeTrustConflictEvidence` | `TrustConflictEvidenceRequestV1={incident_id,artifacts}` with one or two `SignedConflictArtifactV1={kind,canonical_bytes}` | `TrustConflictEvidenceAckV1={incident_id,result}` | exact incident-bound node mTLS; one allowed 1 MiB exception |
| `createNodeRecoveryAttestation` | `RecoveryAttestationV1={recovery_id,nonce,recovery_snapshot_digest,recovery_generation,certificate_id,proof_of_possession,agent_build_digest,supervisor_build_digest,agent_guard_digest,supervisor_guard_digest,latch_guard_digest,trusted_time_evidence_digest,all_slots_stopped}` | `RecoveryAttestationAckV1={recovery_id,attestation_digest,result}` | exact recovery node mTLS |
| POP/failure-domain/node `create` | exact `CreateNodePOPV1`, `CreateFailureDomainV1`, or `CreateNodeV1` | created resource + strong `ETag` | operator mTLS, `Idempotency-Key`; these three root creates alone omit `If-Match` |
| POP/failure-domain/node `get/list/update` | closed filter/update schema | resource/list + `ETag` where singular | operator mTLS; update requires both headers; list page size default 50/max 200 |
| endpoint/slot `create/get/list/update` | `Create/UpdateNodeEndpointV1` or `Create/UpdateNodeProcessSlotV1` | resource/list + parent/node `ETag` | operator mTLS; every create/update requires parent `If-Match` and `Idempotency-Key` |
| `createNodeEnrollmentGrant` | `CreateEnrollmentGrantV1={command_id,csr_der_sha256,reason_code}` | metadata plus one-time `enrollment_grant` only on first committed response | operator mTLS + both headers |
| `putNodeDesiredState` | `PutNodeDesiredStateV1={command_id,reason_code,requested_valid_until,inventory_version,resource_envelope_version,resource_envelope_digest,sorted_processes}` | `SigningOperationV1={signing_id,status,reserved_generation}` | operator mTLS + both headers |
| `drainNode` | `DrainNodeV1={command_id,reason_code}` | `SigningOperationV1` | operator mTLS + both headers |
| `disableNode` | `DisableNodeV1={command_id,reason_code}` where reason is `administrative_disable|retire` | `RecoveryOperationV1={recovery_id,status}` | security-admin operator mTLS + both headers |
| `reenrollNode` | `ReenrollNodeV1={command_id,recovery_id,csr_der_sha256,reason_code}` | `RecoveryEnrollmentGrantV1` | security-admin operator mTLS + both headers |
| `completeNodeReenrollment` | `CompleteReenrollmentV1={command_id,recovery_id,certificate_id,recovery_attestation_digest,host_remediation_evidence}` | `RecoveryOperationV1` | security-admin operator mTLS + both headers |
| `registerNodeHostSecurityIncident` | `RegisterHostSecurityIncidentV1={command_id,subtype,host_remediation_evidence}` | `RecoveryOperationV1` | security-admin operator mTLS + both headers |
| `registerNodeResourceEnvelope` | `RegisterResourceEnvelopeV1={command_id,envelope_package,host_remediation_evidence}` | `ResourceEnvelopeActivationV1={version,authority_sequence,digest}` | security-admin operator mTLS + both headers |
| `clearNodeSecurityQuarantine` | `ClearSecurityQuarantineV1={command_id,incident_id,host_remediation_evidence,recovery_snapshot_digest,recovery_attestation_digest}` | `RecoveryOperationV1` | security-admin operator mTLS + both headers |
| `resumeNodeAfterSecurity` | `ResumeAfterSecurityV1={command_id,recovery_id,target_operator_state,recovery_attestation_digest}` | `SigningOperationV1` | security-admin operator mTLS + both headers |
| `reauthorizeNodeAfterRestore` | `ReauthorizeAfterRestoreV1={command_id,recovery_id,phase,proposal_id?,target_operator_state,effect_digest,host_remediation_evidence}` where phase is `proposal|approval` and approval requires proposal ID | `RestoreReauthorizationV1={proposal_id,approval_id?,status,expires_at}` | security-admin operator mTLS + both headers; uncached authorization |

All operation schemas reference these normative enum registries rather than free text:

| Registry | Exact values |
| --- | --- |
| `OperatorReasonCodeV1` | `provision`, `inventory_update`, `capacity_change`, `drain_maintenance`, `administrative_disable`, `retire`, `identity_compromise`, `host_remediation`, `security_recovery`, `authority_restore`, `release_update` |
| `DesiredReasonV1` | `initial`, `operator_update`, `drain`, `resume`, `lease_refresh`, `clear_slot_quarantine`, `restore_reauthorize` |
| `FaultSubtypeV1` | `identity_compromise`, `online_signer_equivocation`, `metadata_rollback`, `root_rollback`, `root_equivocation`, `unverified_client_highwater_conflict`, `client_highwater_ahead`, `server_trust_bundle_conflict`, `trusted_time_rollback_or_unavailable`, `local_state_corruption_or_rollback`, `release_or_process_integrity`, `profile_binding_mismatch`, `incident_overflow` |
| `ProtocolCapabilityV1` | `fixture_loopback`, `xray_loopback`, `sing_box_loopback` |
| `AdapterV1` | `fixture`, `xray`, `sing_box` |
| `OperatorStateV1` | `provisioning`, `enabled`, `draining`, `disabled` |
| `SecurityStateV1` | `normal`, `quarantined` |
| `IdentityStateV1` | `never_enrolled`, `active`, `recovery_pending`, `recovery_limited`, `revoked` |
| `RecoveryReasonV1` | `identity_compromise`, `administrative_disable`, `retire`, `authority_restore`, `security_incident` |
| `RecoveryActionV1` | `hold_stopped`, `submit_recovery_attestation`, `clear_security_latches` |
| `RestorePhaseV1` | `proposal`, `approval` |
| `OperationResultV1` | `pending`, `active`, `accepted`, `completed`, `superseded`, `rejected` |

`artifacts.maxItems=2`; incident IDs and all UUIDs use `CanonicalUUID`; every digest uses `DigestHex32` unless the canonical DTO explicitly transports 32 raw bytes. Contract tests enumerate every operation and assert its exact request ref, success statuses, auth extension, header allowlist, wire/decoded cap, and that no additional path/schema/security scheme exists. The operator contract test also requires each list success response to reference its resource-specific list schema, requires `items` to have the exact resource `$ref`, `minItems: 0`, and `maxItems: 200`, recursively requires `additionalProperties: false` and a nonempty `required` array on every object schema, and rejects any anonymous/free-form item schema or generic list component.

- [ ] **Step 4: Add exact generation configs and generation script commands**

Use one config per generated package:

```yaml
package: nodebootstrapv1
output: gen/go/talenro/nodebootstrap/v1/server.gen.go
generate:
  models: true
  client: true
  std-http-server: true
  embedded-spec: true
output-options:
  skip-prune: false
compatibility:
  preserve-original-operation-id-casing-in-embedded-spec: true
```

Use `nodeagentv1` and `nodeoperatorv1` with their corresponding output paths in the other two configs. Add three explicit `go tool oapi-codegen --config <config> <spec>` invocations to each generation script; do not loop over filesystem discovery.

- [ ] **Step 5: Generate and run the GREEN contract test**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS and creation of the three exact `server.gen.go` paths, each containing both strict server adapters and typed response clients.

Run: `go test ./internal/nodecontrol/contracts -run TestNodeOpenAPISpecsAreIsolatedAndValid -count=1`

Expected: PASS.

- [ ] **Step 6: REFACTOR the test into an exact operation/security matrix**

Add a table that asserts every frozen operation ID exists once; bootstrap, agent and operator have the exact `x-talenro-listener-auth` value; no spec invents a client-controlled mTLS security scheme; ordinary request schemas stay under 64 KiB; and only `createNodeTrustConflictEvidence` admits the 1 MiB exception. Run:

`go test ./internal/nodecontrol/contracts -run 'TestNodeOpenAPI' -count=1`

Expected: PASS with no cross-listener path, pseudo-mTLS security scheme or auth-extension leakage.

- [ ] **Step 7: Verify deterministic OpenAPI regeneration**

Run: `git diff --exit-code -- api/openapi gen/go/talenro/nodebootstrap gen/go/talenro/nodeagent gen/go/talenro/nodeoperator`

Expected: exit 0 immediately after a second `scripts/generate.ps1` run.

- [ ] **Step 8: Commit the three contracts**

```bash
git add api/openapi/node-bootstrap-api.v1.yaml api/openapi/node-bootstrap-oapi-codegen.yaml api/openapi/node-agent-api.v1.yaml api/openapi/node-agent-oapi-codegen.yaml api/openapi/node-operator-api.v1.yaml api/openapi/node-operator-oapi-codegen.yaml gen/go/talenro/nodebootstrap/v1/server.gen.go gen/go/talenro/nodeagent/v1/server.gen.go gen/go/talenro/nodeoperator/v1/server.gen.go scripts/generate.ps1 scripts/generate.sh internal/nodecontrol/contracts/openapi_contract_test.go
git commit -m "feat(nodecontrol): add isolated node API contracts"
```

### Task 3: Add the closed nodecontrol Protobuf event contract

**Files:**
- Create: `api/proto/talenro/nodecontrol/v1/events.proto`
- Create: `gen/go/talenro/nodecontrol/v1/events.pb.go`
- Create: `internal/nodecontrol/contracts/events.go`
- Test: `internal/nodecontrol/contracts/events_test.go`
- Test: `internal/nodecontrol/contracts/testdata/node_control_event_v1.bin`

**Interfaces:**
- Consumes: deterministic `proto.MarshalOptions`, canonical UUID rules, UTC `google.protobuf.Timestamp`.
- Produces: `nodecontrolv1.NodeControlEventV1`; `func MarshalNodeControlEvent(*nodecontrolv1.NodeControlEventV1) ([]byte, error)`; `func UnmarshalNodeControlEvent([]byte) (*nodecontrolv1.NodeControlEventV1, error)`; `func NodeControlSubject(string) (string, error)`.

- [ ] **Step 1: Write the RED privacy and golden-wire tests**

```go
func TestNodeControlEventGoldenAndPrivate(t *testing.T) {
	event := &nodecontrolv1.NodeControlEventV1{
		EventId: "11111111-1111-4111-8111-111111111111",
		OccurredAt: timestamppb.New(time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)),
		AggregateType: "node",
		AggregateId: "22222222-2222-4222-8222-222222222222",
		AggregateVersion: 7,
		EventType: "node_inventory_changed.v1",
		Payload: &nodecontrolv1.NodeControlEventV1_NodeInventoryChanged{
			NodeInventoryChanged: &nodecontrolv1.NodeInventoryChangedV1{NodeId: "22222222-2222-4222-8222-222222222222", InventoryVersion: 7, PopCode: "us-east", OperatorState: "enabled"},
		},
	}
	encoded, err := MarshalNodeControlEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/node_control_event_v1.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, want) {
		t.Fatal("nodecontrol event wire drift")
	}
	for _, canary := range [][]byte{[]byte("serial-canary"), []byte("csr-canary"), []byte("endpoint-canary")} {
		if bytes.Contains(encoded, canary) {
			t.Fatal("private canary reached event bytes")
		}
	}
}
```

- [ ] **Step 2: Run the focused RED test**

Run: `go test ./internal/nodecontrol/contracts -run TestNodeControlEventGoldenAndPrivate -count=1`

Expected: FAIL because the proto package and event helpers do not exist.

- [ ] **Step 3: Define the exact envelope and six oneof payloads**

```proto
syntax = "proto3";

package talenro.nodecontrol.v1;

import "google/protobuf/timestamp.proto";

option go_package = "talenro.local/platform/gen/go/talenro/nodecontrol/v1;nodecontrolv1";

message NodeControlEventV1 {
  string event_id = 1;
  google.protobuf.Timestamp occurred_at = 2;
  string aggregate_type = 3;
  string aggregate_id = 4;
  uint64 aggregate_version = 5;
  string event_type = 6;
  oneof payload {
    NodeInventoryChangedV1 node_inventory_changed = 10;
    NodeDesiredStatePublishedV1 node_desired_state_published = 11;
    NodeAvailabilityChangedV1 node_availability_changed = 12;
    NodeSecurityStateChangedV1 node_security_state_changed = 13;
    NodeCertificateStatusChangedV1 node_certificate_status_changed = 14;
    NodeOperatorActionRecordedV1 node_operator_action_recorded = 15;
  }
}

message NodeInventoryChangedV1 { string node_id = 1; uint64 inventory_version = 2; string pop_code = 3; string operator_state = 4; }
message NodeDesiredStatePublishedV1 { string node_id = 1; uint64 generation = 2; bytes content_digest = 3; google.protobuf.Timestamp valid_until = 4; string reason = 5; }
message NodeAvailabilityChangedV1 { string node_id = 1; string boot_id = 2; uint64 sequence = 3; string health = 4; optional bool accepting_new = 5; string reason = 6; }
message NodeSecurityStateChangedV1 { string node_id = 1; string incident_id = 2; string fault_subtype = 3; string security_state = 4; string reason = 5; }
message NodeCertificateStatusChangedV1 { string node_id = 1; string certificate_record_id = 2; string status = 3; }
message NodeOperatorActionRecordedV1 { string audit_id = 1; string operator_id = 2; string action = 3; string target = 4; string result = 5; string reason = 6; }
```

- [ ] **Step 4: Implement strict deterministic marshal/unmarshal and subject mapping**

`events.go` must reject unknown fields, noncanonical protobuf bytes, event/payload mismatch, invalid timestamps, noncanonical IDs, zero versions, `content_digest` lengths other than exactly 32 bytes, absent `accepting_new` presence (including when its value is false), unsafe strings, and all unregistered event types. `NodeControlSubject` returns only `talenro.nodecontrol.v1.<event_type>` for the six registered values.

- [ ] **Step 5: Generate the proto and create the reviewed golden vector**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS and creation of `gen/go/talenro/nodecontrol/v1/events.pb.go`.

Generate the golden bytes once through a test-only `-update` flag, review with `Format-Hex`, remove the update path from normal execution, and retain the fixed binary vector.

- [ ] **Step 6: Run the GREEN tests**

Run: `go test ./internal/nodecontrol/contracts -run 'TestNodeControlEvent|TestNodeControlSubject' -count=1`

Expected: PASS for all six payloads, event mismatch rejection, unknown-field rejection, privacy canaries, and golden bytes.

- [ ] **Step 7: REFACTOR with deterministic round-trip and descriptor safety checks**

Add a test that enumerates all six `oneof` descriptors, verifies unique field numbers, rejects the forbidden field terms from `internal/contracts/events/validate.go`, and proves `Marshal(Unmarshal(bytes))` is byte-identical.

Run: `go test ./internal/nodecontrol/contracts -count=1`

Expected: PASS.

- [ ] **Step 8: Commit the event contract**

```bash
git add api/proto/talenro/nodecontrol/v1/events.proto gen/go/talenro/nodecontrol/v1/events.pb.go internal/nodecontrol/contracts/events.go internal/nodecontrol/contracts/events_test.go internal/nodecontrol/contracts/testdata/node_control_event_v1.bin
git commit -m "feat(nodecontrol): add bounded node control events"
```

### Task 4: Create the complete C1.2 authority schema in one reversible migration

**Files:**
- Create: `db/migrations/00006_nodecontrol.sql`
- Create: `db/schema/nodecontrol.v1.yaml`
- Test: `internal/store/nodecontrol_schema_test.go`
- Test: `internal/store/nodecontrol_schema_manifest_test.go`
- Test: `internal/store/nodecontrol_schema_integration_test.go`

**Interfaces:**
- Consumes: goose migration conventions, PostgreSQL 18 partial unique indexes/check constraints, existing UUID/timestamptz/sqlc overrides.
- Produces: schema `nodecontrol`, a machine-readable exact column/constraint/trigger manifest, and every table named in spec §9; no later plan may merge or reinterpret these authority semantics.

- [ ] **Step 1: Write the RED schema manifest test**

```go
func TestNodeControlMigrationContainsEveryAuthorityTable(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00006_nodecontrol.sql")
	if err != nil {
		t.Fatal(err)
	}
	required := []string{
		"node_pops", "node_failure_domains", "node_inventory", "node_failure_domain_membership",
		"node_endpoints", "node_process_slots", "node_capacity_profiles", "node_resource_envelopes",
		"node_enrollment_grants", "node_certificate_issuances", "node_certificates",
		"node_state_signing_intents", "node_root_metadata_publish_intents", "node_root_metadata_signature_shares",
		"node_desired_states", "node_recovery_states", "node_observed_states", "node_state_transitions",
		"node_security_incidents", "node_security_fault_receipts", "node_recovery_sessions",
		"node_restore_reauthorization_approvals", "node_operator_audit",
		"control_plane_trust_bundle_high_waters", "control_plane_authority_fences",
	}
	for _, table := range required {
		if !bytes.Contains(body, []byte("CREATE TABLE nodecontrol."+table)) {
			t.Fatalf("missing table %s", table)
		}
	}
}
```

Add a second test that loads `db/schema/nodecontrol.v1.yaml`, requires exactly the 25 table names below, rejects duplicate/unknown table, column, constraint, enum, index or trigger names, and requires every manifest object to have an exact migration/catalog counterpart. Decode into typed structs with `KnownFields(true)`; reject YAML anchors/aliases/merge keys and any scalar that uses a compact semantic token instead of an exact type/default/check/predicate. Require every column object to carry explicit nullability/default/collation fields even when their value is null, every constraint/index/trigger to have a stable name, and every `authority_operation_id` to have the exact composite FK. The YAML parser is used only by tests; the runtime never reads the manifest.

- [ ] **Step 2: Run the RED schema test**

Run: `go test ./internal/store -run TestNodeControlMigrationContainsEveryAuthorityTable -count=1`

Expected: FAIL because `00006_nodecontrol.sql` does not exist.

- [ ] **Step 3: Add the machine manifest, migration header, authority table, and immutable fence constraints**

Encode all 25 entries below in `db/schema/nodecontrol.v1.yaml`, then use this exact authority core in the migration:

```sql
-- +goose Up
CREATE SCHEMA nodecontrol;

CREATE TABLE nodecontrol.control_plane_authority_fences (
  operation_id uuid PRIMARY KEY,
  effect_kind text NOT NULL CHECK (effect_kind IN (
    'trust_bundle_publish','root_publish','metadata_publish','grant_create','grant_claim',
    'certificate_activate','certificate_revoke','identity_epoch_advance',
    'security_incident_open','security_incident_resolve','resource_envelope_activate',
    'desired_activate','recovery_activate','operator_transition','operator_authorizer_change'
  )),
  scope_kind text NOT NULL CHECK (scope_kind IN ('node','global_node_trust','global_operator_trust')),
  authority_epoch bigint NOT NULL CHECK (authority_epoch > 0),
  authority_sequence bigint NOT NULL CHECK (authority_sequence > 0),
  scope_digest bytea NOT NULL CHECK (octet_length(scope_digest) = 32),
  provider_reservation_digest bytea NOT NULL CHECK (octet_length(provider_reservation_digest) = 32),
  effect_digest bytea CHECK (effect_digest IS NULL OR octet_length(effect_digest) = 32),
  provider_status text NOT NULL CHECK (provider_status IN ('reserved','committed','aborted')),
  provider_receipt_digest bytea CHECK (provider_receipt_digest IS NULL OR octet_length(provider_receipt_digest) = 32),
  db_system_id numeric(20,0) CHECK (db_system_id IS NULL OR db_system_id BETWEEN 1 AND 18446744073709551615),
  db_timeline bigint CHECK (db_timeline IS NULL OR db_timeline BETWEEN 1 AND 4294967295),
  required_lsn pg_lsn,
  abort_reason text CHECK (abort_reason IS NULL OR abort_reason IN ('validation_failed','superseded','provider_dependency_failed','activation_deadline_expired')),
  visibility_state text NOT NULL CHECK (visibility_state IN ('fence_pending','active','aborted')),
  reserved_at timestamptz NOT NULL,
  effect_bound_at timestamptz,
  terminal_at timestamptz,
  UNIQUE (operation_id, authority_epoch, authority_sequence),
  UNIQUE (authority_epoch, authority_sequence),
  CHECK ((effect_digest IS NULL AND db_system_id IS NULL AND db_timeline IS NULL AND required_lsn IS NULL AND effect_bound_at IS NULL)
      OR (effect_digest IS NOT NULL AND db_system_id IS NOT NULL AND db_timeline IS NOT NULL AND required_lsn IS NOT NULL AND effect_bound_at IS NOT NULL)),
  CHECK (provider_status <> 'committed' OR effect_digest IS NOT NULL),
  CHECK ((provider_status = 'reserved') = (provider_receipt_digest IS NULL)),
  CHECK ((provider_status = 'reserved') = (terminal_at IS NULL)),
  CHECK ((visibility_state = 'active') = (provider_status = 'committed')),
  CHECK ((visibility_state = 'aborted') = (provider_status = 'aborted')),
  CHECK ((provider_status = 'aborted') = (abort_reason IS NOT NULL)),
  CHECK (terminal_at IS NULL OR terminal_at >= reserved_at),
  CHECK (effect_bound_at IS NULL OR effect_bound_at >= reserved_at)
);

CREATE INDEX authority_checkpoint_node
ON nodecontrol.control_plane_authority_fences(authority_epoch, scope_digest, authority_sequence DESC)
WHERE scope_kind = 'node' AND provider_status = 'committed' AND visibility_state = 'active';

CREATE INDEX authority_checkpoint_global_node_trust
ON nodecontrol.control_plane_authority_fences(authority_epoch, authority_sequence DESC)
WHERE scope_kind = 'global_node_trust' AND provider_status = 'committed' AND visibility_state = 'active';
```

The legal pre-finalize state is therefore `reserved/fence_pending` with either an entirely unbound database group or an entirely bound group; binding an effect cannot violate the table check. Add a `BEFORE UPDATE` trigger that rejects changes to operation/kind/scope/epoch/sequence/reserved time, any terminal-row update, a second different effect/database binding, or any status transition except `reserved→committed|aborted`. Its integration test attempts each forbidden direct SQL mutation and requires a constraint/trigger failure.

- [ ] **Step 4: Add all inventory, identity, state, recovery, observation, and audit tables**

Create `db/schema/nodecontrol.v1.yaml` first and make the migration/catalog test implement it exactly. The YAML is a literal catalog manifest, not prose and not a semantic schema language. Every table entry enumerates every column with `name`, canonical PostgreSQL `sql_type`, `nullable`, exact `default_sql` (or explicit null), `collation` (or explicit null), and every named `check_sql`; it separately enumerates the ordered primary-key columns, every named unique/FK/check constraint with canonical SQL and deferrability, every index with ordered key/include columns and an exact predicate, and every trigger with timing/events/function/WHEN expression. It is invalid to put shorthand or qualitative tokens such as `V`, `N`, `D32`, `D32?`, `B64K`, `bounded`, `closed`, `paired`, `optional`, `authority_group`, `retention`, or `same as` anywhere in a manifest field. Repeated authority columns are expanded on every table as `authority_operation_id uuid NOT NULL`, `authority_epoch bigint NOT NULL CHECK (authority_epoch BETWEEN 1 AND 9223372036854775807)`, and `authority_sequence bigint NOT NULL CHECK (authority_sequence BETWEEN 1 AND 9223372036854775807)`, with the exact named composite FK to `control_plane_authority_fences(operation_id,authority_epoch,authority_sequence)`. Every digest column spells out `bytea` plus its named `octet_length(...)=32` check, every 64-byte signature does the same for 64, every text column spells out collation/pattern/octet bound, and every time column spells out `timestamp with time zone` and its ordering checks. No YAML anchor, alias, merge key, templating reference, or inherited default is allowed.

The following 25-row matrix is the human review coverage list; it is not permitted as YAML content and none of its compact prose substitutes for the per-column catalog objects required above:

1. `node_pops`: `pop_code text` PK (lowercase 1–32 bytes, `^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`), `iso_country char(2)` uppercase, `region text` lowercase safe 1–64, `operator_state`=`provisioning|enabled|draining|disabled`, `version V DEFAULT 1`, `created_at`, `updated_at`; unique `(iso_country,region,pop_code)`.
2. `node_failure_domains`: `failure_domain_id uuid` PK, `domain_type`=`facility|compute|upstream`, `stable_id` printable non-whitespace ASCII 1–128, `version V DEFAULT 1`, timestamps; unique `(domain_type,stable_id)` and `(failure_domain_id,domain_type)`.
3. `node_inventory`: `node_id uuid` PK, `pop_code` FK, separate `operator_state`, `security_state=normal|quarantined`, `identity_state=never_enrolled|active|recovery_pending|recovery_limited|revoked`, nullable `resume_operator_state`, nullable `pending_operator_transition=draining|resume|restore_reauthorize` plus paired `pending_transition_signing_id`, `identity_epoch N DEFAULT 0`, nullable `lineage_id`, `inventory_version V DEFAULT 1`, `security_version V DEFAULT 1`, paired resource-envelope version/digest, desired/recovery active and next generations, paired active root/metadata publish IDs and versions, nullable all-or-none last-authority group, timestamps. Never-enrolled means epoch zero/no lineage; every other identity state requires lineage. Active generation is null or lower than next. Quarantine implies operator disabled and a saved resume state. Deferred pointer triggers require the exact active signed/published row and committed fence.
4. `node_failure_domain_membership`: `node_id` FK, `failure_domain_id`, denormalized `domain_type`, `inventory_version V`, `created_at`; PK `(node_id,failure_domain_id)`, composite domain FK, unique `(node_id,domain_type)`.
5. `node_endpoints`: `endpoint_id uuid` PK, `node_id` FK, `address inet`, port 1–65535, `transport=tcp|udp`, closed protocol capability, operator state, `inventory_version V`, timestamps; unique exact endpoint tuple; no credential/material column.
6. `node_process_slots`: `node_id` FK, bounded `slot_id`, `adapter=fixture|xray|sing_box`, capacity profile ID/version FK, `required boolean`, operator state, `inventory_version V`, timestamps; PK `(node_id,slot_id)`. A node-row-locked trigger rejects a ninth slot, including concurrent slot-8/9 insert races.
7. `node_capacity_profiles`: `(profile_id,version V)` PK, adapter, all ten non-null limits: egress 1,000,000–1,000,000,000,000; connections 1–10,000,000; handshakes/s 1–1,000,000; CPU millicores 100–64,000; CPU basis points 1–10,000; memory 67,108,864–1,099,511,627,776; tasks 32–4,096; FDs 64–1,000,000; queue 1–1,000,000; packet-loss basis points 1–10,000; nonempty sorted/unique `required_metrics` max 9 from the closed metric registry; `created_at`. A trigger rejects update/delete after any reference and application validation requires release capability support.
8. `node_resource_envelopes`: `(node_id,envelope_version V)` PK plus authority group, per-node unique envelope digest, canonical package `B64K`, deployment key ID `D32`, 64-byte signature, `max_slots=8`, bounded agent/supervisor/core-parent CPU/memory/task/FD limits, aggregate slot-FD/tmpfs byte/inode limits, detected-host-capacity digest, issued/created times. Immutable; inventory pointer requires exact digest and committed fence.
9. `node_enrollment_grants`: grant UUID PK, create-authority group, node, identity epoch, unique token digest, CSR digest, idempotency digest, created/expiry (`<=10m`), all-or-none consumption group including attempt/request/claim authority/result issuance, mutually exclusive expiry/invalidation groups and closed reason. One live grant per node/epoch; plaintext grant and CSR DER columns are forbidden.
10. `node_certificate_issuances`: issuance UUID PK, unique node/attempt and operation, node, `issuance_kind=initial|rotation|recovery`, identity epoch/lineage/issuer, CSR/public-key/template/request digests, `status=pending|active|rejected|superseded|failed`, all-or-none validated result group (serial 1–20 bytes, bounded leaf/chain DER and digests, validity), closed failure, timestamps and 30-day retention. Only pending→one terminal; inputs and populated result are immutable.
11. `node_certificates`: certificate UUID PK, unique issuance FK, activation authority group, exact node/epoch/lineage/issuer/serial/leaf DER and leaf/key/chain digests, validity, `status=active|recovery_pending|recovery_limited|revoked`, all-or-none revoke authority/time/reason group, timestamps/retention. Unique issuer+serial and leaf digest; active lookup index includes every exact authorization component. Only recovery-pending→limited/active/revoked, limited→active/revoked, active→revoked.
12. `node_state_signing_intents`: signing UUID PK plus authority group, node, `kind=desired|recovery`, idempotency/base/reserved generation, canonical payload/digest, root/metadata/key identity, captured inventory/identity/security values, the complete Plan 02 `RecoveryIntentBinding` nullable group, optional 64-byte signature/verified time, deadline, `status=pending|active|failed|superseded`, closed failure and timestamps. Unique node/kind/generation and one pending per node/kind. Recovery group is absent for desired and complete for recovery; remediation digest is present iff action is clear-latches. Only pending→terminal; activation rechecks every captured field.
13. `node_root_metadata_publish_intents`: publish UUID PK plus authority group, `kind=root|metadata`, `reason=normal|root_rotation|emergency_revoke`, optional exact incident, base/reserved versions, canonical payload/digest, key-set/threshold capture, deadline, optional envelope/digest, `status=pending|active|failed|superseded`, failure/timestamps. Global pending count max 8. Old active may become superseded only in the next fence-finalized same-kind activation transaction; payload/share remain immutable. Root rotation checks current and new thresholds; emergency requires its exact open incident.
14. `node_root_metadata_signature_shares`: publish FK, key ID, physical key ID, payload digest, role=`current_root|new_root|metadata`, exact 64-byte signature and verified time; PK by publish/key/role and unique physical-key/role. Immutable; trigger binds payload/role/key to intent and forbids online state-signer identities.
15. `node_desired_states`: `(node_id,generation V)` PK, unique signing FK, authority group, inventory/resource-envelope/root/metadata/key bindings, canonical payload/digest, 64-byte signature, issued/valid/effective-deadline times, closed reason, created/retention. Require issued < effective <= valid <= issued+24h. Immutable and pointer-gated to the exact active intent/fence.
16. `node_recovery_states`: `(node_id,recovery_generation V)` PK, unique signing FK, authority group, identity epoch, recovery session/reason/version, incident/local/supervisor set digests and counts capped 16/64/16, optional remediation digest, root/metadata/action, `all_slots_stopped=true`, canonical payload/digest/key/signature, issued/valid (`<=15m`), created/retention. Remediation iff clear-latches; empty incidents only for administrative-disable/retire/authority-restore. Immutable and pointer-gated.
17. `node_observed_states`: one row per node with boot/sequence/request/observation digests, bounded canonical observation, monotonic sample end, arrival time, seen/applied generation+authority+digest groups, inventory/reducer version/digest/state, closed health/reason, capacity/agent/final accepting flags and timestamps. Generation zero iff its digest is null. Repository enforces new boot starts at 1, exact retry or higher same-boot sequence; trigger forbids final true when agent false. No raw-history table.
18. `node_state_transitions`: transition UUID PK, node, optional authority group, dimension=`operator|security|identity|health|certificate`, dimension-validated from/to values, closed reason, optional paired observation/incident/audit references, aggregate version, occurrence/retention. Immutable, from != to, no raw sample/provider/core text.
19. `node_security_incidents`: incident UUID PK, open-authority group, node/identity epoch, closed fault subtype, fixed subtype slot 1–12 with 13–15 reserved/unusable and overflow slot 16, `status=open|resolution_pending_agent_ack|overflow|resolved`, fixed first/last evidence digests, saturating occurrence count, trust-context digest, first/last times, optional all-or-none resolution authority/remediation/time and retention. One nonresolved row per node/subtype and slot. Overflow starts status overflow and may advance to pending-ack/resolved. Direct resolution is allowed only with no active local/supervisor binding; open rows have null retention.
20. `node_security_fault_receipts`: receipt UUID PK plus authority group, node/epoch/local fault/request/subtype/evidence/boot/incident, optional all-or-none supervisor binding, node-locked binding slots 1–64 and supervisor slots 1–16, `result='accepted'`, `delivery=pending|deliverable`, `binding=active|cleared`, deliver/clear metadata. Unique local retry key and supervisor fault key; changed request/subtype/evidence cannot reuse. Deliverable requires committed active fence.
21. `node_recovery_sessions`: recovery UUID PK plus authority group, node/identity epoch, `reason=identity_compromise|administrative_disable|retire|authority_restore|security_incident`, version, `status=pending|completed|superseded`, incident-set digest, resume operator state (forced disabled for retire/restore), optional recovery certificate and attestation digest, terminal/timestamps/retention. One pending session per node; only pending→one terminal; restore never revives saved intent.
22. `node_restore_reauthorization_approvals`: approval UUID PK plus authority group, node/recovery/effect/scope, `role=proposal|approval`, exact operator/credential/leaf/authority/authorizer/security-admin/POP-scope bindings, created/expiry/status/terminal/retention. Expiry is no later than min(created+15m, credential expiry, evidence completion+15m). Unique node/recovery/effect/operator and role enforce two distinct operators. Only pending→consumed/superseded; activation consumes both exact rows atomically.
23. `node_operator_audit`: audit UUID PK, unique command/operation ID, bounded operator ID, credential digest, `role=inventory_reader|inventory_writer|node_security_admin`, closed action/target/reason/result, optional paired authority and before/after versions, occurrence and >=180-day retention. Immutable; no credential bytes/provider errors/free text.
24. `control_plane_trust_bundle_high_waters`: PK `(purpose,listener_kind,trust_domain)` with closed purpose/listener mapping, lower-case DNS domain, authority group, bundle version/digest, cumulative-set digest/count and updated time. Trigger accepts exact retry or strict advance with append-only cumulative set; rollback/fork fails; activation requires committed fence.
25. `control_plane_authority_fences`: exactly the DDL in Step 3, including closed effect/scope matrix, reservation/receipt digests, legal unbound-or-fully-bound intermediate state, immutable reserve fields/terminal rows, and node/global checkpoint indexes.

Required retention is enforced in both constraints and delete predicates: terminal grants/issuances and expired/revoked certificates at least 30 days; desired/operator audit/resolved incidents at least 180 days; open incidents have no retention deadline. No job deletes a row still referenced by an active pointer, session, receipt, recovery state, audit or outbox. Manifest tests also reject an extra table/column/enum value, not only missing ones.

Required partial unique constraints include:

```sql
CREATE UNIQUE INDEX node_one_unconsumed_grant_per_epoch
ON nodecontrol.node_enrollment_grants(node_id, identity_epoch)
WHERE consumed_at IS NULL AND expired_at IS NULL AND invalidated_at IS NULL;

CREATE UNIQUE INDEX node_one_nonterminal_signing_intent_per_kind
ON nodecontrol.node_state_signing_intents(node_id, signing_kind)
WHERE status = 'pending';

CREATE UNIQUE INDEX node_root_publish_one_per_base_pointer
ON nodecontrol.node_root_metadata_publish_intents(base_root_version, base_metadata_version)
WHERE status = 'pending';

CREATE UNIQUE INDEX node_certificate_issuer_serial_unique
ON nodecontrol.node_certificates(issuer_id, serial_bytes);

CREATE UNIQUE INDEX node_certificate_leaf_digest_unique
ON nodecontrol.node_certificates(leaf_der_sha256);

CREATE UNIQUE INDEX node_open_incident_per_subtype
ON nodecontrol.node_security_incidents(node_id, fault_subtype)
WHERE status IN ('open','resolution_pending_agent_ack','overflow');

CREATE UNIQUE INDEX node_restore_approval_operator_unique
ON nodecontrol.node_restore_reauthorization_approvals(node_id, recovery_id, effect_digest, operator_id);

CREATE UNIQUE INDEX node_restore_approval_role_unique
ON nodecontrol.node_restore_reauthorization_approvals(node_id, recovery_id, effect_digest, role);
```

Add catalog-tested triggers for terminal immutability, active-pointer/fence relationships, slot/root-pending/incident/binding caps, capacity-profile immutability, trust high-water monotonicity and the closed authority effect/scope matrix. Down migration drops triggers/functions after tables and before the schema; it never uses `CASCADE`.

- [ ] **Step 5: Add an exact reverse-order goose Down section**

Drop child tables before parents and finish with `DROP SCHEMA nodecontrol;`. Do not use `CASCADE`; the down migration must expose forgotten dependencies.

- [ ] **Step 6: Run the GREEN manifest test and migration roundtrip**

Run: `go test ./internal/store -run TestNodeControlMigrationContainsEveryAuthorityTable -count=1`

Expected: PASS.

Run with a fresh owned integration database: `go test -tags=integration ./internal/store -run TestNodeControlMigrationUpDownUp -count=1 -timeout 3m`

Expected: PASS; Up, Down, and second Up all complete, and the second schema digest equals the first.

- [ ] **Step 7: REFACTOR the integration test to inspect constraints, not migration text**

Query `pg_catalog.pg_constraint`, `pg_indexes`, `pg_trigger`, `pg_proc`, and `information_schema.columns`; normalize each result and compare it field-for-field with `db/schema/nodecontrol.v1.yaml`, including enum members, defaults/nullability, fixed digest lengths, numeric bounds, partial predicates, FKs, trigger names, pointer relationships and retention columns. Fail on extra as well as missing objects. Run:

`go test -tags=integration ./internal/store -run 'TestNodeControlMigration' -count=1 -timeout 3m`

Expected: PASS against PostgreSQL 18.4.

- [ ] **Step 8: Commit the complete migration**

```bash
git add db/migrations/00006_nodecontrol.sql db/schema/nodecontrol.v1.yaml internal/store/nodecontrol_schema_test.go internal/store/nodecontrol_schema_manifest_test.go internal/store/nodecontrol_schema_integration_test.go
git commit -m "feat(nodecontrol): add authority database schema"
```

### Task 5: Implement the provider-neutral fence and deterministic local/test provider

**Files:**
- Create: `internal/nodecontrol/authority/provider.go`
- Create: `internal/nodecontrol/authority/deterministic_fake.go`
- Test: `internal/nodecontrol/authority/provider_test.go`
- Test: `internal/nodecontrol/authority/deterministic_fake_test.go`

**Interfaces:**
- Consumes: `contracts.AuthorityVersion`, `contracts.Digest`, canonical UUID operation IDs, `context.Context`.
- Produces: the suite-index `authority.Provider` interface and exact request/result types below; `func NewDeterministicProvider(epoch uint64) (*DeterministicProvider, error)`.

```go
type Provider interface {
	Reserve(context.Context, ReserveRequest) (Reservation, error)
	Finalize(context.Context, FinalizeRequest) (Receipt, error)
	Abort(context.Context, AbortRequest) (Receipt, error)
	Inspect(context.Context, uuid.UUID) (Record, error)
	Head(context.Context) (Head, error)
	CommittedNodeCheckpoint(context.Context, contracts.Digest) (NodeCheckpoint, error)
}

type ReserveRequest struct {
	OperationID uuid.UUID
	Kind        EffectKind
	ScopeKind   ScopeKind
	ScopeDigest contracts.Digest
}

type FinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
	DBSystemID   uint64
	DBTimeline   uint32
	RequiredLSN  WALPosition
}

type AbortRequest struct {
	OperationID uuid.UUID
	Reason      AbortReason
}

type ScopeKind string

const (
	ScopeNode                ScopeKind = "node"
	ScopeGlobalNodeTrust     ScopeKind = "global_node_trust"
	ScopeGlobalOperatorTrust ScopeKind = "global_operator_trust"
)

type WALPosition string

type Reservation struct {
	OperationID uuid.UUID
	Kind        EffectKind
	ScopeKind   ScopeKind
	ScopeDigest contracts.Digest
	Epoch       uint64
	Sequence    uint64
	ReservationDigest contracts.Digest
}

type DatabasePoint struct {
	SystemID    uint64
	Timeline    uint32
	RequiredLSN WALPosition
}

type ReceiptStatus string

const (
	StatusCommitted ReceiptStatus = "committed"
	StatusAborted   ReceiptStatus = "aborted"
)

type Receipt struct {
	Reservation
	EffectDigest *contracts.Digest
	DatabasePoint *DatabasePoint
	Status        ReceiptStatus
	AbortReason   *AbortReason
	ReceiptDigest contracts.Digest
}

type Record struct {
	Reservation
	BoundEffectDigest *contracts.Digest
	BoundDatabasePoint *DatabasePoint
	TerminalReceipt   *Receipt
}

type Head struct {
	Epoch                        uint64
	LatestReservedSequence       uint64
	LatestCommittedSequence      uint64
	LatestReservationDigest      contracts.Digest
	LatestCommittedOperationID   uuid.UUID
	LatestCommittedReceiptDigest contracts.Digest
	LatestCommittedDatabasePoint *DatabasePoint
}

type NodeCheckpoint struct {
	AuthorityEpoch uint64
	Sequence       uint64
	ReceiptDigest  contracts.Digest
}
```

`EffectKind` is a closed enum containing `trust_bundle_publish`, `root_publish`, `metadata_publish`, `grant_create`, `grant_claim`, `certificate_activate`, `certificate_revoke`, `identity_epoch_advance`, `security_incident_open`, `security_incident_resolve`, `resource_envelope_activate`, `desired_activate`, `recovery_activate`, `operator_transition`, and `operator_authorizer_change`. `AbortReason` contains only `validation_failed`, `superseded`, `provider_dependency_failed`, and `activation_deadline_expired`.

Node scope digests are exactly `SHA-256(ASCII("TALENRO-NODE-AUTHORITY-SCOPE-V1") || 0x00 || canonical 16-byte UUID)`. The two global transcripts and results are constants, not configuration:

| Scope | Exact SHA-256 input bytes | Golden lowercase hex digest |
| --- | --- | --- |
| `ScopeGlobalNodeTrust` | `ASCII("TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1") || 0x00 || ASCII("global_node_trust")` | `f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1` |
| `ScopeGlobalOperatorTrust` | `ASCII("TALENRO-GLOBAL-OPERATOR-TRUST-SCOPE-V1") || 0x00 || ASCII("global_operator_trust")` | `128f2a7c781f323050ad3ef1541bfad20b7eb32fb2dfae1ca03d59cf671547c8` |

There is exactly one NUL separator and no length prefix, trailing NUL, Unicode normalization, JSON quoting, newline, or caller-supplied suffix. `provider_test.go` recomputes both golden vectors byte-for-byte and rejects a reserve whose scope kind is global but whose digest differs by one bit. Node effects require `ScopeNode`; root/metadata and node-purpose bundle publication require `ScopeGlobalNodeTrust`; operator-authorizer/operator-purpose trust changes require `ScopeGlobalOperatorTrust`. No caller-selected digest can change that kind/scope matrix, and no fourth escape-hatch scope exists.

- [ ] **Step 1: Write RED tests for idempotency, total order, field conflicts, and terminal exclusion**

```go
func TestDeterministicProviderIsIdempotentAndTerminal(t *testing.T) {
	provider, err := NewDeterministicProvider(7)
	if err != nil {
		t.Fatal(err)
	}
	request := ReserveRequest{OperationID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Kind: EffectDesiredActivate, ScopeKind: ScopeNode, ScopeDigest: validNodeScopeDigest(t, uuid.MustParse("22222222-2222-4222-8222-222222222222"))}
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
	receipt, err := provider.Finalize(t.Context(), FinalizeRequest{OperationID: first.OperationID, EffectDigest: contracts.Digest{3}, DBSystemID: 72057594037927937, DBTimeline: 1, RequiredLSN: "0/16B6C50"})
	if err != nil || receipt.Status != StatusCommitted {
		t.Fatalf("finalize = %v, %v", receipt, err)
	}
	if _, err := provider.Abort(t.Context(), AbortRequest{OperationID: first.OperationID, Reason: AbortSuperseded}); !errors.Is(err, ErrTerminalConflict) {
		t.Fatalf("abort after commit error = %v", err)
	}
}
```

In the same RED table, assert the configured-epoch empty `Head` zero rules, the two global-scope golden vectors, the one-bit global digest rejection, and that a committed head returns the operation ID, receipt digest, and database point from one exact receipt while later reserved/aborted rows cannot splice those anchors.

- [ ] **Step 2: Run the RED provider test**

Run: `go test ./internal/nodecontrol/authority -run TestDeterministicProviderIsIdempotentAndTerminal -count=1`

Expected: FAIL because the package does not exist.

- [ ] **Step 3: Implement closed values and provider result validation**

Validate the exact structs above. All epoch/sequence values are positive and at most `math.MaxInt64`, except a `Head`/`NodeCheckpoint` may have sequence zero before the first matching reservation. `WALPosition` accepts only uppercase PostgreSQL `X/Y` hexadecimal syntax and canonicalizes away leading zeroes; `DBSystemID` is a decimal `uint64`, never an opaque label. Committed receipts require effect and database point and forbid abort reason; aborted receipts require an abort reason and may retain an already bound effect/database point. Pointer fields are defensively copied.

`ReservationDigest` and `ReceiptDigest` are locally recomputed, never provider-selected. The receipt transcript is exactly `TALENRO-CONTROL-PLANE-AUTHORITY-RECEIPT-V1\x00 || JCS(AuthorityReceiptPayloadV1)` where every integer is a canonical decimal string (avoiding JCS numbers above 2^53), every digest is lowercase hex, and the LSN is canonical text. The payload contains every reservation field, nullable bound-effect/database group, terminal status and nullable abort reason. A valid `Head` has epoch in 1..`math.MaxInt64`, committed sequence no greater than reserved sequence, and the following exact zero rules: `LatestReservationDigest` is zero iff reserved sequence is zero; `LatestCommittedOperationID == uuid.Nil`, `LatestCommittedReceiptDigest` is zero, and `LatestCommittedDatabasePoint == nil` iff committed sequence is zero; otherwise all three committed anchors are present and match the receipt at `LatestCommittedSequence`. The database point is defensively copied. This is not a journal hash head because a lower sequence may finalize after a higher reservation.

- [ ] **Step 4: Implement the mutex-protected deterministic provider**

The fake starts at sequence zero, allocates exactly one increasing sequence per new operation ID, returns byte-identical results for exact retries, rejects changed fields, and keeps committed/aborted terminal states mutually exclusive. `CommittedNodeCheckpoint(nodeScopeDigest)` returns the maximum committed sequence in the current epoch among that exact node scope and `global_node_trust`; another node's effects, pending rows and aborted rows never advance it. Context cancellation returns `ErrCanceled` without allocating or mutating state.

- [ ] **Step 5: Run the GREEN provider package**

Run: `go test ./internal/nodecontrol/authority -run 'TestDeterministicProvider|TestProviderValue' -count=1`

Expected: PASS, including 1,000 concurrent unique reservations with sequences 1 through 1,000 and no duplicates, plus checkpoint tests proving a second node cannot advance the first node while global root/metadata publication advances both.

- [ ] **Step 6: REFACTOR with deterministic fault injection points**

Add test-only methods `FailNext(Operation, Failure)` and `Snapshot() []Record`; `Operation` is `reserve/finalize/abort/inspect/head`, and `Failure` is `before_mutation/after_mutation_response_lost`. After-mutation response loss must be recoverable with the same operation ID and never allocate a new sequence.

Run: `go test -race ./internal/nodecontrol/authority -count=1`

Expected: PASS under the race detector.

- [ ] **Step 7: Commit the provider boundary and fake**

```bash
git add internal/nodecontrol/authority/provider.go internal/nodecontrol/authority/deterministic_fake.go internal/nodecontrol/authority/provider_test.go internal/nodecontrol/authority/deterministic_fake_test.go
git commit -m "feat(nodecontrol): add deterministic authority fence"
```

### Task 6: Add sqlc authority queries and transaction-bound PostgreSQL repository

**Files:**
- Create: `db/queries/nodecontrol_authority.sql`
- Create: `internal/store/nodecontrol_authority.sql.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/querier.go`
- Create: `internal/nodecontrol/authority/repository.go`
- Create: `internal/nodecontrol/authority/postgres_repository.go`
- Test: `internal/nodecontrol/authority/postgres_repository_test.go`
- Test: `internal/nodecontrol/authority/postgres_repository_integration_test.go`

**Interfaces:**
- Consumes: `store.DBTX`, provider `Reservation`/`Receipt`, caller-owned pgx transactions.
- Produces: `authority.Repository`; `func NewPostgresRepository(store.DBTX) (*PostgresRepository, error)`; transaction-safe pending/terminal recording used by batches 02 and 03.

```go
type Repository interface {
	RecordPending(context.Context, store.DBTX, Reservation, time.Time) error
	CaptureDatabasePoint(context.Context) (DatabasePoint, error)
	BindEffect(context.Context, store.DBTX, uuid.UUID, contracts.Digest, DatabasePoint, time.Time) error
	ActivateCommitted(context.Context, store.DBTX, Receipt, time.Time) error
	RecordAborted(context.Context, store.DBTX, Receipt, time.Time) error
	Get(context.Context, uuid.UUID) (Record, error)
	Head(context.Context) (DatabaseHead, error)
	ListPending(context.Context, uint64) ([]PendingFence, error)
	CommittedNodeCheckpoint(context.Context, uint64, contracts.Digest) (NodeCheckpoint, error)
}

type DatabaseHead struct {
	Epoch                        uint64
	RecordCount                  uint64
	LatestReservedSequence       uint64
	LatestCommittedSequence      uint64
	PendingCount                 uint64
	HasSequenceGap               bool
	LatestReservationDigest      contracts.Digest
	LatestCommittedOperationID   uuid.UUID
	LatestCommittedReceiptDigest contracts.Digest
	LatestCommittedDatabasePoint *DatabasePoint
}

type PendingFence struct {
	OperationID       uuid.UUID
	Kind              EffectKind
	ScopeKind         ScopeKind
	ScopeDigest       contracts.Digest
	Epoch             uint64
	Sequence          uint64
	ReservationDigest contracts.Digest
	BoundEffectDigest *contracts.Digest
	BoundDatabasePoint *DatabasePoint
	ReservedAt        time.Time
	EffectBoundAt     *time.Time
}
```

`Repository.Head` returns only rows from the greatest epoch present in the table. For a nonempty result, epoch/record count/reserved sequence are positive, `HasSequenceGap` is exactly `RecordCount != LatestReservedSequence`, and the reservation digest is nonzero. The three latest-committed anchors obey the same all-zero/all-present rule as provider `Head`; `LatestCommittedOperationID` and its exact database point come from the same active row as `LatestCommittedReceiptDigest`, never from independent aggregates. `ListPending(ctx, epoch)` returns every `reserved/fence_pending` row in ascending sequence with its reservation digest and either a wholly absent or wholly present effect/database binding; it rejects epoch zero and never returns committed, aborted, or another-epoch rows. Returned digest, database-point, and time pointers are defensive copies.

- [ ] **Step 1: Write the RED repository test around caller-owned transaction identity**

Test that `RecordPending` uses the supplied transaction, stores exact epoch/sequence/scope, exact retry succeeds, changed retry is `ErrConflict`, and no method begins or commits a transaction itself. Add RED cases for the exact empty `DatabaseHead`, a latest committed operation/database-point anchor from one row, an all-fields `PendingFence`, defensive copies, and `ListPending` excluding committed/aborted/other-epoch rows while preserving ascending sequence.

- [ ] **Step 2: Run the RED repository test**

Run: `go test ./internal/nodecontrol/authority -run TestPostgresRepositoryRecordsExactFenceState -count=1`

Expected: FAIL because the repository and generated queries do not exist.

- [ ] **Step 3: Add exact authority queries**

```sql
-- name: InsertAuthorityFencePending :execrows
INSERT INTO nodecontrol.control_plane_authority_fences
  (operation_id, effect_kind, scope_kind, authority_epoch, authority_sequence, scope_digest,
   provider_reservation_digest, provider_status, visibility_state, reserved_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'reserved', 'fence_pending', $8)
ON CONFLICT (operation_id) DO NOTHING;

-- name: GetAuthorityFenceForUpdate :one
SELECT * FROM nodecontrol.control_plane_authority_fences
WHERE operation_id = $1
FOR UPDATE;

-- name: BindAuthorityFenceEffect :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET effect_digest = $2, db_system_id = $3, db_timeline = $4,
    required_lsn = $5, effect_bound_at = $6
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest IS NULL;

-- name: ActivateCommittedAuthorityFence :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status = 'committed', provider_receipt_digest = $2, visibility_state = 'active', terminal_at = $3
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
  AND effect_digest = $4 AND db_system_id = $5 AND db_timeline = $6 AND required_lsn = $7;

-- name: AbortAuthorityFence :execrows
UPDATE nodecontrol.control_plane_authority_fences
SET provider_status = 'aborted', provider_receipt_digest = $2, abort_reason = $3,
    visibility_state = 'aborted', terminal_at = $4
WHERE operation_id = $1 AND provider_status = 'reserved' AND visibility_state = 'fence_pending';

-- name: GetAuthorityFence :one
SELECT * FROM nodecontrol.control_plane_authority_fences WHERE operation_id = $1;

-- name: GetAuthorityFenceHead :one
WITH latest_epoch AS (
  SELECT COALESCE(MAX(authority_epoch), 0)::bigint AS authority_epoch
  FROM nodecontrol.control_plane_authority_fences
), epoch_rows AS (
  SELECT f.* FROM nodecontrol.control_plane_authority_fences f
  JOIN latest_epoch e ON f.authority_epoch = e.authority_epoch
), aggregate_head AS (
  SELECT COUNT(*)::bigint AS record_count,
         COALESCE(MAX(authority_sequence), 0)::bigint AS latest_reserved_sequence,
         COALESCE(MAX(authority_sequence) FILTER (
           WHERE provider_status = 'committed' AND visibility_state = 'active'), 0)::bigint AS latest_committed_sequence,
         COUNT(*) FILTER (WHERE provider_status = 'reserved')::bigint AS pending_count
  FROM epoch_rows
), latest_record AS (
  SELECT provider_reservation_digest FROM epoch_rows ORDER BY authority_sequence DESC LIMIT 1
), latest_committed AS (
  SELECT operation_id, provider_receipt_digest, db_system_id, db_timeline, required_lsn
  FROM epoch_rows
  WHERE provider_status = 'committed' AND visibility_state = 'active'
  ORDER BY authority_sequence DESC LIMIT 1
)
SELECT e.authority_epoch, a.record_count, a.latest_reserved_sequence,
       a.latest_committed_sequence, a.pending_count, r.provider_reservation_digest,
       c.operation_id AS latest_committed_operation_id,
       c.provider_receipt_digest, c.db_system_id, c.db_timeline, c.required_lsn,
       (a.record_count <> a.latest_reserved_sequence) AS has_sequence_gap
FROM latest_epoch e CROSS JOIN aggregate_head a
LEFT JOIN latest_record r ON true
LEFT JOIN latest_committed c ON true;

-- name: ListPendingAuthorityFences :many
SELECT operation_id, effect_kind, scope_kind, scope_digest, authority_epoch,
       authority_sequence, provider_reservation_digest, effect_digest,
       db_system_id, db_timeline, required_lsn, reserved_at, effect_bound_at
FROM nodecontrol.control_plane_authority_fences
WHERE authority_epoch = sqlc.arg(authority_epoch)
  AND provider_status = 'reserved' AND visibility_state = 'fence_pending'
ORDER BY authority_sequence;

-- name: GetCommittedNodeCheckpoint :one
SELECT authority_epoch, authority_sequence, provider_receipt_digest
FROM nodecontrol.control_plane_authority_fences
WHERE authority_epoch = sqlc.arg(authority_epoch)
  AND provider_status = 'committed' AND visibility_state = 'active'
  AND (scope_kind = 'global_node_trust'
       OR (scope_kind = 'node' AND scope_digest = sqlc.arg(node_scope_digest)))
ORDER BY authority_sequence DESC
LIMIT 1;

-- name: GetNodeControlDatabaseIdentity :one
SELECT system_identifier::numeric(20,0) AS system_id,
       timeline_id::bigint AS timeline,
       pg_current_wal_insert_lsn()::text AS required_lsn
FROM pg_control_system(), pg_control_checkpoint();
```

The authority-effect workflow has three database boundaries: the domain transaction commits its immutable intent/effect and references the reserved operation; only after that commit does `CaptureDatabasePoint` read the same primary's system ID/timeline and a safe `pg_current_wal_insert_lsn()` not earlier than the effect commit; a second short transaction binds that point to the fence row before provider finalize. Recovery between those boundaries reloads the immutable domain intent by operation ID and binds the same effect digest. It never guesses an LSN inside the uncommitted domain transaction. A missing checkpoint row maps to `(current provider epoch, sequence 0, zero digest)` only when provider and database both report no committed node/global-node-trust effect.

Empty-head epoch normalization is one explicit coordinator rule: SQL returns `DatabaseHead{Epoch:0}` only when `RecordCount`, both sequences, `PendingCount`, both digests, `LatestCommittedOperationID`, and `HasSequenceGap` are all zero/false and `LatestCommittedDatabasePoint` is nil. If and only if provider `Head` is also empty by its zero rules, the coordinator copies the provider's positive epoch into that exact empty database head for comparison. It never rewrites an epoch on a nonempty or partially populated head, never maps provider epoch zero, and never treats an empty database as equal to a nonempty provider. `CommittedNodeCheckpoint` uses the provider's positive current epoch explicitly and returns that epoch with sequence zero/zero receipt digest only when both sides prove no matching committed node/global-node-trust row.

- [ ] **Step 4: Generate sqlc and implement strict row conversion**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS and generated authority methods appear in `internal/store/nodecontrol_authority.sql.go` and `Querier`.

Implement all row conversions with exact 32-byte copies, positive integer checks, closed status checks, canonical UUID checks, and no raw provider/database error wrapping.

- [ ] **Step 5: Run GREEN unit and integration tests**

Run: `go test ./internal/nodecontrol/authority -run TestPostgresRepository -count=1`

Expected: PASS.

Run: `go test -tags=integration ./internal/nodecontrol/authority -run TestPostgresRepositoryRecordsExactFenceState -count=1 -timeout 3m`

Expected: PASS against a fresh migrated PostgreSQL 18.4 database.

- [ ] **Step 6: REFACTOR to prove pending rows are invisible to active-head checks**

Add concurrent integration cases for reserved-only, effect-bound, committed-active, aborted, same-sequence fork, epoch rollover with sequence reset, node A/node B/global checkpoint effects, and timeline mismatch. Assert `DatabaseHead` aggregates only the latest epoch, reports pending count/sequence gaps and exact reservation/receipt/database anchors, lists every pending operation ID, and advances committed sequence/checkpoint only for `provider_status='committed' AND visibility_state='active'`.

Run: `go test -tags=integration ./internal/nodecontrol/authority -run 'TestPostgresRepository|TestAuthorityHead' -count=1 -timeout 3m`

Expected: PASS.

- [ ] **Step 7: Verify generation drift and commit**

Run: `git diff --exit-code -- internal/store`

Expected: exit 0 after a second generation run.

```bash
git add db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/authority/repository.go internal/nodecontrol/authority/postgres_repository.go internal/nodecontrol/authority/postgres_repository_test.go internal/nodecontrol/authority/postgres_repository_integration_test.go
git commit -m "feat(nodecontrol): persist authority fence visibility"
```

### Task 7: Prove deterministic crash recovery and PITR fail-closed behavior

**Files:**
- Create: `internal/nodecontrol/authority/coordinator.go`
- Test: `internal/nodecontrol/authority/coordinator_test.go`
- Test: `internal/nodecontrol/authority/authority_crash_integration_test.go`
- Test: `internal/nodecontrol/authority/authority_pitr_integration_test.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`

**Interfaces:**
- Consumes: `authority.Provider`, `authority.Repository`, read-only `EffectResolver`, `securitykit.Clock`.
- Produces: `func NewCoordinator(Provider, Repository, EffectResolver, securitykit.Clock) (*Coordinator, error)`; `Reserve`, `Finalize`, `Abort`, `Recover`, and `CheckReady` methods; readiness result that keeps bootstrap/agent/operator/signing gates closed on any unexplained mismatch.

```go
type Readiness struct {
	ProviderHead Head
	DatabaseHead DatabaseHead
	Ready        bool
	Reason       ReadinessReason
}

type EffectState string

const (
	EffectAbsent    EffectState = "absent"
	EffectPrepared  EffectState = "prepared"
	EffectCommitted EffectState = "committed"
	EffectTerminal  EffectState = "terminal"
)

type ResolvedEffect struct {
	Kind         EffectKind
	ScopeKind    ScopeKind
	ScopeDigest  contracts.Digest
	EffectDigest contracts.Digest
	State        EffectState
}

type EffectResolver interface {
	ResolveAuthorityEffect(context.Context, uuid.UUID) (ResolvedEffect, error)
}

type CoordinatorFinalizeRequest struct {
	OperationID  uuid.UUID
	EffectDigest contracts.Digest
}

func (c *Coordinator) Reserve(context.Context, ReserveRequest) (Reservation, error)
func (c *Coordinator) Finalize(context.Context, CoordinatorFinalizeRequest) (Receipt, error)
func (c *Coordinator) Abort(context.Context, AbortRequest) (Receipt, error)
func (c *Coordinator) Recover(context.Context, uuid.UUID) (Receipt, error)
func (c *Coordinator) CheckReady(context.Context) (Readiness, error)
```

- [ ] **Step 1: Write the RED crash matrix test**

The table has these exact crash points: provider reserve committed before response; DB pending commit before caller response; domain effect transaction committed before database-point capture; database point captured before bind; DB effect/database bind before provider finalize; provider finalize committed before response; provider finalize response before visibility activation; visibility activation commit before response. For each point, restart a new coordinator over the same fake/provider state and assert `Recover(operationID)` reaches one terminal receipt without a second sequence or second effect and with `required_lsn` not earlier than the domain-effect commit.

- [ ] **Step 2: Run the RED coordinator test**

Run: `go test ./internal/nodecontrol/authority -run TestCoordinatorCrashRecoveryMatrix -count=1`

Expected: FAIL because `Coordinator` does not exist.

- [ ] **Step 3: Implement bounded provider calls and operation-ID recovery**

`Reserve` calls the provider and validates the reservation but does not begin a database transaction. Domain code records that reservation with `Repository.RecordPending` in its own transaction. `CoordinatorFinalizeRequest` deliberately contains only `operation_id` and `effect_digest`; callers cannot submit a system ID, timeline, LSN, epoch, sequence, kind, scope, or receipt. After the immutable effect transaction commits, `Finalize` loads the pending row and resolved effect and requires the operation/kind/scope/scope-digest/effect-digest tuple to match exactly. If the row is unbound, `Finalize` calls `Repository.CaptureDatabasePoint` itself on the same configured primary and binds that exact effect/database group in a second short transaction; if the row is already identically bound, it reuses the stored database point without capturing a later WAL position; any partial or different binding is a conflict. It constructs the provider-only `FinalizeRequest` from the stored operation/effect and the resulting bound point, then calls the provider. `EffectResolver` is a composition-root dispatcher over the finite effect kinds; it is read-only and cannot call the provider. `Recover` compares provider `Inspect`, the fence row and `ResolveAuthorityEffect(operationID)`: absent unbound work may be aborted, prepared work stays fail closed, and committed work may bind/finalize only when kind/scope/effect all match. It never fabricates a receipt, accepts caller-supplied database coordinates, reuses a pre-commit LSN or chooses between conflicting values.

- [ ] **Step 4: Run the GREEN crash matrix**

Run: `go test ./internal/nodecontrol/authority -run TestCoordinatorCrashRecoveryMatrix -count=1`

Expected: PASS at all eight crash points.

- [ ] **Step 5: Write and run the PITR RED integration test**

The test reaches sequence 11 and takes a run-owned physical backup, then commits a revoke at sequence 12, restores the sequence-11 backup through the harness while preserving its PostgreSQL system identity/timeline, and reconnects it to the unchanged provider head at sequence 12. `CheckReady` must return `Ready=false`, reason `database_behind_provider`; active-authority query helpers return `ErrAuthorityUnavailable`. A separate deliberately different cluster asserts `database_identity_mismatch`, so reason precedence is deterministic rather than calling the historical restore “fresh.”

Run: `go test -tags=integration ./internal/nodecontrol/authority -run TestPITRBeforeRevocationFailsClosed -count=1 -timeout 3m`

Expected before readiness comparison is implemented: FAIL because the restored database is incorrectly accepted.

- [ ] **Step 6: Implement provider/database head and timeline comparison**

Apply the exact empty-head normalization from Task 6, then compare only the provider's latest epoch with database rows from that same epoch; sequence zero is legal only for the fully empty normalized pair and never passes readiness once the provider is nonempty. Match latest reserved sequence/digest and latest committed sequence/operation ID/receipt digest/database point as indivisible anchors. Reject provider unavailable, epoch mismatch, database lower sequence, database higher unexplained sequence, `record_count != latest_reserved_sequence`, same coordinate different reservation/receipt digest, missing or different latest committed operation/database point, system ID mismatch, lower timeline, or current WAL position older than the committed provider receipt's required LSN. Require `len(Repository.ListPending(ctx, epoch)) == DatabaseHead.PendingCount`; inspect every returned pending row through both provider and `EffectResolver`, including its exact reservation digest and all-or-none bound group. Any absent conflict, still-prepared operation or unexplained terminal mismatch keeps readiness false until exact recovery finishes. Return only finite reasons and never include provider or database values in public output.

- [ ] **Step 7: Run the GREEN PITR and readiness tests**

Run: `go test -tags=integration ./internal/nodecontrol/authority -run 'TestPITRBeforeRevocationFailsClosed|TestAuthorityReadiness' -count=1 -timeout 3m`

Expected: PASS; old certificate/desired fixture rows remain unreachable while the mismatch exists.

- [ ] **Step 8: REFACTOR readiness into an explicit probe without widening existing public names**

Add an internal authority probe whose public `Name()` is the fixed low-cardinality value `authority`; extend the readiness checked-name allowlist and tests. Its `Ping` returns a value-free error unless `CheckReady.Ready` is true.

Run: `go test ./internal/readiness ./internal/nodecontrol/authority -count=1`

Expected: PASS.

- [ ] **Step 9: Run the Batch 01 exit gate**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/readiness ./internal/store -count=1 -timeout 5m`

Expected: PASS.

Run: `go test -tags=integration ./internal/nodecontrol/authority ./internal/store -count=1 -timeout 5m`

Expected: PASS with migration roundtrip, crash recovery, same-sequence fork rejection, and PITR fail-closed coverage.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Expected: PASS followed by `git diff --exit-code -- api gen internal/store` exiting 0.

- [ ] **Step 10: Commit the Batch 01 gate**

```bash
git add internal/nodecontrol/authority/coordinator.go internal/nodecontrol/authority/coordinator_test.go internal/nodecontrol/authority/authority_crash_integration_test.go internal/nodecontrol/authority/authority_pitr_integration_test.go internal/readiness/checker.go internal/readiness/checker_test.go
git commit -m "test(nodecontrol): prove authority crash and PITR safety"
```

## Batch 01 completion check

- Three generated OpenAPI packages contain disjoint route sets, exact listener-auth extensions and no pseudo-mTLS security schemes.
- Nodecontrol event wire vectors are deterministic, closed, and privacy-safe.
- Migration Up/Down/Up recreates every §9 authority table and constraint.
- Deterministic provider gives a total, nonreused sequence order, node-scoped committed checkpoints and exact retry semantics.
- PostgreSQL exposes only committed/active fence records.
- Every crash point recovers by operation ID without a second authority effect.
- A pre-revocation PITR database remains fail closed against the unchanged external provider head.
- No Task 2 or Task 3 production behavior has entered this batch.
