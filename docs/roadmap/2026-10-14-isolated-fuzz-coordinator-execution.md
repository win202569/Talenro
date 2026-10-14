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
