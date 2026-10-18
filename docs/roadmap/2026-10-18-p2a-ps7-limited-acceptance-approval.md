# P2-A 指定 PS7 离线范围验收批准记录

日期：2026-10-18。
状态：P2A_PS7_DECLARED_INTEGER_VECTORS_ACCEPTED。
LimitedAcceptanceApproved=true；RuntimeApproved=false；MainMergeApproved=false。

## 批准来源与精确范围

用户在提案提交 0228e7eb6f9535f09cfff1458d60bc0ebc24c531 后，对以下明确问题回答“是”：

> 建议仅验收：普通 Windows PowerShell 7.6.5 下，P2-A 指定离线整数换算测试范围。不包含真实运行、整体 C12 或 main 合并。
>
> 是否批准这项限定验收？

这是对该具体范围的人工验收决定，不是代理自行扩大测试或运行授权。
获批[提案](2026-10-18-p2a-ps7-limited-acceptance-proposal.md)完整范围、证据对应和
全部限制继续有效；提案自身的待批准状态保留为历史，不回写。
提案物理 SHA256：
2D5046F7AA325E6B5B9AB5832E78C13D3EEC2083581978B30598FE29A2162DA7。

验收对象为候选提交 cc764dba8617933cc1dda341cf46322f0135cf7f 中
docs/diagnostics/isolated-fuzz/p2a-type-receiver-repair/ 的固定版本：

| 文件 | SHA256 |
| --- | --- |
| offline-timebase-plan.ps1 | 350CAB1A36CECEA57D8335581ADC780CE1F17BDC9379D63F03130C6EEF4FC35B |
| full.tests.ps1 | B5C4E0CE14F0B08C647A966884BE6496AA9C05A5D5BEA827B6F86F87EAFCAA65 |
| fixture-builder.ps1 | C720E247298FA84B8BD348A4A3FBDB3759B7F1F2A510EA0D67D9479E00045A58 |
| expected-output.json | DA3CF1837213F9451A2CB2941DC248B663116CD52EEC199D1D01646A90CA0225 |

## 验收依据与剩余边界

[已保存的五轮结果和独立只读审查](2026-10-18-p2a-type-receiver-repair.md)
支持上述普通 PS7.6.5 Core x64 指定向量：
修正前失败、修正后24组通过、两项指定变异被检出、还原后24组再次通过；
限定审查无 Critical/Important/Minor findings。
这些是先前执行结果，本批准登记未重新运行或扩大覆盖。

接受固定版本在已测范围内的数值、保守拒绝、独立返回及指定顺序隔离结果；
不接受所有输入/平台/引擎/并发条件的普遍保证。特别保留：

- PS5、其他系统/引擎/提权环境、完整 P1/P2/C12、P2-B/C、P3、Stage B/I2/I3、
  安全放行、main 合并和发布均未由本记录批准。
- 真实 QPC/UTC、调度或硬截止、身份/会话所有权、进程停止/取消、管道、清理、
  持久化及真实运行均不在验收内。
- CodeProperty/internal-error 动态路径、灾难恢复、敌对 type-data/并发修改、
  字符串重新读取元数据持久性、控制器故障/安全边界等，继承提案的未验证限制。
- 产品 SourceVerified、LiveActionsPerformed、RuntimeApproved 三字段保持 false；
  人工离线验收不改写运行事实或权限。
- 最新5/5测试和1/1审查已消费，原2/5、后续3/6停止路线的余项仍冻结；
  两次单独诊断已消费。旧失败不改写，任何额度不因验收恢复。

## 本次登记与换机恢复

本轮仅只读核对提案和四份核心文件身份，新增本批准记录，更新进度和恢复说明。
0新增测试宿主、0新增审查、0产品导入、0源码变更、0真实运行。
按长期授权提交推送必要文档到现有开发分支，并核对远端一致；
不提交或删除 ignored 原始 packet、缓存、二进制或凭据。

换机从最新 codex/c12-b01-task9-coordinator 恢复源码、预期、结果和批准记录；
不从 main 推定进度。clone 不恢复本机原始记录、历史映射、引擎缓存或会话，
也不授予自动重播权。

本次动作到批准登记和保存为止。下一阶段需另行明确目标、设计与执行范围；
本记录不自动启动下一阶段，也不提升其他入口。
