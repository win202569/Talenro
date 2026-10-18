# P2-B Bounded Control Frame Inbox Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended Native execution, or superpowers:subagent-driven-development only after a separately approved replacement budget. Steps use checkbox syntax.

**Goal:** 实现无I/O、串行、有界的控制帧组帧/接收队列，不产生会话能力或启动许可。
**Architecture:** 私有弱键状态表识别本模块创建的opaque token；每实例四方向固定缓冲/FIFO/心跳槽。复用固定codec decode及被动检查，接收序号在提交时推进，出队不改变序号。
**Tech Stack:** 普通Windows PS7.6.5 Core x64；PowerShell模块、CLR固定数组/ConditionalWeakTable；无产品编译、原生调用、新依赖或真实I/O。
**Spec:** [已批准书面设计](../specs/2026-10-18-p2b-bounded-inbox-design.md)，物理SHA256=43036F17E712CA2AA57F51CA09BAFF874DB53F954FADF82CBCD6388763B2585B。

用户在设计提交d7e45e1ea889f89193014a41f7cb0d5c9350090a后明确回答“是”，
批准该书面设计并进入计划编写。DesignApproved=true；
ImplementationApproved=false；RuntimeApproved=false。
原设计待批准标题留作历史。本计划、方法及新5+1额度均待另行批准。
本轮只创建计划并更新进度/恢复，无实现、目标宿主、编译或新审查。

## Global Constraints

- 复用现有worktree及codex/c12-b01-task9-coordinator；无main/PR/强推/环境或安全设置变更。
- 四端点固定1:L→W、2:W→L、3:W→C、4:C→W；每方向4普通槽、1心跳槽、1组帧缓冲，均256字节。
- 普通16帧4096字节、心跳1024、组帧1024；稳态原字节容量6144，不是CLR总内存硬上限。
- 无读时钟/管道/Process/Job/发送/取消/线程/能力或完整握手状态机，不调用外部callback。
- 串行、稳定输入合同；身份防误用不是对抗私有反射、敌对module/runspace或跨重启防重放。
- SourceVerified/LiveActionsPerformed/RuntimeApproved每结果固定false；queued不是许可。
- 既有codec9A532E17、所有历史源码/报告/回执及已用/冻结预算不变；不得把旧完整套件当warmup重跑。
- 新意外失败立即停；不补测/重试/换测试/加审查；变异必须立即恢复封存good身份。
- 测试专用原有CounterNote C# fixture仅在两个完整suite宿主内各编译一次，无产品Add-Type或原生编译。

## Review Focus

1. 心跳覆盖与普通FIFO序号交错：入队顺序消费序号，Take不回退；HeartbeatMixed/SequenceAfterTake固定反例。
2. 共享空数组与opaque token泄漏状态：EmptyInbox/OutputOwnership/MultiInstance检查空及非空ETS反向污染和原引用。
3. 入队满/EOF后错误继续：EndpointCapacity/GlobalCapacity/FirstFailure/StoppedDelivery检查全局首因与禁止出队。
4. 隐藏/子类/冷缓存入口：InputPassivity用独立可观察CounterNote，绝不靠公共Members/自报MemberType。
5. 高频/最高序号只可用测试内私有状态构造：SequenceHigh/SequenceExhaustion明确是合成边界，不增加生产恢复/注入入口。

---

## File map and dependencies

未来新目录docs/diagnostics/isolated-fuzz/p2b-bounded-inbox/，本轮不创建：

| 文件 | 职责 |
| --- | --- |
| inbox.psm1 | 五public操作、私有弱键registry、固定状态与快照helper |
| inbox.tests.ps1 | Group all/capacity/sequence；32组独立断言 |
| test-helpers.ps1 | typed literal读取、完整结果比较、受控测试内私有计数器设置 |
| fixtures.json | 完整256字节hex和独立声明；数字以十进制string保存后显式解析 |
| descriptor-fixture.cs | 原CounterNote字节相同副本，SHA87882565265A40D01777C467FBFB2E40DAEDC8F270D47F896D390BD8CF4E09FE |
| expected-output.json | 有序完整输出及两项指定失败，不从产品输出学习 |
| target.ps1 | 固定Group路由/UTF8/context，不额外导入或warmup |
| capture.ps1 | Prepare/SealProduct/Run、顺序预消费、独立有界原字节捕获 |
| README.md / .gitattributes | 范围/预算/恢复；psm1/ps1/json/cs/md均LF |

inbox.psm1仅从固定相对路径../p2b-cold-cache-repair/control-frame.ps1
加载依赖至自身模块作用域；先核完整SHA
9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C。
禁止环境自动发现/任意path参数/影子命令替换。该模块导入只初始化私有
ConditionalWeakTable<object,object>，无I/O能力。模块remove/reimport后旧token失效。
允许复用New-P2BCMetadata、Assert-P2BCPlain、Get-P2BCByteSnapshot、
Copy-P2BCScalar、Read-P2BCUInt、Get-P2BCDeclaredBounds和decode。
Binding的六字段根验证在新helper内按固定root raw descriptor规则实现，
不改变旧Get-P2BCSnapshot形状，不假造12/13字段输入绕过检查。
所有内部helper非导出；仅下面五个public入口。

## Exact interfaces and shapes

五public函数参数全部[object]防止入口CLR强转：
New-P2BInbox(Binding)、Add-P2BInboxFragment(Inbox,EndpointId,Bytes)、
Close-P2BInboxInput(Inbox,EndpointId)、Receive-P2BInboxFrame(Inbox,EndpointId)、
Get-P2BInboxSnapshot(Inbox)。均返回单个精确PSCustomObject结果。

Binding精确六NoteProperty，大小写敏感，无额外/隐藏/子类/可执行成员：
Schema Int32=1；BindingDigest一维byte[32]；SessionId byte[16]；
OriginTick/Frequency/DeadlineTick均Int64，沿用codec Get-P2BCDeclaredBounds。
EndpointId底层精确Byte=1..4；不接受Int32/string/null或装饰值。

结果根固定12字段：
Schema Int32=1、Mode String=p2b-inbox-only、Accepted Boolean、Reason String、
Inbox object或null、EndpointId Byte或null、AcceptedBytes Int32、
Frame byte[256]或null、Snapshot PSCustomObject或null、
SourceVerified/LiveActionsPerformed/RuntimeApproved Boolean=false。
Inbox仅New成功时非null，为原生System.Object opaque token，不是能力/快照；
其他结果不回传token。New失败无token，也没有可锁存状态。
Token持有人只能引用身份，不能传同形对象、克隆/序列化物重建状态。
私有registry以原生token引用为键，查找不读其ETS属性/方法；
不将PSObject包装引用误当相等base或生成第二个token。
对token的ETS装饰不得改变内部状态；它不是声明数据，不通过其metadata解析状态。

Reason成功created/partial/queued/heartbeat-updated/empty/dequeued/inspected。
上述Accepted=true；所有拒绝/stopped为false。AcceptedBytes仅Feed复制时
为本次长度（即使完整帧随后被拒），其他操作及复制前拒绝为0。
EndpointId仅通过被动Byte准入后非null，否则null。Frame仅dequeued非null；
Snapshot仅inspected非null。失败首因从Snapshot读，不在其他根夹带内部状态。

Snapshot精确4字段：Stopped Boolean、FirstFailure PSCustomObject或null、
Binding（独立六字段）、Endpoints（精确object[4]）。
FirstFailure固定Reason String、EndpointId Byte或null、Stage String、
Sequence UInt64或null。Stage仅input/frame/binding/sequence/reference/
capacity/eof/commit；Sequence仅已成功decode后可用，否则null。
新建/未知token的错误不制造伪快照；已停止有效token仍允许Inspect。
Endpoint快照固定8字段：EndpointId Byte、Count Int32、
NextSequence UInt64、LastNonHeartbeatSequence UInt64、PartialBytes byte[0..256]、
Queue object[0..4]（各元素独立byte[256]）、Heartbeat byte[256]或null、
PartialLength Int32。数组空时也必须new，不使用共享Array.Empty/ToArray空结果。

私有每方向四预分配byte[256]槽、head/count、一个heartbeat及present位、
一个partial及length、两UInt64。Take复制head后清槽/推进head/count，
Inspect只复制实际有效字节/帧。registry不留全局强引用token列表，
状态不得含decoded Message或每次结果历史。

故障按设计§4顺序锁存；已停止token除Inspect外直接stopped，不再读
EndpointId/Bytes（测试含getter毒物）。任何异常提交失败锁存internal-error/
commit，不保证回滚。未知token只invalid-input，不改变其他实例。
入参scalar和输出scalar的独立所有权沿用固定copy guard，token身份是唯一例外。

## Task 1: Inbox kernel and five bounded verification hosts

**Consumes:** 固定设计/codec/typed Binding/literal帧/稳定片段。
**Produces:** 上述五入口与固定形状；不含真实适配或握手状态迁移。

- [ ] **Step 1 — freeze independent fixtures, assertions and capture inputs before implementation.**
  准备目录中除inbox.psm1外全部文件；RED时模块必须不存在。
  fixtures.json以冻结12种codec FrameHex为字节基准，测试独立按offset
  逐字节写Sender/Receiver/Sequence/Heartbeat引用和时基，不调用产品encoder。
  常用Binding取digest00..1F/sessionA0..AF、Origin0/F1000/D120000；
  Hello为Kind1/payload16/Profile1/FrameBytes256，Observed4。四方向逐一独立literal。
  HB Kind8/payload16、序号和最近非心跳引用使用明确常量。

- [ ] **Step 2 — run RED once.**
  capture.ps1 -Action Run -Slot red。context/host后精确
  ASSERTION FAILED: P2BInboxApiRequired，exit1/3行/空stderr。
  RED不得import codec或编译fixture。模块已存在或任何其他结果停止。

- [ ] **Step 3 — implement inbox.psm1.**
  按上文接口实现registry/准入/固定状态、Feed/EOF/Take/Inspect。
  保持接收提交与语义引用分离；不调用Test-P2BControlBinding充当接收门槛。
  静态parse不执行/import；核fixture字节独立布局/32组oracle后SealProduct。
  公共Feed中普通容量条件唯一规范写为
  if($endpoint.Count -ge 4){...capacity...}，序号条件唯一写为
  if($message.Sequence -ne $endpoint.NextSequence){...sequence-mismatch...}，
  便于指定单处变异。内部NextSequence以BigInteger加1且显式UInt64返回。

- [ ] **Step 4 — run full GREEN once.**
  capture -Action Run -Slot full，Group all，exit0/32PASS/35行精确。
  任一失败冻结后续三个slot和条件审查，不在余slot修复。

- [ ] **Step 5 — capacity mutant once and immediate restore.**
  仅将上述唯一Count -ge 4改-ge 5，Group capacity仅ApiRequired/
  EndpointCapacity。独立第五帧seq5应被拒；mutant错误接受必须
  ASSERTION FAILED: P2BInboxEndpointCapacity，exit1/4行。
  私有物理4槽可ring覆盖但不得先数组越界；测试先断言第5结果capacity，
  internal-error/任意异常不算命中。立即apply_patch恢复good字节及SHA。
  若实际实现无法保持该特定变异可观察，实施前停下修订计划，不换成其他失败。

- [ ] **Step 6 — sequence mutant once and immediate restore.**
  唯一-ne改-lt，Group sequence仅ApiRequired/SequenceExact。
  新实例Next1时独立literal seq2应sequence-mismatch；
  mutant错误接受必须ASSERTION FAILED: P2BInboxSequenceExact，
  exit1/4行；其他失败无效。立即恢复good并核SHA。

- [ ] **Step 7 — restored full GREEN once.**
  同good产品/32组/oracle，exit0/35精确行。无额外原codec33组或全仓库运行。
  确认两完整宿主仅各一次fixture managed编译；counter正向可观察、产品路径0。

### Frozen ordered 32 test groups

前缀P2BInbox，成功每组一条PASS。组名/关键独立断言如下；
test-helpers比较完整字段/类型/字节及操作前后状态，不能只检查Accepted。

| Group | 必须断言 |
| --- | --- |
| ApiRequired | 五入口恰好导出；没有恢复/注入/public helper |
| BindingShape | 六字段exact；缺/多/大小写/错类型/长度/时基拒绝，无token |
| BindingIsolation | 返回/原Binding反向修改不影响冻结绑定 |
| TokenIdentity | 同形、独立object、克隆/序列化token拒绝；本tokenETS不改变registry |
| EmptyInbox | 四Count0/Next1/Last0/partial空/HBnull；Take empty不停止 |
| WholeFrame | 四方向独立256字节Hello seq1queued、Next2/Last1、出队相等 |
| SplitEdges | 1+255和255+1，partial时无消息/序号不变，合成后完整一次 |
| SingleByteChunks | 256个1字节，仅最后queued，exactBytes且总接受256 |
| ChunkBounds | 0/257/512、partial1后256整块拒绝/保留前1，不截取 |
| LiteralDirections | codec合法Hello错误端点拒绝；各方向角色固定 |
| BindingMismatch | digest/session/Origin/F/D各独立合法帧不符拒绝；不推进 |
| SequenceExact | Next1实际0由decode拒绝；实际2为sequence-mismatch；1成功 |
| SequenceAfterTake | 入队内和Take后重seq1均拒绝；Take不回退Next |
| SequenceHigh | 测试内构造Next=9223372036854775808，等值成功Next+1；相邻不等拒绝 |
| SequenceExhaustion | 私有Next=UInt64.MaxValue-1等值后NextMax，Max拒绝不wrap |
| EndpointCapacity | Hello seq1..4排满；同方向seq5capacity、Count4/Next5，其他端点不借槽 |
| GlobalCapacity | 四方向各4总16；填满四HB槽并各保留1字节partial，Inspect精确有界 |
| HeartbeatMixed | 普通1/HB2/HB3/普通4→FIFO1,4/HB3/Next5/Last4 |
| HeartbeatReference | HB Last错误、重复/旧/跳号均拒绝，不覆盖已有HB，不推进 |
| HeartbeatFullQueue | FIFO1..4满后HB5Last4仍成功/HB槽替换；普通6capacity |
| EofEmpty | 新实例/完整帧后EOF均peer-closed；empty不是EOF |
| EofPartial | 1/255字节后EOF均truncated-frame，保留exact前缀 |
| InvalidFrame | magic/reserved/padding/长度损坏满帧拒绝；不滑窗补救 |
| FirstFailure | capacity后坏frame/EOF保持最初端点/阶段/序号；invalid-input优先 |
| StoppedDelivery | 故障后四方向Feed/Take/Close stopped；毒物参数不读取；Inspect可用 |
| InputPassivity | Binding子类/隐藏/可执行描述符；Byte端点、片段可见/隐藏/子类及冷共享metadata拒绝，真实counter0 |
| InputOwnership | Binding/片段复制后原对象metadata/bytes突变不改变partial/queued |
| OutputOwnership | Take/Inspect空及非空数组/标量metadata变更可观察但不污染后续快照 |
| SnapshotBounds | 固定四端点、最多4Queue元素、256字节HB/partial；连续Take/覆盖无history |
| MultiInstance | 相同Binding两token独立；A失败或ETS污染不影响B |
| AuthorityBoundary | 全形状/flagsfalse；Arm/StopAck合法wire仅queued，不声称引用/释放已验证 |
| Complete | 组选择完整、counter0、无意外error；LIMIT only declarations/inbox, no runtime authority |

SequenceHigh/Exhaustion只允许test-helpers用模块私有scope持有registry，
对新fixture token的Next/Last字段作一次显式设置；设置前后记录字段，
不导出setter、不增参数/环境开关，不作为真实长序列接收证据。
InputPassivity以既有CounterNote/独立SDK原始表操作证明实际receiver/描述符
保留且counter正向可观察，不能把框架丢弃metadata当成产品拒绝；
只在all宿主编译fixture，不另开诊断host。测试内固定reflection不是产品依赖。
样例完整比较：普通1/HB2/HB3/普通4后Snapshot.Endpoints[0]要求
Count Int32=2、Next UInt64=5、Last UInt64=4、Queue逐字节为literal1/literal4、
Heartbeat为literal3、PartialLength Int32=0且新建空byte[]；其余端点初始值不变。

## Proposed finite allowance and controller — NOT APPROVED

推荐Native：同一代理实施上述耦合单元；恰好5个新普通宿主+1次条件独立只读审查。
每宿主最多120000ms执行+2000ms owned exit+2000ms并发独立stdout/stderr drain。
这是新测试宿主上限，不修改真实会话120000ms合同或旧60000ms预算。
固定engine C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，
SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
同目录SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
-NoProfile -NonInteractive -File target.ps1；PS7.6.5/Core/x64/adminFalse。

| Slot | Group | 精确预期 |
| --- | --- | --- |
| red | API absent | exit1/3行/ApiRequired assertion |
| full | all | exit0/32PASS/35行 |
| capacity | capacity | exit1/4行/EndpointCapacity assertion |
| sequence | sequence | exit1/4行/SequenceExact assertion |
| final | all/restored | exit0/32PASS/35行，与full相同oracle |

每完整输出先context/host、后32PASS、末LIMIT；失败无LIMIT。
只容许删除至多一个终止换行，禁止TrimEnd吞空行。
新packet .superpowers/sdd/2026-10-18-p2b-bounded-inbox/必须事先不存在；
Prepare继承旧1185条VerificationPath映射，不擅自改成OriginalPath，
另列先前packet全部普通文件以及17个portable叶文件、修复报告、
提案/批准/本设计、获批本计划快照、新模块之外输入、engine/SDK完整身份；
逐条保存role/path/hash及总数，具体数量由实施前明确列出的完整清单机械计算，
不能省略或动态放宽。当前status/recovery不作为可变历史pin，历史快照若已有映射仍按其VerificationPath核对。
Prepare前任何旧映射缺失/漂移立即停，不自动补建/重跑。新模块RED后SealProduct；
good/两个mutant/恢复文件各保留源码快照和精确diff/SHA，不改旧输入。

控制器每次launch前持久预消费、固定顺序与前轮exact结果，核所有pins/
ordinary上下文/engine/SDK/product身份，分别并发保存原始字节及SHA；
只有当前创建的确切Process可停止，不扫描其他PID。超时、stderr非空、
流不完整、owned exit不明、输出不精确或漂移立即冻结余轮/审查；
回执保留原判定，禁止事后解析修正自动升级。无额外import/fixture预热目标。
完整suite输出预期在Prepare前独立写定，若测试不可观察/错预期也停止。
每次mutant立即恢复；失败清理或恢复不明按unfinished保存，不能继续next slot。

## Task 2: Conditional readonly review and durable handoff

**Consumes:** 相同good模块、全部测试/fixture/oracle/controller、五轮exact回执和冻结身份。
**Produces:** 一次限定代码/已存证据审查及必要devbranch检查点，不是正式验收。

- [ ] **Step 1 — dispatch one fresh readonly reviewer only after all five exact.**
  不执行/import/compile/复现/编辑/网络或安全查询/嵌套代理；
  审当前单元、固定codec复用边界、测试/controller/oracle和保存证据，
  其他历史只核hash。不要求whole-branch历史正确性。
- [ ] **Step 2 — disposition every finding and declined behavior.**
  Critical/Important/审查不完整立即停，不自动修复/补测/替换reviewer。
  保留完整输出、Minor和所有declined并逐项给出父代理结论。
  明确合成高序号、私有scope测试、性能/CLR内存/真实I/O等证据限制。
- [ ] **Step 3 — verify and save explicit necessary files.**
  添加portable-summary.json、完整review.md和执行报告
  docs/roadmap/2026-10-18-p2b-bounded-inbox-execution.md；
  更新本计划实际状态/current-status/recovery。原批准计划快照冻结，
  源/fixture/原始记录不事后改判。必要文件显式暂存/普通提交推送既有devbranch，
  核远端SHA与clean；失败可保存unfinished，禁止main/自动验收。
  clone只恢复portable源/fixture/记录，不恢复token/registry/原始packet/
  SDK/cache/凭据或预算。不要持久化opaque token冒充恢复能力。

## Self-review and handoff

设计§2固定容量→GlobalCapacity/SnapshotBounds；§3所有权→TokenIdentity/
OutputOwnership/MultiInstance；§4准入提交→Split/Binding/Sequence/
Capacity；§5心跳→Heartbeat三组；§6失败→EOF/FirstFailure/Stopped；
§7每类向量对应上表；§8真实适配/发送/状态机均明确排除。五条Review Focus
全部落在Task1具名测试；高序号只有合成测试scope，无生产入口。
32组=35行，RED3，mutant4；所有接口五名/12字段/快照类型统一。
计划不修改已批准设计或codec，0执行/新审查。计划批准后Native执行时先
复核隔离worktree，Prepare冻结后严格按5+1；额外实现代理/任务级审查必须
另定预算，不能塞入本路线。

请用户审阅本计划并选择方式。推荐Native；可选Subagent-driven但需重定额。
只有明确批准本计划、Native及新5+1有限预算后才创建源/fixture/controller并执行。
