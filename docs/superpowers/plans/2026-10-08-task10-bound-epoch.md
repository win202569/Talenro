# Task 10 Bound Epoch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 私有证书读取器从同连接、同一次前置校验取得 epoch，后置校验成功才发布结果。

**Architecture:** authority 创建带原 DBTX 与原子 epoch 的私有包装，保持冻结回调签名；回调返回时清零 epoch。公开只读访问函数只认确切私有类型；serving 的私有组合函数固定调用真实访问函数，再复用 adapter 与 guardedRead。

**Tech Stack:** Go 1.26.5、现有 pgx v5.10.0、sqlc 生成查询；Windows PowerShell，离线测试。

**Spec:** `docs/superpowers/specs/2026-10-08-task10-bound-epoch-design.md`（已批准；执行者须完整阅读）。

## Global Constraints

- 工作目录：`D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；复用现有工作树，保留此前未提交改动。
- 保留冻结签名：`WithConsistentReadyRead(context.Context, func(context.Context, store.DBTX) error) error`。
- epoch 范围为 `1..MaxInt64`，ProviderHead.Epoch 与 DatabaseHead.Epoch 必须相等。
- 不新增生产 reader/source 工厂、HTTP 路由、数据库迁移或依赖。
- 普通 readiness 仍不能替代 Task 15 的 v7 incarnation、activation 与 lease 证明。
- 不暂存、提交或推送；本计划的本地检查点取代技能模板中的逐任务提交。
- 仅 serving、contracts、readiness 全包；authority 仅 `^(TestBoundRead|TestCoordinator)`；store 仅 `-run '^$'` 编译。
- 不运行真实数据库、全仓库/C11/C12/bootstrap、Docker/PITR，不改安全配置、下载依赖或修复共享缓存。
- 已按 Native 完成本计划的限定增量；测试与审查证据、明确偏差见 `docs/roadmap/2026-10-08-bound-epoch-execution.md`。

## Review Focus

- 额外 DBTX 装饰器即使携带同名方法也不能冒充绑定：Task 1 ExactType 测试。
- 回调内访问 epoch 与失效交错只允许原 epoch 或零值错误，不得数据竞争：Task 1 ConcurrentExpiry 测试及原子状态审查。
- 原 readiness 合法空库归一结果能通过，但包装层不得修正零 database epoch：Task 1 EpochRange 测试。
- 先成功暂存再 panic 必须清理 facts，释放连接并失效绑定：Tasks 1/2 Panic 测试（分层证据）。
- 缺失绑定且用户取消时，取消仍优先；解析器错误不能透出私密字符串：Task 2 Errors 测试。

## 文件与职责

- 新建 `internal/nodecontrol/authority/bound_epoch.go`：私有包装、转发、只读访问与失效。
- 新建 `internal/nodecontrol/authority/bound_epoch_test.go`：包装和实际私有 source 的隔离测试。
- 修改 `internal/nodecontrol/authority/bound_read.go`：前置 epoch 一致性检查与回调包装生命周期。
- 修改 `internal/nodecontrol/authority/bound_read_test.go`：只调整业务回调的对象身份假设，保留 readiness 的原连接断言。
- 新建 `internal/nodecontrol/serving/certificate_reader.go`、`certificate_reader_test.go`：私有组合与分层测试。
- 修改 `internal/nodecontrol/serving/repository.go`：仅补充包装/epoch 契约注释。
- 更新两份 roadmap 文档和本计划；不改 SQL、生成 store 或锁文件。

## 统一验证环境

每次运行测试前，在该工作树 PowerShell 设置：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$taskGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

每条测试命令分别检查退出码；0 且输出 ok 才是通过，非零记录实际失败。
编译缓存需要权限时请求限定提权，不扩大测试范围。缺工具/缓存即报告限制，不下载。

### Task 1: 实际 source 的 epoch 包装与有效期

**Files:** authority 的四个文件，见上表。

**Interfaces:**
- Consumes: `store.DBTX`、`boundReadSnapshot`、`boundAuthorityReadSource.WithConsistentReadyRead`。
- Produces: `BoundReadEpoch(db store.DBTX) (uint64, error)`；私有 `boundEpochDB`，字段 `db store.DBTX`、`epoch atomic.Uint64`；方法 `expire()` 与 DBTX 三个转发方法（签名严格同 store/db.go）。零 epoch 兼任无效标记；不复制原子对象、不导出构造器。

- [x] **Step 1 — 写失败测试。** 在 bound_epoch_test.go 使用已有 boundTestSnapshot/boundTestConnection，以 recording DBTX 替代数据库。新增以下 TestBoundRead 前缀组；断言中的 got/err 来自真实访问函数，计数来自传输替身：

```go
// TestBoundReadEpochBinding: pre/post 都取 epoch 19；回调内：
if got != 19 || err != nil { t.Fatal("lost source epoch") }
// Exec/Query/QueryRow 分别转发一次，ctx、SQL、参数、返回值/错误原样传递。
// pre/post 仍收到原 conn；全流程 acquire==1、release==1。
// TestBoundReadEpochExactType: nil、typed nil、普通 DBTX、同名方法假对象、外层包装：
if got != 0 || err != ErrAuthorityUnavailable { t.Fatal("unbound accepted") }
// TestBoundReadEpochExpiry: 回调返回后、postcheck 内、panic 展开后：
if got != 0 || err != ErrAuthorityUnavailable { t.Fatal("binding survived") }
// TestBoundReadEpochRange: (0,0)、(19,0)、(19,20)、(MaxInt64+1,同值) 不回调；
// (1,1)、(19,19)、(MaxInt64,MaxInt64) 通过；已有归一空库 fixture 通过。
// TestBoundReadEpochConcurrentExpiry: channel/barrier 协调读取与 expire，无 sleep；
if !((got == 19 && err == nil) || (got == 0 && err == ErrAuthorityUnavailable)) {
    t.Fatal("invalid atomic state")
}
```

- [x] **Step 2 — 记录 RED。** 运行 `& $taskGo test ./internal/nodecontrol/authority -run '^TestBoundReadEpoch' -count=1 -timeout=2m`。先确认缺失符号导致编译失败；加最小 fail-closed 声明后重跑，确认正常绑定/转发行为失败，而不是环境错误。
- [x] **Step 3 — 实现包装。** 三个 DBTX 方法只转发 `db`；访问函数确切断言 `*boundEpochDB`，检查非 nil、底层 DBTX 非 nil、atomic.Load 的范围，失败仅有限 sentinel；expire 使用 Store(0)。不得将失效误称为撤销底层 SQL 权限。
- [x] **Step 4 — 接入 source。** 前置快照冻结后检查双 epoch 等值及范围。创建包装、Store(epoch)，用局部同步函数 `defer wrapper.expire()` 包住回调；确保回调返回后、postcheck 前失效，panic 时也失效。前后 readiness 与物理身份检查仍用原 conn。修订旧回调 `db==conn` 断言为真实转发/epoch 断言，不能删掉绑定验证。
- [x] **Step 5 — 验证 GREEN。** 运行 `& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m`。保留已有取消、连接替换、head 漂移、WAL 和错误优先级测试；全选中组必须通过。没有 -race 运行则不得声称 race detector 通过；原子操作与交错测试仅提供相应证据。
- [x] **Step 6 — 本地检查点。** 检查 diff 和格式，记录 RED/GREEN；不提交。按选定执行方式完成该阶段审查或留待最终独立审查。

### Task 2: 私有证书读取组合

**Files:** serving/certificate_reader.go、certificate_reader_test.go、repository.go 注释。

**Interfaces:**
- Consumes: Task 1 的 `authority.BoundReadEpoch`；既有 `guardedRead`、`queryCertificateFacts`、`discardCertificateFacts`。
- Produces: `readCertificateFacts(ctx context.Context, source BoundAuthorityReadSource, lookup CertificateLookup) (CertificateFacts, error)`。
- Private test seam: `readCertificateFactsWithEpoch(ctx context.Context, source BoundAuthorityReadSource, lookup CertificateLookup, resolve func(store.DBTX) (uint64, error)) (CertificateFacts, error)`。主入口固定传 `authority.BoundReadEpoch`；无导出 seam、全局替换或配置注入。

- [x] **Step 1 — 写失败测试。** 复用 adapterDB/adapterRows 与 certificateRowFixture；真实主入口测试无绑定拒绝。成功路径通过私有 resolve 接缝只替代跨包 provenance，实际运行 guardedRead、adapter、生成查询与 facts 校验。

```go
// TestCertificateReaderRequiresBinding: 普通/typed nil/同名方法假对象：
if err != authority.ErrAuthorityUnavailable || db.calls != 0 || !reflect.DeepEqual(got, CertificateFacts{}) {
    t.Fatal("unbound query or facts escaped")
}
// TestCertificateReaderBoundEpoch: resolver 返回 19，fixture 的证书/trust epoch 全改 19：
if args[9] != int64(19) { t.Fatal("wrong query epoch") }
// 同一个 callback DBTX 交给 resolver 及 adapter，resolver 恰好一次；agent/nonagent 均比较完整 snapshot。
// TestCertificateReaderErrors: nil resolver、0/MaxInt64+1、未知私密错误均不查询，返回 unavailable；
// 已取消 context 与上述错误组合返回 ErrCanceled；不暴露错误字符串。
// TestCertificateReaderPostcheck: 后置拒绝 => 零 facts，保留的 serial/DER/trust aliases 全零；
// 后置成功 => 完整 facts；查询取消 => 零 facts；缺失/重复回调 => unavailable。
// TestCertificateReaderPanic: 暂存成功后 source panic，recover 原值且暂存 aliases 全零。
```

- [x] **Step 2 — 记录 RED。** `& $taskGo test ./internal/nodecontrol/serving -run '^TestCertificateReader' -count=1 -timeout=2m`；确认缺失实现，补 fail-closed 声明后确认成功路径断言失败。
- [x] **Step 3 — 实现组合。** 主入口仅转调固定解析器的私有 helper。helper 用 guardedRead 的 query 回调解析 epoch；nil/未知错误/非法数值收敛 unavailable，取消按现有规则优先；同 DBTX 调 adapter，discard 固定真实清理函数。无外部 epoch 参数；不更改底层 adapter 的职责。
- [x] **Step 4 — 验证 GREEN。** `& $taskGo test ./internal/nodecontrol/serving -count=1 -timeout=2m`。补注释说明跨包分层测试与真实生产来源证明的区别；不新增导出测试工厂。
- [x] **Step 5 — 本地检查点。** 核对真实入口始终用 BoundReadEpoch、失败不发 SQL、暂存不提前发布；记录测试输出，不提交。

### Task 3: 最终回归、独立审查与进度落盘

**Files:** `docs/roadmap/current-status.md`、`docs/roadmap/2026-10-08-task10-read-boundary.md`、本计划。

**Interfaces:** 消费 Tasks 1–2 的本地实现与证据；产出限定范围结论，不产出生产工厂或发布。

- [x] **Step 1 — 独立审查。** 按选定执行方式审查本增量及关联契约；最终分支审查可只读查看累积变更，但不得自动扩大修复/测试范围。明确 source epoch 来源、确切类型、生命周期、私有测试接缝、取消与 panic 清理。审查不等于真实 DB 或完整分支验收。
- [x] **Step 2 — 最终新鲜回归。** 分别运行并检查退出码：

```powershell
& $taskGo test ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=2m
& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m
& $taskGo test ./internal/store -run '^$' -count=1 -timeout=2m
```

- [x] **Step 3 — 文件检查。** 固定工具链 gofmt 检查/格式化本增量 Go 文件，`git diff --check`，核对新文件也无尾随空白；只读检查 git status，不暂存。四个锁文件 SHA256 必须仍分别为：
  - go.mod：`A0883514AF23F6947A8CF90F52B20D56AF5FC51BCECEEC2DD7C98082DF59E9E0`
  - go.sum：`642D51BD341B270B35EC976F6196152A33057ED38E54FBCEBF143A50271733D7`
  - tools/devtools/go.mod：`0A8312541FEC1A9FDC6555DB05D1F1D1342EE523B3273A800347714C146A71B9`
  - tools/devtools/go.sum：`85A30247BB0F2A8568D8081A31F029CFA8FEA89C9F2CFE3FF17EEDD2A0340164`
- [x] **Step 4 — 记录结论。** 写明实际 RED/GREEN、测试范围/耗时、审查结论及排除项、仍未提交推送；保留真实数据库、v7、安全验收和之前 fixture minor。完成后无测试后台进程；不宣称 Task 10 或 C12 全部完成。

## 计划自查与执行交接

覆盖映射：来源/类型/有效期/连接转发归 Task 1，组合/清理归 Task 2，回归与安全限制归 Task 3。
仅本计划新增的两项实现任务依赖顺序执行；不存在可并行修改的独立接口任务。
推荐 Native：由当前代理连续实现 Tasks 1–2，最后独立审查；也可选逐任务子代理实施与审查。
用户已批准并选择 Native；限定增量已实施、回归和审查完成。仍未提交、推送或放开安全门禁；分层清理测试裁定和审查排除项见执行记录。
