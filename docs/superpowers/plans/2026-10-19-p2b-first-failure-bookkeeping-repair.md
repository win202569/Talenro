# P2-B First-Failure Bookkeeping Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for proposed Native execution. Track steps with checkboxes. Proposed scope is not implementation or runtime permission.

**Goal:** 修复 R1：最终记账的读取、序列化或写入失败不得取代先前首因，且失败后不得启动后续目标或宣称完成。
**Architecture:** 新目录中把三个记账出口收敛到同一私有首因容器/安全记账函数；真实生命周期仍用固定 owned-run 和关联观察函数。先用真实文件错误复现旧未保护记账的异常逃逸，再只修共享记账文件；独立验证记账、完整七组收束和一次原 seam 集成。
**Tech Stack:** 普通 Windows PowerShell7.6.5 Core x64、固定托管 SDK、本地文件 I/O；无安装、编译、新原生 API、安全设置或产品依赖。
**Spec:** [R1/R2 完整审查](../../diagnostics/isolated-fuzz/p2b-process-association-guard-repair/review.md)、[协调者处置](../../diagnostics/isolated-fuzz/p2b-process-association-guard-repair/review-disposition.md)、[关联护栏合同](2026-10-19-p2b-process-association-guard-repair.md)及[握手合同 §9](../specs/2026-10-19-p2b-handshake-contract-design.md)。

状态：PROPOSED_NOT_APPROVED。依据 b458b35238c8fe5566716c50cc9461bc33a2751a。
本次“按建议继续”只进入书面 R1 修复计划；0代码/目标/测试/编译/代理/审查。
推荐 Native。新实现和新额度待批准；旧结果不改判。R2 Minor 继续明确延期，
不得趁适配修掉 PreStart/prestart 键碰撞，也不得重建已丢失的历史详情。

## 根因、边界与选型

R1 的共同原因不是进程漏清理，而是 finally 中的最终台账操作仍可抛出：
旧 test:118–119 的读取/更新/保存、旧 capture:151 的子台账保存、:216 的父台账
保存不在次级错误处理内；后置最终保存也有相同前提。首个异常被新的 finally
异常遮盖，外层 catch 只能看见后者。此结论已由作者与独立审查源码核对；
未做动态注入。原七组和 FINAL 的成功观察仍有效，不代表此异常路径已通过。

选择显式首因锁存和统一安全记账，不只包住一行 WriteAllText；保护读取、
格式转换、更新、序列化和最终写入。把异常吞掉而仍记 SUCCESS 或无限重试，
都不满足 R1。只改原 owned-run 无法覆盖外层台账，拒绝；不重写产品生命周期。

保证是存活控制器内首因关系、有限次级错误和失败封闭，不是磁盘永久故障时
仍能落盘的保证。所有证据出口都失效时只能保留 unknown/INCOMPLETE，
尽力向现有错误流报告；错误流也失败不能凭空声称持久化、恢复或验收成功。

## Global Constraints

- 仅现有隔离工作树、codex/c12-b01-task9-coordinator，PowerShell-only。
  不合并 main、发布、删除旧证据、改 Git 信任/安全设置或安装依赖。
- 固定 owned-run SHA256：7A425C661968CD74BFD8CD4EAA25CBA53ED36CC09E3D2946F9D095FC095B25E1。
- 固定关联观察函数：CD6E9564E77FF55812F9E96291B0F4E511758245A9BBC53AD2802B95E809C6E5；
  fixture：EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F。
- codec：9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C；
  inbox：47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC 不变。
- engine：362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
  SDK：2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- 原关联护栏六输入、33 raw、9 portable、完整审查/处置/报告/计划不变；
  其运行3宿主+4子进程和审查1额度已耗尽，不重开。所有其他旧冻结余量继续冻结。
- SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/
  RealGateImplemented 保持 false；原有限离线批准、Minor2 和 R2 延期不变。
- 只处置调用者创建并仍持有的确切 Process；不按 PID/名称重找、杀树或查询无关进程。
  无故障 callback/setter、任意脚本块、假能力、公共事件回放或产品开关。

## 文件与私有接口

新 leaf：docs/diagnostics/isolated-fuzz/p2b-first-failure-bookkeeping-repair/。
七输入：bookkeeping.ps1（共享首因/记账）、bookkeeping.tests.ps1（RED/GREEN）、
containment.tests.ps1（原七组的窄适配）、capture.ps1（Prepare/Run/Integrate）、
expected-output.json、README.md、.gitattributes（物理 LF）。
原 owned-run/观察函数/fixture/seam 只读引用，不复制产品。新捕获器和测试从
上阶段固定源码适配，全部在 RED 前完成，之后只允许 bookkeeping.ps1 一次修复。
派生 summary/review 是输出，不作为执行输入。

新 packet：.superpowers/sdd/2026-10-19-p2b-first-failure-bookkeeping-repair/。
执行报告：docs/roadmap/2026-10-19-p2b-first-failure-bookkeeping-repair.md；
保存交接时更新 current-status/repository-recovery，显式文件列表提交现有分支。

bookkeeping.ps1 只定义以下控制器私有函数，不在加载时运行目标：

- New-P2BFailureState()：返回普通有序根 PrimaryError=ErrorRecord|null、
  SecondaryErrors=ErrorRecord[]、SecondaryOverflow=bool、Outcome='INCOMPLETE'、
  FinalLedgerWritten=false。初始主因空、次因空、overflow=false。
- Add-P2BFailure([object]State,[ErrorRecord]Record)：首项直接锁存实际传入记录；
  已有首因则仅追加次因，最多8项，第9项置 overflow；从不替换首因或清空状态。
- Save-P2BLedger([object]State,[string]Path,[object]Rows)：捕获序列化/写入错误，
  交给同一锁存函数，返回 bool。false 不抛出覆盖旧因，不重试、不自己写另一个文件。
- Complete-P2BLedger([object]State,[string]Path,[int]RowIndex,[object]Row,
  [ValidateSet('child','integration','parent')][string]Scope)：真实读取/解析已有
  台账、校验索引/更新行并调用同一 Save；任何失败均锁存后返回 false。
  Scope 只是三处固定调用方的诊断标签，不改变处理算法或给予能力。所有 Path
  必须在新 packet 内；坏台账不静默重建，不在失败时伪造完整行。
- Write-P2BFailureFallback([object]State)：对现有 Console.Error 做至多一次
  有界摘要输出；输出错误作为次因保留，不递归记录、不重试。摘要不序列化
  ErrorRecord 对象图；首/次因类型和截断消息各最多1024字符、次因最多8项。
  若错误流不可用，没有额外保证，不升级成功。

三处调用方在自己的 catch 中、进入 finally 之前，把实际首次错误交给同一
State；finally 的资源清理错误和全部记账/记录错误只能另记。没有先前错误时，
记账错误本身成为首因。同一调用不重复添加同一个错误，以免挤掉后续记录。
生命周期首先安全收束，记账不跳过 Kill/Wait/双流/Dispose。主因与次因可在
存活宿主内直接与其独立 catch 锚点比较；不再与自动 Error 列表包装器比身份。

初始预消费写入同样用安全 Save：失败则没有 Start，保留 INCOMPLETE/未知
消费记录并停止；不得因为磁盘上缺行而自动重试。成功、失败后置写入及最终
trace/details 输出均纳入同一错误边界；任何失败不得先发布全运行成功。
已输出的单组 PASS 只是观察，不能覆盖后续非零退出/记录失败。

## Review Focus

1. 后写失败替换首因：三个 scope 的真实目录写失败必须不逃逸、首因同锚点、次因单列。
2. 最终读取/解析失败仍逃逸：缺失文件/真实坏 JSON 拒绝，首因不改、不重建台账。
3. 无先前错误/次因溢出：首次记账错误成为首因；8项满后overflow，始终INCOMPLETE。
4. 正常记录被一律判失败或错误后继续启动：健康更新成功，无错误；所有启动和成功门核 State。
5. 失败 sink 又被用于记录其自身失败：无递归写盘/重试，错误流仅best-effort；无法持久化保持unknown。

## 拟批准新预算、冻结与停线

NEW4顶层PS7宿主+4受控PS7子进程=8新进程；条件性1次独立只读审查。
既有普通父上下文负责 Prepare/监督，不为记账另起PS7；无实施代理/额外探针。

| 顺序 | 新宿主 | 新子进程 | 目标与预期 |
| --- | --- | --- | --- |
| RED |1|0|共享最终记账旧逻辑的真实写失败；预定逃逸断言失败 |
| GREEN |1|0|共享记账七组固定正反例 |
| CONTAINMENT |1|3|原七组关联/生命周期收束，使用同一新记账函数 |
| FINAL |1|1|新 Integrate 对原 seam 一次调用 |

这些 I/O 测试不创建/模拟 OS 进程；三个写失败向量调用实际共享 Save 函数，
Complete 的读取/解析及健康更新另有固定向量。不得用目录读失败冒称写失败，
也不宣称逐一动态注入三种真实生命周期 finally。其调用方接线、预消费失败
不 Start、错误流失效及未知API/OS错误由完整源码审查补足，动态覆盖边界须明示。

每次 Start 先预消费；启动失败也消费，已有/不明台账禁止继续。只有精确 RED
允许 bookkeeping.ps1 一次预定修复。任何未知、非精确、漂移、退出/双流不明、
救援、记录或解析失败，完成确切对象有限安全收束后停，冻结全部余槽/审查。
测试内预先列明且正确收集的 I/O 故障是被测向量，不是允许修脚本的开放额度。

顶层 MaxMs=120000、退出等待2000ms、双流共用2000ms；fixture Hold=45000ms，
fault/normal MaxMs=2000、timeout=200、helper ExitWait/Drain各2000。
FINAL seam MaxMs=110000，外层120000总监督不重置；硬实时/所有出口失效不保证。

Prepare 继承上阶段1487映射项，加33 raw、9 portable、报告和计划各1、
新七输入、本批准计划快照1，共1539项。完整列出路径/角色/哈希；重复路径
不得冲突，缺原始记录就停，不从 summary 重建。RED/GREEN分别保存源码快照/
完整清单，只有 bookkeeping.ps1 一项哈希变化，其余1538项必须相同；之后不改。
新 packet 只可预存本计划 progress.md，不能存在先前槽/结果；前后全清单核对。

## Task 1: 最终记账首因保留的真实 RED→GREEN

**Files:** 新七输入；所有历史只读。
**Interfaces:** bookkeeping.tests.ps1 -Phase RED|GREEN -TracePath <fixed path>
-DetailsPath <fixed path>；capture.ps1 -Action Prepare|Run|Integrate
-Slot red|green|containment|final -FrozenSHA256 <64hex>。

- [ ] Step1：先写字面 oracle/七组断言与调用方，再准备旧行为基线。基线 Save/
  Complete 直接沿旧读取/修改/保存路径，最终读写失败仍逃逸；不是缺 API、假错误或产品 mutant。
  所有调用方先接入首因容器和新函数；RED后只修改共享记账实现，不改预期/调用方。
- [ ] Step2：静态解析、Prepare冻结RED，普通父仅一次 Run -Slot red。
  新 packet 中先以真实文件错误取得第一 catch 锚点并锁存，随后直接调用共享
  Save，以目录作输出路径触发第二个真实写错误；外层仅观察逃逸，不覆盖已存记录。
  Expected：恰好 FAIL: FinalLedgerMustNotEscape、exit1、无子槽，保存
  Escaped=true/ReturnedNormally=false 和两个不同实际错误；非该结果立即停线。
- [ ] Step3：一次最小修复 bookkeeping.ps1，覆盖读取/解析/更新/序列化/写入的
  真实异常边界，保留首因/有限次因并返回false；保留RED快照，仅一项pin更新。
- [ ] Step4：普通父唯一 Run -Slot green，七组均输出同名PASS，完整trace零差异：

| 组名 | 固定断言 |
| --- | --- |
| ChildFinalWritePreservesFirst | child scope真实失败、不逃逸、返回false；Primary与独立第一catch同引用，恰1独立次因；INCOMPLETE/未落盘 |
| IntegrationFinalWritePreservesFirst | integration scope同上，不以错误文本相等代替身份关系 |
| ParentFinalWritePreservesFirst | parent scope同上，不能因父记录失败而宣称完成 |
| FinalReadParseFailurePreservesFirst | 缺失路径和已写坏JSON分别失败、不逃逸、不重建；两次真实次因有序，首因同引用 |
| FirstBookkeepingFailureLatches | 无初始首因时，首次真实记账错误成为Primary，次因0、未落盘、INCOMPLETE |
| SecondaryOverflowStops | 初始首因不变；9次真实最终错误只保留8次因、overflow=true，不重试、不成功 |
| HealthyFinalLedger | 正常已有单行台账替换后内容逐字段等于独立字面根、写成功、首因null/次因0/无overflow；不改变任何真实能力标志 |

  RED/GREEN均0子进程、不加载产品。State根/作用域关系是控制器数据验证，
  不是生命周期回执。宿主exit0/空stderr仅在GREEN全精确，父退出/流/pins完整。
  details保存实际类型链/身份关系，raw保留；不以序列化声称恢复ErrorRecord引用。
- [ ] Step5：保存明确仅Task1完成的必要开发检查点；失败只保存unfinished。
  PowerShell/manual账本替代Bash技能包装，不增加额外task-done/历史suite/删除。

## Task 2: 真实收束、FINAL、条件审查及恢复

**Consumes:** Task1精确RED/GREEN、共享GREEN清单、无后续槽消费；所有调用方已冻结。
**Produces:** 原七组和一次seam观察、完整独立review/处置、可换机恢复开发检查点。

- [ ] Step1：唯一 Run -Slot containment，执行上阶段相同七组/三子进程，并保持
  首因因果负例、正PID关联、无救援、PreStart无Kill/Wait等完整原预期。
  新错误容器/最终记录成功状态进入独立完整oracle；R2详情缺口仍声明，不改键。
  所有PASS、exit0/空stderr、退出/双流/pins均精确才进入FINAL。
- [ ] Step2：唯一 Run -Slot final 启动新Integrate宿主和原seam子进程各1次。
  必须原15行/12PASS、完整原oracle零差异，stdout SHA
  F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE、trace SHA
  4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E；
  无主/次错误/overflow/timeout/kill/rescue，原对象退出/双流/正PID/后置pins完整。
- [ ] Step3：仅全部精确后消费唯一fresh只读review。核全部七输入、三处实际接线、
  全部记录/后置输出边界、预消费停线、首因与清理分离、fallback限制、七组记账/
  七组收束/FINAL及五个Review Focus。完整反馈逐项disposition，R1必须明确评估；
  阻塞或审查不完整则停，无自动fix/补跑/二审，R2/Minor2不自动修。
- [ ] Step4：保存完整raw JSON字面数字/流字节/源码快照/清单哈希/预算/审查，
  作者逐项核便携数据与raw；raw仍ignored且保留。更新状态/恢复、显式列表提交
  推送现有分支并核远端等于本地。正式/main/live/fullC12验收仍需独立决定。

## 自检与下一道批准

已自检：R1三处同源出口、首次/后续/溢出/正常/读解析故障都有指定断言；
调用方接线与不可动态保证的出口失效明确留给源码审查，不虚构全路径注入。
流程从预消费到最终发布失败封闭；首因保存在内存不等于磁盘必定可用。
没有新产品层/通用故障框架/能力或旧额度复用，R2和Minor2仍显式延期。

请审阅并批准 Native实施本计划、NEW4宿主+4子进程=8新PS7进程、条件1只读审查。
批准前本轮仅文档；批准也不授权预期外修复、重试或自行正式验收。
