# Task 10：完整主状态覆盖契约与分层校验边界设计

日期：2026-10-10。

状态：用户已批准本书面设计及第一增量实施计划；2026-10-10 开始 Native 实施。
仅在批准的隔离工作区实现 L0；不修改 Git 索引，不提交、推送。
实现、测试及最终审查结果另见执行记录，本文不代表已验收。

## 1. 意图、依据与成功标准

为后续完整 sealed state 建立可核查的字段覆盖、所有权、校验责任和提交
前置条件，防止把现有部分信任模型包装成完整持久状态。依据为父设计
`2026-08-23-node-pop-control-plane-design.md` §6.1–6.3、§7.1、§8.2、
§11–13、§18，以及已批准的 statepolicy、statetime、staterecovery 设计。

用户要求先做完整契约和分层校验，不同时接入真实磁盘/硬件。本设计采用
完整覆盖清单、封闭的逻辑类型和明确的未验证结果；不选择“先把现有
statepolicy 序列化、缺什么以后再补”，也不并行构建真实平台适配器。

成功标准：父设计要求的每个主状态安全域都有明确归属；缺失不能默认补齐；
结构有效不等于业务验证、可信来源或可提交；三个持久防护域不互相替代。
这里的“完整”指覆盖契约完整，**不是现有完整磁盘格式已经实现**。

## 2. 分期与本阶段边界

第一实施增量建议为独立纯包 `internal/nodecontrol/statecontract`：
验证按值/自有集合的主状态覆盖模型、字段结构及明确可判定的跨引用一致性，
返回 `ContractModel` 和仍需履行的验证责任。它不生成可提交候选或恢复槽。
这一个增量可单独测试，不要求同时实现证书、恢复协议、主机策略和进程
执行等多个新子系统。

后续独立设计阶段依次为：

1. 各域完整材料与业务转换校验器，包括证书/CA、recovery、部署材料和 LKG。
2. 完整内容的确定性编码、材料闭包、摘要重算与平台 seal 边界。
3. 单写入者、同一版本快照与真实持久化适配器；最终才接入启动编排。

本文不规定生产 wire schema/字节布局、不登记新网络协议、不实现 codec、
hash/seal、文件查找、provider、计数器操作或安全锁清理。后续不得把
第一增量的引用模型直接序列化并命名为完整 production sealed blob。

## 3. 三个防护域及时间的所有权

| 防护域 | 自己拥有的权威状态 | 主状态可保存的内容 | 不允许的替代 |
| --- | --- | --- | --- |
| Agent main | 本文的安全高水位和主状态内容；自己的 counter/key/双槽 | 自身完整内容及其他域的必要绑定/历史高水位 | 不用普通文件或摘要替代 hardware guard |
| Supervisor | 自己的 verified/applied 高水位、pending fault/overflow、transition capsule、独立 counter/key | 已接受的 supervisor 身份、材料绑定和相关高水位镜像 | 镜像不证明当前无故障，不授权 lease 或清 fault |
| LocalSecurityLatch | 独立第三 counter/key/双槽、fault/overflow、ACK 和 clear tombstone | 独立 guard 身份绑定；不复制可清除的权威 fault 集合 | 不能根据主状态“空集合”创建空锁或清除旧锁 |
| TrustedTimeSource | 自己的跨 reboot 防回滚状态和写入责任 | 配置的 provider 身份，以及只收紧约束的已持久 floor（若有） | 不用签名服务器时间或主状态频繁写 NV 代替安全时间 |

三个 counter 身份彼此不同，三个 seal key 身份彼此不同；这只检查声明值
是否隔离，真实隔离仍需平台证明。不能靠不同标签证明底层是不同 NV index。

主状态恢复使用已批准 staterecovery 的主状态规则。父设计 §13.2 的
Supervisor pending-fault future candidate 有专门恢复义务，不能调用普通
“counter 未递增即可丢弃未来槽”诊断代替该义务。安全锁启动检查仍在主状态
读取/任何 LKG 授权之前；本文不改变启动顺序，也不实现该顺序。

## 4. 逻辑封装层次与身份

分开三层，避免自引用摘要与无意义 counter 写入：

- `MainStateScope`：canonical node UUID、authority epoch、host-deployment 身份、
  独立预置 trust-anchor/key-set 身份、三个 guard/key 身份及 time-provider 身份。
- `MainStateContentModel`：§5 的固定安全域，不含本次 commit epoch、scan ID、
  attempt ID、pointer、poll nonce 或易失观测。它是逻辑覆盖模型，不是字节对象。
- 未来物理封装：格式版本、main guard/key binding、state_epoch、完整内容与
  ContentDigest、authenticated seal。BlobDigest 标识包含 epoch/binding 的完整
  已密封对象，存放/计算位置由后续格式设计确定，不把自身摘要算入自身。

模型版本固定为 1；未来物理版本不得借用模型版本声称兼容。authority epoch、
流版本/generation/creation sequence 取 `1..MaxInt64`；guard 值/capacity 沿用
staterecovery 的 uint64 边界。UUID 非零、规范小写 RFC4122；摘要为非零
32-byte 值。不同类型的身份不得互换使用。

同 authority epoch、同 node、同部署/trust-anchor 绑定内的已初始化状态
是第一增量的范围。新安装、跨 authority/identity epoch 迁移、reenroll、
counter reset 与 key replacement 不推测，不作为一般更新接受。

每个材料引用有封闭的 `ArtifactKind` 和明确的摘要角色：签名材料分别记录
canonical payload 身份与原始 envelope 身份；证书记录 exact DER 身份；
本地部署材料记录其完整材料身份。没有签名信封的材料不得伪造 envelope
摘要字段。第一增量不接受任意 kind/string map，也不依据摘要验证内容。

未来完整 blob 必须拥有恢复所需的完整材料闭包，或拥有另经设计批准的
不可变、同等认证且可证明可用的闭包存储。只有外部文件路径/一个摘要不满足
该条件。本设计默认后续完整内容自包含；不授权另建外部闭包存储。

## 5. 主状态必需覆盖表

顶层固定具名字段，不用可缺失的通用 sections map。每个域至少有独立
domain tag、显式覆盖状态、类型化材料引用与该域自己的高水位结构。
下表的“材料”描述完整内容阶段必须覆盖的对象；第一增量只验证逻辑绑定，
所有材料真实性、完整性、语义校验仍在 §7 的未履行责任中明确保留。

| 安全域 | 必须覆盖的逻辑字段与完整材料责任 | 校验依据/所有者 |
| --- | --- | --- |
| AuthorityCheckpoint | authority epoch、node-scoped committed checkpoint；不得混作全局 fence 或任意流的创建序列 | §7.3、§9.1；权威 checkpoint 校验器 |
| RootTrust | active root version/creation sequence/payload、exact envelope 或预置 anchor 身份、threshold/key set；验证本次更新所需的连续 rotation chain | §7.1；root 验签/链校验器 |
| MetadataAndRevocations | metadata version/creation sequence/P/E、完整 key-set、累计 revoked IDs、集合摘要、保留 tombstone 及已知最后授权期限 | §7.1；metadata 验签和撤销历史策略 |
| Identity | identity epoch、lineage/期限、选中 exact leaf 的 issuer/serial/DER/public-key binding、key handle 身份、证书状态；恢复/重试所需的已接受证书与切换绑定 | §6.2–6.3；身份与证书状态机 |
| ServerTrust | bootstrap_server 和 agent_server 两个固定 purpose 的 trust-domain、authority/version/sequence/manifest 身份、CA DER 与状态、累计 deauthorized set/摘要、部署签名绑定 | §6.1；独立 host-deployment trust-bundle 校验器 |
| ReleaseDeployment | approved manifest 的 build 身份/version/digest；installed map 的 version/digest；两者交集与完整材料绑定 | §11.1；build/deployment/release 校验器 |
| ResourceEnvelope | envelope version/creation sequence/digest、node/authority/deployment-role binding；完整资源包及对应 inventory 引用 | §11.1；资源包校验器，不从 desired 放宽资源上限 |
| HostMemoryPolicy | policy version/creation sequence/digest、deployment role/key 与完整签名材料绑定 | §12；主机策略校验器；持续 readback 是外部实时条件 |
| Desired | highest-accepted generation/creation sequence/P/E、inventory version、resource-envelope binding、授权 root/metadata/key 身份、有效期与完整 snapshot 引用 | §7、§11；desired 签名/业务策略；最多 8 个 slot |
| Recovery | 独立 recovery generation/creation sequence/P/E、identity/recovery ID/reason/session version、完整 signed incident/fault bindings、required_action 与有效期 | §8.2；独立 recovery 校验器；不能转成 running desired |
| SupervisorBinding | 独立 guard/key/受信 supervisor 身份、已接受 verified/applied generation 及相应 desired/material digests，必要的 transition 引用 | §13.2；独立 supervisor 证据校验器；不包含易失 lease 表 |
| LKG | 按 slot 排序的已接受 LKG 绑定：exact desired/version/sequence、配置/进程/release 材料、授权 root/metadata/key、composite deadline、supervisor applied/transition 对应关系；保留拒绝回退所需高水位 | §12–13；LKG/transition 校验器；不生成启动能力 |

Scope 的 provider 身份/floor 约束与以上十二域共同构成覆盖契约，不能因为
当前实现只有 Root/Metadata/Desired 就把其他域标记为默认成功。

Root rotation chain 单次最多 32、root keys 1–5、metadata keys 最多 64、累计
revoked IDs 最多 4096；这些是各自对象的限制，**32 不是可截断终身 root
历史的许可**。每个 CA 累计失权集合最多 128；desired/LKG slot 最多 8。
Recovery 中 local/supervisor fault binding 上限分别 64/16。其他完整材料的
字节编码、总大小和各域历史保留上限由对应材料契约阶段定义；第一增量
只承载固定引用及上述有界集合，不读取这些尚未纳入的原始 bytes。

Identity 不能把“最多 4 张服务端 active leaf”误用为完整终身本地证书历史
上限；本地切换/重试材料由身份状态机决定，第一增量只检查其引用绑定。
证书私钥、恢复凭据和原始运行配置 secret 不进入本纯模型，不写进错误信息。

## 6. 缺失、合法空集合及高水位规则

覆盖状态固定为 `Unknown / Present / NeverEstablished`，不用 bool 或 nil
推测含义。Unknown 一律拒绝。Present 必须具备本域全部逻辑必需字段。
空 keys、零摘要或零版本不被解释为 NeverEstablished。

NeverEstablished 仅允许 Desired、Recovery、LKG，表达已初始化但该域从未
接受过对象；它的所有材料/高水位字段必须为空。已有 Present 不能退回
NeverEstablished。普通 stopped desired 仍是 Present，不是“没有 desired”。
LKG 的 Present 可有零个当前可用条目，但不能因此删除已建立的安全高水位；
“无可运行 LKG”与“未曾建立 LKG”必须分开表达。

SupervisorBinding 顶层始终 Present，但内部 verified/applied 两条记录各有
显式 `NeverEstablished / Present` 标签，以表示尚未接受/应用过 desired。
这不是顶层缺域例外；NeverEstablished 子记录必须无版本/材料字段。
applied 为 Present 时 verified 必须 Present 且其 generation 不低于 applied；
相同 generation 的 desired 身份必须一致。LKG 有条目时 applied 必须存在，
每条绑定对应的已接受状态；合法 transition 的更深语义仍由 L2 校验。
Desired 为 NeverEstablished 时，verified/applied 及 LKG 都不得伪装成已建立。

第一增量不引入通用 `NotApplicable` 逃生值。ServerTrust 两个固定 purpose
都不能缺失；是否准许其他部署 profile 是另行设计，不在这里默认放行。
空累计撤销集合可合法，但必须显式存在并与完整材料阶段重新计算的集合摘要
一致，不能以零摘要充当空集合。纯引用层无法重算时不声称该义务已满足。

比较 previous/proposed 时遵循各域规则，不使用一个“所有字段都取 max”的
通用算法：

- 相同流版本/generation 与 creation sequence 的不同 payload 是冲突。
- 已接受的同流版本或序列降低拒绝；版本提升所需的严格序列提升由该域规则
  判定，不能把 checkpoint 直接复制到各流的创建序列。
- root 更新还要求连续链；metadata 撤销集合只能追加、不能重新授权；CA 的
  累计失权集合和 status 转换走自己的规则，不能复用 online-key 状态机。
- certificate serial、leaf 到期时间、wall-clock timestamp 不是通用单调版本。
- Supervisor highest-verified 与 highest-applied 不合并；合法 rollback capsule
  不降低 verified，高水位镜像不替代 Supervisor 自己的验证。
- LKG 删除/失效不清空历史授权限制。Recovery session 完成不重置 recovery
  generation，不恢复普通运行意图。

第一增量必须拒绝明显缺失、错域、错 node/authority、域 tag 重复、不可接受的
覆盖退化与其明确定义的基本高水位冲突；其余业务规则列为未履行责任，不能
返回“没有发现问题，所以可提交”。

## 7. 分层校验与结果强度

| 层 | 输入与职责 | 允许的成功结论 |
| --- | --- | --- |
| L0 结构/覆盖 | 显式 logical scope、previous/proposed 固定十二域、引用类型/数量/范围、合法空值和跨引用 | 仅 ContractShapeValid |
| L1 完整材料 | 拥有的原始对象；严格解析、拒绝重复/未知字段、材料闭包、摘要重算、P/E/DER 精确绑定 | 材料形状和身份闭包成立，仍非授权 |
| L2 各域语义 | root/metadata 验签、撤销历史、证书/CA、部署角色、desired/recovery、LKG 等各自策略和新鲜前提 | 对具体内容与旧基线的有条件业务证据 |
| L3 组合/提交准备 | 所有必需域证据、同 node/authority/base state、内容身份、依赖版本与有效期、独占写入下最终重检 | 完整候选可交给后续 seal/commit 适配器 |
| L4 物理提交/恢复 | 编码、seal、双槽 fsync、硬件计数器、新扫描确认 | 经相应平台验收后的真实提交事实 |
| L5 运行授权 | 两类安全锁、可信时间、证书/当前策略/期限/进程条件的独立新鲜检查 | 仅具体操作所需权限；不是保存成功即可运行 |

第一增量**只实施 L0**。`ContractModel` 是私有拥有的结构化结果，公开摘要
只说明覆盖与结构状态，固定带回 L1–L5 未完成的责任分类。无
`ReadyToCommit`、`Trusted`、`Validated=true`、生产 receipt 或转换为
staterecovery.ProposalObservation 的便捷接口。调用方提交一个同名状态
或“验证成功”布尔值，不能清空这些责任。

未来 L2/L3 的证据必须由对应校验器构造并绑定 exact scope、旧状态身份、
新内容身份、材料 P/E、provider/source/dependency identity、适用期限和
校验规则版本；不是通用 caller-supplied receipt。私有构造只能限制程序
误用，不替代平台真实性。相同 ID 下实际字段变化也必须重检并拒绝。

已过期材料可以是恢复选择/历史约束输入，但不能仅因读取成功重新取得运行
授权。加载持久状态与授权新的在线操作是不同检查，避免把正常到期误当成
“可以清除高水位”，或反过来允许过期 LKG 启动。

## 8. 与现有包的接口边界

- stateformat 仍是它已支持的签名对象的唯一严格解析器，不另造宽松解析器。
- stateverify/statepolicy/statetime 的现有结果分别覆盖自己的条件性证据；
  不扩大其含义，不把 StateModel.Status 从 Unknown 改为 Available。
- staterecovery 仍只判断双槽选择及事件顺序。它的 ContentDigest/Intact
  假设不能被用来倒推 L1–L3 已完成。
- statecontract 第一增量无上述包的生产消费者，不接入启动/存储/网络，
  也不增设对现有隔离 guard 的例外。建议只依赖有限标准库和既有 UUID/
  contracts 基础值类型；确切类型名及允许导入表由实施计划固定。

主状态业务内容不纳入 pointer、scan/attempt ID、time nonce、poll/report/
heartbeat/probe 样本或易失 lease deadline。普通采样不触发主 counter 更新。
安全 floor 若变化也不能把主 guard 变成时间 provider 的写入后端；历史
statepolicy.TimeFloor 的含义需在未来适配时逐项核对，不能直接照搬字段。

## 9. 错误、所有权与安全默认值

第一增量的固定错误分类为 InvalidInput、IncompleteCoverage、BindingMismatch、
Rollback、Conflict、StaleModel 以及 context 错误；错误仅给出有限 domain/reason，
不包含 UUID、摘要值、证书、材料 bytes、路径或秘密。结构不成立返回 nil
结果；结构成立但 L1–L5 未做时，责任清单不能为空，也不能自动调用它们。

nil context 拒绝，预取消优先；成功返回前复查取消。集合按值复制到独立所有者；
摘要不泄露可变内部切片。Clear nil-safe、幂等，清理后结果不可使用；不保证
并发 Clear、调用方并发改输入或运行时所有副本擦除。长度先检查再复制；
不排序输入来掩盖重复/乱序，不截断超限集合来“恢复可用”。

任何失败都不删除内容、不推进 counter、不写 latch、不重试、不恢复 LKG。
未来调用方负责安全停止/故障记录，纯模型错误不是这些副作用已执行的证据。

## 10. 第一增量验收与后续门禁

实施计划至少包含以下命名场景对应的测试，不把后续未实现层伪造成通过：

1. 十二域逐一缺失、Unknown、非法 NeverEstablished、已建立域退化全部拒绝。
2. scope/三个 guard/key 交叉复用、authority/node/材料 kind 互换拒绝；结构
   正确的不同身份也不证明真实 provider 隔离。
3. same version/sequence different payload、较低高水位、counter 与 authority
   数域混淆拒绝；不能把 certificate serial 代作版本。
4. 两种 server purpose、资源 envelope 与 desired、LKG 与 exact applied
   desired 的逻辑引用一致性；恢复快照不能填入普通 desired 域。
5. 合法 NeverEstablished、显式空累计集合和零个可用 LKG 的差别；无隐式补齐。
6. 各集合合法边界、超限、重复/乱序；不把单次链上限当终身历史截断许可。
7. 完整 L0 输入只能得到结构结论与非空固定责任清单，不能变成提交 proposal、
   可信状态或启动许可。所有缺少原始材料的场景明确未通过 L1。
8. 输入/输出独立所有权、nil/取消/Clear、错误不泄露输入；AST 禁 I/O、真实
   时间/随机数/网络/provider 调用和生产消费者。

本轮文档没有运行或新增测试，不产生新的运行时验证结论。后续仅在计划
批准后做固定工具链离线测试；沿用已有包限定回归，authority 限定
TestBoundRead/TestCoordinator，store 只编译。无整仓、DB/Docker、race、
硬件/reboot 或 C11/C12 验收，不新增依赖，不提交/推送。

## 11. 自查与交接

已按父设计核对：主状态必需覆盖项没有被 statepolicy 的现有字段裁剪；
resource envelope、host-memory policy、CA 累计失权历史与 recovery 独立
generation 均有归属；supervisor/latch/time 独立所有权与特殊恢复规则保留。
没有用 opaque digest、私有构造或调用方布尔值冒充材料验证/平台证明。

完整契约与第一可实施增量分开：本设计覆盖所有域，但第一计划只实现
L0 的类型/覆盖/一致性模型。完整材料、格式编码、业务校验器及真实平台
集成各自需要后续设计与验收，不能在第一计划内顺带实现。

请先审阅本文。批准后使用 writing-plans 编写第一增量实施计划，再由用户
审阅计划后沿用 Native 执行。当前未实施任何新代码；生产兼容性 UNKNOWN，
整体 Task10/B01/C12/main 验收条件仍未解除。
