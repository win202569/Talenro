# Talenro C1.2 Authority Canonical Evidence and Transactional Dispatcher Addendum

- 状态：待书面复核
- 日期：2026-08-28
- 架构方案：用户于 2026-08-28 选择并批准“最小规范附录”方案；本文逐字内容仍须完成书面复核后才可标记为已批准
- 修订对象：[Authority Abort Linearization and Guarded Serving Design Amendment](./2026-08-24-nodecontrol-authority-abort-serving-design.md) §§5.5、6、6.1、10.1，以及引用这些边界的 suite index、B01/B02/B03 实施计划
- 触发来源：修订版 B01 Task 7 的 RED-test preflight；批准文本无法唯一推导持久 digest、transactional resolver routing 和 admission-token API
- 规范优先级：本文获批后，仅在下述 canonical transcript、dispatcher 和 evidence admission 范围内优先于 2026-08-24 amendment；其余 Abort claim、guarded serving、Goose、PITR 和 runtime ownership 保持不变

## 1. 决策摘要

本文补齐六类实现阻塞，不扩大 B01 的业务能力：

1. 冻结 absent digest/time、final-not-applied activation-input digest 和完整 provider Head transcript。
2. 冻结 v1 commitment reason、policy 和 resolution anchor 的有限矩阵。
3. 让 transactional resolver 显式接收 expected reservation，并返回 operation/epoch/sequence echo，使 dispatcher 可以验证 routing、tuple 和跨表 cardinality。
4. 冻结 fixed unsupported registration 的表示法，不把缺失能力解释为 `EffectAbsent`。
5. 区分 fresh trusted-time evidence、持久化后无 token 的 parsed evidence，以及 package-private admission 消费权。
6. 明确 Task 7 只证明 canonical primitive 和 token 的安全性质；BEGIN、锁、写入、紧邻 Commit 消费、Commit 不确定结果由 Task 9 证明。

本文不创建 production provider、领域 resolver/activator、Abort claim、serving reader 或 runtime composition；这些仍由既有 Tasks 8–10 和 B02/B03 承担。

## 2. 通用 canonical 规则

### 2.1 Strict JSON 与 JCS

本文所有 `V1` body：

- UTF-8 JSON object，字段集合逐字等于本文顺序无关的 required registry；
- `schema_version` 的唯一合法值是 JSON string literal `"1"`；
- 禁止缺失、额外、重复字段，禁止 `null`，禁止 JSON number；
- enum 使用本文列出的 lowercase ASCII token；
- constructor 输出 RFC 8785 JCS；
- `Parse*` 只接受与其 RFC 8785 JCS 结果 byte-for-byte 相同的输入；
- 每个 parser 在 JSON/JCS 处理前拒绝空输入或大于 `4096` bytes 的输入；
- parse 后必须重新验证该 body 可自证的字段、enum、canonical、digest 和 branch 规则；涉及 commitment、Receipt、Head、checkpoint 或 trusted-time bounds 的关系必须再调用本文指定的 contextual validator；
- dynamic payload、certificate、credential、错误文本和外部 locator 不进入 transcript。

### 2.2 UUID、decimal、digest 与 time

- UUID 是 non-nil RFC 4122 lowercase canonical text，且 `uuid.Parse` 后重新格式化必须逐字相同。
- epoch、sequence 和普通 policy version 是无前导零、范围 `1..MaxInt64` 的 decimal string；DB system ID 使用 `1..MaxUint64`，timeline 使用 `1..MaxUint32`，同样禁止前导零。
- final-not-applied policy version 的唯一非 decimal token 是 literal `none`。
- present digest 是 lowercase 64-hex 且不得为 64 个 `0`。
- absent digest 的 canonical text 固定为 64 个 lowercase `0`：

```text
0000000000000000000000000000000000000000000000000000000000000000
```

- present instant 是 UTC `RFC3339Nano`，必须以 `Z` 结尾并在 parse/format 后逐字相同。
- absent instant 的 canonical text 固定为空字符串 `""`。
- `WALPosition` 沿用 provider 的 canonical uppercase `X/Y` 表示。

Absent digest 是 transcript 中的 branch marker，不可当成有效 `contracts.Digest` 返回给业务代码。`Facts()` 把 absent digest 映射为 zero `contracts.Digest`，绝不把 64-zero marker 暴露成 present digest。

## 3. 固定 activation-input empty digest

Final-not-applied commitment 不接受 caller-supplied activation-input digest。Constructor 对该 branch 固定使用：

```text
AuthorityEffectActivationInputsEmptyV1 = SHA-256(
  ASCII("talenro.nodecontrol.authority-effect-activation-inputs.empty.v1\x00")
)
```

其 lowercase hex 固定为：

```text
eafc8658c2b1ebcfce129176ac4e10ce698e5dc6fc012012184d604e6e9b69aa
```

这是 domain-only semantic constant，不是可扩展 JSON body。Conditional commitment 必须携带 nonzero 且不等于该常量的 activation-input digest。

## 4. Canonical provider Head

### 4.1 Body 与 digest

`ActivationDecisionEvidenceV1.provider_head_digest` 不能由 caller 直接提供。它必须由完整 `authority.Head` 构造以下 exact body：

```text
AuthorityProviderHeadV1 = {
  schema_version,
  authority_epoch,
  latest_reserved_sequence,
  latest_committed_sequence,
  latest_reservation_digest,
  latest_committed_operation_id,
  latest_committed_receipt_digest,
  db_system_id,
  db_timeline,
  required_lsn
}
```

```text
provider_head_digest = SHA-256(
  ASCII("talenro.nodecontrol.authority-provider-head.v1\x00") ||
  JCS(AuthorityProviderHeadV1)
)
```

### 4.2 Validation

Evidence capture occurs only after an exact committed receipt exists, so this transcript has no empty branch. Constructor requires：

- `Head` 通过现有 strict validation；
- positive epoch、reserved sequence 和 committed sequence；
- `latest_reserved_sequence >= latest_committed_sequence`；
- nonzero latest reservation and receipt digests；
- non-nil latest committed operation ID；
- complete positive latest committed database point and canonical WAL position。

Evidence input also carries the exact current committed `Receipt`. Task 7 validates `Receipt.Validate()` and requires its operation、kind、scope、epoch、sequence and effect digest to match the commitment. It does **not** claim that the Receipt database point equals the locked `StoredFence` binding；Task 9 must compare that exact database point while holding the activation transaction's fence lock。

Provider Head must be in the same epoch and have committed sequence at least the Receipt sequence. When the sequences are equal, `latest_committed_operation_id`、`latest_committed_receipt_digest` and the complete latest committed database point must equal the Receipt exactly. When Head is later, it is accepted only as a fresh observation from the same Coordinator-owned `Provider` instance；the v1 transcript is not an ancestry proof and no caller-supplied Head is accepted. A lower、empty、same-sequence mismatch、internally inconsistent or malformed Head is rejected. Evidence for a later epoch requires a separately approved epoch-transition anchor and is not inferred by this v1 helper。

The authority package owns one opaque, persistable preimage value：

```go
func NewAuthorityProviderHeadSnapshot(Head) (AuthorityProviderHeadSnapshot, error)
func ParseAuthorityProviderHeadSnapshot([]byte) (AuthorityProviderHeadSnapshot, error)
func (v AuthorityProviderHeadSnapshot) Facts() Head
func (v AuthorityProviderHeadSnapshot) CanonicalJCS() []byte
func (v AuthorityProviderHeadSnapshot) Digest() contracts.Digest
func AuthorityProviderHeadDigest(Head) (contracts.Digest, error)
```

`Facts()` deep-copies the database-point pointer；`CanonicalJCS()` returns a clone. B02/B03 must not duplicate the transcript or construct this snapshot from anything except a Coordinator-supplied `Head`。

## 5. Authority effect commitment v1

### 5.1 Input and opaque value

The public constructor remains：

```go
func NewAuthorityEffectCommitment(AuthorityEffectCommitmentInput) (AuthorityEffectCommitment, error)
func ParseAuthorityEffectCommitment([]byte) (AuthorityEffectCommitment, error)
```

`AuthorityEffectCommitmentInput` contains exact operation ID、effect kind、scope kind/digest、authority epoch/sequence、base effect digest、commitment mode、reason code、numeric activation policy version and one digest field whose zero value means absent input. The conditional branch requires that field present；the final branch requires zero and derives the fixed empty digest. The returned value owns immutable normalized facts、canonical JCS and digest；byte accessors return defensive copies。

```go
type AuthorityEffectCommitmentInput struct {
    OperationID              uuid.UUID
    Kind                     EffectKind
    ScopeKind                ScopeKind
    ScopeDigest              contracts.Digest
    Epoch                    uint64
    Sequence                 uint64
    BaseEffectDigest         contracts.Digest
    Mode                     CommitmentMode
    Reason                   AuthorityEffectReason
    ActivationPolicyVersion uint64
    ActivationInputsDigest  contracts.Digest
}
```

The immutable public projection is exact：

```go
type AuthorityEffectCommitmentFacts struct {
    OperationID             uuid.UUID
    Kind                    EffectKind
    ScopeKind               ScopeKind
    ScopeDigest             contracts.Digest
    Epoch                   uint64
    Sequence                uint64
    BaseEffectDigest        contracts.Digest
    Mode                    CommitmentMode
    Reason                  AuthorityEffectReason
    ActivationPolicyVersion uint64 // zero means transcript literal "none"
    ActivationInputsDigest  contracts.Digest
}

func (v AuthorityEffectCommitment) Facts() AuthorityEffectCommitmentFacts
func (v AuthorityEffectCommitment) CanonicalJCS() []byte
func (v AuthorityEffectCommitment) Digest() contracts.Digest
```

`Facts` is value-only；`CanonicalJCS` always returns a clone. The opaque value has no exported fields and no raw constructor。

### 5.2 Closed enums

```text
commitment_mode = conditional_apply | final_not_applied
reason_code     = none | failed | superseded |
                  activation_deadline_expired | validation_rejected
```

The Go names are frozen as `CommitmentConditionalApply`、`CommitmentFinalNotApplied` and `EffectReasonNone`、`EffectReasonFailed`、`EffectReasonSuperseded`、`EffectReasonActivationDeadlineExpired`、`EffectReasonValidationRejected`。

The 13 currently supported kinds share exactly those four final-not-applied reason tokens in v1. This deliberately resolves the earlier “per-kind finite reason” ambiguity as one shared registry；it is an availability/audit choice, not permission to invent new domain outcomes. Domain activators may reject a token that is inapplicable to a concrete workflow state, but cannot accept another token. Adding or redistributing a reason requires a new approved amendment and golden update。

`trust_bundle_publish` and `operator_authorizer_change` cannot construct a commitment until their durable workflows are separately approved。

### 5.3 Branch matrix

| Mode | Reason | Policy | Activation-input digest |
| --- | --- | --- | --- |
| `conditional_apply` | `none` | caller input `1..MaxInt64`, canonical decimal | caller input present, nonzero, not fixed empty digest |
| `final_not_applied` | one of the four finite reasons | input numeric value must be zero; transcript literal `none` | input must be absent/zero; constructor inserts the fixed empty digest |

The canonical body and domain remain exactly those approved in the 2026-08-24 amendment. Parse validates the same matrix；no parser path infers a default mode、reason or policy。

## 6. Authority checkpoint anchor v1

The canonical body and existing domain remain：

```text
AuthorityCheckpointAnchorV1 = {
  schema_version,
  checkpoint_kind,
  checkpoint_scope_digest,
  authority_epoch,
  authority_sequence,
  receipt_digest
}

checkpoint_digest = SHA-256(
  ASCII("talenro.nodecontrol.authority-checkpoint-anchor.v1\x00") ||
  JCS(AuthorityCheckpointAnchorV1)
)
```

`checkpoint_kind = node | global`. Constructor input contains the typed provider checkpoint plus the requested kind/scope. It derives epoch、sequence and receipt digest from `NodeCheckpoint` rather than requiring a full Receipt that the existing `Provider.CommittedNodeCheckpoint` API cannot return。

The Go names are `CheckpointNone`、`CheckpointNode` and `CheckpointGlobal`; `CheckpointNone` is used only by evidence facts and never constructs an anchor。

- `node` requires a valid non-global node scope digest. In an evidence context it is legal only for a node-scoped commitment and its scope digest must equal that commitment exactly. The provider checkpoint may name the latest exact node receipt **or** a later `global_node_trust` receipt because the existing node-checkpoint API intentionally aggregates both；it therefore does not require a full node-scoped Receipt preimage。
- `global` requires `global_node_trust` or `global_operator_trust` plus that scope's exact fixed global digest. In an evidence context it must equal the commitment's global scope exactly。
- The `NodeCheckpoint` must pass strict validation and have positive sequence/nonzero receipt digest；an empty checkpoint cannot construct an anchor。
- A higher-authority branch in this v1 requires the same epoch as the commitment and a strictly greater sequence. Cross-epoch comparison is invalid until an epoch-transition proof is separately approved；component-wise or lexicographic guesses are forbidden。

Only Coordinator may obtain the authority observation used to construct an anchor. For `node`, it calls the same `Provider` instance's `CommittedNodeCheckpoint(commitment.ScopeDigest)` after Finalize/Inspect. For `global`, where the existing Provider has no historical global-checkpoint API, v1 is deliberately conservative：Coordinator may construct an observation only from `Head()` plus `Inspect(head.LatestCommittedOperationID)` when that exact terminal committed Receipt matches the Head epoch、sequence、receipt digest and requested global scope. If the latest Head is unrelated, the workflow remains pending；handler or database state cannot synthesize a historical global checkpoint。

The opaque anchor exposes normalized defensive facts、canonical JCS and digest only。

```go
type AuthorityCheckpointAnchorFacts struct {
    Kind          CheckpointKind
    ScopeDigest   contracts.Digest
    Epoch         uint64
    Sequence      uint64
    ReceiptDigest contracts.Digest
}

type AuthorityCheckpointAnchorInput struct {
    Kind       CheckpointKind
    ScopeDigest contracts.Digest
    Checkpoint NodeCheckpoint
}

func NewAuthorityCheckpointAnchor(AuthorityCheckpointAnchorInput) (AuthorityCheckpointAnchor, error)
func ParseAuthorityCheckpointAnchor([]byte) (AuthorityCheckpointAnchor, error)
func (v AuthorityCheckpointAnchor) Facts() AuthorityCheckpointAnchorFacts
func (v AuthorityCheckpointAnchor) CanonicalJCS() []byte
func (v AuthorityCheckpointAnchor) Digest() contracts.Digest
```

The pure anchor constructor proves only canonical/value relations. `ValidateActivationDecisionEvidence` performs the commitment/Receipt/Head relation checks, and Task 9 proves that the observation came from the Coordinator-owned Provider call described above。

## 7. Authority effect resolution v1

### 7.1 Constructors and contextual validation

```go
type AuthorityEffectResolutionInput struct {
    Commitment AuthorityEffectCommitment
    Disposition EffectDisposition
    Reason      AuthorityEffectReason
    AnchorKind  DecisionAnchorKind
    AnchorDigest contracts.Digest
    Evidence    ValidatedActivationDecisionEvidence
}

func NewAuthorityEffectResolution(AuthorityEffectResolutionInput) (AuthorityEffectResolution, error)
func ParseAuthorityEffectResolution([]byte) (AuthorityEffectResolution, error)
func ValidateAuthorityEffectResolution(
    AuthorityEffectResolution,
    AuthorityEffectCommitment,
    ValidatedActivationDecisionEvidence,
) error
```

Every matrix row requires exact evidence. The constructor input contains the opaque commitment、disposition、reason、decision-anchor kind/digest and the branded proof obtained by context-validating evidence produced by `NewActivationDecisionEvidence`/`Complete` in that construction flow；a proof derived from parsed-only evidence is rejected by this constructor. Constructor performs the resolution/commitment/evidence matrix validation. `ParseAuthorityEffectResolution` performs only self-contained strict validation because the body intentionally contains commitment/evidence digests rather than their preimages；recovery must obtain the opaque proof returned by `ValidateActivationDecisionEvidence` and pass that proof to `ValidateAuthorityEffectResolution`. The type boundary makes skipping evidence contextual validation impossible。

### 7.2 Closed matrix

```text
disposition         = applied | not_applied
decision_anchor_kind = final_commitment | exact_capture |
                       higher_authority | trusted_time
```

The Go names are `DispositionApplied`、`DispositionNotApplied` and `DecisionAnchorFinalCommitment`、`DecisionAnchorExactCapture`、`DecisionAnchorHigherAuthority`、`DecisionAnchorTrustedTime`。

| Commitment / disposition | Reason | Anchor | Required evidence |
| --- | --- | --- | --- |
| final-not-applied / not-applied | exact commitment reason | `final_commitment`; exact commitment digest | none-time, not-applied-only evidence |
| conditional / applied | `none` | `trusted_time`; exact evidence digest | rollback-resistant, may-apply evidence with live token for a new commit |
| conditional / not-applied from immutable captured facts | `failed` or `validation_rejected` | `exact_capture`; exact commitment activation-input digest | rollback-resistant, not-applied-only evidence |
| conditional / not-applied from higher authority | `superseded` | `higher_authority`; exact checkpoint digest | none-time, not-applied-only evidence with a strictly higher committed checkpoint |
| conditional / not-applied from trusted deadline | `activation_deadline_expired` | `trusted_time`; exact evidence digest | rollback-resistant, not-applied-only evidence |

All other combinations are invalid. Every rollback-resistant row requires checkpoint `none`. The higher-authority row requires exactly one present checkpoint and can only be `not_applied_only`; final-not-applied requires checkpoint `none`. Thus a higher checkpoint can never accompany `may_apply`、`exact_capture` or `trusted_time` resolution. This conservative v1 matrix may keep an operation pending when evidence is insufficient；it never upgrades authority by guessing another anchor。

Persisted not-applied resolution remains monotonic：once validated against rollback-stable evidence, replay cannot replace it with applied. `NewAuthorityEffectResolution` requires fresh construction origin for every row and a live, matching admission state for every rollback-resistant row；the two none-time rows must be tokenless. Parsing and contextual validation never create a token. A parsed applied resolution can only be accepted as an already committed outcome after Task 9 reads the exact evidence、resolution and all validation preimages from the same locked domain record/transactional snapshot and verifies their atomic binding；it can never authorize a new commit。

```go
type AuthorityEffectResolutionFacts struct {
    CommitmentDigest contracts.Digest
    Disposition      EffectDisposition
    Reason           AuthorityEffectReason
    AnchorKind       DecisionAnchorKind
    AnchorDigest     contracts.Digest
}

func (v AuthorityEffectResolution) Facts() AuthorityEffectResolutionFacts
func (v AuthorityEffectResolution) CanonicalJCS() []byte
func (v AuthorityEffectResolution) Digest() contracts.Digest
```

## 8. Activation decision evidence v1

### 8.1 Public and package-private APIs

```go
func NewActivationDecisionEvidence(ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error)
func ParseActivationDecisionEvidence([]byte) (ActivationDecisionEvidence, error)
func ValidateActivationDecisionEvidence(
    ActivationDecisionEvidence,
    ActivationDecisionEvidenceInput,
) (ValidatedActivationDecisionEvidence, error)
func BeginActivationEvidenceCapture() *ActivationEvidenceCapture
func (c *ActivationEvidenceCapture) Complete(ActivationDecisionEvidenceInput) (ActivationDecisionEvidence, error)
```

`NewActivationDecisionEvidence` is a pure constructor only for the two approved `trusted_time_kind=none` branches：exact final-not-applied and strictly higher committed authority. It never creates an admission token。Rollback-resistant evidence must use `BeginActivationEvidenceCapture` before **any** handler trusted-time/attestation or Coordinator Provider observation and `Complete` afterward. The package retains a private test seam `beginActivationEvidenceCaptureForTest(func() time.Time)`；production callers cannot inject a clock。

`ParseActivationDecisionEvidence` validates only the persisted evidence body's self-contained canonical rules and always returns a value without an admission token. `ValidateActivationDecisionEvidence` recomputes and compares every external preimage/relation using the exact input originally used for construction and returns an opaque `ValidatedActivationDecisionEvidence` accepted by the resolution validator；the proof has no public constructor or admission method. Validation never creates a token：a fresh rollback-resistant proof retains only the original shared opaque admission state, while a parsed or none-time proof has none. Parsing or validation is suitable only for exact terminal recovery/verification, never for authorizing a new applied commit。

The authority package owns exact unexported `consumeActivationAdmission(context.Context, ValidatedActivationDecisionEvidence) error`, used later by Coordinator Task 9. No B02/B03 caller receives a raw token or consume method。

The branded proof has no public constructor and exposes only defensive persistence views：

```go
func (v ValidatedActivationDecisionEvidence) Evidence() ActivationDecisionEvidence
func (v ValidatedActivationDecisionEvidence) Input() ActivationDecisionEvidenceInput
```

`Input()` deep-copies every pointer/byte-bearing opaque value. A proof made from fresh `Complete` may retain the same unexported shared admission state so Coordinator can consume it；a proof made from parsed evidence has none. Neither accessor exposes a token API, and copying the proof cannot duplicate consumption authority。

The public projection is exact and never exposes token state：

```go
type ActivationDecisionEvidenceFacts struct {
    OperationID            uuid.UUID
    CommitmentDigest       contracts.Digest
    ProviderHeadDigest     contracts.Digest
    CheckpointKind         CheckpointKind
    CheckpointScopeDigest  contracts.Digest
    CheckpointDigest       contracts.Digest
    TrustedTimeKind        TrustedTimeKind
    TrustedInstant         time.Time // zero iff kind=none
    EvidenceValidUntil     time.Time // zero iff kind=none
    ProviderIdentityDigest contracts.Digest
    FloorAttestationDigest contracts.Digest
    Capability             DecisionCapability
}

func (v ActivationDecisionEvidence) Facts() ActivationDecisionEvidenceFacts
func (v ActivationDecisionEvidence) CanonicalJCS() []byte
func (v ActivationDecisionEvidence) Digest() contracts.Digest
```

### 8.2 Split ownership and exact validation input

A registered B02/B03 handler owns only domain commitment/policy evaluation and the trusted-time source it is approved to use. It returns material；it never receives raw `Provider` and never constructs final evidence. Coordinator owns Receipt、Provider Head、checkpoint observation and the final B01 constructor/capture call。

```go
type ActivationDecisionMaterial struct {
    Commitment                    AuthorityEffectCommitment
    Reason                        AuthorityEffectReason
    CheckpointKind                CheckpointKind
    CheckpointScopeDigest         contracts.Digest
    TrustedTimeKind               TrustedTimeKind
    TrustedInstant                time.Time
    EvidenceValidUntil            time.Time
    AttestationExpiresAt          time.Time
    ActivationDeadline            time.Time
    ProviderIdentityDigest        contracts.Digest
    ExpectedProviderIdentityDigest contracts.Digest
    FloorAttestationDigest        contracts.Digest
    Capability                    DecisionCapability
}

type ActivationDecisionEvidenceInput struct {
    Material     ActivationDecisionMaterial
    Receipt      Receipt
    ProviderHead AuthorityProviderHeadSnapshot
    Checkpoint   *AuthorityCheckpointAnchor
}
```

The handler must cryptographically/semantically verify its trusted-time observation、floor attestation、expected provider identity and commitment-bound deadline before returning material. The expected identity comes from immutable activator configuration, never the per-call business API；the deadline comes from exact immutable activation-input facts whose domain helper recomputes the commitment's activation-input digest. B01 does not authenticate arbitrary external bytes, but it independently requires nonzero observed/expected identity digests, exact equality between them, canonical bounds and every inequality below；a buggy handler therefore cannot extend a supplied deadline merely by choosing `evidence_valid_until`。

The final input contains no caller-supplied provider-head/checkpoint/evidence digest. Constructor and contextual validator derive those digests from the opaque preimages, validate `Receipt.Validate()` with `StatusCommitted`, and require Receipt operation、kind、scope、epoch、sequence and effect digest to equal the commitment exactly. The `StoredFence`/database-point comparison remains Task 9's locked-transaction obligation。

```text
checkpoint_kind     = none | node | global
trusted_time_kind   = none | rollback_resistant
decision_capability = may_apply | not_applied_only
```

The remaining Go names are `TrustedTimeNone`、`TrustedTimeRollbackResistant` and `DecisionCapabilityMayApply`、`DecisionCapabilityNotAppliedOnly`。The canonical evidence body/domain remain exactly those approved in the 2026-08-24 amendment。

### 8.3 Closed evidence-material matrix

| Commitment / intended reason | Checkpoint | Trusted time | Capability | Additional relation |
| --- | --- | --- | --- | --- |
| final-not-applied / exact commitment reason | none | none | not-applied-only | no time/bound fields |
| conditional / none | none | rollback-resistant | may-apply | trusted instant strictly before activation deadline |
| conditional / failed or validation-rejected | none | rollback-resistant | not-applied-only | trusted instant strictly before activation deadline |
| conditional / activation-deadline-expired | none | rollback-resistant | not-applied-only | trusted instant equal to or after activation deadline |
| conditional / superseded | exact strictly-higher node/global anchor | none | not-applied-only | same epoch and checkpoint sequence strictly greater than commitment |

No other row is valid. In particular, a present checkpoint cannot coexist with rollback-resistant time or `may_apply`; a final commitment cannot carry a checkpoint；and exact-capture/deadline rows cannot use a higher-authority anchor。

For checkpoint `none`, material scope、evidence checkpoint scope and evidence checkpoint digest map to the canonical absent digest, while input `Checkpoint` must be nil. For a present checkpoint, material kind/scope and opaque anchor facts must agree exactly；the evidence validator also requires Head committed sequence at least the anchor sequence, and at equality requires Head receipt digest equal the anchor receipt digest。

For trusted time `none`, all four time fields are zero and provider identity、expected provider identity and floor/attestation digests are zero；canonical evidence instants use `""` and its two digests use the absent marker. For rollback-resistant time, all four instants and all three digests are present；every instant is canonical UTC and observed provider identity equals expected identity。

### 8.4 Exact freshness bounds and admission

Every rollback-resistant row requires：

```text
trusted_instant < evidence_valid_until
trusted_instant < attestation_expires_at
```

For `may_apply` and the pre-deadline `failed|validation_rejected` rows：

```text
trusted_instant < activation_deadline
evidence_valid_until <= min(
  attestation_expires_at,
  activation_deadline,
  trusted_instant + 5 seconds
)
```

For `activation_deadline_expired`：

```text
trusted_instant >= activation_deadline
evidence_valid_until <= min(
  attestation_expires_at,
  trusted_instant + 5 seconds
)
```

All additions/durations are checked. Zero、negative、noncanonical、`trusted_instant == evidence_valid_until`、`trusted_instant == attestation_expires_at` or overflow is invalid；`evidence_valid_until` may equal an upper bound because the approved inequality is inclusive. `trusted_instant == activation_deadline` is expired and therefore legal only for the deadline-expired row。

`ActivationEvidenceCapture.Complete` accepts only `trusted_time_kind=rollback_resistant`; none-time callers use the pure constructor. Capture records `capture_started` before the first external call. Token deadline is：

```text
capture_started + min(evidence_valid_until - trusted_instant, 5 seconds)
```

The duration must be positive and representable without overflow. `Complete` is one-shot even on failure and requires its monotonic reading to be strictly before the deadline；equal is expired。

The returned opaque value points to one shared atomic admission state bound to operation ID、commitment digest and evidence digest. Copying the Go value shares, rather than duplicates, that state. The first consume attempt is one-shot even when it fails：when shared state exists, consume atomically transitions `live → spent` before deciding cancellation、binding or expiry, and every failure leaves it spent. Concurrent/copy attempts therefore have at most one state claimant, and a canceled、expired or binding-mismatched first attempt can never be retried with another copy；Task 9 rolls back and recaptures. Missing/parsed/none-time values have no state and fail consume without manufacturing one。

### 8.5 Recovery preimages and persistence ownership

The evidence body intentionally stores digests, so Task 8/9 storage must atomically retain enough immutable data to reconstruct the exact `ActivationDecisionEvidenceInput`：commitment JCS；the full committed Receipt fields already bound to the fence；provider-head snapshot JCS；nullable checkpoint-anchor JCS；reason、attestation expiry、activation deadline and expected provider-identity digest；the exact immutable activation-input preimage/typed domain columns needed by the registered handler to recompute the commitment's activation-input digest；plus evidence JCS and resolution JCS/digests. Each parsed JCS value is re-canonicalized and digest-checked before constructing the validation input。A mutable “current Head” lookup is not a substitute for the stored Head preimage。

If any preimage is absent、noncanonical or does not reproduce the stored evidence/resolution exactly, recovery remains unavailable/fail-closed. Task 9 may accept a tokenless parsed applied outcome only after reading this exact set from one locked domain record/transactional snapshot and proving it was atomically committed；Task 7 parse/context validation alone is never such a proof。

For rollback-resistant evidence only, Task 9 must call the package-private consume helper after all activation writes and immediately before its one Commit attempt. A fresh none-time branded proof carries no token and must not call consume；it proceeds directly from completed writes to the one Commit attempt. Task 7 tests caller delay、just-under/equal/over、copy/reuse and parse/restart；it does not claim to prove transaction ordering、fence binding、atomic persistence or Commit uncertainty。

## 9. Transactional effect dispatcher

### 9.1 Revised resolver boundary

The earlier operation-ID-only interface is replaced because it could neither select a handler safely nor verify operation/epoch/sequence echo. The handler-facing query and Task-9-facing resolver are deliberately different：

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
    ResolveRegisteredAuthorityEffectForUpdate(
        context.Context,
        store.DBTX,
        TransactionalEffectQuery,
    ) (TransactionalResolvedEffect, error)
}

type TransactionalEffectResolver interface {
    ResolveAuthorityEffectForUpdate(
        context.Context,
        store.DBTX,
        Reservation,
    ) (ResolvedEffect, error)
}

type RegisteredEffectActivator interface {
    CaptureActivationDecisionMaterial(context.Context, Receipt) (ActivationDecisionMaterial, error)
    ActivateAuthorityEffect(context.Context, store.DBTX, Receipt, ValidatedActivationDecisionEvidence) error
    ValidatePersistedAuthorityEffect(
        context.Context,
        store.DBTX,
        Receipt,
        ValidatedActivationDecisionEvidence,
        AuthorityEffectResolution,
    ) error
}

type TransactionalEffectActivator interface {
    RegisteredEffectActivator
}

type EffectDispatcher interface {
    TransactionalEffectResolver
    TransactionalEffectActivator
    authorityEffectDispatcher()
}
```

`Expected()` is the exact locked fence expectation. `RegisteredKind()` is the dispatcher-selected handler being probed, so one domain handler object may safely own several registrations without guessing which table to query. The query fields are unexported and there is no public constructor；only `NewEffectDispatcher`'s private concrete dispatcher can create a nonzero query. Task 9 obtains a `Reservation` from `Repository.Lock` inside the same `READ COMMITTED` transaction and calls only the dispatcher-facing `ResolveAuthorityEffectForUpdate(ctx, tx, expected)`；it cannot choose a registered probe kind。

The package-private marker prevents ordinary direct implementations, but Go method promotion permits an external wrapper to embed a valid dispatcher and override exported methods. Therefore Task 9 `NewCoordinator` must additionally require the interface's exact non-nil private concrete dynamic type returned by `NewEffectDispatcher` and store/use that concrete value；wrapper、embedded、proxy or alternate dynamic types return exact `ErrInvalidArgument` before any method call. B02/B03 can implement only registered handler interfaces and must pass them through `NewEffectDispatcher`, so no external aggregate can bypass the exact registry、unsupported entries、query construction or cross-table probes。

Every registered resolver receives the same expected reservation plus its own exact registered kind and queries that kind's real domain table by operation ID using the supplied DBTX. It returns an exact echo：

- absent：expected operation/epoch/sequence plus exact zero-field `ResolvedEffect{State: EffectAbsent}`；
- non-absent：expected operation/epoch/sequence plus valid kind/scope/digest/state from that handler’s domain row。

The dispatcher validates echoes、registration kind、reservation scope and finite state, then strips the echo and returns only the unique `ResolvedEffect` to Task 9. The real domain resolver remains responsible for verifying its row’s stored operation ID/authority tuple before returning；later integration tests mutate those rows to prove this obligation。

Material capture is separately direct-routed by the exact committed Receipt kind. Only after Finalize/Inspect has produced that exact committed Receipt and every earlier DB lock/transaction is released, Coordinator calls `BeginActivationEvidenceCapture()` immediately before the first evidence-related external observation；it then calls dispatcher `CaptureActivationDecisionMaterial`, the Coordinator-owned same-`Provider`-instance Head/checkpoint methods, `Complete` or the none-time constructor, and finally `ValidateActivationDecisionEvidence` to obtain the branded proof passed to activation. This sequence gives handlers no raw Provider while charging every trusted-time、Head and checkpoint observation against the same monotonic budget. The dispatcher rejects a proof whose evidence originated from Parse before invoking `ActivateAuthorityEffect`；parsed proofs are recovery-validation-only. During fresh activation the handler uses only the proof's defensive persistence views plus its locked immutable domain row to recompute the activation-input binding and atomically store every §8.5 preimage with resolution；only Coordinator can consume admission。

`ValidatePersistedAuthorityEffect` is the sole domain-specific recovery entry for a parsed proof；dispatcher rejects fresh-origin proofs at this entry. Dispatcher direct-routes it by the exact committed Receipt kind. Under the caller-owned DBTX it must perform zero writes/external calls；it rereads one exact terminal domain row, recomputes the handler-specific activation-input digest from its immutable typed preimage, proves that commitment、evidence preimages、resolution and terminal columns were atomically stored together, and rejects prepared/nonterminal、missing、duplicate or mismatched state. It cannot create a resolution、advance a pointer/generation、recapture time or obtain admission。Task 9 calls it only after generic evidence/resolution contextual validation and while holding the same recovery snapshot/fence lock。

Before either activation/recovery handler call, dispatcher requires the separately supplied Receipt to equal `proof.Input().Receipt` in every Reservation、effect、database-point、status/reason and receipt-digest field；it also requires the proof commitment/evidence operation、kind、scope、epoch、sequence and effect digest to match that Receipt, and for recovery requires the supplied resolution commitment digest to match the proof commitment. A cross-operation、cross-receipt、tuple、database-point、commitment or evidence binding mismatch returns exact `ErrConflict` before any handler call；a zero/malformed proof or resolution returns exact `ErrInvalidArgument`。

### 9.2 Closed registrations

`EffectRegistration` keeps its approved three fields：

```go
type EffectRegistration struct {
    Kind      EffectKind
    Resolver  RegisteredEffectResolver
    Activator RegisteredEffectActivator
}

func NewEffectDispatcher([]EffectRegistration) (EffectDispatcher, error)
```

`NewEffectDispatcher` requires exactly one registration for every known `EffectKind`：

- the 13 supported kinds require non-nil resolver and activator；
- `trust_bundle_publish` and `operator_authorizer_change` require both interfaces nil and are recorded as fixed unsupported；
- missing、duplicate、unknown、typed-nil、non-nil unsupported or nil supported registration returns `ErrInvalidArgument`。

The exact supported registry and probe order is：

```text
certificate_activate
certificate_revoke
desired_activate
grant_claim
grant_create
identity_epoch_advance
metadata_publish
operator_transition
recovery_activate
resource_envelope_activate
root_publish
security_incident_open
security_incident_resolve
```

The two remaining known kinds are fixed unsupported：

```text
operator_authorizer_change
trust_bundle_publish
```

For a supported expected reservation, resolution probes all 13 supported resolvers in the listed bytewise effect-kind order. Zero non-absent results returns exact `EffectAbsent`；one returns that exact effect；multiple、malformed、cross-kind or mismatched results fail closed. This cross-table exact-one rule prevents a duplicate operation in another domain table from being silently ignored。

For unknown or fixed unsupported expected kind, resolution returns finite conflict before any handler call and never reports absent. Capture/activation dispatches directly by the exact committed receipt kind；unknown/fixed-unsupported kind returns conflict, while a known-kind malformed Reservation/Receipt returns invalid argument, always before a handler call。

Result classification is exact：

| Handler/caller value | Exact result |
| --- | --- |
| absent result with expected operation/epoch/sequence echo and exact `ResolvedEffect{State: EffectAbsent}` | valid absent |
| absent result with a different but individually valid echo, or any nonzero/dirty effect field | `ErrInjectedFailure` |
| non-absent result with unknown enum/state、zero required digest、invalid echo coordinate or another internally invalid field | `ErrInjectedFailure` |
| structurally valid non-absent result whose operation/epoch/sequence echo differs from expected | `ErrConflict` |
| structurally valid non-absent result whose kind differs from registered probe or expected kind | `ErrConflict` |
| structurally valid non-absent result whose scope kind/digest differs from expected | `ErrConflict` |
| two or more structurally valid non-absent results | `ErrConflict` |
| internally malformed handler `ActivationDecisionMaterial` | `ErrInjectedFailure` |
| structurally valid material whose commitment tuple/effect digest differs from Receipt | `ErrConflict` |
| nil or typed-nil caller DBTX、zero/malformed caller proof or resolution | `ErrInvalidArgument` |
| structurally valid proof/resolution bound to another Receipt/commitment | `ErrConflict` |
| parsed-origin proof supplied to fresh activation, or fresh-origin proof supplied to persisted validation | `ErrConflict` |

Validation/mismatch decisions occur before the corresponding handler call where the value is caller-supplied, and immediately after that one handler returns where it is handler-supplied。

No handler error is ever converted to absence or returned with its dynamic text. Exact error mapping for resolve、material capture、activation and persisted-outcome validation is：

- constructor shape/registry failure or malformed non-kind caller value → exact `ErrInvalidArgument`；
- unknown/fixed-unsupported runtime kind、multiple valid rows、cross-kind tuple、Receipt/proof/resolution binding mismatch or any handler error satisfying `errors.Is(err, ErrConflict)` → exact `ErrConflict`；
- canceled context、handler error satisfying `errors.Is(err, context.Canceled)`/`errors.Is(err, context.DeadlineExceeded)` or `errors.Is(err, ErrCanceled)` → exact `ErrCanceled`；
- malformed handler output and every other handler/dependency error, including dynamic SQL/domain text → exact `ErrInjectedFailure`。

The dispatcher returns only those finite sentinels, never wraps handler text, and never includes operation、scope or digest values。

Runtime precedence is fixed：a canceled context wins at entry and after every handler return；otherwise kind classification precedes full value validation so unknown/fixed-unsupported is conflict and known-kind malformed input is invalid argument. Supported resolution invokes handlers in the frozen order and stops at the first mapped handler error or malformed echo；only after all 13 valid echoes are collected does it apply zero/one/multiple cardinality。No later result can mask an earlier failure。

## 10. TDD and acceptance boundary

### 10.1 Task 7 must prove

- independent literal JCS/digest vectors for commitment、resolution、checkpoint、provider Head and evidence；
- fixed empty-input digest and every absent marker；
- conditional/final mode separation and every field mutation；
- same-sequence Head exact match、later same-epoch Head、lower/forked mismatch and Head snapshot round trip；
- every valid evidence/resolution matrix row and rejection of every checkpoint/time/capability cross-row combination；
- strict parse rejection of over-4096、missing/extra/duplicate/null/number/noncanonical data and proof that parse alone cannot satisfy contextual validation；
- attestation/deadline min-bound just-under/equal/over、overflow、one-shot Complete、copy/reuse and parsed/restarted no-token behavior；
- failed consume burns shared state for cancellation、expiry and binding mismatch, including copy/concurrent retry；
- exact 15-kind registration coverage、two fixed unsupported entries、sealed dispatcher、opaque query、zero/one/multiple/corrupt/cross-kind resolver results and deterministic order；
- fresh-vs-Parse proof routing、Receipt/proof/resolution pre-handler binding and persisted-validator direct routing；
- finite error mapping, wrapped-sentinel normalization and nil/typed-nil DBTX rejection that never returns dynamic handler text；
- fuzz no panic/aliasing and race no duplicate token consumption。

The vector fixture is a strict version-1 object whose entries contain only hard-coded `name`、`artifact_kind`、semantic `input`、literal `canonical_jcs` and literal lowercase `digest_hex`. Expected JCS/digests are reviewed constants produced by an independent test-local encoder/preimage, never production helpers or values regenerated during the assertion. Its mutation manifest labels every field mutation either `reject` with the exact finite sentinel or `valid_alternate` with a literal different JCS/digest；tests must not misclassify a legal branch as rejection。

### 10.2 Explicitly deferred to Task 9

- fence lock and dispatcher resolve use the same DBTX；
- claim commit happens before provider Abort；
- evidence capture happens outside DB locks and only after the exact committed Receipt exists；
- the exact `Begin → registered material/trusted-time → same-Provider Head/checkpoint → Complete/New → contextual validation` order and Provider-instance provenance are observed；
- `NewCoordinator` rejects nil、typed-nil、embedded/wrapped/proxy and alternate dispatcher dynamic types and stores only the exact private concrete returned by `NewEffectDispatcher`；
- exact Receipt database point is compared with the locked StoredFence binding；
- rollback-resistant token consumption occurs after all writes and immediately before exactly one Commit attempt, while none-time rows never consume；
- evidence/resolution and every contextual-validation preimage are persisted atomically and reloaded from one locked snapshot；
- Commit response loss reads and contextually validates that exact persisted outcome before recapture；
- parsed-outcome recovery calls `ValidatePersistedAuthorityEffect` in the same locked DBTX and proves it performs zero writes/external calls；
- two PostgreSQL connections deterministically prove one Abort/effect winner。

Task 7 primitive tests cannot be reported as proof of those Coordinator properties。

### 10.3 Explicitly owned by B02/B03 domain plans

- each registered handler authenticates its approved trusted-time/floor attestation, derives expected provider identity from immutable configuration and rejects every identity/signature/expiry mutation；
- each handler derives activation deadline and all policy inputs from immutable domain facts and independently recomputes the commitment activation-input digest；
- real resolver/capture/activate/recovery tests mutate operation、registered/effect kind、scope、epoch、sequence、effect/commitment digest、every validation preimage and terminal column；each mismatch fails before visibility change；
- `ValidatePersistedAuthorityEffect` accepts only one exact terminal row in the caller DBTX and performs zero write、provider、trusted-time、signer or issuer calls。

Task 7 fake handlers prove dispatcher mechanics only；they are not evidence of trusted-source authenticity or a real domain-row binding。

## 11. Downstream plan amendments required before implementation

After this document is marked approved：

1. Add its exact SHA-256 to `c12-spec-set.v1.json` in bytewise path order and recompute the canonical spec-set digest where owned。
2. Update suite index canonical APIs with provider Head snapshot、checkpoint/evidence parse、contextual evidence/resolution validation and the two-level resolver signatures。
3. Update Plan 01 Task 7 tests/files without changing its eight production/test paths；Task 9 consumes the new resolver echo and private admission helper。
4. Update B02/B03 handlers to implement `RegisteredEffectResolver` with the opaque query and all three `RegisteredEffectActivator` methods, including zero-write persisted-outcome validation；add every §10.3 authenticity/binding mutation test. Only the sealed dispatcher exposes the Reservation-in/`ResolvedEffect`-out Task 9 interface。
5. Amend Task 8/9 schema and repository plans to persist/reload every §8.5 preimage atomically with evidence/resolution；no plan may validate recovery from digest-only rows or a fresh current Head。
6. Preserve Tasks 8–10 ownership：Task 8 adds claim-aware storage/guards and exact proof columns, Task 9 integrates Coordinator atomicity, Task 10 implements real guarded serving/PITR。

No production implementation or golden generation begins until the written spec and the amended implementation plan are both approved。

## 12. Completion conditions

This addendum is complete only when：

1. no canonical field、domain、absent literal、enum or matrix remains implicit；
2. the dispatcher can distinguish unsupported、absent、one exact effect and corrupt/multiple effects without caller guesses；
3. parsed evidence cannot recreate a commit right；
4. Task 7 and Task 9 claims are clearly separated；
5. suite index and B01/B02/B03 plans consume the same interface；
6. independent review reports zero Critical/Important ambiguity in this scope。
