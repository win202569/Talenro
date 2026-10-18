# P2-A Pure Integer Timebase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended Native execution after explicit approval. Alternative subagent-driven execution requires the user to select it separately. Steps use checkbox syntax; no work is authorized by this document's existence.

**Goal:** 实现只消费声明整数的时基换算模块，返回独立数值计划或保守拒绝，不读取真实时钟或取得任何运行权限。

**Architecture:** 独立新目录内的单一纯模块完成被动输入准入、精确有界整数运算和独立输出构造。测试使用冻结完整 literal；一个有界控制器仅启动下列五个指定测试宿主。与 P1/真实适配器/旧归档单向隔离，不导入或修改它们。

**Tech Stack:** Windows x64 PS7.6.5、内置 System.Numerics.BigInteger 精确整数计算及固定 SDK 元数据；无安装、Add-Type、原生编译或新依赖。

**Spec:** [P2 书面设计 §5/§9/§10](../specs/2026-10-18-p2-adapter-boundaries-design.md)。用户在设计提交 7e5550974257f3edce5355d1b7acc34bb6d0b497 后，对“批准设计，仅进入 P2-A 实施计划编写”回复“是”。设计固定物理 SHA256=A8E8AD45081E73D526324AF15153C0DDB90E531A92A10C03BF6DAF40A1D52705；原设计起草状态留作历史，不回写。

状态：用户在 cc0663a5 后批准 Native 路线和有限额度；现已 STOPPED_AT_SLOT_2_INPUT_MUTATION。
已用 2/5 个普通宿主、0/1 reviewer。slot1 精确 RED；slot2 在 InputMutation 失败，
余下三次及条件审查冻结，不自动修复/重跑。详见
[实际失败记录](../../roadmap/2026-10-18-p2a-timebase-execution.md)。未验收，无真实运行权限。

## Global Constraints

- 复用 D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator，codex/c12-b01-task9-coordinator；不新建 worktree、不合并 main、不发布。
- 产品只定义函数：无 Process/Job/pipe/Stream、时钟/UTC/QPC/Stopwatch、file/network/security API、外部句柄、callback、imports 或 load-time 动作。
- 输入精确五字段；输出根八字段、成功 Plan 八字段。SourceVerified/LiveActionsPerformed/RuntimeApproved 始终 false；Created 不是授权或真实时基证明。
- OriginTick/ObservedTick/StopTick 精确 Int64；Frequency 精确 Int64 且 >=1000；Schema 精确 Int32=1。不通过参数绑定或数值强制转换先于验证改变输入。
- 固定 120000ms 总量、113500ms 停收点、6500ms 保留及 500/5000/1000ms 阶段；floor 执行预算、ceil 预留预算；等值到期，无浮点或溢出裁剪。
- P1、既有 store/raw coordinator/测试/旧报告及 O1–O23/N1–N9 不变；PS5/完整 P1、P2-B/C/P3、Stage B/I2/I3/C12/main/live/安全放行均不在范围。
- 本次拟议独立额度最多五个普通目标宿主和一个条件只读 reviewer；旧已用/冻结额度不转移、不续期。任何非预期结果先恢复获准临时 mutant，再停止余项，不自动修复/重跑。
- 获准后，此精确有限路线覆盖通用技能的额外 baseline/RED/task-done/全项目/复审/修复循环；代价是无更宽执行证据。保留所有原始包，不清理 scratch。

## Review Focus

1. 超过 double 精确范围的整数仍必须精确：FractionalFrequency、FrequencyLimit 的完整 literal 和逐字段 CLR 类型核对。
2. min 的另一分支可表示也不能掩盖中间溢出：IntermediateOverflow 必须返回 range-refused、Plan=null。
3. 调用者的属性/方法及标量附加元数据不能执行或混入结果：PassiveDescriptors、DecoratedScalars、callback-zero 和输入/输出所有权核对。
4. 相同/迟到 tick 不能获得额外时间：DeadlineEquality、FractionalPhaseRounding、LateStop 的独立边界与两种指定 mutant。
5. 正常/拒绝/返回后突变都不能获得真实权限或污染下一调用：OutputOwnership、SequentialIsolation、RefusalShape、ConservativeFlags。

## File map and interfaces

唯一新代码目录：docs/diagnostics/isolated-fuzz/p2a-timebase-plan/。

| 文件 | 职责 |
| --- | --- |
| offline-timebase-plan.ps1 | public New-OfflineTimebasePlan([object]$Request)，独立纯实现 |
| offline-timebase-plan.tests.ps1 | 独立 fixtures/full-literal 比较器、被动性/所有权断言，Group=all/phase-rounding/deadline-equality |
| capture.ps1 | Prepare/SealProduct/Run 三动作，只有 Run 可启动指定有界宿主 |
| expected-output.json | 固定 RED/GREEN/两个 mutant 的有序输出预期，不由产品生成 |
| source-pins.json | 设计身份、保护链、引擎、输入/测试/oracle 与 allowance 声明 |
| README.md | 可移植执行 ledger、使用边界、结果报告入口 |
| .gitattributes | 新目录 *.ps1/*.json/*.md 均 text eol=lf；不改原入口规则 |

允许另新增 docs/roadmap/2026-10-18-p2a-timebase-execution.md 和完成后
docs/roadmap/2026-10-18-p2a-timebase-readonly-review.md；更新本计划/current-status/repository-recovery。
ignored packet 为 .superpowers/sdd/2026-10-18-p2a-timebase-plan/，必须新建且不存在；不覆盖旧包。
禁止为准备计划提前创建这些代码/测试/packet 文件。

输入/输出精确字段和拒绝分类以设计 §5 为准，不增参数或运行模式。Plan 中非null
数值输出固定 Int64，根 Schema 精确 Int32=1；Plan 无额外 Schema 字段。
根 Created/三个事实字段与 Plan.ObservedDeadlineExpired 精确 Boolean，Kind/Reason 精确 String。
拒绝输出为完整根八字段，Created=false、Plan=null，Reason 只能是设计四枚举中的拒绝项。

内部 helpers 统一 TimebasePlan 名称段，仅接受验证后的本地值：
New-TimebasePlanMetadata() 创建固定 SDK 元数据；Assert-TimebasePlanRequest([object]$Request,[object]$Metadata)
做完整被动检查并取得本地快照；Get-TimebasePlanDelta([long]$Frequency,[int]$Milliseconds,[bool]$Ceiling)
仅接受固定 ms 集 {500,1000,5000,6000,6500,113500,120000}；
Add-TimebasePlanTicks([long]$Left,[long]$Right)、Subtract-TimebasePlanTicks([long]$Left,[long]$Right)
检查 0..Int64.MaxValue；Copy-TimebasePlanScalar([object]$Value,[object]$Metadata)
及 New-TimebasePlanResult([bool]$Created,[string]$Reason,[object]$Plan,[object]$Metadata)
构造固定根/Plan，不是通用 serializer 或真实能力容器。
这些 dot-source 可见 helper 不构成安全隔离边界；唯一支持的业务入口是 public 函数。

## Task 1: 纯模块、独立测试与五次有界验证

**Consumes:** 设计的五字段声明；固定 literal、保护链和普通 PS7 宿主。
**Produces:** New-OfflineTimebasePlan([object]$Request) 的独立根/Plan，五次精确结果与停止/完成 ledger；不产出真实时基证据。

- [x] **Step 1 — preflight and write tests before module.** 核对当时 HEAD/干净分支/origin/实际批准。先写完整测试、固定 oracle、控制器及 metadata，模块文件必须不存在。保护上次 promotion 的 762 个映射/当前条目，再加入其 frozen/index/四份原始流六个 anchors 和本设计身份；全部只读核对。新目录 LF 物理字节与 Git 过滤字节一致后再 freeze，避免修复前次换行错误来假装旧回执通过。
- [ ] **Step 2 — pin full expected values and all 24 groups.** 测试专有 New-TBRequest 强制构造正确 Int32/Int64 fixture，TB-Need、TB-Case、TB-Equal 只做断言/严格完整递归比较；预期 root/Plan 不调用任何产品 helper。所有有效 Plan 比较全部八字段/精确 CLR 类型；所有拒绝比较全部根八字段和 Plan=null。确定下面表格的向量，并在完整套件开始前验证可执行属性/附加元数据 fixture 真正可观察、确实改变原对象；没有发生的突变不能当作通过证据。
- [x] **Step 3 — slot 1 RED.** 在产品不存在时，以固定宿主运行 tests -Group all。Expected:exit1，CONTEXT/HOST 两行后仅 ASSERTION FAILED: P2AApiRequired，无 PASS。非预期异常/模块已存在即停止，不先实现或自行换 RED。
- [x] **Step 4 — implement public function and helpers.** 依下述精确算法完成准入、范围/关系检查、算术和独立返回；不导入 P1 helper，不读任何真实时间。静态 parse/差异/类型路径核对，不另作产品导入试跑。冻结 good 产品 SHA 和 green manifest，再进入下一 slot。
- [ ] **Step 5 — slot 2 GREEN.** 相同普通宿主、相同 tests -Group all。Expected:exit0，固定 24 个 PASS、27 有序行（CONTEXT/HOST/24PASS/LIMIT），stderr 空、完整流、owned exit、身份完整；任何失败停止不修复。
- [ ] **Step 6 — slot 3 phase-rounding mutant and immediate restore.** 仅将计算 NaturalUntil 的 500ms delta 调用参数 Ceiling 从 false 改 true，其他字节/测试不变；保存 good/mutant 摘要与精确差异。Group phase-rounding 必须 exit1，在 P2AFractionalPhaseRounding 断言失败，先前仅 P2AApiRequired PASS。无论结果如何立即通过补丁恢复 good SHA；若非预期则冻结余项。
- [ ] **Step 7 — slot 4 deadline-equality mutant and immediate restore.** 仅将最终 ObservedTick>=DeadlineTick 改为 >，Group deadline-equality 必须 exit1，在 P2ADeadlineEquality 失败，先前仅 API PASS。保存差异/摘要并立即恢复 good；非预期即停止。
- [ ] **Step 8 — slot 5 restored full GREEN.** 恢复身份完全一致后，tests -Group all 再得与 slot2 相同的 24 PASS/27 行、exit0 和所有严格证据。不可把任意 exit1 当作 mutant 命中，也不可把 slot2 结果代替 restored final。
- [ ] **Step 9 — readonly reconcile and checkpoint.** 独立逐行读五次 stdout/stderr/索引，核对模块/测试/期望/保护链、顺序/普通上下文/消费前置、实际 elapsed/清理状态，记录失败与所有未验证项。审查明确文件名单，只保存必要文件到现有开发分支，普通提交推送并核对远端一致；不作正式验收/main 声明。

### Frozen test inventory and assertions

完整有序 PASS 标签如下，每个组可以有多个具名固定子向量；组数不是覆盖率：

| 标签（P2A 前缀） | 必须断言 |
| --- | --- |
| ApiRequired | public 函数存在；RED 精确此处失败 |
| NullStop | 0/1000/0/null，D120000、Stop113500、Until 全 null、Expired=false |
| NormalStop | 设计例 0/1000/4/2，全字段 Plan=0,1000,120000,113500,502,5502,6502,false |
| NonzeroOrigin | 1000/1000/1004/1002，D121000、Stop114500、N1502/C6502/P7502 |
| FractionalFrequency | 设计例 10/1001/120130/null，D120130、Stop113623、Expired=true；大整数见下表 |
| FractionalPhaseRounding | 10/1001/14/12，D120130、Stop113623、N512/C5517/P6518、Expired=false；500ms 取 floor，捕获指定 mutant |
| DeadlineEquality | F1000/T0=0/N119999、120000、120001，Stop=null；Expired=false/true/true |
| StopCutoff | F1000 和1001 的 113500ms 换算值及前/等/后观察；断言固定 StopAt，不能新增一个未设计的 stop-performed 字段 |
| StageWindows | 固定 S=2、N=501/502/5501/5502/6501/6502，Until 恒502/5502/6502；根不因阶段等值而发明真实停止 |
| LateStop | T0=0/F1000/N=Stop=120000 或120001，N114000/C119000/P120000、Expired=true，绝不追加时间 |
| OriginLimit | T0=9223372036854655807/F1000/N=Int64.MaxValue/Stop=null 成功，D=Max、Stop=9223372036854769307；T0再加1为 range-refused |
| FrequencyLimit | F999 拒绝、1000正常、安全大值成功、Max 拒绝；边界值见下表 |
| IntermediateOverflow | 上述安全大 Origin、N=Stop=Int64.MaxValue，s+500 溢出必须 range-refused，即使 min 的另一项可表示 |
| RequestShape | null、hashtable、非精确 PSCustomObject、缺/增字段、大小写变体、非 NoteProperty 均 invalid-input；fixture 无法构造的重复属性不能记成产品拒绝证据 |
| ExactTypes | Schema的Int64/非1、其余数值Int32/float/bool/string/null非法，Stop可null；不得通过强转准入 |
| PassiveDescriptors | 可构造的 ScriptProperty/AliasProperty/ScriptMethod 及脚本 GetType/ToString 计数器保持0；原 fixture 可观察，产品不得执行它。CodeProperty 的非 NoteProperty 拒绝路径静态核对，不为构造 fixture 新增编译或冒称动态回调已覆盖 |
| DecoratedScalars | 对每个输入字段的可附加元数据 Int32/Int64 leaf 逐一装饰，真实 descriptor/payload 在原视图可观察时必须 invalid-input |
| RangeRelationships | Origin/Observed/Stop负数、Observed<Origin、Stop<Origin 或 Stop>Observed 为 range-refused；类型正确但范围错不得归成已创建 |
| InputMutation | 返回后改 root/原 leaf 的 descriptor/payload，不改变已保存结果或后续合法结果 |
| OutputOwnership | 对根 Schema/Created/Kind/Reason/三个 false 字段及全部非null Plan 标量分别 held-view附加/突变，保留突变证据且 A/B/输入独立；跨调用控制区分 held-view 和 fresh reread，不把 raw String held 保留说成字段重读保证 |
| SequentialIsolation | normal→refused→normal 及 A/B 反向突变；完整 literal 不变，无全局输出缓存；固定真/假值及 Int64 也检查 metadata 不串扰 |
| RefusalShape | 对各拒绝分类核对根全部字段/类型，Created=false、Plan=null、保守 flags；internal-error 仅静态追踪，不添加可执行故障注入缝 |
| ConservativeFlags | 全体成功/拒绝向量 SourceVerified/LiveActionsPerformed/RuntimeApproved=false；Kind 固定，无新增能力/验收字段 |
| Complete | 只在当前所选组都完成时输出末尾 PASS；LIMIT 固定说明仅声明整数、无真实时钟/运行证据 |

大整数独立 literal（Origin=Observed=0、Stop=null）：

| Frequency | DeadlineTick | StopAtTick |
| ---: | ---: | ---: |
| 9007199254740993 | 1080863910568919160 | 1022317115413102705 |
| 76861433640456465 | 9223372036854775800 | 8723772718191808777 |

Until 仍全部 null，Expired=false。下一频率 76861433640456466 的 120000ms delta
不能落入 Int64，必须 range-refused。上述 literal 是本轮文档精确整数复算，不是产品测试结果。

测试 helper 的固定接口：New-TBRequest([object]$OriginTick=[long]0,
[object]$Frequency=[long]1000,[object]$ObservedTick=[long]0,[object]$StopTick=$null,
[object]$Schema=[int]1) 仅构造五字段对象，object 参数保留故意错误类型；
TB-Need([bool]$Value)、TB-Case([string]$Name,[scriptblock]$Body)、
TB-Equal([object]$Actual,[object]$Expected,[int]$Depth=0)。Name 不含 P2A 前缀，由
TB-Case 添加并冻结行输出。脚本块仅在测试定义/执行，不进入产品输入或实现。
例如 NormalStop 必须包含如下独立 literal，而非调用产品的结果构造 helper：

```powershell
TB-Case NormalStop {
    $actual=New-OfflineTimebasePlan (New-TBRequest ([long]0) ([long]1000) ([long]4) ([long]2))
    $expected=[pscustomobject]@{
        Schema=[int]1;Kind='timebase-plan-only';Created=$true;Reason='plan-created'
        Plan=[pscustomobject]@{
            OriginTick=[long]0;Frequency=[long]1000;DeadlineTick=[long]120000
            StopAtTick=[long]113500;NaturalUntilTick=[long]502
            CleanupUntilTick=[long]5502;PublishUntilTick=[long]6502
            ObservedDeadlineExpired=$false
        }
        SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false
    }
    TB-Need (TB-Equal $actual $expected)
}
```

### Exact implementation decisions

公开参数保持 [object]，先完整准入五属性和精确类型，再判断数值范围/关系。
读取类型使用固定 SDK 元数据，不能调用输入覆盖的 GetType/ToString/getter。
准入/拒绝正常返回都经同一个固定根/Plan 构造路径；内部异常采用保守根，不泄漏异常。

算术 helper 使用 BigInteger 的 Divide/Remainder 得到 q=F div1000、r=F mod1000；
分别计算并检查 q*ms、r*ms，floor=q*ms+(r*ms div1000)，ceil 仅在余数非零时加1。
每个逻辑结果均检查 0..Int64.MaxValue 后才输出精确 Int64；加/减也先宽整数计算
再显式范围拒绝，不使用 PowerShell 可能自动提升为 double 的 Int64 乘除。
Deadline、StopAt 和每个 Until 按设计 §5 次序计算；s=min(StopTick,D)，再计算
s+floor500 等所有指定分支并检查，不能因最终 min 较小而跳过溢出。

所有输出 scalar 独立拥有：Int32/Int64/Boolean 使用本地 typed-array storage 的新
装箱值，非空固定 String 重建新字符存储，固定 SDK 元数据只在调用内使用。
不复制输入 metadata、不用序列化/全局 type-data/cache。不对输入调用任意方法；
可只读参考既有 owned-return 构造经验，但不能导入或修改旧代码来共享返回对象。
输出所有权测试必须确认突变真的保留，不能把 PowerShell raw String 的短暂包装视图
当成跨字段读取持久性。任何 fixture/断言不成立按失败停止，不临时重写测试求通过。

两个指定 mutant 不改 helper 通用规则/预期。选择 natural500 的 floor→ceil，
因为 F1001 时该差异可观察；不选择会被 StopAt min 的另一项遮蔽的冗余改动。

## Approved finite allowance — 2/5 targets, 0/1 reviewer; remainder FROZEN after slot2 failure

固定已有宿主 C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，
SHA256=362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139。
每个目标 -NoProfile -NonInteractive -File <本目录测试> -Group <指定组>，继承普通用户
令牌，max30000ms 执行+2000ms owned exit+2000ms stream drain；不加时、不提权求通过。
CONTEXT/HOST 在断言前输出，HOST 冻结为7.6.5/Core/x64；只有 CONTEXT 用户名占位符
在 Prepare 时根据普通父上下文解析；不修改其他期望。

| 顺序 | 模块状态 / Group | 必须结果 |
| --- | --- | --- |
| 1 red | absent / all | exit1，精确3行，仅 ApiRequired assertion，无 PASS |
| 2 full | good / all | exit0，24PASS/27行 |
| 3 phase-mutant | 500ms delta Ceiling=false→true / phase-rounding | exit1，精确4行，API PASS后指定 FractionalPhaseRounding assertion；立即恢复 |
| 4 expiry-mutant | Observed>=D 改为 > / deadline-equality | exit1，精确4行，API PASS后指定 DeadlineEquality assertion；立即恢复 |
| 5 final | restored good / all | exit0，与 slot2 同有序27行 |

Prepare 不导入产品，冻结公共输入/六文件/独立 oracle 及继承证据；模块此时必须 absent。
SealProduct 仅在精确 RED 后冻结 good 产品和 GREEN 身份，不执行目标，不回写 RED。
每次 Run 持久预消费后才 launch，顺序及前项精确匹配是硬门槛；每个 mutant 在运行前
冻结实际源码摘要/固定差异，返回后恢复 good。捕获 stdout/stderr 原字节并发且独立，
stderr 必须空、精确有序行、owned exit/complete drain/保护链/模块身份一致才承认结果。
失败索引/缺失/漂移/超时/未知上下文/清理不明全部停止；超时只处理精确 owned 句柄，
不扫描/结束其他 PID，未证实清理记 unknown。Git-only 权限与普通测试权限分开。
源/测试/期望/controller 在相应 freeze 后不为失败修正；无额外产品导入、warmup、PS5、
旧套件重跑、benchmark、根本不同 mutant、Go/Docker/服务/安全查询或 nonce/运行包。

## Task 2: 一个条件限定只读审查与交接

**Consumes:** Task1 的已恢复 good 模块、独立测试/预期、控制器、完整五次原始结果、设计和本计划。
**Produces:** 源码/证据限定判断、每个 finding 和 declined-to-judge 的处置；不是验收或真实许可。

特别披露：CodeProperty 元数据拒绝与普通 recoverable internal-error 分支仅要求静态
追踪；本预算不动态编译 getter 或强制内部异常，不把它们说成已执行覆盖。

- [ ] **Step 1 — request exactly one fresh readonly reviewer only after all five precise results.** 推荐 Native 执行后一次 fresh-context 审查，范围是这份新交付的全部源码/测试/controller/oracle、设计/计划及确切五次保存证据；不是整个历史分支。按实际可用模型/技能的最高判断能力派发，禁止 imports/runs/edits/nested agents/安全或网络查询。继承旧条目仅核摘要，不读无关历史内容。
- [ ] **Step 2 — reconcile every finding and limitation.** Critical/Important 或审查不完整即停止，无自动修复、额外测试、召回/替换 reviewer。Minor 和拒绝判断逐项记录父代理处置，不悄悄改代码。有限向量、普通可恢复 internal-error 静态路径、灾难恢复、并发/敌对 type-data、真实时基/调度/身份/OS/采集/清理/持久化及 PS5 明确保留，不用审查无 blocker 推定全局保证。
- [ ] **Step 3 — preserve final state.** 结果/审查/本计划勾选、current-status/recovery 记录实际预算与未完成门槛；必要文件普通提交推送原开发分支、核对远端 SHA 和 clean。原始 packet 不提交、不删除；不自动正式验收、提升到其他入口、进入 B/C/P3、合并 main 或授权真实运行。

## Historical preparation self-review and execution-method handoff

设计 §1–4/§6–10 作为边界继承；仅实现 §5 数值单元，§9 的 P2-A 类别均在 Task1
具名组及两个 mutant 中落实；不承诺动态触发灾难/内部错误。五条 Review Focus
分别映射完整 literal/范围、溢出、被动性、时间等值、返回所有权和保守 flags。
完整组数人工核对为24，绿色27行，RED3行，mutant4行；大数用独立精确算术复算。
公私接口名/字段类型一致，无真实适配器占位开关；实现细节只规定必须精确的算法，
不在本计划写出产品或测试全文。只有一个实现单元及一个审查交接，不拆成重复上下文任务。

本轮仅新增本计划，更新 current-status/recovery；未创建模块/测试/控制器/raw packet，
未运行产品/测试或派发 reviewer。源代码和全部旧额度仍保持原状。
推荐 **Native（由当前代理直接实现）**：这是接口高度相关的单个算术模块，直接实施
并在最后做一次独立限定审查即可；不需要分代理并行实现。若选择分代理，应先另定
其派发/审查预算，不默认在本计划的0/1席位外增加代理。
请用户审阅计划并明确选择执行方式、批准最多五个新普通 PS7 目标和一次条件审查、
指定变异/恢复及 fail-stop 条款后再实施。当前答复只批准了设计和计划编写。
