# Live Completion Supervisor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for the recommended Native execution. Track the steps below with checkboxes. Subagent-driven execution requires a revised review budget before starting.

**Goal:** 将离线诊断序列的继续权限收敛到存活会话确认，并在失败出口隔离后仍安全收束确切对象。

**Architecture:** 会话/确认登记与出口管理共享私有状态；新拥有对象执行器只能通过出口网关记账。固定序列在同一 PS7 宿主内推进；独立 RED/GREEN 宿主使用各自授权，不迁移确认。普通父上下文只负责固定宿主的有限运输监督，不从磁盘结果授予下一序列权限。

**Tech Stack:** Windows PowerShell7.6.5 Core x64、现有固定托管 SDK、System.Diagnostics.Process、原始字节双流与本地文件；无安装/编译/新原生 API。

**Spec:** [已批准书面规范](../specs/2026-10-19-p2b-live-completion-supervisor-design.md)。批准对象为提交34339923f5e256b95a0f599cab5338b2083d212b中的规范，SHA256：6687E0196D1B8DF5BAEE3E603F5EA86018AE84292D9D38116FCF17A35F6FD49F。

状态：PROPOSED_NOT_APPROVED。用户对34339923后的书面审阅问题回答“是”，
只批准规范及本计划编写，未批准实现/运行。规范原稿的待审标题是历史编写状态；
当前批准记录在此保留，不修改已批准源文件及其哈希。本轮仅文档，0代码/目标/
测试/编译/代理/新审查。推荐 Native，以下新预算及执行方式待用户审阅批准。

## Global Constraints

- 仅现有隔离工作树D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator，
  分支codex/c12-b01-task9-coordinator，PowerShell-only；不合并main/发布/删除历史。
- 磁盘只作证据。不得恢复Session/确认、凭EXPECTED_*放行、复用旧批准或返还已消费尝试。
- 每个会话冻结身份固定，不热修、不动态重冻结；核心RED/GREEN和执行器RED/GREEN
  是分别授权的新序列，不把旧RED文件/确认当作GREEN首步权限。
- 所有出口失败后不再发起逻辑读写/解析/重开/Complete/Publish；有限资源关闭不作为
  重新序列化或显式Write/Flush重试借口。首因直接保留，次因最多8项，第9项overflow。
- SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/RealGateImplemented
  保持false；无生产握手/Job/watchdog/持久化事务或监督器死亡全体退出保证。
- C1/C2仍是未关闭的重要项；旧B1/B2窄修观察、全部历史失败、冻结额度不改判。
  R2和Minor2继续延期，不重建丢失详情、不修改旧owned-run或旧七组oracle。
- 新正常路径不调用旧owned-run的重复写回执逻辑。固定观察器只读复用：
  CD6E9564E77FF55812F9E96291B0F4E511758245A9BBC53AD2802B95E809C6E5。
- 固定fixture：EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F。
  固定engine：362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
  SDK：2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- codec9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C、
  inbox47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC不变。
- 不传任意命令/脚本块/故障callback/伪造观察setter；Process只来自固定目录/参数表，
  只处置仍持有的确切对象，不按PID/名称重找或杀树，不查询无关进程/安全状态。
- PowerShell/manual ledger代替Bash包装及额外baseline/full-suite/task-done；有限批准
  优先，无自动追加fix/retry/re-review，不删除ignored raw。不同任务接口按本计划核对。

## Review Focus

1. 文件已出现完整成功字节，但写入失败/unknown或无活确认：真实入口仍拒绝，不靠人工停线补门。
2. finally/汇总/输出异常再次进入同一失败出口，或通过别名/API绕过网关：清理继续，后续逻辑I/O为0。
3. 最后输出、flush、关闭或调用返回发生在确认之后：确认不得提前可消费，末步也不得遗漏完成动作。
4. 子步骤自报/序列化摘要/跨Session与冻结身份确认/重放：只能是观察，不能制造登记身份或实际启动。
5. 宿主丢失、阻塞文件I/O、流关闭失败或不完整后代观察：unknown不可成功；不能夸大硬期限/实际进程覆盖。

对应测试在任务1–3明确列出；不能安全动态诱发的晚写失败、全部输出边界及宿主死亡
留给源码审查并明确证据限制，不用通用故障钩子冒充已测。

## 文件、身份与私有接口

新leaf：docs/diagnostics/isolated-fuzz/p2b-live-completion-supervisor/。
以下12输入在首轮前全部写好/物理LF/静态解析，然后冻结：

| 文件 | 责任 |
| --- | --- |
| .gitattributes、README.md、expected-output.json | LF、运行边界、独立字面根/输出/固定profile与case表 |
| session.psm1 | 唯一私有Session/确认登记、步骤消费及失败/关闭；只导出固定外观 |
| sinks.ps1 | 在模块私有作用域内使用出口表；单次I/O、失败隔离、首因及有界fallback |
| owned-step.ps1 | 在模块私有作用域内收束新拥有对象执行；不包含旧Save-Receipt重试 |
| sequence.ps1 | 模块私有的固定步骤完成/发布/确认和序列组合；无通用命令执行 |
| core.tests.ps1、owned.tests.ps1、sequence.tests.ps1 | 固定用例及真实消费者断言，同一宿主内执行，不另起测试宿主 |
| capture.ps1 | 一个新PS7宿主的固定Grant入口；执行相应测试并有界保存本轮结果 |
| driver.ps1 | 既有普通父上下文内Prepare/RunOne；监督确切capture宿主，不从磁盘授予权限 |

新packet：.superpowers/sdd/2026-10-19-p2b-live-completion-supervisor/；
执行报告：docs/roadmap/2026-10-19-p2b-live-completion-supervisor.md。
派生portable-summary/review/disposition和状态/恢复文档是输出，不是运行输入。

固定engine路径：C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe；
SDK为同目录System.Management.Automation.dll。固定fixture与原seam路径沿用旧leaf，
不复制产品、不执行旧控制器。stdout/stderr各唯一原始字节消费者，并行开始排空。

接口统一用[object]不透明句柄；实际登记以ReferenceEquals查私有有界条目，不能由
调用者字段值构造。仅以下外观供capture/tests使用，其他操作模块私有：

- New-LSSession([string]Grant,[string]Case,[string]FrozenSHA256) → opaque handle；
  Grant/Case只允许expected-output.json冻结表，路径由表派生，不接受任意Root/命令。
- Test-LSAdvance([object]Session,[int]Ordinal,[object]PreviousAck) → bool；
  无副作用的真实预门，Invoke-LSStep再次检查后同串行路径消费，不从磁盘补状态。
- Invoke-LSStep([object]Session,[int]Ordinal,[object]PreviousAck) → 有序根
  {Accepted:bool,Rejection:string|null,Ack:object|null,Observation:object|null,Error:ErrorRecord|null}。
  私有固定profile决定目标/预期。错误参数只是拒绝、不Start或消费；实际控制器I/O/
  生命周期错误使Session不可逆Failed。拒绝不能临时放宽下一次的真实门。
- Write-LSLedger(Session,SinkId,[object]Value)、Read-LSLedger(Session,SinkId)、
  Complete-LSLedger(Session,SinkId,[int]RowIndex,[object]Row) → {Ok:bool,Value:object|null,Error:ErrorRecord|null}。
  SinkId只对应注册路径/所有者，返回数据不能作为确认。规范化路径共用失败状态。
- Write-LSOutput(Session,[string]SinkId,[string]Text) → 同一I/O根；ConsoleOut/ConsoleError
  也是固定登记出口；未知SinkId拒绝，不尝试另找日志出口。
- Get-LSView(Session) → 只读语义快照：Status/FailureLatched/NextOrdinal/Reserved/
  StartCalls/AckIssued/AckConsumed/Sinks及首/次因/overflow；不是可回传的运行身份。
- Add-LSFailure([object]Session,[Management.Automation.ErrorRecord]Record) → void；
  只能锁存实际错误/使状态失败，不能撤销失败或产生确认。任务1的后续错误由测试
  工具独立catch真实文件错误后送入同一锁存路径，不对失败Session发起更多文件访问。
- Close-LSSession(Session,[ValidateSet('Finish','Stop')]Mode) → 最终快照；
  Stop撤销确认并结束；Finish仅在全部规定动作正常返回后结束。关闭后所有启动门拒绝。

内部Invoke-LSOwnedStep仅由固定序列调用，持有实际Process/异步任务，观察直接注册
到所属Session，不能接受调用者传入Verified/HasExited等“观察”。确认只由私有完成
函数在成功返回/冻结检查/全部发布结束后登记，永不从公开Observation回填。
记录的实际退出不等于所有后代独立OS观察；本计划夹具不派生未列入预算的后代。

登记上限固定：每宿主32个被测Session、每Session32个Sink、最多2个步骤确认条目；
超限拒绝且不分配新目标。每Sink的LogicalIOAttempts/RejectedWithoutIO分开记账，
前者仅在真实文件/输出调用前递增；拒绝不增加实际访问计数。计数属于运行观察，
不是替代实际文件不变/错误身份/源代码所有访问路径审查的唯一证明。

## 拟批准新预算、运输边界及冻结

推荐Native：父代理实施，无实现子代理/逐任务review；仅末尾条件1fresh只读review。
NEW5个普通PS7宿主+8个受控PS7子进程=13新进程，条件1审查。以下是完整预算，
不是旧4+4预算的追加槽；书面计划批准前均不可执行。

| Grant顺序 | 宿主 | 子进程 | 固定内容 |
| --- | ---: | ---: | --- |
| core-red |1|0|核心两项旧机制，恰2预期FAIL/exit1 |
| core-green |1|0|核心7组完整正反例/exit0 |
| owned-red |1|1|Hold一次；真实回执首次失败后旧最终写重试，恰1预期FAIL/exit1 |
| owned-green |1|3|回执故障Hold、EmitAndExit、超时Hold；预启动失败不Start；6组/exit0 |
| sequence-green |1|4|A首步EmitAndExit、B首步EmitAndExit、输出故障C步EmitAndExit、A末步原seam；4组及原15行/12PASS/exit0 |

core测试没有Process创建。owned/sequence中独立被测失败Session与测试宿主自身
运输/记录状态分开；只接受冻结负例，不能把宿主自身错误归类为预期RED。
序列A两步在最后同一PS7进程内使用实际确认推进，B/C是固定的隔离测试Session，
各自独立批准范围/目录/身份；不把其失败当A的成功证据，不授予额外目标。
B只批准首步目标，余下请求即使有确认也无剩余目标预算；测试后Stop撤销它。
末步不再签发下一步确认，所有属于该会话的报告/关闭动作在其成功结束前完成。

固定Case绑定：A=sequence-a，两个已批准步骤[fixture EmitAndExit,original seam]、
目标预算2；B=sequence-b，描述两个步骤但本批准仅含B0 EmitAndExit、目标预算1，
B1任何请求在创建Process前被预算门拒绝；C=sequence-output-fault，只有C0
EmitAndExit用于实际发布失败负例、目标预算1。身份tuple包含各Case的固定步骤表
摘要，不通过热换源码构造跨代次Case。foreign Ack的会话/冻结身份绑定静态检查
与真实跨Case拒绝组合验证，不冒称在同一活Session内动态更换冻结代次。

driver在普通父上下文内被调用，不为Prepare/读哈希/记账再起PS7。RunOne只接受
上述5个固定Grant，创建一次确切capture进程。每Grant的intent预消费记录和final
结果文件分开、各至多一次写；失败不原地补写，已有/不明记录禁止重试。进程退出、
流排空和文件关闭在最终结果发布前完成。任何driver故障立即停线；残留final文件
不授权其他Grant。后续独立Grant来自明确批准的本计划和执行者实际观察，不是
从旧Outcome自动分派；工具调用返回失败/不明也停止，不凭磁盘反推成功。

driver是最小固定运输夹具，不导入正在变异的被测模块来决定自己是否成功。
它的新异常路径只能静态审查/5次实际运输覆盖，不冒称已有独立动态全路径证明。
这项bootstrap边界须在最终审查明确判断，不能默默继承旧运输器的finally重写。
本提案明确请求这一首次装配边界：driver/capture的运输外壳及固定组合接线没有
额外独立RED/故障注入宿主，其下的权限门/出口核心和拥有对象失败路径才有指定
RED→GREEN。不能把首次装配通过写成外壳异常路径均已RED验证；若需该独立验证，
必须在执行前修订本计划及预算，不能临时借用本轮13进程以外额度。

core/owned宿主MaxMs=120000；sequence宿主MaxMs=150000，内部整轮预算140000。
退出等待2000ms、双流共用2000ms。fixture Hold=45000ms；故障/正常MaxMs=2000，
超时200；seam上限110000。各等待取本步上限与整轮剩余（预留退出/双流4000ms）
较小值，剩余不足则拒绝Start。不能以每步重置整轮期限。文件调用/宿主挂起不
提供硬实时保证；外层终止监督器时不得声称其后代已退出，结果unknown并停线。

首轮前继承上一GREEN1599项，新增45 raw、11portable、旧执行报告1、已批准规范
快照1、本批准计划快照1、新输入12，共1670项，逐路径/角色/哈希列全并查冲突。
旧manifest固定3A763103C93625C522CF8BFF997F9280EA07F3278A6D3FCE6D221759368D4237。
原件缺失即停，不从portable重建。新packet只允许预存本计划progress.md。
core-red/ core-green/owned-red/owned-green四清单/对应源码快照分开：core修复只改
session.psm1+sinks.ps1两项，其余1668项相同；owned-red与core-green输入相同，
owned修复只改owned-step.ps1一项，其余1669项相同。sequence完全复用owned-green，
所有测试、oracles、控制面和运输文件始终不改。每次目标前后全清单核对。

任一非精确RED、未知、漂移、未退出/未排空、救援、未计划错误或记录失败，在
完成确切对象有限清理后冻结全部剩余Grant/审查。仅两个精确RED各允许对应一次
预定修复，无自动再修、重试、替代运行或二审。输出PASS只是观察，不覆盖后续失败。

## Task 1: 私有存活确认与失败出口核心 RED→GREEN

**Files:** 创建全部12输入；本任务只在RED后修改session.psm1、sinks.ps1。
**Consumes:** 已批准规范、固定Grant/Case目录、独立oracle；无旧运行权限。
**Produces:** 上述Session/I/O外观、core-green冻结清单；实际观察登记保持私有。

- [ ] Step1：先写core.tests和字面oracle。基线为旧磁盘Outcome放行及失败后仍访问
  出口的真实机制窄提取，API完整，不以缺函数/假ErrorRecord产生RED。
  测试通过真实Test-LSAdvance/文件网关观察行为，不grep源码或断言模拟器存在。
- [ ] Step2：静态解析全部PS脚本/模块、检查LF/输入表，Prepare冻结1670项。
  Run：driver.ps1 -Action Prepare -Grant core-red；随后
  driver.ps1 -Action RunOne -Grant core-red -FrozenSHA256 <打印的清单哈希>。
  Expected：exit1；恰FAIL: DiskMustNotMintAck及FAIL: FailedSinkMustNotBeReused；
  完整trace为无真实确认却获准Advance，以及同一目录写故障后再访问的事实；
  StartCalls=0、无子槽、空stderr、运输退出/双流/后置pins完整。
  driver参数固定为-Action Prepare|RunOne、-Grant上述五值、-FrozenSHA256；
  后续Prepare另接受-PreviousFrozenSHA256以核对唯一允许的输入变化，不执行目标。
  RunOne不得自动派发其他Grant。首次Prepare保存规范/计划批准快照，后续只保存
  对应清单及变更前后源码；不能因首轮缺失改用当前摘要生成历史快照。
- [ ] Step3：只修两项核心文件：登记身份不来自磁盘；访问失败永久隔离路径，
  先锁存错误再拒绝后续I/O；只接受实际模块内完成路径签发的确认。
  保存RED源码后Prepare core-green，Expected恰2项变化，无其他输入漂移。
- [ ] Step4：RunOne core-green唯一一次。Expected：以下7组同名PASS、exit0/
  空stderr、完整trace0差异、0子进程/StartCalls、全部记录/pins完整。

| 组名 | 独立断言 |
| --- | --- |
| DiskCannotMintAck | 写字面EXPECTED_GREEN文件；新Session对下一序号/空确认拒绝，原件不变，无权限恢复 |
| FailedSinkRejectsAccess | 真实目录写错误；后续Read/Complete/Write全部拒绝，失败后LogicalIO增量0、首因同引用 |
| InvalidLedgerStops | 缺文件、坏JSON、无效索引各自独立Case，永久Failed、不重建、不发确认 |
| FirstAndSecondaryBounded | 实际首次错误、重复引用、9个独立后续错误；首因不变，8次因/overflow明确 |
| OutputFailureNoAck | 测试工具中实际disposed StreamWriter作为ConsoleOut，网关真实失败，SessionFailed/无确认，finally恢复原writer |
| PreconsumeFailureNoStart | 固定启动记录目录写失败；Reservation已消费，StartCalls=0/无返还，无第二次访问 |
| HandleCopiesRejected | Session摘要/复制/反序列化对象及自填Verified=true不被登记接受，不能创建或恢复确认 |

例如测试值必须直接等价于：
```powershell
# 固定用例产生的实际返回/快照，不重新用被测函数生成期待值。
Assert-Equal $actual.AdvanceAccepted $false
Assert-Equal $actual.PostFailureLogicalIO 0
Assert-Equal $actual.StartCalls 0
```
真实确认的正常消费/跨会话/重放在任务3用实际拥有目标产生后验证；不在本任务
通过setter制造一次成功OS观察。输出末段/flush/关闭及晚写失败不冒称本组已动态覆盖。
- [ ] Step5：显式12输入及执行/状态/恢复文档开发检查点；只声明任务1核心验证，
  推送并核远端相等。未完成项/旧阻塞/覆盖限制明确，不合并main。

## Task 2: 新拥有对象执行边界 RED→GREEN

**Files:** 使用owned.tests.ps1；RED后唯一修改owned-step.ps1。
**Consumes:** 任务1的模块/出口表/失败状态；core-green固定输入；原fixture/观察器。
**Produces:** 私有Invoke-LSOwnedStep及直接观察，owned-green清单供任务3使用。

内部签名：Invoke-LSOwnedStep(SessionEntry,StepDescriptor) → 私有OwnedObservation。
StepDescriptor只来自冻结表，含可执行文件/参数/流与回执SinkId/等待上限及字面预期；
OwnedObservation来自实际持有对象，包含Started/Association/PID/OwnedExit/ExitCode/
StreamsComplete/TimedOut/TerminationRequested/ElapsedMs及实际错误。禁止公共接受该
类型作为“已观察”输入。无Start前失败不得调用未关联对象的Kill/Wait。

- [ ] Step1：先完成字面测试；旧机制基线直接保留最终Save-Receipt重复写行为，
  只作新隔离目录的RED实现，不能偷偷复用旧helper或让网关提前掩盖待测重试。
  Prepare owned-red只另存清单/快照，输入与core-green完全相同。
- [ ] Step2：RunOne owned-red一次：Hold子进程1个，实际首次回执目录写失败后
  基线最终写再次失败。Expected恰FAIL: ReceiptSinkMustNotBeRetried、exit1，
  Primary保留、额外最终写错误确实存在；子进程实际退出/双流完成、0救援。
- [ ] Step3：仅修owned-step：所有记账经过出口表，Failed后finally只清理不记账；
  启动后及时并行排空，所有异常先锁存，有限终止/退出/双流/Dispose不依赖保存。
  Prepare owned-green，Expected恰1输入变化，core与测试完全不变。
- [ ] Step4：RunOne owned-green一次。Expected6组PASS、exit0/空stderr、完整
  字面trace0差异，3子进程全收束/0救援/pins完整：
  ReceiptFaultNoRetry、FirstCauseSurvivesCleanup、NormalExit、Timeout、
  PreStartFaultNoStart、UnassociatedNoKillWait。
  故障子Case：Hold/2000ms，Primary实际IO错误、失败后LogicalIO=0、无第二次回执
  写次因；正常Case：EmitAndExit/2000ms，exit0，两流恰P2B-CONTAINMENT-OUT/
  P2B-CONTAINMENT-ERR；超时Case：Hold/200ms，超时/终止/退出/双流已观察。
  预启动输出路径目录故障Start=0、Unassociated/无PID、无Kill/Wait；以上负例
  各独立被测Session，宿主控制State不得失败。不是旧七组同oracle复跑。
- [ ] Step5：保存任务2检查点、完整角色/直接观察范围和剩余预算，普通开发分支
  提交推送核相等。不得因通过而回填C1/C2为正式关闭。

## Task 3: 实际存活序列、原seam、条件审查与可恢复交接

**Files:** sequence.ps1/sequence.tests.ps1/capture.ps1使用首轮已冻结版本；仅新增输出。
**Consumes:** owned-green全输入、固定Grant/Case表、真实OwnedObservation及模块确认。
**Produces:** sequence-green完整观察、唯一完整review/disposition、portable与恢复检查点。

私有完成路径只从所属OwnedObservation和已返回的受管I/O推导结果；Test-LSAdvance
检查通过不签发确认，真正消费/预消费在Invoke-LSStep内再次核对。没有真实确认
的文件负例与真实确认之间，唯一差异必须是私有登记来源，不靠名字或摘要字段。

- [ ] Step1：RunOne sequence-green唯一一次，保持同一宿主/同一模块私有登记：
  A0正常夹具产生真实AckA；B0正常夹具产生另一个Session确认，固定身份tuple
  (冻结哈希,Grant,Case)不同；B只批准一个目标，不可借测试启动第二个。
  测试工具在调用C0的Invoke-LSStep之前，将ConsoleOut换为实际closed writer，
  并在finally恢复原writer。该固定profile在启动/等待时不向ConsoleOut发布，
  子进程两流独立重定向；故C0正常退出并收束后，第一次正常完成发布才真实失败。
  要求无确认/Failed/NextStart=0；不需要生命周期中途callback或伪造观察。
  对A1入口依次检查缺确认、复制AckA、序列化摘要、AckB/错误身份/错误步，均拒绝
  且不消费AckA/不创建Process；合法AckA只消费一次后，A1实际启动原seam。
  旧AckA再用拒绝，不产生第五个子进程。B/C隔离测试结束均关闭并撤销确认。
  末步无下一步确认，全部必要输出/关闭完成后A结束；新Session不从A文件恢复。
  Expected4组PASS：LiveAckBindingsAndReplay、OutputFailureNoContinuation、
  PositiveOwnedContinuation、SessionClosedNoResume。4个子进程均exit0、实际
  关联/退出/双流已观察，无救援；C的预期输出故障仅被测State失败，不污染宿主。
- [ ] Step2：同次运行内核原seam，不能另起FINAL或补跑。Expected原15行/12PASS、
  完整ordered JSON/raw numeric token oracle0差异，stdout
  F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE及trace
  4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E。
  完整宿主stdout模板在首轮前固定，依次为STEP: sequence-a/0、
  STEP: sequence-b/0、STEP: sequence-a/1各一行，接原seam15行，接上述4PASS各一行。
  STEP只在各步实际退出/流关闭/后置检查后的完成发布边界输出，不在Start前输出。
  C在首次实际输出即失败，所以没有STEP: sequence-output-fault/0。fixture原始
  两流只保存，不转发；只有seam流原样转发。每个STEP行/本步必须的flush与文件
  关闭都在该步确认之前；不是先签确认再输出STEP。比较仅允许固定Context替换和
  事先声明的CRLF/LF处理，不临时过滤。组合流哈希从字面模板生成，不从实测回填。
  4条PASS是capture的被测Case报告，不是A/B/C的确认。A的所有必需输出关闭已完成、
  B/C已Stop、全部确认已消费/撤销后才发表此报告；报告失败使整个Grant不精确、
  driver停止，无任何可启动的存活会话，不能用这些报告文件启动另一个Grant。
- [ ] Step3：仅前5宿主/8子进程/所有trace/pins/记录全部精确，才消费唯一fresh
  只读审查。审查12输入、所有实际调用/文件绕行、全部原件、两次修复快照、五项
  Review Focus和规范全部边界。最强可用模型、独立上下文；无执行/导入/网络/
  文件或index/HEAD改动/额外代理。完整返回所有发现及每项Declined-to-judge。
  Critical/Important或审查不完整即停，无自动fix/再测/二审；Minor只延期记录。
- [ ] Step4：保存完整审查/逐项处置；C1/C2只有被证据支持的边界可窄结论，不升级
  产品/正式验收。portable保存raw JSON原文本（不JS数值重发大整数）、流/源码
  快照字节、清单/批准规范/计划哈希、预算、unknown及全部限制，raw留存ignored。
  便携/后续报告/处置/状态/恢复为作者逐字节/文本/哈希核对，不标成已独立审查。
- [ ] Step5：更新执行报告、current-status、repository-recovery；显式必要文件
  提交推送既有开发分支，核远端SHA等于本地、工作区状态；不包含缓存/密钥/工具/
  Docker数据/ignored原件，不删除worktree或raw。失败只保存明确未验收检查点。

## 计划自检与执行选择

规范§1–3目标/选型→核心门与Task3；§4–8所有权/顺序/出口/首因→Tasks1–3；
§9旧helper兼容性→Task2新执行器及新oracle；§10固定身份/无恢复→独立Grant与
同宿主A序列；§11证据矩阵→7核心+6执行器+4序列组，安全晚写/flush/关闭/死亡
仅明确静态边界；§12恢复/预算→13进程/1条件审查与交接；§13无实现结论保留。
接口名称/属性统一，无未定义的观察注入接口；源码修复只有两个精确RED窗口。

推荐Native：三个任务共享私有登记和出口合同，内联实施可减少接口漂移；完成后
保留一次真正独立整体审查。Subagent-driven可选，但逐任务实现/审查需先修订
代理/审查预算，不能把当前唯一条件review拆作多次。
请审阅本计划并选择执行方式；推荐批准Native以及NEW5宿主+8子进程=13新PS7进程、
条件1次只读审查。批准前仅文档，旧预算/失败/C1/C2阻塞/R2/Minor2延期均不变。
