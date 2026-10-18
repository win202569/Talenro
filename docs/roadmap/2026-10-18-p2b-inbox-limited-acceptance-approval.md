# P2-B 有界接收层：限定离线验收批准记录

日期：2026-10-18。状态：LIMITED_OFFLINE_ACCEPTANCE_APPROVED。
LimitedAcceptanceApproved=true；RuntimeApproved=false；MainMergeApproved=false。
FullP2BAcceptanceApproved=false；FullC12AcceptanceApproved=false。

## 用户决定及不可扩大的对象

用户在提交b459894dd791f8dad578278248a264b59d6ff504之后，对已展示的
“是否批准仅验收固定版本在PS7.6.5下已测离线接收功能，保留标量附加属性
隔离测试覆盖缺口及全部限制，不包含真实运行、完整C12或main合并”
问题明确回答“是”。本记录仅登记该决定，不把此前泛化的“继续”追认为验收。

被批准提案：
[有界接收层限定离线验收提案](2026-10-18-p2b-inbox-limited-acceptance-proposal.md)，
SHA256为26B1FBA8D2542BD4CF93963EA6481B4A8A68E68007B7A23BE4C2FE908C355F11。
该提案所有固定文件、限制1–10及后续边界一并约束本批准，原提案保持历史原文。

验收产品为提交3c56edc35d7463734eb4dc2bf8bfbbdc99ac80e9中的
docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/inbox.psm1，
SHA256为47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC。
仅普通Windows PS7.6.5 Core x64、固定SDK、稳定被动输入、串行调用和已测向量。
包括已测四方向256字节组帧、有限FIFO/心跳槽、声明绑定/连续序号、
EOF/首因停止/Inspect及已测所有权隔离；不是真实端点认证或握手能力。

## 明确接受的保留项

Minor2继续延期：新Count.Probe修改缺少直接证明附加属性在修改对象保留、
在其他快照缺失的断言。IB-Eq不比较scalar ETS metadata；已有fresh-scalar
guard/引用检查不替代缺失的动态断言。本决定不声称覆盖已补齐，不授权
修改冻结套件、补测或自动修复。

历史Important1是派生摘要大整数舍入，已经用户单独授权仅校正8个整数
文本，并由父代理只读比较9份嵌入记录与原始文件零差异。原审查仍为
历史Critical0/Important1/Minor1，不改称独立审查零意见或重新审查通过。
[完整审查及处置](../diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/review.md)和
[校正记录](2026-10-18-p2b-inbox-evidence-correction.md)继续适用。

历史对象身份不可回溯；中间还原时点缺独立packet记录；私有helper不是
任意/循环/篡改树的反序列化或安全边界；其他环境/并发/真实I/O未验收；
只读审查不是独立复现。6144字节仅原始帧稳态容量，非CLR总内存硬上限。
所有其他提案限制保持不变，不能因人工限定验收而省略。

## 未获批准的事项与证据保留

本决定不包含真实运行/能力/释放/清理、上层握手、完整P2-B会话、
P2-C/P3、Task9集成、全仓库/全C12、owner/vendor安全放行、发布或main。
源返回的SourceVerified/LiveActionsPerformed/RuntimeApproved仍false。

原保存结果及portable-summary中的FormalAcceptance=false等字段不改写：
它们记录验证发生时的权限和结论；后来的人类限定决定以本记录为准。
不修改codec、产品、测试、原始packet、审查原文或历史失败，不追认旧候选。

本轮只读核对提案中8项固定文件一致，登记决定并更新进度/恢复文档。
0新产品目标、0测试、0编译、0审查、0产品变更。
新5/5宿主及1/1审查已耗尽；旧原2spent+3FROZEN、独立诊断1spent不变。
人工限定验收不恢复任何测试、实现、运行或审查预算。

## 后续与换机恢复

下一步应另行界定会话状态/握手子单元的设计、计划和有限验证范围；
本轮没有开始其设计或实现，不把本批准当作live开关或后续执行许可。
Minor2若需补齐，也需单独范围与有限验证安排。

按长期授权普通提交推送既有codex/c12-b01-task9-coordinator开发分支。
换机从该分支最新提交读取current-status、本记录及repository-recovery，
而非默认main。Clone恢复文档、源码与可携带结果，不恢复ignored原始
packet、CLR身份/token、缓存/依赖/凭据或执行权限，不得自动重播旧验证。
