# Offline Diagnostic Review Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans inline, task-by-task. Preserve Native execution. No implementation agents/per-task reviews; only the conditional single fresh readonly reviewer below. Track execution in this plan's own ledger, never edit approved plan bytes.

**Goal:** 修复独立审查 I1 首因被清理错误覆盖、I2 非Step观察/错误判定不完整；只提出新修订离线入口的窄验收，不修改旧验收结论。

**Architecture:** 保留da81a0fd全部冻结源码和证据，新建版本化入口/私有目录及独立测试/packet。两个小型生产边界从现有实现机械迁移，先复现缺陷，再分别修复；所有测试/字面oracle先冻结。最终三个实际profile必须经过同一个完整观察判定器。

**Tech Stack:** Windows PowerShell7.6.5 Core x64、.NET Stream/Process、CreateNew/原始字节双流。两次宿主内Add-Type内存夹具编译，不安装/下载、不生成磁盘程序集或另起编译目标。

**Spec:** 沿用[已批准设计](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md)，SHA DA27DA1C6BD3578512FB60A7F38B6635EAC9ACEF6570534F8B868BD07D088D9A；[原合同§2](2026-10-19-offline-diagnostic-entry-migration.md)及[迁移计划](2026-10-31-offline-diagnostic-runtime-migration.md)的接口、真实测试组和有限生命周期界限继续规范性适用，以下新路径/清单/预算/定向负例替换原运行安排。[完整独立审查](../../roadmap/2026-10-31-offline-diagnostic-runtime-review.md)是本次I1/I2的需求来源，SHA D6F3DB4A607BE202C88C6D7664050DEB2346C6FB435EBCF8ED41D4D813A6A2BD。

状态：PLAN_PENDING_USER_REVIEW_AND_NEW_EXECUTION_APPROVAL。基线da81a0fd5fa531a23156849af45fcabdb01876fb。用户本次仅批准制定修复计划/新额度建议；本轮0运行源码、测试、Prepare、模块导入、编译、目标、代理或审查。旧6宿主/7目标/1审查全部耗尽。

## Global Constraints

- “只涉及离线诊断，不做生产接入”；不改产品、现有验证链、CI、Docker，不合并main/发布。
- “默认为 Inspect”；无模块导入、Session、运行目录或目标创建；资料/参数不是运行许可。
- “Session/Ack 不序列化，不从文件加载，不跨 shell/进程转移”；真实引用/步骤/预算与失败不可逆锁存不变。
- “首次实际失败后，后续逻辑 Read/Parse/Save/Complete/Publish/重新打开在文件访问前拒绝”；清理不重试失败出口，首次真实原因先于次因。
- “不加入 PID/名称重查、杀树、Job/watchdog、无限等待、启动预热或额外 readiness 等待”。
- 旧源码/原件/portable/manifest/审查全部原字节不变；不编辑旧oracle来适配本次结果，不重跑旧Grant，不以新预算返还旧额度。
- 仅手动有限PowerShell/既有linked worktree。批准本计划也须明确批准：无技能默认Bash/setup/baseline/full-suite/task-done重跑/实现代理/额外审查/自动fix/删除。
- 现有I1/I2、历史C1/C2保持open，R2/Minor2延期，五flagsfalse，直到分别取得限定证据；任何结论不回填旧报告。

## Review Focus

1. 写入/flush与Dispose同时失败：两个实际生产写边界均须传播同一个真实首因引用，次因有界且只在首因之后；Task1两边界各5向量，不能把已写PASS当记录成功。
2. 判定器自身发生缺字段/未知类型/错误位置异常：不得把任意异常当“成功拒绝”；Task2固定拒绝签名、未知/缺失/顺序和类型变异覆盖。
3. 非Step的View/IORoot/Error和summary不能略过：Task2完整逐项/递归合同，Task3三个真实profile完整消费，不以子集PASS代替。
4. 新路径/正常执行不能偷偷启用测试工厂或历史上下文：Task1隔离测试工厂，Task3只用真实CreateNew/新固定根及源身份；最终审查核所有入口调用绑定。
5. 记录失败、工具提前返回、旧portable缺失不授予下一轮许可：完整结果先留存再解析、阶段清单/原始双流/唯一final；任何真实收集/写入不明即停，不靠补跑或新目录补证。

## 1. 选择、范围和新归属

本次是已批准架构下的修订，不重选框架/平台。推荐独立v2归属：原位修改会破坏冻结身份；仅降级文档不能修复I1/I2。因此本计划提出新路径，尚未声称用户已批准这些路径/新预算。代价是文档必须清楚区分新旧入口；旧入口不重定向、不删除，也不冒称被OS禁用。

- 新公开入口：`scripts/offline-diagnostics-v2/diagnose-offline.ps1`，repo根取上溯../..。
- 新私有目录：同目录`private/`，module上溯../../..。
- 新leaf：`docs/diagnostics/isolated-fuzz/p2b-review-repair/`，repo上溯../../../..。
- 新packet：`.superpowers/sdd/2026-10-31-offline-diagnostic-review-repair/`。
- 执行报告：`docs/roadmap/2026-10-31-offline-diagnostic-review-repair.md`。
- 当前受支持操作/恢复说明只在最终交接更新；旧迁移执行报告、完整审查、旧操作者说明作为本次冻结依赖，不原位改写。

公开Action/Profile、3profile ID、固定目标/预算、唯一导出Invoke-ODProfile、私有LS签名及原有有序报告根均不变。只有新路径/packet/manifest绑定改变。RunApproved只硬编码三个profile；同宿主再次调用在任何IO前拒绝，无Force/Remove-Module。runtime只读本packet的`repair-green.frozen.json`，单位验证不导入运行module、不调用RunApproved。

### 1.1 首冻前25源码，两个精确变更窗口

| 归属 | 明确文件 | 职责 |
| --- | --- | --- |
| 入口树10项 | .gitattributes、diagnose-offline.ps1、private/catalog.json、private/session.psm1、sinks.ps1、owned-step.ps1、sequence.ps1、profiles.ps1、observations.ps1、write-boundary.ps1 | 局部LF、固定公开入口/描述、私有引用/IO/确切对象/完成/固定分发/实际错误分类、两个记录边界 |
| leaf15项 | .gitattributes、README.md、expected-output.json、test-helpers.ps1、entry.tests.ps1、core.tests.ps1、prefix.tests.ps1、owned.tests.ps1、live.tests.ps1、capture.ps1、driver.ps1、record-boundary.tests.ps1、observation-contract.tests.ps1、fault-stream.cs、observation-contract.ps1 | LF/边界、全部字面合同、原测试迁移、单宿主捕获/原始父传输、I1回归/I2变异/内存夹具/实际完整判定器 |

首冻前25项全部Present，测试和oracle全部写齐。只做路径/固定配置适配，机械分离两个写边界与观察判定器，不修复其已知缺陷；两个RED必须揭示原缺陷而不是missing-function/加载/编译错误。代码移动的唯一目的为测试实际生产调用点，不能造一个未被入口使用的“演示实现”。首冻前静态核对与旧算法控制流等价。

允许修改窗口仅：writer-green恰`private/write-boundary.ps1`一项；repair-green恰`leaf/observation-contract.ps1`一项。其余24项及所有历史输入不变，特别是测试/oracle/capture/profiles/public/driver不可在看见结果后改动。所有生产调用从首冻即绑定这两文件，修复不能绕过它们。

### 1.2 清单/历史资料

继承旧runtime-green.frozen.json全部2129Present Pins，SHA C3E24036ED56FFC89C8742418342D02A9B447B8C8643B0F8F1860D0DCAEF3175；保留VerificationPath/SHA256/ExpectedState，Role可选属性安全查询，重复路径只允许一致state/hash。不把旧8路径再标Absent，不覆盖旧清单或复制运行额度。

再加188项：旧packet真实142文件＋旧leaf15文件＋旧执行报告/完整审查/旧操作者说明3＋本计划原件1＋新packet批准设计/本计划快照2＋25新源码。2129+188=2317，全部Present；三次冻结seed/writer-green/repair-green各25快照，共75。以明确文件集合而不是leaf文件总数静态检查，后加portable不进入运行输入。

旧portable-final SHA 01889C7A479EFD418D19BF2D5B2E8B279077139006E1FB5E923706C4A99A300F，只读核142原件的路径/长度/字节/hash和49目录，不从其重建原件。旧执行报告SHA F68CF2B1B093B412F14E2FC56007A437EC9A754E8425642F603B526390B2C798，旧操作者说明SHA1526710D06AA99C76A5595C5074AA7CBD917C36EE5D2A6DC0EB9ADB1D13B9B72。完整审查SHA见头部。

第一次Prepare仅新packet自己的progress.md，其他存在内容拒绝；不换目录。后两次Prepare验证前阶段SHA，允许指定新源码Present→Present hash变化，恰1项，其余2316项原字节不变；旧源码永远不能进入变更窗口。全量前后验证/快照/manifest单次CreateNew，实际SHA即时记ledger。

## 2. I1：实际生产记录边界与可控夹具

新`write-boundary.ps1`私有定义：

- `New-ODRecordStream([string]Path) -> Stream`：仅File.Open(CreateNew,Write,None)，无外部工厂参数；正常capture/profile永远使用此默认定义。
- `Write-ODCaptureRecord([string]Path,[byte[]]Bytes)`：capture序列化/1MB限制之后实际调用；迁入旧Capture-Once的写/flush/dispose控制流。
- `Write-ODProfileIntentRecord([string]Path,[byte[]]Bytes)`：profile锁存/固定根/intent序列化后实际调用；迁入旧profile.intent写/flush/dispose控制流。

二者均不导出到公开入口。正常caller所有Bytes/Path固定内部生成，不新增任意callback/path/stream用户参数。受测副作用以实际资源操作发生，不构造ErrorRecord或模拟Process/Session/Ack。

unit capture从新冻结文件建立私有隔离测试作用域，使用同样Capture-Failure首因/8次因引用规则。只在该作用域替换New-ODRecordStream，交给固定故障Stream；不改变捕获宿主最终证据自身工厂、不改运行module或全局工厂。用后即释放该作用域，不依赖Force卸载/重导入。

测试工厂调用冻结默认工厂的原始ScriptBlock打开固定case文件，健康分支只增加计数包装并最终核文件字节；OpenFailure直接调用默认工厂，已存在文件不得改变。夹具计数是测试流操作计数，不冒充Process生命周期。捕获每个case最终实际抛出ErrorRecord并按正常Capture-Failure去重；从该隔离作用域真实收集的错误中按独立字面Write/Flush/Dispose签名识别首因对象，再比较Thrown/HostPrimary/Secondary引用，不在夹具中制造“预期ErrorRecord”。缺失原始错误对象本身就是非精确失败。

`fault-stream.cs`经Add-Type -TypeDefinition在writer-red/writer-green各编译一次到内存。固定模式WriteDispose/FlushDispose/DisposeOnly/Success：Write/Flush代理到已关闭的真实FileStream以得到真实ObjectDisposedException；Dispose代理实际读取固定不存在文件以得到真实FileNotFoundException；Success代理健康流。夹具不new/throw预制异常，不伪造ErrorRecord，记录Write/Flush/Dispose调用数。其双故障是**受控Stream测试夹具**，不是OS磁盘晚写/真实FileStream.Dispose失败的动态证明。默认生产工厂另外用真实CreateNew正常/已存在文件检查。

具体分派固定：WriteDispose仅Write代理closed流；FlushDispose的Write代理健康流、仅Flush代理closed流；DisposeOnly的Write/Flush均健康。三种失败模式Dispose均先释放已持有的健康底层流，再执行一次缺失文件ReadAllText产生真实次因/唯一原因；Success只释放健康流。故障不能靠遗漏资源释放产生，不另外Dispose失败的同一包装对象，也不使用循环/额外进程。所有夹具输入路径固定在本Grant的unit/<Boundary>/<Case>/目录内。

两边界各5例，固定10组：Capture/Intent分别加WriteDisposeFirstCause、FlushDisposeFirstCause、DisposeOnly、SuccessOnce、OpenFailureOnce。断言：

- Write+Dispose：首因/实际抛出引用均为Write的原始ErrorRecord，secondary恰Dispose原始引用1项，Write1/Flush0/Dispose1；无重试/替代出口。
- Flush+Dispose：同上首因Flush，Write1/Flush1/Dispose1，secondary1。
- DisposeOnly：唯一首因Dispose，secondary0；SuccessOnce：无首因/次因，写入原字节，三操作各1。
- OpenFailureOnce：真实CreateNew已存在文件导致IOException，Open1/Write0/Flush0/Dispose0，原文件未改。

Write/Flush错误Type=System.Management.Automation.MethodInvocationException、ErrorId=ObjectDisposedException；Dispose错误同Type、ErrorId=FileNotFoundException。消息分别绑定实际Write(3)/Flush(0)/Dispose(0)方法和固定closed-file/缺失文件路径文字；Open失败绑定Open(4)/IOException/固定已存在路径。首冻前把这些确定消息写为字面模板，不在RED/GREEN后回填。位置必须来自本轮冻结write-boundary及有效行号；所有比较以实际ErrorRecord引用/投影为依据。错误摘要本身失败也必须停线，不能因summary健康推导操作成功。

上述消息内文固定为closed FileStream的`Cannot access a closed file.`、Dispose缺失文件的`Could not find file '{path}'.`、Open已存在文件的`The file '{path}' already exists.`；外层均为`Exception calling "{method}" with "{argc}" argument(s): "{inner}"`。不允许接受任意ObjectDisposed/IOException文字。RED四组失败的事实也必须预定：Capture两例Thrown仍为首因但HostPrimary不是首因；Intent两例HostPrimary仍为首因但Thrown不是首因。其余原因丢失/错误类型/计数偏离都不是该精确RED。

修复策略：操作期间只保存第一次真实ErrorRecord；finally只释放持有对象并暂存Dispose错误，不在已有首因尚未登记时发布次因，也不在finally抛替代异常。资源收束后按first→secondary顺序有界登记，随后只抛原始首因；既有HostPrimary不得被覆盖，重复同引用不增次因。无成功finally、无重开/重写/第二次Dispose。

## 3. I2：独立完整观察合同与拒绝变异

新`observation-contract.ps1`定义`Assert-ODProfileObservations([object[]]Rows,[string]Profile,[Collections.IDictionary]Context)`：成功无输出，合同不匹配仅抛RuntimeException，Message/FQID均`OD: observation contract mismatch`。driver的每个真实ProfileReturn必须调用它，不保留只校验Step的替代路径。Context由driver固定构造，含repo/packet/manifest/source bindings；它是判定数据，不授予运行权，不暴露给公开入口。

冻结expected-output.json须覆盖**完整**观察数组/顺序/根、每个View的全部Sink键与Path/State/IO计数、IORoot的Ok/Value/Error、首因/有界次因和summary递归所有确定值；布尔/数值/字符串/null类型均约束。禁止仅检查Contains/子集或忽略未知字段/重复/缺记录。只允许固定repo/packet路径模板、实际源码有效行号/栈及原合同PID/ElapsedMs的显式动态规则；不通配任意错误文本或将非null错误视为匹配。

确定观察inventory沿用实际测试程序的行为，首冻前独立列定：

- core恰22项：disk,Write,Read,Complete,Write,sink,Read,missing,Read,bad,Complete,index,Write,errors,Output,output,Step,Step,preconsume,Step,Step,Step。
- owned恰8项：Step,prestart,Step,receipt-fault,Step,normal,Step,timeout。
- live恰15项：14个Step后恰1个sequence；sequence内部A/B/C/Fresh及A0/B0/C0/A1完整匹配，不只重用Groups摘要。

原profile Case/Steps/Budget和全部group/facts固定不变。core错误合同至少包括以下完整模板；`{path}`只可展开为该固定case受管路径：

| 情形 | Type / ErrorId | Message |
| --- | --- | --- |
| sink/errors初次Write | MethodInvocationException / UnauthorizedAccessException | Exception calling "WriteAllText" with "3" argument(s): "Access to the path '{path}' is denied." |
| missing Read | MethodInvocationException / FileNotFoundException | Exception calling "ReadAllText" with "1" argument(s): "Could not find file '{path}'." |
| bad Read | System.ArgumentException / System.ArgumentException,Microsoft.PowerShell.Commands.ConvertFromJsonCommand | Conversion from JSON failed with error: Unexpected end when reading JSON. Path '', line 1, position 1. |
| index Complete | System.Management.Automation.RuntimeException / Invalid ledger row index | Invalid ledger row index |
| output Write | MethodInvocationException / ObjectDisposedException | Exception calling "Write" with "1" argument(s): "Cannot write to a closed TextWriter.{NL}Object name: 'StreamWriter'." |

表内MethodInvocationException全名System.Management.Automation.MethodInvocationException。Failed后拒绝的Read/Complete/Write必须Ok=false/Value=null/Error=null且LogicalIO不增加。errors summary保留真实首因、8个独立FileNotFound次因及Overflow=true；每个次因绑定其absent-0..8固定路径/实际ReadAllText错误，前8保留、第9仅Overflow。四种Step期待错误仍按原合同完整检查，不以本表替代。core测试本身也在新副本首次冻结前加入相同确定错误断言，不能只靠父判定器弥补弱断言。

unit使用旧runtime-green/lifecycle/continuation的**实际保存**ProfileReturn作为3个正例；先核旧manifest/证据hash及源绑定，保留原字节/数字，仅内存克隆用于改坏字段。固定unit Context指向该只读历史namespace/source集合；正式driver Context只能指向新namespace。仅载入新判定器及其冻结纯schema助手，不运行旧driver/capture/module或历史进程、也不恢复旧Session/Ack。旧输出是待判数据，不是oracle生成器；独立字面合同须先从测试行为/固定异常规则逐字段写定，禁止导出实测值成为期望。

固定20个单变异，expected均为上述**精确拒绝签名**，其他异常/空结果不算拒绝成功：

1–4：删除core首Read、重复该Read、交换该Read与Complete、Read改名Unknown。
5–9：选择missing case的Read，Result增加键、移除Error、Error.Type改System.Exception、ErrorId改UnrelatedError、Message改unrelated（不对Error=null的sink拒绝记录改子属性）。
10–13：missing的View.Status改Active、Result.Ok改true、失败Value改空对象、Write.Result移除Value键。
14–16：output的ConsoleOut.State改Healthy、errors summary删除第8个保留次因、sequence.A.AckConsumed改0。
17–20（原Step拒绝阳性控制）：owned/normal的Step.Observation.OwnedExit改false、同Step.PID改0、core/preconsume的Step.Result.Error.ErrorId改UnrelatedError、同Step.Result.Accepted改字符串true。

3正例+20变异=23组。seed判定器机械保留当前漏洞，预期正例3和控制4通过、前16拒绝测试失败；修复后23全部通过。检查函数的崩溃不能冒充拒绝；新增未知动态项不得静默加入白名单。正式三个profile原始观察须逐项经过同函数并全量前后pin，不能从unit通过推导运行证据已完整。

四个unit宿主同样使用原Schema/Grant/Status/Groups/CurrentGroup/NotRun/Observations/PrimaryError/SecondaryErrors/Overflow/Probe/五falseflags有序捕获根；所有组实际完成，Status=COMPLETE、NotRun=[]、CurrentGroup/PrimaryError/Probe=null、SecondaryErrors=[]、Overflow=false，RED也不例外。I1有恰10个RecordCase观察，固定Boundary/Case/Thrown/Primary/Secondary/调用计数/文件未改或字节匹配/工厂隔离事实，错误投影全字段校验；I2恰23个SchemaCase观察，固定向量编号/Profile/ActualRejected/实际合同拒绝Error，全部按首冻字面模板验证。操作异常不能因group结果符合RED而略过。unit正常证据写入与被测故障Stream必须隔离，完整原始stdout/stderr/exit/pins/证据/final均精确才放行。

## 4. 新额度申请、期限和停线

| 顺序 / NEW Grant | capture宿主 | 真实目标 | 固定完整结果 |
| --- | ---: | ---: | --- |
| writer-red | 1 | 0 | I1恰4FAIL/6PASS，分别是两边界WriteDispose/FlushDispose首因失败；exit1，记录健康，无额外异常 |
| contract-red | 1 | 0 | I2恰16FAIL/7PASS，前16拒绝未生效；exit1，记录健康，无额外异常 |
| writer-green | 1 | 0 | I1恰10PASS/exit0/空stderr |
| contract-green | 1 | 0 | I2恰23PASS/exit0/空stderr |
| core | 1 | 0 | 新入口6组＋prefix/core/绑定/重入16组，共22PASS；22个完整private观察全部核验 |
| lifecycle | 1 | 3 | 新入口6＋owned6＋绑定/重入2=14PASS；8个private观察完整核验 |
| continuation | 1 | 4 | 新入口6＋live5＋绑定/重入2=13PASS，另原12seamPASS；15个private观察完整核验 |

申请NEW7capture宿主＋7受控PS7目标＋两次宿主内内存夹具编译，全部精确才允许NEW1fresh只读审查。前4unit不创建Process/Session/Ack，不调用RunApproved；仅后3经实际新公开入口运行，当前profile根开始不存在，后续Inspect只观察之前完成profile根且字节不变。不给空壳包装器/探针/旧Grant重跑额度。

Native父工具内联driver，仅driver持有一个实际capture Process，capture内联入口。每个非continuation宿主120000ms，continuation150000ms；内部全序列140000ms/清理预留4000ms，确切退出2000ms，双流共享排空2000ms。Hold45000ms、normal/receipt2000ms、timeout200ms、seam110000ms不变。normal各流19ASCII字节、timeout仅字面前缀含真实byte[0]。Session≤32/sink≤32/Ack≤2/观察≤32/不同次因≤8，第9项Overflow不变。

固定engine路径/版本/身份与旧计划一致：C:/Users/Lenovo/.cache/codex-runtimes/codex-primary-runtime/dependencies/native/powershell/pwsh.exe，SHA362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；SMA同目录SHA2D8AF026E589C5603629BFC3E7C57265B4167B12ECA3AE549C6A6C3CA7E4967B；DESKTOP-H4SVCLU\Lenovo/adminFalse。fixture/observer/originalseam路径及hash沿用原合同，不发现即停、不安装或放宽。

三次Prepare：seed一次；writer-green以seed实际SHA；repair-green以writer-green实际SHA。前2Grant用seed，writer-green用同名freeze，余4Grant用repair-green。所有参数/实际SHA/完整工具返回即时记ledger，先保存原对象及实际session_id/片段/退出，再解析；不把10秒yield当超时或丢空输出重启。Tool存档用PowerShell UTF8/字面保留输出字符串，不依赖JS btoa，不做JS Number往返处理原始JSON。

仅表内两种完整精确RED是预期负例。编译/实际断言/输入/Prepare/来源/预算/进程退出/排空/原始收集/summary/final/evidence失败或unknown立即冻结所有剩余Grant/review；不自动修复/重跑/重冻/返还/新namespace/二审/重建原件。纯只读显示语法错误或首冻前确认整批拒绝且未落盘的补丁可记录后纠正；不得将真实采集/写入失败改称显示错误。本计划批准后才生效，不追溯更改任何旧停线结论。

## Task1：机械迁移、预冻结、I1 RED→GREEN

**Files:** 首冻前建立明确25源码；RED后只改write-boundary.ps1。
**Interfaces:** 消费冻结旧边界/完整审查；产出实际两生产写边界及Task2/3始终调用的默认工厂，测试替换不泄漏正常调用。

- [ ] Step1：确认同linked dev worktree/clean/非submodule；本计划ledger记录用户批准/新额度。只读核历史2317来源、原件/portable；25源码机械适配，写全I1/I2/实际profile测试与独立oracle。预检调用链、类型/参数、源LF/AST/JSON，核seed保留缺陷而不是修复后伪RED。
- [ ] Step2：`& <new-leaf>/driver.ps1 -Action Prepare -Grant seed`一次。Expected2317Present/25snapshots/无Absent/无目标。随后RunOne writer-red/contract-red各一次，均显式传实际seedSHA。Expected依表精确签名，否则停。
- [ ] Step3：修复write-boundary.ps1首因/次因顺序和原始引用传播。其他源码/工厂默认/测试/oracle不变。
- [ ] Step4：Prepare writer-green一次，PreviousFrozenSHA256=实际seedSHA，Expected恰1新源码hash变化/25snapshots/其余2316不变。RunOne writer-green一次10PASS/0目标/空stderr/完整记录；不另起编译或helper自检宿主。
- [ ] Step5：必要源码/便携记录/执行current-status/recovery显式提交推送核相等。Spent3hosts/0children/2内存编译，剩4hosts/7children/条件1review；不声称完整观察合同修复。

## Task2：I2完整判定器修复与23组GREEN

**Files:** 仅改新leaf/observation-contract.ps1；测试/oracle/driver及其他24源码不变。
**Interfaces:** 消费首冻23组schema向量和完整字面合同；产出正式driver唯一调用的全量观察断言。

- [ ] Step1：读本任务固定inventory/异常模板/20变异，确认Task1 contract-red恰16FAIL/7PASS且没有任意异常冒充拒绝；不重跑RED。
- [ ] Step2：实现全部22/8/15观察的有序完整匹配，递归View/IORoot/summary/Error，拒绝未知/缺失/重复/错序/错类型/错误签名；动态字段限固定规则，匹配失败只发固定合同拒绝错误，不进行任何运行IO。
- [ ] Step3：Prepare repair-green一次，PreviousFrozenSHA256=实际writer-greenSHA，Expected恰该1新源码变化/25snapshots/其余2316不变。RunOne contract-green一次，Expected23PASS/0目标/空stderr/全量pins与证据健康；此后全部源码不可改。
- [ ] Step4：提交推送/核远端；Spent4hosts/0children/2编译，剩3hosts/7children/条件1review。schema通过不是实际新入口全部通过。

## Task3：新实际入口、唯一审查、可恢复交接

**Files:** 不改运行源码/测试；仅产生固定记录、审查/处置和执行/操作/恢复文档、最终portable。
**Interfaces:** 消费repair-green实际SHA及固定source绑定；产出实际profile完整证据及限定评估。

- [ ] Step1：RunOne core一次；Expected22PASS/0目标/22private观察，真实默认/显式Inspect/14拒绝、模块来源/重入、prefix98条件/core原组及新增确定错误断言全部通过。profile调用位未被Inspect占用。
- [ ] Step2：RunOne lifecycle一次14PASS/3目标；预启动失败FIRST/Start0/Unassociated/noKillWait，receipt失败一次/首因保持/有限退出排空，normal及timeout原元组；父判定器完整8观察。
- [ ] Step3：RunOne continuation一次13PASS另12seamPASS/4目标；Active/Next1/Reserved1/预算2余1的独立缺Ack负例、6无效请求/重放、C真实输出失败及B预算、A1原seam真实续接一次，父判定器完整15观察。保留原seam stdout/trace hashes，不另起FINAL。
- [ ] Step4：仅7hosts/7targets/2编译/2317pins/75snapshots及所有完整字面证据精确，保存完整源码检查点后开始唯一fresh readonly review。显式核当时最强可用model/独立上下文，只读全新调用链、spec/本计划/旧I1-I2、所有rulings/5focus、原始记录；不运行/导入/网络/写文件/Git或派代理。重要项/不完整审查即停，保存未验收检查点；无自动fix/额外测试/二审。
- [ ] Step5：完整保存审查及每项Declined-to-judge处置。即使0Important也只讨论新v2对应窄边界；复合Stream夹具≠真实磁盘晚写/close故障、动态classifier异常/宿主死亡/同步IO无限阻塞等未证实限制仍保留，不回填旧I1/I2或C1/C2报告。
- [ ] Step6：逐字节导出本packet所有真实原件/工具文本/批准/manifests/75snapshots/失败/unknown/审查和目录；byte[0]直接保留，不从portable“重建原件”。报告明确哪些通过、谁做的核对、未验收项/耗尽预算/下一步。更新current-status/repository-recovery及**新增**v2操作者说明并从恢复页链接，不改已冻结旧runbook。显式清单stage，原字节与暂存blob一致、提交推送现有dev分支/核远端SHA等于HEAD；不合main/删除任何历史或raw。作者导出核对不冒称独立审查。

## 5. 作者自检与批准门

已对照：I1两实际边界/两复合向量→Task1，I2所有非Step递归字段/异常签名/拒绝不能崩溃冒充→Task1 RED/Task2 GREEN，默认只读/固定参数/源绑定/不混淆Ack/实际退出排空→Task3；5Review Focus各有对应测试/最终限制。2317=2129+142+15+3+1+2+25；25×3=75；4unit＋3profile=7hosts，0+0+0+0+0+3+4=7children，2内存编译，条件1review。

建议Native继续。尚未批准本书面修复方案、新v2路径、内部测试工厂替换/内存编译或NEW7/7/条件1额度；本轮只读调查/文档，0代码/测试/Prepare/编译/代理。请审阅并明确批准后再执行，不把本次“制定计划”的同意解释成运行许可。
