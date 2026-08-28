# Talenro C1.2 Batch 01 Contracts, Schema, and Authority Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付三份隔离的 node OpenAPI、nodecontrol Protobuf event、Abort/serving 原语，以及 v7 全部 canonical body/envelope/evidence/golden、typed verifier、base 25 + additive 26 的 51-table PostgreSQL catalog、registered Go 00007/Down guard、exact-epoch repository/readiness、`BoundAuthorityReadSource` 与 catalog/crash/PITR 证据。

**Architecture:** 契约先行：OpenAPI、Protobuf 和 `internal/nodecontrol/contracts` 冻结跨分册 wire/domain 名称、JCS/digest/signature/evidence DAG 与 v7 provider-facing contracts。`00006` 只补 Goose parser annotations；authoritative executor 才注册的 Go `00007` 在同一 Goose transaction 安装 26 张默认关闭的 additive protocol table、ACL、guard 与 lock registry。B01 实现 pure verifier、DB evidence/repository 和 deterministic fakes；B03 实现 production provider/incarnation/runtime/timeline/WAL archive，B11 不在 DB transaction 内调用 provider/attestor。唯一外部调用例外是冻结 v7 §9.3 的 disposable Down seam：`authority_protocol_downgrade_authorizer` 在 Goose-owned transaction 内签入实际 txid，且任何其他 signer/API/phase 都不得复用该例外。`internal/nodecontrol/serving` 仍只通过 Coordinator-owned 单连接能力执行 exact-epoch readiness-query-readiness。

**Tech Stack:** Go 1.26.5、OpenAPI 3.0.3、oapi-codegen 2.8.0、Protobuf Go 1.36.11、Buf 1.72.0、PostgreSQL 18.4、pgx 5.10.0、sqlc 1.31.1、goose 3.27.1、RFC 8785 JCS、SHA-256、现有 strictjson/outbox 约束。

**Spec:** [Approved C1.2 node and POP control-plane base design](../specs/2026-08-23-node-pop-control-plane-design.md), approved SHA-256 `B0B0DDBBD11546FC07A25CB76992E375481192A3261270FF442095501CC2B4B5`; [approved authority Abort/serving amendment](../specs/2026-08-24-nodecontrol-authority-abort-serving-design.md), approved content SHA-256 `86996084462A5DE1E7667D56A135E93099EE38E1089464CCDFCFF7284EA0D1D7`; [approved authority v7 upgrade amendment](../specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md), approved content SHA-256 `EFAEBE52BDC3D70BDA8737C893B02752E60ACEC0A08441813CADF1425079FC8E`; [approved canonical authority evidence/dispatcher addendum](../specs/2026-08-28-nodecontrol-authority-canonical-dispatcher-addendum-design.md), approved content SHA-256 `7D480607A92214627C1CEF3E81AF76610EC861508CE8D510D5BD046E82600FE2`; canonical set manifest [c12-spec-set.v1.json](../specs/c12-spec-set.v1.json) contains the base and those three amendments in bytewise path order.

## Global Constraints

- 本册只完成索引中的 `C1.2-B01`；不得实现 operator mutation、state signing、certificate issuance、agent、supervisor 或 core adapter。
- B01 拥有全部 v7 canonical schema/body/envelope/evidence/golden、typed verifier、DB physical objects/ACL/guards/repository/readiness 与 deterministic fakes，但不实现 B03 production provider、non-exportable incarnation/runtime/timeline attestor、trusted WAL decoder、rollback-resistant commit archive/Fence/Inspect/first-consumer runtime，不计算 B10 `SpecDigest`，不执行 B11 production orchestration。
- v7 不支持 mixed-version rolling upgrade。durable/shared Up 需要 fresh `LocalRuntimeIsolationV1`、typed production grant 与 migration lock；普通 Goose CLI 未注册 00007，disposable catalog fixture 也必须显式传 `installation_kind=disposable_fixture`。
- `00007` 安装后默认关闭。activation/completion/release、runtime/incarnation result 链、provider stable Head 与 serving lease 未 exact 对齐时，ordinary writer/reader/listener/signer 都不可用；`fresh_v7_staging_closed` 仍是 `restore_incomplete`。
- 所有 provider/attestor 调用位于 SQL transaction/lock 之外；DB transaction 只消费已验证、完整 preimage 可用的 typed evidence，不从 opaque digest 或 caller bytes 猜测。
- `AuthorityEffectCommitmentV1`、`AuthorityEffectResolutionV1`、`ActivationDecisionEvidenceV1` 仍是 Abort/serving amendment 的历史 schema；Task 7 维护其现有 transcript，v7 fixed schema registry/golden 不得再登记三个同名新类型。
- [Suite index](2026-08-23-node-pop-control-plane.md) 冻结的路径与跨册接口逐字生效；本册不得重命名 `contracts.AuthorityVersion`、`authority.Provider` 或三个生成包。
- C1.1 `internal/trust`、C1.1 root schema、config signer key 与 `securitykit.EnrollmentGrantToken` 不得成为 C1.2 authority。
- PostgreSQL 是领域事实源；production anti-rollback authority 必须在独立 provider，普通表、文件、wall clock 和 Redis 都不能冒充 provider。
- Access-granting effect 在 provider committed receipt 和最终 visibility transaction 前不可被 serving query、listener authorizer、signer 或 active pointer读取。
- Fail-closed effect 可以先写数据库，但 provider finalize 前不得返回成功；任何不确定结果保持不可服务并按 operation ID 恢复。
- 只有 exact unbound/unclaimed fence 与同一显式 `READ COMMITTED` DBTX 中的 `EffectAbsent` 才能 durable-claim Abort；任何 prepared/committed/terminal、missing resolver、unknown kind 或 tuple mismatch均不得调用 provider Abort。
- 每个 authority-bearing INSERT/UPDATE/DELETE writer先锁 exact fence；test-only `authorityEffectGuardCoverage` 与 `authorityEffectOutcomeCoverage` 必须对新增路径缺失、错误 phase/event/parent 或重复 kind fail closed。
- Provider committed Receipt、stored Head/checkpoint、reason/time/identity/typed activation-input preimage、evidence/resolution JCS+digest、DB fence `committed/active`、domain disposition/pointer与 required audit/outbox只允许在 Coordinator-owned activation transaction 中一起提交；rollback-resistant proof在所有写后紧邻唯一 Commit前 consume且首次失败也 burn，none-time proof绝不 consume，Commit不确定先从同一locked snapshot做parsed contextual/domain validation。
- Rollback-resistant time evidence的 monotonic budget在 external time/attestation调用前开始，最长5秒；constructor完成与所有锁/写入后唯一 `Commit` 紧邻前均使用 strict `< deadline`，uncertain COMMIT先读 stored resolution。
- 仓库没有完整、签名、operator-owned deployment ledger，因此本计划固定采用 additive `00007` 承载 semantic upgrade；`00006` 仅加入 StatementBegin/StatementEnd。不得以本地 Goose status、空测试库或未找到部署记录改成 in-place semantic edit。
- Final catalog 数字固定为 base `00006` 25 张 + v7 `00007` 26 张 = 51 张 `nodecontrol` table；Task 4 的 25-table 清单只是实施中间点，Task 8 后 `db/schema/nodecontrol.v1.yaml` 必须逐 object 精确覆盖 51 张，不得把 26 张 v7 protocol table 视为外部 provider state。
- V7 body digest 固定为 `SHA-256(ASCII(schema) || 0x00 || JCS(body))`；signature envelope、schema↔signer-role/policy/key/root/algorithm 映射、CanonicalEvidenceBundleV1 必须封闭验证 unknown/missing/extra/duplicate/noncanonical 输入。
- Generic commit archive contracts 必须覆盖 fixed 10-row-kind registry、related-row-set、purpose/generation evidence、application+decision sorted multi-key Fence、semantic-ID journal + cross-reason epoch prefix subject-slot/reverse unique、stable-holder once-row、global latch/reverse owner、Inspect candidate echo/history 与 staging evidence mode。同 holder lease rollover 只更新 freshness，不产生新 once-key。
- DB lock order 和 archive contract order 分别固定在 v7 设计的 canonical registry；实现和 test 不使用 sleep 猜测竞态，不允许 primary stream 后再取 related-edge lock，不允许 provider call 持有 DB lock。
- Epoch recovery 保持唯一 non-null prefix key、apply-vs-replacement DB 仲裁、candidate-independent decision、pre-CAS fresh candidate proof、successful-CAS-only consume、first-consumer/history tail、pre/post-catchup 与 fork-cut/result reconcile 见证；Plan 01 测试 pure DAG/repository，Plan 03 实现 production CAS。
- Down 仅允许 disposable fixture 在 provider/member 永久 retirement 完整、fixed forbidden registry 为零时，由 registered Go migration 在同一 transaction 执行 `1/0 -> 1/1 -> 0/0`。production 或任意 protocol/source/staging/claim-v1 row 使 Down 永久失败。
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
db/migrations/00007_nodecontrol_authority_abort_serving.go
db/migrations/assets/nodecontrol_authority_v7_up.sql
db/migrations/assets/nodecontrol_authority_v7_down.sql
sqlc.yaml
db/schema/nodecontrol.v1.yaml
scripts/run-c12-integration.ps1
internal/testinfra/c12_dependencies_integration_test.go
internal/testinfra/c12_authority_pitr_integration.go
internal/testinfra/c12_integration_manifest_test.go
testdata/c12/integration-contracts-schema-authority.v1.json
db/queries/nodecontrol_authority.sql
db/queries/nodecontrol_serving.sql
internal/store/nodecontrol_authority.sql.go
internal/store/nodecontrol_serving.sql.go
internal/nodecontrol/contracts/authority.go
internal/nodecontrol/contracts/events.go
internal/nodecontrol/contracts/canonical.go
internal/nodecontrol/contracts/signature_envelope.go
internal/nodecontrol/contracts/evidence_bundle.go
internal/nodecontrol/contracts/authority_v7_genesis.go
internal/nodecontrol/contracts/authority_v7_epoch.go
internal/nodecontrol/contracts/authority_v7_staging.go
internal/nodecontrol/contracts/authority_v7_source.go
internal/nodecontrol/contracts/authority_v7_down.go
internal/nodecontrol/contracts/commit_archive.go
internal/nodecontrol/authority/provider.go
internal/nodecontrol/authority/deterministic_fake.go
internal/nodecontrol/authority/repository.go
internal/nodecontrol/authority/postgres_repository.go
internal/nodecontrol/authority/commitment.go
internal/nodecontrol/authority/resolution.go
internal/nodecontrol/authority/activation_evidence.go
internal/nodecontrol/authority/effect_dispatcher.go
internal/nodecontrol/authority/v7_verifier.go
internal/nodecontrol/authority/commit_archive_verifier.go
internal/nodecontrol/authority/database_head.go
internal/nodecontrol/authority/exact_epoch_readiness.go
internal/nodecontrol/serving/types.go
internal/nodecontrol/serving/repository.go
internal/nodecontrol/serving/postgres_repository.go
```

`db/queries/nodecontrol_inventory.sql`、`nodecontrol_state.sql`、`nodecontrol_identity.sql` 和 `nodecontrol_observation.sql` 由后续对应分册创建；本册 migration 必须已经建立它们将使用的表。

`testdata/c12/authority-v7/` 中的 literal cross-language golden 也属 B01。B03 只实现这些 contract 的 production provider/attestor/archive adapter，B10 只消费 spec-set digest，B11 只调用 typed API 并提供 exact preimage；三者均不得复制 B01 canonicalizer/verifier 或创建 v7 table。

### Task 1: Freeze pure authority comparison contracts

**Files:**
- Create: `internal/nodecontrol/contracts/authority.go`
- Test: `internal/nodecontrol/contracts/authority_test.go`
- Test: `internal/nodecontrol/contracts/authority_fuzz_test.go`

**Interfaces:**
- Consumes: `errors.New`, fixed-width `[32]byte` values, unsigned integer ordering.
- Produces: `type Digest [32]byte`; `type AuthorityVersion struct { Epoch uint64; Sequence uint64 }`; authority-bearing `type VersionedDigest struct { Version uint64; AuthoritySequence uint64; Digest Digest }`; local-artifact-only `type LocalVersionedDigestV1 struct { Version uint64; Digest Digest }`; `func (AuthorityVersion) Validate() error`; `func CompareVersionedDigest(VersionedDigest, VersionedDigest) (Comparison, error)`; and `func CompareLocalVersionedDigest(LocalVersionedDigestV1, LocalVersionedDigestV1) (Comparison, error)`.

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

type LocalVersionedDigestV1 struct {
	Version uint64
	Digest  Digest
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

`LocalVersionedDigestV1` is allowed only for host-local reviewed artifacts that have no control-plane authority sequence, specifically the approved-release manifest and installed-release map in Plans 04–05. `CompareLocalVersionedDigest` requires versions in `1..MaxInt64` and nonzero digests；lower version is rollback, equal version/equal digest is idempotent, equal version/different digest is fork, and higher version is advance. It must never be used for root、metadata、desired、recovery、resource envelope or host-memory policy. Extend the table/fuzz tests with every local outcome, zero/overflow, same-version fork and a compile-time assertion that the two types are not aliases.

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

Run: `go test ./internal/nodecontrol/contracts -run '^$' -fuzz '^FuzzCompareVersionedDigestNeverInventsOutcome$' -fuzztime=10s -timeout 30s`

Expected: PASS after an actual fuzz-engine run；a seed-only `-run Fuzz...` is not this gate.

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
- Create: `internal/nodecontrol/contracts/trust_conflict.go`
- Test: `internal/nodecontrol/contracts/trust_conflict_test.go`
- Test: `internal/nodecontrol/contracts/openapi_contract_test.go`

**Interfaces:**
- Consumes: OpenAPI 3.0.3, existing `oapi-codegen` std-http server conventions, canonical JSON names from the approved spec.
- Produces: packages `nodebootstrapv1`, `nodeagentv1`, and `nodeoperatorv1`; their generated `ServerInterface`, typed clients (`ClientInterface`/`ClientWithResponsesInterface`), request/response models, parameter structs, `HandlerWithOptions`, and embedded specs；plus the shared pure `contracts.DeriveTrustConflictLocalFaultIDV1` helper used independently by server and agent.

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

`NodeActorHighWaterV1.node_authority_checkpoint` is intentionally only the last scalar checkpoint sequence observed by the node；it is not a `NodeHighWaterV1`, carries no caller-supplied digest and cannot be promoted to an artifact-backed per-node trust-conflict stream. The authoritative `CommittedNodeCheckpoint` additionally has its provider-derived receipt digest, but that digest is server-side evidence and is never inferred from the scalar request field. A client scalar ahead of the DB/provider-consistent checkpoint must fail through the headerless global authority-readiness/Head/Inspect path；it cannot create a trust-conflict incident/header/notice/local latch or request nonexistent checkpoint evidence. Contract/route tests freeze this distinction and reject any schema extension that supplies a client checkpoint digest outside a separately approved contract.

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
| `pollNodeDesiredState` | `DesiredPollRequestV1={boot_id,agent_build_digest,capability_schema_digest,high_water,time_nonce?}` | `200 NodeStatePollResponseV1={root_chain?,metadata?,desired_state?,time_attestation?}`、empty `204`, or bodyless `409 NodePollConflict` | node mTLS; 25-second wait; no delta |
| `pollNodeRecoveryState` | `RecoveryPollRequestV1={boot_id,recovery_id,sorted_known_incident_ids,high_water,time_nonce}` | `200 NodeRecoveryPollResponseV1={root_chain?,metadata?,recovery_state,time_attestation}`、empty `204`, or bodyless `409 NodePollConflict` | node mTLS and exact recovery session/incident |
| `createNodeObservation` | `NodeObservationV1={node_id,identity_epoch,boot_id,sequence,observed_at,agent_build_digest,capability_schema_digest,seen_generation,seen_authority_sequence,seen_digest,applied_generation,applied_authority_sequence,applied_digest,sorted_slot_facts,reducer_version,reducer_digest,reducer_output}` | `ObservationAckV1={boot_id,sequence,report_digest,result}` | active node mTLS; at most 8 slot facts |
| `createNodeSecurityFault` | `SecurityFaultReportV1={operation_id,local_fault_id,subtype,evidence_digest,identity_epoch,boot_id,request_digest,supervisor_fault?}` | fixed `SecurityFaultReceiptV1` only | active/recovery-pending/recovery-limited node mTLS; specialized self-invalidating receipt gate |
| `createNodeTrustConflictEvidence` | `TrustConflictEvidenceRequestV1={incident_id,artifacts}` with one or two `SignedConflictArtifactV1={kind,canonical_bytes}`；`kind` is closed to declaration-order `desired_state`、`recovery_state`、`node_state_trust_metadata`、`node_state_root_set`、`server_ca_trust_bundle` | `TrustConflictEvidenceAckV1={incident_id,result}` with result only `accepted` or `escalated` | exact incident-bound node mTLS; one allowed 1 MiB exception |
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
| `ConflictArtifactKindV1` | declaration-order `desired_state`, `recovery_state`, `node_state_trust_metadata`, `node_state_root_set`, `server_ca_trust_bundle` |
| `TrustConflictEvidenceResultV1` | `accepted`, `escalated` |
| `ProtocolCapabilityV1` | `fixture_loopback`, `xray_loopback`, `sing_box_loopback` |
| `AdapterV1` | `fixture`, `xray`, `sing_box` |
| `OperatorStateV1` | `provisioning`, `enabled`, `draining`, `disabled` |
| `SecurityStateV1` | `normal`, `quarantined` |
| `IdentityStateV1` | `never_enrolled`, `active`, `recovery_pending`, `recovery_limited`, `revoked` |
| `RecoveryReasonV1` | `identity_compromise`, `administrative_disable`, `retire`, `authority_restore`, `security_incident` |
| `RecoveryActionV1` | `hold_stopped`, `submit_recovery_attestation`, `clear_security_latches` |
| `RestorePhaseV1` | `proposal`, `approval` |
| `OperationResultV1` | `pending`, `active`, `accepted`, `completed`, `superseded`, `rejected` |

`artifacts.maxItems=2`; incident IDs and all UUIDs use `CanonicalUUID`; every digest uses `DigestHex32` unless the canonical DTO explicitly transports 32 raw bytes. Replace the pre-existing `ConflictArtifactKindV1` values completely—none of `server_ca_bundle`、`node_state_root`、`node_state_metadata`、`node_desired_state` or `node_recovery_state` remains as an alias—and make `TrustConflictEvidenceAckV1.result` reference only the route-specific `TrustConflictEvidenceResultV1`, never the broader `OperationResultV1`. Contract and generated-Go tests enumerate those two exact constant sets in declaration order, reject missing/extra/alias/case values, compile-assign `SignedConflictArtifactV1.Kind` and `TrustConflictEvidenceAckV1.Result` to their respective generated named types, and require the generated ACK field itself to use the route-specific named type.

Both poll operations replace the generic `Conflict` response at status 409 with one shared route-specific `NodePollConflict` response: it has no content/body and only the optional response header `Talenro-Trust-Conflict-Incident-ID`, whose schema is `CanonicalUUID`. The header is present if and only if the server has fence-finalized an `unverified_client_highwater_conflict` or `client_highwater_ahead` incident for that exact request/credential；its value is that exact incident ID. Every other poll conflict is bodyless with the header absent. The header is never accepted on a request, never caller-selected, and no pending/unfinalized incident may emit it. Generated strict server/client response types must expose the header without a free-form map. Contract tests cover present/absent/malformed/duplicate header cases, 200/204/header exclusion, response-loss exact retry returning the same ID, and prove the header's ID is accepted only by the separately authorized evidence route for the same credential.

Freeze `TrustConflictLocalFaultBindingV1={node_id,incident_id,poll_request_digest,certificate_digest,identity_epoch}` in `internal/nodecontrol/contracts/trust_conflict.go`, where `certificate_digest` is exactly SHA-256 of the active leaf DER. `DeriveTrustConflictLocalFaultIDV1` validates canonical nonzero fields and hashes the unambiguous fixed-width transcript `TALENRO-TRUST-CONFLICT-LOCAL-FAULT-ID-V1\x00 || node_uuid[16] || incident_uuid[16] || poll_request_digest[32] || certificate_der_sha256[32] || identity_epoch_uint64_be`, then derives one canonical RFC 4122 variant/version-5 UUID from a frozen package namespace plus that 32-byte transcript digest. It does not depend on the later v7 canonical-schema registry. There is no random、caller-selected or header-carried local-fault ID. Server and agent vector tests use identical bytes and mutate every field；response loss/restart must derive the same ID, while another node、incident、request、certificate or identity epoch must derive a different ID.

Contract tests also enumerate every operation and assert its exact request ref, success statuses, auth extension, header allowlist, wire/decoded cap, and that no additional path/schema/security scheme exists. The operator contract test requires each list success response to reference its resource-specific list schema, requires `items` to have the exact resource `$ref`, `minItems: 0`, and `maxItems: 200`, recursively requires `additionalProperties: false` and a nonempty `required` array on every object schema, and rejects any anonymous/free-form item schema or generic list component.

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

Stage only this task's exact generator inputs/outputs after the first successful generation, then run generation a second time:

Run: `git add api/openapi/node-bootstrap-api.v1.yaml api/openapi/node-bootstrap-oapi-codegen.yaml api/openapi/node-agent-api.v1.yaml api/openapi/node-agent-oapi-codegen.yaml api/openapi/node-operator-api.v1.yaml api/openapi/node-operator-oapi-codegen.yaml gen/go/talenro/nodebootstrap/v1/server.gen.go gen/go/talenro/nodeagent/v1/server.gen.go gen/go/talenro/nodeoperator/v1/server.gen.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- api/openapi gen/go/talenro/nodebootstrap gen/go/talenro/nodeagent gen/go/talenro/nodeoperator`

Run: `git ls-files --others --exclude-standard -- api/openapi gen/go/talenro/nodebootstrap gen/go/talenro/nodeagent gen/go/talenro/nodeoperator`

Expected: both diff and untracked output are empty. Because the exact first-generation outputs are already in the index, this compares the second generation against that baseline and cannot miss newly generated files.

- [ ] **Step 8: Commit the three contracts**

```bash
git add api/openapi/node-bootstrap-api.v1.yaml api/openapi/node-bootstrap-oapi-codegen.yaml api/openapi/node-agent-api.v1.yaml api/openapi/node-agent-oapi-codegen.yaml api/openapi/node-operator-api.v1.yaml api/openapi/node-operator-oapi-codegen.yaml gen/go/talenro/nodebootstrap/v1/server.gen.go gen/go/talenro/nodeagent/v1/server.gen.go gen/go/talenro/nodeoperator/v1/server.gen.go scripts/generate.ps1 scripts/generate.sh internal/nodecontrol/contracts/trust_conflict.go internal/nodecontrol/contracts/trust_conflict_test.go internal/nodecontrol/contracts/openapi_contract_test.go
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

### Task 4: Freeze the 25-table base catalog before the registered v7 migration

**Files:**
- Create: `db/migrations/00006_nodecontrol.sql`
- Create: `db/schema/nodecontrol.v1.yaml`
- Create: `scripts/run-c12-integration.ps1`
- Create (first line `//go:build integration`): `internal/testinfra/c12_dependencies_integration_test.go`
- Test: `internal/testinfra/c12_integration_manifest_test.go`
- Test: `internal/store/nodecontrol_schema_test.go`
- Test: `internal/store/nodecontrol_schema_manifest_test.go`
- Test: `internal/store/nodecontrol_schema_integration_test.go`

**Interfaces:**
- Consumes: goose migration conventions, PostgreSQL 18 partial unique indexes/check constraints, existing UUID/timestamptz/sqlc overrides.
- Produces: the base `00006` 25-table schema, the first 25 entries of the exact catalog manifest, the owned isolated dependency runner's `base` profile, and the Go/parser exact integration-manifest validator. Task 8 extends the same catalog to 51 tables and the runner to registered `authority-v7`/PITR infrastructure；25/base is an intermediate state, never the Batch 01 exit count.

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

Add a second staged test that loads `db/schema/nodecontrol.v1.yaml`, requires exactly the 25 base table names below before Task 8, rejects duplicate/unknown table, column, constraint, enum, index or trigger names, and requires every manifest object to have an exact migration/catalog counterpart. Name it `TestNodeControlBaseCatalogHasExact25Tables` so it cannot be mistaken for the final catalog gate. Task 8 replaces the count assertion with `TestNodeControlV7CatalogHasExact51Tables`, retaining all 25 base object comparisons and adding all 26 v7 objects. Decode into typed structs with `KnownFields(true)`; reject YAML anchors/aliases/merge keys and any scalar that uses a compact semantic token instead of an exact type/default/check/predicate. Require every column object to carry explicit nullability/default/collation fields even when their value is null, every constraint/index/trigger to have a stable name, and every `authority_operation_id` to have the exact composite FK. The YAML parser is used only by tests; the runtime never reads the manifest.

- [ ] **Step 2: Run the RED schema test**

Run: `go test ./internal/store -run TestNodeControlMigrationContainsEveryAuthorityTable -count=1`

Expected: FAIL because `00006_nodecontrol.sql` does not exist.

- [ ] **Step 3: Add the machine manifest, migration header, authority table, and immutable fence constraints**

Encode the 25 base entries below in `db/schema/nodecontrol.v1.yaml` as the Task 4 intermediate artifact, then use this exact authority core in `00006`:

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

The following base-25 matrix is the human review coverage list for `00006`; it is not the final catalog count, is not permitted as YAML content, and none of its compact prose substitutes for the per-column catalog objects required above. Task 8 appends the v7 26-table registry and makes the final exact count 51:

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
24. `control_plane_trust_bundle_high_waters`: PK `(purpose,listener_kind,trust_domain)` with closed purpose/listener mapping, lower-case DNS domain, authority group, bundle version/digest, cumulative-set digest/count and updated time. Trigger accepts exact retry or strict advance with append-only cumulative set；rollback/fork fails；activation requires committed fence. Per the approved Abort/serving and v7 amendments this is active high-water only: it has no pending immutable publish effect, no production publisher/activator and no invented OOB protocol. `trust_bundle_publish` remains the dispatcher’s fixed unsupported registration until a separately approved durable workflow or equivalent OOB protocol changes the specification.
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

- [ ] **Step 6: Add the owned integration runner and run the GREEN migration roundtrip**

Implement the runner boundary frozen by the suite index. It generates one cryptographically random run suffix and starts exactly three run-owned containers: `postgres:18.4-alpine3.23`, `redis:8.8.1-alpine3.23`, and `nats:2.14.3-alpine3.22` with JetStream enabled. Docker assigns every published port on `127.0.0.1`; the runner captures and validates each exact name/ID/mapped port and waits at most 60 seconds for PostgreSQL health plus successful Redis/NATS protocol probes. It sets only child-process `TALENRO_DATABASE_URL`, `TALENRO_REDIS_ADDRESS`, and `TALENRO_NATS_URL`, runs ordinary Goose through 00006, then invokes the requested tagged Go test via a bounded argument array. Its Windows PowerShell 5.1-compatible public package input is one `-Packages` string with strict pipe-delimited explicit `./package` tokens；validate the entire string, reject recursive `./...`、empty/duplicate/metacharacter segments, then split without evaluation. Every Docker/Go invocation uses the native call operator and splatted validated array (`& docker @dockerArgs`, `& go @goArgs`) followed immediately by `$LASTEXITCODE` validation；static tests reject `Invoke-Expression`, `cmd /c`, nested `powershell -Command`, `.Arguments` string construction, concatenated command lines or output-based cleanup targeting. `base` is the only accepted profile in Task 4. A `finally` block re-inspects all three names/IDs before stopping/removing only those exact containers and reports test and each cleanup failure separately. Static/unit tests also reject inherited/fixed dependency URLs or addresses、non-loopback/fixed ports、production markers、unknown packages/flags、cleanup ID mismatch and a focused/group timeout above 30 minutes；suite-mode global budgets are the larger exact values frozen by the index and are validated separately. `internal/testinfra/c12_dependencies_integration_test.go` verifies the three run markers/endpoints, PostgreSQL base migration high-water, Redis isolation, and NATS JetStream availability；it never starts Docker itself.

Run: `go test ./internal/store -run TestNodeControlMigrationContainsEveryAuthorityTable -count=1`

Expected: PASS.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/store' -Run '^TestNodeControlMigrationUpDownUp$' -Timeout 3m`

Expected: PASS; Up, Down, and second Up all complete, and the second schema digest equals the first.

- [ ] **Step 7: REFACTOR the integration test to inspect constraints, not migration text**

Query `pg_catalog.pg_constraint`, `pg_indexes`, `pg_trigger`, `pg_proc`, and `information_schema.columns`; normalize each result and compare it field-for-field with `db/schema/nodecontrol.v1.yaml`, including enum members, defaults/nullability, fixed digest lengths, numeric bounds, partial predicates, FKs, trigger names, pointer relationships and retention columns. Fail on extra as well as missing objects. Run:

`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/store' -Run '^TestNodeControlMigration' -Timeout 3m`

Expected: PASS against PostgreSQL 18.4.

- [ ] **Step 8: Commit the independently green base migration**

```bash
git add db/migrations/00006_nodecontrol.sql db/schema/nodecontrol.v1.yaml scripts/run-c12-integration.ps1 internal/testinfra/c12_dependencies_integration_test.go internal/testinfra/c12_integration_manifest_test.go internal/store/nodecontrol_schema_test.go internal/store/nodecontrol_schema_manifest_test.go internal/store/nodecontrol_schema_integration_test.go
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

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/nodecontrol/authority' -Run '^TestPostgresRepositoryRecordsExactFenceState$' -Timeout 3m`

Expected: PASS against a fresh migrated PostgreSQL 18.4 database.

- [ ] **Step 6: REFACTOR to prove pending rows are invisible to active-head checks**

Add concurrent integration cases for reserved-only, effect-bound, committed-active, aborted, same-sequence fork, epoch rollover with sequence reset, node A/node B/global checkpoint effects, and timeline mismatch. Assert `DatabaseHead` aggregates only the latest epoch, reports pending count/sequence gaps and exact reservation/receipt/database anchors, lists every pending operation ID, and advances committed sequence/checkpoint only for `provider_status='committed' AND visibility_state='active'`.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/nodecontrol/authority' -Run '^(TestPostgresRepository|TestAuthorityHead)$' -Timeout 3m`

Expected: PASS.

- [ ] **Step 7: Verify generation drift and commit**

Stage the exact first-generation query/store outputs, run generation again, then compare the worktree to that index baseline:

Run: `git add db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_authority.sql internal/store`

Expected: both outputs are empty after the second generation；untracked generated files cannot evade the gate.

```bash
git add db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/authority/repository.go internal/nodecontrol/authority/postgres_repository.go internal/nodecontrol/authority/postgres_repository_test.go internal/nodecontrol/authority/postgres_repository_integration_test.go
git commit -m "feat(nodecontrol): persist authority fence visibility"
```

### Task 7: Freeze effect commitment, resolution, evidence, and dispatcher contracts

This task implements the approved Abort/serving amendment as amended by the approved 2026-08-28 canonical evidence/dispatcher addendum. `AuthorityEffectCommitmentV1`, `AuthorityProviderHeadV1`, `AuthorityCheckpointAnchorV1`, `ActivationDecisionEvidenceV1`, and `AuthorityEffectResolutionV1` remain historical authority-layer schemas/domains rather than v7 signer-role bodies；Task 11's v7 schema registry must reject duplicate registration. Task 7 proves canonical primitives、opaque admission and dispatcher mechanics only；Task 9 owns DBTX/lock ordering、Provider provenance、atomic persistence and Commit uncertainty.

**Files:**
- Create: `internal/nodecontrol/authority/commitment.go`
- Create: `internal/nodecontrol/authority/resolution.go`
- Create: `internal/nodecontrol/authority/activation_evidence.go`
- Create: `internal/nodecontrol/authority/effect_dispatcher.go`
- Test: `internal/nodecontrol/authority/commitment_test.go`
- Test: `internal/nodecontrol/authority/activation_evidence_test.go`
- Test: `internal/nodecontrol/authority/effect_dispatcher_test.go`
- Test: `internal/nodecontrol/authority/testdata/authority-effect-vectors.json`

**Interfaces:**
- Consumes: Task 1 `contracts.Digest`, Task 5 `EffectKind`/`ScopeKind`/`Reservation`/`Receipt`/`Head`/`NodeCheckpoint`/`ResolvedEffect`, RFC 8785 JCS, `store.DBTX`.
- Produces: strict New/Parse/accessor APIs for commitment、provider Head snapshot、checkpoint anchor、evidence and resolution；contextual branded proof validation；`RegisteredEffectResolver`、three-method `RegisteredEffectActivator` and the sole sealed dispatcher. B02/B03 may implement registered handlers but may not construct opaque queries、aggregate dispatch or reimplement transcripts.

```go
type TransactionalEffectQuery struct {
	registeredKind EffectKind
	expected       Reservation
}

func (q TransactionalEffectQuery) RegisteredKind() EffectKind
func (q TransactionalEffectQuery) Expected() Reservation

type TransactionalResolvedEffect struct {
	OperationID uuid.UUID
	Epoch       uint64
	Sequence    uint64
	Effect      ResolvedEffect
}

type RegisteredEffectResolver interface {
	ResolveRegisteredAuthorityEffectForUpdate(context.Context, store.DBTX, TransactionalEffectQuery) (TransactionalResolvedEffect, error)
}

type TransactionalEffectResolver interface {
	ResolveAuthorityEffectForUpdate(context.Context, store.DBTX, Reservation) (ResolvedEffect, error)
}

type RegisteredEffectActivator interface {
	CaptureActivationDecisionMaterial(context.Context, Receipt) (ActivationDecisionMaterial, error)
	ActivateAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence) error
	ValidatePersistedAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence, AuthorityEffectResolution) error
}

type TransactionalEffectActivator interface {
	RegisteredEffectActivator
}

type EffectDispatcher interface {
	TransactionalEffectResolver
	TransactionalEffectActivator
	authorityEffectDispatcher()
}

type EffectRegistration struct {
	Kind      EffectKind
	Resolver  RegisteredEffectResolver
	Activator RegisteredEffectActivator
}

func NewEffectDispatcher([]EffectRegistration) (EffectDispatcher, error)

type ActivationDecisionMaterial struct {
	Commitment                     AuthorityEffectCommitment
	Reason                         AuthorityEffectReason
	CheckpointKind                 CheckpointKind
	CheckpointScopeDigest          contracts.Digest
	TrustedTimeKind                TrustedTimeKind
	TrustedInstant                 time.Time
	EvidenceValidUntil             time.Time
	AttestationExpiresAt           time.Time
	ActivationDeadline             time.Time
	ProviderIdentityDigest         contracts.Digest
	ExpectedProviderIdentityDigest contracts.Digest
	FloorAttestationDigest         contracts.Digest
	Capability                     DecisionCapability
}

type ActivationDecisionEvidenceInput struct {
	Material     ActivationDecisionMaterial
	Receipt      Receipt
	ProviderHead AuthorityProviderHeadSnapshot
	Checkpoint   *AuthorityCheckpointAnchor
}

func NewAuthorityProviderHeadSnapshot(Head) (AuthorityProviderHeadSnapshot, error)
func ParseAuthorityProviderHeadSnapshot([]byte) (AuthorityProviderHeadSnapshot, error)
func NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput) (AuthorityCheckpointAnchor, error)
func ParseAuthorityCheckpointAnchor([]byte) (AuthorityCheckpointAnchor, error)
func ParseActivationDecisionEvidence([]byte) (ActivationDecisionEvidence, error)
func ValidateActivationDecisionEvidence(ActivationDecisionEvidence, ActivationDecisionEvidenceInput) (ValidatedActivationDecisionEvidence, error)
func ValidateAuthorityEffectResolution(AuthorityEffectResolution, AuthorityEffectCommitment, ValidatedActivationDecisionEvidence) error
func BeginActivationEvidenceCapture() *ActivationEvidenceCapture
func (c *ActivationEvidenceCapture) Complete(ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error)
```

- [ ] **Step 1: Write RED independent literal commitment/resolution/evidence vectors**

Freeze one strict version-1 fixture whose entries contain only hard-coded `name`、`artifact_kind`、semantic `input`、literal `canonical_jcs` and literal lowercase `digest_hex`, plus a mutation manifest classifying every field mutation as `reject` with an exact finite sentinel or `valid_alternate` with a different literal JCS/digest. Expected bytes/digests are reviewed constants, never regenerated during assertions. Cover the fixed empty activation-input digest and every absent marker plus exact domains `talenro.nodecontrol.authority-effect-commitment.v1\x00`, `talenro.nodecontrol.authority-provider-head.v1\x00`, `talenro.nodecontrol.authority-checkpoint-anchor.v1\x00`, `talenro.nodecontrol.activation-decision-evidence.v1\x00`, and `talenro.nodecontrol.authority-effect-resolution.v1\x00`; include conditional/final-not-applied, every resolution anchor, same/later Head, node/global/none checkpoint, rollback-resistant/none time and every valid matrix row.

```go
func TestAuthorityEffectCanonicalVectors(t *testing.T) {
	vectors := loadLiteralVectors(t, "testdata/authority-effect-vectors.json")
	for _, vector := range vectors {
		t.Run(vector.Name, func(t *testing.T) {
			gotJCS, gotDigest := constructByArtifactKind(t, vector.ArtifactKind, vector.Input)
			if !bytes.Equal(gotJCS, vector.CanonicalJCS) || gotDigest != vector.Digest { t.Fatalf("literal vector mismatch") }
		})
	}
}
```

- [ ] **Step 2: Run the RED canonical contract tests**

Run: `go test ./internal/nodecontrol/authority -run 'TestAuthorityEffectCanonicalVectors|TestAuthorityCanonicalMutationManifest|TestActivationDecisionEvidence' -count=1`

Expected: FAIL because the canonical values、parsers、contextual proof and dispatcher do not exist.

- [ ] **Step 3: Implement strict canonical constructors and monotonic capture**

Implement every approved required field、closed enum and branch matrix. Every parser rejects empty or over-4096-byte input before JSON/JCS processing；unknown/missing/extra/duplicate/null/JSON-number/noncanonical values fail, `schema_version` is exact string `"1"`, UUID/decimal/digest/time encoding is exact, and all Facts/JCS/Input accessors deep-copy. Provider Head accepts same-sequence only on exact operation/receipt/database-point match and later only in the same epoch；checkpoint anchors are same-epoch and strictly higher for higher-authority evidence. `NewActivationDecisionEvidence` constructs only the two none-time branches；rollback-resistant rows require `BeginActivationEvidenceCapture`/one-shot `Complete`, and Parse never mints origin or admission. Contextual validation recomputes every preimage/relation and returns the unconstructable `ValidatedActivationDecisionEvidence` while retaining admission only from the matching fresh Complete path.

- [ ] **Step 4: Write and run the RED finite dispatcher tests**

Require exactly 15 registrations：13 supported with non-nil resolver/activator and two fixed unsupported (`operator_authorizer_change`,`trust_bundle_publish`) with both nil. Assert missing/duplicate/unknown/typed-nil/non-nil-unsupported fail construction；the dispatcher is privately sealed and only it can construct a nonzero opaque query. Resolution probes all 13 supported kinds in frozen bytewise order with the same expected Reservation, validates operation/epoch/sequence echoes, and enforces zero/one/multiple cardinality. Cover absent dirty fields、corrupt echo、cross-kind/scope/tuple/digest、handler errors、direct material/activation/persisted-validation routing、fresh-vs-Parse origin、Receipt/proof/resolution binding and nil/typed-nil DBTX. Exact precedence and mapping are `ErrInvalidArgument`、`ErrConflict`、`ErrCanceled` or `ErrInjectedFailure` only；wrapped sentinels normalize and dynamic handler/SQL/domain text never escapes.

Run: `go test ./internal/nodecontrol/authority -run 'TestEffectDispatcherClosedRegistry|TestEffectDispatcherOpaqueQueries|TestEffectDispatcherProofRouting|TestEffectDispatcherFiniteErrors' -count=1`

Expected: FAIL before `EffectDispatcher` exists, then PASS after the minimal closed registry is implemented.

- [ ] **Step 5: Add admission boundary and uncertain-result tests**

Cover caller delay and trusted-time response just-under/equal/over bounds、checked overflow、one-shot Complete、copy/concurrent reuse、parse/restart no-token、none-time tokenless、exact final-not-applied and higher-authority branches, and rollback-resistant admission. The package-private consume helper changes shared `live → spent` before checking cancellation、expiry or binding；the first failed attempt therefore burns every copy. Fake handler tests prove only fresh proof reaches activation and only parsed proof reaches persisted validation. Explicitly do not claim BEGIN/lock/write/Commit ordering、atomic persistence or Commit uncertainty here；Task 9 owns those properties.

Run: `go test ./internal/nodecontrol/authority -run 'TestActivationDecisionEvidence|TestActivationAdmission' -count=1`

Expected: PASS with exact boundary handling, no origin upgrade and at most one admission claimant even when the first claim fails.

- [ ] **Step 6: Run fuzz/race and commit the canonical authority layer**

Run: `go test ./internal/nodecontrol/authority -run '^$' -fuzz '^FuzzAuthorityEffectCanonical$' -fuzztime=10s -timeout 30s`

Run: `go test ./internal/nodecontrol/authority -run '^$' -fuzz '^FuzzAuthorityEffectParsers$' -fuzztime=10s -timeout 30s`

Run: `go test -race ./internal/nodecontrol/authority -run 'TestAuthorityEffect|TestActivation|TestEffectDispatcher' -count=1`

Expected: PASS with no parser panic、aliasing、dynamic error leakage、data race or duplicate admission consumption.

```bash
git add internal/nodecontrol/authority/commitment.go internal/nodecontrol/authority/resolution.go internal/nodecontrol/authority/activation_evidence.go internal/nodecontrol/authority/effect_dispatcher.go internal/nodecontrol/authority/commitment_test.go internal/nodecontrol/authority/activation_evidence_test.go internal/nodecontrol/authority/effect_dispatcher_test.go internal/nodecontrol/authority/testdata/authority-effect-vectors.json
git commit -m "feat(nodecontrol): freeze authority effect decisions"
```

### Task 8: Install the registered additive v7 catalog, ACLs, guards, and Down latch

**Files:**
- Modify: `api/openapi/node-operator-api.v1.yaml`
- Modify: `gen/go/talenro/nodeoperator/v1/server.gen.go`
- Modify: `internal/nodecontrol/contracts/openapi_contract_test.go`
- Create: `internal/nodecontrol/contracts/authority_v7_boundary.go`
- Test: `internal/nodecontrol/contracts/authority_v7_boundary_test.go`
- Modify: `db/migrations/00006_nodecontrol.sql`
- Create: `db/migrations/00007_nodecontrol_authority_abort_serving.go`
- Create: `db/migrations/assets/nodecontrol_authority_v7_up.sql`
- Create: `db/migrations/assets/nodecontrol_authority_v7_down.sql`
- Modify: `sqlc.yaml`
- Modify: `scripts/run-c12-integration.ps1`
- Modify: `internal/testinfra/c12_dependencies_integration_test.go`
- Create (first line `//go:build integration`): `internal/testinfra/c12_authority_pitr_integration.go`
- Modify: `db/schema/nodecontrol.v1.yaml`
- Modify: `db/queries/nodecontrol_authority.sql`
- Modify: `internal/store/nodecontrol_authority.sql.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/querier.go`
- Modify: `internal/nodecontrol/authority/repository.go`
- Modify: `internal/nodecontrol/authority/postgres_repository.go`
- Create: `internal/nodecontrol/authority/v7_migration_grant.go`
- Create: `internal/nodecontrol/authority/v7_staging_import.go`
- Create: `internal/nodecontrol/authority/v7_down_authorization.go`
- Create (first line `//go:build integration`): `internal/nodecontrol/authority/v7_opaque_integration_fixture.go`
- Test: `internal/nodecontrol/authority/v7_migration_grant_test.go`
- Test: `internal/nodecontrol/authority/v7_staging_import_test.go`
- Test: `internal/nodecontrol/authority/v7_down_authorization_test.go`
- Test: `internal/store/nodecontrol_v7_catalog_test.go`
- Test: `internal/store/nodecontrol_v7_migration_integration_test.go`
- Test: `internal/store/nodecontrol_v7_acl_integration_test.go`
- Test: `internal/store/nodecontrol_v7_down_integration_test.go`
- Test: `internal/nodecontrol/authority/postgres_repository_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_staging_import_repository_integration_test.go`

**Interfaces:**
- Consumes: Task 4 base-25 manifest, Task 6 transaction-bound repository, and Task 7 canonical commitment/Head/checkpoint/evidence/resolution plus branded contextual-proof contracts.
- Produces: sealed `VerifiedAuthorityV7UpGrant`/`VerifiedAuthorityV7DownGrant` and `VerifiedFreshRestoreImportAdmission` types with unexported fields plus the exact defensive cross-package use APIs below, provider-scoped registered-Go migration, a final exact 51-table manifest, v7 output-only `IdentityStateV1=unauthorized`, profile-aware `StoredFence`/`AbortClaim` with a defensive locked persisted-outcome projection, the 26 additive tables, exact proof-column lifecycle on the eight existing domain owner tables, NOLOGIN role/ACL/guard/lock registries, exact `begin_staging_import` repository/store boundary, and the transactional Down `1/0 -> 1/1 -> 0/0` primitive consumed only by B11/B02 as assigned.

```go
package migrations

type AuthorityV7MigrationContext struct {
	InstallationKind InstallationKind
	UpGrant          *authority.VerifiedAuthorityV7UpGrant
	DownGrant        *authority.VerifiedAuthorityV7DownGrant
	DownAuthorizer   authority.AuthorityProtocolDowngradeAuthorizer
}

func NodeControlAuthorityV7Migration() *goose.Migration
func WithAuthorityV7MigrationContext(ctx context.Context, in AuthorityV7MigrationContext) context.Context
```

```go
package authority

type PersistedAuthorityEffectOutcome struct {
	CommitmentJCS                  []byte
	CommitmentDigest               contracts.Digest
	Receipt                        Receipt
	ProviderHeadJCS                []byte
	ProviderHeadDigest             contracts.Digest
	CheckpointAnchorJCS            []byte
	CheckpointAnchorDigest         contracts.Digest
	Reason                         AuthorityEffectReason
	AttestationExpiresAt           time.Time
	ActivationDeadline             time.Time
	ExpectedProviderIdentityDigest contracts.Digest
	EvidenceJCS                    []byte
	EvidenceDigest                 contracts.Digest
	ResolutionJCS                  []byte
	ResolutionDigest               contracts.Digest
}

type StoredFence struct {
	Record           Record
	AbortClaim       *AbortClaim
	PersistedOutcome *PersistedAuthorityEffectOutcome
}

func ConsumeVerifiedAuthorityV7UpGrant(VerifiedAuthorityV7UpGrant) (contracts.AuthorityV7UpMigrationFactsV1, error)
func ConsumeVerifiedAuthorityV7DownGrant(VerifiedAuthorityV7DownGrant) (contracts.AuthorityV7DownMigrationFactsV1, error)
func ConsumeVerifiedAuthorityV7DownAuthorization(VerifiedAuthorityV7DownAuthorization) (contracts.AuthorityV7DownAuthorizationFactsV1, error)
func ViewVerifiedFreshRestoreImportAdmission(VerifiedFreshRestoreImportAdmission) (contracts.FreshRestoreImportProjectionInputV1, error)

type AuthorityProtocolDowngradeAuthorizer interface {
	AuthorizeAuthorityProtocolDowngrade(context.Context, contracts.AuthorityV7DownAuthorizationRequestV1) (VerifiedAuthorityV7DownAuthorization, error)
}

type FreshRestoreImportRepository interface {
	ConsumeVerifiedFreshRestoreImport(context.Context, VerifiedFreshRestoreImportAdmission, contracts.FreshImportTopologyProjectionV1) (contracts.FreshRestoreImportApplicationV1, error)
}
```

Task 8 extends the Task 6 `Repository` interface with exact method `Lock(context.Context, store.DBTX, uuid.UUID) (StoredFence, error)`；the operation ID selects the fence while the supplied DBTX owns its lock and snapshot. `Repository.Lock` returns the `StoredFence` and `PersistedOutcome` from that one caller-owned locked snapshot；all byte slices and pointers are defensive copies. `Receipt` is reconstructed only from the exact fence-bound provider terminal fields and database point in that snapshot. A nil outcome means no atomically terminal proof group exists；a non-nil outcome is complete under the branch matrix below and is never synthesized from a mutable current Head. Task 9 parses/recanonicalizes every JCS preimage and validates all digests before invoking the domain persisted validator.

`NodeControlAuthorityV7Migration` is exported by package `db/migrations`; the opaque grants、Down authorization and import admission remain in package `internal/nodecontrol/authority`. Each opaque value contains an unexported pointer to sealed normalized facts plus a shared monotonic consume state, so copying the Go value cannot duplicate its one production write right. The `Consume*` functions defensively return only fixed normalized facts and consume the corresponding write right exactly once；Up facts contain installation/profile/upgrade-intent/expiry/nonce bindings, while the pre-transaction Down grant contains disposable classification、complete retirement/catalog/latch bindings and the fixed authorization seed but no fabricated database txid. After Goose supplies the real txid and the callback constructs/stores `PristineDowngradeInventoryV1`, `DownAuthorizer` is invoked exactly once inside that same transaction；it must return `VerifiedAuthorityV7DownAuthorization` whose normalized facts bind the exact request、actual txid、transaction nonce、inventory/retirement/catalog/anchor digests、role/policy and two-minute window. `ViewVerifiedFreshRestoreImportAdmission` is non-consuming and returns defensive copies of only the allowlisted manifest topology plus exact admission bindings needed by B02；the repository method performs the one consuming write. No accessor exposes signer key、credential、provider/archive handle or raw constructor.

Task 8 creates `contracts/authority_v7_boundary.go` before any authority package references it. That file owns the closed normalized DTOs `AuthorityV7UpMigrationFactsV1`、`AuthorityV7DownMigrationFactsV1`、`AuthorityV7DownAuthorizationRequestV1`、`AuthorityV7DownAuthorizationFactsV1`、`FreshRestoreImportProjectionInputV1`、`FreshImportTopologyProjectionV1` and `FreshRestoreImportApplicationV1` plus only their nested value records. Their literal field registries are the exact migration-latch/upgrade-intent subset of v7 §5.1, the exact pristine/authorization subset of §9.3, and the exact manifest/projection/application fields of §7.2；they use only already-defined contracts primitives、UUID/time/value slices, never a type introduced by Tasks 11–13. `authority_v7_boundary_test.go` must use external `package contracts_test`；it compile-instantiates every DTO through `contracts`、`authority` and `migrations`, asserts the literal JSON/field registry and defensive-copy behavior, and prevents later files from redeclaring or widening them without creating a test-only import cycle. Any v7 integration test that imports `db/migrations` likewise uses external `package authority_test` or a separate no-back-edge harness；same-package authority seam tests are forbidden from importing migrations. Tasks 12–13 add canonical signed bodies/verifiers but reuse these boundary DTOs.

Task 8 defines all three opaque types/use APIs. Same-package `_test.go` helpers cover unit tests；the build-tagged `v7_opaque_integration_fixture.go` uses `//go:build integration` and exposes narrowly typed fixed factories so `internal/store` integration tests can import sealed fixtures. A production-build static test must prove both those factory symbols and the `c12_authority_pitr_integration.go` control helper are absent without the tag, while exact-first-line tests require both files to declare `//go:build integration`. Task 12 implements production verifiers returning migration grants and Task 13 the fresh-import admission；later tasks consume Task 8 types while Task 8 does not import later verifier files, so there is no compile/runtime cycle. raw operators、ordinary Goose CLI、B02/B11 and caller-built structs cannot instantiate the values. B11 authoritative executor constructs the dedicated provider with `goose.WithDisableGlobalRegistry(true)` and `goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration())`；global `AddMigration` remains forbidden, so ordinary CLI never sees 00007.

`sqlc.yaml` must not rely on directory recursion. Change its one PostgreSQL unit's `schema` to the ordered list `[db/migrations, db/migrations/assets/nodecontrol_authority_v7_up.sql]`: the top-level directory supplies 00001–00006 and the literal Up asset supplies the 26 additive relations used by v7 queries. The Down asset is never a sqlc schema input. This is generator metadata only；it neither registers 00007 with Goose nor makes the asset directly executable. `TestSQLCV7SchemaInputOrder` strict-loads the config, requires that exact two-member order/no Down asset, runs pinned sqlc over v7 queries, and in the same gate proves an ordinary/global Goose provider still cannot see migration 00007.

Extend the B01 integration runner with `authority-v7`: after its ordinary base migration, it launches only `TestPrepareC12AuthorityV7Database` from the first-line integration-tagged `internal/testinfra/c12_dependencies_integration_test.go`, passing a fresh run nonce in process environment. That helper uses the Task 8 integration-only sealed fixture factory and a provider-scoped `NodeControlAuthorityV7Migration()` to install disposable 00007, records the exact 51-table/schema/run marker, and exits；the runner clears the init capability variables before launching the requested test packages. The helper skips/denies ordinary invocation, rejects production kind/non-owned database/wrong nonce, and is absent from non-integration builds. No normal command, global Goose registry or raw SQL path gains 00007.

Also prebuild the infrastructure-only `authority-v7-pitr` profile consumed later by B03 integration groups. Its primary is an isolated `authority-v7` database configured with logical WAL, a least-privilege TLS read-only observer/slot, run-owned archive/base-backup volumes, and exact COMMIT/COMMIT PREPARED end-LSN visibility. `internal/testinfra/c12_authority_pitr_integration.go` exposes only a nonce- and descriptor-bound test state machine (`CreateBaseBackup`, `CrashPrimary`, `RestoreAtCut`, `PromoteCandidate`, `InspectTimeline`)；it accepts no arbitrary Docker args, host path, image, port or cleanup target. The runner creates a repository-external ACL-restricted append-only ownership WAL and passes its path plus an HMAC key only in the child environment. Before every helper-created volume/candidate it durably appends a canonical HMAC-authenticated intent containing the preallocated exact name/nonce/labels；after `docker create` it appends the verified actual ID/port. A helper crash between intent/actual is recovered by inspecting only that precommitted exact name and run label, then appending the actual record or proving absence. The runner's `finally` verifies the complete WAL chain and exact-covers/removes the primary, every recorded or intent-recovered candidate and its run-owned volumes/directories；it never enumerates or cleans by broad prefix. Kill tests at intent-before-create、create-before-actual、actual-before-return and cleanup-before/after-remove prove no leaked or foreign resource. This profile proves real WAL/backup/restore/promotion mechanics while B03 uses a test-only external archive/provider adapter；only P09 Tasks 6B/8 may call the production provider/PITR gate authoritative. Focused B03 WAL/PITR commands and their manifest groups must select `authority-v7-pitr`, never the ordinary shared profile.

- [ ] **Step 1: Write RED registered-migration and exact-51 catalog tests**

Add `TestAuthorityV7BoundaryDTOsCompileAndRemainClosed`, `TestNodeControlV7CatalogHasExact51Tables`, `TestNodeControlV7RegisteredMigrationOnly`, `TestNodeControlV7ProviderScopedMigration`, `TestSQLCV7SchemaInputOrder`, `TestNodeControlV7SealedGrantTypes`, `TestNodeControlV7OpaqueCrossPackageUse`, `TestNodeControlV7ProductionBuildOmitsFixtureFactories`, `TestNodeInventoryV7FreshImportShape`, `TestNodeOperatorV7UnauthorizedIsOutputOnly`, `TestNodeControlV7RoleAndGuardRegistry`, `TestNodeControlV7DownRegistry`, `TestAuthorityEffectProofColumnRegistry`, and `TestStoredFencePersistedOutcomeProjection`. Opaque tests prove copies share one consume right, views are defensive/non-secret, a second grant/authorization/repository consume fails, malformed zero values fail, and the integration-tag factory is importable only under `-tags=integration`. Provider tests must prove only `goose.NewProvider(..., goose.WithDisableGlobalRegistry(true), goose.WithGoMigrations(migrations.NodeControlAuthorityV7Migration()))` sees 00007；a provider without the option and the global registry see no 00007。The proof-column registry exact-covers the 13 supported kinds、eight tables、ten operation-group prefixes and typed activation-input owner columns, rejects a duplicate/missing/extra kind or generic outcome table, and exercises legacy-null、prepared-only、terminal none-time、terminal rollback-resistant and malformed partial groups. Repository tests mutate every stored JCS/digest/time/reason/identity/Receipt field and prove one locked defensive projection or fail-closed. The sqlc test requires the exact base-directory+Up-asset schema list, rejects Down/asset-directory recursion, and proves every v7 query resolves. The catalog test retains all 25 Task 4 entries and requires these exact 26 additional table names, no aliases and no 27th v7 table:

```text
control_plane_authority_protocol_migration_latches
control_plane_authority_protocol_downgrade_authorizations
control_plane_authority_protocol_upgrade_intents
control_plane_authority_runtime_registration_results
control_plane_authority_runtime_rebind_results
control_plane_authority_protocol_upgrade_attempts
control_plane_authority_epoch_transition_intents
control_plane_authority_epoch_transition_applications
control_plane_authority_epoch_transition_resolutions
control_plane_authority_epoch_transition_cancellations
control_plane_authority_epoch_transition_terminal_applications
control_plane_authority_epoch_transition_recovery_intents
control_plane_authority_epoch_transition_recovery_prefix_decisions
control_plane_authority_epoch_transition_recovery_applications
control_plane_authority_legacy_database_source_retirements
control_plane_authority_indeterminate_source_seals
control_plane_authority_legacy_source_seals
control_plane_authority_fresh_restore_requirements
control_plane_authority_protocol_activations
control_plane_authority_protocol_activation_completions
control_plane_authority_protocol_activation_releases
control_plane_authority_staging_import_capabilities
control_plane_authority_staging_import_capability_revocation_applications
control_plane_authority_staging_import_capability_recovery_intents
control_plane_authority_staging_import_capability_recovery_applications
control_plane_authority_fresh_restore_import_applications
```

Run: `go test ./internal/nodecontrol/contracts ./internal/store ./internal/nodecontrol/authority -run '^(TestAuthorityV7BoundaryDTOsCompileAndRemainClosed|TestNodeControlV7CatalogHasExact51Tables|TestNodeControlV7RegisteredMigrationOnly|TestNodeControlV7ProviderScopedMigration|TestSQLCV7SchemaInputOrder|TestNodeControlV7SealedGrantTypes|TestNodeControlV7OpaqueCrossPackageUse|TestNodeControlV7ProductionBuildOmitsFixtureFactories|TestNodeInventoryV7FreshImportShape|TestNodeOperatorV7UnauthorizedIsOutputOnly|TestNodeControlV7RoleAndGuardRegistry|TestNodeControlV7DownRegistry|TestAuthorityEffectProofColumnRegistry|TestStoredFencePersistedOutcomeProjection)$' -count=1`

Expected: FAIL because only the base-25 manifest exists and no registered 00007 package/assets are present.

- [ ] **Step 2: Add the exact 00006 parser annotations and registered Go shell**

Wrap each existing `LANGUAGE plpgsql` function in exactly one Goose `StatementBegin`/`StatementEnd` pair without changing its body or catalog semantics; leave `LANGUAGE sql` functions unwrapped. Implement `00007_nodecontrol_authority_abort_serving.go` in package `migrations` as `goose.NewGoMigration(7, &goose.GoFunc{RunTx: up}, &goose.GoFunc{RunTx: down})` over package-adjacent embedded literal assets and return it from `migrations.NodeControlAuthorityV7Migration`；do not duplicate the factory under `internal/nodecontrol/authority`, mutate a constructed provider or touch the global registry. The callback rejects missing typed context, ordinary CLI execution, expired grants, a second consume and unknown `installation_kind`, calls only the exported authority consume APIs to obtain defensive normalized facts, and never reflects/unmarshals opaque internals. Production Up atomically inserts the down-locked migration latch and production upgrade intent, while disposable fixtures insert only the fixture latch. It receives Goose's exact `*sql.Tx`, never opens a second connection or calls provider/attestor. Down alone may invoke the exact `DownAuthorizer` interface once at the frozen §9.3 point；Up and every other callback path reject a non-nil/attempted external dependency.

- [ ] **Step 3: Implement the profile-aware fence amendment and all 26 tables**

Extend the base fence with `authority_protocol_profile`, `abort_claimed_at`, and `protocol_activation_id`. Preserve all legacy_v6 R0/R1/C/A0/A1 rows byte-for-byte: the abort claim pair is required only for claim_v1, legacy aborted rows remain legal, and no default may make an old binary appear claim-v1. Add the 26 tables exactly as v7 §5.1 defines; encode every canonical body field as a literal column/check/FK/unique/INSERT-only trigger in `nodecontrol.v1.yaml`. The five recovery tables enforce INSERT-only, exact intent/application FKs, non-null `recovery_prefix_key_digest` UNIQUE, `count=0 <=> tail=null`, and the closed staging five-table matrix. Gap/PostRecovery/timeline attestations remain signed envelopes stored in result/evidence columns and do not become extra tables.

Without adding a generic outcome table or a 52nd table, add one exact prefixed proof group for each authority operation group on these eight existing domain owners：`node_enrollment_grants` (`create_` and `claim_` groups), `node_certificate_issuances` (`activation_`), `node_certificates` (`revoke_`), `node_state_transitions` (one row-level group with closed `authority_effect_kind` for `identity_epoch_advance|operator_transition`), `node_security_incidents` (`open_` and `resolve_`), `node_resource_envelopes` (`activation_`), `node_state_signing_intents` (`activation_` for desired/recovery), and `node_root_metadata_publish_intents` (`activation_` for root/metadata). Every group has the exact suffixes `authority_effect_commitment_jcs`、`authority_effect_commitment_digest`、`authority_provider_head_jcs`、`authority_provider_head_digest`、`authority_checkpoint_anchor_jcs`、`authority_checkpoint_anchor_digest`、`authority_effect_reason`、`authority_attestation_expires_at`、`authority_activation_deadline`、`authority_expected_provider_identity_digest`、`authority_activation_evidence_jcs`、`authority_activation_evidence_digest`、`authority_effect_resolution_jcs` and `authority_effect_resolution_digest`；all JCS values are `bytea` length `1..4096`, every present digest is exactly 32 bytes, and each JCS/digest pair is all-or-none. The full Receipt remains on the exact fence and is returned in the same locked projection；the real typed activation-input preimage remains in the owner row and is not copied into a generic blob.

Freeze the lifecycle exactly：legacy_v6 rows have the entire new group null；a claim_v1 prepared/committed domain row has exact commitment JCS+digest and all later proof fields null；terminalization changes the remaining group from all-null to exact-once complete in the same transaction as domain disposition/pointer、fence visibility、audit and outbox. Checkpoint JCS/digest is either both absent for `none` or both present；none-time rows require all three external bound columns (attestation expiry、activation deadline、expected provider identity) null, while rollback-resistant rows require all three present and a 32-byte identity digest. `node_state_transitions` gains a unique non-null authority operation key and explicit `applied|not_applied` disposition so identity/operator tombstones are representable without pretending a state changed. Each `node_security_incidents` open/resolve group has an independently unique operation ID and exact-one subtype cardinality across all security-fault/trust-conflict paths. `node_resource_envelopes` keeps package/policy columns immutable but permits exactly one guarded null→terminal proof-group transition；it cannot rewrite the envelope or pointer. Grant create/claim and incident open/resolve groups cannot cross-fill or reuse one another.

The v7 Up asset also replaces the base `node_inventory` identity/quarantine checks with one additive exact exception. `identity_state=unauthorized` is legal only for a row created through the verified fresh-import security-definer path and only with `operator_state=disabled`、`security_state=quarantined`、`identity_epoch=0`、`lineage_id=NULL`、`resume_operator_state=NULL`、no pending transition、no active desired/recovery/root/metadata pointer、empty authority-anchor group、next desired/recovery generation exactly 1 and the verified imported topology versions. Every ordinary/legacy path, any mixed field, nonzero epoch/lineage, saved resume, active/pending pointer or direct role write rejects. The v7 Down asset restores the exact base enum/check text after proving the disposable fixture contains no such rows. Add `unauthorized` to the operator API's output `IdentityStateV1` and regenerated `NodeV1`, but no create/update/action request accepts it or allows a caller to select identity state.

- [ ] **Step 4: Add NOLOGIN ACLs, source-freeze/barrier guards, and canonical lock order**

Create only `nodecontrol_upgrade_executor`, `nodecontrol_migration_downgrader`, and `nodecontrol_staging_importer` as NOLOGIN roles. Revoke direct table/function execution from ordinary runtime, operator and Goose roles. Catalog-test the fixed writer/parent/outcome/forbidden registries, all 19 authority-bearing domain paths, credential/serving/pointer/pending/recovery tables, claim-v1 activation/completion/release barrier, staging-import exception and source-seal freeze. Freeze the exact security-definer `begin_staging_import` signature and generated params: the Go adapter first uses `ViewVerifiedFreshRestoreImportAdmission` only to supply the B02 projection builder, then its one consuming repository call expands normalized fields from the same opaque admission plus the recomputed projection. The function rechecks held exclusion lease、acquisition/current Head、clock、route、identity/lineage/incarnation/rebind/runtime、the exact unauthorized/quarantined/disabled zero-state shape and `0/1/0/0/0`, then atomically writes manifest objects and `FreshRestoreImportApplicationV1` or zero rows. `FreshRestoreImportRepository` is the only Go adapter；B02 cannot call sqlc/store directly. Every restricted function rechecks exact role, transaction, identity, activation, runtime/incarnation chain and preimage; database functions do not implement JCS or public-key verification. All database writers acquire the v7 canonical lock order and issue zero external calls while locked except the sole bounded `authority_protocol_downgrade_authorizer` call in registered Down after pristine `1/0` capture and before authorization insertion；static/runtime tests reject that interface in every other path.

The fixed writer/outcome registry also enumerates every allowed proof-group transition by table、prefix、effect kind and caller role. Direct SQL、wrong prefix/kind、second terminal write、partial JCS/digest group、over-4096 canonical bytes、checkpoint pair mismatch、none-time external-bound columns、rollback-resistant missing-bound columns or a Receipt/fence mismatch fails before any pointer/disposition/audit/outbox visibility. Repository integration tests lock the fence and real domain row in canonical order and prove the returned `PersistedAuthorityEffectOutcome` is byte-defensive and snapshot-consistent；it never issues a Provider/current-Head lookup.

- [ ] **Step 5: Keep Abort/serving claim semantics green under all five legacy shapes**

Implement `LockAuthorityFence`, `ClaimAuthorityAbort`, `BindAuthorityFenceEffect`, `ActivateCommittedAuthorityFence`, and `AbortAuthorityFence` with profile-aware predicates. Add real v6 catalog fixtures R0/R1/C/A0/A1；prove Up only classifies them `legacy_v6`, leaves every new commitment/Head/checkpoint/reason/time/identity/evidence/resolution column in all ten operation groups null, returns no persisted outcome, accepts legacy aborted rows, and makes every source-sealed shape read-only. Claim-first versus writer-first remains a two-connection barrier with one semantic winner and no provider call inside repository tests.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/nodecontrol/authority|./internal/store' -Run '^(TestNodeControlV7LegacyShapes|TestPostgresRepositoryAbortClaim|TestFenceFirstWriterAbortRace|TestAuthorityEffectGuards)$' -Timeout 5m`

Expected: PASS for R0/R1/C/A0/A1, exact retry, direct INSERT/UPDATE/DELETE bypass rejection, and no legacy row rewrite.

- [ ] **Step 6: Implement and RED/GREEN the three-stage disposable-only Down**

In the registered Down callback lock the fixed table set `ACCESS EXCLUSIVE`, consume a sealed Down grant proving a trusted complete permanent provider/member retirement set and `installation_kind=disposable_fixture`, then execute the exact states. Inside Goose's one owned `*sql.Tx`, obtain the actual database transaction identity and construct/store the exact `PristineDowngradeInventoryV1` at latch/auth=`1/0`. Invoke only the injected B11 `authority_protocol_downgrade_authorizer` once with that actual txid、transaction nonce and exact inventory/retirement/catalog/anchor preimages；B01 verifies its role/policy/root/signature and two-minute window into the sealed authorization, then the callback consumes the sealed facts and inserts the exact authorization to reach `1/1`. Atomically consume authorization+latch to reach `0/0`, and only then run reverse-order DDL plus Goose version update in that same transaction. No provider、attestor、other signer or second authorizer call is permitted while the transaction is open. Production, any nonzero fixed-forbidden row, missing provider retirement, extra latch/auth, cross-tx/wrong nonce, changed signed preimage, second opaque consume or direct raw Down fails.

Any signer/verification/SQL/DDL error rolls back the database, but an opaque grant consumed before that rollback remains consumed in process memory. Therefore retry first performs signed Inspect/catalog/version/latch checks outside a transaction；only after it proves the prior transaction did not commit may B11 re-run the sole production verifier over the same permanent retirement/preflight evidence to mint a fresh one-use `VerifiedAuthorityV7DownGrant`, open a new Goose transaction with a new actual txid/nonce, and obtain a fresh txid-bound Down authorization. Commit uncertainty always Inspect-first；a committed transaction never re-verifies/reissues either value. Apply the same rule to Up: rollback/known-not-committed may re-verify the unchanged upgrade evidence into a fresh `VerifiedAuthorityV7UpGrant`, while uncertain/committed Up is inspected and never replayed blindly. Add consume-before-SQL-rollback、consume-before-DDL-rollback、Commit-response-loss and copied-grant tests for both directions.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/store' -Run '^(TestNodeControlV7DownThreeStageMatrix|TestNodeControlV7DownCrashSeams|TestNodeControlV7ProductionDownRejected)$' -Timeout 5m`

Expected: PASS for only the fully retired pristine fixture and for every rollback/response-loss seam.

- [ ] **Step 7: Run the authoritative registered-Go catalog roundtrip**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/store' -Run '^(TestNodeControlV7UpFromR0R1CA0A1|TestNodeControlV7CatalogRoundTrip|TestNodeControlV7ACLAndGuardCatalog)$' -Timeout 8m`

Expected: PASS with normalized 51-table catalog equality after authorized fixture `up -> down-to 5 -> up`, including v7 unauthorized-shape check installation and exact base identity/quarantine check restoration；ordinary Goose CLI、raw SQL and caller-selected unauthorized state remain rejected.

- [ ] **Step 8: Verify generation and commit only the registered migration boundary**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add api/openapi/node-operator-api.v1.yaml gen/go/talenro/nodeoperator/v1/server.gen.go sqlc.yaml db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- api/openapi/node-operator-api.v1.yaml gen/go/talenro/nodeoperator/v1/server.gen.go sqlc.yaml db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- api/openapi/node-operator-api.v1.yaml gen/go/talenro/nodeoperator/v1 sqlc.yaml db/queries/nodecontrol_authority.sql internal/store`

Run: `go test ./internal/nodecontrol/contracts ./internal/store ./internal/nodecontrol/authority -count=1`

Run in the current PowerShell process:

```powershell
$expectedBuildTags = @{
    'internal/testinfra/c12_authority_pitr_integration.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_opaque_integration_fixture.go' = '//go:build integration'
}
$badBuildTags = @($expectedBuildTags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
})
if ($badBuildTags.Count -ne 0) { throw 'B01 integration capability missing exact build tag' }
```

Expected: second-generation diff and untracked output are empty before tests/commit.

```bash
git add api/openapi/node-operator-api.v1.yaml gen/go/talenro/nodeoperator/v1/server.gen.go internal/nodecontrol/contracts/openapi_contract_test.go internal/nodecontrol/contracts/authority_v7_boundary.go internal/nodecontrol/contracts/authority_v7_boundary_test.go db/migrations/00006_nodecontrol.sql db/migrations/00007_nodecontrol_authority_abort_serving.go db/migrations/assets/nodecontrol_authority_v7_up.sql db/migrations/assets/nodecontrol_authority_v7_down.sql sqlc.yaml db/schema/nodecontrol.v1.yaml scripts/run-c12-integration.ps1 internal/testinfra/c12_dependencies_integration_test.go internal/testinfra/c12_authority_pitr_integration.go db/queries/nodecontrol_authority.sql internal/store/nodecontrol_authority.sql.go internal/store/models.go internal/store/querier.go internal/store/nodecontrol_v7_catalog_test.go internal/store/nodecontrol_v7_migration_integration_test.go internal/store/nodecontrol_v7_acl_integration_test.go internal/store/nodecontrol_v7_down_integration_test.go internal/nodecontrol/authority/repository.go internal/nodecontrol/authority/postgres_repository.go internal/nodecontrol/authority/v7_migration_grant.go internal/nodecontrol/authority/v7_staging_import.go internal/nodecontrol/authority/v7_down_authorization.go internal/nodecontrol/authority/v7_opaque_integration_fixture.go internal/nodecontrol/authority/v7_migration_grant_test.go internal/nodecontrol/authority/v7_staging_import_test.go internal/nodecontrol/authority/v7_down_authorization_test.go internal/nodecontrol/authority/postgres_repository_integration_test.go internal/nodecontrol/authority/v7_staging_import_repository_integration_test.go
git commit -m "feat(nodecontrol): install registered authority v7 catalog"
```

### Task 9: Make Coordinator Abort and provider commitment atomically recoverable

**Files:**
- Modify: `internal/nodecontrol/authority/coordinator.go`
- Modify: `internal/nodecontrol/authority/coordinator_test.go`
- Modify: `internal/nodecontrol/authority/authority_crash_integration_test.go`
- Modify: `internal/nodecontrol/authority/authority_pitr_integration_test.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`

**Interfaces:**
- Consumes: provider/repository Tasks 5–6, Task 7 sealed dispatcher/canonical proof APIs, and Task 8 claim-aware `StoredFence.PersistedOutcome` from one locked snapshot.
- Produces: `func NewCoordinator(Provider, Repository, EffectDispatcher) (*Coordinator, error)` plus `Reserve`, `Finalize`, explicit-reason `Abort`, `Recover`, DBTX-bound and public `CheckReady`, and read-only `CommittedNodeCheckpoint(context.Context, contracts.Digest) (NodeCheckpoint, error)`；no pointer-to-interface、caller-supplied DB point/evidence or raw-provider escape. Construction requires and stores only the exact non-nil private concrete dynamic type returned by `NewEffectDispatcher`；nil、typed-nil、embedded、wrapped、proxy or alternate implementations return exact `ErrInvalidArgument` before any dependency method call.

- [ ] **Step 1: Write the RED multi-process Abort/Finalize/activation crash matrix**

Cover the full ownership matrix：fence lock and resolve share one DBTX；claim commit precedes provider Abort；capture starts only after an exact committed Receipt and after every prior transaction/lock is released；the observable order is exact `Begin → registered material/trusted-time → same-Provider Head/checkpoint → Complete/New → ValidateActivationDecisionEvidence`；the constructor rejects every non-exact dispatcher dynamic type before calls；Receipt database point equals the locked fence binding；rollback-resistant proof consumes after all writes and immediately before one Commit while none-time never consumes；every §8.5 preimage commits atomically；Commit response loss reloads and context-validates stored bytes before any recapture；parsed recovery calls the zero-write/external `ValidatePersistedAuthorityEffect` in the same locked DBTX；and two PostgreSQL connections racing fence-first effect versus Abort have one semantic winner. Include claim/Abort response loss、PITR-lost claim、provider reserved with no guessed reason、provider committed before evidence、capture/activation failures、token burn failure、Commit invocation/response loss and exact terminal replay.

Run: `go test ./internal/nodecontrol/authority -run 'TestCoordinatorDispatcherConstruction|TestCoordinatorAbortLinearization|TestCoordinatorEvidenceCaptureOrder|TestCoordinatorAtomicActivationCrashMatrix|TestCoordinatorPersistedOutcomeRecovery' -count=1`

Expected: FAIL against the pre-amendment Coordinator.

- [ ] **Step 2: Implement claim-before-provider Abort**

Inspect provider outside the transaction, then begin explicit PostgreSQL `READ COMMITTED`, call `Repository.Lock` to obtain the exact locked `Reservation expected`, and invoke only the sealed dispatcher-facing `ResolveAuthorityEffectForUpdate(ctx, tx, expected)` with that same DBTX. The dispatcher privately constructs all 13 registered queries and validates echoes/cardinality；Task 9 cannot select a probe kind. Claim only exact absent/unbound/unclaimed, commit and release locks before provider Abort, and terminalize in a new short transaction. Unknown/fixed-unsupported kinds conflict before any handler call；cancellation/SQL/resolver failure rolls back an uncommitted claim. Recovery may reconstruct a provider-aborted claim only from the receipt's exact reason under the same lock/resolver predicates；provider-reserved recovery never guesses a reason.

- [ ] **Step 3: Implement provider Finalize plus one activation transaction**

After binding and provider Finalize/Inspect returns the exact committed Receipt outside locks, call `BeginActivationEvidenceCapture` immediately before the first evidence-related observation. Direct-route registered `CaptureActivationDecisionMaterial` (including its authenticated trusted-time/floor facts), call `Head` and the approved node/global checkpoint observation on the same Coordinator-owned `Provider` instance, construct `AuthorityProviderHeadSnapshot`/optional `AuthorityCheckpointAnchor`, then use `Complete` for rollback-resistant material or `NewActivationDecisionEvidence` for none-time material and context-validate to the branded proof. No handler receives Provider or constructs final evidence.

Open one short activation transaction, lock/revalidate the exact fence/commitment/domain row, compare every Receipt field and its complete database point with the locked fence, call `ActivateCommitted`, and direct-route `ActivateAuthorityEffect` with the fresh proof in the same DBTX. The handler recomputes the typed activation-input digest and atomically stores its terminal disposition/pointer plus every proof preimage；the Coordinator completes fence visibility、audit/outbox and `PersistedOutcome`. For rollback-resistant proof only, call the package-private consume after every write and immediately invoke exactly one Commit attempt；the first canceled/expired/mismatched consume is already spent and forces rollback/recapture. None-time proof has no token and proceeds from completed writes directly to that one Commit without calling consume. Any origin、binding、missing handler/preimage、mutable-only fact、timeout or SQL failure rolls back all visibility together.

On Commit uncertainty, begin one locked recovery snapshot and load the exact `StoredFence.PersistedOutcome` plus domain row. Reparse/recanonicalize commitment、stored Head、optional checkpoint、evidence and resolution；reconstruct the original `ActivationDecisionEvidenceInput` from stored reason/time/expected identity and the exact fence Receipt；call `ValidateActivationDecisionEvidence` and `ValidateAuthorityEffectResolution`, then direct-route `ValidatePersistedAuthorityEffect` with the tokenless parsed proof in that same DBTX. It must prove the one exact terminal domain group and typed activation-input binding with zero writes/provider/time/signer/issuer calls. Fresh proof at recovery or parsed proof at activation conflicts before the handler. Never replace the stored Head with current Head；only a proved-absent outcome after resolving commit uncertainty may recapture.

- [ ] **Step 4: Run GREEN crash, admission, and race tests**

Run: `go test ./internal/nodecontrol/authority -run 'TestCoordinatorDispatcherConstruction|TestCoordinatorAbortLinearization|TestCoordinatorEvidenceCaptureOrder|TestCoordinatorAtomicActivationCrashMatrix|TestCoordinatorPersistedOutcomeRecovery|TestActivationAdmission' -count=1`

Run: `go test -race ./internal/nodecontrol/authority -run 'TestCoordinatorAbort|TestCoordinatorAtomicActivation' -count=1`

Expected: PASS with one Provider instance, exact call/DBTX ordering, no second provider mutation, no fence/domain/proof split, no none-time consume, no token reuse and no recovery write/external call.

- [ ] **Step 5: Rebuild readiness over exact claims/outcomes**

Match latest reserved/committed anchors and every pending `StoredFence`; a pending claim、prepared effect、missing handler、unsupported kind、provider/DB mismatch、bound-aborted record or any missing/noncanonical/mismatched commitment/Head/checkpoint/evidence/resolution JCS/digest、Receipt database point、reason/time/expected identity、typed activation-input or proof origin/binding keeps readiness false with a finite value-free reason. Terminal recovery uses the same locked snapshot and zero-write persisted validator；it never substitutes mutable current Head. DBTX-bound readiness reuses a caller-supplied connection for DB reads, while all Provider calls remain outside SQL statements/locks. Public probe name remains `authority`.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority|./internal/readiness' -Run '^(TestAuthorityReadiness|TestAuthorityProbe)$' -Timeout 5m`

Expected: PASS with deterministic reason precedence and no values in public errors.

- [ ] **Step 6: Commit the Coordinator protocol**

```bash
git add internal/nodecontrol/authority/coordinator.go internal/nodecontrol/authority/coordinator_test.go internal/nodecontrol/authority/authority_crash_integration_test.go internal/nodecontrol/authority/authority_pitr_integration_test.go internal/readiness/checker.go internal/readiness/checker_test.go
git commit -m "feat(nodecontrol): recover authority effects atomically"
```

### Task 10: Add bound guarded serving and real certificate/desired PITR proof

**Files:**
- Create: `db/queries/nodecontrol_serving.sql`
- Create: `internal/store/nodecontrol_serving.sql.go`
- Modify generated: `internal/store/models.go`
- Modify generated: `internal/store/querier.go`
- Create: `internal/nodecontrol/serving/types.go`
- Create: `internal/nodecontrol/serving/repository.go`
- Create: `internal/nodecontrol/serving/postgres_repository.go`
- Test: `internal/nodecontrol/serving/repository_test.go`
- Test: `internal/nodecontrol/serving/postgres_repository_integration_test.go`
- Modify: `internal/nodecontrol/authority/authority_pitr_integration_test.go`

**Interfaces:**
- Consumes: Task 9 Coordinator and its exact `PostgresRepository`, generated sqlc store, committed certificate/desired fixtures.
- Produces: suite-index `BoundAuthorityReadSource`, `Reader.GetAuthorizedNodeCertificate`, and `Reader.GetActiveDesiredState`; B02/B03 consume these facts later, but B01 does not wire HTTP/listeners or `cmd/control-api`.

```go
type BoundAuthorityReadSource interface {
	WithConsistentReadyRead(context.Context, func(context.Context, store.DBTX) error) error
}

func (r *Reader) GetAuthorizedNodeCertificate(context.Context, CertificateLookup) (CertificateFacts, error)
func (r *Reader) GetActiveDesiredState(context.Context, uuid.UUID) (DesiredStateFacts, error)
```

For an agent-listener lookup, `CertificateLookup` additionally contains one constructor-validated fixed server-trust selector `{purpose=agent_server,listener_kind=agent,trust_domain=<configured exact domain>}`. `CertificateFacts` includes the matching finalized `ServerCABundleHighWater` as a defensive `contracts.VersionedDigest` plus its exact purpose/listener/domain and authority epoch；the group is complete/nonzero or the whole read fails. The certificate and trust-high-water rows are read inside the same `WithConsistentReadyRead` callback and both exact fences must be committed-active, so no caller can splice a trust row from another listener/domain or a second transaction. Non-agent lookup construction cannot select another trust purpose through free strings.

`DesiredStateFacts` is a complete defensive-copy serving projection, not merely desired bytes: it contains at most 32 canonical finalized root-rotation envelopes ending at the exact active root pointer, the exact active full metadata envelope, and the exact active signed desired envelope, with each artifact's publish/signing ID、version/generation、digest and authority tuple. It contains no repository/DBTX handle and exposes no pending intent or mutable backing slice. B03 may use the poll request high-waters only to choose a continuous suffix or decide unchanged/conflict from these facts；its response builder performs zero database reads.

- [ ] **Step 1: Write RED same-connection/read-error-priority tests**

Use a recording bound source to assert pre-check, callback and post-check order; callback NotFound/corrupt still triggers post-check; cancellation wins over authority, authority wins over domain error; callback result/error remains private until post-check exact-head equality. A fake pool/repository mismatch and forced connection replacement return `ErrAuthorityUnavailable` with zero material.

Run: `go test ./internal/nodecontrol/serving -run 'TestGuardedReadOrder|TestGuardedReadErrorPriority' -count=1`

Expected: FAIL because the serving package does not exist.

- [ ] **Step 2: Add exact sqlc queries and the bound reader**

Certificate query matches exact node/issuer/unsigned serial/leaf DER digest/public-key digest/current identity epoch/lineage, allowed status, and exact committed-active `certificate_activate` fence；for the fixed agent selector the same callback also joins exactly one committed-active `control_plane_trust_bundle_high_waters` row and rejects absent/duplicate/purpose/listener/domain/epoch/fence mismatch. The one combined desired-serving query starts only from `node_inventory.active_desired_generation` and the same row's active root/metadata pointers；it joins the exact desired row, active signing intent and exact committed-active `desired_activate` fence, the exact active root/metadata publish intents and their committed-active `root_publish`/`metadata_publish` fences, plus a gap-free committed root-rotation chain ending at that root pointer and capped at 32. A missing/gapped/oversized chain, pointer/envelope mismatch, pending/partial/superseded current artifact or any fence mismatch returns no facts/fails closed；there is no `MAX`, latest-row or orphan fallback. The production source is created only by the Coordinator that owns the same `PostgresRepository`, uses one physical connection for both DB readiness passes and the query, and clears every certificate/trust/root/metadata/desired temporary byte on failure. The tagged serving fixture's exact top-level name is `TestServingQueryExactCommittedProjection`.

- [ ] **Step 3: Generate sqlc and run focused GREEN tests**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_serving.sql internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_serving.sql internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_serving.sql internal/store`

Run: `go test ./internal/nodecontrol/serving ./internal/store -run 'TestGuardedRead|TestServingQuery' -count=1`

Expected: PASS and both second-generation drift/untracked checks are empty against the exact staged first-generation baseline.

- [ ] **Step 4: Write the real-fixture PITR RED/GREEN integration test**

Materialize inventory, issuer dependencies, certificate issuance/certificate row, a multi-step root chain, metadata/signing intents, desired row, every exact fence and all active pointers. First prove both readers return exact defensive fixture facts in root-chain→metadata→desired order；independently replace each active root/metadata artifact or fence with pending/partial/gapped/mismatched data and require zero serving bytes. Restore a revoke/disable-preceding backup while the provider remains at the higher head; raw SQL must still see old certificate/desired rows, but both readers return exact `ErrAuthorityUnavailable` with empty DER/root/metadata/payload/derived authorization. Reconcile to the exact ready head and prove the control fixture reads again.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/authority|./internal/nodecontrol/serving' -Run '^TestPITRBeforeRevocationFailsClosed$' -Timeout 5m`

Expected: PASS for baseline, restored fail-closed and reconciled control phases.

- [ ] **Step 5: Run the Abort/serving sub-gate and commit guarded serving**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/serving ./internal/readiness ./internal/store -count=1 -timeout 5m`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority|./internal/readiness|./internal/nodecontrol/serving' -Run '^(TestCoordinatorPostgresCrashRecoveryMatrix|TestCoordinatorPostgresAbortDomainSafetyAndIdempotence|TestCoordinatorPostgresCrashResolverFailsClosedOnSQLDeletionOrCorruption|TestAuthorityReadiness|TestAuthorityProbe|TestServingQueryExactCommittedProjection)$' -Timeout 7m`

Run: `go test -race ./internal/nodecontrol/authority ./internal/nodecontrol/serving -count=1 -timeout 7m`

Expected: PASS with pinned-Goose roundtrip, claim/effect races, exact recovery, guarded serving and real-fixture PITR coverage.

This sub-gate covers only the base Coordinator、Abort and guarded-serving boundary implemented through Task 10. It does not close Batch 01 and must not be reported as the Batch 01 exit；Tasks 11–18 remain required, and only Task 18 Step 3 is the complete B01 verification boundary.

```bash
git add db/queries/nodecontrol_serving.sql internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/serving/types.go internal/nodecontrol/serving/repository.go internal/nodecontrol/serving/postgres_repository.go internal/nodecontrol/serving/repository_test.go internal/nodecontrol/serving/postgres_repository_integration_test.go internal/nodecontrol/authority/authority_pitr_integration_test.go
git commit -m "feat(nodecontrol): guard certificate and desired reads"
```

### Task 11: Freeze v7 canonical bodies, signature envelopes, evidence bundles, and golden registry

**Files:**
- Create: `internal/nodecontrol/contracts/canonical.go`
- Create: `internal/nodecontrol/contracts/signature_envelope.go`
- Create: `internal/nodecontrol/contracts/evidence_bundle.go`
- Create: `internal/nodecontrol/contracts/schema_registry.go`
- Test: `internal/nodecontrol/contracts/canonical_v7_test.go`
- Test: `internal/nodecontrol/contracts/signature_envelope_test.go`
- Test: `internal/nodecontrol/contracts/evidence_bundle_test.go`
- Test: `internal/nodecontrol/contracts/canonical_v7_fuzz_test.go`
- Create: `testdata/c12/authority-v7/schema-registry.v1.json`
- Create: `testdata/c12/authority-v7/canonical-envelope-golden.v1.json`

**Interfaces:**
- Consumes: the exact schema/body/role mapping and JCS/envelope rules in v7 §4.1; repository-pinned Ed25519 and ECDSA P-256 verification primitives.
- Produces: one schema registry, `CanonicalBodyDigest`, `ParseCanonicalBody`, `VerifySignatureEnvelope`, and `VerifyCanonicalEvidenceBundle`. Tasks 12–14 register typed bodies through this sole registry; B03/B11 consume these APIs without another canonicalizer.

```go
func CanonicalBodyDigest(schema SchemaID, body any) (Digest, []byte, error)
func ParseCanonicalBody(schema SchemaID, raw []byte, dst any) error
func VerifySignatureEnvelope(ctx context.Context, policy TrustedSignerPolicy, raw []byte) (VerifiedEnvelope, error)
func VerifyCanonicalEvidenceBundle(messageSchema SchemaID, messageDigest Digest, raw []byte, registry SchemaRegistry) (VerifiedEvidenceBundle, error)
```

- [ ] **Step 1: Write RED literal vectors independent of production helpers**

Build `schema-registry.v1.json` from the amendment's fixed mapping, with exact schema, Go type, signer role, algorithm policy and external-vs-DB ownership. Assert one-to-one mapping and reject aliases/case conversion. Include negative rows proving the Abort/serving historical `AuthorityEffectCommitmentV1`, `AuthorityEffectResolutionV1`, and `ActivationDecisionEvidenceV1` are not new v7 registry entries.

Run: `go test ./internal/nodecontrol/contracts -run 'TestAuthorityV7SchemaRegistryLiteral|TestAuthorityV7HistoricalSchemasNotReregistered' -count=1`

Expected: FAIL because the v7 registry and parser do not exist.

- [ ] **Step 2: Implement minimal canonical body and envelope verification**

Use strict RFC 8785 JCS and the exact `SHA-256(ASCII(schema) || 0x00 || JCS(body))` formula. Reject missing/extra/duplicate fields, invalid UTF-8, BOM, JSON numbers where canonical decimal strings are required, noncanonical UUID/time/digest/base64url, unordered/duplicate sets, unknown schema/role/key/policy/root/algorithm, non-64-byte Ed25519, non-fixed-width/high-S P-256 and signature metadata/body mismatch. `VerifiedEnvelope` stores defensive canonical bytes and typed metadata; it has no signing method.

- [ ] **Step 3: Implement the acyclic evidence-bundle verifier**

Require exact `message_schema`, `message_body_digest`, sorted `(body_digest,schema)` items, closed `database_immutable_body|external_signed_envelope` kind and exactly one preimage arm. Recompute every digest/signature, reject missing/extra/duplicate evidence against the per-message registry, and never write bundle digest back into the message body. Add mutation tests for every field and a golden proving message/bundle/envelope have no self-reference.

- [ ] **Step 4: Run cross-language literals, fuzz and package tests**

Run: `go test ./internal/nodecontrol/contracts -run 'TestAuthorityV7Canonical|TestSignatureEnvelope|TestCanonicalEvidenceBundle' -count=1`

Run: `go test ./internal/nodecontrol/contracts -run '^$' -fuzz '^FuzzAuthorityV7Canonical$' -fuzztime=20s -timeout 40s`

Run: `go test ./internal/nodecontrol/contracts -run '^$' -fuzz '^FuzzSignatureEnvelope$' -fuzztime=20s -timeout 40s`

Expected: all golden vectors pass and every one-field mutation is rejected with finite value-free errors.

- [ ] **Step 5: Commit the sole canonical layer**

```bash
git add internal/nodecontrol/contracts/canonical.go internal/nodecontrol/contracts/signature_envelope.go internal/nodecontrol/contracts/evidence_bundle.go internal/nodecontrol/contracts/schema_registry.go internal/nodecontrol/contracts/canonical_v7_test.go internal/nodecontrol/contracts/signature_envelope_test.go internal/nodecontrol/contracts/evidence_bundle_test.go internal/nodecontrol/contracts/canonical_v7_fuzz_test.go testdata/c12/authority-v7/schema-registry.v1.json testdata/c12/authority-v7/canonical-envelope-golden.v1.json
git commit -m "feat(nodecontrol): freeze authority v7 canonical envelopes"
```

### Task 12: Add environment, shutdown, incarnation, lineage, empty/source, and Down verifiers

**Files:**
- Create: `internal/nodecontrol/contracts/authority_v7_source.go`
- Create: `internal/nodecontrol/contracts/authority_v7_down.go`
- Create: `internal/nodecontrol/authority/v7_verifier.go`
- Create: `internal/nodecontrol/authority/v7_migration_grant_verifier.go`
- Test: `internal/nodecontrol/authority/v7_environment_verifier_test.go`
- Test: `internal/nodecontrol/authority/v7_empty_source_verifier_test.go`
- Test: `internal/nodecontrol/authority/v7_down_verifier_test.go`
- Test: `internal/nodecontrol/authority/v7_migration_grant_verifier_test.go`
- Create: `testdata/c12/authority-v7/environment-source-down-golden.v1.json`

**Interfaces:**
- Consumes: Task 11 registry/envelope/evidence verifier and Task 8 fixed 51-table catalog manifest.
- Produces: typed verified values for EnvironmentInventory/AnchorSet, LocalRuntimeIsolation/LegacyRuntimeShutdown(Set), DatabaseIncarnation/runtime lease/timeline lineage/registration result, ProviderNamespaceAbsence/EmptyDatabaseInventory/Epoch evidence, source retirement/seal/exact-cover/archive, and Down retirement/three-stage inventory；plus `VerifyAuthorityV7UpGrant`/`VerifyAuthorityV7DownGrant`, which are the only production factories for Task 8 sealed migration grants, and `VerifyAuthorityProtocolDowngradeAuthorization`, which is the only production factory for Task 8's sealed actual-txid-bound Down authorization. Tasks 12–13 reuse, never redeclare, Task 8's normalized boundary DTOs. B11 obtains the retirement/grant values before opening a DB transaction；only the exact §9.3 authorization is signed and verified inside the registered Down transaction.

- [ ] **Step 1: Write RED exact-field and signer-policy vectors**

For every family above, literal tests enumerate the exact required field set, item ordering, count/nullability, signer role, maximum validity window and identity tuple. Add negatives for inventory truncation/alias namespace, shutdown-set missing member, credential/provider mismatch, clone OID/name mismatch, lineage sibling/gap/65th link/fork-before-anchor, empty proof with `environment_count>1`, source kind/schema splice and Down production/partial-retirement input.

Run: `go test ./internal/nodecontrol/authority -run 'TestV7EnvironmentVerifier|TestV7IncarnationLineageVerifier|TestV7EmptySourceVerifier|TestV7DownVerifier' -count=1`

Expected: FAIL because the typed bodies and verifiers are absent.

- [ ] **Step 2: Implement mechanical environment and incarnation verification**

Implement the exact four-field database identity digest and timeline lineage root/next formulas. Require one live runtime binding generation, current non-superseded lease, complete continuous 1..64 history links, old-holder permanent termination for rebind, exact system/OID/name/incarnation and fork cut not earlier than the signed anchor. Environment inventory/anchor/shutdown sets use exact sorted membership; historical admission objects are checked at recorded transition time, while a new transition requires current unexpired evidence. `VerifyAuthorityV7UpGrant` additionally binds installation/transaction nonce, LocalRuntimeIsolation, production-vs-fixture classification and exact preallocated upgrade intent IDs, returning Task 8's sealed value without exposing a raw constructor.

- [ ] **Step 3: Implement empty/source/Down verified values**

Derive candidate classification only from the fixed catalog plus signed inventory/provider inputs: `empty_adoptable`, `legacy_or_nonempty`, or `indeterminate`. Only singular exact DB/provider absence produces an empty proof. Minimal indeterminate seal cannot export/archive/activate; full source requires membership tombstone, per-member DB/environment/provider retirement, post-seal exact cover and source archive. `VerifyAuthorityV7DownGrant` requires disposable fixture, latest inventory membership retirement, all-provider permanent retirement, exact 51-table catalog digest, fixed-forbidden zero projection and unconsumed `1/0` latch；production/missing/expired evidence cannot construct the sealed preflight grant, and it cannot guess a future txid. After the registered callback obtains the actual txid and stores the pristine inventory, B11 signs the exact v7 §9.3 `AuthorityProtocolDowngradeAuthorizationV1` inside that Goose transaction；`VerifyAuthorityProtocolDowngradeAuthorization` checks actual txid、transaction nonce、inventory/retirement/anchor/catalog bindings、role/policy/root/signature and expiry before constructing the one-use sealed authorization.

- [ ] **Step 4: Run focused tests and commit**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority -run 'TestV7Environment|TestV7Incarnation|TestV7Lineage|TestV7Empty|TestV7Source|TestV7Down' -count=1`

```bash
git add internal/nodecontrol/contracts/authority_v7_source.go internal/nodecontrol/contracts/authority_v7_down.go internal/nodecontrol/authority/v7_verifier.go internal/nodecontrol/authority/v7_migration_grant_verifier.go internal/nodecontrol/authority/v7_environment_verifier_test.go internal/nodecontrol/authority/v7_empty_source_verifier_test.go internal/nodecontrol/authority/v7_down_verifier_test.go internal/nodecontrol/authority/v7_migration_grant_verifier_test.go testdata/c12/authority-v7/environment-source-down-golden.v1.json
git commit -m "feat(nodecontrol): verify authority v7 source and downgrade evidence"
```

### Task 13: Freeze genesis, epoch, lease, rebind, and staging provider contracts

**Files:**
- Create: `internal/nodecontrol/contracts/authority_v7_genesis.go`
- Create: `internal/nodecontrol/contracts/authority_v7_epoch.go`
- Create: `internal/nodecontrol/contracts/authority_v7_staging.go`
- Create: `internal/nodecontrol/authority/v7_deterministic_local_test_provider.go`
- Create: `internal/nodecontrol/authority/v7_staging_admission_verifier.go`
- Test: `internal/nodecontrol/contracts/authority_v7_provider_contract_test.go`
- Test: `internal/nodecontrol/authority/v7_deterministic_local_test_provider_test.go`
- Test: `internal/nodecontrol/authority/v7_staging_admission_verifier_test.go`
- Create: `testdata/c12/authority-v7/provider-genesis-epoch-staging-golden.v1.json`

**Interfaces:**
- Consumes: Tasks 11–12 canonical/verified evidence and Task 8 profile/catalog enums.
- Produces: exact request/response/Inspect/Head schemas for credential policy; incarnation registration/rebind; genesis Prepare/Complete/PrepareRelease/Open; ordinary Reserve/Finalize/Abort/Record; epoch prepare/application/resolution/cancellation/terminal/exact recovery/prefix decision/deferred suffix; runtime and serving lease initial/renew/revoke/descendant/terminal catchup; staging hold/intent/recovery/challenge/Release/Abort; `DatabaseAuthorityRebindGapAttestationV1` and `DatabaseAuthorityPostRecoveryRebindAttestationV1`；and `VerifyFreshRestoreImportAdmission`, the only production factory for Task 8's opaque normalized admission consumed by B02. The deterministic provider is test-only by name and constructor; Plan 03 supplies production implementations.

- [ ] **Step 1: Write RED literal field-set, phase-matrix, and DAG tests**

For every contract in the interface list, add a golden exact body and mutate each field, schema, signer role, request digest, evidence-set member and present/null branch. Assert Provider Head phases from `registered_pending_genesis` through active/staging/recovery, effective and terminal epoch chains, lease purpose, ordinary six-field projection and response-loss Inspect. Assert every request/response DAG excludes a backward digest edge.

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority -run 'TestAuthorityV7ProviderContracts|TestAuthorityV7ProviderPhaseMatrix|TestAuthorityV7ProviderDAG' -count=1`

Expected: FAIL because the contract types and deterministic state machine are absent.

- [ ] **Step 2: Implement exact genesis and lease contracts**

Implement the field sets from v7 §6, deterministic IDs and closed phase/present-null matrices. Preparation consumes exact namespace absence and remains closed; completion/release/open require their predecessor digest and current runtime/incarnation tuple. Runtime lease renewals keep binding/ID/generation and form a strictly continuous descendant chain; serving leases bind activation/genesis root/current identity-lineage/policy/effective+terminal epoch and exact ordinary prefix. No response or lease digest is included in the request that causes it.

- [ ] **Step 3: Implement epoch arbitration and recovery contracts**

Implement intent→optional gap-attested deferred suffix→provider transcript→non-null prefix key decision. Apply and replacement share one decision ID/subject and race one DB prefix; decision body never contains candidate. Replacement requires a fresh gap/challenge/attestation for each candidate, consumes only on successful provider CAS, and allows pre-CAS candidate death without changing the decision. Apply requires same-transaction application+decision, per-candidate fresh proof, first-consumer/history CAS, `pre_catchup|post_catchup_marker_cleared`, immediate result/reconcile, and the exact next-winning PostRecovery or catchup Head fork-cut witness.

- [ ] **Step 4: Implement staging and source/Down provider contract surfaces**

Implement hold/required/abort-only, five-table eight-state matrix, immutable capability/recovery/revocation IDs, generic/staging commit proof fields, Release/Abort terminal rules, source membership/retirement and provider permanent downgrade retirement. `VerifyFreshRestoreImportAdmission`在开写transaction前验证manifest/capability envelope、exact held exclusion lease、acquisition-locked/current Head、current clock/route、target activation/database identity/lineage/incarnation/rebind/runtime、normalized 51-table catalog、pre/expected-post projection与apply ID，并把normalized fields封装进Task 8 unforgeable value。B02只能通过Task 8 `ViewVerifiedFreshRestoreImportAdmission`取得defensive `FreshRestoreImportProjectionInputV1`，而repository对同一opaque value执行唯一consuming write；view不能制造新token或绕过lease/Head/recovery重验。The consumer interface exposes typed calls only; it has no DB handle and tests assert provider call counters remain zero in restricted DB transactions.

- [ ] **Step 5: Run deterministic response-loss/CAS tests and commit**

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority -run 'TestAuthorityV7DeterministicLocalTestProvider|TestAuthorityV7EpochArbitration|TestAuthorityV7StagingMatrix|TestFreshRestoreImportAdmission' -count=1`

```bash
git add internal/nodecontrol/contracts/authority_v7_genesis.go internal/nodecontrol/contracts/authority_v7_epoch.go internal/nodecontrol/contracts/authority_v7_staging.go internal/nodecontrol/authority/v7_deterministic_local_test_provider.go internal/nodecontrol/authority/v7_staging_admission_verifier.go internal/nodecontrol/contracts/authority_v7_provider_contract_test.go internal/nodecontrol/authority/v7_deterministic_local_test_provider_test.go internal/nodecontrol/authority/v7_staging_admission_verifier_test.go testdata/c12/authority-v7/provider-genesis-epoch-staging-golden.v1.json
git commit -m "feat(nodecontrol): freeze authority v7 provider contracts"
```

### Task 14: Freeze generic commit archive, Fence, Inspect, and continuity contracts

**Files:**
- Create: `internal/nodecontrol/contracts/commit_archive.go`
- Create: `internal/nodecontrol/authority/commit_archive_verifier.go`
- Test: `internal/nodecontrol/authority/commit_archive_verifier_test.go`
- Test: `internal/nodecontrol/authority/commit_archive_deterministic_model_test.go`
- Test: `internal/nodecontrol/authority/commit_archive_concurrency_test.go`
- Create: `testdata/c12/authority-v7/commit-archive-golden.v1.json`

**Interfaces:**
- Consumes: Task 11 canonical registry, Task 13 gap/PostRecovery/provider Head/history types and Task 8 fixed v7 row registry.
- Produces: generic related-row-set, purpose/generation evidence, fixed 10-row-kind registry, atomic multi-key Fence request/response, semantic-ID journal reservation plus cross-reason epoch prefix subject-slot/reverse uniqueness, stable-holder once-key/once-row, global primary/related latches and reverse owner, signed Inspect candidate/history fields, and staging exact-existing/missing-suffix evidence-mode verifier. B03 owns the production rollback-resistant archive and WAL decoder; B01 provides only schema, verifier and deterministic model tests.

- [ ] **Step 1: Write RED fixed-registry and literal digest tests**

Add positive vectors for all 10 row kinds and negatives for schema alias, wrong semantic ID field, related cardinality/schema, purpose, sorted key generations and gap/PostRecovery exclusivity. Add literal JCS vectors for Fence reason authorization, semantic journal reservation, epoch prefix subject-slot, stable-holder once-key, response, Inspect request/response, provider-consumption event and staging evidence-set modes.

Run: `go test ./internal/nodecontrol/authority -run 'TestCommitArchiveFixedRegistry|TestCommitArchiveFenceGolden|TestCommitArchiveInspectGolden' -count=1`

Expected: FAIL because `commit_archive.go` and the verifier do not exist.

- [ ] **Step 2: Implement application+decision closure and Inspect selection**

Model global key `(row_schema,semantic_row_id)`, absent/exact/different body latch, application forward related edge, decision reverse-owner uniqueness, primary stream and exact challenge retry. Application proof locks sorted application+decision latches, then edge/reverse owner, then primary stream; apply-as-primary and replacement-as-related fail. Inspect returns `related_only` with zero entry/head, signs the full live candidate seven-field tuple and 0/1 row observation, and mechanically selects current-parent, unconsumed-candidate-lineage or recovery-continuity entries from complete preimages rather than caller choice.

- [ ] **Step 3: Implement Fence reservation and stable once semantics**

The reason registry accepts only pair, replacement, staging generic and staging revocation outcome targets. Semantic reservation writes one append-only record; epoch pair/replacement first CAS `(activation_id,recovery_id,recovery_prefix_key_digest)` to one decision ID with reverse uniqueness, then write reason-specific records atomically. Stable once-key includes journal/context/key set plus stable identity/lineage/registration/binding/runtime ID/generation, but excludes renewable runtime lease, Head, evidence/Inspect digest, archive generation and request/time; the complete request still signs and CAS-revalidates all freshness fields. Fence atomically locks sorted global latches, checks no body/stream/owner, advances all generations, writes one successful once-row and exact response; partial key advancement and orphan rows are impossible.

- [ ] **Step 4: Implement continuity history and staging evidence modes**

History events commit ordinal/order/request/application proof without provider response, carry full response/result/Head preimages in signed Inspect, enforce first-consumer and result-pending latch, and use exact next-winning PostRecovery/catchup-Head/unconsumed-tail reconcile branches. Staging candidate universe derives only from provider journal; exact-existing requires live candidate count=1/digest match and may have null canonical parent, while archived-missing-suffix requires count=0, immediate-parent canonical proof and exact commit LSN strictly after fork cut.

- [ ] **Step 5: Run deterministic lock/crash/race matrix**

Run: `go test -race ./internal/nodecontrol/authority -run 'TestCommitArchiveApplicationReplacementRace|TestCommitArchiveFenceOnce|TestCommitArchiveInspectContinuity|TestCommitArchiveStagingModes' -count=1 -timeout 5m`

Expected: only one application-vs-replacement owner, append-vs-Fence winner or first consumer; same holder with a descendant lease cannot Fence twice, while a strictly higher generation can; every latch/edge/generation/once-row/response crash seam is all-or-nothing.

- [ ] **Step 6: Commit the archive contract boundary**

```bash
git add internal/nodecontrol/contracts/commit_archive.go internal/nodecontrol/authority/commit_archive_verifier.go internal/nodecontrol/authority/commit_archive_verifier_test.go internal/nodecontrol/authority/commit_archive_deterministic_model_test.go internal/nodecontrol/authority/commit_archive_concurrency_test.go testdata/c12/authority-v7/commit-archive-golden.v1.json
git commit -m "feat(nodecontrol): freeze commit archive and fence contracts"
```

### Task 15: Extend the repository, DatabaseAuthorityHead, exact-epoch readiness, and bound reads

**Files:**
- Create: `internal/nodecontrol/authority/database_head.go`
- Create: `internal/nodecontrol/authority/exact_epoch_readiness.go`
- Modify: `internal/nodecontrol/authority/repository.go`
- Modify: `internal/nodecontrol/authority/postgres_repository.go`
- Modify: `db/queries/nodecontrol_authority.sql`
- Modify: `db/queries/nodecontrol_serving.sql`
- Modify: `internal/store/nodecontrol_authority.sql.go`
- Modify: `internal/store/nodecontrol_serving.sql.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/querier.go`
- Modify: `internal/nodecontrol/serving/repository.go`
- Modify: `internal/nodecontrol/serving/postgres_repository.go`
- Test: `internal/nodecontrol/authority/database_head_integration_test.go`
- Test: `internal/nodecontrol/authority/exact_epoch_readiness_test.go`
- Test: `internal/nodecontrol/serving/postgres_repository_v7_integration_test.go`

**Interfaces:**
- Consumes: Task 8 full catalog, Tasks 11–14 verified envelopes/provider/archive facts, and Task 10 `BoundAuthorityReadSource`.
- Produces: locked `DatabaseAuthorityHeadV1`, exact-epoch repository projections and Coordinator-owned `BoundAuthorityReadSource` that B02/B03 may consume but cannot construct from an arbitrary pool.

- [ ] **Step 1: Write RED fixed-manifest DatabaseAuthorityHead tests**

Materialize each progressive genesis/epoch/recovery/staging state and require the exact Head field/null matrix. Current identity/lineage/runtime comes only from continuous registration/rebind results; effective/terminal epoch comes only from terminal-application chains; recovery application clears pending recovery without advancing effective/terminal; ordinary six-field sequence/receipt/database-point projection comes from exact current epoch, never `MAX(epoch)` or an unresolved provider response.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^(TestDatabaseAuthorityHeadV7Projection|TestDatabaseAuthorityHeadV7RejectsBrokenChain)$' -Timeout 5m`

Expected: FAIL because the v7 Head projector is absent.

- [ ] **Step 2: Implement locked Head projection and exact repository conversions**

Use one caller-owned read transaction, the fixed 51-table manifest and explicit queries for singleton/chain cardinalities. Reject unknown catalog objects, missing/duplicate predecessor, partially present groups, unresolved runtime result, sequence gap, source/staging matrix violation, legacy row under claim-v1 and any mismatched canonical digest/preimage. Return defensive bytes and finite corruption/unavailable reasons.

- [ ] **Step 3: Implement exact-epoch readiness and bound serving checks**

Before a bound read obtain stable Provider Head plus caller-challenge-bound DatabaseIncarnationProof; on one physical PostgreSQL connection perform DB readiness, exact certificate/desired query and DB readiness again; then obtain a second fresh proof and stable Head. Require activation/completion/release/Open, profile, selected genesis/root/current epoch/effective+terminal chain, current identity/lineage/registration/runtime, serving lease and all six ordinary fields to match. Pending epoch/recovery/staging, restricted/expired lease, source seal, restore-incomplete or any pre/query/post change returns empty material.

- [ ] **Step 4: Run readiness/PITR tests and commit**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git add db/queries/nodecontrol_authority.sql db/queries/nodecontrol_serving.sql internal/store/nodecontrol_authority.sql.go internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code -- db/queries/nodecontrol_authority.sql db/queries/nodecontrol_serving.sql internal/store/nodecontrol_authority.sql.go internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go`

Run: `git ls-files --others --exclude-standard -- db/queries/nodecontrol_authority.sql db/queries/nodecontrol_serving.sql internal/store`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/authority|./internal/nodecontrol/serving' -Run '^(TestExactEpochReadiness|TestBoundServingV7|TestBoundServingV7PITR)$' -Timeout 8m`

Expected: both second-generation checks are empty and all tagged tests PASS.

```bash
git add internal/nodecontrol/authority/database_head.go internal/nodecontrol/authority/exact_epoch_readiness.go internal/nodecontrol/authority/repository.go internal/nodecontrol/authority/postgres_repository.go db/queries/nodecontrol_authority.sql db/queries/nodecontrol_serving.sql internal/store/nodecontrol_authority.sql.go internal/store/nodecontrol_serving.sql.go internal/store/models.go internal/store/querier.go internal/nodecontrol/serving/repository.go internal/nodecontrol/serving/postgres_repository.go internal/nodecontrol/authority/database_head_integration_test.go internal/nodecontrol/authority/exact_epoch_readiness_test.go internal/nodecontrol/serving/postgres_repository_v7_integration_test.go
git commit -m "feat(nodecontrol): enforce authority v7 exact-epoch readiness"
```

### Task 16: Prove catalog, legacy classification, empty/source, and Down conformance

**Files:**
- Test: `internal/store/nodecontrol_v7_catalog_conformance_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_empty_source_integration_test.go`
- Test: `internal/store/nodecontrol_v7_down_crash_integration_test.go`
- Create: `internal/nodecontrol/authority/testdata/v7-legacy-shapes.sql`

**Interfaces:**
- Consumes: Tasks 8, 12 and 15; produces a reviewer-sized §11.1–11.2 and Down conformance gate without production provider code.

- [ ] **Step 1: Add RED R0/R1/C/A0/A1 and catalog fixtures**

Load each real v6 shape, unknown-table/column/FK/trigger variants and concurrent legacy-session barrier. Assert registered Up classifies only legacy_v6, preserves abort rows, creates latch/production intent atomically and leaves every new claim/commitment/resolution/evidence field empty. Assert all protocol/source/staging/claim-v1 rows and every policy/registration precursor permanently block Down.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/store' -Run '^(TestV7LegacyCatalogConformance|TestV7MigrationIsolationBarrier)$' -Timeout 8m`

Expected: initial new fixtures fail until all exact catalog/guard cases are covered.

- [ ] **Step 2: Complete the minimal GREEN empty/source and Down crash matrices**

Cross product DB empty/nonempty, provider empty/nonempty, single/multi environment, inventory/anchor/shutdown mismatch, unknown catalog, source retirement/post-seal splice and exact fork-cut. Inject failure at Down authorization insert, consume, DDL, version update and commit response; require rollback `1/0` or fully absent schema, never an intermediate state.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile base -Packages './internal/nodecontrol/authority|./internal/store' -Run '^(TestV7EmptySourceConformance|TestNodeControlV7DownCrashSeams)$' -Timeout 10m`

- [ ] **Step 3: Commit only conformance fixtures/tests**

```bash
git add internal/store/nodecontrol_v7_catalog_conformance_integration_test.go internal/nodecontrol/authority/v7_empty_source_integration_test.go internal/store/nodecontrol_v7_down_crash_integration_test.go internal/nodecontrol/authority/testdata/v7-legacy-shapes.sql
git commit -m "test(nodecontrol): prove authority v7 catalog and source gates"
```

### Task 17: Prove genesis, epoch arbitration, archive, and PITR crash recovery

**Files:**
- Test: `internal/nodecontrol/authority/v7_activation_crash_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_epoch_recovery_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_commit_archive_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_pitr_integration_test.go`

**Interfaces:**
- Consumes: Tasks 13–15 deterministic provider/archive contracts and real PostgreSQL 18.4; produces the v7 §11.3 authority-fence/PITR test gate. It does not substitute for B03 production provider/WAL runtime conformance.

- [ ] **Step 1: Add RED five-stage activation crash seams**

Inject before/after provider registration, DB registration result, attempt, Prepare, activation, activation commit proof, Complete, completion proof, PrepareRelease, release proof and Open. Persist IDs before calls, use exact Inspect after response loss, and assert ordinary writer/reader/listener/signer counts stay zero before Open.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestV7ActivationCrashMatrix$' -Timeout 10m`

Expected: FAIL until every seam has an exact retry/closed result.

- [ ] **Step 2: Complete the minimal GREEN epoch apply-vs-replacement and candidate progression matrix**

Cover same-holder suffix count 0, deferred 1/64, apply count 0/64, replacement 0..63, one non-null prefix UNIQUE, same subject-slot decision ID across reasons, application two-key proof, candidate A death before CAS, fresh candidate B proof, successful-CAS-only consume, direct-catchup versus first applied-rebind, continuous holder death, pre/post-catchup, pending-result reconcile and marker-cleared resume. Prove no application/replacement dual win or decision/proof digest cycle.

- [ ] **Step 3: Extend GREEN with archive/Fence and recursive PITR matrices**

Cover append-before-Fence and Fence-before-append; paired generations; related-only owner corruption; semantic journal/subject-slot alias attempts; same-holder lease rollover once-key; higher-generation retry; first-consumer history; two consecutive PITRs; result missing with next-winning PostRecovery, catchup Head or unconsumed-tail reconcile; exact commit LSN `<`, `=`, and `>` fork cut. Every provider call counter inside DB transactions remains zero.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/authority' -Run '^(TestV7EpochRecoveryMatrix|TestV7CommitArchiveMatrix|TestV7RecursivePITRMatrix)$' -Timeout 15m -Race`

- [ ] **Step 4: Commit only the epoch/PITR gate**

```bash
git add internal/nodecontrol/authority/v7_activation_crash_integration_test.go internal/nodecontrol/authority/v7_epoch_recovery_integration_test.go internal/nodecontrol/authority/v7_commit_archive_integration_test.go internal/nodecontrol/authority/v7_pitr_integration_test.go
git commit -m "test(nodecontrol): prove authority v7 epoch and PITR recovery"
```

### Task 18: Prove staging, readiness, canonical transcript coverage, and Batch 01 exit

**Files:**
- Create: `testdata/c12/integration-contracts-schema-authority.v1.json`
- Test: `internal/nodecontrol/authority/v7_staging_integration_test.go`
- Test: `internal/nodecontrol/authority/v7_readiness_integration_test.go`
- Test: `internal/nodecontrol/contracts/authority_v7_all_schema_golden_test.go`
- Test: `internal/nodecontrol/authority/v7_batch01_acceptance_test.go`

**Interfaces:**
- Consumes: every prior Batch 01 task; produces the final §11.4–11.6 B01 gate and explicit handoff to B02/B03/B10/B11.

- [ ] **Step 1: Add RED staging five-table and recovery matrix**

Exhaust all 32 row-count combinations and accept only the eight specified states; cross with provider null/held/required/abort-only. Cover capability commit proof before Acquire, import-vs-revocation barrier, normal and recovery Abort, holder rebind, two PITRs, generic/outcome Fence, exact-existing/missing-suffix evidence, stale candidate splice, strict fork cut, and `FreshRestoreImportApplicationV1` rejection under every Fence reason. Assert `fresh_v7_staging_closed` never becomes ready or complete.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestV7StagingRecoveryMatrix$' -Timeout 12m`

Expected: FAIL until all eight states and negative branches are explicit.

- [ ] **Step 2: Complete the minimal GREEN readiness and all-schema transcript gates**

Mutate every Provider/DB Head phase/profile/activation/identity/lineage/lease/epoch/terminal/ordinary field across the pre/query/post checks. Enumerate every v7 schema in the registry and require at least one positive literal vector plus one mutation vector; reject missing/extra registry entries and the three historical-name duplicates. Verify generated sqlc/OpenAPI/Protobuf artifacts have no drift. Create the sorted Batch 01 integration manifest with one explicit package per group, exact top-level test-name lists, the required `base`、`authority-v7` or `authority-v7-pitr` profile and bounded group timeout. The parser gate exact-covers every tracked first-line integration test inside the selected manifest's closed package set exactly once；base migration/registered-Up/Down groups never share a database with preinstalled-v7 or PITR groups.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/contracts|./internal/nodecontrol/authority|./internal/nodecontrol/serving' -Run '^(TestAuthorityV7AllSchemasHaveGolden|TestV7ReadinessMatrix|TestV7Batch01Acceptance)$' -Timeout 12m`

- [ ] **Step 3: Run the complete Batch 01 verification boundary**

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `git diff --exit-code HEAD -- api gen internal/store`

Run: `powershell -NoProfile -Command "if (git ls-files --others --exclude-standard -- api gen internal/store) { throw 'untracked generated artifact' }"`

Run: `go test ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/serving ./internal/store -count=1 -timeout 10m`

Run: `go test -race ./internal/nodecontrol/contracts ./internal/nodecontrol/authority ./internal/nodecontrol/serving -count=1 -timeout 10m`

Run in the current PowerShell process:

```powershell
$expectedBatch01IntegrationTags = [ordered]@{
    'internal/testinfra/c12_dependencies_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_schema_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_v7_migration_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_v7_acl_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_v7_down_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_v7_catalog_conformance_integration_test.go' = '//go:build integration'
    'internal/store/nodecontrol_v7_down_crash_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/authority_crash_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/authority_pitr_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_staging_import_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/database_head_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_empty_source_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_activation_crash_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_epoch_recovery_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_commit_archive_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_pitr_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_staging_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/authority/v7_readiness_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/serving/postgres_repository_integration_test.go' = '//go:build integration'
    'internal/nodecontrol/serving/postgres_repository_v7_integration_test.go' = '//go:build integration'
}
$badBatch01IntegrationTags = @($expectedBatch01IntegrationTags.GetEnumerator() | Where-Object {
    -not (Test-Path -LiteralPath $_.Key -PathType Leaf) -or
    (Get-Content -LiteralPath $_.Key -TotalCount 1) -cne $_.Value
} | ForEach-Object Key)
if ($badBatch01IntegrationTags.Count -ne 0) {
    $badBatch01IntegrationTags
    throw 'B01 integration test missing exact first-line tag'
}
```

Run: `git add testdata/c12/integration-contracts-schema-authority.v1.json internal/nodecontrol/authority/v7_staging_integration_test.go internal/nodecontrol/authority/v7_readiness_integration_test.go internal/nodecontrol/contracts/authority_v7_all_schema_golden_test.go internal/nodecontrol/authority/v7_batch01_acceptance_test.go`

Expected: those exact new integration tests and manifest are tracked in the index before the tracked-file parser runs；no broader path is staged.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Suite batch01 -Timeout 120m`

Expected: PASS with exact 51-table catalog, no generated drift, zero provider/attestor calls inside DB transactions, exactly one bounded `authority_protocol_downgrade_authorizer` call only at the registered disposable-Down §9.3 seam, zero other external calls, and explicit unsupported results for trust/authorizer rotation and legacy destructive restore.

- [ ] **Step 4: Commit the final B01 tests only**

```bash
git add testdata/c12/integration-contracts-schema-authority.v1.json internal/nodecontrol/authority/v7_staging_integration_test.go internal/nodecontrol/authority/v7_readiness_integration_test.go internal/nodecontrol/contracts/authority_v7_all_schema_golden_test.go internal/nodecontrol/authority/v7_batch01_acceptance_test.go
git commit -m "test(nodecontrol): close authority v7 batch01 acceptance"
```

## Batch 01 completion check

- Three generated OpenAPI packages contain disjoint route sets, exact listener-auth extensions and no pseudo-mTLS security schemes.
- Nodecontrol event wire vectors are deterministic, closed, and privacy-safe.
- Registered Go Up creates the exact base-25 + v7-26 = 51-table catalog, NOLOGIN ACLs, fixed guards and down-locked latch; ordinary Goose/raw SQL cannot execute 00007 or Down.
- Disposable-only Down proves the exact permanent retirement set and atomic `1/0 -> 1/1 -> 0/0` matrix; production and every fixed-forbidden row remain non-downgradable.
- Every v7 canonical body/envelope/evidence schema has a literal golden and mutation coverage; the three Abort/serving historical schema names are not duplicated.
- Environment/shutdown/incarnation/lineage/empty/source/Down verifiers, genesis/epoch/staging/provider contracts, gap/PostRecovery attestations and exact phase matrices are complete.
- Related set/purpose/generation, multi-key Fence, semantic journal/epoch subject-slot/stable-holder once-row, global latch/reverse owner, Inspect/history and staging evidence modes pass deterministic race/crash tests.
- DatabaseAuthorityHead, exact-epoch repository/readiness and Coordinator-owned `BoundAuthorityReadSource` fail closed for unresolved epoch/recovery/staging/source state and pre-revocation PITR.
- Every crash point recovers by immutable ID, exact preimage/Inspect and the specified first-consumer/fork-cut witness without a second authority effect or provider call inside a DB transaction.
- `fresh_v7_staging_closed` remains `restore_incomplete`; B01 has not implemented production provider/WAL archive runtime, `SpecDigest`, orchestration, trust/authorizer rotation or destructive legacy restore.
- Independent correctness/security review has zero Critical and zero Important findings; any Medium has an explicit disposition.
