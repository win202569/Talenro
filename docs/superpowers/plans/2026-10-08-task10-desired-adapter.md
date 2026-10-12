# Task 10 Desired Adapter and Bound Composition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 完成私有 desired 查询生命周期与受保护读取组合，失败时不交付暂存材料。

**Architecture:** 专用扫描包装器接管真实生成查询的解码目标与关闭；真实 converter 构造自有投影。私有 reader 在原始 callback DBTX 上取得 epoch，再调用 adapter，并由既有 guardedRead 控制交付与丢弃。没有公开 Reader 或生产接线。

**Tech Stack:** Go 1.26.5、pgx/v5、现有 sqlc 1.31.1 生成方法、PowerShell 离线测试。

**Spec:** `docs/superpowers/specs/2026-10-08-task10-desired-adapter-design.md`（已确认，执行前完整阅读）。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，保留所有既有未提交改动。
- 使用真实 ListExactDesiredProjection、desiredProjectionFromRows、discardDesiredProjection、guardedRead；不改 SQL、生成文件、schema、证书路径或 authority 工厂。
- epoch 范围 1..MaxInt64；node 必须通过既有 factsUUID 校验；固定 scope transcript 中分隔符为单个 NUL。
- 最多 33 次 Scan；33 行交 converter 返回 Conflict；第 34 行不解码，返回 AuthorityUnavailable，不能接受截断前缀。
- cancellation > source/authority failure > domain error；正常参数前置拒绝不启动 source。
- 成功交付晚于 adapter 清理；关闭 panic 也执行内存清理，不吞掉 panic。
- 不导出入口，不保留 DBTX/context，不新增生产 hook、可配置 epoch 或通用查询框架。
- 不暂存、提交、推送；本地 ledger 替代默认 commit/删除工作区步骤。
- 不运行 DB/Docker/完整 C11/C12/bootstrap/PITR/race/全仓测试，不改安全设置或绕过拒绝，不下载依赖或改锁文件。
- 仅 serving/contracts/readiness、authority `^(TestBoundRead|TestCoordinator)`、store `-run '^$'`。

## Review Focus

- Query 同时返回 rows+error 时生成器直接退出，adapter 仍须关闭：Task 1 Transport。
- Close panic 可能阻止后续清理或错误交付自有投影：Task 1 Panic/Cleanup。
- 底层提供超过 SQL 上限的行，不能无界保存扫描指针：Task 1 RowLimit。
- 包装 DBTX 后再取 epoch 会丢失精确能力身份：Task 2 BindingOrder。
- source 成功后取消它自己的 scoped callback context 不等于请求取消：Task 2 ScopedContext。

## 文件结构

- Create `internal/nodecontrol/serving/desired_adapter.go`：参数派生、真实生成查询、错误映射和 staged 生命周期。
- Create `internal/nodecontrol/serving/desired_scan.go`：专用 DBTX/Rows 包装器、33 行边界、目标清理；不影响证书包装器。
- Create `internal/nodecontrol/serving/desired_adapter_test.go`：真实查询/converter 搭配传输替身和清理断言。
- Create `internal/nodecontrol/serving/desired_reader.go`、`desired_reader_test.go`：固定 epoch 来源的私有 guard 组合及测试。
- Create `docs/roadmap/2026-10-08-desired-adapter-execution.md`：本计划 ledger，记录基线、裁定、RED/GREEN、审查和未验收项。
- Update `docs/roadmap/current-status.md`、`docs/roadmap/2026-10-08-task10-read-boundary.md`、本计划：只记录实际证据。

### Task 1: 私有 adapter 与有界扫描清理

**Files:** desired_adapter.go、desired_scan.go、desired_adapter_test.go、ledger。

**Interfaces:** 消费既有生成接口、行转换与投影；产出：

```go
func queryDesiredProjection(ctx context.Context, db store.DBTX, nodeID uuid.UUID, epoch uint64) (desiredProjection, error)
```

辅助类型 `desiredScanDB` 嵌入 store.DBTX、持有唯一 `*desiredScanRows`；Query 返回接管的包装器。`desiredScanRows` 嵌入 pgx.Rows，记录 targets、closed、成功 Next 数量和边界错误；实现 Next、Scan、Err、Close。`(*desiredScanDB).cleanup()` 重复调用安全。其他成员保持私有。

- [x] **Step 1 — 基线与前置检查。** 核对 worktree 分支/HEAD/既有 dirty 清单；保存四个锁与 SQL、生成 query、Querier 的 SHA256。离线全 serving 基线通过后建 ledger。记录 Tasks 1→2 的 queryDesiredProjection 签名一致，Task 3 消费证据而非新的 API。只读 Git 所有权问题用限定提权，不改 safe.directory。
- [x] **Step 2 — 写正常与参数 RED。** 用 `desiredRowsFixture(t, rotations)`，新增 `TestDesiredAdapterQuery` 覆盖 0/1/32 段完整快照，Query 只调用一次且四参数精确。scope 黄金摘要用独立 .NET SHA256 计算规范 UUID bytes 后固化字面值，勿用 Guid.ToByteArray 的混合字节序，也不调用产品 helper 当期望。`TestDesiredAdapterPreflight` 表测 canceled/nil context、nil/typed-nil DBTX、无效 UUID、epoch 0/MaxInt64+1，精确错误且零查询。

  传输替身 `desiredAdapterDB` 记录 query/args/context，`desiredAdapterRows` 提供全部 generated row 字段、独立复制每个解码 []byte、记录 Scan targets 与解码 buffers。按生成 Scan 次序显式列出值，不依靠反射字段顺序假定 SQL 列顺序。Next/Scan/Err/Close 的可控错误仅存在测试文件。

```go
got, err := queryDesiredProjection(ctx, db, node, 19)
if err != nil || !reflect.DeepEqual(got.Snapshot(), want) { t.Fatal("projection mismatch") }
// 每个错误用例同时要求无材料：
if err != wantErr || !reflect.DeepEqual(got, desiredProjection{}) { t.Fatal("failed query leaked material") }
```

- [x] **Step 3 — 观察 RED。** 设置下方离线环境，运行 `& $taskGo test ./internal/nodecontrol/serving -run '^TestDesiredAdapter' -count=1 -timeout=2m`。先记录缺失符号，再加仅声明/fail-closed stub，正常 Query 用例必须出现行为断言失败；环境失败不算 RED。
- [x] **Step 4 — 实现参数、扫描和正常路径。** adapter 按 spec 顺序验证参数；固定派生 scope，创建 cleanup guard，再调用真实生成方法。Scan 在底层调用前登记 targets；Query 在检查 error 前接管 rows；Close 先置 closed 再转发。Next 最多允许 33 行进入 Scan，第 34 行仅标记不可用并停止；Err 优先报告已记录边界错误，否则转发底层错误。生成成功后真实 converter 创建独立投影，不复制业务校验。
- [x] **Step 5 — 写故障清理测试。** `TestDesiredAdapterTransport`：Query 私有错误/NotFound/Conflict 伪装、rows+error、nil/typed nil rows，第一行与中间行部分 Scan 错误、末尾 Rows.Err，均为有限不可用或取消，不泄漏原始消息。`TestDesiredAdapterCleanup` 保留每个 target、byte alias、参数 digest alias：每条路径退出后零值、底层显式 Close 一次；重复 cleanup 不重复 Close。正常输出与 fixture 保持完整，修改输出快照不影响下一快照。
- [x] **Step 6 — 写 panic/取消/数量测试。** `TestDesiredAdapterPanic` 在 Scan 部分赋值后、Next 已读若干行后、Close 抛出唯一测试 panic，边界 recover 后检查临时材料清零。`TestDesiredAdapterRowLimit` 分别输入 33/34/无限 Next：33 返回 Conflict，后两者不可用，Scan 不超过 33，不能再次查询。`TestDesiredAdapterCancellation` 覆盖查询前、Query/Scan/Err 回调取消、deadline 错误，以及转换后最后一次 context 检查返回取消；最后一种可用记录 Err 调用序列的同步 context 替身，不用 sleep、并发竞态或生产 hook。
- [x] **Step 7 — 观察新增测试 RED 并完成生命周期实现。** 逐个失败条件修复：defer 顺序必须在 Close panic 下仍清 decoded targets；只有 query/rows 完整成功才允许 converter 错误分类。使用明确的阶段所有权标记，确保 converter 产生的投影在取消/后续清理 panic 时 discard；成功交付标记只能在清理完成及最终取消检查后设置。即使错误返还零值，也不能先覆盖仍需清理的自有投影。

  生成方法本身在转换前关闭 rows。对“转换后清理 panic”的防御分支不伪称真实 pgx 常态；若需要观察 staged aliases，可将 adapter 内部收尾逻辑提取为私有、无可替换回调的所有权方法并直接测试其 panic/取消行为，不能添加向生产外泄 staged bytes 的 hook。任何这种提取须 ledger 说明接口，不改变主函数签名。

- [x] **Step 8 — GREEN 与敏感性检查。** 重跑全部 `^TestDesiredAdapter`。若新增清理测试直接通过，临时将 cleanup 的 byte clear 或 target 清零移除一个，观察 Cleanup/Panic 测试失败，再立即恢复。运行全 serving，要求 exit 0；gofmt 新文件，保存本地证据，不提交。

### Task 2: bound epoch 与 guardedRead 接线

**Files:** desired_reader.go、desired_reader_test.go、ledger。

**Interfaces:** 消费 Task 1 adapter、authority.BoundReadEpoch 与现有 guardedRead/discardDesiredProjection；产出：

```go
func readDesiredProjection(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID) (desiredProjection, error)
func readDesiredProjectionWithEpoch(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID, resolve func(store.DBTX) (uint64, error)) (desiredProjection, error)
```

- [x] **Step 1 — 写来源与前置 RED。** `TestDesiredReaderRequiresBinding`：真实固定入口对普通/typed-nil/仿造 BoundReadEpoch 方法/外层装饰的 DBTX 均不可用、零查询。`TestDesiredReaderPreflight`：取消优先；nil context 不可用；无效 node 返回 InvalidArgument、不调用 source；其余 nil/typed-nil source 交 guard 拒绝。`TestDesiredReaderBindingOrder` 用私有 resolver seam 返回 19，断言传入的正是原始 callback DBTX、事件为 resolve→Query，Query 收到同一 scoped context，成功得到完整独立快照。
- [x] **Step 2 — 观察 RED。** `& $taskGo test ./internal/nodecontrol/serving -run '^TestDesiredReader' -count=1 -timeout=2m`；补声明和 fail-closed stub 后确认 BindingOrder 正常案例确实失败。
- [x] **Step 3 — 实现私有 reader。** 固定入口仅传 authority.BoundReadEpoch。WithEpoch 在外层完成取消/nil context/node 前置检查，再调用 guardedRead；callback 内先解析 epoch，nil resolver、0/越界值或非取消错误映射不可用，取消错误映射取消；成功才调用 Task 1 adapter。向 guard 传 discardDesiredProjection，不改变 guard 或 authority 代码，不导出工厂。
- [x] **Step 4 — 写错误顺序与清理组合测试。** `TestDesiredReaderPostcheck` 覆盖成功、NotFound、Conflict、后检失败、两类领域错误+后检失败、请求取消+后检失败、漏 callback、重复 callback。领域错误时 callback 对 source 返回 nil，以完成后检；取消例外。所有失败零投影，重复 callback 仅一次 Query。`TestDesiredReaderResolverErrors` 覆盖 nil、0、越界、私有错误、取消/deadline。`TestDesiredReaderScopedContext` source 成功后取消自己的 callback context，外层仍活动时保留成功。
- [x] **Step 5 — 证明 staged 丢弃。** `TestDesiredReaderDiscardComposition` 直接以真实 guardedRead+Task 1 adapter 构建测试组合，保留自有 payload/envelope、根链容器、签名和投影地址；postcheck reject、outer cancel、source panic 后全部清零，成功则保持。panic 仅在测试边界 recover；测试组合不新增生产观察 hook。另以固定 reader/seam 路径验证相同输出分类、一次 Query 和 decoded cleanup。
- [x] **Step 6 — GREEN/变异检查。** 重跑全部 `^TestDesiredReader` 和全 serving；若清理测试直接通过，临时在测试组合传 no-op discard（不改生产 guard），断言测试能发现 retained staged material 后恢复。记录真实 authority accessor 的现有专用测试与 seam 组合测试各自能证明什么；seam 成功不称作真实 Coordinator/连接证明。

### Task 3: 限定验收与最终独立审查

**Files:** 两份 roadmap、本计划、执行 ledger。

**Interfaces:** 消费 Tasks 1–2 实现及证据；产出局部完成记录，不开放生产。

- [x] **Step 1 — 新鲜限定验证。** 逐条运行下方三条命令，检查 exit 0；store 必须说明 no tests to run。失败先排错，不扩大安全范围。
- [x] **Step 2 — 新上下文独立审查。** 按 Native 执行技能，最强可用审查模型只读检查相关分支变更、spec/plan/ledger 与五项 Review Focus；明确本增量和历史分支实际审查范围。Critical/Important 先复现 RED，再一轮修复与限定全回归；Minor 记录延期，不顺手扩大修改。
- [x] **Step 3 — 边界核验。** gofmt -l 新 Go 文件为空；git diff --check 及未跟踪文件空白检查通过。SQL、全部生成文件、四个锁 hashes 与初始基线一致；不得重新生成来掩盖漂移。确认无公开入口、生产调用者或改写证书路径。保留所有既有改动。
- [x] **Step 4 — 记录并交付。** 写实际 RED/GREEN、审查发现/裁定/排除项、是否有遗留 minor；注明本地未提交未推送。更新 checklist 只勾选完成项；保留真实 DB/PITR、canonical/签名、可信时间/授权、writer/runbook、v7/生产/B01/C12/main 与安全门禁。

## 固定离线验证命令

在指定 worktree 内逐条执行；本计划编写时不执行这些命令。

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$taskGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
& $taskGo test ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=2m
& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m
& $taskGo test ./internal/store -run '^$' -count=1 -timeout=2m
```

缓存写权限仅按需申请限定提权；依赖缺失报告阻塞，不下载/修复缓存。不执行任何 sqlc 或全生成脚本。

## 自查与交接

Spec 参数/生命周期对应 Task 1，来源/后置优先级对应 Task 2，所有权与异常路径分别有真实组件组合测试；五项 Review Focus 均有具名覆盖。Task 3 只验证许可范围，不把模拟 transport 当数据库证据。保持三个主函数签名不变，无新增授权能力。

沿用用户已选择的 Native（当前会话实现＋最终独立审查）。用户已批准实施；Tasks 1–3 已在本地完成，证据和全部裁定见 execution ledger。未提交或推送，生产与安全验收门禁保持开放。
