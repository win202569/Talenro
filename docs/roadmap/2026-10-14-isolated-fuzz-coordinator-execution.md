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
