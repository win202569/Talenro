# P2-B Controller Causality Reverification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for proposed Native execution. Steps use checkboxes. No implementation or execution is authorized by this document alone.

**Goal:** 在不改控制器的前提下修订首因断言，补验完整五组收束行为，再验证一次原握手 seam 集成。
**Architecture:** 新测试只读加载原 GREEN 候选 owned-run 与原子进程 fixture；以已校准的有界 I/O 叶因身份替代跨来源 ErrorRecord 包装器身份比较。新目录、新冻结包、新预消费账本；普通父控制器监督两个宿主，第二宿主仅在首宿主完整通过后调用原 seam 一次。
**Tech Stack:** Windows、普通 PowerShell 7.6.5 Core x64、固定 SDK；无安装、编译、新产品依赖或安全设置变更。
**Spec:** [原 I1 收束修复合同](2026-10-19-p2b-controller-exception-containment.md)、[判据诊断计划](2026-10-19-p2b-error-causality-characterization.md)及[握手合同 §9](../specs/2026-10-19-p2b-handshake-contract-design.md)。本计划不实现真实握手状态或能力。

状态：GREEN_FAILED_FROZEN_UNFINISHED。用户在 62f7ad23 后以“是”批准 Native、
NEW2 宿主+4 子进程+条件1审查及首次 GREEN 判据修订例外。
执行基线 62f7ad23058fc13dbea1387a41cbe6fab5da00b8；原批准快照 SHA
5264D0FBE0A26258E4DDC1F9A54E2C06786E1CFEAD4A141B8458CDDB77BD2FAC 单独保留。
五输入/1454pins 已冻结。GREEN 宿主40880/1556ms/exit2：前四组通过，
PreStartOutputFault 的 Unassociated=false 未达预期；失败原因未确认，不改判。
三子进程确切退出、0外层救援；本轮用1宿主+3子进程，余1宿主+1子进程+1审查冻结。
FINAL/审查未运行，I1 未验收。源码/原始记录不修不重跑，只保存 unfinished 检查点。

## Global Constraints

- 仅现有 codex/c12-b01-task9-coordinator 开发工作树；PowerShell 执行。
  不合并 main、发布、删旧文件、改 Git 信任或安全设置、安装依赖。
- 原 seam、containment、causality 的全部冻结输入、raw、portable-summary、
  审查与失败记录保持不变；旧 GREEN 仍为失败，FirstRecordIdentity=false、
  CauseConfirmed=false 不补写、不改判。诊断 5/5 不等于控制器验收。
- 固定 codec SHA256：9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C。
- 固定 inbox SHA256：47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC。
- 固定 engine SHA256：362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139。
- 固定 SDK SHA256：2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- SourceVerified、LiveActionsPerformed、RuntimeApproved、FormalAcceptance、
  RealGateImplemented 均不升级；原有限离线批准与 Minor2 延后不变。
- 旧 containment 余 1 宿主+3 子进程+1 审查保持 FROZEN；诊断 1/1 已耗尽。
  不借用任何旧额度，不自动重跑 RED、旧 GREEN、旧诊断或旧 FINAL。
- 不修改/复制产品控制器、不增加故障开关、setter、callback；仅新测试私有判据。
  只持有并监督确切 Process 对象，不按 PID/名称重找、查询或终止其他进程。

## 固定复用与文件图

以下相对路径以 docs/diagnostics/isolated-fuzz/ 为根，只读复用：

| 输入 | SHA256 / 用途 |
| --- | --- |
| p2b-controller-exception-containment/owned-run.ps1 | 7A425C661968CD74BFD8CD4EAA25CBA53ED36CC09E3D2946F9D095FC095B25E1；唯一受测控制器候选 |
| p2b-controller-exception-containment/fixture-child.ps1 | EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F；Hold / EmitAndExit |
| p2b-error-causality-characterization/causality.tests.ps1 | 3524ADD68A3C0F487D530F3979CB6721D8E08B840FA80F092A6EFF0F85ACA823；只读提取两个判据定义，不执行该脚本 |
| p2b-handshake-seam-verification/handshake-seam.tests.ps1 | 复用原冻结 manifest 中的精确哈希与原 oracle，不修改或复制目标 |

新目录 docs/diagnostics/isolated-fuzz/p2b-controller-causality-reverification/：

- containment.tests.ps1：五组测试、私有因果断言与三次子进程账本。
- capture.ps1：Prepare / Run / Integrate 三种入口；普通父监督、预消费、完整 oracle 比较。
- expected-output.json：运行前人工固定的 GREEN/FINAL 输出及完整观测预期。
- README.md / .gitattributes：范围、恢复限制、ps1/json/md 物理 LF。

共五个新冻结输入。后续 portable-summary.json、review.md、review-disposition.md
是输出，不列为执行输入。新 ignored packet：
.superpowers/sdd/2026-10-19-p2b-controller-causality-reverification/。
执行报告另建 docs/roadmap/2026-10-19-p2b-controller-causality-reverification.md；
更新 current-status.md 与 repository-recovery.md，必要文件提交现有分支。

## Review Focus

1. 同一底层错误的 catch 包装器不同：第 2 组比较 I/O 叶因，拒绝用包装器身份代替。
2. 两次同路径、同文本错误混同：第 2 组要求独立锚点的叶因不同，消息不得决定通过。
3. 以返回值自证首因、或清理错误替换首因：第 2 组从独立 Error 列表取锚点，核三种替换负例。
4. 额外错误或无界链被静默忽略：第 2 组要求恰好两条独立列表记录；空、缺失、深九层均拒绝。
5. 救援或持久化失败被误报成功：第 1/3/4 组和 FINAL 要求确切退出/完整双流；外层救援即停止。

## 新预算、顺序与停线

拟批准 NEW 2 个顶层 PS7 宿主 + 4 个受控 PS7 子进程 = 6 个新进程，
外加条件性 1 次只读独立审查。普通既有父上下文不另启动 PS7；Prepare 不启动目标。
无实现代理、额外诊断宿主、探针、重试或产品 mutant。

| 顺序 | 宿主槽 | 子进程槽 | 唯一用途 |
| --- | --- | --- | --- |
| 1 | green | fault、normal、timeout | 新 containment.tests.ps1，故障子进程由前两组共享 |
| 2 | final | final-seam | 新 capture.ps1 -Action Integrate，调用原 seam 一次 |
| 3 | 无 | 无 | 前两行完整精确成功后，一次独立只读审查 |

PreStartOutputFault 仅创建未启动的 Process 对象，无 Start 尝试；不得消耗或增加子进程。
每个 Start 尝试之前先持久化消费槽；即使启动失败也已消费，不可回退。
只允许顺序前缀；已有对应槽、账本未知、父记录写失败或 pin 不符，禁止 Start。

每个顶层宿主 MaxMs=120000；超时后只对持有的同一 Process 请求终止，
ExitWaitMs=2000、双流合用 DrainMs=2000。它们是有界观测窗口，不是硬实时保证，
超限如实失败，不延长后重判。受测 helper 的 fault/normal MaxMs=2000、timeout
MaxMs=200；ExitWaitMs/DrainMs 各 2000。fixture Hold 保持原 45000ms 不改。
FINAL seam 使用 MaxMs=110000（不改变原测试脚本）；其外层 final 宿主受
120000ms 总监督，保留 10000ms 给子层有限收束与宿主落盘，不重置父截止。
若子层不能在父界内完成，记录 unknown 并停线；父终止宿主不等于已证明
其子进程退出，不得把宿主退出或 fixture 自行到期当作子进程收束证据。

任何非精确结果、未知所有权/退出、超时宿主、流不完整、意外 stderr、额外错误、
pin 漂移、外层救援、记录/解析失败：先按确切对象完成有限安全收束，再停止后续槽。
保留原始记录，余量标 FROZEN；不自动修代码、修 oracle、重跑或发审查。

## Task 1: 新断言及五组补验（一次 GREEN）

**Files:** 创建新五输入；只读上述固定源，不修改旧目录。
**Interfaces:**

- Invoke-P2BOwnedRun -Process <Diagnostics.Process> -StdoutPath <string>
  -StderrPath <string> -ReceiptPath <string> -MaxMs <int>
  -ExitWaitMs 2000 -DrainMs 2000：保持原签名与返回字段。
- containment.tests.ps1 -TracePath <fixed green.trace.json> -DetailsPath
  <fixed green.details.json>：仅固定新 packet 路径，成功 exit0，失败 exit2。
- 私有 C-SameIOCause([object]Left,[object]Right) 与
  C-PrimaryPreserved([object]Result,[object]FirstAnchor,[object]FinalAnchor)
  原样提取诊断中的函数定义；静态比对函数体，不 import/执行旧诊断目标。
  前者仅接受 ErrorRecord，完整走 Exception.InnerException 到 null，最多 8 个
  不重复节点，取最后一个 UnauthorizedAccessException/IOException 派生异常；
  仅同一叶因引用为 true。后者要求 PSCustomObject/PSNoteProperty、精确一个
  ErrorRecord[] 清理项，首因匹配 First 不匹配 Final，清理反向匹配。

- [ ] Step 1：复核前置证据、五输入文件边界与全部固定哈希。确认仅新 oracle 修订，
  不做新的产品 RED；批准本计划即明确同意“新套件首次运行可 GREEN”的测试判据修订例外。
  历史 RED 只说明已保存的旧缺陷，不冒称本轮重新完成 RED→GREEN。
- [ ] Step 2：先写独立预期及五组断言，再写新入口/账本。第 2 组在 helper 调用前
  Error.Clear()，返回后立即保存该宿主 Error 列表的对象快照，早于其他可能报错操作。
  快照必须恰好两条 ErrorRecord；旧在前、新在后，分别以 oldest、newest 为独立锚点。
  这是固定调用路径的严格预期，不是已测事实：数量不等于 2 即失败，不能当场分组、
  忽略额外项或放宽 oracle。锚点不得从 r.PrimaryError/CleanupErrors 导出。
  固定 helper 的两次 Save-Receipt 顺序为首写失败、finally 末写失败；保持该源码 pin。
- [ ] Step 3：实现下列固定测试；第 2 组所用替换根仅是断言负例，绝不修改实际 r：

| 组名（按顺序） | 必须成立的预期 |
| --- | --- |
| PostStartReceiptFault | 新目录作 ReceiptPath，Hold 子进程；Started/PrimaryPresent/OwnedExit/ActualHasExited/TerminationRequested/StreamsComplete=true；TimedOut/CleanupOverflow=false；无外层救援 |
| PrimarySurvivesFinalReceiptFault | 两条独立锚点，互不同叶因；C-PrimaryPreserved(r,first,final)=true、CleanupCount=1；用 final 替换 primary、交换两者、用 first 替换 cleanup 的三个普通负例根均 false；null 锚点/缺 Primary/九层包装均 false |
| NormalExit | EmitAndExit；exit0、无 primary/cleanup/overflow/timeout/kill，确切退出/双流完成；原 ASCII stdout=P2B-CONTAINMENT-OUT、stderr=P2B-CONTAINMENT-ERR 精确匹配 |
| Timeout | Hold；TimedOut/TerminationRequested/OwnedExit/ActualHasExited/StreamsComplete=true，已知退出码，无 primary/cleanup/overflow；不要求特定终止码 |
| PreStartOutputFault | StdoutPath 指向目录；Started=false、primary 存在，Process 未关联、OwnedExit=false、ExitCode=null、无 kill；该异常不是启动成功 |

- [ ] Step 4：普通父 Prepare。验证 diagnostic frozen.json SHA
  6B31CB6ABDD0C610CF955732207D3DC478B9052418269539181003481411428E，继承
  1434 映射 pin；加入该诊断全部 7 raw、6 portable 文件、执行报告、新五输入和
  本批准计划快照（总 1454 映射项，允许重复路径但禁止矛盾哈希）。输出完整清单后冻结。
  所有新代码、oracle、预算、槽序与计划快照哈希固定；运行前后全清单检查。
  缺旧 raw/固定依赖时停在准备阶段，不从 portable 数据伪造原件。
- [ ] Step 5：仅一次普通父 Run -Slot green -FrozenSHA256 <prepare 输出>。
  它启动上述新测试宿主一次，后者消费三子槽；5 行 PASS 必须与上述组名逐字同序，
  宿主 exit0/空 stderr/无父错误/完整退出双流/0救援，完整 trace 对固定 oracle 零差异。
  动态 PID、时间、终止退出码只在明确列出的 receipts 字段保存与范围核验，
  不从完整固定 trace 比较中临时删字段；details 保留对象关系布尔值与有界错误链。
  不序列化声称恢复 CLR 引用身份。失败执行停线；不得进入 Task 2。
- [ ] Step 6：保存清晰标注的开发检查点及恢复说明；即使 GREEN 通过也只表示本组观察，
  不关闭 I1、不声称整体完成。普通提交不改变冻结执行输入或已消费槽。

## Task 2: 一次 FINAL 集成、条件审查和交接

**Files:** 同一新 capture.ps1/expected-output.json；派生 summary、执行报告与审查记录。
**Interfaces:** capture.ps1 -Action Prepare；-Action Run -Slot green|final
  -FrozenSHA256 <64 hex>；内部 -Action Integrate -FrozenSHA256 <same>。
Run 只从既有普通父调用；Integrate 必须核 final 已预消费且 green 精确通过，禁止任意直跑。

- [ ] Step 1：普通父 Run -Slot final，仅在 Task 1 全部精确通过且所有 pin 完整后，
  预消费 final 宿主槽并启动 capture.ps1 -Action Integrate 一次。Integrate 只读加载
  固定 owned-run，对确切原 seam Process 预消费 final-seam 子槽并调用 helper 一次。
  Run 父使用独立的既有诊断父监督模式，不依赖被测 helper 来保证顶层清理。
  每层在 Start 之后、持久化之前必须进入 finally 安全边界，持有实际启动状态；
  保存失败不能绕过有限退出/双流收束，救援不被归作验证成功。
- [ ] Step 2：FINAL 要求原 15 行/12 PASS、原完整 trace oracle 零差异；stdout SHA
  F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE、trace SHA
  4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E。
  helper 必须无 primary/cleanup/overflow/timeout/kill，exit0、确切退出/双流完整；
  final 宿主同样 exit0、空 stderr、完整父监督、无救援且后置 pins 完整。
  这是一项新明确预算的集成观察，不是自动恢复旧 FINAL。
- [ ] Step 3：仅上述成功后消费 1 次独立只读审查。授权时只派一个 reviewer，
  无执行代理或额外审查；不执行脚本/探针，核全部新输入、原 helper、独立预期、
  锚点独立性/数量拒绝、替换负例、账本/收束与完整结果。报告分开区分 I1 是否有
  足够修复证据、剩余问题和范围排除；完整反馈逐项 disposition，不静默丢弃。
  发现阻塞项则保留未验收状态，不自动修复、复验或二审。
- [ ] Step 4：仅当两次运行及独立审查均满足合同，提交“有限离线控制器补验证据待用户
  验收”的交接；不自行赋予正式/运行/live/main/full-C12 验收。portable summary 保留
  精确源码/计划/manifest/原始流/trace/details/账本哈希及实际计数；raw 仍 ignored。
  明确旧失败未改判、旧额度冻结、新额度已用/未用/冻结数和全部剩余阻塞。
  更新恢复文档、显式文件列表暂存提交推送并核远端 SHA；不提交缓存、凭据、二进制。

## 计划自检与批准门

已对照原 I1 合同和握手 §9 自检：五组、最终完整 seam、独立 oracle、
首因负例、旧证据保护、确切进程所有权、有限时窗、新额度及条件审查均有归属。
新锚点路径尚未执行，恰好两条的预期可能触发停线；不能用诊断直接替它验收。
不新增产品行为，不以合成结果替代真实控制器回执，不扩展真实能力范围。

请审阅并批准：Native 实施本计划；NEW 2 宿主+4 子进程（6 新 PS7 进程），
条件性 1 次独立只读审查；新套件首次 GREEN 的判据修订例外；失败即停、不借旧额度。
用户已批准该范围；本轮 GREEN 第五组失败，停线生效。该批准不授权自动修失败、
重试或正式验收；下一步需要另行决定只读原因分析，再决定修订与新验证范围。
