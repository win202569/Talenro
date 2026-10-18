# P2-B Inbox Tree Copy Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for recommended Native execution. Use superpowers:subagent-driven-development only with a separately approved replacement budget. Track checkboxes in order.

**Goal:** 修复新inbox自身构造快照被误送入标量复制分支的问题，保持独立所有权和被动入口。
**Architecture:** 仅修改Copy-IBTree及一个私有节点分类helper，分离原始包装器身份、底层类型和对象字段所属wrapper；不改codec或接收语义。新的专项测试加原32组回归，再做两项指定变异和还原完整验证。
**Tech Stack:** 普通Windows PS7.6.5Corex64、固定PowerShell SDK、原模块/固定CLR元数据；无产品编译、真实I/O或新依赖。
**Spec:** [已批准inbox设计](../specs/2026-10-18-p2b-bounded-inbox-design.md)§3所有权/§6快照/§7验证；SHA43036F17E712CA2AA57F51CA09BAFF874DB53F954FADF82CBCD6388763B2585B。

状态：UNFINISHED_REVIEW_BLOCKED_PORTABLE_UINT64。用户在8f0d8ba2后批准只释放未启动3轮
及条件审查，不重跑RED/full、不改冻结输入。此前容量缩进错误已精确还原，
历史前置拒绝记录保留。容量/序号mutants均指定断言，随后最终组合49行/
44PASS，与原full一致。ImplementationApproved=true；RuntimeApproved=false。
NEW5/5SPENT，0remaining；只读review1/1完成，Critical0/Important1/Minor1。
Important：可携带摘要的UInt64边界值被代理JSON转换舍入；原始记录与哈希
正确不变，但证据交付未完成。Minor：Count.Probe隔离未直接断言，接受延期。
未自动修复/补测/更换reviewer；完整审查及5项declined处置已保存。
原2/5SPENT+3FROZEN及独立诊断1/1SPENT不变。
实测/处置见[执行记录](../../roadmap/2026-10-18-p2b-inbox-tree-copy-repair.md)。

## Global Constraints

- 复用既有隔离worktree/codex/c12-b01-task9-coordinator；不改main、设置或依赖。
- 保留五public入口、12字段结果根、4字段Snapshot/8字段Endpoint、原Reason顺序。
- 16普通槽/4心跳槽/4组帧缓冲及6144字节原始稳态容量不变；不是CLR总内存上限。
- token仍只在created结果保留原身份，禁止树复制token；无生产恢复/注入/状态导入接口。
- SourceVerified/LiveActionsPerformed/RuntimeApproved始终false；无真实管道/时钟/进程/能力。
- 固定codec9A532E17及其标量拒绝/新值guard不可改；普通数据准入、隐藏描述符/CWT检查不放宽。
- 原目录、测试、源码、回执和诊断永久保留；新目录独立冻结，不修旧失败。
- 意外结果/输入漂移/审查阻塞立即停；不自动改测试、补跑、扩大修复或换reviewer。

## Evidence and allowed repair

[单次诊断](../../roadmap/2026-10-18-p2b-first-snapshot-diagnostic.md)实际捕获
首次snapshot返回internal-error，Unexpected scalar堆栈：
Copy-P2BCScalar:146 <- Copy-IBTree:21 <- New-IBResult:29 <- Snapshot:256。
产品基线SHA43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8。
具体运行时wrapper类型此前未记录，不当作已测事实。

修复前RED须补充独立固定反射观测：使用object[]反射参数保留真实PSObject
引用，记录内部测试构造节点的原始CLR类型和_immediateBaseObject类型。
不通过产品分类helper生成预期。若实测表示不支持拟议包装器解释，
停止而非继续猜修。该观测只能说明对应测试对象，不冒称直接观测了
先前历史失败中已消失的那个运行时对象。

仅允许Copy-IBTree与新增私有Get-IBTreeNode([object]$Value)改变：
- helper返回保留引用的object[3]：Kind String(null/object/array/scalar)、
  NativeValue、OriginalOwner；内部约定，非公开API。
- raw type若是PSObject，仅通过固定WrapperBaseField/Read-P2BCField有界
  最多8层追踪底层类型；null/异常/未终止wrapper链保守拒绝，不调用输入方法。
- 底层精确PSCustomObject走object分支，但必须保留原所属wrapper枚举自身
  可信NoteProperties；不能先用PSCustomObject底层singleton替代owner导致字段丢失。
- 底层精确object[]走array分支，以真实CLR长度/元素逐一递归，new object[n]；
  空数组也new，不借共享空实例，不用可shadow的Length/Rank。
- byte[]及原支持数值/Boolean/String叶仍由原Copy-P2BCScalar复制；
  null保留。未知引用类/hashtable/Version/object叶拒绝，不扩大标量白名单。
- helper只处理本模块已拥有的固定树；不导出、不改变公共输入校验路径。
  这不使任意调用者提供的PSCustomObject脚本属性变成可准入数据。
- New-IBResult、registry、Binding验证、序号/容量/EOF/心跳、codec从原位置起
  保持字节相同。若需要改这些部分才能通过，按新问题停止，不悄悄扩大修复。

## Review Focus

1. wrapper被识别却丢失NoteProperty所属关系：OwnedObjectTree/NestedValues核完整字段/类型/顺序。
2. 空object[]/byte[]和嵌套元素仍共享：EmptyArrays/ReturnedMutation检查实际ETS和byte修改保留但互不污染。
3. 为兼容对象而吞掉未知引用或绕过标量guard：UnsupportedLeaves/ScalarLeaves及原InputPassivity保持严格拒绝/独立值。
4. 首次成功后后续snapshot或故障snapshot仍失败：FirstSnapshot/TwoSnapshots/StoppedSnapshot检查完整literal。
5. 私有测试scope误变生产入口或证据误标：Representation记录只测试构造对象；原ApiRequired/AuthorityBoundary和五导出检查不变。

---

## File map and fixed inputs

实施目录docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/已创建并保存。

| 文件 | 实施时职责 |
| --- | --- |
| inbox.psm1 | 从43067AD0逐字节起步，RED后仅改允许helper |
| inbox.tests.ps1 / test-helpers.ps1 | p2b-bounded-inbox-tick-correction原32组及helper逐字节副本 |
| fixtures.json / descriptor-fixture.cs | 原literal/CounterNote逐字节副本 |
| tree.tests.ps1 | 新12组专项、类型观测和完整独立预期/自有trace，不改原suite |
| target.ps1 | red/focused-full/capacity/sequence/final固定路由和两suite同宿主串行组合 |
| expected-output.json | RED3行、组合49行、mutant4行，预先写定 |
| capture.ps1 | 新5槽预消费、freeze/seal/捕获/失败即停 |
| README.md / .gitattributes | 新范围、预算、结果；psm1/ps1/json/md/cs物理LF |

原32组test SHA904F20D4D60540D7A312F5ACF14B0E1E01D263C7A905B18F578FB0846C7267CE；
原codec加载路径仍../p2b-cold-cache-repair/control-frame.ps1，SHA
9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C。
不重新执行codec独立33组，也不复制改写其scalar helper。
测试内访问私有helper仅限模块scope构造固定测试树/捕获返回和异常，无生产setter。

## Task 1: RED, narrow repair, complete bounded validation

**Consumes:** 固定设计/上述基线、已有诊断及新独立typed预期。
**Produces:** 同五入口/相同结果合同、修复树复制和新5轮原字节回执。

- [x] **Step1 — prepare tests/oracle/controller and freeze.**
  复制旧产品/五项固定测试输入；核SHA。写专项12组和新路由，不import/
  compile/probe产品。静态parse和literal/source自检后Prepare，冻结新模块外
  全输入及baseline-product.psm1快照，旧1289条VerificationPath映射原样继承。
- [x] **Step2 — RED once on43067AD0.**
  capture -Action Run -Slot red。先记录独立测试节点的原始/底层类型及owner；
  重现同Binding创建及digest0=99/session0=99/frequency1001修改，只调用
  一次public snapshot，完整保存ActualResult。必须Accepted=false/
  Reason=internal-error、payloadsnull/flagsfalse，随后精确
  ASSERTION FAILED: P2BTreeFirstSnapshot，exit1/3行。其他错误停止。
  无新的API缺失RED或原full重跑；不启动断点/managed编译。
- [x] **Step3 — implement allowed tree dispatch only.**
  新Get-IBTreeNode按上文签名/三元素合同，Copy-IBTree保留owner并递归独立构造。
  比较完整diff证明其他函数原字节不变。静态parse后SealProduct冻结good，
  同时生成两项准确mutant源码快照及SHA；不导入试跑。
- [x] **Step4 — full once: new12groups then unchanged32groups.**
  capture -Action Run -Slot full。先执行tree.tests -Group all：
  context/host+12PASS(14行)，完成才在同一宿主执行原inbox.tests -Group all，
  原35行逐字节oracle不改。总49行、44PASS，exit0/空stderr；
  两份trace分别保存，均Passed/Errornull/authorityfalse。
  原suite必须CompileCount1/Observabletrue/Counters0/Synthetic4/Decorations8。
  任一失败冻结余3轮/条件审查，不借mutant槽修复。
- [x] **Step5 — capacity mutant and immediate restore.**
  原唯一Count -ge4→-ge5；target仅调用原Group capacity，原4行/
  exit1/ASSERTION FAILED: P2BInboxEndpointCapacity。必须是错误接受第五帧，
  不是internal-error/数组越界。立即apply_patch还原good并核SHA。
- [x] **Step6 — sequence mutant and immediate restore.**
  原唯一Sequence -ne Next→-lt Next；只跑原Group sequence，
  原4行/exit1/P2BInboxSequenceExact错误接受反例。立即恢复good身份。
- [x] **Step7 — final combined full after restoration.**
  完全相同good/专项/32组/oracle，总49行/44PASS/exit0。原fixture仅在两次
  组合full宿主各managed编译一次，无额外编译host/预热。所有pins/恢复身份核对。

### Twelve ordered focused groups and independent expectations

成功前缀P2BTree，组名及断言：

| Group | 必须独立验证 |
| --- | --- |
| FirstSnapshot | 同诊断Binding修改后，public返回inspected/完整初始Snapshot，四方向Count0/Next1/Last0/HBnull/新空数组；冻结绑定未变 |
| OwnedObjectTree | 私有scope构造两NoteProperty对象，逐字段精确复制，raw/underlying类型及owner记录，副本根不同引用 |
| OwnedArrayTree | object[2]包含独立byte[2]{7,9}和typedUInt64高位；长度/元素类型/字节精确且各引用独立 |
| EmptyArrays | object[0]/byte[0]各复制两次，原/两副本引用都不同；ETS Length/Rank修改真实保留，不污染其他副本 |
| NestedValues | Root.Label='probe'、Root.Child.Enabled=true、Root.Child.Bytes={7,9}、Root.Items=[Int32(3),String('x')]；完整字段顺序/类型不丢失 |
| ScalarLeaves | null、Byte7、UInt16(258)、UInt32(65537)、UInt64(9223372036854775808)、Int32(-7)、Int64(-9)、Booleantrue、freshString('probe')、byte[2]{7,9}；值与原型一致且copy guard成立 |
| UnsupportedLeaves | 私有scope未知System.Object/System.Version/hashtable叶拒绝；不当作树scalar接受或返回原引用 |
| ReturnedMutation | 修改首个snapshot的Binding/Queue/Partial/Count metadata和实际byte，后续snapshot完整literal不变 |
| TwoSnapshots | 同token连续取两份独立完整快照，根/端点/空数组/标量metadata无交叉污染 |
| TwoInboxes | 同Binding两个token；A的快照修改/停止不改变B的初始快照，token身份不被复制 |
| StoppedSnapshot | 通过CloseInput触发peer-closed后仍可inspected；首因完整、queued记录保留、后续Feed/Take stopped |
| Complete | 前11组全完成、所有副作用限本host，无新导出/生产测试开关；权限flagsfalse |

初始Snapshot完整literal由原计划精确4字段/每方向8字段/Binding六字段构造，
不调用产品/helper生成expected。根schema/mode等12字段也逐字段核对。
NestedValues只为私有测试scope中的受控自有树，不是公开新增输入种类。
FirstSnapshot在RED先独立记录wrapper表示；若不满足固定SDK下的可解释
PSObject/PSCustomObject表示，停止，不用该轮临时学出的值重写oracle。
GREEN对相同冻结向量执行，其余11组首次运行；不能声称它们各自跑过RED。

## Approved NEW5+1 allowance — exhausted; review blocks evidence handoff

推荐Native，同一代理直接实施，一名新独立只读reviewer仅在五轮全部精确后启动。
本新预算不是原3FROZEN的复用。实际新5host已用，0remaining，审查条件已满足。
以下为原批准边界，不允许补跑。每host最多120000ms执行+2000ms ownedexit+
2000ms并发分流drain；不改真实运行合同。固定PS7.6.5Corex64/adminFalse；
engineSHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139，
SDK SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B，
路径沿用原计划，不自动安装/搜索替代版本。

| Slot | Product / target | Exact expected |
| --- | --- | --- |
| red |43067AD0 /FirstSnapshot |exit1/3行/指定public失败及表示记录完整 |
| full |newgood /tree12+original32 |exit0/49行/44PASS，两trace精确 |
| capacity |oneCountmutant /originalcapacity |exit1/4行/指定capacity错误接受 |
| sequence |oneSequencemutant /originalsequence |exit1/4行/指定sequence错误接受 |
| final |restoredgood /tree12+original32 |exit0/49行/44PASS，和full同oracle |

新packet .superpowers/sdd/2026-10-18-p2b-inbox-tree-copy-repair/必须此前不存在。
Prepare继承1289映射，增加诊断packet全部文件/5portable文件/诊断报告，
前一tick-correction固定11portable与失败报告（若已有映射仍保留），
本设计/批准计划快照、新目录非模块固定输入及baseline快照、engine/SDK。
每条明确role/path/hash并机械核总数；Prepare前完整路径清单单独显示后冻结，
不能漏项/改旧VerificationPath/补建遗失历史。动态数量只是清单计数，不是准许模糊输入。
新product只在RED后变更，seal及good/mutant/两次恢复身份全部记录。
任何历史pin漂移/原始流缺失即停，不重建或重播。

控制器先持久预消费，顺序固定，前一轮必须精确；rawstdout/stderr分开并发捕获，
只拥有当前确切target句柄，无PID扫描/其他进程清理。时限/流/退出/普通上下文/
身份/输出任一异常即停，原回执不改判。只去除最多一个终止换行，不吞空行。
target/composite输出oracle在Prepare前冻结；不根据失败输出学习预期。
无额外原full/codec/仓库/Go/Docker测试或import。unexpectedmutant先恢复good，
随后仍停；不把恢复成功当后续运行许可。新审查未满足条件不得启动。

## Task2: Conditional readonly review and durable handoff

**Consumes:** 相同sealedgood、专项/32组、literal/oracle/controller、五轮原证据。
**Produces:** 一次限定审查、完整findings/declined处置和devbranch保存，不是验收。

- [x] **Step1 — one fresh readonly reviewer after5exact.**
  审允许diff、wrapper/owner/数组所有权/标量拒绝、测试独立性、五轮证据；
  原codec及历史其余材料只hash/必要固定helper边界，不扩大历史wholebranch审计。
  禁止execute/import/compile/修改/额外探测/网络安全查询/嵌套代理。
- [x] **Step2 — preserve and disposition all feedback.**
  Critical/Important/审查不完整停止，无自动修复/补测/换reviewer。
  原样保留所有findings/declined；父代理逐项说明有限成立范围与缺口。
- [x] **Step3 — checkpoint necessary files explicitly (unfinished status).**
  新portable-summary/review.md、docs/roadmap/2026-10-18-p2b-inbox-tree-copy-repair.md、
  本计划实际状态/current-status/recovery，普通提交推送既有devbranch并核remote=local/clean。
  ignored原始记录/缓存/凭据不提交或删除。失败也保存为unfinished，禁止main/
  自动验收。Clone不恢复token/CLR身份/原始packet/预算或运行权限。

## Self-review and execution handoff

修复仅服务既有Spec§3/§6/§7，不改角色/协议/能力合同。
五条ReviewFocus均有具名专项及原suite覆盖；scalar白名单/codec不改；
对象保留所属wrapper，array走原生类型和独立分配，token例外不泛化。
12专项+32原组=44PASS；2专项header+12PASS+原35行=49行；
RED3、mutants4。原32组完整性未被此前两PASS替代，未知后续缺陷仍可能停止。
实际完成Task1Steps1–7；Step5前置拒绝后未启动目标，经用户续批才执行3轮。
Task2只读审查已完成，证据精度Important阻塞，Minor延期，5项declined全部处置。
必要源码、结果、完整审查和恢复文档作为明确unfinished检查点保存。

用户已批准Native/新5+1及停止后的3轮/条件审查续行，现五轮均精确。
唯一只读review的Important已触发停止；不自动修复/补测/更换审查者。
下一步建议显式批准仅无损校正派生摘要并只读逐字比较，0目标/编译/审查，
不改冻结产品/测试/预期/控制器，也不把该校正当成Minor修复或正式验收。
原2/5SPENT+3FROZEN和诊断1/1SPENT始终不变。
