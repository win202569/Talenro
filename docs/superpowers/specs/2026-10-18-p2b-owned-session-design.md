# P2-B 会话所有权、启动握手与失联处置详细设计（待书面批准）

日期：2026-10-18。文档基线：5864e4b8477c05ef87e2f95ab448732fcddf8429。
状态：WRITTEN_DESIGN_PENDING_APPROVAL。
DesignApproved=false；ImplementationApproved=false；RuntimeApproved=false。

用户对“仅编写 P2-B 详细设计，不实现、不运行探测”明确回答“是”，只授权本稿、
设计自检和必要进度保存。本文没有测试、审查、编译、系统/安全查询或运行额度；
不生成实际会话身份、命令体、运行目录、nonce 或可执行运行包。

## 1. 目标、当前事实与成功标准

在任何获准目标运行前建立明确所有权，只有一次释放机会；监督方失联后，
仍有独立、预先持有的停止通路。所有权、控制消息、实际观察和最终清理必须可区分。

[P2-A 限定离线验收](../../roadmap/2026-10-18-p2a-ps7-limited-acceptance-approval.md)
已批准普通 Windows PS7.6.5 指定声明整数向量；产品350CAB1A、测试B5C4E0CE、
fixtureC720E247 不变。这不证明真实时钟、Job、管道或安全条件已经成立。
P1 已有后续限定处置以当前报告为准，不采用旧文档开头的历史未完成状态作为新结论。

继承[P2 上层设计](2026-10-18-p2-adapter-boundaries-design.md)、
[联合架构](2026-10-17-raw-capture-supervision-design.md)和
[真实运行门槛](../../roadmap/2026-10-17-ordinary-ledger-stageb-readiness.md)。
安全 owner/vendor 处置、当前相关安全证据及精确运行批准仍是独立前提；
本轮未刷新这些事实，不宣称历史检测是误报，也不恢复或重放历史载荷。

本稿成功标准是冻结可审查的组件边界、消息格式、失败路径与证据地图，
不是创建能靠 live 开关运行的脚手架，也不是复制第三套生命周期模拟器。

## 2. 选型与一项明确的上层职责细化

选定：L 作为唯一启动代理，W 独立监督，C 只读数据，分层验证后才接真实能力。
同进程计时/采集方案组件少，但不能承担进程级独立监督保证，故不选。
先完成 PS5 是独立兼容路线，不是当前 B 的能力前提；本稿只设计普通 Windows PS7 路线。

**拟议细化，需随本稿审阅批准：**上层曾写“C 创建 T”，本稿改为 L 创建并释放 T。
C 只提交读通路就绪；W 只向 L 发放一次许可。理由是把创建/恢复权限从 C 移出，
缩小提前释放、能力泄漏及退出后重新创建的面。拓扑仍是 L/W 在 J 外、C/T 在 J 内；
不改变实际目标、原 Root/Child 合同或原有限额度。旧上层文本不回写成已实现的新职责。

另一项明确边界：P2-B 不在此冻结真正字节采集或 P3 记录格式。
没有另行设计并验证的 P2-C Ready 绑定时，任何真实 T 都不得释放。

## 3. 角色、依赖与能力表

| 角色 | 仅可持有和执行 | 不得取得/宣称 |
| --- | --- | --- |
| L 调用方/启动代理 | 本次冻结时基、J 的设置/赋属能力、独立停止/查询能力、W/C/T 的创建时确切句柄、T 主线程释放能力 | 不能凭 PID/名称重找目标或替 W 宣称已观察清理 |
| W 独立监督者 | 同一 J 的独立停止/查询能力、L/C/T 的受限观察句柄、有限控制端点和自身时钟 | 无目标启动/恢复能力，不等待 C 的 I/O/finally 才能停止 |
| C 采集方 | 已绑定两通道的独占读端及 C↔W 控制端点；固定内存/队列额度 | 无 J 控制句柄、无 T 恢复权限、无正向发布或清理证明权限 |
| T 目标 | 精确批准的执行输入、目标标准流端点 | 不继承 L/W 控制端点或 J 控制句柄 |
| P3 未来接收方 | B 的有限观察摘要与 C 的独立数据证据 | 不将旧 synthetic 格式当成真实证据或自动发布成功 |

J 为新外层拥有会话；原 Root/Child 工作负载 Job 与 cap32 保持独立。
外层 J 不给旧成员计数“增加包装器配额”：C/T 单独标识、单独有界，
原 cap32 仍只约束原工作负载。首个 B 能力验证方案仅包含两个直接成员 C/T，
不执行 Root/Child 工作负载；未来集成须另定两层 Job 限额和记账规则，未批准则拒绝。
不在本稿暗设34或其他集成限额，不声称已有能力支持嵌套。
Profile1拟将外层J的ActiveProcessLimit固定为2且禁止breakaway，仅容纳C/T；
设置/查询不一致即拒绝。C/T不得再创建其他成员。这个独立能力验证限额不修改
原工作负载cap32，也不能用于宣称包含真实Root/Child的集成已经就绪。

能力对象只存在于其所属进程的受控原生资源表，不能接受调用者传入的裸整数、
PID、反序列化 SafeHandle、字符串 Job 名称或“Verified=true”作为能力。
日志中的本地索引/句柄显示值只用于关联，不能重建控制能力。
同用户/管理员恶意注入、任意进程复制句柄不在本合同的防御保证内；
这不是安全沙箱，敌对目标需要新的隔离设计，不能沿用本验收。

## 4. Windows 候选适配与启动前归属

本节是设计选择与平台资料支持，不是宿主验证：

- Job 的层级、breakaway、最后句柄关闭与通知都有平台条件。
  选择不启用逃逸，通知只作唤醒线索，不用“未收到消息”证明没有成员变化。
  [官方 Job 说明](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects)。
- 选择启动属性中的 JOB_LIST 建立创建时赋属，并用 CREATE_SUSPENDED 保留主线程闸门；
  不回退成“运行后尽快加入”。属性不支持或创建失败即拒绝。
  [启动属性](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute)、
  [创建标志](https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags)。
- 原观察 OpenProcess 的0x101000及inherit=false不变。外层初始化句柄传递是另一操作：
  明确的 HANDLE_LIST 要求其条目可继承并按 API 要求设置创建参数，不能同时声称
  CreateProcess 的全局 bInheritHandles=false。仅启动瞬间的专用副本入白名单；
  无关句柄不加入，初始化完成即关闭多余副本并清除继续继承属性。
  [HANDLE_LIST 官方约束](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-updateprocthreadattribute)。
- 不通过修改系统保护、执行策略、宿主 Job 限制、进程身份或 breakaway 标志解决兼容。
  本机是否允许上述拓扑、属性、嵌套和最小权限，均待另行获准验证。

L 创建 J 的临时设置句柄，不向 C/T 传递。W 得 JOB_OBJECT_QUERY|JOB_OBJECT_TERMINATE；
L 设置期需要 ASSIGN_PROCESS/SET_ATTRIBUTES/QUERY/TERMINATE，设置后只保留所需的
本地受限句柄。任何额外权限须在实施计划逐项列明，禁止 JOB_ALL_ACCESS 作为默认长期能力。
原观察权限与这些新控制能力分表，不扩大旧观察句柄。
[Job 权限说明](https://learn.microsoft.com/en-us/windows/win32/procthread/job-object-security-and-access-rights)。

初次 W 引导只执行固定、审阅过的监督程序，不执行目标；先获 L 的受限观察句柄与
J 受限能力，启动后立即自行验证并监视 L。W 不以无限 suspended 状态持有唯一 J 副本。
L 在 W 完成绑定前不得创建 C/T；W 引导失败或 L 已退出则停止/关闭自己的会话能力。
所有初始化中间状态都必须有表中下一步失败处置，不能遗留无人恢复/停止的挂起进程。

L 创建 C/T 时必须创建时入 J 并保留实际 process/thread 句柄。C 在 W 绑定完成后
可恢复至固定“未武装/等待读通路”程序；T 主线程仍挂起。T 的恢复权限只留在 L。
任何创建成功但绑定/校验失败的成员均在同一预算内走已拥有资源停止路径；
不重新按 PID 打开，不发布“完全未产生进程”的错误结论。

## 5. 内部逻辑接口与数据边界

以下是拟议接口职责，不是本稿新增的函数、脚本或可调用 API：

| 接口 | 输入 | 输出/失败边界 |
| --- | --- | --- |
| AdmitRunBinding | 精确版本/文件身份/参数字节/环境白名单及合法外部运行决定 | 冻结 BindingDigest；未知即拒绝，不由调用者布尔自证 |
| BindOwnedSession | 已准入绑定、真实时基、受控创建资源 | 进程内 OwnedSession 或失败摘要；不返回可序列化控制句柄 |
| ExchangeControlFrame | 精确256字节帧、固定端点角色 | 解析后的固定消息或拒绝；不调用 getter/delegate/脚本块 |
| PrepareRelease | W 的一次 Arm、L 当前状态、C Ready 和实际绑定 | 单次本地释放票据；不接受 P1 的 Would* 替代条件 |
| CommitRelease | 已消费票据、唯一 T 主线程句柄 | released、not-released 或 unknown；失败不重试 |
| StopOwnedSession | 首个故障、固定截止与已持能力 | 请求/观察/剩余未知分开；不把发送停止当完成 |
| SealControlEvidence | L/W 分别形成的有限本地观察 | 版本化只读摘要；不等于 C 字节完整或 P3 归档 |

普通数据入口精确被动字段/整数/字节类型，拒绝额外字段、可执行描述符、隐式强转和装饰标量。
对已准入普通数据做独立快照；真正能力只能由受控创建路径取得，不能靠复制普通对象产生。
角色程序、桥接实现和依赖都须后续实施计划冻结；不把任意外部 Stream 或 callback 注入控制路径。

## 6. 控制传输、固定帧与鉴别边界

选择预建、点对点、支持有界异步读写的本地字节通道，L↔W、W↔C 各两个单向端点；
每端点只有一个写者、一个读者。端点在启动白名单中精确交付，不按可猜名字重新发现。
精确 Windows 管道构造/取消路径由真实能力阶段验证；不采用同步无限 Write/Read。

本地端点所有权和启动文件/能力绑定提供当前威胁模型内的来源关联，不宣称加密认证。
SessionId 是未来运行时生成的16字节关联值，不是权限或身份认证；
BindingDigest 是冻结运行描述的32字节 SHA256，不是签名。本轮不生成它们。

帧总长固定256字节，全部多字节整数 unsigned little-endian；
tick/frequency必须再满足0..Int64.MaxValue及P2-A关系规则。禁止尾随字节和可变长字符串。

| 偏移/长度 | 字段 |
| --- | --- |
| 0/4 | Magic，ASCII P2BC |
| 4/2；6/1；7/1；8/1 | Version=1、Kind、Sender、Receiver |
| 9/7 | Reserved，全部0 |
| 16/32；48/16 | BindingDigest、SessionId |
| 64/8 | 方向内Sequence，从1开始，严格加1，溢出拒绝 |
| 72/8；80/8；88/8；96/8 | OriginTick、Frequency、DeadlineTick、ObservedTick |
| 104/2；106/6 | PayloadLength，0..112；Reserved全部0 |
| 112/112 | Kind专用payload；未使用字节全部0 |
| 224/32 | Reserved，全部0 |

角色编码 L=1、W=2、C=3。种类固定1..12：
Hello、Bound、Ready、Arm、ArmAck、ReleaseAttempted、ReleaseObserved、
Heartbeat、Stop、StopAck、Observation、Fault。其他值拒绝。
方向/状态白名单：Hello/Bound限L↔W或W↔C；Ready为C→W；
Arm为W→L；ArmAck/ReleaseAttempted/ReleaseObserved为L→W；
Heartbeat仅已绑定端点；Stop为L↔W或W→C；StopAck为对应接收方回复；
Observation为C→W或W→L；Fault为已绑定的报告方向。
T无控制消息。C不可提交Arm、释放、清理已证或许可结论。

payload前8字节固定为Subtype=0及7个零字节。下表长度包含这8字节；
后续字段按列出的顺序紧邻编码，U32/U64分别为4/8字节，不使用原生结构体对齐。
payload与帧尾部未使用空间必须全部0，不得包含路径、脚本、裸句柄或可执行动作。

| Kind | PayloadLength | 第8字节起的精确字段 |
| --- | ---: | --- |
| Hello | 16 | Profile U32=1，FrameBytes U32=256 |
| Bound | 16 | Checks U32=7，Reserved U32=0；Checks仅表示发送方声明已完成绑定三检查 |
| Ready | 32 | ReadyBits U32=7，StdoutBytes U32=1048576，StderrBytes U32=1048576，TotalBytes U32=2097152，Events U32=4096，ReadBytes U32=65536 |
| Arm | 24 | ArmSequence U64=本帧Sequence，ReleaseBeforeTick U64=冻结StopAtTick |
| ArmAck | 16 | ArmSequence U64=已接收的唯一Arm序号 |
| ReleaseAttempted | 16 | ArmSequence U64 |
| ReleaseObserved | 24 | ArmSequence U64，Result U32，Win32Error U32；Result=1已恢复/2未恢复/3未知，2或3必须停止 |
| Heartbeat | 16 | LastNonHeartbeatSequence U64，0或该方向已发的最近非心跳序号，必须小于本帧Sequence |
| Stop | 24 | Cause U32，Reserved U32=0，FirstFailureSequence U64 |
| StopAck | 16 | StopSequence U64=对应Stop帧序号 |
| Observation | 32 | ObservationKind U32，State U32，Subject U32，Status U32，Value U64 |
| Fault | 24 | Cause U32，Step U32，Win32Error U32，Reserved U32=0 |

State编码依次为Prepared=1、Bound=2、Ready=3、Armed=4、ReleaseAttempted=5、
Running=6、Stopping=7、Sealed=8、Refused=9、Aborted=10。
Cause仅1..13：binding、protocol、capacity、peer-lost、clock、deadline、create、
ownership、release-unknown、adapter、internal、cancel、unsupported。
Step为同一State编码；0只允许表示状态建立前的准入失败。
非失败帧不额外附带原因。Win32Error=0不能单独解释为成功。

ObservationKind=1本地状态/2进程退出/3本次J活动数/4目标释放状态；
Subject=L1/W2/C3/T4/J5，Status=1已观察/2尚未观察/3未知。
Kind1仅Subject=发送者且Status=1、Value固定0；Kind2仅Subject=C/T，Value仅已观察时为UInt32退出码，否则0；
Kind3仅Subject=J且已观察时Value为0..2，否则0；Kind4仅Subject=T，Value为
Status=1时的1已恢复/2未恢复，或Status=2/3时的3未知。C只可报告Kind1、Subject=C；
W才可向L报告Kind2/3/4。
状态/方向/Subject组合不在上述白名单即拒绝。未观察/未知退出码不由占位0升级为exit0。
Bound/Ready是声明，接收者仍须本地核对能力、固定实现与已批准的P2-C绑定，不能透传为真实权限。
Bound的Checks三位分别为固定程序版本、实际对端能力绑定、角色与J位置核对；
ReadyBits三位分别为stdout读端、stderr读端和共同容量预留。字段为7不替代独立核对。
Profile1只有C/T两个成员的能力验证范围；任何工作负载集成需新profile及独立审阅，
不能扩展这里的计数/字段。此编码表是协议数据合同，不是已可运行的live接口。
CRC/哈希不作为鉴权替代；帧绑定不符、乱序/重复/缺号、未知状态或截断一律停止。

每个物理接收端固定256字节组帧缓冲，单次处理最大一帧；分片只拼满一帧，不无限积累。
每接收端最多一个在途读/写，各固定256字节发送缓冲。未完成写不得再排无界任务。
控制系统全局非心跳待处理上限16帧，共4096字节，静态分为四个接收端各4槽，
不借用其他端点额度，不额外添加排队任务。心跳每接收端一个256字节最近槽，共4槽，
覆盖而不追加历史；本地接收时间独立记，不要求跨进程共享一个可竞争的计数器。
故障摘要总共最多16条、每条1024字节，静态分为L/W各8槽，C只发固定Fault帧而不保存
额外故障历史；每方第0槽保留首因，其他7槽存后续摘要。上限包括序列化后的固定字段。
满则置固定overflow标志并停止，不覆盖首因，不能靠旁路日志继续推进。

## 7. 状态、启动顺序与一次性语义

状态单调：Prepared→Bound→Ready→Armed→ReleaseAttempted→Running→Stopping→Sealed；
任意未释放阶段可Refused，已发生资源活动后故障为Aborted。Sealed仅表示B摘要封存。
失败状态不可恢复Running，也没有Persisted或Acceptance状态。

1. L先冻结真实T0/Frequency/D及准入绑定，创建J与W引导通路。
2. W核对L/自身绑定与时基，确认自己持有J独立停止能力；双方互相Bound。
3. L在J中创建C/T，保持T挂起。向W交付实际受限观察绑定；C仅恢复为未武装采集程序。
4. C准备两条已绑定数据读端和容量后发Ready。W核对固定C版本、方向、绑定及未过期，
   L核对T实际身份/输入/归属。缺少已获验证的P2-C准备能力则拒绝，不用假Ready代替。
5. W发一次Arm，包含该方向序号与固定绑定。L进入Armed并回ArmAck；不是可重发命令。
6. L先原子消费本地票据、标记ReleaseAttempted，再检查取消/失联/停收点，最后只对
   唯一T主线程尝试一次恢复。任何返回不明/ACK丢失按unknown并停止，不再Resume/重建T。
7. L报告本地观察，W基于自己的状态/句柄核对Running。消息到达不重置时基或票据。
8. 正常结束也要进入Stopping并核对J记账；B封存摘要，完整发布留给C/P3独立合取条件。

一次性保证限定为同一存活控制会话。内存票据不是跨掉电事务日志；
L崩溃后不恢复会话、不从文件重放Arm，也不重新创建目标。
未来运行包消费/恢复纪律另行批准；本稿不承诺crash-atomic或exactly-once业务执行。
ArmAck丢失、ReleaseObserved缺失都不能触发第二次目标。

取消/故障与释放竞争由L同一释放临界区串行化，不在锁内读盘/解析/网络I/O；
若释放先发生而故障后观察，记录实际已释放并立即停止，不伪称从未执行。
W的计时/停止线程与解析、数据读取及任何等待ACK的线程分离；L兜底同理。

## 8. 时基、心跳、等待与停止预算

总120000ms包含W启动、绑定、C/T准备、读取、停止和未来交接，不在Ready/Arm处起新表。
真实时基不可校验、倒退或溢出则拒绝；P2-A只用于数值换算，不成为真实QPC证明。
绝对停收点/阶段Until沿用已审阅P2-A公式，now>=边界即到期。
到113500ms对应保守停收点后不得再释放目标；自然500、清理5000、交接1000均在D内。

拟定心跳发送间隔100ms；连续500ms没有本地接收的有效新序号帧即失联线索。
重复/旧帧不刷新最近时间；对端ObservedTick不决定本地超时。
心跳失联阈值不是额外自然等待预算，也不替代process-handle退出观察。
等待上限为min(下一100ms检查点、阶段剩余、D剩余)，不足1ms向下为0，
不调用无参无限wait，不在D后加drain/cleanup时间。OS调度不提供硬实时保证，
超限如实记录；计时线程仍可能因系统失效不能及时执行。

首因一经记录保持不变；自然完成观察不得覆盖早先失联、身份不符、容量或截止故障。
StopAck只证明控制消息被收到；TerminateJobObject成功只证明请求返回，不是全员退出。
到期或错误后的停止动作只作用于已有J和持有的确切W引导进程能力，不按PID扫描清理。

## 9. 失联、双失效与清理证据

| 情形 | 有限处置 | 证据边界 |
| --- | --- | --- |
| W退出/失联 | L不再释放，使用预持J停止能力，独立查询/观察，收束已拥有W资源 | W的文件或exit0不证明J已空 |
| L退出/失联 | W停止J、不授权新的Arm/发布 | 自己的停止请求与实际记账分开 |
| C退出/读错/控制断开 | W停止J，L保留兜底；C前缀留给C层判定 | 不补造EOF或完整字节 |
| L/W同时失效 | 仅依赖预先设置且后来获证的kill-on-last-close策略 | 无存活观察者，CleanupStatus=unknown，不发布成功 |
| 创建/绑定中断 | 未释放目标不再推进；已拥有挂起成员也通过J收束 | 不因“未运行用户代码”就忽略进程残留 |
| 句柄泄漏/Job属性/嵌套不支持 | 释放前拒绝，已产生资源则按失败收束 | 不开breakaway/不加权限/不改系统求通过 |
| D到达仍不完整 | 停止正向判定、保存超限及unknown | 不重开时间窗或宣称准时清理 |

L/W持有J副本数量与关闭顺序在自己的资源表逐项核对；C/T白名单不得含J控制能力。
进入Stopping后封闭所有创建、赋属和恢复入口；不得在最终记账后重新加入成员。
只有所有本次已建资源均有处置记录，才可说“本地句柄表已收束”；不等于OS无成员。
不使用Job signaled作为“任意原因全部成员退出”的通用条件。
最终J活动计数为0、成员枚举为空且未截断、已持成员确切句柄退出证据，
以及查询窗口/资源身份均完整，才形成有限的CleanupStatus=observed-empty。
计数/枚举冲突、未知句柄、查询失败、失去观察者或超时一律unknown。
通知与PID仅作线索，不能用重用PID补齐丢失的实际句柄证据。

## 10. B摘要与C/P3边界

逻辑B摘要包括Schema、Mode=p2b-control-evidence、BindingDigest、SessionId、
State、FirstFailure、T0/Frequency/D、ReleaseAttempted/ReleaseObserved状态、
StopRequested/StopObserved、CleanupStatus、各方有限观察及溢出/缺失标志。
unknown不是false；exit未知保持null。普通字段和数组由持有方独立构造，不夹带能力。
精确真实持久化编码属于P3，不能把本段当作已经冻结的外部archive API。

B自身不声称CaptureComplete、双EOF、原始字节完整、来源已认证或归档已验证。
事实摘要与外部运行授权分离；声明运行获准的字段不能由调用者输入布尔透传。
本稿所有当前状态的ImplementationApproved/RuntimeApproved保持false。
未来能否产生正向技术候选，须同时满足B、C、P3和真实授权的各自条件。

## 11. 验证地图与后续实施边界（本轮0次）

| 类别 | 未来必须独立覆盖 | 不可由离线通过推断 |
| --- | --- | --- |
| 协议与有限内存 | 精确256字节、字段/方向、序号溢出/重放、分片/截断、全局16帧、心跳槽/故障满 | 真实端点鉴别或异步取消有效 |
| 启动闸门 | 无Ready/无Arm/过期/重复/绑定错、消费后失败、取消竞争、ACK丢失不得重发 | OS没有先执行后归属窗口 |
| 创建/能力 | JOB_LIST、暂停/赋属、W外/C/T内、每个中间故障、最小权限与泄漏拒绝 | 本机兼容、控制句柄从未被外部复制 |
| 失联与截止 | L死/W死/双死/C挂、解析阻塞、113500/120000等值、旧帧不能续期 | 硬实时停止或掉电恢复 |
| 清理与证据 | 请求未完成、计数冲突、截断/未知、成员句柄退出、最终0/空 | 仅exit0或无通知证明全部清理 |
| 集成边界 | C没有发布/释放权、P2-C缺失禁止T释放、原cap32独立、P3未批准不发布 | 完整采集/归档/安全/C12/main验收 |

本稿批准后，**唯一建议的第一份实施计划是纯控制帧编解码与准入校验**：
实现§6逐kind精确payload合同、解析/独立literal测试，不含Process/Job/时钟/外部流/
句柄/系统调用、状态机重建或任意callback；角色/时基/状态值只作为声明核对，
输出不得升级为真实所有权或启动许可。它不是P1的替代，也不实现本稿逻辑能力接口。
计划须另定有限次数/每次上限/顺序/失败即停/条件审查，明确批准后才可写代码运行。

真实能力适配、故障验证、与C/P3集成各需后续详细计划、安全前提和精确运行批准。
纯协议成功不绕过这些门槛；当前尚无任何新目标或reviewer额度。

## 12. 文档自检、提交和下一道批准

本轮仅只读现有设计/批准记录与官方资料，新增本稿、更新current-status/recovery。
无产品导入、测试/能力/安全探测、原生编译、nonce、运行目录、目标脚本或新审查代理。
原P2-A/P1/合成采集器/store/分类器、旧计划/结果/额度和ignored原始记录保持不变。

自检：一处上层职责细化已显式标出；真实能力与普通数据不互换；
启动时赋属与挂起闸门分工明确；继承白名单与旧inherit=false不混淆；
全局16帧计量、不追加心跳历史；一次尝试不等于crash-atomic；
失联、未知与正向证据分离；固定payload表无自由扩展；未支持的集成限额/live能力明确拒绝而非默认放行。
资料支持候选机制，不证明本机事实；没有以文档补写owner/vendor放行结论。

提交推送现有开发分支并核对远端一致，换机从本稿/当前状态/恢复说明继续；
clone不恢复能力、原始记录、缓存、会话或执行权限。
请用户审阅本稿。明确批准后只编写§11第一份纯协议实施计划，
实施计划仍须单独审阅和有限执行批准；本稿当前不授权实现或运行。
