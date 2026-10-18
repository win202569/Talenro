# P2-B 会话与握手消费合同详细设计（待书面批准）

日期：2026-10-19。基线提交66e457272b3952d90c8212bd9fef4c4383c6f039。
状态：WRITTEN_DESIGN_PENDING_APPROVAL。
DesignApproved=false；ImplementationApproved=false；RuntimeApproved=false。

用户明确批准“按推荐路线编写详细设计，仅文档、不实现、不运行、不授予释放权限”。
本稿是对既有P2-B设计的合同细化，不是新的执行许可、第三套生命周期模拟器、
可运行脚手架、测试包或真实能力证明。本轮0产品目标/测试/编译/新审查/安全查询。

## 1. 意图、成功标准及不变基线

意图：让后续实现者知道谁消费已接收消息、谁拥有可信的本地状态、
何时只能记为对端声明、何时才可能取得一次性本地票据，以及任何不确定性
如何关闭正向路径。已验收codec/inbox不应被重复实现，也不能升级成真实能力。

成功标准是合同无歧义：普通数据准入、协议消费、实际能力和最终证据分开；
缺少真实前提时保持拒绝，而不是让传入的Verified=true补齐条件。

| 依据 | 固定身份/当前结论 |
| --- | --- |
| P2适配边界设计 |A8E8AD45081E73D526324AF15153C0DDB90E531A92A10C03BF6DAF40A1D52705|
| P2-B会话所有权设计 |41BD5F7225D65D7BAE903E693946A2595C5036A2408B16BECE37965C622E36CE|
| codec限定批准 |4BF7F1779AA98FC4682FBE94E32B11F59D81726B98A1F744E3E0DA553D9292BB|
| inbox限定批准 |8381F8674BC86D6D2C6F6CCDDA90430977F69F2A7BF87CE8D48DE148FC6D8644|
| 固定codec产品 |9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C|
| 固定inbox产品 |47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC|

依据分别见[适配边界](2026-10-18-p2-adapter-boundaries-design.md)、
[会话所有权](2026-10-18-p2b-owned-session-design.md)、
[codec批准](../../roadmap/2026-10-18-p2b-codec-limited-acceptance-approval.md)和
[inbox批准](../../roadmap/2026-10-18-p2b-inbox-limited-acceptance-approval.md)。

P1仍是限定离线政策，不代替真实观察；P2-A只做已批准数值换算。
codec/inbox的版本、wire种类1..12及字段、四方向/16普通槽/4心跳槽/
4组帧缓冲、原工作负载cap32、总120000ms和真实前提均不变。
inbox Minor2的scalar ETS直接断言缺口继续按原批准延期，不在本稿补齐。

## 2. 选型、组件边界及逻辑接口

选择细化真实适配之前的消费/所有权合同。另一方案是另写可回放全局事件
的纯状态模拟器，会重复P1且不能证明真实前提，故不选；直接实现进程/
管道/Job的方案需要独立安全与能力前提，当前不进入。

| 层 | 唯一责任 | 不能据此宣称 |
| --- | --- | --- |
| codec |精确帧/类型/payload、声明绑定及Expected引用检查|来源、状态、实际能力|
| inbox |字节组帧、逐方向入队连续序号、容量、心跳覆盖和首因停止|握手合法或允许释放|
| 本地握手消费方 |从自己拥有的inbox消费一次、核对本地阶段与跨方向引用|对端声明已变成本地OS事实|
| 本地资源/能力所有者 |受控创建表、确切句柄、真实时钟、释放临界区、独立停止|任意消息或普通对象能创建能力|
| 观察/交接方 |分别保存声明、本地尝试、本地观察及unknown|发送Stop等于退出或发布成功|

以下仅为逻辑接口职责，不冻结新公开API：
ConsumeOwnedControl只接受内部固定inbox一次交付，不接受外部Message/事件列表；
DeriveLocalExpectation从所属角色私有账本取引用，不接收调用者Expected覆盖；
PrepareLocalRelease在真实本地前提成立时建立一个私有票据；
CommitLocalRelease原子消费该票据并限定一次尝试；
LatchFailureAndStop先关闭正向闸门，再处置已拥有资源；
SnapshotLocalEvidence产生有界独立普通数据，不夹带能力或可恢复票据。

这些逻辑操作的实现必须位于固定可信角色程序内。不能接受Stream、
任意回调/脚本块、用户提供的句柄整数、PID、名称、反序列化token，
也不能把纯数据对象注册为真实能力。测试控制器不充当生产角色。

## 3. 谁拥有状态，谁只能报告声明

L是唯一创建/恢复T的主体；W在J外独立监督；C在J内只读双通道；
T没有控制协议。L/W分别持自己的确切资源表和独立停止能力。
真实资源构造仍按上层设计另行验证，本稿不实现或证明它。

每个角色只拥有一个本地握手账本：固定Binding/时基、实际端点关联、
逐对端Hello/Bound一次标志、C Ready声明、唯一Arm引用、票据状态、
逐方向一次Stop引用、最近有效本地接收时刻、首因和有限观察。
没有共享可竞争的“全局State”；同一时刻各方对对端的知识可以不同。
收到Observation.State不直接覆盖本地阶段。

逐方向发序号由唯一发送者拥有；接收准入序号只由inbox拥有。
普通消息消费方记录最近已消费普通序号供单调检查，不另建准入序号表。
ArmSequence来自W实际唯一Arm帧的方向序号；StopSequence来自对应
方向实际唯一Stop帧。发送方不能由caller提供这些值或从普通快照恢复它们。
为写入分配的序号不在写入不明后重用；UInt64.MaxValue之前关闭发送，
不wrap、不重新建立会话续号。写失败后不通过跳号/重试重新同步。

资源表内的关联记录不是普通布尔字段：
对端程序/角色/位置、J设置/赋属、T实际身份、主线程唯一持有、两通道
P2-C Ready绑定必须由已批准的受控创建/查询路径关联到本地资源。
本地代码负责其生命周期，外部不能通过设字段令其成立。
P2-C真实准备绑定目前未完成；没有它就不能出票或发Arm。
安全owner/vendor放行、精确运行包签发和宿主能力验证也都是外部前置；
本轮没有刷新或取得这些前提。写作批准和调用者RuntimeApproved布尔不能代替它们。

## 4. 从inbox到握手：两个序号边界不可混同

固定数据流是字节片段进入已拥有inbox，经准入后，普通帧由唯一本地
消费者Take一次并立即解码/消费；心跳走固定最近槽，不复制到普通队列。
不开放“给定任意帧就推进本地状态”的公共入口。

**入队序号不是普通出队序号。** 例如普通1、心跳2、心跳3、普通4，
普通消费者看到1/4是合法的；不得按普通消费游标要求4=1+1。
inbox已经检查整个方向的连续准入及心跳LastNonHeartbeatSequence。
消费方可要求普通序号严格大于自己最近消费的普通序号，不能据此声称
再次证明中间帧完整；心跳缺口的准入保证来源仍是固定inbox。

所有权关联必须来自同一可信调用路径对同一私有token的直接Take结果。
调用者复制出来的Frame、相似返回根、Accepted=true或Snapshot不能建立该关联。
Take后的独立byte[256]在本地固定路径立即消费；不外借可修改引用，
不在中途接受外部数据替换。每次交付最多一次消费，不重放旧帧。

消费方复用固定decoder核类型/payload。codec的全量Test-P2BControlBinding
假定Expected.NextSequence是准入时的方向游标，不能拿最新inbox游标校验旧队列帧，
也不能把本帧Sequence抄作Expected来冒称独立验证连续序号。
在上述可信inbox交付路径，连续准入由inbox负责；握手层仅补本地阶段和
Arm/Stop等关联检查，沿用codec定义的引用语义，不另写第二套wire解析器。
未来若需要复用全量checker，必须有真正的准入时预提交游标接缝和单独计划；
本稿不为此改inbox或增加每帧历史/凭空构造receipt。

四个方向独立排序，没有跨方向到达全序。Ready→Arm等跨通道顺序由W自己的
账本和串行操作建立，不比较不相关方向的序号大小。
消费异常、绑定错或非法阶段先锁存首因并停止，不把错误帧放回队列继续尝试。

为避免心跳覆盖掩盖更早普通帧的阶段检查，未来角色在处理某方向新的
heartbeat-updated事件前，先按FIFO消费该方向更早的普通项，最多4项，
每项都检查本地截止/停止；不为等待空队列作阻塞读取。
入队与该处理纪律由同一角色的串行控制路径协调，无额外无界事件队列。
不能仅晚些读取Heartbeat快照就认定它在正确阶段到达。
协议不合格的新帧即使已经入队，也不得刷新握手有效接收时刻。

## 5. Hello、Bound、Ready与Arm的消费顺序

每一双向链路各方向最多一次Hello和一次Bound。发送Hello需已有实际关联
的本地端点；发送Bound需已收合法Hello并完成本地固定三检查。
双向Hello均成立、自己Bound发送完成且对端Bound已合法消费，才记该链路Bound。
Checks=7只记对端声明，不替代自己的检查。

| 消息/本地事件 | 接收方前提 | 允许的本地推进 | 不允许的推断 |
| --- | --- | --- | --- |
| L↔W Hello/Bound |实际引导端点关联、固定Binding/时基；每项一次|分别记录；双方完成后LWBound|尚未创建C/T也能宣称其就绪|
| W↔C Hello/Bound |LWBound；受控创建的C/T及观察关联已建立|分别记录；双方完成后WCBound|C得到J控制能力|
| C→W Ready |WCBound，固定payload，首次Ready|只记Ready声明；实际P2-C绑定另核|ReadyBits=7等于双读端已验证|
| W本地准备Arm |LWBound/WCBound、Ready声明与实际绑定、T/停止能力及有效时基均满足|分配唯一W→L Arm序号并发送一次|W获得T恢复句柄|
| W→L Arm |L已Bound，持T确切身份/主线程/停止资源；唯一Arm；本地未到StopAt|记录Arm引用，进入Armed，准备ArmAck|仅凭合法Arm即可释放|
| L→W ArmAck |W已经实际发过唯一Arm，未停止；引用相等，首次ACK|只记对应ACK|T已释放或可再发Arm|
| L→W ReleaseAttempted |同Arm引用，ArmAck已消费；首次声明|W记录L的尝试声明|W已经实际观测到恢复|
| L→W ReleaseObserved |同Arm引用，Attempted已消费；首次声明|记录结果；仅Result1有资格结合本地实际观察推进|Win32Error=0单独代表成功|
| Heartbeat |链路已Bound，方向准入有效且本地未终止|更新本地有效接收时刻及最近槽视图|重置起点、票据或截止|
| Observation |链路Bound、角色/kind/subject符合既有表|更新有限对应观察项|用远端State替换本地阶段|

任何重复Hello/Bound/Ready/Arm/ACK/Attempted/Observed，或提前/反向引用、
引用不等、错会话，均为protocol/binding故障；不靠幂等重发继续会话。
C不能发送Arm/释放结论/清理已证。Result2未恢复或Result3未知均关闭正向
路径；Result1也只是L声明，W还需自己的受限观察关联和实际查询。
缺失或不明不补为false/exit0/成功，亦不触发第二个T。

L/W对同一资源的检查各自保留，不能以“对端已Checked”替代。
W无法独立取得所需真实恢复观察时仍为unknown，不能仅凭L的Result1发送
“已观察恢复”或建立真实Running结论。C自报Running至多表示自身读循环，
不证明T运行；C自报Armed/ReleaseAttempted与其角色不符，应拒绝。
LWBound后L才可创建C/T；C仅恢复至未武装采集程序，T仍挂起。
P2-C前提缺失时设计路径停在拒绝，不发送假Ready，不为演示填入真实就绪。

## 6. 一次性本地票据：存在、消费与实际尝试分开

票据仅属于L的存活会话、固定Binding和唯一T主线程资源项；不进入wire、
快照或持久化，不是nonce字符串/普通对象可重建的能力。本轮不创建票据。
票据状态只有Absent→Prepared→Consumed或Prepared→Revoked；
Consumed/Revoked均不可回到Prepared。同一会话最多一个Prepared机会。

Prepare需要§5全部本地资源前提、唯一Arm以及未停止/未取消/未过期。
准备票据本身不执行恢复。ArmAck只作协议报告；其有界发送不明或失败
即撤销并停止，不把“可能已收到ACK”当作可以尝试的理由。

Commit在L同一释放临界区内执行：
先确认唯一票据身份和Prepared，原子改Consumed并记本地ReleaseAttempted，
随后再次核本地停止/取消/对端有效性/真实now<StopAt及资源仍属本次；
任一失败只停止，Consumed也不恢复。通过后仅对同一个T主线程尝试一次。
此处本地ReleaseAttempted表示票据已消费、进入尝试阶段，不证明原生调用
已经发出；须另区分确定未调用、观察到调用返回和调用是否发生不明。
不能拿调用前的标记、控制消息或调用后缺失记录补造实际执行证据。
临界区不做消息解析、写盘、网络/管道等待或外部回调。
原生调用失败/返回不明/未完成不重试，不重建T，不再开票。

本地Attempted先记后调用，不依赖控制消息是否送达。ArmAck先完成其
本地有界写；Commit返回后，固定发送者按Attempted→Observed顺序有界报告，
每端点仍仅一个在途写/一个256字节发送缓冲，不增无界队列。
发送不明会导致停止/unknown，不重做恢复。W等待报告时仍独立计时/停止，
不能等待ACK或C的I/O才能行使停止能力。

取消与Commit竞争使用同一L正向闸门：停止先锁存则不能尝试；若恢复先发生，
如实记为已尝试/可能已释放并立即停止，不伪称从未执行。
W的实际停止也可与L交错；这不是跨进程原子事务。查询不明按unknown收束。
保证仅限同一存活会话的至多一次尝试，不承诺crash-atomic或业务exactly-once。

## 7. 本地阶段、Stop、首因及有限证据

上层State编码1..10不变。阶段是每角色自己的可解释进度，不是全局同步事实。
健康路径遵循Prepared/Bound/Ready/Armed/ReleaseAttempted/Running/
Stopping/Sealed的适用子集；C不获得Armed/恢复能力。
未创建任何本会话OS资源的准入失败可Refused；一旦有资源活动，失败先
封闭闸门并进入Stopping，最终Aborted，即使T未恢复也不能当作无资源拒绝。
故障摘要可封存，但State=Aborted不能改成成功Sealed；Sealed也非Acceptance。

| 触发/消息 | 本地动作及引用 | 证据边界 |
| --- | --- | --- |
| 本地取消/截止/peer lost/协议错 |立即锁存首因、封闭创建/赋属/恢复，再用预持能力停止|通知不是停止本身|
| Stop（仅既有合法方向） |链路Bound后可在未终止阶段接收；首次记录该方向Stop序号；进入Stopping|对端原因只是声明，不能覆盖本地更早首因|
| StopAck |仅对应本地曾发的Stop且引用相等；记录一次|收到不等于进程退出或J空|
| Fault |仅Bound链路；沿既有Cause/Step编码锁存并收束|不接受C伪造能力/清理成功|
| inbox自身失败/EOF |冻结正常Feed/Take；停止/查询走已拥有能力|不能新建inbox继续收ACK或清除首因|
| 双方同时停止 |每个合法发送方向最多一个Stop；允许互为对应ACK|ACK缺失不能阻塞独立停止|
| D到达/观察失败 |正向判断停止；已有未知保持unknown|不在D后另开窗口补证|

未Bound端点不因报错而取得可信发Fault/Stop的资格；初始化失败仅走
已拥有资源本地收束。对已终止本地会话不重新解释新消息恢复运行。
重复/不匹配StopAck是协议异常，记录有限后续原因但不覆盖首因或重新启动。
发送Stop/ACK失败不重发，不用消息确认替代实际停止/查询。

内部原因沿既有13种Cause映射：绑定错binding，非法阶段/引用protocol，
接收/摘要满capacity，对端丢失peer-lost，时钟异常clock，到界deadline，
创建create，资源不符ownership，恢复不明release-unknown，适配器adapter，
内部internal，取消cancel，不支持unsupported；不扩展wire枚举。
首因关联不到已收到帧时，内部Sequence可null，Stop的FirstFailureSequence
用现有合法0表示无关联帧，不伪造一个消息序号或成功观察。

故障摘要仍L/W各8槽、总16条，每条含编码后字段最多1024字节。
第0槽固定首因，余槽填满即overflow并停止，不覆盖首因，不写旁路无限日志。
每个固定subject只保留必要的最新声明/本地观察及unknown，不追加消息历史。
不可将LatestObserved覆盖成“全过程从未失败”；首次失败/尝试状态单调保留。
精确外部持久化格式留给P3，本稿不新增archive/schema可运行接口。

## 8. 时基与独立停止不从消息借权

复用原T0/Frequency/D与P2-A保守边界：总120000ms包括启动至交接；
StopAt沿现有113500ms及6500ms保留公式取更早值，now>=StopAt禁止释放，
now>=D停止正向判定。Ready/Arm/ACK/心跳均不重新计时。
固定100ms发送/检查与500ms失联线索不变，不新增deadline外drain/cleanup预算。

每个本地对端watch从受控初始化安装时开始，不等待成功握手才监视。
只有新序号、声明绑定正确、角色/当前阶段合法的实际本地接收才刷新；
初始化Hello/Bound可以作为合法进展，未Bound心跳则是协议故障。
重复、非法帧、定时读取同一心跳槽及对端ObservedTick均不刷新。
本地真实时钟由未来获准适配器提供，不接受调用者tick声明冒充真实采样；
无法核对、倒退或溢出即关闭闸门。

独立计时/停止路径不能被消费解析、写入等待、C读取或finally阻塞。
停止只作用于已持J及确切引导资源，不按PID/名称重找。
observed-empty沿上层计数0、枚举空且未截断、确切成员退出及查询身份完整
的合取条件；请求成功/ACK/exit0/Job信号或无通知都不足。
L/W双失效仅保留原预设策略的有限依赖，CleanupStatus=unknown，不发布成功。
OS调度/系统失效不提供硬实时保证，超限须如实记录。

## 9. 后续验证地图与实施隔离（本轮不执行）

| 类别 | 必须覆盖的独立反例/预期 |
| --- | --- |
| 层间衔接 |普通1/心跳2/3/普通4合法；普通游标不重判连续；外部复制Frame不能取得交付关联|
| 握手顺序 |缺Hello/Bound/Ready、重复/逆向消息、错会话、跨方向序号不能比较大小|
| 可信前提 |Checks/ReadyBits=7但无实际资源/P2-C绑定仍拒绝；无布尔升级/序列化票据恢复|
| 一次性 |双Arm、双Commit、消费后检查失败/返回不明/ACK丢失均不再次尝试|
| 竞争 |停止先发生不尝试；恢复先发生则如实记录再停止；不承诺跨进程原子事务|
| 时间 |StopAt/D/500ms等值到界、原点不重置、旧心跳不续期、无真实采样时拒绝|
| 故障 |inbox停止后不读新ACK/不换token，首因不覆盖，摘要满/缺观察保持unknown|
| 所有权 |各角色资源不互换，T无协议，C无J控制/恢复权，普通快照不含能力|
| 证据 |尝试/声明/实际观察分开，unknown退出码为null，StopAck不能生成observed-empty|

预期必须是独立固定表/字节/引用和明确拒绝结果，不由被测逻辑动态生成。
内部普通数据、真实能力、调度/原生失败应分层验证；不能用synthetic
通过结果伪装实际能力成立。本稿不授权向当前产品注入setter、故障开关或通用callback。

书面批准本稿后只可编写合同落实计划，区分现有数据衔接可做的有限离线
拒绝验证与仍需安全/能力前提的真实阶段。不得另造公共事件回放状态机或
WouldRelease产品来代替后者；未有实际能力验证与P2-C准备绑定，正向出票/
真实恢复阶段继续阻塞。具体新代码/文件/API/次数/时限/顺序/失败即停/
审查条件须在后续计划中冻结并另行批准；旧耗尽/冻结额度不能借用。

## 10. 自检、保存及下一道批准

自检：本稿只细化既有消息与角色合同，不改wire/codec/inbox；入队准入与
普通消费分离，心跳不伪造连续证明；Expected不由自身帧反抄以自证；
状态报告不等于能力；唯一票据、一次尝试、失败unknown及资源停止分开；
首因/有限缓冲/原时基不变；未引入第三套可运行模拟器或live开关。
本轮不选择新Windows API，不声称本机支持、当前安全证据或owner/vendor放行。

仅新增本稿并更新current-status/repository-recovery，普通提交推送现有
codex/c12-b01-task9-coordinator；保留所有源代码、测试、审查、原始记录及预算。
换机恢复文档/源码不恢复能力、CLR身份、token、依赖/凭据或运行权限。
请用户审阅本书面设计；批准后仅进入实施计划编写，不直接实现或运行。
