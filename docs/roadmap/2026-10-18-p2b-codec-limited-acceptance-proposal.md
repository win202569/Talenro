# P2-B 纯控制帧：限定离线验收提案与后续边界

日期：2026-10-18。状态：PROPOSED_PENDING_EXPLICIT_HUMAN_ACCEPTANCE。
LimitedAcceptanceApproved=false；RuntimeApproved=false；MainMergeApproved=false。

## 本次请求与提案

用户在上轮修复保存后回答“继续”，本轮据此只读梳理证据与剩余工作，
编写本提案并按长期授权保存。该回答不是对尚未展示的具体验收范围作出批准；
本提案不登记人工验收、不新增实现或执行预算。

建议仅验收提交8500aef9a7c4be52f793944ad4ac356451d2d638中
docs/diagnostics/isolated-fuzz/p2b-cold-cache-repair/ 的固定版本，在普通
Windows PS7.6.5 Core x64、SDK7.6.0.500、稳定输入和已测向量范围内的：

- 精确256字节控制帧编解码、独立literal对应和声明绑定校验。
- 指定方向、角色、声明序号/引用/时基检查及保守拒绝。
- 已测根描述符与标量/数组附加描述符拒绝、独立返回和连续调用隔离。
- 空缓存与真实无共享表的区分、已测共享键/独立键路径和原36个漏检拒绝。

Accepted仅表示纯数据操作成功，不是角色身份、会话能力、释放许可或运行批准。
SourceVerified、LiveActionsPerformed、RuntimeApproved始终false。
这是纯codec子单元，不验收完整P2-B会话所有权架构。

## 固定对象与证据地图

| 文件 | SHA256 |
| --- | --- |
| control-frame.ps1 | 9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C |
| control-frame.tests.ps1 | 7CBC263600948F9806D8749B43D0B9DB55C100F48757B1C5CC49CF94A92C60B1 |
| scalar.tests.ps1 | 809BCC35F8231CFD7356BBAB884EDB89DA03CE10651BAA54EB6B76A0F90A6AE5 |
| cache-fixture.ps1 | 0010A0FC2A8C1830E1DDFDA732DF83AF5E11412AADF19CF159A950F9E82EACD6 |
| fixtures.json | 5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D |
| expected-output.json | 42BF546A3D9ACFE9F9BAC30A2DBBB06E5D6BFC417BFA5F3A1EF97D150E15BFF1 |
| portable-summary.json | 362808EA91F100312B6720C72B68BBC26A74017BC0E603D60DCCF622964F2595 |
| review.md | EED4BC3C7937279403ED0CF53FEDFCD895B522169676AB5B21080CD8FF8EBC1C |

[修复执行报告](2026-10-18-p2b-cold-cache-repair.md)和
[完整独立审查](../diagnostics/isolated-fuzz/p2b-cold-cache-repair/review.md)
提供最新固定版本的主要依据。保存的结果为：

| 验收对象 | 已保存依据 | 不可推导的结论 |
| --- | --- | --- |
| 编解码与声明绑定 | 原33组完整回归、36行精确oracle，literal非产品生成 | 真实端点/身份认证、运行状态机 |
| 空缓存共享元数据 | RED恰好36漏检；修复后84/84，38接受/46拒绝 | 所有SDK历史与所有输入的普遍正确性 |
| 描述符与所有权 | 27proofs、14API histories、35matrix、182malformed、45guards、54key controls | 任意私有反射下不可伪造的来源 |
| 限定代码质量 | 独立只读审查0Critical/0Important/0Minor，无未完成项 | 独立重跑、完整安全审计或全项目正确性 |

本轮只读重算1185个mapped pins均一致，三个保存回执均ExactExpected、
无timeout、owned exit和streams complete。未运行产品/测试/控制器。
先前指定两项变异及还原结果见
[历史negative controls](2026-10-18-p2b-codec-negative-controls.md)；
它们属于旧FA6047A4版本，不能声称在最新9A532E17重新执行过。
历史Important、失败和冻结额度保留；最新结论仅由其自身修复证据支持，
不追认旧候选为已验收，也不宣称重新执行了完整历史变异链。

## 必须一并接受的限制

1. 仅已测普通PS7版本；PS5、其他SDK/架构/平台/提权环境未验收。
2. 冷缓存nested-wrapper、empty-existing-table组合没有专门动态向量；
   继承键计算的静态检查不能替代这些组合的运行证据。
3. CounterNote计数为零不是对所有可能getter的普遍动态插桩证明；
   灾难失败、完整internal-error注入和fresh-copy失败异常链未全面验证。
4. 稳定输入之外的并发修改、TOCTOU、敌对type-data、私有反射、伪造helper、
   SDK内部破坏、进程隔离与不可伪造factory provenance不在合同内。
5. 控制器仅有普通工作流顺序/预消费保护，不是对抗性防重放账本；
   成功有限宿主不证明超时/kill/drain故障分支或硬实时保证。
6. 真实QPC/UTC、端点、分片组帧/队列、心跳计时、Job/进程能力、
   一次释放、失联处置、实际清理、字节来源、EOF和持久化均不在codec内。
7. P2-C、P3、Stage B/I2/I3、Task9完整集成及其后任务、全C12、
   安全owner/vendor放行、main合并和发布没有被批准。
8. 最新3/3宿主及1/1审查已用完；旧停止路线和余项保持冻结。
   本轮0新增宿主、0新增审查、0代码变更，不恢复任何预算。

## 剩余工作顺序（依赖图，不是实施计划或执行许可）

1. 用户明确决定是否接受上述固定版本的限定纯codec范围。
2. 若接受，另写后续P2-B子单元设计/计划：控制状态、有限队列与
   一次性握手/失败语义须明确分离纯声明和真实能力；不默认实现或运行。
3. 真实P2-B适配仍须独立安全前提、精确版本/输入、有限验证与运行批准。
   P2-C就绪绑定未经独立设计验证，真实T不得释放。
4. P2-C双通道采集与P3真实记录/持久化分别设计、验证和批准；
   不复用synthetic成功记录冒充真实采集或归档证据。
5. 然后再回到Task9 Coordinator集成和独立验收阻塞；
   Task10、Tasks11–14、Tasks15–18及其余C1.2批次仍依原依赖顺序推进。
   纯codec成功不能关闭这些事项；本轮未重新审查它们的源码或当前环境。

职责和边界依据
[P2-B设计§7–11](../superpowers/specs/2026-10-18-p2b-owned-session-design.md)；
原设计起草/原codec计划停止状态留作历史，不覆盖后续批准和新结果。
本提案不决定新的实现技术或增加功能，因此没有新实现计划或运行包。

## 保存与待用户决定

仅新增本提案、更新current-status和repository-recovery，提交推送现有
codex/c12-b01-task9-coordinator分支。clone恢复文档、源代码和便携结果，
不恢复原始记录、实际CLR引用、缓存/SDK/凭据，也不恢复测试或运行权限。

待明确问题：是否仅批准上述固定版本在普通Windows PS7.6.5下的指定
纯控制帧codec离线验收，保留全部限制，不包含完整P2-B、真实运行、全C12或main？
