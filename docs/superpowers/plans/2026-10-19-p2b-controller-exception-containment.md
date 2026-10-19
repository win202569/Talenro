# P2-B Controller Exception Containment Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended Native execution. Subagent-driven execution requires a separately approved replacement budget. Track steps with checkboxes.

**Goal:** 修复只读审查I1涉及的控制器启动后异常收尾，不实现真实握手或改变固定codec/inbox。
**Architecture:** 保留旧冻结目录；在新目录将同一控制器的进程生命周期收束到一个仅供离线控制器使用的私有函数。调用者始终持有原始Process对象，函数负责有界退出/流收尾，调用者最后释放对象；测试调用同一函数，不修改产品、不提供通用故障callback。
**Tech Stack:** Windows PowerShell7.6.5 Core x64、固定.NET Process/Task/文件流，无Add-Type、安装或原生编译。
**Spec:** [已批准握手合同§9的验证隔离](../specs/2026-10-19-p2b-handshake-contract-design.md)；本计划仅落实[审查I1及协调者处置](../../diagnostics/isolated-fuzz/p2b-handshake-seam-verification/review-disposition.md)，不落实整个握手合同。

状态：GREEN_FAILED_FROZEN_UNFINISHED。用户在19e3880d后批准Native及本新预算。
计划基线229dd2a5707a93f09a28300885117783b71050ee，
分支codex/c12-b01-task9-coordinator。执行基线19e3880d5133ecdb75bed9981ceff25ceb0547ce。
已准备7输入/1398pins；RED符合预期并救援；仅owned-run做一次预定修复。
GREEN第二组PrimarySurvivesFinalReceiptFault失败，1/5组通过，原因尚未确认。
已用2/3宿主及2/5子进程；余1宿主/3子进程/1条件审查冻结，0编译/产品改动。
本计划未完成/未验收。完整失败及恢复见同日执行报告，不改判/自动修复或重跑。

## Global Constraints

- 仅现有Windows/PowerShell隔离工作树；不合并main、发布、删文件/分支/原始记录或改安全设置。
- 旧p2b-handshake-seam-verification五冻结输入及全部原始记录、摘要、审查均不改写。
- 旧两次EXPECTED_WITHIN_SCOPE观察有效但任务仍REVIEW_BLOCKED_I1_UNFINISHED；不得回写成异常路径已验证。
- codec9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C、
  inbox47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC不变。
- SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/RealGateImplemented保持false。
- 不重跑旧33/32/12suite；仅预算内final一次引用已冻结12组seam夹具，无新产品接口/回调/setter。
- 普通非管理员PS7；engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139，
  SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- 停止只作用于创建时持有的确切Process对象；不按PID/名称查找，不杀树或无关进程。
- 旧预算：seam2/2+review1/1、tree5/5+review1/1用尽；original2spent+3FROZEN、diagnostic1spent不变。

## 选型与成功标准

这是既有控制器I1的窄修复；原书面计划已获批准，执行在GREEN失败后停止。
推荐新目录+单一生命周期函数：可保留旧证据并让真实失败路径直接覆盖新控制器。
原地修改旧冻结控制器会破坏证据身份，拒绝；仅补finally但无异常路径验证不足，
也不选择第三套产品模拟器或独立重写整个进程框架。

成功仅指：特定普通宿主下，新控制器已覆盖所列异常/正常/超时路径；原始
失败不会被二次持久化失败替换，确切子进程退出得到有界观察。OS调度/强制
终止API失败仍可能导致unknown，不能宣称任何环境下必定清理成功。
真实握手、资源能力、时钟、P2-C与安全前提仍阻塞。

## 文件图与接口

新目录docs/diagnostics/isolated-fuzz/p2b-controller-exception-containment/，仅创建7输入：

| 文件 | 职责 |
| --- | --- |
| owned-run.ps1 |只定义Invoke-P2BOwnedRun，无加载时执行/产品导入/导出模块 |
| capture.ps1 |新Prepare/Run入口、槽预消费/冻结/完整流与旧seam oracle比较 |
| containment.tests.ps1 |同一函数的RED/GREEN真实子进程检查，外层保有原Process |
| fixture-child.ps1 |无产品依赖；EmitAndExit输出固定ASCII后exit0；Hold输出固定ASCII后等待45000ms再exit0 |
| expected-output.json |新5组测试固定断言/输出；final引用旧oracle，不学习目标输出 |
| README.md / .gitattributes |范围、有限恢复与物理LF规则 |

后续派生输出portable-summary.json/review.md/review-disposition.md不作为执行输入。
新ignored packet .superpowers/sdd/2026-10-19-p2b-controller-exception-containment/
保存每次消费/原流/退出/故障/RED及GREEN源码快照；不得复制旧raw进Git。

Invoke-P2BOwnedRun([Diagnostics.Process]Process,[string]StdoutPath,
[string]StderrPath,[string]ReceiptPath,[int]MaxMs=120000,
[int]ExitWaitMs=2000,[int]DrainMs=2000)：
- 接受调用者新建、尚未Start且仅此处Start一次的Process；调用者是唯一最终Dispose方。
- 无任意scriptblock/callback/故障开关；ReceiptPath只是控制器的本地记录路径。
- 返回有序普通结果：Started bool,PrimaryError ErrorRecord或null,
  CleanupErrors ErrorRecord[]（最多8，超出记CleanupOverflow=true）,
  TimedOut/TerminationRequested/OwnedExit/StreamsComplete bool,
  ExitCode nullable Int32,ElapsedMs Int64,CleanupOverflow bool。
- PrimaryError保留首次原始ErrorRecord；各项清理失败单独收集；不得将null退出码变0。
- 在Start前准备路径/流；Start成功立即记Started，再并发排空双流，然后持久化。
  任一post-Start异常进入同一收尾；正常完成不额外Kill。
- 收尾先判断原Process是否退出；仍运行才Kill该对象；最多2000ms观察退出、
  双流共用最多2000ms等待，不逐流累计预算。Kill失败/退出未确认/流未完成均失败或unknown。
- 最终记录写入失败不得遮盖PrimaryError或阻止其余清理；只有调用者完成
  最后独立退出检查和必要救援后才Dispose。无重试Kill循环/无限Wait/旁路无限日志。

测试路径故障用新packet内预建的目录充当ReceiptPath，触发真实文件写入异常；
无权限/ACL更改。双重写入异常用同一目录使最终写入也失败，不添加产品钩子。

## Review Focus

1. Start后首个写盘失败遗漏Kill/Exit：PostStartReceiptFault核原Process真实HasExited、OwnedExit，GREEN外层救援次数必须0。
2. 最终写盘二次失败覆盖首因：PrimarySurvivesFinalReceiptFault核同一首次ErrorRecord身份及有序CleanupErrors。
3. 正常退出/退出竞争误杀或伪报exit0：NormalExit核固定双流/exit0/无Kill，Timeout核Kill请求不等于已观察退出。
4. 预消费或Start前准备失败意外启动：PreStartOutputFault核Started=false及Process未关联；不消耗子进程启动额外额度。
5. 异常排流、退出/流未确认却宣称成功：故障/超时案例核OwnedExit和StreamsComplete；任何unknown保持失败。Kill API本身失败/OS停顿仅源码审查，不能宣称已动态验证。

## 已批准有限预算：3验证宿主+5子进程，共8新PS进程+条件1审查；余量冻结

| 槽 | 顶层普通PS7宿主 | 子进程上限 | 预期 |
| --- | --- | --- | --- |
| RED |1次，只跑PostStartReceiptFault|1个Hold|恰好已知退出断言失败，外层以原Process救援并确认退出 |
| GREEN |1次，固定5组|3个：写盘故障共用1、正常1、超时1|5组PASS，外层救援0 |
| FINAL |1次，新capture调用旧固定seam脚本|1个|旧15行/12PASS和完整trace精确一致，正常退出无Kill |

GREEN的PostStartReceiptFault与PrimarySurvivesFinalReceiptFault使用同一故障子进程/
同一结果；PreStartOutputFault不启动子进程，不通过增加New来补数。
总计顶层3、子进程5；任一启动尝试均预消费，启动失败也算，禁止隐藏额外probe。
这3个顶层目标分别是RED测试脚本、GREEN测试脚本、FINAL新capture脚本；
FINAL的seam进程已计入子进程5，不另开一个验证宿主。现有普通父控制上下文
负责预消费、捕获和每个顶层目标的确切Process护栏；不为记账再启动PS7。
顶层每次最多120000ms+2000退出+2000并发drain；固定fixture子进程
故障/超时MaxMs=2000（Timeout单例MaxMs=200），ExitWait/Drain各2000ms。
Hold自限45000ms，不派生孙进程。RED外层救援只对同一个已持Process，
一次Kill+2000ms退出观察；救援失败立即停止，不能进入GREEN。

只有完整且恰好符合预期的RED允许一次预定修复，冻结GREEN源码/预期后
再跑GREEN/FINAL。RED的非预期失败、GREEN/FINAL任一非精确/unknown立即冻结余槽及审查。
不得把RED发现未知问题解释成任意调试额度；无第四宿主/第六子进程。
最多1名fresh只读reviewer，仅GREEN及FINAL均精确后启动；不允许实施代理。
用户已批准Native；GREEN失败后剩余额度冻结。Subagent-driven不在本8+1内。

## Task1: 同一生命周期函数的受控RED→GREEN修复

**Files:** 新7输入；旧控制器/测试/产品只读。
**Consumes:** I1源路径、固定宿主与同一Process对象。
**Produces:** Invoke-P2BOwnedRun及5组真实子进程测试、完整RED/GREEN证据。

- [x] **Step1 — 写5组断言与独立预期，准备未修复生命周期基线。**
  同时完成新capture及其Prepare/Run final两入口，已调用同一个基线函数；
  其余6输入在RED前完成，之后GREEN只允许修改owned-run，不另改捕获器。
  未修复owned-run只将旧Start/Save/Wait/finally逻辑移入上述接口，外层持有
  Process且统一负责Dispose；保留post-Start异常不收尾的原I1，不能制造缺API RED。
  差异逐项标注：外移Dispose仅供原对象独立观察，不修复子进程存活问题。
  PostStartReceiptFault断言OwnedExit与原对象HasExited均true；旧路径应违反此断言。
  PrimarySurvivesFinalReceiptFault核首次文件写入异常对象不被最终写入异常替换。
  NormalExit核固定stdout/stderr、ExitCode0/无Kill；Timeout核TimedOut、Kill、
  原对象HasExited与OwnedExit；PreStartOutputFault核未Start且无关联对象。
- [x] **Step2 — 静态核对并冻结RED，先预消费再启动唯一RED宿主。**
  先执行Task2 Step1规定的无启动Prepare；该步骤是所有执行的前置条件，
  不是等Task1运行后才创建清单。读取/解析/哈希不import产品。
  Run:固定PS7 -NoProfile -NonInteractive -File containment.tests.ps1 -Phase RED。
  Expected:只有PostStartReceiptFault出现预定退出断言失败；Started=true，
  被测路径未确认退出，外层直接持有同一对象观察未退出并完成唯一救援。
  先原样保存失败，再救援，RED不能被救援后改成PASS。无其他测试/产品import。
- [x] **Step3 — 一次最小修复并冻结GREEN。**
  按接口收尾合同重排try/catch/finally，分离主因/清理错误与记录写入；
  新capture实际调用同一owned-run函数，不能只修测试专用副本。
  保留RED源码和哈希；其余6输入/预期不得从RED输出学习或改写。
- [ ] **Step4 — 唯一GREEN宿主执行固定5组。**
  已执行但未达预期：host20428/824ms/exit2；第一组PASS、第二组身份断言失败。
  子进程47688退出/流完成/救援0；其余三组未运行，不关闭本步骤或I1。
  Run:同一固定PS7 -File containment.tests.ps1 -Phase GREEN。
  Expected:5组精确PASS/exit0/无宿主错误流；3子进程上限、
  故障子进程复用同一结果、救援0；首因身份/清理二次错误完整。
  如失败/不完整停止，不使用FINAL调试，不自动修复。
- [x] **Step5 — 保存修复但未复审检查点（因GREEN失败，仅unfinished保存）。**
  只提交新7输入及必要便携RED/GREEN结果，不声称旧控制器已变更或I1已验收关闭。

## Task2: 新capture的一次集成观察、条件审查与恢复

**Consumes:** Task1同一函数/GREEN冻结输入、旧seam脚本与原oracle。
**Produces:** 一次FINAL观察、唯一完整只读review及所有意见处置、必要开发检查点。

- [x] **Step1 — 前置Prepare无启动冻结完整映射；GREEN后核对应映射。**
  继承旧1378条VerificationPath，不改写遗失文件。新增旧seam原packet的
  frozen/index及full/final各stdout/stderr/trace共8项；旧leaf portable-summary/
  review/review-disposition与执行报告共4项；新7输入与批准Plan快照1项。
  总1398条（1378+8+4+7+1），列出role/path/hash后冻结。
  RED/GREEN差异仅owned-run及其显式源码快照，分别完整封存清单；
  RED清单永不覆盖；GREEN清单保留同1398条身份，仅更新获准修复的
  owned-run哈希，其他1397条必须逐项相同，不能借修复重新学习oracle。
  FINAL只用GREEN清单，禁止再改输入。任一缺失或数量不符停止，不重建历史raw。
- [ ] **Step2 — 预消费FINAL后新capture调用固定旧脚本一次。**
  GREEN失败，未消费/未运行，冻结。
  Run:固定PS7 -File 新capture.ps1 -Action Run -Slot final -FrozenSHA256 <实际GREEN清单SHA>。
  Expected:同旧15行/12PASS/exit0/空stderr及完整原oracle；无Kill、OwnedExit/
  StreamsComplete/postpins true，stdout及trace哈希分别F693A64A.../4751D03D...。
  不要求PID/耗时等动态宿主元数据相等，不声称再次两轮验证。
- [ ] **Step3 — 仅完整GREEN/FINAL后，消费唯一fresh只读review。**
  条件不成立，未消费/未派发，冻结。
  新7输入/RED→GREEN差异/3槽证据及5个Review Focus；禁止执行/import/compile/
  编辑/index/HEAD/网络/安全查询/嵌套代理。历史只读必要边界，否则hash-only。
  Important/Critical/不完整则停，无自动fix/补测/换reviewer。
- [x] **Step4 — 保存明确未完成开发检查点（无审查可保留，不表示Task2完成）。**
  portable-summary保留raw JSON数字token，逐项与原始记录核对；不经JS Number回写。
  更新新执行报告、current-status、repository-recovery、本Plan实际状态，
  明确新预算消耗/未测项/旧I1记录原样保留。源代码/便携夹具不能依赖私有raw目录，
  但重建完整历史验证所需raw/宿主/权限不因clone自动恢复。
  必要文件明确暂存/提交/推送原开发分支，确认remote==local且工作树干净。
  无main合并/删除/正式验收/live权限；任何失败只保存unfinished。

## 自检与交接

覆盖：I1主路径/二次持久化/正常退出/超时/启动前错误均有具名断言；
真实Kill失败、Start成功后API不可观察、OS调度停顿及并发/恶意篡改不作动态
保证，保持unknown并纳入只读审查；不把请求终止等同实际退出。
生命周期函数与新capture用同一实现；测试没有伪造产品权限或第三握手模拟器。
预算明确计入每个宿主/子进程及唯一审查，只有已知RED允许一次预定修复；
generic额外baseline/task-done/suite/fix pass/deletion不得借机增加。
本合同其他真实握手条款不在本计划内，仍未实现/未验收。

用户已批准Native/8新进程+条件1审查；实际4进程，余4进程及审查冻结。
下一步建议仅另行授权失败身份断言的只读分析；不得据本失败自行修夹具/
改判、继续FINAL或复审。任何改动/新验证需新的明确范围及有限预算。
