# Task11 C2：元数据准入统一澄清附录

状态：用户在f6814efa后以“是”批准本文作为本地对照依据，未授权实施或部署。
现有[完整初审表](../../roadmap/2026-10-16-task11-registry-metadata-audit.md)记录93项有据、
53项未闭合；批准统一方法不等于认证全表。canonical登记及兼容性仍开放。
依据：[注册表设计](2026-10-16-authority-v7-registry-design.md)、
[已本地批准C1](2026-10-16-authority-v7-request-auth-clarification.md)、
[146项来源索引](../../roadmap/2026-10-16-task11-registry-reconciliation.md)。
本轮无产品代码、测试、配置、密钥或原批准规格集变更，无负载执行。

## 1. 目的与选择

推荐采用“明示能力准入＋独立请求认证＋未决项阻塞出版”：统一目录的分类方法，
不因名称、存储位置、签名父对象或相邻段落推定权限。代价是仍须逐项提交来源，
但可一次汇总例外，不再按每个同类对象重复建立不同默认规则。

不采用所有Request自动继承provider角色、所有unsigned body自动变DB evidence，
也不默改为仅名称目录；后者虽然更小，但会删减已批准的元数据交付。

## 2. 三个独立维度

| 维度 | 含义 | 不意味着 |
| --- | --- | --- |
| 请求认证 | 调用方能否提出该操作，由对应调用契约负责 | schema在目录中就可调用；无envelope角色就无需认证 |
| SignerRoleOrNull | 通用envelope允许的精确角色；不采用该封装时为null | 非null就能进入bundle；替代可信key/root/policy检查 |
| EvidenceKinds | 该schema可作为独立bundle项的形式集合 | 该项在每个message中都被需要或允许；替代message-specific exact-cover |

C1请求保持专用角色claim_v1_security_admin_request、EvidenceKinds=[]与独立管理员
认证；不推广其角色、key策略或签名封装到其他请求。认证模式不塞进SignerRoleOrNull
或用null表示认证关闭；本切片不新增认证运行时API或通用认证fallback。

## 3. 各类对象的统一裁决规则（拟议）

1. **请求**：只有对应契约明确采用通用envelope且明示角色，才登记该角色；
   认证渠道、已有capability或body/preimage校验不自动构成通用签名封装。
   “同一signer role”须追踪到具体先行对象/角色，不能取同段最近的角色词。
   认证方式或封装资格确有歧义的请求标为对照未决，不能直接填null。
2. **嵌入体、集合与container**：父对象签名绑定其内容，不给子对象独立签名或
   evidence资格。独立body digest只证明域分离内容标识，不证明签发者或DB提交。
   若完整对应契约仅允许嵌入/集合传输，且无独立证据通道，则可经逐项审查登记
   SignerRoleOrNull=null、EvidenceKinds=[]；不能仅凭名称后缀套用此规则。
3. **临时/事务投影**：不因来自DB就登记database_immutable_body。临时Head应走
   规格明示的attestation路径；事务内inventory/state不因被引用就变成已提交
   immutable row。只有明确的immutable写入/身份/preimage及对应证据消费契约
   才可登记DB形式；未明确者保留未决，不以严格默认值伪称协议已有禁止规则。
4. **外部签名对象**：角色须来自其明示定义或无歧义引用；其外部证据资格也须
   独立有据。DB保存外部envelope不改变其证据形式；不得降格为裸DB body。

这些规则不新造签名角色、不增加协议对象、不删除既有证据通道。任何与原契约
冲突或需要新增行为的行必须作为例外单列，不能以“统一规则”改写它。

## 4. 完整对照与阻塞机制

对146项逐行记录：精确schema/object、定义位置、请求认证契约位置（适用时）、
envelope角色依据、独立证据资格依据、结论和例外。此为计划审查附表，不新增wire字段。
每项只能处于“有据已裁决”或“未决”；未决不是生产enum，也不进入产品目录。
null/[]必须有肯定分类依据，不能用于隐藏未找到定义或未完成阅读。

全表对照后一次列出所有剩余例外及其具体问题。任何未决行阻塞完整注册表产品
交付和实施就绪声明；不得省略它后仍称覆盖146项，也不得让调用方选择另一份表。
本附录批准不等于146项结论自动获认证，亦不保证后续无新的实质契约缺口。

## 5. 后续验证边界

未来独立字面量向量须覆盖：有角色但空EvidenceKinds的C1；嵌入体不继承父角色；
临时Head拒绝DB形式；合法外部envelope不降格；请求认证不因null/[]被跳过；
“未找到来源”不得通过；同数目missing+extra替换仍失败。角色/形式变异测试
只证明目录规则，不宣称真正完成验签、认证或业务mutation0验证。

message-specific bundle exact-cover、动态trust policy、可信时间、签名验证与
物理DB证据仍属后续实现。保持摘要前缀、历史schema排除、防御性复制和有限错误。
具体测试命令由后续实施计划冻结；本轮未运行任何测试。

## 6. 审批与恢复

本提案只澄清本地目录准入，不登记canonical spec-set、不部署策略、不推断现有
provider兼容性。兼容性UNKNOWN；原批准规格及摘要保持不动。
批准本文后，下一步是按统一规则完成146项审查附表并集中报告例外，再写可执行
计划；实施仍需计划及执行方式批准。当前C2方法已本地批准，全表仍有53项待闭合，
全部生产/main门槛不变。

自检：没有把absence当作null的充分依据，没有把签名等同evidence资格，没有
请求免认证路径，没有把统一规则冒称全表认证，没有实现占位或虚构测试结果。
