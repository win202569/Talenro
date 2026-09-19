# C12 Controlled PITR Test Interface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让旧 authority PITR 回归通过执行器拥有的真实恢复实例，证明恢复到实际提交 A 后，外部 provider 已到 B 时精确失败关闭，同一候选库与 A-provider 对照真实就绪。

**Architecture:** integration-only controller 签发窄 DBTX/事务能力，固定 SQL 策略限制用途，查询结果在返回前受限物化。真实 activation 事务内写 marker，事务结束后独立观察普通 COMMIT；保留 A 为恢复目标、B 为崩溃归档边界。authority 同包测试 bridge 连接既有私有事务 seam，不修改生产 Coordinator。

**Tech Stack:** Go、现有 pgx/v5 与 sqlc、PostgreSQL 18.4、现有 PowerShell C12 runner、integration build tag；不增加依赖。

**Spec:** [已批准的受控 PITR 附录](../specs/2026-09-19-c12-controlled-pitr-test-interface-design.md)，原文提交 `c23a64c1f2bcc45f24493828deb753795f60e7fb`，用户 2026-09-19 批准。另读 [B01](2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md) Tasks 8–9、[执行架构](../specs/2026-09-04-c12-secure-execution-architecture-design.md)、附录引用的 Abort/serving 与 canonical/dispatcher 约束。

## Global Constraints

- 实施目录：`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；分支 `codex/c12-b01-task9-coordinator`。不在主工作区 main 实现，不合并 main。
- 原批准前约束：计划审阅及执行方式确认前，不修改实现、不执行测试或 Docker。用户已于 2026-09-19 以“是”批准本计划及逐任务子代理执行；实际状态与未通过 gate 见文末执行记录。
- 生产 wire schema、canonical transcript、签名角色、数据库迁移及 `c12-spec-set.v1.json` 不变。
- 不实现 Task 10 serving、Task 15/18 exact-epoch/runtime/lease/staging readiness、B03 decoder/provider 或完全丢失 fence 的 ProtocolActivationID 修复。
- 最大 64 个 observation；单个 run 只选定一个 recovery cut；candidate 上限 8 个。无驱逐、无跨 controller/run/database/generation 转让。
- 每次访问/观察最长 30 秒，取调用 context、30 秒、group 剩余绝对期限的最早值。每次访问/观察最多 65,536 行、64 MiB；单行最多 1 MiB。包含结果元数据与复制成本，不能靠 SQL 行数提示替代客户端限额。
- 一个 callback 最多一个活动 READ COMMITTED 事务、最多两个私有连接；每个 controller 同时最多一个访问 callback。同步 callback 不响应取消时，由既有 runner 进程截止兜底，不宣称 Go 强制终止。
- admission consume → 唯一真实 driver Commit 之间只有本地一次性状态转移。无 SQL、marker、日志/WAL、时间查询、观察或 pre-Commit hook；结果丢失只在 driver 返回后注入。
- 候选库允许固定 SELECT 及必要 FOR UPDATE；不设置 blanket read-only transaction，不允许领域 DML/DDL/COPY/角色切换/任意函数，不给 SUPERUSER 或 GRANT ALL。
- 新 observer 支持普通 COMMIT；既有私有 COMMIT PREPARED 测试保留。部分 destructive drain 的结果不确定时 poison 通道；不得重新 Open/retry drain 伪装可恢复。
- Docker 资源只归既有 runner 清理；authority 不增加 cleanup API。既有 testinfra 私有清理故障测试原样保留，不把其权限暴露给消费者。
- `.superpowers`、`.cache`、`.task19-go` 不暂存、不删除、不作测试根。所有手工文件编辑用 apply_patch，只暂存明确列出的本任务文件。
- 不扩大 runner 安全预算、环境白名单、构建输入或清理范围；物理 gate 用原定 5m test timeout。当前主机缺 Docker，不能因此声称物理验收通过。

## Review Focus

1. 读结果逃出 callback、未读完 Rows、QueryRow 延迟 Scan：必须已关闭 driver cursor、无底层 Conn、失效后无 I/O。Task 1 的 `eager_rows_and_expired_views`。
2. 真实 Commit 已成功但响应丢失，或 destructive drain 只读一半：前者同 handle 可观察，后者不得重新生成 cut。Task 3 的 `commit_response_lost`、`partial_drain_poison`。
3. A/B 共处一次 slot drain、XID wraparound、同 marker 出现两次：不能丢 A、混块或用 full-XID 猜提交。Task 3 的 `multi_result_xid_wrap`、`ambiguous_marker`。
4. 候选最小权限拒绝行锁，掩盖成 database_unavailable：必须在同一个真实恢复库跑 provider-B 精确拒绝与 provider-A 精确 ready。Tasks 2、5 的参数/权限负例及物理双对照。
5. 有效句柄被复制/改 DTO/跨 controller 使用、或 callback 与 crash 并发：副作用前拒绝，不能烧掉另一个合法句柄或生成未登记资源。Tasks 1、3、4 的 identity/capacity/concurrency 测试。

---

## File Map and Dependency Order

所有下面新增 Go 文件第一行均为 `//go:build integration`，除明确列出的普通静态 guard。无需新增 store/authority 生产 `.go` 文件。

| Task | 文件 | 唯一责任 |
| --- | --- | --- |
| 1 | 新 `internal/testinfra/c12_authority_access_integration.go` | 窄能力、私有 lease/事务、错误分类与回收 |
| 1 | 新 `internal/testinfra/c12_authority_rows_integration.go` | 受限物化及无连接 Row/Rows 视图 |
| 1 | 新 `internal/testinfra/c12_authority_access_integration_test.go` | fake driver 生命周期/无逃逸/Commit 次数 |
| 2 | 新 `internal/testinfra/c12_authority_sql_policy_integration.go` | 编译期 SQL 字面量、参数/fixture ledger、候选精确 grants |
| 2 | 新 `internal/testinfra/c12_authority_sql_policy_integration_test.go` | SQL/类型/权限/静态漂移行为 |
| 3 | 新 `internal/testinfra/c12_authority_observation_integration.go` | marker、完整事务解码、观察登记与 poison |
| 3 | 新 `internal/testinfra/c12_authority_observation_integration_test.go` | 观察 RED/GREEN 与不确定结果 |
| 4 | 新 `internal/testinfra/c12_authority_cut_integration.go` | 不透明 cut selection、retained identity 检查 |
| 4 | 新 `internal/testinfra/c12_authority_cut_integration_test.go` | A/B 分离、candidate 与顺序负例 |
| 1–4 | 改 `internal/testinfra/c12_authority_pitr_integration.go` | 接入新状态；保留既有资源执行及兼容路径 |
| 4、6 | 改 `internal/testinfra/c12_dependencies_integration_test.go` | 既有 profile/prepared/ownership 测试兼容断言 |
| 5 | 新 `internal/nodecontrol/authority/authority_pitr_bridge_integration_test.go` | 同包私有事务 bridge、绑定 activator、fixture/provider 对照 |
| 5 | 改 `internal/nodecontrol/authority/authority_crash_integration_test.go` | 仅抽取窄 DBTX seed/domain helper、resolver 的 DBTX 依赖 |
| 5 | 改 `internal/nodecontrol/authority/authority_pitr_integration_test.go` | 替换自管 Docker harness，真实 P/A/B 恢复流程 |
| 6 | 改 `internal/nodecontrol/authority/v7_migration_grant_test.go` | 普通构建排除测试 bridge/能力 |
| 6 | 改 `internal/testinfra/c12_integration_manifest_test.go` | 固定 API、构建标签、runner 注册及无逃逸静态 guard |
| 6 | 改 `scripts/run-c12-integration.ps1` | 仅登记精确 PITR-only 顶层测试并拒绝同组多 controller 消费者 |
| 6 | 改 `docs/roadmap/current-status.md`、本计划 | 真实证据、未通过项、物理覆盖变化 |

依赖：1 → 2 → 3 → 4 → 5 → 6。Task 1 的 deny-all 策略在 Task 2 打开固定集合。代码顺序执行；独立审查可以在每个提交后进行，不能并行修改共同 controller 状态文件。

## Frozen Public Interfaces

下列签名是本计划唯一新增消费者接口；不要临时增加 raw connection、SQL 注册、重开、cleanup、任意 fixture 配置或 caller LSN API。

```go
type C12AuthorityAccess interface {
    store.DBTX
    Begin(context.Context) (C12AuthorityTransaction, error)
}

type C12AuthorityTransaction interface {
    store.DBTX
    Commit(context.Context) error
    Rollback(context.Context) error
}

type C12AuthorityPITRObservation struct { state *c12ObservationState }
type C12AuthorityPITRObservedCommit struct { state *c12ObservedCommitState }
type C12AuthorityPITRCutSelection struct { state *c12CutSelectionState }

func (c C12AuthorityPITRController) WithPrimaryAuthorityAccess(
    context.Context, func(C12AuthorityAccess) error) error
func (c C12AuthorityPITRController) WithCandidateAuthorityAccess(
    context.Context, C12AuthorityPITRCandidate, func(C12AuthorityAccess) error) error
func (c C12AuthorityPITRController) NewCommitObservation() (C12AuthorityPITRObservation, error)
func (c C12AuthorityPITRController) BindCommitObservation(
    context.Context, C12AuthorityPITRObservation, C12AuthorityTransaction) error
func (c C12AuthorityPITRController) ObserveCommit(
    context.Context, C12AuthorityPITRObservation) (C12AuthorityPITRObservedCommit, C12AuthorityPITRCommit, error)
func (c C12AuthorityPITRController) SelectRecoveryCut(
    C12AuthorityPITRBaseBackup, C12AuthorityPITRObservedCommit) (C12AuthorityPITRCutSelection, error)
func (c C12AuthorityPITRController) CrashPrimaryAtCut(
    context.Context, C12AuthorityPITRCutSelection, C12AuthorityPITRObservedCommit) (C12AuthorityPITRCut, error)
```

`C12AuthorityPITRCommit` 继续是既有纯值诊断 DTO；新路径不能仅凭它授权。普通 observation 返回 `Kind=commit`、`GID=""`、decoder block 的 XID 与实际 terminal LSN；SQLXID 字段在此是兼容诊断，不是 full-XID 关联证据。

保留 Open/CreateBaseBackup/CrashPrimary/RestoreAtCut/PromoteCandidate/InspectTimeline 的原签名。给既有 backup/cut/candidate 加未导出绑定字段，比较 retained record 全部字段，而非只比较可写名字。不得删除或改变既有 ownership WAL schema。

## Task 1: Narrow Access, Eager Rows, and Lifecycle

**Files:** File Map Task 1，controller 状态字段和 access 接入点。

**Interfaces:** 消费 `store.DBTX`、现有 controller 私有连接配置；产出上述 Access/Transaction/With 方法、七个有限 error 值，以及下面私有 seam。真实 driver 只保存在 testinfra 私有实现中，不导出。

```go
type C12AuthorityPITRError string
func (e C12AuthorityPITRError) Error() string { return string(e) }
const (
    C12PITRInvalidHandle C12AuthorityPITRError = "pitr invalid handle"
    C12PITRWrongPhase C12AuthorityPITRError = "pitr wrong phase"
    C12PITRNotObserved C12AuthorityPITRError = "pitr not observed"
    C12PITRCapacityExceeded C12AuthorityPITRError = "pitr capacity exceeded"
    C12PITRIndeterminate C12AuthorityPITRError = "pitr indeterminate"
    C12PITRCanceled C12AuthorityPITRError = "pitr canceled"
    C12PITRDependencyFailure C12AuthorityPITRError = "pitr dependency failure"
)
type c12ReadBudget struct { rows, bytes int64 }
func (b *c12ReadBudget) take(rowBytes int64) error
func c12MaterializeRows(context.Context, *c12AccessLease, pgx.Rows) (pgx.Rows, error)
func c12WithAccess(context.Context, *c12AuthorityPITRState, *c12CandidateBinding,
    func(C12AuthorityAccess) error) error
```

`c12CandidateBinding` 为 Task 4 签发的私有 retained identity；primary 传 nil。`c12AccessLease` 包含 owner、generation、context/cancel、原子 live 位、共享 budget、一个活动事务及最多两个私有连接；零 lease 无效。controller 初始 access SQL policy 只拒绝，不能临时放行任意 SELECT。

controller 的 run generation 在同一次 Open→backup→crash→restore 链内保持不变；phase改变不使已授权restore cut失效。每个lease/transaction另有递增generation与revoked位，关闭后不复用。Restore后只有该candidate的binding合法，不重新签发primary能力。禁止用一次全局generation自增误杀正常restore链，也禁止保留旧tx能力跨phase使用。

- [ ] **Step 1: 写限额与窄接口 RED 测试。** 在 `TestC12AuthorityPITRScopedAccessPolicy` 下先加入：

```go
func TestC12AuthorityPITRScopedAccessPolicy(t *testing.T) {
    t.Run("bounded_materialization", func(t *testing.T) {
        b := c12ReadBudget{}
        for i := 0; i < 65536; i++ {
            if err := b.take(0); err != nil { t.Fatal(err) }
        }
        if !errors.Is(b.take(0), C12PITRCapacityExceeded) { t.Fatal("row cap") }
        if !errors.Is((&c12ReadBudget{}).take(1048577), C12PITRCapacityExceeded) {
            t.Fatal("single row cap")
        }
        b = c12ReadBudget{bytes: 67108864}
        if !errors.Is(b.take(1), C12PITRCapacityExceeded) { t.Fatal("byte cap") }
    })
    t.Run("zero_controller", func(t *testing.T) {
        called := false
        err := (C12AuthorityPITRController{}).WithPrimaryAuthorityAccess(
            t.Context(), func(C12AuthorityAccess) error { called = true; return nil })
        if !errors.Is(err, C12PITRInvalidHandle) || called { t.Fatal("zero authorized") }
    })
    t.Run("no_raw_method_set", func(t *testing.T) {
        for _, typ := range []reflect.Type{
            reflect.TypeFor[*c12AuthorityAccess](), reflect.TypeFor[*c12AuthorityTx](),
        } {
            for _, name := range []string{"Conn", "Config", "BeginTx", "Prepare", "SendBatch", "CopyFrom", "LargeObjects"} {
                if _, ok := typ.MethodByName(name); ok { t.Fatalf("escape: %s", name) }
            }
        }
        if reflect.TypeFor[*c12AuthorityTx]().Implements(reflect.TypeFor[pgx.Tx]()) {
            t.Fatal("transaction exposes pgx.Tx")
        }
    })
}
```

- [ ] **Step 2: 执行 RED。** `go test -tags=integration ./internal/testinfra -run '^TestC12AuthorityPITRScopedAccessPolicy$' -count=1 -timeout=3m`。预期新增符号未定义；记录实际输出，不能把环境故障算 RED。

- [ ] **Step 3: 实现限额和物化。** 私有 `c12AuthorityAccess`/`c12AuthorityTx` 不嵌入 pgx.Tx/pool。`c12AuthorityTx` 持有私有 driver、lease、原子 terminal 状态、transaction generation。预算用减法防溢出：

```go
func (b *c12ReadBudget) take(n int64) error {
    if n < 0 || n > 1<<20 || b.rows >= 65536 || b.bytes > (64<<20)-n {
        return C12PITRCapacityExceeded
    }
    b.rows++
    b.bytes += n
    return nil
}
```

物化使用 `pgx.Rows.RawValues()` 的逐列深拷贝、复制 FieldDescriptions、保存 CommandTag；驱动 Rows 必须 defer Close，读至正常 EOF 后才发布结果。`QueryRow` 调同一物化器：保存首行/无行结果，关闭驱动，再返回本地 Scan 视图。读取失败立即返回，不在 Scan 或 Commit 时补读。lease 预算聚合所有查询，包含 metadata 与返回副本；零行/巨量字段也收费，先检查字段长度和列数再分配。pgtype.Map.Scan 按 OID/format 扫描复制的 raw bytes；Values 经 codec 解码后复制 byte slices，不接受未知可变值类型。`Rows.Conn() *pgx.Conn` 永远 nil。Close/Err 可安全收尾，失效 Next=false，Scan/Values 返回 wrong phase/canceled；RawValues/FieldDescriptions 失效返回 nil。正常提前 Close 后不允许再次读取，不能暴露内部切片。

DBTX 的“无行”是必要查询语义，不能把它误作数据库故障：私有无行 error 的 Error 仍是有限 dependency failure，`Is` 只匹配 `pgx.ErrNoRows` 和 `C12PITRDependencyFailure`；不保留 raw PgError/DSN。生产 repository 当前用 `errors.Is` 判无行，不需生产修改。

```go
type c12NoRows struct{}
func (c12NoRows) Error() string { return C12PITRDependencyFailure.Error() }
func (c12NoRows) Is(target error) bool {
    return target == pgx.ErrNoRows || target == C12PITRDependencyFailure
}
```

- [ ] **Step 4: 加 fake driver 的生命周期 RED 表。** 测试内定义 `c12AccessTestDriver`：嵌入 nil `pgx.Tx` 仅用于未实现方法立刻 panic，覆盖 Exec/Query/Commit/Rollback；字段 `events []string`、`commits, rollbacks, cursorCloses int`、`rows pgx.Rows`、`commitErr error`。另定义 `c12AccessTestRows`，按 pgx.Rows 全部方法实现固定两行 bytea、可配置第 N 行错误、Close 计数和每次读取事件，Conn 故意返回一个非 nil 测试指针以证明 wrapper 不转发。所有生产 adapter methods 仍逐个实现，不嵌入 driver。

新增子测试与精确断言：

```go
var c12AccessCases = []string{
    "eager_rows_and_expired_views", "queryrow_is_not_lazy", "raw_values_are_copied",
    "callback_error", "callback_panic", "parent_cancel", "lease_deadline",
    "one_active_transaction", "one_active_callback", "double_commit",
    "unconsumed_views_need_no_commit_io", "foreign_candidate", "no_rows_is_preserved",
}
```

每例用上述 fake 接到私有 `c12WithAccess` opener seam，不公开注入点。`eager_rows_and_expired_views` 必须断言 Query 返回时 cursorCloses=1；修改 RawValues/Values/FieldDescriptions 不影响第二次读取；保存 row/rows/tx 到 callback 外再使用只返回有限错误，events 不增长。panic 捕获为 dependency failure、不输出 panic 值，未提交事务 rollback 一次，连接关闭；error/cancel/timeout 同样回收。单个 callback 内第二个 Begin 拒绝；第一个完成后可新 Begin，generation 不相同。取消并发用 channels/barriers，不 sleep 30 秒；子 context 短 deadline 证明取最早值。

- [ ] **Step 5: 实现 Begin/Commit 和作用域关闭。** Begin 只用 `pgx.TxOptions{IsoLevel: pgx.ReadCommitted}`，不提供参数。access 始终使用 lease context 派生调用，不能让 callback 用 Background 延长寿命。driver Commit 前仅 CAS terminal 状态；先前绑定/查询错误必须已阻止正常返回。事务完成后在 driver 返回之后使 tx 视图失效并更新 lease，取消后的 cleanup 不可重发 Commit。第一次 Commit 的唯一真实调用模式：

```go
func (tx *c12AuthorityTx) Commit(ctx context.Context) error {
    if !tx.terminal.CompareAndSwap(0, 1) { return C12PITRWrongPhase }
    err := tx.driver.Commit(tx.lease.ctx)
    tx.finishAfterDriver(err)
    return tx.commitResult(err)
}
```

`finishAfterDriver(error)`、`commitResult(error) error` 是私有本地收尾/有限错误映射，定义在本文件；不能做新的数据库写入。`ctx` 不用于绕开 lease：调用 ctx 的取消已在正常 operation/lease 绑定，Commit 前不能插入新检查。Rollback 使用同一 terminal gate、重复安全，返回 finite wrong phase。callback 收尾用有界 cleanup context 撤销能力并 rollback/close 连接，不清理容器。并发取消与 Commit 共用 terminal gate，若 Commit 已开始只能等待受限 driver 调用/关闭连接，不能第二次 Commit。

- [ ] **Step 6: GREEN 与提交。** 重跑 Step 2，再用 `-race` 跑同一精确测试（平台支持时）；预期 PASS、无数据竞争。只提交本 task 三个新文件及 controller 接入：`feat(testinfra): add scoped PITR authority access`。若 race 工具链缺失，记录未执行，不把普通 PASS 称为 race PASS。

## Task 2: Closed SQL Registry, Fixture Binding, and Candidate Grants

**Files:** Task 2 两个新文件；Task 1 access 的私有执行前策略调用；controller Open/backup 内部初始化。

**Interfaces:** 消费 DBTX adapter；产出 `c12AuthorizeSQL(mode c12SQLMode, call c12SQLCall, sql string, args []any, fixture *c12FixtureLedger) error`。`c12SQLMode` 为内部位集：primary=1、setup=2、candidate=4；backup 前 primary callback 为 primary|setup，之后仅 primary，candidate 不可组合。`c12SQLCall` 仅 exec/query/queryRow 三值。`c12FixtureLedger` 保留同 controller 的 activation、三个 node、operation、certificate、epoch/sequence/scope 绑定；候选视图只读。无 caller 配置构造器。

```go
type c12SQLMode uint8
const ( c12SQLPrimary c12SQLMode=1; c12SQLSetup c12SQLMode=2; c12SQLCandidate c12SQLMode=4 )
type c12SQLCall uint8
const ( c12SQLExec c12SQLCall=iota; c12SQLQuery; c12SQLQueryRow )
type c12FixtureOperation struct {
    operationID,nodeID,certificateID uuid.UUID
    epoch,sequence int64
    scopeDigest [32]byte
}
type c12FixtureCertificate struct {
    nodeID,issuanceID,historyOperationID uuid.UUID
    historySequence int64
    scopeDigest [32]byte
}
type c12FixtureLedger struct {
    activationID uuid.UUID
    epoch int64
    certificates map[uuid.UUID]c12FixtureCertificate
    operations []c12FixtureOperation
}
```

pending overlay 是ledger的深复制，最多3组；本地提交后再合并，不aliasmap/slice。setup闭包latch/activation依赖身份单独保存在同一事务的固定字段中，最多一套，不接受任意registration集合。

### Closed statement inventory

把下表 **现有实际传给 DBTX 的完整字节字面量** 复制到新 policy Go constants；包括 sqlc `-- name:` 注释、换行。不编辑 sqlc 生成文件、不运行时读源文件/规范化 SQL。AST parity test 从以下确切 owner 读取并比对 SHA256+字节；执行时只用编译期登记。

| 来源/名字 | 调用/参数数 | 允许模式 |
| --- | --- | --- |
| `internal/store/nodecontrol_authority.sql.go`: GetNodeControlDatabaseIdentity、GetAuthorityFenceHead | QueryRow / 0 | primary、candidate |
| 同文件 ListPendingAuthorityFences | Query / 1 | primary、candidate |
| 同文件 LockAuthorityFence | QueryRow / 1 | primary、candidate |
| 同文件 LockCertificateRevocationOutcome | QueryRow / 1 | primary、candidate |
| 同文件 GetStoredAuthorityFence、GetAuthorityFenceForUpdate | QueryRow / 1 | primary |
| 同文件 InsertClaimV1AuthorityFencePending | Exec / 9 | primary |
| 同文件 BindAuthorityFenceEffect | Exec / 6 | primary |
| 同文件 ActivateCommittedAuthorityFence | Exec / 7 | primary |
| crash fixture ResolveRegisteredAuthorityEffectForUpdate: aux effect SELECT FOR UPDATE | Query / 1 | primary、candidate |
| 同方法: certificate resolution IS NOT NULL FOR UPDATE | QueryRow / 1 | primary、candidate |
| task9ReadCertificateInput: certificate input/commitment FOR UPDATE | QueryRow / 1 | primary、candidate |
| ValidatePersistedAuthorityEffect: status/evidence/resolution FOR UPDATE | QueryRow / 1 | primary、candidate |
| 同方法: audit+outbox counts | QueryRow / 1 | primary、candidate |
| 同方法: outbox payload WHERE event/aggregate/version | QueryRow / 3 | primary、candidate |
| CaptureActivationDecisionMaterial: commitment SELECT | QueryRow / 1 | primary |
| commitDomain: certificate commitment UPDATE；aux effect INSERT | Exec / 6、6 | primary |
| ActivateAuthorityEffect: certificate UPDATE；audit INSERT；outbox INSERT | Exec / 12、7、6 | primary |
| task9SeedProofActivation: installed latch SELECT | QueryRow / 0 | setup |
| 同 helper: exact replica/origin SET LOCAL | Exec / 0 | setup 事务内 |
| 同 helper: upgrade_intents、runtime_registration_results、upgrade_attempts、protocol_activations、activation_completions、activation_releases INSERT | Exec / 7、4、8、3、3、4 | setup |
| seedCertificate: exact node_pops、node_inventory、legacy issuance fence、node_certificate_issuances、node_certificates INSERT | Exec / 1、3、7、10、10 | setup |
| task9InstallCrashAuxiliaryTables: authority_task7_crash_effects CREATE TABLE | Exec / 0 | 仅 controller Open 内部，不向消费者放行 |

不登记旧 InsertAuthorityFencePending、Abort、其他 effect kind 的 outcome locks、其他 protocol fixture initializer、复制 outbox 的 DDL。当前场景只支持 certificate_revoke；注册 dispatcher 其他 handler 的 wrong-kind 分支必须不访问 SQL。候选 positive fixture 必须零 pending。

新增三个固定断言 SELECT（定义在 authority bridge；policy 复制并做 AST parity）及仅控制器可调用的 marker INSERT：

```go
const pitrCertificateSnapshotSQL = `SELECT status,revoke_authority_operation_id,revoke_authority_effect_commitment_jcs,revoke_authority_activation_evidence_jcs,revoke_authority_effect_resolution_jcs FROM nodecontrol.node_certificates WHERE certificate_id=$1`
const pitrOperationSnapshotSQL = `SELECT provider_status,visibility_state,authority_protocol_profile FROM nodecontrol.control_plane_authority_fences WHERE operation_id=$1`
const pitrAuxCountSQL = `SELECT count(*) FROM nodecontrol.authority_task7_crash_effects WHERE operation_id=$1`
const c12MarkerInsertSQL = `INSERT INTO public.c12_authority_pitr_markers(marker) VALUES($1)`
```

快照 SELECT primary/candidate，参数均1个已登记 uuid.UUID；marker 不在消费者 SQL 集合，由 Bind 的私有路径执行。marker 表 `marker text PRIMARY KEY CHECK(marker ~ '^[0-9a-f]{64}$')` 由 controller Open 建立，无 caller DDL。

参数策略逐条固定：operation/node/certificate 是 uuid.UUID；LockCertificateRevocationOutcome 为 `uuid.NullUUID{Valid:true}`；list epoch 是 int64(31)，当前协议序列限1..3；digest 是32字节副本；JCS/DER/outbox 沿用所属生产字段边界且单值不超1MiB；scope/effect/state 只接受本 fixture 的闭集常量，命名 string 类型按 exact package/type 或基础 string 的显式分支处理，禁止 fmt.Stringer/driver.Valuer 执行 caller 代码。sqlc `pgtype.Timestamptz/Int8/Numeric` 要求 Valid、有限、正值及预期 SystemID/timeline；Numeric 深复制 big.Int。RequiredLsn 只接受 repository 当前传入的 WAL 字符串形状，不能接受任意 interface codec。所有入参先复制，再授权/执行，防 callback 并发修改切片。

fixture ledger 的新增身份仅来自 **成功提交的已授权 setup/claim 语句**：setup 中建立 activation、最多3组 certificate→node/issuance/history；当前 claim INSERT 校验已存在 activation、节点 scopeDigest、epoch31、下一个sequence，再登记 operation。同事务后续 Lock/UPDATE 必须能读取该事务私有 pending ledger overlay，否则 RecordPending→Lock 会自拒绝；其他连接/事务看不到 overlay。certificate commitment UPDATE 要求 certificate 为 `uuid.NewSHA1(operationID, []byte("certificate"))` 且同 node/scope。rollback 丢弃 pending delta，Commit 响应不确定时不猜结果；后续写授权失败关闭，不凭 caller 声称已提交扩表，但独立 Observe 仍能核验已绑定 marker。候选只使用原 primary 已提交身份集合，因此 B 的不存在查询仍是已登记操作。legacy fence 的 epoch1/序列1..3只是 setup 历史，不得成为当前 provider31 的 claim。

### Exact role/object inventory

controller 派生私有 candidate 登录名 `talenro_c12_<既有 run suffix 前24hex>_candidate`、随机密码，Open 创建 NOSUPERUSER/NOCREATEDB/NOCREATEROLE/NOREPLICATION/NOBYPASSRLS；无需新 descriptor 字段或宿主资源。角色、marker、aux 表和 grants 在 Open 内按固定次序全部建好，aux DDL 复制原 helper 的精确表定义；authority 消费者不再次 CREATE。这样原基础设施 profile 不调用 authority fixture 也能正常备份。备份前验证必需对象存在，不在 candidate 补建；原隔离 child-DB crash helper 仍自己创建其 aux/outbox，两个路径不得混用。

| 对象 | candidate 权限 |
| --- | --- |
| exact runner database | CONNECT |
| nodecontrol、public | USAGE |
| control_plane_authority_fences、node_certificates、authority_task7_crash_effects、node_operator_audit、transactional_outbox | SELECT |
| control_plane_authority_fences、authority_task7_crash_effects | UPDATE(operation_id)，仅使 FOR UPDATE 可用 |
| node_certificates | UPDATE(certificate_id)，仅使 FOR UPDATE 可用 |
| pg_catalog.pg_control_system()、pg_control_checkpoint()、pg_current_wal_insert_lsn() | EXECUTE |

不 GRANT sequence/temp/DDL/角色成员资格；不改变其他数据库 PUBLIC 权限。继承的 PUBLIC schema 权限不是 SQL wrapper 授权，wrapper 仍精确拒绝 DDL。任何额外函数/表权限需求须报告具体失败，不自动 GRANT ALL。primary setup 仍使用现有私有受控管理员通道，但只放行上述固定 setup，CreateBaseBackup 关闭 setup phase；新 access 不能 caller 选择该模式。

列级 UPDATE 用于行锁的依据已核对 [PostgreSQL 18 SELECT 文档](https://www.postgresql.org/docs/18/sql-select.html)：每个被锁表至少一列需有 UPDATE 权限。具体本项目 grants 是否足够仍须 Task5 真实数据库对照，文档不能替代运行证据。

- [ ] **Step 1: 写 policy RED。** 定义测试里的 `c12PolicyFixture()` 返回固定31 epoch、1个已登记 operation/node/certificate/activation 的 ledger；UUID 用 uuid.MustParse 固定非零值，无数据库访问。验证完整字节/调用类型/参数拒绝：

```go
func c12PolicyFixture() *c12FixtureLedger {
    operationID:=uuid.MustParse("81000000-0000-4000-8000-000000000001")
    nodeID:=uuid.MustParse("81000000-0000-4000-8000-000000000002")
    certificateID:=uuid.NewSHA1(operationID,[]byte("certificate"))
    scope:=sha256.Sum256(append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"),nodeID[:]...))
    return &c12FixtureLedger{
        activationID:uuid.MustParse("79000000-0000-4000-8000-000000000001"),epoch:31,
        certificates:map[uuid.UUID]c12FixtureCertificate{certificateID:{nodeID:nodeID,scopeDigest:scope,historySequence:1}},
        operations:[]c12FixtureOperation{{operationID:operationID,nodeID:nodeID,certificateID:certificateID,epoch:31,sequence:1,scopeDigest:scope}},
    }
}
func TestC12AuthorityPITRSQLPolicy(t *testing.T) {
    ledger := c12PolicyFixture()
    good := pitrAuxCountSQL
    id := ledger.operations[0].operationID
    for _, bad := range []string{
        good + "; DELETE FROM nodecontrol.node_certificates",
        good + " -- comment", " " + good,
        "WITH x AS (DELETE FROM nodecontrol.node_certificates RETURNING *) SELECT * FROM x",
        "SELECT pg_advisory_lock(1)", "SET ROLE talenro",
    } {
        if c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, bad, []any{id}, ledger) == nil {
            t.Fatal("non-registered SQL authorized")
        }
    }
    for _, args := range [][]any{{}, {id, id}, {id.String()}, {uuid.New()}, {uuid.Nil}} {
        if c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, good, args, ledger) == nil {
            t.Fatal("wrong binding authorized")
        }
    }
    if err := c12AuthorizeSQL(c12SQLCandidate, c12SQLQueryRow, good, []any{id}, ledger); err != nil {
        t.Fatal(err)
    }
    if c12AuthorizeSQL(c12SQLCandidate, c12SQLExec, good, []any{id}, ledger) == nil {
        t.Fatal("wrong call kind")
    }
}
```

- [ ] **Step 2: 运行 RED。** `go test -tags=integration ./internal/testinfra -run '^TestC12AuthorityPITRSQLPolicy$' -count=1 -timeout=3m`。预期缺少固定 policy，而非连接 Docker。

- [ ] **Step 3: 实现逐条 registry 和 private grants。** 使用 `map[string]c12SQLRule` 以完整 SQL 为 key；rule 保存 call/modes/validator，validator 按上表 exact arity/type/ledger 检查；找不到直接 invalid handle。不能先执行再判断，也不能让 QueryRow 延迟授权。role 标识符只由 controller 验证的 run suffix 派生并用 pgx.Identifier.Sanitize 引用，密码不进 error/log/返回值；参数化不能替代标识符验证。

```go
type c12SQLRule struct {
    call c12SQLCall
    modes uint8
    validate func([]any, *c12FixtureLedger) error
}
func c12AuthorizeSQL(mode c12SQLMode, call c12SQLCall, query string, args []any, f *c12FixtureLedger) error {
    rule, ok := c12AuthoritySQLRules[query]
    if !ok || rule.call != call || rule.modes&uint8(mode) == 0 || f == nil {
        return C12PITRInvalidHandle
    }
    return rule.validate(args, f)
}
```

- [ ] **Step 4: 加 drift/ledger/grants RED→GREEN。** `TestC12AuthorityPITRSQLRegistryMatchesSources` 用 go/parser 定位上表确切 const 或函数 call 的 BasicLit，经 strconv.Unquote 后逐字比对；拒绝来源新增未映射 DBTX calls。对源副本改一字节、参数位置或调用 Exec/Query，必须失败。另测 setup 提交后登记、rollback 不登记、backup 后 replica/DDL 拒绝、statement argument slices 复制、外来 fixture UUID 拒绝、candidate 每个 UPDATE/INSERT 常量拒绝。授予列 UPDATE 的实际 PostgreSQL 充分性由 Task 5 live ready 对照验证，不能只凭静态检查宣布有效。

- [ ] **Step 5: GREEN 与提交。** 运行两个 policy 顶层及 Task1 测试精确 alternation；预期全部 PASS。提交 `feat(testinfra): freeze PITR SQL capabilities and candidate grants`，仅 policy 文件/controller/access 必需接线。

## Task 3: Actual Commit Observation and Same-Handle Retry

**Files:** Task 3 两个新文件；controller state 增加固定64槽 observations、observationPoison、generation；不改变 ownership WAL 消费协议。

**Interfaces:** 消费 exact `*c12AuthorityTx`，产出 New/Bind/Observe。私有 parser 与原 prepared parser分开入口，共用有界读取，不允许新普通路径接受 prepared terminal。

```go
type c12ObservationState struct {
    owner *c12AuthorityPITRState
    generation uint64
    marker string
    transaction *c12AuthorityTx
    bound bool
    result *c12ObservedCommitState
}
type c12ObservedCommitState struct {
    owner *c12AuthorityPITRState
    generation uint64
    observation *c12ObservationState
    facts C12AuthorityPITRCommit
}
func c12ParseObservedTransactions([]c12AuthorityPITRDecodedRow,
    map[string]*c12ObservationState) (map[*c12ObservationState]C12AuthorityPITRCommit, error)
func c12DrainObservations(context.Context, *c12AuthorityPITRState) error
```

所有字段访问由 controller mutex 或明确一次性状态保护；marker 不导出、不出错消息。marker 32随机字节→64hex，容量/phase检查先于随机数与 SQL。复制 handle 指向同登记项，不生成新 marker。

New/Bind/Observe 在 primary phase1（backup前P）和phase2（backup后A/B）均可使用；Observe必须在callback结束后。Select只允许已经有base backup的phase2。这样P是真实前缀而不需要另一套不受控提交接口。

- [ ] **Step 1: 写完整 block parser RED。** 测试代码生成固定 test_decoding 事务块，不将 BEGIN/change LSN 当 terminal：

```go
func c12DecodedCommit(xid uint64, marker, end string) []c12AuthorityPITRDecodedRow {
    return []c12AuthorityPITRDecodedRow{
        {LSN:"0/10", XID:xid, Data:fmt.Sprintf("BEGIN %d", xid)},
        {LSN:"0/18", XID:xid, Data:"table public.c12_authority_pitr_markers: INSERT: marker[text]:'"+marker+"'"},
        {LSN:end, XID:xid, Data:fmt.Sprintf("COMMIT %d", xid)},
    }
}
func TestC12AuthorityPITRCommitObservationStateMachine(t *testing.T) {
    t.Run("multi_result_xid_wrap", func(t *testing.T) {
        a := &c12ObservationState{marker:strings.Repeat("a",64), bound:true}
        b := &c12ObservationState{marker:strings.Repeat("b",64), bound:true}
        rows := append(c12DecodedCommit(4294967295,a.marker,"0/20"),
            c12DecodedCommit(3,b.marker,"0/40")...)
        got, err := c12ParseObservedTransactions(rows, map[string]*c12ObservationState{a.marker:a,b.marker:b})
        if err != nil || len(got)!=2 || got[a].EndLSN!="0/20" || got[b].EndLSN!="0/40" {
            t.Fatalf("lost or mixed transaction: %#v %v", got, err)
        }
    })
    t.Run("ambiguous_marker", func(t *testing.T) {
        a := &c12ObservationState{marker:strings.Repeat("a",64),bound:true}
        rows := append(c12DecodedCommit(7,a.marker,"0/20"),c12DecodedCommit(8,a.marker,"0/40")...)
        if _,err:=c12ParseObservedTransactions(rows,map[string]*c12ObservationState{a.marker:a}); err==nil {
            t.Fatal("ambiguous marker proved commit")
        }
    })
}
```

- [ ] **Step 2: 运行 RED。** `go test -tags=integration ./internal/testinfra -run '^TestC12AuthorityPITRCommitObservationStateMachine$' -count=1 -timeout=3m`。预期新 parser/state 未定义。

- [ ] **Step 3: 实现 New/Bind。** Bind 只接受接口动态类型 exact private `*c12AuthorityTx`，且 owner/current primary/generation/live 与 retained transaction 指针全部一致；foreign wrapper 即使转发全部方法也拒绝。observation 必须未绑定，事务不能已绑定第二个 observation。用该 driver 的当前事务执行唯一固定 marker INSERT，成功后安装绑定；任何写入错误不重新绑定为另一事务。

```go
tx, ok := transaction.(*c12AuthorityTx)
if !ok || tx == nil || tx.lease.owner != controller.state {
    return C12PITRInvalidHandle
}
// 在持锁完成 phase/generation/once 验证并登记 binding-in-flight 后执行。
_, err := tx.driver.Exec(tx.lease.ctx, c12MarkerInsertSQL, observation.state.marker)
if err != nil { return C12PITRDependencyFailure }
```

Bind 必须是 activator 正常 SQL 阶段的末尾动作，不能在 Commit 包装中插入。失败 binding-in-flight 变不可再用，不清空 marker 让别的事务重试。每个 observation 只支持一笔普通事务，rollback 保留“未证明”，不产出 commit。

- [ ] **Step 4: 实现有界 drain 和严格 parser。** 独立私有连接，在没有活动 access/transaction 时读取固定 slot。`pg_logical_slot_get_changes` 使用固定 rows hint，但每读一行先累计行/字节限额；不无界 append。按 BEGIN→changes→COMMIT 顺序验证完整 block、XID 文本与行 xid 一致，普通 marker INSERT 必须完整匹配固定表/字段/64hex。未知合法事务可忽略，但不能将其中 marker 搬到另一个 block；已登记 marker 的 UPDATE/DELETE、重复、跨块、prepared/缺失终端、重复 terminal 全拒绝。LSN 必须有效且 terminal 不早于 block 前面的 LSN，数字比较不用字符串词典序。

成功 drain 先完整校验全部 rows，再一次安装 **所有** 已绑定 observation 的结果，返回其中请求项的值副本。若请求 marker 未出现且 drain 完整成功，返回 not observed，不推断 rollback。已安装结果的 Observe 重试零 slot I/O；post-install response-loss seam 仅私有测试可用。真实 driver Commit 的 error 不阻止观察，不能再提交业务事务探测。

Query 发出后，任何取消/读取错误/超限/parse 歧义或无法确认 destructive read 完整性，都将通道 poison 为 indeterminate；后续 New/Bind/Select/Crash 拒绝，已有结果可作诊断读取但不能用于新 crash。发出查询前的 canceled 不需 poison。不得把 ownership WAL 当 commit archive，不新增持久化重开承诺。

- [ ] **Step 5: 用 fake stream/tx 增加故障 RED→GREEN。** 在同一顶层添加下列明确子例；fake drain 私有 seam `func(context.Context) (pgx.Rows,error)` 返回计数/可中途失败 Rows，复用 Task1 fake，不公开给消费者：

```go
var c12ObservationCases = []string{
    "zero_foreign_and_wrapped_transaction", "copy_cannot_rebind", "closed_transaction",
    "capacity_64_no_side_effect_65", "rollback_not_observed", "commit_response_lost",
    "observe_response_lost_same_facts", "partial_drain_poison", "cancel_before_query",
    "cancel_after_query_poison", "row_byte_and_single_row_limits", "missing_terminal",
    "duplicate_terminal", "prepared_terminal_rejected", "cross_block_marker",
    "consume_then_one_driver_commit", "completed_handle_after_generation_change",
}
```

`commit_response_lost` fake driver 先记录 committed marker，再返回 sentinel error，Observe 必须得到唯一 terminal、driver commit count=1。`observe_response_lost_same_facts` 先安装结果后丢响应，重试逐字段相等且 drain count 不增。`partial_drain_poison` 在第二行返回错误，之后 Select/Crash 必须 indeterminate、Docker call count=0。64容量测试比较整个 backend event 列表，65不发随机数/SQL。顺序测试注入只读 event recorder 到正常 driver方法；观察到 bind→consume→driver_commit，consume 与 driver_commit 间没有 Query/Exec/rows.Close/time/WAL seam；double Commit 不增加次数。不能只做字符串静态审查。

- [ ] **Step 6: GREEN、race、提交。** 重跑 Step2 及 Tasks1–2 精确测试；支持时 `-race`。提交 `feat(testinfra): observe controller-bound authority commits`。旧 prepared parser测试必须保留，不顺手删除。

## Task 4: Earlier Recovery Target and Exact Candidate Identity

**Files:** Task4 两个新文件；controller backup/crash/restore/promote/inspect；既有基础设施 profile 断言。

**Interfaces:** 消费 observed handles；产出 Select/CrashAtCut、私有 candidate binding 和 shared crash implementation。

```go
type c12CutSelectionState struct {
    owner *c12AuthorityPITRState
    generation uint64
    backup C12AuthorityPITRBaseBackup
    target *c12ObservedCommitState
}
type c12CandidateBinding struct {
    owner *c12AuthorityPITRState
    generation uint64
    index uint8
    cut *c12CutSelectionState
}
func c12ValidateRecoveryOrder(backupEnd, targetEnd, boundaryEnd string) error
func c12CrashAtValidatedCut(context.Context, *c12AuthorityPITRState,
    C12AuthorityPITRBaseBackup, C12AuthorityPITRCommit, C12AuthorityPITRCommit) (C12AuthorityPITRCut,error)
```

last helper 两个 commit 参数分别 target/boundary；不能互换。phase transition mutex 和 active-access 标志同一锁域，先完整验证再认领副作用 phase，避免 bad handle 先 CAS 烧掉正常 state。

- [ ] **Step 1: 写切点顺序 RED。**

```go
func TestC12AuthorityPITRRecoveryCutStateMachine(t *testing.T) {
    for _, tc := range []struct{backup,target,boundary string; ok bool}{
        {"0/10","0/20","0/40",true}, {"0/10","0/20","0/20",true},
        {"0/20","0/10","0/40",false}, {"0/10","0/40","0/20",false},
        {"0/F","0/10","1/0",true}, {"0/10","garbage","0/40",false},
    } {
        err:=c12ValidateRecoveryOrder(tc.backup,tc.target,tc.boundary)
        if (err==nil)!=tc.ok { t.Fatalf("order %v: %v",tc,err) }
    }
}
```

- [ ] **Step 2: 运行 RED。** `go test -tags=integration ./internal/testinfra -run '^TestC12AuthorityPITRRecoveryCutStateMachine$' -count=1 -timeout=3m`。

- [ ] **Step 3: 实现 selection 与共同 crash。** parser 把 PostgreSQL LSN 高低32位转 uint64，拒绝越界/非规范值；只比较，不加减。Select 验证 retained backup 和 observed record 属主、generation、phase、非 poison、A>=backup.End；重复同一个 selection 幂等返回原handle，不签发第二份；不同 A capacity exceeded。CrashAtCut 再验证 B>=A、active callback/tx=0，冻结 boundary。归档/停止完全沿用现有 exact primary/ownership逻辑，归档等待覆盖 B。

```go
cut := C12AuthorityPITRCut{
    BackupID: backup.BackupID,
    TerminalCommit: boundary,
    RecoveryTargetLSN: target.EndLSN,
}
```

cut 加私有 retained binding；TerminalCommit 表示物理崩溃证明边界，恢复目标只由 target 决定。旧 CrashPrimary 验证已保留的私有 probe record 后进入同一 helper，target=boundary；不能接受外部伪造新 observation。旧 prepared probe 可走兼容入口，不可进入新 Select。

原批准文本在此要求`recovery_target_inclusive=on`；R17 已纠正此计划缺陷为 `off/false`，保留原样逻辑 terminal EndLSN，不作 LSN 运算。[REL_18_4 logical.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/replication/logical/logical.c) 与 [logicalfuncs.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/replication/logical/logicalfuncs.c) 输出 commit record END；[xlogrecovery.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/access/transam/xlogrecovery.c) 按 record START 在 replay 前/后判断非 inclusive/inclusive。由此推论，off 在 A 已重放后、下一 record 前停止，on 会多重放下一 record。此修正保留已批准 A 包含/B 排除语义，见 [PostgreSQL 18 Recovery Target 文档](https://www.postgresql.org/docs/18/runtime-config-wal.html#RUNTIME-CONFIG-WAL-RECOVERY-TARGET) 和文末 R17 记录。实际恢复参数断言已执行；真实 A/B 领域数据物理 gate 仍未执行，不能以 LSN 文本或源代码证明替代。

- [ ] **Step 4: candidate 所有操作验证完整状态。** retained candidate 存 exact index/name/dataName/target/promoted、container ID/system identity/timeline、binding。Restore 必须 exact cut、固定8槽，容量检查先于 intent/create；Promote 返回新 promoted DTO，旧 pre-promote DTO 随即无效；Inspect/Access 只接受已晋升状态4或已检查状态5的 exact DTO。InspectTimeline 保留原有一次性行为，不放宽为任意再晋升。检查 systemID=backup systemID、timeline > backup timeline。callback 存活时 CreateBaseBackup/Crash/Restore/Promote 均拒绝，不接管该 callback 的连接。

- [ ] **Step 5: 状态/副作用 RED→GREEN。** 增加 `foreign_backup`、`foreign_observed`、`forged_cut`、`altered_candidate_target`、`altered_candidate_name`、`stale_pre_promote_dto`、`candidate_capacity_8`、`active_callback_blocks_crash`、`failed_validation_preserves_phase`、`poison_blocks_crash`、`a_survives_b` 子例。用私有 backend event recorder（复用 state 的测试 seam，不新公开 API），每个负例断言 event count 不变、phase不变；合法重试仍可用。`a_survives_b` 断言 return cut 的 terminal=B而 target=A，restore 使用 inclusive A，不依赖最近 terminal字段。

- [ ] **Step 6: GREEN、兼容编译、提交。** 运行 cut+observation 精确 tests；`go test -tags=integration ./internal/testinfra -run '^$' -count=1 -timeout=5m` 仅编译。提交 `feat(testinfra): separate PITR recovery target from crash boundary`。实际旧 profile compatibility 由 Task6 单独物理 gate验收。

## Task 5: Real Authority Bridge and P/A/B Migration

**Files:** authority 三个 integration `_test.go`；policy 仅为本计划固定源字面量的提取同步，不增加 statement范围。

**Interfaces:** 消费 controller 窄 API；产出以下仅同包测试 helper，不导出、不改变生产：

```go
type pitrAuthorityRepository struct {
    *PostgresRepository
    access testinfra.C12AuthorityAccess
}
func (r *pitrAuthorityRepository) beginAuthorityTransaction(ctx context.Context) (authorityTransaction,error) {
    return r.access.Begin(ctx)
}

type pitrOperationFixture struct {
    request ReserveRequest
    nodeID, certificateID uuid.UUID
    sequence uint64
}
func pitrSeedClosure(context.Context, store.DBTX, uuid.UUID, int, time.Time) error
func pitrSeedCertificate(context.Context, store.DBTX, DatabasePoint, pitrOperationFixture, time.Time) error
func pitrCommitDomain(context.Context, store.DBTX, *PostgresRepository, Reservation, uuid.UUID) (contracts.Digest,error)
func pitrCopyCommittedProvider(context.Context, *DeterministicProvider) (*DeterministicProvider,error)
func pitrFinalizeObserved(*testing.T, testinfra.C12AuthorityPITRController,
    *DeterministicProvider, pitrOperationFixture, uuid.UUID,
    coordinatorClock) (Receipt,testinfra.C12AuthorityPITRObservedCommit,testinfra.C12AuthorityPITRCommit)
```

`pitrSeedClosure` 从现有 task9SeedProofActivation 抽出：context/DBTX/error，原 helper 保持签名并调用它，维持原crash测试。`pitrSeedCertificate` 从原 seedCertificate抽出、独立接收已有 operationID/预计sequence/nodeID，不 Reserve、不创建子库；每个操作独立node避免重复inventory。`pitrCommitDomain` 抽出既有 Lock→commitment→certificate UPDATE→aux INSERT，原 crash helper 委托且行为不变。三者迁移 SQL 字面量时同步 AST parity owner，不扩大 SQL。

resolver 的 pool 字段/constructor 参数改为 `store.DBTX`，其余 registered methods 保留。现有 fault repository 仅服务原 crash tests；新 bridge 不使用 postgresCrashTransaction、不暴露任何 raw pool。为原事务活跃断言保留可选 repository 字段；新 bridge 测试单独验证 provider 调用不发生在持锁时。

- [ ] **Step 1: bridge/handler 顺序 RED。** 新顶层 `TestAuthorityPITRBridgeUsesControlledTransaction` 用测试本地 `pitrBridgeFakeAccess` 和 `pitrBridgeFakeTx` 实现窄接口；只覆盖调用所需 DBTX，其他方法计数后返回 finite错误。断言私有 begin 返回相同动态对象（不是 wrapper），Coordinator.transact operation正常返回后 Commit恰好1次；operation失败0次Commit、一次Rollback。定义 test fake的 Commits/Rollbacks/Events 字段及 Begin 返回现有 tx，不能嵌入真实pgx.Tx。

```go
tx, err := repository.beginAuthorityTransaction(t.Context())
if err != nil || tx != access.tx { t.Fatal("bridge replaced exact controlled transaction") }
if _, ok := any(tx).(pgx.Tx); ok { t.Fatal("bridge exposed full pgx.Tx") }
```

另以 AST 对所有复用 handler 的方法体拒绝 `.Commit/.Rollback/Begin` 调用、类型断言到完整 pgx.Tx；动态 trace证实该测试输入没有提前 Commit，不能把 AST当唯一证据。运行 `go test -tags=integration ./internal/nodecontrol/authority -run '^TestAuthorityPITRBridgeUsesControlledTransaction$' -count=1 -timeout=2m`，记录实现前 RED。

- [ ] **Step 2: 实现 bridge 和绑定 activator。** 新 `pitrObservedResolver` 嵌入真实 `*postgresCrashEffectResolver`，仅覆盖 Activate，先跑全部原certificate/audit/outbox写入，成功后 Bind当前exact窄tx：

```go
type pitrObservedResolver struct {
    *postgresCrashEffectResolver
    controller testinfra.C12AuthorityPITRController
    observation testinfra.C12AuthorityPITRObservation
}
func (r *pitrObservedResolver) ActivateAuthorityEffect(ctx context.Context, db store.DBTX,
    receipt Receipt, proof ValidatedActivationDecisionEvidence) error {
    err := r.postgresCrashEffectResolver.ActivateAuthorityEffect(ctx,db,receipt,proof)
    if err != nil { return err }
    tx,ok:=db.(testinfra.C12AuthorityTransaction)
    if !ok { return ErrInvalidArgument }
    if err=r.controller.BindCommitObservation(ctx,r.observation,tx); err!=nil {
        return err
    }
    return nil
}
```

该签名与现有 registered activator 一致。实际 Coordinator 先写 ActivateCommitted，再调用 handler；Bind 之后它执行既有 persisted outcome 读取/比对，然后消费 admission 并唯一 Commit，没有新增pre-Commit钩子。callback返回后再Observe，不能在持锁时drain。`pitrFinalizeObserved` 的具体接线为：

```go
func pitrFinalizeObserved(t *testing.T, c testinfra.C12AuthorityPITRController,
    provider *DeterministicProvider, op pitrOperationFixture, activationID uuid.UUID,
    clock coordinatorClock) (Receipt,testinfra.C12AuthorityPITRObservedCommit,testinfra.C12AuthorityPITRCommit) {
    t.Helper()
    observation,err:=c.NewCommitObservation()
    if err!=nil { t.Fatal(err) }
    var receipt Receipt
    err=c.WithPrimaryAuthorityAccess(t.Context(),func(access testinfra.C12AuthorityAccess) error {
        ctx:=t.Context()
        raw,err:=NewPostgresRepository(access)
        if err!=nil { return err }
        repository:=&pitrAuthorityRepository{PostgresRepository:raw,access:access}
        resolver:=&pitrObservedResolver{postgresCrashEffectResolver:newPostgresCrashEffectResolver(access),
            controller:c,observation:observation}
        coordinator:=mustNewCoordinatorForTest(t,provider,repository,resolver,clock)
        reservation,err:=coordinator.Reserve(ctx,op.request)
        if err!=nil { return err }
        if reservation.Sequence!=op.sequence { return ErrConflict }
        tx,err:=access.Begin(ctx)
        if err!=nil { return err }
        defer tx.Rollback(ctx)
        if err=raw.RecordPendingClaimV1(ctx,tx,ClaimV1Reservation{Reservation:reservation,
            ProtocolActivationID:activationID},clock.Now().Add(-time.Second)); err!=nil { return err }
        digest,err:=pitrCommitDomain(ctx,tx,raw,reservation,op.nodeID)
        if err!=nil { return err }
        if err=tx.Commit(ctx); err!=nil { return err }
        receipt,err=coordinator.Finalize(ctx,CoordinatorFinalizeRequest{OperationID:op.request.OperationID,EffectDigest:digest})
        return err
    })
    if err!=nil { t.Fatal(err) }
    observed,facts,err:=c.ObserveCommit(t.Context(),observation)
    if err!=nil { t.Fatal(err) }
    return receipt,observed,facts
}
```

callback最多30秒，access 将上述t.Context绑定到自己的较短lease，不能延长I/O。实际primary Ready/原子结果断言在后续独立access执行，避免持有30秒能力跨越备份/观察；没有重新Finalize来探测Commit。

- [ ] **Step 3: 抽取 fixture 并建立真实前缀。** 顶层固定P/A/B三个operationID（uuid.NewSHA1固定测试namespace+"prefix"/"cut"/"revoke"）、三个独立node，epoch31 sequence1/2/3；同一个protocol activation使用原`79000000-0000-4000-8000-000000000001`。Open已建立aux并验证outbox来自00004，消费者不CREATE。仅在backup前setup事务seed closure、三个历史epoch1依赖及active certificate，恢复origin再Commit。CaptureDatabasePoint走固定primary查询。此时历史fence已非空，不能用空provider声称ready：先完成真实P(seq1) activation，再真实CheckReady必须Ready且PendingCount=0。

每个操作目标 certificateID=uuid.NewSHA1(operationID,"certificate")。A与B的Reserve只在各自真正提交阶段发生，不为预置历史依赖提前推进provider。旧helper中reservation.Sequence用于历史数据的地方改用fixture预计sequence，协议执行后断言实际等于预计值。

- [ ] **Step 4: 用 replay 建独立 provider-A 对照。** 不改原provider、不浅拷贝其mutex/map：在A完成时读Snapshot，按sequence有序，调用一个新NewDeterministicProvider(31)的Reserve/Finalize重放同请求/point；每条reservation和receipt逐字段相同，否则失败。拒绝pending/aborted记录，本场景只含P和A：

```go
func pitrCopyCommittedProvider(ctx context.Context, src *DeterministicProvider) (*DeterministicProvider,error) {
    head,err:=src.Head(ctx)
    if err!=nil { return nil,err }
    dst,err:=NewDeterministicProvider(head.Epoch)
    if err!=nil { return nil,err }
    for _,record:=range src.Snapshot() {
        receipt:=record.TerminalReceipt
        if receipt==nil || receipt.Status!=StatusCommitted || receipt.DatabasePoint==nil || receipt.EffectDigest==nil {
            return nil,ErrInvalidArgument
        }
        reservation,err:=dst.Reserve(ctx,ReserveRequest{OperationID:record.OperationID,Kind:record.Kind,ScopeKind:record.ScopeKind,ScopeDigest:record.ScopeDigest})
        if err!=nil || reservation!=record.Reservation { return nil,ErrConflict }
        point:=receipt.DatabasePoint
        copied,err:=dst.Finalize(ctx,FinalizeRequest{OperationID:record.OperationID,EffectDigest:*receipt.EffectDigest,
            DBSystemID:point.SystemID,DBTimeline:point.Timeline,RequiredLSN:point.RequiredLSN})
        if err!=nil || !receiptEquals(copied,*receipt) { return nil,ErrConflict }
    }
    return dst,nil
}
```

现有Snapshot按sequence排序；用单元测试固定该前提及无alias，推进原provider到B后dst.Head仍A。不引入生产provider回退能力。

- [ ] **Step 5: 替换旧物理测试，先取得真实 RED（环境可用时）。** 删除本文件的pitrDockerHarness、exec/net/rawpool、密码/端口选择、raw00006安装、自己cleanup及无关fresh-cluster分支。唯一顶层仍叫`TestPITRBeforeRevocationFailsClosed`，流程：

```go
controller,err:=testinfra.OpenC12AuthorityPITR()
if err!=nil { t.Fatal(err) }
// setup + P + primary Ready 已由 Step 3 的固定 helper 执行。
backup,err:=controller.CreateBaseBackup(t.Context())
if err!=nil { t.Fatal(err) }
receiptA,observedA,factsA:=pitrFinalizeObserved(t,controller,provider,operationA,activationID,clock)
selection,err:=controller.SelectRecoveryCut(backup,observedA)
if err!=nil { t.Fatal(err) }
providerA,err:=pitrCopyCommittedProvider(t.Context(),provider)
if err!=nil { t.Fatal(err) }
receiptB,observedB,factsB:=pitrFinalizeObserved(t,controller,provider,operationB,activationID,clock)
if factsA.EndLSN==factsB.EndLSN { t.Fatal("B must be later than A") }
cut,err:=controller.CrashPrimaryAtCut(t.Context(),selection,observedB)
if err!=nil { t.Fatal(err) }
if cut.RecoveryTargetLSN!=factsA.EndLSN { t.Fatal("B replaced A") }
candidate,err:=controller.RestoreAtCut(t.Context(),cut)
if err!=nil { t.Fatal(err) }
candidate,err=controller.PromoteCandidate(t.Context(),candidate)
if err!=nil { t.Fatal(err) }
timeline,err:=controller.InspectTimeline(t.Context(),candidate)
if err!=nil || !timeline.Promoted { t.Fatalf("promotion: %#v %v",timeline,err) }
```

上段变量来自Step3 fixture（provider、operationA/B、activationID、clock）且receiptA/B在下一步使用，不保留unused变量。借真实gate先见旧测试所述缺陷或新增断言RED；若Docker缺失，仅记录阻塞，不能用连接失败替代业务RED。

- [ ] **Step 6: 加恢复后的严格状态与双对照。** 在同一 candidate access创建真实PostgresRepository+bridge+registeredresolver：A fence claim_v1/committed/active，A certificate/revoke proof/audit/outbox恰好一次；B fence/aux/audit/outbox不存在，B certificate status=active、revokeoperation/JCS/evidence/resolution全NULL。用上表快照SQL和已有validate查询，不增加rawSQL逃逸。状态读取在ready前执行，避免把wrapper权限失败误算成业务拒绝。

```go
behind,err:=coordinatorB.CheckReady(ctx)
if err!=nil || behind.Ready || behind.Reason!=ReadinessDatabaseBehindProvider {
    t.Fatalf("provider B must be exactly behind: %#v %v",behind,err)
}
ready,err:=coordinatorA.CheckReady(ctx)
if err!=nil || !ready.Ready || ready.Reason!=ReadinessReady {
    t.Fatalf("same candidate provider A must really be ready: %#v %v",ready,err)
}
```

coordinatorA/B由同access/bridge创建，仅provider不同；provider B 的Head/receipt在crash前和两次readiness后逐字段相同，A-copy的Head为receiptA。assert timeline前进和systemID不变。candidate DML用固定certificate UPDATE常量发起须wrapper提前拒绝且随后快照未变；driver权限实测检查readonly=false、superuser=false、准确column UPDATE可rowlock，不给消费者任意权限探针SQL（testinfra私有验证连接做固定检查）。保留一项测试内readiness-gated读取失败返回零数据，但明确不是Task10同连接serving交付。

- [ ] **Step 7: GREEN、crash 回归、提交。** 先运行bridge单元、policy source parity和integration编译；真实数据库可用时由Task6所列runner分别跑迁移测试与以下6个原crash顶层，每个使用`-Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^精确名称$' -Timeout 5m`：`TestCoordinatorPostgresCrashRecoveryMatrix`、`TestCoordinatorPostgresAbortDomainSafetyAndIdempotence`、`TestCoordinatorPostgresCrashResolverFailsClosedOnSQLDeletionOrCorruption`、`TestCoordinatorPostgresAtomicActivationCrashMatrix`、`TestCoordinatorPostgresFenceFirstAbortRace`、`TestCoordinatorPostgresFinalizeCapturesAfterEffectWriterCommit`。原故障注入、重复域效果、Abort锁竞争不因helper抽取改变。提交 `test(authority): migrate PITR regression to controlled real commits`。Docker缺失时只提交明确未验收checkpoint，不能关闭Task9。

## Task 6: Closed Registration, Verification, and Evidence

**Files:** File Map Task6，docs；不增加CI、后台服务、安装Docker或修改runner执行/清理架构。

**Interfaces:** runner仍用现有Focused参数；仅登记`TestPITRBeforeRevocationFailsClosed`为authority-v7-pitr专属，保留`TestC12AuthorityPITRProfile`和私有failure-seam。每个controller消费者单独调用；同一group不并列两个会Open/Crash的顶层测试。

- [x] **Step 1: 写普通静态 guard RED。** 扩展 `TestC12RunnerAuthorityProfilesAreClosed` 和新增 `TestC12PITRConsumerInterfacesAreIntegrationOnly`：读取本计划列出的新文件，首行tag精确匹配；go/build普通构建排除它们；AST枚举controller公开方法与有限新types，仅为上面冻结集合；authority bridge只在`_test.go`。读旧PITR文件确认无os/exec、net、rawpool、密码、raw迁移加载与Docker cleanup符号。source guard解析调用/方法集，不依赖一个易绕过的substring。

```go
for _, path := range []string{
    "c12_authority_access_integration.go", "c12_authority_rows_integration.go",
    "c12_authority_sql_policy_integration.go", "c12_authority_observation_integration.go",
    "c12_authority_cut_integration.go",
} {
    raw,err:=os.ReadFile(path)
    if err!=nil { t.Fatal(err) }
    if !bytes.HasPrefix(raw,[]byte("//go:build integration\n")) &&
       !bytes.HasPrefix(raw,[]byte("//go:build integration\r\n")) { t.Fatal(path) }
}
```

测试runner注册规则：authority测试在base/authority-v7拒绝；exact pitr profile接收；两个controller consumers同组拒绝；无关既有精确选择仍允许。执行普通`go test ./internal/testinfra -run '^(TestC12RunnerAuthorityProfilesAreClosed|TestC12PITRConsumerInterfacesAreIntegrationOnly)$' -count=1 -timeout=5m`，先记录RED。

- [x] **Step 2: 仅修改runner closed登记。** 把现有单一public-test判断扩为固定name→package映射；保持prepared artifact、HMAC、five-leafrunroot、Dockerendpoint、inherited-env、deadline、finally cleanup原样。检查selectedgroup中属于该映射的数量<=1，所有这些测试要求pitr profile；私有failureseam仍必须exact单测且指定现有五值之一。写固定注册内容：

```powershell
$script:c12PITRConsumerTests = @{
  'TestC12AuthorityPITRProfile' = './internal/testinfra'
  'TestC12AuthorityPITROwnershipWALFailureSeam' = './internal/testinfra'
  'TestPITRBeforeRevocationFailsClosed' = './internal/nodecontrol/authority'
}
```

当前runner的`-Race`未加入最终goArgs，不能拿它当race证据；本计划不顺手修此缺陷，race只用下面直接无Docker精确命令，留明确记录。

- [x] **Step 3: 无Docker验证。** 从worktree根运行，每条保留命令/exit code/测试名/耗时；真正RED/GREEN在各task已记录，不允许只最后一次大测：

```powershell
go test -tags=integration ./internal/testinfra -run '^(TestC12AuthorityPITRScopedAccessPolicy|TestC12AuthorityPITRSQLPolicy|TestC12AuthorityPITRSQLRegistryMatchesSources|TestC12AuthorityPITRCommitObservationStateMachine|TestC12AuthorityPITRRecoveryCutStateMachine)$' -count=1 -timeout=3m
go test -tags=integration ./internal/nodecontrol/authority -run '^TestAuthorityPITRBridgeUsesControlledTransaction$' -count=1 -timeout=2m
go test -race -tags=integration ./internal/testinfra -run '^(TestC12AuthorityPITRScopedAccessPolicy|TestC12AuthorityPITRCommitObservationStateMachine|TestC12AuthorityPITRRecoveryCutStateMachine)$' -count=1 -timeout=3m
go test -race -tags=integration ./internal/nodecontrol/authority -run '^TestAuthorityPITRBridgeUsesControlledTransaction$' -count=1 -timeout=2m
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
go test ./internal/nodecontrol/authority ./internal/readiness -count=1 -timeout=5m
go test ./... -count=1 -timeout=15m
```

完整ordinary suite既有testinfra约9分钟，15m是普通suite的每包timeout，不修改runner物理/cleanup预算。compile-only明确标注；race不可用单独记阻塞，不quiet skip。严禁不带精确-run就执行integration包全体测试。

- [ ] **Step 4: 权威物理gate。** 先检查Docker工具/daemon及现有runner前置条件。只暂存已审阅本任务文件以满足runnerstaged-tree证据；不要暂存无关scratch。沿用合法干净子进程环境，不从runner代码移除inherited-env guard。下面每条独立调用，不能把多个顶层以alternation合在同组：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITRProfile$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/nodecontrol/authority' -Run '^TestPITRBeforeRevocationFailsClosed$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITROwnershipWALFailureSeam$' -Timeout 5m -PITRFailureSeam 'after-intent-before-create'
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITROwnershipWALFailureSeam$' -Timeout 5m -PITRFailureSeam 'after-create-before-actual'
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITROwnershipWALFailureSeam$' -Timeout 5m -PITRFailureSeam 'after-actual-before-return'
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITROwnershipWALFailureSeam$' -Timeout 5m -PITRFailureSeam 'after-clean-intent-before-remove'
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7-pitr -Packages './internal/testinfra' -Run '^TestC12AuthorityPITROwnershipWALFailureSeam$' -Timeout 5m -PITRFailureSeam 'after-remove-before-clean-result'
```

要求go-json明确PASS非skip、runnercleanup/foreigncanary检查通过、无非owned删除。迁移测试物理证据必须包含P ready、A/B真实terminal、A数据有/B无、candidate timeline前进、精确database_behind_provider和同candidate Ready。若工具缺失/依赖校验超时/预算耗尽，保留失败证据并停止该gate；不放宽deadline、不用原harness替代、不标完成。

- [x] **Step 5: 更新事实记录。** 在current-status新增本计划链接、准确提交/测试证据、legacy harness替换范围及剩余阻塞。删除fresh-unrelated physicalcluster分支的覆盖由既有`TestAuthorityReadiness` identity-mismatch单元保持，明确它不是等价物理覆盖。Task10证书/desired serving与reconcile恢复仍未交付；B01未完成，除非所有本计划gate和后续原计划要求真正满足。不能把实现完成、可编译、live验收混写。

- [ ] **Step 6: 独立最终review及提交推送。** 使用requesting-code-review检查整个实施diff，特别ReviewFocus五项、candidate列权限、marker→consume→Commit、原crash抽取回归、没有放大公开能力。Critical/Important先修复再对应复测。只在实际验证支持时声明完成；否则提交标为checkpoint并列未通过gate。提交 `test(c12): close controlled PITR gates and record evidence`，推送当前codex分支；不创建PR、不合并、不删worktree。Git索引/网络权限不足时按既有审批流程请求，不绕过。

## Plan Self-Review and Handoff

本计划的自检由主代理执行，不派子代理代审计划：

- [x] 按spec§4–7核对句柄、64/1/8容量、30秒/65536/64MiB/1MiB、完整transaction块、同进程retry、poison、窄DBTX、物化、候选角色与生命周期，覆盖于Tasks1–4。
- [x] 按spec§8–9核对真实P前缀→backup→A→B→restoreA、同候选正向对照、profile-only、prepared/cleanup兼容及ordinary/race/review，覆盖于Tasks5–6；未把计划中的gate当作执行证据。
- [x] 已按source修正activator为`ValidatedActivationDecisionEvidence → error`、确认controller返回值、sqlc参数数量、Snapshot排序及6个crash测试名称；补充pre-backup P、事务ledger overlay和run generation语义。
- [x] 已检查接口/步骤及ReviewFocus归属；aux表/角色在Open固定建立，避免旧profile未调用authority setup时被新backup前置条件拒绝。
- [x] 保存文档并交用户审阅，确认执行方式后才开始Task1。

建议逐任务子代理实现与独立审查：六个task共享能力身份、事务时序与清理边界，逐步审查比只在末尾发现接口偏差更合适。也可由主代理顺序实施、最后独立审查；这是本轮需用户选择的执行方式，不由过去的“1”自动推定。

## Execution record — 2026-09-19 (development checkpoint, not acceptance)

The user approved this written plan and subagent-driven execution with “是”
after the `d8b8e014` handoff. The pre-implementation approval condition was
satisfied before Task1. Implementation ran only in the named isolated worktree
and branch; main was not modified. The historical task text above remains the
approved sequence; this ledger distinguishes work performed from gates accepted.

- [x] Task1 implementation and three fix rounds, `d8b8e014..29e03bc8`: narrow
  access/eager rows/lifecycle. Independent source review closed all seven initial
  Important findings and follow-on findings; ordinary failures below remain evidence.
- [x] Task2 implementation, `816fa7fa`: fixed SQL/argument capabilities and
  candidate grants; independent source review approved. No physical privilege claim.
- [x] Task3 implementation and legacy poison fix, `816fa7fa..374a8098`:
  ordinary commit observation and bounded destructive drain; independent source review
  approved after fixing semantic rejection outside the legacy poison scope.
- [x] Task4 source implementation, `0cb6c720`: earlier A target/B crash boundary,
  retained candidate identity and private pre-callback privilege verification.
  Source review found no Critical/Important source defect; its Important ordinary
  validation finding remains OPEN under U1. It is not a quality-acceptance approval.
- [x] Task5 implementation, `e3b5ed93726dd0b7d02d8a609ae24963e48087fb`:
  real controlled P/A/B authority bridge, shared fixture extraction and required
  22-owner AST parity (including all three snapshot owners); source review approved
  only for an unaccepted development checkpoint. Frozen ordinary26567 PASS below.
- [x] Task6 implementation in the commit containing this ledger: exact three-name
  PITR-only package registry and common pre-resource selection boundary; actual
  PowerShell and ordinary integration-only API guards. Base is `e3b5ed93`;
  see Task6 source-freeze hashes and command results below.
- [ ] Physical acceptance: all 13 calls below unavailable/not executed.
- [ ] Root independent Task6 review, whole-branch review and authorized checkpoint
  push: pending at this implementation handoff. No PR, merge or worktree deletion.
- [ ] Task9/B01 acceptance and Task10 serving/reconcile recovery: not delivered.

### Transparent amendments and coverage boundaries

R17 corrects an original plan defect, not the approved semantic requirement:
the original Task4 text prescribed `recovery_target_inclusive=on`. PostgreSQL
18.4 logical callbacks emit `txn->end_lsn`, the end of the commit record, whereas
recovery compares the record-start `ReadRecPtr`; inclusive on can replay the first
following record. Retain the unmodified terminal EndLSN and use off/false so A is
included and the following record is not replayed. This is an inference from
[REL_18_4 logical.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/replication/logical/logical.c),
[logicalfuncs.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/replication/logical/logicalfuncs.c),
[xlogrecovery.c](https://github.com/postgres/postgres/blob/REL_18_4/src/backend/access/transam/xlogrecovery.c)
and the [PostgreSQL18 recovery-target documentation](https://www.postgresql.org/docs/18/runtime-config-wal.html#RUNTIME-CONFIG-WAL-RECOVERY-TARGET).
Task4 asserts actual ordinary and legacy prepared restore arguments. Source proof
and those tests do not replace the unavailable real A-present/B-absent acceptance.

R20 retains the private two-column unsigned SystemID/timeline identity projection
while the candidate is paused in recovery. Current insert/flush WAL functions
cannot run during recovery; original frozen identity SQL remains for primary and
promoted stages. Replay position is checked separately before promotion. This
adds no consumer query, grant, connection or public API. See
[PostgreSQL18 administration functions §§9.28.3–4](https://www.postgresql.org/docs/18/functions-admin.html).

Task5 removed the old authority-owned Docker/exec/network/password/raw-pool/raw
migration loading and cleanup harness. The replacement exercises real
PostgresRepository/registered handler/controller transactions, pre-backup P Ready,
actual A/B terminals, A recovery with B archive/crash boundary, exact recovered
domain/null state, B-provider `database_behind_provider`, and same-candidate
A-provider Ready. These assertions are implemented, not physically verified here.
The unrelated fresh-cluster physical branch was removed. Existing
`TestAuthorityReadiness` identity-mismatch unit coverage remains, but is NOT
equivalent physical coverage and does not establish fresh-cluster behavior.
R22's memory adapter proves provider/lock scheduling only; it is not SQL,
transaction-atomicity or physical acceptance. Fixed SQL authorization does not
replace the trusted real activation handler or enforce a second activation protocol.
Task10 authorized-certificate/desired-state serving and reconcile recovery remain
outside this delivery.

### Historical verification retained

All native commands below used process-local `GOOS=windows`; no persistent Go
configuration, ACL, concurrency, harness timeout or runner allowance was altered.

| Source/checkpoint | Command/result and disposition |
| --- | --- |
| Initial default elevated baseline74809 | `go test ./...` EXIT1 before assertions: GOOS=linux/GOHOSTOS=windows produced invalid Win32 executables. Process-local native correction only. |
| Native pre-change baseline93669 | `go test ./...` EXIT1; testinfra677.290s, watchdog30s helper timeout. Exact unchanged watchdog reruns PASS18.64s and two18.70s/18.87s; not a full-suite pass or proven cause. |
| Task1 initial implementation, before later review fixes | Intermediate ordinary full PASS; then-final tree ordinary full failed prepared-artifact previous-digest CreateProcessW Win325. Unchanged exact previous-digest passed. Both retained; no all-green claim. |
| Task2 ordinary92487 and final `816fa7fa` | `go test ./... -count=1 -timeout=15m` EXIT0, testinfra584.942s. An integration-test registry mutation guard was added during that run, and the run preceded the integration-only domain-separated commitment-digest correction; it is NOT a frozen-final-tree gate. Final focused2.285s/race3.518s(no race report)/compile testinfra0.254s+authority0.236s PASS after the fix; physical acceptance not run. |
| Task3 frozen1641 | `go test ./... -count=1 -timeout=15m` EXIT1; testinfra595.495s, watchdog30s timeout and prepared-cleanup-retry CreateProcessW Win325. Unchanged serial exact reruns PASS watchdog19.80s/package20.931s and cleanup3.75s/package6.360s. No cause established. |
| Task4 frozen80961, `0cb6c720` tree | Same ordinary15m command EXIT1; transient-validator test220.58s exceeded unchanged PowerShell150s helper, then package timeout/testinfra903.139s. |
| Root unchanged focused21969 | `go test ./internal/testinfra -run '^TestC12Batch01SuiteRejectsTransientCandidateTestMainValidatorBypass$' -count=1 -v -timeout=5m` EXIT0, test130.82s/package132.117s. Not proof of cause or erasure of80961. |
| R21 unchanged frozen26986, same `0cb6c720` | Same ordinary15m command EXIT1: `TestOpenHonorsCancelledContext`1.2828222s>1s (platform18.393s), and `TestC12PreparedProtocolRejectsMalformedFrames/oversized-frame` CreateProcessW Win325 (parent173.13s/sub3.47s, testinfra583.287s). No package timeout that round; no source/budget/permission change. Important validation finding stays OPEN under U1. |
| Task5 frozen26567, `e3b5ed93` source | Same ordinary15m command EXIT0, all packages PASS/testinfra563.456s; all four frozen hashes matched. Final focused authority0.319s/testinfra0.277s, direct race1.476s/1.686s and compile0.337s/0.322s PASS. This fresh PASS does not explain or erase older failures. |

### Task6 exact verification and source identity

Task6 RED: ordinary exact registration/API pair EXIT1,testinfra4.138s:
`TestC12AuthorityPITRProfile/base allowed=False reached=True` at the actual
PowerShell first-resource sentinel. Minimal common guard GREEN3.154s. Existing
pre-run-root cleanup harness then failed for its former empty/all selection
(EXIT1,testinfra5.715s); narrowing only its call to exact public PITR selection
restored the original forced-initializer/cleanup assertions (combined GREEN5.845s).
Final API/DTO/signature and actual PS guard pair PASS2.473s.

Source/test freeze: `2026-09-19T17:37:17.4497348+04:00`, base `e3b5ed93`.
No source/test edits during ordinary81642. Documentation appended after results,
not relabeled as part of the earlier source freeze.

| Frozen path | SHA256 |
| --- | --- |
| internal/nodecontrol/authority/v7_migration_grant_test.go | ABFC0667D67D97581233091300E097A1D56B2D2C1EE2CA24A2BB7E4ADB2288AC |
| internal/testinfra/c12_integration_manifest_test.go | 97329E6C186D1D74D55AB549F8E930CB906D6EF493E3C7400DA018E8CBD34E91 |
| scripts/run-c12-integration.ps1 | 047B07EEC9615FC43AA112E867AF861493309D4C91083A7A2755F0665D42B359 |

Commands executed from the worktree root; each exact no-Docker command exited0.
The integration compile-only line executes no tests.

```powershell
$env:GOOS='windows'
go test ./internal/testinfra -run '^(TestC12RunnerAuthorityProfilesAreClosed|TestC12PITRConsumerInterfacesAreIntegrationOnly)$' -count=1 -timeout=5m
go test -tags=integration ./internal/testinfra -run '^(TestC12AuthorityPITRScopedAccessPolicy|TestC12AuthorityPITRSQLPolicy|TestC12AuthorityPITRSQLRegistryMatchesSources|TestC12AuthorityPITRCommitObservationStateMachine|TestC12AuthorityPITRRecoveryCutStateMachine)$' -count=1 -timeout=3m
go test -tags=integration ./internal/nodecontrol/authority -run '^TestAuthorityPITRBridgeUsesControlledTransaction$' -count=1 -timeout=2m
go test -tags=integration ./internal/nodecontrol/authority ./internal/testinfra -run '^$' -count=1 -timeout=5m
go test ./internal/nodecontrol/authority ./internal/readiness -count=1 -timeout=5m
$env:CGO_ENABLED='1'; $env:CC='C:/Programs/mingw64/bin/gcc.exe'
go test -race -tags=integration ./internal/testinfra -run '^(TestC12AuthorityPITRScopedAccessPolicy|TestC12AuthorityPITRCommitObservationStateMachine|TestC12AuthorityPITRRecoveryCutStateMachine)$' -count=1 -timeout=3m
go test -race -tags=integration ./internal/nodecontrol/authority -run '^TestAuthorityPITRBridgeUsesControlledTransaction$' -count=1 -timeout=2m
```

Results in order: guard2.473s; behavior testinfra2.799s; bridge authority0.318s;
compile-only authority0.324s/testinfra0.367s; ordinary authority10.336s/readiness2.822s;
native direct race testinfra6.337s/authority1.384s with no race report.
Runner `-Race` remains unforwarded and is NOT evidence of race coverage.

Final ordinary81642 used its own process-local `GOOS=windows` child:
`go test ./... -count=1 -timeout=15m`. EXIT0; all packages passed, testinfra587.277s/trust4.295s/trustclient2.024s. Postrun17:48:21.997+04 hashes all MATCH the three values above.
The15m ordinary per-package limit is not a change to the5m physical gates.

### Physical gate prerequisites and all unavailable calls

2026-09-19 recheck: `Get-Command docker.exe -CommandType Application` found no
Docker executable, so daemon/endpoint/image identity could not be checked.
PowerShell/Git/Go exist; Go reports1.26.5 windows/amd64. Existing gcc is present.
The module declares goose3.27.1 as a Go tool; optional `go tool -n goose` lookup
produced no output while waiting approximately3m and was interrupted (EXIT1).
Goose/dependency resolution is unverified, not a runner-gate PASS or timeout.
Ambient inherited GIT_* names were present; a future run needs the unchanged
legal clean-child environment. No guard was removed and no physical runner was
launched/staged-tree receipt acquired after Docker was found absent. No installation,
legacy harness substitute, deadline extension or non-owned deletion occurred.

All seven separate PITR commands in Task6 Step4 remain UNAVAILABLE / NOT EXECUTED
(not failed/skipped/passed): public testinfra profile; authority migration; and
private seams after-intent-before-create, after-create-before-actual,
after-actual-before-return, after-clean-intent-before-remove,
after-remove-before-clean-result. Each retains its own5m budget.

The six preserved crash regressions also remain UNAVAILABLE / NOT EXECUTED; each
must run separately with unchanged authority-v7 profile and5m budget:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresCrashRecoveryMatrix$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresAbortDomainSafetyAndIdempotence$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresCrashResolverFailsClosedOnSQLDeletionOrCorruption$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresAtomicActivationCrashMatrix$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresFenceFirstAbortRace$' -Timeout 5m
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-c12-integration.ps1 -Profile authority-v7 -Packages './internal/nodecontrol/authority' -Run '^TestCoordinatorPostgresFinalizeCapturesAfterEffectWriterCommit$' -Timeout 5m
```

Without those calls, explicit non-skip go-json PASS, cleanup/foreign-canary and
exact ownership checks, real row-lock privilege sufficiency, P Ready, A/B actual
terminals, A-present/B-absent data, timeline advance and B-denied/A-Ready on the
same candidate are all unverified. Unit/static/compile evidence cannot close them.

### Ordered decisions and explicit user exception

The following ordered record preserves the root's reasons and cost-if-wrong.
U1 is an explicit user exception; it is not a root waiver, general permission to
suppress new source findings, acceptance, or authority to merge main.

- Ruling R1: Introduce the minimum private candidate identity carrier with Task1 (owner, run generation, index), leaving its cut association/issuance to Task4; consumer candidate access denies absent binding. Task1's real default SQL policy is deny-all; tests may inject only private backend/fixed-fixture authorization to exercise real adapter behavior — Task1 otherwise cannot compile or test its adapter before Tasks2/4 — cost if wrong: cross-task type churn or an unintended authorization path, caught by API/identity tests and reviews.
- Ruling R2: The approved spec's cancellation deadline and immediate sole driver Commit both apply. Do not blindly ignore a supplied Commit context as the illustrative plan pseudocode does. Pass that context directly to the sole driver Commit while a lease guardian, established before admission, independently cancels/closes the private connection on the earlier lease/parent deadline. No context merge, clock query, SQL, WAL, or hook is inserted between admission and driver Commit — preserves per-call cancellation and the harder ordering rule — cost if wrong: cancellation leak or an extra pre-Commit operation; explicit event/cancellation/race tests must catch it. If this cannot be implemented safely in pgx, escalate to controller before changing scope.
- Ruling R3: Preserve the spec-required exact SQL/source/build-tag/API parity guards, with runtime rejection tests and real positive controls alongside them — these enforce an explicit compile-time authority boundary rather than substitute source matching for behavior — cost if wrong: brittle guards or missed behavioral coverage, reviewed independently.
- Ruling R4: A missing-symbol compile failure can record interface introduction, but each substantive behavior also needs an assertion-level RED after minimal declarations exist; environment failures never count — reconciles plan examples with TDD evidence — cost if wrong: extra test cycles, not scope expansion.
- Ruling R5: Retain this plan's .superpowers workspace after finishing; never delete it or unrelated scratch — the user-approved plan explicitly prohibits deletion and takes precedence over the generic SDD cleanup step — cost if wrong: small ignored local scratch remains; portable progress stays in tracked docs/commits.
- Ruling R6: Explicitly set GOOS=windows only in the child PowerShell processes used for native Windows tests, because actual elevated Go reports GOOS=linux with GOHOSTOS=windows — restores host-native executable generation without modifying user configuration or runner policy — cost if wrong: host-specific verification would be mislabeled; commands and GOOS/GOHOSTOS evidence are recorded, and Linux/Docker acceptance remains separate.
- Ruling R7: Keep the existing runner as the outer absolute group-deadline enforcer; derive only min(caller,30s) inside the access capability, with earlier parent-context tests. Do not invent a new deadline env/descriptor field — source inspection shows runner:5945 establishes groupDeadline and :6070 bounds Invoke-C12Go by it, while the current descriptor has no transport for that value — cost if wrong: group-bound termination uses the runner's process cutoff instead of a pre-cutoff Go cancellation; runner remains the sole resource cleaner and no total budget is extended.
- Ruling R8: Proceed with Task1 after recording the pre-existing full-suite watchdog timeout and three passing unchanged focused reruns; retain the full-suite failure and require ordinary post-change verification, without editing runner timing or declaring baseline all-green — the failure predates implementation, source timing supports load sensitivity, and the scoped adapter work does not alter those paths — cost if wrong: an existing cleanup defect may remain hidden by focused success; final evidence must retain and separately report any recurrence.
- Ruling R9: Task2 freezes/tests the three new snapshot SQL literals in its own policy while enforcing all existing-source parity. Task5, when it creates the bridge owner, must replace the staged literal check with required bridge AST parity and update moved fixture owners; never silently skip a missing owner — the Task2 plan references a source file created only in Task5, so final-owner parity cannot execute yet without violating file ownership — cost if wrong: bridge drift might escape until the later gate; Task5/6 completion must explicitly verify the transition.
- Ruling R10: Task3 uses the fixed numeric get_changes hint65536 plus independent client row65536/byte64MiB/per-row1MiB caps; successful completion means normal EOF, no stream error, and complete validated transaction blocks for that response, not proof the entire slot is empty. A complete response with no requested marker is not-observed; any partial/error/over-limit response poisons. PostgreSQL18 docs confirm the hint is checked after whole transaction output and can be exceeded (https://www.postgresql.org/docs/18/functions-admin.html#FUNCTIONS-REPLICATION) — resolves the brief's unspecified hint without treating a server hint as a safety bound or looping unbounded drains — cost if wrong: false poison or lost proof; explicit exact-cap/over-cap/incomplete-block tests and the live gate must cover it.
- Ruling R11: Task3 proves poison rejection through existing CrashPrimary and its internal gates; Task4 must add the exact public SelectRecoveryCut/CrashPrimaryAtCut poison-with-zero-Docker assertions when those methods exist, without introducing premature Task4 APIs — Task3's example calls methods only introduced by its dependent task — cost if wrong: a new path could omit poison enforcement; Task4 review must resolve this explicit follow-on obligation before completion.
- Ruling R12: Use the already-frozen GetNodeControlDatabaseIdentity SQL via controller-private access to establish actual system ID/timeline in Task2 for policy validation, and Task4 captures/rechecks that tuple in retained backup/candidate records before authorizing their lifecycle. The descriptor database digest is not a PostgreSQL system identifier; no new public DTO field/SQL permission is needed — baseline backup lacks the facts the approved candidate equality/timeline comparison requires — cost if wrong: identity could be accepted from the wrong checkpoint; tests must reject mismatched retained system IDs/timelines and live acceptance remains required.
- Ruling R13: Task4 may add the minimal unexported controller backend seam needed to substitute/record external Docker, SQL and WAL operations in its tests; default paths keep the existing exact resource execution. Test actual validation/state transitions, not a fake whole crash/cut algorithm — the plan says reuse a recorder, but immutable baseline contains no Docker-capable one — cost if wrong: tests could validate a mock instead of controller behavior; independent review must verify seam placement and unchanged public surface.
- Ruling R14: Bind setup fixture groups in two stages: retain at most3 structurally consistent node/scope/sequence/lineage/history/issuance/attempt/certificate groups from authorized setup; when claim supplies the original operationID, validate every required NewSHA1 derivation before registering that current operation or permitting its domain writes — setup SQL carries derived UUIDs only, while the approved Task5 plan does not supply three concrete original UUIDs to Task2; neither SHA1 inversion nor invented hardcoded IDs/new registration API is justified — cost if wrong: a malformed setup group could become an authorized claim; test partial/conflicting/reused groups and every late-binding mismatch and review independently.
- Ruling R15: Keep Task2 a fixed-SQL/argument/fixture capability, not a replacement activation-sequence validator. Do not add pre-Commit completeness checks or post-commit poisoning solely for intentionally truncated already-authorized activation DML; merge only verified NEW setup/claim identities, retain uncertain-Commit failclosed behavior, and allow legitimate committed pending claims — approved spec7.3 treats handler DBTX usage as trusted fixed test code, while the real Coordinator/handler plus Task5 prove atomic activation and the consume→Commit interval forbids added checks — cost if wrong: a deliberately truncated trusted callback can physically commit partial allowed DML; explicitly document this boundary and never claim wrapper-enforced protocol atomicity/rollback or skip the real atomicity acceptance assertions.
- Ruling R16: Route the already-approved Task5 Step6 testinfra-private candidate privilege verification into Task4's candidate-access lifecycle work. Task2 currently installs grants but only has an observer read-only/TLS probe, not a candidate-role probe; Task5 owns authority tests and must not gain arbitrary probe SQL. Task4 adds a fixed unexported verifier on the existing restricted candidate connection before callback entry, with failure preventing capability exposure, no additional connection/public API/consumer SQL/grants. Verify effective non-superuser, non-read-only and exact required column privileges; actual FOR UPDATE sufficiency remains the real Task5/6 Coordinator gate — resolves a cross-task ownership omission without expanding authority — cost if wrong: private catalog checks could reject a valid restored candidate or falsely certify insufficient grants; negative behavioral tests and real same-candidate Ready remain mandatory, and no live acceptance is claimed without them.
- Ruling R17: Correct Task4's literal recovery_target_inclusive=on to off/false when the unchanged target is the logical terminal EndLSN, for both new ordinary and legacy prepared restore paths. REL_18_4 logical.c reports commit record END, while xlogrecovery.c compares record START before replay for off and after replay for on; on replays one subsequent record. The approved spec section6 requires semantic A inclusion/B exclusion, not that literal on, so off implements its stronger exact-boundary intent without LSN arithmetic or a new API. Task4 tests actual restore arguments; Task5/6 retain real A/B and legacy prepared gates; Task6 corrects the tracked plan with citations — cost if wrong: the selected commit might be omitted or later WAL included; only real data checks can close physical acceptance. Evidence: postgresql18-terminal-boundary-facts.md.
- Ruling R18: Allow private candidate phases6/7 as started/in-flight-or-failed reservations for valid Promote/Inspect, preserving4=successfully promoted and5=successfully inspected; expose4/5 only after success. Full caller identity validation still precedes any phase claim, while an uncertain started external operation must not roll back to an authorizing/retryable state. Baseline grep confirms candidatePhase is only private controller/access state plus one policy fixture, not WAL/runner/public protocol. This avoids the baseline bug that claimed success4 before promotion SQL and avoids redundant success/failure flags — cost if wrong: a valid failed operation can strand a candidate or an overlooked phase consumer can misinterpret it; tests must cover valid4/5, failed6/7 access/retry refusal, zero later effects and unchanged cleanup compatibility.
- Ruling R19: Adapt only Task2's c12TestOriginalCallKind candidate fixture/helpers in c12_authority_sql_policy_integration_test.go to the new exact promoted DTO and private pre-callback verification; do not introduce an optional-unverified candidate admission route merely to retain its zero-open assertion. Record backend events after valid candidate admission, then prove rejected consumer calls add zero Query/Exec/driver events (also after Begin for tx); primary behavior remains. No runtime SQL/grant changes and no opportunistic deferred-Minor edit. Optional nil DTO may mean primary only, never bypass validation for a nonnil candidate binding — cost if wrong: fixture adaptation could weaken authorization-before-driver regression coverage; scoped tests/review must verify event-delta assertions, not merely opens==1.
- Ruling R20: Refine R12's fixed identity-query reuse by adding one unexported two-column systemID/timeline SELECT for the paused pre-promote candidate on its existing private database connection. Preserve the frozen query's signed-to-unsigned CASE projections/ranges, but omit current insert/flush LSN; keep original frozen SQL for backup and promoted stages. PostgreSQL18 functions-admin9.28.3 forbids pg_current_wal_insert_lsn during recovery, while existing cut-wait already verifies replay LSN. Do not defer candidate identity verification until after promotion and do not add consumer SQL/grants/API/connections — cost if wrong: projection drift or a fake-only recovery success could accept wrong identity or fail live; tests must reject insert/flush queries while fake recovery is active, check actual Restore uses the narrow query, unsigned bounds/wrong tuple, and retain real physical gate. Primary source: https://www.postgresql.org/docs/18/functions-admin.html .
- Ruling R21: Review1's only Task4 Important finding is the failed ordinary validation, without an identified source defect. After exact unchanged diagnostic21969 passed130.82s (package132.117s) and helper/stack trace proved only an outer150s timeout before semantic assertions, authorize one frozen full-suite rerun by the original implementer with identical budgets/permissions/concurrency and no source edits. If it passes without a patch, rereview the original immutable diff with appended evidence rather than fabricate an empty commit/fix diff. Preserve80961 failure and unknown causation; final Task6 gate remains required — cost if wrong: another up-to15m run may not reproduce intermittent failure and cannot prove its cause; if still failing, stop speculative retries and route the concrete blocker.
- User direction U1 (2026-09-19, async choice after R21 emitted another ordinary failure): user explicitly selected “继续第 5–6 项，保留验收阻塞” in response to a choice stating only an unaccepted development-branch checkpoint would be pushed and main would not be merged. This overrides the normal per-task quality gate solely to permit remaining implementation; it does not declare either ordinary failure fixed, waive final reporting, authorize broader harness/security changes, or imply final acceptance. Keep Task4 Important validation finding OPEN in final review. Task4 source is spec-compliant with no identified source defect; next implementation is Task5. While26986 runs, its whole source tree remains frozen; Task5 may read/prepare only until root explicitly releases edits after test exit.
- Ruling R22: Allow a narrow Task5 test-local adapter over existing coordinatorMemoryRepository solely for provider-outside-lock scheduling tests using pitrBridgeFakeTx. Root source check confirmed existing memory Lock/readAuthoritySnapshot require exact *coordinatorMemoryTransaction; adapting/delegating existing memory behavior avoids building a second SQL-result simulator for unrelated scheduling. Real Coordinator Finalize/CheckReady and fake active-state guard must execute, with positive method counts and negative guard check; retain separate exact real bridge/transact and dynamic real-handler assertions. No production/existing-memory-test changes or algorithm duplication; actual P/A/B stays real PostgresRepository/handler/controller — cost if wrong: memory adaptation can give false confidence about real SQL/atomicity, so explicitly restrict its claim to scheduling and preserve independent live/handler gates.
