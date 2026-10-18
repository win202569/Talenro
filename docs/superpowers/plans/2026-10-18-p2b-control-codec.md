# P2-B Pure Control Frame Codec Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Native is recommended here; implementation awaits explicit plan/method/budget approval.

**Goal:** 实现固定256字节控制帧的纯编解码与声明绑定校验，不产生真实身份、会话能力或启动许可。

**Architecture:** 一个无副作用模块处理完整帧，先验证精确被动类型再独立复制和编码/解析；独立literal字节预期是主要oracle。声明绑定校验单独返回结果，不维护会话状态、接收队列或真实计时器。

**Tech Stack:** 普通 Windows PS7.6.5 Core x64、固定SDK元数据、byte[]与精确整数；无新依赖、编译、原生调用或外部I/O。

**Spec:** [已批准P2-B设计§5/§6/§11](../specs/2026-10-18-p2b-owned-session-design.md)。
设计提交2a27ceccbb995711033119584c8e19957d4429ae后，用户明确回答“是”，仅批准进入本计划编写。
设计物理SHA256=41BD5F7225D65D7BAE903E693946A2595C5036A2408B16BECE37965C622E36CE。
设计起草状态留作历史，不回写冻结内容。

状态：PLAN_PENDING_APPROVAL；ImplementationApproved=false；RuntimeApproved=false。
拟议5次NEW普通测试宿主+1次条件限定只读审查，当前0/5、0/1，均未批准/执行。

## Global Constraints

- 复用现有隔离worktree及codex/c12-b01-task9-coordinator，不新建worktree或合并main。
- 产品模块只定义函数；无load-time动作、Process/Job/时钟/外部Stream/文件/网络/安全API、callback或产品导入。
- 精确256字节；全部多字节整数unsigned little-endian；逐kind长度/常量/方向/Observation组合及全部零保留区以设计§6为准。
- Profile=1仅协议声明；不实现W/L/C/T、资源表、创建/恢复/停止、心跳计时、队列/分片传输或清理证明。
- 不改P2-A已验收产品350CAB1A/测试B5C4E0CE/fixtureC720E247/oracleDA3CF183，不导入它们作为编解码逻辑。
- SourceVerified、LiveActionsPerformed、RuntimeApproved在每个结果上固定false；Accepted只是纯数据操作成功。
- 无PS5/完整C12/P2-C/P3/安全放行/真实运行/入口提升授权；所有旧已用/冻结额度保持原状。
- 获准后，精确有限路线覆盖通用技能的额外baseline/全项目/修复/重跑/替换审查循环；任何异常停止，临时mutant先恢复再封存。
- 本轮只写计划/进度/恢复文档，不提前创建下述代码、fixtures、控制器或packet，不生成真实运行身份/命令体。

## Review Focus

1. UInt64超过2^53及最高位不能经过double/有符号移位损坏；HelloLiteral、SequenceExact及Encode/DecodeAllKinds核对固定字节和精确CLR类型。
2. 编解码双方共享错误不能靠round-trip隐藏；12种独立完整hex帧和逐字节reserved反例为主要oracle，RoundTripSecondary仅补充。
3. 假/未知退出不升级为exit0，C不能伪造W观察；ObservationSemantics逐组合比较Status与Value，所有权限flags均false。
4. 输入getter/伪GetType/数组Length或Rank装饰不能执行或改变准入；PassiveDescriptors/DecoratedScalars/ByteArrayShape记录真实可观察反例和零callback。
5. 输出byte数组、标量metadata及连续调用不能串扰；InputMutation/OutputOwnership/SequentialIsolation确认突变真实保留，保留raw String held-view与fresh-reread差别。

## File map and public interfaces

唯一新代码目录：docs/diagnostics/isolated-fuzz/p2b-control-codec/。

| 文件 | 职责 |
| --- | --- |
| control-frame.ps1 | 三个public入口、固定类型/帧/payload验证和独立返回helper |
| control-frame.tests.ps1 | 三入口测试，Group=all/reserved-tail/sequence-equality |
| fixtures.json | 12种完整256字节literal hex与独立预期字段；UInt64以十进制字符串保存，避免JSON数值精度丢失 |
| expected-output.json | RED、完整32PASS/35行和两mutant的有序输出oracle |
| capture.ps1 | Prepare/SealProduct/Run，固定顺序、预消费、身份和有界原字节捕获 |
| source-pins.json | 设计/批准计划/保护链/引擎/阶段身份与拟议额度 |
| README.md | 仅本单元范围、实际结果入口、已消费状态与换机限制 |
| .gitattributes | 新目录ps1/json/md均text eol=lf，不改旧目录规则 |

另新增docs/roadmap/2026-10-18-p2b-control-codec-execution.md，内含实际审查及限制处置；
仅更新current-status/repository-recovery和本计划的实际状态。旧计划及失败不回写。

公开签名：
- ConvertTo-P2BControlFrame([object]$Message)
- ConvertFrom-P2BControlFrame([object]$Bytes)
- Test-P2BControlBinding([object]$Message,[object]$Expected)

Message为精确PSCustomObject，恰好12个NoteProperty：
Version UInt16=1；Kind/Sender/Receiver各Byte；BindingDigest精确一维byte[32]；
SessionId精确一维byte[16]；Sequence UInt64>=1；
OriginTick/Frequency/DeadlineTick/ObservedTick各Int64；
Payload精确一维byte[]，长度必须等于该kind在设计表中的PayloadLength。
拒绝额外/大小写变体字段、可执行成员、装饰标量/数组、null与隐式强转。

Expected为精确PSCustomObject，恰好13个NoteProperty：
Schema Int32=1；Sender/Receiver Byte；BindingDigest byte[32]；SessionId byte[16]；
NextSequence UInt64>=1；OriginTick/Frequency/DeadlineTick/StopAtTick Int64；
ArmSequence/StopSequence/LastNonHeartbeatSequence UInt64（0表示尚无对应声明）。
普通字段/数组也按Message相同被动规则准入。Expected不包含句柄或“已授权”布尔。

统一结果精确10字段：
Schema Int32=1；Mode String='p2b-codec-only'；Accepted Boolean；
Reason String；Bytes（精确byte[]或null）；Message（上述Message或null）；
NextSequence（UInt64或null）；SourceVerified/LiveActionsPerformed/RuntimeApproved Boolean=false。
Encode成功Reason=encoded且仅Bytes非null；Decode成功Reason=decoded且仅Message非null；
Binding成功Reason=binding-matched且仅NextSequence非null。拒绝Accepted=false，后三载荷均null。
Reason其他枚举：invalid-input、invalid-frame、binding-mismatch、sequence-mismatch、
sequence-exhausted、reference-mismatch、expired-declaration、internal-error。
普通内部异常保守返回internal-error，不泄露异常；灾难恢复不在保证内。

### Validation and implementation decisions

类型/描述符必须通过固定SDK元数据核对，先于值/长度使用，不调用输入GetType/ToString/getter。
数组准入用真实CLR byte[]身份及固定SDK长度/维度访问，不用可被shadow的Length/Rank。
拒绝实例扩展成员；稳定输入先复制，返回数组和scalar均重新拥有，非空String重建字符存储。
复用已验证的固定元数据/typed-array经验但不导入旧产品或共享其返回缓存。

读写LE数值采用非负BigInteger乘除256，检查位宽后返回精确UInt16/UInt32/UInt64；
禁止double、BitConverter平台端序依赖和会截断高位的PS有符号移位。
小型私有helper统一P2BC名称段，只负责Read/WriteLE、被动快照、声明边界、payload及独立返回；
不得引入通用对象序列化器或callback接口。尾零检查的精确签名为
Assert-P2BCZeroBytes([byte[]]$Bytes,[int]$Start,[int]$Count)，只接受已复制的内部数组，
范围必须在数组内，Count=0为空区间且无输出，非零字节抛内部拒绝标记；成功不输出对象。
其他helper签名可按这几个固定职责拆分，但不新增public入口或扩展输入合同；
dot-source可见的helper不是隔离/权限边界，不支持外部绕过public准入调用。

声明时基验证：Origin>=0、Frequency>=1000、Observed>=Origin，四字段均<=Int64.MaxValue。
按P2-A既定商/余数、每个中间值范围检查计算D=Origin+floor(F*120000/1000)和
StopAt=min(Origin+floor(F*113500/1000),D-ceil(F*6500/1000))；D必须与帧Deadline一致。
这只是算声明整数，不读取真实QPC。StopAt在Expected/Arm中必须等于计算值。
Observed可以晚于D，以表示迟到故障；不能把这种消息当成按时许可。

Encode先完整验证Message，再构造新256字节零数组，按设计固定偏移写字段/payload；
不允许任意Header/Reserved输入字段。Decode先验证Bytes精确类型/被动性，
再复制恰好256字节，完整验证magic/version/所有保留区/长度/padding/payload/角色，
之后才公布独立Message。Bytes错误类型/装饰为invalid-input；二进制内容坏为invalid-frame。
Encode或Binding的普通对象结构/声明字段非法为invalid-input。
允许表示ReleaseObserved.Result=2/3等合法失败消息；Accepted不等于允许继续运行。

Binding先验证两对象，再依次比较方向、digest/session和时基绑定；
不一致为binding-mismatch；Sequence不等Expected.NextSequence为sequence-mismatch；
相等但为UInt64.MaxValue时sequence-exhausted，不计算溢出的next。
否则核对Arm/StopAck引用、Heartbeat.LastNonHeartbeatSequence与Expected，
不一致为reference-mismatch；Arm声明Observed>=StopAt为expired-declaration；
最后仅返回独立NextSequence=当前Sequence+1，不修改Expected或保存会话状态。
ArmSequence=本帧Sequence；ArmAck/ReleaseAttempted/ReleaseObserved须引用Expected非零ArmSequence；
StopAck须引用Expected非零StopSequence；Heartbeat引用Expected.LastNonHeartbeatSequence且<本帧Sequence。
其他kind的Expected引用可以为0，不凭其存在推定已绑定/已释放。

静态合法性检查适用全部设计§6 kind/payload/方向/Observation条件；
方向精确展开：Hello/Bound/Heartbeat/Fault为1→2、2→1、2→3、3→2；Ready仅3→2，
Arm仅2→1；ArmAck/ReleaseAttempted/ReleaseObserved仅1→2；Stop为1→2、2→1、2→3；
StopAck为1→2、2→1、3→2（没有2→3）；Observation仅3→2或2→1，且满足payload角色限制。
真正生命周期状态、队列额度、实际超时、端点身份与票据原子消费留给未来能力单元。
同一Message/Expected重复调用纯函数可以返回同一结果，不宣称跨调用防重放或一次启动。
分片/255/257字节只拒绝；组帧、异步等待和16槽队列不是本单元。

## Task 1: Pure codec and five bounded verification hosts

**Consumes:** 上述Message/Bytes/Expected和设计§6固定布局。
**Produces:** 三public入口、严格独立结果、五轮有界回执；无真实会话或运行能力。
**Files:** 创建上表8文件；实现开始前模块必须不存在，其余7输入先冻结。

- [ ] **Step 1 — prepare independent tests and oracle.** 写全部32组测试、12份完整hex、
  控制器和metadata。先静态核对layout/payload常量和fixture构造源码，再冻结；
  实际getter/metadata突变的可观察性在相应full测试内证明，不另行导入/执行预检。
  所有数值输入显式CLR类型；纯测试helper BC-Need([bool])、BC-Equal($Actual,$Expected)、
  BC-Case([string], [scriptblock])负责断言、完整类型/字段比较和Group筛选。
  BC-Equal不得调用产品，类型反射receiver用BaseObject；byte[]逐字节独立比较。
- [ ] **Step 2 — run RED once.** capture Run red，模块不存在时仅检查三public函数缺失；
  exit1，context/host后ASSERTION FAILED: P2BCodecApiRequired，精确3行，stderr空。
  模块已存在或任何非预期结果即停止，不自行换测试/重跑。
- [ ] **Step 3 — implement module.** 按上文签名/返回/优先级及设计表实现，不新增额外入口。
  全部普通数据快照/校验完成才返回成功；无I/O/load-time动作。静态parse、exact
  fixture字节和独立oracle核对后SealProduct冻结good身份，不导入试跑。
- [ ] **Step 4 — run full GREEN once.** capture Run full -Group all，必须exit0、
  下列32个精确PASS、35行、stderr空/普通上下文/完整流/owned exit/身份完整。
  任一失败先停止；不在剩余slot中修复。
- [ ] **Step 5 — reserved-tail mutant then immediate restore.** 唯一尾部零检查调用固定写为
  `Assert-P2BCZeroBytes $snapshot 224 32`；仅把Count32改0（helper的0长度是空区间）。
  Group reserved-tail仅API和ReservedZero；必须exit1、4行，失败ReservedZero。
  立即apply_patch恢复good字节/摘要；不以任意失败当命中。
- [ ] **Step 6 — next-sequence mutant then immediate restore.** 唯一声明序号等值拒绝条件
  `$message.Sequence -ne $expected.NextSequence`仅改为`-lt`。
  Group sequence-equality仅API和SequenceExact；actual8/expected7的独立反例必须
  exit1、4行，失败SequenceExact。立即恢复good身份；非预期冻结余项。
- [ ] **Step 7 — restored final full GREEN.** 同good模块/测试/fixtures/oracle身份再次
  all→exit0、32PASS/35行。不能用先前GREEN代替还原后结果。
- [ ] **Step 8 — reconcile and checkpoint.** 逐行及hash核对五轮、保留所有原记录，
  记录每个失败/未测项。只保存必要源/fixture/报告及恢复文件到现有dev分支，
  普通提交推送并核对远端。完成条件不是正式验收/main/下一能力阶段许可。

### Independent literal fixture and assertions

主要Hello基准：Version UInt16=1、Kind/Sender/Receiver Byte=1/1/2；
BindingDigest bytes00..1F，SessionId bytesA0..AF；Sequence UInt64=72623859790382856；
Origin/Frequency/D/Observed Int64=10/1000/120010/11；
payload16字节，前8零，Profile LE1、FrameBytes LE256。
完整期望帧由下列128字节前缀+128个零字节组成（不是调用encoder生成）：

```text
50 32 42 43 01 00 01 01 02 00 00 00 00 00 00 00
00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F
10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F
A0 A1 A2 A3 A4 A5 A6 A7 A8 A9 AA AB AC AD AE AF
08 07 06 05 04 03 02 01 0A 00 00 00 00 00 00 00
E8 03 00 00 00 00 00 00 CA D4 01 00 00 00 00 00
0B 00 00 00 00 00 00 00 10 00 00 00 00 00 00 00
00 00 00 00 00 00 00 00 01 00 00 00 00 01 00 00
```

Step1须将每种kind各自的完整512hex字符独立写入fixtures.json，不能借产品生成，
不能仅存“encode后decode”预期。其余kind固定header取Sequence7、Origin0/F1000/D120000/
Observed4、相同digest/session；方向选合法的一项，payload按设计表的字面常量：
Bound1→2、Ready3→2、Arm2→1(ArmSeq7/ReleaseBefore113500)、
ArmAck/Attempt/Released1→2(引用Arm3，Released.Result1/Error0)、
Heartbeat1→2(LastNonHeartbeat6)、Stop2→1(Cause12/FirstFailureSequence0)、
StopAck1→2(引用Stop2)、Observation3→2(Kind1/State3/Subject3/Status1/Value0)、
Fault3→2(Cause10/Step3/Error5)。Bound.Checks7及Ready六常量不计算自产品。
独立decode预期为完整12字段/精确类型，不只核Kind或round-trip。

示例断言（只使用Step1已定义的BC-Need，literalBytes/literalMessage来自独立fixture）：

```powershell
$encoded=ConvertTo-P2BControlFrame $literalMessage
BC-Need ($encoded.Accepted -and $encoded.Reason -ceq 'encoded')
BC-Need (BC-Equal $encoded.Bytes $literalBytes)
$decoded=ConvertFrom-P2BControlFrame $literalBytes
BC-Need ($decoded.Accepted -and (BC-Equal $decoded.Message $literalMessage))
```

完整有序组（前缀P2BCodec；每组成功才写PASS）：

| Group | 必须断言的独立向量 |
| --- | --- |
| ApiRequired | 三public函数存在；无模块RED只在此失败 |
| HelloLiteral | 上述高于2^53的完整LE字节及逐字段类型 |
| EncodeAllKinds | 12份独立Message→完整literal字节、完整成功根形状 |
| DecodeAllKinds | 12份literal→完整Message、精确UInt/Int类型 |
| RoundTripSecondary | 12种往返不变，仅辅助已有独立oracle |
| ExactMessageShape | null/hashtable/额外/缺失/小写变体字段拒绝；Expected13字段也核对 |
| ExactTypes | Byte/UInt16/UInt64/Int64的错误数值类型、string/double/bool/null拒绝，不强转 |
| PassiveDescriptors | 可观察ScriptProperty/AliasProperty/ScriptMethod/GetType/ToString，产品callback0 |
| DecoratedScalars | 每类numeric及数组的附加metadata真实保留后拒绝；不原地清除 |
| ByteArrayShape | digest31/33、session15/17、非byte[]、多维数组、Length/Rank装饰拒绝 |
| FrameLength | 0/255/257/512字节拒绝；不截取第一帧 |
| HeaderMagicVersion | 每个magic字节及Version0/2拒绝 |
| ReservedZero | 分别9..15、106..111、224..255逐位置非零拒绝；mutant包含单独byte255=1反例 |
| PayloadLength | 各kind固定长度±1、113以及payload-padding非零拒绝 |
| PayloadConstants | HelloProfile/256、Bound7/reserved、Ready7/1048576/1048576/2097152/4096/65536逐常量反例 |
| MalformedPayloadRange | Subtype非0、prefix保留、Cause0/14、State11、Step非法和payload尾部非零拒绝；ReleaseObserved.Result1/2/3均是可表示数据且flagsfalse，0/4拒绝，Error0不改变Result2/3含义 |
| DirectionRoles | 四合法边、各kind方向；T=4/未知role及C伪Arm/释放拒绝 |
| ObservationSemantics | 全4种Observation及Status1/2/3；未知退出Value0保留未知；C伪J/释放和矛盾Value拒绝 |
| SequenceExact | 等7通过Next8；实际6/8均sequence-mismatch；独立UInt64最高位序号9223372036854775808无损 |
| SequenceExhaustion | UInt64.MaxValue可编码/解析但相等Binding为sequence-exhausted；拒绝0，不wrap |
| ArmReferences | Arm自引用、非零关联Arm3、错误/零引用拒绝 |
| StopReferences | StopAck引用Stop2；零/错误引用拒绝；本地FirstFailureSequence0允许 |
| HeartbeatReference | Last6/Sequence7通过；Last7/8拒绝；Expected不符reference-mismatch |
| Timebase | 0/1000/D120000与10/1001/D120130/Stop113623；负数/低频/错D/中间溢出拒绝，Fault迟到Observed120001允许表示 |
| ArmEquality | ArmObserved113499可Binding；113500/113501为expired-declaration，不影响合法Stop/Fault编码 |
| BindingMismatch | digest/session任一字节、方向、Origin/F/D错配拒绝，Expected不变 |
| FragmentRefusal | 两段各128字节单独调用均拒绝，随后完整256仍独立成功；不保存拼接状态 |
| InputMutation | 返回后改原Message/Payload/Bytes/Expected不能改变保存结果或后续合法结果 |
| OutputOwnership | 两结果根、Message、三个byte数组、所有scalar held-view附加/突变互不串扰；数组必须保留变更，不接受“未突变”假通过 |
| SequentialIsolation | encode→refused→decode→binding，A/B反向metadata/数组突变不污染下一调用 |
| ConservativeFlags | 全成功/拒绝的10字段根、载荷互斥/类型和三false；Accepted不能被当capability |
| Complete | 本组选定全部完成、callback0才PASS；固定LIMIT说明仅协议声明 |

CodeProperty和ordinary internal-error路径只静态追踪，不新增原生编译或故障注入缝；
灾难恢复、并发/敌对type-data、真实端点/时基/清理、完整状态机和队列明确未测。
测试自身fixture若共享/不可观察、类型或字节预期错，即非预期停止，不修测试求通过。

## Proposed finite allowance and controller contract — NOT APPROVED

固定引擎沿用已知路径及SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
-NoProfile -NonInteractive -File <tests> -Group <指定组>，普通令牌，无PS5。
本次拟议每host最多60000ms执行+2000ms owned exit+2000ms并发分流drain，
不延长/重试；这是测试宿主预算，不更改未来真实运行的120000ms合同。

| Slot | 模块/Group | 精确预期 |
| --- | --- | --- |
| 1 red | absent/all | exit1，3行，只有API assertion |
| 2 full | good/all | exit0，32PASS/35行 |
| 3 reserved-tail | 单处尾零检查Count32→0/reserved-tail | exit1，4行，API PASS后ReservedZero assertion |
| 4 sequence-equality | 单处Sequence -ne→-lt/sequence-equality | exit1，4行，API PASS后SequenceExact assertion |
| 5 final | restored good/all | exit0，与full相同有序35行 |

每个完整输出context/host先写，PASS固定P2BCodec前缀；
末行LIMIT: protocol declarations only; no identity, capability or runtime approval。
RED/mutant不输出LIMIT。预期不得从当前产品输出动态学习。

新ignored packet：.superpowers/sdd/2026-10-18-p2b-control-codec/，必须此前不存在。
Prepare保留上一type-receiver-repair的839个VerificationPath映射、不改写为旧Path；
新增该packet16个anchor（manifest/seal/index/old/newtests/baseline/十个原始流），
本设计1、批准计划快照1、P2-A固定四核心身份4、新目录模块外7个输入，
合计868映射项（不是868个唯一文件）。批准计划原文先核SHA再复制到新packet；
后续本计划状态更新不修改已冻结快照。任何历史映射缺失/漂移先停止，不探测补建。
模块在RED后通过SealProduct独立冻结；每个mutant有摘要/精确差异，恢复核goodSHA。
LF物理/Git过滤字节须一致，冻结后不改源/test/oracle/controller去解释失败。

Run在启动前持久预消费，强制顺序/前项精确结果/普通上下文/引擎/全部pins一致，
stdout/stderr原字节分别并发保存。只处理当前拥有的确切Process句柄，不扫描其他PID。
host超时、stream不完整、stderr非空、输出不精确、身份漂移、清理不明全部停止；
原始流/索引不删除或改判。无额外warmup、产品import、旧套件/全仓库/Go/Docker重跑。

## Task 2: One conditional readonly review and handoff

**Consumes:** 同一还原good模块、测试、fixtures/oracle/controller、全部五轮精确回执和设计/计划。
**Produces:** 一次限定源码/保存证据审查、全部finding/declined处置和可移植交接，不是验收。

- [ ] **Step 1 — dispatch exactly one fresh readonly reviewer only after all five expected outcomes.**
  限定本新单元源码/测试/控制器/oracle和五轮证据，继承项只核hash不审无关历史；
  禁止imports/runs/edits/新增目标/嵌套代理/安全或网络查询。
- [ ] **Step 2 — reconcile findings and limits.** Critical/Important/审查不完整则停止；
  无自动修复/补测/替换reviewer。逐项处置Minor及declined，不把无blocker当真实能力证明。
- [ ] **Step 3 — save necessary progress.** 更新执行报告/本计划实际状态/current-status/recovery，
  明确入口、固定版本、已用额度/停止位置/未测项；只提交明确名单并推送现有dev分支，
  核对远端与clean。原packet保留不提交；不自动验收/main/真实能力或下一阶段实施。

## Self-review and execution-method handoff

覆盖边界：设计§6帧/静态kind语义由Task1实现；§5/7—10真实能力/状态/计时/清理和§6
传输组帧/4槽队列不属于首单元，明确留在后续独立计划，不用纯codec结果代替。
五条Review Focus各映射上述具名测试；32组=35完整行，RED3，mutant4。
接口12/13/10字段、三入口名字、Reason和输出载荷互斥、精确类型前后统一；
literal前缀128+零128=256，尾部检查和Sequence反例各有指定单处mutation。
没有运行代码或调用产品来生成计划的预期。

仅本计划及两份进度/恢复文档改动；设计批准包括L唯一启动代理的职责细化，
不改变旧P2-A验收或任何运行门槛。实施计划本身仍未批准。

推荐 **Native（当前代理直接实施）**：编解码和绑定校验共享同一固定类型/布局，
作为一个高度相关单元实现后一次独立限定审查即可。若选Subagent-driven，
其额外实施/任务审查席位需重定预算，不能默认挤入本5+1路线。
请用户审阅计划，选择方法并明确批准有限5+1预算后再实施；当前只保存计划。
