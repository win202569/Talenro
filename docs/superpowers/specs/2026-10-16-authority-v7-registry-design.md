# Task11 第一切片：v7 注册表与独立字面量向量设计

最新实施：[metadata计划](../plans/2026-10-17-authority-v7-metadata-registry.md)在c549a519后
获Native/inline批准，但首轮在普通runtime退出检查失败后停止，见[未验收记录](../../roadmap/2026-10-17-task11-metadata-registry-implementation.md)。
只新增第一条inventory测试，未交付metadata API/fixture；不得重跑或视为TDD RED通过。
以下计划待批准等描述保留为更早时点。

最新进展：名称切片已完成；九项/27项/16项附录均书面确认后，
[最终146项本地metadata对照](../../roadmap/2026-10-17-task11-final-metadata-reconciliation.md)
为146/0。[metadata-only实施计划](../plans/2026-10-17-authority-v7-metadata-registry.md)
已编写，待审阅和选择执行方法；尚未实施，不是独立认证或部署验收。
下文早期范围决定、待审阅与93/53计数保留为历史过程，不代表最新状态。

最新范围决定：用户在7c6194bb后批准首切片缩为名称/object只读目录。
[书面范围修订](2026-10-16-authority-v7-name-directory-design.md)待审阅；下文完整
元数据设计作为后续目标保留，不再据其API/文件清单自动启动首切片实施。

状态：书面设计已批准；C1附录随后在247f1f41后获用户批准，作为本地对照依据。
当前计划对照停在[分组元数据疑点C2及计数更正](../../roadmap/2026-10-16-task11-registry-reconciliation.md)，
没有可执行计划或实施批准。此前147误计表头，正确为146个V1对象；协议表本身不变。
后续C2方法已本地批准；[完整逐项初审表](../../roadmap/2026-10-16-task11-registry-metadata-audit.md)
记录93项有据、53项待闭合，尚无可执行元数据实施计划；未默改为仅名称目录。
基线：21b55a99b6f625e88fa8ca42c4cea8de68622b76，现有开发分支
codex/c12-b01-task9-coordinator。本文不是已批准实施计划，不授权代码、测试、
fuzz、数据库、Docker、诊断或 main 合并操作。

## 1. 意图、依据和成功含义

为 Task12–14 及后续 Task15 提供唯一、可审查的 v7 schema 名称和策略元数据
目录，先关闭名称、角色、证据归属和字面量向量的一致性缺口，不一次实现完整
Task11 的规范化、验签和 evidence-bundle 引擎。

权威依据是已批准 [v7 修订 §4.1 及各 object 定义小节](2026-08-24-nodecontrol-authority-v7-upgrade-design.md)。
[原 Batch01 Task11](../plans/2026-08-23-node-pop-control-plane-01-contracts-schema-authority.md)
是实施分解，不得覆盖规格。当前只读提取 §4.1 表得到 **146 行、146 个唯一
object 名称、146 个唯一 schema**。实现时须把这份完整固定集合写成独立字面量，
不在运行时读取 Markdown、不通过大小写转换或 object 名称推导 schema。

成功仅表示完整目录/元数据及其正反向向量一致。它不表示146个 Go DTO、codec、
签名验证器或生产调用路径已实现，不表示 Task11/Batch01/C12 整体验收通过。

## 2. 已确认冲突与兼容性决定

v7 §4.1 和 authority/v7_migration_grant.go:authorityV7DomainDigest 使用：

```text
SHA-256(ASCII("talenro.c12." + schema) || 0x00 || JCS(body))
```

原计划 Global Constraints 与 Task11 Step2 写成不带此前缀的摘要。这两处存在
文字冲突，不能据此构造另一种 v7 摘要。本设计按已批准规格保留前缀；**不改变
wire protocol、既有摘要、私有 helper、历史向量或 approved spec-set 内容/摘要**。
原计划只加醒目冲突提示，保留旧文字供追溯；后续局部实施计划必须引用正确式。

三个历史 object：AuthorityEffectCommitmentV1、AuthorityEffectResolutionV1、
ActivationDecisionEvidenceV1 不进入146项 v7 目录；其已有 Abort/serving transcript
保持原样，不能把“未登记为新 v7 schema”解释为删除历史协议支持。

签名 metadata 的 domain 仍是 talenro.c12.signature-envelope.v1 加 NUL 与
JCS(metadata)，不与 body digest domain 混用。本切片不实现或修改签名逻辑。

## 3. 方案比较和选择

采用**完整名称集合、限元数据能力的独立注册表**：所有146项一次精确登记，明确
哪些只是对象名、允许的证据形式与外部签名角色，不提供认证成功值。代价是后续
codec 还需逐项接线；好处是不会把部分私有 fixture 的支持范围冒充全体协议。

不采用“从现有 helper switch 自动生成”：它们是局部、私有的用途/fixture 策略，
既不覆盖全表，也不是生产 trust policy。也不一次实现完整 Task11：规范化、
验签、信任策略及 bundle 验证属于后续切片，扩大本轮会混淆目录与授权的证据。

## 4. 注册表职责和拟定接口

拟定在 internal/nodecontrol/contracts/schema_registry.go 提供唯一查询入口：

```go
func LookupAuthorityV7Schema(schema string) (AuthorityV7SchemaDescriptor, error)
func AuthorityV7Schemas() []AuthorityV7SchemaDescriptor
```

Descriptor 的信息固定为 Schema、ObjectName、EvidenceKinds、SignerRoleOrNull。
Schema/ObjectName 是 §4.1 精确字面量。ObjectName 是**协议 object 标签**，不是
reflect.Type 或已经实现的 Go 类型凭据，不提供 NewBody/Decode/Verify/Sign 方法。
EvidenceKinds 仅可包含规格允许的 database_immutable_body、external_signed_envelope；
不作为 bundle evidence 的 request/container 使用显式空集合，而非默认赋予一种。
依已批准C1澄清，SignerRoleOrNull表示通用envelope角色，与EvidenceKinds独立；
外部签名对象须有其定义小节唯一指定的角色；非通用envelope项使用显式null。
C1请求使用claim_v1_security_admin_request且EvidenceKinds=[]；空集合不免除请求认证。

证据形式、存储位置和签名角色不是同一概念：外部 envelope 可存入 DB，但这不
赋予 DB 自签权，也不允许把它换成裸 database_immutable_body。request 经鉴权
传输不自动等于可当作外部 attestation 验证；aggregate/container 不能凭名称
继承 provider 签名角色。各项必须逐条对照其定义小节，不按前缀或邻近项猜测。

注册表不保存私钥、公钥、key ID、trust root 或当前 policy version；这些属于
后续 TrustedSignerPolicy 的独立输入。ed25519 和 ecdsa-p256-sha256 是规格规定
的算法词汇，不是“任意角色允许任选算法”的授权。后续验签器仍须精确匹配角色、
key、root、policy、algorithm；目录查询不得返回 VerifiedEnvelope 或授权 capability。

实现将使用固定内部表，无运行时登记、caller override、环境变量或文件热更新。
列表按 schema 的 ASCII/bytewise 顺序返回；返回 descriptor 及嵌套集合均防御性
复制。修改调用方得到的值不能影响以后查询。查找精确区分大小写，不做 TrimSpace、
别名、版本降级或 case conversion。

未知/空/大小写变体/历史同名项返回唯一有限 ErrUnknownAuthorityV7Schema，失败
descriptor 为零值；错误文本不回显 caller 输入。已知条目也只说明登记事实。
不提供接受 caller 注册表后返回“验证成功”的替代入口。

## 5. 完整元数据对照与停止规则

后续实施计划必须将146项名称、各自 evidence kind/role 及规格小节来源逐行对照，
冻结为 schema-registry.v1.json 的字面量 rows，并在计划中列出完整审查清单。
来源记录只作审查依据，不进入 body/envelope wire 格式或摘要。

每个角色/证据形式只能来自对应 object 定义的明示契约。若角色、签名资格或
证据形式存在未明示/冲突，必须报告具体 schema/小节并暂停该交付，取得书面
契约澄清后再继续；不得用 unknown、通配 role、fallback、猜测值或“将来补齐”
条目使全表校验通过。**本设计不宣称已完成146项角色和证据资格的逐项认证**。
它冻结完整覆盖和不猜测的准入规则；完整对照是实施前计划的必需交付，不能留到
写产品代码时决定。若对照发现需要改动原协议，另行提请批准，不在本设计中默改。

## 6. 文件边界与现有实现关系

未来第一切片可创建（本轮不创建这些文件）：

- internal/nodecontrol/contracts/schema_registry.go：固定目录及只读查询。
- internal/nodecontrol/contracts/schema_registry_test.go：独立期望集、拒绝与所有权测试。
- testdata/c12/authority-v7/schema-registry.v1.json：完整字面量元数据目录。
- testdata/c12/authority-v7/schema-registry-golden.v1.json：仅本切片的目录与摘要域向量，
  明确标记 coverage=registry-and-domain-framing，不冒称完整 canonical envelope golden。

这份局部 golden **不替代**原 Task11 的 canonical-envelope-golden.v1.json。
本阶段不创建完整 codec 的占位文件，也不提前声明 CanonicalBodyDigest、
ParseCanonicalBody、VerifySignatureEnvelope、VerifyCanonicalEvidenceBundle 已交付。

authority/v7_migration_grant.go 的 domain helper、strict JSON、私有 signer policy，
v7_down_authorization.go、v7_staging_import.go 及 stateformat/stateverify 均保持原样；
不得从 contracts 反向 import authority，也不为了复用而导出 fixture signer。
已有私有 role 映射仅作为差异核对资料，不能用作146项期望值的自动 oracle。
后续统一 codec 迁移须单独列出调用方、兼容向量和迁移前后证据，本切片不接线。

不改 SQL、schema、生成物、OpenAPI、provider、reader/readiness、权限、安全设置或
canonical approved spec-set；不创建 Batch01 集成清单以绕过 Task18。

## 7. 独立向量和后续验收要求

本节规定未来测试，不代表本轮已运行，具体离线命令和工具边界由后续计划冻结。

1. 精确覆盖146项 schema/object 集合，正反向一对一；元数据逐条与独立期望字面量
   比较。测试不得调用生产查询函数生成自己的 expected，也不以“各146项”代替
   内容比较；missing+extra 即使抵消总数仍须失败。
2. 对 role/evidence kind 的每种实际类别给出正例和单字段变异反例；覆盖未知 schema、
   大小写/空白/别名、重复 schema、重复 object、漏项、多项、错误 role、证据形式
   替换及三个历史 object 误登记。JSON fixture 拒绝额外/缺失/重复字段和尾随内容。
3. 防御性复制：修改列表、descriptor 或 EvidenceKinds 后，再查仍返回原值；稳定
   排序，不让 map 遍历顺序影响 snapshot 或错误。测试不改变产品注册表本身。
4. 摘要域向量逐字保存 schema、已经 canonical 的 body 字节、完整 preimage hex、
   期望 SHA256；期望值通过独立标准 SHA256 计算/人工拆分核对，不从私有生产
   helper 生成。本切片的测试可重算 framing，但不新增公开 canonical/digest API。
   包含漏前缀、双前缀、漏 NUL、schema 改变、body 字节改变的区分向量。
5. 摘要 framing 测试不证明 body schema/JCS/签名合法；合成 body 必须明确标为
   framing-only，不能当成完整 v7 正例。完整 envelope 正例、Ed25519/P-256 验签、
   low-S、可信时间、bundle DAG 和业务 admission 验收留给后续批准切片。
6. 先固定独立 oracle，再 RED/GREEN 实施；测试不得借助 ignored SDK/原电脑日志
   重建 fixture。所有必要文本随 Git 保存，不需要新依赖、网络或真正私钥。

实现验收须同时满足完整名称/元数据对照、正反控制、复制/稳定性、摘要域不兼容
反例和明确的局部覆盖声明；任何未澄清条目阻塞，不计作部分生产注册成功。
不把向量通过当成所有 schema codec 或全项目安全/物理验收通过。

## 8. 已完成的设计自检与交接

已检查本设计未把完整表与完整 codec 混为一谈，未创建新的 trust authority，未
把局部 golden 占用完整 golden 文件名，也未改写既有 protocol digest。角色/
证据来源不足的处置是明确停止并提请澄清，不授权实现者填默认值。

当前仅完成书面设计；未运行测试、Go/native、SDK/Docker/DB 操作。旧 error87、
matrix/fuzz/超时、Task9/10/15/18、物理 PITR、Batch01、C12、生产及 main 门槛不变。
保留现有工作区/开发分支，按 AGENTS.md 保存文档；Git 不恢复本机缓存和原始回执。

本文及C1局部附录已获批准；完整策略对照仍停在C2。先完成分组契约裁决，
再继续146项对照和实施计划。canonical集合登记、部署兼容和实施批准均未完成。
