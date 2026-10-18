# P2-B 纯控制帧：指定 PS7 离线范围验收批准记录

日期：2026-10-18。
状态：P2B_PURE_CODEC_PS7_SPECIFIED_OFFLINE_VECTORS_ACCEPTED。
LimitedAcceptanceApproved=true；RuntimeApproved=false；MainMergeApproved=false。

## 批准来源与固定对象

用户在提案提交3da72edd014ef9e2fc88a1cc49cf952df79e2495后，对以下明确问题回答“是”：

> 是否批准：仅验收普通 Windows PS7.6.5 下已测的纯控制帧编解码与声明绑定校验范围，保留全部限制，不包含完整 P2-B、真实运行、整体 C12 或 main 合并？

这是对该具体范围的人工决定，不是代理自行扩大验收或执行许可。
[获批提案](2026-10-18-p2b-codec-limited-acceptance-proposal.md)全部范围、
固定文件身份、证据地图、八项限制和预算约束继续有效。
提案物理SHA256：
A4A88B7834CF57BCC3090F6948591675444933DCE09419228BACB9C3A0EAAD60。
原提案待批准状态保留为历史，不回写为已经获批。

验收对象为提交8500aef9a7c4be52f793944ad4ac356451d2d638中的
docs/diagnostics/isolated-fuzz/p2b-cold-cache-repair/固定版本。
产品control-frame.ps1 SHA256：
9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C。
本轮只读重算提案列出的八份文件，均与固定摘要一致。

## 验收范围与证据

仅普通Windows PS7.6.5 Core x64、SDK7.6.0.500、稳定输入、指定已测向量：
精确256字节编解码、独立literal对应、声明方向/序号/引用/时基绑定、
保守拒绝、已测被动描述符检查、独立返回及连续隔离，
包括已测冷缓存共享/独立键路径和36项漏检修复。

[执行报告](2026-10-18-p2b-cold-cache-repair.md)和
[完整独立审查](../diagnostics/isolated-fuzz/p2b-cold-cache-repair/review.md)
记载修复前36个指定漏检、修复后84/84、原33组完整回归/36行oracle，
以及0Critical/0Important/0Minor审查。均为保存的历史结果；
本次批准登记没有重新运行，也不声称独立重现运行环境。
旧版本两项变异不冒称在本固定新版本重跑。

## 仍然未验收或未授权

- 完整P2-B会话所有权、真实能力/启动/停止/清理、端点/时钟、
  分片组帧/队列/心跳、真实双通道采集和归档均不在本批准内。
- PS5及其他SDK/平台/架构/提权环境、所有输入的普遍正确性、
  冷缓存nested-wrapper/empty-existing-table专门动态覆盖未验收。
- 敌对反射/type-data、并发修改/TOCTOU、隔离、不可伪造factory来源、
  灾难故障、完整internal-error和控制器故障路径继承提案限制。
- P2-C、P3、Stage B/I2/I3、Task9完整集成和后续任务、全C12、
  安全owner/vendor放行、main合并及发布未批准。
- SourceVerified、LiveActionsPerformed、RuntimeApproved仍false；
  人工离线验收不是运行事实或能力。
- 最新3/3宿主、1/1审查已消费，旧失败及冻结余项不改判、不解冻；
  clone或本次验收均不恢复预算，不授予重播许可。

## 本轮登记与交接

本轮仅只读核对提案/八份固定文件和Git身份，新增批准记录，
更新current-status及repository-recovery。0新增测试宿主、0新增审查、
0产品导入/源码修改、0真实运行。文档表格核对器初次匹配遗漏含数字的
文件扩展名，在任何文件hash检查前停止；修正只读匹配后八项全部一致。
这不是产品测试失败、目标重跑或修改原始结果。

按长期授权提交推送这三份必要文档到现有
codex/c12-b01-task9-coordinator，并确认远端一致/工作区干净。
换机从最新开发分支及本记录恢复，不自动使用main；不恢复原始packet、
实际CLR引用、运行时/缓存/凭据，也不继承新的执行权限。

本轮到批准登记与保存为止。下一建议：另行编写后续P2-B子单元的
详细设计/计划，分离控制状态、有限队列、一次性握手和真实能力；
须确定新范围，不自动实现、测试、启动真实目标或提升入口。
