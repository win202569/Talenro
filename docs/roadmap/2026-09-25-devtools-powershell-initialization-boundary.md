# PowerShell 开发工具初始化边界检查

历史记录：下文暂停状态属于 5.1 检查当时。后续已批准精确要求 Core 7.6.5 的设计和计划，
并形成部分实现；5.1 初始化证据仍保留。当前状态见[换机接续说明](2026-09-25-progress-recovery-handoff.md)。

日期：2026-09-25。状态：修订计划已获用户确认；任务 1 因明确的初始化停止条件暂停，未验收。

## 检查目标

已批准附录要求四个 Windows PowerShell 工具入口在任何子进程启动前建立私有 Job。
修订计划 Task 1 Step 4a 明确要求检查互操作加载是否自行启动编译器，不能满足时停止报告。

当前未提交的 verify-devtools.ps1 中，Initialize-DevtoolsProcessOwnership 先调用
Add-Type -TypeDefinition，再调用生成类型的 Initialize；后者才创建 Job 并加入当前进程。

## 本机证据

使用系统 Windows PowerShell 独立进程，以 -NoProfile -NonInteractive -File 调用专属诊断脚本。
诊断只编译随机命名的空静态类型，无方法、无系统调用、无项目代码加载。编译前订阅
Win32_ProcessStartTrace，仅筛选父 PID 为诊断自身且名称为 csc.exe 的进程启动事件。
结束时注销订阅并移除该订阅事件，不记录全机进程或命令行。

- PowerShell：5.1.26100.9444。
- 诊断父 PID：40452。
- 观察到编译器子 PID：48968。
- CompilerChildObserved：true；诊断退出码：0。
- 编译器提示空类型没有公开方法或属性，符合该无行为输入。

首次在受限工具宿主中订阅被拒绝，未执行编译；随后通过限定提权在目标 Windows PowerShell
中完成上述检查。订阅权限不足不是产品测试失败，也不计作 RED。

## 结论与限制

本机默认 Add-Type 源码编译路径会启动外部编译器。因此，当前“先 Add-Type，再建立 Job”
的初始化顺序不能证明或保证所有子进程从启动起就有归属。先前针对 Go 子孙的取消测试通过，
不覆盖这个更早的初始化窗口。

本次没有运行真实 Job 初始化函数，没有证明编译器泄漏，也没有测试所有 PowerShell 版本。
结论仅为当前实现方案与已批准先决条件冲突，不能泛化为所有 PowerShell 方案都不可行。

## 下一步边界

按已批准计划暂停此实现，并请求是否评估不启动编译子进程的初始化方式。
不自动放宽为“仅 Go 子进程受保护”，不引入外层监督程序或编码加载器，不修改安全设置。
仍保留 Windows PowerShell-only 方向、Unix Bash 和既有 Bash smoke 范围。

未修改产品代码、未执行 C12/bootstrap/完整普通测试、未下载或上传内容。
任务 1 未完成，任务 2/3 未开始；I2/I3 及 Defender 事件状态不变。
诊断脚本留在本计划的忽略工作目录中，不纳入产品提交。
