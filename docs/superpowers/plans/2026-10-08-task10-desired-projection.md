# Task 10 Desired Projection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 实现私有期望状态暂存投影的结构校验、根链连续性、深拷贝与清理，不将其当作已验签的 serving facts。

**Architecture:** 一组私有专用值类型持有 inventory 选择、root/metadata 发布物与 desired 签名材料；构造器先验证完整结构，再复制输入。查询、规范解析、密码学和生产接入不在本计划内。

**Tech Stack:** Go 1.26.5、现有 google/uuid 与 contracts 类型、标准库 bytes/sha256/time；PowerShell 离线验证。

**Spec:** `docs/superpowers/specs/2026-10-08-task10-desired-facts-design.md`（已确认；执行前完整阅读）。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator` 和现有 codex 分支，保留所有先前未提交工作。
- 不导出 `DesiredStateFacts` 或 `Reader.GetActiveDesiredState`；不改现有证书读取流程。
- epoch 与正坐标范围 `1..MaxInt64`；base version 范围 `0..MaxInt64`，ChainBaseRootVersion 必须为正。
- RootRotations 最多 32 段；逐段版本精确加 1，不排序、截断或填补。
- payload 长 `1..65536`；root/metadata envelope 长 `1..262144`；签名固定 64 字节，digest/key ID 固定 32 字节。
- `IssuedAt < EffectiveDeadline <= ValidUntil <= IssuedAt+24h`；不依据本机当前时间授权。
- 不暂存、提交、推送；用本地执行记录代替技能模板的 commit 步骤。
- 不运行数据库、Docker、完整 C11/C12/bootstrap/PITR 或全仓库测试；不改安全设置、下载依赖或变更锁文件。
- 仅允许 serving/contracts/readiness、authority 的 `^(TestBoundRead|TestCoordinator)` 组，以及 store 的 `-run '^$'` 编译检查。

## Review Focus

- 共享底层 slice（特别是链尾与 ActiveRoot）必须被隔离复制：Task 1 Ownership。
- 修改链尾字段同时保持 digest 有效仍必须因精确末端不一致被拒绝：Task 1 Chain。
- MaxInt64 末端与加法越界必须分别允许和拒绝，不能 wrap：Task 1 Chain。
- 摘要正确的任意 bytes 不是规范 envelope；测试不能误称其可信：Task 1 OpaqueBoundary。
- 错误构造与重复清理不得改变输入或已交付快照：Task 1 Ownership/Discard。

## 文件结构

- Create: `internal/nodecontrol/serving/desired_projection.go`：全部私有值类型、校验、复制和清理；不加 DBTX/context/provider 依赖。
- Create: `internal/nodecontrol/serving/desired_projection_test.go`：真实纯函数测试，无数据库替身。
- Modify: `docs/roadmap/current-status.md`、`docs/roadmap/2026-10-08-task10-read-boundary.md`：限定证据与后续门禁。
- Create: `docs/roadmap/2026-10-08-desired-projection-execution.md`：本计划执行记录，含裁定、RED/GREEN、审查和未完成项。
- 不改 SQL、生成文件、schema、既有 facts 或 reader。

## 固定接口和数据字典

均位于 serving 包且类型名私有；结构字段可用以下首字母大写名称，方便显式 fixture。

| 类型 | 字段及 Go 类型 |
| --- | --- |
| `desiredPublishProjection` | `PublishID, AuthorityOperationID uuid.UUID`; `Version, BaseRootVersion, BaseMetadataVersion uint64`; `Authority contracts.AuthorityVersion`; `Payload, PublishedEnvelope []byte`; `PayloadDigest, EnvelopeDigest contracts.Digest` |
| `desiredArtifactProjection` | `NodeID, SigningID, AuthorityOperationID, RootPublishID, MetadataPublishID uuid.UUID`; `Generation, InventoryVersion, ResourceEnvelopeVersion, RootVersion, MetadataVersion uint64`; `Authority contracts.AuthorityVersion`; `ResourceEnvelopeDigest, SigningKeyID, PayloadDigest contracts.Digest`; `Payload []byte`; `Signature [64]byte`; `IssuedAt, EffectiveDeadline, ValidUntil time.Time` |
| `desiredProjectionSnapshot` | `NodeID, ActiveRootPublishID, ActiveMetadataPublishID uuid.UUID`; `ActiveDesiredGeneration, ActiveRootVersion, ActiveMetadataVersion, ChainBaseRootVersion uint64`; `RootRotations []desiredPublishProjection`; `ActiveRoot, ActiveMetadata desiredPublishProjection`; `Desired desiredArtifactProjection` |
| `desiredProjection` | 仅 `snapshot desiredProjectionSnapshot`，无 DBTX 句柄 |

公开字典不意味着 exported API。Root 与 metadata 位置决定校验规则，不引入自由 kind 字符串或可选类型分支。

固定函数：

```go
func newDesiredProjection(nodeID uuid.UUID, epoch uint64, s desiredProjectionSnapshot) (desiredProjection, error)
func (p desiredProjection) Snapshot() desiredProjectionSnapshot
func cloneDesiredProjectionSnapshot(s desiredProjectionSnapshot) desiredProjectionSnapshot
func discardDesiredProjection(p *desiredProjection)
```

复用 `factsUUID`、`factsCoordinate` 与 `contracts.AuthorityVersion.Validate()`，不改它们。
可提取私有发布物校验/复制小 helper，但不得引入通用 repository、签名框架或测试用可替换生产全局变量。

### Task 1: 私有投影、验证与所有权边界

**Files:** desired_projection.go、desired_projection_test.go 和执行记录。

**Interfaces:** 消费上述既有验证 helper；产出数据字典四种私有类型及四个固定函数。

- [x] **Step 1 — 写测试 fixture 和正常/错误用例。** `desiredProjectionFixture(t, rotations int)` 返回 expected node、epoch 19 和独立 snapshot；UUID 固定合法且各角色不同；根版本从 1 开始，链节点从 2 递增，authority sequence 用 10、11…，metadata/desired sequence 分别 100/101。desired generation=7、inventory=8、resource version=9，metadata version=4；时间固定 2026-01-01 UTC，deadline=+1h、validUntil=+24h。非空 payload/envelope 使用固定 bytes，测试自行计算 SHA-256，而非调用产品校验函数生成期望值。active root 与最后一段内容相同但为独立期望对象；空链 active root version=1/base=0。

  新增测试组 `TestDesiredProjectionValid`、`TestDesiredProjectionReject`、`TestDesiredProjectionBounds`、`TestDesiredProjectionChain`、`TestDesiredProjectionOpaqueBoundary`。断言模式：

```go
// Valid: 0、1、32 段，完整 snapshot 比较，所有字段均保留。
got, err := newDesiredProjection(node, 19, input)
if err != nil || !reflect.DeepEqual(got.Snapshot(), want) { t.Fatal("projection changed") }
// Reject: 每次仅改变一个独立条件；无效 expected node/epoch => InvalidArgument，
// 其余内部不一致 => Conflict；每个失败都要求零值：
if err != wantErr || !reflect.DeepEqual(got, desiredProjection{}) { t.Fatal("invalid material escaped") }
```

  Reject 逐字段覆盖所有 UUID 的零值/非 RFC variant、正整数的 0/MaxInt64+1、允许零的 base 字段越界；所有制品 epoch/sequence，node 与 active 指针、desired 引用、metadata base root、root/metadata base continuity 错配。逐项校验摘要/key ID/签名/resource digest 非零与 payload/envelope SHA-256 不符。不得只测试链首而遗漏中间节点。

  Bounds 对 active root、metadata、某个中间 rotation、desired payload 分别覆盖 0/1/65536/65537；各 publish envelope 覆盖 0/1/262144/262145。改变长度后更新 fixture 摘要，避免长度测试实际仅在测 hash mismatch。时间覆盖每个零值、相等 issued/deadline、逆序、deadline=validUntil 正例、24h 正例和 24h+1ns 反例。

  Chain 覆盖 33 段、重复 PublishID、重复版本、乱序、断段、sequence 相等/回退、错 epoch、错误末端；分别变更尾端 ID/版本/operation/authority/base 字段/bytes（同步更新摘要）来验证完整尾端相等。空链 base 不等于 active root 拒绝；base=MaxInt64 空链允许，base=MaxInt64-1 到 MaxInt64 一段允许，base=MaxInt64 加一拒绝。不要给首段与不存在的 base sequence 发明比较规则。

  OpaqueBoundary 明确以任意非 JSON bytes 和非零假签名证明本层只做结构检查，不做 canonical/验签；测试名称不得含 authenticated/authorized。

- [x] **Step 2 — 观察 RED。** 使用下方离线环境运行 `& $taskGo test ./internal/nodecontrol/serving -run '^TestDesiredProjection' -count=1 -timeout=2m`。先观察缺失类型/函数；只补声明和 fail-closed stub 后再运行，必须看到正常投影断言失败，而非环境错误。
- [x] **Step 3 — 实现验证与复制。** `newDesiredProjection` 先检查 expected 参数，全部内部规则成功才深拷贝；失败不修改输入。发布物通用形状检查后，按 root/metadata 位置检查 base 规则；chain 长度先于遍历，先检查当前版本 < MaxInt64 再加一。链尾比较所有字段（bytes 按内容），不只比较摘要；空链也完整验证 ActiveRoot。成功对象仅持有自有复制。
- [x] **Step 4 — 第一轮 GREEN。** 重跑相同选择器，输出 ok/退出码 0；随后 `& $taskGo test ./internal/nodecontrol/serving -count=1 -timeout=2m` 验证既有逻辑未回归。
- [x] **Step 5 — 写所有权与清理测试。** `TestDesiredProjectionOwnership` 修改输入的链容器、每类嵌套 bytes、第一次快照，断言再次 Snapshot 完整等于独立原期望；包含输入 ActiveRoot 和链尾共享 backing slice 的情况。`TestDesiredProjectionDiscard` 保留 p.snapshot 内所有 byte aliases，清理后逐字节为零且 p 为零值，输入和先前 Snapshot 不变；重复清理零值也不 panic，零投影 Snapshot 返回零结构。nil 指针不是函数契约，不新增无用分支。
- [x] **Step 6 — 证明新测试有效。** 清理先保留 no-op stub，确认 Discard 测试失败后实现：清每段 payload/envelope、active root/metadata 的各 bytes、desired payload，再清 root slice 与整个对象（包括固定签名数组）。所有权若因 Step 3 已复制而直接通过，临时去掉 clone 的一个关键复制验证测试失败，再立即恢复，不留下变异实现。
- [x] **Step 7 — 完成本地检查点。** 全 serving 再次通过，记录各次 RED/GREEN 和必要裁定；gofmt 本增量两文件，不提交。第一阶段完成不使结构投影成为来源证明或已验签 facts。

### Task 2: 限定回归、独立审查与交付

**Files:** 两份 roadmap、本计划及执行记录。

**Interfaces:** 消费 Task 1 私有实现及测试证据；产出限定完成结论，不产出公开 reader 或 SQL。

- [x] **Step 1 — 独立审查。** 按执行方式安排新上下文审查，重点看 Review Focus 五项、spec 覆盖、状态命名和未误开放入口。可只读查看关联分支变更，不将未审部分称作通过。不运行安全阻塞的集成测试；重要问题需复现 RED 再修复。
- [x] **Step 2 — 新鲜回归。** 分别执行下面三条测试命令，逐条检查退出码，失败则停止报告完成：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$taskGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
& $taskGo test ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=2m
& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m
& $taskGo test ./internal/store -run '^$' -count=1 -timeout=2m
```

  Task 1 也先设置同一环境。缓存写权限仅申请限定提权；离线依赖缺失则报告，不下载。Expected：退出码 0、各包 ok；store 明确 no tests to run。

- [x] **Step 3 — 检查文件边界。** gofmt -l 新 Go 文件输出为空；git diff --check 无输出，另检查新文件尾随空白。只读 git status 确认无无关修改；四个锁文件的 SHA256 与执行前基线完全相同，生成 SQL/store 无本次变化。先前未提交文件不得清理或重写。
- [x] **Step 4 — 记录结果。** 更新当前进度、Task 10 记录和执行 ledger：实际测试结果、审查范围/发现/排除项、未提交未推送。明确根链可信窗口、SQL/fence/active intent、canonical envelope、验签、v7、真实 DB/PITR、安全与完整 C12 仍未完成。此前证书 fixture minor 不因新 fixture 改善而标记修复。

## 自查与执行交接

Spec 的类型/校验/根链/所有权全部映射 Task 1；限定验证和审查映射 Task 2。
本计划不解决未来根链窗口策略，也不假定起点已获信任；无额外制品/API/依赖。
用户已确认并按 Native 执行。限定实现、回归及独立审查完成；证据与延期 M1 见
`docs/roadmap/2026-10-08-desired-projection-execution.md`。未提交、未推送，后续阶段仍未开放。
