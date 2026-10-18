# P2-B Handshake Seam Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended Native execution. Use superpowers:subagent-driven-development only with a separately approved replacement budget. Track steps with checkboxes.

**Goal:** 交付固定codec/inbox之间的12组离线衔接验证及可恢复证据，不实现会话状态机、票据或真实能力。
**Architecture:** 新测试夹具在同一普通宿主中调用已验收inbox及其固定codec依赖；预期为独立literal。不添加产品入口，不把测试辅助函数变成握手消费器。真实阶段列为明确阻塞项，不以模拟器填补。
**Tech Stack:** Windows PowerShell7.6.5 Core x64、固定SDK、现有byte[256] fixture，无Add-Type/原生编译/新依赖。
**Spec:** [书面合同](../specs/2026-10-19-p2b-handshake-contract-design.md)，SHA256=9C37A7AD0156DDB7732503E78F919FAA887B0708D3F91CFA5494C9BD367309CC。用户在f64d0ff7后明确批准书面稿并仅同意编写计划。

状态：PLAN_PENDING_APPROVAL。ImplementationApproved=false；RuntimeApproved=false。
本轮只写计划/自检；0新产品/夹具文件/packet/目标/测试/编译/审查/安全查询。
提议NEW2目标+条件1只读reviewer，未批准。此预算只适用于Native。
旧inbox5/5和1/1已用完，旧原2spent+3FROZEN及独立诊断1spent均不变。

## Global Constraints

- 复用codex/c12-b01-task9-coordinator隔离worktree；基线f64d0ff799a22390fbff9726c851d436015e7b7f。禁止main/发布/设置/依赖变更。
- 固定codec9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C与inbox47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC不得改写或复制成另一产品。
- 固定wire种类1..12、4方向、16普通槽/4心跳槽/4组帧缓冲、6144原始字节容量和旧cap32不变。
- 只验证已有数据接缝；不创建ConsumeOwnedControl/DeriveLocalExpectation/PrepareLocalRelease/CommitLocalRelease等生产入口、公共事件回放或WouldRelease产品。
- 不改变或重跑原codec33/inbox32/tree12套件；只读引用其固定fixture/helper，Minor2明确延期，不补测试或修改旧证据。
- 真实120000ms合同、100/500ms失联设计、StopAt和D不变；测试宿主护栏不是实际运行时基证明。
- SourceVerified/LiveActionsPerformed/RuntimeApproved保持false；不产生实际SessionId、nonce、运行目录、票据、句柄或进程能力。
- 无实际能力/P2-C准备绑定不得出票/发Arm/恢复；测试中的合法Arm数据不构成消息发送或许可。
- 所有预期/脚本/数据在首个目标前冻结；异常即停，不按输出改预期、自动修复/重跑/加审查者。
- 原始packet/失败/审查/耗尽与冻结额度保留。普通阶段提交按明确清单，不删除raw/worktree；换机不恢复执行许可。

## Review Focus

1. 把普通1/4当作缺号而拒绝：OrdinaryGap和StaleExpectedRejected必须同时证明inbox合法交付与错误全量checker游标拒绝，不能构造自证生产Expected。
2. “binding-matched”被当作能力：DeclaredExpectedIsNotAuthority逐项核false flags，并明确真实握手/消费来源尚未实现，测试辅助函数不导出。
3. fixture原引用不是1：ArmReference固定引用3、StopReference固定引用2；误用帧本身序号或默认1必须由独立反例检出。
4. 停止后为收ACK换token/重置准入：StoppedInboxNoRecovery检查原token首因/保留队列、Feed/Take均stopped；TakeDoesNotResetAdmission检查重复帧仍拒绝。
5. JSON整数精度或证据标签再漂移：只由PowerShell保留原JSON token；保存前比较嵌入数据与raw，不经JS Number再序列化，不把两轮夹具通过称为状态机/能力验证。

---

## 分期及书面合同覆盖账本

这是书面合同§9允许的第一份“现有数据衔接有限离线验证”计划，不是整个合同落实完毕。

| 合同范围 | 本计划覆盖 | 仍未实现/不得声称通过 |
| --- | --- | --- |
| §1–2层间边界 |FixedBoundary、DeclaredExpectedIsNotAuthority|新的生产握手消费器/资源所有者|
| §3–4序号/引用关联 |OrdinaryGap、TakeDoesNotResetAdmission、StaleExpectedRejected、3类引用|真实交付来源认证、发送者游标/资源表|
| §5消息准入声明 |固定Arm/ACK/StopAck/Heartbeat与codec拒绝反例|Hello/Bound/Ready本地阶段、重复阶段、对端实际绑定|
| §6一次性票据/竞争 |不新增票据或正向接口；NoAuthority声明|出票/双Commit/原生调用/取消竞争/崩溃窗口|
| §7inbox停止边界 |StoppedInboxNoRecovery、完整首因/队列快照|真实Stop/ACK发送、J停止/计数/成员观察|
| §8时基 |ArmDeclaredExpiry仅声明整数113499/113500|真实时钟、500ms看门狗、独立停止、硬实时|
| §9–10证据纪律 |冻结/两轮捕获/限定审查/无损便携结果|完整P2-B/P2-C/P3/Task9/fullC12或人工验收|

未覆盖部分是阻塞的后续独立阶段，不是可跳过的任务：真实能力验证、安全
owner/vendor及精确运行批准、P2-C准备绑定尚未成立。不得在此计划执行中
临时加模拟器、测试setter、fake Ready、live开关或通过传入布尔补齐。
如后续要落实这些部分，必须另定范围/计划/前提/有限预算；本计划没有授予它们。

## 文件图与固定输入

未来新目录docs/diagnostics/isolated-fuzz/p2b-handshake-seam-verification/，本轮不创建。

| 文件 | 唯一职责 |
| --- | --- |
| handshake-seam.tests.ps1 |12组受限夹具、私有断言/Expected构造、独立trace|
| expected-output.json |full/final同一15行输出及固定trace观测预期|
| capture.ps1 |Prepare/Run full或final，NEW2顺序预消费/分流/哈希/上限/停止|
| README.md / .gitattributes |范围/恢复限制与ps1/json/md物理LF|
| portable-summary.json / review.md |两轮之后的派生记录及唯一条件review，非冻结执行输入|

引用旧目录p2b-inbox-tree-copy-repair/而不复制产品：
inbox.psm1=47BABFED...；test-helpers.ps1=4F418D2BE07ACE9597CE7A90438E7325F692A3C578010580548299016A2A5E8D；
fixtures.json=5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D。
fixture/helper仍读固定旧路径；不得编辑其内容。测试只用IB-Binding/Frame/
Put/Result/Type/Eq/Initial/Expected/Failure等普通数据辅助函数，不调用
IB-PrivateSequence、IB-InitDescriptors/Decorate/Cold或任何registry写入。

## Task1: 冻结并交付12组衔接验证，执行两轮相同夹具

**Files:** 新目录前5项；旧产品/helper/fixtures只读。
**Consumes:** 原New/Add/Close/Receive/Snapshot五入口及固定codec普通数据合同。
**Produces:** 新handshake-seam.tests.ps1脚本、15行独立oracle、full/final两条精确回执；无新产品API。

测试内部仅允许以下局部辅助函数，全部留在脚本scope且不导出：
HS-Need([bool]Condition)失败抛ASSERTION FAILED: P2BHSeam+当前Group；
HS-Case([string]Name,[scriptblock]Body)顺序记录/输出PASS；
HS-Codec([ValidateSet('Decode','Check')][string]Action,[object]Value,[object]Expected=$null)
在已载入固定module scope调用ConvertFrom-P2BControlFrame或Test-P2BControlBinding；
HS-Expected([byte]Sender,[byte]Receiver,[uint64]Next,[uint64]Arm=0,[uint64]Stop=0,[uint64]Last=0)
只构造下述typed literal；HS-AssertCodec([object]Result,[string]Reason,[object]Next=$null)
独立核完整10字段codec根、Reason/Accepted/Next及false flags。
这些是夹具调用便利，不是DeriveLocalExpectation或可信交付凭证。

HS-Expected固定字段及顺序：
Schema Int32=1,Sender/Receiver Byte,BindingDigest byte[32]=00..1F,
SessionId byte[16]=A0..AF,NextSequence UInt64=Next,
OriginTick Int64=0,Frequency Int64=1000,DeadlineTick Int64=120000,
StopAtTick Int64=113500,ArmSequence/StopSequence/LastNonHeartbeatSequence UInt64为参数。
不从被测Message复制这些值。IB-Frame从固定hex独立改写方向/序号/声明时基；
Kind5的payload引用保持3、Kind10保持2，不能假设为1。

- [ ] **Step1 — 先写完整夹具/预期，静态检查后冻结，产品不动。**
  只Import固定inbox一次；固定codec定义由其已有依赖提供，不另加预热。
  新trace字段固定Schema1/Mode=p2b-handshake-seam-tests/Groups string[]/
  Observations固定观测对象/Passed bool/Error null或固定异常摘要/
  CompileCount Int32=0/ProductChanged=false/RealGateImplemented=false/
  SourceVerified=false/LiveActionsPerformed=false/RuntimeApproved=false。
  所有结果在序列化前核typed值；控制器核完整Group列表及各观测literal。
  在Prepare前完成下列12组，所有预期写进expected-output.json，禁止学习输出。

  **1 FixedBoundary**：五导出恰好原五入口，固定module/helper/fixture哈希；
  HS-Need (ExportsExact -and PinsExact)。不得出现新的生产消费/出票入口。

  **2 OrdinaryGap**：新token，依次入队端点1的Hello seq1、
  Heartbeat seq2/Last1、Heartbeat seq3/Last1、Bound seq4。四次Reason依次
  queued/heartbeat-updated/heartbeat-updated/queued。
  HS-Need ((IB-Eq SnapshotQueue ([object[]]@($f1,$f4))) -and Count -eq 2
  -and AdmissionNext -eq 5 -and LastNonHeartbeat -eq 4 -and HeartbeatSequence -eq 3)。
  两次Take按f1/f4精确字节，第三次empty；完整预期Snapshot仍Next5/Last4，
  Queue新空数组，Heartbeat=f3，Partial空。这只是声明传输，不证明Bound阶段合法。

  **3 TakeDoesNotResetAdmission**：另用同一固定1/HB2/HB3/4向量并Take两帧，
  再Feed f4必须sequence-mismatch/AcceptedBytes256；首因Sequence UInt64=4、
  Endpoint Byte1/Stage sequence；Next5/Last4不变，拒绝f4保留Partial256。
  HS-Need (-not Duplicate.Accepted -and Duplicate.Reason -ceq 'sequence-mismatch')。

  **4 StaleExpectedRejected**：对已由该向量Take的f4调用Decode，然后Check，
  Expected=HS-Expected 1 2 2。完整结果sequence-mismatch/Next=null。
  HS-AssertCodec $result 'sequence-mismatch'。不能由此否认组2的inbox合法交付。

  **5 DeclaredExpectedIsNotAuthority**：相同固定f4，独立Expected.Next=4，
  Check必须binding-matched/UInt64 Next5，但三authority flags均false。
  HS-AssertCodec $result 'binding-matched' ([uint64]5)。
  这是已知fixture的纯声明checker控制，不是从本帧反抄游标的生产方案。

  **6 ArmReference**：IB-Frame 1 7 5保留ArmAck payload引用3；
  Decode后Expected.Next7/Arm3应binding-matched/Next8；
  Arm0与Arm2分别reference-mismatch/Next=null。
  HS-AssertCodec $good 'binding-matched' ([uint64]8)，两反例各核完整拒绝根。

  **7 StopReference**：IB-Frame 1 7 10保留StopAck引用2；
  Expected.Next7/Stop2应binding-matched/Next8；
  Stop0与Stop1各reference-mismatch。不得用本帧Sequence7替代跨方向引用2。

  **8 HeartbeatReference**：IB-Frame 1 2 8 1；
  Expected.Next2/Last1应binding-matched/Next3；Last0应reference-mismatch。
  HS-AssertCodec $good 'binding-matched' ([uint64]3)，坏Expected核Next=null。

  **9 ArmDeclaredExpiry**：IB-Frame 2 1 4，ArmSequence=1/ReleaseBefore=113500；
  独立IB-Put修改ObservedTick为113499时，Expected2→1/Next1应binding-matched/Next2；
  ObservedTick=113500时expired-declaration/Next=null。
  HS-AssertCodec $expired 'expired-declaration'。仅声明整数，不证明真实时钟或出票。

  **10 LookalikeInboxRejected**：先创建一个正常token，分别用全新Object、
  PSCustomObject含Inbox/Accepted=true的相似根、PSSerializer往返所得新身份，
  调用Get-P2BInboxSnapshot均invalid-input，完整12字段根/false flags。
  原token初始完整Snapshot不变。此项仅旧token误用防护，不证明尚未实现的
  ConsumeOwnedControl来源认证，也不把相似Frame宣称为生产交付。

  **11 StoppedInboxNoRecovery**：本组只New一次，Feed Hello seq1后Close端点1，
  peer-closed；原token再Feed StopAck seq2与Take均stopped，Snapshot完整首因仍
  peer-closed/Endpoint1/Stage eof/Sequence null，保留Queue=f1/Next2/Last1。
  HS-Need (CreatedCount -eq 1 -and PostFeed.Reason -ceq 'stopped'
  -and PostTake.Reason -ceq 'stopped')，再核完整snapshot literal。
  不新建token收ACK、不触碰registry。

  **12 Complete**：前11组全完成；CompileCount0、ProductChanged/RealGateImplemented
  与全部authority flags false，五导出仍原集合；HS-Need (Groups.Count -eq 11
  -and CompileCount -eq 0 -and -not RealGateImplemented)。写trace不序列化token。

  上述名称/断言中的SnapshotQueue/Count等是局部观测值，不是新接口或caller事实。
  所有codec结果均按10字段根核查；所有inbox结果用固定IB-Result核12字段根。

- [ ] **Step2 — Prepare，冻结1378条映射，不启动目标。**
  新packet .superpowers/sdd/2026-10-19-p2b-handshake-seam-verification/此前不存在。
  继承旧tree-copy frozen.json的1326条VerificationPath映射，逐项原样核对。
  先核旧manifest DB3FFAFC5D8A9E9501B3C12FFBFBF3735CB5B52D8414CCB020EC71ED52E58418、
  seal DB47535AF5D567C23D94DAC03E491E9E6D3A9C3BC4C92708AC0B44416AA305C8、
  invocations447B896F231E8B7DD714887D919B12700A0AA0F214D3F55F8F45EB3E1DE693F6。
  增加旧packet全部26文件、旧portable叶全部13文件、以下4报告：
  docs/roadmap/2026-10-18-p2b-inbox-tree-copy-repair.md、
  docs/roadmap/2026-10-18-p2b-inbox-evidence-correction.md、
  docs/roadmap/2026-10-18-p2b-inbox-limited-acceptance-proposal.md、
  docs/roadmap/2026-10-18-p2b-inbox-limited-acceptance-approval.md；
  新5输入、engine/SDK2项、批准Spec及本Plan各一份packet快照，共52新增=1378。
  每条role/path/hash列出并展示后才Freeze；旧路径不改写、不补建遗失记录。
  新5个文本输入物理LF；旧文件/二进制保留原字节，不做行尾改写。
  Spec快照必须9C37A7AD...；Plan快照绑定用户批准版本。
  静态parse/JSON/literal/source检查不得import产品或编译。清单数量/旧身份任何异常停。

- [ ] **Step3 — full一次，首次执行新12组夹具。**
  Run: capture.ps1 -Action Run -Slot full -FrozenSHA256 <Prepare实际SHA>。
  新controller预消费row1后才启动同一固定普通PS7目标，直接-File
  handshake-seam.tests.ps1 -TracePath <packet/full.trace.json>。
  Expected:exit0，空stderr，精确15行：Context/Host两行、12条
  PASS: P2BHSeam+上述Group按顺序、末行
  LIMIT: declarations/seam only; no handshake state or runtime authority。
  Trace Groups=12及Observations固定预期全相等，CompileCount0，所有标志符合Step1，
  pins/产品身份不变，ownedexit/drains完成，无timeout。
  任何失败立即冻结final/review，不借第二轮修复或改冻结预期。

- [ ] **Step4 — final一次，不改任一输入或产品。**
  Run:同controller -Action Run -Slot final -FrozenSHA256 <同一SHA>。
  Expected:同15行/12PASS/exit0/空stderr，stdout和完整trace与full哈希相同。
  两轮是新夹具有限重复，不是旧suite回归或新产品RED→GREEN。

- [ ] **Step5 — 保存有限结果，勿声称真实阶段完成。**
  portable-summary保留两raw流/trace哈希及精确可携带trace，读取时保留JSON数字原文，
  不经JavaScript Number回写；逐字段/数字token与raw做只读比较。
  原始记录留ignored，不复制旧raw进Git，不把full/final改判。
  按明确列表提交必要新测试/控制器/预期/记录，旧产品与历史文档不变。

### 测试驱动与无产品变更的明确例外

本任务是为已固定产品新增衔接夹具，首次运行预期应通过；没有新产品实现
可以合理要求先失败再修复。因此不制造缺API的RED，不编写故意坏的产品，
不冒称TDD产品修复或产品变异。先冻结独立断言/预期，反例在12组内核对旧API
保守拒绝。若发现真实产品缺陷或测试向量有误，停止并另定处置，不改既有验收。
这里的有限例外不适用于今后新增生产行为；后者仍须各自TDD与明确预算。

## 拟申请NEW2+1，当前未批准

推荐Native：一个紧耦合夹具交付+一次末尾只读review，避免额外实施代理。
Subagent-driven需要重新提出包含实施/逐任务/末尾review的预算，不能塞入本2+1。

| 槽 | 目标 | 精确预期 |
| --- | --- | --- |
| full |固定产品/冻结夹具12组|exit0/15行/12PASS，完整trace固定预期|
| final |完全同输入/产品/12组|exit0/15行/12PASS，stdout及trace与full同hash|

每宿主最多120000ms执行+2000ms owned exit+2000ms并发drain。
这是离线验证控制器的护栏，不改变真实B在D内完成全部处置的合同。
固定普通PS7.6.5 Core x64/adminFalse；engine：
C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，
SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
同目录SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
不得安装、搜索替代、提权目标、查询/改变安全设置或借用旧额度。

Prepare/Run父控制器也须普通上下文；只拥有当前target确切Process句柄，
预消费先持久化，两个流并发捕获，超限只停止该owned目标，不扫描PID/杀其他进程。
strict UTF8，只去最多一个终止换行，不吞空行；输入/退出/时限/流/预期任一异常停。
两次目标以外不准额外import/probe/compile/测试宿主；无Go/Docker/全仓库测试。

## Task2: 条件只读审查与可恢复保存

**Files:** 新review.md/portable-summary，执行报告
docs/roadmap/2026-10-19-p2b-handshake-seam-verification.md、此Plan实际状态、
current-status/repository-recovery。
**Consumes:** 相同冻结夹具/1378pins、两轮raw/精确oracle、固定产品身份。
**Produces:** 一次完整限定review及所有findings/declined处置、普通开发检查点；不是验收。

- [ ] **Step1 — 仅在两轮均精确后，消耗唯一fresh只读reviewer。**
  按Review Focus审新5文件/预期/两raw记录及Spec范围，不重审整个历史分支。
  固定codec/helper只读必要边界，其他历史hash-only。
  禁止execute/import/compile/编辑/index/HEAD/网络/安全查询/嵌套代理。
- [ ] **Step2 — 保留所有意见及范围弃判，逐项处置。**
  Critical/Important/不完整即停，不自动fix pass/补测/换reviewer。
  Minor仅明确延期；不能据原产品限定验收忽略新夹具问题。
  审查不能把“无握手/能力实现”误称已经实现；允许部分交付须显式标注。
- [ ] **Step3 — 核对明确必要文件后普通提交推送现有开发分支。**
  更新本Plan/执行与恢复记录，写清完成组数/未测项/失败/旧新预算。
  校对便携结果与raw完全一致、远端SHA=local/工作树干净；失败只存unfinished。
  不合并main、不自动验收、不删除原始/worktree/分支，不附加生产阶段。

## 计划自检与执行交接

Spec§1–10覆盖账本见上；真正本地阶段/票据/能力/竞争/时钟/清理未实现，
是明确受阻的独立后续范围，不以通过本夹具闭环它们。五个Review Focus均有
具名向量；literal引用3/2来自旧原始hex，不假设1；心跳使普通序号1/4合法；
两次15行=2header+12PASS+1LIMIT；0编译，无伪RED/新产品变异声明。
所有新辅助函数都在脚本scope，模块五导出和全部产品哈希不变；
没有未定义的生产接口、TODO占位、实际能力或可运行live包。

请用户审阅本计划并明确方法和新2+1预算。推荐Native；批准前不创建
新夹具/packet或执行任何目标。书面Spec批准不等于本计划执行批准。
