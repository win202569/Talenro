# 隔离框架诊断执行记录

计划：2026-10-14-isolated-fuzz-coordinator-diagnosis；基线96a70850。
用户选择1，批准 Native 执行。任务2–5尚未完成，最终独立审查未进行。

## Task1：安全复制逻辑

RED：安全stub拒绝合法独立目标，测试exit1，明确断言失败，而不是解析错误。
GREEN：普通沙箱测试13项通过，真实junction/hardlink创建受限、标为UNRUN。
随后在批准目录进行提权测试：16项全部通过、exit0，包含真实junction、
目标祖先junction和多个hardlink拒绝；源/副本清单一致、修改副本不改源。
仅微型fixture进行了复制；此记录时未复制/构建/插桩固定工具链。
测试目录均保留在专属scope，不提交临时数据，也未删除任何用户文件。

## 执行裁决与依赖

- Bash助手改用PowerShell等价台账/brief记录：机器仅支持PowerShell；
  保留计划身份和BASE，代价是减少自动化记账。
- Gate单元测试使用单独验证副本，六个框架臂稍后使用新的run副本：
  解决“Task2先测试钩子”与“baseline必须先于patch”的依赖；代价是额外磁盘/复制时间。
- 不运行通用依赖下载或产品完整套件：批准spec明确禁止；代价是没有新的产品验收证据。

下一步是Task2专属副本/真实上下文钩子TDD。原matrix3失败、Pair未运行及
全部生产验收门槛继续开放。当前仅诊断工具进度，不是修复或完整框架复现。

## Task2：上下文门已验证，框架追踪未完成

首次copy校验反复检查每个文件的全套祖先路径，持续数分钟仍未完成。
只读抽样7文件祖先检查耗时103ms；核实所属进程后仅停止本任务慢速校验，
不删除文件。改为根祖先一次检查、逐层每组件检查、复制时每不同源父目录
检查一次；每文件reparse/hardlink检查及三份hash清单仍保留。
这是GREEN后的去重重构，随后16项隔离测试再次全部通过。

新的gate-tests-002副本复制并核对11536文件，与源清单一致，原目录前后不变。
实际go version为go1.26.5 windows/amd64，GOROOT明确指向独立副本。
只在此副本src/context增加诊断代码，原模块缓存不修改。

RED：无行为stub构建成功；第一次未加引号的-test.run参数被PowerShell误解析，
该参数错误不算RED。改成整参数字符串后TestDiagnosticFuzzSelectedParent真实
exit1，提示selected parent window not established。
GREEN：实现原上下文定向传播门后，5项真实context测试全部通过、exit0：
SelectedParent、Unregistered、Watchdog、EarlyUnregisterAndCancel、RejectRegistration。
看门狗测试故意验证超时释放（1.01s）；这不等于实际框架臂允许看门狗触发。

测试使用外部context_test包，避免testing对context的依赖形成内部测试导入环；
代价是只验证导出通道/行为而非未导出字段。该裁决已记入本计划台账。
coordinator.patch目前只包含context部分；internal/fuzz插桩、追踪测试、runner、
六个框架臂及最终独立审查尚未实施，Task2整体未完成，Task3–5待办。
必要模板/局部patch与记录保存为明确未完成检查点，不能当作完整复现或修复。

## Task2追踪补齐检查点

新增fuzz-diag.go.txt/fuzz-diag_test.go.txt组织追踪格式和测试；新增materialize.tests.ps1
验证复建器。这些必要模板仍仅进入复制的标准库，不进入产品包，代价是多3份素材。
追踪格式RED：哨兵身份缺失、未读取值错误标为已读取，两项测试真实exit1；
GREEN两项通过。错误ID按相等性登记，不读Error正文、不用errors.Is吞错；
不可比较的错误明确标unsupported，其实际框架集成尚未验证。

实际stop仍保留原错误相等、interrupt短路及旧错误替换条件；子Err和条件性
父Err各在原求值位置捕获，不额外读取父状态。先释放传播门再写日志/取消workers。
截止、worker、种子crasher及最终返回有独立来源记录；未知路径标other，不假定来源。
观察模式立即释放门；其他模式只操控已登记父context。最终收尾记录worker计数。

复建器RED缺少应安装的context helper；GREEN8项断言通过，含双源码hash钉定、
全守卫先于修改、原工具链拒绝、重复安装拒绝、实际patch反向核对。
最新范围复核：5 context +2 trace Go测试通过；16 isolation +8 materialize
PowerShell断言通过。原context/fuzz源码hash仍765F99F7/D9DE1256（全值见spec）。
这些是组件验证；runner、六个真实框架臂及最终独立审查尚未执行。

## Task3运行器检查点及基线构建阻塞

已新增run.ps1/run.tests.ps1、process-native.cs及独立fixture模板。Native helper
先挂起启动，再绑定独立非breakaway job并恢复，只有stdout/stderr/NUL标准句柄
列入继承白名单；超限只处理该job，不按猜测PID清理。代价是Windows专用素材。
原退出码丢失stub RED后修复；wrong-mode和缺worker证据对照也分别RED再修复。
当前16项运行器断言通过，包含真实子进程清理、重复收据/臂拒绝和子环境隔离。

模型复核首次在Go启动之前因PSObject环境值cast失败，失败收据保留；
显式包装值回归重现RED，解包字符串修复后16项通过。新model-check-002
只运行controls-only，11项通过，没有随机采样或产品验收重跑。
返回日志不应伪称读取child/old/stopping，新回归RED后修复为unread及
stoppingRead=false；context5+trace3共8项Go测试。退出源parent-done
准确指ctx.Done分支，incoming字段区分deadline/canceled（不新增Err读）。

新framework-001副本11536文件校验成功；准备UTC11:54:25–12:05:26
（Asia/Tbilisi15:54:25–16:05:26）。副本仍未插桩，仅开始baseline预备构建。
baseline构建UTC12:30:49–12:32:23（Asia/Tbilisi16:30:49–16:32:23）：
exit0、TimedOut=false、job累计278进程、根退出时记录剩余1；
已执行该job清理，CleanupVerified=true、ActiveAfterCleanup=0。
因此运行器构建门槛失败并按设计停止；不把exit0当完整基线通过。
剩余1是否实际子进程、根job记录滞后或其他记账问题尚未确定；历史收据
未含该时刻成员PID列表，不能事后按旧PID排查/终止或改写证据。

实际baseline/execute目录不存在，只有baseline预备目录；六个fuzz臂全部未运行，
没有重试，后续臂未启动。另一个静态复核发现编译日志中的Windows路径转义：
原始Contains匹配context/fuzz路径均false，规范化双反斜线后均true；
这项尚未修复，不允许直接跳过来源检查。Task3整体/Task4–5仍未完成。
最终独立审查未进行，不能声称完整框架复现、历史根因或修复。

产品前后701文件清单Path/SHA256差异0；固定原context/fuzz hash仍为
765F99F7/D9DE1256全值。没有重跑原产品套件/Envelope/Pair，验收阻塞及main禁合并保持。
下一步需批准：先排查运行器收尾/路径证据判断，建立回归；保留本次失败收据，
再以明确新收据身份启动基线。不得在未获准前复用当前臂目录或重复运行直到通过。

## 批准后的运行器修复与新基线准备

用户批准修复收尾判断/路径解析，并用新收据重启基线。使用系统化排查、
TDD与原计划Native执行，范围仅现有诊断工具，不修改产品或正式工具链。
历史构建收据仍保留，不能据新结果推定上次remaining1的真实身份。

任务成员列表现在单独记录，DescendantsRemaining只排除确切创建的根ID，
不把根直接当子进程。记录ActiveAtRootWait、RootExited、成员列表及是否
请求清理；既有5s清理上限不变，最终计数和成员列表必须均为空。
真实存活子进程即使后来被清理，仍拒绝基线；列表无法核实则失败关闭。
编译来源按Go日志的引号/双反斜线解析，要求同条命令包含精确副本compiler、
对应package及source token；错根、后缀伪装、错包、错compiler均拒绝。

根计数、转义路径、新收据身份分别真实RED→GREEN；首个uint64别名错误仅是
测试编写错误，不计RED。12项修复回归+18项运行器检查全部通过；包括根退出
而真实子进程仍在的对照及清理后仍拒绝。另8项复建、16项真实链接隔离通过。

裁决：重验framework-001完整11536文件/hash/link清单后复用未插桩副本，
不采用旧fixture.exe。retry-001单独创建baseline-retry-001和plain-retry-001，
缓存/构建/收据均为新目录，旧记录不覆盖；该入口只支持未启动fuzz的失败
baseline准备，不支持已运行臂的自动重试。代价是重新构建及完整清单核对；
若复用核对错误可能漏漂移，所以必须在启动前通过全清单验证。
本节先记录修复/准备，不提前宣称基线成功或后续臂已运行。

### retry-001实际结果与新的收尾阻塞

副本全11536文件清单差异0；原工具链全11536文件清单差异0（UTC13:05:03），
产品701摘要差异0。旧失败构建收据hash仍
E1989051F7B7630B04A457165E35D31A08AA8EA3B52191DC60F0CE570198A1EF。
新build UTC13:03:50.049–13:04:23.178（Asia/Tbilisi17:03:50–17:04:23），
exit0、278进程、计数/成员均0、未请求终止。编译来源token守卫通过，
fixture hash9FAA1E69E2DDFFB81339CC465B4300C10CA77F1ACFEC6D6CC2D4575B0433FBAB，
binary hashC3E6CD21E1327574B2FA499D16BDBCDF5F4C8F5DF9233BC0CED94C41D355F23F。

实际baseline UTC13:04:24.543–13:04:30.291（Asia/Tbilisi17:04:24–17:04:30），
exit0、238579次、5s预算、parallel1、count1、内部30s/外部60s不变。
TotalProcesses3、根退出时计数和成员均0、未请求终止、CleanupVerified=true；
实际版本go1.26.5 windows/amd64，GOROOT/GOTOOLDIR都在未插桩副本。
这是独立无错误fixture基线通过，不是产品或Envelope/Pair验收。

随后Task3现有controls-only模型重新核对源模板相同，在独立新收据
controls-repair-001运行：11项Preserved=true，但process gate真实失败。
UTC13:06:47.016–13:06:47.064，exit0、TimedOut=false、TotalProcesses2；
根ID23688已退出，job仍列不同成员24584、ActiveAtRootWait1、DescendantsRemaining1。
只清理所属job后计数和列表均为空、CleanupVerified=true，不能因此改判通过。
此记录首次明确非根成员；未捕获其映像/活性，不能断言真实工作进程或退出记账滞后。
不按历史PID追查/终止，不回写旧模型收据或推定旧构建失败原因。

停止后续臂：observe、parent-first、synchronized、crash、cancel仍UNRUN，
副本未插桩；Task3整体仍阻塞，Task4仅基线完成，Task5/最终独立审查未进行。
额外重试入口守卫对“retry已启动fuzz后再次运行”真实RED→GREEN，最终13项
回归+18项运行器通过；不重跑真实基线。下一步需要有界实时成员身份/活性
诊断以解释这个新阻塞，不能放宽收尾门槛或循环重试。原生产/main阻塞不变。

## 有界实时成员身份排查

用户同意有界排查，不授权重跑基线或放宽门槛。系统化排查/TDD先补证据：
实时job列表→仅query-limited/synchronize打开→同句柄IsProcessInJob核实归属→
读取映像与零等待状态/退出码，查询失败/归属变化标未知；每次最多32条证据。
新增可选运行期采样，10ms轮询、同一既有进程时限，不默认开启；正常仍100ms。
采样改变时序，不能算自然观察fuzz。任何证据不改剩余成员或清理判据。

已知真实PowerShell子进程的身份/运行态检查RED→GREEN；第一版取缺失属性报错
不是RED，修正测试后得到明确断言失败。运行期子进程采样也RED→GREEN。
最终20运行器+13修复回归断言通过；无需安装/系统设置变化/旧PID操作。

controls-member-001（UTC13:20:44.205–13:20:44.261，Asia/Tbilisi17:20:44）：
exit0、11模型对照保留、活跃计数曾1、实际捕获列表空，清理后0/空，门槛通过；
没有身份捕获，不把这次通过当旧失败修复。为弥补短进程结束后错过成员，仅再做
一次提前采样controls-member-live-001，随后停止；不是循环重试到通过。

该次UTC13:23:01.459–13:23:01.521（Asia/Tbilisi17:23:01），根26924、
TotalProcesses2、exit0、TimedOut=false、11模型对照保留；运行期与根退出时
均捕获归属确切job的成员26932，映像C:/Windows/System32/conhost.exe，
查询错误0。运行期wait状态running/退出码259；根退出后wait仍running，
随后另一读取退出码0。这些非原子读可能跨越退出，不能合并为同一时刻状态。
DescendantsRemaining1，实际gate再次失败；只清理本job后0/空，未改判通过。

本次残留映像身份已确认；为何创建及退出/记账先后关系未证实，旧失败成员
身份仍未知。原Go runtime/os/syscall源码未找到AllocConsole/AttachConsole
引用；这是只读搜索结果，不证明整个系统没有其他路径。
没有默认排除conhost、增加等待预算、改启动flags或进程全局清理。
后续五个fuzz臂UNRUN，Task3阻塞/Task5最终审查待办，原产品/main阻塞保持。
下一步建议另行批准收尾设计：在既有清理上限内研究自然退出与强制清理的
证据区分，保留超时/未知成员失败；本次不实现该语义变化。

## 批准的500ms自然收尾增量

用户批准短设计后实现：根退出后的证据收集、最多500ms自然观察和强制清理，
共用原有5000ms清理时钟。不按conhost或其他名字豁免；超时、未知/截断成员
证据或请求强制终止均不能通过。初始DescendantsRemaining/活跃计数/列表原样
保存，schema2另记自然收尾、决策时零计数/空列表、强制清理及实际耗时。
schema-less旧收据仍用旧严格剩余判据，旧失败不改判。版本/来源/构建/臂共用
完成验证器。Windows调度/native调用不保证硬实时；记录实际耗时，成功收据
超过500ms观察或5000ms收尾上限也拒绝。

真实短命子进程检查先RED；第一版fixture在ready信号后才首次导入Start-Sleep
模块，超出500ms，运行器正确拒绝。改为直接Thread.Sleep300，未改窗口或
程序判据；native检查GREEN后旧validator对初始剩余拒绝（validator RED），
新增schema2验证GREEN。真实长期子进程和超时仍拒绝；强制清理、未知证据、
超预算、决策时非零计数/成员、未知schema负面对照通过。最终31 runner+
13 regression+8 materialization断言全部通过、exit0。没有产品全套重跑。

唯一新模型对照controls-natural-001：UTC13:35:58.895–13:35:58.958
（Asia/Tbilisi17:35:58），exit0、TimedOut=false、TotalProcesses2、
根25008退出时原始剩余1、成员11884确认为本job的System32/conhost.exe。
自然收尾25ms，NaturalShutdownVerified=true、ShutdownEvidenceComplete=true，
决策时/最终活跃计数0、成员空，TerminationRequested=false、CleanupVerified=true。
11项模型Preserved=true，当前完成验证器通过；没有再次启动模型或基线。
仍保留wait running与随后退出码0的非原子读，不用它们伪造同刻退出状态。

新对照证明本次残留能在短窗口自然收尾，不证明为何创建、所有历史剩余均是
conhost或原Envelope失败根因。701产品摘要无变化，正式context/fuzz钉定hash
仍765F99F7/D9DE1256全值。本轮仅有界收尾增量完成；后续observe/parent-first/
synchronized/crash/cancel五臂UNRUN，完整框架复现/最终审查未完成，生产及main
门槛保持。源/测试/报告照常保存开发分支，不提交原始收据或工具链。
