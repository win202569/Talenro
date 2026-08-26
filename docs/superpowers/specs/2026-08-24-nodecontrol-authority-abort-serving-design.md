# Talenro C1.2 Authority Abort Linearization and Guarded Serving Design Amendment

- 状态：已批准
- 日期：2026-08-24
- 修订对象：[节点与 POP 控制面设计](./2026-08-23-node-pop-control-plane-design.md) §§6.2–6.4、§§7.4–7.5、§9.1、§§10.1–10.3、§17.2、§18、§21，以及这些章节引用的 post-provider/post-signer/post-issuer activation freshness gate
- 触发来源：C1.2-B01 Task 7 独立审查
- 实施前提：本文件获批后，另行生成并批准实施计划；在此之前不得继续扩大实现修改
- 规范优先级：上述范围内凡基础设计要求在 external result/fence finalize 后的最终数据库事务中用当前时间重新判定 activation deadline、approval/evidence freshness 或 key/material window，本修订第 5.5 节的 rollback-resistant evidence-capture 加 same-process monotonic admission 线性化规则取代该时间判定；其余 immutable/authority-fenced 状态重检以及 serving/request/response-commit 的当前有效性检查不变

## 1. 决策摘要

本修订关闭三个在 C1.2-B01 Task 7 审查中暴露的边界缺口：

1. 只读 `EffectResolver` 与 provider `Abort` 之间存在跨事务、跨进程的 absent-to-committed 竞态；一次读取或进程内锁不能成为线性化证据。
2. PITR 测试只读取 authority fence 行，未通过生产代码读取真实 certificate 与 desired-state 关系，因而不能证明旧授权载荷不可服务。
3. `00006_nodecontrol.sql` 中的 PL/pgSQL 函数没有 Goose statement annotations，普通 pinned Goose `up` 无法应用完整迁移。

批准的解决方案是：

- 在既有 `control_plane_authority_fences` 行上增加持久的 `abort_claimed_at`，并复用现有 `abort_reason`；不新增 claim 表，也不增加 `prepared_effect_digest`。
- 所有 authority-bearing 领域写入先锁定对应 fence 行，数据库 trigger 再强制拒绝已有 abort claim；`Abort` 在同一 `READ COMMITTED` 事务内锁 fence、解析领域效果并写入 claim。
- 新增 `internal/nodecontrol/serving` 内部生产能力边界，以 readiness 前后双检和 exact head equality 保护真实 certificate/desired 查询。
- 修复 migration 00006 的 Goose annotations，并以真实 pinned Goose 完成 fresh `up -> down-to 5 -> up` 验证。

本修订不会把 deterministic fake 接入 `control-api`，不会发明尚不存在的 production authority provider，也不会提前实现 B02/B03 的 operator、signer、identity、listener 或 runtime composition。

## 2. 必须保持的安全不变量

1. PostgreSQL 仍是领域事实源；外部 provider 仍是 rollback-resistant epoch/sequence/effect receipt 权威。Abort claim 只负责协调并发，不能替代任一方。
2. provider 调用始终发生在数据库锁和事务之外。
3. 只有 exact absent、unbound、unclaimed 的 operation 才能获得 Abort claim。
4. 领域效果与 Abort claim 对同一 operation 只能有一个先行者；两者不能都成功。
5. 任何未知 effect kind、缺失 resolver、删除/篡改/重复领域行、tuple/digest 不匹配或 provider/DB 分叉都 fail closed。
6. terminal receipt 与 claim 均不可变；exact retry 不创建第二个 provider effect，不重写 terminal time，也不改变原始 abort reason。
7. authority readiness 关闭时，certificate、desired bytes 及其可用于授权的派生值均不得离开 serving reader。
8. PITR 后即使旧领域行仍可由原始 SQL 看见，guarded production reader 也必须返回固定 `ErrAuthorityUnavailable` 且不返回载荷。

## 3. 方案比较

### 3.1 已选择：既有 fence 行上的 durable Abort claim

每个 operation 已有唯一、持久且跨进程可锁定的 fence 行。把 claim 放在该行上，可以让领域 writer 与 Abort 围绕同一个 PostgreSQL row lock 排序，并让 crash、failover、response loss 和 PITR 后的恢复都存在数据库证据。

### 3.2 未选择：session advisory lock

Session advisory lock 在连接或进程消失后不保留证据，无法跨 crash/PITR 证明谁先获得权利；若跨 provider 调用持有它，又会违反短事务和连接池边界。它最多可作为 durable protocol 之上的性能优化，不能承担正确性。

### 3.3 未选择：停用自动 Abort

永久保留所有尚未物化领域事实的 reservation 可以避免误 abort，却会破坏这类合法失败的有界恢复与 provider pending 清理。本设计只为 exact absent/unbound operation 保留 Abort，并要求先取得可审计的 durable claim。任何已经物化的 prepared/terminal intent 不再通过 Abort 清理，而必须收敛为第 6 节定义的 immutable fail-closed committed outcome。

### 3.4 未选择：第二张 claim 表或额外 prepared digest

第二张表会重新引入 claim 行尚不存在时的插入竞态，并增加双锚点恢复问题。经批准的最小协议只增加 `abort_claimed_at`，复用 `abort_reason`，依靠 fence-first writer、数据库 trigger 和事务内 resolver 关闭 TOCTOU；本修订不引入 `prepared_effect_digest`。

## 4. Fence 数据模型与有限状态

`control_plane_authority_fences` 增加：

```sql
abort_claimed_at timestamp with time zone NULL
```

`abort_reason` 同时表示 pending Abort claim 的不可变原因和最终 aborted receipt 的原因。它不能由 terminalization 重新选择。

| 状态 | provider / visibility | DB binding | Abort claim | 允许的后继 |
| --- | --- | --- | --- | --- |
| pending open | `reserved / fence_pending` | absent | absent | bound、abort-claimed |
| pending bound | `reserved / fence_pending` | exact all-or-none group | absent | committed |
| pending abort-claimed | `reserved / fence_pending` | absent | exact reason/time | aborted |
| committed | `committed / active` | exact all-or-none group | absent | 无 |
| aborted | `aborted / aborted` | absent | exact reason/time | 无 |

`pending open` 只描述 fence 行的物理形态，不等价于领域 effect absent。领域事实可能已在另一个事务中 prepared/committed，因此取得 claim 前必须在同一 fence-lock transaction 内调用 resolver。

Provider wire contract 仍可表示 bound-aborted receipt，既有 canonical digest/vector 不因本修订删除；但本 Coordinator profile 从不在 binding 后调用 Abort。观察到 bound-aborted provider record 时保持 mismatch/unavailable，不把它复制为 durable-claim terminal。若未来需要支持该分支，必须另行批准其领域语义和迁移协议。

Migration、manifest 和 catalog tests 必须共同强制：

- `abort_reason` 与 `abort_claimed_at` 在 pending claim 和 aborted terminal 中成对存在，其余状态同时为空。
- claim 与 `effect_digest/db_system_id/db_timeline/required_lsn/effect_bound_at` binding group 互斥。
- committed 必须有完整 binding 且无 claim；aborted 必须有 claim、无 binding，receipt reason 必须等于 stored reason。
- `abort_claimed_at >= reserved_at`，`terminal_at >= abort_claimed_at`。
- claim tuple `(abort_reason, abort_claimed_at)` 写入后不可改变。
- bound tuple 写入后不可改变；terminal row 完全不可变。
- `BindEffect`、Finalize/activation 和所有领域 effect trigger 遇到 claim 时拒绝。
- exact already-aborted retry 返回原始 stored receipt，不执行 UPDATE、不产生新的 terminal timestamp，也不发起第二次 provider mutation。

Repository 的 Go 状态必须显式承载 claim，不能把它塞进 provider `Record` 或从 nullable columns 临时猜测：

```go
type AbortClaim struct {
    Reason    AbortReason
    ClaimedAt time.Time
}

type StoredFence struct {
    Record     Record // provider-compatible reservation/binding/terminal mirror
    AbortClaim *AbortClaim
}
```

`Repository.Get` 与 transaction-bound `Repository.Lock` 返回 `StoredFence`；`PendingFence` 同样携带 optional `AbortClaim`。Repository 增加 `ClaimAbort(ctx, dbtx, operationID, reason, claimedAt)`，只允许 absent claim 的第一次写入或 exact same claim 的零写入重试。`BindEffect`、`ActivateCommitted`、`RecordAborted` 均在 SQL predicate 和 Go validation 两层检查 claim 状态，其中 `RecordAborted` 只能消费已存储且 reason exact 相同的 claim。任何 impossible nullable shape 都返回 finite conflict/injected-failure，而不是降级为无 claim。

## 5. Abort 线性化协议

### 5.1 Authority-bearing 领域 writer

每个会创建、推进或激活 authority effect 的 repository transaction 使用固定锁顺序：

1. 按 operation ID `SELECT ... FOR UPDATE` 锁定 fence。
2. 验证 operation、effect kind、scope kind、scope digest、epoch、sequence 与领域 tuple exact 一致。
3. 验证未 Abort-claimed，并验证该 writer path 对应的 exact phase。
4. 写入或推进领域 effect。
5. 在同一事务提交；provider 调用仍在事务外。

Coverage 对每条 path 还必须冻结其 phase：

- `pre_finalize_effect`：创建/推进 pending effect 或立即生效的 fail-closed effect；fence 必须为 `reserved/fence_pending` 且无 claim。
- `post_finalize_activation`：消费已经验证的 exact committed provider receipt；同一事务先把 DB fence 转为 `committed/active`，再推进 active pointer/serving visibility，且全程无 claim。
- `parent_protected`：只能在已按上述 phase 锁定的 immutable parent intent 下写入，不能自带旁路 authority tuple。

不能把所有 writer 一概要求为 pending：access-granting effect 的最终 activation 必须看到同事务内已经 committed/active 的 DB fence。反过来，pre-finalize writer 也不能接受 committed fence。Coverage phase 与 trigger condition 不一致必须使 schema test 失败。

应用 repository 必须显式遵守 fence-first 顺序。数据库 `BEFORE INSERT/UPDATE/DELETE` trigger 作为不可绕过的防线，再次锁定/验证 exact fence 并拒绝已有 claim。DELETE 使用 `OLD` authority tuple；任何仍为 pending/claimed 或仍被 resolver 用来解释 nonterminal operation 的 evidence 都不可删除。Terminal retention delete 只有在 exact fence 已 terminal、保留期与引用条件全部满足时才允许。直接 SQL、后续新 repository 或遗漏的 workflow 不能绕过 claim。

新增 test-only Go literal `authorityEffectGuardCoverage`（放在 store schema tests，而不是运行时或 YAML 中）必须枚举每一个 authority operation FK/path，并将其分类为：

- 直接 effect writer：必须声明 INSERT/UPDATE/DELETE 中实际可发生的 events，并有 exact guard trigger；
- 由父 intent 间接保护：必须指明父表及不可绕过的 FK/trigger 链；
- 非 effect 元数据：必须给出不影响授权/撤销/active pointer 的理由。

至少覆盖 inventory authority pointer、resource envelope、issuance、grant create/claim、certificate activate/revoke、incident open/resolve、security-fault receipt、recovery session、restore approval、state/root publish intent、desired/recovery state、trust-bundle high-water，以及携带 authority tuple 的 operator audit/state transition 路径。未来 migration 增加任何 authority tuple 时，缺少 coverage 分类或 guard 必须使测试失败。`db/schema/nodecontrol.v1.yaml` 继续只描述 exact catalog objects，不承载这些语义分类。

### 5.2 Coordinator Abort

`Coordinator.Abort` 的有效路径固定为：

1. 验证请求，并读取 DB/provider record 以识别 exact terminal retry 或明显分叉；此阶段不得写入。
2. 开启显式 PostgreSQL `READ COMMITTED` 短事务。
3. 在该事务内按 operation ID 锁定并重读 fence。
4. 通过同一个 `store.DBTX` 调用 transactional resolver。
5. 仅当 fence exact unbound/unclaimed 且 resolver 返回 exact `EffectAbsent` 时写入 `(abort_reason, abort_claimed_at)`；exact 相同 claim 可恢复，reason 不同则冲突。
6. 提交 claim transaction，释放所有数据库锁。
7. 在事务外调用 provider `Abort`。
8. 验证 exact provider receipt 后，在新的短事务中把同一 claim terminalize 为 aborted。

取消、deadline、SQL error 或 resolver error 必须回滚尚未提交的 claim，并保留既有有限 sentinel。Provider 调用期间不持有 PostgreSQL transaction 或 row lock。

### 5.3 线性化证明

- 领域 writer 先锁 fence：Abort 等待；writer 提交后，`READ COMMITTED` resolver 看见领域 effect，claim 被拒绝，provider Abort 不会发生。
- Abort 先锁 fence：claim 与 absent observation 原子提交；后续 writer 锁到 fence 后看见 claim，并在触碰有效领域状态前被拒绝。
- 两者等待同一个持久 row lock，而 resolver 与 claim 使用同一个 DBTX，因此不存在 absent read 与 claim write 之间的跨事务窗口。
- 进程终止不会释放语义权利：已提交 claim 留在 fence 行，未提交 claim 随事务回滚。

### 5.4 Recovery

Recovery 按 operation ID 和 exact tuple 执行：

- claim 已提交、provider 仍 reserved：重试同 reason 的 provider Abort。
- claim 已提交、provider 已 aborted：验证 receipt 后 terminalize DB。
- terminal DB commit response loss：返回 exact stored receipt，零重写。
- provider 显示 aborted、PITR 丢失 DB claim：receipt 已保存 reason；只能在 fence lock 下、transactional resolver exact absent、DB/provider tuple exact 且 unbound 时重建该 exact reason claim，随后 terminalize。若 PITR 连 fence row 一并丢失，pending insert、row lock、resolver 和 claim 必须在一个事务内完成；不能提交一个中间 unclaimed fence，也不能直接写 terminal receipt。
- provider 仍 reserved、PITR 丢失 claim：provider 没有 abort reason，`Recover(operationID)` 不得猜测或选择默认 reason。它可以重建 exact pending fence、解析并 finalize exact committed effect；若 effect absent，则保持 unclaimed pending 并返回 finite unresolved/conflict，只有携带显式 reason 的新 `Abort` 调用才能竞争 claim。
- claim 存在但 provider committed、领域 effect 存在但 provider aborted、reason/digest/scope/epoch/sequence 任一不一致：保持 pending/unavailable 并返回 finite conflict。
- claim pending 本身使 readiness 保持关闭，直到 exact recovery 得到可验证 terminal。

### 5.5 Provider commit 与领域 visibility 的原子收敛

Provider Finalize 已成功后，DB fence terminalization 不能先于领域 active pointer/terminal outcome 单独提交。Coordinator 增加有限、事务绑定的 activator dispatcher：

```go
type TransactionalEffectActivator interface {
    CaptureActivationDecisionEvidence(
        context.Context,
        Receipt,
    ) (ActivationDecisionEvidence, error)
    ActivateAuthorityEffect(
        context.Context,
        store.DBTX,
        Receipt,
        ActivationDecisionEvidence,
    ) error
}
```

`ActivationDecisionEvidence` 是 B01 定义并严格验证的有限内部值，至少包含 exact provider head、按 scope 可选的 committed node/global checkpoint anchor、按 policy 可选的 rollback-resistant trusted-time anchor、有限 `decision_capability`、`evidence_valid_until`，以及这些字段的 domain-separated digest。Trusted-time anchor 只含 canonical trusted instant、provider-identity digest和floor/attestation digest，不接受 wall clock、HTTP Date或普通 caller timestamp。Policy不需要某类证据时使用 exact absent branch，不能填零值 fake。

Checkpoint present branch 携带 strict `AuthorityCheckpointAnchorV1={schema_version,checkpoint_kind,checkpoint_scope_digest,authority_epoch,authority_sequence,receipt_digest}`。`checkpoint_kind` 只有 `node/global`，`checkpoint_scope_digest` 必须逐字等于 operation/commitment 的 canonical node scope digest或 B01 固定的 canonical global scope digest；epoch/sequence 使用 canonical nonempty decimal string，receipt digest 必须是 exact committed provider receipt digest。`checkpoint_digest=SHA-256("talenro.nodecontrol.authority-checkpoint-anchor.v1\x00" || JCS(AuthorityCheckpointAnchorV1))`。Validator 必须从完整 anchor 重算 digest并验证 receipt、provider head、scope与 commitment；只给 digest、错误 node/global scope、较低/未提交 receipt 或跨 operation anchor均拒绝。

Canonical evidence transcript 固定为 `ActivationDecisionEvidenceV1={schema_version,operation_id,commitment_digest,provider_head_digest,checkpoint_kind,checkpoint_scope_digest,checkpoint_digest,trusted_time_kind,trusted_instant,evidence_valid_until,provider_identity_digest,floor_attestation_digest,decision_capability}`。Checkpoint kind 只有 `none/node/global`，trusted-time kind只有 `none/rollback_resistant`，decision capability只有 `may_apply/not_applied_only`；absent branch 的对应值使用 B01 固定 empty digest/empty instant，present branch禁止 empty值。Trusted instant与valid-until使用 UTC `RFC3339Nano` canonical text且禁止非零 offset。`evidence_digest=SHA-256("talenro.nodecontrol.activation-decision-evidence.v1\x00" || JCS(...))`，其余 string/digest encoding 与第 6.1 节一致。Evidence 不能跨 operation、commitment或不同 transcript复用；retry只有在 exact transcript逐字相同时才可复用。

Rollback-resistant evidence capture 建立 post-Finalize 时间/freshness 的外部权威样本，随后对该 exact evidence 的 same-process monotonic admission 单次消费完成线性化；DB COMMIT time不是时间权威。任何 `decision_capability=may_apply` 的 evidence 都必须使用 `trusted_time_kind=rollback_resistant`，并在 capture 时满足 `trusted_instant < evidence_valid_until <= min(trusted-time attestation expiry,effect activation deadline,trusted_instant+5s)`；没有 rollback-resistant trusted time 就不能构造 applied-capable evidence。若 trusted time 已越过 deadline，capability只能是 `not_applied_only`，并仍须满足 `trusted_instant < evidence_valid_until <= min(trusted-time attestation expiry,trusted_instant+5s)`。`trusted_time_kind=none` 时 trusted instant、valid-until、provider identity与floor/attestation digest必须全部使用 canonical empty值，capability也必须是 `not_applied_only`；该 branch 只有在 commitment mode 已是 exact `final_not_applied`，或 exact higher committed checkpoint/receipt已经独立证明 rollback-stable stale时合法。Applied resolution不得使用 capture 时已过期、trusted-time absent或 not-applied-only evidence；trusted-time absent 只允许上述两类 not-applied terminalization。

对 `trusted_time_kind=rollback_resistant`，B01 capture 必须在调用 `TrustedTimeSource`/attestation provider **之前**记录 same-process monotonic `capture_started`；constructor 再在 opaque evidence value 中附带不序列化、不入 canonical digest 的 operation/commitment/evidence-digest-bound single-use monotonic admission token。Token deadline 使用 checked positive duration，固定为 `capture_started + min(evidence_valid_until-trusted_instant,5s)`；负值、零值、overflow或不可表示值均拒绝。外部调用、验证与其余 capture 工作都消耗这个 budget；constructor 完成时 monotonic reading 必须严格小于 deadline，等于或超过均拒绝且不得返回 evidence/token。复制 value 不能复制消费权，反序列化或进程重启不能重建 token。Coordinator 在 activation transaction 已取得全部锁、完成所有重检与写入、即将调用 `Commit` 的紧邻前一步原子消费 token；消费时 monotonic reading同样必须严格小于 deadline，到期、缺失、重复消费或 context 已取消时必须 rollback 整个事务，保持 fence pending，且只在 rollback/commit absence 已确定后重新capture。调用 `Commit` 前成功消费是 freshness admission action；此后 COMMIT response arrival或数据库 durable time不再参与时间判定。

这条 supersession 对基础设计 §§6.2–6.4、§§7.4–7.5 与 §§10.1–10.3 的所有 post-provider/post-signer/post-issuer activation 一致适用：最终短事务仍逐字重检 exact intent、base pointer、authority/identity/inventory/security/key/credential/approval scope与distinct-operator等 immutable或authority-fenced facts，不再次读取 wall clock或在锁内调用 `TrustedTimeSource`，但必须按上段在 COMMIT attempt 前消费 monotonic admission。基础设计 §6.3 的 15 分钟 proposal/approval/HostRemediationEvidence freshness、§7.4 的 intent activation deadline与key/material window、§7.5 的 publication deadline，都在 capture 时进入 exact evidence/commitment，并要求 admission token在 derived valid-until前消费；capture 时已过期只能得到 not-applied，capture后暂停越过窗口也不能 applied。此规则不延长 leaf、metadata、key、desired/recovery payload或 endpoint authorization 的运行时有效期。

Evidence transcript/digest与 final resolution 在同一个 activation transaction 持久化；monotonic token只证明本进程何时获准发出该次 COMMIT，不持久化为时间权威。Context、lock/statement/idle timeout仍用于有界执行，但不能被描述为“证明没有提交”。COMMIT response loss、timeout或crash后，recovery先按 operation ID读取 exact stored resolution：若 evidence/decision digest已原子提交，就验证并接受该已获 admission 的结果；若未提交，才重新capture当前 evidence。不得因进程重启、超时信号或丢失COMMIT响应就覆盖已提交resolution，也不得假定 uncertain commit失败。

该线性化只决定 workflow是否可把候选记录为applied，不延长certificate、key、metadata、desired payload或endpoint authorization的有效期。Guarded serving与B03 request/response-commit authorization始终重新读取当前TrustedTimeSource；因此在DB transaction较晚提交时，已经过期的material仍不可服务。

Coordinator/registered production handler 在 provider Finalize/Inspect 之后、开启 activation transaction 之前捕获该 evidence；业务 API caller不能提交或覆盖它。Capture 期间不持有 DB transaction/lock。Evidence 缺失、格式错误、低于当前 receipt sequence、checkpoint scope错误、trusted-time provider identity错误或 digest tamper 均保持 pending并 fail closed。

`Finalize` 或 provider-committed recovery 在 provider call/Inspect 和 activation evidence capture 完成且所有数据库锁均已释放后，开启一个新的短事务：

1. 锁定并重读 `StoredFence`，要求 exact binding、无 claim、provider committed receipt exact。
2. 通过 `Repository.ActivateCommitted` 把 DB fence 变为 `committed/active`。
3. 在同一个 DBTX 将 Coordinator 捕获的 exact `ActivationDecisionEvidence` 交给 finite activator，按第 6.1 节 commitment mode 得到 terminal resolution：`conditional_apply` 可以成为 applied，也可以在 rollback-stable stale 证据下成为 not-applied；`final_not_applied` 只能验证并落相同 tombstone。需要与 resolution 原子出现的 audit/outbox 也在此完成。
4. 若 evidence 使用 trusted time，Coordinator 在所有 DB 写入完成后原子消费 exact monotonic admission token；失败则 rollback，成功则立即调用一次 `Commit` 提交全部变化。Trusted-time absent 的 exact final-not-applied 或 higher-authority not-applied branch不创建该 token。

Activator 不能调用 provider、trusted-time source、signer、issuer 或其他外部 dependency，只能验证传入 evidence 并读取 caller-owned DBTX。永久 stale 只有在 evidence 中更高且 committed 的 external authority receipt/checkpoint与 exact DB fact一致，或 rollback-resistant trusted time 已越过不可逆 deadline 时，才作为一次成功的 not-applied terminal resolution 与 fence 同事务提交；仅有更高 pending reservation不够。Missing activator/evidence、DB/provider 尚未追平、非 authority-fenced mutable fact、tuple mismatch、SQL error 或 crash 则回滚整个事务，使 DB fence 继续 pending并按 operation ID重试。Exact terminal retry 只验证 stored resolution，不能再次推进 generation、pointer、audit 或 outbox。

这项 seam 取代 B02/B03 计划中“Coordinator 先单独 ActivateCommitted，领域服务随后另开 activation transaction”的两事务描述。获批后的实施计划必须修改 `Coordinator` 构造/API 和各领域 activator 所有权，不能仅靠 service-level best-effort recovery 弥补中间窗口。

## 6. Transactional Effect Resolver

普通 readiness/recovery 可以保留现有只读 resolver，但 Abort claim 必须依赖显式事务接口：

```go
type TransactionalEffectResolver interface {
    ResolveAuthorityEffectForUpdate(
        context.Context,
        store.DBTX,
        uuid.UUID,
    ) (ResolvedEffect, error)
}
```

Coordinator 构造时必须获得该能力；不得在运行时把缺失能力解释为 `EffectAbsent`。Dispatcher 对 `EffectKind` 做有限、穷尽分派，每个 resolver 读取真实领域表并验证 operation/kind/scope/scope-digest/effect-digest 绑定。零行、多行、损坏行、跨 operation 行、未知状态和未知 kind 一律 fail closed。

Resolver state 描述 operation 的 authority lifecycle，而不直接等同于领域表的 `status` 文本：

- `EffectAbsent`：没有任何领域事实被物化；只有这一状态且 fence unbound/unclaimed 时可 Abort。
- `EffectPrepared`：存在尚未形成 immutable final outcome 的 intent；不得 Abort，workflow 必须继续或写 fail-closed outcome。
- `EffectCommitted`：immutable commitment 已落库并有 exact effect digest，既可以是待确定执行的 conditional candidate，也可以是 failed/superseded/deadline-expired 的 final not-applied tombstone；该状态必须 Bind/Finalize。
- `EffectTerminal`：领域 visibility 或 fail-closed tombstone 已与 exact terminal fence receipt 原子收敛；只能用于 exact retry/verification，不能重新 Abort/Finalize。

因此 signer/issuer/provider dependency failure、deadline 或 supersession 若在 provider Finalize 前已经确定，必须先在 fence-first transaction 中形成 immutable final-not-applied commitment，并由 `EffectCommitted` 路径 finalize。`AbortActivationDeadlineExpired` 等 Abort reason 只适用于 deadline 到达前尚未物化任何领域事实的 reservation；不能借 reason 名称绕过 prepared-work 禁止 Abort 的规则。

### 6.1 Domain-separated effect commitment 与单调 resolution

成功候选与已确定的 fail-closed tombstone 不能复用原 intent/payload digest。B01 在 `internal/nodecontrol/authority` 独占 canonical constructor、strict validator、JCS encoder 与 digest helper；B02/B03 只能调用该 helper，不能各自重算 transcript。

每个可 Finalize 的 `EffectCommitted` 必须由 exact domain record 提供一个 canonical commitment：

```text
AuthorityEffectCommitmentV1 = {
  schema_version,
  operation_id,
  effect_kind,
  scope_kind,
  scope_digest,
  authority_epoch,
  authority_sequence,
  base_effect_digest,
  commitment_mode,
  reason_code,
  activation_policy_version,
  activation_inputs_digest
}
```

- `commitment_mode` 只有 `conditional_apply` 与 `final_not_applied`。
- `conditional_apply` 的 `reason_code=none`，并绑定非 `none` 的 `activation_policy_version` 与 exact `activation_inputs_digest`；provider receipt 承诺的是候选与确定性激活规则，不宣称它已经 active。
- `final_not_applied` 使用该 effect kind 注册的有限 reason，例如 `failed`、`superseded`、`activation_deadline_expired`、`validation_rejected`；其 policy version固定为 `none`，inputs digest使用 B01 固定的 domain-separated empty-input digest。未知或跨 kind reason拒绝。
- `base_effect_digest` 绑定原本要执行的 immutable payload/action；commitment 再绑定 operation、authority tuple、mode、reason与 activation inputs。
- `schema_version`、epoch、sequence 与 policy version使用 canonical nonempty decimal strings；UUID为 lowercase canonical text，digest为 lowercase 64-hex，enum为 closed lowercase token。所有字段 required，禁止 JSON number、null、额外字段和动态错误文本。
- `effect_digest = SHA-256("talenro.nodecontrol.authority-effect-commitment.v1\x00" || JCS(AuthorityEffectCommitmentV1))`；transcript 不含 payload、certificate material或credential。

相同 base payload 的 conditional 与 final-not-applied commitment 必须得到不同 digest；不同 failure reason、policy version或activation inputs也必须不同。Commitment fields在 Bind前一次形成，之后不可改。

`conditional_apply` 的 post-Finalize resolution 只能读取三类 rollback-stable inputs：commitment中已捕获且不可变的 facts、由更高 provider authority sequence/checkpoint 锚定的状态、以及 rollback-resistant trusted time。任何会决定安全性但只存在于可回放 PostgreSQL 的 mutable fact，必须在 provider Finalize 前变成 final-not-applied、改为 authority-fenced input，或从 post-Finalize policy移除。

Activator 将 final disposition与有限 reason写回该 effect kind 的 existing domain intent/record，并保存可由 B01 helper重算的 resolution：

```text
AuthorityEffectResolutionV1 = {
  schema_version,
  commitment_digest,
  disposition,
  reason_code,
  decision_anchor_kind,
  decision_anchor_digest
}
```

`disposition` 只有 `applied/not_applied`。Applied 的 reason 固定 `none`；not-applied reason 使用 per-kind有限 enum。`decision_anchor_kind` 只有 `final_commitment`、`exact_capture`、`higher_authority`、`trusted_time`；digest分别绑定 exact commitment、captured-input set、第 5.5 节 strict `checkpoint_digest`/committed receipt set或trusted-time evidence。`resolution_digest = SHA-256("talenro.nodecontrol.authority-effect-resolution.v1\x00" || JCS(AuthorityEffectResolutionV1))`，使用与 commitment相同的 strict string/digest encoding。

`authorityEffectOutcomeCoverage` test-only registry 为每个 effect kind唯一声明 domain table/key、commitment columns、terminal disposition/reason/decision-anchor columns、resolution digest column、resolver和activator；不新增 generic outcome table，也不把领域事实搬进 fence。若现有领域表缺少 required字段，migration在该表增加 exact typed columns并更新 literal catalog manifest。

Resolution 规则必须单调不升权：重放可以因更高 authority sequence或更晚 trusted time把原先 applied候选降为 not-applied，但任何已经以 rollback-stable证据得到 not-applied 的候选永远不能在 PITR/retry后升级为 applied。DB 落后于 provider、决定性证据缺失或只有 DB-local stale 时，activator 回滚并保持 readiness关闭，不能猜测 terminal outcome。

PITR 若恢复到 commitment 写入前的 prepared intent，而 provider 已 committed一个 final-not-applied digest，resolver必须报告 prepared/mismatch；不得重造 tombstone或改用 conditional digest。若 provider committed conditional digest，restored DB只有在 exact commitment存在、全部后续 authority records已追平且同一 policy按 rollback-stable inputs求值后才可 resolution。否则保持 unavailable并进入 authority-restore/数据恢复流程。

当前 schema 中 `operator_authorizer_change` 缺少完整 durable authorizer record/workflow，`trust_bundle_publish` 只有 active high-water、缺少完整 pending immutable effect workflow。本修订不虚构这些 workflow；在对应后续设计落地前，dispatcher 对缺失能力返回固定 unsupported/error，Abort 不得据此取得 claim。

## 7. Certificate 与 Desired Guarded Reader

### 7.1 内部生产能力边界

新增 `internal/nodecontrol/serving` 和 `db/queries/nodecontrol_serving.sql`。该包是可由未来 B02/B03 runtime 组合的真实内部 reader，但本批次不把它接入 HTTP handler、listener 或 `cmd/control-api`。

Serving reader 不得分别接收一个任意 DBTX 和一个独立 availability，否则可把数据库 A 的 ready head 用来授权数据库 B 的 query。最小依赖必须是由同一个 `PostgresRepository`/Coordinator 生产、不可拆分的 bound source：

```go
type BoundAuthorityReadSource interface {
    WithConsistentReadyRead(
        context.Context,
        func(context.Context, store.DBTX) error,
    ) error
}
```

Production constructor 只从 Coordinator 实际使用的同一个 `PostgresRepository` 创建该 source。Source 从该 repository 的 pool 获取一个物理连接，pre-check 的全部 DB reads、serving query 和 post-check 的全部 DB reads 都使用这一个连接；Coordinator 为此提供 internal DBTX-bound readiness path。Provider calls 发生在 SQL statement 之外，期间不持有显式数据库 transaction 或 row lock。连接失效、failover 或 database identity 改变均使本次读取失败关闭。

Serving package 不能自行组合 availability 与 database。构造时错配 pool/repository、同 cluster 不同 database 或 test fake database 必须被拒绝或在 integration test 中确定性失败。

每次读取固定执行：

1. Bound source 在固定连接上执行 DBTX-bound `CheckReady`，要求 `Ready=true`，保存完整 provider/database head snapshot。
2. 通过同一个连接执行一条 exact certificate 或 desired SQL query。
3. Bound source 在同一连接上再次执行 DBTX-bound `CheckReady`。
4. 要求第二次仍 ready，且两次 provider head 与 database head 逐字段 exact 相等。
5. `WithConsistentReadyRead` 只有第 4 步成功后才返回 nil；serving reader 在此之前不暴露 callback 的临时结果。失败时清零/丢弃内部结果并返回固定 `ErrAuthorityUnavailable`。

Readiness reason 与 head 只用于内部判断，不进入 public error、日志高基数 label 或 response。取消/deadline 保留既有 cancellation sentinel；所有 authority mismatch 对 serving caller 统一为 value-free `ErrAuthorityUnavailable`。

Query 先返回 no-row、corrupt-row 或 domain validation error时，bound source仍必须执行 post-check；只有 context 已取消或物理连接已不可继续使用时可以跳过。错误优先级固定为：cancellation/deadline、authority/post-check unavailable、随后才是 post-check GREEN 下的有限 domain not-found/corrupt classification。连接终止视为 authority unavailable。临时 query result与临时 error都不得在 post-check前写入 caller-visible对象、日志或 metrics。

### 7.2 Exact certificate query

Certificate reader 必须同时匹配：

- exact node ID、issuer ID 和 unsigned serial bytes；
- exact leaf DER digest 与 public-key digest；
- inventory 当前 identity epoch 与 lineage；
- certificate 的允许有限状态和相同 node/epoch/lineage；
- exact activation fence operation/epoch/sequence；
- fence kind=`certificate_activate`、scope=`node`、scope digest exact，且 provider/visibility 为 `committed/active`。

不得只按 serial、leaf row 或 partial index 命中授权。证书时间、endpoint policy 与 trusted-time 判定仍由 B03 authorizer 完成；本 reader 只提供经过 authority gate 的 exact durable facts。

### 7.3 Exact desired query

Desired reader 必须从 `node_inventory.active_desired_generation` 出发，连接：

- exact `(node_id, generation)` desired row；
- exact active `node_state_signing_intents` desired intent；
- intent、payload、generation、digest、key 与 authority binding；
- exact committed/active `desired_activate` fence。

禁止 `MAX(generation)`、最近一行、orphan envelope、pending intent 或缺失 inventory pointer 的 fallback。签名/有效期/trusted-time 与 endpoint response-commit authorization 仍由 B02/B03 完成。

### 7.4 PITR 证明

PITR integration 必须创建真实 inventory、certificate、desired state、signing intent、fence 及所有 required dependency rows。测试先证明未恢复数据库可由 production reader 读取 exact fixture，再从 revoke/disable 前备份恢复，并保持外部 provider 位于更高 head。

恢复后测试必须同时证明：

- 原始 SQL 仍能看见旧 certificate 与 desired rows，确认 fixture 确实存在；
- 两个 production reader 都返回 exact `ErrAuthorityUnavailable`；
- certificate material、desired canonical bytes 和派生授权结构均为空；
- authority 恢复到 exact ready head 后，control fixture 可再次读取。

## 8. Goose 00006 修复

原地扩展 `00006_nodecontrol.sql` 的前提是 C1.2 migration 00006 尚未发布、未进入任何需要升级兼容的 durable/shared environment。实施计划的第一项只读 preflight 必须核对由 release/deployment operator持有的权威环境清单：它逐一列出所有 durable/shared PostgreSQL environment、当前 migration version、00006 apply timestamp/absence、deployment ID与签名/attestation digest，并由该 operator确认范围完整。若已有 C1.2 completion/deployment manifest，也必须逐字交叉核对。

仓库当前没有定义可验证的 release/deployment ledger；“代码库中没找到部署记录”、本地 `goose status` 或测试数据库为空都不能证明未发布。若实施时仍无法取得上述完整、受签环境清单，默认前提不成立：只允许修复 00006 的 parser annotations 以保证 fresh install，并必须另行设计 additive semantic migration/upgrade path，不能原地加入 claim/outcome schema。若清单证明任一 durable environment 已应用 00006，也执行同一 additive 路径。

在该 pre-release 前提成立时，`00006_nodecontrol.sql` 保持单一 C1.2 schema migration，不新建补丁 migration。两个 `LANGUAGE sql` 函数无需 wrapping；现有 17 个 `LANGUAGE plpgsql` 函数以及本修订新增的任何 PL/pgSQL guard function，均分别使用：

```sql
-- +goose StatementBegin
CREATE FUNCTION ...
...
$$;
-- +goose StatementEnd
```

不得用一个大 block 包裹整份 migration。Annotations 位于函数定义之外，不改变 `pg_get_functiondef`、catalog semantics 或 manifest hash source。

权威 migration gate 必须调用 repository pinned Goose，而不是测试中切片 SQL 后直接 `pool.Exec`：

1. 创建 fresh、run-owned PostgreSQL database。
2. 对完整 migrations 执行普通 `up`，验证版本 6、manifest 和 catalog digest。
3. 执行 `down-to 5`，验证 nodecontrol schema 消失。
4. 再执行普通 `up`，验证版本 6 和相同 catalog digest。
5. 任何命令输出不得泄露数据库 URL 或 credential；只清理 exact run-owned database/container/volume。

静态测试还必须证明每个 dollar-quoted PL/pgSQL function 恰好位于一个匹配的 StatementBegin/StatementEnd block 内。

## 9. 实施边界与所有权

本修订批准 B01 的最小扩展：

- 修改 migration、literal schema manifest、authority sqlc queries/generated store、repository、Coordinator、transactional resolver/activator dispatcher、canonical commitment/resolution helper、readiness 与相关测试。
- 新增内部 `nodecontrol/serving` reader、serving sqlc query 和真实 certificate/desired PITR fixtures。
- 合入当前冻结的 Task 7 五项 in-scope 审查修复：Abort terminal retry、provider-aborted exact effect validation、真实 PostgreSQL resolver crash matrix、empty-head epoch normalization、pending `ErrNotFound` classification。
- 修复 pinned Goose 对 migration 00006 的正常 up/down/up 支持。

明确延后：

- production `authority.Provider` adapter、credential/config、provider lifecycle；
- operator mutation、state signer、issuer、identity/recovery service；
- bootstrap/agent/operator handler、listener、response-commit authorization；
- `cmd/control-api` runtime composition 与 production readiness wiring。

这些仍由 B02/B03 及最终 B11 authority-fence gate 完成。`DeterministicProvider` 仍只能用于 local/test，不能进入 production composition。

获批后的计划修订必须消除重复所有权：

| 边界 | B01 amendment | B02/B03 保留责任 |
| --- | --- | --- |
| desired read | `nodecontrol_serving.sql` 与 authority 双检 reader | B02 保留 intent/allocator/activation/write workflow；服务层消费 guarded reader，不再另写可绕过的 active-load query |
| certificate read | `nodecontrol_serving.sql` 与 authority 双检 reader | B03 保留 issuance/rotation/revoke、trusted-time 与 endpoint policy；authorizer 消费 guarded certificate facts |
| effect resolution/activation | transactional resolver、transactional activator、claim protocol与缺失能力 fail-closed | B02/B03 各领域 resolver/activator 接受 caller-owned `store.DBTX` 并注册到同一个 finite dispatcher；领域服务不另开旁路 visibility transaction |
| runtime | 无 | B03 Task 10 构造 production provider、dispatcher、shared Coordinator、services、listeners 和 lifecycle |

因此 `db/queries/nodecontrol_state.sql` 与 `db/queries/nodecontrol_identity.sql` 继续拥有各自 workflow mutation/activation 查询，但不得复制或绕过 B01 的 authority-gated serving 查询。

Writing-plans 阶段必须把下列 normative 文档列为显式修改目标，缺一项就不能开始实现：

- Suite index `2026-08-23-node-pop-control-plane.md`：repository map 增加 `db/queries/nodecontrol_serving.sql`、generated artifact 与 `internal/nodecontrol/serving`；frozen cross-plan interfaces/traceability 登记 bound read source、commitment/resolution helper、transactional resolver/activator和Coordinator evidence capture。
- B01 Plan 01：替换原只读 resolver/独立 fence activation API，加入 claim、commitment、evidence、atomic activator、guard coverage、Goose与真实 serving/PITR tasks。
- B02 Plan 02：删除 `StateRepository.LoadActive` 作为 production serving path；workflow 若需内部读取只能使用明确非 serving 的 intent/capture query，所有 desired payload serving消费 guarded reader。改写 signer stale/failure与 transactional resolver/activator步骤。
- B03 Plan 03：删除 `Repository.LookupExactCertificate` 作为 authorizer serving path；issuance workflow 如需内部 lookup必须使用不同的非授权名称/返回类型，mTLS authorizer只消费 guarded certificate facts。Task 10注册同一个 production dispatcher/source。
- Plan 08 verifier/supply-chain 与 Plan 09 operations/completion：evidence schema/builder/validator统一消费下述 `C12SpecSetV1` digest。

这些修改不是可延后的 implementation note；原 plans 中的旁路 query/API 在修订完成前仍具规范效力，因此 writing-plans 输出必须逐项删除或重写。

原规格 hash 继续标识 2026-08-23 base spec。本文获批后，受影响实施计划必须同时列出 base spec 与本 amendment 的 approved content hash；不得静默替换旧 hash，也不得把旧计划描述成已经包含本修订。

现有 evidence 只有一个 `SpecDigest`，因此它统一绑定 canonical spec set，而不是任取其中一个文件：

```text
C12SpecSetV1 = {
  schema_version: "1",
  base: {path, sha256},
  amendments: [{path, sha256}, ...]
}

SpecDigest = SHA-256(
  "talenro.c12.spec-set.v1\x00" || JCS(C12SpecSetV1)
)
```

Path 使用 repository-relative、forward-slash、UTF-8 canonical text；SHA-256 使用 lowercase 64-hex；amendments 按 path bytewise 升序且不得重复。Plan 08/B10 与 Plan 09/B11 的 evidence builder、validator和 completion manifest 必须消费这个单一 spec-set digest。本文最终批准前不生成 approved amendment hash或 `SpecDigest`；内容或状态变化后必须重新计算，旧 evidence不得复用。

Approved-set 的唯一清单为 tracked strict JSON `docs/superpowers/specs/c12-spec-set.v1.json`；它不包含自身 hash，恰好含一个 base和全部 approved amendments。Plan 08 在 `internal/c12evidence/specset.go` 独占 strict decode、path confinement、file hashing、JCS constructor与`SpecDigest` helper；其他 package/脚本只能调用该 helper。Approval流程必须在同一受审查变更中把新 amendment状态改为已批准并加入清单；builder要求每个listed Markdown状态为已批准、每个digest exact，且拒绝missing file、extra/duplicate entry、wrong base、非canonical order/path和hash drift。Plan 09/B11只验证/消费，不另写 builder。

## 10. TDD 与验收矩阵

实现必须严格 RED -> GREEN -> REFACTOR，并至少覆盖：

### 10.1 Unit/contract

- claim/binding/terminal 全部合法状态与每个非法组合；
- exact claim retry、reason conflict、exact terminal retry零重写；
- Finalize/Bind/Activate 对 claimed row 的拒绝；
- transactional dispatcher 的未知 kind、零行、多行、tuple/digest tamper；
- prepared failure 到 immutable fail-closed `EffectCommitted` outcome 的有限转换；该路径 Finalize 而不调用 provider Abort；
- `AuthorityEffectCommitmentV1`/`AuthorityEffectResolutionV1` 的 independent literal JCS/digest goldens，证明相同 base effect 的 conditional/final-not-applied、不同 reason/policy/input/decision anchor 均不相等，并对每个字段做 mutation sensitivity；
- `AuthorityCheckpointAnchorV1` 的 node/global scope、epoch/sequence/receipt goldens与 domain separation；`ActivationDecisionEvidenceV1` 的 none/node/global、trusted-time present/absent、may-apply/not-applied-only goldens，以及 missing、tampered、wrong-scope、lower/uncommitted sequence、wrong-provider-identity、expired/not-applied-only misuse拒绝；trusted-time absent永不能 applied，且只接受 exact final-not-applied commitment或higher committed authority anchor；capture调用当刻 attestation/valid-until 已过期时不能产生 may-apply；fake TrustedTimeSource/attestation response delay在budget just-under时保留剩余窗口、equal/over时不返回 token，证明外部调用不能重置窗口；fake monotonic pause在 BEGIN 前、锁等待中或 DB 写入后跨过admission deadline都使事务零提交并重新capture，token copy/reuse/restart均拒绝；token成功消费后发生的 in-flight COMMIT response loss必须先按stored evidence/resolution判定，已原子提交即接受、未提交才重新capture；
- transactional activator 的 missing-kind、rollback-stable permanent stale terminalization、DB-local/transient stale rollback 与 exact idempotent replay；
- provider bound-aborted canonical vector保持不变，但 Coordinator 将其分类为 mismatch/unavailable；
- guarded reader 的 pre-check、post-check、head changed、result clearing、callback NotFound/corrupt后仍post-check，以及 cancellation > authority > domain错误优先级；
- schema coverage manifest 对遗漏/重复/错误 indirect-parent、错误 phase 或缺失 DELETE event 分类的 mutation sensitivity。

### 10.2 PostgreSQL concurrency/crash

- 两个独立连接模拟 writer-first 与 Abort-first，不使用 sleep 猜测锁竞争；通过可观察 lock/wait seam 确认顺序。
- claim commit 前 crash、claim commit response loss、provider Abort 前失败、provider response loss、terminal DB response loss。
- concurrent Finalize/Abort 只能得到一个合法 terminal，不能产生第二次 provider mutation。
- provider Finalize 后、fence/domain activation transaction 的每个 callback/commit/response-loss crash point均由 operation ID 重放；fence 与领域 visibility 不能半提交。
- provider committed final-not-applied commitment 后，PITR 分别恢复到 commitment 前 prepared row、commitment 后但 fence terminalization 前、terminalization 后；只有后两者在 exact transcript 存在时可恢复，第一种必须持续 unavailable且零 activation。
- provider committed conditional commitment 后，以更高 authority sequence和trusted-time expiry分别造成永久 stale；PITR/retry最多保持或降低权限，永远不能把已锚定 not-applied resolution升级为 applied。
- SQL effect row DELETE 在 pending/claimed 时被 `OLD` fence guard拒绝；测试绕过 guard 后的删除、字段篡改、duplicate match 和 cross-bound tuple仍全部保持 fail closed。
- readiness 检查全部 pending row，并正确处理 claimed pending。

### 10.3 Serving/PITR/Goose

- control database 的真实 certificate/desired exact reader GREEN。
- pre-revocation/disable PITR 后 raw rows 存在但两个 reader均不可服务。
- pinned Goose fresh `up -> down-to 5 -> up` 与 catalog digest稳定。
- migration annotation 静态 coverage完整。
- migration preflight 在完整受签 environment inventory证明未发布时选择 in-place semantic path；ledger缺失/不完整或任一 version 6 environment存在时选择 additive path，禁止 repo-absence fallback。

### 10.4 Whole-branch gate

- pinned generator 连续运行两次且 tracked generated tree无 drift；
- focused、package、integration、race、vet、format/diff checks通过；
- uninterrupted `go test ./... -count=1` 通过；
- fresh PostgreSQL/Docker authoritative run通过，所有 exact-owned资源清理并二次确认不存在；
- 完成独立代码/安全审查，Critical/Important findings 归零。

## 11. 完成条件

只有同时满足下列条件，Task 7 修复和扩展才可标记 complete：

1. Abort 与所有已实现 authority-bearing writer 共享持久 fence-row 线性化点。
2. 任何跨进程 absent-to-effect 竞态都由确定性 PostgreSQL concurrency test 证明只有一方成功。
3. Claim、provider receipt、DB terminal 与 recovery 在每个 crash/response-loss 点保持 exact 幂等。
4. Provider commitment、领域 terminal resolution、fence visibility、audit/outbox 在同一 DBTX 收敛；PITR/retry 对 conditional effect 只能保持或降低权限。
5. 真实 certificate/desired production-capable reader 在 PITR mismatch 下返回无载荷的 `ErrAuthorityUnavailable`。
6. 普通 pinned Goose 可以完整应用、回退并重应用 migration 00006。
7. 没有 fake provider 或未实现 runtime 被描述为 production wiring。
8. 后续 B02/B03 计划明确消费本 amendment，而不是复制或绕过 guarded reader、transactional resolver/activator 与 canonical commitment/resolution helper。
9. Plan 08/B10 与 Plan 09/B11 evidence 使用 canonical `C12SpecSetV1` 的单一 `SpecDigest` 绑定 base spec和全部 approved amendments。
