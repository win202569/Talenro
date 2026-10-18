# P2-B 有界接收层：限定离线验收提案

日期：2026-10-18。状态：PROPOSED_PENDING_EXPLICIT_HUMAN_ACCEPTANCE。
LimitedAcceptanceApproved=false；RuntimeApproved=false；MainMergeApproved=false。

## 提案及固定对象

用户在摘要校正保存后回答“继续”。本轮据此只读核对现有证据并编写
本提案，不把该回答登记为对尚未展示的具体验收范围的批准。
不新增实现、运行、测试、编译或审查预算。

建议仅接受提交3c56edc35d7463734eb4dc2bf8bfbbdc99ac80e9中
docs/diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/的固定产品版本，
在普通Windows PS7.6.5 Core x64、固定SDK、稳定被动输入、串行调用、
已测离线向量范围内的有界字节接收内核：

- 四固定方向的256字节分片组帧、声明方向/Binding及连续序号检查。
- 每方向4普通FIFO槽、1最近心跳槽及1组帧缓冲；普通槽总计16。
- EOF/截断、首因锁存、停止后拒绝Feed/Take及可独立Inspect的已测语义。
- 已测输入复制、队列/快照/空数组所有权、跨快照及跨实例隔离。
- 本次仅Copy-IBTree/私有分类helper修复及其已测wrapper/字段所属关系。

Accepted只表示本层纯数据操作结果，不代表真实身份、握手许可或能力。
SourceVerified、LiveActionsPerformed、RuntimeApproved始终false。
保留完整独立审查的限制及下述明确测试覆盖例外，不接受为全项目正确。

| 固定文件 | SHA256 |
| --- | --- |
| inbox.psm1 |47BABFEDFDA73C850EA49CC74E3B6348919EAC69E7AFBFD3DCA907B4988C02FC|
| inbox.tests.ps1 |904F20D4D60540D7A312F5ACF14B0E1E01D263C7A905B18F578FB0846C7267CE|
| tree.tests.ps1 |854909B72BF0066BB03858E7801130AD106F741F6C0082D8364B227E7CF57F3D|
| test-helpers.ps1 |4F418D2BE07ACE9597CE7A90438E7325F692A3C578010580548299016A2A5E8D|
| fixtures.json |5045D321677C5FC3ABFA8674CF23CA9062C86A92D64E0B596001564E303D2A8D|
| expected-output.json |EFE0E669AA7F1980ED6A0CDBFDC31646E6701D2A47F1F2879D8DD974FAB50D4A|
| portable-summary.json |7E64C8EED19BEFC8C75775E8077A79568ABBEA06A6DD3FD035E1CBFE12D2B185|
| review.md |7EB19C477EAEE54D3BCB31A5C2A5120C0BBE7EF7F6DB120825915B275F8E3F49|

固定codec依赖9A532E17D33CAB1FEE9D8BB24C41D5A4B867E42F5EBAB0EE3C3C29C9CDC1093C
未改变，其[原有限批准](2026-10-18-p2b-codec-limited-acceptance-approval.md)不扩大。

## 已保存证据，不是本轮重跑

[执行记录](2026-10-18-p2b-inbox-tree-copy-repair.md)、
[完整审查及所有处置](../diagnostics/isolated-fuzz/p2b-inbox-tree-copy-repair/review.md)、
[无损校正记录](2026-10-18-p2b-inbox-evidence-correction.md)共同限定本提案。

| 保存轮次 | 实际记录 |
| --- | --- |
| RED |基线43067AD0，指定首次Snapshot失败，exit1/3行|
| full |修复47BABFED，新12+原32组，exit0/49行/44PASS|
| capacity |同修复版本唯一容量变异，指定错误接受断言，exit1/4行|
| sequence |同修复版本唯一序号变异，指定错误接受断言，exit1/4行|
| final |精确还原47BABFED，exit0/49行/44PASS，和full输出/两trace哈希相同|

五条回执均EXPECTED_WITHIN_SCOPE，无timeout/stderr，owned exit与drains完成。
原32组的两份full/final trace各记录CompileCount1、Observabletrue、
Counters0、Synthetic4、Decorations8。只有FirstSnapshot单独跑过RED；
不能声称其他11专项各自具备RED运行证据。

独立审查未识别允许修复内的产品缺陷，但历史为Critical0/Important1/
Minor1，不得改称独立审查零意见。Important1是派生摘要大整数舍入，
经用户另行批准，仅恢复8个JSON整数原文；父代理只读比较9份嵌入记录
与原始文件零差异。不是重审，不重写原始记录或原审查结论。
本轮再核1326pins及五条保存回执/流/trace哈希一致；0产品/测试/控制器调用。

## 必须连同接受的限制

1. **Minor2明确延期，不冒充覆盖完成。** ReturnedMutation/TwoSnapshots
   给Count附加Probe，却未直接证明该附加属性在修改对象上保留且在另一
   快照上缺失；IB-Eq不比较scalar ETS metadata。原codec fresh-scalar
   guard及引用断言提供支持，但不是这些缺失断言的动态替代。
   如要求补齐，须另定测试变更/有限验证范围，不能修改已冻结套件或补跑。
2. 仅指定普通PS7/SDK、稳定输入与串行合同；PS5、其他运行时/平台/架构、
   提权、并发修改/TOCTOU及敌对type-data不在验收内。
3. Representation只观测新构造的测试对象，不能回溯历史消失对象的身份。
   中间两次还原由父代理工具输出中的精确哈希检查观测，五条packet回执
   未独立记录两次还原时点；当前及final使用的good身份可核。
4. helper只服务内部固定自有树；不接受任意调用者树/循环图，不抵抗
   同进程反射、私有scope调用或SDK内部篡改，不是安全反序列化边界。
5. 6144字节是原始槽/组帧稳态留存容量，不是CLR总内存硬上限。
   本层不保存无界消息历史；测试不证明所有灾难失败或硬实时保证。
6. 宿主控制器的预消费/顺序/owned-process保护不构成对抗性防重放账本；
   成功完成这些有限宿主不验证所有timeout/kill/drain异常分支。
   独立审查严格只读，并未独立重跑。
7. 不包含真实端点/管道/时钟/心跳计时、进程/Job/释放能力、上层握手许可、
   跨进程或重启防重放、真实清理、持久化或来源认证。
8. 不包含完整P2-B会话、P2-C/P3、Task9完整集成、全C12、全仓库验收、
   owner/vendor安全放行、main合并、发布或真实运行。
9. 新5/5宿主与1/1审查已用完；旧原2spent+3FROZEN和独立诊断1spent不变。
   本轮0新目标/编译/审查/产品变更，不恢复任何预算。
10. 摘要保留精确JSON整数文本；消费者不得用JavaScript Number解析后再
    回写这些大整数。Clone不恢复ignored原始packet、CLR身份、token、
    缓存/依赖/凭据或执行权限；不得重建缺失旧证据或自动重播。

## 后续依赖及待决定事项

若用户明确接受，应仅登记本固定版本的上述限定离线验收及Minor2延期，
再另行界定后续会话状态/握手子单元的设计、计划与有限执行预算。
不能仅加live开关即视为真实P2-B适配获准；P2-C/P3及Task9集成仍需各自
前提和独立验证。若用户不接受Minor2延期，先为该覆盖缺口另定工作范围。
本提案本身不实现这些功能、不分配预算、不表示人工验收已发生。

按长期授权仅保存本提案、current-status及repository-recovery到现有
codex/c12-b01-task9-coordinator开发分支；不合并main。

待用户明确决定：是否仅接受固定47BABFED版本在普通PS7.6.5下、指定
离线向量内的有界接收层验收，接受上述全部限制与Minor2延期，不包含
真实运行、完整P2-B/全C12或main？
