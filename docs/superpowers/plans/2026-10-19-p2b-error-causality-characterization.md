# P2-B Error Causality Characterization Verification Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for proposed Native execution. Track steps with checkboxes. This is a finite diagnostic stage, not controller repair or integration acceptance.

**Goal:** 校准首因判据：区分同一底层I/O失败的不同ErrorRecord包装，与两次内容相同但因果不同的失败。
**Architecture:** 在新独立诊断目录内，用两次真实本地写入失败生成独立错误，再验证仅诊断脚本内部的因果判定函数及拒绝反例。不加载或修改控制器/旧测试/codec/inbox；不创建子进程或通用产品故障接口。
**Tech Stack:** Windows普通PowerShell7.6.5 Core x64、固定SDK、File.WriteAllText；无新依赖、原生编译或安全设置变更。
**Spec:** [原I1修复合同](2026-10-19-p2b-controller-exception-containment.md)中的“首因保留/清理错误分离”，受[握手合同§9](../specs/2026-10-19-p2b-handshake-contract-design.md)的验证隔离约束。本计划只验证判据，不完成上述合同。

状态：CHARACTERIZED_WITHIN_SCOPE_NO_ACCEPTANCE。用户在7e2184f6后批准Native/
新1诊断宿主、0子进程、0独立审查。执行基线7e2184f67182a9759015bc7ccd92bf291e7bcec2。
开发分支codex/c12-b01-task9-coordinator。5输入/1434pins已冻结，
唯一host2844/749ms/exit0，5组精确PASS、空stderr、trace零差异、details有效。
输入完整、ownedexit/drains完成，无子进程/编译/产品导入；额度1/1用尽。
仅判据诊断完成，未关闭I1、未改判旧GREEN；只读核对并保存开发检查点。

## 静态依据、选择与不变结论

旧containment.tests.ps1:65、81用最早Error列表项与返回PrimaryError做
ErrorRecord引用相等判断；owned-run.ps1:28保存catch收到的记录，随后
finally只另记CleanupErrors，源码中没有第二次赋值PrimaryError。

[PowerShell7.6.5 Compiler.cs:4783–4812](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/parser/Compiler.cs#L4783-L4812)
显示单一catch-all使用新ErrorRecord设置当前错误；
[ErrorPackage.cs:1303–1349](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/ErrorPackage.cs#L1303-L1349)
展示复制及WrapException新建记录。由此静态确认：跨这两个来源要求同一
ErrorRecord实例，不是可靠的首因保留判据。它不说明旧进程中的所有对象关系
已经重新测得；旧GREEN失败、FirstRecordIdentity=false和CauseConfirmed=false
等原始字段原样保留，不改判、不补写缺失的CLR观察。

推荐先校准底层异常因果，成功后另定控制器补验计划。直接删掉失败断言会
丢掉首因保障，拒绝；只比较错误字符串会把同路径同内容的两次失败混同，
拒绝；本次不重跑整套控制器、不恢复旧FINAL、不借用冻结余量。

## Global Constraints

- 复用既有隔离Windows/PowerShell开发工作树；不合并main、发布、删除旧文件/
  原始记录、改Git信任/安全设置、安装或联网获取运行依赖。
- 原seam五冻结输入、containment七冻结输入、两阶段源码快照及所有raw/
  portable-summary/审查/失败记录不变；诊断另建目录、packet和预消费账本。
- 固定codec9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C、
  inbox47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC不变。
- engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
  SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/
  RealGateImplemented始终false；异常引用仅同一诊断宿主内有效，不可序列化恢复。
- 不import产品/旧owned-run，不执行旧套件或任何进程创建、Kill、PID/名称查询。
  唯一允许的Process对象由普通父控制器创建，用于监督本次诊断宿主本身。
- 旧containment已用2宿主+2子进程；余1宿主+3子进程+1review继续FROZEN。
  seam2+1/tree5+1已用完、original2spent+3FROZEN/diagnostic1spent均不变。

## 文件图与唯一职责

未来新目录docs/diagnostics/isolated-fuzz/p2b-error-causality-characterization/：

| 五个新输入 | 职责 |
| --- | --- |
| causality.tests.ps1 |两次真实I/O失败、局部判据、5组断言与trace |
| capture.ps1 |普通父上下文Prepare/Run、预消费、确切宿主护栏、捕获与完整预期检查 |
| expected-output.json |5行PASS及独立固定观测预期，不从目标输出生成 |
| README.md / .gitattributes |诊断范围/恢复限制，ps1/json/md物理LF |

派生portable-summary.json及执行说明不是冻结输入。
新ignored packet .superpowers/sdd/2026-10-19-p2b-error-causality-characterization/
记录源码/计划快照、manifest、一次invocation、原双流、trace及有界错误详情。
所有输入/断言先冻结；本计划保存时不得创建这些代码或packet。

## 判据接口与错误样本

只在新测试脚本scope定义，不导出：
C-SameIOCause([object]Left,[object]Right) -> bool；
C-PrimaryPreserved([object]Result,[object]FirstAnchor,[object]FinalAnchor) -> bool。

C-SameIOCause只接受两个真实ErrorRecord。分别从Exception沿InnerException
取最多8个不同引用节点；循环/超过8/空值/没有I/O叶因均false。
允许的I/O叶因仅System.UnauthorizedAccessException或System.IO.IOException
及其派生类；仅比较这个叶因的真实引用身份，不能凭ErrorRecord、包装器、
消息、类型名、HResult、哈希或“存在某个公共祖先”就判true。
一条链含多个符合类型时取最后一个；必须完整走到null且未超限。
该有界规则服务固定本地I/O诊断，不宣称通用PowerShell错误等价算法。

先清空本诊断Error列表，在新packet内新建receipt-is-directory目录，
固定同一Write-Receipt诊断函数两次调用同一路径的File.WriteAllText。
两次都是独立真实方法调用，不throw已有异常、不构造第二个“相同引用”失败。
每次catch直接保存当前记录；在第二次写入前保存第一阶段Error列表锚点，
第二阶段结束另保存最后失败锚点。具体锚点取各阶段结束时Error[0]；
不假定Error列表恰好只有两条。判定旧“最早项”假设的结果另作观测，
不得让其决定哪个真实I/O调用属于首因。

C-PrimaryPreserved只检查普通诊断结果根的PrimaryError与CleanupErrors：
Primary必须与FirstAnchor同I/O叶因、与FinalAnchor不同；CleanupErrors必须
恰好1项，与FinalAnchor同叶因、与FirstAnchor不同。缺字段/类型不符/
未知因果/截断均false。普通结果根仅为断言单元样本，不冒充控制器运行回执。
不把同文本不同失败合并；不弱化成“PrimaryPresent=true就通过”。

## Review Focus

1. catch复制记录被误当作丢失首因：CatchCopyNotIdentity核记录引用不同、同叶因仍匹配。
2. 同路径/同内容两次错误被字符串比较混同：DistinctEqualText核叶因引用不同而判据拒绝。
3. 后续错误替换Primary仍误通过：ValidatorRejectsReplacement核原根true、替换/交换根false。
4. Error列表多包装/索引变化污染锚点：OriginalCauseRetained核两次独立阶段锚点和对应catch，记录来源，不假定全局最早项正确。
5. 缺失、循环或过深链被当成功：MissingOrUnboundedCauseRejected核缺锚点/缺Primary/超过8节点拒绝；循环保护须源码自检，不伪造产品钩子。

## 已批准且用尽：1诊断宿主，0子进程，0独立审查

推荐Native，仅一组紧耦合诊断，不派发代理或追加独立review席位。
现有普通父上下文执行capture Prepare/Run，不再开一个PS7做记账。
Run只Start一个固定PS7 -NoProfile -NonInteractive -File causality.tests.ps1，
无子进程。预消费持久化先于Start，启动失败也用尽；最多120000ms执行+
2000ms确切退出观察+2000ms双流共同drain。父路径异常先收束同一确切宿主，
记录写入失败不能绕过收尾；不扫描PID、不无限重试。
任何异常、unknown、非精确/超时立即停止；无修复/再跑/换宿主/追加审查。
旧冻结余量不释放，不进入控制器GREEN/FINAL。

这是新固定语义表征，首次预期通过；不制造缺API RED，不改产品，也不声称
产品TDD修复。真实I/O反例及显式错误替换样本在同一宿主内检查判据。
用户已明确批准上述有限首绿及0独立审查，覆盖通用技能的额外baseline/
task-done/suite/最终review建议；代价是只有诊断作者自检，没有独立验收。
以后拟采用判据补验控制器，必须另定含运行/审查的有限计划。

## Task1: 冻结独立判据与5组表征，再执行唯一宿主

**Files:** 上述新5输入；旧源码/packet只读。
**Consumes:** 两次实际本地写入失败及独立阶段锚点。
**Produces:** 固定5组observations及完整原始诊断，不改变旧测试。

- [x] **Step1 — 先写固定断言与预期，静态检查后Prepare。**
  五组顺序及PASS名称恰好如下；每组断言为literal true/false，不由判据计算预期：
  CatchCopyNotIdentity：第一catch与第一列表锚点记录引用不同=true，
  两者叶因匹配=true。
  OriginalCauseRetained：第一/第二catch各自对应阶段锚点匹配=true，
  第一catch对第二锚点匹配=false。
  DistinctEqualText：两个I/O叶因消息非空且相等=true，叶因引用不同=true，
  C-SameIOCause(first,second)=false。
  ValidatorRejectsReplacement：真实first主因+second清理的普通根true；
  Primary替换成second、主清理交换、Cleanup替换成first三反例均false。
  MissingOrUnboundedCauseRejected：null锚点、缺Primary、9层RuntimeException
  包装后才到I/O叶因的普通反例均false；不执行控制器或模拟其生命周期。
- [x] **Step2 — Freeze全部1434条映射，不启动。**
  继承旧GREEN清单9AAF1DA8...的1398条VerificationPath及哈希，逐项核对；
  增加旧containment packet递归19个文件、旧portable叶10文件、旧执行报告1、
  新5输入及批准Plan快照1。1434=1398+19+10+1+5+1。
  旧packet/leaf文件数量变化或缺失停止，不能补建。清单role/path/hash展示后
  才冻结；审批快照取用户实际批准的本稿版本，不随进度文档更新而覆盖。
- [x] **Step3 — Run一次，检查完整预期。**
  Run:现有普通父上下文capture.ps1 -Action Run -FrozenSHA256 <Prepare实际SHA>。
  Expected:exit0/空stderr/恰好5行PASS，按上述5组顺序；无Context额外输出。
  trace固定Schema1/Mode=p2b-error-causality-characterization/Groups string[]/
  Observations五组及对应上述literal布尔项/Passed=true/Error=null/
  CompileCount0/ProductChanged=false及全部authority flags false。
  实际类型/有限异常链摘要另保留details，最多两事件、每链8节点、每项
  1024字符；类型须在上述I/O允许集合内，不能将字符串裁剪当匹配依据。
  pre/post1434pins一致、ownedexit/drains确认；无子进程Start尝试。
  不符合任何项停止且保存失败，不按输出校准oracle。

## Task2: 无损保存有限结论与下一道边界

**Consumes:** 同一冻结输入/清单/一次诊断回执。
**Produces:** 可恢复诊断结果，不是控制器修复验收。

- [x] **Step1 — 保存便携结果并只读核对。**
  保留trace/双流字节哈希及准确JSON数字token，不经JS Number回写。
  稀缺CLR引用只能在同一宿主中判断；JSON里的true或数字ID不能恢复引用。
  核全部映射/原始哈希/便携子树，失败也保留、不改旧GREEN。
- [x] **Step2 — 明确完成范围，必要文件提交推送开发分支。**
  更新本Plan/新执行说明/current-status/repository-recovery；
  结果只能称“限定宿主下判据表征”，不称I1关闭/原GREEN通过/
  正常退出或超时路径通过/完整首因保证。
  必要文件逐项暂存，忽略raw/依赖/凭据，remote==local及工作树干净再报告。
  不删除raw/工作树，不合并main。若成功，下一步仅建议另行制定控制器
  断言修订+完整5组/FINAL/独立审查计划，不自动实施该后续阶段。

## 自检与批准交接

范围覆盖错误判据本身的正例/相同文本独立反例/替换清理反例/缺失和深度拒绝；
循环仅代码保护自检，无动态循环输入证明。此处不采用通用GetBaseException
或字符串等价，不修改I1合同的“真实首因保留”目标，只避免跨包装记录引用假设。
本稿是后续方案细分，不修改旧合同/历史失败/冻结清单；旧进程CLR对象关系
无法补测，单次新表征也不宣称已还原其每个引用。
用户已批准本有限诊断；实际1宿主/0子进程/0独立审查，五组符合预期。
批准稿快照866221E9...，manifest6B31CB6A...；原始回执和失败历史保留。
本有限诊断没有进一步目标或审查额度。下一步只建议另定控制器断言修订+
完整5组/FINAL/独立审查计划；不得自动改旧冻结文件、重跑或解冻原额度。
