# Task 10 Desired Query Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增固定版本 1 起点的精确 desired SQL 及私有行转换，不开放服务入口。

**Architecture:** 单语句从 inventory 当前指针连接 desired、signing、root/metadata 与精确 fences；有界预期版本槽位保留缺失历史。生成行经私有转换校验后交给既有 desiredProjection 构造器复制。数据库执行适配器和 bound reader 不在本阶段。

**Tech Stack:** Go 1.26.5、sqlc 1.31.1、pgx/v5、PostgreSQL SQL；PowerShell 离线检查。

**Spec:** `docs/superpowers/specs/2026-10-08-task10-desired-query-design.md`（用户已确认，执行前完整阅读）。

## Global Constraints

- 复用 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`；保留已有未提交工作，不创建替代 worktree。
- ChainBaseRootVersion 固定为 1；active root version 必须在 1..33；轮换版本 2..active version，最多 32 段；版本 34 拒绝。
- 不按当前时间过滤历史、不截取最后 32 段、不使用 MAX/latest 或孤立行回退。
- payload 1..65536 字节、publish envelope 1..262144 字节、digest/key ID 32 字节、signature 64 字节；沿用第一阶段非零和摘要验证。
- 同一 epoch、exact operation/epoch/sequence/scope/effect、committed-active fences；历史发布 status=active、reason=root_rotation。
- 初始根不是 0→1 轮换；active version=1 也必须有真实发布物，不伪造制品。
- 不增加公开 Reader/DesiredStateFacts、查询 adapter、生产接线、规范 envelope 构造、验签、迁移或依赖。
- 不暂存、提交、推送；技能默认 commit 步骤替换为本地 ledger 检查点。
- 不运行真实数据库、Docker、全仓库/C11/C12/bootstrap/PITR/race；不改安全设置、不绕过保护、不下载依赖、不改锁文件。
- 仅 serving/contracts/readiness 全包，authority 的 `^(TestBoundRead|TestCoordinator)`，store `-run '^$'` 编译。

## Review Focus

- NULL 的三值逻辑可能把缺 fence 判为成功：Task 1 GuardShape 和 Task 2 Reject，要求所有有效性标记明确 false 而非默认 true。
- 恶意 MaxInt64 active root 可能触发巨大序列：Task 1 GuardShape 限定生成器实参，Task 2 Bounds 拒绝；仅 WHERE 过滤不算资源边界。
- desired 尚未设置但 root/metadata 已发布是正常状态：Task 2 Absent；部分指针组仍为 Conflict。
- 重复固定头被换成另一个合法对象可能混合多个状态：Task 2 HeaderEquality，逐类合法错配并验证零输出。
- 链尾与 ActiveRoot 共享输入缓冲可能掩盖浅复制：Task 2 Ownership 逐次修改立即断言，包括 envelope，不能成对修改后才检查。

## 文件结构和执行记录

- Modify `db/queries/nodecontrol_serving.sql`：只追加 ListExactDesiredProjection，不改证书语句。
- Generated `internal/store/nodecontrol_serving.sql.go`、`internal/store/querier.go`：只通过固定 sqlc 生成；models.go 预期不变。
- Create `internal/nodecontrol/serving/desired_query_test.go`：借 recording DBTX 捕获生成查询文本与参数，做静态契约检查，不模拟 SQL 引擎。
- Create `internal/nodecontrol/serving/desired_row.go`、`desired_row_test.go`：私有生成行到既有投影的转换及纯单元测试。
- Create `docs/roadmap/2026-10-08-desired-query-execution.md`：基线、裁定、RED/GREEN、生成 hashes、审查范围和阻塞。
- Update `docs/roadmap/current-status.md`、`docs/roadmap/2026-10-08-task10-read-boundary.md`、本计划：准确记录局部结果。

## 生成接口与行契约

Task 1 产出 sqlc 标准接口：

```go
func (q *Queries) ListExactDesiredProjection(ctx context.Context, arg ListExactDesiredProjectionParams) ([]ListExactDesiredProjectionRow, error)
```

四个具名参数：`node_id::uuid`、`authority_epoch::bigint`、`node_scope_digest::bytea`、`trust_scope_digest::bytea`。
对应 Go 字段 `NodeID uuid.UUID`、`AuthorityEpoch int64`、`NodeScopeDigest, TrustScopeDigest []byte`。
生成接口可见不等于已授权生产来源；本阶段不添加调用它的产品入口。

输出 aliases 固定如下，避免实现者自行改变相邻任务接口：

- `node_id`；`pointer_groups_valid`、`has_desired`、`current_valid`、`has_rotation`、`rotation_valid` 为非 NULL boolean。
- `active_desired_generation`、`active_root_publish_id`、`active_root_version`、`active_metadata_publish_id`、`active_metadata_version`；`rotation_expected_version`。
- `desired_` 前缀字段对应 desiredArtifactProjection 的全部字段：node_id、signing_id、authority_operation_id、root_publish_id、metadata_publish_id、generation、inventory_version、resource_envelope_version、root_version、metadata_version、authority_epoch、authority_sequence、resource_envelope_digest、signing_key_id、payload_digest、payload、signature、issued_at、effective_deadline、valid_until。
- `root_`、`metadata_`、`rotation_` 三组前缀字段对应 desiredPublishProjection：publish_id、authority_operation_id、version、base_root_version、base_metadata_version、authority_epoch、authority_sequence、payload、published_envelope、payload_digest、envelope_digest。

UUID、整数及 bytes 的缺失值用显式零哨兵统一生成类型（uuid.UUID/int64/[]byte），但必须先从原始 NULL 检查指针完整性/关系标记，不能用 COALESCE 掩盖部分指针。时间保留 nullable pgtype.Timestamptz，不转字符串或丢弃 InfinityModifier。sqlc 的具体 Go 大小写以生成结果为准，SQL alias 不变；不手改生成器类型。

`pointer_groups_valid`：每对 root/metadata ID/version 同时 NULL 或同时为合法非零 ID、正版本，desired generation 为 NULL 或正值。
`has_desired` 仅表示原 active_desired_generation 非 NULL，不表示已找到 desired。
`current_valid` 合并 spec 第 2 节的全部精确关系及有限终结时间/材料长度约束；NULL 结果转 false。
`rotation_valid` 对历史套用相同 publish 完整性及 fence 检查，另要求 root_rotation。
缺失或错误 current 只返回一条诊断行，所有制品 bytes 置空；空链 has_rotation=false 且 rotation 组全零。

### Task 1: 组合 SQL 与稳定生成接口

**Files:** SQL、两份生成文件、desired_query_test.go、执行 ledger。

**Interfaces:** 消费现有 schema、authority scope 规则及上述行契约；产出 ListExactDesiredProjection 生成接口，不产出 adapter。

- [x] **Step 1 — 建立基线。** 记录分支/HEAD/状态、四个锁文件及 internal/store 全部文件 hashes；读取既有 recording DBTX 测试、schema 和原 Task 10。首次不运行全生成脚本。保存 ledger，不覆盖既有变更。Git 所有权限制通过只读限定提权解决，不改 safe.directory。
- [x] **Step 2 — 写生成查询契约 RED。** 新增 `TestDesiredQueryArgumentMapping`，使用不同 UUID、epoch=19、两种不同固定 32 字节 scope 值；recording DBTX Query 记录参数并返回零行，检查生成方法完整转交四参数。新增 `TestDesiredQueryGuardShape`，捕获实际生成 SQL，检查 inventory 起点、精确指针关联、三类 effect/scope、committed/active、完整 signing 绑定、有限时间、预期槽位 LEFT JOIN、排序与 LIMIT 33；禁止 MAX、当前时间过滤、OFFSET。检查生成器上界表达式先将越界值变为安全空区间，不把 WHERE 当保护。仅做关键契约检查，不复制一份整段 SQL 当测试 oracle。

```go
// 使用项目现有 recording DBTX 的签名实现，不连接数据库。
rows, err := store.New(db).ListExactDesiredProjection(ctx, params)
if err != nil || len(rows) != 0 { t.Fatal("unexpected generated query result") }
if !reflect.DeepEqual(capturedArgs, wantArgs) { t.Fatal("argument binding changed") }
// wantArgs 按已生成语句中的占位符顺序列出四个独立字面 fixture 值。
```

- [x] **Step 3 — 观察 RED。** 运行下方环境中的 `& $taskGo test ./internal/nodecontrol/serving -run '^TestDesiredQuery' -count=1 -timeout=2m`，记录缺失生成接口。Task 1 的 RED 是生成契约缺失证据，不冒充 SQL 执行失败证据。
- [x] **Step 4 — 编写 SQL。** inventory exact node CTE 保留原始指针存在性；LEFT JOIN exact desired/signing/current publish/fences，计算非空 boolean 标记。所有 SQL 字节输出先限制长度，无效 current 分支仅输出诊断标记。历史采用 `generate_series(2::bigint, CASE WHEN current_valid AND active_root_version BETWEEN 1 AND 33 THEN active_root_version ELSE 1::bigint END)`，LEFT JOIN 保留空链和缺失槽位；每个历史 fence 固定 root_publish/global scope。按 rotation_expected_version 升序，LIMIT 33；版本 34/MaxInt64 返回诊断而不是被隐藏成零行。逐一对照 spec 全部 equality，避免只靠 FK。
- [x] **Step 5 — 离线生成并 GREEN。** 运行下方固定 sqlc 命令，要求 exit 0、生成头 v1.31.1；运行 Step 3 与 store 编译。若 sqlc 不支持表达式，保持契约改写 SQL，不升级依赖。依赖缺失/安全阻断则停止该步骤并报告，不修缓存或重试被检测命令。
- [x] **Step 6 — 验证静态检查敏感性。** 临时从捕获文本检查输入中去掉 root fence 或生成器 CASE，要求 GuardShape 拒绝，随后恢复测试；不要保留变异 SQL。该检查仅证明静态测试能发现指定缺项，不证明 PostgreSQL 执行语义。
- [x] **Step 7 — 重生成稳定性检查。** 保存首次生成后 internal/store 全文件名与 SHA256，第二次执行同一生成命令后集合及内容必须相同。相对初始基线仅 serving 生成文件及 Querier 可变化；证书语句/方法及 models 保持原样。记录本地检查点，不暂存提交。

### Task 2: 私有行转换与错误/所有权边界

**Files:** desired_row.go、desired_row_test.go、执行 ledger。

**Interfaces:** 消费 Task 1 生成行与第一阶段 `newDesiredProjection`；产出：

```go
func desiredProjectionFromRows(nodeID uuid.UUID, expectedEpoch uint64, rows []store.ListExactDesiredProjectionRow) (desiredProjection, error)
```

- [x] **Step 1 — 写有效与拒绝测试。** `desiredRowsFixture(t, rotations int)` 以第一阶段独立期望 snapshot 为基础显式映射生成行，填充诊断/有效性标记；每个角色用不同合法 UUID，epoch=19，generation=7，metadata version=4，固定有限 UTC 时间。新增 `TestDesiredRowValid`（0/1/32 段逐字段等于独立期望），`TestDesiredRowReject`（各 false 标记、负坐标、错 node/epoch、32 字节摘要的 31/33 长度、签名 63/65、NULL/无限时间、缺槽位/乱序/重复/错末端），`TestDesiredRowBounds`（active root=0/34/MaxInt64、33 行、版本 33 正例）。

```go
got, err := desiredProjectionFromRows(node, 19, input)
if err != nil || !reflect.DeepEqual(got.Snapshot(), want) { t.Fatal("projection mismatch") }
// 每个拒绝子用例均要求精确有限错误与零输出：
if err != wantErr || !reflect.DeepEqual(got, desiredProjection{}) { t.Fatal("invalid row escaped") }
```

- [x] **Step 2 — 写状态区分测试。** `TestDesiredRowAbsent`：nil/零行=>NotFound；一行合法指针组且 has_desired=false=>NotFound（root/metadata 先发布也如此）；同一行 pointer_groups_valid=false=>Conflict；has_desired=true/current_valid=false=>Conflict；无效 expected node/epoch 优先 InvalidArgument。错误数量先拒绝，不能从多行中选一个 NotFound。空链存在非零 rotation 数据、错误 has_rotation/expected version 均 Conflict。
- [x] **Step 3 — 写固定头一致性测试。** `TestDesiredRowHeaderEquality` 在至少两行中只改第二行每一组头字段，分别保持字段本身合法：UUID、版本、authority、摘要与重算摘要后的 bytes、时间、签名、所有标记；均须零输出 Conflict。rotation 组不属于固定头，不误比较相邻轮换。摘要正确但非 JSON payload 仍是结构测试，不称为已验签。
- [x] **Step 4 — 观察 RED。** `& $taskGo test ./internal/nodecontrol/serving -run '^TestDesiredRow' -count=1 -timeout=2m`；先缺函数，再加 fail-closed stub 观察 Valid 行为断言失败，区别环境错误。
- [x] **Step 5 — 实现最小转换。** 参数、数量、指针组与 absence 判定先于复制；检查固定头相等、active version 1..33、预期行数、槽位标记，逐字段验证 signed 值/长度再转换。所有时间须 Valid 且 Finite。头和历史分别映射到 snapshot，ChainBaseRootVersion=1，调用既有构造器完成内容/链验证及拥有式复制。不在转换器内排序、执行 SQL、接受 caller high-water 或发明授权检查。
- [x] **Step 6 — GREEN 与 Ownership。** 重跑选择器后新增 `TestDesiredRowOwnership`：修改每个输入 bytes 后立即检查已构造投影不变，恢复 fixture 再测下一个；覆盖链尾与 root 共享 payload/envelope。修改一个返回快照后下一快照不变。失败前后输入完整相等。若直接通过，临时去掉一个 clone 证明用例失败后立即恢复原实现，记录变异证据，不将前次 M1 自动标为修复。
- [x] **Step 7 — 本地检查点。** 全 serving 通过，gofmt 新 Go 文件；记录 RED/GREEN、类型推断裁定和本阶段未提供查询生命周期清理。不得把模拟生成行当成真实来源证明。

### Task 3: 限定验证与独立审查交付

**Files:** roadmap 两文件、本计划、执行 ledger。

**Interfaces:** 消费 Tasks 1–2 的代码/静态及单元证据；产出局部完成结论。

- [x] **Step 1 — 新鲜限定回归。** 逐条运行下方三条测试，检查每条 exit 0；失败按排错技能处理，不扩大范围。store 必须标明 no tests to run。
- [x] **Step 2 — 独立审查。** 沿用 Native：实现由当前会话完成；完成后按执行技能要求由新上下文、最强可用模型审查分支相关变更，明确整体分支与本增量各自的实际覆盖范围。重点看 Review Focus、SQL 三值逻辑、槽位资源边界、未开放生产入口。重要问题需 RED/修复/GREEN；不得以审查代替被阻塞的 DB 测试。
- [x] **Step 3 — 核对工作区。** 再核对四个锁 hashes、生成稳定性和文件集合；gofmt -l 新文件应空，git diff --check 应无错误，另查未跟踪文件尾随空白。不得清理之前未提交内容；记录未审/延期事项。
- [x] **Step 4 — 记录交付。** roadmap/ledger 写实际测试、审查发现和限定结论；保留真实 SQL/PITR、安全、canonical/验签、adapter/reader/v7 门禁。计划只勾选已完成项；不暂存、不提交、不推送。

## 固定命令与预期

所有命令在上述 worktree 执行；先设置离线环境（不是现在执行）：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$taskGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
& $taskGo tool '-modfile=tools/devtools/go.mod' sqlc generate
```

生成仅此 sqlc 入口，不运行整个 generate.ps1 或 devtools 全验证。预期 exit 0；缺缓存不下载。缓存写权限可申请此范围提权，不能扩大到数据库或安全设置。

```powershell
& $taskGo test ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1 -timeout=2m
& $taskGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1 -timeout=2m
& $taskGo test ./internal/store -run '^$' -count=1 -timeout=2m
```

## 自查与交接

Spec 1–3 对应 Task 1 和 Task 2 边界测试；Spec 4 对应 Task 2；Spec 5–6 对应 Task 3 与限制。五项 Review Focus 均有具名测试，但 SQL 静态检查/模拟行不补足真实数据库执行证据。没有新依赖或超范围入口。
用户已确认计划，按 Native 执行；Tasks 1–3 完成。独立审查的一项 Important 测试缺口已 RED→GREEN 修复，最终限定回归通过。完整证据及排除项见 2026-10-08-desired-query-execution.md；未提交或推送。
