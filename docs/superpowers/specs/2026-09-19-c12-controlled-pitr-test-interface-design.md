# C1.2 受控 PITR 测试接口与旧夹具迁移附录

- 日期：2026-09-19
- 状态：用户于 2026-09-19 以“是”批准本文；实施计划及本轮执行方式仍待审阅确认，尚未开始实现。
- 基线：`codex/c12-b01-task9-coordinator`，`226e2f11`。
- 目的：修复旧 PITR 测试与 claim-v1 Coordinator 的不兼容，让测试能够观察真实事务提交，并通过执行器拥有的恢复实例运行真实就绪检查。
- 本文不代表 PostgreSQL/PITR 测试已经通过，也不关闭 Task 9 或 Batch 01 验收。

## 1. 适用范围与规范优先级

本文获批后，仅扩展 [B01 计划](../plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md) Task 8 的 integration-only PITR 接口，以及 Task 9 的 PITR 测试适配范围。Task 8 原先冻结的五个控制器方法不再是完整公开方法集合；新增能力必须符合本文边界。原有五个方法的基础设施测试行为继续兼容。

以下约束不变：

- [Abort/serving amendment](2026-08-24-nodecontrol-authority-abort-serving-design.md) 的 claim、事务、可见性和失败关闭要求。
- [canonical/dispatcher addendum](2026-08-28-nodecontrol-authority-canonical-dispatcher-addendum-design.md) 的封闭 dispatcher、真实/解析证明来源、原子持久化和 admission 消费规则。
- [C12 执行架构](2026-09-04-c12-secure-execution-architecture-design.md) 的进程隔离、构建输入、绝对期限、认证 ownership WAL 和唯一顶层清理者。
- 原有生产 wire schema、canonical transcript、签名角色、数据库迁移及 `c12-spec-set.v1.json` 不因这个测试附录而改变。

不实现 Task 10 的 certificate/desired bound serving，不实现 Task 15/18 的 exact-epoch、runtime/lease/staging readiness，不实现 B03 生产 WAL decoder/provider，也不修复完全丢失 fence 时缺少权威 `ProtocolActivationID` 的问题。

## 2. 已确认的缺口

1. `authority_pitr_integration_test.go` 自行创建 Docker 资源，只安装 `00006`。新的 `GetStoredFence` 查询要求 `00007` 的 claim-v1 字段，测试在进入恢复断言前就可能失败。
2. 旧测试把内存 `coordinatorEffectResolver` 配给真实 pgx 事务；它没有 registered transactional resolver/activator，不能用于新的 sealed dispatcher。
3. `issueC12AuthorityPITRCommit` 只提交自己的 probe，不能证明 Coordinator 的实际事务已经提交；控制器只保留一个 latest terminal。
4. `CrashPrimary` 将崩溃终端与恢复目标设为同一个值，无法表达“已观察 A，之后提交 B/provider 前进，但恢复到 A”。
5. 候选实例的连接仍是私有的，现有候选 DTO 不能单凭 caller 可写字段成为数据访问授权。

普通 Go suite 和 integration-tag 编译通过不能消除这些缺口。

## 3. 选择的方案

采用“控制器签发的作用域数据库能力 + authority 同包 integration bridge”。控制器保留连接、端点、事务及资源身份；bridge 只连接既有 Coordinator 的私有事务入口，不改变生产 API。

纯数据快照方案接口较小，但无法运行真实 Coordinator 或证明正向就绪检查。直接给出 `*pgxpool.Pool` / `pgx.Tx` 便于复用，但会暴露连接配置、原始连接、batch/COPY/prepare 等额外能力。因此本文选择窄接口，并接受维护固定语句登记表的成本。

这是批准的测试代码之间的能力边界，不宣称能在同一 Go 进程中沙箱化任意恶意测试代码。不可用接口限制替代执行器的进程、网络和资源归属约束。

## 4. 接口与所有权

### 4.1 新增的逻辑操作

下表冻结语义；实施计划将把签名落实为这些命名的 integration-only Go API。类型均有私有状态，不能由 caller 填字段获得授权。

| 操作 | 输入与结果 | 所有权要求 |
| --- | --- | --- |
| `WithPrimaryAuthorityAccess` | context、接收窄数据库能力的同步 callback；返回 error | 只连接本 controller 的 primary；crash 开始后拒绝 |
| `WithCandidateAuthorityAccess` | context、候选句柄、同步 callback；返回 error | 只连接本 controller 已晋升的 exact candidate |
| `NewCommitObservation` | 返回不透明 observation 句柄 | 由当前 controller 签发，不接受 caller XID、LSN 或 marker |
| `BindCommitObservation` | observation、controller 签发的 exact transaction；返回 error | 在真实事务正常写入阶段绑定 marker，只能绑定一次 |
| `ObserveCommit` | observation；返回不透明 observed-commit 句柄及只读提交事实 | 事务/锁释放后，由独立受控连接观察 |
| `SelectRecoveryCut` | 本 controller 的 backup 和 observed commit；返回不透明 cut selection | 固定较早恢复目标，不随后续提交改变 |
| `CrashPrimaryAtCut` | cut selection、已观察的较晚 crash-boundary commit；返回现有 restore-cut 值 | 先验证完整关联与归档，再停止本 controller 的 exact primary |

现有 `OpenC12AuthorityPITR`、`CreateBaseBackup`、`CrashPrimary`、`RestoreAtCut`、`PromoteCandidate`、`InspectTimeline` 继续可用于已有基础设施测试。旧 `CrashPrimary(backup, terminal)` 等价于 target 与 crash boundary 相同的兼容路径，不能绕过新增路径的归属验证。

### 4.2 句柄与身份

- Observation、observed commit、cut selection 和新签发的 candidate 都绑定 controller 实例、run nonce、数据库身份及内部登记项；零值、外部构造、跨 controller、旧 phase/generation 的句柄拒绝。
- 现有 candidate DTO 可保留 index/target/promoted 等诊断事实，但只有其私有绑定与 retained state 全部匹配才有授权。Name、DataName 或调用方修改后的 TargetLSN 都不是权限来源。
- 复制一个合法 Go 句柄只产生同一状态的别名，不产生第二个绑定机会或独立授权。允许同一已完成 observation 的幂等读取；不允许复制后重新绑定。
- 所有返回事实为 defensive copies；不返回 DSN、密码、Docker 参数、宿主路径、端口选择、底层 connection/config 或清理目标。
- 最大 64 个 observation，单个 run 只选定一个 recovery cut；candidate 上限仍为现有 8 个。记录保留到 run 结束，不驱逐 cut 所引用的提交。达到上限在副作用前返回确定错误。
- 对外错误采用有限、无敏感值的类别，区分 invalid handle、wrong phase、not observed、capacity exceeded、indeterminate、canceled 和 dependency failure；不拼接连接串、密码、nonce、任意 SQL 或底层 Docker 输出。

## 5. 真实事务提交观察

### 5.1 在正确的事务内绑定

Observation 经 registered test activator 在真实 activation 事务的正常写入阶段绑定；需要观察其他真实领域事务时，同样在该事务的固定测试写入点绑定。绑定验证 controller、primary、transaction generation，并在同一事务内写入控制器生成的唯一 marker。Marker 使用固定内部表/参数化语句，不能由 caller 指定内容或任意 SQL。

测试 bridge 使用控制器签发的原始窄事务能力，不通过任意 caller wrapper、反射或 unwrap 接口伪造归属。现有 crash-test registered handler 可以复用，但其 repository/transaction 故障包装器不能自动成为这个控制器的授权事务。

绑定必须发生在 admission consume 之前。`Commit` 路径只允许内部一次性状态转移并立即调用一次真实 driver Commit；不能在 consume 与该调用之间插入 marker 写入、XID 查询、逻辑解码、WAL 文件写入、独立时间检查或测试 hook。结果丢失故障只能在真实 Commit 调用之后注入。

### 5.2 观察与判定

`ObserveCommit` 不在原事务或任何领域行锁内执行。它从受控 primary 连接及固定 logical slot 读取，关联唯一 marker 所在的完整 decoded transaction block，并要求恰好一个匹配终端。

- 只接受实际 COMMIT 的已验证终端位置，不接受 caller LSN、当前 flush LSN、BEGIN/普通 row-change LSN 作为提交证明。
- 不假设 `txid_current()` 的 full-XID 数值永远等于 logical decoding 的 32-bit XID；用本 run 的唯一 marker 与对应 transaction block 关联，拒绝歧义或跨 transaction 拼接。
- Caller-facing 新接口首先支持 Coordinator 的普通 COMMIT；既有私有 COMMIT PREPARED 基础设施测试保留，但不得将其类型混入普通提交 observation。
- 一个 drain 包含多个已绑定 observation 时，必须保留每个匹配结果，不能只保存被请求的/latest 一个。先完成验证和内存结果安装，再返回。
- 真实 Commit 返回错误不等于 rollback。只要同一 controller 和句柄仍有效，独立观察可以确认已提交 marker；不得通过重做领域提交来“探测”。
- Marker 尚未出现只能表示尚未证明提交；除非有可信 rollback 结果，不得推断已回滚，更不能据此生成可用 cut。
- 每次逻辑读取总量上限为 65,536 行、64 MiB 解码数据；单行上限 1 MiB。SQL 的行数提示不是硬上限，客户端也必须在读取过程中计数、限制和取消。超限拒绝，不截断后继续证明。
- 每次观察及数据库访问能力的寿命最长 30 秒，并受调用 context 与现有 group 剩余绝对期限约束，取三者最早值；这是更小的局部上限，不增加 runner 总预算。到期撤销能力并取消数据库 I/O；同步 callback 必须响应取消。不能宣称 Go 能强制终止忽略 context 的任意 callback，此类挂起仍由执行器现有进程截止机制终止。

### 5.3 重试承诺的边界

支持同一进程、同一 controller、同一句柄的 Commit-response-loss 观察，以及已安装结果后的 Observe-response-loss 重试；已完成结果必须逐字段相同。

不承诺跨进程重启后恢复 observation，不把 ownership WAL 当成新的可恢复 commit archive。现有 destructive slot read 若发生部分读取/取消、无法确认是否已消费完整数据，应将观察通道标为 indeterminate，拒绝新 cut/crash；保留资源归属，交由执行器收尾。不能静默重新打开 controller、接受重新构造的句柄或伪称本次可幂等恢复。

跨进程 peek/persist/ack 协议不在本附录内。Ownership WAL 的既有记录仍用于资源归属和诊断，不因此获得生产提交证明语义。

## 6. 恢复切点与物理崩溃分离

选定目标 A 时，backup、A 和数据库身份必须同属当前 controller，A 的提交位置不得早于 backup 完成位置。只接受 controller 已观察的实际位置，禁止 LSN 加减、caller 自选字符串或以 wall clock 猜测边界。

之后允许提交并观察 B，要求 `B >= A`；真正的 pre-revocation 场景要求 `B > A`。B 不覆盖 A。`CrashPrimaryAtCut` 要求没有仍活动的数据库 callback/事务，再固定 crash boundary，切换并等待所需 WAL 归档，停止 exact primary。返回的 restore cut 保留 A，不把 B 复制为恢复目标。

恢复采用明确的 inclusive 目标：A 提交可见，后续 B 的领域状态不可见。该性质必须通过恢复后数据验证，而不是仅比较 LSN 文本；含/不含边界遵循 [PostgreSQL 18 recovery target 语义](https://www.postgresql.org/docs/18/runtime-config-wal.html)。

Restore/Promote/Inspect/Access 都重新检查 retained candidate、cut、phase 和完整身份。候选晋升后 timeline 必须前进，system identity 仍与 backup 匹配。重复、错序或归属错误操作不能创建第二份未登记资源，也不能扩大清理范围。

## 7. 窄数据库能力与 authority bridge

### 7.1 不暴露完整 pgx 能力

公开数据库能力只提供 `store.DBTX` 所需的查询形状，以及固定 READ COMMITTED 的窄事务创建操作。窄事务仅提供 DBTX、一次 Commit 和 Rollback；不得返回或实现可供 caller 使用的完整 `pgx.Tx` / `*pgxpool.Pool`。

authority 同包、首行 `//go:build integration` 的 test bridge 复用真实 `PostgresRepository`，仅实现现有私有 `beginAuthorityTransaction` seam，将窄事务交给未修改的 Coordinator。它不复制 readiness 算法，不创建替代生产 repository，不增加生产 raw-provider/DB escape。

禁止公开 Conn、Config、Begin/savepoint、Prepare、SendBatch、CopyFrom、LargeObjects、通知订阅等旁路。查询结果在读 API 返回前按固定上限物化并关闭底层 driver Rows；不向 caller 移交活动数据库游标。返回的 `pgx.Rows` 只是有作用域的结果视图，`Conn()` 返回 nil；QueryRow 同样在返回前完成受限读取，不能把原始 driver Row 藏在延迟 Scan 内。Row.Scan、Rows.Next/Values/RawValues 等均检查存活期，返回值不别名内部缓冲区。关闭后只能 Close/Err 等安全收尾或返回确定失效结果。

### 7.2 固定语句策略

控制器持有编译期固定语句登记表，不接受 caller 注入的 SQL allowlist、权限模式或表名。匹配完整语句身份、参数数量/类型以及 fixture identity；不能靠 SELECT 前缀、关键字黑名单或移除注释来授权。生成查询发生漂移时，静态测试失败，必须审查登记表变更。

| 能力 | 允许范围 | 禁止范围 |
| --- | --- | --- |
| Primary fixture setup | 现有 v7 test closure、历史依赖、辅助表的精确固定建置语句；只在 backup 前 | 安装/替代 00006/00007、任意 DDL、配置外部资源 |
| Primary protocol | 真实 claim/fence/certificate/audit/outbox fixture 的固定读写、observer marker，以及其必要事务 | 任意对象、任意函数、外部文件/网络/扩展执行 |
| Promoted candidate | 真实 repository/registered-handler/readiness 所需固定 SELECT、必要的 SELECT FOR UPDATE、固定快照断言 | 领域 DML、DDL、COPY、SET ROLE/session_replication_role、LISTEN、advisory lock、外部副作用函数 |

Primary setup 的 replica-only 历史依赖仅限已有批准夹具；目标协议写入前必须恢复 origin。Primary 事务不能利用 setup 权限在 backup 后重新安装或重置状态。

Candidate 使用 controller 内部最小权限连接，仅对行锁实际需要的精确表赋予相应数据库权限；wrapper 仍拒绝 DML。不能一律开启 SQL read-only transaction，因为这会拒绝行锁并把真正可用的正向对照变成 database_unavailable。必须保留真实成功对照来证明差异。

固定查询的参数和结果受既有领域字段/集合边界约束，单次访问另有上述 30 秒、64 MiB 和 65,536 行的聚合上限。执行角色、语句及对象集合在计划中逐项列出；不得使用 GRANT ALL、SUPERUSER candidate 连接或 caller 配置的授权集合替代。

### 7.3 生命周期

- 一次 callback 最多一个活动 READ COMMITTED 事务；数据库能力最多持有两个内部连接，以免固定 primary 夹具在事务内需要独立只读探针时自锁。
- 同一 controller 同时最多一个 authority-access callback。不得在活动 callback 中 crash、restore、promote 或改变其 generation。
- 底层 driver Rows 在每次读 API 返回前关闭；超限或物化错误在正常 operation 阶段返回，不延迟到 admission consume 后处理。未消费的已物化结果视图在事务结束时失效，不要求 Commit 路径先执行额外查询或关闭游标 I/O。
- callback 返回、抛出 panic、取消或超时均撤销数据库/事务/row 视图，关闭结果、回滚未完成事务并关闭内部连接；清理连接不是清理 Docker 资源。
- 保留的句柄在 callback 后不能再次查询或提交。不能因 leaked borrowed connection 无限等待 callback 收尾；接口不提供 borrow/raw connection。
- Handler 对其 DBTX 的使用属于可信固定测试代码约束。Go structural type assertion 不能被宣称为权限沙箱；独立回归检查 handler 不提前 Commit，并验证 Coordinator 仅调用一次 Commit。

## 8. 旧 PITR 测试迁移与真实对照

迁移 `TestPITRBeforeRevocationFailsClosed`，移除 authority 包内的 `pitrDockerHarness`、硬编码数据库密码、端口分配、直接 Docker 执行、raw 00006 安装和自管 cleanup。只使用 `authority-v7-pitr` profile 已准备的 registered v7 数据库。

复用 crash fixture 的 registered pgx resolver/material/activator/persisted validator，提取可接收受控 DBTX 的测试初始化逻辑；不能复用旧内存 resolver，也不能为了复用强迫控制器返回 raw pool。

测试按以下顺序运行：

1. 在 primary 建立真实 v7 closure 和领域 fixture，提前准备 A/B 所需的历史依赖和目标证书，提交前缀；真实 Coordinator.CheckReady 必须为 ready。后续 backup 之后只允许协议读写，不再开启 setup/replica 权限。
2. 创建 base backup。之后再通过真实 Coordinator 提交并观察 A，选定 A 为恢复 cut；该提交不使用 synthetic probe 冒充领域提交。
3. 对目标证书执行真实撤销 B，确认 fence/domain/proof/audit/outbox 原子结果、真实 primary readiness 和外部 provider 最新 receipt；观察 B。
4. 释放所有访问能力，经控制器 crash、restore A、promote、inspect；外部 deterministic provider 保持 B，不随数据库恢复。
5. 候选库固定查询证明 A 已存在、B 不存在、目标证书仍是撤销前数据；不能只断言 fence count。
6. 同一真实候选数据库与停留在 B 的 provider 运行 Coordinator，要求精确 `database_behind_provider`，不能接受 database_unavailable 作为替代。
7. 用独立保留在 A 的 deterministic provider 对照运行同一候选读取路径，要求真实 ready；它不替换步骤 6 的外部 provider，也不声称生产回退 provider 合法。

仅在测试中保留的 readiness-gated 查询可以验证失败时返回零数据，但不能称为 Task 10 的同连接 readiness→query→readiness serving。Task 10 的证书/desired 投影与 reconcile 后服务恢复不在此处交付。

旧测试“全新、不相关物理集群”的 identity-mismatch 分支不改造成 fake physical restore。本文不增加创建任意 fresh cluster 的控制器能力；保持已有 deterministic identity-mismatch 单元覆盖，并在迁移记录中明确物理覆盖范围变化。

## 9. 验收矩阵

| 类别 | 必须证明 |
| --- | --- |
| 构建隔离 | 所有新控制器能力/bridge 只在 integration tag 下存在；生产 build 不导出这些符号 |
| 真正事务绑定 | 外来/包装/已关闭事务、重复绑定、错误 controller/数据库、caller fabricated marker 均拒绝；合法 marker 与领域效果同事务 |
| 提交观察 | 普通 COMMIT、rollback、Commit 响应丢失、相同 handle 重试、同一 drain 多个 observation、XID wraparound 关联、缺失/重复/错类终端、indeterminate drain |
| 严格顺序 | marker 绑定在 consume 前；consume→唯一 driver Commit 之间没有额外查询、写入、观察或时间检查 |
| 切点 | A→B→crash→restore A；B 不能覆盖 A；foreign/forged cut、backup 之前的位置、错序、容量耗尽均失败 |
| 访问边界 | 正确的 row-lock/readiness 成功；候选领域写入、SQL 变体/多语句/CTE 副作用、错误参数、raw-connection escape、退出后 Row/Rows/transaction 复用均失败 |
| 生命周期 | callback error/panic/cancel/timeout、未关闭 Rows、双 Commit、post-crash primary、未晋升 candidate、旧 generation 都安全收尾 |
| 物理恢复 | 实际 backup/archive/restore/promotion；目标真实领域状态及精确失败原因；同候选正向 ready 对照；provider B 不变 |
| 资源归属 | 现有 intent/create/actual/cleanup 故障测试及 foreign-canary 仍通过；caller 无新增 cleanup 权限 |

先完成有实际 RED→GREEN 输出的单元/负向行为测试，再执行 compile-only 和真实数据库 gate。编译通过、静态 SQL 检查或 fake controller 不能替代物理恢复证据。最终合并还需要普通 suite、相关 race 检查和独立审查无阻塞项。

权威物理 gate 使用既有 runner 的 `authority-v7-pitr` profile，明确 package、顶层测试名及有限 timeout；不改成直接 `go test` 自管 Docker，也不放宽 inherited-env、构建校验、ownership 或期限限制。缺少 Docker、依赖验证超时等必须如实报告，不标记验收完成。

## 10. 文件边界与实施前置条件

预计实现范围：

- `internal/testinfra/c12_authority_pitr_integration.go`：句柄、观察、cut、scoped access 状态机。
- 同目录 integration-only 能力实现/测试文件：窄 DBTX、固定策略及行为回归；不扩大普通生产构建。
- `internal/testinfra/c12_dependencies_integration_test.go`：保留旧控制路径，增加新接口与错误场景。
- `internal/nodecontrol/authority/authority_pitr_integration_test.go`：替换旧 harness 和旧 handler 接线。
- `internal/nodecontrol/authority/authority_crash_integration_test.go` 及同包 integration-only bridge：仅为共用真实 fixture 做必要提取。
- `internal/nodecontrol/authority/v7_migration_grant_test.go` 及相关静态 guard：冻结新的 build-tag、公开接口和无逃逸约束。
- 文档/实施计划：记录证据、迁移范围与未通过的 gate。

不修改生产 Coordinator、provider、领域协议或迁移来迁就测试。若固定桥接无法接入现有私有 seam，实施应停下报告设计冲突，不能临时放出 pgx/raw pool。Runner 的唯一清理者和安全预算不变；必要的纯测试/符号登记调整须逐项列入实施计划，不授权重构执行器。

本文已获用户批准，下一步编写实施计划并选择执行方式。当前分支继续是未完成检查点，不因设计文档提交而成为 merge-ready。
