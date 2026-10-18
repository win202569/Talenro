# P1 owned-return：PS7 限定离线验收提案（待用户决定）

日期：2026-10-18。整理基线：1506490922b64fecfcb93acfe2085f0dcb50a775。
状态：PROPOSED_NOT_APPROVED；FormalAcceptance=false；PromotionApproved=false。

用户“按你的建议继续”仅授权起草本提案及保存进度，不是批准下文拟议验收结论。
本轮只读整理和文档提交；没有产品导入、测试、变异、修复、独立审查或 main 合并。

## 1. 建议用户批准的精确结论

> 仅对下表固定 owned-return 候选版本，在已有 Windows 普通用户 PS7 保存证据
> 所覆盖的合成输入和顺序调用向量内，作出“PS7 限定离线功能验收”判断。
> 采用同版本 API、旧所有权回归、后返回值矩阵及最新限定独立审查的组合证据。
> I1/R1 仅记为 AddressedWithinStatedPS7Vectors，不全局关闭历史问题。
> 接受本文逐项披露的证据限制；不声称原 PS7/PS5.1 双引擎 P1 合同已经完成。

这是明确提出的较窄验收档案，不是修改源码合同、删除测试或重写旧审查。
只有用户对第 1、4、5、6 节明确批准后，才能另写实际批准记录。
批准后仅更新文档层面的上述限定验收状态，产品 Acceptance 等字段仍为 false。
本提案本身以及普通保存/推送操作不构成该批准。

| 固定对象 | 文件与本机已验证字节 SHA256 |
| --- | --- |
| 候选产品 | docs/diagnostics/isolated-fuzz/scalar-owned-return-candidate/offline-supervision-policy.ps1；1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8 |
| 全 API 测试 | 同目录 offline-supervision-policy.tests.ps1；61A3BACF26EDE1F8B137CBF77CC4DBC35E4D6E6FB8FE6365953CDA9F41CEDEA3 |
| 旧所有权回归 | 同目录 offline-supervision-policy.ownership.tests.ps1；C4905925B309B00E94834B2FDA3EE8034265B23CDF274276C9ECA5511250C6E7 |
| 已保存 PS7 引擎身份 | 362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139；不是所有 PS7 版本的兼容性声明 |
| 独立审查报告 | [完整审查和全部处置](2026-10-18-p1-owned-return-readonly-review.md)；8CD62874C9375561592BE1BE3ED29C579B62FC5E464980B9C0392715B92EC463 |

当前原产品 953CC2A8 和 Integer/Token 候选 F9B4E408 不在此次验收对象中。
不覆盖整个分支、整个项目、已验收 store/raw coordinator 的重新验收或后续候选版本。
摘要是本机物理执行快照身份，不是 Git blob 摘要或任意换机行尾配置的字节承诺。

## 2. 依据与不变合同

依据[原 P1 计划](../superpowers/plans/2026-10-17-offline-supervision-policy.md)、
[联合设计第 9/10 节](../superpowers/specs/2026-10-17-raw-capture-supervision-design.md)
以及最新独立审查。旧设计/计划中的停止状态是历史记录，不据此重复启动旧额度。

候选仍只提供 Get-OfflineSupervisionDecision 的被动合成监督决定：

- 精确输入预检、13 个根输出字段、22 个 Decision 字段和独立返回快照。
- 单一绝对 120000ms 合成预算，113500ms 等值截止；natural/cleanup/publication
  为共享 500/5000/1000ms，不能追加时间、重置起点或抹去先前失败。
- 每通道 1048576、合计 2097152 字节的模型计数，数据事件 4096，控制槽 16、
  每条 1024 字节；这是模型约束，不是实际传输/内存/性能保证。
- Mode=policy-only，EvidenceClass=synthetic；OriginVerified/Acceptance/
  RuntimeApproved/LiveActionsPerformed=false，CleanupStatus=unknown。
- Would* 只表达假设决定。没有真实时钟、进程、管道、持久化、系统查询、回调、
  imports 或运行授权；不会把声明的 cleanup verified 转成真实清理事实。

## 3. 已有证据采用方式

| 证据 | 保存结果 | 可以支持 / 不可以推导 |
| --- | --- | --- |
| [同版本 API 与旧所有权回归](2026-10-18-p1-owned-return-regression-v2-result.md) | 35 组/38 行，23290ms；8 fixture+10 regression 组/22 行，6464ms；两次 exit0、stderr 空 | 同候选 PS7 正反例；不是 PS5 结果、实时保证或覆盖率百分比 |
| [优化后实际所有权矩阵](2026-10-18-p1-inline-ownership-timing-result.md) | 39 case、80 decision、29 scalar、1 metadata；19954ms；53 FACT | 指定顺序输入/A/B 所有权及字段检查；不是所有叶子/状态穷尽证明 |
| [observer 等价性](2026-10-18-p1-inline-observer-v2-result.md) | 45 例 | 检查器具名分支辨别；不是产品自身变异证明 |
| [限定独立审查](2026-10-18-p1-owned-return-readonly-review.md) | 0 Critical/Important/Minor 被识别；66 个输入 pin、676 个继承条目完整 | 完整候选及保存证据限定判断；不是独立见证执行、全分支审查或正式验收 |
| 历史原 I1、Token R1 与实际污染 witness（审查报告内索引） | 历史问题和失败保留 | 解释修复动机；不得改写为旧版本通过 |

独立审查报告完整列出六组原始输出、存储转换证据、摘要及父代理核对方式。
本提案不再执行这些测试，也不把摘要当成执行认证。原始记录仍留在本机 ignored
目录；便携报告可从 Git 恢复，原始记录不会随 clone 自动恢复。

特别保留 String 语义区别：32 条 Int32/Int64/Boolean 加 3 条 stored PSObject
String 的 held 与 fresh-field reread 均保留；18 条 raw String 仅 held-view
保留、fresh-field reread 不保留。不能把最后 18 条描述成“字段重读附加元数据也保留”。

## 4. 必须明确作出的两项范围决定（推荐 A）

| 议题 | A：本提案推荐的限定路线 | B：保留完整双引擎路线 |
| --- | --- | --- |
| PS5.1 / 双引擎 | 明确只接受已有 PS7 向量；PS5.1 兼容性未验证、未验收的状态不变（不等于已证不兼容），完整双引擎 P1 门槛继续开放，不标为完成 | 不作本次限定验收；先另拟同候选 PS5 证据的有限计划，再经批准执行 |
| 新产品变异 / omission mutant | 明确接受此证据缺口，不作为此次狭窄文档验收的先决条件；依据正向 backing/descriptor/payload 检查及限定审查，不冒称新产品变异已通过 | 保留为待决定证据义务；先确定具体能区分的缺陷与有限预算，不默认无限探测 |

选择 A 不是判定 PS5 已兼容，不是将观察器负对照冒称产品 mutation，也不是
永久豁免未来推广/完整合同的证据义务。新产品 omission mutant 尚未运行是事实。
它不自动使全部正向证据无效，但必须随这次限定判断明确接受其不确定性。
两项范围决定必须一并包含在用户批准中；不能从“测试通过”推定接受。

## 5. 随批准完整保留的 23 项旧限制与 9 项新增限制

以下编号与独立审查报告对应；本节是继承清单，不改变该报告的具体理由和处置。

| 旧项 | 保留的边界 |
| --- | --- |
| O1 | 审查未见证当前执行，只读保存证据 |
| O2 | 有限向量，不是穷尽或普遍所有权证明 |
| O3 | 合成毫秒，不是 QPC/跨进程时基证明 |
| O4 | 无真实容器、启动、句柄适配/OS 兼容证明 |
| O5 | session/actor 标签不是身份认证 |
| O6 | 模型计数不证明并发管道、背压或字节来源 |
| O7 | Would* 不证明实际取消、停止或清理 |
| O8 | 模型截止和保存耗时不是物理硬实时保证 |
| O9 | 模型 readback 不是持久化、原子性或掉电证明 |
| O10 | 成功捕获未遍历 timeout/kill/drain 分支，旧失败不改判 |
| O11 | 所阅 stderr 为空，非空 CLIXML 分支未动态证明 |
| O12 | 普通 catch 不保证灾难性恢复 |
| O13 | 并发/敌对 runspace 不覆盖；普通顺序调用并非被排除 |
| O14 | 哈希与记录不是历史执行或变异时序认证 |
| O15 | 继承依赖仅核哈希，不是重新行为验证 |
| O16 | Git 身份由父代理核验，不是审查者独立认证 |
| O17 | 无全分支、全项目正确性结论 |
| O18 | 安全 owner、P2/P3、runtime、main、C12、release 仍分别设门槛 |
| O19 | 同候选 PS5/原双引擎合同未验证完成 |
| O20 | reflection/shape 额外套件不能被不同目的检查追溯替代 |
| O21 | 无新的 Token/产品特定变异运行 |
| O22 | 单次 timing 不是 benchmark，旧 timeout 仍保留 |
| O23 | 未覆盖的 binding/descriptor 上下文不获普遍保证 |

| 新项 | 保留的边界 |
| --- | --- |
| N1 | 同候选 PS5/完整双引擎门槛仍开放；本次只拟议限定 PS7 |
| N2 | 无 owned-return omission/product mutant；按第 4 节明确处置而非冒称已测 |
| N3 | recoverable internal-error reset 未动态强制；结构拒绝走 catch，internal-error 静态评估，不保证分配器/SDK 持续失败恢复 |
| N4 | 未对每个输出叶子/状态逐个做返回后附加；共享构建器仅支持限定判断 |
| N5 | raw String held-view 不等于 fresh-field reread 持续保留 |
| N6 | 人工 reread decoy 仅分支敏感性；sharing 控制仅覆盖 Int32/wrapped String |
| N7 | 单次 matched-prefix 比较不是重复基准，也不修复历史时序认证 |
| N8 | 物理摘要受换机行尾配置影响，不是普遍 checkout 字节承诺 |
| N9 | 文本检查器不是进程回执验证或执行认证 |

没有一项因本提案被记成“已测试通过”。不把这些限制泛化成每项都必须自动加测；
也不把原双引擎要求或 Important 历史问题无声降级。I1/R1 仅采用最新指定向量
处置，原问题记录、旧失败、旧停止与全部已用/冻结/不合格额度不变。

## 6. 明确不批准的行动

即使用户批准第 1 节，仍不授权源码替换/候选提升、main 合并、发布、PS5 兼容声明、
P2/P3 实现/实验、真实采集/载荷/取消/清理/持久化、安全设置或信任变更、全 C12
验收、新测试/审查/变异/修复/重跑、安装或新额度。已有运行字段不改变。
原 accepted store/raw coordinator 及其依赖不修改，不从本提案继承更宽授权。

## 7. 用户决定后如何记录（不是现在执行）

- 若明确批准 A 以及第 1/4/5/6 节：仅另建有限批准记录，引用本提案固定版本、
  用户实际回复及基线；更新状态/恢复文档并普通提交推送。不得回写旧回执的
  FormalAcceptance=false，不启动新测试，不将候选复制至原产品路径。
- 若选择 B 或不同条件：维持未验收；先起草精确缺失证据计划，未批准不执行。
- 若输入缺失、源码/证据漂移或出现矛盾：停止作出验收结论，保留原记录并报告；
  不修复、不重跑、不自动召回审查。缺失 ignored 原始记录不能视为仍可本机复核。

## 8. 本轮文档范围、自检和恢复

仅新增本提案，更新 current-status 和 repository-recovery；无源码/测试/旧报告改动。
采用 writing-plans 的范围、依赖和逐步核对方法；这是文档验收提案而非实施计划，
因此没有新 RED/GREEN、实现任务、代理派发或执行方式选择，不能机械启动这些环节。

自检：已对应原合同输入/输出/预算和设计 P1 唯一边界；列出版本、证据和 23+9
全部限制；PS5 与 mutant 给出明确推荐而非 TBD；待批准状态、原字段和推广门槛
分离。提交前只核链接、固定身份、差异范围与保存状态，不宣称新运行验证。

换机从 codex/c12-b01-task9-coordinator 最新提交恢复，先读本提案、最新审查、
current-status 与 repository-recovery。不要从 main、clone、文档存在或历史“继续”
推定验收/运行授权；ignored 原始记录、引擎缓存和凭据不在 Git 交付范围。
