# P2-B Process Association Guard Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for proposed Native execution. Steps use checkboxes. No code or runtime authority follows from a proposed plan.

**Goal:** 修正未启动进程断言及两处同源清理护栏的错误异常前提，重新完成有限离线收束与 seam 验证。
**Architecture:** 在新目录引入一个仅供离线控制器使用的私有进程关联观察函数，测试断言与两处清理护栏共用它。先在无子进程 RED 中复现旧判断的实际错误，再仅修该函数；完整 GREEN 和 FINAL 均使用新冻结清单，原 owned-run/产品/历史证据不动。
**Tech Stack:** Windows、普通 PowerShell 7.6.5 Core x64、固定 SDK、托管反射；无安装、Add-Type、原生编译、安全设置或依赖变更。
**Spec:** [原 I1 收束合同](2026-10-19-p2b-controller-exception-containment.md)、[因果补验合同](2026-10-19-p2b-controller-causality-reverification.md)及[握手合同 §9](../specs/2026-10-19-p2b-handshake-contract-design.md)。仅窄修离线测试控制器，不实现真实握手/能力。

状态：REVIEW_BLOCKED_R1_UNFINISHED。用户在979c9cf1后批准Native/新3宿主+
4子进程+条件1审查，执行基线979c9cf179d0aed0e23d91cbb1efefe181802b90。
原批准快照98EB496FE5ACB02E0675A14116BE18FA099A8A2F731E50E5AEA718D0FF0A60D5不变。
RED精确、仅observer一次修复、GREEN七组精确、FINAL原15行12PASS/完整oracle精确。
唯一审查发现R1 Important：最终记账失败可遮盖首因；R2 Minor：PreStart详情键覆盖。
全部7进程及1审查额度用尽；不修不重跑不二审，旧失败/原始记录/冻结余量不改。
本任务仍未完成/验收，具体证据及逐项处置见同日执行报告和新leaf review文件。

## 已知依据与范围选择

保存的 GREEN 前四组通过，第五组 Unassociated=false，完整运行仍失败。
测试 containment.tests.ps1:150 用 try{$null=$p.Id}catch 推断未关联；
同目录测试:83 和 capture.ps1:106 则把“未捕获异常”推断为已关联/已启动。
这三处依赖同一个错误前提。原 owned-run:20 的文件创建在 Start:21 之前；
保存的 prestart 回执显示 Create 失败、Started=false，但未保存 Id getter 的实际返回值。

[PowerShell 7.6.5 Binders.cs:5027–5040](https://github.com/PowerShell/PowerShell/blob/v7.6.5/src/System.Management.Automation/engine/runtime/Binding/Binders.cs#L5027-L5040)
将一般成员读取异常转成 null；因此外层 catch 是否触发不是进程关联证明。
这是静态原因定位，不是一次新的运行观察，不补写旧 CauseConfirmed/失败字段。
[PropertyInfo.GetValue 文档](https://learn.microsoft.com/en-us/dotnet/api/system.reflection.propertyinfo.getvalue?view=net-10.0)
提供显式方法调用路径；反射异常的实际包装仍须本计划 GREEN 验证，不能凭文档预先宣称通过。

选共享私有观察函数，使三处不再重复不同判断。只把第五组改成 null 比较，
仍遗漏清理护栏，拒绝；原地改冻结目录会损坏历史身份，拒绝；不重写产品生命周期。

## Global Constraints

- 使用当前隔离工作树和 codex/c12-b01-task9-coordinator；PowerShell-only。
  不合并 main、发布、删除历史或原始记录、改 Git 信任/安全设置或安装依赖。
- 固定 owned-run SHA256：7A425C661968CD74BFD8CD4EAA25CBA53ED36CC09E3D2946F9D095FC095B25E1。
- 固定 fixture-child SHA256：EE5CDA1C54739EBADD9578391CCE56B5CA13CDE764EC878D5782EFAE389B251F。
- 固定 codec：9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C；
  inbox：47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC。
- engine：362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；
  SDK：2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B。
- 原 causality-reverification 五输入、18 raw、6 portable、执行报告与原计划不变。
  其余历史源码/原始记录/审查/失败/预算不变，不重分类旧 GREEN 或消除 I1 阻塞。
- 旧因果补验余 1 宿主+1 子进程+1 审查、旧 containment 余 1 宿主+3 子进程+
  1 审查全部 FROZEN；不借用旧额度，不自动继续旧 FINAL。
- SourceVerified/LiveActionsPerformed/RuntimeApproved/FormalAcceptance/
  RealGateImplemented 保持 false；Minor2 延后和所有原有限批准不变。
- 只观察/终止调用者创建并仍持有的确切 Process；PID 只作诊断输出，不用于重找。
  无进程枚举、按名称终止、Kill-tree、产品故障 setter/callback、探针或编译。

## 文件与接口

新 leaf：docs/diagnostics/isolated-fuzz/p2b-process-association-guard-repair/。
六个冻结输入：owned-process-observation.ps1（仅定义私有观察函数）、
containment.tests.ps1（RED/GREEN）、capture.ps1（Prepare/Run/Integrate）、
expected-output.json、README.md、.gitattributes（物理 LF）。
以因果补验的测试/capture 为适配起点，不执行其旧入口；只读复用原 helper、
fixture 及原 seam 脚本/oracle，不复制或修改产品。派生 summary/review 为输出。

新 ignored packet：.superpowers/sdd/2026-10-19-p2b-process-association-guard-repair/。
执行报告：docs/roadmap/2026-10-19-p2b-process-association-guard-repair.md；
交接更新 current-status 与 repository-recovery，显式文件列表提交现有开发分支。

Read-P2BProcessAssociation([object]Process) 返回普通有序根：
State='Associated'|'Unassociated'|'Unknown'、ProcessId=nullable Int32、
Error=ErrorRecord|null。它只读、不 Start/Kill/Dispose，不接受外部 getter 或回调。

GREEN 合同：

- 非真实 Diagnostics.Process（含 null/普通相似对象）直接 Unknown/null/null。
- 从固定 [Diagnostics.Process] 类型取 public instance Id PropertyInfo，并用
  GetValue 方法调用读取原对象，而不是对外部对象取动态属性或执行外部代码。
- 只有实际返回 Int32 且 >0 才 Associated，Error=null；null/其他类型/非正值为 Unknown。
- 捕获的方法调用错误必须保留 ErrorRecord；完整走 InnerException 到 null，最多
  8 个不重复节点。只允许 MethodInvocationException/TargetInvocationException 包装，
  最内层恰为 InvalidOperationException 才 Unassociated，ProcessId=null。
  缺失、循环、超限、其他异常或未知包装均 Unknown；不按消息字符串分类。
- 不把 Associated 等同仍存活，不把 Unassociated 扩展为任何资源清理已成功；
  此函数只在调用者自己创建、尚未 Dispose 的对象上供生命周期判断。

正关联测试以实际启动/确切退出的原对象直接读取正 Int32 Id 作独立编号锚点，
新函数的 ProcessId 必须相等；直接属性若返回 null 即未知/失败，不以“未抛异常”
自证。动态编号不写入固定 oracle，编号关系布尔值必须进入完整 trace 比较。

三处使用：第五组要求 Unassociated、null PID、有实际错误；两处清理护栏只有
Associated 才能据观察填 Started/PID。Unassociated 且没有已知启动证据不调用
Kill/Wait，不伪造 OwnedExit；与已知 Started 冲突则记 Unknown/失败。
Unknown 永不推断“未启动”或“退出”：如可能已启动，只能对持有的同一对象做
一次有界安全收束，记录每个错误，不据此自动通过；不能按 PID 重新找目标。
所有流/记录/Dispose 仍在 finally 内，外层救援即停线；主因不被清理/写盘错误替换。

## Review Focus

1. 属性异常被吞：UnstartedAssociation 对新未启动对象要求显式方法路径得 Unassociated。
2. null/相似普通对象被当成进程：UnknownAssociationRejected 要求 Unknown，不触发任何生命周期操作。
3. null PID 被当成已启动：RED 直接暴露旧前提；GREEN 三子进程与 FINAL 在 Dispose 前核正 Int32 PID。
4. 未关联对象进入清理：PreStartOutputFault 实际调用新测试侧清理护栏，要求不 Kill/Wait、不伪报 Started/OwnedExit。
5. Unknown 或清理失败被宣称成功：两护栏源码审查和真实故障/超时结果均核失败封闭；未知 API/OS 失败仅静态覆盖，不冒充已动态注入。

## 新有限预算及冻结

拟批准 NEW 3 顶层宿主 +4 子进程 =7 新 PS7 进程，条件性 1 次独立只读审查。
既有普通父上下文 Prepare/Run 记账，不另启动 PS7；无实施代理、额外基线、探针或重试。

| 槽 | 新宿主 | 新子进程 | 必须精确满足 |
| --- | --- | --- | --- |
| RED | 1 | 0 | 新未启动对象暴露旧关联判断，预定断言失败；从未调用 Start |
| GREEN | 1 | 3 | 两组关联检查+原五组收束全部 PASS，0外层救援 |
| FINAL | 1 | 1 | 新 Integrate 调用原 seam 一次，原完整 oracle/字节不变 |

每个 Start 尝试先消费独立槽（启动失败也消费）；只允许顺序前缀。RED 只能是
预定失败，随后仅 owned-process-observation.ps1 允许一次预定修复；其余五输入/
oracle 不变。RED/GREEN 各存全清单与观察源码快照。任意未知/非精确/写盘或
解析失败/漂移/退出或流不明/救援，先有限安全收束，再冻结全部余槽及审查。
不自动修新问题、改 oracle、补跑或二审。

顶层 MaxMs=120000、退出观察 2000ms、双流共用 2000ms；fault/normal
MaxMs=2000、timeout=200、helper ExitWait/Drain 各2000；fixture Hold 原45000不改。
FINAL seam MaxMs=110000，外层总120000，留10000给有限收束/记录；不能重置父截止。
超限如实 unknown，不宣称硬实时或父宿主退出等于子进程退出。

Prepare 继承因果补验 manifest 的1454条；加该阶段18 raw、6 portable、
执行报告和计划各1、新六输入、本批准计划快照1，共1487映射项。
先打印完整路径/哈希清单；重复映射必须同哈希，缺原始文件就停，不从 summary 重建。
新 packet 可预存仅 progress.md；不得存在旧 manifest/槽/结果。GREEN 仅观察函数
一项哈希变化，其余1486项逐项相同；FINAL 完全使用 GREEN 清单。前后全量核对。

## Task 1: 共享关联判据的真实 RED 与一次修复

**Files:** 新六输入；历史目录只读。
**Interfaces:** Read-P2BProcessAssociation 如上；containment.tests.ps1
-Phase RED|GREEN -TracePath <fixed packet path> -DetailsPath <fixed packet path>。
capture.ps1 -Action Prepare|Run|Integrate -Slot red|green|final -FrozenSHA256 <64hex>。
Integrate 为内部入口，核 FINAL 已消费且 GREEN 全精确；不允许任意直跑。

- [ ] Step1：先写字面 oracle 和具名断言，再完成调用方。RED 基线观察函数仅将
  旧两行 try{$id=$Process.Id;…关联为真}catch{…未关联} 原义移入该私有接口；
  不加 null/正值校验，不制造缺 API 失败。RED 只加载观察函数，不加载 helper/产品。
- [ ] Step2：静态解析、复核六输入/预算、Prepare 冻结 RED。普通父 Run -Slot red。
  Expected：恰好 FAIL: Association.UnstartedMustBeUnassociated、exit1；
  保存实际旧判定 Associated/PID=null，观察对象在 finally Dispose，0 Start/子槽。
  结果若与此不同停线，不能调整 RED 预期或直接继续 GREEN。
- [ ] Step3：仅在 RED 精确后，将共享观察函数改为上述显式反射方法/有界异常分类。
  不改两调用方、测试或 oracle；保留原 RED 源码/清单并生成 GREEN 清单。
- [ ] Step4：普通父仅一次 Run -Slot green，按以下七组执行并比较完整独立 trace：

| 固定组顺序 | 必须断言 |
| --- | --- |
| UnstartedAssociation | 新未启动真实 Process，State=Unassociated、PID=null、Error 存在；无 Start/子槽 |
| UnknownAssociationRejected | null 和普通相似根都 Unknown/null/null；无 Start/子槽 |
| PostStartReceiptFault | 同原真实目录写失败；确切退出/双流完整、要求终止、非timeout、无overflow/救援 |
| PrimarySurvivesFinalReceiptFault | 原两个独立 Error列表锚点、不同叶因；原根 true，primary替换/交换/cleanup替换/null/缺失/九层负例 false；恰一清理错误 |
| NormalExit | 原 EmitAndExit；固定双流、exit0、无错误/kill/timeout/overflow；Dispose前关联为正 Int32 PID |
| Timeout | 原 Hold/200ms；请求终止且确切退出/双流完整、已知退出码，无额外错误/overflow；Dispose前关联仍为原 PID |
| PreStartOutputFault | 原输出路径故障；Started=false、primary存在、Unassociated/null PID、OwnedExit=false、ExitCode=null、无Kill；实际走新测试清理护栏，记录未尝试Kill/Wait |

  Expected：7行 PASS 同序、宿主exit0/空stderr、完整退出与双流、0救援、后置pins完整，
  完整trace对oracle零差异。PID/耗时/终止码仅在预先列出的动态 receipts 保存与范围检查，
  不临时删字段放宽比较；details 保存有界异常类型链和实际观察，不恢复 CLR 身份。
- [ ] Step5：保存必要开发检查点，标明仅此任务观察完成；失败只保存 unfinished。
  PowerShell/manual账本替代Bash技能包装，不额外运行 task-done/全套历史测试或删除 raw。

## Task 2: 一次 FINAL、条件审查和交接

**Files:** 同一新 capture/oracle；输出 summary、执行报告、review/disposition。
**Consumes:** Task1完整 GREEN、三已退出子槽及未漂移 GREEN清单；原 seam/oracle。

- [ ] Step1：GREEN完全精确后，普通父预消费 FINAL 并启动 Integrate宿主一次；
  它预消费唯一 seam子槽，用固定 owned-run 调原 seam。新 capture侧护栏调用同一
  Read-P2BProcessAssociation；与测试侧同合同。普通顶层父仍使用独立 Started标记监督。
- [ ] Step2：要求原15行/12PASS、exit0/空stderr、完整原trace oracle零差异，
  stdout SHA F693A64A1402A0AF882E30D62F6EB79601F5F380974900DD982B7ED94641A0FE、
  trace SHA 4751D03D28FB8342AFFC56BF0E8343F0E0B1D5E548A7DF7C1B164F719C10A17E。
  helper无primary/cleanup/overflow/timeout/kill、确切退出/双流完整、无救援；
  FINAL宿主同样完整、无父错误、后置pins完整；关联 PID 与持有对象一致。
- [ ] Step3：只有全部上述条件成立，消费唯一 fresh只读review，核六输入/RED→GREEN
  唯一改动、两护栏及第五组、所有Review Focus、七组/FINAL/预算/证据。无执行探针、
  嵌套代理或修改；完整报告逐项 disposition。阻塞/不完整即停，无自动fix/补验/二审。
- [ ] Step4：保存源码/计划身份、全部原始流字节/trace/details/账本哈希与实际预算，
  summary与raw逐项一致；raw继续ignored且保留。更新状态/恢复，显式列表提交推送，
  核remote==local。不自行关闭正式验收，不改旧失败/冻结余量，不合并main或授予live权限。

## 自检与批准门

已自检：三处错误前提在同一私有边界修订；真实未启动 RED 先行、只改一输入、
七组 GREEN/一次 FINAL/条件审查的接口与预算一致。对 null/未知拒绝、原首因负例、
清理写盘失败封闭、完整 oracle 和历史保护均有任务；OS/API未知异常实际注入仍未覆盖，
只能静态审查并如实 unknown。反射包装预期未动态验证，非精确就停，不边跑边改分类。

原预算已批准并全部消费，不解冻任何旧槽。R1触发停线，下一步需另行批准窄范围
首因/最终记账修复计划及新有限预算；本轮仅补齐未验收交接，不自动实施建议。
