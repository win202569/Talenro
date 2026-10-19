# Offline Diagnostic Entry Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Recommend Native/inline here; the proposed allowance contains only one final fresh readonly reviewer, not per-task agents/reviews.

**Goal:** 为未来固定离线诊断建立实际接到私有 Session/Ack 与失败出口管理的新操作入口，保留所有历史源码、证据及耗尽/冻结额度。

**Architecture:** 新 scripts 入口缺省只读 Inspect；RunApproved 只选择本计划的三个固定新 profile。迁移机制到新的私有目录，实际执行测试程序仍是冻结的离线消费者；不转发历史 driver/capture，不从文件恢复确认。原始传输层只监督它直接拥有的宿主，固定执行器监督目标。

**Tech Stack:** Windows、PowerShell7.6.5 Core x64、.NET Process/原始字节流；不新增依赖或独立编译工具。

**Spec:** [已批准书面设计](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md)，批准对象在7e375b8af7b09c85d63ee10152c2d40ded73a86b，SHA256 DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A。用户此后明确回复“批准”。规范原 pending 头保留为历史，本计划记录实际批准，不修改规范字节。

状态：PLAN_PENDING_USER_REVIEW_AND_EXECUTION_APPROVAL。本轮仅计划，0实现/导入/测试/目标/编译/代理/审查。规范批准仅允许编写计划；下面 NEW6宿主+7子进程和条件1审查是申请，不是现有额度。

## Global Constraints

- “只涉及离线诊断，不做生产接入”；不改 api/cmd/internal、CI、Docker 或现有 verify/run-c12-integration 入口，不合并 main。
- “默认为 Inspect”；不导入执行模块、不创建运行目录或 Session、不启动目标、不安装/下载依赖；资料不授予运行权。
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”；只用单个存活运行宿主的私有登记。
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”；清理不重试失败出口。
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”。
- “唯一受支持入口”不等于“仓库中其他 PowerShell 文件不可执行”；不伪装成 OS 层禁用历史脚本。
- “正负对照除提交的真实/缺失确认之外，其他相关前提保持可解释”；不能用 Budget0/Steps 空冒充 C1 的独立证据。
- 历史 leaf/raw/portable/批准/审查不动；C1/C2 open、R2/Minor2延期、五项能力标记 false；不从便携记录重建原件或授权。
- 全程手动有限 PowerShell。未来批准本计划时，一并批准此预算对技能默认额外 Bash/setup/full-suite/task-done/修复轮/二审/删除的限制；无额外重跑或自动补救。

## Review Focus

1. Inspect/参数拒绝若间接触发模块导入、目录或目标创建，默认只读就失效；Task1 的默认/显式/未知参数实际调用及 Task3 的实现存在状态回归固定此边界。
2. 同一宿主重复导入或重复 RunApproved 不能重置 Session、消耗标记或出口状态；Task2 的实际入口二次调用拒绝，Task3 真实续接/重放检查固定正常边界，恶意 Force 重载不作安全沙箱承诺。
3. 缺确认拒绝不能被空步骤/零额度混淆；Task3 增加 AvailableNextStepWithCandidateLedger，合法 Active/步骤/余量相同，缺 Ack 拒绝而真实 Ack 后续可用。
4. 新路径适配不能遗漏 core 预消费失败或把 receipt/open/output 预期负例当全局成功；Task2 真实 Start 写失败，Task3 三种真实故障及 no-retry/首因/清理完整性固定已覆盖路径。动态 classifier 自身异常/晚期关闭并发错误仍需静态核对和明确限制。
5. 便携文件、目录存在或 final 写入不明不能变成下一轮授权；Task1 的 Inspect 历史资料输出、Task3 的 live 文件负例/关闭会话检查及父传输完整签名固定可测边界。实际晚写/宿主死亡/所有出口失效不声称已动态诱发。

---

## 1. 文件、责任和不可变输入

执行工作树继续使用现有 codex/c12-b01-task9-coordinator，先只读确认 linked worktree/分支/clean 状态。无新 worktree、安装或宽泛 baseline suite。

以下20个映射是新输入集合；未实现文件以 **Absent** 记录，不创建假文件。bootstrap 在首轮前写完，测试/oracle/driver之后不可变：

| 类别 | 文件 | 唯一职责 |
| --- | --- | --- |
| bootstrap | scripts/.gitattributes | 只为 /diagnose-offline.ps1 和 /private/offline-diagnostics/** 指定 text eol=lf；不改其他脚本行尾 |
| bootstrap | docs/diagnostics/isolated-fuzz/p2b-entry-migration/.gitattributes | 本 leaf 全部 text eol=lf |
| bootstrap | 同 leaf/README.md | 固定新预算、入口、运行限制、RED归属和恢复规则 |
| bootstrap | 同 leaf/expected-output.json | 所有六轮完整字面 stdout/stderr、顺序 groups/facts/根字段、profile/case表和固定哈希 |
| bootstrap | 同 leaf/test-helpers.ps1 | LS-Need/New/Path/Details/Result/Step/Write/Read/Complete/Output 的固定测试消费者，先采集实际返回再断言 |
| bootstrap | 同 leaf/entry.tests.ps1 | 真实公开入口 Inspect/拒绝/调用来源/二次 Run 检查 |
| bootstrap | 同 leaf/core.tests.ps1 | 七组实际 Session/出口/错误保留用例，不含测试工具定义 |
| bootstrap | 同 leaf/prefix.tests.ps1 | 七组98个字面字节条件；模块内直接调用纯 byte predicate，不暴露内部模块入口 |
| bootstrap | 同 leaf/owned.tests.ps1 | 零启动 prestart FIRST，然后 receipt/normal/timeout，共3实际目标 |
| bootstrap | 同 leaf/live.tests.ps1 | 真实新入口存活 A/B/C 序列、非混淆文件负例、原 seam；共4实际目标 |
| bootstrap | 同 leaf/capture.ps1 | 一个固定验证宿主中调用真实新入口，收束后单次证据写入；不另起入口进程 |
| bootstrap | 同 leaf/driver.ps1 | 原始双流父传输、静态冻结/新增窗口、独立完整证据判定、单次 intent/final |
| Task1新增 | scripts/diagnose-offline.ps1 | 只接受 Action/Profile，先参数拒绝；缺省 Inspect；合法 RunApproved 才惰性导入新的私有模块 |
| Task1新增 | scripts/private/offline-diagnostics/catalog.json | 纯描述数据：三个固定 profile、依赖/历史资料路径；不是授权清单 |
| Task2新增 | 同私有目录/session.psm1 | 私有 Session/Ack/一次调用状态与唯一导出 Invoke-ODProfile |
| Task2新增 | 同私有目录/sinks.ps1 | 全部受管路径状态和 IO 网关，保留实际 ErrorRecord |
| Task2新增 | 同私有目录/owned-step.ps1 | 固定目标、关联、单消费者排空、确切对象清理；不调用旧 owned-run |
| Task2新增 | 同私有目录/sequence.ps1 | 步骤验证/预消费/完成顺序/字节及 seam 验证/私有 Ack 登记 |
| Task2新增 | 同私有目录/profiles.ps1 | 固定 profile 到冻结测试程序的映射，关闭全部 Session 后返回报告；无调用者 callback |
| Task2新增 | 同私有目录/observations.ps1 | 有界实际观察、步骤错误分类/首因保留、只读报告；不发 Ack、不写最终捕获证据 |

执行 packet：.superpowers/sdd/2026-10-19-offline-diagnostic-entry-migration/。
运行源报告：docs/roadmap/2026-10-19-offline-diagnostic-entry-migration.md。
操作者说明：docs/runbooks/offline-diagnostics.md。交接另更新 current-status.md 和 repository-recovery.md。portable-summary.json 仅最后加入新 leaf，不是运行输入。

不得直接套旧 module/driver。迁移来源是 p2b-prestart-cause-repair 的 session/sinks/owned-step/sequence 和 capture 的观察边界，重新绑定新根/新profile/冻结清单；去掉旧 generation/Grant 分支，只保留所列新配置。测试函数先作为 bootstrap 冻结；profiles.ps1 只固定分发，不能在 Task2 追加或重写测试。

根路径必须按新归属计算：公开入口取 scripts 的父目录；私有 session.psm1 从 scripts/private/offline-diagnostics 上溯 ../../..；capture/driver 从新leaf上溯 ../../../..。fixture、关联观察器、原seam和新测试leaf均由 repo＋固定相对路径定位，不沿用旧module相邻目录表达式或调用者cwd。所有新输入物理LF，scripts局部属性规则不得改变现有验证脚本行尾。

## 2. 接口和数据合同

公开脚本：[CmdletBinding(PositionalBinding=$false)]，仅 param([string]Action='Inspect',[string]Profile='')；Action 只允许精确大小写 Inspect/RunApproved，Profile 为三个精确 Id 或 Inspect 下空值。无额外脚本参数/位置参数/旧参数别名；标准 PowerShell CommonParameters 不作为运行许可。

Inspect 返回单个有序对象，根字段固定为 Schema/Action/Profiles/Dependencies/History/Authority/SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/RealGateImplemented。Schema=1、Action=Inspect、Authority=None、五flags均false；无 CanRun/剩余额度/Session/Ack。Profiles固定顺序：entry-migration-core(0)、entry-migration-owned(3)、entry-migration-live(4)。Dependencies每项有 Role/Path/Exists/SHA256/MatchesPinned，缺失SHA为null；History只描述固定历史路径与 consumed/evidence-only，不解析成功字段来授权。

Inspect 只读 catalog/依赖；六份运行实现未建立时须如实列 missing，不因此导入/创建模块。记录实现依赖缺失数组和冻结清单是否存在；Task1预期缺6份、无 runtime-green 清单，Task3预期缺0份、清单存在。输出模板用首轮前固定的 {Repo}/{Engine} 路径替换，不从实测回填。

RunApproved三个 profile 的字面选择写在公开脚本中，catalog修改不能添加运行分支。缺模块实际抛出 RuntimeException，Message/FQID均为 OD: backend unavailable；未知 profile/动作/旧参数在导入前拒绝。不得 exit 整个验证宿主；脚本用正常返回/实际抛出，让 capture 获取原始错误。

唯一模块导出：Invoke-ODProfile([string]Profile) → 单个有序 ProfileReport。根字段：Schema/Profile/EntryPath/EntrySHA256/KernelPath/KernelSHA256/FrozenSHA256/Status/Groups/Observations/PrimaryError/SecondaryErrors/Overflow/Closed/五falseflags。Status为COMPLETE或INCOMPLETE；不返回不透明Session/Ack。模块记录并核对实际加载路径/源码哈希；父传输另核冻结身份，不只信自报字段。

每个宿主仅允许首次 Invoke-ODProfile；第二次在文件IO/目标创建前实际抛出 OD: profile already invoked。公开脚本不 Force 导入或 Remove-Module；重入拒绝不能重置私有登记。Inspect不占用此运行调用位。

Inspect在所列PS7运行时内保证只读依赖检查，不宣称另有Windows PowerShell5.1实现。若新电脑尚无此解释器，先读仓库静态恢复说明；本计划不自动安装或承诺“缺少解释器仍执行脚本”。

Profile内部保留 LS 接口签名以减少无关改名：New-LSSession(Profile,Case,FrozenSHA256)、Test-LSAdvance(Session,Ordinal,PreviousAck)、Invoke-LSStep(Session,Ordinal,PreviousAck)、Get-LSView(Session)、Close-LSSession(Session,Finish|Stop)，以及既有四种网关/失败函数；全部私有、不导出。Step根仍为 Accepted/Rejection/Ack/Observation/Error；IORoot仍为 Ok/Value/Error；引用身份判断在内存完成，序列化仅保留不含Ack的投影。

模块固定读取 packet/runtime-green.frozen.json 并绑定其 SHA；调用方不能传新清单/路径来授予运行权。当前独立批准与父传输的已知冻结哈希仍是执行前提，磁盘文件本身不是许可。每个 profile 输出固定到 packet/profiles/<Id>/，先预消费/CreateNew profile.intent.json；已有/未知即拒绝，无自动RunId、删除或换目录。父宿主独立 intent/final 不构成下一 Grant 的门。

Core运行 Invoke-LSPrefixTests→Invoke-LSCoreTests；Owned运行 Invoke-LSOwnedTests；Live运行 Invoke-LSLiveTests。定义分别来自冻结的新测试文件，经固定路径载入私有模块，不提供执行任意源文件的入口。test-helpers先定义且同一宿主采集当前Profile/Case实际步骤/视图，错误分类在调用方断言之前。

capture沿用有序证据根 Schema/Grant/Status/Groups/CurrentGroup/NotRun/Observations/PrimaryError/SecondaryErrors/Overflow/Probe/五falseflags，Schema=1，Probe恒null；Status在RED未运行组存在时为INCOMPLETE，健康GREEN为COMPLETE。公开Run返回的ProfileReport只进入一次 Name=ProfileReturn 的观察；其Groups按实际输出顺序合入宿主Groups，再记录绑定/重入组。父传输仅从该唯一ProfileReturn.Observations中 Name=Step 的Started=true根计目标，不能重复统计summary中的同一观察。Inspect观察不被计作目标。

oracle字面冻结所有确定字段和顺序；仅PID/ElapsedMs/实际源码位置允许首轮前列明的动态规则：PID正整数且实际目标互异，ElapsedMs非负且满足固定时限/清理元组，错误位置/栈绑定冻结源文件和有效行号。实际ErrorRecord的类型/id/消息和关联事实必须固定匹配，不允许任意异常文本；源码哈希从已绑定manifest取预期，不能从自报来源回填。完整字节比较只允许固定路径模板和既定CRLF/LF差异。

## 3. 新预算、冻结及失败规则（待批准）

| 顺序 | 新Grant | 真实公开动作 | 宿主 | 实际子进程上限 | 完整预期 |
| --- | --- | --- | ---: | ---: | --- |
| 1 | entry-red | 新脚本尚不存在 |1|0|exit1；唯一 FAIL: EntryPresent，其他5入口组NotRun；未调用历史脚本 |
| 2 | entry-green | 缺省/显式Inspect及拒绝向量 |1|0|exit0；6入口组PASS；运行模块实际缺6份/未导入 |
| 3 | runtime-red | RunApproved entry-migration-core，模块尚不存在 |1|0|exit1；FAIL: RuntimeProfileBound；实际 OD: backend unavailable 被记录；14组与重入组NotRun |
| 4 | runtime-green | RunApproved entry-migration-core |1|0|exit0；7prefix+7core+RuntimeProfileBound+ProfileReentryRejected，共16PASS |
| 5 | lifecycle | Inspect/拒绝回归→RunApproved entry-migration-owned |1|3|exit0；6入口+6owned+2绑定/重入，共14PASS |
| 6 | continuation | Inspect/拒绝回归→RunApproved entry-migration-live |1|4|exit0；6入口+5live+2绑定/重入，共13PASS，另原12seamPASS |

NEW6宿主+7子进程=13个新受控PS7进程，全部精确才允许1次fresh readonly review。Native工具宿主作固定父传输，不再通过 &pwsh 启动driver/入口；capture宿主内以 & scripts/diagnose-offline.ps1 实际调用，不另起PS包装层。不能为测试未知参数、inspect、FINAL或工具自检偷偷增加宿主。

每个非continuation宿主上限120000ms，continuation150000ms；内部全序列140000ms/保留4000ms清理；确切对象退出等待2000ms，两流共享排空2000ms。Hold45000ms，normal/receipt目标2000ms，timeout200ms，原seam110000ms。normal两流各恰19ASCII字节，timeout只可接受其字面前缀（含真正byte[0]）。PS7固定engine SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；SMA SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B；非管理员固定当前上下文。不发现则停止，不安装/更改路径或放宽版本。

engine固定路径C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，SMA为同目录System.Management.Automation.dll；上下文为DESKTOP-H4SVCLU\Lenovo、admin=False。固定只读依赖：docs/diagnostics/isolated-fuzz/p2b-controller-exception-containment/fixture-child.ps1，SHA EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F；p2b-process-association-guard-repair/owned-process-observation.ps1，SHA CD6E9564E77FF55812F9E96291B0F4E511758245A9BBC53AD2802B95E809C6E5；原p2b-handshake-seam-verification脚本/oracle及codec/inbox身份继续由继承清单核验，不能复制成可变替代品。

原fixture/关联观察器/seam继续只读固定依赖；不新增独立编译或目标，原seam内部已有行为仍属于同一获准目标。Session≤32、sink≤32/Session、Ack≤2/Session、观察≤32、不同次因≤8、第9项Overflow。证据根/顺序和全部值在expected-output首轮前固定，原seam stdout F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE、trace4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E。

冻结共2064映射：继承当前1930＋旧packet92＋旧leaf16＋最新执行报告1＋closure-map1＋已批准设计原件1＋本计划原件1＋两份批准快照2＋新20输入。相同路径重复映射仅允许同hash。继承原件不齐或冲突即停止，不从portable恢复。Schema2的Pin增加ExpectedState=Present|Absent；Absent必须不存在且SHA为null，不能忽略。初始2056Present/8Absent；Task1后2058/6；Task2后2064/0。

三次冻结：entry-red(12源码快照)、entry-green(14)、runtime-green(20)，共46真实源码快照；不存在的文件只记Absent，不伪造快照。runtime-red复用entry-green；后两轮复用runtime-green。第一次新增窗口恰公开脚本/catalog两项，第二次恰六份运行实现；不允许改bootstrap、测试、oracle/driver或其他已存在源码。

批准设计/计划原件和两份批准快照均保持字节不变。执行期间不勾改本计划原件；checkbox完成状态、实际清单SHA/提交/预算/ruling写入本计划专属progress.md，避免进度更新造成冻结输入漂移。

每轮父传输先完整校验/消费一次intent，持有确切capture Process及单消费者原始双流；实际退出/排空/全量后置pin/完整输出/证据类型、顺序、值全部匹配才精确。一次serialization/CreateNew/write/flush/close证据，不重试/替代路径；final写入失败即unknown。entry-red/runtime-red仅接受上述全部固定签名，不接受任意exit1/缺文件。其他缺原件、漂移、未知、错误签名、控制器/传输/记录失败或非精确，立即冻结剩余全部Grant和review；不自动修复、返还额度、重跑、再冻结、二审或重建记录。

## Task 1: 可检查、不可意外运行的公开入口

**Files:** 首轮前建立12bootstrap；仅在RED后新增 scripts/diagnose-offline.ps1 和私有catalog.json。其余六实现保持Absent。
**Interfaces — Consumes:** 已批准设计、固定原依赖、冻结catalog描述合同。**Produces:** Inspect有序根及合法RunApproved惰性分支；给Task2实际缺模块错误，不是假成功。

- [ ] **Step1：先写entry.tests和全部后续测试/oracle/driver。** 定义 Invoke-ODEntryContract([string]Grant)；EntryPresent先判断实际文件，RED缺失则记FAIL/其他5NotRun并返回，不调用不存在文件或旧入口。GREEN的固定6组：EntryPresent、InspectDefaultReadOnly、InspectExplicitReadOnly、UnsupportedRequestNoRuntime、InspectDependencyStateReadOnly、InspectHistoryDataOnly。
  断言示例：`$before=Get-Module -All | Where-Object Path -eq $newKernel; $r=& $entry; LS-Need ($r.Action -ceq 'Inspect' -and $r.Authority -ceq 'None') 'default'; LS-Need ($null -eq $before -and $null -eq (Get-Module -All | Where-Object Path -eq $newKernel)) 'no runtime import'`。比较固定profiles子树的前/后目录名、文件名/长度/内容hash快照相等，不比较访问时间；本轮三个profile根全部不存在，亦不产生运行intent/Session。Task3允许之前已完成的core/owned根存在但字节不变：lifecycle前只有core存在，continuation前只有core/owned存在，各轮当前运行根仍不存在。不能只信自报TargetsStarted字段。Profile列表/falseflags/根顺序和依赖缺失数组按字面oracle核。
  13拒绝向量逐项实际调用并抓原始ErrorRecord：Action=Run、旧Profile=classification-red，以及Grant/Slot/Resume/Retry/Force/SkipValidation/ScriptPath/Command/Arguments/OutputRoot/RunId参数。全部导入前拒绝；位置参数拒绝另在同组核，总Rejected=14，不增加目标。
- [ ] **Step2：只做静态LF/AST/JSON检查，Prepare首冻并Run entry-red一次。** `& <new-leaf>/driver.ps1 -Action Prepare -Grant entry-red`；再以返回SHA显式调用 `-Action RunOne -Grant entry-red -FrozenSHA256 <本次实际SHA>`。Expected固定唯一FAIL/NotRun5、exit1、0子进程、2056Present/8Absent全匹配，无其他错误；实际SHA即时写ledger，不预填假值。
- [ ] **Step3：实现公开脚本和纯catalog。** 按§2；参数先决策，Inspect不触及Import或运行准备；RunApproved仅固定三profile，六运行实现不存在则抛指定真实错误。只读依赖缺失如实报告；不增加隐式宿主、许可参数或旧分支。
- [ ] **Step4：Prepare entry-green，Expected恰2个Absent→Present，其他2062映射不变；Run entry-green一次。** Expected6PASS/exit0/0子进程/空stderr，全部实际模块和输出目录观察一致，完整冻结/证据精确。此后入口/catalog不可再改。
- [ ] **Step5：保存Task1检查点。** 显式源码/测试/批准资料＋报告/状态/恢复，diff检查后提交现有dev分支并按长期授权推送/核远端相等。记录新spent2宿主/0子进程，剩4宿主/7子进程/条件1review；不声称RunApproved已可运行。

## Task 2: 新私有实现及真实公开core消费

**Files:** 仅新增六运行实现；不改Task1/冻结测试。**Interfaces — Consumes:** 公开RunApproved/实际模块缺失错误；固定新profile、LS私有接口和ProfileReport。**Produces:** 三profile真实运行分发，core16组结果，以及Task3所需的实际执行/完成控制链。

- [ ] **Step1：读取已冻结core/prefix/entry测试brief，确认下面断言已在首轮前写完。** capture使用公开脚本而非直接模块导出；源模块身份从Get-Module -All和实际路径/hash交叉核；Run后再次调用同一公开RunApproved必须得到指定OD重入错误，Start/IO无新增。
  七core组为 DiskCannotMintAck、FailedSinkRejectsAccess、InvalidLedgerStops、FirstAndSecondaryBounded、OutputFailureNoAck、PreconsumeFailureNoStart、HandleCopiesRejected；七prefix/98字面条件沿用原独立向量，但直接私有调用byte predicate。旧disk空步骤组仅辅助，不当新C1独立证明。
- [ ] **Step2：Run runtime-red一次，复用entry-green清单。** Expected实际公开入口抛 OD: backend unavailable，原始Type/Message/FQID和模块缺失事实保留；唯一FAIL RuntimeProfileBound，其余14组/重入组NotRun；exit1/0子进程。不得将其他配置/加载错误当同一个RED。
- [ ] **Step3：实现六文件及唯一导出。** 重绑定新Profile/packet/root/manifest；不Import旧module、不调用旧driver或owned-run；复用只读关联观察器及原fixture/seam。profiles只按固定程序选择Core/Owned/Live，运行一次位和profile.intent预消费不可重置，finally关闭全部Session；所有必要完成动作先于私有Ack。
  observations迁移已验证的实际Step→观察→分类→调用方断言顺序。固定期待Step故障共4种：core/preconsume的Start Failed/UnauthorizedAccessException、Observation=null/Reserved1/Start0/StartIO1/OutIO0；owned/prestart的Out0 Failed/Start0/Unassociated/Kill0/Wait0；owned/receipt-fault的Receipt0 Failed/1尝试/实际退出排空；live/output-fault的ConsoleOut Failed/ObjectDisposedException/目标正常退出排空。匹配实际ErrorRecord类型/id/Session primary引用、固定Case/Profile、唯一failed sink及生命周期；后3种Start sink须Healthy。其他Step错误在断言前提升真实首因，分类异常也先留实际步骤原因。无“expected=true”外部参数、模拟ErrorRecord或生命周期setter。
  core网关自身实际首因/8次因、Failed后LogicalIO0仍按独立固定测试核对；这不等于穷举所有宿主级错误传播或动态classifier异常，保留审查限制。
- [ ] **Step4：Prepare runtime-green，Expected恰6个Absent→Present、2058其他映射不变；Run runtime-green一次。** Expectedprefix7→core7→RuntimeProfileBound→ProfileReentryRejected共16PASS、exit0/空stderr、0子进程、所有2064Present验证精确；冻结后不能改实现。
- [ ] **Step5：保存Task2检查点并提交推送/核相等。** 记录新spent4宿主/0子进程，剩2宿主/7子进程/条件1review；源调用链已连通不等于生命周期/新C1/C2边界通过。bootstrap首次组装只由缺入口/缺模块RED支持，不冒称每个迁移内部分支均有独立RED。

## Task 3: 新入口生命周期、非混淆确认、原seam、审查与交接

**Files:** 所有运行源码/测试/driver不再改变；只产生固定输出和交接文档。**Interfaces — Consumes:** Task2完成的真实公开profile链、实际ErrorRecord/OwnedObservation、私有Ack及runtime-green冻结身份。**Produces:** lifecycle/continuation完整实际入口证据、唯一审查/处置、portable及新操作者说明。

- [ ] **Step1：Run lifecycle一次，继续复用runtime-green清单。** 先跑6入口组（此时代码存在、实现missing数组为空），再经真实RunApproved进入owned程序；prestart零启动必须FIRST，随后receipt/normal/timeout恰3目标。完整6owned顺序与事实沿用最新已修版本，末尾RuntimeProfileBound/ProfileReentryRejected；Expected总14PASS/exit0，3关联/退出/双流收束、无外层终止/救援，失败receipt1尝试/0后续逻辑IO/首因同引用。入口Inspect不得占用profile调用位或创建运行目录。
- [ ] **Step2：Run continuation唯一一次。** 6入口组后，公开RunApproved创建A=[EmitAndExit,Seam]/预算2、B同固定步骤但预算1、C=[EmitAndExit]/预算1。A0、B0真实运行，各产生不同私有Ack。C在完成输出前置换实际closed StreamWriter、finally恢复；正常目标退出/排空之后真实ConsoleOut失败，无Ack、下一步拒绝。B固定预算不启动第二个目标。
  **新AvailableNextStepWithCandidateLedger断言代码必须已首轮冻结：** `LS-Need ($av.Status -ceq 'Active' -and $av.NextOrdinal -eq 1 -and $av.Reserved -eq 1 -and (Test-LSAdvance $a 1 $a0.Ack)) 'available real next step'; $before=(Get-LSView $a).StartCalls; $bad=LS-Step $a 1 $null; LS-Need (-not $bad.Accepted -and (Get-LSView $a).StartCalls -eq $before -and (Test-LSAdvance $a 1 $a0.Ack)) 'candidate ledger cannot substitute ack'`。A0真实正常保存的Ledger含CANDIDATE_COMPLETED；按只读字节前后比相等，不能改写受管台账来伪造晚写失败。案例profile字面预算2/步骤2/余量1独立核；现有Get-LSView无Budget字段，不凭不存在属性断言。
  再核缺失/复制/序列化摘要/跨B确认和错误ordinal共6原无效请求（已先测null可重复拒绝、不创建目标），真实A0确认仍可用；A1仅消费一次实际运行原seam，重放拒绝，不产生第五个子进程。A/B/C全部关闭，fresh磁盘摘要不恢复Session。五live最终组顺序：AvailableNextStepWithCandidateLedger、LiveAckBindingsAndReplay、OutputFailureNoContinuation、PositiveOwnedContinuation、SessionClosedNoResume，之后2绑定/重入组；总13PASS另原12seamPASS。原始seam子输出/完整trace/hash不变；组合宿主stdout顺序全部在首轮oracle固定，不能实测后回填或另起FINAL。
- [ ] **Step3：仅六宿主/七目标/全部记录和输入精确，消费一次fresh readonly review。** 先提交新入口完整源码检查点；最强当前可用模型/独立上下文审查整个新入口链、20输入/46快照/完整原件、批准设计/计划、所有ledger rulings和五Review Focus。只读，不导入/测试/网络/改文件或Git状态，不再派代理；0Critical/Important且审查完整才可提窄结论。发现重要项或不完整就停，保存未验收状态，无代码fix/额外测试/二审。
- [ ] **Step4：逐项保存审查及所有Declined-to-judge处置。** 即使干净，也只能说新离线入口对应边界获支持，不能回填旧C1/C2报告或升级产品/正式验收。晚写/flush/close、宿主死亡、所有出口失败、动态classifier异常等未诱发路径明确列出；R2/Minor2仍延期。
- [ ] **Step5：写操作者说明并制作portable。** 说明默认Inspect、新profile须新批准、历史命令只作证据、无旧参数兼容、旧脚本未被OS禁用、换机无自动续跑。保存全部现存packet原始JSON文本和其他字节/base64，数字不经JS重发；所有46源码快照/manifest/批准/未知/审查/目录也包含。直接赋值保留byte[0]，逐字节/长度/hash核全部原件与暂存blob；不重写原件或重建缺失记录。后续export/报告/状态/恢复仅作者核对，不冒充已独立审查。
- [ ] **Step6：更新execution/current-status/recovery、显式文件清单提交并推送既有dev分支，核远端SHA等于本地和工作树状态。** 不提交ignored raw/缓存/工具/密钥/Docker数据，不删除raw/worktree，不合并main。任何失败只保存明确未验收检查点，不以commit/push成功代替运行验收。

## 4. 作者计划自检及执行选择

规范§1–3目的/选型/归属→§1文件表与Task1–2；§4只读/拒绝/固定新授权→Task1及Task2重入；§5真实链/完成/出口→Task2与Task3；§6兼容含义→Task1拒绝及Task3操作者说明；§7矩阵→三任务完整向量，未诱发晚期边界对应五Review Focus；§8分层C1/C2→Task3审查处置；§9预算/恢复→六轮13进程/条件1review/portable。来源哈希、接口、2064映射/Absent转换、46快照、固定输出组数和预算逐项交叉核；无额外隐含宿主。

推荐 **Native**：三任务共享私有登记、固定profile与冻结身份，内联能减少接口漂移，保留最后一次独立整体审查。若选择Subagent-driven，须先重新列实现/逐任务审查数量及预算，不把本计划的一次review拆成多次。

请审阅本计划并确认执行方式与 **NEW6宿主+7子进程=13个新受控进程、条件1次只读审查**。在该批准之前，仅可维护本计划及恢复资料；不建立上述代码/测试骨架、不Prepare、不运行目标、不新增审查。
