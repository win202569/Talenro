# 主状态材料校验：根、元数据与 desired 子集设计

日期：2026-10-10。

状态：用户于 2026-10-10 批准本书面设计及实施计划；本地实现和限定验证完成。
证据见对应执行记录；Git 索引未修改，未提交或推送，项目整体验收仍有阻塞。

## 1. 意图与成功标准

在主状态 L0 结构契约之后，增加可独立验证的原始材料一致性组件。
选择“整体边界设计、分批实现”：第一批只检查根集合、根轮换、元数据和
desired 的真实字节、内容摘要、信封摘要及明确列出的引用关系。
不一次实现十二个安全域，不把部分材料通过解释为完整主状态验收。

依据为父设计 `2026-08-23-node-pop-control-plane-design.md`、
`2026-10-10-task10-main-state-contract-design.md`，以及现有
`stateformat`、`statecontract` 的格式与隔离边界。

成功标准：替换内容、替换信封、遗漏本批依赖或篡改预期字段均被拒绝；
合法历史材料可保留；结果只证明指定批次的材料与预期引用一致，不能用于
启动、业务授权、提交、恢复放行或消除完整材料闭包义务。

## 2. 组件与依赖

新增纯内存包 `internal/nodecontrol/statematerials`。它只接受显式输入，
依赖基础 `contracts.Digest`、UUID 值类型和唯一严格解析器 `stateformat`。
标准库仅用于上下文取消、错误、值比较、受限集合操作等纯计算。

不导入 `statecontract`，不增加其生产消费者例外，不读取其私有模型。
本批输入是独立、明确的字段投影；今后将投影与完整 L0 模型绑定，需要
单独的组合校验设计。本批不提供自动适配器或“全模型已检查”的入口。
也不调用 `stateverify`、`statepolicy`、启动流程、存储或平台 provider。

所有 wire 解析、原始字节摘要计算仍属于 `stateformat`。
新包不得复制 JSON schema、重新编码材料以凑摘要或增加宽松解析路径。
没有新网络协议、持久格式、依赖安装、磁盘、网络、硬件、随机数、环境
变量或当前时间读取。

## 3. 输入与边界

输入包含公共 scope（canonical 非零 RFC4122 node UUID、正 authority epoch），
一个有界材料表及同一批次的预期记录。材料类型封闭为 RootSet、
RootRotation、Metadata、Desired；未知类型拒绝。

RootSet 接受已有 `node-state-root-set.v1` 原始 canonical payload，
允许历史或当前版本；不是 initial-root 信封，也不具有初始信任含义。
其余三种接受现有 canonical 信封完整字节。RootSet 的预期信封摘要必须为零，
其余类型 payload/envelope 摘要均必须非零。

每个材料表项恰有一个对应的预期记录；记录可被多个关系引用。
重复表项、重复预期记录、未使用材料、缺失材料均拒绝，不自动去重、查找或补齐。
同一类型和 payload 摘要只能对应一个信封摘要；同内容的另一信封不得静默替代。
多处引用同一个唯一表项则允许。

本地批次预算：材料表与预期记录各最多 128 项，轮换记录最多 32 项；
全部输入原始字节总计最多 16 MiB。RootSet payload 最多 65,536 字节；
信封最多 262,144 字节，解码 payload 最多 65,536 字节。
先检查长度和溢出安全的累计预算，再复制或解析。空批次拒绝。
这些是单次 API 工作量限制，不是根版本、历史寿命或全系统存量上限。

预期集合必须显式列出以下字段，不使用“零值表示忽略”的可选比较方式：

| 类型 | 从材料核对的字段 | 仅属外部关联的字段 |
| --- | --- | --- |
| RootSet | epoch、sequence、version、threshold、有序 key IDs、payload 摘要 | 无信任锚证明 |
| RootRotation | epoch、sequence、previous/new version、时间窗、payload/envelope 摘要、内嵌新根字节身份 | old/new RootSet 引用 |
| Metadata | epoch、sequence、metadata/root version、时间窗、有序 key IDs、revoked IDs、payload/envelope 摘要 | RootSet payload 引用 |
| Desired | epoch、node ID、sequence、generation、inventory version、resource version/digest、时间窗、slot IDs、信封 signing key ID、payload/envelope 摘要 | Metadata 与 RootSet payload 引用 |

epoch、版本、sequence 等正整数字段遵循既有 MaxInt64 边界；集合顺序与
唯一性遵循现有严格格式。所有材料的 epoch 均必须等于公共 scope。
时间比较使用显式 Unix 秒与解析后的值，不读取时钟。
输入预期字段本身不是可信来源，只是本次调用要求核对的声明。

## 4. 唯一解析器的必要扩展

现有解析器已验证 canonical schema、字段格式、payload SHA-256 和信封
坐标一致性，并保存原始 payload/envelope。只增加以下格式事实能力：

1. RootSet 原始 payload 的独立格式入口，复用 `decodeRootPayload`。
   使用独立的 ParsedRootSet 结果类型，不把任意版本标记为 InitialRoot。
   既有 InitialRoot 信封入口仍只接受 version 1，不放宽其约束。
2. RootRotation 暴露严格解析后内嵌 `new_root_set` 的原始 canonical
   JSON 字节及其 SHA-256。通过解析器内部取得原始子对象片段，不从 DTO
   marshal 重建；片段仍走根集合严格校验。初始 payload 与内嵌根使用相同
   根集合身份算法。轮换整体 payload 摘要保持原语义，绝不替换为根摘要。
3. Desired 增加有界、独立复制的 slot IDs 格式事实访问器；全部 process
   字段仍由现有解析路径严格检查，不增加业务投影或进程执行模型。

新增所有者及访问器沿用现有 detached-copy、nil-safe Clear 规则。
原有解析入口和访问器结果保持兼容；扩展不增加签名有效性结论。

## 5. 校验流程与历史关系

入口先检查 context、scope、预算、输入结构及唯一性，随后复制输入，逐件
调用唯一解析器。逐件核对预期字段和实际摘要，再建立只含本批材料的索引，
检查关系。所有成功返回前再次检查取消；错误路径清理临时拥有的材料。

根关系：轮换的前后版本必须连续，old/new 引用必须命中本批根集合，
new 引用的摘要及字节必须等于解析得到的内嵌根。每个根版本在本批唯一，
同版本不同内容拒绝；同一轮换连接不得分叉、重复或形成循环。
同一连续链中的 sequence 严格递增，轮换 sequence 等于内嵌新根 sequence；
新根 sequence 大于旧根 sequence。重复根 payload 或轮换 payload 拒绝。
允许保存多个已断开历史片段；不把片段连接解释为从预置信任锚验证的完整链。

元数据关系：引用必须命中本批 RootSet；其 root version 与该根版本相等，
epoch 一致。元数据 wire 没有 root payload 摘要字段，根摘要匹配属于
调用者声明的关联一致性，不能称为已签名的根摘要绑定。

Desired 关系：node/epoch 必须匹配 scope；关联的元数据及其根必须在本批，
签名 key ID 必须属于所关联元数据的 keys，外部根引用必须与该元数据
声明的根关联一致。检查 key ID 存在不等于检查 key 状态、时间授权或签名。
Desired wire 没有 root/metadata payload 摘要字段，这两条关系不能作为
签名真实性证明。资源只核对 desired 内嵌 version/digest 与预期相等；
不声称资源信封本身已经校验。

历史材料不要求等于当前版本，也不因当前时钟下过期或较新元数据撤销
而删除。现有解析器对材料内部时间窗、状态和撤销字段一致性的检查保留。
本批相同 desired payload 的原始 metadata/root/key/resource/inventory
关联不得出现相互矛盾的声明。跨批次原始关联保持与防回滚仍属后续组合
校验责任；本包没有持久记忆，不伪称能检测未提供的旧绑定。

## 6. 输出、所有权与错误

概念入口为 Evaluate(ctx, Input) 与 Recheck(ctx, Model, Input)。具体 Go
命名在实施计划中定稿；必须维持本节语义，不增加授权入口。

成功返回私有、自有输入的 MaterialSubsetModel；摘要只显示 scope、核对
数量、逐类型覆盖和固定未完成责任。状态命名限定为 SubsetConsistent，
不出现 Trusted、Ready、Committed 或完整 MaterialClosureValid。

未完成责任固定包含：其他材料及历史闭包、签名及外部信任锚、域业务语义、
与完整主状态的同快照组合复核、物理提交、运行授权。MetadataHistory、
RevokedSetDigest 及撤销 tombstone 的完整性不在本批覆盖内，不为其发明
摘要编码算法，也不以当前 revoked IDs 列表替代完整历史。即使批次四类材料
全有，也不得移除这些责任。某类未提供时覆盖显示 NotProvided，不能显示
已验证该域不存在；关联所需的本批类型缺失仍必须报错。

所有输入 bytes/slices、结果摘要中的集合及解析器访问器都不共享可变存储。
Recheck 重新解析所提供原始材料并核对与已拥有输入一致，不只比较缓存摘要。
材料表顺序不影响身份比较；嵌套有序集合仍必须使用规范顺序。
Clear 清零可达自有可变存储，幂等、nil-safe；清理后不可重检为成功。
不保证清除 Go 运行时、不可变 string 或调用者独立副本；不支持 Clear 与
同一模型的访问并发。调用期间调用者不得并发修改输入。

错误至少区分 InvalidInput、LimitExceeded、InvalidMaterial、MissingMaterial、
DuplicateMaterial、ReferenceMismatch、ClearedModel 和 context 取消/超时。
任何失败均返回 nil 模型，不返回部分成功；错误信息不回显原始材料。
取消检查位于入口、每件材料前后、关系遍历及最终返回；不承诺中断单次
有界解析器调用，也不通过后台 goroutine 绕过取消。

## 7. 测试与验收边界

采用 Native 执行方法，后续计划先安排失败测试，再最小实现。覆盖：

- 原始内容替换、声明摘要伪造、同 payload 不同 envelope、未知格式及多余字段。
- RootSet 与 rotation payload 摘要不可混用；内嵌根字节及身份一致；
  version 1 初始限制保持；高版本历史根可解析但不产生信任结论。
- 缺失、重复、分叉、旧根或新根错误、版本/sequence/epoch/node 不一致。
- 元数据 key/revoked 列表、desired slots、资源引用和每个预期标量的篡改。
- 旧元数据/根、历史过期材料、当前撤销不抹去旧材料；同材料关联冲突拒绝。
- 每项数量/字节预算的边界及超限、空批次、未知类型、nil/清理后重检。
- 每个嵌套集合独立所有权、访问器复制、Clear 不修改调用者输入。
- 预取消、遍历中取消、成功返回前取消、失败路径无结果泄漏。
- 产品依赖白名单、禁止 statematerials 的生产消费者、原 L0 隔离测试不放宽。

回归限于新包与 stateformat，以及已批准的 statecontract、staterecovery、
statetime、stateverify、statepolicy、serving、contracts、readiness；authority
仅 TestBoundRead/TestCoordinator，store 仅编译。保持固定离线 Go 环境。
不运行全仓库、数据库、Docker、race、硬件、重启或 C11/C12/PITR 验收。
测试成功只表示此增量行为满足本文，不改变项目整体验收阻塞。

## 8. 交付与后续批准

本设计预计只涉及新 statematerials 包、stateformat 的上述小范围事实扩展、
对应测试及进度文档；不触碰其余产品流程。实施时保留现有未提交内容。

书面设计获批后才编写实施计划；计划审阅通过后才实现。沿用用户已选的
Native 方法，不重复询问执行方式。不暂存、提交、推送或合并。
本文件已自查范围、材料身份、历史绑定、隔离、所有权及未验证责任；
这不是独立安全审查，也不是产品测试或实现完成声明。
