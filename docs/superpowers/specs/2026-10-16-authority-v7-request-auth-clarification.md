# Task11 C1 局部澄清提案：请求认证与证据签名分离

状态：用户在247f1f41后以“是”批准本文作为C1局部澄清依据，允许继续对照/规划。
本文未纳入canonical approved spec-set，不授权代码、信任配置、密钥或协议部署变更。
原提案基线29f35563ac3d103aafdddd2b3af8268853cd9262；新增角色是本地已批契约补充，
不是原规格已有规则。对象数更正为146（旧147误计表头），其余对象为145项。

## 1. 目的、依据及非目标

关闭 [C1 计划阻塞](../../roadmap/2026-10-16-task11-registry-planning-blocker.md)：
v7 §4.3 要求 PublishClaimV1CredentialPolicyRequestV1 按 provider 固定的
security-admin request signer policy 认证，但未明确该请求采用通用 envelope
时的角色和证据资格；ClaimV1CredentialPolicyV1 响应的 claim_v1_security_policy
角色不能代给请求。

依据：[v7 §4.1/§4.3](2026-08-24-nodecontrol-authority-v7-upgrade-design.md) 与
[已批准注册表设计 §4–5](2026-10-16-authority-v7-registry-design.md)。本文拟补充
这个请求的契约，不声称原规格已经包含以下新增角色，也不泛化到其余145项。
保留 schema、body 字段、body digest 域、响应字段及其签名角色。

不实现验签器、provider、请求路由、生产 trust bundle、策略发布、证据验证器或
注册表，不迁移已有消费者，不宣称完成146项对照或完整 Task11。

## 2. 方案比较与推荐

**推荐：通用 envelope＋专用请求角色＋独立证据资格。** 复用 §4.1 的字段、
metadata 签名输入、算法和严格编码规则；以 schema/role/trusted policy 将请求
与响应区分。代价是新增一个必须显式批准、配置和兼容核对的角色。

另一方案是另定义 security-admin 的独立请求认证 wire 协议。它同样能隔离用途，
但必须另冻结 transcript、传输和错误契约，当前没有可直接引用的完整定义，
会扩大本切片。仅把请求登记为 null/空证据集合而不明确如何认证，不是解决方案。

以下各节已获本地设计批准；canonical登记与部署生效仍须各自的显式流程。

## 3. 请求与响应的精确分离

| 项目 | 请求 | 响应 |
| --- | --- | --- |
| Object | PublishClaimV1CredentialPolicyRequestV1 | ClaimV1CredentialPolicyV1 |
| Schema | publish-claim-v1-credential-policy-request.v1 | claim-v1-credential-policy.v1 |
| 通用 envelope signer_role | **本地批准新增 claim_v1_security_admin_request** | 保留 claim_v1_security_policy |
| 信任归属 | provider 固定的、独立管理的 security-admin request signer policy | 原有响应签名策略 |
| Bundle EvidenceKinds | 显式空集合，不作可复用 bundle evidence | 保持 external_signed_envelope，不改既有响应要求 |

请求采用 §4.1 的九字段 envelope，schema 必须为上表精确请求值；body 仍是
§4.3 已冻结的原字段，不能新增字段或用响应 body 代替。body digest 保持：

```text
SHA-256(ASCII("talenro.c12." + schema) || 0x00 || JCS(body))
```

签名仍覆盖原 §4.1 七字段 metadata，输入为
ASCII("talenro.c12.signature-envelope.v1") || 0x00 || JCS(metadata)。
不增加新加密算法或另一个签名 domain；角色、schema、key、policy、root 和
algorithm 都必须签名绑定。Ed25519/P-256 的编码及 low-S 规则不变；具体 key
仅能使用其可信策略固定的单一算法，不允许请求自行任选算法。

请求验签只证明匹配该调用的请求认证条件，不返回通用证据成功值、生产就绪或
任意发布能力；还须逐项检查 §4.3 原有 body、expected Head/policy、sequence、
activation/namespace、nonce 和 exact-retry/conflict 条件。缺少请求认证、只有
TLS/连接身份、只有 body digest、只有合法响应或 signer_role 声明均不得发布策略。

## 4. 信任与启动顺序

provider 验证请求所用的 security-admin policy 必须独立预置并受管理控制，
不得从本次请求的 accepted_claim_v1_key_set_digest、正在发布的 credential
policy、未验证响应或 caller 传入的 key/root 推导。否则首次策略发布会出现
“用待批准策略批准自己”的循环。

本提案要求请求和响应角色使用分离的授权 key identity/公钥及签名能力；响应
signer 不因能签响应而能提交策略管理请求，普通 claim-v1 caller 也不因属于
accepted key set 而取得 security-admin 权限。不在本轮生成密钥或选择真实
key ID、trust root、policy version；部署值由未来独立可信策略管理步骤冻结。
不存在可用于生产的默认开发 key、fixture key 或 wildcard-role 回退。

未知/撤销 key、错误角色/root/policy/algorithm、非规范 bytes、无效签名或任何
认证不确定结果必须在 provider mutation 前拒绝，不把请求体或密钥材料写入
错误文本。认证不通过不授权弱策略重试；合法 body 的 exact retry 仍由既有
policy_update_id/body 规则处理，本提案不扩张重放或权限生命周期。

## 5. 对注册表设计的局部澄清

依本地批准，注册表 descriptor 的 SignerRoleOrNull 明确表示“该 schema 使用
通用 envelope 时允许的唯一角色”，而非“只要有角色就能作 bundle evidence”。
它与 EvidenceKinds 是独立轴：本请求可为非 null 角色而 EvidenceKinds 为空。
这不新增运行时字段或公共验证接口，也不让目录查找替代请求认证。

本地批准的本请求单行元数据为：

```text
Schema: publish-claim-v1-credential-policy-request.v1
ObjectName: PublishClaimV1CredentialPolicyRequestV1
EvidenceKinds: []
SignerRoleOrNull: claim_v1_security_admin_request
Source: v7 §4.3 + 本澄清附录（已本地批准，未完成canonical集合登记）
```

后续 bundle verifier 对本请求必须拒绝 external_signed_envelope 和
database_immutable_body 两种包装，即便通用验签器认定其签名有效。
将签过名的请求写入日志/DB 或 provider 永久请求档案，不会把它升级为可复用
bundle evidence。已有响应证据资格、request_digest 引用和存储义务不变。

其余请求、嵌入对象和 container 的角色/证据资格仍须各自明示，不能把本行规则
复制给所有 RequestV1，也不能把本次批准当作整张146项策略表的认证。

## 6. 未来验收要求（本轮未实施/未运行）

- 独立字面量正例分别覆盖请求与响应；拒绝互换 role/schema、把响应 key 当请求
  key、普通 accepted key 冒充 security-admin、caller 提供 key/root 以及 bootstrap
  策略自授权。不得以“签名数学成立”代替策略匹配。
- 认证成功也须保留 body/Head/policy/sequence/activation/nonce/conflict 拒绝；
  用 recording provider 断言所有拒绝路径 mutation0，不运行真实策略发布。
- 使用合法请求签名尝试两种 bundle evidence kind 都应拒绝；响应合法 envelope
  的既有资格不变。null/空 evidence 元数据不能关闭 mandatory request authentication。
- role/key/policy/root/algorithm/body_digest 的单字段变异、缺失/额外/重复字段、
  非规范编码和 signature mismatch 均拒绝；摘要前缀和 signature metadata 域不变。
- fixture 使用独立测试密钥与固定向量，永不成为生产默认；无 live provider、
  新依赖、网络、Docker/DB 或完整矩阵。具体命令和工具界限须后续实施计划批准。

上述验收属于后续相应 codec/provider 实施，不偷渡进当前仅注册表/向量切片。
第一切片只冻结获批单行元数据及相应拒绝向量，不声称已实现请求认证能力。

## 7. 兼容性、审批与下一步

本文明确提出新增角色、用途限制和策略/key 分离要求，不是对未审阅部署的兼容性
证明。现有 provider 是否使用其他 request wire/role 未核对，兼容性 UNKNOWN；
不能迁移、开启生产或重新解释旧请求。未来接线前必须冻结版本/策略管理、兼容
向量和拒绝旧格式的边界，不提供无签名/旧角色/响应角色的自动兼容回退。

当前 canonical 四份规格、c12-spec-set.v1.json 和其摘要均不改。获批后可将
本附录作为本地开发设计的 C1 对照依据，并明确更新注册表设计的语义说明；
是否纳入 canonical approved spec-set、如何更新批准摘要/跨册消费者仍须独立
的显式流程，不能因为存在此文件就改变 canonical 集合。本附录获批也不关闭
该集合的登记/部署兼容门槛。

当前C1为 LOCALLY CLARIFIED；canonical登记/部署兼容仍开放。已恢复的146项
对照遇到分组疑点C2；见恢复记录，不以C1批准推定全表完成。计划及执行方式另行批准，
不得直接写产品代码/运行测试；完整 Task11、物理/C12/生产/main 门槛不变。

自检：请求与响应角色没有复用，认证和 evidence 资格独立，没有用 null 取消
认证，没有自授权 bootstrap，新增规则均标为提案而非旧规格事实，没有私钥、
默认生产配置、实现占位符或虚构测试结果。
