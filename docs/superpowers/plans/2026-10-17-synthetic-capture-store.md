# Full Synthetic Capture Store Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended native/inline execution,or superpowers:subagent-driven-development only if the user chooses that method. Steps use checkbox syntax. This plan is not approved for execution yet.

**Goal:** 保存并恢复20字段完整合成记录，验证原字节、事件顺序及派生事实一致性，不执行归档内容或提升权限声明。

**Architecture:** 独立有界固定二进制编解码及新建式本地归档，单向依赖现有合同与isolation。读回保留磁盘派生事实，再调用公开合同校验；旧输出包格式/API不变。

**Tech Stack:** Windows PowerShell5.1/7、.NET BinaryReader/BinaryWriter、严格UTF8、SHA256、既有SafeFileHandle文件信息检查；不新增依赖或安装工具。

**Spec:** [已批准书面设计](../specs/2026-10-17-synthetic-capture-store-design.md)，批准基于b33e04bd。执行者先完整阅读设计及本计划。

## Global Constraints

- 当前仅计划写作权限；待用户批准计划、执行方式和下表有限验证包后才能写产品代码/运行测试。
- 复用当前隔离worktree及 `codex/c12-b01-task9-coordinator`，遵循AGENTS.md；不新建/删除worktree，不合并main、force-push、改历史或安全设置。
- 新模块只导入 `capture-adapter-contract.ps1` 与 `isolation.ps1`；不改旧store、合同、模拟器、输出包、分类器及其测试或建立反向导入。
- payload magic TLRSYN01，Int32存储版本1；marker magic TLRSYC01、精确52字节，8+4+8+32；payload上限8388608字节，所有整数小端，Bool/存在位只允许0/1。
- 精确20字段记录、Lifecycle17字段、可选Package13字段及顺序按设计第3/5节；Snapshot盘上原字节，恢复规范Base64/ByteLength/SHA256；Digest盘上32字节。
- SessionLabel ASCII1..80；exe/origin各1..1024字符、严格UTF8<=4096字节；UTC精确28字符；encoding仅utf-8/utf-16le；参数<=65536、严格UTF16LE偶数长度且无前导BOM。
- Events精确object[]<=4096；chunk<=65536；全量chunk总和<=2097152；admitted stdout/stderr各<=1048576；ElapsedMs为0..Int64.MaxValue，exit为Int32；不改120000ms虚拟deadline。
- 只接纳现有记录校验Verified=true；四种内部一致生命周期可归档。clock-order不一致、Created=false描述符与非法/超限输入拒绝；CleanupStatus=unknown，四个顶层行动/权限flag全部false。
- 先本地被动类型/形状/标量检查再读值及上游校验；拒绝ScriptProperty/CodeProperty、对象非NoteProperty字段及Length/Rank覆盖；不派发调用方GetType/ToString，不规范化Unicode。
- 专用ignored根 `.superpowers/diagnostics/offline-synthetic-capture-store`；直接子目录leaf1..80；record.bin/complete.bin/stopped；路径OrdinalIgnoreCase，协议Ordinal。
- 拒绝UNC/ADS/相对/穿越/reparse/硬链接数非1；CreateNew、独占写，读FileShare.Read并持有安全句柄，前后检查stopped；已有目标不覆盖。
- 校验/编码成功前连缺失专用根也不创建。创建目标后的错误保留文件，尽力写stopped但不覆盖原错误；无删除/重试/自动修复/静默迁移。
- 校验完整不等于Lifecycle complete/真实执行；不承诺来源认证、并发敌对对象/目录竞态防护、事务原子性、fsync/断电持久性。无真实采集/命令执行/安全/Stage B/I2/I3/full C12/main权限。
- 新执行包拟定最多12主机、每次120000ms、一次只读审查、无修复或重跑预算；全项目测试不在包内。所有旧/冻结额度不变；任何异常停止并保存未完成开发检查点。

## Review Focus

这五类易被正常往返漏掉的条件，均在下方测试任务内指定断言；不是排除项：

1. 合法exit0与缺失exit/索引null混淆：NullableExitAndIndexDistinct，恢复值/存在位不可混同。
2. Int64.MaxValue虚拟事件时间被double/Int32截断：ExactInt64TimeRetained，恢复精确long并保持timed-out状态。
3. 完成标记存在但stopped也存在：StoppedOverridesCompleteMarker，无Record泄露且仍失败。
4. 文件内或额外文件含可执行文本：ArchiveTextNeverExecuted，只读取固定文件，无执行/外部路径访问。
5. 空事件缓存或调用方修改污染另一次读取：IndependentEmptyEventOwnership，两次结果独立，损坏一份不改变下一份。

## File Structure

- Create `docs/diagnostics/isolated-fuzz/synthetic-capture-store.ps1`：完整归档的本地被动校验、固定编解码与两公开API。私有函数统一SyntheticStore前缀，避免污染旧函数名。
- Create `docs/diagnostics/isolated-fuzz/synthetic-capture-store.tests.ps1`：真实公开API、独立手工二进制oracle、有界文件fixture；Group参数format/boundary/all，默认all。不得改生产函数以制造测试通过。
- Create `docs/roadmap/2026-10-17-synthetic-capture-store.md`：批准范围、逐主机结果、hash、审查处置和限制。
- Modify 当前计划：仅追踪进度/结果；Modify `docs/roadmap/current-status.md`、`docs/runbooks/repository-recovery.md`：实际可恢复功能和未完成门槛。
- 仅执行期建立ignored `.superpowers/sdd/2026-10-17-synthetic-capture-store/`：progress、顺序调用索引、stdout/stderr及有限runner。旧索引/receipt不改不删；tracked测试不依赖它。

## Proposed finite execution packet — approval required

|调用|新主机|唯一工作|通过条件|
|---|---|---|---|
|1|PS7 RED|新suite Group format|缺少公开API，精确SyntheticCaptureStoreRequired断言失败；不是fixture/解析错误|
|2|PS5 RED|同上，仍未实现|同一缺失API失败|
|3|PS7 GREEN|新suite Group format|组内全部具名检查通过、完成标记、exit0/空stderr|
|4|PS5 GREEN|同上|同上|
|5|PS7 GREEN|新suite Group boundary|组内全部具名检查通过、完成标记、exit0/空stderr|
|6|PS5 GREEN|同上|同上|
|7|PS7 regression|既有output-evidence-store.tests.ps1|81PASS、exit0/空stderr/原终结标记|
|8|PS5 regression|同上|同上|
|9|PS7 regression|既有output-capture-simulation.tests.ps1|106PASS、exit0/空stderr/原终结标记|
|10|PS5 regression|同上|同上|
|11|PS7 regression|既有capture-adapter-contract.tests.ps1|110PASS、exit0/空stderr/原终结标记|
|12|PS5 regression|同上|同上|

每次独立新host、NoLogo/NoProfile/NonInteractive，PS5使用Windows模块路径；顺序执行。
每一host启动前登记已消费槽位，完成后记录真实退出码/耗时/每个组完整具名结果和stderr；
超时、取消、启动失败、fixture错误或非预期输出都停止，未消费槽位冻结而非转作修复。
RED预期失败是仅有的允许继续的非零退出；不声称每个后续用例都单独观察过RED。
GREEN后不得在下一组前悄悄改代码/测试。新组各自完整覆盖冻结的具名用例集合，
首个RED前固定其预期断言计数与终结标记，之后不按实际输出调低预期。
format完成标记SyntheticCaptureStoreFormatComplete；boundary完成标记SyntheticCaptureStoreBoundaryComplete。
分组是为了减少单host负担，不保证能在120秒内完成；超时如实失败，不延长上限。
all仅是便于将来显式使用的组合入口，本包不额外执行all。

## Task 1: 完整离线合成记录归档及有限验证

**Files:** 新store模块及新tests；本计划、执行report、状态与恢复文档随同交付。

**Interfaces:**

- Consumes：`New-SyntheticCaptureRecord([object]Packet)`、`Test-SyntheticCaptureRecord([object]Record)`、`Get-IsolatedFullPath([string]Path)`、`[IsolatedFuzzFileInfo]::Snapshot(SafeFileHandle)`；不调用旧store私有函数。
- Produces：`Write-OfflineSyntheticCaptureStore([string]Directory,[object]Record)`；成功返回下述结果，前置/写盘错误抛出。
- Produces：`Read-OfflineSyntheticCaptureStore([string]Directory)`；精确7字段Schema=1,Verified,Reason,Record,OriginVerified=false,Acceptance=false,RuntimeApproved=false。
- Reader Reason：成功consistent-synthetic-archive-only；失败固定代码invalid-path/missing-archive/incomplete-archive/unsupported-format/invalid-payload/integrity-mismatch/invalid-record/io-error。失败Record=null，不返回异常/路径/原输入。
- Test local：`Check-CaptureStore([bool]Value,[string]Name)`失败抛ASSERTION FAILED，成功打印PASS；`-Group format|boundary|all`选择字面用例表；不引入运行时测试开关。

- [ ] **1. 只读执行前核对。** 阅读AGENTS/设计/计划，核对当前分支、干净状态、现有依赖hash与12槽位批准。复用已有worktree；任何范围/源漂移先停，不顺带处理。按选定方式读取执行技能。只在此后建立新ignored索引，不改旧receipt。

- [ ] **2. 写format组固定fixtures和断言。** API guard名SyntheticCaptureStoreRequired先检查两函数存在；fixture内仅声明合成exe='x'、session='s'、origin='o'、UTC='2026-10-17T00:00:00.0000000Z'，不执行它们。预期payload与52字节marker采用手工编码的固定literal/oracle，不用生产encoder或生产hash helper生成预期。

  具名行为：PreflightPreservesRootState、IndependentWireLiteral、TwentyFieldsRoundtrip、FourLifecycleRoundtrips、NullableExitAndIndexDistinct、ExactInt64TimeRetained、RejectedLateTraceRetained、Utf8Utf16UnicodeAndBytes、IndependentEmptyEventOwnership、CallerMutationCannotChangeArchive、ArchiveTextNeverExecuted。
  Snapshot字面预期stdout='YWJj'/3、stderr='AP8NCg=='/4；参数字面Base64='QQAtTg=='；所有权限false；新结果字段数7、Record字段数20，reason精确consistent-synthetic-archive-only。
  原始参数fixture用byte[]@(65,0,45,78)，stdout用byte[]@(97,98,99)，stderr用byte[]@(0,255,13,10)。示例独立断言（actual为公开writer/reader返回结果，不是生产encoder的预期副本）：

```powershell
Check-CaptureStore ($actual.Verified -and @($actual.PSObject.Properties).Count -eq 7 -and @($actual.Record.PSObject.Properties).Count -eq 20) 'TwentyFieldsRoundtrip'
Check-CaptureStore ([string]::Equals($actual.Record.Arguments.Base64,'QQAtTg==',[StringComparison]::Ordinal) -and [string]::Equals($actual.Record.Lifecycle.Stdout.Base64,'YWJj',[StringComparison]::Ordinal) -and [string]::Equals($actual.Record.Lifecycle.Stderr.Base64,'AP8NCg==',[StringComparison]::Ordinal)) 'LiteralArgumentAndStreamBytes'
```

  FourLifecycle分别空事件incomplete、exit+双EOF complete、tick120000 timed-out、重复EOF refused；末者必须仍可完整归档。Int64.MaxValue tick精确保留，不能以超时丢弃全量trace。
  PreflightPreservesRootState在任何建根fixture之前执行：坏Record不得创建目标目录，专用根存在性须等于调用前。缺根分支仅首次根确实不存在时动态证明；后续主机保留已存在根，不能假称再次证明缺根分支。若执行前根已存在，则报告缺根分支未动态覆盖并提交审查，不删除/移动根或额外启动host。

- [ ] **3. 写boundary组真实输入/文件断言。** 具名行为：PreflightCreatesNoDirectory、ClockReversalAndDescriptorRefused、ExactTokensAndShapes、MethodsAndGettersNeverInvoked、LengthRankShadowsRefused、ExactFieldAndTraceBounds、MalformedPrimitiveRefused、ValidHashInvalidDerivedStateRefused、StoppedOverridesCompleteMarker、PartialArchiveNeverPublishesRecord、ExistingTargetNeverOverwritten、UnsafePathsAndLinksRefused、ExclusiveWriterRefused、OldNewFormatCrossReadRefused。
  格式损坏用独立字节fixture及重算文件marker覆盖，不替换生产decoder/validator。针对bool2、存在位2、版本2、负/溢出/超限长度与计数、非法UTF8、截断与tail逐项断言Verified=false且Record=null。marker长度51/53与hash不符单独测试。边界8MiB总长不等于合法记录：合法字段/事件界限仍须分别断言。
  被动验证用真实调用计数器0、无新目录和反例后的正常记录；未调用GetType方法正例仍可往返。CodeProperty使用固定测试计数类的MethodInfo构造getter；只允许Add-Type编译测试中硬编码的无害fixture，不编译/执行归档输入。路径fixture仅在专用合成根及其已验证父目录，危险路径只送入拒绝检查，不实际访问UNC；保留所有创建文件/链接，不删除根。
  OldNewFormatCrossReadRefused具体覆盖：新根内旧magic payload配有效新marker，必须被新reader拒绝；旧reader对新根目录拒绝。旧格式自身可用由调用7/8完整回归证明；不在新suite写入旧根或声称路径拒绝已覆盖旧reader的所有新payload解析路径。

- [ ] **4. 冻结测试目录与运行包。** 只读解析脚本，检查独立oracle、具名用例集合/计数、Group终结标记和不泄露数据；记录hash。准备顺序runner，不增加内部PowerShell测试子host或额外PS诊断host。没有实现时执行调用1、2，预期均为guard缺失；其他结果停止。

- [ ] **5. 实现本地被动校验与有界固定codec。** 仅在新模块，单向导入批准依赖；按设计第4/5节读取真实CLR类型、描述符/长度、精确字段及固定线格式。所有整数长度计算有界Int64；严格Bool/存在位；不泛化对象序列化或分发调用方方法。完整保存盘上Lifecycle/Binding等，不用重算值替换它们。这个步骤不运行额外host。

- [ ] **6. 实现两个公开API及发布。** 按设计第6/7节在创建任何目录前冻结已验证payload；本地路径/安全句柄检查、CreateNew、重读核对、最后发布marker、最终公开reader验证；读取先检查总长/hash再解码/合同重算。错误只映射固定Reason，回收本调用打开的资源，不删现场。失败后写stopped尽力而为，不吞原错误；不把一致性标记为运行成功。

- [ ] **7. 有限GREEN。** 按表执行调用3..6，任何失败立即停止、保留记录，不修改代码重跑。两组每种host均完整通过才算新store得到本范围验证；不拿format通过替代boundary。代码/测试hash须在各次间和最后一致。运行表达式示例如下，必须由120000ms计数runner启动而非手动额外再跑：

```powershell
pwsh.exe -NoLogo -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/synthetic-capture-store.tests.ps1 -Group format
powershell.exe -NoLogo -NoProfile -NonInteractive -File docs/diagnostics/isolated-fuzz/synthetic-capture-store.tests.ps1 -Group boundary
```

- [ ] **8. 有限依赖回归。** 调用7..12分别按表执行既有三个完整suite，精确81/106/110PASS并核对终结标记。原PS5合同历史耗时102530ms接近上限，不作本次耗时保证；超时仍停。分类器/输出包 standalone及全项目测试未额外运行，报告清楚这一限制。

- [ ] **9. 一次独立只读审查。** 仅全部验证完成后，按requesting-code-review技能派一名新审查者，传入设计/计划、实际base/head及hash、完整文件清单和逐host证据；不继承会话史。只读源码/receipt，不dot-source、不执行测试/网络/Git写入/修改/再委派。重点检查Review Focus、wire oracle、全字段派生事实保留、本地被动读取、路径/错误状态与隔离限制。
  记录Critical/Important/Minor及排除项；有Critical/Important即未完成检查点，不自动修复/二审或增加host。可按证据说明不适用意见，但不隐去未解决事项。无缺陷结论也只限离线合成归档。

- [ ] **10. 保存交付。** 执行报告逐项记录真实结果/耗时/退出/stderr、已消费与冻结槽位、源/测试/索引hash、审查结论、设计偏差及已验证/未验证范围；更新状态和恢复文档。失败照样明确标记未验收。保留ignored现场，仅显式暂存必要source/tests/spec/plan/report/status/recovery；提交推送现有dev分支并确认远端SHA=本地、工作区无遗漏。无PR/main合并或worktree删除。

## Self-review and handoff

设计§1..3对应步骤2/5/6；§4/5边界及wire对应步骤2/3/5；§6/7重建/发布/错误对应
步骤3/6；§8恢复对应步骤10；§9覆盖对应步骤2/3/7/8/9。五项Review Focus全部
有具名测试归属。成功/失败7字段接口与设计一致，20/17/13字段不改，三种magic
不混用。缺根分支只按真实前提报告，不移动旧根制造覆盖；跨格式拒绝不夸大路径层证据。
测试本身仍需实施时写出并冻结计数；本计划没有现成实现或测试通过声明。
额外全项目测试、复测、强制中断真实进程和安全放行均不包含，成本是没有全项目/
真实采集/断电认证。书面计划自审，不派审查者或启动测试。

推荐native/inline：一个紧密耦合的store任务，由当前会话实现，最后一次独立只读
审查，避免跨实现者传递私有codec状态。另一选择是subagent-driven：任务由独立
实现者完成并审查，增加上下文成本；同样不得增加12host预算。
请用户审阅并批准本计划（包括12host/一次审查/无修复预算），选择执行方式后再实施。
