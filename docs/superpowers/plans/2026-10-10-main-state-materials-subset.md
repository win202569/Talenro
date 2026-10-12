# Main-state Materials Subset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 核对根、轮换、元数据和 desired 原始材料与显式预期的一致性，不产生可信、运行或提交授权。

**Architecture:** 新增隔离纯包 statematerials，唯一解析与字节摘要仍由 stateformat 提供。保持 statecontract 的生产消费者禁令；结果拥有输入副本，并固定报告未完成责任。历史材料允许存在，但不能在本批或同一模型重检中改写绑定。

**Tech Stack:** Go 1.26.5、现有 contracts/UUID/JCS、PowerShell、离线依赖；无新增依赖。

**Spec:** `docs/superpowers/specs/2026-10-10-main-state-materials-subset-design.md`（2026-10-10 已批准）。执行前完整阅读该设计和本计划。

状态：用户于 2026-10-10 批准计划；任务 1–5 已完成。Native 本地实现、限定回归及一次独立审查完成，三项测试覆盖建议已补齐；未提交/未推送。

## Global Constraints

- 工作区仅 `D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator`，复用现有分支；保留所有已有改动。
- 不暂存、提交、推送、合并、重置、删除分支或新建 worktree；下列记录步骤替代技能模板中的 commit。
- 材料表与预期记录各最多 128 项；轮换最多 32 项；原始字节总计最多 16 MiB。
- RootSet payload 最多 65,536 字节；信封最多 262,144 字节；解码 payload 最多 65,536 字节。
- 所有 epoch 均匹配公共 scope；正整数上限 MaxInt64；根 keys 1–5、metadata keys 1–64、revoked IDs 0–4096、desired slots 0–8。
- 不读取当前时间、磁盘、网络、环境、随机数或硬件；测试与基线记录可读取本地文件。
- 不修改生产 wire 格式；不新增宽松解析器、序列化适配器或 L0 消费者例外。
- 不运行全仓库、数据库、Docker、race、硬件、重启、C11/C12/PITR 验收。
- 所有失败返回 nil 模型；错误不回显原始材料；未验证责任不得因材料齐全而消失。
- 所有手工文件编辑使用 apply_patch；格式化可使用固定工具链 gofmt。

## Review Focus

以下五项从接口使用场景补入任务测试，不以设计未枚举为理由忽略：

1. 非活动 union 字段携带残留声明：必须 InvalidInput，不能悄悄忽略（任务 2）。
2. 同一批次调整表顺序及 nil/空集合：合法顺序变化、等价空集合重检通过，规范有序列表乱序拒绝（任务 4）。
3. 上限检查发生于复制前：超大原始切片或预期集合必须 LimitExceeded，不能先解析或复制（任务 2）。
4. 独立历史片段和高根版本：允许缺少片段外轮换，不误称锚链闭合；实际声明的缺失关系仍失败（任务 3）。
5. 相同 payload 的信封替换或外部绑定更换：单次声明可解析不等于可替换模型，Recheck 必须拒绝（任务 4）。

## 文件与接口分工

修改 stateformat 的 `artifact.go`（新私有拥有字段和 Clear）、`root.go`（内嵌原始根提取）、`desired.go`（一次解析产生 slots）、`envelope.go`（接收新增事实）。
新增 `stateformat/subset_material.go`（公开事实 API）、`subset_material_test.go`（兼容与所有权）。不重构无关解析器。

新包 statematerials：`doc.go` 边界；`input.go` 类型；`errors.go` 错误；`shape.go` 预算/结构；`parse.go` 单件核对；`relations.go` 图关系；`model.go` 所有权；`evaluate.go` 入口。
测试文件按任务为 `fixtures_test.go`、`parse_test.go`、`relations_test.go`、`model_test.go`、`boundary_test.go`。
执行证据写入 `docs/roadmap/2026-10-10-main-state-materials-subset-execution.md`；最后更新 `docs/roadmap/current-status.md`，保留历史记录。

## 固定执行环境与验证命令

PowerShell 每个新执行会话先设置下列任务环境；不修改用户永久环境：

```powershell
$env:GOENV='off'; $env:GOWORK='off'; $env:GOTOOLCHAIN='local'
$env:GOPROXY='off'; $env:GOSUMDB='off'; $env:GOAUTH='off'
$env:GOVCS='all:off'; $env:GOFLAGS='-mod=readonly'
$env:GOPATH='C:/Users/Lenovo/go'; $env:GOMODCACHE='C:/Users/Lenovo/go/pkg/mod'
$materialsGo='C:/Users/Lenovo/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe'
```

工作目录为上述 worktree。下文命令每次独立运行，记录退出码，不让后一次覆盖前一次失败。
已在运行的工具会话必须等结束，不以输出迟到为理由重复启动测试。
RED 必须是预期缺失接口或断言失败；环境失败不算 RED。若原行为已通过，记录已有覆盖，不制造失败。

### Task 1: 唯一解析器补充原始根与 slots 事实

**Files:** 修改上述四个 stateformat 产品文件；新增 subset_material.go、subset_material_test.go。

**Interfaces:**

```go
type RootSetBindings struct { Epoch, Sequence, Version uint64 }
func ParseRootSet(payload []byte) (*ParsedRootSet, error)
func (r *ParsedRootSet) Bindings() RootSetBindings
func (r *ParsedRootSet) Payload() []byte
func (r *ParsedRootSet) PayloadDigest() contracts.Digest
func (r *ParsedRootSet) RootMaterial() (RootMaterial, error)
func (r *ParsedRootSet) Clear()
func (a *ParsedArtifact) EmbeddedRootSet() (*ParsedRootSet, error)
func (a *ParsedArtifact) DesiredSlotIDs() ([]string, error)
```

ParsedRootSet 字段私有，无 Kind=InitialRoot、无 envelope 身份。空对象的值访问器返回零值，RootMaterial 返回 ErrInvalidFormat；Clear nil-safe。
EmbeddedRootSet 仅 RootRotation 可用，每次返回独立所有者；DesiredSlotIDs 仅 Desired 可用，零 slots 返回 nil、nil；错 kind/清理后返回 ErrInvalidFormat。

- [x] 在首次产品改动前记录当前 HEAD、工作树状态及全部现有 Go/SQL/go.mod/go.sum 哈希到 `.superpowers/sdd/2026-10-10-main-state-materials-subset/`。使用 apply_patch 保存清单，不改 Git 配置或索引。
- [x] 写 `TestSubsetRootFacts`：从现有 fixture 验证 v1、内嵌 v2、MaxInt64 版本；`got.PayloadDigest()==sha256.Sum256(raw)`；内嵌根摘要不等于 rotation payload 摘要；InitialRoot 信封的 v2 仍报 ErrInvalidFormat。
- [x] 写 `TestSubsetFactOwnership` 和 `TestSubsetSlots`：分别篡改输入、返回 payload、keys、slots，二次读取不变；Clear 不改外部副本；nil/错类型/清理后拒绝；slots 0/8 通过、9 拒绝，乱序与重复拒绝。
- [x] 运行 `& $materialsGo test ./internal/nodecontrol/stateformat -run '^TestSubset' -count=1`，记录预期 RED。
- [x] 实现上述 API。RootSet 复用 decodeRootPayload；rotation 在已有 strictCanonical 成功后用 json.RawMessage 取得原始 new_root_set，再用根解析器核对；不得从 DTO 重编码。既有 rootBindings/rotation wrapper 的外部语义不变。
- [x] desired 内部增加 `decodeDesiredPayload(body []byte) (artifactBindings, []string, error)`，原 parseDesiredPayload 作为兼容 wrapper 并清理不返回的 slots；ParseEnvelope 接收其 slots。ParsedArtifact 持有并清理 slots 和内嵌根所有者；失败路径统一清理。
- [x] 重跑 targeted，再运行 `& $materialsGo test ./internal/nodecontrol/stateformat -count=1`；必须 exit 0，已有 vectors、格式和边界测试不放宽。记录证据，不提交。

### Task 2: 有界批次及逐件材料核对

**Files:** 新增 statematerials 的 doc/input/errors/shape/parse/evaluate/model；测试 fixtures_test.go、parse_test.go。model/evaluate 此步提供最小所有权结果，任务 3/4 完善。

**Interfaces:** 输入类型固定如下，Digest 表示 contracts.Digest；产品代码不定义另一种摘要算法。

```go
type MaterialKind uint8 // 0 Unknown; 1 RootSet; 2 RootRotation; 3 Metadata; 4 Desired
type Scope struct { NodeID string; Epoch uint64 }
type MaterialRef struct { Kind MaterialKind; Payload, Envelope contracts.Digest }
type Material struct { Ref MaterialRef; Raw []byte }
type Window struct { Issued, Until int64 }
type RootExpectation struct { Version, Threshold uint64; KeyIDs []contracts.Digest }
type RotationExpectation struct {
    PreviousVersion, NewVersion uint64
    OldRootPayload, NewRootPayload contracts.Digest
    Window Window
}
type MetadataExpectation struct {
    Version, RootVersion uint64
    RootPayload contracts.Digest
    KeyIDs, RevokedIDs []contracts.Digest
    Window Window
}
type DesiredExpectation struct {
    NodeID string
    Generation, InventoryVersion, ResourceVersion uint64
    ResourcePayload, RootPayload, MetadataPayload, SigningKeyID contracts.Digest
    Window Window
    SlotIDs []string
}
type Expectation struct {
    Ref MaterialRef
    Epoch, Sequence uint64
    Root RootExpectation
    Rotation RotationExpectation
    Metadata MetadataExpectation
    Desired DesiredExpectation
}
type Input struct { Scope Scope; Materials []Material; Expected []Expectation }
func Evaluate(ctx context.Context, in Input) (*MaterialSubsetModel, error)
func Recheck(ctx context.Context, model *MaterialSubsetModel, in Input) error
func (m *MaterialSubsetModel) Clear()
```

Expectation 是封闭 tagged union；只有 Ref.Kind 对应的一个分支可非零，其余分支必须零值（空 slice 视同 nil）。Ref 的 (Kind, Payload) 是唯一键；Envelope 是必须核对的值，不作为允许重复 payload 的借口。
同一表内唯一键重复统一为 ErrDuplicateMaterial；两张表各自唯一但对应项 E 不同则为 ErrReferenceMismatch。
所有错误使用 errors.New 哨兵 `ErrInvalidInput/ErrLimitExceeded/ErrInvalidMaterial/ErrMissingMaterial/ErrDuplicateMaterial/ErrReferenceMismatch/ErrClearedModel`；context 错误原样返回。
私有函数 `validateInput(in Input) error`、`parseBatch(ctx context.Context, in Input) (*parsedBatch, error)`；parsedBatch 拥有解析结果并实现 Clear。

- [x] 创建测试专用 `subsetFixture(t)`，基于现有 stateformat testdata 构造相互匹配的四类批次。预期坐标显式填写；原始 bytes 摘要由测试标准 SHA-256 独立算，不从待测模型回填期望。纯测试 JSON/JCS 修改器为变异测试同步重建信封。
- [x] 写 `TestSubsetParse`：正常批次通过；逐一篡改从材料核对的预期标量/列表或 Raw，要求错误且模型 nil；外部关系变异留到任务 3。签名内容格式合法但不真实仍可通过格式及字段核对，未验证标记由任务 4 固定断言。
- [x] 写 `TestSubsetInputShape`：nil ctx、空批次、非法 node/epoch/类型、root 非零 E、其他类型零 E、union 残留、重复记录和表项拒绝；无对应预期或材料报 MissingMaterial；同 P 不同 E 报 ReferenceMismatch；规范集合乱序/重复报 InvalidInput。
- [x] 写 `TestSubsetBudgets`：128/129 项、32/33 轮换、65,536/65,537 payload、262,144/262,145 envelope、16 MiB/16 MiB+1；允许边界的预算检查通过，超限为 LimitExceeded。未构成合法 JSON 的边界 bytes 只测试 validateInput，不虚称 Evaluate 成功。根 keys 5/6、metadata keys 64/65、revoked 4096/4097、slots 8/9 在复制前检查；使用同时无效/超限的输入验证超限优先于解析。
- [x] 运行 `& $materialsGo test ./internal/nodecontrol/statematerials -run '^TestSubset(Parse|InputShape|Budgets)$' -count=1`，记录 RED。
- [x] 实现 validateInput：ctx 入口检查后先预算，再结构、唯一性和表/预期配对；采用 `limit-total` 剩余量比较避免加法溢出。检查所有 union 分支集合上限后再比较零值；node 文本长度 36、slot 长度 1–64 在复制前检查。缺少一侧属于 MissingMaterial。canonical UUID 校验不接受可解析但非 canonical 文本。复制后调用 parseBatch，用任务 1 API 和现有 Bindings/MetadataMaterial 核对所有显式预期；只丢弃解析临时对象，不丢弃自有输入。
- [x] 运行同一命令至 GREEN；本阶段不声称引用关系已完成。记录执行状态，不提交。

### Task 3: 根图、元数据与 desired 外部关系

**Files:** 新增 relations.go、relations_test.go；修改 evaluate.go 调用顺序。

**Interfaces:** `validateRelations(ctx context.Context, in Input, parsed *parsedBatch) error`；位于 parseBatch 成功之后、模型公开返回之前。索引只用本批 (Kind, Payload)，不使用外部查找。

- [x] 写 `TestSubsetRelations`：缺 old/new root、metadata/root version 不符、desired metadata/key/root 不符均拒绝且模型 nil；字节不同的新根拒绝；相同根版本不同 P、轮换分叉/重复、倒退 sequence 拒绝；外部关系的缺失为 MissingMaterial，冲突为 ReferenceMismatch。
- [x] 写 `TestSubsetHistory`：根版本大于 32 的合法单根及断开历史片段通过；声明的轮换两端必须都在本批；旧 metadata/desired 在较新 revoked 材料并存时仍可保持原关联；不读取当前时间；不要求没有声明的片段外轮换。使用显式旧日期材料，保留格式内部窗口约束。
- [x] 写 `TestSubsetReferenceIdentity`：同 P 的重复声明不得绕过唯一性，篡改原 root/metadata/key/resource/inventory 声明不能被不同表顺序隐藏；旧/new root 与 rotation P 不可互换；metadata 与 desired 的外部根引用只验证关系，不声称被 wire 签名覆盖。
- [x] 运行 `& $materialsGo test ./internal/nodecontrol/statematerials -run '^TestSubset(Relations|History|ReferenceIdentity)$' -count=1`，确认断言 RED。
- [x] 实现关系检查：根版本唯一；每条轮换版本 +1（先检查上限）、新根原始 bytes 相等、old sequence < new sequence = rotation sequence；相同端点/分叉拒绝。所有材料 scope epoch 一致。metadata 所指 root 版本相等；desired 所指 metadata 存在 signing key，且双方 root 引用一致。不要因 key 当前 revoked 或已过时间窗添加业务授权判断。关系遍历检查 ctx。
- [x] 重跑 targeted 和 `& $materialsGo test ./internal/nodecontrol/statematerials -count=1` 至 GREEN，记录证据。

### Task 4: 封闭结果、所有权、取消与完整重检

**Files:** 完善 model.go/evaluate.go；新增 model_test.go。

**Interfaces:**

```go
type ResultStatus uint8 // 0 InvalidResult; 1 SubsetConsistent
type CoverageState uint8 // 0 NotProvided; 1 Checked
type Coverage struct { Kind MaterialKind; State CoverageState; Count int }
type Obligation uint8 // 0 UnknownObligation; 后六值见下
type Summary struct {
    Status ResultStatus
    Scope Scope
    MaterialCount int
    Coverage [4]Coverage
    Obligations [6]Obligation
}
func (m *MaterialSubsetModel) Summary() Summary
```

六义务按固定顺序：OtherMaterialsAndHistory、SignaturesAndAnchor、DomainSemantics、CompositionRecheck、PhysicalCommit、RuntimeAuthorization。有效 Summary 的 coverage 按四种 kind 排列；空/nil/已清理模型 Summary 返回零值，不伪造成功或责任清单。模型字段全部私有，无转换为可信候选的方法。
Recheck 先检查 ctx、模型存活，再 Evaluate(in) 并比较完整规范化自有输入；差异 ErrReferenceMismatch，缺失/无效模型 ErrClearedModel。输入表按唯一身份排序；仅等价空集合归一为 nil；不修复调用者的非规范嵌套列表。

- [x] 写 `TestSubsetSummary`：单根批次允许且其余三种 NotProvided；四类齐全仍恰六项义务；模型零值/清理后不能显示 SubsetConsistent。
- [x] 写 `TestSubsetOwnership`：通过测试反射枚举每个切片路径，每个用新 fixture、独立 expected，逐一修改 Raw、Expected 内所有列表和外层表；已生成模型不变；清理自有存储不修改 caller；所有者不可浅复制的限制写入 doc.go。
- [x] 写 `TestSubsetRecheck`：同输入、表换序、等价 nil/空集合成功；换 scope、合法另一信封（相同 P 不同 E）、同 raw 换合法外部关联、替换任一原始字节均拒绝；确认实际重解析，不仅匹配摘要；nil/已清理模型拒绝。
- [x] 写 `TestSubsetCancellation`：nil ctx、预取消、预超时；用只在测试中的计数 Context 在逐件解析前后、关系遍历、最终返回处触发 Canceled；模型必须 nil。Recheck 最终比较后也检查 ctx；不引入 sleep、后台任务或产品测试钩子。通过包内可观察私有存储检查 Clear 清零；不声称证明运行时所有复制已擦除。
- [x] 运行 `& $materialsGo test ./internal/nodecontrol/statematerials -run '^TestSubset(Summary|Ownership|Recheck|Cancellation)$' -count=1`，记录真实 RED 或已有通过项。
- [x] 实现深复制/清理/摘要/重检，Evaluate 仅成功时转交所有权；每个错误分支清理自有输入和临时解析器对象，最终 ctx 检查失败时清理新模型。Recheck 的 fresh 模型总是 defer Clear。
- [x] 重跑 targeted 与新包全测至 GREEN，记录证据。

### Task 5: 边界回归、一次 Native 独立审查及交付

**Files:** 新增 statematerials/boundary_test.go；更新本计划勾选项、execution.md、current-status.md。必要修复限定于任务 1–4 文件。

**Interfaces:** `materialsImportsAllowed(src []byte) bool` 测试辅助；生产白名单逐成员限定：context 的 Context/Canceled/DeadlineExceeded、errors.New、bytes.Compare/Equal/Clone、reflect.DeepEqual、sort.Slice、uuid.Parse/Nil/RFC4122、contracts.Digest，以及 stateformat 本计划实际使用的 API/类型/Kind/ErrInvalidFormat。拒绝 dot/blank import；不放行前缀子包或额外别名绕过。

- [x] 写 `TestSubsetBoundary`：导入 net/os/time/crypto/statecontract/stateverify/statepolicy 均拒绝；允许的库中 uuid.New、context.WithCancel 等未授权成员也拒绝；解析错误拒绝。扫描仓库非测试 Go 文件，禁止导入 statematerials 的生产消费者；保持现有 L0 边界原样。
- [x] 运行 `& $materialsGo test ./internal/nodecontrol/statematerials -run '^TestSubsetBoundary$' -count=1`，新 guard 缺失先 RED，再实现 guard 并 GREEN；若扫描发现实际越界，修复产品而非放宽白名单。
- [x] 用固定 gofmt 格式化本次 Go 文件；确认格式列表为空；逐个运行以下三条命令，均须 exit 0：

```powershell
& $materialsGo test ./internal/nodecontrol/statematerials ./internal/nodecontrol/stateformat ./internal/nodecontrol/statecontract ./internal/nodecontrol/staterecovery ./internal/nodecontrol/statetime ./internal/nodecontrol/stateverify ./internal/nodecontrol/statepolicy ./internal/nodecontrol/serving ./internal/nodecontrol/contracts ./internal/readiness -count=1
& $materialsGo test ./internal/nodecontrol/authority -run '^(TestBoundRead|TestCoordinator)' -count=1
& $materialsGo test ./internal/store -run '^$' -count=1
```

- [x] 按 Native 方法调用 requesting-code-review 技能，派发一次新上下文、当时可用最强模型的只读 reviewer；提供批准设计、计划、基线、实际 diff/新增文件清单和测试结果。不要用 HEAD diff 漏掉 untracked 新文件，也不要把所有旧脏文件冒充本次改动。审查范围只为本增量，不做外部写操作。
- [x] 对 Critical/Important 发现由作者复现、按 TDD 修复，并重跑上述限定回归；Minor 记录解决或明确待办。若审查要求改变批准设计，暂停并请求方向，不自行扩大范围。不自动发起第二次独立审查。
- [x] 比较基线：仅列明批准文件改变，其余已有 Go/SQL/module 文件哈希不变；核对 HEAD 未变，Git diff --check 无空白错误。Git 若受 ownership 限制，只申请只读提升，不修改 safe.directory。
- [x] execution.md 记录每轮真实退出码、范围、审查发现与处理、未解决项；current-status.md 加入“本地实现、未提交/未推送、不是完整材料闭包或生产验收”。结束所有测试会话后交付，不把仍运行当通过。

## 计划自查与批准门槛

设计 §1–2 对应边界及任务 5；§3 对应任务 2；§4 对应任务 1；§5 对应任务 3；§6 对应任务 4；§7–8 对应任务 5。
Review Focus 五项分别已分配明确测试；结构类型、入口名称与错误在各任务之间保持一致。
执行结果见 docs/roadmap/2026-10-10-main-state-materials-subset-execution.md。一次审查 0 Critical / 0 Important / 3 Minor；三项覆盖建议均以补充测试及定向变异验证闭合。补充后新包/stateformat exit0，20 个审查过的产品文件哈希未变；不重复无改动的其他限定回归，不发起第二次审查。
