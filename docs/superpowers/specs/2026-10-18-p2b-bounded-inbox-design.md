# P2-B 有界控制帧接收层详细设计（待书面批准）

日期：2026-10-18。基线提交78aa0f2c967e2c5035e6dcb376757df4699160d9。
状态：WRITTEN_DESIGN_PENDING_APPROVAL。
DesignApproved=false；ImplementationApproved=false；RuntimeApproved=false。

用户已同意编写这一范围的详细设计：固定容量接收层、分片/截断、
重复帧及心跳覆盖，不接真实管道/进程/启动权限。本稿仅文档和自检，
0测试/编译/新审查/能力探测；没有新的执行预算或可运行包。

## 1. 意图、依据与选型

目标是给已验收纯codec补上有限的字节组帧和接收排队合同，
使未来状态层能区分“字节收到”“帧已入队”和“握手语义获准”，
不创建第三套完整生命周期模拟器，不实现任何真实控制能力。

继承[会话设计§6–11](2026-10-18-p2b-owned-session-design.md)，
该文档SHA256为41BD5F7225D65D7BAE903E693946A2595C5036A2408B16BECE37965C622E36CE。
[纯codec限定批准](../../roadmap/2026-10-18-p2b-codec-limited-acceptance-approval.md)
仅覆盖指定PS7向量；依赖固定产品9A532E17，不能扩大其验收范围。

选定先做单线程、无I/O的有界接收内核；另外两个方案是先做完整握手
状态机（状态/能力混合过早），或直接接真实管道（需要独立环境/安全前提）。
本单元可单独拒绝损坏数据，不提供PrepareRelease/CommitRelease接口。
未来真实适配不能只加live开关就被视为已获批。

## 2. 固定拓扑与容量

| EndpointId Byte | Sender→Receiver Byte | 普通FIFO | 最近心跳 | 组帧 |
| --- | --- | --- | --- | --- |
| 1 | L1→W2 | 4×256字节 | 1×256字节 | 1×256字节 |
| 2 | W2→L1 | 4×256字节 | 1×256字节 | 1×256字节 |
| 3 | W2→C3 | 4×256字节 | 1×256字节 | 1×256字节 |
| 4 | C3→W2 | 4×256字节 | 1×256字节 | 1×256字节 |

普通队列全局最多16帧/4096字节，不借槽、不丢旧帧、不使用旁路队列。
心跳4槽/1024字节和组帧4缓冲/1024字节单独计量；
稳态留存原始帧/片段容量合计6144字节，不包括固定元数据、对象开销、
调用者输入、返回快照及单次decode暂存，不能称作CLR总内存硬上限。
每调用最多处理一个完整帧，不留decoded Message或无界历史。

未来每物理端点最多一个在途读/写，256字节发送缓冲由后续适配器负责；
本单元无读写任务，也不证明真实并发约束。既有L/W各8条、
每条1024字节故障摘要仍属未来监督层；本单元只留一条固定失败码/端点/
阶段/可选序号摘要，不创建第二套16条日志。

## 3. 状态所有权与逻辑接口

这是接口合同，不是本轮新增代码。实现计划须冻结具体返回形状和类型，
但不得改变本稿语义或容量。本单元只接受稳定、精确被动数据；
不接受外部Stream、回调、脚本块、任意方法对象或“Verified=true”。

- CreateInbox(Binding)：固定Schema Int32=1、BindingDigest byte[32]、
  SessionId byte[16]、OriginTick/Frequency/DeadlineTick Int64；
  校验沿用codec声明关系。四方向固定，NextSequence均UInt64=1，
  LastNonHeartbeatSequence均UInt64=0，部分长度/队列均空。
  不接受恢复序号、序列化旧状态或外来能力。
- FeedFragment(Inbox, EndpointId Byte, Bytes byte[])：Bytes精确一维、
  无附加metadata，长度1..(256-当前部分长度)。独立复制，不截掉多余字节。
  超量/空输入拒绝并锁存失败，不接受一块含多帧或自动拆分旁路。
- EndInput(Inbox, EndpointId Byte)：表示调用者声明EOF；不是实际I/O证据。
- TakeOrdinary(Inbox, EndpointId Byte)：取该方向FIFO头的独立帧快照；
  不重新绑定或推进序号；空队列正常返回empty，不忙等。
- InspectInbox(Inbox)：返回有界独立诊断快照，包括各长度/计数/
  序号、最近心跳及首因；不返回内部数组、句柄或可执行动作。

Inbox为受实现所有的进程内状态，非可导入恢复格式；入口须识别本单元
创建的状态，不把同形PSCustomObject当作状态。该身份机制是普通API
误用防护，不抵抗同进程私有反射。不同Inbox可用同一声明Binding，
其结果互相独立；不宣称跨实例/进程/重启防重放。
本单元按串行调用合同工作；并发调用不在支持范围。
Caller修改输入或返回的空/非空数组、scalar metadata不得污染状态、
后续调用或其他实例；不得通过共享空数组绕过所有权。

## 4. 分片、入队和序号提交

每次Feed按以下固定顺序处理：

1. 检查状态身份、全局失败锁存，再检查端点及Bytes精确类型/被动性/长度；
   非法输入不复制字节，触发首因invalid-input。未知状态仅拒绝，
   无有效状态时无法在其上锁存；不得触发输入getter。
2. 将整个合法片段独立复制到该端点固定缓冲。未满256字节返回partial，
   尚无消息、序号不变；保留精确部分长度。
3. 满256后调用固定codec ConvertFrom-P2BControlFrame；失败则锁存
   invalid-frame或internal-error，不搜索magic、不滑窗重同步。
4. 核对物理方向、digest/session和Origin/Frequency/Deadline等于冻结Binding；
   错为binding-mismatch。NextSequence必须精确相等，重复/旧帧/跳号均
   sequence-mismatch；UInt64.MaxValue即sequence-exhausted，不wrap。
5. Heartbeat的LastNonHeartbeatSequence须等于该方向内部已提交的最近
   非心跳序号，否则reference-mismatch。普通Arm/StopAck等跨方向引用及
   生命周期合法性不在本层作最终判定。
6. 非心跳检查该方向Count<4，否则capacity；心跳不占普通槽。
7. 同一串行提交中：复制帧入普通FIFO或替换最近心跳，推进NextSequence，
   非心跳更新LastNonHeartbeatSequence；清除组帧已用数据/长度后返回queued
   或heartbeat-updated。任何检查失败都不入队、不推进序号、不覆盖旧心跳。

完整但被拒的帧可保留在唯一组帧缓冲作有界诊断，长度256；
不另存第二份错误载荷。此前成功片段已经接收，失败不能伪称整条流未收。
每次返回AcceptedBytes为本次已复制片段长度或0，不代表完整消息获准。
内部提交失败锁存internal-error、禁止继续交付，不声称故障原子回滚。

注意：不能把Test-P2BControlBinding直接当接收检查器使用。
该函数还核对Arm/Stop等依赖握手状态的Expected引用；本层没有这些状态。
本层只复用decode并实现上述接收方向/声明绑定/连续序号规则，
后续状态层负责其余引用和释放语义，不以伪造Expected绕过检查。

## 5. 心跳和上层交付边界

全部帧包括心跳共享本方向连续序号；心跳覆盖只减少留存历史，
不能使后续普通帧等待已被覆盖的心跳序号。普通FIFO保持自身接收顺序；
最近心跳与FIFO是两个视图，不宣称合成完整全序消息日志。
示例：普通seq1、心跳seq2、心跳seq3、普通seq4：
普通FIFO为1/4，心跳槽为3，NextSequence为5。

本层不读取真实时钟、不更新真实“最近有效心跳时间”、不决定失联，
也不将Inspect重复读取视为新心跳。heartbeat-updated仅是新合法帧事件，
接收时间/100ms发送/500ms失联判定由后续真实适配和状态层独立处理。
未来状态层不能凭心跳槽重放Arm、建立Bound或证明C读通路就绪。
重复普通帧即使原帧仍在FIFO，或已Take，均拒绝；Take不撤销接收提交。

## 6. EOF、故障锁存和有限诊断

全局首个失败一经锁存，四端点停止正向接收与Take；
Inspect仍可返回独立诊断，队列/心跳/部分数据留在原固定容量内。
后续操作返回stopped，不清首因、不恢复队列、不自动创建新Inbox。

EndInput遇部分长度1..255时为truncated-frame；对齐无部分时为peer-closed；
本层没有授权的正常会话结束状态，因此二者均锁存失败。
不把空片段当EOF、不以队列empty或对端exit暗示EOF，不补零拼帧。
失败后EOF也只返回stopped。Take非法端点触发invalid-input；
合法端点空队列为empty且不停止。Inspect不消费普通帧或清空心跳。

失败码固定：invalid-input、invalid-frame、binding-mismatch、
sequence-mismatch、sequence-exhausted、reference-mismatch、capacity、
truncated-frame、peer-closed、internal-error；stopped是后续操作状态，
不覆盖FirstFailure。失败不发送Stop或终止进程，只供未来拥有能力层处置。

每个操作结果的Mode固定p2b-inbox-only，SourceVerified/
LiveActionsPerformed/RuntimeApproved固定false。queued/partial/empty/
heartbeat-updated均不等于真实端点认证、握手许可、发布或清理已证。
此内部结果不冒充codec原10字段根，不修改其公开签名/Reason或wire格式。

## 7. 后续验证地图（本轮不执行）

| 类别 | 必须独立验证的精确边界 |
| --- | --- |
| 组帧 | 256一次、1+255、255+1、256个1；部分后超剩余整块拒绝；0/257拒绝；多帧块不截取 |
| 方向绑定 | 四固定方向，错误endpoint/role/digest/session/时基；合法codec帧不能跨端点 |
| 序号 | 从1开始，心跳/普通混排，队列内/Take后重复，旧帧/跳号，UInt64高位和MaxValue不wrap |
| 容量 | 每方向4个成功第5个失败；其他方向空也不借槽；四方向共16帧；心跳独立覆盖 |
| 心跳 | 上述1/2/3/4例、错误LastNonHeartbeat、重复不覆盖、不推进、不产生新事件 |
| EOF与故障 | 长度0/1/255的EOF，首因保持，错误完整帧不重同步，失败后四方向不可Take/Feed |
| 所有权 | 空/非空结果、组帧、出队、Inspect、Binding及多实例交叉修改和ETS装饰不可串扰 |
| 被动入口 | getter/描述符子类/隐藏metadata及Length/Rank欺骗，类型精确、无隐式转换，拒绝时不调用输入代码 |
| 有限留存 | 满队列+心跳+部分缓冲；覆盖/Take不遗留无界列表或decoded消息历史 |
| 上层边界 | Arm/Stop引用不被本层声称验证；queued不形成状态迁移、能力或真实清理 |

预期必须是独立literal和明确结果快照，不由被测产品动态生成。
计划须说明如何覆盖高序号边界而不新增生产恢复序号/故障注入入口；
若无法在单元测试层隔离验证，应明确限制并提出新设计，不偷偷添加绕过。
产品输入结构检查不能仅靠自报MemberType或公共成员视图，需避免重开
已修复的冷缓存/隐藏描述符漏洞；具体复用边界由实施计划冻结审查。

## 8. 实施分解、批准门槛和自检

建议一个独立可测试交付：有界inbox内核及其测试；
不同时实现真实适配、监督状态机、发送队列、完整故障摘要或P2-C/P3。
实施计划再明确文件图、具体类型/完整结果形状、固定引擎与依赖身份、
有限RED/GREEN/针对性负对照/还原验证、次数/顺序/每轮上限、
失败即停及一次条件只读审查；本稿不授予任何次数或新审查席位。
所有旧预算原样冻结，已验收codec不可原地修改或重跑以作warmup。

自检完成：16普通槽与4心跳槽/4组帧缓冲分别计量；序号提交与出队分离；
跨方向引用未假定可由接收层验证；EOF统一故障避免伪正常结束；
全局首因和诊断保留有界；原codec验收范围/RootChild cap32/真实120000ms
合同均不变；没有TODO占位、真实会话身份或可执行代码。

本轮只新增本稿并更新进度/恢复文档，普通提交推送现有开发分支。
换机只恢复文档/源码，不恢复原始记录、能力、凭据或执行权限。
请用户审阅书面设计；明确批准后才编写实施计划，计划仍须另行批准、
选择执行方式和有限预算。未批准前不写实现或运行验证。
