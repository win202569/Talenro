# Offline Diagnostic Runtime Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task-by-task. No implementation agents or per-task reviews. Only the conditional final fresh readonly reviewer below is proposed. Track progress in this plan's own ledger; do not edit approved plan bytes.

**Goal:** 实现目前仍缺失的8个公开/私有离线诊断文件，以真实新入口验证只读检查、存活确认和失败出口隔离。

**Architecture:** 沿用已批准的“公开 Inspect/RunApproved＋私有实现”方案。新增独立运行测试 leaf 和 packet；只读取历史原件，不调用历史 driver/capture/module，不恢复旧额度。2078项准备清单是输入身份基线，不是运行许可。

**Tech Stack:** Windows PowerShell7.6.5 Core x64、.NET Process、原始字节双流；Native工具宿主作为父传输；不安装依赖。

**Spec:** [已批准设计](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md)，SHA256 DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A。[原实施合同](2026-10-19-offline-diagnostic-entry-migration.md)，SHA256 062AAC17317B5F8FBE5E29A88A67ABF6E53D5BD4BC6CA8151AC04DE783F81205；[准备完成证据](../../roadmap/2026-10-31-offline-diagnostic-entry-preparation-v2-complete.md)。设计选型已批准，不重启设计选择；原合同的§2全部接口/数据字段和Task1–3字面组名、事实、拒绝向量是本计划的规范性引用，执行者须完整阅读。本文件的新目录、清单、预算及停线规则替代原运行安排，不修改原件。

状态：PLAN_PENDING_USER_REVIEW_AND_NEW_EXECUTION_APPROVAL。基线39d55e3050f0712fbc7f14a3188f1bdea2dd0e11。本轮只编写计划/恢复说明，0运行代码、测试、Prepare、导入、目标、代理或审查。

## Global Constraints

- “只涉及离线诊断，不做生产接入”；不改api/cmd/internal、CI、Docker、verify/run-c12-integration，不合并main、不发布。
- “默认为 Inspect”；不导入执行模块、不创建运行目录或Session、不启动目标、不安装/下载依赖。资料不授予运行权。
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”；只在单个存活宿主中按真实引用、步骤和私有预算消费。
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”；清理不重试失败出口。
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”。
- “唯一受支持入口”不等于OS禁用所有旧脚本；不转译旧Grant/Slot/Resume，不新增任意命令/路径/回调运行接口。
- 原始1930及后续2078映射、所有旧leaf/packet/portable/批准/审查/报告保持原字节。C1/C2 open、R2/Minor2延期、五项能力标记false。
- 所有历史额度维持耗尽/冻结。下面NEW6宿主/7子进程/条件1审查须另获批准，绝非恢复原6/7/1。
- 仅手动有限PowerShell；批准本计划也批准其对技能默认额外Bash/setup/baseline/full-suite/task-done重跑/实现代理/修复轮/二审/删除的限制。保留全部原始记录。

## Review Focus

1. Prepare身份：可选Role、必需ExpectedState、8个Absent转换、重复路径身份一致，以及portable加入后的显式文件清单。Task1静态全量前置检查，Task1–2精确新增窗口及每轮后置验证覆盖；不额外运行准备探针。
2. 默认只读：模块已存在时Inspect/未知参数也不能准备运行；同宿主二次Run不能重置私有状态。Task1实际调用及Task2重入、Task3存在状态回归覆盖。
3. 确认负例不能被零预算/空步骤混淆；Task3在Active、Next1、Reserved1、预算2/余量1且真实Ack仍有效时检查成功磁盘台账不能代替Ack。
4. 分类器/路径迁移必须保留实际首因和确切资源清理；Task2预消费故障及Task3预启动、receipt、console-output真实故障覆盖。未诱发的晚期关闭/并发、分类器自身异常由最终审查明确限制，不当成已动态通过。
5. 工具提前返回与真实运行错误必须区分；每次完整保留工具结果/真实session_id/所有输出片段/退出，再解析。未知或最终记录失败不能从已有PASS/portable/final推导下一轮权限；Task1传输实现和Task3整体审查检查此规则，不加自检宿主。

## 1. 文件、目录、冻结身份

继续使用现有linked worktree，分支codex/c12-b01-task9-coordinator，执行前只读确认clean/非submodule；不另建worktree。

新测试leaf：`docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/`。
新packet：`.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration/`。
执行报告：`docs/roadmap/2026-10-31-offline-diagnostic-runtime-migration.md`。
操作者说明：`docs/runbooks/offline-diagnostics.md`；交接更新current-status/repository-recovery。
本轮不建立这些运行文件/目录。执行时新packet初始只允许本计划progress.md；任何历史或部分新运行输出均拒绝，不自动换目录。

| 来源/阶段 | 精确文件 | 责任与约束 |
| --- | --- | --- |
| 已有且冻结 | scripts/.gitattributes | 沿用已建立的两个LF规则，不修改 |
| 新leaf首冻前11项 | .gitattributes、README.md、expected-output.json、test-helpers.ps1、entry.tests.ps1、core.tests.ps1、prefix.tests.ps1、owned.tests.ps1、live.tests.ps1、capture.ps1、driver.ps1 | 分别负责LF、操作边界、独立字面oracle、测试工具、真实入口合同、core7组、prefix7组98条件、owned6组、live5组、单宿主捕获、原始父传输/清单；首次冻结后全部不改 |
| Task1新增2项 | scripts/diagnose-offline.ps1；scripts/private/offline-diagnostics/catalog.json | 实际公开入口与纯描述catalog；从继承Absent转Present |
| Task2新增6项 | scripts/private/offline-diagnostics/session.psm1、sinks.ps1、owned-step.ps1、sequence.ps1、profiles.ps1、observations.ps1 | 私有登记、失败出口、确切目标、完成/确认顺序、固定profile分发、首因观察；只允许继承Absent转Present |

新leaf测试/oracle可从已冻结p2b-entry-migration的对应文件迁移，但只在新文件中做固定路径/配置适配。原leaf不可改，也不可作为运行入口。私有机制迁移来源仍为p2b-prestart-cause-repair的session/sinks/owned-step/sequence及capture观察分类，不Import该历史module。

公开入口repo根为scripts父目录；私有module上溯../../..；新leaf上溯../../../..。所有fixture/观察器/seam/测试/packet采用repo＋固定相对路径，不依赖cwd。新测试中Grant是新父捕获阶段，私有测试作用域的Grant固定为实际profile标识，禁止让历史phase名称渗入会话预算选择。

### 1.1 初始2129映射，不重复追加8个Absent运行文件

继承`.superpowers/sdd/2026-10-19-offline-diagnostic-entry-preparation-v2/preparation.frozen.json`的全部2078Pins；文件SHA256必须为895E06C510D5466DF397AF63B5B713A62095FBE8ECE67A1321C2963CFBAC915D。保持每条VerificationPath/SHA256/ExpectedState；Role按已验证的可选属性查询保留/缺省null。未知状态、缺必需身份、漂移即拒绝。

再增加：V2真实packet28文件＋V2 leaf8文件＋V2完成报告1＋本计划原件1＋新packet的批准设计/本计划快照2＋新leaf11文件，共51项。2078+28+8+1+1+2+11=2129；V2 manifest已包含在28原件中，不另加。V2 portable-prepared SHA256为80C34D471B77D4B97D3227E78F74D0574532CD757042976888D00EBB1EA7C208，须只读逐字节验证28原件，不从其重建缺失文件。

脚本属性与8运行路径已经在2078中，不能再加重复“新输入pin”。重复历史路径只允许同state/hash；实际源身份读取允许多个一致Present映射，必须拒绝冲突，不简单要求所有历史引用路径只有1条。

完整源码快照输入20项＝已有scripts属性1＋新leaf11＋8个继承Absent。初冻2121Present/8Absent，快照12；入口后2123Present/6Absent，快照14；私有实现后2129Present/0Absent，快照20。三次快照共46，不创建Absent假文件。

只允许两个固定新增窗口：entry-green恰公开入口/catalog2项，runtime-green恰私有实现6项。每次同时核验全部2129项；剩余不相关Present输入不得改变，未到窗口的Absent必须不存在。不重新写旧2078清单或其准备预算字段。

driver静态检查采用显式11源码名称；最终portable仅交接数据，不属于运行输入，不靠目录总文件数推断。首次freeze前写齐全部测试/oracle/driver，之后不能修改期望来匹配实测。对保存原工具输出的JSON不做JS Number往返重发。

## 2. 合同绑定与新额度申请

原合同§2的公开参数、Inspect完整有序根、三profile、私有LS签名、ProfileReport、capture证据根及全部界限不变；只替换测试leaf/packet为本计划新固定目录。模块只读新packet/runtime-green.frozen.json，不读取旧packet的运行授权。三个profile仍为entry-migration-core/owned/live；不接受调用者提供FrozenPath/ScriptPath/Command/回调。

Inspect只读依赖/历史，缺6份实现时如实报告，无导入/Session/目录/目标。公开RunApproved三分支硬编码，catalog不能扩展许可。缺模块真实异常Type=RuntimeException，Message/FQID均为OD: backend unavailable；同宿主第二次调用在IO前真实拒绝OD: profile already invoked。不Force/Remove-Module、不重置调用位。

唯一模块导出Invoke-ODProfile([string]Profile)返回原合同ProfileReport，opaque Session/Ack均不返回。prefix/core/owned/live测试定义只能从新冻结leaf加载到私有作用域，测试工具先于程序；profile只固定分发，不接受可变callback。捕获观察ProfileReturn恰1条，父传输只从其Observations中Name=Step/Started=true计目标，不能重复计summary。

| 顺序/新Grant | 宿主 | 目标 | 必须出现的完整结果 |
| --- | ---: | ---: | --- |
| entry-red | 1 | 0 | 新公开脚本确实Absent；exit1，唯一FAIL EntryPresent，其余5入口组NotRun |
| entry-green | 1 | 0 | 缺省/显式Inspect、14实际拒绝、依赖和历史，共6PASS；exit0/空stderr，无模块导入 |
| runtime-red | 1 | 0 | 经真实公开RunApproved得到指定backend unavailable原始错误；唯一FAIL RuntimeProfileBound，其他15组NotRun；exit1 |
| runtime-green | 1 | 0 | prefix7＋core7＋RuntimeProfileBound＋ProfileReentryRejected，共16PASS；exit0/空stderr |
| lifecycle | 1 | 3 | 入口6＋owned6（预启动故障FIRST）＋绑定/重入2，共14PASS；3实际目标完整退出/排空 |
| continuation | 1 | 4 | 入口6＋live5＋绑定/重入2，共13PASS，另原seam12PASS；4实际目标完整退出/排空 |

申请NEW6capture宿主＋7受控子进程，全部精确才申请1次fresh readonly review。Native工具宿主仅作为父传输，不嵌套pwsh运行driver/入口；公开脚本在capture内调用，不再另起包装进程。无额外工具自检/FINAL/探针/重跑/独立编译目标。每轮退出、完整原始双流、全量前后pin、证据字段/类型/顺序/值、目标数和有限清理必须全部精确；任意exit1不能冒充上述两种RED。

非continuation宿主120000ms，continuation150000ms；内部全序列140000ms，清理预留4000ms；确切对象退出2000ms，双流共享排空2000ms。Hold45000ms、normal/receipt2000ms、timeout200ms、原seam110000ms。normal各流19ASCII字节；timeout只能是各自字面前缀，允许真实byte[0]。

固定PS7路径为C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；同目录SMA SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。固定机器DESKTOP-H4SVCLU、账户Lenovo、admin=False（WindowsIdentity.Name按原合同精确匹配）；不满足即停止，不自动安装或放宽。

固定fixture为p2b-controller-exception-containment/fixture-child.ps1，SHA EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F；关联观察器p2b-process-association-guard-repair/owned-process-observation.ps1，SHA CD6E9564E77FF55812F9E96291B0F4E511758245A9BBC53AD2802B95E809C6E5。两路径均以docs/diagnostics/isolated-fuzz/为前缀。原seam/codec/inbox继续由继承pins验证。seam stdout F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE，trace4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E。

Session≤32、sink≤32/Session、Ack≤2/Session、观察≤32、不同次因≤8，第9项Overflow。原始ErrorRecord先于调用者断言采集，错误位置绑定冻结文件及有效行号；PID/ElapsedMs只允许原合同规定的有限动态值，不接受泛化异常文本。

## 3. 单次、输出收集及停线

三次Prepare仅entry-red/entry-green/runtime-green；runtime-red复用entry-green，后两轮复用runtime-green。Prepare和每轮Run的参数/SHA/原始完整工具结果即时记本计划ledger，保留单次CreateNew原始记录；不为补证据重跑。

工具调用先保留完整返回对象，再看实际session_id或exit_code。若仍运行，只对该实际session调用等待/读取并保存各片段，直至退出；之后才解析合并输出。不得丢掉session、猜PID、把空/截断内容当JSON、另启同命令或把包装器exit0当测试通过。工具10秒yield不是测试超时，不自动停掉已知仍在运行的确切父传输；依据固定期限收束。

真实Prepare/验证失败、输入漂移、来源或预算不明、原始记录/退出/排空/final/evidence失败或非精确，立即冻结后续全部新Grant/review；保留真实首因，不自动修复、重冻、换输出目录、返还额度、二审或重建原件。仅完整固定签名的entry-red/runtime-red为预期负例，不触发该停线。

申请的窄区分：纯只读查询的语法/显示错误，或首冻前整批被拒且确认未落盘的源码补丁，可原样记录后纠正该查询/补丁；不消耗、恢复或绕过运行额度。它必须未调用受测代码、未写运行记录、未启动目标。冻结后源码不可改，真实输出收集/安全检查/记录失败不能改称显示错误；涉及实际验证或未知副作用仍停线。此区别须随本计划明确获批，不追溯更改旧停线结论。

## Task 1: 新bootstrap与实际公开只读入口

**Files:** 首冻前建立新leaf11文件；复用scripts/.gitattributes。RED后只新增公开脚本/catalog2项。
**Interfaces:** 消费原合同§2及本计划新路径/pins；产出Inspect与指定真实缺模块错误，供Task2调用。

- [ ] Step1：先完成全部测试/oracle/父传输。按原Task1固定6组、14实际拒绝、完整Inspect模板和导入/目录前后字节观察；Task2–3全部测试也在此时写定。Preserve Role/ExpectedState及state-aware继承，显式源码集合；使用本计划§1.1的2129映射，不导入或运行旧driver来“检查”。只读核对原件/哈希、LF/AST/JSON及全部模板自洽，不额外启动入口探针。
- [ ] Step2：Native执行新driver `-Action Prepare -Grant entry-red` 一次，再以实际返回SHA `-Action RunOne -Grant entry-red -FrozenSHA256 <实际SHA>` 一次。Expected2121Present/8Absent、12快照、唯一预定FAIL/其余5NotRun/exit1/0子进程及完整证据，无其他错误。由真实文件缺失支持入口RED，不调用不存在脚本。
- [ ] Step3：实现公开脚本/catalog。精确参数/分支及有序输出按原合同；Inspect无执行准备，未知参数导入前拒绝；RunApproved只惰性导入新module，缺module抛指定真实错误，不exit整个capture。为后续core调用保留真实错误。
- [ ] Step4：Prepare entry-green一次，恰2个继承Absent转Present，其余2127项不变，2123Present/6Absent/14快照；Run entry-green一次，Expected6PASS/exit0/0子进程/空stderr，包含完整14拒绝及历史资料不授予许可的观察。
- [ ] Step5：Task1完成检查点，显式必要文件提交推送/核远端。记录spent2host/0child、剩4host/7child/条件1review；不声称runtime已可运行。

## Task 2: 私有运行链及实际公开core消费

**Files:** 仅新增session.psm1/sinks.ps1/owned-step.ps1/sequence.ps1/profiles.ps1/observations.ps1；测试、oracle、driver和公开入口均不改。
**Interfaces:** 消费Task1公开RunApproved及预冻LS测试；唯一产出Invoke-ODProfile/ProfileReport与Task3真实运行机制。私有LS函数签名、Step/IORoot和一次性Ack按原合同§2；不向外导出会话句柄。

- [ ] Step1：读取预冻core/prefix测试；Run runtime-red一次，复用entry-green实际SHA。Expected真实公开入口OD: backend unavailable、原始Type/Message/FQID/未导入事实、FAIL RuntimeProfileBound/其余15NotRun/exit1/0目标；其他错误一律不能替代。
- [ ] Step2：在六新文件中实现私有登记、sink状态、确切对象运行、完成后Ack、固定profile分发和bounded观察。module只读取本计划新runtime-green清单，不接受外部路径/回调。单次调用位先锁存，再任何IO；finally关闭全部Session，第二次真实公开调用拒绝且模块引用/目录字节不变。
- [ ] Step3：保留原Task2四种期待Step故障分类的全部实际事实：core/preconsume的Start失败；owned/prestart的Out0失败/Start0；owned/receipt-fault的Receipt0一次失败及退出排空；live/output-fault的真实ConsoleOut失败及退出排空。先保存实际ErrorRecord首因再断言/分类，不伪造异常、观测或生命周期setter。普通网关first/secondary/failed后LogicalIO0按七core实际测试；不冒称所有晚期边界已诱发。
- [ ] Step4：Prepare runtime-green一次，恰6个Absent转Present，其余2123不变，2129Present/0Absent/20快照。Run runtime-green一次，Expectedprefix7/98条件→core7→绑定/重入2共16PASS/exit0/空stderr/0子进程；仅从唯一ProfileReturn计目标，来源哈希另从pins核对。
- [ ] Step5：提交推送/核相等；记录spent4host/0child、剩2host/7child/条件1review。初始缺入口/缺backend RED证明新组装边界，不声称每个移植分支独立RED诱发。

## Task 3: 生命周期、确认对照、独立审查与可恢复交接

**Files:** 运行源码及预冻测试不再改变；仅产生固定记录、审查/处置、操作者/执行/恢复文档与最终portable。
**Interfaces:** 消费Task2真实公开profile链和runtime-green SHA；产出完整实际生命周期证据及限定结论。

- [ ] Step1：Run lifecycle一次。先6入口组（此时6实现均存在、Inspect仍不导入/准备），再owned6，最后绑定/重入2，共14PASS/exit0/3目标。prestart FIRST验证Start0/Unassociated/Kill0/Wait0，随后receipt/normal/timeout3真实目标均关联/退出/双流收束，receipt失败出口1次/后续LogicalIO0，无外层救援。
- [ ] Step2：Run continuation一次。先入口6，再真实公开live程序；A0/B0/C0/A1恰4目标，原seam只能在A1跑一次。保留原Task3全部断言与5live组顺序；关键独立对照为A0后Active/Next1/Reserved1、预算2/步骤2/余量1、真实Ack仍可推进，实际Ledger为CANDIDATE_COMPLETED，缺Ack请求拒绝/Start不增加/台账字节不变。随后真实Ack恰消费一次，复制/跨Session/错ordinal/重放拒绝，C真实输出失败无Ack/下一步拒绝，关闭后磁盘摘要不能恢复会话。Expected13PASS另原12seamPASS，4子进程及原stdout/trace完整匹配。
- [ ] Step3：仅6host/7child/2129pins/46快照/全部字面结果和原始记录精确才提交完整源码检查点，申请的唯一fresh reviewer开始。执行时核当前最强可用模型/独立上下文，只读审查整个新入口链、批准设计/原合同/本计划、所有rulings及5Review Focus；不导入/测试/网络/改文件/Git或再派代理。重要缺陷或不完整审查即停，保存未验收检查点；无自动fix/额外测试/二审。没有审查不能声称新C1/C2窄边界获验收支持。
- [ ] Step4：保存完整审查及每项Declined-to-judge处置，明确未动态验证的晚写/flush-close、宿主死亡、所有出口同时失败、分类器自身异常。即使审查干净，也不回填旧C1/C2或升级生产/正式/整体C12验收；五flags仍false，R2/Minor2仍延期。
- [ ] Step5：写offline-diagnostics操作说明：支持新入口默认Inspect，未来运行仍需独立批准；无旧参数兼容、无OS层禁用旧脚本、无自动恢复Session/Ack。保存所有实际packet原始JSON文本/其他字节、46源码快照、批准/manifest/intent/final/失败/unknown/审查和目录；portable最终才加入新leaf，不改输入清单。逐字节/长度/hash核原件/暂存blob，直接保留byte[0]，不重写或重建缺失记录。
- [ ] Step6：更新执行/current-status/recovery，显式文件清单提交推送既有dev分支并核远端SHA/工作树。导出与文档最后机械核对是作者交接，不冒充独立审查。无main合并、删除raw/worktree或发布。

## 4. 作者自检、授权边界与交接

已复核：设计§1–3→新归属/原件保持，§4→Task1默认/拒绝，§5→Task2会话/出口及Task3真实确认，§6→操作者说明，§7矩阵→三任务及未诱发限制，§8→条件审查/分层结论，§9→新预算/恢复。2129=2078+51；源码20项已有12/14/20→46快照；2121/8→2123/6→2129/0；目标3+4=7。不会重新追加继承8Absent或把准备manifest的零额度复制成运行许可。

本计划不更改架构选型；Native继续作为推荐/既有选择。用户本次“按建议继续”仅批准编写本计划，未批准此书面新执行方案、NEW6/7/条件1或新的窄停线区别。请审阅并明确批准；此前只可维护计划及恢复资料，不能写bootstrap/实现、Prepare、导入/测试/目标或审查。
